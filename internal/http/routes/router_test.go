package routes

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"clawbot-server/internal/platform/bots"
	"clawbot-server/internal/platform/ops"
	"clawbot-server/internal/platform/policies"
	"clawbot-server/internal/platform/runs"
	"clawbot-server/internal/platform/store"
	"clawbot-server/internal/version"
)

type runsStub struct{}

func (runsStub) List(context.Context) ([]runs.Run, error) { return nil, nil }
func (runsStub) Get(context.Context, string) (runs.Run, error) {
	return runs.Run{}, nil
}
func (runsStub) Create(context.Context, runs.CreateInput, string) (runs.Run, error) {
	return runs.Run{}, nil
}
func (runsStub) Update(context.Context, string, runs.UpdateInput, string) (runs.Run, error) {
	return runs.Run{}, nil
}
func (runsStub) StartRun(context.Context, string, runs.ExecuteRunInput, string) (runs.ExecuteRunResult, error) {
	return runs.ExecuteRunResult{}, nil
}
func (runsStub) ExecuteCycleRun(context.Context, string, string, runs.ExecuteRunInput, string) (runs.ExecuteRunResult, error) {
	return runs.ExecuteRunResult{}, nil
}
func (runsStub) CreateCycle(context.Context, string, runs.CreateCycleInput, string) (runs.Cycle, error) {
	return runs.Cycle{}, nil
}
func (runsStub) GetCycle(context.Context, string, string) (runs.Cycle, error) {
	return runs.Cycle{}, nil
}
func (runsStub) UpdateCycle(context.Context, string, string, runs.UpdateCycleInput, string) (runs.Cycle, error) {
	return runs.Cycle{}, nil
}
func (runsStub) AttachArtifact(context.Context, string, runs.AttachArtifactInput, string) (runs.Artifact, error) {
	return runs.Artifact{}, nil
}
func (runsStub) ListArtifacts(context.Context, string) ([]runs.Artifact, error) {
	return nil, nil
}
func (runsStub) UpsertComparison(context.Context, string, runs.UpsertComparisonInput, string) (runs.Comparison, error) {
	return runs.Comparison{}, nil
}
func (runsStub) GetComparison(context.Context, string) (runs.Comparison, error) {
	return runs.Comparison{}, nil
}
func (runsStub) RegisterModelProfile(context.Context, runs.RegisterModelProfileInput, string) (runs.ModelProfile, error) {
	return runs.ModelProfile{}, nil
}
func (runsStub) GetModelProfile(context.Context, string) (runs.ModelProfile, error) {
	return runs.ModelProfile{}, nil
}
func (runsStub) DependencyHealth(context.Context) (runs.DependencyHealth, error) {
	return runs.DependencyHealth{Status: "healthy"}, nil
}

type botsStub struct{}

func (botsStub) List(context.Context) ([]bots.Bot, error) { return nil, nil }
func (botsStub) Get(context.Context, string) (bots.Bot, error) {
	return bots.Bot{}, nil
}
func (botsStub) Create(context.Context, bots.CreateInput, string) (bots.Bot, error) {
	return bots.Bot{}, nil
}
func (botsStub) Update(context.Context, string, bots.UpdateInput, string) (bots.Bot, error) {
	return bots.Bot{}, nil
}

type policiesStub struct{}

func (policiesStub) List(context.Context) ([]policies.Policy, error) { return nil, nil }
func (policiesStub) Get(context.Context, string) (policies.Policy, error) {
	return policies.Policy{}, nil
}
func (policiesStub) Create(context.Context, policies.CreateInput, string) (policies.Policy, error) {
	return policies.Policy{}, nil
}
func (policiesStub) Update(context.Context, string, policies.UpdateInput, string) (policies.Policy, error) {
	return policies.Policy{}, nil
}

type dashboardStub struct{}

