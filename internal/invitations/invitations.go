// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package invitations

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/tookly/internal/sessions"
)

const InvitationTTL = 7 * 24 * time.Hour

var (
	ErrInvitationNotFound  = errors.New("invitation not found")
	ErrInvitationExpired   = errors.New("invitation expired")
	ErrInvitationRevoked   = errors.New("invitation revoked")
	ErrInvitationUsed      = errors.New("invitation already accepted")
	ErrDuplicateInvitation = errors.New("pending invitation already exists for this email")
	ErrAlreadyMember       = errors.New("user is already a workspace member")
)

var validInviteRoles = map[string]bool{"admin": true, "member": true}

type Invitation struct {
	ID          string     `db:"id"           json:"id"`
	WorkspaceID string     `db:"workspace_id" json:"workspace_id"`
	Email       string     `db:"email"        json:"email"`
	Role        string     `db:"role"         json:"role"`
	InvitedBy   string     `db:"invited_by"   json:"invited_by"`
	TokenHash   string     `db:"token_hash"   json:"-"`
	Status      string     `db:"status"       json:"status"`
	ExpiresAt   time.Time  `db:"expires_at"   json:"expires_at"`
	AcceptedAt  *time.Time `db:"accepted_at"  json:"accepted_at,omitempty"`
	CreatedAt   time.Time  `db:"created_at"   json:"created_at"`
}

type CreateInvitationParams struct {
	WorkspaceID string
	Email       string
	Role        string
	InvitedBy   string
}

func (p CreateInvitationParams) Validate() error {
	if p.WorkspaceID == "" {
		return errors.New("workspace_id is required")
	}
	if p.Email == "" {
		return errors.New("email is required")
	}
	if !validInviteRoles[p.Role] {
		return errors.New("role must be 'admin' or 'member'")
	}
	if p.InvitedBy == "" {
		return errors.New("invited_by is required")
	}
	return nil
}

func CreateInvitation(ctx context.Context, db *sqlx.DB, params CreateInvitationParams) (string, Invitation, error) {
	if db == nil {
		return "", Invitation{}, errors.New("db is required")
	}
	if err := params.Validate(); err != nil {
		return "", Invitation{}, err
	}

	// Check if already a member
	isMember, err := isWorkspaceMemberByEmail(ctx, db, params.WorkspaceID, params.Email)
	if err != nil {
		return "", Invitation{}, fmt.Errorf("check membership: %w", err)
	}
	if isMember {
		return "", Invitation{}, ErrAlreadyMember
	}

	rawToken, err := sessions.GenerateToken()
	if err != nil {
		return "", Invitation{}, fmt.Errorf("generate token: %w", err)
	}
	tokenHash := sessions.HashToken(rawToken)
	expiresAt := time.Now().Add(InvitationTTL)

	inv, err := createInvitation(ctx, db, params, tokenHash, expiresAt)
	if err != nil {
		return "", Invitation{}, err
	}
	return rawToken, inv, nil
}

func GetInvitation(ctx context.Context, db *sqlx.DB, rawToken string) (Invitation, error) {
	if db == nil {
		return Invitation{}, errors.New("db is required")
	}
	if rawToken == "" {
		return Invitation{}, errors.New("token is required")
	}
	tokenHash := sessions.HashToken(rawToken)
	inv, err := getInvitationByToken(ctx, db, tokenHash)
	if err != nil {
		return Invitation{}, err
	}
	if inv.Status == "revoked" {
		return Invitation{}, ErrInvitationRevoked
	}
	if inv.Status == "accepted" {
		return Invitation{}, ErrInvitationUsed
	}
	if time.Now().After(inv.ExpiresAt) {
		return Invitation{}, ErrInvitationExpired
	}
	return inv, nil
}

func AcceptInvitation(ctx context.Context, db *sqlx.DB, rawToken, userID string) error {
	if db == nil {
		return errors.New("db is required")
	}
	inv, err := GetInvitation(ctx, db, rawToken)
	if err != nil {
		return err
	}

	tokenHash := sessions.HashToken(rawToken)

	// Atomic: add member + mark accepted in one transaction
	tx, txErr := db.BeginTxx(ctx, nil)
	if txErr != nil {
		return fmt.Errorf("begin tx: %w", txErr)
	}
	defer tx.Rollback()

	// Add workspace member via tx
	_, err = tx.ExecContext(ctx,
		`INSERT INTO workspace_members (workspace_id, user_id, role)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (workspace_id, user_id)
		 DO UPDATE SET role = excluded.role, archived_at = NULL`,
		inv.WorkspaceID, userID, inv.Role,
	)
	if err != nil {
		return fmt.Errorf("add member: %w", err)
	}

	// Mark invitation as accepted
	_, err = tx.ExecContext(ctx,
		`UPDATE invitations SET status = 'accepted', accepted_at = NOW() WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("accept invitation: %w", err)
	}

	return tx.Commit()
}

func ListPending(ctx context.Context, db *sqlx.DB, workspaceID string) ([]Invitation, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	if workspaceID == "" {
		return nil, errors.New("workspace_id is required")
	}
	return listPending(ctx, db, workspaceID)
}

func Revoke(ctx context.Context, db *sqlx.DB, invitationID string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if invitationID == "" {
		return errors.New("invitation_id is required")
	}
	return revokeInvitation(ctx, db, invitationID)
}

func Resend(ctx context.Context, db *sqlx.DB, invitationID string) (string, error) {
	if db == nil {
		return "", errors.New("db is required")
	}
	if invitationID == "" {
		return "", errors.New("invitation_id is required")
	}
	rawToken, err := sessions.GenerateToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	tokenHash := sessions.HashToken(rawToken)
	expiresAt := time.Now().Add(InvitationTTL)
	if err := resendInvitation(ctx, db, invitationID, tokenHash, expiresAt); err != nil {
		return "", err
	}
	return rawToken, nil
}

func GetInvitationByID(ctx context.Context, db *sqlx.DB, id string) (Invitation, error) {
	if db == nil {
		return Invitation{}, errors.New("db is required")
	}
	if id == "" {
		return Invitation{}, errors.New("id is required")
	}
	return getInvitationByID(ctx, db, id)
}
