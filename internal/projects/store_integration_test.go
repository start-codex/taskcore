package projects

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

func TestCreateProject(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (CreateProjectParams, func(*testing.T))
		wantErr error
	}{
		{
			name: "creates project successfully",
			arrange: func(t *testing.T, db *sqlx.DB) (CreateProjectParams, func(*testing.T)) {
				ws := seedWorkspace(t, db)
				params := CreateProjectParams{WorkspaceID: ws, Name: "Engineering", Key: "ENG", Description: "eng team"}
				return params, func(t *testing.T) {}
			},
		},
		{
			name: "returned project has correct fields",
			arrange: func(t *testing.T, db *sqlx.DB) (CreateProjectParams, func(*testing.T)) {
				ws := seedWorkspace(t, db)
				params := CreateProjectParams{WorkspaceID: ws, Name: "Marketing", Key: "MKT", Description: "mkt team"}
				return params, func(t *testing.T) {}
			},
		},
		{
			name:    "duplicate key in same workspace",
			wantErr: ErrDuplicateProjectKey,
			arrange: func(t *testing.T, db *sqlx.DB) (CreateProjectParams, func(*testing.T)) {
				ws := seedWorkspace(t, db)
				_, err := CreateProject(context.Background(), db, CreateProjectParams{WorkspaceID: ws, Name: "First", Key: "DUP"})
				if err != nil {
					t.Fatalf("seed project: %v", err)
				}
				return CreateProjectParams{WorkspaceID: ws, Name: "Second", Key: "DUP"}, nil
			},
		},
		{
			name: "same key in different workspaces is allowed",
			arrange: func(t *testing.T, db *sqlx.DB) (CreateProjectParams, func(*testing.T)) {
				ws1 := seedWorkspace(t, db)
				ws2 := seedWorkspace(t, db)
				_, err := CreateProject(context.Background(), db, CreateProjectParams{WorkspaceID: ws1, Name: "First", Key: "ENG"})
				if err != nil {
					t.Fatalf("seed project ws1: %v", err)
				}
				return CreateProjectParams{WorkspaceID: ws2, Name: "Second", Key: "ENG"}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, check := tt.arrange(t, db)
			got, err := CreateProject(context.Background(), db, params)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateProject() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil {
				if got.ID == "" {
					t.Fatal("expected non-empty id")
				}
				if got.Key != params.Key {
					t.Fatalf("key: got %q, want %q", got.Key, params.Key)
				}
				if got.WorkspaceID != params.WorkspaceID {
					t.Fatalf("workspace_id: got %q, want %q", got.WorkspaceID, params.WorkspaceID)
				}
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestGetProject(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (string, func(*testing.T))
		wantErr error
	}{
		{
			name: "returns existing project",
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				ws := seedWorkspace(t, db)
				proj, err := CreateProject(context.Background(), db, CreateProjectParams{WorkspaceID: ws, Name: "Engineering", Key: "ENG"})
				if err != nil {
					t.Fatalf("seed project: %v", err)
				}
				return proj.ID, func(t *testing.T) {}
			},
		},
		{
			name:    "not found",
			wantErr: ErrProjectNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				return "00000000-0000-0000-0000-000000000000", nil
			},
		},
		{
			name: "returns archived project",
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				ws := seedWorkspace(t, db)
				proj, err := CreateProject(context.Background(), db, CreateProjectParams{WorkspaceID: ws, Name: "Old", Key: "OLD"})
				if err != nil {
					t.Fatalf("seed project: %v", err)
				}
				if err := ArchiveProject(context.Background(), db, proj.ID); err != nil {
					t.Fatalf("archive project: %v", err)
				}
				return proj.ID, func(t *testing.T) {}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, check := tt.arrange(t, db)
			got, err := GetProject(context.Background(), db, id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetProject() error = %v, wantErr = %v", err, tt.wantErr)
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

func TestListProjects(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (string, func(*testing.T, []Project))
		wantErr error
	}{
		{
			name: "returns only active projects",
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T, []Project)) {
				ws := seedWorkspace(t, db)
				active, err := CreateProject(context.Background(), db, CreateProjectParams{WorkspaceID: ws, Name: "Active", Key: "ACT"})
				if err != nil {
					t.Fatalf("seed active project: %v", err)
				}
				archived, err := CreateProject(context.Background(), db, CreateProjectParams{WorkspaceID: ws, Name: "Archived", Key: "ARC"})
				if err != nil {
					t.Fatalf("seed archived project: %v", err)
				}
				if err := ArchiveProject(context.Background(), db, archived.ID); err != nil {
					t.Fatalf("archive project: %v", err)
				}
				return ws, func(t *testing.T, got []Project) {
					if len(got) != 1 {
						t.Fatalf("len: got %d, want 1", len(got))
					}
					if got[0].ID != active.ID {
						t.Fatalf("id: got %q, want %q", got[0].ID, active.ID)
					}
				}
			},
		},
		{
			name: "empty workspace returns empty slice",
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T, []Project)) {
				ws := seedWorkspace(t, db)
				return ws, func(t *testing.T, got []Project) {
					if len(got) != 0 {
						t.Fatalf("len: got %d, want 0", len(got))
					}
				}
			},
		},
		{
			name: "does not return projects from other workspaces",
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T, []Project)) {
				ws1 := seedWorkspace(t, db)
				ws2 := seedWorkspace(t, db)
				if _, err := CreateProject(context.Background(), db, CreateProjectParams{WorkspaceID: ws2, Name: "Other", Key: "OTH"}); err != nil {
					t.Fatalf("seed other project: %v", err)
				}
				return ws1, func(t *testing.T, got []Project) {
					if len(got) != 0 {
						t.Fatalf("len: got %d, want 0", len(got))
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wsID, check := tt.arrange(t, db)
			got, err := ListProjects(context.Background(), db, wsID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ListProjects() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if check != nil {
				check(t, got)
			}
		})
	}
}

func TestArchiveProject(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) string
		wantErr error
	}{
		{
			name: "archives active project",
			arrange: func(t *testing.T, db *sqlx.DB) string {
				ws := seedWorkspace(t, db)
				proj, err := CreateProject(context.Background(), db, CreateProjectParams{WorkspaceID: ws, Name: "Engineering", Key: "ENG"})
				if err != nil {
					t.Fatalf("seed project: %v", err)
				}
				return proj.ID
			},
		},
		{
			name:    "not found",
			wantErr: ErrProjectNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) string {
				return "00000000-0000-0000-0000-000000000000"
			},
		},
		{
			name:    "already archived",
			wantErr: ErrProjectNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) string {
				ws := seedWorkspace(t, db)
				proj, err := CreateProject(context.Background(), db, CreateProjectParams{WorkspaceID: ws, Name: "Old", Key: "OLD"})
				if err != nil {
					t.Fatalf("seed project: %v", err)
				}
				if err := ArchiveProject(context.Background(), db, proj.ID); err != nil {
					t.Fatalf("first archive: %v", err)
				}
				return proj.ID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.arrange(t, db)
			err := ArchiveProject(context.Background(), db, id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ArchiveProject() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil {
				got, err := GetProject(context.Background(), db, id)
				if err != nil {
					t.Fatalf("get archived project: %v", err)
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

func seedUser(t *testing.T, db *sqlx.DB) string {
	t.Helper()
	var id string
	if err := db.GetContext(context.Background(), &id,
		`INSERT INTO app_users (email, name, password_hash)
		 VALUES (gen_random_uuid()::text || '@test.local', 'Test', '') RETURNING id`,
	); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), `DELETE FROM app_users WHERE id = $1`, id)
	})
	return id
}

