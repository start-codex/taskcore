package users

import (
	"context"
	"testing"
)

func TestCreateUserParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  CreateUserParams
		wantErr bool
	}{
		{name: "valid", params: CreateUserParams{Email: "alice@example.com", Name: "Alice", Password: "secret"}, wantErr: false},
		{name: "missing name", params: CreateUserParams{Email: "alice@example.com", Name: "", Password: "secret"}, wantErr: true},
		{name: "missing email", params: CreateUserParams{Email: "", Name: "Alice", Password: "secret"}, wantErr: true},
		{name: "email without @", params: CreateUserParams{Email: "notanemail", Name: "Alice", Password: "secret"}, wantErr: true},
		{name: "missing password", params: CreateUserParams{Email: "alice@example.com", Name: "Alice", Password: ""}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateUser_NilDB(t *testing.T) {
	_, err := CreateUser(context.Background(), nil, CreateUserParams{Email: "a@b.com", Name: "A", Password: "p"})
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

func TestAuthenticateUser_NilDB(t *testing.T) {
	_, err := AuthenticateUser(context.Background(), nil, "a@b.com", "pass")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("AuthenticateUser() error = %v, want %q", err, "db is required")
	}
}
