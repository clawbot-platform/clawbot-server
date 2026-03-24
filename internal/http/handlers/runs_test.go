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
	listResult []runs.Run
	getResult  runs.Run
	createFn   func(runs.CreateInput, string) (runs.Run, error)
	updateFn   func(string, runs.UpdateInput, string) (runs.Run, error)
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

func TestRunsHandlerCreate(t *testing.T) {
	handler := NewRunsHandler(&runsServiceStub{
		createFn: func(input runs.CreateInput, actor string) (runs.Run, error) {
			return runs.Run{ID: "run-1", Name: input.Name, Status: input.Status, CreatedBy: actor}, nil
		},
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

func routeContext(key string, value string) *chi.Context {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return rctx
}
