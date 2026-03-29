package users

import (
	"errors"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/authz"
	"github.com/start-codex/taskcode/internal/respond"
	"github.com/start-codex/taskcode/internal/sessions"
)

func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB) {
	mux.HandleFunc("POST /users", handleCreate(db))
	mux.HandleFunc("GET /users/{userID}", handleGet(db))
	mux.HandleFunc("POST /auth/login", handleLogin(db))
	mux.HandleFunc("GET /auth/me", handleMe(db))
	mux.HandleFunc("POST /auth/logout", handleLogout(db))
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

func setSessionCookie(w http.ResponseWriter, rawToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    rawToken,
		Path:     "/",
		MaxAge:   604800,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   os.Getenv("SECURE_COOKIES") == "true",
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
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
		if user.ArchivedAt != nil {
			respond.Error(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		result, err := sessions.Create(r.Context(), db, user.ID)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		setSessionCookie(w, result.RawToken)
		respond.JSON(w, http.StatusOK, user)
	}
}

func handleMe(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil || cookie.Value == "" {
			respond.JSON(w, http.StatusOK, map[string]any{"authenticated": false})
			return
		}
		session, err := sessions.Validate(r.Context(), db, cookie.Value)
		if err != nil {
			if sessions.IsAuthError(err) {
				clearSessionCookie(w)
				respond.JSON(w, http.StatusOK, map[string]any{"authenticated": false})
				return
			}
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		user, err := GetUser(r.Context(), db, session.UserID)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				clearSessionCookie(w)
				respond.JSON(w, http.StatusOK, map[string]any{"authenticated": false})
				return
			}
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		respond.JSON(w, http.StatusOK, map[string]any{"authenticated": true, "user": user})
	}
}

func handleLogout(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err == nil && cookie.Value != "" {
			_ = sessions.Delete(r.Context(), db, cookie.Value)
		}
		clearSessionCookie(w)
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleGet(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authedUserID, err := authz.UserIDFromContext(r.Context())
		if err != nil {
			respond.Error(w, http.StatusUnauthorized, "authentication required")
			return
		}
		if authedUserID != r.PathValue("userID") {
			respond.Error(w, http.StatusForbidden, "access denied")
			return
		}
		user, err := GetUser(r.Context(), db, r.PathValue("userID"))
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, user)
	}
}
