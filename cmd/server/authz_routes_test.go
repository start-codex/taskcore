package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/start-codex/taskcode/internal/authz"
	"github.com/start-codex/taskcode/internal/sessions"
	"github.com/start-codex/taskcode/internal/testpg"
)

// setupTestServer creates a test HTTP server with the full API handler stack.
func setupTestServer(t *testing.T, db *sqlx.DB) *httptest.Server {
	t.Helper()
	handler := newAPIHandler(db)
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

// loginCookie creates a session for the given user and returns the raw token.
func loginCookie(t *testing.T, db *sqlx.DB, userID string) string {
	t.Helper()
	result, err := sessions.Create(context.Background(), db, userID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	t.Cleanup(func() {
		_ = sessions.Delete(context.Background(), db, result.RawToken)
	})
	return result.RawToken
}

func seedMember(t *testing.T, db *sqlx.DB, workspaceID, userID, role string) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, $3)`,
		workspaceID, userID, role,
	)
	if err != nil {
		t.Fatalf("seed member: %v", err)
	}
}

func seedBoard(t *testing.T, db *sqlx.DB, projectID string) string {
	t.Helper()
	var id string
	err := db.QueryRowContext(context.Background(),
		`INSERT INTO boards (project_id, name, type, filter_query) VALUES ($1, $2, 'kanban', '') RETURNING id`,
		projectID, "Board "+testpg.UniqueSuffix(t, db),
	).Scan(&id)
	if err != nil {
		t.Fatalf("seed board: %v", err)
	}
	return id
}

type envelope struct {
	Status int    `json:"status"`
	Error  string `json:"error,omitempty"`
}

func doRequest(t *testing.T, srv *httptest.Server, method, path, token string) envelope {
	t.Helper()
	req, err := http.NewRequest(method, srv.URL+path, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "session_id", Value: token})
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	var env envelope
	_ = json.NewDecoder(resp.Body).Decode(&env)
	env.Status = resp.StatusCode
	return env
}

// TestAuthzWiring_WorkspaceGet verifies GET /workspaces/{id} returns
// 403 for non-members and 200 for members.
func TestAuthzWiring_WorkspaceGet(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)
	srv := setupTestServer(t, db)

	member := testpg.SeedUser(t, db)
	nonMember := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, member, "member")

	memberToken := loginCookie(t, db, member)
	nonMemberToken := loginCookie(t, db, nonMember)

	// Member → 200
	env := doRequest(t, srv, "GET", "/workspaces/"+wsID, memberToken)
	if env.Status != 200 {
		t.Fatalf("member GET /workspaces/{id}: status = %d, want 200 (error: %s)", env.Status, env.Error)
	}

	// Non-member → 403
	env = doRequest(t, srv, "GET", "/workspaces/"+wsID, nonMemberToken)
	if env.Status != 403 {
		t.Fatalf("non-member GET /workspaces/{id}: status = %d, want 403", env.Status)
	}
}

// TestAuthzWiring_ProjectGet verifies GET /projects/{id} returns
// 403 for non-members and 200 for members.
func TestAuthzWiring_ProjectGet(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)
	srv := setupTestServer(t, db)

	member := testpg.SeedUser(t, db)
	nonMember := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, member, "member")
	projID := testpg.SeedProject(t, db, wsID, "PRJAZ")

	memberToken := loginCookie(t, db, member)
	nonMemberToken := loginCookie(t, db, nonMember)

	env := doRequest(t, srv, "GET", "/projects/"+projID, memberToken)
	if env.Status != 200 {
		t.Fatalf("member GET /projects/{id}: status = %d, want 200 (error: %s)", env.Status, env.Error)
	}

	env = doRequest(t, srv, "GET", "/projects/"+projID, nonMemberToken)
	if env.Status != 403 {
		t.Fatalf("non-member GET /projects/{id}: status = %d, want 403", env.Status)
	}
}

// TestAuthzWiring_ProjectBoards verifies GET /projects/{id}/boards returns
// 403 for non-members and 200 for members.
func TestAuthzWiring_ProjectBoards(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)
	srv := setupTestServer(t, db)

	member := testpg.SeedUser(t, db)
	nonMember := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, member, "member")
	projID := testpg.SeedProject(t, db, wsID, "BRDWI")

	memberToken := loginCookie(t, db, member)
	nonMemberToken := loginCookie(t, db, nonMember)

	env := doRequest(t, srv, "GET", "/projects/"+projID+"/boards", memberToken)
	if env.Status != 200 {
		t.Fatalf("member GET /projects/{id}/boards: status = %d, want 200 (error: %s)", env.Status, env.Error)
	}

	env = doRequest(t, srv, "GET", "/projects/"+projID+"/boards", nonMemberToken)
	if env.Status != 403 {
		t.Fatalf("non-member GET /projects/{id}/boards: status = %d, want 403", env.Status)
	}
}

// TestAuthzWiring_BoardGet verifies GET /boards/{id} returns
// 403 for non-members and 200 for members.
func TestAuthzWiring_BoardGet(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)
	srv := setupTestServer(t, db)

	member := testpg.SeedUser(t, db)
	nonMember := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, member, "member")
	projID := testpg.SeedProject(t, db, wsID, "BRDGT")
	boardID := seedBoard(t, db, projID)

	memberToken := loginCookie(t, db, member)
	nonMemberToken := loginCookie(t, db, nonMember)

	env := doRequest(t, srv, "GET", "/boards/"+boardID, memberToken)
	if env.Status != 200 {
		t.Fatalf("member GET /boards/{id}: status = %d, want 200 (error: %s)", env.Status, env.Error)
	}

	env = doRequest(t, srv, "GET", "/boards/"+boardID, nonMemberToken)
	if env.Status != 403 {
		t.Fatalf("non-member GET /boards/{id}: status = %d, want 403", env.Status)
	}
}

// Ensure authz import is used (it's needed for the test to compile with the right module).
var _ = authz.ErrForbidden
