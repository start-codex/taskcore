// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package oidc

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrProviderNotFound = errors.New("OIDC provider not found")
	ErrProviderDisabled = errors.New("OIDC provider is disabled")
	ErrDuplicateSlug    = errors.New("OIDC provider slug already exists")
	ErrIdentityNotFound = errors.New("OIDC identity not found")
)

var slugRegexp = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}[a-z0-9]?$`)

type Provider struct {
	ID           string    `db:"id"            json:"id"`
	Name         string    `db:"name"          json:"name"`
	Slug         string    `db:"slug"          json:"slug"`
	IssuerURL    string    `db:"issuer_url"    json:"issuer_url"`
	ClientID     string    `db:"client_id"     json:"client_id"`
	ClientSecret string    `db:"client_secret" json:"-"`
	RedirectURI  string    `db:"redirect_uri"  json:"redirect_uri"`
	Scopes       string    `db:"scopes"        json:"scopes"`
	AutoRegister bool      `db:"auto_register" json:"auto_register"`
	Enabled      bool      `db:"enabled"       json:"enabled"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"    json:"updated_at"`
}

type PublicProvider struct {
	ID   string `db:"id"   json:"id"`
	Name string `db:"name" json:"name"`
	Slug string `db:"slug" json:"slug"`
}

type Identity struct {
	ID         string    `db:"id"          json:"id"`
	UserID     string    `db:"user_id"     json:"user_id"`
	ProviderID string    `db:"provider_id" json:"provider_id"`
	Subject    string    `db:"subject"     json:"subject"`
	Email      string    `db:"email"       json:"email"`
	CreatedAt  time.Time `db:"created_at"  json:"created_at"`
}

type CreateProviderParams struct {
	Name         string
	Slug         string
	IssuerURL    string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       string
	AutoRegister bool
	Enabled      bool
}

func (p CreateProviderParams) Validate() error {
	if p.Name == "" {
		return errors.New("name is required")
	}
	if !slugRegexp.MatchString(p.Slug) {
		return errors.New("slug must be lowercase alphanumeric with hyphens, 2-64 characters")
	}
	if p.IssuerURL == "" {
		return errors.New("issuer_url is required")
	}
	if !strings.HasPrefix(p.IssuerURL, "https://") && !strings.HasPrefix(p.IssuerURL, "http://") {
		return errors.New("issuer_url must be a valid URL")
	}
	if p.ClientID == "" {
		return errors.New("client_id is required")
	}
	if p.ClientSecret == "" {
		return errors.New("client_secret is required")
	}
	if p.RedirectURI == "" {
		return errors.New("redirect_uri is required")
	}
	return nil
}

type UpdateProviderParams struct {
	Name         string
	IssuerURL    string
	ClientID     string
	ClientSecret string // "********" means keep existing
	RedirectURI  string
	Scopes       string
	AutoRegister bool
	Enabled      bool
}

func (p UpdateProviderParams) Validate() error {
	if p.Name == "" {
		return errors.New("name is required")
	}
	if p.IssuerURL == "" {
		return errors.New("issuer_url is required")
	}
	if !strings.HasPrefix(p.IssuerURL, "https://") && !strings.HasPrefix(p.IssuerURL, "http://") {
		return errors.New("issuer_url must be a valid URL")
	}
	if p.ClientID == "" {
		return errors.New("client_id is required")
	}
	if p.RedirectURI == "" {
		return errors.New("redirect_uri is required")
	}
	return nil
}

func CreateProvider(ctx context.Context, db *sqlx.DB, params CreateProviderParams) (Provider, error) {
	if db == nil {
		return Provider{}, errors.New("db is required")
	}
	if err := params.Validate(); err != nil {
		return Provider{}, err
	}
	return createProvider(ctx, db, params)
}

func UpdateProvider(ctx context.Context, db *sqlx.DB, id string, params UpdateProviderParams) (Provider, error) {
	if db == nil {
		return Provider{}, errors.New("db is required")
	}
	if id == "" {
		return Provider{}, errors.New("id is required")
	}
	if err := params.Validate(); err != nil {
		return Provider{}, err
	}
	return updateProvider(ctx, db, id, params)
}

func DeleteProvider(ctx context.Context, db *sqlx.DB, id string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if id == "" {
		return errors.New("id is required")
	}
	return deleteProvider(ctx, db, id)
}

func GetProvider(ctx context.Context, db *sqlx.DB, id string) (Provider, error) {
	if db == nil {
		return Provider{}, errors.New("db is required")
	}
	if id == "" {
		return Provider{}, errors.New("id is required")
	}
	return getProvider(ctx, db, id)
}

func GetProviderBySlug(ctx context.Context, db *sqlx.DB, slug string) (Provider, error) {
	if db == nil {
		return Provider{}, errors.New("db is required")
	}
	if slug == "" {
		return Provider{}, errors.New("slug is required")
	}
	return getProviderBySlug(ctx, db, slug)
}

func ListProviders(ctx context.Context, db *sqlx.DB) ([]Provider, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	return listProviders(ctx, db)
}

func ListEnabledProviders(ctx context.Context, db *sqlx.DB) ([]PublicProvider, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	return listEnabledProviders(ctx, db)
}

func GetIdentityByProviderSubject(ctx context.Context, db *sqlx.DB, providerID, subject string) (Identity, error) {
	if db == nil {
		return Identity{}, errors.New("db is required")
	}
	return getIdentityByProviderSubject(ctx, db, providerID, subject)
}

func CreateIdentity(ctx context.Context, tx *sqlx.Tx, userID, providerID, subject, email string) (Identity, error) {
	if tx == nil {
		return Identity{}, errors.New("tx is required")
	}
	if userID == "" || providerID == "" || subject == "" || email == "" {
		return Identity{}, fmt.Errorf("all identity fields are required")
	}
	return createIdentity(ctx, tx, userID, providerID, subject, email)
}

func SetEmailVerifiedTx(ctx context.Context, tx *sqlx.Tx, userID string) error {
	if tx == nil {
		return errors.New("tx is required")
	}
	return setEmailVerifiedTx(ctx, tx, userID)
}
