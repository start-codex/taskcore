// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package passwordreset

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

func createToken(ctx context.Context, db *sqlx.DB, userID, tokenHash string, expiresAt time.Time) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert reset token: %w", err)
	}
	return nil
}

func getTokenByHash(ctx context.Context, db *sqlx.DB, tokenHash string) (ResetToken, error) {
	var token ResetToken
	err := db.GetContext(ctx, &token,
		`SELECT id, user_id, token_hash, expires_at, used_at, created_at
		 FROM password_reset_tokens WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return ResetToken{}, ErrTokenNotFound
		}
		return ResetToken{}, fmt.Errorf("get reset token: %w", err)
	}
	return token, nil
}

func markTokenUsed(ctx context.Context, db *sqlx.DB, tokenHash string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = NOW() WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("mark token used: %w", err)
	}
	return nil
}

func markTokenUsedTx(ctx context.Context, tx *sqlx.Tx, tokenHash string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = NOW() WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("mark token used tx: %w", err)
	}
	return nil
}
