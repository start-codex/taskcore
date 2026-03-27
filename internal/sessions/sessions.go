package sessions

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
	ErrUserArchived    = errors.New("user account is archived")
)

// DefaultSessionTTL is the default time-to-live for a session (7 days)
const DefaultSessionTTL = 7 * 24 * time.Hour

// Session represents a user session.
// ID is the hashed token stored in the database, not the raw bearer token.
type Session struct {
	ID         string     `db:"id"           json:"id"`
	UserID     string     `db:"user_id"      json:"user_id"`
	CreatedAt  time.Time  `db:"created_at"   json:"created_at"`
	ExpiresAt  time.Time  `db:"expires_at"   json:"expires_at"`
	LastUsedAt *time.Time `db:"last_used_at" json:"last_used_at,omitempty"`
}

// GenerateToken generates a cryptographically random 32-byte session token
// encoded as a hex string.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashToken returns the SHA-256 hex digest of a raw token.
// The hash is what gets stored in the database; the raw token is the bearer
// credential returned to the client.
func HashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// Create creates a new session for the given user.
// It returns the Session whose ID is the hashed token. The raw token is
// available only at creation time via the second return value of
// createSession (the caller of this function receives it in Session.ID
// temporarily replaced — see store.go for details).
//
// The caller must capture the raw token from CreateResult and send it to
// the client; it cannot be recovered from the database.
func Create(ctx context.Context, db *sqlx.DB, userID string) (CreateResult, error) {
	if db == nil {
		return CreateResult{}, errors.New("db is required")
	}
	if userID == "" {
		return CreateResult{}, errors.New("userID is required")
	}
	return createSession(ctx, db, userID, DefaultSessionTTL)
}

// CreateResult holds both the persisted session and the raw (unhashed)
// token that must be sent to the client exactly once.
type CreateResult struct {
	Session  Session
	RawToken string
}

// IsAuthError reports whether err means the session should be treated as
// unauthenticated rather than as an internal server failure.
func IsAuthError(err error) bool {
	return errors.Is(err, ErrSessionNotFound) ||
		errors.Is(err, ErrSessionExpired) ||
		errors.Is(err, ErrUserArchived)
}

// Validate validates a session by its raw token.
// The token is hashed before lookup. Returns ErrSessionNotFound if the
// session does not exist, ErrSessionExpired if expired, or ErrUserArchived
// if the owning user account has been archived.
func Validate(ctx context.Context, db *sqlx.DB, token string) (Session, error) {
	if db == nil {
		return Session{}, errors.New("db is required")
	}
	if token == "" {
		return Session{}, errors.New("token is required")
	}
	return validateSession(ctx, db, token)
}

// Delete deletes a session by its raw token.
// The token is hashed before lookup.
// Does not return an error if the session does not exist (idempotent).
func Delete(ctx context.Context, db *sqlx.DB, token string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if token == "" {
		return errors.New("token is required")
	}
	return deleteSession(ctx, db, token)
}
