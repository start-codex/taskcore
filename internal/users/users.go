package users

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrDuplicateEmail     = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type User struct {
	ID           string     `db:"id"            json:"id"`
	Email        string     `db:"email"         json:"email"`
	Name         string     `db:"name"          json:"name"`
	CreatedAt    time.Time  `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"    json:"updated_at"`
	ArchivedAt   *time.Time `db:"archived_at"   json:"archived_at,omitempty"`
	PasswordHash string     `db:"password_hash" json:"-"`
}

type CreateUserParams struct {
	Email    string
	Name     string
	Password string
}

func (params CreateUserParams) Validate() error {
	if params.Name == "" {
		return errors.New("name is required")
	}
	if !strings.Contains(params.Email, "@") || params.Email == "" {
		return errors.New("email is required and must contain @")
	}
	if params.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

func CreateUser(ctx context.Context, db *sqlx.DB, params CreateUserParams) (User, error) {
	if db == nil {
		return User{}, errors.New("db is required")
	}
	if err := params.Validate(); err != nil {
		return User{}, err
	}
	return createUser(ctx, db, params)
}

func GetUser(ctx context.Context, db *sqlx.DB, id string) (User, error) {
	if db == nil {
		return User{}, errors.New("db is required")
	}
	if id == "" {
		return User{}, errors.New("id is required")
	}
	return getUser(ctx, db, id)
}

func GetUserByEmail(ctx context.Context, db *sqlx.DB, email string) (User, error) {
	if db == nil {
		return User{}, errors.New("db is required")
	}
	if email == "" {
		return User{}, errors.New("email is required")
	}
	return getUserByEmail(ctx, db, email)
}

func ArchiveUser(ctx context.Context, db *sqlx.DB, id string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if id == "" {
		return errors.New("id is required")
	}
	return archiveUser(ctx, db, id)
}

func AuthenticateUser(ctx context.Context, db *sqlx.DB, email, password string) (User, error) {
	if db == nil {
		return User{}, errors.New("db is required")
	}
	if email == "" || password == "" {
		return User{}, ErrInvalidCredentials
	}
	return authenticateUser(ctx, db, email, password)
}
