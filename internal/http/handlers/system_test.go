package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"clawbot-server/internal/version"
)

func TestSystemHandlerHealthReadyVersion(t *testing.T) {
	handler := NewSystemHandler(func(context.Context) error { return nil }, version.Info{
		Version:   "1.2.3",
		Commit:    "abc123",
		BuildDate: "2026-03-25",
	})

	healthReq := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	healthResp := httptest.NewRecorder()
	handler.Health(healthResp, healthReq)
	if healthResp.Code != http.StatusOK {
		t.Fatalf("expected health 200, got %d", healthResp.Code)
	}

	readyReq := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	readyResp := httptest.NewRecorder()
	handler.Ready(readyResp, readyReq)
	if readyResp.Code != http.StatusOK {
		t.Fatalf("expected ready 200, got %d", readyResp.Code)
	}

	versionReq := httptest.NewRequest(http.MethodGet, "/version", nil)
	versionResp := httptest.NewRecorder()
	handler.Version(versionResp, versionReq)
	if versionResp.Code != http.StatusOK {
		t.Fatalf("expected version 200, got %d", versionResp.Code)
	}
	if body := versionResp.Body.String(); body == "" || !containsAll(body, "1.2.3", "abc123", "2026-03-25") {
		t.Fatalf("unexpected version body %s", body)
	}
}

func TestSystemHandlerReadyNotReady(t *testing.T) {
	handler := NewSystemHandler(func(context.Context) error { return errors.New("db unavailable") }, version.Info{})

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	resp := httptest.NewRecorder()
	handler.Ready(resp, req)

	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", resp.Code)
	}
}

func containsAll(body string, values ...string) bool {
	for _, value := range values {
		if !strings.Contains(body, value) {
			return false
		}
	}
	return true
}
