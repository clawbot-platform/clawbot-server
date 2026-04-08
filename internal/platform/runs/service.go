package runs

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"clawbot-server/internal/platform/audit"
	"clawbot-server/internal/platform/scheduler"
	"clawbot-server/internal/platform/store"
)

type DependencyConfig struct {
	ClawmemBaseURL               string
	InferenceBaseURL             string
	GuardrailTimeout             time.Duration
	HelperTimeout                time.Duration
	DisableLocalOllamaGuardrails bool
	EnableCompactDualPayload     bool
	PolicyBundleID               string
	PolicyBundleVersion          string
	Environment                  string
}

type Manager struct {
	repo      Repository
	tx        store.Transactor
	audits    audit.Recorder
	scheduler scheduler.Service
	db        store.DBTX
	memory    MemoryClient
	inference InferenceClient
	deps      DependencyConfig
}

func NewManager(db store.DBTX, tx store.Transactor, repo Repository, audits audit.Recorder, scheduler scheduler.Service) *Manager {
	return NewManagerWithIntegrations(db, tx, repo, audits, scheduler, NewNoopMemoryClient(), NewNoopInferenceClient(), DependencyConfig{})
}

func NewManagerWithIntegrations(
	db store.DBTX,
	tx store.Transactor,
	repo Repository,
	audits audit.Recorder,
	scheduler scheduler.Service,
	memoryClient MemoryClient,
	inferenceClient InferenceClient,
	deps DependencyConfig,
) *Manager {
	if memoryClient == nil {
		memoryClient = NewNoopMemoryClient()
	}
	if inferenceClient == nil {
		inferenceClient = NewNoopInferenceClient()
	}
	if deps.ClawmemBaseURL == "" {
		deps.ClawmemBaseURL = memoryClient.BaseURL()
	}
	if deps.InferenceBaseURL == "" {
		deps.InferenceBaseURL = inferenceClient.BaseURL()
	}
	if strings.TrimSpace(deps.PolicyBundleID) == "" {
		deps.PolicyBundleID = "ach-governance"
	}
	if strings.TrimSpace(deps.PolicyBundleVersion) == "" {
		deps.PolicyBundleVersion = "2026.1"
	}
	if strings.TrimSpace(deps.Environment) == "" {
		deps.Environment = "unknown"
	}

	return &Manager{
		repo:      repo,
		tx:        tx,
		audits:    audits,
		scheduler: scheduler,
		db:        db,
		memory:    memoryClient,
		inference: inferenceClient,
		deps:      deps,
	}
}

func (m *Manager) List(ctx context.Context) ([]Run, error) {
	return m.repo.List(ctx, m.db)
}

func (m *Manager) Get(ctx context.Context, id string) (Run, error) {
	return m.repo.Get(ctx, m.db, id)
}

func (m *Manager) StartRun(ctx context.Context, runID string, input ExecuteRunInput, actor string) (ExecuteRunResult, error) {
	return m.executeRun(ctx, runID, nil, input, actor)
}

func (m *Manager) ExecuteCycleRun(ctx context.Context, runID string, cycleID string, input ExecuteRunInput, actor string) (ExecuteRunResult, error) {
	cycleID = strings.TrimSpace(cycleID)
	if cycleID == "" {
		return ExecuteRunResult{}, fmt.Errorf("cycle id is required")
	}
	input.CycleID = &cycleID
	return m.executeRun(ctx, runID, &cycleID, input, actor)
}

func (m *Manager) ReviewAction(ctx context.Context, runID string, input ReviewActionInput, actor string) (Run, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return Run{}, fmt.Errorf("run id is required")
	}
	normalizeReviewActionInput(&input, actor)
	if err := validateReviewActionInput(input); err != nil {
		return Run{}, err
	}

	runForPolicy, err := m.repo.Get(ctx, m.db, runID)
	if err != nil {
		return Run{}, err
	}
	var cycleForPolicy *Cycle
	if input.CycleID != nil && strings.TrimSpace(*input.CycleID) != "" {
		cycle, cycleErr := m.repo.GetCycle(ctx, m.db, runID, strings.TrimSpace(*input.CycleID))
		if cycleErr != nil {
			return Run{}, cycleErr
		}
		cycleForPolicy = &cycle
	}

	policyDecision, err := m.evaluateAndRecordPolicy(ctx, "run.review."+input.Action, actor, &runForPolicy, cycleForPolicy, nil, nil, nil, nil)
	if err != nil {
		return Run{}, err
	}
	if !policyDecision.Allow {
		return Run{}, policyDeniedError(policyDecision)
	}

	var updated Run
	err = m.tx.InTx(ctx, func(ctx context.Context, q store.DBTX) error {
		run, err := m.repo.Get(ctx, q, runID)
		if err != nil {
			return err
		}
		priorStatus := run.Status
		newStatus, err := mapReviewActionToRunStatus(input.Action)
		if err != nil {
			return err
		}

		run.Status = newStatus
		run.ReviewMetadataJSON = mergeJSONObjects(run.ReviewMetadataJSON, map[string]any{
			"review_action": map[string]any{
				"action":        input.Action,
				"reviewer_id":   input.ReviewerID,
				"reviewer_type": input.ReviewerType,
				"rationale":     input.Rationale,
				"at":            time.Now().UTC().Format(time.RFC3339Nano),
				"prior_status":  priorStatus,
				"new_status":    newStatus,
			},
		})

		updated, err = m.repo.Update(ctx, q, run)
		if err != nil {
			return err
		}

		reviewInput := input
		if reviewInput.PolicyDecisionID == nil {
			reviewInput.PolicyDecisionID = &policyDecision.ID
		}
		if _, err := m.repo.RecordReviewAction(ctx, q, runID, reviewInput, priorStatus, newStatus); err != nil {
			return err
		}

		if input.CycleID != nil && strings.TrimSpace(*input.CycleID) != "" {
			cycle, err := m.repo.GetCycle(ctx, q, runID, strings.TrimSpace(*input.CycleID))
			if err != nil {
				return err
			}
			nextCycleStatus, err := mapReviewActionToCycleStatus(input.Action)
			if err != nil {
				return err
			}
			cycle.Status = nextCycleStatus
			if _, err := m.repo.UpdateCycle(ctx, q, cycle); err != nil {
				return err
			}
		}

		if err := recordAudit(ctx, m.audits, q, "run", "run.review.action", actor, updated.ID, map[string]any{
			"action":       input.Action,
			"prior_status": priorStatus,
			"new_status":   newStatus,
			"reviewer_id":  input.ReviewerID,
		}); err != nil {
			return err
		}
		if _, err := m.repo.AppendGovernanceAuditEvent(ctx, q, GovernanceAuditEventInput{
			ActorID:          input.ReviewerID,
			ActorType:        input.ReviewerType,
			ActionType:       "run.review." + input.Action,
			TargetRunID:      &updated.ID,
			TargetCycleID:    input.CycleID,
			PolicyDecisionID: &policyDecision.ID,
			PayloadSummary: mustJSON(map[string]any{
				"prior_status": priorStatus,
				"new_status":   newStatus,
				"rationale":    input.Rationale,
			}, map[string]any{}),
		}); err != nil {
			return err
		}

		return m.scheduler.RecordRunIntent(ctx, q, scheduler.Signal{
			RunID:   updated.ID,
			RunName: updated.Name,
			Status:  updated.Status,
			Actor:   actor,
			Reason:  "run.review.action",
		})
	})
	if err != nil {
		return Run{}, err
	}

	return updated, nil
}

func (m *Manager) Create(ctx context.Context, input CreateInput, actor string) (Run, error) {
	normalizeCreateInput(&input, actor)
	if err := validateCreateInput(input); err != nil {
		return Run{}, err
	}

	policyDecision, err := m.evaluateAndRecordPolicy(ctx, "run.create", actor, nil, nil, &input, nil, nil, nil)
	if err != nil {
		return Run{}, err
	}
	if !policyDecision.Allow {
		return Run{}, policyDeniedError(policyDecision)
	}
	input.ReviewMetadataJSON = mergeJSONObjects(input.ReviewMetadataJSON, map[string]any{
		"last_policy_decision_id": policyDecision.ID,
		"execution_ring":          input.ExecutionRing,
	})

	var created Run
	err = m.tx.InTx(ctx, func(ctx context.Context, q store.DBTX) error {
		var err error
		created, err = m.repo.Create(ctx, q, input)
		if err != nil {
			return err
		}
		runID := created.ID

		if err := recordAudit(ctx, m.audits, q, "run", "run.created", actor, created.ID, created); err != nil {
			return err
		}
		if _, err := m.repo.AppendGovernanceAuditEvent(ctx, q, GovernanceAuditEventInput{
			ActorID:          actor,
			ActorType:        "user",
			ActionType:       "run.create",
			TargetRunID:      &runID,
			PolicyDecisionID: &policyDecision.ID,
			PayloadSummary: mustJSON(map[string]any{
				"execution_mode": created.ExecutionMode,
				"execution_ring": created.ExecutionRing,
			}, map[string]any{}),
		}); err != nil {
			return err
		}

		return m.scheduler.RecordRunIntent(ctx, q, scheduler.Signal{
			RunID:   created.ID,
			RunName: created.Name,
			Status:  created.Status,
			Actor:   actor,
			Reason:  "run.created",
		})
	})
	if err != nil {
		return Run{}, err
	}

	return created, nil
}

func (m *Manager) Update(ctx context.Context, id string, input UpdateInput, actor string) (Run, error) {
	var updated Run

	err := m.tx.InTx(ctx, func(ctx context.Context, q store.DBTX) error {
		existing, err := m.repo.Get(ctx, q, id)
		if err != nil {
			return err
		}

		merged, err := mergeRun(existing, input)
		if err != nil {
			return err
		}

		updated, err = m.repo.Update(ctx, q, merged)
		if err != nil {
			return err
		}

		if err := recordAudit(ctx, m.audits, q, "run", "run.updated", actor, updated.ID, updated); err != nil {
			return err
		}

		return m.scheduler.RecordRunIntent(ctx, q, scheduler.Signal{
			RunID:   updated.ID,
			RunName: updated.Name,
			Status:  updated.Status,
			Actor:   actor,
			Reason:  "run.updated",
		})
	})
	if err != nil {
		return Run{}, err
	}

	return updated, nil
}

