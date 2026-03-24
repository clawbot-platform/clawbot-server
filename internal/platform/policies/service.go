package policies

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"clawbot-server/internal/platform/audit"
	"clawbot-server/internal/platform/store"
)

type Manager struct {
	repo   Repository
	tx     store.Transactor
	audits audit.Recorder
	db     store.DBTX
}

func NewManager(db store.DBTX, tx store.Transactor, repo Repository, audits audit.Recorder) *Manager {
	return &Manager{repo: repo, tx: tx, audits: audits, db: db}
}

func (m *Manager) List(ctx context.Context) ([]Policy, error) {
	return m.repo.List(ctx, m.db)
}

func (m *Manager) Get(ctx context.Context, id string) (Policy, error) {
	return m.repo.Get(ctx, m.db, id)
}

func (m *Manager) Create(ctx context.Context, input CreateInput, actor string) (Policy, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return Policy{}, fmt.Errorf("name is required")
	}
	if len(input.RulesJSON) == 0 {
		input.RulesJSON = json.RawMessage(`{}`)
	}

	var created Policy
	err := m.tx.InTx(ctx, func(ctx context.Context, q store.DBTX) error {
		var err error
		created, err = m.repo.Create(ctx, q, input)
		if err != nil {
			return err
		}
		return recordAudit(ctx, m.audits, q, "policy.created", actor, created.ID, created)
	})
	if err != nil {
		return Policy{}, err
	}
	return created, nil
}

func (m *Manager) Update(ctx context.Context, id string, input UpdateInput, actor string) (Policy, error) {
	var updated Policy
	err := m.tx.InTx(ctx, func(ctx context.Context, q store.DBTX) error {
		existing, err := m.repo.Get(ctx, q, id)
		if err != nil {
			return err
		}
		merged, err := merge(existing, input)
		if err != nil {
			return err
		}
		updated, err = m.repo.Update(ctx, q, merged)
		if err != nil {
			return err
		}
		return recordAudit(ctx, m.audits, q, "policy.updated", actor, updated.ID, updated)
	})
	if err != nil {
		return Policy{}, err
	}
	return updated, nil
}

func merge(existing Policy, input UpdateInput) (Policy, error) {
	if input.Name != nil {
		existing.Name = strings.TrimSpace(*input.Name)
	}
	if input.Category != nil {
		existing.Category = strings.TrimSpace(*input.Category)
	}
	if input.Version != nil {
		existing.Version = strings.TrimSpace(*input.Version)
	}
	if input.Enabled != nil {
		existing.Enabled = *input.Enabled
	}
	if input.Description != nil {
		existing.Description = strings.TrimSpace(*input.Description)
	}
	if input.RulesJSON != nil {
		existing.RulesJSON = *input.RulesJSON
	}
	if existing.Name == "" {
		return Policy{}, fmt.Errorf("name is required")
	}
	if len(existing.RulesJSON) == 0 {
		existing.RulesJSON = json.RawMessage(`{}`)
	}
	return existing, nil
}

func recordAudit(ctx context.Context, recorder audit.Recorder, q store.DBTX, eventType string, actor string, entityID string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal audit payload: %w", err)
	}
	return recorder.Record(ctx, q, audit.Event{
		EventType:  eventType,
		EntityType: "policy",
		EntityID:   entityID,
		Actor:      actor,
		Payload:    body,
	})
}
