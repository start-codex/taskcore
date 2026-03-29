package authz

import (
	"context"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/start-codex/taskcode/internal/testpg"
)

func seedMember(t *testing.T, db *sqlx.DB, workspaceID, userID, role string) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, $3)`,
		workspaceID, userID, role,
	)
	if err != nil {
		t.Fatalf("seed member: %v", err)
	}
}

func seedBoard(t *testing.T, db *sqlx.DB, projectID string) string {
	t.Helper()
	var id string
	err := db.QueryRowContext(context.Background(),
		`INSERT INTO boards (project_id, name, type, filter_query) VALUES ($1, $2, 'kanban', '') RETURNING id`,
		projectID, "Board "+testpg.UniqueSuffix(t, db),
	).Scan(&id)
	if err != nil {
		t.Fatalf("seed board: %v", err)
	}
	return id
}

func seedColumn(t *testing.T, db *sqlx.DB, boardID string) string {
	t.Helper()
	var id string
	err := db.QueryRowContext(context.Background(),
		`INSERT INTO board_columns (board_id, name, position) VALUES ($1, $2, 0) RETURNING id`,
		boardID, "Column "+testpg.UniqueSuffix(t, db),
	).Scan(&id)
	if err != nil {
		t.Fatalf("seed column: %v", err)
	}
	return id
}

func TestRequireWorkspaceMembership_Integration(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	member := testpg.SeedUser(t, db)
	nonMember := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, member, "member")

	tests := []struct {
		name    string
		userID  string
		wsID    string
		wantErr error
	}{
		{name: "member ok", userID: member, wsID: wsID},
		{name: "non-member forbidden", userID: nonMember, wsID: wsID, wantErr: ErrForbidden},
		{name: "workspace not found", userID: member, wsID: "00000000-0000-0000-0000-000000000000", wantErr: ErrWorkspaceNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithUserID(context.Background(), tt.userID)
			err := RequireWorkspaceMembership(ctx, db, tt.wsID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestRequireWorkspaceMembership_ArchivedWorkspace(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	userID := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, userID, "member")

	// Archive workspace
	db.ExecContext(context.Background(), `UPDATE workspaces SET archived_at = NOW() WHERE id = $1`, wsID)
	t.Cleanup(func() {
		db.ExecContext(context.Background(), `UPDATE workspaces SET archived_at = NULL WHERE id = $1`, wsID)
	})

	ctx := WithUserID(context.Background(), userID)
	err := RequireWorkspaceMembership(ctx, db, wsID)
	if !errors.Is(err, ErrWorkspaceNotFound) {
		t.Fatalf("error = %v, want ErrWorkspaceNotFound", err)
	}
}

func TestRequireProjectMembership_Integration(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	member := testpg.SeedUser(t, db)
	nonMember := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, member, "member")
	projID := testpg.SeedProject(t, db, wsID, "AUTHZ")

	tests := []struct {
		name    string
		userID  string
		projID  string
		wantWS  string
		wantErr error
	}{
		{name: "member ok", userID: member, projID: projID, wantWS: wsID},
		{name: "non-member forbidden", userID: nonMember, projID: projID, wantErr: ErrForbidden},
		{name: "project not found", userID: member, projID: "00000000-0000-0000-0000-000000000000", wantErr: ErrProjectNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithUserID(context.Background(), tt.userID)
			wsID, err := RequireProjectMembership(ctx, db, tt.projID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if err == nil && wsID != tt.wantWS {
				t.Fatalf("workspaceID = %q, want %q", wsID, tt.wantWS)
			}
		})
	}
}

func TestRequireBoardAccess_Integration(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	member := testpg.SeedUser(t, db)
	nonMember := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, member, "member")
	projID := testpg.SeedProject(t, db, wsID, "BRDAZ")
	boardID := seedBoard(t, db, projID)

	tests := []struct {
		name     string
		userID   string
		boardID  string
		wantWS   string
		wantProj string
		wantErr  error
	}{
		{name: "member ok", userID: member, boardID: boardID, wantWS: wsID, wantProj: projID},
		{name: "non-member forbidden", userID: nonMember, boardID: boardID, wantErr: ErrForbidden},
		{name: "board not found", userID: member, boardID: "00000000-0000-0000-0000-000000000000", wantErr: ErrBoardNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithUserID(context.Background(), tt.userID)
			ws, proj, err := RequireBoardAccess(ctx, db, tt.boardID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if err == nil {
				if ws != tt.wantWS {
					t.Fatalf("workspaceID = %q, want %q", ws, tt.wantWS)
				}
				if proj != tt.wantProj {
					t.Fatalf("projectID = %q, want %q", proj, tt.wantProj)
				}
			}
		})
	}
}

func TestRequireColumnAccess_Integration(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	member := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, member, "member")
	projID := testpg.SeedProject(t, db, wsID, "COLAZ")
	boardID := seedBoard(t, db, projID)
	colID := seedColumn(t, db, boardID)

	ctx := WithUserID(context.Background(), member)
	ws, proj, brd, err := RequireColumnAccess(ctx, db, colID)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if ws != wsID {
		t.Fatalf("workspaceID = %q, want %q", ws, wsID)
	}
	if proj != projID {
		t.Fatalf("projectID = %q, want %q", proj, projID)
	}
	if brd != boardID {
		t.Fatalf("boardID = %q, want %q", brd, boardID)
	}

	// Column not found
	_, _, _, err = RequireColumnAccess(ctx, db, "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, ErrColumnNotFound) {
		t.Fatalf("error = %v, want ErrColumnNotFound", err)
	}
}
