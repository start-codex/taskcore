package boards

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

func TestCreateBoard(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (CreateBoardParams, func(*testing.T))
		wantErr error
	}{
		{
			name: "creates kanban board",
			arrange: func(t *testing.T, db *sqlx.DB) (CreateBoardParams, func(*testing.T)) {
				proj := seedProject(t, db)
				p := CreateBoardParams{ProjectID: proj, Name: "Main Board", Type: "kanban"}
				return p, func(t *testing.T) {}
			},
		},
		{
			name: "creates scrum board",
			arrange: func(t *testing.T, db *sqlx.DB) (CreateBoardParams, func(*testing.T)) {
				proj := seedProject(t, db)
				p := CreateBoardParams{ProjectID: proj, Name: "Sprint Board", Type: "scrum", FilterQuery: "type=story"}
				return p, func(t *testing.T) {}
			},
		},
		{
			name:    "duplicate name in same project",
			wantErr: ErrDuplicateBoardName,
			arrange: func(t *testing.T, db *sqlx.DB) (CreateBoardParams, func(*testing.T)) {
				proj := seedProject(t, db)
				if _, err := CreateBoard(context.Background(), db, CreateBoardParams{ProjectID: proj, Name: "Dup", Type: "kanban"}); err != nil {
					t.Fatalf("seed board: %v", err)
				}
				return CreateBoardParams{ProjectID: proj, Name: "Dup", Type: "kanban"}, nil
			},
		},
		{
			name: "same name in different projects is allowed",
			arrange: func(t *testing.T, db *sqlx.DB) (CreateBoardParams, func(*testing.T)) {
				ws := seedWorkspace(t, db)
				proj1 := seedProjectInWorkspace(t, db, ws, "PRJ")
				proj2 := seedProjectInWorkspace(t, db, ws, "OTH")
				if _, err := CreateBoard(context.Background(), db, CreateBoardParams{ProjectID: proj1, Name: "Board", Type: "kanban"}); err != nil {
					t.Fatalf("seed board proj1: %v", err)
				}
				return CreateBoardParams{ProjectID: proj2, Name: "Board", Type: "kanban"}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, check := tt.arrange(t, db)
			got, err := CreateBoard(context.Background(), db, p)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateBoard() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil {
				if got.ID == "" {
					t.Fatal("expected non-empty id")
				}
				if got.ProjectID != p.ProjectID {
					t.Fatalf("project_id: got %q, want %q", got.ProjectID, p.ProjectID)
				}
				if got.Type != p.Type {
					t.Fatalf("type: got %q, want %q", got.Type, p.Type)
				}
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestGetBoard(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (string, func(*testing.T))
		wantErr error
	}{
		{
			name: "returns existing board",
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				proj := seedProject(t, db)
				board, err := CreateBoard(context.Background(), db, CreateBoardParams{ProjectID: proj, Name: "Board", Type: "kanban"})
				if err != nil {
					t.Fatalf("seed board: %v", err)
				}
				return board.ID, func(t *testing.T) {}
			},
		},
		{
			name:    "not found",
			wantErr: ErrBoardNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				return "00000000-0000-0000-0000-000000000000", nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, check := tt.arrange(t, db)
			got, err := GetBoard(context.Background(), db, id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetBoard() error = %v, wantErr = %v", err, tt.wantErr)
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

func TestListBoards(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (string, func(*testing.T, []Board))
		wantErr error
	}{
		{
			name: "returns only active boards",
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T, []Board)) {
				proj := seedProject(t, db)
				active, err := CreateBoard(context.Background(), db, CreateBoardParams{ProjectID: proj, Name: "Active", Type: "kanban"})
				if err != nil {
					t.Fatalf("seed active board: %v", err)
				}
				archived, err := CreateBoard(context.Background(), db, CreateBoardParams{ProjectID: proj, Name: "Archived", Type: "kanban"})
				if err != nil {
					t.Fatalf("seed archived board: %v", err)
				}
				if err := ArchiveBoard(context.Background(), db, archived.ID); err != nil {
					t.Fatalf("archive board: %v", err)
				}
				return proj, func(t *testing.T, got []Board) {
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
			name: "empty project returns empty slice",
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T, []Board)) {
				proj := seedProject(t, db)
				return proj, func(t *testing.T, got []Board) {
					if len(got) != 0 {
						t.Fatalf("len: got %d, want 0", len(got))
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projID, check := tt.arrange(t, db)
			got, err := ListBoards(context.Background(), db, projID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ListBoards() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if check != nil {
				check(t, got)
			}
		})
	}
}

func TestAddColumn(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (AddColumnParams, func(*testing.T))
		wantErr error
	}{
		{
			name: "adds first column at position 0",
			arrange: func(t *testing.T, db *sqlx.DB) (AddColumnParams, func(*testing.T)) {
				board := seedBoard(t, db)
				p := AddColumnParams{BoardID: board, Name: "To Do"}
				return p, func(t *testing.T) {}
			},
		},
		{
			name: "columns get sequential positions",
			arrange: func(t *testing.T, db *sqlx.DB) (AddColumnParams, func(*testing.T)) {
				board := seedBoard(t, db)
				if _, err := AddColumn(context.Background(), db, AddColumnParams{BoardID: board, Name: "To Do"}); err != nil {
					t.Fatalf("add first column: %v", err)
				}
				if _, err := AddColumn(context.Background(), db, AddColumnParams{BoardID: board, Name: "In Progress"}); err != nil {
					t.Fatalf("add second column: %v", err)
				}
				p := AddColumnParams{BoardID: board, Name: "Done"}
				return p, func(t *testing.T) {
					cols, err := ListColumns(context.Background(), db, board)
					if err != nil {
						t.Fatalf("list columns: %v", err)
					}
					if len(cols) != 3 {
						t.Fatalf("len: got %d, want 3", len(cols))
					}
					for i, col := range cols {
						if col.Position != i {
							t.Fatalf("col[%d].Position: got %d, want %d", i, col.Position, i)
						}
					}
				}
			},
		},
		{
			name:    "duplicate column name in same board",
			wantErr: ErrDuplicateColumnName,
			arrange: func(t *testing.T, db *sqlx.DB) (AddColumnParams, func(*testing.T)) {
				board := seedBoard(t, db)
				if _, err := AddColumn(context.Background(), db, AddColumnParams{BoardID: board, Name: "Dup"}); err != nil {
					t.Fatalf("seed column: %v", err)
				}
				return AddColumnParams{BoardID: board, Name: "Dup"}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, check := tt.arrange(t, db)
			got, err := AddColumn(context.Background(), db, p)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("AddColumn() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil && got.ID == "" {
				t.Fatal("expected non-empty id")
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestAssignUnassignStatus(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (boardColumnID, statusID string, check func(*testing.T))
		wantErr error
	}{
		{
			name: "assigns status to column",
			arrange: func(t *testing.T, db *sqlx.DB) (string, string, func(*testing.T)) {
				proj, status := seedProjectWithStatus(t, db)
				board, err := CreateBoard(context.Background(), db, CreateBoardParams{ProjectID: proj, Name: "Board", Type: "kanban"})
				if err != nil {
					t.Fatalf("seed board: %v", err)
				}
				col, err := AddColumn(context.Background(), db, AddColumnParams{BoardID: board.ID, Name: "To Do"})
				if err != nil {
					t.Fatalf("seed column: %v", err)
				}
				return col.ID, status, func(t *testing.T) {}
			},
		},
		{
			name: "assign same status twice is idempotent",
			arrange: func(t *testing.T, db *sqlx.DB) (string, string, func(*testing.T)) {
				proj, status := seedProjectWithStatus(t, db)
				board, err := CreateBoard(context.Background(), db, CreateBoardParams{ProjectID: proj, Name: "Board", Type: "kanban"})
				if err != nil {
					t.Fatalf("seed board: %v", err)
				}
				col, err := AddColumn(context.Background(), db, AddColumnParams{BoardID: board.ID, Name: "To Do"})
				if err != nil {
					t.Fatalf("seed column: %v", err)
				}
				if err := AssignStatus(context.Background(), db, col.ID, status); err != nil {
					t.Fatalf("first assign: %v", err)
				}
				return col.ID, status, func(t *testing.T) {}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			colID, statusID, check := tt.arrange(t, db)
			err := AssignStatus(context.Background(), db, colID, statusID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("AssignStatus() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil {
				if err := UnassignStatus(context.Background(), db, colID, statusID); err != nil {
					t.Fatalf("UnassignStatus() error = %v", err)
				}
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestArchiveColumn(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) string
		wantErr error
	}{
		{
			name: "archives active column",
			arrange: func(t *testing.T, db *sqlx.DB) string {
				board := seedBoard(t, db)
				col, err := AddColumn(context.Background(), db, AddColumnParams{BoardID: board, Name: "To Do"})
				if err != nil {
					t.Fatalf("seed column: %v", err)
				}
				return col.ID
			},
		},
		{
			name:    "not found",
			wantErr: ErrColumnNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) string {
				return "00000000-0000-0000-0000-000000000000"
			},
		},
		{
			name:    "already archived",
			wantErr: ErrColumnNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) string {
				board := seedBoard(t, db)
				col, err := AddColumn(context.Background(), db, AddColumnParams{BoardID: board, Name: "Old"})
				if err != nil {
					t.Fatalf("seed column: %v", err)
				}
				if err := ArchiveColumn(context.Background(), db, col.ID); err != nil {
					t.Fatalf("first archive: %v", err)
				}
				return col.ID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.arrange(t, db)
			err := ArchiveColumn(context.Background(), db, id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ArchiveColumn() error = %v, wantErr = %v", err, tt.wantErr)
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

func seedProjectInWorkspace(t *testing.T, db *sqlx.DB, workspaceID, key string) string {
	t.Helper()
	var id string
	if err := db.GetContext(context.Background(), &id,
		`INSERT INTO projects (workspace_id, name, key, description) VALUES ($1, $2, $3, '') RETURNING id`,
		workspaceID, key, key,
	); err != nil {
		t.Fatalf("seed project: %v", err)
	}
	return id
}

func seedProject(t *testing.T, db *sqlx.DB) string {
	t.Helper()
	ws := seedWorkspace(t, db)
	return seedProjectInWorkspace(t, db, ws, "BRD")
}

func seedBoard(t *testing.T, db *sqlx.DB) string {
	t.Helper()
	proj := seedProject(t, db)
	board, err := CreateBoard(context.Background(), db, CreateBoardParams{ProjectID: proj, Name: "Board", Type: "kanban"})
	if err != nil {
		t.Fatalf("seed board: %v", err)
	}
	return board.ID
}

func seedProjectWithStatus(t *testing.T, db *sqlx.DB) (projectID, statusID string) {
	t.Helper()
	proj := seedProject(t, db)
	if err := db.GetContext(context.Background(), &statusID,
		`INSERT INTO statuses (project_id, name, category, position) VALUES ($1, 'To Do', 'todo', 0) RETURNING id`,
		proj,
	); err != nil {
		t.Fatalf("seed status: %v", err)
	}
	return proj, statusID
}
