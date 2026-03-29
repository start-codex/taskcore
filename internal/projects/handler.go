package projects

import (
	"errors"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/authz"
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
	case errors.Is(err, authz.ErrUnauthenticated):
		respond.Error(w, http.StatusUnauthorized, "authentication required")
	case errors.Is(err, authz.ErrForbidden):
		respond.Error(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, authz.ErrWorkspaceNotFound),
		errors.Is(err, authz.ErrProjectNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
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
		wsID := r.PathValue("workspaceID")
		if err := authz.RequireWorkspaceMembership(r.Context(), db, wsID); err != nil {
			fail(w, err)
			return
		}
		var body struct {
			Name        string `json:"name"`
			Key         string `json:"key"`
			Description string `json:"description"`
			Template    string `json:"template"`
			Locale      string `json:"locale"`
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
			Template:    body.Template,
			Locale:      body.Locale,
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
		wsID := r.PathValue("workspaceID")
		if err := authz.RequireWorkspaceMembership(r.Context(), db, wsID); err != nil {
			fail(w, err)
			return
		}
		list, err := ListProjects(r.Context(), db, wsID)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, list)
	}
}

func handleGet(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projID := r.PathValue("projectID")
		if _, err := authz.RequireProjectMembership(r.Context(), db, projID); err != nil {
			fail(w, err)
			return
		}
		project, err := GetProject(r.Context(), db, projID)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, project)
	}
}

func handleArchive(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projID := r.PathValue("projectID")
		if _, err := authz.RequireProjectMembership(r.Context(), db, projID); err != nil {
			fail(w, err)
			return
		}
		if err := ArchiveProject(r.Context(), db, projID); err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleListMembers(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projID := r.PathValue("projectID")
		if _, err := authz.RequireProjectMembership(r.Context(), db, projID); err != nil {
			fail(w, err)
			return
		}
		members, err := ListMembers(r.Context(), db, projID)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, members)
	}
}

func handleAddMember(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projID := r.PathValue("projectID")
		wsID, err := authz.RequireProjectMembership(r.Context(), db, projID)
		if err != nil {
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
		// The target user must be a workspace member before being added to a project.
		targetCtx := authz.WithUserID(r.Context(), body.UserID)
		if err := authz.RequireWorkspaceMembership(targetCtx, db, wsID); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, "user is not a workspace member")
			return
		}
		params := AddMemberParams{
			ProjectID: projID,
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
		projID := r.PathValue("projectID")
		if _, err := authz.RequireProjectMembership(r.Context(), db, projID); err != nil {
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
		projID := r.PathValue("projectID")
		if _, err := authz.RequireProjectMembership(r.Context(), db, projID); err != nil {
			fail(w, err)
			return
		}
		err := RemoveMember(r.Context(), db, projID, r.PathValue("userID"))
		if err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
