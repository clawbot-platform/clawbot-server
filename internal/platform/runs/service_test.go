package runs

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"clawbot-server/internal/platform/audit"
	"clawbot-server/internal/platform/scheduler"
	"clawbot-server/internal/platform/store"
)

type transactorStub struct{}

func (transactorStub) InTx(ctx context.Context, fn func(context.Context, store.DBTX) error) error {
	return fn(ctx, nil)
}

type repositoryStub struct {
	runsByID        map[string]Run
	runOrder        []string
	cyclesByID      map[string]Cycle
	artifacts       []Artifact
	comparisonByRun map[string]Comparison
	profilesByName  map[string]ModelProfile
	profilesByID    map[string]ModelProfile
	nextRunID       int
	nextCycleID     int
	nextArtifactID  int
	nextCompareID   int
	nextProfileID   int
}

func newRepositoryStub() *repositoryStub {
	return &repositoryStub{
		runsByID:        map[string]Run{},
		cyclesByID:      map[string]Cycle{},
		artifacts:       []Artifact{},
		comparisonByRun: map[string]Comparison{},
		profilesByName:  map[string]ModelProfile{},
		profilesByID:    map[string]ModelProfile{},
	}
}

func (s *repositoryStub) List(context.Context, store.DBTX) ([]Run, error) {
	items := make([]Run, 0, len(s.runOrder))
	for _, id := range s.runOrder {
		items = append(items, s.runsByID[id])
	}
	return items, nil
}

func (s *repositoryStub) Get(_ context.Context, _ store.DBTX, id string) (Run, error) {
	item, ok := s.runsByID[id]
	if !ok {
		return Run{}, store.ErrNotFound
	}
	return item, nil
}

func (s *repositoryStub) Create(_ context.Context, _ store.DBTX, input CreateInput) (Run, error) {
	s.nextRunID++
	id := fmt.Sprintf("run-%d", s.nextRunID)
	item := Run{
		ID:                 id,
		Name:               input.Name,
		Description:        input.Description,
		Status:             input.Status,
		ScenarioType:       input.ScenarioType,
		RunType:            input.RunType,
		ExecutionMode:      input.ExecutionMode,
		Repo:               input.Repo,
		Domain:             input.Domain,
		DatasetRefs:        input.DatasetRefs,
		PromptPackVersion:  input.PromptPackVersion,
		RulePackVersion:    input.RulePackVersion,
		ModelProfile:       input.ModelProfile,
		GuardrailProfile:   input.GuardrailProfile,
		MemoryNamespace:    input.MemoryNamespace,
		RequestedBy:        input.RequestedBy,
		CreatedBy:          input.CreatedBy,
		StartedAt:          input.StartedAt,
		FinishedAt:         input.FinishedAt,
		CompletedAt:        input.FinishedAt,
		ArtifactBundleRefs: input.ArtifactBundleRefs,
		MemorySnapshotRefs: input.MemorySnapshotRefs,
		ReviewMetadataJSON: input.ReviewMetadataJSON,
		Notes:              input.Notes,
		MetadataJSON:       input.MetadataJSON,
	}
	s.runsByID[id] = item
	s.runOrder = append([]string{id}, s.runOrder...)
	return item, nil
}

func (s *repositoryStub) Update(_ context.Context, _ store.DBTX, item Run) (Run, error) {
	if _, ok := s.runsByID[item.ID]; !ok {
		return Run{}, store.ErrNotFound
	}
	s.runsByID[item.ID] = item
	return item, nil
}

func (s *repositoryStub) CreateCycle(_ context.Context, _ store.DBTX, runID string, input CreateCycleInput) (Cycle, error) {
	s.nextCycleID++
	id := fmt.Sprintf("cycle-%d", s.nextCycleID)
	item := Cycle{
		ID:                     id,
		RunID:                  runID,
		CycleKey:               input.CycleKey,
		Focus:                  input.Focus,
		Objective:              input.Objective,
		DetectorPack:           input.DetectorPack,
		SummaryRef:             input.SummaryRef,
		CarryForwardSummaryRef: input.CarryForwardSummaryRef,
		Status:                 input.Status,
		MetadataJSON:           input.MetadataJSON,
	}
	s.cyclesByID[id] = item
	return item, nil
}