func (m *Manager) CreateCycle(ctx context.Context, runID string, input CreateCycleInput, actor string) (Cycle, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return Cycle{}, fmt.Errorf("run id is required")
	}

	if err := normalizeCreateCycleInput(&input); err != nil {
		return Cycle{}, err
	}

	runForPolicy, err := m.repo.Get(ctx, m.db, runID)
	if err != nil {
		return Cycle{}, err
	}
	policyDecision, err := m.evaluateAndRecordPolicy(ctx, "run.cycle.create", actor, &runForPolicy, nil, nil, &input, nil, nil)
	if err != nil {
		return Cycle{}, err
	}
	if !policyDecision.Allow {
		return Cycle{}, policyDeniedError(policyDecision)
	}

	var created Cycle
	err = m.tx.InTx(ctx, func(ctx context.Context, q store.DBTX) error {
		run, err := m.repo.Get(ctx, q, runID)
		if err != nil {
			return err
		}

		if run.RunType != string(RunTypeWeekRun) {
			return fmt.Errorf("cycles can only be created for week_run run types")
		}

		created, err = m.repo.CreateCycle(ctx, q, runID, input)
		if err != nil {
			return err
		}

		if err := recordAudit(ctx, m.audits, q, "cycle", "run.cycle.created", actor, created.ID, created); err != nil {
			return err
		}
		cycleID := created.ID
		if _, err := m.repo.AppendGovernanceAuditEvent(ctx, q, GovernanceAuditEventInput{
			ActorID:          actor,
			ActorType:        "user",
			ActionType:       "run.cycle.create",
			TargetRunID:      &run.ID,
			TargetCycleID:    &cycleID,
			PolicyDecisionID: &policyDecision.ID,
			PayloadSummary: mustJSON(map[string]any{
				"cycle_key":      created.CycleKey,
				"execution_ring": created.ExecutionRing,
			}, map[string]any{}),
		}); err != nil {
			return err
		}

		return m.scheduler.RecordRunIntent(ctx, q, scheduler.Signal{
			RunID:   run.ID,
			RunName: run.Name,
			Status:  run.Status,
			Actor:   actor,
			Reason:  "run.cycle.created",
		})
	})
	if err != nil {
		return Cycle{}, err
	}

	return created, nil
}

func (m *Manager) GetCycle(ctx context.Context, runID string, cycleID string) (Cycle, error) {
	return m.repo.GetCycle(ctx, m.db, strings.TrimSpace(runID), strings.TrimSpace(cycleID))
}

func (m *Manager) UpdateCycle(ctx context.Context, runID string, cycleID string, input UpdateCycleInput, actor string) (Cycle, error) {
	runID = strings.TrimSpace(runID)
	cycleID = strings.TrimSpace(cycleID)
	if runID == "" || cycleID == "" {
		return Cycle{}, fmt.Errorf("run id and cycle id are required")
	}

	var updated Cycle
	err := m.tx.InTx(ctx, func(ctx context.Context, q store.DBTX) error {
		run, err := m.repo.Get(ctx, q, runID)
		if err != nil {
			return err
		}

		existing, err := m.repo.GetCycle(ctx, q, runID, cycleID)
		if err != nil {
			return err
		}

		merged, err := mergeCycle(existing, input)
		if err != nil {
			return err
		}

		if input.MemoryNote != nil {
			note := strings.TrimSpace(*input.MemoryNote)
			if note != "" {
				namespace := composeMemoryNamespace(run, merged, input.AgentNamespace)
				snapshotRef, err := m.memory.PersistScopedNotes(ctx, namespace, MemoryWriteInput{Note: note})
				if err != nil {
					return err
				}
				if snapshotRef != "" {
					merged.MemorySnapshotRef = snapshotRef
					run.MemorySnapshotRefs = appendUnique(run.MemorySnapshotRefs, snapshotRef)
					run, err = m.repo.Update(ctx, q, run)
					if err != nil {
						return err
					}
				}
			}
		}

		updated, err = m.repo.UpdateCycle(ctx, q, merged)
		if err != nil {
			return err
		}

		if err := recordAudit(ctx, m.audits, q, "cycle", "run.cycle.updated", actor, updated.ID, updated); err != nil {
			return err
		}

		return m.scheduler.RecordRunIntent(ctx, q, scheduler.Signal{
			RunID:   run.ID,
			RunName: run.Name,
			Status:  run.Status,
			Actor:   actor,
			Reason:  "run.cycle.updated",
		})
	})
	if err != nil {
		return Cycle{}, err
	}

	return updated, nil
}

func (m *Manager) AttachArtifact(ctx context.Context, runID string, input AttachArtifactInput, actor string) (Artifact, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return Artifact{}, fmt.Errorf("run id is required")
	}
	if err := normalizeAttachArtifactInput(&input); err != nil {
		return Artifact{}, err
	}
	runForPolicy, err := m.repo.Get(ctx, m.db, runID)
	if err != nil {
		return Artifact{}, err
	}
	policyDecision, err := m.evaluateAndRecordPolicy(ctx, "run.artifact.attach", actor, &runForPolicy, nil, nil, nil, &input, nil)
	if err != nil {
		return Artifact{}, err
	}
	if !policyDecision.Allow {
		return Artifact{}, policyDeniedError(policyDecision)
	}

	var artifact Artifact
	err = m.tx.InTx(ctx, func(ctx context.Context, q store.DBTX) error {
		run, err := m.repo.Get(ctx, q, runID)
		if err != nil {
			return err
		}

		if input.CycleID != nil && strings.TrimSpace(*input.CycleID) != "" {
			if _, err := m.repo.GetCycle(ctx, q, runID, strings.TrimSpace(*input.CycleID)); err != nil {
				return err
			}
		}

		artifact, err = m.repo.CreateArtifact(ctx, q, runID, input)
		if err != nil {
			return err
		}

		run.ArtifactBundleRefs = appendUnique(run.ArtifactBundleRefs, artifact.URI)
		if _, err := m.repo.Update(ctx, q, run); err != nil {
			return err
		}
		artifactID := artifact.ID
		if _, err := m.repo.AppendGovernanceAuditEvent(ctx, q, GovernanceAuditEventInput{
			ActorID:          actor,
			ActorType:        "user",
			ActionType:       "run.artifact.attach",
			TargetRunID:      &run.ID,
			TargetArtifactID: &artifactID,
			PolicyDecisionID: &policyDecision.ID,
			PayloadSummary: mustJSON(map[string]any{
				"artifact_type": artifact.ArtifactType,
				"content_type":  artifact.ContentType,
			}, map[string]any{}),
		}); err != nil {
			return err
		}

		return recordAudit(ctx, m.audits, q, "artifact", "run.artifact.attached", actor, artifact.ID, artifact)
	})
	if err != nil {
		return Artifact{}, err
	}

	return artifact, nil
}

func (m *Manager) ListArtifacts(ctx context.Context, runID string) ([]Artifact, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return nil, fmt.Errorf("run id is required")
	}
	if _, err := m.repo.Get(ctx, m.db, runID); err != nil {
		return nil, err
	}
	return m.repo.ListArtifacts(ctx, m.db, runID)
}

func (m *Manager) UpsertComparison(ctx context.Context, runID string, input UpsertComparisonInput, actor string) (Comparison, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return Comparison{}, fmt.Errorf("run id is required")
	}
	if err := normalizeUpsertComparisonInput(&input); err != nil {
		return Comparison{}, err
	}

	var comparison Comparison
	err := m.tx.InTx(ctx, func(ctx context.Context, q store.DBTX) error {
		if _, err := m.repo.Get(ctx, q, runID); err != nil {
			return err
		}
		if input.CycleID != nil && strings.TrimSpace(*input.CycleID) != "" {
			if _, err := m.repo.GetCycle(ctx, q, runID, strings.TrimSpace(*input.CycleID)); err != nil {
				return err
			}
		}

		var err error
		comparison, err = m.repo.UpsertComparison(ctx, q, runID, input)
		if err != nil {
			return err
		}
		if _, err := m.repo.AppendGovernanceAuditEvent(ctx, q, GovernanceAuditEventInput{
			ActorID:     actor,
			ActorType:   "user",
			ActionType:  "run.comparison.upsert",
			TargetRunID: &runID,
			PayloadSummary: mustJSON(map[string]any{
				"review_status": comparison.ReviewStatus,
				"guardrail_present": guardrailStatusFromSummary(comparison.GuardrailSummary) == string(GuardrailStatusPassed) ||
					guardrailStatusFromSummary(comparison.GuardrailSummary) == string(GuardrailStatusFlagged),
			}, map[string]any{}),
		}); err != nil {
			return err
		}

		return recordAudit(ctx, m.audits, q, "comparison", "run.comparison.upserted", actor, comparison.ID, comparison)
	})
	if err != nil {
		return Comparison{}, err
	}

	return comparison, nil
}

func (m *Manager) GetComparison(ctx context.Context, runID string) (Comparison, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return Comparison{}, fmt.Errorf("run id is required")
	}
	return m.repo.GetComparison(ctx, m.db, runID)
}

func (m *Manager) RegisterModelProfile(ctx context.Context, input RegisterModelProfileInput, actor string) (ModelProfile, error) {
	normalizeModelProfileInput(&input)
	if err := validateModelProfileInput(input); err != nil {
		return ModelProfile{}, err
	}

	var profile ModelProfile
	err := m.tx.InTx(ctx, func(ctx context.Context, q store.DBTX) error {
		var err error
		profile, err = m.repo.RegisterModelProfile(ctx, q, input, actor)
		if err != nil {
			return err
		}

		return recordAudit(ctx, m.audits, q, "model_profile", "model_profile.registered", actor, profile.ID, profile)
	})
	if err != nil {
		return ModelProfile{}, err
	}

	return profile, nil
}

func (m *Manager) GetModelProfile(ctx context.Context, idOrName string) (ModelProfile, error) {
	idOrName = strings.TrimSpace(idOrName)
	if idOrName == "" {
		return ModelProfile{}, fmt.Errorf("model profile id or name is required")
	}
	return m.repo.GetModelProfile(ctx, m.db, idOrName)
}

func (m *Manager) DependencyHealth(ctx context.Context) (DependencyHealth, error) {
	health := DependencyHealth{
		Status: "healthy",
		Dependencies: []DependencyStatus{
			{Name: "postgres", Status: "healthy"},
			{Name: "clawmem", Status: "disabled", Endpoint: m.deps.ClawmemBaseURL},
			{Name: "inference", Status: "disabled", Endpoint: m.deps.InferenceBaseURL},
		},
	}

	var one int
	if err := m.db.QueryRow(ctx, "SELECT 1").Scan(&one); err != nil {
		health.Status = "degraded"
		health.Dependencies[0].Status = "down"
		health.Dependencies[0].Error = err.Error()
	}

	if strings.TrimSpace(m.deps.ClawmemBaseURL) != "" {
		health.Dependencies[1].Status = "configured"
	}
	if strings.TrimSpace(m.deps.InferenceBaseURL) != "" {
		health.Dependencies[2].Status = "configured"
	}

	return health, nil
}

