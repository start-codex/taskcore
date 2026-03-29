// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package main

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/trazawork/internal/authz"
	"github.com/start-codex/trazawork/internal/respond"
	"github.com/start-codex/trazawork/internal/sessions"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(status int) {
	sw.status = status
	sw.ResponseWriter.WriteHeader(status)
}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		w.Header().Set("X-Request-ID", hex.EncodeToString(b))
		next.ServeHTTP(w, r)
	})
}

func withLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(sw, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", w.Header().Get("X-Request-ID"),
		)
	})
}

var authAllowlist = []struct{ method, path string }{
	{"POST", "/users"},
	{"POST", "/auth/login"},
	{"GET", "/auth/me"},
	{"POST", "/auth/logout"},
}

func withAuth(next http.Handler, db *sqlx.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, route := range authAllowlist {
			if r.Method == route.method && r.URL.Path == route.path {
				next.ServeHTTP(w, r)
				return
			}
		}

		cookie, err := r.Cookie("session_id")
		if err != nil || cookie.Value == "" {
			respond.Error(w, http.StatusUnauthorized, "authentication required")
			return
		}

		session, err := sessions.Validate(r.Context(), db, cookie.Value)
		if err != nil {
			if sessions.IsAuthError(err) {
				respond.Error(w, http.StatusUnauthorized, "authentication required")
				return
			}
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}

		ctx := authz.WithUserID(r.Context(), session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered", "error", rec)
				respond.Error(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
