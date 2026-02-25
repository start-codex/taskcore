package boards

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrBoardNotFound       = errors.New("board not found")
	ErrColumnNotFound      = errors.New("board column not found")
	ErrDuplicateBoardName  = errors.New("board name already exists in project")
	ErrDuplicateColumnName = errors.New("column name already exists in board")
)

var validBoardTypes = map[string]bool{"kanban": true, "scrum": true}

type Board struct {
	ID          string     `db:"id"           json:"id"`
	ProjectID   string     `db:"project_id"   json:"project_id"`
	Name        string     `db:"name"         json:"name"`
	Type        string     `db:"type"         json:"type"`
	FilterQuery string     `db:"filter_query" json:"filter_query"`
	CreatedAt   time.Time  `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"   json:"updated_at"`
	ArchivedAt  *time.Time `db:"archived_at"  json:"archived_at,omitempty"`
}

type BoardColumn struct {
	ID         string     `db:"id"          json:"id"`
	BoardID    string     `db:"board_id"    json:"board_id"`
	Name       string     `db:"name"        json:"name"`
	Position   int        `db:"position"    json:"position"`
	CreatedAt  time.Time  `db:"created_at"  json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"  json:"updated_at"`
	ArchivedAt *time.Time `db:"archived_at" json:"archived_at,omitempty"`
}

type CreateBoardParams struct {
	ProjectID   string
	Name        string
	Type        string
	FilterQuery string
}

func (p CreateBoardParams) Validate() error {
	if p.ProjectID == "" {
		return errors.New("project_id is required")
	}
	if p.Name == "" {
		return errors.New("name is required")
	}
	if !validBoardTypes[p.Type] {
		return errors.New("type must be 'kanban' or 'scrum'")
	}
	return nil
}

type AddColumnParams struct {
	BoardID string
	Name    string
}

func (p AddColumnParams) Validate() error {
	if p.BoardID == "" {
		return errors.New("board_id is required")
	}
	if p.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

func CreateBoard(ctx context.Context, db *sqlx.DB, p CreateBoardParams) (Board, error) {
	if db == nil {
		return Board{}, errors.New("db is required")
	}
	if err := p.Validate(); err != nil {
		return Board{}, err
	}
	return createBoard(ctx, db, p)
}

func GetBoard(ctx context.Context, db *sqlx.DB, id string) (Board, error) {
	if db == nil {
		return Board{}, errors.New("db is required")
	}
	if id == "" {
		return Board{}, errors.New("id is required")
	}
	return getBoard(ctx, db, id)
}

func ListBoards(ctx context.Context, db *sqlx.DB, projectID string) ([]Board, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	if projectID == "" {
		return nil, errors.New("project_id is required")
	}
	return listBoards(ctx, db, projectID)
}

func ArchiveBoard(ctx context.Context, db *sqlx.DB, id string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if id == "" {
		return errors.New("id is required")
	}
	return archiveBoard(ctx, db, id)
}

func AddColumn(ctx context.Context, db *sqlx.DB, p AddColumnParams) (BoardColumn, error) {
	if db == nil {
		return BoardColumn{}, errors.New("db is required")
	}
	if err := p.Validate(); err != nil {
		return BoardColumn{}, err
	}
	return addColumn(ctx, db, p)
}

func ListColumns(ctx context.Context, db *sqlx.DB, boardID string) ([]BoardColumn, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	if boardID == "" {
		return nil, errors.New("board_id is required")
	}
	return listColumns(ctx, db, boardID)
}

func ArchiveColumn(ctx context.Context, db *sqlx.DB, id string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if id == "" {
		return errors.New("id is required")
	}
	return archiveColumn(ctx, db, id)
}

func AssignStatus(ctx context.Context, db *sqlx.DB, boardColumnID, statusID string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if boardColumnID == "" {
		return errors.New("board_column_id is required")
	}
	if statusID == "" {
		return errors.New("status_id is required")
	}
	return assignStatus(ctx, db, boardColumnID, statusID)
}

func UnassignStatus(ctx context.Context, db *sqlx.DB, boardColumnID, statusID string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if boardColumnID == "" {
		return errors.New("board_column_id is required")
	}
	if statusID == "" {
		return errors.New("status_id is required")
	}
	return unassignStatus(ctx, db, boardColumnID, statusID)
}
