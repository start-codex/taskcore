package issues

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrIssueNotFound   = errors.New("issue not found")
	ErrInvalidPriority = errors.New("priority must be 'low', 'medium', 'high' or 'critical'")
)

var validPriorities = map[string]bool{
	"low": true, "medium": true, "high": true, "critical": true,
}

type Issue struct {
	ID             string     `db:"id"              json:"id"`
	ProjectID      string     `db:"project_id"      json:"project_id"`
	Number         int        `db:"number"          json:"number"`
	IssueTypeID    string     `db:"issue_type_id"   json:"issue_type_id"`
	StatusID       string     `db:"status_id"       json:"status_id"`
	ParentIssueID  *string    `db:"parent_issue_id" json:"parent_issue_id,omitempty"`
	Title          string     `db:"title"           json:"title"`
	Description    string     `db:"description"     json:"description"`
	Priority       string     `db:"priority"        json:"priority"`
	AssigneeID     *string    `db:"assignee_id"     json:"assignee_id,omitempty"`
	ReporterID     string     `db:"reporter_id"     json:"reporter_id"`
	DueDate        *time.Time `db:"due_date"        json:"due_date,omitempty"`
	StatusPosition int        `db:"status_position" json:"status_position"`
	CreatedAt      time.Time  `db:"created_at"      json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"      json:"updated_at"`
	ArchivedAt     *time.Time `db:"archived_at"     json:"archived_at,omitempty"`
}

type CreateIssueParams struct {
	ProjectID     string
	IssueTypeID   string
	StatusID      string
	ParentIssueID string
	Title         string
	Description   string
	Priority      string
	AssigneeID    string
	ReporterID    string
	DueDate       *time.Time
}

func (p CreateIssueParams) Validate() error {
	if p.ProjectID == "" {
		return errors.New("project_id is required")
	}
	if p.IssueTypeID == "" {
		return errors.New("issue_type_id is required")
	}
	if p.StatusID == "" {
		return errors.New("status_id is required")
	}
	if p.Title == "" {
		return errors.New("title is required")
	}
	if p.ReporterID == "" {
		return errors.New("reporter_id is required")
	}
	priority := p.Priority
	if priority == "" {
		priority = "medium"
	}
	if !validPriorities[priority] {
		return ErrInvalidPriority
	}
	return nil
}

type UpdateIssueParams struct {
	IssueID     string
	ProjectID   string
	Title       string
	Description string
	Priority    string
	AssigneeID  *string
	DueDate     *time.Time
}

func (p UpdateIssueParams) Validate() error {
	if p.IssueID == "" {
		return errors.New("issue_id is required")
	}
	if p.ProjectID == "" {
		return errors.New("project_id is required")
	}
	if p.Title == "" {
		return errors.New("title is required")
	}
	if !validPriorities[p.Priority] {
		return ErrInvalidPriority
	}
	return nil
}

type ListIssuesParams struct {
	ProjectID  string
	StatusID   string
	AssigneeID string
}

func CreateIssue(ctx context.Context, db *sqlx.DB, p CreateIssueParams) (Issue, error) {
	if db == nil {
		return Issue{}, errors.New("db is required")
	}
	if err := p.Validate(); err != nil {
		return Issue{}, err
	}
	if p.Priority == "" {
		p.Priority = "medium"
	}
	return createIssue(ctx, db, p)
}

func GetIssue(ctx context.Context, db *sqlx.DB, projectID, issueID string) (Issue, error) {
	if db == nil {
		return Issue{}, errors.New("db is required")
	}
	if projectID == "" {
		return Issue{}, errors.New("project_id is required")
	}
	if issueID == "" {
		return Issue{}, errors.New("issue_id is required")
	}
	return getIssue(ctx, db, projectID, issueID)
}

func ListIssues(ctx context.Context, db *sqlx.DB, p ListIssuesParams) ([]Issue, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	if p.ProjectID == "" {
		return nil, errors.New("project_id is required")
	}
	return listIssues(ctx, db, p)
}

func UpdateIssue(ctx context.Context, db *sqlx.DB, p UpdateIssueParams) (Issue, error) {
	if db == nil {
		return Issue{}, errors.New("db is required")
	}
	if err := p.Validate(); err != nil {
		return Issue{}, err
	}
	return updateIssue(ctx, db, p)
}

func ArchiveIssue(ctx context.Context, db *sqlx.DB, projectID, issueID string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if projectID == "" {
		return errors.New("project_id is required")
	}
	if issueID == "" {
		return errors.New("issue_id is required")
	}
	return archiveIssue(ctx, db, projectID, issueID)
}

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