func (m *Manager) executeRun(ctx context.Context, runID string, forcedCycleID *string, input ExecuteRunInput, actor string) (ExecuteRunResult, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return ExecuteRunResult{}, fmt.Errorf("run id is required")
	}
	normalizeExecuteRunInput(&input)

	runForPolicy, err := m.repo.Get(ctx, m.db, runID)
	if err != nil {
		return ExecuteRunResult{}, err
	}
	cycleForPolicy, _, err := resolveExecutionCycle(ctx, m.repo, m.db, runForPolicy, forcedCycleID, input.CycleID)
	if err != nil {
		return ExecuteRunResult{}, err
	}

	var profileForPolicy *ModelProfile
	if runForPolicy.ExecutionMode == string(ExecutionModeLLM) || runForPolicy.ExecutionMode == string(ExecutionModeDual) {
		profile, err := m.repo.GetModelProfile(ctx, m.db, runForPolicy.ModelProfile)
		if err != nil {
			return ExecuteRunResult{}, err
		}
		profileForPolicy = &profile
	}

	policyDecision, err := m.evaluateAndRecordPolicy(ctx, "run.execute", actor, &runForPolicy, cycleForPolicy, nil, nil, nil, profileForPolicy)
	if err != nil {
		return ExecuteRunResult{}, err
	}
	if !policyDecision.Allow {
		_ = m.markExecutionPolicyFailure(ctx, runID, cycleForPolicy)
		return ExecuteRunResult{}, policyDeniedError(policyDecision)
	}

	var result ExecuteRunResult
	err = m.tx.InTx(ctx, func(ctx context.Context, q store.DBTX) error {
		run, err := m.repo.Get(ctx, q, runID)
		if err != nil {
			return err
		}

		if run.RunType != string(RunTypeAgentRun) && run.RunType != string(RunTypeWeekRun) {
			return fmt.Errorf("run type %q does not support execution", run.RunType)
		}

		cycle, cycleID, err := resolveExecutionCycle(ctx, m.repo, q, run, forcedCycleID, input.CycleID)
		if err != nil {
			return err
		}

		now := time.Now().UTC()
		if run.StartedAt == nil {
			run.StartedAt = &now
		}
		run.Status = string(RunStatusRunning)
		if _, err := m.repo.Update(ctx, q, run); err != nil {
			return err
		}

		if cycle != nil {
			if cycle.StartedAt == nil {
				cycle.StartedAt = &now
			}
			cycle.Status = string(CycleStatusRunning)
			if _, err := m.repo.UpdateCycle(ctx, q, *cycle); err != nil {
				return err
			}
		}

		namespace := normalizeMemoryNamespace(run.MemoryNamespace, run.Repo, run.Domain, run.ID)
		if cycle != nil {
			namespace = composeMemoryNamespace(run, *cycle, input.AgentNamespace)
		} else if input.AgentNamespace != nil {
			namespace.AgentNamespace = strings.TrimSpace(*input.AgentNamespace)
		}

		memoryContext, err := m.memory.FetchScopedContext(ctx, namespace)
		if err != nil {
			return fmt.Errorf("fetch scoped memory context: %w", err)
		}

		deterministicSummary := json.RawMessage(`{}`)
		llmSummary := json.RawMessage(`{}`)
		guardrailSummary := json.RawMessage(`{}`)
		effectiveGuardrailStatus := string(GuardrailStatusDisabled)
		createdArtifacts := make([]Artifact, 0, 4)
		createdArtifactURIs := make([]string, 0, 4)
		var comparison *Comparison

		if run.ExecutionMode == string(ExecutionModeDeterministic) || run.ExecutionMode == string(ExecutionModeDual) {
			deterministicSummary = buildDeterministicExecutionSummary(run, cycle, memoryContext, input)

			deterministicArtifact, err := createExecutionArtifact(ctx, m.repo, q, run, cycleID, "deterministic_output", deterministicSummary, "application/json", "v1")
			if err != nil {
				return err
			}
			createdArtifacts = append(createdArtifacts, deterministicArtifact)
			createdArtifactURIs = append(createdArtifactURIs, deterministicArtifact.URI)
		}

		if run.ExecutionMode == string(ExecutionModeLLM) || run.ExecutionMode == string(ExecutionModeDual) {
			profile := profileForPolicy
			if profile == nil {
				loaded, loadErr := m.repo.GetModelProfile(ctx, q, run.ModelProfile)
				if loadErr != nil {
					return fmt.Errorf("get model profile %q: %w", run.ModelProfile, loadErr)
				}
				profile = &loaded
			}

			inferencePayload := buildInferencePayload(run, cycle, memoryContext, deterministicSummary, input, m.deps.EnableCompactDualPayload)
			phaseTimeouts := deriveInferencePhaseTimeouts(*profile, m.deps)
			enableGuardrails := profile.EnableGuardrails
			if m.deps.DisableLocalOllamaGuardrails && isLocalOllamaProvider(profile.Provider) {
				enableGuardrails = false
			}
			if !enableGuardrails {
				effectiveGuardrailStatus = string(GuardrailStatusDisabled)
			}

			inferenceResponse, err := m.inference.Execute(ctx, InferenceRequest{
				Provider:                profile.Provider,
				BaseURL:                 profile.BaseURL,
				Prompt:                  choosePrompt(input.Prompt, run, cycle),
				SystemPrompt:            chooseSystemPrompt(input.SystemPrompt, run),
				InputJSON:               inferencePayload,
				PrimaryModel:            profile.PrimaryModel,
				GuardrailModel:          profile.GuardrailModel,
				HelperModel:             profile.HelperModel,
				Temperature:             profile.Temperature,
				MaxTokens:               profile.MaxTokens,
				ExpectJSON:              profile.JSONMode || profile.StructuredOutput,
				EnableGuardrails:        enableGuardrails,
				EnableHelperModel:       profile.EnableHelperModel,
				HelperRequested:         helperRequested(input.InputJSON),
				PrimaryTimeoutSeconds:   phaseTimeouts.Primary,
				GuardrailTimeoutSeconds: phaseTimeouts.Guardrail,
				HelperTimeoutSeconds:    phaseTimeouts.Helper,
				ConnectionMeta:          profile.ConnectionMetadata,
			})
			if err != nil {
				return fmt.Errorf("execute llm inference: %w", err)
			}

			llmSummary = defaultRaw(inferenceResponse.PrimaryOutput, json.RawMessage(`{}`))
			guardrailSummary = defaultRaw(inferenceResponse.GuardrailOutput, json.RawMessage(`{}`))
			if string(guardrailSummary) == "{}" {
				guardrailSummary = defaultRaw(inferenceResponse.Guardrail, json.RawMessage(`{}`))
			}
			effectiveGuardrailStatus = resolveGuardrailStatus(enableGuardrails, inferenceResponse.GuardrailStatus, guardrailSummary)
			guardrailSummary = ensureGuardrailSummary(guardrailSummary, effectiveGuardrailStatus, inferenceResponse.GuardrailText, inferenceResponse.GuardrailScore)

			llmArtifact, err := createExecutionArtifact(ctx, m.repo, q, run, cycleID, "llm_output", llmSummary, "application/json", "v1")
			if err != nil {
				return err
			}
			createdArtifacts = append(createdArtifacts, llmArtifact)
			createdArtifactURIs = append(createdArtifactURIs, llmArtifact.URI)

			if len(guardrailSummary) > 0 && string(guardrailSummary) != "{}" {
				guardrailArtifact, err := createExecutionArtifact(ctx, m.repo, q, run, cycleID, "guardrail_report", guardrailSummary, "application/json", "v1")
				if err != nil {
					return err
				}
				createdArtifacts = append(createdArtifacts, guardrailArtifact)
				createdArtifactURIs = append(createdArtifactURIs, guardrailArtifact.URI)
			}
		}

		if run.ExecutionMode == string(ExecutionModeDual) {
			comparisonValue, err := m.repo.UpsertComparison(ctx, q, run.ID, UpsertComparisonInput{
				CycleID:              cycleID,
				DeterministicSummary: deterministicSummary,
				LLMSummary:           llmSummary,
				GuardrailSummary:     guardrailSummary,
				Deltas:               buildComparisonDeltas(deterministicSummary, llmSummary, guardrailSummary),
				ReviewStatus:         string(ReviewStatusReviewPending),
				ReviewerNotes:        "",
				FinalDisposition:     "pending_review",
				FinalOutput:          json.RawMessage(`{}`),
			})
			if err != nil {
				return err
			}
			comparison = &comparisonValue
		}

		snapshotRef, err := m.memory.PersistScopedNotes(ctx, namespace, buildMemoryWriteInput(input, memoryContext, run, cycle, deterministicSummary, llmSummary))
		if err != nil {
			return fmt.Errorf("persist scoped memory notes: %w", err)
		}

		for _, uri := range createdArtifactURIs {
			run.ArtifactBundleRefs = appendUnique(run.ArtifactBundleRefs, uri)
		}
		if snapshotRef != "" {
			run.MemorySnapshotRefs = appendUnique(run.MemorySnapshotRefs, snapshotRef)
		}
		run.GuardrailStatus = effectiveGuardrailStatus
		run.ReviewMetadataJSON = mergeJSONObjects(run.ReviewMetadataJSON, map[string]any{
			"last_policy_decision_id": policyDecision.ID,
			"guardrail_status":        effectiveGuardrailStatus,
		})

		finishedAt := time.Now().UTC()
		run.FinishedAt = &finishedAt
		run.CompletedAt = &finishedAt
		if run.ExecutionMode == string(ExecutionModeDeterministic) {
			run.Status = string(RunStatusCompleted)
		} else {
			if effectiveGuardrailStatus == string(GuardrailStatusTimeout) || effectiveGuardrailStatus == string(GuardrailStatusUnavailable) {
				run.Status = string(RunStatusGuardrailDeferred)
			} else {
				run.Status = string(RunStatusReviewPending)
			}
		}

		updatedRun, err := m.repo.Update(ctx, q, run)
		if err != nil {
			return err
		}

		if cycle != nil {
			cycle.FinishedAt = &finishedAt
			cycle.GuardrailStatus = effectiveGuardrailStatus
			if run.ExecutionMode == string(ExecutionModeDeterministic) {
				cycle.Status = string(CycleStatusCompleted)
			} else {
				if effectiveGuardrailStatus == string(GuardrailStatusTimeout) || effectiveGuardrailStatus == string(GuardrailStatusUnavailable) {
					cycle.Status = string(CycleStatusGuardrailDeferred)
				} else {
					cycle.Status = string(CycleStatusReviewPending)
				}
			}
			if snapshotRef != "" {
				cycle.MemorySnapshotRef = snapshotRef
				cycle.CarryForwardSummaryRef = snapshotRef
			}
			if len(createdArtifacts) > 0 && cycle.SummaryRef == "" {
				cycle.SummaryRef = createdArtifacts[0].URI
			}
			if _, err := m.repo.UpdateCycle(ctx, q, *cycle); err != nil {
				return err
			}
		}

		if err := recordAudit(ctx, m.audits, q, "run", "run.executed", actor, updatedRun.ID, map[string]any{
			"run_id":           updatedRun.ID,
			"run_type":         updatedRun.RunType,
			"execution_mode":   updatedRun.ExecutionMode,
			"guardrail_status": effectiveGuardrailStatus,
			"artifact_count":   len(createdArtifacts),
		}); err != nil {
			return err
		}
		if _, err := m.repo.AppendGovernanceAuditEvent(ctx, q, GovernanceAuditEventInput{
			ActorID:          actor,
			ActorType:        "user",
			ActionType:       "run.execute",
			TargetRunID:      &updatedRun.ID,
			TargetCycleID:    cycleID,
			PolicyDecisionID: &policyDecision.ID,
			PayloadSummary: mustJSON(map[string]any{
				"execution_mode":   updatedRun.ExecutionMode,
				"execution_ring":   updatedRun.ExecutionRing,
				"guardrail_status": effectiveGuardrailStatus,
				"artifact_count":   len(createdArtifacts),
			}, map[string]any{}),
		}); err != nil {
			return err
		}
		if effectiveGuardrailStatus == string(GuardrailStatusTimeout) || effectiveGuardrailStatus == string(GuardrailStatusUnavailable) || effectiveGuardrailStatus == string(GuardrailStatusDisabled) {
			guardrailAction := "guardrail.deferred"
			if effectiveGuardrailStatus == string(GuardrailStatusDisabled) {
				guardrailAction = "guardrail.disabled"
			}
			if _, err := m.repo.AppendGovernanceAuditEvent(ctx, q, GovernanceAuditEventInput{
				ActorID:          actor,
				ActorType:        "user",
				ActionType:       guardrailAction,
				TargetRunID:      &updatedRun.ID,
				TargetCycleID:    cycleID,
				PolicyDecisionID: &policyDecision.ID,
				PayloadSummary: mustJSON(map[string]any{
					"guardrail_status": effectiveGuardrailStatus,
				}, map[string]any{}),
			}); err != nil {
				return err
			}
		}

		if err := m.scheduler.RecordRunIntent(ctx, q, scheduler.Signal{
			RunID:   updatedRun.ID,
			RunName: updatedRun.Name,
			Status:  updatedRun.Status,
			Actor:   actor,
			Reason:  "run.executed",
		}); err != nil {
			return err
		}

		result = ExecuteRunResult{
			RunID:                updatedRun.ID,
			RunType:              updatedRun.RunType,
			ExecutionMode:        updatedRun.ExecutionMode,
			Status:               updatedRun.Status,
			CycleID:              cycleID,
			MemorySnapshotRef:    snapshotRef,
			DeterministicSummary: deterministicSummary,
			LLMSummary:           llmSummary,
			GuardrailSummary:     guardrailSummary,
			GuardrailStatus:      effectiveGuardrailStatus,
			Artifacts:            createdArtifacts,
			Comparison:           comparison,
		}

		return nil
	})
	if err != nil {
		return ExecuteRunResult{}, err
	}

	return result, nil
}

