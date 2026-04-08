package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"clawbot-server/internal/platform/runs"

	"github.com/go-chi/chi/v5"
)

type runsServiceStub struct {
	listResult  []runs.Run
	getResult   runs.Run
	createFn    func(runs.CreateInput, string) (runs.Run, error)
	updateFn    func(string, runs.UpdateInput, string) (runs.Run, error)
	reviewFn    func(string, runs.ReviewActionInput, string) (runs.Run, error)
	startRunFn  func(string, runs.ExecuteRunInput, string) (runs.ExecuteRunResult, error)
	execCycleFn func(string, string, runs.ExecuteRunInput, string) (runs.ExecuteRunResult, error)

	attachArtifactFn  func(string, runs.AttachArtifactInput, string) (runs.Artifact, error)
	listArtifactsFn   func(string) ([]runs.Artifact, error)
	createCycleFn     func(string, runs.CreateCycleInput, string) (runs.Cycle, error)
	updateCycleFn     func(string, string, runs.UpdateCycleInput, string) (runs.Cycle, error)
	getCycleFn        func(string, string) (runs.Cycle, error)
	upsertCompareFn   func(string, runs.UpsertComparisonInput, string) (runs.Comparison, error)
	getCompareFn      func(string) (runs.Comparison, error)
	registerProfileFn func(runs.RegisterModelProfileInput, string) (runs.ModelProfile, error)
	getProfileFn      func(string) (runs.ModelProfile, error)
	dependencyFn      func() (runs.DependencyHealth, error)
}

func (s *runsServiceStub) List(context.Context) ([]runs.Run, error) { return s.listResult, nil }
func (s *runsServiceStub) Get(context.Context, string) (runs.Run, error) {
	return s.getResult, nil
}
func (s *runsServiceStub) Create(_ context.Context, input runs.CreateInput, actor string) (runs.Run, error) {
	return s.createFn(input, actor)
}
func (s *runsServiceStub) Update(_ context.Context, id string, input runs.UpdateInput, actor string) (runs.Run, error) {
	return s.updateFn(id, input, actor)
}
func (s *runsServiceStub) ReviewAction(_ context.Context, runID string, input runs.ReviewActionInput, actor string) (runs.Run, error) {
	return s.reviewFn(runID, input, actor)
}
func (s *runsServiceStub) StartRun(_ context.Context, runID string, input runs.ExecuteRunInput, actor string) (runs.ExecuteRunResult, error) {
	return s.startRunFn(runID, input, actor)
}
func (s *runsServiceStub) ExecuteCycleRun(_ context.Context, runID string, cycleID string, input runs.ExecuteRunInput, actor string) (runs.ExecuteRunResult, error) {
	return s.execCycleFn(runID, cycleID, input, actor)
}
func (s *runsServiceStub) CreateCycle(_ context.Context, runID string, input runs.CreateCycleInput, actor string) (runs.Cycle, error) {
	return s.createCycleFn(runID, input, actor)
}
func (s *runsServiceStub) GetCycle(_ context.Context, runID string, cycleID string) (runs.Cycle, error) {
	return s.getCycleFn(runID, cycleID)
}
func (s *runsServiceStub) UpdateCycle(_ context.Context, runID string, cycleID string, input runs.UpdateCycleInput, actor string) (runs.Cycle, error) {
	return s.updateCycleFn(runID, cycleID, input, actor)
}
func (s *runsServiceStub) AttachArtifact(_ context.Context, runID string, input runs.AttachArtifactInput, actor string) (runs.Artifact, error) {
	return s.attachArtifactFn(runID, input, actor)
}
func (s *runsServiceStub) ListArtifacts(_ context.Context, runID string) ([]runs.Artifact, error) {
	return s.listArtifactsFn(runID)
}
func (s *runsServiceStub) UpsertComparison(_ context.Context, runID string, input runs.UpsertComparisonInput, actor string) (runs.Comparison, error) {
	return s.upsertCompareFn(runID, input, actor)
}
func (s *runsServiceStub) GetComparison(_ context.Context, runID string) (runs.Comparison, error) {
	return s.getCompareFn(runID)
}
func (s *runsServiceStub) RegisterModelProfile(_ context.Context, input runs.RegisterModelProfileInput, actor string) (runs.ModelProfile, error) {
	return s.registerProfileFn(input, actor)
}
func (s *runsServiceStub) GetModelProfile(_ context.Context, idOrName string) (runs.ModelProfile, error) {
	return s.getProfileFn(idOrName)
}
func (s *runsServiceStub) DependencyHealth(context.Context) (runs.DependencyHealth, error) {
	return s.dependencyFn()
}

