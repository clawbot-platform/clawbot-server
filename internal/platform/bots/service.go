package bots

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

func (m *Manager) List(ctx context.Context) ([]Bot, error) {
	return m.repo.List(ctx, m.db)
}

func (m *Manager) Get(ctx context.Context, id string) (Bot, error) {
	return m.repo.Get(ctx, m.db, id)
}

func (m *Manager) Create(ctx context.Context, input CreateInput, actor string) (Bot, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return Bot{}, fmt.Errorf("name is required")
	}
	if input.Status == "" {
		input.Status = "active"
	}
	if !isValidStatus(input.Status) {
		return Bot{}, fmt.Errorf("invalid status %q", input.Status)
	}
	if len(input.ConfigJSON) == 0 {
		input.ConfigJSON = json.RawMessage(`{}`)
	}

	var created Bot
	err := m.tx.InTx(ctx, func(ctx context.Context, q store.DBTX) error {
		var err error
		created, err = m.repo.Create(ctx, q, input)
		if err != nil {
			return err
		}
		return recordAudit(ctx, m.audits, q, "bot.created", actor, created.ID, created)
	})
	if err != nil {
		return Bot{}, err
	}

	return created, nil
}

func (m *Manager) Update(ctx context.Context, id string, input UpdateInput, actor string) (Bot, error) {
	var updated Bot
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
		return recordAudit(ctx, m.audits, q, "bot.updated", actor, updated.ID, updated)
	})
	if err != nil {
		return Bot{}, err
	}
	return updated, nil
}

func merge(existing Bot, input UpdateInput) (Bot, error) {
	if input.Name != nil {
		existing.Name = strings.TrimSpace(*input.Name)
	}
	if input.Role != nil {
		existing.Role = strings.TrimSpace(*input.Role)
	}
	if input.Runtime != nil {
		existing.Runtime = strings.TrimSpace(*input.Runtime)
	}
	if input.Status != nil {
		existing.Status = strings.TrimSpace(*input.Status)
	}
	if input.RepoHint != nil {
		existing.RepoHint = strings.TrimSpace(*input.RepoHint)
	}
	if input.Version != nil {
		existing.Version = strings.TrimSpace(*input.Version)
	}
	if input.ConfigJSON != nil {
		existing.ConfigJSON = *input.ConfigJSON
	}
	if existing.Name == "" {
		return Bot{}, fmt.Errorf("name is required")
	}
	if !isValidStatus(existing.Status) {
		return Bot{}, fmt.Errorf("invalid status %q", existing.Status)
	}
	if len(existing.ConfigJSON) == 0 {
		existing.ConfigJSON = json.RawMessage(`{}`)
	}
	return existing, nil
}

func isValidStatus(status string) bool {
	switch status {
	case "active", "inactive", "deprecated":
		return true
	default:
		return false
	}
}

func recordAudit(ctx context.Context, recorder audit.Recorder, q store.DBTX, eventType string, actor string, entityID string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal audit payload: %w", err)
	}
	return recorder.Record(ctx, q, audit.Event{
		EventType:  eventType,
		EntityType: "bot",
		EntityID:   entityID,
		Actor:      actor,
		Payload:    body,
	})
}
