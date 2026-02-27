package issues

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

const reorderOffset = 1000000

const issueCols = `id, project_id, number, issue_type_id, status_id, parent_issue_id,
	title, description, priority, assignee_id, reporter_id, due_date,
	status_position, created_at, updated_at, archived_at`

func createIssue(ctx context.Context, db *sqlx.DB, params CreateIssueParams) (Issue, error) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return Issue{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var number int
	if err := tx.QueryRowxContext(ctx,
		`INSERT INTO project_issue_counters (project_id, last_number)
		 VALUES ($1, 1)
		 ON CONFLICT (project_id)
		 DO UPDATE SET last_number = project_issue_counters.last_number + 1
		 RETURNING last_number`,
		params.ProjectID,
	).Scan(&number); err != nil {
		return Issue{}, fmt.Errorf("upsert issue counter: %w", err)
	}

	var parentIssueID *string
	if params.ParentIssueID != "" {
		parentIssueID = &params.ParentIssueID
	}
	var assigneeID *string
	if params.AssigneeID != "" {
		assigneeID = &params.AssigneeID
	}

	var issue Issue
	err = tx.QueryRowxContext(ctx,
		`INSERT INTO issues (
			project_id, number, issue_type_id, status_id, parent_issue_id,
			title, description, priority, assignee_id, reporter_id, due_date,
			status_position
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, $11,
			(SELECT COALESCE(MAX(status_position), -1) + 1
			 FROM issues
			 WHERE project_id = $1 AND status_id = $4 AND archived_at IS NULL)
		)
		RETURNING `+issueCols,
		params.ProjectID, number, params.IssueTypeID, params.StatusID, parentIssueID,
		params.Title, params.Description, params.Priority, assigneeID, params.ReporterID, params.DueDate,
	).StructScan(&issue)
	if err != nil {
		return Issue{}, fmt.Errorf("insert issue: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return Issue{}, fmt.Errorf("commit create issue: %w", err)
	}
	return issue, nil
}

func getIssue(ctx context.Context, db *sqlx.DB, projectID, issueID string) (Issue, error) {
	var issue Issue
	err := db.GetContext(ctx, &issue,
		`SELECT `+issueCols+`
		 FROM issues
		 WHERE id = $1 AND project_id = $2`,
		issueID, projectID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Issue{}, ErrIssueNotFound
		}
		return Issue{}, fmt.Errorf("get issue: %w", err)
	}
	return issue, nil
}

func listIssues(ctx context.Context, db *sqlx.DB, params ListIssuesParams) ([]Issue, error) {
	query := `SELECT ` + issueCols + `
		 FROM issues
		 WHERE project_id = $1
		   AND archived_at IS NULL`
	args := []any{params.ProjectID}

	if params.StatusID != "" {
		args = append(args, params.StatusID)
		query += fmt.Sprintf(" AND status_id = $%d", len(args))
	}
	if params.AssigneeID != "" {
		args = append(args, params.AssigneeID)
		query += fmt.Sprintf(" AND assignee_id = $%d", len(args))
	}

	query += ` ORDER BY status_id, status_position ASC`

	var issues []Issue
	if err := db.SelectContext(ctx, &issues, query, args...); err != nil {
		return nil, fmt.Errorf("list issues: %w", err)
	}
	return issues, nil
}

func updateIssue(ctx context.Context, db *sqlx.DB, params UpdateIssueParams) (Issue, error) {
	var issue Issue
	err := db.QueryRowxContext(ctx,
		`UPDATE issues
		 SET title       = $1,
		     description = $2,
		     priority    = $3,
		     assignee_id = $4,
		     due_date    = $5
		 WHERE id = $6
		   AND project_id = $7
		   AND archived_at IS NULL
		 RETURNING `+issueCols,
		params.Title, params.Description, params.Priority, params.AssigneeID, params.DueDate,
		params.IssueID, params.ProjectID,
	).StructScan(&issue)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Issue{}, ErrIssueNotFound
		}
		return Issue{}, fmt.Errorf("update issue: %w", err)
	}
	return issue, nil
}

func archiveIssue(ctx context.Context, db *sqlx.DB, projectID, issueID string) error {
	res, err := db.ExecContext(ctx,
		`UPDATE issues
		 SET archived_at = NOW()
		 WHERE id = $1
		   AND project_id = $2
		   AND archived_at IS NULL`,
		issueID, projectID,
	)
	if err != nil {
		return fmt.Errorf("archive issue: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive issue rows affected: %w", err)
	}
	if n == 0 {
		return ErrIssueNotFound
	}
	return nil
}

type issuePosition struct {
	StatusID       string `db:"status_id"`
	StatusPosition int    `db:"status_position"`
}

