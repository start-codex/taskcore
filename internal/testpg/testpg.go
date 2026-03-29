// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package testpg

import (
	"context"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/start-codex/trazawork/migrations"
)

const testDSNEnv = "MINI_JIRA_TEST_DSN"

// Open opens a test database connection from MINI_JIRA_TEST_DSN.
// Skips the test if the variable is not set.
// Registers t.Cleanup to close the connection.
func Open(t *testing.T) *sqlx.DB {
	t.Helper()
	dsn := os.Getenv(testDSNEnv)
	if dsn == "" {
		t.Skipf("%s not set; skipping integration test", testDSNEnv)
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("testpg.Open: connect: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// EnsureMigrated applies all pending migrations via migrations.Up.
// If the database was previously bootstrapped with the old inline ensureSchema helper
// (tables present but schema_migrations missing), it fails fast with a clear message.
func EnsureMigrated(t *testing.T, db *sqlx.DB) {
	t.Helper()
	ctx := context.Background()

	var schemaMigrationsExists bool
	if err := db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'schema_migrations'
		)`,
	).Scan(&schemaMigrationsExists); err != nil {
		t.Fatalf("testpg.EnsureMigrated: check schema_migrations: %v", err)
	}

	if !schemaMigrationsExists {
		for _, table := range []string{"workspaces", "app_users", "projects"} {
			var exists bool
			if err := db.QueryRowContext(ctx,
				`SELECT EXISTS(
					SELECT 1 FROM information_schema.tables
					WHERE table_schema = 'public' AND table_name = $1
				)`, table,
			).Scan(&exists); err != nil {
				t.Fatalf("testpg.EnsureMigrated: check legacy table %q: %v", table, err)
			}
			if exists {
				t.Fatalf(
					"testpg.EnsureMigrated: legacy schema detected: test database was initialized "+
						"with the old ensureSchema helper (schema_migrations table missing). "+
						"Reset the test database and re-run.",
				)
			}
		}
	}

	if err := migrations.Up(ctx, db.DB); err != nil {
		t.Fatalf("testpg.EnsureMigrated: apply migrations: %v", err)
	}
}

// UniqueSuffix returns an 8-character random hex string suitable for use
// in unique identifiers like emails or slugs within test fixtures.
func UniqueSuffix(t *testing.T, db *sqlx.DB) string {
	t.Helper()
	var suffix string
	if err := db.QueryRowContext(context.Background(),
		`SELECT substr(replace(gen_random_uuid()::text, '-', ''), 1, 8)`,
	).Scan(&suffix); err != nil {
		t.Fatalf("testpg.UniqueSuffix: %v", err)
	}
	return suffix
}

// SeedUser inserts a minimal user row and registers cleanup.
// Returns the user ID.
func SeedUser(t *testing.T, db *sqlx.DB) string {
	t.Helper()
	suffix := UniqueSuffix(t, db)
	var id string
	if err := db.QueryRowContext(context.Background(),
		`INSERT INTO app_users (email, name, password_hash)
		 VALUES ($1, $2, '')
		 RETURNING id`,
		suffix+"@test.local", "Test User "+suffix,
	).Scan(&id); err != nil {
		t.Fatalf("testpg.SeedUser: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM app_users WHERE id = $1`, id)
	})
	return id
}

// SeedWorkspace inserts a minimal workspace row and registers cleanup.
// Cleanup deletes the workspace; cascade removes members and projects.
// Returns the workspace ID.
func SeedWorkspace(t *testing.T, db *sqlx.DB) string {
	t.Helper()
	suffix := UniqueSuffix(t, db)
	var id string
	if err := db.QueryRowContext(context.Background(),
		`INSERT INTO workspaces (name, slug)
		 VALUES ($1, $2)
		 RETURNING id`,
		"Workspace "+suffix, "ws-"+suffix,
	).Scan(&id); err != nil {
		t.Fatalf("testpg.SeedWorkspace: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM workspaces WHERE id = $1`, id)
	})
	return id
}

// SeedProject inserts a minimal project row under workspaceID with the given key.
// Does not register t.Cleanup — cleanup is handled by cascade from SeedWorkspace
// or by the local fixture that owns the workspace.
// Returns the project ID.
func SeedProject(t *testing.T, db *sqlx.DB, workspaceID, key string) string {
	t.Helper()
	var id string
	if err := db.QueryRowContext(context.Background(),
		`INSERT INTO projects (workspace_id, name, key, description)
		 VALUES ($1, $2, $3, '')
		 RETURNING id`,
		workspaceID, "Project "+key, key,
	).Scan(&id); err != nil {
		t.Fatalf("testpg.SeedProject: %v", err)
	}
	return id
}
