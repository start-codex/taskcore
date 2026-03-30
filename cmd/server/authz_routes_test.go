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
	"github.com/start-codex/tookly/internal/authz"
	"github.com/start-codex/tookly/internal/sessions"
	"github.com/start-codex/tookly/internal/testpg"
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

// TestAdminWiring_WorkspaceArchive verifies DELETE /workspaces/{id}
// returns 403 for members and 204 for admins.
func TestAdminWiring_WorkspaceArchive(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)
	srv := setupTestServer(t, db)

	admin := testpg.SeedUser(t, db)
	member := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, admin, "admin")
	seedMember(t, db, wsID, member, "member")

	memberToken := loginCookie(t, db, member)
	adminToken := loginCookie(t, db, admin)

	// Member → 403
	env := doRequest(t, srv, "DELETE", "/workspaces/"+wsID, memberToken)
	if env.Status != 403 {
		t.Fatalf("member DELETE /workspaces/{id}: status = %d, want 403", env.Status)
	}

	// Admin → 204
	env = doRequest(t, srv, "DELETE", "/workspaces/"+wsID, adminToken)
	if env.Status != 204 {
		t.Fatalf("admin DELETE /workspaces/{id}: status = %d, want 204 (error: %s)", env.Status, env.Error)
	}
}

// TestAdminWiring_WorkspaceMembers verifies GET/POST /workspaces/{id}/members
// returns 403 for members and success for admins.
func TestAdminWiring_WorkspaceMembers(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)
	srv := setupTestServer(t, db)

	admin := testpg.SeedUser(t, db)
	member := testpg.SeedUser(t, db)
	newUser := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, admin, "admin")
	seedMember(t, db, wsID, member, "member")

	memberToken := loginCookie(t, db, member)
	adminToken := loginCookie(t, db, admin)

	// Member GET /workspaces/{id}/members → 403
	env := doRequest(t, srv, "GET", "/workspaces/"+wsID+"/members", memberToken)
	if env.Status != 403 {
		t.Fatalf("member GET /workspaces/{id}/members: status = %d, want 403", env.Status)
	}

	// Admin GET /workspaces/{id}/members → 200
	env = doRequest(t, srv, "GET", "/workspaces/"+wsID+"/members", adminToken)
	if env.Status != 200 {
		t.Fatalf("admin GET /workspaces/{id}/members: status = %d, want 200 (error: %s)", env.Status, env.Error)
	}

	// Member POST /workspaces/{id}/members → 403
	envD := doRequestWithBody(t, srv, "POST", "/workspaces/"+wsID+"/members", memberToken, map[string]string{
		"user_id": newUser, "role": "member",
	})
	if envD.Status != 403 {
		t.Fatalf("member POST /workspaces/{id}/members: status = %d, want 403", envD.Status)
	}

	// Admin POST /workspaces/{id}/members → 201
	envD = doRequestWithBody(t, srv, "POST", "/workspaces/"+wsID+"/members", adminToken, map[string]string{
		"user_id": newUser, "role": "member",
	})
	if envD.Status != 201 {
		t.Fatalf("admin POST /workspaces/{id}/members: status = %d, want 201 (error: %s)", envD.Status, envD.Error)
	}
}

// TestAdminWiring_ProjectCreate verifies POST /workspaces/{id}/projects
// returns 403 for members and 201 for admins.
func TestAdminWiring_ProjectCreate(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)
	srv := setupTestServer(t, db)

	admin := testpg.SeedUser(t, db)
	member := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, admin, "admin")
	seedMember(t, db, wsID, member, "member")

	memberToken := loginCookie(t, db, member)
	adminToken := loginCookie(t, db, admin)

	suffix := testpg.UniqueSuffix(t, db)

	// Member → 403
	envD := doRequestWithBody(t, srv, "POST", "/workspaces/"+wsID+"/projects", memberToken, map[string]string{
		"name": "Proj " + suffix, "key": "PM" + suffix[:2],
	})
	if envD.Status != 403 {
		t.Fatalf("member POST /workspaces/{id}/projects: status = %d, want 403", envD.Status)
	}

	// Admin → 201
	envD = doRequestWithBody(t, srv, "POST", "/workspaces/"+wsID+"/projects", adminToken, map[string]string{
		"name": "Proj " + suffix, "key": "PA" + suffix[:2],
	})
	if envD.Status != 201 {
		t.Fatalf("admin POST /workspaces/{id}/projects: status = %d, want 201 (error: %s)", envD.Status, envD.Error)
	}
}

