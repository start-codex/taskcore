package issuetypes

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrIssueTypeNotFound  = errors.New("issue type not found")
	ErrDuplicateIssueType = errors.New("issue type name already exists in project")
)

type IssueType struct {
	ID         string     `db:"id"          json:"id"`
	ProjectID  string     `db:"project_id"  json:"project_id"`
	Name       string     `db:"name"        json:"name"`
	Icon       string     `db:"icon"        json:"icon"`
	Level      int        `db:"level"       json:"level"`
	CreatedAt  time.Time  `db:"created_at"  json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"  json:"updated_at"`
	ArchivedAt *time.Time `db:"archived_at" json:"archived_at,omitempty"`
}

type CreateIssueTypeParams struct {
	ProjectID string
	Name      string
	Icon      string
	Level     int
}

func (params CreateIssueTypeParams) Validate() error {
	if params.ProjectID == "" {
		return errors.New("project_id is required")
	}
	if params.Name == "" {
		return errors.New("name is required")
	}
	if params.Level < 0 {
		return errors.New("level must be >= 0")
	}
	return nil
}

func CreateIssueType(ctx context.Context, db *sqlx.DB, params CreateIssueTypeParams) (IssueType, error) {
	if db == nil {
		return IssueType{}, errors.New("db is required")
	}
	if err := params.Validate(); err != nil {
		return IssueType{}, err
	}
	return createIssueType(ctx, db, params)
}

func ListIssueTypes(ctx context.Context, db *sqlx.DB, projectID string) ([]IssueType, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	if projectID == "" {
		return nil, errors.New("project_id is required")
	}
	return listIssueTypes(ctx, db, projectID)
}

func ArchiveIssueType(ctx context.Context, db *sqlx.DB, projectID, issueTypeID string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if projectID == "" {
		return errors.New("project_id is required")
	}
	if issueTypeID == "" {
		return errors.New("issue_type_id is required")
	}
	return archiveIssueType(ctx, db, projectID, issueTypeID)
}
