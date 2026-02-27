package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const userCols = `id, email, name, created_at, updated_at, archived_at`

func createUser(ctx context.Context, db *sqlx.DB, params CreateUserParams) (User, error) {
	hash, err := hashPassword(params.Password)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}
	var user User
	err = db.QueryRowxContext(ctx,
		`INSERT INTO app_users (email, name, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING `+userCols,
		params.Email, params.Name, hash,
	).StructScan(&user)
	if err != nil {
		if isUniqueViolation(err) {
			return User{}, ErrDuplicateEmail
		}
		return User{}, fmt.Errorf("insert user: %w", err)
	}
	return user, nil
}

func getUser(ctx context.Context, db *sqlx.DB, id string) (User, error) {
	var user User
	err := db.GetContext(ctx, &user,
		`SELECT `+userCols+` FROM app_users WHERE id = $1`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func getUserByEmail(ctx context.Context, db *sqlx.DB, email string) (User, error) {
	var user User
	err := db.GetContext(ctx, &user,
		`SELECT `+userCols+` FROM app_users WHERE email = $1`,
		email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func authenticateUser(ctx context.Context, db *sqlx.DB, email, password string) (User, error) {
	var user User
	err := db.GetContext(ctx, &user,
		`SELECT `+userCols+`, password_hash FROM app_users WHERE email = $1`,
		email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrInvalidCredentials
		}
		return User{}, fmt.Errorf("get user for auth: %w", err)
	}
	ok, err := verifyPassword(user.PasswordHash, password)
	if err != nil {
		return User{}, fmt.Errorf("verify password: %w", err)
	}
	if !ok {
		return User{}, ErrInvalidCredentials
	}
	user.PasswordHash = ""
	return user, nil
}

func archiveUser(ctx context.Context, db *sqlx.DB, id string) error {
	res, err := db.ExecContext(ctx,
		`UPDATE app_users
		 SET archived_at = NOW()
		 WHERE id = $1 AND archived_at IS NULL`,
		id,
	)
	if err != nil {
		return fmt.Errorf("archive user: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive user rows affected: %w", err)
	}
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
