package policies

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
	list         []Policy
	item         Policy
	createdInput CreateInput
	updated      Policy
}

func (s *repositoryStub) List(context.Context, store.DBTX) ([]Policy, error) {
	return s.list, nil
}

func (s *repositoryStub) Get(context.Context, store.DBTX, string) (Policy, error) {
	return s.item, nil
}

func (s *repositoryStub) Create(_ context.Context, _ store.DBTX, input CreateInput) (Policy, error) {
	s.createdInput = input
	s.item = Policy{
		ID:          "policy-1",
		Name:        input.Name,
		Category:    input.Category,
		Version:     input.Version,
		Enabled:     input.Enabled,
		Description: input.Description,
		RulesJSON:   input.RulesJSON,
	}
	return s.item, nil
}

func (s *repositoryStub) Update(_ context.Context, _ store.DBTX, item Policy) (Policy, error) {
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
		list: []Policy{{ID: "policy-1", Name: "Generic Control", RulesJSON: json.RawMessage(`{}`)}},
		item: Policy{ID: "policy-1", Name: "Generic Control", RulesJSON: json.RawMessage(`{}`)},
	}
	audits := &auditStub{}
	manager := NewManager(nil, transactorStub{}, repo, audits)

	items, err := manager.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != "policy-1" {
		t.Fatalf("unexpected List() result %#v", items)
	}

	item, err := manager.Get(context.Background(), "policy-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if item.ID != "policy-1" {
		t.Fatalf("unexpected Get() result %#v", item)
	}

	created, err := manager.Create(context.Background(), CreateInput{
		Name:     "  Generic Control  ",
		Category: "runtime",
		Version:  "v1",
	}, "platform-admin")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Name != "Generic Control" {
		t.Fatalf("expected trimmed name, got %#v", created)
	}
	if string(repo.createdInput.RulesJSON) != "{}" {
		t.Fatalf("expected default rules json, got %s", repo.createdInput.RulesJSON)
	}

	description := "Reusable platform policy"
	enabled := true
	updated, err := manager.Update(context.Background(), "policy-1", UpdateInput{
		Description: &description,
		Enabled:     &enabled,
	}, "platform-admin")
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Description != description || !repo.updated.Enabled {
		t.Fatalf("unexpected Update() result %#v / %#v", updated, repo.updated)
	}
	if len(audits.events) != 2 {
		t.Fatalf("expected 2 audit events, got %d", len(audits.events))
	}
}

func TestMergeRejectsEmptyName(t *testing.T) {
	name := " "
	_, err := merge(Policy{Name: "Policy", RulesJSON: json.RawMessage(`{}`)}, UpdateInput{Name: &name})
	if err == nil {
		t.Fatal("expected empty name error")
	}
}
