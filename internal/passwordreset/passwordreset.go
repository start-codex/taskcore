// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package passwordreset

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/tookly/internal/sessions"
	"github.com/start-codex/tookly/internal/users"
)

const TokenTTL = 1 * time.Hour

var (
	ErrTokenNotFound = errors.New("reset token not found")
	ErrTokenExpired  = errors.New("reset token expired")
	ErrTokenUsed     = errors.New("reset token already used")
)

type ResetToken struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	TokenHash string     `db:"token_hash"`
	ExpiresAt time.Time  `db:"expires_at"`
	UsedAt    *time.Time `db:"used_at"`
	CreatedAt time.Time  `db:"created_at"`
}

// CreateToken generates a password reset token for the given user.
// Returns the raw token (to be sent via email) and an error.
func CreateToken(ctx context.Context, db *sqlx.DB, userID string) (string, error) {
	if db == nil {
		return "", errors.New("db is required")
	}
	if userID == "" {
		return "", errors.New("userID is required")
	}
	rawToken, err := sessions.GenerateToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	tokenHash := sessions.HashToken(rawToken)
	expiresAt := time.Now().Add(TokenTTL)
	if err := createToken(ctx, db, userID, tokenHash, expiresAt); err != nil {
		return "", err
	}
	return rawToken, nil
}

// ValidateToken checks if a reset token is valid (exists, not used, not expired).
// Returns the user ID associated with the token.
func ValidateToken(ctx context.Context, db *sqlx.DB, rawToken string) (string, error) {
	if db == nil {
		return "", errors.New("db is required")
	}
	if rawToken == "" {
		return "", errors.New("token is required")
	}
	tokenHash := sessions.HashToken(rawToken)
	token, err := getTokenByHash(ctx, db, tokenHash)
	if err != nil {
		return "", err
	}
	if token.UsedAt != nil {
		return "", ErrTokenUsed
	}
	if time.Now().After(token.ExpiresAt) {
		return "", ErrTokenExpired
	}
	return token.UserID, nil
}

// ResetPassword validates the token and resets the user's password atomically.
// All operations (password update, token mark used) happen within a transaction.
// Session invalidation happens after commit (best-effort).
func ResetPassword(ctx context.Context, db *sqlx.DB, rawToken, newPassword string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if rawToken == "" {
		return errors.New("token is required")
	}
	if len(newPassword) < users.MinPasswordLength {
		return users.ErrPasswordTooShort
	}

	// Validate token first (outside tx to fail fast)
	tokenHash := sessions.HashToken(rawToken)
	token, err := getTokenByHash(ctx, db, tokenHash)
	if err != nil {
		return err
	}
	if token.UsedAt != nil {
		return ErrTokenUsed
	}
	if time.Now().After(token.ExpiresAt) {
		return ErrTokenExpired
	}

	// Atomic: update password + mark token used
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := users.SetPasswordTx(ctx, tx, token.UserID, newPassword); err != nil {
		return fmt.Errorf("set password: %w", err)
	}

	if err := markTokenUsedTx(ctx, tx, tokenHash); err != nil {
		return fmt.Errorf("mark token used: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	// Best-effort session invalidation after commit
	_ = sessions.DeleteByUserID(ctx, db, token.UserID, "")

	return nil
}
