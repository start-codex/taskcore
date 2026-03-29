package authz

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

var (
	ErrUnauthenticated   = errors.New("unauthenticated")
	ErrForbidden         = errors.New("forbidden")
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrProjectNotFound   = errors.New("project not found")
	ErrBoardNotFound     = errors.New("board not found")
	ErrColumnNotFound    = errors.New("column not found")
)

type ctxKey struct{}

// WithUserID stores the authenticated user ID in the context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, userID)
}

// UserIDFromContext retrieves the authenticated user ID from the context.
// Returns ErrUnauthenticated if no user ID is present.
func UserIDFromContext(ctx context.Context) (string, error) {
	v, ok := ctx.Value(ctxKey{}).(string)
	if !ok || v == "" {
		return "", ErrUnauthenticated
	}
	return v, nil
}

// RequireWorkspaceMembership verifies that the authenticated user is a member
// of the given workspace. Returns ErrWorkspaceNotFound if the workspace does
// not exist (or is archived), ErrForbidden if the user is not a member.
func RequireWorkspaceMembership(ctx context.Context, db *sqlx.DB, workspaceID string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if workspaceID == "" {
		return errors.New("workspaceID is required")
	}
	userID, err := UserIDFromContext(ctx)
	if err != nil {
		return err
	}
	exists, err := workspaceExists(ctx, db, workspaceID)
	if err != nil {
		return fmt.Errorf("require workspace membership: %w", err)
	}
	if !exists {
		return ErrWorkspaceNotFound
	}
	member, err := isMember(ctx, db, workspaceID, userID)
	if err != nil {
		return fmt.Errorf("require workspace membership: %w", err)
	}
	if !member {
		return ErrForbidden
	}
	return nil
}

// RequireProjectMembership verifies that the authenticated user is a member
// of the workspace that owns the given project. Returns the resolved
// workspaceID on success.
func RequireProjectMembership(ctx context.Context, db *sqlx.DB, projectID string) (string, error) {
	if db == nil {
		return "", errors.New("db is required")
	}
	if projectID == "" {
		return "", errors.New("projectID is required")
	}
	wsID, err := projectWorkspaceID(ctx, db, projectID)
	if err != nil {
		return "", err
	}
	if err := RequireWorkspaceMembership(ctx, db, wsID); err != nil {
		return "", err
	}
	return wsID, nil
}

// RequireBoardAccess verifies that the authenticated user is a member of the
// workspace that owns the board's project. Returns workspaceID and projectID.
func RequireBoardAccess(ctx context.Context, db *sqlx.DB, boardID string) (string, string, error) {
	if db == nil {
		return "", "", errors.New("db is required")
	}
	if boardID == "" {
		return "", "", errors.New("boardID is required")
	}
	projID, err := boardProjectID(ctx, db, boardID)
	if err != nil {
		return "", "", err
	}
	wsID, err := RequireProjectMembership(ctx, db, projID)
	if err != nil {
		return "", "", err
	}
	return wsID, projID, nil
}

// RequireColumnAccess verifies that the authenticated user is a member of the
// workspace that owns the column's board's project. Returns workspaceID,
// projectID, and boardID.
func RequireColumnAccess(ctx context.Context, db *sqlx.DB, columnID string) (string, string, string, error) {
	if db == nil {
		return "", "", "", errors.New("db is required")
	}
	if columnID == "" {
		return "", "", "", errors.New("columnID is required")
	}
	bID, err := columnBoardID(ctx, db, columnID)
	if err != nil {
		return "", "", "", err
	}
	wsID, projID, err := RequireBoardAccess(ctx, db, bID)
	if err != nil {
		return "", "", "", err
	}
	return wsID, projID, bID, nil
}
