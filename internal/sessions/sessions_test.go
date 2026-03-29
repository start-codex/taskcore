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
	"testing"

	"github.com/jmoiron/sqlx"
)

// fakeDB returns a non-nil *sqlx.DB that is not connected to any database.
// Useful for exercising validation branches that check parameters before
// touching the database.
func fakeDB(t *testing.T) *sqlx.DB {
	t.Helper()
	raw, err := sql.Open("postgres", "")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	return sqlx.NewDb(raw, "postgres")
}

func TestGenerateToken(t *testing.T) {
	token1, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if len(token1) != 64 { // 32 bytes = 64 hex characters
		t.Fatalf("GenerateToken() length = %d, want 64", len(token1))
	}

	token2, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if token1 == token2 {
		t.Fatalf("GenerateToken() generated identical tokens")
	}
}

func TestHashToken(t *testing.T) {
	token := "abc123"
	h1 := HashToken(token)
	h2 := HashToken(token)
	if h1 != h2 {
		t.Fatalf("HashToken() not deterministic: %q != %q", h1, h2)
	}
	if len(h1) != 64 { // SHA-256 = 32 bytes = 64 hex chars
		t.Fatalf("HashToken() length = %d, want 64", len(h1))
	}
	if h1 == token {
		t.Fatal("HashToken() returned the raw token")
	}
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "session not found", err: ErrSessionNotFound, want: true},
		{name: "session expired", err: ErrSessionExpired, want: true},
		{name: "user archived", err: ErrUserArchived, want: true},
		{name: "wrapped auth error", err: fmt.Errorf("validate session: %w", ErrSessionExpired), want: true},
		{name: "generic error", err: errors.New("db down"), want: false},
		{name: "nil", err: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAuthError(tt.err); got != tt.want {
				t.Fatalf("IsAuthError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name    string
		db      *sqlx.DB
		userID  string
		wantErr string
	}{
		{name: "nil db", db: nil, userID: "u1", wantErr: "db is required"},
		{name: "empty userID", db: fakeDB(t), userID: "", wantErr: "userID is required"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Create(context.Background(), tc.db, tc.userID)
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("Create() error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		db      *sqlx.DB
		token   string
		wantErr string
	}{
		{name: "nil db", db: nil, token: "tok", wantErr: "db is required"},
		{name: "empty token", db: fakeDB(t), token: "", wantErr: "token is required"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Validate(context.Background(), tc.db, tc.token)
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("Validate() error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name    string
		db      *sqlx.DB
		token   string
		wantErr string
	}{
		{name: "nil db", db: nil, token: "tok", wantErr: "db is required"},
		{name: "empty token", db: fakeDB(t), token: "", wantErr: "token is required"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Delete(context.Background(), tc.db, tc.token)
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("Delete() error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}
