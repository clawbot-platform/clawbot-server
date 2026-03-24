package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"clawbot-server/internal/platform/store"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, q store.DBTX, event Event) error {
	if len(event.Payload) == 0 {
		event.Payload = json.RawMessage(`{}`)
	}

	if err := s.repo.Create(ctx, q, event); err != nil {
		return fmt.Errorf("create audit event: %w", err)
	}

	return nil
}

type PostgresRepository struct{}

func NewPostgresRepository() *PostgresRepository {
	return &PostgresRepository{}
}

func (r *PostgresRepository) Create(ctx context.Context, q store.DBTX, event Event) error {
	const query = `
INSERT INTO audit_events (
  event_type,
  entity_type,
  entity_id,
  actor,
  payload_json
)
VALUES ($1, $2, $3, $4, $5)
`

	_, err := q.Exec(ctx, query,
		event.EventType,
		event.EntityType,
		event.EntityID,
		event.Actor,
		event.Payload,
	)

	return err
}
