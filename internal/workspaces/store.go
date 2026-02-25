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

func createWorkspace(ctx context.Context, db *sqlx.DB, p CreateWorkspaceParams) (Workspace, error) {
	var out Workspace
	err := db.QueryRowxContext(
		ctx,
		`INSERT INTO workspaces (name, slug)
		 VALUES ($1, $2)
		 RETURNING `+selectCols,
		p.Name,
		p.Slug,
	).StructScan(&out)
	if err != nil {
		if isUniqueViolation(err) {
			return Workspace{}, ErrDuplicateSlug
		}
		return Workspace{}, fmt.Errorf("insert workspace: %w", err)
	}
	return out, nil
}

func getWorkspace(ctx context.Context, db *sqlx.DB, id string) (Workspace, error) {
	var out Workspace
	err := db.GetContext(
		ctx,
		&out,
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
	return out, nil
}

func getWorkspaceBySlug(ctx context.Context, db *sqlx.DB, slug string) (Workspace, error) {
	var out Workspace
	err := db.GetContext(
		ctx,
		&out,
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
	return out, nil
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

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
