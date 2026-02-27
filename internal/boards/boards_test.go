package boards

import (
	"context"
	"testing"
)

func TestCreateBoardParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  CreateBoardParams
		wantErr bool
	}{
		{
			name:    "valid kanban",
			params:  CreateBoardParams{ProjectID: "proj-1", Name: "Main Board", Type: "kanban"},
			wantErr: false,
		},
		{
			name:    "valid scrum",
			params:  CreateBoardParams{ProjectID: "proj-1", Name: "Sprint Board", Type: "scrum"},
			wantErr: false,
		},
		{
			name:    "missing project_id",
			params:  CreateBoardParams{ProjectID: "", Name: "Main Board", Type: "kanban"},
			wantErr: true,
		},
		{
			name:    "missing name",
			params:  CreateBoardParams{ProjectID: "proj-1", Name: "", Type: "kanban"},
			wantErr: true,
		},
		{
			name:    "invalid type",
			params:  CreateBoardParams{ProjectID: "proj-1", Name: "Main Board", Type: "other"},
			wantErr: true,
		},
		{
			name:    "empty type",
			params:  CreateBoardParams{ProjectID: "proj-1", Name: "Main Board", Type: ""},
			wantErr: true,
		},
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

func TestAddColumnParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  AddColumnParams
		wantErr bool
	}{
		{
			name:    "valid",
			params:  AddColumnParams{BoardID: "board-1", Name: "To Do"},
			wantErr: false,
		},
		{
			name:    "missing board_id",
			params:  AddColumnParams{BoardID: "", Name: "To Do"},
			wantErr: true,
		},
		{
			name:    "missing name",
			params:  AddColumnParams{BoardID: "board-1", Name: ""},
			wantErr: true,
		},
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

func TestCreateBoard_NilDB(t *testing.T) {
	_, err := CreateBoard(context.Background(), nil, CreateBoardParams{ProjectID: "p", Name: "B", Type: "kanban"})
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("CreateBoard() error = %v, want %q", err, "db is required")
	}
}

func TestGetBoard_NilDB(t *testing.T) {
	_, err := GetBoard(context.Background(), nil, "some-id")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("GetBoard() error = %v, want %q", err, "db is required")
	}
}

func TestListBoards_NilDB(t *testing.T) {
	_, err := ListBoards(context.Background(), nil, "proj-1")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("ListBoards() error = %v, want %q", err, "db is required")
	}
}

func TestArchiveBoard_NilDB(t *testing.T) {
	err := ArchiveBoard(context.Background(), nil, "some-id")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("ArchiveBoard() error = %v, want %q", err, "db is required")
	}
}

func TestAddColumn_NilDB(t *testing.T) {
	_, err := AddColumn(context.Background(), nil, AddColumnParams{BoardID: "b", Name: "Col"})
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("AddColumn() error = %v, want %q", err, "db is required")
	}
}

func TestListColumns_NilDB(t *testing.T) {
	_, err := ListColumns(context.Background(), nil, "board-1")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("ListColumns() error = %v, want %q", err, "db is required")
	}
}

func TestArchiveColumn_NilDB(t *testing.T) {
	err := ArchiveColumn(context.Background(), nil, "some-id")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("ArchiveColumn() error = %v, want %q", err, "db is required")
	}
}

func TestAssignStatus_NilDB(t *testing.T) {
	err := AssignStatus(context.Background(), nil, "col-1", "status-1")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("AssignStatus() error = %v, want %q", err, "db is required")
	}
}

func TestUnassignStatus_NilDB(t *testing.T) {
	err := UnassignStatus(context.Background(), nil, "col-1", "status-1")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("UnassignStatus() error = %v, want %q", err, "db is required")
	}
}
