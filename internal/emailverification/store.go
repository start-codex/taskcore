// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package emailverification

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type verificationToken struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	TokenHash string     `db:"token_hash"`
	ExpiresAt time.Time  `db:"expires_at"`
	UsedAt    *time.Time `db:"used_at"`
	CreatedAt time.Time  `db:"created_at"`
}

func createToken(ctx context.Context, db *sqlx.DB, userID, tokenHash string, expiresAt time.Time) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert verification token: %w", err)
	}
	return nil
}

func getTokenByHash(ctx context.Context, db *sqlx.DB, tokenHash string) (verificationToken, error) {
	var token verificationToken
	err := db.GetContext(ctx, &token,
		`SELECT id, user_id, token_hash, expires_at, used_at, created_at
		 FROM email_verification_tokens WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return verificationToken{}, ErrTokenNotFound
		}
		return verificationToken{}, fmt.Errorf("get verification token: %w", err)
	}
	return token, nil
}

func markTokenUsed(ctx context.Context, db *sqlx.DB, tokenHash string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE email_verification_tokens SET used_at = NOW() WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("mark verification token used: %w", err)
	}
	return nil
}

func setEmailVerified(ctx context.Context, db *sqlx.DB, userID string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE app_users SET email_verified_at = NOW(), updated_at = NOW() WHERE id = $1 AND email_verified_at IS NULL`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("set email verified: %w", err)
	}
	return nil
}
