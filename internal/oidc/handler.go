// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package oidc

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/tookly/internal/authz"
	"github.com/start-codex/tookly/internal/respond"
	"github.com/start-codex/tookly/internal/sessions"
	"github.com/start-codex/tookly/internal/users"
	"golang.org/x/oauth2"
)

func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB) {
	// Public
	mux.HandleFunc("GET /auth/oidc/providers", handleListEnabled(db))
	mux.HandleFunc("GET /auth/oidc/{slug}", handleStartFlow(db))
	mux.HandleFunc("GET /auth/oidc/{slug}/callback", handleCallback(db))
	// Admin CRUD
	mux.HandleFunc("GET /instance/oidc/providers", handleAdminList(db))
	mux.HandleFunc("POST /instance/oidc/providers", handleAdminCreate(db))
	mux.HandleFunc("PUT /instance/oidc/providers/{id}", handleAdminUpdate(db))
	mux.HandleFunc("DELETE /instance/oidc/providers/{id}", handleAdminDelete(db))
}

func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrProviderNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrDuplicateSlug):
		respond.Error(w, http.StatusConflict, err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

// --- Public endpoints ---

func handleListEnabled(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		providers, err := ListEnabledProviders(r.Context(), db)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		respond.JSON(w, http.StatusOK, providers)
	}
}

func handleStartFlow(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		prov, err := GetProviderBySlug(r.Context(), db, slug)
		if err != nil {
			if errors.Is(err, ErrProviderNotFound) {
				respond.Error(w, http.StatusNotFound, "provider not found")
				return
			}
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		if !prov.Enabled {
			respond.Error(w, http.StatusNotFound, "provider not found")
			return
		}

		state, err := sessions.GenerateToken()
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		nonce, err := sessions.GenerateToken()
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}

		next := sanitizeNext(r.URL.Query().Get("next"))

		secure := os.Getenv("SECURE_COOKIES") == "true"
		cookiePath := fmt.Sprintf("/api/auth/oidc/%s", slug)
		for _, c := range []*http.Cookie{
			{Name: "oidc_state", Value: state, Path: cookiePath, MaxAge: 600, HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: secure},
			{Name: "oidc_nonce", Value: nonce, Path: cookiePath, MaxAge: 600, HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: secure},
			{Name: "oidc_next", Value: next, Path: cookiePath, MaxAge: 600, HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: secure},
		} {
			http.SetCookie(w, c)
		}

		oauth2Cfg, _, err := newOAuth2Config(r.Context(), prov)
		if err != nil {
			respond.Error(w, http.StatusBadGateway, "failed to discover OIDC provider")
			return
		}

		authURL := oauth2Cfg.AuthCodeURL(state, oauth2.SetAuthURLParam("nonce", nonce))
		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

func handleCallback(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		cookiePath := fmt.Sprintf("/api/auth/oidc/%s", slug)
		secure := os.Getenv("SECURE_COOKIES") == "true"

		// Read and clear cookies
		stateCookie, _ := r.Cookie("oidc_state")
		nonceCookie, _ := r.Cookie("oidc_nonce")
		nextCookie, _ := r.Cookie("oidc_next")
		for _, name := range []string{"oidc_state", "oidc_nonce", "oidc_next"} {
			http.SetCookie(w, &http.Cookie{Name: name, Value: "", Path: cookiePath, MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: secure})
		}

		next := "/"
		if nextCookie != nil {
			next = sanitizeNext(nextCookie.Value)
		}

		// Validate state
		if stateCookie == nil || stateCookie.Value == "" {
			redirectLoginError(w, r, "oidc_denied", next)
			return
		}
		queryState := r.URL.Query().Get("state")
		if subtle.ConstantTimeCompare([]byte(stateCookie.Value), []byte(queryState)) != 1 {
			redirectLoginError(w, r, "oidc_denied", next)
			return
		}

		// Check for IdP error
		if r.URL.Query().Get("error") != "" {
			redirectLoginError(w, r, "oidc_denied", next)
			return
		}

		nonce := ""
		if nonceCookie != nil {
			nonce = nonceCookie.Value
		}

		prov, err := GetProviderBySlug(r.Context(), db, slug)
		if err != nil || !prov.Enabled {
			redirectLoginError(w, r, "oidc_denied", next)
			return
		}

		oauth2Cfg, oidcProvider, err := newOAuth2Config(r.Context(), prov)
		if err != nil {
			redirectLoginError(w, r, "oidc_denied", next)
			return
		}

		code := r.URL.Query().Get("code")
		token, err := oauth2Cfg.Exchange(r.Context(), code)
		if err != nil {
			redirectLoginError(w, r, "oidc_denied", next)
			return
		}

		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			redirectLoginError(w, r, "oidc_denied", next)
			return
		}

		claims, err := verifyIDToken(r.Context(), prov, oidcProvider, rawIDToken, nonce)
		if err != nil {
			redirectLoginError(w, r, "oidc_denied", next)
			return
		}
		if claims.Email == "" {
			redirectLoginError(w, r, "oidc_denied", next)
			return
		}

		// Account resolution in transaction
		user, err := resolveAccount(r.Context(), db, prov, claims)
		if err != nil {
			switch {
			case errors.Is(err, errAccountArchived):
				redirectLoginError(w, r, "account_archived", next)
			case errors.Is(err, errNoAccount):
				redirectLoginError(w, r, "oidc_no_account", next)
			default:
				redirectLoginError(w, r, "oidc_denied", next)
			}
			return
		}

		// Create session
		result, err := sessions.Create(r.Context(), db, user.ID)
		if err != nil {
			redirectLoginError(w, r, "oidc_denied", next)
			return
		}

		// Set session cookie with SameSite=Lax (cross-site redirect from IdP)
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    result.RawToken,
			Path:     "/",
			MaxAge:   604800,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   secure,
		})

		http.Redirect(w, r, next, http.StatusFound)
	}
}

