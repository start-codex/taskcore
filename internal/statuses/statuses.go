package statuses

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrStatusNotFound  = errors.New("status not found")
	ErrDuplicateStatus = errors.New("status name already exists in project")
)

var validCategories = map[string]bool{"todo": true, "doing": true, "done": true}

type Status struct {
	ID         string     `db:"id"          json:"id"`
	ProjectID  string     `db:"project_id"  json:"project_id"`
	Name       string     `db:"name"        json:"name"`
	Category   string     `db:"category"    json:"category"`
	Position   int        `db:"position"    json:"position"`
	CreatedAt  time.Time  `db:"created_at"  json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"  json:"updated_at"`
	ArchivedAt *time.Time `db:"archived_at" json:"archived_at,omitempty"`
}

type CreateStatusParams struct {
	ProjectID string
	Name      string
	Category  string
}

func (params CreateStatusParams) Validate() error {
	if params.ProjectID == "" {
		return errors.New("project_id is required")
	}
	if params.Name == "" {
		return errors.New("name is required")
	}
	if !validCategories[params.Category] {
		return errors.New("category must be 'todo', 'doing' or 'done'")
	}
	return nil
}

type UpdateStatusParams struct {
	StatusID  string
	ProjectID string
	Name      string
	Category  string
}

func (params UpdateStatusParams) Validate() error {
	if params.StatusID == "" {
		return errors.New("status_id is required")
	}
	if params.ProjectID == "" {
		return errors.New("project_id is required")
	}
	if params.Name == "" {
		return errors.New("name is required")
	}
	if !validCategories[params.Category] {
		return errors.New("category must be 'todo', 'doing' or 'done'")
	}
	return nil
}

func CreateStatus(ctx context.Context, db *sqlx.DB, params CreateStatusParams) (Status, error) {
	if db == nil {
		return Status{}, errors.New("db is required")
	}
	if err := params.Validate(); err != nil {
		return Status{}, err
	}
	return createStatus(ctx, db, params)
}

func ListStatuses(ctx context.Context, db *sqlx.DB, projectID string) ([]Status, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	if projectID == "" {
		return nil, errors.New("project_id is required")
	}
	return listStatuses(ctx, db, projectID)
}

func UpdateStatus(ctx context.Context, db *sqlx.DB, params UpdateStatusParams) (Status, error) {
	if db == nil {
		return Status{}, errors.New("db is required")
	}
	if err := params.Validate(); err != nil {
		return Status{}, err
	}
	return updateStatus(ctx, db, params)
}

func ArchiveStatus(ctx context.Context, db *sqlx.DB, projectID, statusID string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if projectID == "" {
		return errors.New("project_id is required")
	}
	if statusID == "" {
		return errors.New("status_id is required")
	}
	return archiveStatus(ctx, db, projectID, statusID)
}
