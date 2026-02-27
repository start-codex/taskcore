package issues

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/respond"
)

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
		var body struct {
			IssueTypeID   string `json:"issue_type_id"`
			StatusID      string `json:"status_id"`
			ParentIssueID string `json:"parent_issue_id"`
			Title         string `json:"title"`
			Description   string `json:"description"`
			Priority      string `json:"priority"`
			AssigneeID    string `json:"assignee_id"`
			ReporterID    string `json:"reporter_id"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
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
			ReporterID:    body.ReporterID,
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
		var body struct {
			Title       string  `json:"title"`
			Description string  `json:"description"`
			Priority    string  `json:"priority"`
			AssigneeID  *string `json:"assignee_id"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		params := UpdateIssueParams{
			IssueID:     r.PathValue("issueID"),
			ProjectID:   r.PathValue("projectID"),
			Title:       body.Title,
			Description: body.Description,
			Priority:    body.Priority,
			AssigneeID:  body.AssigneeID,
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
		if err := ArchiveIssue(r.Context(), db, r.PathValue("projectID"), r.PathValue("issueID")); err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleMove(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