func mergeRun(existing Run, input UpdateInput) (Run, error) {
	if input.Name != nil {
		existing.Name = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		existing.Description = strings.TrimSpace(*input.Description)
	}
	if input.Status != nil {
		existing.Status = strings.TrimSpace(*input.Status)
	}
	if input.ScenarioType != nil {
		existing.ScenarioType = strings.TrimSpace(*input.ScenarioType)
	}
	if input.RunType != nil {
		existing.RunType = strings.TrimSpace(*input.RunType)
	}
	if input.ExecutionMode != nil {
		existing.ExecutionMode = strings.TrimSpace(*input.ExecutionMode)
	}
	if input.ExecutionRing != nil {
		existing.ExecutionRing = strings.TrimSpace(*input.ExecutionRing)
	}
	if input.GuardrailStatus != nil {
		existing.GuardrailStatus = strings.TrimSpace(*input.GuardrailStatus)
	}
	if input.Repo != nil {
		existing.Repo = strings.TrimSpace(*input.Repo)
	}
	if input.Domain != nil {
		existing.Domain = strings.TrimSpace(*input.Domain)
	}
	if input.DatasetRefs != nil {
		existing.DatasetRefs = dedupeAndTrimStrings(*input.DatasetRefs)
	}
	if input.PromptPackVersion != nil {
		existing.PromptPackVersion = strings.TrimSpace(*input.PromptPackVersion)
	}
	if input.RulePackVersion != nil {
		existing.RulePackVersion = strings.TrimSpace(*input.RulePackVersion)
	}
	if input.ModelProfile != nil {
		existing.ModelProfile = strings.TrimSpace(*input.ModelProfile)
	}
	if input.GuardrailProfile != nil {
		existing.GuardrailProfile = strings.TrimSpace(*input.GuardrailProfile)
	}
	if input.MemoryNamespace != nil {
		existing.MemoryNamespace = *input.MemoryNamespace
	}
	if input.RequestedBy != nil {
		existing.RequestedBy = strings.TrimSpace(*input.RequestedBy)
	}
	if input.StartedAt != nil {
		existing.StartedAt = input.StartedAt
	}
	if input.FinishedAt != nil {
		existing.FinishedAt = input.FinishedAt
		existing.CompletedAt = input.FinishedAt
	}
	if input.CompletedAt != nil {
		existing.CompletedAt = input.CompletedAt
		existing.FinishedAt = input.CompletedAt
	}
	if input.ArtifactBundleRefs != nil {
		existing.ArtifactBundleRefs = dedupeAndTrimStrings(*input.ArtifactBundleRefs)
	}
	if input.MemorySnapshotRefs != nil {
		existing.MemorySnapshotRefs = dedupeAndTrimStrings(*input.MemorySnapshotRefs)
	}
	if input.ReviewMetadataJSON != nil {
		existing.ReviewMetadataJSON = *input.ReviewMetadataJSON
	}
	if input.Notes != nil {
		existing.Notes = strings.TrimSpace(*input.Notes)
	}
	if input.MetadataJSON != nil {
		existing.MetadataJSON = *input.MetadataJSON
	}

	existing.MemoryNamespace = normalizeMemoryNamespace(existing.MemoryNamespace, existing.Repo, existing.Domain, existing.ID)

	if existing.Name == "" {
		return Run{}, fmt.Errorf("name is required")
	}
	if !isValidRunStatus(existing.Status) {
		return Run{}, fmt.Errorf("invalid status %q", existing.Status)
	}
	if !isValidRunType(existing.RunType) {
		return Run{}, fmt.Errorf("invalid run_type %q", existing.RunType)
	}
	if !isValidExecutionMode(existing.ExecutionMode) {
		return Run{}, fmt.Errorf("invalid execution_mode %q", existing.ExecutionMode)
	}
	if !isValidExecutionRing(existing.ExecutionRing) {
		return Run{}, fmt.Errorf("invalid execution_ring %q", existing.ExecutionRing)
	}
	if !isValidGuardrailStatus(existing.GuardrailStatus) {
		return Run{}, fmt.Errorf("invalid guardrail_status %q", existing.GuardrailStatus)
	}
	if existing.ModelProfile == "" {
		existing.ModelProfile = "ach-default"
	}
	if existing.ExecutionRing == "" {
		existing.ExecutionRing = defaultExecutionRingForMode(existing.ExecutionMode)
	}
	if existing.GuardrailStatus == "" {
		existing.GuardrailStatus = string(GuardrailStatusDisabled)
	}
	if existing.GuardrailProfile == "" {
		existing.GuardrailProfile = "ach-guardian-default"
	}
	if len(existing.MetadataJSON) == 0 {
		existing.MetadataJSON = json.RawMessage(`{}`)
	}
	if len(existing.ReviewMetadataJSON) == 0 {
		existing.ReviewMetadataJSON = json.RawMessage(`{}`)
	}

	return existing, nil
}

func mergeCycle(existing Cycle, input UpdateCycleInput) (Cycle, error) {
	if input.Focus != nil {
		existing.Focus = strings.TrimSpace(*input.Focus)
	}
	if input.Objective != nil {
		existing.Objective = strings.TrimSpace(*input.Objective)
	}
	if input.DetectorPack != nil {
		existing.DetectorPack = strings.TrimSpace(*input.DetectorPack)
	}
	if input.ExecutionRing != nil {
		existing.ExecutionRing = strings.TrimSpace(*input.ExecutionRing)
	}
	if input.GuardrailStatus != nil {
		existing.GuardrailStatus = strings.TrimSpace(*input.GuardrailStatus)
	}
	if input.SummaryRef != nil {
		existing.SummaryRef = strings.TrimSpace(*input.SummaryRef)
	}
	if input.CarryForwardSummaryRef != nil {
		existing.CarryForwardSummaryRef = strings.TrimSpace(*input.CarryForwardSummaryRef)
	}
	if input.Status != nil {
		next := strings.TrimSpace(*input.Status)
		if !isValidCycleStatus(next) {
			return Cycle{}, fmt.Errorf("invalid cycle status %q", next)
		}
		if !isValidCycleTransition(existing.Status, next) {
			return Cycle{}, fmt.Errorf("invalid cycle status transition %q -> %q", existing.Status, next)
		}
		existing.Status = next
	}
	if input.StartedAt != nil {
		existing.StartedAt = input.StartedAt
	}
	if input.FinishedAt != nil {
		existing.FinishedAt = input.FinishedAt
	}
	if input.MetadataJSON != nil {
		existing.MetadataJSON = *input.MetadataJSON
	}
	if !isValidExecutionRing(existing.ExecutionRing) {
		return Cycle{}, fmt.Errorf("invalid execution_ring %q", existing.ExecutionRing)
	}
	if !isValidGuardrailStatus(existing.GuardrailStatus) {
		return Cycle{}, fmt.Errorf("invalid guardrail_status %q", existing.GuardrailStatus)
	}
	if len(existing.MetadataJSON) == 0 {
		existing.MetadataJSON = json.RawMessage(`{}`)
	}
	return existing, nil
}

