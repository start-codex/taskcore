package projects

import (
	"context"
	"testing"
)

func TestCreateProjectParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		p       CreateProjectParams
		wantErr bool
	}{
		{
			name:    "valid",
			p:       CreateProjectParams{WorkspaceID: "ws-1", Name: "Engineering", Key: "ENG"},
			wantErr: false,
		},
		{
			name:    "key exactly 2 chars",
			p:       CreateProjectParams{WorkspaceID: "ws-1", Name: "Engineering", Key: "AB"},
			wantErr: false,
		},
		{
			name:    "key exactly 10 chars",
			p:       CreateProjectParams{WorkspaceID: "ws-1", Name: "Engineering", Key: "ABCDEFGHIJ"},
			wantErr: false,
		},
		{
			name:    "missing workspace_id",
			p:       CreateProjectParams{WorkspaceID: "", Name: "Engineering", Key: "ENG"},
			wantErr: true,
		},
		{
			name:    "missing name",
			p:       CreateProjectParams{WorkspaceID: "ws-1", Name: "", Key: "ENG"},
			wantErr: true,
		},
		{
			name:    "key too short",
			p:       CreateProjectParams{WorkspaceID: "ws-1", Name: "Engineering", Key: "E"},
			wantErr: true,
		},
		{
			name:    "key too long",
			p:       CreateProjectParams{WorkspaceID: "ws-1", Name: "Engineering", Key: "ABCDEFGHIJK"},
			wantErr: true,
		},
		{
			name:    "key lowercase",
			p:       CreateProjectParams{WorkspaceID: "ws-1", Name: "Engineering", Key: "eng"},
			wantErr: true,
		},
		{
			name:    "key with digits",
			p:       CreateProjectParams{WorkspaceID: "ws-1", Name: "Engineering", Key: "ENG1"},
			wantErr: true,
		},
		{
			name:    "key with spaces",
			p:       CreateProjectParams{WorkspaceID: "ws-1", Name: "Engineering", Key: "EN G"},
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

func TestCreateProject_NilDB(t *testing.T) {
	_, err := CreateProject(context.Background(), nil, CreateProjectParams{
		WorkspaceID: "ws-1",
		Name:        "Engineering",
		Key:         "ENG",
	})
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("CreateProject() error = %v, want %q", err, "db is required")
	}
}

func TestGetProject_NilDB(t *testing.T) {
	_, err := GetProject(context.Background(), nil, "some-id")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("GetProject() error = %v, want %q", err, "db is required")
	}
}

func TestGetProject_EmptyID(t *testing.T) {
	_, err := GetProject(context.Background(), nil, "")
	if err == nil {
		t.Fatal("GetProject() with empty id should return error")
	}
}

func TestListProjects_NilDB(t *testing.T) {
	_, err := ListProjects(context.Background(), nil, "ws-1")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("ListProjects() error = %v, want %q", err, "db is required")
	}
}

func TestArchiveProject_NilDB(t *testing.T) {
	err := ArchiveProject(context.Background(), nil, "some-id")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("ArchiveProject() error = %v, want %q", err, "db is required")
	}
}