var (
	errAccountArchived = errors.New("account archived")
	errNoAccount       = errors.New("no account for this email")
)

func resolveAccount(ctx context.Context, db *sqlx.DB, prov Provider, claims *IDTokenClaims) (users.User, error) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return users.User{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. Check existing identity
	ident, err := getIdentityByProviderSubject(ctx, db, prov.ID, claims.Subject)
	if err == nil {
		// Identity found — load user
		user, err := users.GetUser(ctx, db, ident.UserID)
		if err != nil {
			return users.User{}, fmt.Errorf("get user: %w", err)
		}
		if user.ArchivedAt != nil {
			return users.User{}, errAccountArchived
		}
		_ = tx.Commit()
		return user, nil
	}
	if !errors.Is(err, ErrIdentityNotFound) {
		return users.User{}, fmt.Errorf("lookup identity: %w", err)
	}

	// 2. Identity not found — try to link by email (exact match)
	user, err := users.GetUserByEmail(ctx, db, claims.Email)
	if err == nil {
		// User exists — create identity link
		if user.ArchivedAt != nil {
			return users.User{}, errAccountArchived
		}
		if _, err := createIdentity(ctx, tx, user.ID, prov.ID, claims.Subject, claims.Email); err != nil {
			return users.User{}, fmt.Errorf("create identity link: %w", err)
		}
		// Mark email as verified if not already
		if err := setEmailVerifiedTx(ctx, tx, user.ID); err != nil {
			return users.User{}, fmt.Errorf("set verified: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return users.User{}, fmt.Errorf("commit: %w", err)
		}
		return user, nil
	}
	if !errors.Is(err, users.ErrUserNotFound) {
		return users.User{}, fmt.Errorf("lookup user by email: %w", err)
	}

	// 3. No user found — JIT provisioning if allowed
	if !prov.AutoRegister {
		return users.User{}, errNoAccount
	}

	name := claims.Name
	if name == "" {
		name = claims.Email
	}
	newUser, err := users.CreateOIDCUserTx(ctx, tx, users.CreateOIDCUserParams{
		Email: claims.Email,
		Name:  name,
	})
	if err != nil {
		return users.User{}, fmt.Errorf("create oidc user: %w", err)
	}
	if _, err := createIdentity(ctx, tx, newUser.ID, prov.ID, claims.Subject, claims.Email); err != nil {
		return users.User{}, fmt.Errorf("create identity: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return users.User{}, fmt.Errorf("commit: %w", err)
	}
	return newUser, nil
}

// sanitizeNext validates that next is a safe local relative path.
// Rejects protocol-relative (//evil.example), scheme-based, or empty values.
func sanitizeNext(next string) string {
	if next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") || strings.Contains(next, ":") {
		return "/"
	}
	return next
}

func redirectLoginError(w http.ResponseWriter, r *http.Request, errorCode, next string) {
	u := "/login?error=" + url.QueryEscape(errorCode)
	if next != "" && next != "/" {
		u += "&next=" + url.QueryEscape(next)
	}
	http.Redirect(w, r, u, http.StatusFound)
}

// --- Admin endpoints ---

func handleAdminList(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := authz.RequireInstanceAdmin(r.Context(), db); err != nil {
			respond.Error(w, http.StatusForbidden, "forbidden")
			return
		}
		providers, err := ListProviders(r.Context(), db)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		respond.JSON(w, http.StatusOK, providers)
	}
}

