package boards

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const boardCols = `id, project_id, name, type, filter_query, created_at, updated_at, archived_at`
const columnCols = `id, board_id, name, position, created_at, updated_at, archived_at`

func createBoard(ctx context.Context, db *sqlx.DB, params CreateBoardParams) (Board, error) {
	var board Board
	err := db.QueryRowxContext(
		ctx,
		`INSERT INTO boards (project_id, name, type, filter_query)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+boardCols,
		params.ProjectID,
		params.Name,
		params.Type,
		params.FilterQuery,
	).StructScan(&board)
	if err != nil {
		if isUniqueViolation(err) {
			return Board{}, ErrDuplicateBoardName
		}
		return Board{}, fmt.Errorf("insert board: %w", err)
	}
	return board, nil
}

func getBoard(ctx context.Context, db *sqlx.DB, id string) (Board, error) {
	var board Board
	err := db.GetContext(
		ctx,
		&board,
		`SELECT `+boardCols+`
		 FROM boards
		 WHERE id = $1`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Board{}, ErrBoardNotFound
		}
		return Board{}, fmt.Errorf("get board: %w", err)
	}
	return board, nil
}

func listBoards(ctx context.Context, db *sqlx.DB, projectID string) ([]Board, error) {
	var boards []Board
	err := db.SelectContext(
		ctx,
		&boards,
		`SELECT `+boardCols+`
		 FROM boards
		 WHERE project_id = $1
		   AND archived_at IS NULL
		 ORDER BY created_at ASC`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list boards: %w", err)
	}
	return boards, nil
}

func archiveBoard(ctx context.Context, db *sqlx.DB, id string) error {
	res, err := db.ExecContext(
		ctx,
		`UPDATE boards
		 SET archived_at = NOW()
		 WHERE id = $1
		   AND archived_at IS NULL`,
		id,
	)
	if err != nil {
		return fmt.Errorf("archive board: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive board rows affected: %w", err)
	}
	if n == 0 {
		return ErrBoardNotFound
	}
	return nil
}

func addColumn(ctx context.Context, db *sqlx.DB, params AddColumnParams) (BoardColumn, error) {
	var column BoardColumn
	err := db.QueryRowxContext(
		ctx,
		`INSERT INTO board_columns (board_id, name, position)
		 VALUES (
		   $1, $2,
		   (SELECT COALESCE(MAX(position), -1) + 1
		    FROM board_columns
		    WHERE board_id = $1 AND archived_at IS NULL)
		 )
		 RETURNING `+columnCols,
		params.BoardID,
		params.Name,
	).StructScan(&column)
	if err != nil {
		if isUniqueViolation(err) {
			return BoardColumn{}, ErrDuplicateColumnName
		}
		return BoardColumn{}, fmt.Errorf("insert board column: %w", err)
	}
	return column, nil
}

func listColumns(ctx context.Context, db *sqlx.DB, boardID string) ([]BoardColumn, error) {
	var columns []BoardColumn
	err := db.SelectContext(
		ctx,
		&columns,
		`SELECT `+columnCols+`
		 FROM board_columns
		 WHERE board_id = $1
		   AND archived_at IS NULL
		 ORDER BY position ASC`,
		boardID,
	)
	if err != nil {
		return nil, fmt.Errorf("list board columns: %w", err)
	}
	return columns, nil
}

func archiveColumn(ctx context.Context, db *sqlx.DB, id string) error {
	res, err := db.ExecContext(
		ctx,
		`UPDATE board_columns
		 SET archived_at = NOW()
		 WHERE id = $1
		   AND archived_at IS NULL`,
		id,
	)
	if err != nil {
		return fmt.Errorf("archive board column: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive board column rows affected: %w", err)
	}
	if n == 0 {
		return ErrColumnNotFound
	}
	return nil
}

func assignStatus(ctx context.Context, db *sqlx.DB, boardColumnID, statusID string) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO board_column_statuses (board_column_id, status_id)
		 VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`,
		boardColumnID,
		statusID,
	)
	if err != nil {
		return fmt.Errorf("assign status to column: %w", err)
	}
	return nil
}

func unassignStatus(ctx context.Context, db *sqlx.DB, boardColumnID, statusID string) error {
	_, err := db.ExecContext(
		ctx,
		`DELETE FROM board_column_statuses
		 WHERE board_column_id = $1
		   AND status_id = $2`,
		boardColumnID,
		statusID,
	)
	if err != nil {
		return fmt.Errorf("unassign status from column: %w", err)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
