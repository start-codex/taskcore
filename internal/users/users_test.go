package users

import (
	"context"
	"testing"
)

func TestCreateUserParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		p       CreateUserParams
		wantErr bool
	}{
		{name: "valid", p: CreateUserParams{Email: "alice@example.com", Name: "Alice"}, wantErr: false},
		{name: "missing name", p: CreateUserParams{Email: "alice@example.com", Name: ""}, wantErr: true},
		{name: "missing email", p: CreateUserParams{Email: "", Name: "Alice"}, wantErr: true},
		{name: "email without @", p: CreateUserParams{Email: "notanemail", Name: "Alice"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.p.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddWorkspaceMemberParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		p       AddWorkspaceMemberParams
		wantErr bool
	}{
		{name: "owner", p: AddWorkspaceMemberParams{WorkspaceID: "w", UserID: "u", Role: "owner"}, wantErr: false},
		{name: "admin", p: AddWorkspaceMemberParams{WorkspaceID: "w", UserID: "u", Role: "admin"}, wantErr: false},
		{name: "member", p: AddWorkspaceMemberParams{WorkspaceID: "w", UserID: "u", Role: "member"}, wantErr: false},
		{name: "missing workspace_id", p: AddWorkspaceMemberParams{WorkspaceID: "", UserID: "u", Role: "member"}, wantErr: true},
		{name: "missing user_id", p: AddWorkspaceMemberParams{WorkspaceID: "w", UserID: "", Role: "member"}, wantErr: true},
		{name: "invalid role", p: AddWorkspaceMemberParams{WorkspaceID: "w", UserID: "u", Role: "viewer"}, wantErr: true},
		{name: "empty role", p: AddWorkspaceMemberParams{WorkspaceID: "w", UserID: "u", Role: ""}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.p.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddProjectMemberParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		p       AddProjectMemberParams
		wantErr bool
	}{
		{name: "admin", p: AddProjectMemberParams{ProjectID: "p", UserID: "u", Role: "admin"}, wantErr: false},
		{name: "member", p: AddProjectMemberParams{ProjectID: "p", UserID: "u", Role: "member"}, wantErr: false},
		{name: "viewer", p: AddProjectMemberParams{ProjectID: "p", UserID: "u", Role: "viewer"}, wantErr: false},
		{name: "missing project_id", p: AddProjectMemberParams{ProjectID: "", UserID: "u", Role: "member"}, wantErr: true},
		{name: "missing user_id", p: AddProjectMemberParams{ProjectID: "p", UserID: "", Role: "member"}, wantErr: true},
		{name: "invalid role", p: AddProjectMemberParams{ProjectID: "p", UserID: "u", Role: "owner"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.p.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateUser_NilDB(t *testing.T) {
	_, err := CreateUser(context.Background(), nil, CreateUserParams{Email: "a@b.com", Name: "A"})
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("CreateUser() error = %v, want %q", err, "db is required")
	}
}

func TestGetUser_NilDB(t *testing.T) {
	_, err := GetUser(context.Background(), nil, "id")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("GetUser() error = %v, want %q", err, "db is required")
	}
}

func TestGetUserByEmail_NilDB(t *testing.T) {
	_, err := GetUserByEmail(context.Background(), nil, "a@b.com")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("GetUserByEmail() error = %v, want %q", err, "db is required")
	}
}

func TestArchiveUser_NilDB(t *testing.T) {
	err := ArchiveUser(context.Background(), nil, "id")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("ArchiveUser() error = %v, want %q", err, "db is required")
	}
}

func TestAddWorkspaceMember_NilDB(t *testing.T) {
	_, err := AddWorkspaceMember(context.Background(), nil, AddWorkspaceMemberParams{WorkspaceID: "w", UserID: "u", Role: "member"})
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("AddWorkspaceMember() error = %v, want %q", err, "db is required")
	}
}

func TestRemoveWorkspaceMember_NilDB(t *testing.T) {
	err := RemoveWorkspaceMember(context.Background(), nil, "w", "u")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("RemoveWorkspaceMember() error = %v, want %q", err, "db is required")
	}
}

func TestAddProjectMember_NilDB(t *testing.T) {
	_, err := AddProjectMember(context.Background(), nil, AddProjectMemberParams{ProjectID: "p", UserID: "u", Role: "member"})
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("AddProjectMember() error = %v, want %q", err, "db is required")
	}
}

func TestRemoveProjectMember_NilDB(t *testing.T) {
	err := RemoveProjectMember(context.Background(), nil, "p", "u")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("RemoveProjectMember() error = %v, want %q", err, "db is required")
	}
}
