package bots

import (
	"context"
	"encoding/json"
	"testing"

	"clawbot-server/internal/platform/audit"
	"clawbot-server/internal/platform/store"
)

type transactorStub struct{}

func (transactorStub) InTx(ctx context.Context, fn func(context.Context, store.DBTX) error) error {
	return fn(ctx, nil)
}

type repositoryStub struct {
	list         []Bot
	item         Bot
	createdInput CreateInput
	updated      Bot
}

func (s *repositoryStub) List(context.Context, store.DBTX) ([]Bot, error) {
	return s.list, nil
}

func (s *repositoryStub) Get(context.Context, store.DBTX, string) (Bot, error) {
	return s.item, nil
}

func (s *repositoryStub) Create(_ context.Context, _ store.DBTX, input CreateInput) (Bot, error) {
	s.createdInput = input
	s.item = Bot{
		ID:         "bot-1",
		Name:       input.Name,
		Role:       input.Role,
		Runtime:    input.Runtime,
		Status:     input.Status,
		RepoHint:   input.RepoHint,
		Version:    input.Version,
		ConfigJSON: input.ConfigJSON,
	}
	return s.item, nil
}

func (s *repositoryStub) Update(_ context.Context, _ store.DBTX, item Bot) (Bot, error) {
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

func TestManagerListGetCreateUpdate(t *testing.T) {
	repo := &repositoryStub{
		list: []Bot{{ID: "bot-1", Name: "Generic Runner", Status: "active", ConfigJSON: json.RawMessage(`{}`)}},
		item: Bot{ID: "bot-1", Name: "Generic Runner", Status: "active", ConfigJSON: json.RawMessage(`{}`)},
	}
	audits := &auditStub{}
	manager := NewManager(nil, transactorStub{}, repo, audits)

	items, err := manager.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != "bot-1" {
		t.Fatalf("unexpected List() result %#v", items)
	}

	item, err := manager.Get(context.Background(), "bot-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if item.ID != "bot-1" {
		t.Fatalf("unexpected Get() result %#v", item)
	}

	created, err := manager.Create(context.Background(), CreateInput{
		Name:    "  Generic Runner  ",
		Role:    "orchestrator",
		Runtime: "zeroclaw",
	}, "platform-admin")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Status != "active" {
		t.Fatalf("expected default status active, got %s", created.Status)
	}
	if string(repo.createdInput.ConfigJSON) != "{}" {
		t.Fatalf("expected default config json, got %s", repo.createdInput.ConfigJSON)
	}

	newName := "Reusable Control Plane Bot"
	newStatus := "deprecated"
	updated, err := manager.Update(context.Background(), "bot-1", UpdateInput{
		Name:   &newName,
		Status: &newStatus,
	}, "platform-admin")
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Name != newName || repo.updated.Status != newStatus {
		t.Fatalf("unexpected Update() result %#v / %#v", updated, repo.updated)
	}
	if len(audits.events) != 2 {
		t.Fatalf("expected 2 audit events, got %d", len(audits.events))
	}
}

func TestMergeRejectsInvalidStatus(t *testing.T) {
	status := "unknown"
	_, err := merge(Bot{Name: "Bot", Status: "active", ConfigJSON: json.RawMessage(`{}`)}, UpdateInput{Status: &status})
	if err == nil {
		t.Fatal("expected invalid status error")
	}
}