// moveIssue persists the move of an issue to a target status/position.
// It uses a two-phase offset strategy to avoid transient unique index collisions.
func moveIssue(ctx context.Context, db *sqlx.DB, params MoveIssueParams) error {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	current, err := getIssuePositionForUpdate(ctx, tx, params.ProjectID, params.IssueID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrIssueNotFound
		}
		return err
	}

	sourceStatusID := current.StatusID
	targetStatusID := params.TargetStatusID
	if targetStatusID == "" {
		targetStatusID = sourceStatusID
	}

	if err := lockStatuses(ctx, tx, params.ProjectID, sourceStatusID, targetStatusID); err != nil {
		return err
	}

	if err := lockAffectedIssues(ctx, tx, params.ProjectID, sourceStatusID, targetStatusID); err != nil {
		return err
	}

	targetPos, err := clampTargetPosition(ctx, tx, params.ProjectID, targetStatusID, params.TargetPosition, sourceStatusID == targetStatusID)
	if err != nil {
		return err
	}

	if sourceStatusID == targetStatusID && targetPos == current.StatusPosition {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit no-op move: %w", err)
		}
		return nil
	}

	if err := parkIssueAtTempPosition(ctx, tx, params.ProjectID, params.IssueID, sourceStatusID); err != nil {
		return err
	}

	if sourceStatusID == targetStatusID {
		if err := reorderWithinSameStatus(ctx, tx, params.ProjectID, params.IssueID, sourceStatusID, current.StatusPosition, targetPos); err != nil {
			return err
		}
	} else {
		if err := collapseSourceStatus(ctx, tx, params.ProjectID, params.IssueID, sourceStatusID, current.StatusPosition); err != nil {
			return err
		}
		if err := openGapInTargetStatus(ctx, tx, params.ProjectID, targetStatusID, targetPos); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE issues
		 SET status_id = $1,
		     status_position = $2
		 WHERE id = $3
		   AND project_id = $4`,
		targetStatusID,
		targetPos,
		params.IssueID,
		params.ProjectID,
	); err != nil {
		return fmt.Errorf("place moved issue: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit move issue: %w", err)
	}

	return nil
}

func getIssuePositionForUpdate(ctx context.Context, tx *sqlx.Tx, projectID, issueID string) (issuePosition, error) {
	var pos issuePosition
	err := tx.GetContext(
		ctx,
		&pos,
		`SELECT status_id, status_position
		 FROM issues
		 WHERE id = $1
		   AND project_id = $2
		   AND archived_at IS NULL
		 FOR UPDATE`,
		issueID,
		projectID,
	)
	if err != nil {
		return issuePosition{}, fmt.Errorf("load issue for update: %w", err)
	}
	return pos, nil
}

func lockStatuses(ctx context.Context, tx *sqlx.Tx, projectID, sourceStatusID, targetStatusID string) error {
	rows, err := tx.QueryxContext(
		ctx,
		`SELECT id
		 FROM statuses
		 WHERE project_id = $1
		   AND (id = $2 OR id = $3)
		 ORDER BY id
		 FOR UPDATE`,
		projectID,
		sourceStatusID,
		targetStatusID,
	)
	if err != nil {
		return fmt.Errorf("lock statuses: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate status locks: %w", err)
	}

	required := 1
	if sourceStatusID != targetStatusID {
		required = 2
	}
	if count != required {
		return errors.New("source or target status not found in project")
	}

	return nil
}

func lockAffectedIssues(ctx context.Context, tx *sqlx.Tx, projectID, sourceStatusID, targetStatusID string) error {
	if _, err := tx.ExecContext(
		ctx,
		`SELECT id
		 FROM issues
		 WHERE project_id = $1
		   AND archived_at IS NULL
		   AND (status_id = $2 OR status_id = $3)
		 ORDER BY id
		 FOR UPDATE`,
		projectID,
		sourceStatusID,
		targetStatusID,
	); err != nil {
		return fmt.Errorf("lock affected issues: %w", err)
	}
	return nil
}

func clampTargetPosition(ctx context.Context, tx *sqlx.Tx, projectID, targetStatusID string, requested int, sameStatus bool) (int, error) {
	var count int
	if err := tx.GetContext(
		ctx,
		&count,
		`SELECT COUNT(*)
		 FROM issues
		 WHERE project_id = $1
		   AND status_id = $2
		   AND archived_at IS NULL`,
		projectID,
		targetStatusID,
	); err != nil {
		return 0, fmt.Errorf("count target status issues: %w", err)
	}

	maxPos := count
	if sameStatus {
		maxPos = count - 1
	}
	if maxPos < 0 {
		maxPos = 0
	}

	if requested < 0 {
		return 0, nil
	}
	if requested > maxPos {
		return maxPos, nil
	}
	return requested, nil
}

func parkIssueAtTempPosition(ctx context.Context, tx *sqlx.Tx, projectID, issueID, sourceStatusID string) error {
	var tempPos int
	if err := tx.GetContext(
		ctx,
		&tempPos,
		`SELECT COALESCE(MAX(status_position), -1) + 1
		 FROM issues
		 WHERE project_id = $1
		   AND status_id = $2
		   AND archived_at IS NULL`,
		projectID,
		sourceStatusID,
	); err != nil {
		return fmt.Errorf("compute temp position: %w", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE issues
		 SET status_position = $1
		 WHERE id = $2
		   AND project_id = $3`,
		tempPos,
		issueID,
		projectID,
	); err != nil {
		return fmt.Errorf("park moving issue: %w", err)
	}

	return nil
}

