// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package passwordreset

import (
	"context"
	"errors"
	"testing"

	_ "github.com/lib/pq"
	"github.com/start-codex/tookly/internal/sessions"
	"github.com/start-codex/tookly/internal/testpg"
)

func TestCreateAndValidateToken(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	userID := testpg.SeedUser(t, db)
	ctx := context.Background()

	rawToken, err := CreateToken(ctx, db, userID)
	if err != nil {
		t.Fatalf("CreateToken error = %v", err)
	}
	if rawToken == "" {
		t.Fatal("CreateToken returned empty token")
	}

	gotUserID, err := ValidateToken(ctx, db, rawToken)
	if err != nil {
		t.Fatalf("ValidateToken error = %v", err)
	}
	if gotUserID != userID {
		t.Fatalf("ValidateToken userID = %q, want %q", gotUserID, userID)
	}
}

func TestValidateToken_NotFound(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	_, err := ValidateToken(context.Background(), db, "nonexistent_token")
	if !errors.Is(err, ErrTokenNotFound) {
		t.Fatalf("error = %v, want ErrTokenNotFound", err)
	}
}

func TestValidateToken_Used(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	userID := testpg.SeedUser(t, db)
	ctx := context.Background()

	rawToken, _ := CreateToken(ctx, db, userID)
	hash := sessions.HashToken(rawToken)
	db.ExecContext(ctx, `UPDATE password_reset_tokens SET used_at = NOW() WHERE token_hash = $1`, hash)

	_, err := ValidateToken(ctx, db, rawToken)
	if !errors.Is(err, ErrTokenUsed) {
		t.Fatalf("error = %v, want ErrTokenUsed", err)
	}
}

func TestValidateToken_Expired(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	userID := testpg.SeedUser(t, db)
	ctx := context.Background()

	rawToken, _ := CreateToken(ctx, db, userID)
	hash := sessions.HashToken(rawToken)
	db.ExecContext(ctx, `UPDATE password_reset_tokens SET expires_at = NOW() - INTERVAL '1 hour' WHERE token_hash = $1`, hash)

	_, err := ValidateToken(ctx, db, rawToken)
	if !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("error = %v, want ErrTokenExpired", err)
	}
}

func TestResetPassword_EndToEnd(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	userID := testpg.SeedUser(t, db)
	ctx := context.Background()

	rawToken, _ := CreateToken(ctx, db, userID)

	err := ResetPassword(ctx, db, rawToken, "newpassword123")
	if err != nil {
		t.Fatalf("ResetPassword error = %v", err)
	}

	// Token should now be used
	_, err = ValidateToken(ctx, db, rawToken)
	if !errors.Is(err, ErrTokenUsed) {
		t.Fatalf("token should be used after reset, got error = %v", err)
	}
}
