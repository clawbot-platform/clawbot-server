package db

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type migration struct {
	version  int
	upFile   string
	downFile string
}

func ApplyAll(ctx context.Context, pool *pgxpool.Pool) error {
	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	if err := ensureSchemaMigrations(ctx, pool); err != nil {
		return err
	}

	applied, err := appliedVersions(ctx, pool)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if applied[migration.version] {
			continue
		}

		body, err := migrationFiles.ReadFile(migration.upFile)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", migration.upFile, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration transaction: %w", err)
		}

		if _, err := tx.Exec(ctx, string(body)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", migration.upFile, err)
		}
		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", migration.version); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %d: %w", migration.version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %d: %w", migration.version, err)
		}
	}

	return nil
}

func DownOne(ctx context.Context, pool *pgxpool.Pool) error {
	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	if err := ensureSchemaMigrations(ctx, pool); err != nil {
		return err
	}

	var version int
	if err := pool.QueryRow(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version); err != nil {
		return fmt.Errorf("load latest migration version: %w", err)
	}
	if version == 0 {
		return nil
	}

	var target migration
	found := false
	for _, migration := range migrations {
		if migration.version == version {
			target = migration
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("no down migration found for version %d", version)
	}

	body, err := migrationFiles.ReadFile(target.downFile)
	if err != nil {
		return fmt.Errorf("read down migration %s: %w", target.downFile, err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin rollback transaction: %w", err)
	}

	if _, err := tx.Exec(ctx, string(body)); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("apply down migration %s: %w", target.downFile, err)
	}
	if _, err := tx.Exec(ctx, "DELETE FROM schema_migrations WHERE version = $1", version); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("delete migration record %d: %w", version, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit down migration %d: %w", version, err)
	}

	return nil
}

func ensureSchemaMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	const query = `
CREATE TABLE IF NOT EXISTS schema_migrations (
  version BIGINT PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)
`

	if _, err := pool.Exec(ctx, query); err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}
	return nil
}

func appliedVersions(ctx context.Context, pool *pgxpool.Pool) (map[int]bool, error) {
	rows, err := pool.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan migration version: %w", err)
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

func loadMigrations() ([]migration, error) {
	entries, err := fs.Glob(migrationFiles, "migrations/*.sql")
	if err != nil {
		return nil, fmt.Errorf("glob migration files: %w", err)
	}

	byVersion := make(map[int]*migration)
	for _, entry := range entries {
		base := strings.TrimPrefix(entry, "migrations/")
		parts := strings.Split(base, "_")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid migration name %q", base)
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("parse migration version from %q: %w", base, err)
		}

		item := byVersion[version]
		if item == nil {
			item = &migration{version: version}
			byVersion[version] = item
		}

		switch {
		case strings.HasSuffix(base, ".up.sql"):
			item.upFile = entry
		case strings.HasSuffix(base, ".down.sql"):
			item.downFile = entry
		default:
			return nil, fmt.Errorf("invalid migration suffix for %q", base)
		}
	}

	result := make([]migration, 0, len(byVersion))
	for _, item := range byVersion {
		if item.upFile == "" || item.downFile == "" {
			return nil, fmt.Errorf("migration %d is missing up or down file", item.version)
		}
		result = append(result, *item)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].version < result[j].version
	})

	return result, nil
}
