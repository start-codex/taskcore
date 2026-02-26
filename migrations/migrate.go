package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

//go:embed *.sql
var files embed.FS

// Up applies all pending .up.sql migrations in version order.
// Uses the same schema_migrations table format as golang-migrate (single row, current version).
func Up(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version bigint  NOT NULL PRIMARY KEY,
			dirty   boolean NOT NULL
		)`); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	var current int64
	var dirty bool
	err := db.QueryRowContext(ctx, `SELECT version, dirty FROM schema_migrations LIMIT 1`).Scan(&current, &dirty)
	switch {
	case err == sql.ErrNoRows:
		// no migrations applied yet
	case err != nil:
		return fmt.Errorf("read schema_migrations: %w", err)
	case dirty:
		return fmt.Errorf("migration %d left in dirty state, manual fix required", current)
	}

	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}

	type migration struct {
		version int64
		file    string
	}
	var pending []migration
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		v, err := strconv.ParseInt(strings.SplitN(name, "_", 2)[0], 10, 64)
		if err != nil {
			continue
		}
		if v > current {
			pending = append(pending, migration{v, name})
		}
	}
	sort.Slice(pending, func(i, j int) bool { return pending[i].version < pending[j].version })

	for _, m := range pending {
		content, err := fs.ReadFile(files, m.file)
		if err != nil {
			return fmt.Errorf("read %s: %w", m.file, err)
		}

		if _, err := db.ExecContext(ctx, `DELETE FROM schema_migrations`); err != nil {
			return fmt.Errorf("clear schema_migrations before v%d: %w", m.version, err)
		}
		if _, err := db.ExecContext(ctx,
			`INSERT INTO schema_migrations (version, dirty) VALUES ($1, true)`, m.version); err != nil {
			return fmt.Errorf("mark dirty v%d: %w", m.version, err)
		}
		if _, err := db.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("apply v%d (%s): %w", m.version, m.file, err)
		}
		if _, err := db.ExecContext(ctx,
			`UPDATE schema_migrations SET dirty = false WHERE version = $1`, m.version); err != nil {
			return fmt.Errorf("mark clean v%d: %w", m.version, err)
		}
	}

	return nil
}
