package runs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"clawbot-server/internal/platform/store"

	"github.com/jackc/pgx/v5"
)

type PostgresRepository struct{}

func NewPostgresRepository() *PostgresRepository {
	return &PostgresRepository{}
}

func (r *PostgresRepository) List(ctx context.Context, q store.DBTX) ([]Run, error) {
	const query = `
SELECT
  id,
  name,
  description,
  status,
  scenario_type,
  run_type,
  execution_mode,
  execution_ring,
  guardrail_status,
  repo,
  domain,
  dataset_refs_json,
  prompt_pack_version,
  rule_pack_version,
  model_profile,
  guardrail_profile,
  memory_namespace_json,
  requested_by,
  created_by,
  created_at,
  updated_at,
  started_at,
  finished_at,
  completed_at,
  artifact_bundle_refs_json,
  memory_snapshot_refs_json,
  review_metadata_json,
  notes,
  metadata_json
FROM runs
ORDER BY created_at DESC
`

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Run, 0)
	for rows.Next() {
		item, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *PostgresRepository) Get(ctx context.Context, q store.DBTX, id string) (Run, error) {
	const query = `
SELECT
  id,
  name,
  description,
  status,
  scenario_type,
  run_type,
  execution_mode,
  execution_ring,
  guardrail_status,
  repo,
  domain,
  dataset_refs_json,
  prompt_pack_version,
  rule_pack_version,
  model_profile,
  guardrail_profile,
  memory_namespace_json,
  requested_by,
  created_by,
  created_at,
  updated_at,
  started_at,
  finished_at,
  completed_at,
  artifact_bundle_refs_json,
  memory_snapshot_refs_json,
  review_metadata_json,
  notes,
  metadata_json
FROM runs
WHERE id = $1
`

	row := q.QueryRow(ctx, query, id)
	item, err := scanRun(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Run{}, store.ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) Create(ctx context.Context, q store.DBTX, input CreateInput) (Run, error) {
	const query = `
INSERT INTO runs (
  name,
  description,
  status,
  scenario_type,
  run_type,
  execution_mode,
  execution_ring,
  guardrail_status,
  repo,
  domain,
  dataset_refs_json,
  prompt_pack_version,
  rule_pack_version,
  model_profile,
  guardrail_profile,
  memory_namespace_json,
  requested_by,
  created_by,
  started_at,
  finished_at,
  completed_at,
  artifact_bundle_refs_json,
  memory_snapshot_refs_json,
  review_metadata_json,
  notes,
  metadata_json
)
VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9,
  $10, $11, $12, $13, $14, $15, $16, $17, $18, $19,
  $20, $20, $21, $22, $23, $24, $25
)
RETURNING
  id,
  name,
  description,
  status,
  scenario_type,
  run_type,
  execution_mode,
  execution_ring,
  guardrail_status,
  repo,
  domain,
  dataset_refs_json,
  prompt_pack_version,
  rule_pack_version,
  model_profile,
  guardrail_profile,
  memory_namespace_json,
  requested_by,
  created_by,
  created_at,
  updated_at,
  started_at,
  finished_at,
  completed_at,
  artifact_bundle_refs_json,
  memory_snapshot_refs_json,
  review_metadata_json,
  notes,
  metadata_json
`

	row := q.QueryRow(ctx, query,
		input.Name,
		input.Description,
		input.Status,
		input.ScenarioType,
		input.RunType,
		input.ExecutionMode,
		input.ExecutionRing,
		input.GuardrailStatus,
		input.Repo,
		input.Domain,
		mustJSON(input.DatasetRefs, []string{}),
		input.PromptPackVersion,
		input.RulePackVersion,
		input.ModelProfile,
		input.GuardrailProfile,
		mustJSON(input.MemoryNamespace, MemoryNamespace{}),
		input.RequestedBy,
		input.CreatedBy,
		input.StartedAt,
		input.FinishedAt,
		mustJSON(input.ArtifactBundleRefs, []string{}),
		mustJSON(input.MemorySnapshotRefs, []string{}),
		defaultRaw(input.ReviewMetadataJSON, json.RawMessage(`{}`)),
		input.Notes,
		defaultRaw(input.MetadataJSON, json.RawMessage(`{}`)),
	)

	return scanRun(row)
}

func (r *PostgresRepository) Update(ctx context.Context, q store.DBTX, item Run) (Run, error) {
	const query = `
UPDATE runs
SET
  name = $2,
  description = $3,
  status = $4,
  scenario_type = $5,
  run_type = $6,
  execution_mode = $7,
  execution_ring = $8,
  guardrail_status = $9,
  repo = $10,
  domain = $11,
  dataset_refs_json = $12,
  prompt_pack_version = $13,
  rule_pack_version = $14,
  model_profile = $15,
  guardrail_profile = $16,
  memory_namespace_json = $17,
  requested_by = $18,
  started_at = $19,
  finished_at = $20,
  completed_at = $20,
  artifact_bundle_refs_json = $21,
  memory_snapshot_refs_json = $22,
  review_metadata_json = $23,
  notes = $24,
  metadata_json = $25,
  updated_at = NOW()
WHERE id = $1
RETURNING
  id,
  name,
  description,
  status,
  scenario_type,
  run_type,
  execution_mode,
  execution_ring,
  guardrail_status,
  repo,
  domain,
  dataset_refs_json,
  prompt_pack_version,
  rule_pack_version,
  model_profile,
  guardrail_profile,
  memory_namespace_json,
  requested_by,
  created_by,
  created_at,
  updated_at,
  started_at,
  finished_at,
  completed_at,
  artifact_bundle_refs_json,
  memory_snapshot_refs_json,
  review_metadata_json,
  notes,
  metadata_json
`

	row := q.QueryRow(ctx, query,
		item.ID,
		item.Name,
		item.Description,
		item.Status,
		item.ScenarioType,
		item.RunType,
		item.ExecutionMode,
		item.ExecutionRing,
		item.GuardrailStatus,
		item.Repo,
		item.Domain,
		mustJSON(item.DatasetRefs, []string{}),
		item.PromptPackVersion,
		item.RulePackVersion,
		item.ModelProfile,
		item.GuardrailProfile,
		mustJSON(item.MemoryNamespace, MemoryNamespace{}),
		item.RequestedBy,
		item.StartedAt,
		item.FinishedAt,
		mustJSON(item.ArtifactBundleRefs, []string{}),
		mustJSON(item.MemorySnapshotRefs, []string{}),
		defaultRaw(item.ReviewMetadataJSON, json.RawMessage(`{}`)),
		item.Notes,
		defaultRaw(item.MetadataJSON, json.RawMessage(`{}`)),
	)

	return scanRun(row)
}

func (r *PostgresRepository) CreateCycle(ctx context.Context, q store.DBTX, runID string, input CreateCycleInput) (Cycle, error) {
	const query = `
INSERT INTO run_cycles (
  run_id,
  cycle_key,
  focus,
  objective,
  detector_pack,
  execution_ring,
  guardrail_status,
  summary_ref,
  carry_forward_summary_ref,
  status,
  metadata_json
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, run_id, cycle_key, focus, objective, detector_pack, execution_ring, guardrail_status, summary_ref, carry_forward_summary_ref, memory_snapshot_ref, status, started_at, finished_at, metadata_json, created_at, updated_at
`

	row := q.QueryRow(ctx, query,
		runID,
		input.CycleKey,
		input.Focus,
		input.Objective,
		input.DetectorPack,
		input.ExecutionRing,
		GuardrailStatusDisabled,
		input.SummaryRef,
		input.CarryForwardSummaryRef,
		input.Status,
		defaultRaw(input.MetadataJSON, json.RawMessage(`{}`)),
	)

	return scanCycle(row)
}

func (r *PostgresRepository) GetCycle(ctx context.Context, q store.DBTX, runID string, cycleID string) (Cycle, error) {
	const query = `
SELECT id, run_id, cycle_key, focus, objective, detector_pack, execution_ring, guardrail_status, summary_ref, carry_forward_summary_ref, memory_snapshot_ref, status, started_at, finished_at, metadata_json, created_at, updated_at
FROM run_cycles
WHERE run_id = $1 AND id = $2
`

	row := q.QueryRow(ctx, query, runID, cycleID)
	item, err := scanCycle(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Cycle{}, store.ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) UpdateCycle(ctx context.Context, q store.DBTX, item Cycle) (Cycle, error) {
	const query = `
UPDATE run_cycles
SET
  focus = $2,
  objective = $3,
  detector_pack = $4,
  execution_ring = $5,
  guardrail_status = $6,
  summary_ref = $7,
  carry_forward_summary_ref = $8,
  memory_snapshot_ref = $9,
  status = $10,
  started_at = $11,
  finished_at = $12,
  metadata_json = $13,
  updated_at = NOW()
WHERE id = $1
RETURNING id, run_id, cycle_key, focus, objective, detector_pack, execution_ring, guardrail_status, summary_ref, carry_forward_summary_ref, memory_snapshot_ref, status, started_at, finished_at, metadata_json, created_at, updated_at
`

	row := q.QueryRow(ctx, query,
		item.ID,
		item.Focus,
		item.Objective,
		item.DetectorPack,
		item.ExecutionRing,
		item.GuardrailStatus,
		item.SummaryRef,
		item.CarryForwardSummaryRef,
		item.MemorySnapshotRef,
		item.Status,
		item.StartedAt,
		item.FinishedAt,
		defaultRaw(item.MetadataJSON, json.RawMessage(`{}`)),
	)

	return scanCycle(row)
}

func (r *PostgresRepository) CreateArtifact(ctx context.Context, q store.DBTX, runID string, input AttachArtifactInput) (Artifact, error) {
	const query = `
INSERT INTO run_artifacts (
  run_id,
  cycle_id,
  artifact_type,
  uri,
  content_type,
  version,
  checksum,
  metadata_json
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, run_id, cycle_id, artifact_type, uri, content_type, version, checksum, metadata_json, created_at
`

	row := q.QueryRow(ctx, query,
		runID,
		nullableString(input.CycleID),
		input.ArtifactType,
		input.URI,
		input.ContentType,
		input.Version,
		input.Checksum,
		defaultRaw(input.MetadataJSON, json.RawMessage(`{}`)),
	)

	return scanArtifact(row)
}

func (r *PostgresRepository) ListArtifacts(ctx context.Context, q store.DBTX, runID string) ([]Artifact, error) {
	const query = `
SELECT id, run_id, cycle_id, artifact_type, uri, content_type, version, checksum, metadata_json, created_at
FROM run_artifacts
WHERE run_id = $1
ORDER BY created_at DESC
`

	rows, err := q.Query(ctx, query, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Artifact, 0)
	for rows.Next() {
		item, err := scanArtifact(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) UpsertComparison(ctx context.Context, q store.DBTX, runID string, input UpsertComparisonInput) (Comparison, error) {
	const query = `
INSERT INTO run_comparisons (
  run_id,
  cycle_id,
  deterministic_summary_json,
  llm_summary_json,
  guardrail_summary_json,
  deltas_json,
  review_status,
  reviewer_notes,
  final_disposition,
  final_output_json
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (run_id)
DO UPDATE SET
  cycle_id = EXCLUDED.cycle_id,
  deterministic_summary_json = EXCLUDED.deterministic_summary_json,
  llm_summary_json = EXCLUDED.llm_summary_json,
  guardrail_summary_json = EXCLUDED.guardrail_summary_json,
  deltas_json = EXCLUDED.deltas_json,
  review_status = EXCLUDED.review_status,
  reviewer_notes = EXCLUDED.reviewer_notes,
  final_disposition = EXCLUDED.final_disposition,
  final_output_json = EXCLUDED.final_output_json,
  updated_at = NOW()
RETURNING id, run_id, cycle_id, deterministic_summary_json, llm_summary_json, guardrail_summary_json, deltas_json, review_status, reviewer_notes, final_disposition, final_output_json, created_at, updated_at
`

	row := q.QueryRow(ctx, query,
		runID,
		nullableString(input.CycleID),
		defaultRaw(input.DeterministicSummary, json.RawMessage(`{}`)),
		defaultRaw(input.LLMSummary, json.RawMessage(`{}`)),
		defaultRaw(input.GuardrailSummary, json.RawMessage(`{}`)),
		defaultRaw(input.Deltas, json.RawMessage(`{}`)),
		input.ReviewStatus,
		input.ReviewerNotes,
		input.FinalDisposition,
		defaultRaw(input.FinalOutput, json.RawMessage(`{}`)),
	)

	return scanComparison(row)
}

func (r *PostgresRepository) GetComparison(ctx context.Context, q store.DBTX, runID string) (Comparison, error) {
	const query = `
SELECT id, run_id, cycle_id, deterministic_summary_json, llm_summary_json, guardrail_summary_json, deltas_json, review_status, reviewer_notes, final_disposition, final_output_json, created_at, updated_at
FROM run_comparisons
WHERE run_id = $1
`

	row := q.QueryRow(ctx, query, runID)
	item, err := scanComparison(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Comparison{}, store.ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) RegisterModelProfile(ctx context.Context, q store.DBTX, input RegisterModelProfileInput, actor string) (ModelProfile, error) {
	const query = `
INSERT INTO model_profiles (
  name,
  provider,
  base_url,
  primary_model,
  guardrail_model,
  helper_model,
  timeout_seconds,
  temperature,
  max_tokens,
  json_mode,
  structured_output,
  enable_guardrails,
  enable_helper_model,
  connection_metadata_json,
  created_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
ON CONFLICT (name)
DO UPDATE SET
  provider = EXCLUDED.provider,
  base_url = EXCLUDED.base_url,
  primary_model = EXCLUDED.primary_model,
  guardrail_model = EXCLUDED.guardrail_model,
  helper_model = EXCLUDED.helper_model,
  timeout_seconds = EXCLUDED.timeout_seconds,
  temperature = EXCLUDED.temperature,
  max_tokens = EXCLUDED.max_tokens,
  json_mode = EXCLUDED.json_mode,
  structured_output = EXCLUDED.structured_output,
  enable_guardrails = EXCLUDED.enable_guardrails,
  enable_helper_model = EXCLUDED.enable_helper_model,
  connection_metadata_json = EXCLUDED.connection_metadata_json,
  updated_at = NOW()
RETURNING id, name, provider, base_url, primary_model, guardrail_model, helper_model, timeout_seconds, temperature, max_tokens, json_mode, structured_output, enable_guardrails, enable_helper_model, connection_metadata_json, created_by, created_at, updated_at
`

	jsonMode := valueOrDefaultBool(input.JSONMode, true)
	structuredOutput := valueOrDefaultBool(input.StructuredOutput, true)
	enableGuardrails := valueOrDefaultBool(input.EnableGuardrails, true)
	enableHelper := valueOrDefaultBool(input.EnableHelperModel, true)

	row := q.QueryRow(ctx, query,
		input.Name,
		input.Provider,
		input.BaseURL,
		input.PrimaryModel,
		input.GuardrailModel,
		input.HelperModel,
		input.TimeoutSeconds,
		input.Temperature,
		input.MaxTokens,
		jsonMode,
		structuredOutput,
		enableGuardrails,
		enableHelper,
		defaultRaw(input.ConnectionMetadata, json.RawMessage(`{}`)),
		actor,
	)

	return scanModelProfile(row)
}

func (r *PostgresRepository) GetModelProfile(ctx context.Context, q store.DBTX, idOrName string) (ModelProfile, error) {
	const query = `
SELECT id, name, provider, base_url, primary_model, guardrail_model, helper_model, timeout_seconds, temperature, max_tokens, json_mode, structured_output, enable_guardrails, enable_helper_model, connection_metadata_json, created_by, created_at, updated_at
FROM model_profiles
WHERE id::text = $1 OR name = $1
LIMIT 1
`

	row := q.QueryRow(ctx, query, idOrName)
	item, err := scanModelProfile(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return ModelProfile{}, store.ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) RecordPolicyDecision(ctx context.Context, q store.DBTX, input PolicyDecisionInput) (PolicyDecision, error) {
	const query = `
INSERT INTO policy_decisions (
  action_type,
  target_run_id,
  target_cycle_id,
  actor_id,
  actor_type,
  policy_input_json,
  allow,
  policy_bundle_id,
  policy_bundle_version,
  reason_code,
  conditions_applied_json,
  fallback_mode
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING id, action_type, target_run_id, target_cycle_id, actor_id, actor_type, policy_input_json, allow, policy_bundle_id, policy_bundle_version, reason_code, conditions_applied_json, fallback_mode, created_at
`

	row := q.QueryRow(ctx, query,
		input.ActionType,
		nullableString(input.TargetRunID),
		nullableString(input.TargetCycleID),
		input.ActorID,
		input.ActorType,
		defaultRaw(input.PolicyInput, json.RawMessage(`{}`)),
		input.Allow,
		input.PolicyBundleID,
		input.PolicyBundleVersion,
		input.ReasonCode,
		mustJSON(input.ConditionsApplied, []string{}),
		input.FallbackMode,
	)

	var (
		item           PolicyDecision
		targetRunID    *string
		targetCycleID  *string
		policyInput    []byte
		conditionsJSON []byte
	)
	if err := row.Scan(
		&item.ID,
		&item.ActionType,
		&targetRunID,
		&targetCycleID,
		&item.ActorID,
		&item.ActorType,
		&policyInput,
		&item.Allow,
		&item.PolicyBundleID,
		&item.PolicyBundleVersion,
		&item.ReasonCode,
		&conditionsJSON,
		&item.FallbackMode,
		&item.CreatedAt,
	); err != nil {
		return PolicyDecision{}, err
	}
	item.TargetRunID = targetRunID
	item.TargetCycleID = targetCycleID
	item.PolicyInput = defaultRaw(policyInput, json.RawMessage(`{}`))
	_ = json.Unmarshal(conditionsJSON, &item.ConditionsApplied)
	if len(item.ConditionsApplied) == 0 {
		item.ConditionsApplied = []string{}
	}

	return item, nil
}

func (r *PostgresRepository) RecordReviewAction(ctx context.Context, q store.DBTX, runID string, input ReviewActionInput, priorStatus string, newStatus string) (ReviewActionRecord, error) {
	const query = `
INSERT INTO run_review_actions (
  run_id,
  cycle_id,
  reviewer_id,
  reviewer_type,
  action_type,
  prior_status,
  new_status,
  rationale,
  policy_decision_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, run_id, cycle_id, reviewer_id, reviewer_type, action_type, prior_status, new_status, rationale, policy_decision_id, created_at
`

	row := q.QueryRow(ctx, query,
		runID,
		nullableString(input.CycleID),
		input.ReviewerID,
		input.ReviewerType,
		input.Action,
		priorStatus,
		newStatus,
		input.Rationale,
		nullableString(input.PolicyDecisionID),
	)

	var (
		item             ReviewActionRecord
		cycleID          *string
		policyDecisionID *string
	)
	if err := row.Scan(
		&item.ID,
		&item.RunID,
		&cycleID,
		&item.ReviewerID,
		&item.ReviewerType,
		&item.ActionType,
		&item.PriorStatus,
		&item.NewStatus,
		&item.Rationale,
		&policyDecisionID,
		&item.CreatedAt,
	); err != nil {
		return ReviewActionRecord{}, err
	}
	item.CycleID = cycleID
	item.PolicyDecisionID = policyDecisionID

	return item, nil
}

func (r *PostgresRepository) AppendGovernanceAuditEvent(ctx context.Context, q store.DBTX, input GovernanceAuditEventInput) (GovernanceAuditEvent, error) {
	const latestQuery = `
SELECT current_event_hash
FROM governance_audit_events
ORDER BY created_at DESC
LIMIT 1
`

	previousHash := ""
	if err := q.QueryRow(ctx, latestQuery).Scan(&previousHash); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return GovernanceAuditEvent{}, err
	}

	payload := defaultRaw(input.PayloadSummary, json.RawMessage(`{}`))
	currentHash := governanceEventHash(previousHash, input, payload)

	const insertQuery = `
INSERT INTO governance_audit_events (
  previous_event_hash,
  current_event_hash,
  actor_id,
  actor_type,
  action_type,
  target_run_id,
  target_cycle_id,
  target_artifact_id,
  policy_decision_id,
  payload_summary_json
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, previous_event_hash, current_event_hash, actor_id, actor_type, action_type, target_run_id, target_cycle_id, target_artifact_id, policy_decision_id, payload_summary_json, created_at
`

	row := q.QueryRow(ctx, insertQuery,
		previousHash,
		currentHash,
		input.ActorID,
		input.ActorType,
		input.ActionType,
		nullableString(input.TargetRunID),
		nullableString(input.TargetCycleID),
		nullableString(input.TargetArtifactID),
		nullableString(input.PolicyDecisionID),
		payload,
	)

	var (
		item             GovernanceAuditEvent
		targetRunID      *string
		targetCycleID    *string
		targetArtifactID *string
		policyDecisionID *string
		payloadSummary   []byte
	)
	if err := row.Scan(
		&item.ID,
		&item.PreviousEventHash,
		&item.CurrentEventHash,
		&item.ActorID,
		&item.ActorType,
		&item.ActionType,
		&targetRunID,
		&targetCycleID,
		&targetArtifactID,
		&policyDecisionID,
		&payloadSummary,
		&item.CreatedAt,
	); err != nil {
		return GovernanceAuditEvent{}, err
	}
	item.TargetRunID = targetRunID
	item.TargetCycleID = targetCycleID
	item.TargetArtifactID = targetArtifactID
	item.PolicyDecisionID = policyDecisionID
	item.PayloadSummary = defaultRaw(payloadSummary, json.RawMessage(`{}`))

	return item, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanRun(s scanner) (Run, error) {
	var (
		item                   Run
		datasetRefsJSON        []byte
		memoryNamespaceJSON    []byte
		artifactBundleRefsJSON []byte
		memorySnapshotRefsJSON []byte
		reviewMetadataJSON     []byte
		metadataJSON           []byte
	)

	err := s.Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.Status,
		&item.ScenarioType,
		&item.RunType,
		&item.ExecutionMode,
		&item.ExecutionRing,
		&item.GuardrailStatus,
		&item.Repo,
		&item.Domain,
		&datasetRefsJSON,
		&item.PromptPackVersion,
		&item.RulePackVersion,
		&item.ModelProfile,
		&item.GuardrailProfile,
		&memoryNamespaceJSON,
		&item.RequestedBy,
		&item.CreatedBy,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.StartedAt,
		&item.FinishedAt,
		&item.CompletedAt,
		&artifactBundleRefsJSON,
		&memorySnapshotRefsJSON,
		&reviewMetadataJSON,
		&item.Notes,
		&metadataJSON,
	)
	if err != nil {
		return Run{}, err
	}

	_ = json.Unmarshal(datasetRefsJSON, &item.DatasetRefs)
	if len(item.DatasetRefs) == 0 {
		item.DatasetRefs = []string{}
	}

	_ = json.Unmarshal(memoryNamespaceJSON, &item.MemoryNamespace)
	item.MemoryNamespace = normalizeMemoryNamespace(item.MemoryNamespace, item.Repo, item.Domain, item.ID)

	_ = json.Unmarshal(artifactBundleRefsJSON, &item.ArtifactBundleRefs)
	if len(item.ArtifactBundleRefs) == 0 {
		item.ArtifactBundleRefs = []string{}
	}

	_ = json.Unmarshal(memorySnapshotRefsJSON, &item.MemorySnapshotRefs)
	if len(item.MemorySnapshotRefs) == 0 {
		item.MemorySnapshotRefs = []string{}
	}

	item.ReviewMetadataJSON = defaultRaw(reviewMetadataJSON, json.RawMessage(`{}`))
	item.MetadataJSON = defaultRaw(metadataJSON, json.RawMessage(`{}`))

	if item.RequestedBy == "" {
		item.RequestedBy = item.CreatedBy
	}
	if item.FinishedAt == nil {
		item.FinishedAt = item.CompletedAt
	}
	if strings.TrimSpace(item.ExecutionRing) == "" {
		item.ExecutionRing = defaultExecutionRingForMode(item.ExecutionMode)
	}
	if strings.TrimSpace(item.GuardrailStatus) == "" {
		item.GuardrailStatus = string(GuardrailStatusDisabled)
	}

	return item, nil
}

func scanCycle(s scanner) (Cycle, error) {
	var (
		item         Cycle
		metadataJSON []byte
	)

	err := s.Scan(
		&item.ID,
		&item.RunID,
		&item.CycleKey,
		&item.Focus,
		&item.Objective,
		&item.DetectorPack,
		&item.ExecutionRing,
		&item.GuardrailStatus,
		&item.SummaryRef,
		&item.CarryForwardSummaryRef,
		&item.MemorySnapshotRef,
		&item.Status,
		&item.StartedAt,
		&item.FinishedAt,
		&metadataJSON,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return Cycle{}, err
	}

	item.MetadataJSON = defaultRaw(metadataJSON, json.RawMessage(`{}`))
	if strings.TrimSpace(item.ExecutionRing) == "" {
		item.ExecutionRing = string(ExecutionRing1)
	}
	if strings.TrimSpace(item.GuardrailStatus) == "" {
		item.GuardrailStatus = string(GuardrailStatusDisabled)
	}
	return item, nil
}

func scanArtifact(s scanner) (Artifact, error) {
	var (
		item         Artifact
		cycleID      *string
		metadataJSON []byte
	)

	err := s.Scan(
		&item.ID,
		&item.RunID,
		&cycleID,
		&item.ArtifactType,
		&item.URI,
		&item.ContentType,
		&item.Version,
		&item.Checksum,
		&metadataJSON,
		&item.CreatedAt,
	)
	if err != nil {
		return Artifact{}, err
	}

	item.CycleID = cycleID
	item.MetadataJSON = defaultRaw(metadataJSON, json.RawMessage(`{}`))
	return item, nil
}

func scanComparison(s scanner) (Comparison, error) {
	var (
		item                 Comparison
		cycleID              *string
		deterministicSummary []byte
		llmSummary           []byte
		guardrailSummary     []byte
		deltas               []byte
		finalOutput          []byte
	)

	err := s.Scan(
		&item.ID,
		&item.RunID,
		&cycleID,
		&deterministicSummary,
		&llmSummary,
		&guardrailSummary,
		&deltas,
		&item.ReviewStatus,
		&item.ReviewerNotes,
		&item.FinalDisposition,
		&finalOutput,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return Comparison{}, err
	}

	item.CycleID = cycleID
	item.DeterministicSummary = defaultRaw(deterministicSummary, json.RawMessage(`{}`))
	item.LLMSummary = defaultRaw(llmSummary, json.RawMessage(`{}`))
	item.GuardrailSummary = defaultRaw(guardrailSummary, json.RawMessage(`{}`))
	item.Deltas = defaultRaw(deltas, json.RawMessage(`{}`))
	item.FinalOutput = defaultRaw(finalOutput, json.RawMessage(`{}`))
	return item, nil
}

func scanModelProfile(s scanner) (ModelProfile, error) {
	var (
		item               ModelProfile
		connectionMetadata []byte
	)

	err := s.Scan(
		&item.ID,
		&item.Name,
		&item.Provider,
		&item.BaseURL,
		&item.PrimaryModel,
		&item.GuardrailModel,
		&item.HelperModel,
		&item.TimeoutSeconds,
		&item.Temperature,
		&item.MaxTokens,
		&item.JSONMode,
		&item.StructuredOutput,
		&item.EnableGuardrails,
		&item.EnableHelperModel,
		&connectionMetadata,
		&item.CreatedBy,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return ModelProfile{}, err
	}

	item.ConnectionMetadata = defaultRaw(connectionMetadata, json.RawMessage(`{}`))
	return item, nil
}

func defaultRaw(raw []byte, fallback json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return append(json.RawMessage(nil), fallback...)
	}
	return append(json.RawMessage(nil), raw...)
}

func mustJSON(value any, fallback any) json.RawMessage {
	body, err := json.Marshal(value)
	if err != nil {
		body, _ = json.Marshal(fallback)
	}
	return body
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	if *value == "" {
		return nil
	}
	return *value
}

func valueOrDefaultBool(ptr *bool, fallback bool) bool {
	if ptr == nil {
		return fallback
	}
	return *ptr
}

func governanceEventHash(previousHash string, input GovernanceAuditEventInput, payload json.RawMessage) string {
	type canonical struct {
		PreviousHash     string          `json:"previous_hash"`
		ActorID          string          `json:"actor_id"`
		ActorType        string          `json:"actor_type"`
		ActionType       string          `json:"action_type"`
		TargetRunID      *string         `json:"target_run_id,omitempty"`
		TargetCycleID    *string         `json:"target_cycle_id,omitempty"`
		TargetArtifactID *string         `json:"target_artifact_id,omitempty"`
		PolicyDecisionID *string         `json:"policy_decision_id,omitempty"`
		Payload          json.RawMessage `json:"payload"`
		Timestamp        int64           `json:"timestamp"`
	}

	body, _ := json.Marshal(canonical{
		PreviousHash:     previousHash,
		ActorID:          input.ActorID,
		ActorType:        input.ActorType,
		ActionType:       input.ActionType,
		TargetRunID:      input.TargetRunID,
		TargetCycleID:    input.TargetCycleID,
		TargetArtifactID: input.TargetArtifactID,
		PolicyDecisionID: input.PolicyDecisionID,
		Payload:          payload,
		Timestamp:        time.Now().UTC().UnixNano(),
	})
	digest := sha256.Sum256(body)
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(digest[:]))
}
