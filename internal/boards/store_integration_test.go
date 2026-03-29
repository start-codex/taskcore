// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package boards

import (
	"context"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/start-codex/trazawork/internal/testpg"
)

func TestCreateBoard(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (CreateBoardParams, func(*testing.T))
		wantErr error
	}{
		{
			name: "creates kanban board",
			arrange: func(t *testing.T, db *sqlx.DB) (CreateBoardParams, func(*testing.T)) {
				proj := seedProject(t, db)
				params := CreateBoardParams{ProjectID: proj, Name: "Main Board", Type: "kanban"}
				return params, func(t *testing.T) {}
			},
		},
		{
			name: "creates scrum board",
			arrange: func(t *testing.T, db *sqlx.DB) (CreateBoardParams, func(*testing.T)) {
				proj := seedProject(t, db)
				params := CreateBoardParams{ProjectID: proj, Name: "Sprint Board", Type: "scrum", FilterQuery: "type=story"}
				return params, func(t *testing.T) {}
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
				ws := testpg.SeedWorkspace(t, db)
				proj1 := testpg.SeedProject(t, db, ws, "PRJ")
				proj2 := testpg.SeedProject(t, db, ws, "OTH")
				if _, err := CreateBoard(context.Background(), db, CreateBoardParams{ProjectID: proj1, Name: "Board", Type: "kanban"}); err != nil {
					t.Fatalf("seed board proj1: %v", err)
				}
				return CreateBoardParams{ProjectID: proj2, Name: "Board", Type: "kanban"}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, check := tt.arrange(t, db)
			got, err := CreateBoard(context.Background(), db, params)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateBoard() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil {
				if got.ID == "" {
					t.Fatal("expected non-empty id")
				}
				if got.ProjectID != params.ProjectID {
					t.Fatalf("project_id: got %q, want %q", got.ProjectID, params.ProjectID)
				}
				if got.Type != params.Type {
					t.Fatalf("type: got %q, want %q", got.Type, params.Type)
				}
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestGetBoard(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

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
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

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
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (AddColumnParams, func(*testing.T))
		wantErr error
	}{
		{
			name: "adds first column at position 0",
			arrange: func(t *testing.T, db *sqlx.DB) (AddColumnParams, func(*testing.T)) {
				board := seedBoard(t, db)
				params := AddColumnParams{BoardID: board, Name: "To Do"}
				return params, func(t *testing.T) {}
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
				params := AddColumnParams{BoardID: board, Name: "Done"}
				return params, func(t *testing.T) {
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
			params, check := tt.arrange(t, db)
			got, err := AddColumn(context.Background(), db, params)
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
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

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
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

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

func seedProject(t *testing.T, db *sqlx.DB) string {
	t.Helper()
	ws := testpg.SeedWorkspace(t, db)
	return testpg.SeedProject(t, db, ws, "BRD")
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