func normalizeCreateInput(input *CreateInput, actor string) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.ScenarioType = strings.TrimSpace(input.ScenarioType)
	input.RunType = strings.TrimSpace(input.RunType)
	input.ExecutionMode = strings.TrimSpace(input.ExecutionMode)
	input.ExecutionRing = strings.TrimSpace(input.ExecutionRing)
	input.GuardrailStatus = strings.TrimSpace(input.GuardrailStatus)
	input.Repo = strings.TrimSpace(input.Repo)
	input.Domain = strings.TrimSpace(input.Domain)
	input.PromptPackVersion = strings.TrimSpace(input.PromptPackVersion)
	input.RulePackVersion = strings.TrimSpace(input.RulePackVersion)
	input.ModelProfile = strings.TrimSpace(input.ModelProfile)
	input.GuardrailProfile = strings.TrimSpace(input.GuardrailProfile)
	input.RequestedBy = strings.TrimSpace(input.RequestedBy)
	input.CreatedBy = strings.TrimSpace(input.CreatedBy)
	input.Notes = strings.TrimSpace(input.Notes)
	input.DatasetRefs = dedupeAndTrimStrings(input.DatasetRefs)
	input.ArtifactBundleRefs = dedupeAndTrimStrings(input.ArtifactBundleRefs)
	input.MemorySnapshotRefs = dedupeAndTrimStrings(input.MemorySnapshotRefs)

	if input.Status == "" {
		input.Status = string(RunStatusPending)
	}
	if input.RunType == "" {
		input.RunType = string(RunTypeReplayRun)
	}
	if input.ExecutionMode == "" {
		input.ExecutionMode = string(ExecutionModeDeterministic)
	}
	if input.ExecutionRing == "" {
		input.ExecutionRing = defaultExecutionRingForMode(input.ExecutionMode)
	}
	if input.GuardrailStatus == "" {
		input.GuardrailStatus = string(GuardrailStatusDisabled)
	}
	if input.ModelProfile == "" {
		input.ModelProfile = "ach-default"
	}
	if input.GuardrailProfile == "" {
		input.GuardrailProfile = "ach-guardian-default"
	}
	if input.RequestedBy == "" {
		input.RequestedBy = actor
	}
	if input.CreatedBy == "" {
		input.CreatedBy = actor
	}
	if len(input.MetadataJSON) == 0 {
		input.MetadataJSON = json.RawMessage(`{}`)
	}
	if len(input.ReviewMetadataJSON) == 0 {
		input.ReviewMetadataJSON = json.RawMessage(`{}`)
	}
	input.MemoryNamespace = normalizeMemoryNamespace(input.MemoryNamespace, input.Repo, input.Domain, input.Name)
}

func normalizeExecuteRunInput(input *ExecuteRunInput) {
	input.Prompt = strings.TrimSpace(input.Prompt)
	input.SystemPrompt = strings.TrimSpace(input.SystemPrompt)
	input.MemoryNote = strings.TrimSpace(input.MemoryNote)
	if input.CycleID != nil {
		cycleID := strings.TrimSpace(*input.CycleID)
		input.CycleID = &cycleID
	}
	if input.AgentNamespace != nil {
		agentNamespace := strings.TrimSpace(*input.AgentNamespace)
		input.AgentNamespace = &agentNamespace
	}
	if len(input.InputJSON) == 0 {
		input.InputJSON = json.RawMessage(`{}`)
	}
}

func validateCreateInput(input CreateInput) error {
	if input.Name == "" {
		return fmt.Errorf("name is required")
	}
	if !isValidRunStatus(input.Status) {
		return fmt.Errorf("invalid status %q", input.Status)
	}
	if !isValidRunType(input.RunType) {
		return fmt.Errorf("invalid run_type %q", input.RunType)
	}
	if !isValidExecutionMode(input.ExecutionMode) {
		return fmt.Errorf("invalid execution_mode %q", input.ExecutionMode)
	}
	if !isValidExecutionRing(input.ExecutionRing) {
		return fmt.Errorf("invalid execution_ring %q", input.ExecutionRing)
	}
	if !isValidGuardrailStatus(input.GuardrailStatus) {
		return fmt.Errorf("invalid guardrail_status %q", input.GuardrailStatus)
	}
	if input.ModelProfile == "" {
		return fmt.Errorf("model_profile is required")
	}
	if input.RequestedBy == "" {
		return fmt.Errorf("requested_by is required")
	}
	return nil
}

func normalizeCreateCycleInput(input *CreateCycleInput) error {
	input.CycleKey = normalizeCycleKey(input.CycleKey)
	input.Focus = strings.TrimSpace(input.Focus)
	input.Objective = strings.TrimSpace(input.Objective)
	input.DetectorPack = strings.TrimSpace(input.DetectorPack)
	input.ExecutionRing = strings.TrimSpace(input.ExecutionRing)
	input.SummaryRef = strings.TrimSpace(input.SummaryRef)
	input.CarryForwardSummaryRef = strings.TrimSpace(input.CarryForwardSummaryRef)
	if input.Status == "" {
		input.Status = string(CycleStatusPending)
	}
	if input.ExecutionRing == "" {
		input.ExecutionRing = string(ExecutionRing1)
	}
	input.Status = strings.TrimSpace(input.Status)
	if len(input.MetadataJSON) == 0 {
		input.MetadataJSON = json.RawMessage(`{}`)
	}
	if input.CycleKey == "" {
		return fmt.Errorf("cycle_key is required")
	}
	if !isValidCycleKey(input.CycleKey) {
		return fmt.Errorf("invalid cycle_key %q", input.CycleKey)
	}
	if !isValidCycleStatus(input.Status) {
		return fmt.Errorf("invalid cycle status %q", input.Status)
	}
	if !isValidExecutionRing(input.ExecutionRing) {
		return fmt.Errorf("invalid execution_ring %q", input.ExecutionRing)
	}
	return nil
}

func normalizeAttachArtifactInput(input *AttachArtifactInput) error {
	input.ArtifactType = strings.TrimSpace(input.ArtifactType)
	input.URI = strings.TrimSpace(input.URI)
	input.ContentType = strings.TrimSpace(input.ContentType)
	input.Version = strings.TrimSpace(input.Version)
	input.Checksum = strings.TrimSpace(input.Checksum)
	if input.CycleID != nil {
		cycleID := strings.TrimSpace(*input.CycleID)
		input.CycleID = &cycleID
	}
	if input.ContentType == "" {
		input.ContentType = "application/json"
	}
	if len(input.MetadataJSON) == 0 {
		input.MetadataJSON = json.RawMessage(`{}`)
	}
	if input.ArtifactType == "" {
		return fmt.Errorf("artifact_type is required")
	}
	if input.URI == "" {
		return fmt.Errorf("uri is required")
	}
	return nil
}

func normalizeUpsertComparisonInput(input *UpsertComparisonInput) error {
	if input.CycleID != nil {
		cycleID := strings.TrimSpace(*input.CycleID)
		input.CycleID = &cycleID
	}
	if input.ReviewStatus == "" {
		input.ReviewStatus = string(ReviewStatusReviewPending)
	}
	input.ReviewStatus = strings.TrimSpace(input.ReviewStatus)
	input.ReviewerNotes = strings.TrimSpace(input.ReviewerNotes)
	input.FinalDisposition = strings.TrimSpace(input.FinalDisposition)
	if !isValidComparisonReviewStatus(input.ReviewStatus) {
		return fmt.Errorf("invalid review_status %q", input.ReviewStatus)
	}
	if len(input.DeterministicSummary) == 0 {
		input.DeterministicSummary = json.RawMessage(`{}`)
	}
	if len(input.LLMSummary) == 0 {
		input.LLMSummary = json.RawMessage(`{}`)
	}
	if len(input.GuardrailSummary) == 0 {
		input.GuardrailSummary = json.RawMessage(`{}`)
	}
	if len(input.Deltas) == 0 {
		input.Deltas = json.RawMessage(`{}`)
	}
	if len(input.FinalOutput) == 0 {
		input.FinalOutput = json.RawMessage(`{}`)
	}
	return nil
}

func normalizeModelProfileInput(input *RegisterModelProfileInput) {
	input.Name = strings.TrimSpace(input.Name)
	input.Provider = strings.TrimSpace(input.Provider)
	input.BaseURL = strings.TrimSpace(input.BaseURL)
	input.PrimaryModel = strings.TrimSpace(input.PrimaryModel)
	input.GuardrailModel = strings.TrimSpace(input.GuardrailModel)
	input.HelperModel = strings.TrimSpace(input.HelperModel)

	if input.TimeoutSeconds <= 0 {
		input.TimeoutSeconds = 45
	}
	if input.MaxTokens <= 0 {
		input.MaxTokens = 4096
	}
	if len(input.ConnectionMetadata) == 0 {
		input.ConnectionMetadata = json.RawMessage(`{}`)
	}
}

func validateModelProfileInput(input RegisterModelProfileInput) error {
	if input.Name == "" {
		return fmt.Errorf("name is required")
	}
	if input.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if input.PrimaryModel == "" {
		return fmt.Errorf("primary_model is required")
	}
	if input.TimeoutSeconds <= 0 {
		return fmt.Errorf("timeout_seconds must be greater than 0")
	}
	if input.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be greater than 0")
	}
	if input.Temperature < 0 {
		return fmt.Errorf("temperature must be >= 0")
	}
	return nil
}

func composeMemoryNamespace(run Run, cycle Cycle, agentNamespace *string) MemoryNamespace {
	ns := normalizeMemoryNamespace(run.MemoryNamespace, run.Repo, run.Domain, run.ID)
	ns.CycleNamespace = cycle.CycleKey
	if agentNamespace != nil {
		ns.AgentNamespace = strings.TrimSpace(*agentNamespace)
	}
	return ns
}

