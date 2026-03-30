// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package issues

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/tookly/internal/authz"
	"github.com/start-codex/tookly/internal/respond"
)

func parseDueDate(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil, errors.New("due_date must be YYYY-MM-DD format")
	}
	return &t, nil
}

func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB) {
	mux.HandleFunc("POST /projects/{projectID}/issues", handleCreate(db))
	mux.HandleFunc("GET /projects/{projectID}/issues", handleList(db))
	mux.HandleFunc("GET /projects/{projectID}/issues/{issueID}", handleGet(db))
	mux.HandleFunc("PUT /projects/{projectID}/issues/{issueID}", handleUpdate(db))
	mux.HandleFunc("DELETE /projects/{projectID}/issues/{issueID}", handleArchive(db))
	mux.HandleFunc("POST /projects/{projectID}/issues/{issueID}/move", handleMove(db))
}

func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, authz.ErrUnauthenticated):
		respond.Error(w, http.StatusUnauthorized, "authentication required")
	case errors.Is(err, authz.ErrForbidden):
		respond.Error(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, authz.ErrWorkspaceNotFound),
		errors.Is(err, authz.ErrProjectNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrIssueNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalidPriority):
		respond.Error(w, http.StatusUnprocessableEntity, err.Error())
	default:
		slog.Error("issues handler error", "error", err)
		respond.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func handleCreate(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := authz.RequireProjectMembership(r.Context(), db, r.PathValue("projectID")); err != nil {
			fail(w, err)
			return
		}
		authedUserID, err := authz.UserIDFromContext(r.Context())
		if err != nil {
			fail(w, err)
			return
		}
		var body struct {
			IssueTypeID   string  `json:"issue_type_id"`
			StatusID      string  `json:"status_id"`
			ParentIssueID string  `json:"parent_issue_id"`
			Title         string  `json:"title"`
			Description   string  `json:"description"`
			Priority      string  `json:"priority"`
			AssigneeID    string  `json:"assignee_id"`
			DueDate       *string `json:"due_date"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		dueDate, err := parseDueDate(body.DueDate)
		if err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		params := CreateIssueParams{
			ProjectID:     r.PathValue("projectID"),
			IssueTypeID:   body.IssueTypeID,
			StatusID:      body.StatusID,
			ParentIssueID: body.ParentIssueID,
			Title:         body.Title,
			Description:   body.Description,
			Priority:      body.Priority,
			AssigneeID:    body.AssigneeID,
			ReporterID:    authedUserID,
			DueDate:       dueDate,
		}
		if err := params.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		issue, err := CreateIssue(r.Context(), db, params)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusCreated, issue)
	}
}

func handleList(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := authz.RequireProjectMembership(r.Context(), db, r.PathValue("projectID")); err != nil {
			fail(w, err)
			return
		}
		q := r.URL.Query()
		list, err := ListIssues(r.Context(), db, ListIssuesParams{
			ProjectID:  r.PathValue("projectID"),
			StatusID:   q.Get("status_id"),
			AssigneeID: q.Get("assignee_id"),
		})
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, list)
	}
}

func handleGet(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := authz.RequireProjectMembership(r.Context(), db, r.PathValue("projectID")); err != nil {
			fail(w, err)
			return
		}
		issue, err := GetIssue(r.Context(), db, r.PathValue("projectID"), r.PathValue("issueID"))
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, issue)
	}
}

func handleUpdate(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := authz.RequireProjectMembership(r.Context(), db, r.PathValue("projectID")); err != nil {
			fail(w, err)
			return
		}
		var body struct {
			Title       string  `json:"title"`
			Description string  `json:"description"`
			Priority    string  `json:"priority"`
			AssigneeID  *string `json:"assignee_id"`
			DueDate     *string `json:"due_date"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		dueDate, err := parseDueDate(body.DueDate)
		if err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		params := UpdateIssueParams{
			IssueID:     r.PathValue("issueID"),
			ProjectID:   r.PathValue("projectID"),
			Title:       body.Title,
			Description: body.Description,
			Priority:    body.Priority,
			AssigneeID:  body.AssigneeID,
			DueDate:     dueDate,
		}
		if err := params.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		issue, err := UpdateIssue(r.Context(), db, params)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, issue)
	}
}

func handleArchive(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := authz.RequireProjectMembership(r.Context(), db, r.PathValue("projectID")); err != nil {
			fail(w, err)
			return
		}
		if err := ArchiveIssue(r.Context(), db, r.PathValue("projectID"), r.PathValue("issueID")); err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleMove(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := authz.RequireProjectMembership(r.Context(), db, r.PathValue("projectID")); err != nil {
			fail(w, err)
			return
		}
		var body struct {
			TargetStatusID string `json:"target_status_id"`
			TargetPosition int    `json:"target_position"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		params := MoveIssueParams{
			ProjectID:      r.PathValue("projectID"),
			IssueID:        r.PathValue("issueID"),
			TargetStatusID: body.TargetStatusID,
			TargetPosition: body.TargetPosition,
		}
		if err := params.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		if err := MoveIssue(r.Context(), db, params); err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