// TestAdminWiring_ProjectArchive verifies DELETE /projects/{id}
// returns 403 for members and 204 for admins.
func TestAdminWiring_ProjectArchive(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)
	srv := setupTestServer(t, db)

	admin := testpg.SeedUser(t, db)
	member := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, admin, "admin")
	seedMember(t, db, wsID, member, "member")
	projID := testpg.SeedProject(t, db, wsID, "PADL")

	memberToken := loginCookie(t, db, member)
	adminToken := loginCookie(t, db, admin)

	// Member → 403
	env := doRequest(t, srv, "DELETE", "/projects/"+projID, memberToken)
	if env.Status != 403 {
		t.Fatalf("member DELETE /projects/{id}: status = %d, want 403", env.Status)
	}

	// Admin → 204
	env = doRequest(t, srv, "DELETE", "/projects/"+projID, adminToken)
	if env.Status != 204 {
		t.Fatalf("admin DELETE /projects/{id}: status = %d, want 204 (error: %s)", env.Status, env.Error)
	}
}

// TestAdminWiring_ProjectMembers verifies GET/POST /projects/{id}/members
// returns 403 for members and success for admins.
func TestAdminWiring_ProjectMembers(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)
	srv := setupTestServer(t, db)

	admin := testpg.SeedUser(t, db)
	member := testpg.SeedUser(t, db)
	newUser := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, admin, "admin")
	seedMember(t, db, wsID, member, "member")
	seedMember(t, db, wsID, newUser, "member")
	projID := testpg.SeedProject(t, db, wsID, "PMBR")

	memberToken := loginCookie(t, db, member)
	adminToken := loginCookie(t, db, admin)

	// Member GET /projects/{id}/members → 403
	env := doRequest(t, srv, "GET", "/projects/"+projID+"/members", memberToken)
	if env.Status != 403 {
		t.Fatalf("member GET /projects/{id}/members: status = %d, want 403", env.Status)
	}

	// Admin GET /projects/{id}/members → 200
	env = doRequest(t, srv, "GET", "/projects/"+projID+"/members", adminToken)
	if env.Status != 200 {
		t.Fatalf("admin GET /projects/{id}/members: status = %d, want 200 (error: %s)", env.Status, env.Error)
	}

	// Member POST /projects/{id}/members → 403
	envD := doRequestWithBody(t, srv, "POST", "/projects/"+projID+"/members", memberToken, map[string]string{
		"user_id": newUser, "role": "member",
	})
	if envD.Status != 403 {
		t.Fatalf("member POST /projects/{id}/members: status = %d, want 403", envD.Status)
	}

	// Admin POST /projects/{id}/members → 201
	envD = doRequestWithBody(t, srv, "POST", "/projects/"+projID+"/members", adminToken, map[string]string{
		"user_id": newUser, "role": "member",
	})
	if envD.Status != 201 {
		t.Fatalf("admin POST /projects/{id}/members: status = %d, want 201 (error: %s)", envD.Status, envD.Error)
	}
}

// TestAdminWiring_MemberReadNoRegression verifies GET /workspaces/{id} and
// GET /projects/{id} still return 200 for members (not broken by admin checks).
func TestAdminWiring_MemberReadNoRegression(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)
	srv := setupTestServer(t, db)

	member := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, member, "member")
	projID := testpg.SeedProject(t, db, wsID, "NREG")

	token := loginCookie(t, db, member)

	env := doRequest(t, srv, "GET", "/workspaces/"+wsID, token)
	if env.Status != 200 {
		t.Fatalf("member GET /workspaces/{id}: status = %d, want 200", env.Status)
	}

	env = doRequest(t, srv, "GET", "/projects/"+projID, token)
	if env.Status != 200 {
		t.Fatalf("member GET /projects/{id}: status = %d, want 200", env.Status)
	}
}

