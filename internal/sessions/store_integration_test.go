package sessions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/start-codex/taskcode/internal/testpg"
	"github.com/start-codex/taskcode/internal/users"
)

func TestCreateSession(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (string, func(*testing.T))
		wantErr error
	}{
		{
			name: "creates session successfully",
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				u := seedUser(t, db)
				return u.ID, func(t *testing.T) {}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, check := tt.arrange(t, db)
			got, err := Create(context.Background(), db, userID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Create() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil {
				if got.RawToken == "" {
					t.Fatal("expected non-empty raw token")
				}
				if len(got.RawToken) != 64 {
					t.Fatalf("raw token length = %d, want 64", len(got.RawToken))
				}
				// Session.ID should be the hash, not the raw token
				if got.Session.ID == got.RawToken {
					t.Fatal("session ID should be hashed, not raw token")
				}
				if got.Session.ID != HashToken(got.RawToken) {
					t.Fatal("session ID should equal HashToken(rawToken)")
				}
				if got.Session.UserID != userID {
					t.Fatalf("user_id: got %q, want %q", got.Session.UserID, userID)
				}
				if got.Session.CreatedAt.IsZero() {
					t.Fatal("expected non-zero created_at")
				}
				if got.Session.ExpiresAt.IsZero() {
					t.Fatal("expected non-zero expires_at")
				}
				if got.Session.LastUsedAt == nil || got.Session.LastUsedAt.IsZero() {
					t.Fatal("expected non-zero last_used_at")
				}
				expectedExpiry := time.Now().Add(DefaultSessionTTL)
				if got.Session.ExpiresAt.Before(expectedExpiry.Add(-time.Minute)) || got.Session.ExpiresAt.After(expectedExpiry.Add(time.Minute)) {
					t.Fatalf("expires_at: got %v, want approximately %v", got.Session.ExpiresAt, expectedExpiry)
				}
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestValidateSession(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (string, func(*testing.T))
		wantErr error
	}{
		{
			name: "validates existing session",
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				result := seedSession(t, db)
				return result.RawToken, func(t *testing.T) {}
			},
		},
		{
			name:    "session not found",
			wantErr: ErrSessionNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				return "nonexistent-token-that-does-not-exist-in-database-00000000", nil
			},
		},
		{
			name:    "session expired",
			wantErr: ErrSessionExpired,
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				u := seedUser(t, db)
				result, err := createSession(context.Background(), db, u.ID, -time.Hour)
				if err != nil {
					t.Fatalf("create expired session: %v", err)
				}
				t.Cleanup(func() {
					db.ExecContext(context.Background(), `DELETE FROM sessions WHERE id = $1`, result.Session.ID)
				})
				return result.RawToken, nil
			},
		},
		{
			name:    "user archived",
			wantErr: ErrUserArchived,
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				u := seedUser(t, db)
				result, err := createSession(context.Background(), db, u.ID, DefaultSessionTTL)
				if err != nil {
					t.Fatalf("create session: %v", err)
				}
				// Archive the user
				_, err = db.ExecContext(context.Background(),
					`UPDATE app_users SET archived_at = NOW() WHERE id = $1`, u.ID)
				if err != nil {
					t.Fatalf("archive user: %v", err)
				}
				t.Cleanup(func() {
					db.ExecContext(context.Background(), `DELETE FROM sessions WHERE id = $1`, result.Session.ID)
					db.ExecContext(context.Background(), `UPDATE app_users SET archived_at = NULL WHERE id = $1`, u.ID)
				})
				return result.RawToken, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, check := tt.arrange(t, db)
			got, err := Validate(context.Background(), db, token)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil {
				if got.ID != HashToken(token) {
					t.Fatalf("session id: got %q, want hash of token", got.ID)
				}
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestDeleteSession(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) string
		wantErr error
	}{
		{
			name: "deletes existing session",
			arrange: func(t *testing.T, db *sqlx.DB) string {
				result := seedSession(t, db)
				return result.RawToken
			},
		},
		{
			name: "idempotent for nonexistent session",
			arrange: func(t *testing.T, db *sqlx.DB) string {
				return "nonexistent-token-00000000000000000000000000000000000000"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.arrange(t, db)
			err := Delete(context.Background(), db, token)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Delete() error = %v, wantErr = %v", err, tt.wantErr)
			}

			_, err = Validate(context.Background(), db, token)
			if !errors.Is(err, ErrSessionNotFound) {
				t.Fatalf("after Delete(), Validate() error = %v, want ErrSessionNotFound", err)
			}
		})
	}
}

func TestSessionLifecycle(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	u := seedUser(t, db)

	// Create session
	result, err := Create(context.Background(), db, u.ID)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	rawToken := result.RawToken
	t.Cleanup(func() {
		db.ExecContext(context.Background(), `DELETE FROM sessions WHERE id = $1`, result.Session.ID)
	})

	// Validate session using raw token
	validated, err := Validate(context.Background(), db, rawToken)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if validated.ID != HashToken(rawToken) {
		t.Fatalf("validated session id: got %q, want hash of raw token", validated.ID)
	}

	// Delete session using raw token
	err = Delete(context.Background(), db, rawToken)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify session is deleted
	_, err = Validate(context.Background(), db, rawToken)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("after Delete(), Validate() error = %v, want ErrSessionNotFound", err)
	}

	// Delete again (idempotent)
	err = Delete(context.Background(), db, rawToken)
	if err != nil {
		t.Fatalf("Delete() second time error = %v, want nil (idempotent)", err)
	}
}

// --- helpers ---

func seedUser(t *testing.T, db *sqlx.DB) users.User {
	t.Helper()
	suffix := testpg.UniqueSuffix(t, db)
	u, err := users.CreateUser(context.Background(), db, users.CreateUserParams{
		Email:    "session-test-" + suffix + "@test.local",
		Name:     "Session Test User",
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

func seedSession(t *testing.T, db *sqlx.DB) CreateResult {
	t.Helper()
	u := seedUser(t, db)
	result, err := Create(context.Background(), db, u.ID)
	if err != nil {
		t.Fatalf("seed session: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), `DELETE FROM sessions WHERE id = $1`, result.Session.ID)
	})
	return result
}
