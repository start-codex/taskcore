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
	ErrMemberNotFound    = errors.New("member not found")
)

var validRoles = map[string]bool{"owner": true, "admin": true, "member": true}

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

func (params CreateWorkspaceParams) Validate() error {
	if params.Name == "" {
		return errors.New("name is required")
	}
	if !reSlug.MatchString(params.Slug) {
		return errors.New("slug must be 2-50 lowercase alphanumeric characters or hyphens, starting with a letter or digit")
	}
	return nil
}

func CreateWorkspace(ctx context.Context, db *sqlx.DB, params CreateWorkspaceParams) (Workspace, error) {
	if db == nil {
		return Workspace{}, errors.New("db is required")
	}
	if err := params.Validate(); err != nil {
		return Workspace{}, err
	}
	return createWorkspace(ctx, db, params)
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

type WorkspaceMember struct {
	WorkspaceID string     `db:"workspace_id" json:"workspace_id"`
	UserID      string     `db:"user_id"      json:"user_id"`
	Role        string     `db:"role"         json:"role"`
	CreatedAt   time.Time  `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"   json:"updated_at"`
	ArchivedAt  *time.Time `db:"archived_at"  json:"archived_at,omitempty"`
}

type AddMemberParams struct {
	WorkspaceID string
	UserID      string
	Role        string
}

func (params AddMemberParams) Validate() error {
	if params.WorkspaceID == "" {
		return errors.New("workspace_id is required")
	}
	if params.UserID == "" {
		return errors.New("user_id is required")
	}
	if !validRoles[params.Role] {
		return errors.New("role must be 'owner', 'admin' or 'member'")
	}
	return nil
}

type UpdateMemberRoleParams struct {
	WorkspaceID string
	UserID      string
	Role        string
}

func (params UpdateMemberRoleParams) Validate() error {
	if params.WorkspaceID == "" {
		return errors.New("workspace_id is required")
	}
	if params.UserID == "" {
		return errors.New("user_id is required")
	}
	if !validRoles[params.Role] {
		return errors.New("role must be 'owner', 'admin' or 'member'")
	}
	return nil
}

func AddMember(ctx context.Context, db *sqlx.DB, params AddMemberParams) (WorkspaceMember, error) {
	if db == nil {
		return WorkspaceMember{}, errors.New("db is required")
	}
	if err := params.Validate(); err != nil {
		return WorkspaceMember{}, err
	}
	return addMember(ctx, db, params)
}

func RemoveMember(ctx context.Context, db *sqlx.DB, workspaceID, userID string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if workspaceID == "" {
		return errors.New("workspace_id is required")
	}
	if userID == "" {
		return errors.New("user_id is required")
	}
	return removeMember(ctx, db, workspaceID, userID)
}

func ListMembers(ctx context.Context, db *sqlx.DB, workspaceID string) ([]WorkspaceMember, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	if workspaceID == "" {
		return nil, errors.New("workspace_id is required")
	}
	return listMembers(ctx, db, workspaceID)
}

func UpdateMemberRole(ctx context.Context, db *sqlx.DB, params UpdateMemberRoleParams) (WorkspaceMember, error) {
	if db == nil {
		return WorkspaceMember{}, errors.New("db is required")
	}
	if err := params.Validate(); err != nil {
		return WorkspaceMember{}, err
	}
	return updateMemberRole(ctx, db, params)
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
