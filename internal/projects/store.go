package projects

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const selectCols = `id, workspace_id, name, key, description, created_at, updated_at, archived_at`

func createProject(ctx context.Context, db *sqlx.DB, p CreateProjectParams) (Project, error) {
	var out Project
	err := db.QueryRowxContext(
		ctx,
		`INSERT INTO projects (workspace_id, name, key, description)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+selectCols,
		p.WorkspaceID,
		p.Name,
		p.Key,
		p.Description,
	).StructScan(&out)
	if err != nil {
		if isUniqueViolation(err) {
			return Project{}, ErrDuplicateProjectKey
		}
		return Project{}, fmt.Errorf("insert project: %w", err)
	}
	return out, nil
}

func getProject(ctx context.Context, db *sqlx.DB, id string) (Project, error) {
	var out Project
	err := db.GetContext(
		ctx,
		&out,
		`SELECT `+selectCols+`
		 FROM projects
		 WHERE id = $1`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Project{}, ErrProjectNotFound
		}
		return Project{}, fmt.Errorf("get project: %w", err)
	}
	return out, nil
}

func listProjects(ctx context.Context, db *sqlx.DB, workspaceID string) ([]Project, error) {
	var out []Project
	err := db.SelectContext(
		ctx,
		&out,
		`SELECT `+selectCols+`
		 FROM projects
		 WHERE workspace_id = $1
		   AND archived_at IS NULL
		 ORDER BY created_at ASC`,
		workspaceID,
	)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	return out, nil
}

func archiveProject(ctx context.Context, db *sqlx.DB, id string) error {
	res, err := db.ExecContext(
		ctx,
		`UPDATE projects
		 SET archived_at = NOW()
		 WHERE id = $1
		   AND archived_at IS NULL`,
		id,
	)
	if err != nil {
		return fmt.Errorf("archive project: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive project rows affected: %w", err)
	}
	if n == 0 {
		return ErrProjectNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
