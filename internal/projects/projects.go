package projects

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrProjectNotFound     = errors.New("project not found")
	ErrDuplicateProjectKey = errors.New("project key already exists in workspace")
)

var reKey = regexp.MustCompile(`^[A-Z]{2,10}$`)

type Project struct {
	ID          string     `db:"id"           json:"id"`
	WorkspaceID string     `db:"workspace_id" json:"workspace_id"`
	Name        string     `db:"name"         json:"name"`
	Key         string     `db:"key"          json:"key"`
	Description string     `db:"description"  json:"description"`
	CreatedAt   time.Time  `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"   json:"updated_at"`
	ArchivedAt  *time.Time `db:"archived_at"  json:"archived_at,omitempty"`
}

type CreateProjectParams struct {
	WorkspaceID string
	Name        string
	Key         string
	Description string
}

func (p CreateProjectParams) Validate() error {
	if p.WorkspaceID == "" {
		return errors.New("workspace_id is required")
	}
	if p.Name == "" {
		return errors.New("name is required")
	}
	if !reKey.MatchString(p.Key) {
		return errors.New("key must be 2-10 uppercase letters (A-Z)")
	}
	return nil
}

func CreateProject(ctx context.Context, db *sqlx.DB, p CreateProjectParams) (Project, error) {
	if db == nil {
		return Project{}, errors.New("db is required")
	}
	if err := p.Validate(); err != nil {
		return Project{}, err
	}
	return createProject(ctx, db, p)
}

func GetProject(ctx context.Context, db *sqlx.DB, id string) (Project, error) {
	if db == nil {
		return Project{}, errors.New("db is required")
	}
	if id == "" {
		return Project{}, errors.New("id is required")
	}
	return getProject(ctx, db, id)
}

func ListProjects(ctx context.Context, db *sqlx.DB, workspaceID string) ([]Project, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	if workspaceID == "" {
		return nil, errors.New("workspace_id is required")
	}
	return listProjects(ctx, db, workspaceID)
}

func ArchiveProject(ctx context.Context, db *sqlx.DB, id string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if id == "" {
		return errors.New("id is required")
	}
	return archiveProject(ctx, db, id)
}
