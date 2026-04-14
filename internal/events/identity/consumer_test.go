package identity

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/nats-io/nats.go"
)

type fakeSubscriber struct {
	subjects map[string]nats.MsgHandler
}

func newFakeSubscriber() *fakeSubscriber {
	return &fakeSubscriber{
		subjects: make(map[string]nats.MsgHandler),
	}
}

func (f *fakeSubscriber) Subscribe(subject string, cb nats.MsgHandler) (*nats.Subscription, error) {
	f.subjects[subject] = cb
	return nil, nil
}

func TestNewConsumerSubscribesToIdentitySubjects(t *testing.T) {
	t.Parallel()

	sub := newFakeSubscriber()
	consumer, err := NewConsumer(sub, slog.Default(), Handlers{})
	if err != nil {
		t.Fatalf("NewConsumer() error = %v", err)
	}
	if consumer == nil {
		t.Fatal("expected non-nil consumer")
	}

	for _, subject := range []string{
		SubjectCompareCompleted,
		SubjectOFACScreeningCompleted,
		SubjectOFACScreeningReview,
	} {
		if _, ok := sub.subjects[subject]; !ok {
			t.Fatalf("expected subject %q to be subscribed", subject)
		}
	}
}

func TestConsumerDecodesCompareEvent(t *testing.T) {
	t.Parallel()

	sub := newFakeSubscriber()

	var got Event[CompareCompletedPayload]

	_, err := NewConsumer(sub, slog.Default(), Handlers{
		OnCompareCompleted: func(ctx context.Context, evt Event[CompareCompletedPayload]) error {
			got = evt
			return nil
		},
	})
	if err != nil {
		t.Fatalf("NewConsumer() error = %v", err)
	}

	handler := sub.subjects[SubjectCompareCompleted]
	if handler == nil {
		t.Fatal("expected compare completed handler to be registered")
	}

	raw, err := json.Marshal(Envelope{
		EventID:       "evt_1",
		EventType:     SubjectCompareCompleted,
		TenantID:      "tenant-1",
		CorrelationID: "corr-1",
		CaseID:        "case-1",
		OccurredAt:    "2026-04-14T12:00:00Z",
		Payload: mustRawJSON(t, CompareCompletedPayload{
			TenantID:            "tenant-1",
			CaseID:              "case-1",
			CorrelationID:       "corr-1",
			DecisionTraceID:     "dt_1",
			ExplanationID:       "exp_1",
			Disposition:         "resolved",
			ConfidenceBand:      "high",
			LeftSourceSystem:    "kyc_applications",
			LeftSourceRecordID:  "left-record",
			RightSourceSystem:   "watchlist_candidates",
			RightSourceRecordID: "right-record",
		}),
	})
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}

	handler(&nats.Msg{
		Subject: SubjectCompareCompleted,
		Data:    raw,
	})

	if got.EventID != "evt_1" {
		t.Fatalf("EventID = %q, want %q", got.EventID, "evt_1")
	}
	if got.TenantID != "tenant-1" {
		t.Fatalf("TenantID = %q, want %q", got.TenantID, "tenant-1")
	}
	if got.CaseID != "case-1" {
		t.Fatalf("CaseID = %q, want %q", got.CaseID, "case-1")
	}
	if got.Payload.DecisionTraceID != "dt_1" {
		t.Fatalf("DecisionTraceID = %q, want %q", got.Payload.DecisionTraceID, "dt_1")
	}
	if got.Payload.ExplanationID != "exp_1" {
		t.Fatalf("ExplanationID = %q, want %q", got.Payload.ExplanationID, "exp_1")
	}
	if got.Payload.Disposition != "resolved" {
		t.Fatalf("Disposition = %q, want %q", got.Payload.Disposition, "resolved")
	}
	if got.Payload.ConfidenceBand != "high" {
		t.Fatalf("ConfidenceBand = %q, want %q", got.Payload.ConfidenceBand, "high")
	}
}

func TestConsumerDecodesOFACScreeningReviewEvent(t *testing.T) {
	t.Parallel()

	sub := newFakeSubscriber()

	var got Event[OFACScreeningPayload]

	_, err := NewConsumer(sub, slog.Default(), Handlers{
		OnOFACScreeningReview: func(ctx context.Context, evt Event[OFACScreeningPayload]) error {
			got = evt
			return nil
		},
	})
	if err != nil {
		t.Fatalf("NewConsumer() error = %v", err)
	}

	handler := sub.subjects[SubjectOFACScreeningReview]
	if handler == nil {
		t.Fatal("expected screening review handler to be registered")
	}

	raw, err := json.Marshal(Envelope{
		EventID:       "evt_2",
		EventType:     SubjectOFACScreeningReview,
		TenantID:      "tenant-1",
		CorrelationID: "corr-2",
		CaseID:        "case-2",
		OccurredAt:    "2026-04-14T12:01:00Z",
		Payload: mustRawJSON(t, OFACScreeningPayload{
			TenantID:          "tenant-1",
			CaseID:            "case-2",
			CorrelationID:     "corr-2",
			ScreeningID:       "scr_1",
			Decision:          "manual_review",
			DecisionTraceID:   "dt_2",
			ExplanationID:     "exp_2",
			SubjectName:       "Jane Citizen",
			CandidateCount:    1,
			TopCandidateName:  "Jane Citizen",
			TopCandidateScore: 95,
		}),
	})
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}

	handler(&nats.Msg{
		Subject: SubjectOFACScreeningReview,
		Data:    raw,
	})

	if got.EventID != "evt_2" {
		t.Fatalf("EventID = %q, want %q", got.EventID, "evt_2")
	}
	if got.Payload.ScreeningID != "scr_1" {
		t.Fatalf("ScreeningID = %q, want %q", got.Payload.ScreeningID, "scr_1")
	}
	if got.Payload.Decision != "manual_review" {
		t.Fatalf("Decision = %q, want %q", got.Payload.Decision, "manual_review")
	}
	if got.Payload.TopCandidateName != "Jane Citizen" {
		t.Fatalf("TopCandidateName = %q, want %q", got.Payload.TopCandidateName, "Jane Citizen")
	}
	if got.Payload.TopCandidateScore != 95 {
		t.Fatalf("TopCandidateScore = %v, want %v", got.Payload.TopCandidateScore, 95)
	}
}

func mustRawJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()

	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal raw payload: %v", err)
	}
	return json.RawMessage(b)
}
