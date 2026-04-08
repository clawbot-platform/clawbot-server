package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"clawbot-server/internal/platform/runs"
	"clawbot-server/internal/platform/store"

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

func newRunsServiceStub() *runsServiceStub {
	return &runsServiceStub{
		listResult: []runs.Run{{ID: "run-1", Name: "Run One"}},
		getResult:  runs.Run{ID: "run-1", Name: "Run One", Status: "pending"},
		createFn:   emptyRunCreate,
		updateFn:   emptyRunUpdate,
		reviewFn: func(runID string, input runs.ReviewActionInput, _ string) (runs.Run, error) {
			return runs.Run{ID: runID, Status: input.Action, Notes: input.Rationale}, nil
		},
		startRunFn: func(runID string, input runs.ExecuteRunInput, _ string) (runs.ExecuteRunResult, error) {
			return runs.ExecuteRunResult{RunID: runID, Status: "review_pending", LLMSummary: input.InputJSON}, nil
		},
		execCycleFn: func(runID string, cycleID string, _ runs.ExecuteRunInput, _ string) (runs.ExecuteRunResult, error) {
			return runs.ExecuteRunResult{RunID: runID, CycleID: &cycleID, Status: "review_pending"}, nil
		},
		attachArtifactFn: func(runID string, input runs.AttachArtifactInput, _ string) (runs.Artifact, error) {
			return runs.Artifact{ID: "artifact-1", RunID: runID, ArtifactType: input.ArtifactType, URI: input.URI}, nil
		},
		listArtifactsFn: func(runID string) ([]runs.Artifact, error) {
			return []runs.Artifact{{ID: "artifact-1", RunID: runID}}, nil
		},
		createCycleFn: func(runID string, input runs.CreateCycleInput, _ string) (runs.Cycle, error) {
			if runID == "" {
				return runs.Cycle{}, errors.New("run id is required")
			}
			return runs.Cycle{ID: "cycle-1", RunID: runID, CycleKey: input.CycleKey, Status: input.Status}, nil
		},
		updateCycleFn: func(runID string, cycleID string, input runs.UpdateCycleInput, _ string) (runs.Cycle, error) {
			return runs.Cycle{ID: cycleID, RunID: runID, Status: derefString(input.Status, "pending")}, nil
		},
		getCycleFn: func(runID string, cycleID string) (runs.Cycle, error) {
			return runs.Cycle{ID: cycleID, RunID: runID, CycleKey: "day-1", Status: "pending"}, nil
		},
		upsertCompareFn: func(runID string, input runs.UpsertComparisonInput, _ string) (runs.Comparison, error) {
			return runs.Comparison{ID: "cmp-1", RunID: runID, ReviewStatus: input.ReviewStatus}, nil
		},
		getCompareFn: func(runID string) (runs.Comparison, error) {
			return runs.Comparison{ID: "cmp-1", RunID: runID, ReviewStatus: string(runs.ReviewStatusReviewPending)}, nil
		},
		registerProfileFn: func(input runs.RegisterModelProfileInput, _ string) (runs.ModelProfile, error) {
			return runs.ModelProfile{ID: "profile-1", Name: input.Name, Provider: input.Provider}, nil
		},
		getProfileFn: func(idOrName string) (runs.ModelProfile, error) {
			return runs.ModelProfile{ID: "profile-1", Name: idOrName, Provider: "local_ollama"}, nil
		},
		dependencyFn: func() (runs.DependencyHealth, error) {
			return runs.DependencyHealth{Status: "healthy"}, nil
		},
	}
}

func derefString(value *string, fallback string) string {
	if value == nil {
		return fallback
	}
	return *value
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

func TestRunsHandlerListAndGet(t *testing.T) {
	handler := NewRunsHandler(newRunsServiceStub())

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/runs", nil)
	listRec := httptest.NewRecorder()
	handler.List(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for List, got %d", listRec.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/runs/run-1", nil)
	getReq = getReq.WithContext(context.WithValue(getReq.Context(), chi.RouteCtxKey, routeContext("runID", "run-1")))
	getRec := httptest.NewRecorder()
	handler.Get(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for Get, got %d", getRec.Code)
	}
}

func TestRunsHandlerCreateValidationFailure(t *testing.T) {
	handler := NewRunsHandler(newRunsServiceStub())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs", bytes.NewBufferString(`{"name":"bad","unknown":true}`))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	var payload map[string]map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if code, _ := payload["error"]["code"].(string); code != "invalid_json" {
		t.Fatalf("expected invalid_json code, got %#v", payload)
	}
}

