package workspaces

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var testDSN = os.Getenv("MINI_JIRA_TEST_DSN")

func TestCreateWorkspace(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (CreateWorkspaceParams, func(*testing.T))
		wantErr error
	}{
		{
			name: "creates workspace successfully",
			arrange: func(t *testing.T, db *sqlx.DB) (CreateWorkspaceParams, func(*testing.T)) {
				p := CreateWorkspaceParams{Name: "Acme Corp", Slug: uniqueSlug(t, db)}
				return p, func(t *testing.T) {}
			},
		},
		{
			name: "returned workspace has correct fields",
			arrange: func(t *testing.T, db *sqlx.DB) (CreateWorkspaceParams, func(*testing.T)) {
				slug := uniqueSlug(t, db)
				p := CreateWorkspaceParams{Name: "Check Fields", Slug: slug}
				return p, func(t *testing.T) {}
			},
		},
		{
			name:    "duplicate slug",
			wantErr: ErrDuplicateSlug,
			arrange: func(t *testing.T, db *sqlx.DB) (CreateWorkspaceParams, func(*testing.T)) {
				slug := uniqueSlug(t, db)
				if _, err := CreateWorkspace(context.Background(), db, CreateWorkspaceParams{Name: "First", Slug: slug}); err != nil {
					t.Fatalf("seed workspace: %v", err)
				}
				return CreateWorkspaceParams{Name: "Second", Slug: slug}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, check := tt.arrange(t, db)
			got, err := CreateWorkspace(context.Background(), db, p)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateWorkspace() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil {
				if got.ID == "" {
					t.Fatal("expected non-empty id")
				}
				if got.Slug != p.Slug {
					t.Fatalf("slug: got %q, want %q", got.Slug, p.Slug)
				}
				if got.Name != p.Name {
					t.Fatalf("name: got %q, want %q", got.Name, p.Name)
				}
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestGetWorkspace(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (string, func(*testing.T))
		wantErr error
	}{
		{
			name: "returns existing workspace",
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				ws, err := CreateWorkspace(context.Background(), db, CreateWorkspaceParams{Name: "Acme", Slug: uniqueSlug(t, db)})
				if err != nil {
					t.Fatalf("seed workspace: %v", err)
				}
				return ws.ID, func(t *testing.T) {}
			},
		},
		{
			name:    "not found",
			wantErr: ErrWorkspaceNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				return "00000000-0000-0000-0000-000000000000", nil
			},
		},
		{
			name: "returns archived workspace",
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				ws, err := CreateWorkspace(context.Background(), db, CreateWorkspaceParams{Name: "Old", Slug: uniqueSlug(t, db)})
				if err != nil {
					t.Fatalf("seed workspace: %v", err)
				}
				if err := ArchiveWorkspace(context.Background(), db, ws.ID); err != nil {
					t.Fatalf("archive workspace: %v", err)
				}
				return ws.ID, func(t *testing.T) {}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, check := tt.arrange(t, db)
			got, err := GetWorkspace(context.Background(), db, id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetWorkspace() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil && got.ID != id {
				t.Fatalf("id: got %q, want %q", got.ID, id)
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestGetWorkspaceBySlug(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (string, func(*testing.T))
		wantErr error
	}{
		{
			name: "returns workspace by slug",
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				slug := uniqueSlug(t, db)
				ws, err := CreateWorkspace(context.Background(), db, CreateWorkspaceParams{Name: "Acme", Slug: slug})
				if err != nil {
					t.Fatalf("seed workspace: %v", err)
				}
				return slug, func(t *testing.T) {
					if ws.Slug != slug {
						t.Fatalf("slug: got %q, want %q", ws.Slug, slug)
					}
				}
			},
		},
		{
			name:    "not found",
			wantErr: ErrWorkspaceNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				return "slug-that-does-not-exist", nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slug, check := tt.arrange(t, db)
			got, err := GetWorkspaceBySlug(context.Background(), db, slug)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetWorkspaceBySlug() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil && got.Slug != slug {
				t.Fatalf("slug: got %q, want %q", got.Slug, slug)
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestArchiveWorkspace(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) string
		wantErr error
	}{
		{
			name: "archives active workspace",
			arrange: func(t *testing.T, db *sqlx.DB) string {
				ws, err := CreateWorkspace(context.Background(), db, CreateWorkspaceParams{Name: "Acme", Slug: uniqueSlug(t, db)})
				if err != nil {
					t.Fatalf("seed workspace: %v", err)
				}
				return ws.ID
			},
		},
		{
			name:    "not found",
			wantErr: ErrWorkspaceNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) string {
				return "00000000-0000-0000-0000-000000000000"
			},
		},
		{
			name:    "already archived",
			wantErr: ErrWorkspaceNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) string {
				ws, err := CreateWorkspace(context.Background(), db, CreateWorkspaceParams{Name: "Old", Slug: uniqueSlug(t, db)})
				if err != nil {
					t.Fatalf("seed workspace: %v", err)
				}
				if err := ArchiveWorkspace(context.Background(), db, ws.ID); err != nil {
					t.Fatalf("first archive: %v", err)
				}
				return ws.ID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.arrange(t, db)
			err := ArchiveWorkspace(context.Background(), db, id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ArchiveWorkspace() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil {
				got, err := GetWorkspace(context.Background(), db, id)
				if err != nil {
					t.Fatalf("get archived workspace: %v", err)
				}
				if got.ArchivedAt == nil {
					t.Fatal("expected archived_at to be set")
				}
			}
		})
	}
}

func openTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	if testDSN == "" {
		t.Skip("MINI_JIRA_TEST_DSN is not set; skipping PostgreSQL integration test")
	}
	db, err := sqlx.Connect("postgres", testDSN)
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func ensureSchema(t *testing.T, db *sqlx.DB) {
	t.Helper()

	requiredTables := []string{"workspaces", "app_users", "projects", "statuses", "issues"}

	existing := 0
	for _, table := range requiredTables {
		var exists *string
		if err := db.Get(&exists, `SELECT to_regclass('public.`+table+`')::text`); err != nil {
			t.Fatalf("check table %s exists: %v", table, err)
		}
		if exists != nil && *exists != "" {
			existing++
		}
	}

	if existing == len(requiredTables) {
		return
	}
	if existing > 0 {
		t.Fatalf("partial schema detected: found %d/%d required tables", existing, len(requiredTables))
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
	sqlBytes, err := os.ReadFile(filepath.Join(root, "migrations", "0001_init.up.sql"))
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	if _, err := db.Exec(string(sqlBytes)); err != nil {
		t.Fatalf("apply migration: %v", err)
	}
}

// uniqueSlug genera un slug único usando gen_random_uuid() de PostgreSQL.
// Registra limpieza del workspace creado vía ON DELETE CASCADE desde workspaces.
func uniqueSlug(t *testing.T, db *sqlx.DB) string {
	t.Helper()
	var slug string
	if err := db.GetContext(context.Background(), &slug,
		`SELECT 'ws-' || substr(replace(gen_random_uuid()::text, '-', ''), 1, 8)`,
	); err != nil {
		t.Fatalf("generate unique slug: %v", err)
	}
	return slug
}
