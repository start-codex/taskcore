package authz

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func workspaceExists(ctx context.Context, db *sqlx.DB, workspaceID string) (bool, error) {
	var exists bool
	err := db.GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM workspaces WHERE id = $1 AND archived_at IS NULL)`,
		workspaceID,
	)
	if err != nil {
		return false, fmt.Errorf("check workspace exists: %w", err)
	}
	return exists, nil
}

func isMember(ctx context.Context, db *sqlx.DB, workspaceID, userID string) (bool, error) {
	var exists bool
	err := db.GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM workspace_members WHERE workspace_id = $1 AND user_id = $2 AND archived_at IS NULL)`,
		workspaceID, userID,
	)
	if err != nil {
		return false, fmt.Errorf("check workspace membership: %w", err)
	}
	return exists, nil
}

func projectWorkspaceID(ctx context.Context, db *sqlx.DB, projectID string) (string, error) {
	var wsID string
	err := db.GetContext(ctx, &wsID,
		`SELECT workspace_id FROM projects WHERE id = $1`,
		projectID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrProjectNotFound
		}
		return "", fmt.Errorf("resolve project workspace: %w", err)
	}
	return wsID, nil
}

func boardProjectID(ctx context.Context, db *sqlx.DB, boardID string) (string, error) {
	var projID string
	err := db.GetContext(ctx, &projID,
		`SELECT project_id FROM boards WHERE id = $1`,
		boardID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrBoardNotFound
		}
		return "", fmt.Errorf("resolve board project: %w", err)
	}
	return projID, nil
}

func columnBoardID(ctx context.Context, db *sqlx.DB, columnID string) (string, error) {
	var boardID string
	err := db.GetContext(ctx, &boardID,
		`SELECT board_id FROM board_columns WHERE id = $1`,
		columnID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrColumnNotFound
		}
		return "", fmt.Errorf("resolve column board: %w", err)
	}
	return boardID, nil
}
