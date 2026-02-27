package projects

import (
	"errors"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/respond"
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
	case errors.Is(err, ErrProjectNotFound), errors.Is(err, ErrMemberNotFound):
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
		params := CreateProjectParams{
			WorkspaceID: r.PathValue("workspaceID"),
			Name:        body.Name,
			Key:         body.Key,
			Description: body.Description,
		}
		if err := params.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		project, err := CreateProject(r.Context(), db, params)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusCreated, project)
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
		project, err := GetProject(r.Context(), db, r.PathValue("projectID"))
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, project)
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
		members, err := ListMembers(r.Context(), db, r.PathValue("projectID"))
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
		params := AddMemberParams{
			ProjectID: r.PathValue("projectID"),
			UserID:    body.UserID,
			Role:      body.Role,
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
		var body struct {
			Role string `json:"role"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		params := UpdateMemberRoleParams{
			ProjectID: r.PathValue("projectID"),
			UserID:    r.PathValue("userID"),
			Role:      body.Role,
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
		err := RemoveMember(r.Context(), db, r.PathValue("projectID"), r.PathValue("userID"))
		if err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
