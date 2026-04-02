// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package emailverification

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/tookly/internal/authz"
	"github.com/start-codex/tookly/internal/instance"
	"github.com/start-codex/tookly/internal/respond"
	"github.com/start-codex/tookly/internal/users"
)

func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB) {
	mux.HandleFunc("POST /auth/verify-email", handleVerifyEmail(db))
	mux.HandleFunc("POST /auth/resend-verification", handleResendVerification(db))
	mux.HandleFunc("GET /instance/verification", handleGetVerificationConfig(db))
	mux.HandleFunc("POST /instance/verification", handleSetVerificationConfig(db))
}

func handleVerifyEmail(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Token string `json:"token"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if body.Token == "" {
			respond.Error(w, http.StatusBadRequest, "token is required")
			return
		}
		if err := VerifyEmail(r.Context(), db, body.Token); err != nil {
			if errors.Is(err, ErrTokenNotFound) || errors.Is(err, ErrTokenExpired) || errors.Is(err, ErrTokenUsed) {
				respond.Error(w, http.StatusBadRequest, "invalid_or_expired_token")
				return
			}
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		respond.JSON(w, http.StatusOK, map[string]string{"status": "verified"})
	}
}

func handleResendVerification(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := authz.UserIDFromContext(r.Context())
		if err != nil {
			respond.Error(w, http.StatusUnauthorized, "authentication required")
			return
		}
		user, err := users.GetUser(r.Context(), db, userID)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		// Idempotent: already verified → success without sending
		if user.EmailVerifiedAt != nil {
			respond.JSON(w, http.StatusOK, map[string]string{"status": "already_verified"})
			return
		}
		// Build base URL
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
		if err := SendVerificationEmail(r.Context(), db, userID, user.Email, baseURL); err != nil {
			respond.Error(w, http.StatusInternalServerError, "failed to send verification email")
			return
		}
		respond.JSON(w, http.StatusOK, map[string]string{"status": "sent"})
	}
}

func handleGetVerificationConfig(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := authz.RequireInstanceAdmin(r.Context(), db); err != nil {
			respond.Error(w, http.StatusForbidden, "forbidden")
			return
		}
		required, err := IsVerificationRequired(r.Context(), db)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		respond.JSON(w, http.StatusOK, map[string]bool{"required": required})
	}
}

func handleSetVerificationConfig(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := authz.RequireInstanceAdmin(r.Context(), db); err != nil {
			respond.Error(w, http.StatusForbidden, "forbidden")
			return
		}
		var body struct {
			Required bool `json:"required"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		val := "false"
		if body.Required {
			val = "true"
		}
		if err := instance.SetConfig(r.Context(), db, "email_verification_required", val); err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		respond.JSON(w, http.StatusOK, map[string]string{"status": "saved"})
	}
}
