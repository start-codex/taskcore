package users

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrDuplicateEmail = errors.New("email already exists")
	ErrMemberNotFound = errors.New("member not found")
)

var validWorkspaceRoles = map[string]bool{"owner": true, "admin": true, "member": true}
var validProjectRoles = map[string]bool{"admin": true, "member": true, "viewer": true}

type User struct {
	ID         string     `db:"id"          json:"id"`
	Email      string     `db:"email"       json:"email"`
	Name       string     `db:"name"        json:"name"`
	CreatedAt  time.Time  `db:"created_at"  json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"  json:"updated_at"`
	ArchivedAt *time.Time `db:"archived_at" json:"archived_at,omitempty"`
}

type WorkspaceMember struct {
	WorkspaceID string     `db:"workspace_id" json:"workspace_id"`
	UserID      string     `db:"user_id"      json:"user_id"`
	Role        string     `db:"role"         json:"role"`
	CreatedAt   time.Time  `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"   json:"updated_at"`
	ArchivedAt  *time.Time `db:"archived_at"  json:"archived_at,omitempty"`
}

type ProjectMember struct {
	ProjectID  string     `db:"project_id"  json:"project_id"`
	UserID     string     `db:"user_id"     json:"user_id"`
	Role       string     `db:"role"        json:"role"`
	CreatedAt  time.Time  `db:"created_at"  json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"  json:"updated_at"`
	ArchivedAt *time.Time `db:"archived_at" json:"archived_at,omitempty"`
}

type CreateUserParams struct {
	Email string
	Name  string
}

func (p CreateUserParams) Validate() error {
	if p.Name == "" {
		return errors.New("name is required")
	}
	if !strings.Contains(p.Email, "@") || p.Email == "" {
		return errors.New("email is required and must contain @")
	}
	return nil
}

type AddWorkspaceMemberParams struct {
	WorkspaceID string
	UserID      string
	Role        string
}

func (p AddWorkspaceMemberParams) Validate() error {
	if p.WorkspaceID == "" {
		return errors.New("workspace_id is required")
	}
	if p.UserID == "" {
		return errors.New("user_id is required")
	}
	if !validWorkspaceRoles[p.Role] {
		return errors.New("role must be 'owner', 'admin' or 'member'")
	}
	return nil
}

type UpdateWorkspaceMemberRoleParams struct {
	WorkspaceID string
	UserID      string
	Role        string
}

func (p UpdateWorkspaceMemberRoleParams) Validate() error {
	if p.WorkspaceID == "" {
		return errors.New("workspace_id is required")
	}
	if p.UserID == "" {
		return errors.New("user_id is required")
	}
	if !validWorkspaceRoles[p.Role] {
		return errors.New("role must be 'owner', 'admin' or 'member'")
	}
	return nil
}

type AddProjectMemberParams struct {
	ProjectID string
	UserID    string
	Role      string
}

func (p AddProjectMemberParams) Validate() error {
	if p.ProjectID == "" {
		return errors.New("project_id is required")
	}
	if p.UserID == "" {
		return errors.New("user_id is required")
	}
	if !validProjectRoles[p.Role] {
		return errors.New("role must be 'admin', 'member' or 'viewer'")
	}
	return nil
}

type UpdateProjectMemberRoleParams struct {
	ProjectID string
	UserID    string
	Role      string
}

func (p UpdateProjectMemberRoleParams) Validate() error {
	if p.ProjectID == "" {
		return errors.New("project_id is required")
	}
	if p.UserID == "" {
		return errors.New("user_id is required")
	}
	if !validProjectRoles[p.Role] {
		return errors.New("role must be 'admin', 'member' or 'viewer'")
	}
	return nil
}

// --- User ---

func CreateUser(ctx context.Context, db *sqlx.DB, p CreateUserParams) (User, error) {
	if db == nil {
		return User{}, errors.New("db is required")
	}
	if err := p.Validate(); err != nil {
		return User{}, err
	}
	return createUser(ctx, db, p)
}

func GetUser(ctx context.Context, db *sqlx.DB, id string) (User, error) {
	if db == nil {
		return User{}, errors.New("db is required")
	}
	if id == "" {
		return User{}, errors.New("id is required")
	}
	return getUser(ctx, db, id)
}

func GetUserByEmail(ctx context.Context, db *sqlx.DB, email string) (User, error) {
	if db == nil {
		return User{}, errors.New("db is required")
	}
	if email == "" {
		return User{}, errors.New("email is required")
	}
	return getUserByEmail(ctx, db, email)
}

func ArchiveUser(ctx context.Context, db *sqlx.DB, id string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if id == "" {
		return errors.New("id is required")
	}
	return archiveUser(ctx, db, id)
}

// --- Workspace members ---

func AddWorkspaceMember(ctx context.Context, db *sqlx.DB, p AddWorkspaceMemberParams) (WorkspaceMember, error) {
	if db == nil {
		return WorkspaceMember{}, errors.New("db is required")
	}
	if err := p.Validate(); err != nil {
		return WorkspaceMember{}, err
	}
	return addWorkspaceMember(ctx, db, p)
}

func RemoveWorkspaceMember(ctx context.Context, db *sqlx.DB, workspaceID, userID string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if workspaceID == "" {
		return errors.New("workspace_id is required")
	}
	if userID == "" {
		return errors.New("user_id is required")
	}
	return removeWorkspaceMember(ctx, db, workspaceID, userID)
}

func ListWorkspaceMembers(ctx context.Context, db *sqlx.DB, workspaceID string) ([]WorkspaceMember, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	if workspaceID == "" {
		return nil, errors.New("workspace_id is required")
	}
	return listWorkspaceMembers(ctx, db, workspaceID)
}

func UpdateWorkspaceMemberRole(ctx context.Context, db *sqlx.DB, p UpdateWorkspaceMemberRoleParams) (WorkspaceMember, error) {
	if db == nil {
		return WorkspaceMember{}, errors.New("db is required")
	}
	if err := p.Validate(); err != nil {
		return WorkspaceMember{}, err
	}
	return updateWorkspaceMemberRole(ctx, db, p)
}

// --- Project members ---

func AddProjectMember(ctx context.Context, db *sqlx.DB, p AddProjectMemberParams) (ProjectMember, error) {
	if db == nil {
		return ProjectMember{}, errors.New("db is required")
	}
	if err := p.Validate(); err != nil {
		return ProjectMember{}, err
	}
	return addProjectMember(ctx, db, p)
}

func RemoveProjectMember(ctx context.Context, db *sqlx.DB, projectID, userID string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if projectID == "" {
		return errors.New("project_id is required")
	}
	if userID == "" {
		return errors.New("user_id is required")
	}
	return removeProjectMember(ctx, db, projectID, userID)
}

func ListProjectMembers(ctx context.Context, db *sqlx.DB, projectID string) ([]ProjectMember, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	if projectID == "" {
		return nil, errors.New("project_id is required")
	}
	return listProjectMembers(ctx, db, projectID)
}

func UpdateProjectMemberRole(ctx context.Context, db *sqlx.DB, p UpdateProjectMemberRoleParams) (ProjectMember, error) {
	if db == nil {
		return ProjectMember{}, errors.New("db is required")
	}
	if err := p.Validate(); err != nil {
		return ProjectMember{}, err
	}
	return updateProjectMemberRole(ctx, db, p)
}
