package issuetypes

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const issueTypeCols = `id, project_id, name, icon, level, created_at, updated_at, archived_at`

func createIssueType(ctx context.Context, db *sqlx.DB, p CreateIssueTypeParams) (IssueType, error) {
	var out IssueType
	err := db.QueryRowxContext(ctx,
		`INSERT INTO issue_types (project_id, name, icon, level)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+issueTypeCols,
		p.ProjectID, p.Name, p.Icon, p.Level,
	).StructScan(&out)
	if err != nil {
		if isUniqueViolation(err) {
			return IssueType{}, ErrDuplicateIssueType
		}
		return IssueType{}, fmt.Errorf("create issue type: %w", err)
	}
	return out, nil
}

func listIssueTypes(ctx context.Context, db *sqlx.DB, projectID string) ([]IssueType, error) {
	var out []IssueType
	if err := db.SelectContext(ctx, &out,
		`SELECT `+issueTypeCols+`
		 FROM issue_types
		 WHERE project_id = $1
		   AND archived_at IS NULL
		 ORDER BY level ASC, name ASC`,
		projectID,
	); err != nil {
		return nil, fmt.Errorf("list issue types: %w", err)
	}
	return out, nil
}

func archiveIssueType(ctx context.Context, db *sqlx.DB, projectID, issueTypeID string) error {
	res, err := db.ExecContext(ctx,
		`UPDATE issue_types
		 SET archived_at = NOW()
		 WHERE id         = $1
		   AND project_id = $2
		   AND archived_at IS NULL`,
		issueTypeID, projectID,
	)
	if err != nil {
		return fmt.Errorf("archive issue type: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive issue type rows affected: %w", err)
	}
	if n == 0 {
		return ErrIssueTypeNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
