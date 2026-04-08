package runs

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"clawbot-server/internal/db"

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
}

func TestRepositoryRunCyclesDBBacked(t *testing.T) {
	pool := integrationPool(t)
	repo := NewPostgresRepository()
	ctx := context.Background()

	run := createIntegrationRun(ctx, t, repo, pool, string(RunTypeWeekRun), string(ExecutionModeDual))

	cycle, err := repo.CreateCycle(ctx, pool, run.ID, CreateCycleInput{
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

	loaded, err := repo.GetCycle(ctx, pool, run.ID, cycle.ID)
	if err != nil {
		t.Fatalf("GetCycle() error = %v", err)
	}
	if loaded.CycleKey != "day-1" {
		t.Fatalf("unexpected cycle loaded %#v", loaded)
	}

	loaded.Status = string(CycleStatusRunning)
	updated, err := repo.UpdateCycle(ctx, pool, loaded)
	if err != nil {
		t.Fatalf("UpdateCycle() error = %v", err)
	}
	if updated.Status != string(CycleStatusRunning) {
		t.Fatalf("expected running status, got %s", updated.Status)
	}
}

func TestRepositoryRunArtifactsDBBacked(t *testing.T) {
	pool := integrationPool(t)
	repo := NewPostgresRepository()
	ctx := context.Background()

	run := createIntegrationRun(ctx, t, repo, pool, string(RunTypeWeekRun), string(ExecutionModeDual))
	cycle, err := repo.CreateCycle(ctx, pool, run.ID, CreateCycleInput{CycleKey: "day-2", ExecutionRing: string(ExecutionRing2), Status: string(CycleStatusPending), MetadataJSON: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("CreateCycle() error = %v", err)
	}

	artifact, err := repo.CreateArtifact(ctx, pool, run.ID, AttachArtifactInput{
		CycleID:      &cycle.ID,
		ArtifactType: "replay_output",
		URI:          "s3://integration/replay.json",
		ContentType:  "application/json",
		Version:      "v1",
		Checksum:     "abc123",
		MetadataJSON: json.RawMessage(`{"kind":"replay"}`),
	})
	if err != nil {
		t.Fatalf("CreateArtifact() error = %v", err)
	}
	if artifact.ArtifactType != "replay_output" {
		t.Fatalf("unexpected artifact %#v", artifact)
	}

	artifacts, err := repo.ListArtifacts(ctx, pool, run.ID)
	if err != nil {
		t.Fatalf("ListArtifacts() error = %v", err)
	}
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
}

func TestRepositoryRunComparisonsDBBacked(t *testing.T) {
	pool := integrationPool(t)
	repo := NewPostgresRepository()
	ctx := context.Background()

	run := createIntegrationRun(ctx, t, repo, pool, string(RunTypeWeekRun), string(ExecutionModeDual))
	cycle, err := repo.CreateCycle(ctx, pool, run.ID, CreateCycleInput{CycleKey: "day-3", ExecutionRing: string(ExecutionRing2), Status: string(CycleStatusPending), MetadataJSON: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("CreateCycle() error = %v", err)
	}

	comparison, err := repo.UpsertComparison(ctx, pool, run.ID, UpsertComparisonInput{
		CycleID:              &cycle.ID,
		DeterministicSummary: json.RawMessage(`{"precision":0.92}`),
		LLMSummary:           json.RawMessage(`{"recommendation":"tighten thresholds"}`),
		GuardrailSummary:     json.RawMessage(`{"decision":"review"}`),
		Deltas:               json.RawMessage(`{"alerts":"+2"}`),
		ReviewStatus:         string(ReviewStatusReviewPending),
		FinalDisposition:     "pending_review",
		FinalOutput:          json.RawMessage(`{}`),
	})
	if err != nil {
		t.Fatalf("UpsertComparison() error = %v", err)
	}
	if comparison.RunID != run.ID {
		t.Fatalf("unexpected comparison %#v", comparison)
	}

	loaded, err := repo.GetComparison(ctx, pool, run.ID)
	if err != nil {
		t.Fatalf("GetComparison() error = %v", err)
	}
	if string(loaded.DeterministicSummary) != `{"precision":0.92}` {
		t.Fatalf("unexpected comparison summary %#v", loaded)
	}
}
