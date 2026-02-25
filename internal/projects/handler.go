package projects

import (
	"errors"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/respond"
	"github.com/start-codex/taskcode/internal/users"
)

func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB) {
	mux.HandleFunc("POST /workspaces/{workspaceID}/projects", handleCreate(db))
	mux.HandleFunc("GET /workspaces/{workspaceID}/projects", handleList(db))
	mux.HandleFunc("GET /projects/{projectID}", handleGet(db))
	mux.HandleFunc("DELETE /projects/{projectID}", handleArchive(db))
	mux.HandleFunc("GET /projects/{projectID}/members", handleListMembers(db))
	mux.HandleFunc("POST /projects/{projectID}/members", handleAddMember(db))
	mux.HandleFunc("PUT /projects/{projectID}/members/{userID}", handleUpdateMemberRole(db))
	mux.HandleFunc("DELETE /projects/{projectID}/members/{userID}", handleRemoveMember(db))
}

func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrProjectNotFound),
		errors.Is(err, users.ErrUserNotFound),
		errors.Is(err, users.ErrMemberNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrDuplicateProjectKey):
		respond.Error(w, http.StatusConflict, err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func handleCreate(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name        string `json:"name"`
			Key         string `json:"key"`
			Description string `json:"description"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		p := CreateProjectParams{
			WorkspaceID: r.PathValue("workspaceID"),
			Name:        body.Name,
			Key:         body.Key,
			Description: body.Description,
		}
		if err := p.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		proj, err := CreateProject(r.Context(), db, p)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusCreated, proj)
	}
}

func handleList(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := ListProjects(r.Context(), db, r.PathValue("workspaceID"))
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, list)
	}
}

func handleGet(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proj, err := GetProject(r.Context(), db, r.PathValue("projectID"))
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, proj)
	}
}

func handleArchive(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := ArchiveProject(r.Context(), db, r.PathValue("projectID")); err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleListMembers(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		members, err := users.ListProjectMembers(r.Context(), db, r.PathValue("projectID"))
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, members)
	}
}

func handleAddMember(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			UserID string `json:"user_id"`
			Role   string `json:"role"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		p := users.AddProjectMemberParams{
			ProjectID: r.PathValue("projectID"),
			UserID:    body.UserID,
			Role:      body.Role,
		}
		if err := p.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		m, err := users.AddProjectMember(r.Context(), db, p)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusCreated, m)
	}
}

func handleUpdateMemberRole(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Role string `json:"role"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		p := users.UpdateProjectMemberRoleParams{
			ProjectID: r.PathValue("projectID"),
			UserID:    r.PathValue("userID"),
			Role:      body.Role,
		}
		if err := p.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		m, err := users.UpdateProjectMemberRole(r.Context(), db, p)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, m)
	}
}

func handleRemoveMember(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := users.RemoveProjectMember(r.Context(), db, r.PathValue("projectID"), r.PathValue("userID"))
		if err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