func (dashboardStub) Summary(context.Context) (store.DashboardSummary, error) {
	return store.DashboardSummary{Runs: 2, Bots: 3, Policies: 4, AuditEvents: 5}, nil
}

type opsStub struct{}

func (opsStub) Overview(context.Context) (ops.Overview, error) {
	return ops.Overview{Status: ops.StatusHealthy, ServicesTotal: 3}, nil
}
func (opsStub) ListServices(context.Context) ([]ops.ServiceStatus, error) {
	return []ops.ServiceStatus{{ID: "svc-1", Name: "clawbot-server"}}, nil
}
func (opsStub) GetService(context.Context, string) (ops.ServiceStatus, error) {
	return ops.ServiceStatus{ID: "svc-1", Name: "clawbot-server"}, nil
}
func (opsStub) SetMaintenance(context.Context, string, string) (ops.ServiceStatus, error) {
	return ops.ServiceStatus{ID: "svc-1", Name: "clawbot-server", MaintenanceMode: true}, nil
}
func (opsStub) ResumeService(context.Context, string, string) (ops.ServiceStatus, error) {
	return ops.ServiceStatus{ID: "svc-1", Name: "clawbot-server"}, nil
}
func (opsStub) ListSchedulers(context.Context) ([]ops.SchedulerStatus, error) {
	return []ops.SchedulerStatus{{ID: "sch-1", Name: "sync"}}, nil
}
func (opsStub) GetScheduler(context.Context, string) (ops.SchedulerStatus, error) {
	return ops.SchedulerStatus{ID: "sch-1", Name: "sync"}, nil
}
func (opsStub) PauseScheduler(context.Context, string, string) (ops.SchedulerStatus, error) {
	return ops.SchedulerStatus{ID: "sch-1", Name: "sync"}, nil
}
func (opsStub) ResumeScheduler(context.Context, string, string) (ops.SchedulerStatus, error) {
	return ops.SchedulerStatus{ID: "sch-1", Name: "sync", Enabled: true}, nil
}
func (opsStub) RunSchedulerOnce(context.Context, string, string) (ops.SchedulerStatus, error) {
	return ops.SchedulerStatus{ID: "sch-1", Name: "sync", LastResult: ops.ResultManualTriggered}, nil
}
func (opsStub) ListEvents(context.Context) ([]ops.ActivityEvent, error) {
	return []ops.ActivityEvent{{ID: "evt-1", Source: "svc-1"}}, nil
}

func TestNewRoutesSystemAndDashboardEndpoints(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := New(logger, Services{
		System:    NewSystemHandler(func(context.Context) error { return nil }),
		Runs:      runsStub{},
		Bots:      botsStub{},
		Policies:  policiesStub{},
		Dashboard: dashboardStub{},
		Ops:       opsStub{},
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected health 200, got %d", resp.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/version", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected version 200, got %d", resp.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/summary", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected dashboard 200, got %d", resp.Code)
	}

	var payload struct {
		Data store.DashboardSummary `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if payload.Data.AuditEvents != 5 {
		t.Fatalf("unexpected dashboard payload %#v", payload)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/ops/overview", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected ops overview 200, got %d", resp.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/ops", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected ops page 200, got %d", resp.Code)
	}
}

func TestNewSystemHandlerUsesCurrentVersion(t *testing.T) {
	originalValue := version.Value
	originalCommit := version.Commit
	originalBuildDate := version.BuildDate
	t.Cleanup(func() {
		version.Value = originalValue
		version.Commit = originalCommit
		version.BuildDate = originalBuildDate
	})

	version.Value = "9.9.9"
	version.Commit = "deadbeef"
	version.BuildDate = "2026-03-25"

	handler := NewSystemHandler(func(context.Context) error { return nil })
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	resp := httptest.NewRecorder()
	handler.Version(resp, req)

	if !strings.Contains(resp.Body.String(), "9.9.9") {
		t.Fatalf("expected current version in body, got %s", resp.Body.String())
	}
}
