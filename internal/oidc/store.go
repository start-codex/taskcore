// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package oidc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/tookly/internal/pgutil"
)

const providerCols = `id, name, slug, issuer_url, client_id, client_secret, redirect_uri, scopes, auto_register, enabled, created_at, updated_at`

func createProvider(ctx context.Context, db *sqlx.DB, p CreateProviderParams) (Provider, error) {
	scopes := p.Scopes
	if scopes == "" {
		scopes = "openid email profile"
	}
	var prov Provider
	err := db.QueryRowxContext(ctx,
		`INSERT INTO oidc_providers (name, slug, issuer_url, client_id, client_secret, redirect_uri, scopes, auto_register, enabled)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING `+providerCols,
		p.Name, p.Slug, p.IssuerURL, p.ClientID, p.ClientSecret, p.RedirectURI, scopes, p.AutoRegister, p.Enabled,
	).StructScan(&prov)
	if err != nil {
		if pgutil.IsUniqueViolation(err) {
			return Provider{}, ErrDuplicateSlug
		}
		return Provider{}, fmt.Errorf("insert oidc provider: %w", err)
	}
	return prov, nil
}

const maskedSecret = "********"

func updateProvider(ctx context.Context, db *sqlx.DB, id string, p UpdateProviderParams) (Provider, error) {
	secret := p.ClientSecret
	if secret == maskedSecret || secret == "" {
		var existing string
		if err := db.GetContext(ctx, &existing,
			`SELECT client_secret FROM oidc_providers WHERE id = $1`, id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return Provider{}, ErrProviderNotFound
			}
			return Provider{}, fmt.Errorf("get existing secret: %w", err)
		}
		secret = existing
	}
	scopes := p.Scopes
	if scopes == "" {
		scopes = "openid email profile"
	}
	var prov Provider
	err := db.QueryRowxContext(ctx,
		`UPDATE oidc_providers
		 SET name = $2, issuer_url = $3, client_id = $4, client_secret = $5, redirect_uri = $6,
		     scopes = $7, auto_register = $8, enabled = $9, updated_at = NOW()
		 WHERE id = $1
		 RETURNING `+providerCols,
		id, p.Name, p.IssuerURL, p.ClientID, secret, p.RedirectURI, scopes, p.AutoRegister, p.Enabled,
	).StructScan(&prov)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Provider{}, ErrProviderNotFound
		}
		return Provider{}, fmt.Errorf("update oidc provider: %w", err)
	}
	return prov, nil
}

func deleteProvider(ctx context.Context, db *sqlx.DB, id string) error {
	result, err := db.ExecContext(ctx, `DELETE FROM oidc_providers WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete oidc provider: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrProviderNotFound
	}
	return nil
}

func getProvider(ctx context.Context, db *sqlx.DB, id string) (Provider, error) {
	var prov Provider
	err := db.GetContext(ctx, &prov, `SELECT `+providerCols+` FROM oidc_providers WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Provider{}, ErrProviderNotFound
		}
		return Provider{}, fmt.Errorf("get oidc provider: %w", err)
	}
	return prov, nil
}

func getProviderBySlug(ctx context.Context, db *sqlx.DB, slug string) (Provider, error) {
	var prov Provider
	err := db.GetContext(ctx, &prov, `SELECT `+providerCols+` FROM oidc_providers WHERE slug = $1`, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Provider{}, ErrProviderNotFound
		}
		return Provider{}, fmt.Errorf("get oidc provider by slug: %w", err)
	}
	return prov, nil
}

func listProviders(ctx context.Context, db *sqlx.DB) ([]Provider, error) {
	var providers []Provider
	err := db.SelectContext(ctx, &providers, `SELECT `+providerCols+` FROM oidc_providers ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list oidc providers: %w", err)
	}
	if providers == nil {
		providers = []Provider{}
	}
	return providers, nil
}

func listEnabledProviders(ctx context.Context, db *sqlx.DB) ([]PublicProvider, error) {
	var providers []PublicProvider
	err := db.SelectContext(ctx, &providers, `SELECT id, name, slug FROM oidc_providers WHERE enabled = true ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list enabled oidc providers: %w", err)
	}
	if providers == nil {
		providers = []PublicProvider{}
	}
	return providers, nil
}

func getIdentityByProviderSubject(ctx context.Context, db *sqlx.DB, providerID, subject string) (Identity, error) {
	var ident Identity
	err := db.GetContext(ctx, &ident,
		`SELECT id, user_id, provider_id, subject, email, created_at
		 FROM user_identities WHERE provider_id = $1 AND subject = $2`,
		providerID, subject)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Identity{}, ErrIdentityNotFound
		}
		return Identity{}, fmt.Errorf("get identity: %w", err)
	}
	return ident, nil
}

func createIdentity(ctx context.Context, tx *sqlx.Tx, userID, providerID, subject, email string) (Identity, error) {
	var ident Identity
	err := tx.QueryRowxContext(ctx,
		`INSERT INTO user_identities (user_id, provider_id, subject, email)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, provider_id, subject, email, created_at`,
		userID, providerID, subject, email,
	).StructScan(&ident)
	if err != nil {
		return Identity{}, fmt.Errorf("insert identity: %w", err)
	}
	return ident, nil
}

func setEmailVerifiedTx(ctx context.Context, tx *sqlx.Tx, userID string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE app_users SET email_verified_at = NOW() WHERE id = $1 AND email_verified_at IS NULL`,
		userID)
	if err != nil {
		return fmt.Errorf("set email verified: %w", err)
	}
	return nil
}
