// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package users

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/tookly/internal/authz"
	"github.com/start-codex/tookly/internal/email"
	"github.com/start-codex/tookly/internal/respond"
	"github.com/start-codex/tookly/internal/sessions"
)

func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB) {
	mux.HandleFunc("POST /users", handleCreate(db))
	mux.HandleFunc("GET /users/{userID}", handleGet(db))
	mux.HandleFunc("POST /auth/login", handleLogin(db))
	mux.HandleFunc("GET /auth/me", handleMe(db))
	mux.HandleFunc("POST /auth/logout", handleLogout(db))
	mux.HandleFunc("POST /auth/change-password", handleChangePassword(db))
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

		// Send verification email if required (inline to avoid import cycle)
		TrySendVerificationEmail(r, db, user.ID, user.Email)

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
		// Check if email verification is required (inline to avoid import cycle with instance)
		verificationRequired := false
		var reqVal string
		if err := db.GetContext(r.Context(), &reqVal,
			`SELECT value FROM instance_config WHERE key = 'email_verification_required'`); err == nil {
			verificationRequired = reqVal == "true"
		}
		respond.JSON(w, http.StatusOK, map[string]any{
			"authenticated":               true,
			"user":                         user,
			"email_verification_required":  verificationRequired,
		})
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

func handleChangePassword(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := authz.UserIDFromContext(r.Context())
		if err != nil {
			respond.Error(w, http.StatusUnauthorized, "authentication required")
			return
		}
		var body struct {
			CurrentPassword string `json:"current_password"`
			NewPassword     string `json:"new_password"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if err := ChangePassword(r.Context(), db, userID, body.CurrentPassword, body.NewPassword); err != nil {
			if errors.Is(err, ErrInvalidCredentials) {
				respond.Error(w, http.StatusUnauthorized, "current password is incorrect")
				return
			}
			if errors.Is(err, ErrPasswordTooShort) {
				respond.Error(w, http.StatusUnprocessableEntity, err.Error())
				return
			}
			fail(w, err)
			return
		}
		// Invalidate all other sessions, preserving the current one
		cookie, _ := r.Cookie("session_id")
		if cookie != nil && cookie.Value != "" {
			_ = sessions.DeleteByUserID(r.Context(), db, userID, cookie.Value)
		}
		respond.JSON(w, http.StatusOK, map[string]string{"status": "password_changed"})
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

// TrySendVerificationEmail checks if email verification is required and sends
// the verification email. Errors are logged but never fail the caller.
func TrySendVerificationEmail(r *http.Request, db *sqlx.DB, userID, userEmail string) {
	ctx := r.Context()
	var reqVal string
	if err := db.GetContext(ctx, &reqVal,
		`SELECT value FROM instance_config WHERE key = 'email_verification_required'`); err != nil {
		return
	}
	if reqVal != "true" {
		return
	}

	rawToken, err := sessions.GenerateToken()
	if err != nil {
		slog.Error("verification token generation failed", "error", err)
		return
	}
	tokenHash := sessions.HashToken(rawToken)
	if _, err = db.ExecContext(ctx,
		`INSERT INTO email_verification_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, NOW() + INTERVAL '24 hours')`,
		userID, tokenHash); err != nil {
		slog.Error("verification token insert failed", "error", err)
		return
	}

	baseURL := ""
	_ = db.GetContext(ctx, &baseURL, `SELECT value FROM instance_config WHERE key = 'base_url'`)
	if baseURL == "" {
		baseURL = r.Header.Get("Origin")
	}
	if baseURL == "" {
		proto := r.Header.Get("X-Forwarded-Proto")
		if proto == "" {
			proto = "http"
		}
		baseURL = fmt.Sprintf("%s://%s", proto, r.Host)
	}

	body, err := email.RenderTemplate("email_verification", struct{ VerifyURL string }{
		fmt.Sprintf("%s/verify-email?token=%s", baseURL, rawToken),
	})
	if err != nil {
		slog.Error("verification email render failed", "error", err)
		return
	}

	var smtpHost, smtpPortStr, smtpFrom, smtpUser, smtpPass string
	_ = db.GetContext(ctx, &smtpHost, `SELECT value FROM instance_config WHERE key = 'smtp_host'`)
	_ = db.GetContext(ctx, &smtpPortStr, `SELECT value FROM instance_config WHERE key = 'smtp_port'`)
	_ = db.GetContext(ctx, &smtpFrom, `SELECT value FROM instance_config WHERE key = 'smtp_from'`)
	_ = db.GetContext(ctx, &smtpUser, `SELECT value FROM instance_config WHERE key = 'smtp_username'`)
	_ = db.GetContext(ctx, &smtpPass, `SELECT value FROM instance_config WHERE key = 'smtp_password'`)
	if smtpHost == "" {
		return
	}
	port := 587
	fmt.Sscanf(smtpPortStr, "%d", &port)
	_ = email.Send(&email.SMTPConfig{Host: smtpHost, Port: port, From: smtpFrom, Username: smtpUser, Password: smtpPass},
		email.Message{To: userEmail, Subject: "Verify your Tookly email", Body: body})
}