// TestDueDate_CreateAndUpdate verifies due_date can be set on create,
// changed on update, and cleared by sending null.
func TestDueDate_CreateAndUpdate(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)
	srv := setupTestServer(t, db)

	user := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, user, "member")
	projID := testpg.SeedProject(t, db, wsID, "DUED")
	token := loginCookie(t, db, user)

	statusID := seedStatus(t, db, projID)
	issueTypeID := seedIssueType(t, db, projID)

	// Create issue with due_date
	env := doRequestWithBody(t, srv, "POST", "/projects/"+projID+"/issues", token, map[string]any{
		"issue_type_id": issueTypeID,
		"status_id":     statusID,
		"title":         "Issue with date",
		"due_date":      "2026-04-15",
	})
	if env.Status != 201 {
		t.Fatalf("create with due_date: status = %d, want 201 (error: %s)", env.Status, env.Error)
	}
	var created struct {
		ID      string  `json:"id"`
		DueDate *string `json:"due_date"`
	}
	if err := json.Unmarshal(env.Data, &created); err != nil {
		t.Fatalf("unmarshal created issue: %v", err)
	}
	if created.DueDate == nil || (*created.DueDate)[:10] != "2026-04-15" {
		t.Fatalf("created due_date = %v, want 2026-04-15", created.DueDate)
	}

	// Update due_date to a different value
	env = doRequestWithBody(t, srv, "PUT", "/projects/"+projID+"/issues/"+created.ID, token, map[string]any{
		"title":    "Issue with date",
		"priority": "medium",
		"due_date": "2026-05-01",
	})
	if env.Status != 200 {
		t.Fatalf("update due_date: status = %d, want 200 (error: %s)", env.Status, env.Error)
	}
	var updated struct {
		DueDate *string `json:"due_date"`
	}
	if err := json.Unmarshal(env.Data, &updated); err != nil {
		t.Fatalf("unmarshal updated issue: %v", err)
	}
	if updated.DueDate == nil || (*updated.DueDate)[:10] != "2026-05-01" {
		t.Fatalf("updated due_date = %v, want 2026-05-01", updated.DueDate)
	}

	// Clear due_date by sending null
	env = doRequestWithBody(t, srv, "PUT", "/projects/"+projID+"/issues/"+created.ID, token, map[string]any{
		"title":    "Issue with date",
		"priority": "medium",
		"due_date": nil,
	})
	if env.Status != 200 {
		t.Fatalf("clear due_date: status = %d, want 200 (error: %s)", env.Status, env.Error)
	}
	var cleared struct {
		DueDate *string `json:"due_date"`
	}
	if err := json.Unmarshal(env.Data, &cleared); err != nil {
		t.Fatalf("unmarshal cleared issue: %v", err)
	}
	if cleared.DueDate != nil {
		t.Fatalf("cleared due_date = %v, want nil", *cleared.DueDate)
	}

	// Create issue without due_date (should work, it's optional)
	env = doRequestWithBody(t, srv, "POST", "/projects/"+projID+"/issues", token, map[string]any{
		"issue_type_id": issueTypeID,
		"status_id":     statusID,
		"title":         "Issue no date",
	})
	if env.Status != 201 {
		t.Fatalf("create without due_date: status = %d, want 201 (error: %s)", env.Status, env.Error)
	}
}

func seedColumn(t *testing.T, db *sqlx.DB, boardID string) string {
	t.Helper()
	var id string
	err := db.QueryRowContext(context.Background(),
		`INSERT INTO board_columns (board_id, name, position) VALUES ($1, $2, 0) RETURNING id`,
		boardID, "Col "+testpg.UniqueSuffix(t, db),
	).Scan(&id)
	if err != nil {
		t.Fatalf("seed column: %v", err)
	}
	return id
}

func seedStatus(t *testing.T, db *sqlx.DB, projectID string) string {
	t.Helper()
	var id string
	err := db.QueryRowContext(context.Background(),
		`INSERT INTO statuses (project_id, name, category, position) VALUES ($1, $2, 'todo', 0) RETURNING id`,
		projectID, "Status "+testpg.UniqueSuffix(t, db),
	).Scan(&id)
	if err != nil {
		t.Fatalf("seed status: %v", err)
	}
	return id
}

func seedIssueType(t *testing.T, db *sqlx.DB, projectID string) string {
	t.Helper()
	var id string
	err := db.QueryRowContext(context.Background(),
		`INSERT INTO issue_types (project_id, name, icon, level) VALUES ($1, $2, 'task', 0) RETURNING id`,
		projectID, "Type "+testpg.UniqueSuffix(t, db),
	).Scan(&id)
	if err != nil {
		t.Fatalf("seed issue type: %v", err)
	}
	return id
}

