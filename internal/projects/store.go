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
const memberCols = `project_id, user_id, role, created_at, updated_at, archived_at`

func createProject(ctx context.Context, db *sqlx.DB, params CreateProjectParams) (Project, error) {
	var project Project
	err := db.QueryRowxContext(
		ctx,
		`INSERT INTO projects (workspace_id, name, key, description)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+selectCols,
		params.WorkspaceID,
		params.Name,
		params.Key,
		params.Description,
	).StructScan(&project)
	if err != nil {
		if isUniqueViolation(err) {
			return Project{}, ErrDuplicateProjectKey
		}
		return Project{}, fmt.Errorf("insert project: %w", err)
	}
	return project, nil
}

func getProject(ctx context.Context, db *sqlx.DB, id string) (Project, error) {
	var project Project
	err := db.GetContext(
		ctx,
		&project,
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
	return project, nil
}

func listProjects(ctx context.Context, db *sqlx.DB, workspaceID string) ([]Project, error) {
	var projects []Project
	err := db.SelectContext(
		ctx,
		&projects,
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
	return projects, nil
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

func addMember(ctx context.Context, db *sqlx.DB, params AddMemberParams) (ProjectMember, error) {
	var member ProjectMember
	err := db.QueryRowxContext(ctx,
		`INSERT INTO project_members (project_id, user_id, role)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (project_id, user_id)
		 DO UPDATE SET role = excluded.role, archived_at = NULL
		 RETURNING `+memberCols,
		params.ProjectID, params.UserID, params.Role,
	).StructScan(&member)
	if err != nil {
		return ProjectMember{}, fmt.Errorf("add project member: %w", err)
	}
	return member, nil
}

func removeMember(ctx context.Context, db *sqlx.DB, projectID, userID string) error {
	res, err := db.ExecContext(ctx,
		`UPDATE project_members
		 SET archived_at = NOW()
		 WHERE project_id = $1 AND user_id = $2 AND archived_at IS NULL`,
		projectID, userID,
	)
	if err != nil {
		return fmt.Errorf("remove project member: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("remove project member rows affected: %w", err)
	}
	if n == 0 {
		return ErrMemberNotFound
	}
	return nil
}

func listMembers(ctx context.Context, db *sqlx.DB, projectID string) ([]ProjectMember, error) {
	var members []ProjectMember
	err := db.SelectContext(ctx, &members,
		`SELECT `+memberCols+`
		 FROM project_members
		 WHERE project_id = $1 AND archived_at IS NULL
		 ORDER BY created_at ASC`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list project members: %w", err)
	}
	return members, nil
}

func updateMemberRole(ctx context.Context, db *sqlx.DB, params UpdateMemberRoleParams) (ProjectMember, error) {
	var member ProjectMember
	err := db.QueryRowxContext(ctx,
		`UPDATE project_members
		 SET role = $1
		 WHERE project_id = $2 AND user_id = $3 AND archived_at IS NULL
		 RETURNING `+memberCols,
		params.Role, params.ProjectID, params.UserID,
	).StructScan(&member)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ProjectMember{}, ErrMemberNotFound
		}
		return ProjectMember{}, fmt.Errorf("update project member role: %w", err)
	}
	return member, nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
