// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package users

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/start-codex/trazawork/internal/testpg"
)

func TestCreateUser(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (CreateUserParams, func(*testing.T))
		wantErr error
	}{
		{
			name: "creates user successfully",
			arrange: func(t *testing.T, db *sqlx.DB) (CreateUserParams, func(*testing.T)) {
				params := CreateUserParams{Email: uniqueEmail(t, db), Name: "Alice", Password: "pass123"}
				return params, func(t *testing.T) {}
			},
		},
		{
			name:    "duplicate email",
			wantErr: ErrDuplicateEmail,
			arrange: func(t *testing.T, db *sqlx.DB) (CreateUserParams, func(*testing.T)) {
				email := uniqueEmail(t, db)
				if _, err := CreateUser(context.Background(), db, CreateUserParams{Email: email, Name: "First", Password: "pass"}); err != nil {
					t.Fatalf("seed user: %v", err)
				}
				return CreateUserParams{Email: email, Name: "Second", Password: "pass"}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, check := tt.arrange(t, db)
			got, err := CreateUser(context.Background(), db, params)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateUser() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil {
				if got.ID == "" {
					t.Fatal("expected non-empty id")
				}
				if got.Email != params.Email {
					t.Fatalf("email: got %q, want %q", got.Email, params.Email)
				}
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (string, func(*testing.T))
		wantErr error
	}{
		{
			name: "returns existing user",
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				u := seedUser(t, db)
				return u.ID, func(t *testing.T) {}
			},
		},
		{
			name:    "not found",
			wantErr: ErrUserNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				return "00000000-0000-0000-0000-000000000000", nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, check := tt.arrange(t, db)
			got, err := GetUser(context.Background(), db, id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetUser() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil && got.ID != id {
				t.Fatalf("id: got %q, want %q", got.ID, id)
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestGetUserByEmail(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (string, func(*testing.T))
		wantErr error
	}{
		{
			name: "returns user by email",
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				u := seedUser(t, db)
				return u.Email, func(t *testing.T) {}
			},
		},
		{
			name:    "not found",
			wantErr: ErrUserNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				return "nobody@does-not-exist.local", nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, check := tt.arrange(t, db)
			got, err := GetUserByEmail(context.Background(), db, email)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetUserByEmail() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil && got.Email != email {
				t.Fatalf("email: got %q, want %q", got.Email, email)
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestArchiveUser(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) string
		wantErr error
	}{
		{
			name: "archives active user",
			arrange: func(t *testing.T, db *sqlx.DB) string {
				return seedUser(t, db).ID
			},
		},
		{
			name:    "not found",
			wantErr: ErrUserNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) string {
				return "00000000-0000-0000-0000-000000000000"
			},
		},
		{
			name:    "already archived",
			wantErr: ErrUserNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) string {
				u := seedUser(t, db)
				if err := ArchiveUser(context.Background(), db, u.ID); err != nil {
					t.Fatalf("first archive: %v", err)
				}
				return u.ID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.arrange(t, db)
			err := ArchiveUser(context.Background(), db, id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ArchiveUser() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

// --- helpers ---

func uniqueEmail(t *testing.T, db *sqlx.DB) string {
	t.Helper()
	suffix := testpg.UniqueSuffix(t, db)
	return fmt.Sprintf("user-%s@test.local", suffix)
}

func seedUser(t *testing.T, db *sqlx.DB) User {
	t.Helper()
	u, err := CreateUser(context.Background(), db, CreateUserParams{
		Email:    uniqueEmail(t, db),
		Name:     "Test User",
		Password: "testpass123",
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), `DELETE FROM app_users WHERE id = $1`, u.ID)
	})
	return u
}
