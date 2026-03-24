package runs

import (
	"context"
	"errors"

	"clawbot-server/internal/platform/store"

	"github.com/jackc/pgx/v5"
)

type PostgresRepository struct{}

func NewPostgresRepository() *PostgresRepository {
	return &PostgresRepository{}
}

func (r *PostgresRepository) List(ctx context.Context, q store.DBTX) ([]Run, error) {
	const query = `
SELECT id, name, description, status, scenario_type, created_by, created_at, updated_at, started_at, completed_at, metadata_json
FROM runs
ORDER BY created_at DESC
`

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Run
	for rows.Next() {
		var item Run
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Status,
			&item.ScenarioType,
			&item.CreatedBy,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.StartedAt,
			&item.CompletedAt,
			&item.MetadataJSON,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *PostgresRepository) Get(ctx context.Context, q store.DBTX, id string) (Run, error) {
	const query = `
SELECT id, name, description, status, scenario_type, created_by, created_at, updated_at, started_at, completed_at, metadata_json
FROM runs
WHERE id = $1
`

	var item Run
	err := q.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.Status,
		&item.ScenarioType,
		&item.CreatedBy,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.StartedAt,
		&item.CompletedAt,
		&item.MetadataJSON,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Run{}, store.ErrNotFound
	}

	return item, err
}

func (r *PostgresRepository) Create(ctx context.Context, q store.DBTX, input CreateInput) (Run, error) {
	const query = `
INSERT INTO runs (
  name,
  description,
  status,
  scenario_type,
  created_by,
  metadata_json
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, name, description, status, scenario_type, created_by, created_at, updated_at, started_at, completed_at, metadata_json
`

	var item Run
	err := q.QueryRow(ctx, query,
		input.Name,
		input.Description,
		input.Status,
		input.ScenarioType,
		input.CreatedBy,
		input.MetadataJSON,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.Status,
		&item.ScenarioType,
		&item.CreatedBy,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.StartedAt,
		&item.CompletedAt,
		&item.MetadataJSON,
	)

	return item, err
}

func (r *PostgresRepository) Update(ctx context.Context, q store.DBTX, item Run) (Run, error) {
	const query = `
UPDATE runs
SET
  name = $2,
  description = $3,
  status = $4,
  scenario_type = $5,
  started_at = $6,
  completed_at = $7,
  metadata_json = $8,
  updated_at = NOW()
WHERE id = $1
RETURNING id, name, description, status, scenario_type, created_by, created_at, updated_at, started_at, completed_at, metadata_json
`

	var updated Run
	err := q.QueryRow(ctx, query,
		item.ID,
		item.Name,
		item.Description,
		item.Status,
		item.ScenarioType,
		item.StartedAt,
		item.CompletedAt,
		item.MetadataJSON,
	).Scan(
		&updated.ID,
		&updated.Name,
		&updated.Description,
		&updated.Status,
		&updated.ScenarioType,
		&updated.CreatedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
		&updated.StartedAt,
		&updated.CompletedAt,
		&updated.MetadataJSON,
	)

	return updated, err
}