// TestWorkflowAdmin_Boards verifies board write routes require admin.
func TestWorkflowAdmin_Boards(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)
	srv := setupTestServer(t, db)

	admin := testpg.SeedUser(t, db)
	member := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, admin, "admin")
	seedMember(t, db, wsID, member, "member")
	projID := testpg.SeedProject(t, db, wsID, "BWFL")
	boardID := seedBoard(t, db, projID)
	colID := seedColumn(t, db, boardID)
	statusID := seedStatus(t, db, projID)

	memberToken := loginCookie(t, db, member)
	adminToken := loginCookie(t, db, admin)

	// POST /projects/{id}/boards — member 403, admin 201
	envD := doRequestWithBody(t, srv, "POST", "/projects/"+projID+"/boards", memberToken, map[string]string{
		"name": "B1", "type": "kanban",
	})
	if envD.Status != 403 {
		t.Fatalf("member POST boards: %d, want 403", envD.Status)
	}
	envD = doRequestWithBody(t, srv, "POST", "/projects/"+projID+"/boards", adminToken, map[string]string{
		"name": "B2", "type": "kanban",
	})
	if envD.Status != 201 {
		t.Fatalf("admin POST boards: %d, want 201 (error: %s)", envD.Status, envD.Error)
	}

	// DELETE /boards/{id} — member 403, admin 204
	env := doRequest(t, srv, "DELETE", "/boards/"+boardID, memberToken)
	if env.Status != 403 {
		t.Fatalf("member DELETE board: %d, want 403", env.Status)
	}
	env = doRequest(t, srv, "DELETE", "/boards/"+boardID, adminToken)
	if env.Status != 204 {
		t.Fatalf("admin DELETE board: %d, want 204 (error: %s)", env.Status, env.Error)
	}

	// POST /boards/{id}/columns — member 403, admin 201
	boardID2 := seedBoard(t, db, projID)
	envD = doRequestWithBody(t, srv, "POST", "/boards/"+boardID2+"/columns", memberToken, map[string]string{
		"name": "C1",
	})
	if envD.Status != 403 {
		t.Fatalf("member POST column: %d, want 403", envD.Status)
	}
	envD = doRequestWithBody(t, srv, "POST", "/boards/"+boardID2+"/columns", adminToken, map[string]string{
		"name": "C2",
	})
	if envD.Status != 201 {
		t.Fatalf("admin POST column: %d, want 201 (error: %s)", envD.Status, envD.Error)
	}

	// DELETE /columns/{id} — member 403, admin 204
	env = doRequest(t, srv, "DELETE", "/columns/"+colID, memberToken)
	if env.Status != 403 {
		t.Fatalf("member DELETE column: %d, want 403", env.Status)
	}
	env = doRequest(t, srv, "DELETE", "/columns/"+colID, adminToken)
	if env.Status != 204 {
		t.Fatalf("admin DELETE column: %d, want 204 (error: %s)", env.Status, env.Error)
	}

	// POST /columns/{id}/statuses — member 403, admin 204
	colID2 := seedColumn(t, db, boardID2)
	envD = doRequestWithBody(t, srv, "POST", "/columns/"+colID2+"/statuses", memberToken, map[string]string{
		"status_id": statusID,
	})
	if envD.Status != 403 {
		t.Fatalf("member POST assign status: %d, want 403", envD.Status)
	}
	envD = doRequestWithBody(t, srv, "POST", "/columns/"+colID2+"/statuses", adminToken, map[string]string{
		"status_id": statusID,
	})
	if envD.Status != 204 {
		t.Fatalf("admin POST assign status: %d, want 204 (error: %s)", envD.Status, envD.Error)
	}

	// DELETE /columns/{id}/statuses/{statusID} — member 403, admin 204
	env = doRequest(t, srv, "DELETE", "/columns/"+colID2+"/statuses/"+statusID, memberToken)
	if env.Status != 403 {
		t.Fatalf("member DELETE unassign status: %d, want 403", env.Status)
	}
	env = doRequest(t, srv, "DELETE", "/columns/"+colID2+"/statuses/"+statusID, adminToken)
	if env.Status != 204 {
		t.Fatalf("admin DELETE unassign status: %d, want 204 (error: %s)", env.Status, env.Error)
	}

	// GET /boards/{id} — member 200 (no regression)
	env = doRequest(t, srv, "GET", "/boards/"+boardID2, memberToken)
	if env.Status != 200 {
		t.Fatalf("member GET board: %d, want 200", env.Status)
	}

	// GET /boards/{id}/columns — member 200 (no regression)
	env = doRequest(t, srv, "GET", "/boards/"+boardID2+"/columns", memberToken)
	if env.Status != 200 {
		t.Fatalf("member GET columns: %d, want 200", env.Status)
	}
}