func reorderWithinSameStatus(ctx context.Context, tx *sqlx.Tx, projectID, issueID, statusID string, sourcePos, targetPos int) error {
	if targetPos < sourcePos {
		return shiftUpRange(ctx, tx, projectID, issueID, statusID, targetPos, sourcePos-1)
	}
	if targetPos > sourcePos {
		return shiftDownRange(ctx, tx, projectID, issueID, statusID, sourcePos+1, targetPos)
	}
	return nil
}

func collapseSourceStatus(ctx context.Context, tx *sqlx.Tx, projectID, issueID, statusID string, sourcePos int) error {
	return shiftDownRange(ctx, tx, projectID, issueID, statusID, sourcePos+1, -1)
}

func openGapInTargetStatus(ctx context.Context, tx *sqlx.Tx, projectID, statusID string, targetPos int) error {
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE issues
		 SET status_position = status_position + $1
		 WHERE project_id = $2
		   AND status_id = $3
		   AND archived_at IS NULL
		   AND status_position >= $4`,
		reorderOffset, projectID, statusID, targetPos,
	); err != nil {
		return fmt.Errorf("phase 1 open gap: %w", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE issues
		 SET status_position = status_position - $1 + 1
		 WHERE project_id = $2
		   AND status_id = $3
		   AND archived_at IS NULL
		   AND status_position >= $4`,
		reorderOffset, projectID, statusID, targetPos+reorderOffset,
	); err != nil {
		return fmt.Errorf("phase 2 open gap: %w", err)
	}

	return nil
}

func shiftUpRange(ctx context.Context, tx *sqlx.Tx, projectID, issueID, statusID string, startPos, endPos int) error {
	if startPos > endPos {
		return nil
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE issues
		 SET status_position = status_position + $1
		 WHERE project_id = $2
		   AND status_id = $3
		   AND archived_at IS NULL
		   AND id <> $4
		   AND status_position BETWEEN $5 AND $6`,
		reorderOffset,
		projectID,
		statusID,
		issueID,
		startPos,
		endPos,
	); err != nil {
		return fmt.Errorf("phase 1 shift up range: %w", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE issues
		 SET status_position = status_position - $1 + 1
		 WHERE project_id = $2
		   AND status_id = $3
		   AND archived_at IS NULL
		   AND id <> $4
		   AND status_position BETWEEN $5 AND $6`,
		reorderOffset,
		projectID,
		statusID,
		issueID,
		startPos+reorderOffset,
		endPos+reorderOffset,
	); err != nil {
		return fmt.Errorf("phase 2 shift up range: %w", err)
	}

	return nil
}

// shiftDownRange decreases position by 1 for a range.
// If endPos is -1, it means "to the end".
func shiftDownRange(ctx context.Context, tx *sqlx.Tx, projectID, issueID, statusID string, startPos, endPos int) error {
	if endPos >= 0 && startPos > endPos {
		return nil
	}

	args1 := []any{reorderOffset, projectID, statusID, issueID, startPos}
	phase1 := `UPDATE issues
		 SET status_position = status_position + $1
		 WHERE project_id = $2
		   AND status_id = $3
		   AND archived_at IS NULL
		   AND id <> $4
		   AND status_position >= $5`

	args2 := []any{reorderOffset, projectID, statusID, issueID, startPos + reorderOffset}
	phase2 := `UPDATE issues
		 SET status_position = status_position - $1 - 1
		 WHERE project_id = $2
		   AND status_id = $3
		   AND archived_at IS NULL
		   AND id <> $4
		   AND status_position >= $5`

	if endPos >= 0 {
		args1 = append(args1, endPos)
		phase1 += " AND status_position <= $6"
		args2 = append(args2, endPos+reorderOffset)
		phase2 += " AND status_position <= $6"
	}

	if _, err := tx.ExecContext(ctx, phase1, args1...); err != nil {
		return fmt.Errorf("phase 1 shift down range: %w", err)
	}
	if _, err := tx.ExecContext(ctx, phase2, args2...); err != nil {
		return fmt.Errorf("phase 2 shift down range: %w", err)
	}
	return nil
}
