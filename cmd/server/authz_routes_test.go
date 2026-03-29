// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/start-codex/trazawork/internal/authz"
	"github.com/start-codex/trazawork/internal/sessions"
	"github.com/start-codex/trazawork/internal/testpg"
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

type dataEnvelope struct {
	Status int             `json:"status"`
	Data   json.RawMessage `json:"data,omitempty"`
	Error  string          `json:"error,omitempty"`
}

func doRequestWithBody(t *testing.T, srv *httptest.Server, method, path, token string, body any) dataEnvelope {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req, err := http.NewRequest(method, srv.URL+path, &buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "session_id", Value: token})
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	var env dataEnvelope
	_ = json.NewDecoder(resp.Body).Decode(&env)
	env.Status = resp.StatusCode
	return env
}

// TestContractCleanup_WorkspaceCreate verifies POST /workspaces derives owner
// from session and ignores any spoofed owner_id in the body.
func TestContractCleanup_WorkspaceCreate(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)
	srv := setupTestServer(t, db)

	userID := testpg.SeedUser(t, db)
	token := loginCookie(t, db, userID)
	suffix := testpg.UniqueSuffix(t, db)

	// Create workspace without owner_id
	env := doRequestWithBody(t, srv, "POST", "/workspaces", token, map[string]string{
		"name": "WS " + suffix,
		"slug": "ws-" + suffix,
	})
	if env.Status != 201 {
		t.Fatalf("POST /workspaces without owner_id: status = %d, want 201 (error: %s)", env.Status, env.Error)
	}

	// Verify the authenticated user became the owner
	var ws struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(env.Data, &ws); err != nil {
		t.Fatalf("unmarshal workspace: %v", err)
	}
	var role string
	err := db.QueryRowContext(context.Background(),
		`SELECT role FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`,
		ws.ID, userID,
	).Scan(&role)
	if err != nil {
		t.Fatalf("query owner membership: %v", err)
	}
	if role != "owner" {
		t.Fatalf("owner role = %q, want %q", role, "owner")
	}

	// Spoofed owner_id should be ignored — user still becomes owner
	otherUser := testpg.SeedUser(t, db)
	suffix2 := testpg.UniqueSuffix(t, db)
	env = doRequestWithBody(t, srv, "POST", "/workspaces", token, map[string]string{
		"name":     "WS " + suffix2,
		"slug":     "ws-" + suffix2,
		"owner_id": otherUser,
	})
	if env.Status != 201 {
		t.Fatalf("POST /workspaces with spoofed owner_id: status = %d, want 201 (error: %s)", env.Status, env.Error)
	}
	var ws2 struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(env.Data, &ws2); err != nil {
		t.Fatalf("unmarshal workspace: %v", err)
	}
	err = db.QueryRowContext(context.Background(),
		`SELECT role FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`,
		ws2.ID, userID,
	).Scan(&role)
	if err != nil {
		t.Fatalf("query owner membership after spoof: %v", err)
	}
	if role != "owner" {
		t.Fatalf("spoofed owner_id: authenticated user role = %q, want %q", role, "owner")
	}
}

