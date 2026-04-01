// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package invitations

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/tookly/internal/pgutil"
)

const invCols = `id, workspace_id, email, role, invited_by, token_hash, status, expires_at, accepted_at, created_at`

func createInvitation(ctx context.Context, db *sqlx.DB, params CreateInvitationParams, tokenHash string, expiresAt time.Time) (Invitation, error) {
	var inv Invitation
	err := db.QueryRowxContext(ctx,
		`INSERT INTO invitations (workspace_id, email, role, invited_by, token_hash, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING `+invCols,
		params.WorkspaceID, params.Email, params.Role, params.InvitedBy, tokenHash, expiresAt,
	).StructScan(&inv)
	if err != nil {
		if pgutil.IsUniqueViolation(err) {
			return Invitation{}, ErrDuplicateInvitation
		}
		return Invitation{}, fmt.Errorf("insert invitation: %w", err)
	}
	return inv, nil
}

func getInvitationByToken(ctx context.Context, db *sqlx.DB, tokenHash string) (Invitation, error) {
	var inv Invitation
	err := db.GetContext(ctx, &inv,
		`SELECT `+invCols+` FROM invitations WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Invitation{}, ErrInvitationNotFound
		}
		return Invitation{}, fmt.Errorf("get invitation by token: %w", err)
	}
	return inv, nil
}

func getInvitationByID(ctx context.Context, db *sqlx.DB, id string) (Invitation, error) {
	var inv Invitation
	err := db.GetContext(ctx, &inv,
		`SELECT `+invCols+` FROM invitations WHERE id = $1`,
		id,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Invitation{}, ErrInvitationNotFound
		}
		return Invitation{}, fmt.Errorf("get invitation by id: %w", err)
	}
	return inv, nil
}

func listPending(ctx context.Context, db *sqlx.DB, workspaceID string) ([]Invitation, error) {
	var invs []Invitation
	err := db.SelectContext(ctx, &invs,
		`SELECT `+invCols+` FROM invitations
		 WHERE workspace_id = $1 AND status = 'pending'
		 ORDER BY created_at DESC`,
		workspaceID,
	)
	if err != nil {
		return nil, fmt.Errorf("list pending invitations: %w", err)
	}
	return invs, nil
}

func revokeInvitation(ctx context.Context, db *sqlx.DB, id string) error {
	res, err := db.ExecContext(ctx,
		`UPDATE invitations SET status = 'revoked' WHERE id = $1 AND status = 'pending'`,
		id,
	)
	if err != nil {
		return fmt.Errorf("revoke invitation: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrInvitationNotFound
	}
	return nil
}

func acceptInvitation(ctx context.Context, db *sqlx.DB, tokenHash string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE invitations SET status = 'accepted', accepted_at = NOW() WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("accept invitation: %w", err)
	}
	return nil
}

func resendInvitation(ctx context.Context, db *sqlx.DB, id, newTokenHash string, newExpiresAt time.Time) error {
	res, err := db.ExecContext(ctx,
		`UPDATE invitations SET token_hash = $2, expires_at = $3 WHERE id = $1 AND status = 'pending'`,
		id, newTokenHash, newExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("resend invitation: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrInvitationNotFound
	}
	return nil
}

func isWorkspaceMemberByEmail(ctx context.Context, db *sqlx.DB, workspaceID, email string) (bool, error) {
	var exists bool
	err := db.GetContext(ctx, &exists,
		`SELECT EXISTS(
			SELECT 1 FROM workspace_members wm
			JOIN app_users u ON u.id = wm.user_id
			WHERE wm.workspace_id = $1 AND u.email = $2 AND wm.archived_at IS NULL
		)`,
		workspaceID, email,
	)
	if err != nil {
		return false, fmt.Errorf("check member by email: %w", err)
	}
	return exists, nil
}