func TestProjectMembers(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) func(*testing.T)
	}{
		{
			name: "add and list member",
			arrange: func(t *testing.T, db *sqlx.DB) func(*testing.T) {
				proj := seedProject(t, db)
				uID := seedUser(t, db)
				if _, err := AddMember(context.Background(), db, AddMemberParams{ProjectID: proj, UserID: uID, Role: "member"}); err != nil {
					t.Fatalf("add: %v", err)
				}
				return func(t *testing.T) {
					members, err := ListMembers(context.Background(), db, proj)
					if err != nil {
						t.Fatalf("list: %v", err)
					}
					if len(members) != 1 || members[0].UserID != uID {
						t.Fatalf("expected 1 member with id %q", uID)
					}
				}
			},
		},
		{
			name: "update project member role",
			arrange: func(t *testing.T, db *sqlx.DB) func(*testing.T) {
				proj := seedProject(t, db)
				uID := seedUser(t, db)
				if _, err := AddMember(context.Background(), db, AddMemberParams{ProjectID: proj, UserID: uID, Role: "viewer"}); err != nil {
					t.Fatalf("add: %v", err)
				}
				got, err := UpdateMemberRole(context.Background(), db, UpdateMemberRoleParams{ProjectID: proj, UserID: uID, Role: "admin"})
				if err != nil {
					t.Fatalf("update: %v", err)
				}
				return func(t *testing.T) {
					if got.Role != "admin" {
						t.Fatalf("role: got %q, want admin", got.Role)
					}
				}
			},
		},
		{
			name: "remove and re-add project member",
			arrange: func(t *testing.T, db *sqlx.DB) func(*testing.T) {
				proj := seedProject(t, db)
				uID := seedUser(t, db)
				if _, err := AddMember(context.Background(), db, AddMemberParams{ProjectID: proj, UserID: uID, Role: "viewer"}); err != nil {
					t.Fatalf("add: %v", err)
				}
				if err := RemoveMember(context.Background(), db, proj, uID); err != nil {
					t.Fatalf("remove: %v", err)
				}
				if _, err := AddMember(context.Background(), db, AddMemberParams{ProjectID: proj, UserID: uID, Role: "admin"}); err != nil {
					t.Fatalf("re-add: %v", err)
				}
				return func(t *testing.T) {
					members, err := ListMembers(context.Background(), db, proj)
					if err != nil {
						t.Fatalf("list: %v", err)
					}
					if len(members) != 1 || members[0].Role != "admin" {
						t.Fatalf("expected 1 member with role admin")
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := tt.arrange(t, db)
			if check != nil {
				check(t)
			}
		})
	}
}

func seedWorkspace(t *testing.T, db *sqlx.DB) string {
	t.Helper()
	var id string
	if err := db.GetContext(context.Background(), &id,
		`INSERT INTO workspaces (name, slug) VALUES ('ws', gen_random_uuid()::text) RETURNING id`,
	); err != nil {
		t.Fatalf("seed workspace: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), `DELETE FROM workspaces WHERE id = $1`, id); err != nil {
			t.Fatalf("cleanup workspace: %v", err)
		}
	})
	return id
}

func seedProject(t *testing.T, db *sqlx.DB) string {
	t.Helper()
	ws := seedWorkspace(t, db)
	var id string
	if err := db.GetContext(context.Background(), &id,
		`INSERT INTO projects (workspace_id, name, key, description)
		 VALUES ($1, 'Project', upper(substr(replace(gen_random_uuid()::text,'-',''),1,3)), '') RETURNING id`,
		ws,
	); err != nil {
		t.Fatalf("seed project: %v", err)
	}
	return id
}