func handleAdminCreate(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := authz.RequireInstanceAdmin(r.Context(), db); err != nil {
			respond.Error(w, http.StatusForbidden, "forbidden")
			return
		}
		var body struct {
			Name         string `json:"name"`
			Slug         string `json:"slug"`
			IssuerURL    string `json:"issuer_url"`
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
			RedirectURI  string `json:"redirect_uri"`
			Scopes       string `json:"scopes"`
			AutoRegister bool   `json:"auto_register"`
			Enabled      bool   `json:"enabled"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		params := CreateProviderParams{
			Name:         body.Name,
			Slug:         body.Slug,
			IssuerURL:    body.IssuerURL,
			ClientID:     body.ClientID,
			ClientSecret: body.ClientSecret,
			RedirectURI:  body.RedirectURI,
			Scopes:       body.Scopes,
			AutoRegister: body.AutoRegister,
			Enabled:      body.Enabled,
		}
		if err := params.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		prov, err := CreateProvider(r.Context(), db, params)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusCreated, prov)
	}
}

func handleAdminUpdate(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := authz.RequireInstanceAdmin(r.Context(), db); err != nil {
			respond.Error(w, http.StatusForbidden, "forbidden")
			return
		}
		id := r.PathValue("id")
		var body struct {
			Name         string `json:"name"`
			IssuerURL    string `json:"issuer_url"`
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
			RedirectURI  string `json:"redirect_uri"`
			Scopes       string `json:"scopes"`
			AutoRegister bool   `json:"auto_register"`
			Enabled      bool   `json:"enabled"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		params := UpdateProviderParams{
			Name:         body.Name,
			IssuerURL:    body.IssuerURL,
			ClientID:     body.ClientID,
			ClientSecret: body.ClientSecret,
			RedirectURI:  body.RedirectURI,
			Scopes:       body.Scopes,
			AutoRegister: body.AutoRegister,
			Enabled:      body.Enabled,
		}
		if err := params.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		prov, err := UpdateProvider(r.Context(), db, id, params)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, prov)
	}
}

func handleAdminDelete(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := authz.RequireInstanceAdmin(r.Context(), db); err != nil {
			respond.Error(w, http.StatusForbidden, "forbidden")
			return
		}
		id := r.PathValue("id")
		if err := DeleteProvider(r.Context(), db, id); err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
