package middleware

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func TestCaptureRequestIDCopiesMiddlewareValue(t *testing.T) {
	handler := chimiddleware.RequestID(CaptureRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(requestIDFromContext(r.Context())))
	})))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
	if strings.TrimSpace(resp.Body.String()) == "" {
		t.Fatal("expected captured request id in response body")
	}
}

func TestRequestLoggerEmitsRequestMetadata(t *testing.T) {
	var buffer bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buffer, nil))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs", nil)
	req = req.WithContext(context.WithValue(req.Context(), requestIDKey{}, "req-123"))
	resp := httptest.NewRecorder()

	handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	handler.ServeHTTP(resp, req)

	body := buffer.String()
	if !containsAll(body, `"method":"POST"`, `"path":"/api/v1/runs"`, `"status":201`, `"request_id":"req-123"`) {
		t.Fatalf("unexpected log body %s", body)
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
