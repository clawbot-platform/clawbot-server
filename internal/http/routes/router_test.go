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

func TestNewRoutesSystemAndDashboardEndpoints(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := New(logger, Services{
		System:    NewSystemHandler(func(context.Context) error { return nil }),
		Runs:      runsStub{},
		Bots:      botsStub{},
		Policies:  policiesStub{},
		Dashboard: dashboardStub{},
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
