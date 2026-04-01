// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package invitations

import (
	"context"
	"errors"
	"testing"

	_ "github.com/lib/pq"
	"github.com/start-codex/tookly/internal/sessions"
	"github.com/start-codex/tookly/internal/testpg"
)

func TestCreateInvitation_Success(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	inviter := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	ctx := context.Background()

	rawToken, inv, err := CreateInvitation(ctx, db, CreateInvitationParams{
		WorkspaceID: wsID,
		Email:       "invited@test.local",
		Role:        "member",
		InvitedBy:   inviter,
	})
	if err != nil {
		t.Fatalf("CreateInvitation error = %v", err)
	}
	if rawToken == "" {
		t.Fatal("rawToken is empty")
	}
	if inv.Status != "pending" {
		t.Fatalf("status = %q, want pending", inv.Status)
	}
	if inv.Email != "invited@test.local" {
		t.Fatalf("email = %q, want invited@test.local", inv.Email)
	}
}

func TestCreateInvitation_Duplicate(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	inviter := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	ctx := context.Background()

	params := CreateInvitationParams{
		WorkspaceID: wsID,
		Email:       "dup@test.local",
		Role:        "member",
		InvitedBy:   inviter,
	}
	_, _, err := CreateInvitation(ctx, db, params)
	if err != nil {
		t.Fatalf("first create error = %v", err)
	}

	_, _, err = CreateInvitation(ctx, db, params)
	if !errors.Is(err, ErrDuplicateInvitation) {
		t.Fatalf("duplicate error = %v, want ErrDuplicateInvitation", err)
	}
}

func TestGetInvitation_Valid(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	inviter := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	ctx := context.Background()

	rawToken, _, err := CreateInvitation(ctx, db, CreateInvitationParams{
		WorkspaceID: wsID,
		Email:       "get@test.local",
		Role:        "member",
		InvitedBy:   inviter,
	})
	if err != nil {
		t.Fatalf("create error = %v", err)
	}

	inv, err := GetInvitation(ctx, db, rawToken)
	if err != nil {
		t.Fatalf("GetInvitation error = %v", err)
	}
	if inv.Email != "get@test.local" {
		t.Fatalf("email = %q", inv.Email)
	}
}

func TestGetInvitation_Expired(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	inviter := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	ctx := context.Background()

	rawToken, _, _ := CreateInvitation(ctx, db, CreateInvitationParams{
		WorkspaceID: wsID,
		Email:       "expired@test.local",
		Role:        "member",
		InvitedBy:   inviter,
	})

	// Expire the token
	hash := sessions.HashToken(rawToken)
	db.ExecContext(ctx, `UPDATE invitations SET expires_at = NOW() - INTERVAL '1 day' WHERE token_hash = $1`, hash)

	_, err := GetInvitation(ctx, db, rawToken)
	if !errors.Is(err, ErrInvitationExpired) {
		t.Fatalf("error = %v, want ErrInvitationExpired", err)
	}
}

func TestGetInvitation_Revoked(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	inviter := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	ctx := context.Background()

	rawToken, inv, _ := CreateInvitation(ctx, db, CreateInvitationParams{
		WorkspaceID: wsID,
		Email:       "revoked@test.local",
		Role:        "member",
		InvitedBy:   inviter,
	})

	_ = Revoke(ctx, db, inv.ID)

	_, err := GetInvitation(ctx, db, rawToken)
	if !errors.Is(err, ErrInvitationRevoked) {
		t.Fatalf("error = %v, want ErrInvitationRevoked", err)
	}
}

func TestAcceptInvitation_Success(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	inviter := testpg.SeedUser(t, db)
	accepter := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	ctx := context.Background()

	// Make inviter a member so workspace exists properly
	db.ExecContext(ctx, `INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'owner')`, wsID, inviter)

	rawToken, _, _ := CreateInvitation(ctx, db, CreateInvitationParams{
		WorkspaceID: wsID,
		Email:       "accept@test.local",
		Role:        "member",
		InvitedBy:   inviter,
	})

	err := AcceptInvitation(ctx, db, rawToken, accepter)
	if err != nil {
		t.Fatalf("AcceptInvitation error = %v", err)
	}

	// Verify member was added
	var role string
	db.GetContext(ctx, &role,
		`SELECT role FROM workspace_members WHERE workspace_id = $1 AND user_id = $2 AND archived_at IS NULL`,
		wsID, accepter)
	if role != "member" {
		t.Fatalf("role = %q, want member", role)
	}
}

func TestListPending(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	inviter := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	ctx := context.Background()

	CreateInvitation(ctx, db, CreateInvitationParams{
		WorkspaceID: wsID, Email: "list1@test.local", Role: "member", InvitedBy: inviter,
	})
	CreateInvitation(ctx, db, CreateInvitationParams{
		WorkspaceID: wsID, Email: "list2@test.local", Role: "admin", InvitedBy: inviter,
	})

	invs, err := ListPending(ctx, db, wsID)
	if err != nil {
		t.Fatalf("ListPending error = %v", err)
	}
	if len(invs) < 2 {
		t.Fatalf("len = %d, want >= 2", len(invs))
	}
}

func TestResend(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	inviter := testpg.SeedUser(t, db)
	wsID := testpg.SeedWorkspace(t, db)
	ctx := context.Background()

	_, inv, _ := CreateInvitation(ctx, db, CreateInvitationParams{
		WorkspaceID: wsID, Email: "resend@test.local", Role: "member", InvitedBy: inviter,
	})

	newToken, err := Resend(ctx, db, inv.ID)
	if err != nil {
		t.Fatalf("Resend error = %v", err)
	}
	if newToken == "" {
		t.Fatal("new token is empty")
	}

	// New token should be valid
	_, err = GetInvitation(ctx, db, newToken)
	if err != nil {
		t.Fatalf("GetInvitation with new token error = %v", err)
	}
}
