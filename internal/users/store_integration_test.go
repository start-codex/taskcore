package users

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var testDSN = os.Getenv("MINI_JIRA_TEST_DSN")

func TestCreateUser(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (CreateUserParams, func(*testing.T))
		wantErr error
	}{
		{
			name: "creates user successfully",
			arrange: func(t *testing.T, db *sqlx.DB) (CreateUserParams, func(*testing.T)) {
				p := CreateUserParams{Email: uniqueEmail(t, db), Name: "Alice"}
				return p, func(t *testing.T) {}
			},
		},
		{
			name:    "duplicate email",
			wantErr: ErrDuplicateEmail,
			arrange: func(t *testing.T, db *sqlx.DB) (CreateUserParams, func(*testing.T)) {
				email := uniqueEmail(t, db)
				if _, err := CreateUser(context.Background(), db, CreateUserParams{Email: email, Name: "First"}); err != nil {
					t.Fatalf("seed user: %v", err)
				}
				return CreateUserParams{Email: email, Name: "Second"}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, check := tt.arrange(t, db)
			got, err := CreateUser(context.Background(), db, p)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateUser() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil {
				if got.ID == "" {
					t.Fatal("expected non-empty id")
				}
				if got.Email != p.Email {
					t.Fatalf("email: got %q, want %q", got.Email, p.Email)
				}
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (string, func(*testing.T))
		wantErr error
	}{
		{
			name: "returns existing user",
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				u := seedUser(t, db)
				return u.ID, func(t *testing.T) {}
			},
		},
		{
			name:    "not found",
			wantErr: ErrUserNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				return "00000000-0000-0000-0000-000000000000", nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, check := tt.arrange(t, db)
			got, err := GetUser(context.Background(), db, id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetUser() error = %v, wantErr = %v", err, tt.wantErr)
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

func TestGetUserByEmail(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (string, func(*testing.T))
		wantErr error
	}{
		{
			name: "returns user by email",
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				u := seedUser(t, db)
				return u.Email, func(t *testing.T) {}
			},
		},
		{
			name:    "not found",
			wantErr: ErrUserNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				return "nobody@does-not-exist.local", nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, check := tt.arrange(t, db)
			got, err := GetUserByEmail(context.Background(), db, email)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetUserByEmail() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil && got.Email != email {
				t.Fatalf("email: got %q, want %q", got.Email, email)
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestArchiveUser(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) string
		wantErr error
	}{
		{
			name: "archives active user",
			arrange: func(t *testing.T, db *sqlx.DB) string {
				return seedUser(t, db).ID
			},
		},
		{
			name:    "not found",
			wantErr: ErrUserNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) string {
				return "00000000-0000-0000-0000-000000000000"
			},
		},
		{
			name:    "already archived",
			wantErr: ErrUserNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) string {
				u := seedUser(t, db)
				if err := ArchiveUser(context.Background(), db, u.ID); err != nil {
					t.Fatalf("first archive: %v", err)
				}
				return u.ID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.arrange(t, db)
			err := ArchiveUser(context.Background(), db, id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ArchiveUser() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestWorkspaceMembers(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) func(*testing.T)
		wantErr error
	}{
		{
			name: "add and list member",
			arrange: func(t *testing.T, db *sqlx.DB) func(*testing.T) {
				ws := seedWorkspace(t, db)
				u := seedUser(t, db)
				if _, err := AddWorkspaceMember(context.Background(), db, AddWorkspaceMemberParams{WorkspaceID: ws, UserID: u.ID, Role: "member"}); err != nil {
					t.Fatalf("add member: %v", err)
				}
				return func(t *testing.T) {
					members, err := ListWorkspaceMembers(context.Background(), db, ws)
					if err != nil {
						t.Fatalf("list: %v", err)
					}
					if len(members) != 1 || members[0].UserID != u.ID {
						t.Fatalf("expected 1 member with id %q", u.ID)
					}
				}
			},
		},
		{
			name: "add same member twice updates role",
			arrange: func(t *testing.T, db *sqlx.DB) func(*testing.T) {
				ws := seedWorkspace(t, db)
				u := seedUser(t, db)
				if _, err := AddWorkspaceMember(context.Background(), db, AddWorkspaceMemberParams{WorkspaceID: ws, UserID: u.ID, Role: "member"}); err != nil {
					t.Fatalf("first add: %v", err)
				}
				if _, err := AddWorkspaceMember(context.Background(), db, AddWorkspaceMemberParams{WorkspaceID: ws, UserID: u.ID, Role: "admin"}); err != nil {
					t.Fatalf("second add: %v", err)
				}
				return func(t *testing.T) {
					members, err := ListWorkspaceMembers(context.Background(), db, ws)
					if err != nil {
						t.Fatalf("list: %v", err)
					}
					if len(members) != 1 {
						t.Fatalf("len: got %d, want 1", len(members))
					}
					if members[0].Role != "admin" {
						t.Fatalf("role: got %q, want admin", members[0].Role)
					}
				}
			},
		},
		{
			name: "remove member excludes from list",
			arrange: func(t *testing.T, db *sqlx.DB) func(*testing.T) {
				ws := seedWorkspace(t, db)
				u := seedUser(t, db)
				if _, err := AddWorkspaceMember(context.Background(), db, AddWorkspaceMemberParams{WorkspaceID: ws, UserID: u.ID, Role: "member"}); err != nil {
					t.Fatalf("add: %v", err)
				}
				if err := RemoveWorkspaceMember(context.Background(), db, ws, u.ID); err != nil {
					t.Fatalf("remove: %v", err)
				}
				return func(t *testing.T) {
					members, err := ListWorkspaceMembers(context.Background(), db, ws)
					if err != nil {
						t.Fatalf("list: %v", err)
					}
					if len(members) != 0 {
						t.Fatalf("len: got %d, want 0", len(members))
					}
				}
			},
		},
		{
			name: "re-add removed member restores access",
			arrange: func(t *testing.T, db *sqlx.DB) func(*testing.T) {
				ws := seedWorkspace(t, db)
				u := seedUser(t, db)
				if _, err := AddWorkspaceMember(context.Background(), db, AddWorkspaceMemberParams{WorkspaceID: ws, UserID: u.ID, Role: "member"}); err != nil {
					t.Fatalf("add: %v", err)
				}
				if err := RemoveWorkspaceMember(context.Background(), db, ws, u.ID); err != nil {
					t.Fatalf("remove: %v", err)
				}
				if _, err := AddWorkspaceMember(context.Background(), db, AddWorkspaceMemberParams{WorkspaceID: ws, UserID: u.ID, Role: "admin"}); err != nil {
					t.Fatalf("re-add: %v", err)
				}
				return func(t *testing.T) {
					members, err := ListWorkspaceMembers(context.Background(), db, ws)
					if err != nil {
						t.Fatalf("list: %v", err)
					}
					if len(members) != 1 || members[0].Role != "admin" {
						t.Fatalf("expected 1 member with role admin")
					}
				}
			},
		},
		{
			name: "update workspace member role",
			arrange: func(t *testing.T, db *sqlx.DB) func(*testing.T) {
				ws := seedWorkspace(t, db)
				u := seedUser(t, db)
				if _, err := AddWorkspaceMember(context.Background(), db, AddWorkspaceMemberParams{WorkspaceID: ws, UserID: u.ID, Role: "member"}); err != nil {
					t.Fatalf("add: %v", err)
				}
				got, err := UpdateWorkspaceMemberRole(context.Background(), db, UpdateWorkspaceMemberRoleParams{WorkspaceID: ws, UserID: u.ID, Role: "admin"})
				if err != nil {
					t.Fatalf("update role: %v", err)
				}
				return func(t *testing.T) {
					if got.Role != "admin" {
						t.Fatalf("role: got %q, want admin", got.Role)
					}
				}
			},
		},
		{
			name:    "remove non-existent member",
			wantErr: ErrMemberNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) func(*testing.T) {
				ws := seedWorkspace(t, db)
				err := RemoveWorkspaceMember(context.Background(), db, ws, "00000000-0000-0000-0000-000000000000")
				if !errors.Is(err, ErrMemberNotFound) {
					t.Fatalf("RemoveWorkspaceMember() error = %v, wantErr = %v", err, ErrMemberNotFound)
				}
				return nil
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
				u := seedUser(t, db)
				if _, err := AddProjectMember(context.Background(), db, AddProjectMemberParams{ProjectID: proj, UserID: u.ID, Role: "member"}); err != nil {
					t.Fatalf("add: %v", err)
				}
				return func(t *testing.T) {
					members, err := ListProjectMembers(context.Background(), db, proj)
					if err != nil {
						t.Fatalf("list: %v", err)
					}
					if len(members) != 1 || members[0].UserID != u.ID {
						t.Fatalf("expected 1 member with id %q", u.ID)
					}
				}
			},
		},
		{
			name: "update project member role",
			arrange: func(t *testing.T, db *sqlx.DB) func(*testing.T) {
				proj := seedProject(t, db)
				u := seedUser(t, db)
				if _, err := AddProjectMember(context.Background(), db, AddProjectMemberParams{ProjectID: proj, UserID: u.ID, Role: "viewer"}); err != nil {
					t.Fatalf("add: %v", err)
				}
				got, err := UpdateProjectMemberRole(context.Background(), db, UpdateProjectMemberRoleParams{ProjectID: proj, UserID: u.ID, Role: "admin"})
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
				u := seedUser(t, db)
				if _, err := AddProjectMember(context.Background(), db, AddProjectMemberParams{ProjectID: proj, UserID: u.ID, Role: "viewer"}); err != nil {
					t.Fatalf("add: %v", err)
				}
				if err := RemoveProjectMember(context.Background(), db, proj, u.ID); err != nil {
					t.Fatalf("remove: %v", err)
				}
				if _, err := AddProjectMember(context.Background(), db, AddProjectMemberParams{ProjectID: proj, UserID: u.ID, Role: "admin"}); err != nil {
					t.Fatalf("re-add: %v", err)
				}
				return func(t *testing.T) {
					members, err := ListProjectMembers(context.Background(), db, proj)
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

// --- helpers ---

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
			t.Fatalf("check table %s: %v", table, err)
		}
		if exists != nil && *exists != "" {
			existing++
		}
	}
	if existing == len(requiredTables) {
		return
	}
	if existing > 0 {
		t.Fatalf("partial schema: found %d/%d tables", existing, len(requiredTables))
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

func uniqueEmail(t *testing.T, db *sqlx.DB) string {
	t.Helper()
	var suffix string
	if err := db.GetContext(context.Background(), &suffix,
		`SELECT substr(replace(gen_random_uuid()::text, '-', ''), 1, 8)`,
	); err != nil {
		t.Fatalf("generate unique email: %v", err)
	}
	return fmt.Sprintf("user-%s@test.local", suffix)
}

func seedUser(t *testing.T, db *sqlx.DB) User {
	t.Helper()
	u, err := CreateUser(context.Background(), db, CreateUserParams{
		Email: uniqueEmail(t, db),
		Name:  "Test User",
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), `DELETE FROM app_users WHERE id = $1`, u.ID); err != nil {
			t.Logf("cleanup user %s: %v", u.ID, err)
		}
	})
	return u
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
			t.Logf("cleanup workspace %s: %v", id, err)
		}
	})
	return id
}

func seedProject(t *testing.T, db *sqlx.DB) string {
	t.Helper()
	ws := seedWorkspace(t, db)
	var id string
	if err := db.GetContext(context.Background(), &id,
		`INSERT INTO projects (workspace_id, name, key, description) VALUES ($1, 'Project', upper(substr(replace(gen_random_uuid()::text,'-',''),1,3)), '') RETURNING id`,
		ws,
	); err != nil {
		t.Fatalf("seed project: %v", err)
	}
	return id
}
