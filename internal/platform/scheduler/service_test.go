package scheduler

import (
	"context"
	"encoding/json"
	"testing"

	"clawbot-server/internal/platform/audit"
	"clawbot-server/internal/platform/store"
)

type auditRecorderStub struct {
	events []audit.Event
}

func (s *auditRecorderStub) Record(_ context.Context, _ store.DBTX, event audit.Event) error {
	s.events = append(s.events, event)
	return nil
}

func TestPlaceholderServiceRecordRunIntent(t *testing.T) {
	recorder := &auditRecorderStub{}
	service := NewPlaceholderService(recorder)

	err := service.RecordRunIntent(context.Background(), nil, Signal{
		RunID:   "run-123",
		RunName: "Baseline run",
		Status:  "pending",
		Actor:   "tester",
		Reason:  "run.created",
	})
	if err != nil {
		t.Fatalf("RecordRunIntent() error = %v", err)
	}

	if len(recorder.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(recorder.events))
	}

	if recorder.events[0].EventType != "scheduler.intent.recorded" {
		t.Fatalf("unexpected event type: %s", recorder.events[0].EventType)
	}

	var payload map[string]any
	if err := json.Unmarshal(recorder.events[0].Payload, &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if payload["run_id"] != "run-123" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}
