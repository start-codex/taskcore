package workspaces

import (
	"errors"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/respond"
	"github.com/start-codex/taskcode/internal/users"
)

func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB) {
	mux.HandleFunc("POST /workspaces", handleCreate(db))
	mux.HandleFunc("GET /workspaces/{workspaceID}", handleGet(db))
	mux.HandleFunc("DELETE /workspaces/{workspaceID}", handleArchive(db))
	mux.HandleFunc("GET /workspaces/{workspaceID}/members", handleListMembers(db))
	mux.HandleFunc("POST /workspaces/{workspaceID}/members", handleAddMember(db))
	mux.HandleFunc("PUT /workspaces/{workspaceID}/members/{userID}", handleUpdateMemberRole(db))
	mux.HandleFunc("DELETE /workspaces/{workspaceID}/members/{userID}", handleRemoveMember(db))
}

func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrWorkspaceNotFound),
		errors.Is(err, users.ErrUserNotFound),
		errors.Is(err, users.ErrMemberNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrDuplicateSlug):
		respond.Error(w, http.StatusConflict, err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func handleCreate(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name string `json:"name"`
			Slug string `json:"slug"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		p := CreateWorkspaceParams{Name: body.Name, Slug: body.Slug}
		if err := p.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		ws, err := CreateWorkspace(r.Context(), db, p)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusCreated, ws)
	}
}

func handleGet(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := GetWorkspace(r.Context(), db, r.PathValue("workspaceID"))
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, ws)
	}
}

func handleArchive(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := ArchiveWorkspace(r.Context(), db, r.PathValue("workspaceID")); err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleListMembers(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		members, err := users.ListWorkspaceMembers(r.Context(), db, r.PathValue("workspaceID"))
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
		p := users.AddWorkspaceMemberParams{
			WorkspaceID: r.PathValue("workspaceID"),
			UserID:      body.UserID,
			Role:        body.Role,
		}
		if err := p.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		m, err := users.AddWorkspaceMember(r.Context(), db, p)
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
		p := users.UpdateWorkspaceMemberRoleParams{
			WorkspaceID: r.PathValue("workspaceID"),
			UserID:      r.PathValue("userID"),
			Role:        body.Role,
		}
		if err := p.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		m, err := users.UpdateWorkspaceMemberRole(r.Context(), db, p)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, m)
	}
}

func handleRemoveMember(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := users.RemoveWorkspaceMember(r.Context(), db, r.PathValue("workspaceID"), r.PathValue("userID"))
		if err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
