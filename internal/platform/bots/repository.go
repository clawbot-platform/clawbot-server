package bots

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

func (r *PostgresRepository) List(ctx context.Context, q store.DBTX) ([]Bot, error) {
	const query = `
SELECT id, name, role, runtime, status, repo_hint, version, config_json, created_at, updated_at
FROM bots
ORDER BY created_at DESC
`

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Bot
	for rows.Next() {
		var item Bot
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Role,
			&item.Runtime,
			&item.Status,
			&item.RepoHint,
			&item.Version,
			&item.ConfigJSON,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *PostgresRepository) Get(ctx context.Context, q store.DBTX, id string) (Bot, error) {
	const query = `
SELECT id, name, role, runtime, status, repo_hint, version, config_json, created_at, updated_at
FROM bots
WHERE id = $1
`

	var item Bot
	err := q.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.Name,
		&item.Role,
		&item.Runtime,
		&item.Status,
		&item.RepoHint,
		&item.Version,
		&item.ConfigJSON,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Bot{}, store.ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) Create(ctx context.Context, q store.DBTX, input CreateInput) (Bot, error) {
	const query = `
INSERT INTO bots (name, role, runtime, status, repo_hint, version, config_json)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, name, role, runtime, status, repo_hint, version, config_json, created_at, updated_at
`

	var item Bot
	err := q.QueryRow(ctx, query,
		input.Name,
		input.Role,
		input.Runtime,
		input.Status,
		input.RepoHint,
		input.Version,
		input.ConfigJSON,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Role,
		&item.Runtime,
		&item.Status,
		&item.RepoHint,
		&item.Version,
		&item.ConfigJSON,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	return item, err
}

func (r *PostgresRepository) Update(ctx context.Context, q store.DBTX, item Bot) (Bot, error) {
	const query = `
UPDATE bots
SET
  name = $2,
  role = $3,
  runtime = $4,
  status = $5,
  repo_hint = $6,
  version = $7,
  config_json = $8,
  updated_at = NOW()
WHERE id = $1
RETURNING id, name, role, runtime, status, repo_hint, version, config_json, created_at, updated_at
`

	var updated Bot
	err := q.QueryRow(ctx, query,
		item.ID,
		item.Name,
		item.Role,
		item.Runtime,
		item.Status,
		item.RepoHint,
		item.Version,
		item.ConfigJSON,
	).Scan(
		&updated.ID,
		&updated.Name,
		&updated.Role,
		&updated.Runtime,
		&updated.Status,
		&updated.RepoHint,
		&updated.Version,
		&updated.ConfigJSON,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	return updated, err
}
