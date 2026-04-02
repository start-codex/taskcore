// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/tookly/internal/authz"
	"github.com/start-codex/tookly/internal/instance"
	"github.com/start-codex/tookly/internal/respond"
	"github.com/start-codex/tookly/internal/sessions"
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
		if strings.HasPrefix(r.URL.Path, "/_app/") {
			return
		}
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", w.Header().Get("X-Request-ID"),
		)
	})
}

var staticPublicRoutes = []struct{ method, path string }{
	{"POST", "/auth/login"},
	{"GET", "/auth/me"},
	{"POST", "/auth/logout"},
	{"GET", "/instance/status"},
	{"POST", "/instance/bootstrap"},
	{"POST", "/auth/forgot-password"},
	{"POST", "/auth/reset-password"},
	{"GET", "/invitations/accept"},
	{"POST", "/invitations/accept"},
	{"POST", "/auth/verify-email"},
	{"GET", "/auth/oidc/providers"},
}

// isPublicRoute returns:
//   - public=true: route is allowed without auth
//   - block=true: route is blocked with 409 (pre-bootstrap)
//   - err!=nil: internal error (corrupted state)
func isPublicRoute(ctx context.Context, method, path string, db *sqlx.DB) (public bool, block bool, err error) {
	for _, route := range staticPublicRoutes {
		if method == route.method && path == route.path {
			return true, false, nil
		}
	}
	// OIDC flow routes are public (start + callback)
	if method == "GET" && strings.HasPrefix(path, "/auth/oidc/") && path != "/auth/oidc/providers" {
		return true, false, nil
	}
	// POST /users is public only after the instance is initialized.
	// Before bootstrap, user creation is blocked (409).
	if method == "POST" && path == "/users" {
		init, initErr := instance.IsInitialized(ctx, db)
		if initErr != nil {
			return false, false, initErr
		}
		if !init {
			return false, true, nil // block with 409
		}
		return true, false, nil // public after bootstrap
	}
	return false, false, nil
}

func withAuth(next http.Handler, db *sqlx.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		public, block, routeErr := isPublicRoute(r.Context(), r.Method, r.URL.Path, db)
		if routeErr != nil {
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		if block {
			respond.Error(w, http.StatusConflict, "instance not initialized — use POST /instance/bootstrap")
			return
		}
		if public {
			next.ServeHTTP(w, r)
			return
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
