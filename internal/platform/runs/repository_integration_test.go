package runs

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"clawbot-server/internal/db"
	"clawbot-server/internal/platform/store"

	"github.com/jackc/pgx/v5/pgxpool"
)

func integrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := strings.TrimSpace(os.Getenv("CLAWBOT_TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("CLAWBOT_TEST_DATABASE_URL is not set; skipping DB-backed integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("pgxpool.New() error = %v", err)
	}
	t.Cleanup(pool.Close)

	if err := db.ApplyAll(ctx, pool); err != nil {
		t.Fatalf("ApplyAll() error = %v", err)
	}

	resetIntegrationTables(t, pool)
	return pool
}

func resetIntegrationTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := pool.Exec(ctx, `
TRUNCATE TABLE governance_audit_events, run_review_actions, policy_decisions, run_comparisons, run_artifacts, run_cycles, runs RESTART IDENTITY CASCADE;
DELETE FROM model_profiles WHERE name <> 'ach-default';
`); err != nil {
		t.Fatalf("reset integration tables: %v", err)
	}
}

func createIntegrationRun(ctx context.Context, t *testing.T, repo *PostgresRepository, pool *pgxpool.Pool, runType string, mode string) Run {
	t.Helper()

	run, err := repo.Create(ctx, pool, CreateInput{
		Name:               "integration-run",
		Description:        "integration test",
		Status:             string(RunStatusPending),
		RunType:            runType,
		ExecutionMode:      mode,
		ExecutionRing:      defaultExecutionRingForMode(mode),
		GuardrailStatus:    string(GuardrailStatusDisabled),
		Repo:               "ach-trust-lab",
		Domain:             "ach",
		DatasetRefs:        []string{"data/samples/sample_ach_events.json"},
		PromptPackVersion:  "ach-week/v1",
		RulePackVersion:    "detectors/v1",
		ModelProfile:       "ach-default",
		GuardrailProfile:   "ach-guardian-default",
		MemoryNamespace:    MemoryNamespace{RepoNamespace: "ach-trust-lab", RunNamespace: "integration"},
		RequestedBy:        "integration-suite",
		CreatedBy:          "integration-suite",
		MetadataJSON:       json.RawMessage(`{}`),
		ReviewMetadataJSON: json.RawMessage(`{"approval_required":true}`),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	return run
}

func TestRepositoryModelProfilesDBBacked(t *testing.T) {
	pool := integrationPool(t)
	repo := NewPostgresRepository()
	ctx := context.Background()

	profile, err := repo.RegisterModelProfile(ctx, pool, RegisterModelProfileInput{
		Name:               "ach-integration-profile",
		Provider:           "local_ollama",
		BaseURL:            "http://ai-precision:11434",
		PrimaryModel:       "ibm/granite3.3:8b",
		GuardrailModel:     "ibm/granite3.3-guardian:8b",
		HelperModel:        "granite4:3b",
		TimeoutSeconds:     45,
		Temperature:        0.1,
		MaxTokens:          2048,
		ConnectionMetadata: json.RawMessage(`{"network":"tailscale"}`),
	}, "integration-suite")
	if err != nil {
		t.Fatalf("RegisterModelProfile() error = %v", err)
	}

	loaded, err := repo.GetModelProfile(ctx, pool, profile.Name)
	if err != nil {
		t.Fatalf("GetModelProfile() error = %v", err)
	}
	if loaded.PrimaryModel != "ibm/granite3.3:8b" || loaded.Provider != "local_ollama" {
		t.Fatalf("unexpected loaded model profile %#v", loaded)
	}

	_, err = repo.GetModelProfile(ctx, pool, "missing-profile")
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected store.ErrNotFound for missing model profile, got %v", err)
	}
}

func TestRepositoryRunsLifecycleAndOrderingDBBacked(t *testing.T) {
	pool := integrationPool(t)
	repo := NewPostgresRepository()
	ctx := context.Background()

	run1 := createIntegrationRun(ctx, t, repo, pool, string(RunTypeWeekRun), string(ExecutionModeDual))
	time.Sleep(5 * time.Millisecond)
	run2 := createIntegrationRun(ctx, t, repo, pool, string(RunTypeAgentRun), string(ExecutionModeLLM))

	listed, err := repo.List(ctx, pool)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(listed) < 2 {
		t.Fatalf("expected at least 2 runs, got %d", len(listed))
	}
	if listed[0].ID != run2.ID {
		t.Fatalf("expected most recent run first, got %s then %s", listed[0].ID, listed[1].ID)
	}

	loaded, err := repo.Get(ctx, pool, run1.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if loaded.ExecutionRing != string(ExecutionRing2) || loaded.GuardrailStatus != string(GuardrailStatusDisabled) {
		t.Fatalf("unexpected governance defaults %#v", loaded)
	}

	loaded.Status = string(RunStatusReviewPending)
	loaded.ExecutionRing = string(ExecutionRing3)
	loaded.GuardrailStatus = string(GuardrailStatusFlagged)
	loaded.Notes = "requires reviewer follow-up"
	loaded.ReviewMetadataJSON = json.RawMessage(`{"last_policy_decision_id":"policy-1"}`)
	loaded.MemorySnapshotRefs = []string{"snapshot-1"}
	updated, err := repo.Update(ctx, pool, loaded)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Status != string(RunStatusReviewPending) || updated.ExecutionRing != string(ExecutionRing3) || updated.GuardrailStatus != string(GuardrailStatusFlagged) {
		t.Fatalf("unexpected updated run %#v", updated)
	}
	if updated.Notes != "requires reviewer follow-up" || len(updated.MemorySnapshotRefs) != 1 {
		t.Fatalf("expected notes and memory snapshot refs to persist, got %#v", updated)
	}

	_, err = repo.Get(ctx, pool, "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected store.ErrNotFound for missing run, got %v", err)
	}
}

func TestRepositoryRunCyclesDBBacked(t *testing.T) {
	pool := integrationPool(t)
	repo := NewPostgresRepository()
	ctx := context.Background()

	run := createIntegrationRun(ctx, t, repo, pool, string(RunTypeWeekRun), string(ExecutionModeDual))

	cycle1, err := repo.CreateCycle(ctx, pool, run.ID, CreateCycleInput{
		CycleKey:      "day-1",
		Focus:         "descriptor controls",
		Objective:     "evaluate descriptor risk signals",
		DetectorPack:  "detectors/v1",
		ExecutionRing: string(ExecutionRing2),
		Status:        string(CycleStatusPending),
		MetadataJSON:  json.RawMessage(`{"phase":1}`),
	})
	if err != nil {
		t.Fatalf("CreateCycle() error = %v", err)
	}

	cycle2, err := repo.CreateCycle(ctx, pool, run.ID, CreateCycleInput{
		CycleKey:      "day-2",
		Focus:         "payroll diversion",
		Objective:     "validate transfer controls",
		DetectorPack:  "detectors/v1",
		ExecutionRing: string(ExecutionRing3),
		Status:        string(CycleStatusPending),
		MetadataJSON:  json.RawMessage(`{"phase":2}`),
	})
	if err != nil {
		t.Fatalf("CreateCycle(day-2) error = %v", err)
	}

	loaded, err := repo.GetCycle(ctx, pool, run.ID, cycle1.ID)
	if err != nil {
		t.Fatalf("GetCycle() error = %v", err)
	}
	if loaded.CycleKey != "day-1" {
		t.Fatalf("unexpected cycle loaded %#v", loaded)
	}

	rows, err := pool.Query(ctx, `SELECT cycle_key FROM run_cycles WHERE run_id = $1 ORDER BY created_at ASC`, run.ID)
	if err != nil {
		t.Fatalf("Query(run_cycles) error = %v", err)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			t.Fatalf("Scan(cycle_key) error = %v", err)
		}
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err() = %v", err)
	}
	if len(keys) != 2 || keys[0] != "day-1" || keys[1] != "day-2" {
		t.Fatalf("unexpected cycle order/filtering %#v", keys)
	}

	startedAt := time.Now().UTC()
	finishedAt := startedAt.Add(2 * time.Minute)
	loaded.Status = string(CycleStatusReviewPending)
	loaded.GuardrailStatus = string(GuardrailStatusFlagged)
	loaded.MemorySnapshotRef = "snapshot-cycle-1"
	loaded.StartedAt = &startedAt
	loaded.FinishedAt = &finishedAt
	updated, err := repo.UpdateCycle(ctx, pool, loaded)
	if err != nil {
		t.Fatalf("UpdateCycle() error = %v", err)
	}
	if updated.Status != string(CycleStatusReviewPending) || updated.GuardrailStatus != string(GuardrailStatusFlagged) {
		t.Fatalf("unexpected updated cycle %#v", updated)
	}
	if updated.MemorySnapshotRef != "snapshot-cycle-1" {
		t.Fatalf("expected memory snapshot ref to persist, got %#v", updated)
	}
	if updated.StartedAt == nil || updated.FinishedAt == nil {
		t.Fatalf("expected started_at and finished_at to persist, got %#v", updated)
	}

	_, err = repo.GetCycle(ctx, pool, run.ID, "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected store.ErrNotFound for missing cycle, got %v", err)
	}

	_ = cycle2
}

