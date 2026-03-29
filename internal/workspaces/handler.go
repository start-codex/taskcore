// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package workspaces

import (
	"errors"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/trazawork/internal/authz"
	"github.com/start-codex/trazawork/internal/respond"
)

func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB) {
	mux.HandleFunc("POST /workspaces", handleCreate(db))
	mux.HandleFunc("GET /workspaces/{workspaceID}", handleGet(db))
	mux.HandleFunc("DELETE /workspaces/{workspaceID}", handleArchive(db))
	mux.HandleFunc("GET /workspaces/{workspaceID}/members", handleListMembers(db))
	mux.HandleFunc("POST /workspaces/{workspaceID}/members", handleAddMember(db))
	mux.HandleFunc("PUT /workspaces/{workspaceID}/members/{userID}", handleUpdateMemberRole(db))
	mux.HandleFunc("DELETE /workspaces/{workspaceID}/members/{userID}", handleRemoveMember(db))
	mux.HandleFunc("GET /workspaces", handleListByUser(db))
}

func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, authz.ErrUnauthenticated):
		respond.Error(w, http.StatusUnauthorized, "authentication required")
	case errors.Is(err, authz.ErrForbidden):
		respond.Error(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, authz.ErrWorkspaceNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrWorkspaceNotFound), errors.Is(err, ErrMemberNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrDuplicateSlug):
		respond.Error(w, http.StatusConflict, err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func handleCreate(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authedUserID, err := authz.UserIDFromContext(r.Context())
		if err != nil {
			fail(w, err)
			return
		}
		var body struct {
			Name string `json:"name"`
			Slug string `json:"slug"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		params := CreateWorkspaceParams{Name: body.Name, Slug: body.Slug, OwnerID: authedUserID}
		if err := params.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		ws, err := CreateWorkspace(r.Context(), db, params)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusCreated, ws)
	}
}

func handleGet(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsID := r.PathValue("workspaceID")
		if err := authz.RequireWorkspaceMembership(r.Context(), db, wsID); err != nil {
			fail(w, err)
			return
		}
		ws, err := GetWorkspace(r.Context(), db, wsID)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, ws)
	}
}

func handleArchive(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsID := r.PathValue("workspaceID")
		if err := authz.RequireWorkspaceAdmin(r.Context(), db, wsID); err != nil {
			fail(w, err)
			return
		}
		if err := ArchiveWorkspace(r.Context(), db, wsID); err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleListMembers(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsID := r.PathValue("workspaceID")
		if err := authz.RequireWorkspaceAdmin(r.Context(), db, wsID); err != nil {
			fail(w, err)
			return
		}
		members, err := ListMembers(r.Context(), db, wsID)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, members)
	}
}

func handleAddMember(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsID := r.PathValue("workspaceID")
		if err := authz.RequireWorkspaceAdmin(r.Context(), db, wsID); err != nil {
			fail(w, err)
			return
		}
		var body struct {
			UserID string `json:"user_id"`
			Role   string `json:"role"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		params := AddMemberParams{
			WorkspaceID: r.PathValue("workspaceID"),
			UserID:      body.UserID,
			Role:        body.Role,
		}
		if err := params.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		member, err := AddMember(r.Context(), db, params)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusCreated, member)
	}
}

func handleUpdateMemberRole(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsID := r.PathValue("workspaceID")
		if err := authz.RequireWorkspaceAdmin(r.Context(), db, wsID); err != nil {
			fail(w, err)
			return
		}
		var body struct {
			Role string `json:"role"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		params := UpdateMemberRoleParams{
			WorkspaceID: r.PathValue("workspaceID"),
			UserID:      r.PathValue("userID"),
			Role:        body.Role,
		}
		if err := params.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		member, err := UpdateMemberRole(r.Context(), db, params)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, member)
	}
}

func handleRemoveMember(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsID := r.PathValue("workspaceID")
		if err := authz.RequireWorkspaceAdmin(r.Context(), db, wsID); err != nil {
			fail(w, err)
			return
		}
		err := RemoveMember(r.Context(), db, wsID, r.PathValue("userID"))
		if err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleListByUser(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := authz.UserIDFromContext(r.Context())
		if err != nil {
			fail(w, err)
			return
		}
		workspaceList, err := ListByUser(r.Context(), db, userID)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, workspaceList)
	}
}
