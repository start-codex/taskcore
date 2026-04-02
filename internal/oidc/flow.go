// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package oidc

import (
	"context"
	"fmt"
	"strings"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// IDTokenClaims holds the claims extracted from a verified ID token.
type IDTokenClaims struct {
	Subject string
	Email   string
	Name    string
}

// newOAuth2Config builds an oauth2.Config from a Provider.
// The authorization and token endpoints are discovered from the issuer.
func newOAuth2Config(ctx context.Context, p Provider) (*oauth2.Config, *gooidc.Provider, error) {
	oidcProvider, err := gooidc.NewProvider(ctx, p.IssuerURL)
	if err != nil {
		return nil, nil, fmt.Errorf("oidc discovery for %s: %w", p.IssuerURL, err)
	}
	scopes := strings.Fields(p.Scopes)
	if len(scopes) == 0 {
		scopes = []string{gooidc.ScopeOpenID, "email", "profile"}
	}
	cfg := &oauth2.Config{
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		Endpoint:     oidcProvider.Endpoint(),
		RedirectURL:  p.RedirectURI,
		Scopes:       scopes,
	}
	return cfg, oidcProvider, nil
}

// verifyIDToken verifies the raw ID token and validates the nonce.
// Returns the extracted claims.
func verifyIDToken(ctx context.Context, p Provider, oidcProvider *gooidc.Provider, rawIDToken, expectedNonce string) (*IDTokenClaims, error) {
	verifier := oidcProvider.Verifier(&gooidc.Config{ClientID: p.ClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("verify id token: %w", err)
	}
	if idToken.Nonce != expectedNonce {
		return nil, fmt.Errorf("nonce mismatch")
	}
	var claims struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("parse id token claims: %w", err)
	}
	return &IDTokenClaims{
		Subject: idToken.Subject,
		Email:   claims.Email,
		Name:    claims.Name,
	}, nil
}
