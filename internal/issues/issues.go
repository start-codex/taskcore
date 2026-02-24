package issues

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

var ErrIssueNotFound = errors.New("issue not found")

type MoveIssueParams struct {
	ProjectID      string
	IssueID        string
	TargetStatusID string
	TargetPosition int
}

func (p MoveIssueParams) Validate() error {
	if p.ProjectID == "" || p.IssueID == "" {
		return errors.New("project_id and issue_id are required")
	}
	if p.TargetPosition < 0 {
		return errors.New("target_position must be >= 0")
	}
	return nil
}

func MoveIssue(ctx context.Context, db *sqlx.DB, p MoveIssueParams) error {
	if db == nil {
		return errors.New("db is required")
	}
	if err := p.Validate(); err != nil {
		return err
	}
	return moveIssue(ctx, db, p)
}
