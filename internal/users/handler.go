package users

import (
	"errors"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/respond"
)

func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB) {
	mux.HandleFunc("POST /users", handleCreate(db))
	mux.HandleFunc("GET /users/{userID}", handleGet(db))
}

func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrUserNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrDuplicateEmail):
		respond.Error(w, http.StatusConflict, err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func handleCreate(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		p := CreateUserParams{Email: body.Email, Name: body.Name}
		if err := p.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		u, err := CreateUser(r.Context(), db, p)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusCreated, u)
	}
}

func handleGet(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := GetUser(r.Context(), db, r.PathValue("userID"))
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, u)
	}
}
