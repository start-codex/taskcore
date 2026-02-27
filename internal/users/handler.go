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
	mux.HandleFunc("POST /auth/login", handleLogin(db))
}

func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrUserNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrDuplicateEmail):
		respond.Error(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrInvalidCredentials):
		respond.Error(w, http.StatusUnauthorized, err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func handleCreate(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Password string `json:"password"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		params := CreateUserParams{Email: body.Email, Name: body.Name, Password: body.Password}
		if err := params.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		user, err := CreateUser(r.Context(), db, params)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusCreated, user)
	}
}

func handleLogin(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		user, err := AuthenticateUser(r.Context(), db, body.Email, body.Password)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, user)
	}
}

func handleGet(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := GetUser(r.Context(), db, r.PathValue("userID"))
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, user)
	}
}
