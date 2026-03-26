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
	list         []Run
	item         Run
	createdInput CreateInput
	updated      Run
}

func (s *repositoryStub) List(context.Context, store.DBTX) ([]Run, error) {
	return s.list, nil
}

func (s *repositoryStub) Get(context.Context, store.DBTX, string) (Run, error) {
	return s.item, nil
}

func (s *repositoryStub) Create(_ context.Context, _ store.DBTX, input CreateInput) (Run, error) {
	s.createdInput = input
	s.item = Run{
		ID:           "run-1",
		Name:         input.Name,
		Description:  input.Description,
		Status:       input.Status,
		ScenarioType: input.ScenarioType,
		CreatedBy:    input.CreatedBy,
		MetadataJSON: input.MetadataJSON,
	}
	return s.item, nil
}

func (s *repositoryStub) Update(_ context.Context, _ store.DBTX, item Run) (Run, error) {
	s.updated = item
	s.item = item
	return item, nil
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

func TestManagerListGetCreateUpdate(t *testing.T) {
	repo := &repositoryStub{
		list: []Run{{ID: "run-1", Name: "Platform baseline", Status: "pending", MetadataJSON: json.RawMessage(`{}`)}},
		item: Run{ID: "run-1", Name: "Platform baseline", Status: "pending", MetadataJSON: json.RawMessage(`{}`)},
	}
	audits := &auditStub{}
	scheduler := &schedulerStub{}
	manager := NewManager(nil, transactorStub{}, repo, audits, scheduler)

	items, err := manager.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != "run-1" {
		t.Fatalf("unexpected List() result %#v", items)
	}

	item, err := manager.Get(context.Background(), "run-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if item.ID != "run-1" {
		t.Fatalf("unexpected Get() result %#v", item)
	}

	created, err := manager.Create(context.Background(), CreateInput{
		Name:         "  Platform baseline  ",
		Description:  "  shared stack smoke  ",
		ScenarioType: "  smoke  ",
	}, "platform-admin")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Status != "pending" {
		t.Fatalf("expected default status pending, got %s", created.Status)
	}
	if created.CreatedBy != "platform-admin" {
		t.Fatalf("expected default created_by from actor, got %s", created.CreatedBy)
	}
	if string(repo.createdInput.MetadataJSON) != "{}" {
		t.Fatalf("expected default metadata json, got %s", repo.createdInput.MetadataJSON)
	}
	if len(scheduler.signals) != 1 || scheduler.signals[0].Reason != "run.created" {
		t.Fatalf("unexpected scheduler signals %#v", scheduler.signals)
	}

	newName := "Reusable platform baseline"
	newStatus := "running"
	newMetadata := json.RawMessage(`{"kind":"foundation"}`)
	updated, err := manager.Update(context.Background(), "run-1", UpdateInput{
		Name:         &newName,
		Status:       &newStatus,
		MetadataJSON: &newMetadata,
	}, "platform-admin")
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Name != newName || updated.Status != newStatus {
		t.Fatalf("unexpected Update() result %#v", updated)
	}
	if len(audits.events) != 2 {
		t.Fatalf("expected 2 audit events, got %d", len(audits.events))
	}
	if len(scheduler.signals) != 2 || scheduler.signals[1].Reason != "run.updated" {
		t.Fatalf("unexpected scheduler signals %#v", scheduler.signals)
	}
}

func TestMergeRunRejectsInvalidStatus(t *testing.T) {
	status := "unknown"
	_, err := mergeRun(Run{Name: "Run", Status: "pending", MetadataJSON: json.RawMessage(`{}`)}, UpdateInput{Status: &status})
	if err == nil {
		t.Fatal("expected invalid status error")
	}
}

func TestMergeRunRejectsEmptyName(t *testing.T) {
	name := "   "
	_, err := mergeRun(Run{Name: "Run", Status: "pending", MetadataJSON: json.RawMessage(`{}`)}, UpdateInput{Name: &name})
	if err == nil {
		t.Fatal("expected name validation error")
	}
}