func TestRunsHandlerCreateServiceErrorMapping(t *testing.T) {
	stub := newRunsServiceStub()
	stub.createFn = func(runs.CreateInput, string) (runs.Run, error) { return runs.Run{}, store.ErrNotFound }
	handler := NewRunsHandler(stub)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs", bytes.NewBufferString(`{"name":"missing","status":"pending"}`))
	rec := httptest.NewRecorder()
	handler.Create(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestRunsHandlerCycleEndpoints(t *testing.T) {
	handler := NewRunsHandler(newRunsServiceStub())

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/runs/run-1/cycles", bytes.NewBufferString(`{"cycle_key":"day-2","status":"pending"}`))
	createReq = createReq.WithContext(context.WithValue(createReq.Context(), chi.RouteCtxKey, routeContext("runID", "run-1")))
	createRec := httptest.NewRecorder()
	handler.CreateCycle(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for CreateCycle, got %d", createRec.Code)
	}

	updateReq := httptest.NewRequest(http.MethodPatch, "/api/v1/runs/run-1/cycles/cycle-1", bytes.NewBufferString(`{"status":"running"}`))
	updateReq = updateReq.WithContext(context.WithValue(updateReq.Context(), chi.RouteCtxKey, routeContextMany(map[string]string{"runID": "run-1", "cycleID": "cycle-1"})))
	updateRec := httptest.NewRecorder()
	handler.UpdateCycle(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for UpdateCycle, got %d", updateRec.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/runs/run-1/cycles/cycle-1", nil)
	getReq = getReq.WithContext(context.WithValue(getReq.Context(), chi.RouteCtxKey, routeContextMany(map[string]string{"runID": "run-1", "cycleID": "cycle-1"})))
	getRec := httptest.NewRecorder()
	handler.GetCycle(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GetCycle, got %d", getRec.Code)
	}
}

func TestRunsHandlerCreateCycleMissingRunID(t *testing.T) {
	handler := NewRunsHandler(newRunsServiceStub())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs//cycles", bytes.NewBufferString(`{"cycle_key":"day-2","status":"pending"}`))
	rec := httptest.NewRecorder()
	handler.CreateCycle(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing runID, got %d", rec.Code)
	}
}

func TestRunsHandlerStartRunInvalidJSONAndSuccess(t *testing.T) {
	handler := NewRunsHandler(newRunsServiceStub())

	badReq := httptest.NewRequest(http.MethodPost, "/api/v1/runs/run-1/start", bytes.NewBufferString(`{"prompt":`))
	badReq = badReq.WithContext(context.WithValue(badReq.Context(), chi.RouteCtxKey, routeContext("runID", "run-1")))
	badRec := httptest.NewRecorder()
	handler.StartRun(badRec, badReq)
	if badRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid StartRun JSON, got %d", badRec.Code)
	}

	okReq := httptest.NewRequest(http.MethodPost, "/api/v1/runs/run-1/start", bytes.NewBufferString(`{"prompt":"go","input_json":{"x":1}}`))
	okReq = okReq.WithContext(context.WithValue(okReq.Context(), chi.RouteCtxKey, routeContext("runID", "run-1")))
	okRec := httptest.NewRecorder()
	handler.StartRun(okRec, okReq)
	if okRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for StartRun, got %d", okRec.Code)
	}
}

func TestRunsHandlerArtifactAndComparisonEndpoints(t *testing.T) {
	handler := NewRunsHandler(newRunsServiceStub())

	listArtifactsReq := httptest.NewRequest(http.MethodGet, "/api/v1/runs/run-1/artifacts", nil)
	listArtifactsReq = listArtifactsReq.WithContext(context.WithValue(listArtifactsReq.Context(), chi.RouteCtxKey, routeContext("runID", "run-1")))
	listArtifactsRec := httptest.NewRecorder()
	handler.ListArtifacts(listArtifactsRec, listArtifactsReq)
	if listArtifactsRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for ListArtifacts, got %d", listArtifactsRec.Code)
	}

	upsertReq := httptest.NewRequest(http.MethodPut, "/api/v1/runs/run-1/comparison", bytes.NewBufferString(`{"review_status":"review_pending"}`))
	upsertReq = upsertReq.WithContext(context.WithValue(upsertReq.Context(), chi.RouteCtxKey, routeContext("runID", "run-1")))
	upsertRec := httptest.NewRecorder()
	handler.UpsertComparison(upsertRec, upsertReq)
	if upsertRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for UpsertComparison, got %d", upsertRec.Code)
	}

	getComparisonReq := httptest.NewRequest(http.MethodGet, "/api/v1/runs/run-1/comparison", nil)
	getComparisonReq = getComparisonReq.WithContext(context.WithValue(getComparisonReq.Context(), chi.RouteCtxKey, routeContext("runID", "run-1")))
	getComparisonRec := httptest.NewRecorder()
	handler.GetComparison(getComparisonRec, getComparisonReq)
	if getComparisonRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GetComparison, got %d", getComparisonRec.Code)
	}
}

func TestRunsHandlerModelProfileEndpoints(t *testing.T) {
	handler := NewRunsHandler(newRunsServiceStub())

	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/model-profiles", bytes.NewBufferString(`{"name":"ach-local","provider":"local_ollama"}`))
	registerRec := httptest.NewRecorder()
	handler.RegisterModelProfile(registerRec, registerReq)
	if registerRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for RegisterModelProfile, got %d", registerRec.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/model-profiles/ach-local", nil)
	getReq = getReq.WithContext(context.WithValue(getReq.Context(), chi.RouteCtxKey, routeContext("modelProfileID", "ach-local")))
	getRec := httptest.NewRecorder()
	handler.GetModelProfile(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GetModelProfile, got %d", getRec.Code)
	}
}

func TestRunsHandlerReviewEndpoints(t *testing.T) {
	handler := NewRunsHandler(newRunsServiceStub())

	cases := []struct {
		name   string
		target func(http.ResponseWriter, *http.Request)
	}{
		{name: "reject", target: handler.RejectRun},
		{name: "override", target: handler.OverrideRun},
		{name: "defer", target: handler.DeferRun},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/runs/run-1/"+tc.name, bytes.NewBufferString(`{"rationale":"operator action"}`))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeContext("runID", "run-1")))
			rec := httptest.NewRecorder()
			tc.target(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200 for %s, got %d", tc.name, rec.Code)
			}
		})
	}
}

func emptyRunCreate(runs.CreateInput, string) (runs.Run, error) {
	return runs.Run{ID: "run-1", Name: "Run", Status: "pending"}, nil
}

func emptyRunUpdate(string, runs.UpdateInput, string) (runs.Run, error) {
	return runs.Run{ID: "run-1", Name: "Run", Status: "pending"}, nil
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
