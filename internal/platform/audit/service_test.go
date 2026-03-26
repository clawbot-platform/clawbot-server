package audit

import (
	"context"
	"errors"
	"testing"

	"clawbot-server/internal/platform/store"
)

type repositoryStub struct {
	event Event
	err   error
}

func (s *repositoryStub) Create(_ context.Context, _ store.DBTX, event Event) error {
	s.event = event
	return s.err
}

func TestServiceRecordDefaultsEmptyPayload(t *testing.T) {
	repo := &repositoryStub{}
	service := NewService(repo)

	err := service.Record(context.Background(), nil, Event{
		EventType:  "run.created",
		EntityType: "run",
		EntityID:   "run-1",
		Actor:      "tester",
	})
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if string(repo.event.Payload) != "{}" {
		t.Fatalf("expected default payload, got %s", repo.event.Payload)
	}
}

func TestServiceRecordWrapsRepositoryError(t *testing.T) {
	repo := &repositoryStub{err: errors.New("boom")}
	service := NewService(repo)

	err := service.Record(context.Background(), nil, Event{EventType: "run.created"})
	if err == nil || err.Error() != "create audit event: boom" {
		t.Fatalf("unexpected error %v", err)
	}
}