func TestRepositoryRunArtifactsComparisonsAndGovernanceDBBacked(t *testing.T) {
	pool := integrationPool(t)
	repo := NewPostgresRepository()
	ctx := context.Background()

	run := createIntegrationRun(ctx, t, repo, pool, string(RunTypeWeekRun), string(ExecutionModeDual))
	cycle, err := repo.CreateCycle(ctx, pool, run.ID, CreateCycleInput{CycleKey: "day-2", ExecutionRing: string(ExecutionRing2), Status: string(CycleStatusPending), MetadataJSON: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("CreateCycle() error = %v", err)
	}

	artifactRun, err := repo.CreateArtifact(ctx, pool, run.ID, AttachArtifactInput{
		ArtifactType: "deterministic_output",
		URI:          "s3://integration/deterministic.json",
		ContentType:  "application/json",
		Version:      "v1",
		Checksum:     "det123",
		MetadataJSON: json.RawMessage(`{"kind":"deterministic"}`),
	})
	if err != nil {
		t.Fatalf("CreateArtifact(run) error = %v", err)
	}
	time.Sleep(5 * time.Millisecond)

	artifactCycle, err := repo.CreateArtifact(ctx, pool, run.ID, AttachArtifactInput{
		CycleID:      &cycle.ID,
		ArtifactType: "replay_output",
		URI:          "s3://integration/replay.json",
		ContentType:  "application/json",
		Version:      "v1",
		Checksum:     "abc123",
		MetadataJSON: json.RawMessage(`{"kind":"replay"}`),
	})
	if err != nil {
		t.Fatalf("CreateArtifact(cycle) error = %v", err)
	}
	if artifactCycle.ArtifactType != "replay_output" {
		t.Fatalf("unexpected artifact %#v", artifactCycle)
	}

	artifacts, err := repo.ListArtifacts(ctx, pool, run.ID)
	if err != nil {
		t.Fatalf("ListArtifacts() error = %v", err)
	}
	if len(artifacts) != 2 {
		t.Fatalf("expected 2 artifacts, got %d", len(artifacts))
	}
	if artifacts[0].ID != artifactCycle.ID || artifacts[1].ID != artifactRun.ID {
		t.Fatalf("expected reverse-created ordering, got %#v", artifacts)
	}
	if artifacts[0].CycleID == nil || *artifacts[0].CycleID != cycle.ID {
		t.Fatalf("expected cycle artifact linkage, got %#v", artifacts[0])
	}

	comparison, err := repo.UpsertComparison(ctx, pool, run.ID, UpsertComparisonInput{
		CycleID:              &cycle.ID,
		DeterministicSummary: json.RawMessage(`{"precision":0.92}`),
		LLMSummary:           json.RawMessage(`{"recommendation":"tighten thresholds"}`),
		GuardrailSummary:     json.RawMessage(`{"decision":"review","status":"guardrail_flagged"}`),
		Deltas:               json.RawMessage(`{"alerts":"+2"}`),
		ReviewStatus:         string(ReviewStatusReviewPending),
		FinalDisposition:     "pending_review",
		FinalOutput:          json.RawMessage(`{}`),
	})
	if err != nil {
		t.Fatalf("UpsertComparison(create) error = %v", err)
	}
	if comparison.RunID != run.ID {
		t.Fatalf("unexpected comparison %#v", comparison)
	}
	if comparison.CycleID == nil || *comparison.CycleID != cycle.ID {
		t.Fatalf("expected comparison cycle linkage %#v", comparison)
	}

	comparison, err = repo.UpsertComparison(ctx, pool, run.ID, UpsertComparisonInput{
		CycleID:              &cycle.ID,
		DeterministicSummary: json.RawMessage(`{"precision":0.95}`),
		LLMSummary:           json.RawMessage(`{"recommendation":"promote to shadow mode"}`),
		GuardrailSummary:     json.RawMessage(`{"decision":"allow","status":"guardrail_passed"}`),
		Deltas:               json.RawMessage(`{"alerts":"+1"}`),
		ReviewStatus:         string(ReviewStatusApproved),
		FinalDisposition:     "accepted",
		FinalOutput:          json.RawMessage(`{"decision":"ship"}`),
	})
	if err != nil {
		t.Fatalf("UpsertComparison(update) error = %v", err)
	}
	loaded, err := repo.GetComparison(ctx, pool, run.ID)
	if err != nil {
		t.Fatalf("GetComparison() error = %v", err)
	}
	if string(loaded.DeterministicSummary) != `{"precision":0.95}` || loaded.ReviewStatus != string(ReviewStatusApproved) || loaded.FinalDisposition != "accepted" {
		t.Fatalf("unexpected comparison summary %#v", loaded)
	}

	otherRun := createIntegrationRun(ctx, t, repo, pool, string(RunTypeWeekRun), string(ExecutionModeDual))
	_, err = repo.GetComparison(ctx, pool, otherRun.ID)
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected store.ErrNotFound for missing comparison, got %v", err)
	}

	policyDecision, err := repo.RecordPolicyDecision(ctx, pool, PolicyDecisionInput{
		ActionType:          "run.execute",
		TargetRunID:         &run.ID,
		TargetCycleID:       &cycle.ID,
		ActorID:             "policy-engine",
		ActorType:           "system",
		PolicyInput:         json.RawMessage(`{"execution_mode":"dual","execution_ring":"ring_2"}`),
		Allow:               false,
		PolicyBundleID:      "ach-governance",
		PolicyBundleVersion: "2026.1",
		ReasonCode:          "ring_below_minimum",
		ConditionsApplied:   []string{"requires_ring_2_for_dual"},
		FallbackMode:        "review_required",
	})
	if err != nil {
		t.Fatalf("RecordPolicyDecision() error = %v", err)
	}
	if policyDecision.ID == "" || policyDecision.TargetRunID == nil || *policyDecision.TargetRunID != run.ID {
		t.Fatalf("unexpected policy decision %#v", policyDecision)
	}
	if policyDecision.Allow {
		t.Fatalf("expected deny policy decision %#v", policyDecision)
	}

	reviewAction, err := repo.RecordReviewAction(ctx, pool, run.ID, ReviewActionInput{
		Action:           "defer",
		ReviewerID:       "reviewer-1",
		ReviewerType:     "human",
		Rationale:        "requires additional evidence",
		CycleID:          &cycle.ID,
		PolicyDecisionID: &policyDecision.ID,
	}, string(RunStatusReviewPending), string(RunStatusDeferred))
	if err != nil {
		t.Fatalf("RecordReviewAction() error = %v", err)
	}
	if reviewAction.PolicyDecisionID == nil || *reviewAction.PolicyDecisionID != policyDecision.ID {
		t.Fatalf("expected policy_decision_id linkage %#v", reviewAction)
	}

	firstInput := GovernanceAuditEventInput{
		ActorID:          "policy-engine",
		ActorType:        "system",
		ActionType:       "policy.deny",
		TargetRunID:      &run.ID,
		TargetCycleID:    &cycle.ID,
		PolicyDecisionID: &policyDecision.ID,
		PayloadSummary:   json.RawMessage(`{"reason_code":"ring_below_minimum"}`),
	}
	event1, err := repo.AppendGovernanceAuditEvent(ctx, pool, firstInput)
	if err != nil {
		t.Fatalf("AppendGovernanceAuditEvent(first) error = %v", err)
	}
	if event1.PreviousEventHash != "" {
		t.Fatalf("expected empty previous hash for first event, got %q", event1.PreviousEventHash)
	}
	expectedHash1 := governanceEventHash("", firstInput, firstInput.PayloadSummary)
	if event1.CurrentEventHash != expectedHash1 {
		t.Fatalf("unexpected first hash %q (want %q)", event1.CurrentEventHash, expectedHash1)
	}

	secondInput := GovernanceAuditEventInput{
		ActorID:          "reviewer-1",
		ActorType:        "human",
		ActionType:       "run.review.defer",
		TargetRunID:      &run.ID,
		TargetCycleID:    &cycle.ID,
		TargetArtifactID: &artifactCycle.ID,
		PolicyDecisionID: &policyDecision.ID,
		PayloadSummary:   json.RawMessage(`{"new_status":"deferred"}`),
	}
	event2, err := repo.AppendGovernanceAuditEvent(ctx, pool, secondInput)
	if err != nil {
		t.Fatalf("AppendGovernanceAuditEvent(second) error = %v", err)
	}
	if event2.PreviousEventHash != event1.CurrentEventHash {
		t.Fatalf("expected hash chain linkage, got prev=%q want=%q", event2.PreviousEventHash, event1.CurrentEventHash)
	}
	expectedHash2 := governanceEventHash(event1.CurrentEventHash, secondInput, secondInput.PayloadSummary)
	if event2.CurrentEventHash != expectedHash2 {
		t.Fatalf("unexpected second hash %q (want %q)", event2.CurrentEventHash, expectedHash2)
	}

	var persistedPrev, persistedCurrent string
	if err := pool.QueryRow(ctx, `SELECT previous_event_hash, current_event_hash FROM governance_audit_events WHERE id = $1`, event2.ID).Scan(&persistedPrev, &persistedCurrent); err != nil {
		t.Fatalf("query governance hash fields: %v", err)
	}
	if persistedPrev != event1.CurrentEventHash || persistedCurrent != expectedHash2 {
		t.Fatalf("unexpected persisted governance hash chain prev=%q current=%q", persistedPrev, persistedCurrent)
	}
}
