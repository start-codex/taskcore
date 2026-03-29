// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package projects

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/trazawork/internal/pgutil"
)

const selectCols = `id, workspace_id, name, key, description, created_at, updated_at, archived_at`
const memberCols = `project_id, user_id, role, created_at, updated_at, archived_at`

type templateStatus struct {
	name     string
	category string
}

var templateDefs = map[string]map[string][]templateStatus{
	"en": {
		"kanban": {
			{"To Do", "todo"},
			{"In Progress", "doing"},
			{"Done", "done"},
		},
		"scrum": {
			{"Backlog", "todo"},
			{"To Do", "todo"},
			{"In Progress", "doing"},
			{"In Review", "doing"},
			{"Done", "done"},
		},
	},
	"es": {
		"kanban": {
			{"Por hacer", "todo"},
			{"En progreso", "doing"},
			{"Hecho", "done"},
		},
		"scrum": {
			{"Backlog", "todo"},
			{"Por hacer", "todo"},
			{"En progreso", "doing"},
			{"En revisión", "doing"},
			{"Hecho", "done"},
		},
	},
}

func resolveTemplate(template, locale string) []templateStatus {
	if defs, ok := templateDefs[locale]; ok {
		if statuses, ok := defs[template]; ok {
			return statuses
		}
	}
	return templateDefs["en"][template]
}

func createProject(ctx context.Context, db *sqlx.DB, params CreateProjectParams) (Project, error) {
	if params.Template == "" {
		var project Project
		err := db.QueryRowxContext(
			ctx,
			`INSERT INTO projects (workspace_id, name, key, description)
			 VALUES ($1, $2, $3, $4)
			 RETURNING `+selectCols,
			params.WorkspaceID, params.Name, params.Key, params.Description,
		).StructScan(&project)
		if err != nil {
			if pgutil.IsUniqueViolation(err) {
				return Project{}, ErrDuplicateProjectKey
			}
			return Project{}, fmt.Errorf("insert project: %w", err)
		}
		return project, nil
	}

	var project Project
	if err := pgutil.WithTx(ctx, db, nil, "begin transaction", "commit transaction", func(tx *sqlx.Tx) error {
		if err := tx.QueryRowxContext(
			ctx,
			`INSERT INTO projects (workspace_id, name, key, description)
			 VALUES ($1, $2, $3, $4)
			 RETURNING `+selectCols,
			params.WorkspaceID, params.Name, params.Key, params.Description,
		).StructScan(&project); err != nil {
			if pgutil.IsUniqueViolation(err) {
				return ErrDuplicateProjectKey
			}
			return fmt.Errorf("insert project: %w", err)
		}

		for i, s := range resolveTemplate(params.Template, params.Locale) {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO statuses (project_id, name, category, position) VALUES ($1, $2, $3, $4)`,
				project.ID, s.name, s.category, i,
			); err != nil {
				return fmt.Errorf("insert status %q: %w", s.name, err)
			}
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO boards (project_id, name, type, filter_query) VALUES ($1, $2, $3, '')`,
			project.ID, "Board", params.Template,
		); err != nil {
			return fmt.Errorf("insert board: %w", err)
		}
		return nil
	}); err != nil {
		return Project{}, err
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
	projects := []Project{}
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
	members := []ProjectMember{}
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
