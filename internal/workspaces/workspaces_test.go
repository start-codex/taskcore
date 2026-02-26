package workspaces

import (
	"context"
	"testing"
)

func TestCreateWorkspaceParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		p       CreateWorkspaceParams
		wantErr bool
	}{
		{
			name:    "valid",
			p:       CreateWorkspaceParams{Name: "Acme Corp", Slug: "acme-corp"},
			wantErr: false,
		},
		{
			name:    "slug exactly 2 chars",
			p:       CreateWorkspaceParams{Name: "AB", Slug: "ab"},
			wantErr: false,
		},
		{
			name:    "slug with digits",
			p:       CreateWorkspaceParams{Name: "Team 42", Slug: "team42"},
			wantErr: false,
		},
		{
			name:    "missing name",
			p:       CreateWorkspaceParams{Name: "", Slug: "acme"},
			wantErr: true,
		},
		{
			name:    "slug too short",
			p:       CreateWorkspaceParams{Name: "A", Slug: "a"},
			wantErr: true,
		},
		{
			name:    "slug with uppercase",
			p:       CreateWorkspaceParams{Name: "Acme", Slug: "Acme"},
			wantErr: true,
		},
		{
			name:    "slug starts with hyphen",
			p:       CreateWorkspaceParams{Name: "Acme", Slug: "-acme"},
			wantErr: true,
		},
		{
			name:    "slug with spaces",
			p:       CreateWorkspaceParams{Name: "Acme", Slug: "acme corp"},
			wantErr: true,
		},
		{
			name:    "empty slug",
			p:       CreateWorkspaceParams{Name: "Acme", Slug: ""},
			wantErr: true,
		},
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

func TestCreateWorkspace_NilDB(t *testing.T) {
	_, err := CreateWorkspace(context.Background(), nil, CreateWorkspaceParams{Name: "Acme", Slug: "acme"})
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("CreateWorkspace() error = %v, want %q", err, "db is required")
	}
}

func TestGetWorkspace_NilDB(t *testing.T) {
	_, err := GetWorkspace(context.Background(), nil, "some-id")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("GetWorkspace() error = %v, want %q", err, "db is required")
	}
}

func TestGetWorkspace_EmptyID(t *testing.T) {
	_, err := GetWorkspace(context.Background(), nil, "")
	if err == nil {
		t.Fatal("GetWorkspace() with empty id should return error")
	}
}

func TestGetWorkspaceBySlug_NilDB(t *testing.T) {
	_, err := GetWorkspaceBySlug(context.Background(), nil, "acme")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("GetWorkspaceBySlug() error = %v, want %q", err, "db is required")
	}
}

func TestArchiveWorkspace_NilDB(t *testing.T) {
	err := ArchiveWorkspace(context.Background(), nil, "some-id")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("ArchiveWorkspace() error = %v, want %q", err, "db is required")
	}
}