// TestWorkflowAdmin_Statuses verifies status write routes require admin.
func TestWorkflowAdmin_Statuses(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)
	srv := setupTestServer(t, db)

	admin := testpg.SeedUser(t, db)
	member := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, admin, "admin")
	seedMember(t, db, wsID, member, "member")
	projID := testpg.SeedProject(t, db, wsID, "SWFL")
	statusID := seedStatus(t, db, projID)

	memberToken := loginCookie(t, db, member)
	adminToken := loginCookie(t, db, admin)

	suffix := testpg.UniqueSuffix(t, db)

	// POST /projects/{id}/statuses — member 403, admin 201
	envD := doRequestWithBody(t, srv, "POST", "/projects/"+projID+"/statuses", memberToken, map[string]string{
		"name": "S " + suffix, "category": "todo",
	})
	if envD.Status != 403 {
		t.Fatalf("member POST status: %d, want 403", envD.Status)
	}
	envD = doRequestWithBody(t, srv, "POST", "/projects/"+projID+"/statuses", adminToken, map[string]string{
		"name": "S " + suffix, "category": "in_progress",
	})
	if envD.Status != 201 {
		t.Fatalf("admin POST status: %d, want 201 (error: %s)", envD.Status, envD.Error)
	}

	// PUT /projects/{id}/statuses/{statusID} — member 403, admin 200
	envD = doRequestWithBody(t, srv, "PUT", "/projects/"+projID+"/statuses/"+statusID, memberToken, map[string]string{
		"name": "Updated", "category": "done",
	})
	if envD.Status != 403 {
		t.Fatalf("member PUT status: %d, want 403", envD.Status)
	}
	envD = doRequestWithBody(t, srv, "PUT", "/projects/"+projID+"/statuses/"+statusID, adminToken, map[string]string{
		"name": "Updated", "category": "done",
	})
	if envD.Status != 200 {
		t.Fatalf("admin PUT status: %d, want 200 (error: %s)", envD.Status, envD.Error)
	}

	// DELETE /projects/{id}/statuses/{statusID} — member 403, admin 204
	env := doRequest(t, srv, "DELETE", "/projects/"+projID+"/statuses/"+statusID, memberToken)
	if env.Status != 403 {
		t.Fatalf("member DELETE status: %d, want 403", env.Status)
	}
	env = doRequest(t, srv, "DELETE", "/projects/"+projID+"/statuses/"+statusID, adminToken)
	if env.Status != 204 {
		t.Fatalf("admin DELETE status: %d, want 204 (error: %s)", env.Status, env.Error)
	}

	// GET /projects/{id}/statuses — member 200 (no regression)
	env = doRequest(t, srv, "GET", "/projects/"+projID+"/statuses", memberToken)
	if env.Status != 200 {
		t.Fatalf("member GET statuses: %d, want 200", env.Status)
	}
}

// TestWorkflowAdmin_IssueTypes verifies issue type write routes require admin.
func TestWorkflowAdmin_IssueTypes(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)
	srv := setupTestServer(t, db)

	admin := testpg.SeedUser(t, db)
	member := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	seedMember(t, db, wsID, admin, "admin")
	seedMember(t, db, wsID, member, "member")
	projID := testpg.SeedProject(t, db, wsID, "IWFL")
	issueTypeID := seedIssueType(t, db, projID)

	memberToken := loginCookie(t, db, member)
	adminToken := loginCookie(t, db, admin)

	// POST /projects/{id}/issue-types — member 403, admin 201
	envD := doRequestWithBody(t, srv, "POST", "/projects/"+projID+"/issue-types", memberToken, map[string]any{
		"name": "Bug", "icon": "bug", "level": 0,
	})
	if envD.Status != 403 {
		t.Fatalf("member POST issue-type: %d, want 403", envD.Status)
	}
	envD = doRequestWithBody(t, srv, "POST", "/projects/"+projID+"/issue-types", adminToken, map[string]any{
		"name": "Bug", "icon": "bug", "level": 0,
	})
	if envD.Status != 201 {
		t.Fatalf("admin POST issue-type: %d, want 201 (error: %s)", envD.Status, envD.Error)
	}

	// DELETE /projects/{id}/issue-types/{issueTypeID} — member 403, admin 204
	env := doRequest(t, srv, "DELETE", "/projects/"+projID+"/issue-types/"+issueTypeID, memberToken)
	if env.Status != 403 {
		t.Fatalf("member DELETE issue-type: %d, want 403", env.Status)
	}
	env = doRequest(t, srv, "DELETE", "/projects/"+projID+"/issue-types/"+issueTypeID, adminToken)
	if env.Status != 204 {
		t.Fatalf("admin DELETE issue-type: %d, want 204 (error: %s)", env.Status, env.Error)
	}

	// GET /projects/{id}/issue-types — member 200 (no regression)
	env = doRequest(t, srv, "GET", "/projects/"+projID+"/issue-types", memberToken)
	if env.Status != 200 {
		t.Fatalf("member GET issue-types: %d, want 200", env.Status)
	}
}

// Ensure authz import is used (it's needed for the test to compile with the right module).
var _ = authz.ErrForbidden
