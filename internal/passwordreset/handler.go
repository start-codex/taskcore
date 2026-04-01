// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package passwordreset

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/tookly/internal/email"
	"github.com/start-codex/tookly/internal/instance"
	"github.com/start-codex/tookly/internal/respond"
	"github.com/start-codex/tookly/internal/users"
)

func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB) {
	mux.HandleFunc("POST /auth/forgot-password", handleForgotPassword(db))
	mux.HandleFunc("POST /auth/reset-password", handleResetPassword(db))
}

func handleForgotPassword(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email string `json:"email"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		// Always return 200 — no email enumeration
		user, err := users.GetUserByEmail(r.Context(), db, body.Email)
		if err != nil || user.ArchivedAt != nil {
			respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
			return
		}

		rawToken, err := CreateToken(r.Context(), db, user.ID)
		if err != nil {
			slog.Error("failed to create reset token", "error", err, "email", body.Email)
			respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
			return
		}

		// Build reset URL — prefer configured base_url, fallback to Origin/Host
		baseURL, _ := instance.GetConfig(r.Context(), db, "base_url")
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
		resetURL := fmt.Sprintf("%s/reset-password?token=%s", baseURL, rawToken)

		// Render and send email
		emailBody, err := email.RenderTemplate("password_reset", struct{ ResetURL string }{resetURL})
		if err != nil {
			slog.Error("failed to render reset email template", "error", err)
			respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
			return
		}

		smtpConfig, _ := instance.LoadSMTPConfig(r.Context(), db)
		if err := email.Send(smtpConfig, email.Message{
			To:      user.Email,
			Subject: "Reset your Tookly password",
			Body:    emailBody,
		}); err != nil {
			slog.Error("failed to send reset email", "error", err, "to", user.Email)
		}

		respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func handleResetPassword(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Token       string `json:"token"`
			NewPassword string `json:"new_password"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if body.Token == "" || body.NewPassword == "" {
			respond.Error(w, http.StatusBadRequest, "token and new_password are required")
			return
		}

		if err := ResetPassword(r.Context(), db, body.Token, body.NewPassword); err != nil {
			if errors.Is(err, ErrTokenNotFound) || errors.Is(err, ErrTokenExpired) || errors.Is(err, ErrTokenUsed) {
				respond.Error(w, http.StatusBadRequest, "invalid_or_expired_token")
				return
			}
			if errors.Is(err, users.ErrPasswordTooShort) {
				respond.Error(w, http.StatusUnprocessableEntity, err.Error())
				return
			}
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}

		respond.JSON(w, http.StatusOK, map[string]string{"status": "password_reset"})
	}
}