func (s *repositoryStub) GetCycle(_ context.Context, _ store.DBTX, runID string, cycleID string) (Cycle, error) {
	item, ok := s.cyclesByID[cycleID]
	if !ok || item.RunID != runID {
		return Cycle{}, store.ErrNotFound
	}
	return item, nil
}

func (s *repositoryStub) UpdateCycle(_ context.Context, _ store.DBTX, item Cycle) (Cycle, error) {
	if _, ok := s.cyclesByID[item.ID]; !ok {
		return Cycle{}, store.ErrNotFound
	}
	s.cyclesByID[item.ID] = item
	return item, nil
}

func (s *repositoryStub) CreateArtifact(_ context.Context, _ store.DBTX, runID string, input AttachArtifactInput) (Artifact, error) {
	s.nextArtifactID++
	id := fmt.Sprintf("artifact-%d", s.nextArtifactID)
	item := Artifact{
		ID:           id,
		RunID:        runID,
		CycleID:      input.CycleID,
		ArtifactType: input.ArtifactType,
		URI:          input.URI,
		ContentType:  input.ContentType,
		Version:      input.Version,
		Checksum:     input.Checksum,
		MetadataJSON: input.MetadataJSON,
	}
	s.artifacts = append([]Artifact{item}, s.artifacts...)
	return item, nil
}

