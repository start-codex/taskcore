package statuses

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/respond"
)

func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB) {
	mux.HandleFunc("POST /projects/{projectID}/statuses", handleCreate(db))
	mux.HandleFunc("GET /projects/{projectID}/statuses", handleList(db))
	mux.HandleFunc("PUT /projects/{projectID}/statuses/{statusID}", handleUpdate(db))
	mux.HandleFunc("DELETE /projects/{projectID}/statuses/{statusID}", handleArchive(db))
}

func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrStatusNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrDuplicateStatus):
		respond.Error(w, http.StatusConflict, err.Error())
	default:
		slog.Error("statuses handler error", "error", err)
		respond.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func handleCreate(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name     string `json:"name"`
			Category string `json:"category"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		p := CreateStatusParams{
			ProjectID: r.PathValue("projectID"),
			Name:      body.Name,
			Category:  body.Category,
		}
		if err := p.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		s, err := CreateStatus(r.Context(), db, p)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusCreated, s)
	}
}

func handleList(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := ListStatuses(r.Context(), db, r.PathValue("projectID"))
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, list)
	}
}

func handleUpdate(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name     string `json:"name"`
			Category string `json:"category"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		p := UpdateStatusParams{
			StatusID:  r.PathValue("statusID"),
			ProjectID: r.PathValue("projectID"),
			Name:      body.Name,
			Category:  body.Category,
		}
		if err := p.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		s, err := UpdateStatus(r.Context(), db, p)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, s)
	}
}

func handleArchive(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := ArchiveStatus(r.Context(), db, r.PathValue("projectID"), r.PathValue("statusID")); err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
