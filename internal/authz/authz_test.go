package authz

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
)

func fakeDB(t *testing.T) *sqlx.DB {
	t.Helper()
	raw, err := sql.Open("postgres", "")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	return sqlx.NewDb(raw, "postgres")
}

func TestWithUserIDRoundTrip(t *testing.T) {
	ctx := WithUserID(context.Background(), "user-123")
	got, err := UserIDFromContext(ctx)
	if err != nil {
		t.Fatalf("UserIDFromContext() error = %v", err)
	}
	if got != "user-123" {
		t.Fatalf("UserIDFromContext() = %q, want %q", got, "user-123")
	}
}

func TestUserIDFromContext_Missing(t *testing.T) {
	_, err := UserIDFromContext(context.Background())
	if !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("UserIDFromContext() error = %v, want ErrUnauthenticated", err)
	}
}

func TestUserIDFromContext_EmptyString(t *testing.T) {
	ctx := WithUserID(context.Background(), "")
	_, err := UserIDFromContext(ctx)
	if !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("UserIDFromContext() error = %v, want ErrUnauthenticated", err)
	}
}

func TestSentinelsAreDistinct(t *testing.T) {
	sentinels := []error{
		ErrUnauthenticated,
		ErrForbidden,
		ErrWorkspaceNotFound,
		ErrProjectNotFound,
		ErrBoardNotFound,
		ErrColumnNotFound,
	}
	for i, a := range sentinels {
		for j, b := range sentinels {
			if i != j && errors.Is(a, b) {
				t.Fatalf("sentinels %q and %q should not match", a, b)
			}
		}
	}
}

func TestRequireWorkspaceMembership_Guards(t *testing.T) {
	ctx := WithUserID(context.Background(), "user-1")
	tests := []struct {
		name        string
		db          *sqlx.DB
		workspaceID string
		wantErr     string
	}{
		{name: "nil db", db: nil, workspaceID: "ws-1", wantErr: "db is required"},
		{name: "empty workspaceID", db: fakeDB(t), workspaceID: "", wantErr: "workspaceID is required"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := RequireWorkspaceMembership(ctx, tc.db, tc.workspaceID)
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestRequireWorkspaceMembership_NoContext(t *testing.T) {
	err := RequireWorkspaceMembership(context.Background(), fakeDB(t), "ws-1")
	if !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("error = %v, want ErrUnauthenticated", err)
	}
}

func TestRequireProjectMembership_Guards(t *testing.T) {
	ctx := WithUserID(context.Background(), "user-1")
	tests := []struct {
		name      string
		db        *sqlx.DB
		projectID string
		wantErr   string
	}{
		{name: "nil db", db: nil, projectID: "p-1", wantErr: "db is required"},
		{name: "empty projectID", db: fakeDB(t), projectID: "", wantErr: "projectID is required"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := RequireProjectMembership(ctx, tc.db, tc.projectID)
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestRequireBoardAccess_Guards(t *testing.T) {
	ctx := WithUserID(context.Background(), "user-1")
	tests := []struct {
		name    string
		db      *sqlx.DB
		boardID string
		wantErr string
	}{
		{name: "nil db", db: nil, boardID: "b-1", wantErr: "db is required"},
		{name: "empty boardID", db: fakeDB(t), boardID: "", wantErr: "boardID is required"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := RequireBoardAccess(ctx, tc.db, tc.boardID)
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestRequireColumnAccess_Guards(t *testing.T) {
	ctx := WithUserID(context.Background(), "user-1")
	tests := []struct {
		name     string
		db       *sqlx.DB
		columnID string
		wantErr  string
	}{
		{name: "nil db", db: nil, columnID: "c-1", wantErr: "db is required"},
		{name: "empty columnID", db: fakeDB(t), columnID: "", wantErr: "columnID is required"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, _, err := RequireColumnAccess(ctx, tc.db, tc.columnID)
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}
