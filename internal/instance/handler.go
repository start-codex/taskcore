// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package instance

import (
	"errors"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/tookly/internal/authz"
	"github.com/start-codex/tookly/internal/email"
	"github.com/start-codex/tookly/internal/respond"
	"github.com/start-codex/tookly/internal/users"
)

func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB) {
	mux.HandleFunc("GET /instance/status", handleStatus(db))
	mux.HandleFunc("POST /instance/bootstrap", handleBootstrap(db))
	mux.HandleFunc("GET /instance/smtp", handleGetSMTP(db))
	mux.HandleFunc("POST /instance/smtp", handleSetSMTP(db))
	mux.HandleFunc("POST /instance/smtp/test", handleTestSMTP(db))
}

func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrAlreadyInitialized):
		respond.Error(w, http.StatusConflict, "instance already initialized")
	case errors.Is(err, users.ErrDuplicateEmail):
		respond.Error(w, http.StatusConflict, "email already exists")
	default:
		respond.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func handleStatus(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		init, err := IsInitialized(r.Context(), db)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		respond.JSON(w, http.StatusOK, map[string]bool{"initialized": init})
	}
}

func handleBootstrap(db *sqlx.DB) http.HandlerFunc {
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

		params := BootstrapParams{
			Email:    body.Email,
			Name:     body.Name,
			Password: body.Password,
		}
		if err := params.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}

		result, err := Bootstrap(r.Context(), db, params)
		if err != nil {
			fail(w, err)
			return
		}

		secure := os.Getenv("SECURE_COOKIES") == "true"
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    result.RawToken,
			Path:     "/",
			MaxAge:   604800,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   secure,
		})

		respond.JSON(w, http.StatusCreated, result.User)
	}
}

func handleGetSMTP(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := authz.RequireInstanceAdmin(r.Context(), db); err != nil {
			respond.Error(w, http.StatusForbidden, "forbidden")
			return
		}
		config, err := LoadSMTPConfig(r.Context(), db)
		if err != nil {
			if errors.Is(err, email.ErrSMTPNotConfigured) {
				respond.JSON(w, http.StatusOK, email.SMTPConfig{})
				return
			}
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		// Mask password
		if config.Password != "" {
			config.Password = "********"
		}
		respond.JSON(w, http.StatusOK, config)
	}
}

const maskedPassword = "********"

func handleSetSMTP(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := authz.RequireInstanceAdmin(r.Context(), db); err != nil {
			respond.Error(w, http.StatusForbidden, "forbidden")
			return
		}
		var config email.SMTPConfig
		if err := respond.Decode(r, &config); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if err := config.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		// If password is the masked sentinel, keep the existing stored password
		if config.Password == maskedPassword {
			existing, err := LoadSMTPConfig(r.Context(), db)
			if err == nil && existing != nil {
				config.Password = existing.Password
			} else {
				config.Password = ""
			}
		}
		if err := SaveSMTPConfig(r.Context(), db, config); err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		respond.JSON(w, http.StatusOK, map[string]string{"status": "saved"})
	}
}

func handleTestSMTP(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := authz.RequireInstanceAdmin(r.Context(), db); err != nil {
			respond.Error(w, http.StatusForbidden, "forbidden")
			return
		}
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
		config, err := LoadSMTPConfig(r.Context(), db)
		if err != nil {
			if errors.Is(err, email.ErrSMTPNotConfigured) {
				respond.Error(w, http.StatusUnprocessableEntity, "SMTP not configured")
				return
			}
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		msg := email.Message{
			To:      user.Email,
			Subject: "Tookly SMTP Test",
			Body:    "<h2>SMTP works!</h2><p>This is a test email from your Tookly instance.</p>",
		}
		if err := email.Send(config, msg); err != nil {
			respond.Error(w, http.StatusBadGateway, "failed to send test email: "+err.Error())
			return
		}
		respond.JSON(w, http.StatusOK, map[string]string{"status": "sent", "to": user.Email})
	}
}
