package runs

import (
	"context"
	"encoding/json"
	"testing"

	"clawbot-server/internal/platform/audit"
	"clawbot-server/internal/platform/scheduler"
	"clawbot-server/internal/platform/store"
)

type transactorStub struct{}

func (transactorStub) InTx(ctx context.Context, fn func(context.Context, store.DBTX) error) error {
	return fn(ctx, nil)
}

type repositoryStub struct {
	item Run
}

func (s *repositoryStub) List(context.Context, store.DBTX) ([]Run, error) { return nil, nil }
func (s *repositoryStub) Get(context.Context, store.DBTX, string) (Run, error) {
	return s.item, nil
}
func (s *repositoryStub) Create(context.Context, store.DBTX, CreateInput) (Run, error) {
	return s.item, nil
}
func (s *repositoryStub) Update(context.Context, store.DBTX, Run) (Run, error) {
	return s.item, nil
}

type auditStub struct {
	events []audit.Event
}

func (s *auditStub) Record(_ context.Context, _ store.DBTX, event audit.Event) error {
	s.events = append(s.events, event)
	return nil
}

type schedulerStub struct {
	signals []scheduler.Signal
}

func (s *schedulerStub) RecordRunIntent(_ context.Context, _ store.DBTX, signal scheduler.Signal) error {
	s.signals = append(s.signals, signal)
	return nil
}

func TestManagerCreateRecordsAuditAndSchedulerIntent(t *testing.T) {
	repo := &repositoryStub{item: Run{
		ID:           "run-1",
		Name:         "Phase 1 run",
		Status:       "pending",
		ScenarioType: "placeholder",
		MetadataJSON: json.RawMessage(`{"mode":"test"}`),
	}}
	audits := &auditStub{}
	schedulerService := &schedulerStub{}
	manager := NewManager(nil, transactorStub{}, repo, audits, schedulerService)

	_, err := manager.Create(context.Background(), CreateInput{
		Name:         "Phase 1 run",
		Status:       "pending",
		ScenarioType: "placeholder",
		MetadataJSON: json.RawMessage(`{"mode":"test"}`),
	}, "tester")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if len(audits.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(audits.events))
	}
	if len(schedulerService.signals) != 1 {
		t.Fatalf("expected 1 scheduler signal, got %d", len(schedulerService.signals))
	}
}
