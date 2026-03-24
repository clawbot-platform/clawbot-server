package policies

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

func (r *PostgresRepository) List(ctx context.Context, q store.DBTX) ([]Policy, error) {
	const query = `
SELECT id, name, category, version, enabled, description, rules_json, created_at, updated_at
FROM policies
ORDER BY created_at DESC
`

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Policy
	for rows.Next() {
		var item Policy
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Category,
			&item.Version,
			&item.Enabled,
			&item.Description,
			&item.RulesJSON,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) Get(ctx context.Context, q store.DBTX, id string) (Policy, error) {
	const query = `
SELECT id, name, category, version, enabled, description, rules_json, created_at, updated_at
FROM policies
WHERE id = $1
`

	var item Policy
	err := q.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.Name,
		&item.Category,
		&item.Version,
		&item.Enabled,
		&item.Description,
		&item.RulesJSON,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Policy{}, store.ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) Create(ctx context.Context, q store.DBTX, input CreateInput) (Policy, error) {
	const query = `
INSERT INTO policies (name, category, version, enabled, description, rules_json)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, name, category, version, enabled, description, rules_json, created_at, updated_at
`

	var item Policy
	err := q.QueryRow(ctx, query,
		input.Name,
		input.Category,
		input.Version,
		input.Enabled,
		input.Description,
		input.RulesJSON,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Category,
		&item.Version,
		&item.Enabled,
		&item.Description,
		&item.RulesJSON,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	return item, err
}

func (r *PostgresRepository) Update(ctx context.Context, q store.DBTX, item Policy) (Policy, error) {
	const query = `
UPDATE policies
SET
  name = $2,
  category = $3,
  version = $4,
  enabled = $5,
  description = $6,
  rules_json = $7,
  updated_at = NOW()
WHERE id = $1
RETURNING id, name, category, version, enabled, description, rules_json, created_at, updated_at
`

	var updated Policy
	err := q.QueryRow(ctx, query,
		item.ID,
		item.Name,
		item.Category,
		item.Version,
		item.Enabled,
		item.Description,
		item.RulesJSON,
	).Scan(
		&updated.ID,
		&updated.Name,
		&updated.Category,
		&updated.Version,
		&updated.Enabled,
		&updated.Description,
		&updated.RulesJSON,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	return updated, err
}