// TestContractCleanup_WorkspaceList verifies GET /workspaces derives user
// from session, returns only the authenticated user's workspaces, and ignores
// any spoofed user_id query parameter.
func TestContractCleanup_WorkspaceList(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)
	srv := setupTestServer(t, db)

	member := testpg.SeedUser(t, db)
	other := testpg.SeedUser(t, db)

	// Create two workspaces: one with member, one with other only
	memberWS := testpg.SeedWorkspace(t, db)
	otherWS := testpg.SeedWorkspace(t, db)
	seedMember(t, db, memberWS, member, "member")
	seedMember(t, db, otherWS, other, "member")

	token := loginCookie(t, db, member)

	assertOnlyMemberWorkspace := func(t *testing.T, path string) {
		t.Helper()

		env := doRequestWithBody(t, srv, "GET", path, token, nil)
		if env.Status != 200 {
			t.Fatalf("GET %s: status = %d, want 200 (error: %s)", path, env.Status, env.Error)
		}

		var wsList []struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(env.Data, &wsList); err != nil {
			t.Fatalf("unmarshal workspace list for %s: %v", path, err)
		}
		if len(wsList) != 1 {
			t.Fatalf("GET %s returned %d workspaces, want 1", path, len(wsList))
		}
		if wsList[0].ID != memberWS {
			t.Fatalf("GET %s returned workspace %s, want %s", path, wsList[0].ID, memberWS)
		}
		if wsList[0].ID == otherWS {
			t.Fatalf("GET %s returned other user's workspace %s", path, otherWS)
		}
	}

	// GET /workspaces → should return only member's workspace
	assertOnlyMemberWorkspace(t, "/workspaces")

	// GET /workspaces?user_id=other → query param ignored, still returns member's workspace only
	assertOnlyMemberWorkspace(t, "/workspaces?user_id="+other)
}

// TestContractCleanup_IssueCreate verifies POST /projects/{id}/issues derives
// reporter from session and ignores any spoofed reporter_id.
func TestContractCleanup_IssueCreate(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)
	srv := setupTestServer(t, db)

	member := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, member, "member")
	projID := testpg.SeedProject(t, db, wsID, "ISRC")
	token := loginCookie(t, db, member)

	// Seed a status and issue type for the project
	var statusID, issueTypeID string
	err := db.QueryRowContext(context.Background(),
		`INSERT INTO statuses (project_id, name, category, position) VALUES ($1, 'Todo', 'todo', 0) RETURNING id`,
		projID,
	).Scan(&statusID)
	if err != nil {
		t.Fatalf("seed status: %v", err)
	}
	err = db.QueryRowContext(context.Background(),
		`INSERT INTO issue_types (project_id, name, icon, level) VALUES ($1, 'Task', 'task', 0) RETURNING id`,
		projID,
	).Scan(&issueTypeID)
	if err != nil {
		t.Fatalf("seed issue type: %v", err)
	}

	// Create issue without reporter_id
	env := doRequestWithBody(t, srv, "POST", "/projects/"+projID+"/issues", token, map[string]string{
		"issue_type_id": issueTypeID,
		"status_id":     statusID,
		"title":         "Test issue",
	})
	if env.Status != 201 {
		t.Fatalf("POST /issues without reporter_id: status = %d, want 201 (error: %s)", env.Status, env.Error)
	}

	var issue struct {
		ID         string `json:"id"`
		ReporterID string `json:"reporter_id"`
	}
	if err := json.Unmarshal(env.Data, &issue); err != nil {
		t.Fatalf("unmarshal issue: %v", err)
	}
	if issue.ReporterID != member {
		t.Fatalf("reporter_id = %q, want %q (authenticated user)", issue.ReporterID, member)
	}

	// Spoofed reporter_id should be ignored
	otherUser := testpg.SeedUser(t, db)
	env = doRequestWithBody(t, srv, "POST", "/projects/"+projID+"/issues", token, map[string]string{
		"issue_type_id": issueTypeID,
		"status_id":     statusID,
		"title":         "Spoofed issue",
		"reporter_id":   otherUser,
	})
	if env.Status != 201 {
		t.Fatalf("POST /issues with spoofed reporter_id: status = %d, want 201 (error: %s)", env.Status, env.Error)
	}
	var issue2 struct {
		ReporterID string `json:"reporter_id"`
	}
	if err := json.Unmarshal(env.Data, &issue2); err != nil {
		t.Fatalf("unmarshal issue: %v", err)
	}
	if issue2.ReporterID != member {
		t.Fatalf("spoofed reporter_id: got %q, want %q (authenticated user)", issue2.ReporterID, member)
	}
}

// Ensure authz import is used (it's needed for the test to compile with the right module).
var _ = authz.ErrForbidden
