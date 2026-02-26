package statuses

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const statusCols = `id, project_id, name, category, position, created_at, updated_at, archived_at`

func createStatus(ctx context.Context, db *sqlx.DB, p CreateStatusParams) (Status, error) {
	var out Status
	err := db.QueryRowxContext(ctx,
		`INSERT INTO statuses (project_id, name, category, position)
		 VALUES ($1, $2, $3,
		   COALESCE(
		     (SELECT MAX(position) + 1 FROM statuses WHERE project_id = $1 AND archived_at IS NULL),
		     0
		   )
		 )
		 RETURNING `+statusCols,
		p.ProjectID, p.Name, p.Category,
	).StructScan(&out)
	if err != nil {
		if isUniqueViolation(err) {
			return Status{}, ErrDuplicateStatus
		}
		return Status{}, fmt.Errorf("create status: %w", err)
	}
	return out, nil
}

func listStatuses(ctx context.Context, db *sqlx.DB, projectID string) ([]Status, error) {
	var out []Status
	if err := db.SelectContext(ctx, &out,
		`SELECT `+statusCols+`
		 FROM statuses
		 WHERE project_id = $1
		   AND archived_at IS NULL
		 ORDER BY position ASC`,
		projectID,
	); err != nil {
		return nil, fmt.Errorf("list statuses: %w", err)
	}
	return out, nil
}

func updateStatus(ctx context.Context, db *sqlx.DB, p UpdateStatusParams) (Status, error) {
	var out Status
	err := db.QueryRowxContext(ctx,
		`UPDATE statuses
		 SET name     = $1,
		     category = $2
		 WHERE id         = $3
		   AND project_id = $4
		   AND archived_at IS NULL
		 RETURNING `+statusCols,
		p.Name, p.Category, p.StatusID, p.ProjectID,
	).StructScan(&out)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Status{}, ErrStatusNotFound
		}
		if isUniqueViolation(err) {
			return Status{}, ErrDuplicateStatus
		}
		return Status{}, fmt.Errorf("update status: %w", err)
	}
	return out, nil
}

func archiveStatus(ctx context.Context, db *sqlx.DB, projectID, statusID string) error {
	res, err := db.ExecContext(ctx,
		`UPDATE statuses
		 SET archived_at = NOW()
		 WHERE id         = $1
		   AND project_id = $2
		   AND archived_at IS NULL`,
		statusID, projectID,
	)
	if err != nil {
		return fmt.Errorf("archive status: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive status rows affected: %w", err)
	}
	if n == 0 {
		return ErrStatusNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
