package issuetypes

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/authz"
	"github.com/start-codex/taskcode/internal/respond"
)

func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB) {
	mux.HandleFunc("POST /projects/{projectID}/issue-types", handleCreate(db))
	mux.HandleFunc("GET /projects/{projectID}/issue-types", handleList(db))
	mux.HandleFunc("DELETE /projects/{projectID}/issue-types/{issueTypeID}", handleArchive(db))
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
	case errors.Is(err, ErrIssueTypeNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrDuplicateIssueType):
		respond.Error(w, http.StatusConflict, err.Error())
	default:
		slog.Error("issuetypes handler error", "error", err)
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
			Name  string `json:"name"`
			Icon  string `json:"icon"`
			Level int    `json:"level"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		params := CreateIssueTypeParams{
			ProjectID: r.PathValue("projectID"),
			Name:      body.Name,
			Icon:      body.Icon,
			Level:     body.Level,
		}
		if err := params.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		it, err := CreateIssueType(r.Context(), db, params)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusCreated, it)
	}
}

func handleList(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projID := r.PathValue("projectID")
		if _, err := authz.RequireProjectMembership(r.Context(), db, projID); err != nil {
			fail(w, err)
			return
		}
		list, err := ListIssueTypes(r.Context(), db, projID)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, list)
	}
}

func handleArchive(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projID := r.PathValue("projectID")
		if _, err := authz.RequireProjectMembership(r.Context(), db, projID); err != nil {
			fail(w, err)
			return
		}
		if err := ArchiveIssueType(r.Context(), db, projID, r.PathValue("issueTypeID")); err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