func (s *repositoryStub) ListArtifacts(_ context.Context, _ store.DBTX, runID string) ([]Artifact, error) {
	items := make([]Artifact, 0)
	for _, item := range s.artifacts {
		if item.RunID == runID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (s *repositoryStub) UpsertComparison(_ context.Context, _ store.DBTX, runID string, input UpsertComparisonInput) (Comparison, error) {
	existing, ok := s.comparisonByRun[runID]
	if !ok {
		s.nextCompareID++
		existing.ID = fmt.Sprintf("comparison-%d", s.nextCompareID)
		existing.RunID = runID
	}
	existing.CycleID = input.CycleID
	existing.DeterministicSummary = input.DeterministicSummary
	existing.LLMSummary = input.LLMSummary
	existing.GuardrailSummary = input.GuardrailSummary
	existing.Deltas = input.Deltas
	existing.ReviewStatus = input.ReviewStatus
	existing.ReviewerNotes = input.ReviewerNotes
	existing.FinalDisposition = input.FinalDisposition
	existing.FinalOutput = input.FinalOutput
	s.comparisonByRun[runID] = existing
	return existing, nil
}

func (s *repositoryStub) GetComparison(_ context.Context, _ store.DBTX, runID string) (Comparison, error) {
	item, ok := s.comparisonByRun[runID]
	if !ok {
		return Comparison{}, store.ErrNotFound
	}
	return item, nil
}

func (s *repositoryStub) RegisterModelProfile(_ context.Context, _ store.DBTX, input RegisterModelProfileInput, actor string) (ModelProfile, error) {
	item, ok := s.profilesByName[input.Name]
	if !ok {
		s.nextProfileID++
		item.ID = fmt.Sprintf("profile-%d", s.nextProfileID)
		item.CreatedBy = actor
	}
	jsonMode := true
	if input.JSONMode != nil {
		jsonMode = *input.JSONMode
	}
	structured := true
	if input.StructuredOutput != nil {
		structured = *input.StructuredOutput
	}
	enableGuardrails := true
	if input.EnableGuardrails != nil {
		enableGuardrails = *input.EnableGuardrails
	}
	enableHelper := true
	if input.EnableHelperModel != nil {
		enableHelper = *input.EnableHelperModel
	}

	item.Name = input.Name
	item.Provider = input.Provider
	item.BaseURL = input.BaseURL
	item.PrimaryModel = input.PrimaryModel
	item.GuardrailModel = input.GuardrailModel
	item.HelperModel = input.HelperModel
	item.TimeoutSeconds = input.TimeoutSeconds
	item.Temperature = input.Temperature
	item.MaxTokens = input.MaxTokens
	item.JSONMode = jsonMode
	item.StructuredOutput = structured
	item.EnableGuardrails = enableGuardrails
	item.EnableHelperModel = enableHelper
	item.ConnectionMetadata = input.ConnectionMetadata

	s.profilesByName[item.Name] = item
	s.profilesByID[item.ID] = item
	return item, nil
}

func (s *repositoryStub) GetModelProfile(_ context.Context, _ store.DBTX, idOrName string) (ModelProfile, error) {
	if item, ok := s.profilesByID[idOrName]; ok {
		return item, nil
	}
	if item, ok := s.profilesByName[idOrName]; ok {
		return item, nil
	}
	return ModelProfile{}, store.ErrNotFound
}

type auditStub struct {
	events []audit.Event
}

func (s *auditStub) Record(_ context.Context, _ store.DBTX, event audit.Event) error {
	s.events = append(s.events, event)
	return nil
}

type schedulerStub struct {
	signals []scheduler.Signal
}

func (s *schedulerStub) RecordRunIntent(_ context.Context, _ store.DBTX, signal scheduler.Signal) error {
	s.signals = append(s.signals, signal)
	return nil
}

type memoryStub struct {
	lastNamespace MemoryNamespace
	lastInput     MemoryWriteInput
	snapshotRef   string
	fetched       int
	fetchedNS     MemoryNamespace
	context       MemoryContext
}

func (s *memoryStub) BaseURL() string { return "http://clawmem.internal" }
func (s *memoryStub) FetchScopedContext(_ context.Context, ns MemoryNamespace) (MemoryContext, error) {
	s.fetched++
	s.fetchedNS = ns
	return s.context, nil
}
func (s *memoryStub) PersistScopedNotes(_ context.Context, ns MemoryNamespace, input MemoryWriteInput) (string, error) {
	s.lastNamespace = ns
	s.lastInput = input
	if s.snapshotRef == "" {
		s.snapshotRef = "snapshot-1"
	}
	return s.snapshotRef, nil
}

type inferenceStub struct {
	calls     int
	lastInput InferenceRequest
	response  InferenceResponse
	returnErr error
}

func (s *inferenceStub) BaseURL() string { return "http://ai-precision" }
func (s *inferenceStub) Execute(_ context.Context, input InferenceRequest) (InferenceResponse, error) {
	s.calls++
	s.lastInput = input
	if s.returnErr != nil {
		return InferenceResponse{}, s.returnErr
	}
	return s.response, nil
}

func TestRunSpecCreateAndDefaults(t *testing.T) {
	repo := newRepositoryStub()
	audits := &auditStub{}
	schedulerSvc := &schedulerStub{}
	manager := NewManager(nil, transactorStub{}, repo, audits, schedulerSvc)

	created, err := manager.Create(context.Background(), CreateInput{
		Name:        "  Week Run 1  ",
		Description: "  ach showcase  ",
		Repo:        "ach-trust-lab",
		Domain:      "ach",
	}, "program-owner")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.RunType != string(RunTypeReplayRun) {
		t.Fatalf("expected default run type replay_run, got %s", created.RunType)
	}
	if created.ExecutionMode != string(ExecutionModeDeterministic) {
		t.Fatalf("expected default execution mode deterministic, got %s", created.ExecutionMode)
	}
	if created.ModelProfile != "ach-default" {
		t.Fatalf("expected default model profile, got %s", created.ModelProfile)
	}
	if created.RequestedBy != "program-owner" {
		t.Fatalf("expected requested_by from actor, got %s", created.RequestedBy)
	}
	if created.MemoryNamespace.RepoNamespace != "ach-trust-lab" {
		t.Fatalf("expected repo namespace mapping, got %#v", created.MemoryNamespace)
	}
	if len(audits.events) != 1 || audits.events[0].EventType != "run.created" {
		t.Fatalf("unexpected audit events %#v", audits.events)
	}
	if len(schedulerSvc.signals) != 1 || schedulerSvc.signals[0].Reason != "run.created" {
		t.Fatalf("unexpected scheduler signals %#v", schedulerSvc.signals)
	}
}

func TestExecutionModeValidation(t *testing.T) {
	repo := newRepositoryStub()
	manager := NewManager(nil, transactorStub{}, repo, &auditStub{}, &schedulerStub{})

	_, err := manager.Create(context.Background(), CreateInput{Name: "bad", ExecutionMode: "unsupported"}, "actor")
	if err == nil {
		t.Fatal("expected execution mode validation error")
	}
}

func TestWeekRunCycleLifecycleAndMemoryNamespace(t *testing.T) {
	repo := newRepositoryStub()
	memory := &memoryStub{snapshotRef: "snapshot-cycle-1"}
	manager := NewManagerWithIntegrations(
		nil,
		transactorStub{},
		repo,
		&auditStub{},
		&schedulerStub{},
		memory,
		&inferenceStub{},
		DependencyConfig{ClawmemBaseURL: "http://clawmem.internal", InferenceBaseURL: "http://ai-precision"},
	)

	run, err := manager.Create(context.Background(), CreateInput{
		Name:          "week-run",
		RunType:       string(RunTypeWeekRun),
		ExecutionMode: string(ExecutionModeDual),
		Repo:          "ach-trust-lab",
	}, "owner")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	cycle, err := manager.CreateCycle(context.Background(), run.ID, CreateCycleInput{
		CycleKey: "day-1",
		Focus:    "descriptor controls",
		Status:   string(CycleStatusPending),
	}, "owner")
	if err != nil {
		t.Fatalf("CreateCycle() error = %v", err)
	}

	running := string(CycleStatusRunning)
	updated, err := manager.UpdateCycle(context.Background(), run.ID, cycle.ID, UpdateCycleInput{Status: &running}, "owner")
	if err != nil {
		t.Fatalf("UpdateCycle(running) error = %v", err)
	}
	if updated.Status != string(CycleStatusRunning) {
		t.Fatalf("expected running status, got %s", updated.Status)
	}

	reviewPending := string(CycleStatusReviewPending)
	note := "carry-forward risk needs investigation"
	agentNamespace := "daily-summary"
	updated, err = manager.UpdateCycle(context.Background(), run.ID, cycle.ID, UpdateCycleInput{
		Status:         &reviewPending,
		MemoryNote:     &note,
		AgentNamespace: &agentNamespace,
	}, "owner")
	if err != nil {
		t.Fatalf("UpdateCycle(review_pending) error = %v", err)
	}

	if updated.MemorySnapshotRef != "snapshot-cycle-1" {
		t.Fatalf("expected memory snapshot ref, got %s", updated.MemorySnapshotRef)
	}
	if memory.lastNamespace.RepoNamespace != "ach-trust-lab" || memory.lastNamespace.CycleNamespace != "day-1" || memory.lastNamespace.AgentNamespace != "daily-summary" {
		t.Fatalf("unexpected memory namespace mapping %#v", memory.lastNamespace)
	}
	if memory.lastInput.Note != note {
		t.Fatalf("expected memory note propagation, got %#v", memory.lastInput)
	}
}

func TestArtifactRegistration(t *testing.T) {
	repo := newRepositoryStub()
	manager := NewManager(nil, transactorStub{}, repo, &auditStub{}, &schedulerStub{})

	run, err := manager.Create(context.Background(), CreateInput{Name: "artifact-run"}, "owner")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	artifact, err := manager.AttachArtifact(context.Background(), run.ID, AttachArtifactInput{
		ArtifactType: "replay_output",
		URI:          "s3://runs/run-1/replay.json",
		Version:      "v1",
	}, "owner")
	if err != nil {
		t.Fatalf("AttachArtifact() error = %v", err)
	}
	if artifact.RunID != run.ID {
		t.Fatalf("expected run id %s, got %s", run.ID, artifact.RunID)
	}

	artifacts, err := manager.ListArtifacts(context.Background(), run.ID)
	if err != nil {
		t.Fatalf("ListArtifacts() error = %v", err)
	}
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
}

func TestModelProfileValidationAndRegistration(t *testing.T) {
	repo := newRepositoryStub()
	manager := NewManager(nil, transactorStub{}, repo, &auditStub{}, &schedulerStub{})

	_, err := manager.RegisterModelProfile(context.Background(), RegisterModelProfileInput{Name: ""}, "owner")
	if err == nil {
		t.Fatal("expected model profile validation error")
	}

	registered, err := manager.RegisterModelProfile(context.Background(), RegisterModelProfileInput{
		Name:               "ach-enterprise",
		Provider:           "local_ollama",
		BaseURL:            "http://ai-precision:11434",
		PrimaryModel:       "ibm/granite3.3:8b",
		GuardrailModel:     "ibm/granite3.3-guardian:8b",
		HelperModel:        "granite4:3b",
		TimeoutSeconds:     60,
		Temperature:        0.1,
		MaxTokens:          2048,
		ConnectionMetadata: json.RawMessage(`{"network":"tailscale"}`),
	}, "owner")
	if err != nil {
		t.Fatalf("RegisterModelProfile() error = %v", err)
	}

	loaded, err := manager.GetModelProfile(context.Background(), registered.Name)
	if err != nil {
		t.Fatalf("GetModelProfile() error = %v", err)
	}
	if loaded.PrimaryModel != "ibm/granite3.3:8b" {
		t.Fatalf("unexpected model profile %#v", loaded)
	}
}

func TestComparisonUpsertAndFetch(t *testing.T) {
	repo := newRepositoryStub()
	manager := NewManager(nil, transactorStub{}, repo, &auditStub{}, &schedulerStub{})

	run, err := manager.Create(context.Background(), CreateInput{Name: "compare-run"}, "owner")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	comparison, err := manager.UpsertComparison(context.Background(), run.ID, UpsertComparisonInput{
		DeterministicSummary: json.RawMessage(`{"precision":0.91}`),
		LLMSummary:           json.RawMessage(`{"recommendation":"tighten descriptor monitoring"}`),
		GuardrailSummary:     json.RawMessage(`{"decision":"pass"}`),
		Deltas:               json.RawMessage(`{"alert_growth":"+3%"}`),
		ReviewStatus:         string(ReviewStatusReviewPending),
		ReviewerNotes:        "needs compliance review",
		FinalDisposition:     "pending",
	}, "reviewer")
	if err != nil {
		t.Fatalf("UpsertComparison() error = %v", err)
	}

	if comparison.RunID != run.ID {
		t.Fatalf("expected run id %s, got %s", run.ID, comparison.RunID)
	}

	loaded, err := manager.GetComparison(context.Background(), run.ID)
	if err != nil {
		t.Fatalf("GetComparison() error = %v", err)
	}

	if string(loaded.GuardrailSummary) != `{"decision":"pass"}` {
		t.Fatalf("unexpected comparison payload %#v", loaded)
	}
}

func TestStartRunLLMUsesInferenceAndMemoryFetch(t *testing.T) {
	repo := newRepositoryStub()
	_, _ = repo.RegisterModelProfile(context.Background(), nil, RegisterModelProfileInput{
		Name:               "ach-default",
		Provider:           "local_ollama",
		BaseURL:            "http://ai-precision:11434",
		PrimaryModel:       "ibm/granite3.3:8b",
		GuardrailModel:     "ibm/granite3.3-guardian:8b",
		HelperModel:        "granite4:3b",
		TimeoutSeconds:     45,
		Temperature:        0.1,
		MaxTokens:          4096,
		ConnectionMetadata: json.RawMessage(`{}`),
	}, "system")

	memory := &memoryStub{
		snapshotRef: "snapshot-agent-1",
		context: MemoryContext{
			PriorCycleSummaries: []string{"previous-summary"},
			UnresolvedGaps:      []string{"gap-1"},
			ReviewerNotes:       []string{"note-1"},
		},
	}
	inference := &inferenceStub{
		response: InferenceResponse{
			PrimaryOutput: json.RawMessage(`{"summary":"agent analysis"}`),
			Guardrail:     json.RawMessage(`{"decision":"pass"}`),
		},
	}
	manager := NewManagerWithIntegrations(nil, transactorStub{}, repo, &auditStub{}, &schedulerStub{}, memory, inference, DependencyConfig{})

	run, err := manager.Create(context.Background(), CreateInput{
		Name:          "agent-run",
		RunType:       string(RunTypeAgentRun),
		ExecutionMode: string(ExecutionModeLLM),
		Repo:          "ach-trust-lab",
		Domain:        "ach",
		ModelProfile:  "ach-default",
	}, "owner")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	result, err := manager.StartRun(context.Background(), run.ID, ExecuteRunInput{
		Prompt:    "analyze run",
		InputJSON: json.RawMessage(`{"kind":"agent"}`),
	}, "owner")
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}

	if inference.calls != 1 {
		t.Fatalf("expected one inference call, got %d", inference.calls)
	}
	if inference.lastInput.Provider != "local_ollama" {
		t.Fatalf("expected local_ollama provider, got %q", inference.lastInput.Provider)
	}
	if inference.lastInput.BaseURL != "http://ai-precision:11434" {
		t.Fatalf("expected model profile base_url to be forwarded, got %q", inference.lastInput.BaseURL)
	}
	if inference.lastInput.HelperRequested {
		t.Fatalf("expected helper_requested=false, got true")
	}
	if memory.fetched != 1 {
		t.Fatalf("expected one memory fetch call, got %d", memory.fetched)
	}
	if result.Status != string(RunStatusReviewPending) {
		t.Fatalf("expected review_pending, got %s", result.Status)
	}
	if len(result.Artifacts) == 0 {
		t.Fatalf("expected llm artifacts to be persisted")
	}
}

func TestExecuteCycleRunDualPersistsComparison(t *testing.T) {
	repo := newRepositoryStub()
	_, _ = repo.RegisterModelProfile(context.Background(), nil, RegisterModelProfileInput{
		Name:               "ach-default",
		Provider:           "local_ollama",
		BaseURL:            "http://ai-precision:11434",
		PrimaryModel:       "ibm/granite3.3:8b",
		GuardrailModel:     "ibm/granite3.3-guardian:8b",
		HelperModel:        "granite4:3b",
		TimeoutSeconds:     45,
		Temperature:        0.1,
		MaxTokens:          4096,
		ConnectionMetadata: json.RawMessage(`{}`),
	}, "system")

	memory := &memoryStub{snapshotRef: "snapshot-cycle-2"}
	inference := &inferenceStub{
		response: InferenceResponse{
			PrimaryOutput: json.RawMessage(`{"summary":"cycle llm"}`),
			Guardrail:     json.RawMessage(`{"decision":"review"}`),
		},
	}
	manager := NewManagerWithIntegrations(nil, transactorStub{}, repo, &auditStub{}, &schedulerStub{}, memory, inference, DependencyConfig{})

	run, err := manager.Create(context.Background(), CreateInput{
		Name:          "week-run",
		RunType:       string(RunTypeWeekRun),
		ExecutionMode: string(ExecutionModeDual),
		Repo:          "ach-trust-lab",
		Domain:        "ach",
		ModelProfile:  "ach-default",
	}, "owner")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	cycle, err := manager.CreateCycle(context.Background(), run.ID, CreateCycleInput{
		CycleKey: "day-2",
		Focus:    "receiver monitoring",
		Status:   string(CycleStatusPending),
	}, "owner")
	if err != nil {
		t.Fatalf("CreateCycle() error = %v", err)
	}

	result, err := manager.ExecuteCycleRun(context.Background(), run.ID, cycle.ID, ExecuteRunInput{
		AgentNamespace: stringPtr("daily-summary"),
		Prompt:         "analyze cycle",
	}, "owner")
	if err != nil {
		t.Fatalf("ExecuteCycleRun() error = %v", err)
	}

	if result.Comparison == nil {
		t.Fatal("expected comparison to be persisted in dual mode")
	}
	if string(result.DeterministicSummary) == "{}" || string(result.LLMSummary) == "{}" {
		t.Fatalf("expected deterministic and llm summaries in dual mode, got %#v", result)
	}
	if result.MemorySnapshotRef == "" {
		t.Fatalf("expected memory snapshot ref")
	}
}

func TestBuildInferencePayloadCompactDual(t *testing.T) {
	run := Run{
		ID:                "run-123",
		RunType:           string(RunTypeWeekRun),
		ExecutionMode:     string(ExecutionModeDual),
		DatasetRefs:       []string{"dataset-a.json"},
		PromptPackVersion: "prompt-v2",
		RulePackVersion:   "rules-v9",
	}
	cycle := &Cycle{
		CycleKey:     "day-3",
		Focus:        "descriptor drift",
		Objective:    "tighten false positive controls",
		DetectorPack: "detectors/ach-v3",
	}
	memory := MemoryContext{
		PriorCycleSummaries: []string{"summary-a", "summary-b"},
		CarryForwardRisks:   []string{"risk-a"},
		UnresolvedGaps:      []string{"gap-a"},
		ReviewerNotes:       []string{"note-a"},
	}
	deterministic := json.RawMessage(`{"run_id":"run-123","dataset_ref_count":1,"input":{"large":"payload"}}`)

	payload := buildInferencePayload(run, cycle, memory, deterministic, ExecuteRunInput{}, true)

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if _, ok := decoded["memory_context"]; ok {
		t.Fatal("expected compact dual payload to omit full memory_context")
	}
	if _, ok := decoded["deterministic_summary"]; ok {
		t.Fatal("expected compact dual payload to omit full deterministic_summary")
	}
	if decoded["cycle_key"] != "day-3" || decoded["focus"] != "descriptor drift" {
		t.Fatalf("expected cycle metadata in compact payload, got %#v", decoded)
	}
	if _, ok := decoded["carry_forward_memory_compact"]; !ok {
		t.Fatalf("expected carry_forward_memory_compact in payload %#v", decoded)
	}
	if _, ok := decoded["deterministic_summary_compact"]; !ok {
		t.Fatalf("expected deterministic_summary_compact in payload %#v", decoded)
	}
}

func TestDualExecutionUsesCompactPayloadAndTimeoutWiring(t *testing.T) {
	repo := newRepositoryStub()
	_, _ = repo.RegisterModelProfile(context.Background(), nil, RegisterModelProfileInput{
		Name:               "ach-default",
		Provider:           "local_ollama",
		BaseURL:            "http://ai-precision:11434",
		PrimaryModel:       "ibm/granite3.3:8b",
		GuardrailModel:     "ibm/granite3.3-guardian:8b",
		HelperModel:        "granite4:3b",
		TimeoutSeconds:     40,
		Temperature:        0.1,
		MaxTokens:          4096,
		ConnectionMetadata: json.RawMessage(`{"guardrail_timeout_seconds":55}`),
	}, "system")

	manager := NewManagerWithIntegrations(
		nil,
		transactorStub{},
		repo,
		&auditStub{},
		&schedulerStub{},
		&memoryStub{},
		&inferenceStub{response: InferenceResponse{
			PrimaryOutput: json.RawMessage(`{"summary":"ok"}`),
			Guardrail:     json.RawMessage(`{"decision":"pass"}`),
		}},
		DependencyConfig{
			EnableCompactDualPayload: true,
			HelperTimeout:            25 * time.Second,
		},
	)

	run, err := manager.Create(context.Background(), CreateInput{
		Name:          "week-run",
		RunType:       string(RunTypeWeekRun),
		ExecutionMode: string(ExecutionModeDual),
		Repo:          "ach-trust-lab",
		Domain:        "ach",
		ModelProfile:  "ach-default",
	}, "owner")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	cycle, err := manager.CreateCycle(context.Background(), run.ID, CreateCycleInput{CycleKey: "day-1", Status: string(CycleStatusPending)}, "owner")
	if err != nil {
		t.Fatalf("CreateCycle() error = %v", err)
	}

	stub := manager.inference.(*inferenceStub)
	_, err = manager.ExecuteCycleRun(context.Background(), run.ID, cycle.ID, ExecuteRunInput{}, "owner")
	if err != nil {
		t.Fatalf("ExecuteCycleRun() error = %v", err)
	}

	if stub.lastInput.PrimaryTimeoutSeconds != 40 {
		t.Fatalf("expected primary timeout wiring, got %d", stub.lastInput.PrimaryTimeoutSeconds)
	}
	if stub.lastInput.GuardrailTimeoutSeconds != 55 {
		t.Fatalf("expected guardrail timeout override, got %d", stub.lastInput.GuardrailTimeoutSeconds)
	}
	if stub.lastInput.HelperTimeoutSeconds != 25 {
		t.Fatalf("expected helper timeout from deps, got %d", stub.lastInput.HelperTimeoutSeconds)
	}

	var payload map[string]any
	if err := json.Unmarshal(stub.lastInput.InputJSON, &payload); err != nil {
		t.Fatalf("unmarshal compact input payload: %v", err)
	}
	if _, ok := payload["carry_forward_memory_compact"]; !ok {
		t.Fatalf("expected compact dual payload, got %#v", payload)
	}
	if _, ok := payload["memory_context"]; ok {
		t.Fatalf("expected compact dual payload to omit memory_context, got %#v", payload)
	}
}

func TestGuardrailInvocationPathSelection(t *testing.T) {
	repo := newRepositoryStub()
	_, _ = repo.RegisterModelProfile(context.Background(), nil, RegisterModelProfileInput{
		Name:               "ach-default",
		Provider:           "local_ollama",
		BaseURL:            "http://ai-precision:11434",
		PrimaryModel:       "ibm/granite3.3:8b",
		GuardrailModel:     "ibm/granite3.3-guardian:8b",
		HelperModel:        "granite4:3b",
		TimeoutSeconds:     45,
		Temperature:        0.1,
		MaxTokens:          2048,
		ConnectionMetadata: json.RawMessage(`{}`),
	}, "system")

	inference := &inferenceStub{response: InferenceResponse{PrimaryOutput: json.RawMessage(`{"summary":"ok"}`)}}
	manager := NewManagerWithIntegrations(
		nil,
		transactorStub{},
		repo,
		&auditStub{},
		&schedulerStub{},
		&memoryStub{},
		inference,
		DependencyConfig{DisableLocalOllamaGuardrails: true},
	)

	run, err := manager.Create(context.Background(), CreateInput{
		Name:          "agent-run",
		RunType:       string(RunTypeAgentRun),
		ExecutionMode: string(ExecutionModeLLM),
		Repo:          "ach-trust-lab",
		Domain:        "ach",
		ModelProfile:  "ach-default",
	}, "owner")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := manager.StartRun(context.Background(), run.ID, ExecuteRunInput{}, "owner"); err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}

	if inference.lastInput.EnableGuardrails {
		t.Fatalf("expected guardrails to be disabled for local_ollama when flag is enabled")
	}
}

func stringPtr(value string) *string {
	return &value
}
