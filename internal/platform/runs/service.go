package runs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"clawbot-server/internal/platform/audit"
	"clawbot-server/internal/platform/scheduler"
	"clawbot-server/internal/platform/store"
)

type Manager struct {
	repo      Repository
	tx        store.Transactor
	audits    audit.Recorder
	scheduler scheduler.Service
	db        store.DBTX
}

func NewManager(db store.DBTX, tx store.Transactor, repo Repository, audits audit.Recorder, scheduler scheduler.Service) *Manager {
	return &Manager{
		repo:      repo,
		tx:        tx,
		audits:    audits,
		scheduler: scheduler,
		db:        db,
	}
}

func (m *Manager) List(ctx context.Context) ([]Run, error) {
	return m.repo.List(ctx, m.db)
}

func (m *Manager) Get(ctx context.Context, id string) (Run, error) {
	return m.repo.Get(ctx, m.db, id)
}

func (m *Manager) Create(ctx context.Context, input CreateInput, actor string) (Run, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.ScenarioType = strings.TrimSpace(input.ScenarioType)
	input.CreatedBy = strings.TrimSpace(input.CreatedBy)

	if input.Name == "" {
		return Run{}, fmt.Errorf("name is required")
	}
	if input.Status == "" {
		input.Status = "pending"
	}
	if !isValidStatus(input.Status) {
		return Run{}, fmt.Errorf("invalid status %q", input.Status)
	}
	if len(input.MetadataJSON) == 0 {
		input.MetadataJSON = json.RawMessage(`{}`)
	}
	if input.CreatedBy == "" {
		input.CreatedBy = actor
	}

	var created Run
	err := m.tx.InTx(ctx, func(ctx context.Context, q store.DBTX) error {
		var err error
		created, err = m.repo.Create(ctx, q, input)
		if err != nil {
			return err
		}

		if err := recordEntityAudit(ctx, m.audits, q, "run.created", actor, created.ID, created); err != nil {
			return err
		}

		return m.scheduler.RecordRunIntent(ctx, q, scheduler.Signal{
			RunID:   created.ID,
			RunName: created.Name,
			Status:  created.Status,
			Actor:   actor,
			Reason:  "run.created",
		})
	})
	if err != nil {
		return Run{}, err
	}

	return created, nil
}

func (m *Manager) Update(ctx context.Context, id string, input UpdateInput, actor string) (Run, error) {
	var updated Run

	err := m.tx.InTx(ctx, func(ctx context.Context, q store.DBTX) error {
		existing, err := m.repo.Get(ctx, q, id)
		if err != nil {
			return err
		}

		merged, err := mergeRun(existing, input)
		if err != nil {
			return err
		}

		updated, err = m.repo.Update(ctx, q, merged)
		if err != nil {
			return err
		}

		if err := recordEntityAudit(ctx, m.audits, q, "run.updated", actor, updated.ID, updated); err != nil {
			return err
		}

		return m.scheduler.RecordRunIntent(ctx, q, scheduler.Signal{
			RunID:   updated.ID,
			RunName: updated.Name,
			Status:  updated.Status,
			Actor:   actor,
			Reason:  "run.updated",
		})
	})
	if err != nil {
		return Run{}, err
	}

	return updated, nil
}

func mergeRun(existing Run, input UpdateInput) (Run, error) {
	if input.Name != nil {
		existing.Name = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		existing.Description = strings.TrimSpace(*input.Description)
	}
	if input.Status != nil {
		existing.Status = strings.TrimSpace(*input.Status)
	}
	if input.ScenarioType != nil {
		existing.ScenarioType = strings.TrimSpace(*input.ScenarioType)
	}
	if input.MetadataJSON != nil {
		existing.MetadataJSON = *input.MetadataJSON
	}
	if input.StartedAt != nil {
		existing.StartedAt = input.StartedAt
	}
	if input.CompletedAt != nil {
		existing.CompletedAt = input.CompletedAt
	}

	if existing.Name == "" {
		return Run{}, fmt.Errorf("name is required")
	}
	if !isValidStatus(existing.Status) {
		return Run{}, fmt.Errorf("invalid status %q", existing.Status)
	}
	if len(existing.MetadataJSON) == 0 {
		existing.MetadataJSON = json.RawMessage(`{}`)
	}

	return existing, nil
}

func isValidStatus(status string) bool {
	switch status {
	case "pending", "scheduled", "running", "completed", "failed", "cancelled":
		return true
	default:
		return false
	}
}

func recordEntityAudit(ctx context.Context, recorder audit.Recorder, q store.DBTX, eventType string, actor string, entityID string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal audit payload: %w", err)
	}

	return recorder.Record(ctx, q, audit.Event{
		EventType:  eventType,
		EntityType: "run",
		EntityID:   entityID,
		Actor:      actor,
		Payload:    body,
	})
}