func normalizeMemoryNamespace(ns MemoryNamespace, repo string, domain string, runKey string) MemoryNamespace {
	ns.RepoNamespace = strings.TrimSpace(ns.RepoNamespace)
	ns.RunNamespace = strings.TrimSpace(ns.RunNamespace)
	ns.CycleNamespace = strings.TrimSpace(ns.CycleNamespace)
	ns.AgentNamespace = strings.TrimSpace(ns.AgentNamespace)

	if ns.RepoNamespace == "" {
		ns.RepoNamespace = strings.TrimSpace(repo)
	}
	if ns.RepoNamespace == "" {
		ns.RepoNamespace = strings.TrimSpace(domain)
	}
	if ns.RepoNamespace == "" {
		ns.RepoNamespace = "global"
	}

	if ns.RunNamespace == "" {
		runKey = strings.TrimSpace(runKey)
		if runKey == "" {
			runKey = "unspecified"
		}
		ns.RunNamespace = "run-" + runKey
	}

	return ns
}

func isValidRunStatus(status string) bool {
	switch status {
	case string(RunStatusPending),
		string(RunStatusScheduled),
		string(RunStatusRunning),
		string(RunStatusGuardrailDeferred),
		string(RunStatusFailedRuntime),
		string(RunStatusFailedPolicy),
		string(RunStatusReviewPending),
		string(RunStatusApproved),
		string(RunStatusRejected),
		string(RunStatusOverridden),
		string(RunStatusDeferred),
		string(RunStatusCompleted),
		string(RunStatusFailed),
		string(RunStatusCancelled),
		string(RunStatusClosed):
		return true
	default:
		return false
	}
}

func isValidRunType(runType string) bool {
	switch runType {
	case string(RunTypeReplayRun), string(RunTypeAgentRun), string(RunTypeWeekRun):
		return true
	default:
		return false
	}
}

func isValidExecutionMode(mode string) bool {
	switch mode {
	case string(ExecutionModeDeterministic), string(ExecutionModeLLM), string(ExecutionModeDual):
		return true
	default:
		return false
	}
}

func isValidCycleStatus(status string) bool {
	switch status {
	case string(CycleStatusPending),
		string(CycleStatusRunning),
		string(CycleStatusGuardrailDeferred),
		string(CycleStatusFailedRuntime),
		string(CycleStatusFailedPolicy),
		string(CycleStatusReviewPending),
		string(CycleStatusApproved),
		string(CycleStatusRejected),
		string(CycleStatusOverridden),
		string(CycleStatusDeferred),
		string(CycleStatusCompleted),
		string(CycleStatusFailed),
		string(CycleStatusCancelled):
		return true
	default:
		return false
	}
}

func isValidExecutionRing(value string) bool {
	switch value {
	case string(ExecutionRing0), string(ExecutionRing1), string(ExecutionRing2), string(ExecutionRing3):
		return true
	default:
		return false
	}
}

func isValidGuardrailStatus(value string) bool {
	switch value {
	case string(GuardrailStatusPassed),
		string(GuardrailStatusFlagged),
		string(GuardrailStatusTimeout),
		string(GuardrailStatusUnavailable),
		string(GuardrailStatusDisabled):
		return true
	default:
		return false
	}
}

func isValidComparisonReviewStatus(status string) bool {
	switch status {
	case string(ReviewStatusReviewPending), string(ReviewStatusApproved), string(ReviewStatusRejected), string(ReviewStatusOverridden):
		return true
	default:
		return false
	}
}

func isValidCycleTransition(current string, next string) bool {
	if current == next {
		return true
	}
	allowed := map[string]map[string]bool{
		string(CycleStatusPending): {
			string(CycleStatusRunning):   true,
			string(CycleStatusFailed):    true,
			string(CycleStatusCancelled): true,
		},
		string(CycleStatusRunning): {
			string(CycleStatusReviewPending): true,
			string(CycleStatusCompleted):     true,
			string(CycleStatusFailed):        true,
			string(CycleStatusCancelled):     true,
		},
		string(CycleStatusReviewPending): {
			string(CycleStatusApproved):   true,
			string(CycleStatusRejected):   true,
			string(CycleStatusRunning):    true,
			string(CycleStatusDeferred):   true,
			string(CycleStatusOverridden): true,
			string(CycleStatusCompleted):  true,
		},
		string(CycleStatusGuardrailDeferred): {
			string(CycleStatusReviewPending): true,
			string(CycleStatusDeferred):      true,
			string(CycleStatusCancelled):     true,
			string(CycleStatusFailedRuntime): true,
		},
		string(CycleStatusApproved): {
			string(CycleStatusCompleted): true,
		},
		string(CycleStatusRejected): {
			string(CycleStatusRunning):   true,
			string(CycleStatusCancelled): true,
		},
		string(CycleStatusDeferred): {
			string(CycleStatusRunning): true,
		},
	}
	return allowed[current][next]
}

func normalizeCycleKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "-")
	return value
}

func helperRequested(input json.RawMessage) bool {
	if len(input) == 0 || string(input) == "{}" {
		return false
	}
	var payload map[string]any
	if err := json.Unmarshal(input, &payload); err != nil {
		return false
	}
	keys := []string{"helper_requested", "enable_helper", "needs_helper"}
	for _, key := range keys {
		value, ok := payload[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case bool:
			return typed
		case string:
			normalized := strings.ToLower(strings.TrimSpace(typed))
			return normalized == "true" || normalized == "1" || normalized == "yes"
		case float64:
			return typed > 0
		}
	}
	return false
}

func isValidCycleKey(value string) bool {
	if strings.HasPrefix(value, "day-") {
		day, err := strconv.Atoi(strings.TrimPrefix(value, "day-"))
		if err != nil {
			return false
		}
		return day >= 1 && day <= 7
	}
	return value != ""
}

func dedupeAndTrimStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		result = append(result, trimmed)
	}
	return result
}

func appendUnique(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func recordAudit(ctx context.Context, recorder audit.Recorder, q store.DBTX, entityType string, eventType string, actor string, entityID string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal audit payload: %w", err)
	}

	return recorder.Record(ctx, q, audit.Event{
		EventType:  eventType,
		EntityType: entityType,
		EntityID:   entityID,
		Actor:      actor,
		Payload:    body,
	})
}

func resolveExecutionCycle(
	ctx context.Context,
	repo Repository,
	q store.DBTX,
	run Run,
	forcedCycleID *string,
	requestedCycleID *string,
) (*Cycle, *string, error) {
	if run.RunType == string(RunTypeWeekRun) {
		cycleID := ""
		if forcedCycleID != nil {
			cycleID = strings.TrimSpace(*forcedCycleID)
		}
		if cycleID == "" && requestedCycleID != nil {
			cycleID = strings.TrimSpace(*requestedCycleID)
		}
		if cycleID == "" {
			return nil, nil, fmt.Errorf("cycle id is required for week_run execution")
		}
		cycle, err := repo.GetCycle(ctx, q, run.ID, cycleID)
		if err != nil {
			return nil, nil, err
		}
		return &cycle, &cycleID, nil
	}

	if forcedCycleID != nil || (requestedCycleID != nil && strings.TrimSpace(*requestedCycleID) != "") {
		return nil, nil, fmt.Errorf("cycle id is only supported for week_run execution")
	}
	return nil, nil, nil
}

func buildDeterministicExecutionSummary(run Run, cycle *Cycle, memoryContext MemoryContext, input ExecuteRunInput) json.RawMessage {
	payload := map[string]any{
		"run_id":                       run.ID,
		"run_type":                     run.RunType,
		"execution_mode":               run.ExecutionMode,
		"execution_ring":               run.ExecutionRing,
		"dataset_ref_count":            len(run.DatasetRefs),
		"carry_forward_risk_count":     len(memoryContext.CarryForwardRisks),
		"unresolved_gap_count":         len(memoryContext.UnresolvedGaps),
		"reviewer_note_count":          len(memoryContext.ReviewerNotes),
		"deterministic_signal_version": "v1",
	}
	if cycle != nil {
		payload["cycle_id"] = cycle.ID
		payload["cycle_key"] = cycle.CycleKey
	}
	if len(input.InputJSON) != 0 && string(input.InputJSON) != "{}" {
		payload["input"] = json.RawMessage(input.InputJSON)
	}
	body, _ := json.Marshal(payload)
	return body
}

func buildInferencePayload(
	run Run,
	cycle *Cycle,
	memoryContext MemoryContext,
	deterministicSummary json.RawMessage,
	input ExecuteRunInput,
	compactDualEnabled bool,
) json.RawMessage {
	if compactDualEnabled && run.ExecutionMode == string(ExecutionModeDual) {
		return buildCompactDualInferencePayload(run, cycle, memoryContext, deterministicSummary)
	}

	payload := map[string]any{
		"run_id":                run.ID,
		"run_type":              run.RunType,
		"execution_mode":        run.ExecutionMode,
		"dataset_refs":          run.DatasetRefs,
		"prompt_pack_version":   run.PromptPackVersion,
		"rule_pack_version":     run.RulePackVersion,
		"memory_context":        memoryContext,
		"deterministic_summary": deterministicSummary,
		"input":                 json.RawMessage(input.InputJSON),
	}
	if cycle != nil {
		payload["cycle"] = map[string]any{
			"id":            cycle.ID,
			"cycle_key":     cycle.CycleKey,
			"focus":         cycle.Focus,
			"objective":     cycle.Objective,
			"detector_pack": cycle.DetectorPack,
		}
	}
	body, _ := json.Marshal(payload)
	return body
}

func buildCompactDualInferencePayload(
	run Run,
	cycle *Cycle,
	memoryContext MemoryContext,
	deterministicSummary json.RawMessage,
) json.RawMessage {
	payload := map[string]any{
		"run_id":                        run.ID,
		"run_type":                      run.RunType,
		"execution_mode":                run.ExecutionMode,
		"cycle_key":                     "",
		"focus":                         "",
		"objective":                     "",
		"detector_pack":                 "",
		"dataset_refs":                  run.DatasetRefs,
		"prompt_pack_version":           run.PromptPackVersion,
		"rule_pack_version":             run.RulePackVersion,
		"deterministic_summary_compact": compactDeterministicSummary(deterministicSummary),
		"carry_forward_memory_compact":  compactCarryForwardMemorySummary(memoryContext),
	}
	if cycle != nil {
		payload["cycle_key"] = cycle.CycleKey
		payload["focus"] = cycle.Focus
		payload["objective"] = cycle.Objective
		payload["detector_pack"] = cycle.DetectorPack
	}

	body, _ := json.Marshal(payload)
	return body
}

