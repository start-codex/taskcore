package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const userCols = `id, email, name, created_at, updated_at, archived_at`
const wsMemberCols = `workspace_id, user_id, role, created_at, updated_at, archived_at`
const projMemberCols = `project_id, user_id, role, created_at, updated_at, archived_at`

// --- User ---

func createUser(ctx context.Context, db *sqlx.DB, p CreateUserParams) (User, error) {
	var out User
	err := db.QueryRowxContext(ctx,
		`INSERT INTO app_users (email, name)
		 VALUES ($1, $2)
		 RETURNING `+userCols,
		p.Email, p.Name,
	).StructScan(&out)
	if err != nil {
		if isUniqueViolation(err) {
			return User{}, ErrDuplicateEmail
		}
		return User{}, fmt.Errorf("insert user: %w", err)
	}
	return out, nil
}

func getUser(ctx context.Context, db *sqlx.DB, id string) (User, error) {
	var out User
	err := db.GetContext(ctx, &out,
		`SELECT `+userCols+` FROM app_users WHERE id = $1`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("get user: %w", err)
	}
	return out, nil
}

func getUserByEmail(ctx context.Context, db *sqlx.DB, email string) (User, error) {
	var out User
	err := db.GetContext(ctx, &out,
		`SELECT `+userCols+` FROM app_users WHERE email = $1`,
		email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("get user by email: %w", err)
	}
	return out, nil
}

func archiveUser(ctx context.Context, db *sqlx.DB, id string) error {
	res, err := db.ExecContext(ctx,
		`UPDATE app_users
		 SET archived_at = NOW()
		 WHERE id = $1 AND archived_at IS NULL`,
		id,
	)
	if err != nil {
		return fmt.Errorf("archive user: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive user rows affected: %w", err)
	}
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

// --- Workspace members ---

// addWorkspaceMember upserts a workspace member. If the member already exists
// (even archived), their role is updated and archived_at is cleared.
func addWorkspaceMember(ctx context.Context, db *sqlx.DB, p AddWorkspaceMemberParams) (WorkspaceMember, error) {
	var out WorkspaceMember
	err := db.QueryRowxContext(ctx,
		`INSERT INTO workspace_members (workspace_id, user_id, role)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (workspace_id, user_id)
		 DO UPDATE SET role = excluded.role, archived_at = NULL
		 RETURNING `+wsMemberCols,
		p.WorkspaceID, p.UserID, p.Role,
	).StructScan(&out)
	if err != nil {
		return WorkspaceMember{}, fmt.Errorf("add workspace member: %w", err)
	}
	return out, nil
}

func removeWorkspaceMember(ctx context.Context, db *sqlx.DB, workspaceID, userID string) error {
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

func listWorkspaceMembers(ctx context.Context, db *sqlx.DB, workspaceID string) ([]WorkspaceMember, error) {
	var out []WorkspaceMember
	err := db.SelectContext(ctx, &out,
		`SELECT `+wsMemberCols+`
		 FROM workspace_members
		 WHERE workspace_id = $1 AND archived_at IS NULL
		 ORDER BY created_at ASC`,
		workspaceID,
	)
	if err != nil {
		return nil, fmt.Errorf("list workspace members: %w", err)
	}
	return out, nil
}

func updateWorkspaceMemberRole(ctx context.Context, db *sqlx.DB, p UpdateWorkspaceMemberRoleParams) (WorkspaceMember, error) {
	var out WorkspaceMember
	err := db.QueryRowxContext(ctx,
		`UPDATE workspace_members
		 SET role = $1
		 WHERE workspace_id = $2 AND user_id = $3 AND archived_at IS NULL
		 RETURNING `+wsMemberCols,
		p.Role, p.WorkspaceID, p.UserID,
	).StructScan(&out)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WorkspaceMember{}, ErrMemberNotFound
		}
		return WorkspaceMember{}, fmt.Errorf("update workspace member role: %w", err)
	}
	return out, nil
}

// --- Project members ---

// addProjectMember upserts a project member. If the member already exists
// (even archived), their role is updated and archived_at is cleared.
func addProjectMember(ctx context.Context, db *sqlx.DB, p AddProjectMemberParams) (ProjectMember, error) {
	var out ProjectMember
	err := db.QueryRowxContext(ctx,
		`INSERT INTO project_members (project_id, user_id, role)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (project_id, user_id)
		 DO UPDATE SET role = excluded.role, archived_at = NULL
		 RETURNING `+projMemberCols,
		p.ProjectID, p.UserID, p.Role,
	).StructScan(&out)
	if err != nil {
		return ProjectMember{}, fmt.Errorf("add project member: %w", err)
	}
	return out, nil
}

func removeProjectMember(ctx context.Context, db *sqlx.DB, projectID, userID string) error {
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

func listProjectMembers(ctx context.Context, db *sqlx.DB, projectID string) ([]ProjectMember, error) {
	var out []ProjectMember
	err := db.SelectContext(ctx, &out,
		`SELECT `+projMemberCols+`
		 FROM project_members
		 WHERE project_id = $1 AND archived_at IS NULL
		 ORDER BY created_at ASC`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list project members: %w", err)
	}
	return out, nil
}

func updateProjectMemberRole(ctx context.Context, db *sqlx.DB, p UpdateProjectMemberRoleParams) (ProjectMember, error) {
	var out ProjectMember
	err := db.QueryRowxContext(ctx,
		`UPDATE project_members
		 SET role = $1
		 WHERE project_id = $2 AND user_id = $3 AND archived_at IS NULL
		 RETURNING `+projMemberCols,
		p.Role, p.ProjectID, p.UserID,
	).StructScan(&out)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ProjectMember{}, ErrMemberNotFound
		}
		return ProjectMember{}, fmt.Errorf("update project member role: %w", err)
	}
	return out, nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
