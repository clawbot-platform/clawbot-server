package identityclient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCompareUsesTenantHeadersAndDecodesResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/v1/compare" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("X-Correlation-ID") != "corr-123" {
			t.Fatalf("unexpected X-Correlation-ID %q", r.Header.Get("X-Correlation-ID"))
		}
		if r.Header.Get("X-Case-ID") != "case-123" {
			t.Fatalf("unexpected X-Case-ID %q", r.Header.Get("X-Case-ID"))
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if !strings.Contains(string(body), `"tenant_id":"tenant-default"`) {
			t.Fatalf("expected default tenant in body, got %s", string(body))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"disposition":"resolved","confidence_band":"high","explanation":{"explanation_id":"exp_1","summary":"ok","why":["strong signal"],"why_not":[],"how":["linked records"],"source_refs":[{"source_system":"kyc_applications","source_record_id":"left-record"},{"source_system":"watchlist_candidates","source_record_id":"right-record"}]},"decision_trace_id":"dt_1"}`))
	}))
	defer server.Close()

	client := New(server.URL, time.Second, "tenant-default")
	result, err := client.Compare(context.Background(), CompareRequest{
		Left:    RecordRef{SourceSystem: "kyc_applications", SourceRecordID: "left-record"},
		Right:   RecordRef{SourceSystem: "watchlist_candidates", SourceRecordID: "right-record"},
		Explain: true,
	}, "corr-123", "case-123")
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}

	if result.DecisionTraceID != "dt_1" {
		t.Fatalf("unexpected decision trace id %q", result.DecisionTraceID)
	}
	if result.ConfidenceBand != "high" {
		t.Fatalf("unexpected confidence band %q", result.ConfidenceBand)
	}
}

func TestScreenOFACUsesCaseHeaderAndDecodesResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/v1/watchlist/ofac/screenings" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("X-Correlation-ID") != "corr-999" {
			t.Fatalf("unexpected X-Correlation-ID %q", r.Header.Get("X-Correlation-ID"))
		}
		if r.Header.Get("X-Case-ID") != "case-xyz" {
			t.Fatalf("unexpected X-Case-ID %q", r.Header.Get("X-Case-ID"))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"screening_id":"scr_1","decision":"manual_review","decision_trace_id":"dt_2","candidates":[{"dataset_run_id":"ofac-run","list_kind":"sdn","list_uid":"uid-1","name":"Jane Citizen","matched_on":"name+country+identifier","score":95,"needs_review":true}]}`))
	}))
	defer server.Close()

	client := New(server.URL, time.Second, "tenant-default")
	result, err := client.ScreenOFAC(context.Background(), ScreenOFACRequest{
		CaseID: "case-xyz",
		Subject: OFACSubject{
			Name:    "Jane Citizen",
			Country: "US",
		},
	}, "corr-999")
	if err != nil {
		t.Fatalf("ScreenOFAC() error = %v", err)
	}

	if result.ScreeningID != "scr_1" {
		t.Fatalf("unexpected screening id %q", result.ScreeningID)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("expected one candidate, got %d", len(result.Candidates))
	}
}

func TestCompareReturnsNon2xxErrorBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "identity unavailable", http.StatusBadGateway)
	}))
	defer server.Close()

	client := New(server.URL, time.Second, "tenant-default")
	_, err := client.Compare(context.Background(), CompareRequest{
		Left:    RecordRef{SourceSystem: "left", SourceRecordID: "1"},
		Right:   RecordRef{SourceSystem: "right", SourceRecordID: "2"},
		Explain: true,
	}, "corr-1", "case-1")
	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
	if !strings.Contains(err.Error(), "status 502") {
		t.Fatalf("expected status in error, got %v", err)
	}
	if !strings.Contains(err.Error(), "identity unavailable") {
		t.Fatalf("expected response body in error, got %v", err)
	}
}

func TestHealthz(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL, time.Second, "tenant-default")
	if err := client.Healthz(context.Background()); err != nil {
		t.Fatalf("Healthz() error = %v", err)
	}
}