func compactDeterministicSummary(summary json.RawMessage) map[string]any {
	out := map[string]any{}
	if len(summary) == 0 || string(summary) == "{}" {
		return out
	}

	var decoded map[string]any
	if err := json.Unmarshal(summary, &decoded); err != nil {
		out["summary_present"] = true
		return out
	}

	keys := []string{
		"run_id",
		"cycle_key",
		"dataset_ref_count",
		"carry_forward_risk_count",
		"unresolved_gap_count",
		"reviewer_note_count",
		"deterministic_signal_version",
	}
	for _, key := range keys {
		if value, ok := decoded[key]; ok {
			out[key] = value
		}
	}

	return out
}

func compactCarryForwardMemorySummary(memoryContext MemoryContext) map[string]any {
	return map[string]any{
		"prior_cycle_summaries": limitStrings(memoryContext.PriorCycleSummaries, 3, 220),
		"carry_forward_risks":   limitStrings(memoryContext.CarryForwardRisks, 5, 220),
		"unresolved_gaps":       limitStrings(memoryContext.UnresolvedGaps, 5, 220),
		"reviewer_notes":        limitStrings(memoryContext.ReviewerNotes, 5, 220),
		"counts": map[string]int{
			"prior_cycle_summaries": len(memoryContext.PriorCycleSummaries),
			"carry_forward_risks":   len(memoryContext.CarryForwardRisks),
			"unresolved_gaps":       len(memoryContext.UnresolvedGaps),
			"reviewer_notes":        len(memoryContext.ReviewerNotes),
		},
	}
}

func buildMemoryWriteInput(
	input ExecuteRunInput,
	memoryContext MemoryContext,
	run Run,
	cycle *Cycle,
	deterministicSummary json.RawMessage,
	llmSummary json.RawMessage,
) MemoryWriteInput {
	note := input.MemoryNote
	if note == "" {
		if cycle != nil {
			note = fmt.Sprintf("cycle %s executed in %s mode", cycle.CycleKey, run.ExecutionMode)
		} else {
			note = fmt.Sprintf("run %s executed in %s mode", run.ID, run.ExecutionMode)
		}
	}

	priorSummaries := append([]string{}, memoryContext.PriorCycleSummaries...)
	if string(deterministicSummary) != "{}" {
		priorSummaries = append(priorSummaries, "deterministic_summary_generated")
	}
	if string(llmSummary) != "{}" {
		priorSummaries = append(priorSummaries, "llm_summary_generated")
	}

	return MemoryWriteInput{
		Note:                note,
		PriorCycleSummaries: dedupeAndTrimStrings(priorSummaries),
		CarryForwardRisks:   dedupeAndTrimStrings(memoryContext.CarryForwardRisks),
		UnresolvedGaps:      dedupeAndTrimStrings(memoryContext.UnresolvedGaps),
		BacklogItems:        dedupeAndTrimStrings(memoryContext.BacklogItems),
		ReviewerNotes:       dedupeAndTrimStrings(memoryContext.ReviewerNotes),
	}
}

func createExecutionArtifact(
	ctx context.Context,
	repo Repository,
	q store.DBTX,
	run Run,
	cycleID *string,
	artifactType string,
	payload json.RawMessage,
	contentType string,
	version string,
) (Artifact, error) {
	now := time.Now().UTC().Format("20060102T150405Z")
	uri := fmt.Sprintf("control-plane://runs/%s/%s-%s.json", run.ID, artifactType, now)
	return repo.CreateArtifact(ctx, q, run.ID, AttachArtifactInput{
		CycleID:      cycleID,
		ArtifactType: artifactType,
		URI:          uri,
		ContentType:  contentType,
		Version:      version,
		MetadataJSON: payload,
	})
}

func buildComparisonDeltas(deterministicSummary json.RawMessage, llmSummary json.RawMessage, guardrailSummary json.RawMessage) json.RawMessage {
	guardrailStatus := guardrailStatusFromSummary(guardrailSummary)
	guardrailPresent := guardrailStatus == string(GuardrailStatusPassed) || guardrailStatus == string(GuardrailStatusFlagged)
	payload := map[string]any{
		"deterministic_present": string(deterministicSummary) != "{}",
		"llm_present":           string(llmSummary) != "{}",
		"guardrail_present":     guardrailPresent,
		"guardrail_status":      guardrailStatus,
	}
	body, _ := json.Marshal(payload)
	return body
}

func ensureGuardrailSummary(summary json.RawMessage, status string, text string, score *float64) json.RawMessage {
	if status == "" {
		status = string(GuardrailStatusDisabled)
	}
	payload := map[string]any{}
	if len(summary) != 0 && string(summary) != "{}" {
		_ = json.Unmarshal(summary, &payload)
	}
	payload["status"] = status
	if strings.TrimSpace(text) != "" {
		payload["note"] = strings.TrimSpace(text)
	}
	if score != nil {
		payload["score"] = *score
	}
	body, _ := json.Marshal(payload)
	return body
}

func guardrailStatusFromSummary(summary json.RawMessage) string {
	if len(summary) == 0 || string(summary) == "{}" {
		return string(GuardrailStatusDisabled)
	}
	var payload map[string]any
	if err := json.Unmarshal(summary, &payload); err != nil {
		return string(GuardrailStatusFlagged)
	}
	value, _ := payload["status"].(string)
	value = strings.TrimSpace(strings.ToLower(value))
	switch value {
	case string(GuardrailStatusPassed),
		string(GuardrailStatusFlagged),
		string(GuardrailStatusTimeout),
		string(GuardrailStatusUnavailable),
		string(GuardrailStatusDisabled):
		return value
	default:
		return string(GuardrailStatusFlagged)
	}
}

func resolveGuardrailStatus(enabled bool, returned string, summary json.RawMessage) string {
	if !enabled {
		return string(GuardrailStatusDisabled)
	}
	returned = strings.TrimSpace(strings.ToLower(returned))
	switch returned {
	case string(GuardrailStatusPassed),
		string(GuardrailStatusFlagged),
		string(GuardrailStatusTimeout),
		string(GuardrailStatusUnavailable):
		return returned
	}
	if len(summary) == 0 || string(summary) == "{}" {
		return string(GuardrailStatusUnavailable)
	}
	return guardrailStatusFromSummary(summary)
}

func choosePrompt(candidate string, run Run, cycle *Cycle) string {
	if candidate != "" {
		return candidate
	}
	if cycle != nil {
		return fmt.Sprintf("Analyze cycle %s for run %s using strict JSON output.", cycle.CycleKey, run.ID)
	}
	return fmt.Sprintf("Analyze run %s using strict JSON output.", run.ID)
}

func chooseSystemPrompt(candidate string, run Run) string {
	if candidate != "" {
		return candidate
	}
	return fmt.Sprintf("You are a compliance-oriented assistant for %s. Return JSON only.", run.Domain)
}

func defaultExecutionRingForMode(mode string) string {
	switch mode {
	case string(ExecutionModeLLM), string(ExecutionModeDual):
		return string(ExecutionRing2)
	default:
		return string(ExecutionRing1)
	}
}

func normalizeReviewActionInput(input *ReviewActionInput, actor string) {
	input.Action = strings.TrimSpace(strings.ToLower(input.Action))
	input.ReviewerID = strings.TrimSpace(input.ReviewerID)
	input.ReviewerType = strings.TrimSpace(strings.ToLower(input.ReviewerType))
	input.Rationale = strings.TrimSpace(input.Rationale)
	if input.ReviewerID == "" {
		input.ReviewerID = strings.TrimSpace(actor)
	}
	if input.ReviewerType == "" {
		input.ReviewerType = "human"
	}
	if input.CycleID != nil {
		cycleID := strings.TrimSpace(*input.CycleID)
		input.CycleID = &cycleID
	}
	if input.PolicyDecisionID != nil {
		policyID := strings.TrimSpace(*input.PolicyDecisionID)
		input.PolicyDecisionID = &policyID
	}
}

func validateReviewActionInput(input ReviewActionInput) error {
	if input.ReviewerID == "" {
		return fmt.Errorf("reviewer_id is required")
	}
	switch input.Action {
	case "approve", "reject", "override", "defer":
		return nil
	default:
		return fmt.Errorf("unsupported review action %q", input.Action)
	}
}

func mapReviewActionToRunStatus(action string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(action)) {
	case "approve":
		return string(RunStatusApproved), nil
	case "reject":
		return string(RunStatusRejected), nil
	case "override":
		return string(RunStatusOverridden), nil
	case "defer":
		return string(RunStatusDeferred), nil
	default:
		return "", fmt.Errorf("unsupported review action %q", action)
	}
}

func mapReviewActionToCycleStatus(action string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(action)) {
	case "approve":
		return string(CycleStatusApproved), nil
	case "reject":
		return string(CycleStatusRejected), nil
	case "override":
		return string(CycleStatusOverridden), nil
	case "defer":
		return string(CycleStatusDeferred), nil
	default:
		return "", fmt.Errorf("unsupported review action %q", action)
	}
}

func (m *Manager) evaluateAndRecordPolicy(
	ctx context.Context,
	action string,
	actor string,
	run *Run,
	cycle *Cycle,
	createRunInput *CreateInput,
	createCycleInput *CreateCycleInput,
	attachArtifactInput *AttachArtifactInput,
	profile *ModelProfile,
) (PolicyDecision, error) {
	input := buildPolicyInput(action, actor, m.deps.Environment, run, cycle, createRunInput, createCycleInput, attachArtifactInput, profile)
	outcome := evaluatePolicyOutcome(input)
	decision, err := m.repo.RecordPolicyDecision(ctx, m.db, PolicyDecisionInput{
		ActionType:          action,
		TargetRunID:         optionalRunID(run),
		TargetCycleID:       optionalCycleID(cycle),
		ActorID:             actor,
		ActorType:           "user",
		PolicyInput:         mustJSON(input, map[string]any{}),
		Allow:               outcome.Allow,
		PolicyBundleID:      m.deps.PolicyBundleID,
		PolicyBundleVersion: m.deps.PolicyBundleVersion,
		ReasonCode:          outcome.ReasonCode,
		ConditionsApplied:   outcome.ConditionsApplied,
		FallbackMode:        outcome.FallbackMode,
	})
	if err != nil {
		return PolicyDecision{}, err
	}

	if _, err := m.repo.AppendGovernanceAuditEvent(ctx, m.db, GovernanceAuditEventInput{
		ActorID:          actor,
		ActorType:        "user",
		ActionType:       "policy." + ternary(outcome.Allow, "allow", "deny"),
		TargetRunID:      decision.TargetRunID,
		TargetCycleID:    decision.TargetCycleID,
		PolicyDecisionID: &decision.ID,
		PayloadSummary: mustJSON(map[string]any{
			"action":             action,
			"allow":              outcome.Allow,
			"reason_code":        outcome.ReasonCode,
			"conditions_applied": outcome.ConditionsApplied,
			"fallback_mode":      outcome.FallbackMode,
		}, map[string]any{}),
	}); err != nil {
		return PolicyDecision{}, err
	}

	return decision, nil
}

