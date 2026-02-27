package workspaces

import (
	"errors"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/respond"
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
		var body struct {
			Name string `json:"name"`
			Slug string `json:"slug"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		params := CreateWorkspaceParams{Name: body.Name, Slug: body.Slug}
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
		members, err := ListMembers(r.Context(), db, r.PathValue("workspaceID"))
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
		err := RemoveMember(r.Context(), db, r.PathValue("workspaceID"), r.PathValue("userID"))
		if err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
