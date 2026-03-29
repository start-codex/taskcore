// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package sessions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

const sessionCols = `s.id, s.user_id, s.created_at, s.expires_at, s.last_used_at`

func createSession(ctx context.Context, db *sqlx.DB, userID string, ttl time.Duration) (CreateResult, error) {
	rawToken, err := GenerateToken()
	if err != nil {
		return CreateResult{}, fmt.Errorf("generate token: %w", err)
	}

	hashedToken := HashToken(rawToken)
	now := time.Now()
	expiresAt := now.Add(ttl)

	var session Session
	err = db.QueryRowxContext(ctx,
		`INSERT INTO sessions (id, user_id, created_at, expires_at, last_used_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, created_at, expires_at, last_used_at`,
		hashedToken, userID, now, expiresAt, now,
	).StructScan(&session)
	if err != nil {
		return CreateResult{}, fmt.Errorf("insert session: %w", err)
	}
	return CreateResult{Session: session, RawToken: rawToken}, nil
}

func validateSession(ctx context.Context, db *sqlx.DB, rawToken string) (Session, error) {
	hashedToken := HashToken(rawToken)

	var session Session
	err := db.GetContext(ctx, &session,
		`SELECT `+sessionCols+`
		 FROM sessions s
		 JOIN app_users u ON u.id = s.user_id
		 WHERE s.id = $1
		   AND u.archived_at IS NULL`,
		hashedToken,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Distinguish "session missing" from "user archived":
			// check if the session row exists at all.
			var exists bool
			if err2 := db.GetContext(ctx, &exists,
				`SELECT EXISTS(SELECT 1 FROM sessions WHERE id = $1)`, hashedToken); err2 != nil {
				return Session{}, fmt.Errorf("check session existence: %w", err2)
			}
			if exists {
				return Session{}, ErrUserArchived
			}
			return Session{}, ErrSessionNotFound
		}
		return Session{}, fmt.Errorf("get session: %w", err)
	}

	if time.Now().After(session.ExpiresAt) {
		return Session{}, ErrSessionExpired
	}

	return session, nil
}

func deleteSession(ctx context.Context, db *sqlx.DB, rawToken string) error {
	hashedToken := HashToken(rawToken)
	_, err := db.ExecContext(ctx,
		`DELETE FROM sessions WHERE id = $1`,
		hashedToken,
	)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
