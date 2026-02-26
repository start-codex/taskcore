package issuetypes

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/respond"
)

func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB) {
	mux.HandleFunc("POST /projects/{projectID}/issue-types", handleCreate(db))
	mux.HandleFunc("GET /projects/{projectID}/issue-types", handleList(db))
	mux.HandleFunc("DELETE /projects/{projectID}/issue-types/{issueTypeID}", handleArchive(db))
}

func fail(w http.ResponseWriter, err error) {
	switch {
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
		var body struct {
			Name  string `json:"name"`
			Icon  string `json:"icon"`
			Level int    `json:"level"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		p := CreateIssueTypeParams{
			ProjectID: r.PathValue("projectID"),
			Name:      body.Name,
			Icon:      body.Icon,
			Level:     body.Level,
		}
		if err := p.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		it, err := CreateIssueType(r.Context(), db, p)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusCreated, it)
	}
}

func handleList(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := ListIssueTypes(r.Context(), db, r.PathValue("projectID"))
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, list)
	}
}

func handleArchive(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := ArchiveIssueType(r.Context(), db, r.PathValue("projectID"), r.PathValue("issueTypeID")); err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
