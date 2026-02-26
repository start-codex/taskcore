package workspaces

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrDuplicateSlug     = errors.New("slug already exists")
)

var reSlug = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,49}$`)

type Workspace struct {
	ID         string     `db:"id"          json:"id"`
	Name       string     `db:"name"        json:"name"`
	Slug       string     `db:"slug"        json:"slug"`
	CreatedAt  time.Time  `db:"created_at"  json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"  json:"updated_at"`
	ArchivedAt *time.Time `db:"archived_at" json:"archived_at,omitempty"`
}

type CreateWorkspaceParams struct {
	Name string
	Slug string
}

func (p CreateWorkspaceParams) Validate() error {
	if p.Name == "" {
		return errors.New("name is required")
	}
	if !reSlug.MatchString(p.Slug) {
		return errors.New("slug must be 2-50 lowercase alphanumeric characters or hyphens, starting with a letter or digit")
	}
	return nil
}

func CreateWorkspace(ctx context.Context, db *sqlx.DB, p CreateWorkspaceParams) (Workspace, error) {
	if db == nil {
		return Workspace{}, errors.New("db is required")
	}
	if err := p.Validate(); err != nil {
		return Workspace{}, err
	}
	return createWorkspace(ctx, db, p)
}

func GetWorkspace(ctx context.Context, db *sqlx.DB, id string) (Workspace, error) {
	if db == nil {
		return Workspace{}, errors.New("db is required")
	}
	if id == "" {
		return Workspace{}, errors.New("id is required")
	}
	return getWorkspace(ctx, db, id)
}

func GetWorkspaceBySlug(ctx context.Context, db *sqlx.DB, slug string) (Workspace, error) {
	if db == nil {
		return Workspace{}, errors.New("db is required")
	}
	if slug == "" {
		return Workspace{}, errors.New("slug is required")
	}
	return getWorkspaceBySlug(ctx, db, slug)
}

func ArchiveWorkspace(ctx context.Context, db *sqlx.DB, id string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if id == "" {
		return errors.New("id is required")
	}
	return archiveWorkspace(ctx, db, id)
}