func buildPolicyInput(
	action string,
	actor string,
	environment string,
	run *Run,
	cycle *Cycle,
	createRunInput *CreateInput,
	createCycleInput *CreateCycleInput,
	attachArtifactInput *AttachArtifactInput,
	profile *ModelProfile,
) map[string]any {
	payload := map[string]any{
		"action":      action,
		"actor":       actor,
		"environment": environment,
	}

	if run != nil {
		payload["run_type"] = run.RunType
		payload["execution_mode"] = run.ExecutionMode
		payload["execution_ring"] = run.ExecutionRing
		payload["repo"] = run.Repo
		payload["run_namespace"] = run.MemoryNamespace.RunNamespace
		payload["model_profile"] = run.ModelProfile
		payload["prompt_pack_version"] = run.PromptPackVersion
		payload["rule_pack_version"] = run.RulePackVersion
	}
	if cycle != nil {
		payload["cycle_id"] = cycle.ID
		payload["cycle_key"] = cycle.CycleKey
		payload["cycle_execution_ring"] = cycle.ExecutionRing
	}
	if createRunInput != nil {
		payload["run_type"] = createRunInput.RunType
		payload["execution_mode"] = createRunInput.ExecutionMode
		payload["execution_ring"] = createRunInput.ExecutionRing
		payload["repo"] = createRunInput.Repo
		payload["run_namespace"] = createRunInput.MemoryNamespace.RunNamespace
		payload["model_profile"] = createRunInput.ModelProfile
		payload["prompt_pack_version"] = createRunInput.PromptPackVersion
		payload["rule_pack_version"] = createRunInput.RulePackVersion
	}
	if createCycleInput != nil {
		payload["cycle_key"] = createCycleInput.CycleKey
		payload["cycle_execution_ring"] = createCycleInput.ExecutionRing
	}
	if attachArtifactInput != nil {
		payload["artifact_uri"] = attachArtifactInput.URI
	}
	if profile != nil {
		payload["guardrails_enabled"] = profile.EnableGuardrails
		payload["outbound_targets"] = []string{profile.BaseURL}
		payload["provider"] = profile.Provider
	}

	return payload
}

type policyOutcome struct {
	Allow             bool
	ReasonCode        string
	ConditionsApplied []string
	FallbackMode      string
}

func evaluatePolicyOutcome(input map[string]any) policyOutcome {
	ring := executionRingFromAny(input["execution_ring"])
	mode := strings.TrimSpace(fmt.Sprintf("%v", input["execution_mode"]))
	action := strings.TrimSpace(fmt.Sprintf("%v", input["action"]))
	out := policyOutcome{
		Allow:             true,
		ReasonCode:        "policy_allow",
		ConditionsApplied: []string{},
		FallbackMode:      "",
	}

	if !isValidExecutionRing(ring) {
		return policyOutcome{
			Allow:             false,
			ReasonCode:        "invalid_execution_ring",
			ConditionsApplied: []string{"valid_execution_ring_required"},
			FallbackMode:      "deny",
		}
	}

	required := minimumRingForMode(mode)
	if ringRank(ring) < ringRank(required) {
		return policyOutcome{
			Allow:             false,
			ReasonCode:        "execution_ring_too_low",
			ConditionsApplied: []string{fmt.Sprintf("minimum_ring_%s_required", required)},
			FallbackMode:      "deny",
		}
	}
	out.ConditionsApplied = append(out.ConditionsApplied, fmt.Sprintf("minimum_ring_%s_satisfied", required))

	if targets, ok := input["outbound_targets"].([]string); ok {
		for _, target := range targets {
			if isExternalTarget(target) && ringRank(ring) < ringRank(string(ExecutionRing3)) {
				return policyOutcome{
					Allow:             false,
					ReasonCode:        "external_target_requires_ring_3",
					ConditionsApplied: []string{"ring_3_required_for_external_actions"},
					FallbackMode:      "deny",
				}
			}
		}
	}

	if uri := strings.TrimSpace(fmt.Sprintf("%v", input["artifact_uri"])); uri != "" {
		if (strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://")) && ringRank(ring) < ringRank(string(ExecutionRing3)) {
			return policyOutcome{
				Allow:             false,
				ReasonCode:        "artifact_external_uri_requires_ring_3",
				ConditionsApplied: []string{"ring_3_required_for_external_actions"},
				FallbackMode:      "deny",
			}
		}
	}

	if strings.HasPrefix(action, "run.review.") {
		out.ConditionsApplied = append(out.ConditionsApplied, "review_action_allowed")
	}

	return out
}

func minimumRingForMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case string(ExecutionModeLLM), string(ExecutionModeDual):
		return string(ExecutionRing2)
	default:
		return string(ExecutionRing1)
	}
}

func ringRank(ring string) int {
	switch strings.ToLower(strings.TrimSpace(ring)) {
	case string(ExecutionRing0):
		return 0
	case string(ExecutionRing1):
		return 1
	case string(ExecutionRing2):
		return 2
	case string(ExecutionRing3):
		return 3
	default:
		return -1
	}
}

func executionRingFromAny(value any) string {
	ring := strings.TrimSpace(fmt.Sprintf("%v", value))
	if ring == "" || ring == "<nil>" {
		return string(ExecutionRing1)
	}
	return ring
}

func isExternalTarget(target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" {
		return false
	}
	if strings.Contains(target, "127.0.0.1") || strings.Contains(target, "localhost") || strings.Contains(target, "ai-precision") {
		return false
	}
	return strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://")
}

func optionalRunID(run *Run) *string {
	if run == nil || strings.TrimSpace(run.ID) == "" {
		return nil
	}
	value := run.ID
	return &value
}

func optionalCycleID(cycle *Cycle) *string {
	if cycle == nil || strings.TrimSpace(cycle.ID) == "" {
		return nil
	}
	value := cycle.ID
	return &value
}

func policyDeniedError(decision PolicyDecision) error {
	return fmt.Errorf("policy denied (%s) by %s@%s", decision.ReasonCode, decision.PolicyBundleID, decision.PolicyBundleVersion)
}

func (m *Manager) markExecutionPolicyFailure(ctx context.Context, runID string, cycle *Cycle) error {
	return m.tx.InTx(ctx, func(ctx context.Context, q store.DBTX) error {
		run, err := m.repo.Get(ctx, q, runID)
		if err != nil {
			return err
		}
		run.Status = string(RunStatusFailedPolicy)
		if _, err := m.repo.Update(ctx, q, run); err != nil {
			return err
		}
		if cycle != nil {
			loaded, err := m.repo.GetCycle(ctx, q, runID, cycle.ID)
			if err != nil {
				return err
			}
			loaded.Status = string(CycleStatusFailedPolicy)
			if _, err := m.repo.UpdateCycle(ctx, q, loaded); err != nil {
				return err
			}
		}
		return nil
	})
}

func mergeJSONObjects(raw json.RawMessage, patch map[string]any) json.RawMessage {
	payload := map[string]any{}
	if len(raw) != 0 && string(raw) != "{}" {
		_ = json.Unmarshal(raw, &payload)
	}
	for key, value := range patch {
		payload[key] = value
	}
	body, _ := json.Marshal(payload)
	return body
}

func ternary(condition bool, ifTrue string, ifFalse string) string {
	if condition {
		return ifTrue
	}
	return ifFalse
}

type phaseTimeouts struct {
	Primary   int
	Guardrail int
	Helper    int
}

func deriveInferencePhaseTimeouts(profile ModelProfile, deps DependencyConfig) phaseTimeouts {
	base := profile.TimeoutSeconds
	if base <= 0 {
		base = 45
	}

	result := phaseTimeouts{
		Primary:   base,
		Guardrail: base,
		Helper:    maxInt(15, base/2),
	}

	if result.Helper > base {
		result.Helper = base
	}

	primaryOverride, guardrailOverride, helperOverride := parsePhaseTimeoutOverrides(profile.ConnectionMetadata)
	if primaryOverride > 0 {
		result.Primary = primaryOverride
	}
	if guardrailOverride > 0 {
		result.Guardrail = guardrailOverride
	}
	if helperOverride > 0 {
		result.Helper = helperOverride
	}

	if deps.GuardrailTimeout > 0 {
		result.Guardrail = durationToSeconds(deps.GuardrailTimeout)
	}
	if deps.HelperTimeout > 0 {
		result.Helper = durationToSeconds(deps.HelperTimeout)
	}

	if result.Primary <= 0 {
		result.Primary = 45
	}
	if result.Guardrail <= 0 {
		result.Guardrail = result.Primary
	}
	if result.Helper <= 0 {
		result.Helper = result.Primary
	}

	return result
}

func parsePhaseTimeoutOverrides(raw json.RawMessage) (int, int, int) {
	if len(raw) == 0 || string(raw) == "{}" {
		return 0, 0, 0
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return 0, 0, 0
	}

	return intOverride(payload, "primary_timeout_seconds"), intOverride(payload, "guardrail_timeout_seconds"), intOverride(payload, "helper_timeout_seconds")
}

func intOverride(values map[string]any, key string) int {
	value, ok := values[key]
	if !ok {
		return 0
	}
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case string:
		number, err := strconv.Atoi(strings.TrimSpace(typed))
		if err != nil {
			return 0
		}
		return number
	default:
		return 0
	}
}

func durationToSeconds(value time.Duration) int {
	seconds := int(value.Seconds())
	if seconds <= 0 {
		return 1
	}
	return seconds
}

func isLocalOllamaProvider(provider string) bool {
	provider = strings.ToLower(strings.TrimSpace(provider))
	return provider == "local_ollama" || provider == "ollama"
}

func limitStrings(values []string, maxCount int, maxChars int) []string {
	if maxCount <= 0 {
		return nil
	}
	out := make([]string, 0, maxCount)
	for _, item := range values {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if maxChars > 0 && len(item) > maxChars {
			item = item[:maxChars] + "..."
		}
		out = append(out, item)
		if len(out) >= maxCount {
			break
		}
	}
	return out
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
