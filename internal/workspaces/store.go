package workspaces

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const selectCols = `id, name, slug, created_at, updated_at, archived_at`
const memberCols = `workspace_id, user_id, role, created_at, updated_at, archived_at`

func createWorkspace(ctx context.Context, db *sqlx.DB, params CreateWorkspaceParams) (Workspace, error) {
	var workspace Workspace
	err := db.QueryRowxContext(
		ctx,
		`INSERT INTO workspaces (name, slug)
		 VALUES ($1, $2)
		 RETURNING `+selectCols,
		params.Name,
		params.Slug,
	).StructScan(&workspace)
	if err != nil {
		if isUniqueViolation(err) {
			return Workspace{}, ErrDuplicateSlug
		}
		return Workspace{}, fmt.Errorf("insert workspace: %w", err)
	}
	return workspace, nil
}

func getWorkspace(ctx context.Context, db *sqlx.DB, id string) (Workspace, error) {
	var workspace Workspace
	err := db.GetContext(
		ctx,
		&workspace,
		`SELECT `+selectCols+`
		 FROM workspaces
		 WHERE id = $1`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Workspace{}, ErrWorkspaceNotFound
		}
		return Workspace{}, fmt.Errorf("get workspace: %w", err)
	}
	return workspace, nil
}

func getWorkspaceBySlug(ctx context.Context, db *sqlx.DB, slug string) (Workspace, error) {
	var workspace Workspace
	err := db.GetContext(
		ctx,
		&workspace,
		`SELECT `+selectCols+`
		 FROM workspaces
		 WHERE slug = $1`,
		slug,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Workspace{}, ErrWorkspaceNotFound
		}
		return Workspace{}, fmt.Errorf("get workspace by slug: %w", err)
	}
	return workspace, nil
}

func archiveWorkspace(ctx context.Context, db *sqlx.DB, id string) error {
	res, err := db.ExecContext(
		ctx,
		`UPDATE workspaces
		 SET archived_at = NOW()
		 WHERE id = $1
		   AND archived_at IS NULL`,
		id,
	)
	if err != nil {
		return fmt.Errorf("archive workspace: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive workspace rows affected: %w", err)
	}
	if n == 0 {
		return ErrWorkspaceNotFound
	}
	return nil
}

func addMember(ctx context.Context, db *sqlx.DB, params AddMemberParams) (WorkspaceMember, error) {
	var member WorkspaceMember
	err := db.QueryRowxContext(ctx,
		`INSERT INTO workspace_members (workspace_id, user_id, role)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (workspace_id, user_id)
		 DO UPDATE SET role = excluded.role, archived_at = NULL
		 RETURNING `+memberCols,
		params.WorkspaceID, params.UserID, params.Role,
	).StructScan(&member)
	if err != nil {
		return WorkspaceMember{}, fmt.Errorf("add workspace member: %w", err)
	}
	return member, nil
}

func removeMember(ctx context.Context, db *sqlx.DB, workspaceID, userID string) error {
	res, err := db.ExecContext(ctx,
		`UPDATE workspace_members
		 SET archived_at = NOW()
		 WHERE workspace_id = $1 AND user_id = $2 AND archived_at IS NULL`,
		workspaceID, userID,
	)
	if err != nil {
		return fmt.Errorf("remove workspace member: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("remove workspace member rows affected: %w", err)
	}
	if n == 0 {
		return ErrMemberNotFound
	}
	return nil
}

func listMembers(ctx context.Context, db *sqlx.DB, workspaceID string) ([]WorkspaceMember, error) {
	var members []WorkspaceMember
	err := db.SelectContext(ctx, &members,
		`SELECT `+memberCols+`
		 FROM workspace_members
		 WHERE workspace_id = $1 AND archived_at IS NULL
		 ORDER BY created_at ASC`,
		workspaceID,
	)
	if err != nil {
		return nil, fmt.Errorf("list workspace members: %w", err)
	}
	return members, nil
}

func updateMemberRole(ctx context.Context, db *sqlx.DB, params UpdateMemberRoleParams) (WorkspaceMember, error) {
	var member WorkspaceMember
	err := db.QueryRowxContext(ctx,
		`UPDATE workspace_members
		 SET role = $1
		 WHERE workspace_id = $2 AND user_id = $3 AND archived_at IS NULL
		 RETURNING `+memberCols,
		params.Role, params.WorkspaceID, params.UserID,
	).StructScan(&member)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WorkspaceMember{}, ErrMemberNotFound
		}
		return WorkspaceMember{}, fmt.Errorf("update workspace member role: %w", err)
	}
	return member, nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