func TestRunsHandlerCreate(t *testing.T) {
	handler := NewRunsHandler(&runsServiceStub{
		createFn: func(input runs.CreateInput, actor string) (runs.Run, error) {
			return runs.Run{ID: "run-1", Name: input.Name, Status: input.Status, CreatedBy: actor}, nil
		},
		updateFn: emptyRunUpdate,
	})

	body := bytes.NewBufferString(`{"name":"Test run","status":"pending"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs", body)
	req.Header.Set("X-Actor", "tester")
	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}

	var response map[string]runs.Run
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if response["data"].CreatedBy != "tester" {
		t.Fatalf("unexpected actor propagation: %#v", response["data"])
	}
}

func TestRunsHandlerUpdate(t *testing.T) {
	handler := NewRunsHandler(&runsServiceStub{
		createFn: emptyRunCreate,
		updateFn: func(id string, input runs.UpdateInput, actor string) (runs.Run, error) {
			return runs.Run{ID: id, Name: "Updated", Status: "scheduled", CreatedBy: actor}, nil
		},
	})

	body := bytes.NewBufferString(`{"status":"scheduled"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/runs/run-1", body)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeContext("runID", "run-1")))
	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestRunsHandlerAttachArtifact(t *testing.T) {
	handler := NewRunsHandler(&runsServiceStub{
		createFn: emptyRunCreate,
		updateFn: emptyRunUpdate,
		attachArtifactFn: func(runID string, input runs.AttachArtifactInput, _ string) (runs.Artifact, error) {
			return runs.Artifact{ID: "art-1", RunID: runID, ArtifactType: input.ArtifactType, URI: input.URI}, nil
		},
	})

	body := bytes.NewBufferString(`{"artifact_type":"replay_output","uri":"s3://bundle.json"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs/run-1/artifacts", body)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeContext("runID", "run-1")))
	recorder := httptest.NewRecorder()

	handler.AttachArtifact(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}
}

func TestRunsHandlerStartRun(t *testing.T) {
	handler := NewRunsHandler(&runsServiceStub{
		createFn: emptyRunCreate,
		updateFn: emptyRunUpdate,
		startRunFn: func(runID string, input runs.ExecuteRunInput, _ string) (runs.ExecuteRunResult, error) {
			return runs.ExecuteRunResult{RunID: runID, ExecutionMode: "llm", Status: "review_pending", LLMSummary: input.InputJSON}, nil
		},
	})

	body := bytes.NewBufferString(`{"prompt":"analyze","input_json":{"kind":"agent"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs/run-1/start", body)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeContext("runID", "run-1")))
	recorder := httptest.NewRecorder()

	handler.StartRun(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestRunsHandlerExecuteCycleRun(t *testing.T) {
	handler := NewRunsHandler(&runsServiceStub{
		createFn: emptyRunCreate,
		updateFn: emptyRunUpdate,
		execCycleFn: func(runID string, cycleID string, _ runs.ExecuteRunInput, _ string) (runs.ExecuteRunResult, error) {
			return runs.ExecuteRunResult{RunID: runID, CycleID: &cycleID, ExecutionMode: "dual", Status: "review_pending"}, nil
		},
	})

	body := bytes.NewBufferString(`{"prompt":"cycle analysis"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs/run-1/cycles/cycle-1/execute", body)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeContextMany(map[string]string{"runID": "run-1", "cycleID": "cycle-1"})))
	recorder := httptest.NewRecorder()

	handler.ExecuteCycleRun(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestRunsHandlerDependencyHealth(t *testing.T) {
	handler := NewRunsHandler(&runsServiceStub{
		createFn: emptyRunCreate,
		updateFn: emptyRunUpdate,
		dependencyFn: func() (runs.DependencyHealth, error) {
			return runs.DependencyHealth{Status: "healthy"}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/control-plane/dependencies", nil)
	recorder := httptest.NewRecorder()

	handler.DependencyHealth(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestRunsHandlerApproveRun(t *testing.T) {
	handler := NewRunsHandler(&runsServiceStub{
		createFn: emptyRunCreate,
		updateFn: emptyRunUpdate,
		reviewFn: func(runID string, input runs.ReviewActionInput, _ string) (runs.Run, error) {
			return runs.Run{ID: runID, Status: "approved", Notes: input.Rationale}, nil
		},
	})

	body := bytes.NewBufferString(`{"rationale":"validated and approved"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs/run-1/approve", body)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeContext("runID", "run-1")))
	recorder := httptest.NewRecorder()

	handler.ApproveRun(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func emptyRunCreate(runs.CreateInput, string) (runs.Run, error) {
	return runs.Run{ID: "run-1", Name: "Run", Status: "pending"}, nil
}

func emptyRunUpdate(string, runs.UpdateInput, string) (runs.Run, error) {
	return runs.Run{ID: "run-1", Name: "Run", Status: "pending"}, nil
}

func (s *runsServiceStub) ensureDefaults() {
	if s.createFn == nil {
		s.createFn = emptyRunCreate
	}
	if s.updateFn == nil {
		s.updateFn = emptyRunUpdate
	}
	if s.startRunFn == nil {
		s.startRunFn = func(runID string, _ runs.ExecuteRunInput, _ string) (runs.ExecuteRunResult, error) {
			return runs.ExecuteRunResult{RunID: runID, Status: "completed"}, nil
		}
	}
	if s.reviewFn == nil {
		s.reviewFn = func(runID string, _ runs.ReviewActionInput, _ string) (runs.Run, error) {
			return runs.Run{ID: runID, Status: "approved"}, nil
		}
	}
	if s.execCycleFn == nil {
		s.execCycleFn = func(runID string, cycleID string, _ runs.ExecuteRunInput, _ string) (runs.ExecuteRunResult, error) {
			return runs.ExecuteRunResult{RunID: runID, CycleID: &cycleID, Status: "completed"}, nil
		}
	}
	if s.attachArtifactFn == nil {
		s.attachArtifactFn = func(runID string, _ runs.AttachArtifactInput, _ string) (runs.Artifact, error) {
			return runs.Artifact{ID: "art-1", RunID: runID}, nil
		}
	}
	if s.listArtifactsFn == nil {
		s.listArtifactsFn = func(string) ([]runs.Artifact, error) { return nil, nil }
	}
	if s.createCycleFn == nil {
		s.createCycleFn = func(string, runs.CreateCycleInput, string) (runs.Cycle, error) { return runs.Cycle{}, nil }
	}
	if s.updateCycleFn == nil {
		s.updateCycleFn = func(string, string, runs.UpdateCycleInput, string) (runs.Cycle, error) { return runs.Cycle{}, nil }
	}
	if s.getCycleFn == nil {
		s.getCycleFn = func(string, string) (runs.Cycle, error) { return runs.Cycle{}, nil }
	}
	if s.upsertCompareFn == nil {
		s.upsertCompareFn = func(string, runs.UpsertComparisonInput, string) (runs.Comparison, error) {
			return runs.Comparison{}, nil
		}
	}
	if s.getCompareFn == nil {
		s.getCompareFn = func(string) (runs.Comparison, error) { return runs.Comparison{}, nil }
	}
	if s.registerProfileFn == nil {
		s.registerProfileFn = func(runs.RegisterModelProfileInput, string) (runs.ModelProfile, error) {
			return runs.ModelProfile{}, nil
		}
	}
	if s.getProfileFn == nil {
		s.getProfileFn = func(string) (runs.ModelProfile, error) { return runs.ModelProfile{}, nil }
	}
	if s.dependencyFn == nil {
		s.dependencyFn = func() (runs.DependencyHealth, error) { return runs.DependencyHealth{Status: "healthy"}, nil }
	}
}

func routeContext(key string, value string) *chi.Context {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return rctx
}

func routeContextMany(values map[string]string) *chi.Context {
	rctx := chi.NewRouteContext()
	for key, value := range values {
		rctx.URLParams.Add(key, value)
	}
	return rctx
}
