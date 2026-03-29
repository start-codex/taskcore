// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package boards

import (
	"errors"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/trazawork/internal/authz"
	"github.com/start-codex/trazawork/internal/respond"
)

func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB) {
	mux.HandleFunc("POST /projects/{projectID}/boards", handleCreate(db))
	mux.HandleFunc("GET /projects/{projectID}/boards", handleList(db))
	mux.HandleFunc("GET /boards/{boardID}", handleGet(db))
	mux.HandleFunc("DELETE /boards/{boardID}", handleArchive(db))
	mux.HandleFunc("POST /boards/{boardID}/columns", handleAddColumn(db))
	mux.HandleFunc("GET /boards/{boardID}/columns", handleListColumns(db))
	mux.HandleFunc("DELETE /columns/{columnID}", handleArchiveColumn(db))
	mux.HandleFunc("POST /columns/{columnID}/statuses", handleAssignStatus(db))
	mux.HandleFunc("DELETE /columns/{columnID}/statuses/{statusID}", handleUnassignStatus(db))
}

func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, authz.ErrUnauthenticated):
		respond.Error(w, http.StatusUnauthorized, "authentication required")
	case errors.Is(err, authz.ErrForbidden):
		respond.Error(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, authz.ErrWorkspaceNotFound),
		errors.Is(err, authz.ErrProjectNotFound),
		errors.Is(err, authz.ErrBoardNotFound),
		errors.Is(err, authz.ErrColumnNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrBoardNotFound), errors.Is(err, ErrColumnNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrDuplicateBoardName), errors.Is(err, ErrDuplicateColumnName):
		respond.Error(w, http.StatusConflict, err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func handleCreate(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projID := r.PathValue("projectID")
		if _, err := authz.RequireProjectMembership(r.Context(), db, projID); err != nil {
			fail(w, err)
			return
		}
		var body struct {
			Name        string `json:"name"`
			Type        string `json:"type"`
			FilterQuery string `json:"filter_query"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		params := CreateBoardParams{
			ProjectID:   r.PathValue("projectID"),
			Name:        body.Name,
			Type:        body.Type,
			FilterQuery: body.FilterQuery,
		}
		if err := params.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		b, err := CreateBoard(r.Context(), db, params)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusCreated, b)
	}
}

func handleList(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projID := r.PathValue("projectID")
		if _, err := authz.RequireProjectMembership(r.Context(), db, projID); err != nil {
			fail(w, err)
			return
		}
		list, err := ListBoards(r.Context(), db, projID)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, list)
	}
}

func handleGet(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		boardID := r.PathValue("boardID")
		if _, _, err := authz.RequireBoardAccess(r.Context(), db, boardID); err != nil {
			fail(w, err)
			return
		}
		b, err := GetBoard(r.Context(), db, boardID)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, b)
	}
}

func handleArchive(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		boardID := r.PathValue("boardID")
		if _, _, err := authz.RequireBoardAccess(r.Context(), db, boardID); err != nil {
			fail(w, err)
			return
		}
		if err := ArchiveBoard(r.Context(), db, boardID); err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleAddColumn(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		boardID := r.PathValue("boardID")
		if _, _, err := authz.RequireBoardAccess(r.Context(), db, boardID); err != nil {
			fail(w, err)
			return
		}
		var body struct {
			Name string `json:"name"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		params := AddColumnParams{BoardID: r.PathValue("boardID"), Name: body.Name}
		if err := params.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		col, err := AddColumn(r.Context(), db, params)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusCreated, col)
	}
}

func handleListColumns(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		boardID := r.PathValue("boardID")
		if _, _, err := authz.RequireBoardAccess(r.Context(), db, boardID); err != nil {
			fail(w, err)
			return
		}
		cols, err := ListColumns(r.Context(), db, boardID)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, cols)
	}
}

func handleArchiveColumn(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		colID := r.PathValue("columnID")
		if _, _, _, err := authz.RequireColumnAccess(r.Context(), db, colID); err != nil {
			fail(w, err)
			return
		}
		if err := ArchiveColumn(r.Context(), db, colID); err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleAssignStatus(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		colID := r.PathValue("columnID")
		if _, _, _, err := authz.RequireColumnAccess(r.Context(), db, colID); err != nil {
			fail(w, err)
			return
		}
		var body struct {
			StatusID string `json:"status_id"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if err := AssignStatus(r.Context(), db, r.PathValue("columnID"), body.StatusID); err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleUnassignStatus(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		colID := r.PathValue("columnID")
		if _, _, _, err := authz.RequireColumnAccess(r.Context(), db, colID); err != nil {
			fail(w, err)
			return
		}
		if err := UnassignStatus(r.Context(), db, colID, r.PathValue("statusID")); err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
