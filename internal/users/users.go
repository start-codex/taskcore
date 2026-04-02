// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package users

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

const MinPasswordLength = 8

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrDuplicateEmail     = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrPasswordTooShort   = fmt.Errorf("password must be at least %d characters", MinPasswordLength)
)

type User struct {
	ID              string     `db:"id"                json:"id"`
	Email           string     `db:"email"             json:"email"`
	Name            string     `db:"name"              json:"name"`
	IsInstanceAdmin bool       `db:"is_instance_admin" json:"is_instance_admin"`
	EmailVerifiedAt *time.Time `db:"email_verified_at" json:"email_verified_at,omitempty"`
	HasPassword     bool       `db:"-"                 json:"has_password"`
	CreatedAt       time.Time  `db:"created_at"        json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"        json:"updated_at"`
	ArchivedAt      *time.Time `db:"archived_at"       json:"archived_at,omitempty"`
	PasswordHash    string     `db:"password_hash"     json:"-"`
}

// fillDerived sets computed fields that don't come from the DB directly.
func (u *User) fillDerived() {
	u.HasPassword = u.PasswordHash != ""
}

type CreateOIDCUserParams struct {
	Email string
	Name  string
}

func (p CreateOIDCUserParams) Validate() error {
	if p.Name == "" {
		return errors.New("name is required")
	}
	if !strings.Contains(p.Email, "@") || p.Email == "" {
		return errors.New("email is required and must contain @")
	}
	return nil
}

func CreateOIDCUser(ctx context.Context, db *sqlx.DB, params CreateOIDCUserParams) (User, error) {
	if db == nil {
		return User{}, errors.New("db is required")
	}
	if err := params.Validate(); err != nil {
		return User{}, err
	}
	return createOIDCUser(ctx, db, params)
}

func CreateOIDCUserTx(ctx context.Context, tx *sqlx.Tx, params CreateOIDCUserParams) (User, error) {
	if tx == nil {
		return User{}, errors.New("tx is required")
	}
	if err := params.Validate(); err != nil {
		return User{}, err
	}
	return createOIDCUserTx(ctx, tx, params)
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
	if len(params.Password) < MinPasswordLength {
		return ErrPasswordTooShort
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

// CreateInstanceAdminTx creates a user with is_instance_admin=true within an existing transaction.
func CreateInstanceAdminTx(ctx context.Context, tx *sqlx.Tx, params CreateUserParams) (User, error) {
	if tx == nil {
		return User{}, errors.New("tx is required")
	}
	if err := params.Validate(); err != nil {
		return User{}, err
	}
	return createInstanceAdminTx(ctx, tx, params)
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

func ChangePassword(ctx context.Context, db *sqlx.DB, userID, currentPassword, newPassword string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if userID == "" {
		return errors.New("userID is required")
	}
	if len(newPassword) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	hash, err := getPasswordHash(ctx, db, userID)
	if err != nil {
		return err
	}
	// OIDC-only users have no password — cannot change via this flow
	if hash == "" {
		return ErrInvalidCredentials
	}
	ok, err := verifyPassword(hash, currentPassword)
	if err != nil {
		return fmt.Errorf("verify password: %w", err)
	}
	if !ok {
		return ErrInvalidCredentials
	}
	newHash, err := hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return updatePassword(ctx, db, userID, newHash)
}

// SetPassword sets a new password for the user without verifying the current one.
// Used for password reset flows. Validates minimum length.
func SetPassword(ctx context.Context, db *sqlx.DB, userID, newPassword string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if userID == "" {
		return errors.New("userID is required")
	}
	if len(newPassword) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	newHash, err := hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return updatePassword(ctx, db, userID, newHash)
}

// SetPasswordTx sets a new password within an existing transaction.
// Used for atomic password reset flows.
func SetPasswordTx(ctx context.Context, tx *sqlx.Tx, userID, newPassword string) error {
	if tx == nil {
		return errors.New("tx is required")
	}
	if userID == "" {
		return errors.New("userID is required")
	}
	if len(newPassword) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	newHash, err := hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return updatePasswordTx(ctx, tx, userID, newHash)
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
