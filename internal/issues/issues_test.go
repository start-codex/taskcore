package issues

import (
	"context"
	"testing"
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
