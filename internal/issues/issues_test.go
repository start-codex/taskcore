package issues

import (
	"context"
	"testing"
	"time"
)

func TestMoveIssueParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		p       MoveIssueParams
		wantErr bool
	}{
		{
			name:    "valid params",
			p:       MoveIssueParams{ProjectID: "proj-1", IssueID: "issue-1", TargetPosition: 0},
			wantErr: false,
		},
		{
			name:    "missing project_id",
			p:       MoveIssueParams{ProjectID: "", IssueID: "issue-1", TargetPosition: 0},
			wantErr: true,
		},
		{
			name:    "missing issue_id",
			p:       MoveIssueParams{ProjectID: "proj-1", IssueID: "", TargetPosition: 0},
			wantErr: true,
		},
		{
			name:    "negative target_position",
			p:       MoveIssueParams{ProjectID: "proj-1", IssueID: "issue-1", TargetPosition: -1},
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

func TestMoveIssue_NilDB(t *testing.T) {
	err := MoveIssue(context.Background(), nil, MoveIssueParams{
		ProjectID:      "proj-1",
		IssueID:        "issue-1",
		TargetPosition: 0,
	})
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("MoveIssue() error = %v, want %q", err, "db is required")
	}
}

func TestCreateIssueParams_Validate(t *testing.T) {
	due := time.Now()
	valid := CreateIssueParams{
		ProjectID:   "p",
		IssueTypeID: "t",
		StatusID:    "s",
		Title:       "Fix bug",
		ReporterID:  "r",
		Priority:    "high",
	}

	tests := []struct {
		name    string
		p       CreateIssueParams
		wantErr bool
	}{
		{name: "valid", p: valid, wantErr: false},
		{name: "priority defaults to medium", p: func() CreateIssueParams { c := valid; c.Priority = ""; return c }(), wantErr: false},
		{name: "valid with due date", p: func() CreateIssueParams { c := valid; c.DueDate = &due; return c }(), wantErr: false},
		{name: "missing project_id", p: func() CreateIssueParams { c := valid; c.ProjectID = ""; return c }(), wantErr: true},
		{name: "missing issue_type_id", p: func() CreateIssueParams { c := valid; c.IssueTypeID = ""; return c }(), wantErr: true},
		{name: "missing status_id", p: func() CreateIssueParams { c := valid; c.StatusID = ""; return c }(), wantErr: true},
		{name: "missing title", p: func() CreateIssueParams { c := valid; c.Title = ""; return c }(), wantErr: true},
		{name: "missing reporter_id", p: func() CreateIssueParams { c := valid; c.ReporterID = ""; return c }(), wantErr: true},
		{name: "invalid priority", p: func() CreateIssueParams { c := valid; c.Priority = "urgent"; return c }(), wantErr: true},
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

func TestUpdateIssueParams_Validate(t *testing.T) {
	valid := UpdateIssueParams{
		IssueID:   "i",
		ProjectID: "p",
		Title:     "Fix bug",
		Priority:  "low",
	}

	tests := []struct {
		name    string
		p       UpdateIssueParams
		wantErr bool
	}{
		{name: "valid", p: valid, wantErr: false},
		{name: "missing issue_id", p: func() UpdateIssueParams { c := valid; c.IssueID = ""; return c }(), wantErr: true},
		{name: "missing project_id", p: func() UpdateIssueParams { c := valid; c.ProjectID = ""; return c }(), wantErr: true},
		{name: "missing title", p: func() UpdateIssueParams { c := valid; c.Title = ""; return c }(), wantErr: true},
		{name: "invalid priority", p: func() UpdateIssueParams { c := valid; c.Priority = "asap"; return c }(), wantErr: true},
		{name: "empty priority invalid", p: func() UpdateIssueParams { c := valid; c.Priority = ""; return c }(), wantErr: true},
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

func TestCreateIssue_NilDB(t *testing.T) {
	_, err := CreateIssue(context.Background(), nil, CreateIssueParams{
		ProjectID: "p", IssueTypeID: "t", StatusID: "s", Title: "T", ReporterID: "r", Priority: "medium",
	})
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("CreateIssue() error = %v, want %q", err, "db is required")
	}
}

func TestGetIssue_NilDB(t *testing.T) {
	_, err := GetIssue(context.Background(), nil, "p", "i")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("GetIssue() error = %v, want %q", err, "db is required")
	}
}

func TestListIssues_NilDB(t *testing.T) {
	_, err := ListIssues(context.Background(), nil, ListIssuesParams{ProjectID: "p"})
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("ListIssues() error = %v, want %q", err, "db is required")
	}
}

func TestUpdateIssue_NilDB(t *testing.T) {
	_, err := UpdateIssue(context.Background(), nil, UpdateIssueParams{
		IssueID: "i", ProjectID: "p", Title: "T", Priority: "medium",
	})
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("UpdateIssue() error = %v, want %q", err, "db is required")
	}
}

func TestArchiveIssue_NilDB(t *testing.T) {
	err := ArchiveIssue(context.Background(), nil, "p", "i")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("ArchiveIssue() error = %v, want %q", err, "db is required")
	}
}
