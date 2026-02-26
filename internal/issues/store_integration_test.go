package issues

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var testDSN = os.Getenv("MINI_JIRA_TEST_DSN")

func TestMoveIssue(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB, projectSeed) (MoveIssueParams, func(*testing.T))
		wantErr error
	}{
		{
			name: "within same status",
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (MoveIssueParams, func(*testing.T)) {
				a := insertIssue(t, db, seed, issueSeed{number: 1, title: "A", statusID: seed.statusTodoID, statusPosition: 0})
				b := insertIssue(t, db, seed, issueSeed{number: 2, title: "B", statusID: seed.statusTodoID, statusPosition: 1})
				c := insertIssue(t, db, seed, issueSeed{number: 3, title: "C", statusID: seed.statusTodoID, statusPosition: 2})
				p := MoveIssueParams{ProjectID: seed.projectID, IssueID: c, TargetStatusID: seed.statusTodoID, TargetPosition: 0}
				return p, func(t *testing.T) {
					assertOrder(t,
						fetchStatusOrder(t, db, seed.projectID, seed.statusTodoID),
						[]orderedIssue{{ID: c, Pos: 0}, {ID: a, Pos: 1}, {ID: b, Pos: 2}},
					)
				}
			},
		},
		{
			name: "across statuses",
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (MoveIssueParams, func(*testing.T)) {
				a := insertIssue(t, db, seed, issueSeed{number: 1, title: "A", statusID: seed.statusTodoID, statusPosition: 0})
				b := insertIssue(t, db, seed, issueSeed{number: 2, title: "B", statusID: seed.statusTodoID, statusPosition: 1})
				d := insertIssue(t, db, seed, issueSeed{number: 3, title: "D", statusID: seed.statusDoingID, statusPosition: 0})
				e := insertIssue(t, db, seed, issueSeed{number: 4, title: "E", statusID: seed.statusDoingID, statusPosition: 1})
				p := MoveIssueParams{ProjectID: seed.projectID, IssueID: b, TargetStatusID: seed.statusDoingID, TargetPosition: 1}
				return p, func(t *testing.T) {
					assertOrder(t,
						fetchStatusOrder(t, db, seed.projectID, seed.statusTodoID),
						[]orderedIssue{{ID: a, Pos: 0}},
					)
					assertOrder(t,
						fetchStatusOrder(t, db, seed.projectID, seed.statusDoingID),
						[]orderedIssue{{ID: d, Pos: 0}, {ID: b, Pos: 1}, {ID: e, Pos: 2}},
					)
				}
			},
		},
		{
			name: "no-op same position",
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (MoveIssueParams, func(*testing.T)) {
				a := insertIssue(t, db, seed, issueSeed{number: 1, title: "A", statusID: seed.statusTodoID, statusPosition: 0})
				b := insertIssue(t, db, seed, issueSeed{number: 2, title: "B", statusID: seed.statusTodoID, statusPosition: 1})
				p := MoveIssueParams{ProjectID: seed.projectID, IssueID: a, TargetStatusID: seed.statusTodoID, TargetPosition: 0}
				return p, func(t *testing.T) {
					assertOrder(t,
						fetchStatusOrder(t, db, seed.projectID, seed.statusTodoID),
						[]orderedIssue{{ID: a, Pos: 0}, {ID: b, Pos: 1}},
					)
				}
			},
		},
		{
			name:    "issue not found",
			wantErr: ErrIssueNotFound,
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (MoveIssueParams, func(*testing.T)) {
				p := MoveIssueParams{
					ProjectID:      seed.projectID,
					IssueID:        "00000000-0000-0000-0000-000000000000",
					TargetStatusID: seed.statusTodoID,
					TargetPosition: 0,
				}
				return p, nil
			},
		},
		{
			name: "clamp position beyond max",
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (MoveIssueParams, func(*testing.T)) {
				a := insertIssue(t, db, seed, issueSeed{number: 1, title: "A", statusID: seed.statusTodoID, statusPosition: 0})
				b := insertIssue(t, db, seed, issueSeed{number: 2, title: "B", statusID: seed.statusTodoID, statusPosition: 1})
				c := insertIssue(t, db, seed, issueSeed{number: 3, title: "C", statusID: seed.statusTodoID, statusPosition: 2})
				p := MoveIssueParams{ProjectID: seed.projectID, IssueID: a, TargetStatusID: seed.statusTodoID, TargetPosition: 999}
				return p, func(t *testing.T) {
					assertOrder(t,
						fetchStatusOrder(t, db, seed.projectID, seed.statusTodoID),
						[]orderedIssue{{ID: b, Pos: 0}, {ID: c, Pos: 1}, {ID: a, Pos: 2}},
					)
				}
			},
		},
		{
			name: "move to beginning of another status",
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (MoveIssueParams, func(*testing.T)) {
				a := insertIssue(t, db, seed, issueSeed{number: 1, title: "A", statusID: seed.statusTodoID, statusPosition: 0})
				d := insertIssue(t, db, seed, issueSeed{number: 2, title: "D", statusID: seed.statusDoingID, statusPosition: 0})
				e := insertIssue(t, db, seed, issueSeed{number: 3, title: "E", statusID: seed.statusDoingID, statusPosition: 1})
				p := MoveIssueParams{ProjectID: seed.projectID, IssueID: a, TargetStatusID: seed.statusDoingID, TargetPosition: 0}
				return p, func(t *testing.T) {
					assertOrder(t,
						fetchStatusOrder(t, db, seed.projectID, seed.statusDoingID),
						[]orderedIssue{{ID: a, Pos: 0}, {ID: d, Pos: 1}, {ID: e, Pos: 2}},
					)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed := seedProject(t, db)
			p, check := tt.arrange(t, db, seed)
			err := MoveIssue(context.Background(), db, p)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("MoveIssue() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestMoveIssue_ConcurrentMoves(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	seed := seedProject(t, db)

	const workers = 8
	issueIDs := make([]string, 0, workers)
	for i := range workers {
		issueIDs = append(issueIDs, insertIssue(t, db, seed, issueSeed{
			number:         i + 1,
			title:          "I",
			statusID:       seed.statusTodoID,
			statusPosition: i,
		}))
	}

	start := make(chan struct{})
	errCh := make(chan error, workers)
	var wg sync.WaitGroup
	for _, id := range issueIDs {
		wg.Add(1)
		go func(issueID string) {
			defer wg.Done()
			<-start
			errCh <- MoveIssue(context.Background(), db, MoveIssueParams{
				ProjectID:      seed.projectID,
				IssueID:        issueID,
				TargetStatusID: seed.statusDoingID,
				TargetPosition: 0,
			})
		}(id)
	}

	close(start)
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent MoveIssue returned error: %v", err)
		}
	}

	gotTodo := fetchStatusOrder(t, db, seed.projectID, seed.statusTodoID)
	if len(gotTodo) != 0 {
		t.Fatalf("expected todo status to be empty, got %d issues", len(gotTodo))
	}

	gotDoing := fetchStatusOrder(t, db, seed.projectID, seed.statusDoingID)
	if len(gotDoing) != workers {
		t.Fatalf("expected %d issues in doing status, got %d", workers, len(gotDoing))
	}
	assertContiguousPositions(t, gotDoing)
	assertContainsSameIDs(t, gotDoing, issueIDs)
}

func TestMoveIssue_ConcurrentMixedMoves(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	seed := seedProject(t, db)

	todoIDs := []string{
		insertIssue(t, db, seed, issueSeed{number: 1, title: "T1", statusID: seed.statusTodoID, statusPosition: 0}),
		insertIssue(t, db, seed, issueSeed{number: 2, title: "T2", statusID: seed.statusTodoID, statusPosition: 1}),
		insertIssue(t, db, seed, issueSeed{number: 3, title: "T3", statusID: seed.statusTodoID, statusPosition: 2}),
		insertIssue(t, db, seed, issueSeed{number: 4, title: "T4", statusID: seed.statusTodoID, statusPosition: 3}),
		insertIssue(t, db, seed, issueSeed{number: 5, title: "T5", statusID: seed.statusTodoID, statusPosition: 4}),
	}
	doingIDs := []string{
		insertIssue(t, db, seed, issueSeed{number: 6, title: "D1", statusID: seed.statusDoingID, statusPosition: 0}),
		insertIssue(t, db, seed, issueSeed{number: 7, title: "D2", statusID: seed.statusDoingID, statusPosition: 1}),
		insertIssue(t, db, seed, issueSeed{number: 8, title: "D3", statusID: seed.statusDoingID, statusPosition: 2}),
	}

	type moveCase struct {
		issueID   string
		statusID  string
		targetPos int
	}
	moves := []moveCase{
		{issueID: todoIDs[0], statusID: seed.statusDoingID, targetPos: 0},
		{issueID: todoIDs[1], statusID: seed.statusDoingID, targetPos: 1},
		{issueID: doingIDs[2], statusID: seed.statusDoingID, targetPos: 0},
		{issueID: todoIDs[4], statusID: seed.statusTodoID, targetPos: 0},
		{issueID: doingIDs[0], statusID: seed.statusTodoID, targetPos: 2},
	}

	start := make(chan struct{})
	errCh := make(chan error, len(moves))
	var wg sync.WaitGroup
	for _, m := range moves {
		wg.Add(1)
		go func(mc moveCase) {
			defer wg.Done()
			<-start
			errCh <- MoveIssue(context.Background(), db, MoveIssueParams{
				ProjectID:      seed.projectID,
				IssueID:        mc.issueID,
				TargetStatusID: mc.statusID,
				TargetPosition: mc.targetPos,
			})
		}(m)
	}

	close(start)
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent mixed MoveIssue returned error: %v", err)
		}
	}

	gotTodo := fetchStatusOrder(t, db, seed.projectID, seed.statusTodoID)
	gotDoing := fetchStatusOrder(t, db, seed.projectID, seed.statusDoingID)

	assertContiguousPositions(t, gotTodo)
	assertContiguousPositions(t, gotDoing)

	allExpected := append(append([]string{}, todoIDs...), doingIDs...)
	allGot := append(append([]orderedIssue{}, gotTodo...), gotDoing...)
	assertContainsSameIDs(t, allGot, allExpected)
}

type projectSeed struct {
	workspaceID   string
	reporterID    string
	projectID     string
	statusTodoID  string
	statusDoingID string
	issueTypeID   string
}

type issueSeed struct {
	number         int
	title          string
	statusID       string
	statusPosition int
}

type orderedIssue struct {
	ID  string `db:"id"`
	Pos int    `db:"status_position"`
}

func openTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	if testDSN == "" {
		t.Skip("MINI_JIRA_TEST_DSN is not set; skipping PostgreSQL integration test")
	}

	db, err := sqlx.Connect("postgres", testDSN)
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func ensureSchema(t *testing.T, db *sqlx.DB) {
	t.Helper()

	requiredTables := []string{"workspaces", "app_users", "projects", "statuses", "issues"}

	existing := 0
	for _, table := range requiredTables {
		var exists *string
		if err := db.Get(&exists, `SELECT to_regclass('public.`+table+`')::text`); err != nil {
			t.Fatalf("check table %s exists: %v", table, err)
		}
		if exists != nil && *exists != "" {
			existing++
		}
	}

	if existing == len(requiredTables) {
		return
	}
	if existing > 0 {
		t.Fatalf("partial schema detected: found %d/%d required tables", existing, len(requiredTables))
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
	sqlBytes, err := os.ReadFile(filepath.Join(root, "migrations", "0001_init.up.sql"))
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	if _, err := db.Exec(string(sqlBytes)); err != nil {
		t.Fatalf("apply migration: %v", err)
	}
}

func seedProject(t *testing.T, db *sqlx.DB) projectSeed {
	t.Helper()

	ctx := context.Background()
	out := projectSeed{}

	if err := db.GetContext(ctx, &out.workspaceID, `INSERT INTO workspaces (name, slug) VALUES ('ws', gen_random_uuid()::text) RETURNING id`); err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	if err := db.GetContext(ctx, &out.reporterID, `INSERT INTO app_users (email, name) VALUES (gen_random_uuid()::text || '@test.local', 'Reporter') RETURNING id`); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if err := db.GetContext(ctx, &out.projectID, `INSERT INTO projects (workspace_id, name, key, description) VALUES ($1, 'Project', upper(substr(replace(gen_random_uuid()::text,'-',''),1,3)), '') RETURNING id`, out.workspaceID); err != nil {
		t.Fatalf("insert project: %v", err)
	}
	if err := db.GetContext(ctx, &out.issueTypeID, `INSERT INTO issue_types (project_id, name, level) VALUES ($1, 'Task', 1) RETURNING id`, out.projectID); err != nil {
		t.Fatalf("insert issue_type: %v", err)
	}
	if err := db.GetContext(ctx, &out.statusTodoID, `INSERT INTO statuses (project_id, name, category, position) VALUES ($1, 'Por hacer', 'todo', 0) RETURNING id`, out.projectID); err != nil {
		t.Fatalf("insert todo status: %v", err)
	}
	if err := db.GetContext(ctx, &out.statusDoingID, `INSERT INTO statuses (project_id, name, category, position) VALUES ($1, 'En curso', 'doing', 1) RETURNING id`, out.projectID); err != nil {
		t.Fatalf("insert doing status: %v", err)
	}

	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), `DELETE FROM workspaces WHERE id = $1`, out.workspaceID); err != nil {
			t.Fatalf("cleanup workspace: %v", err)
		}
		if _, err := db.ExecContext(context.Background(), `DELETE FROM app_users WHERE id = $1`, out.reporterID); err != nil {
			t.Fatalf("cleanup user: %v", err)
		}
	})

	return out
}

func insertIssue(t *testing.T, db *sqlx.DB, seed projectSeed, in issueSeed) string {
	t.Helper()
	var id string
	err := db.Get(&id, `
		INSERT INTO issues (
			project_id, number, issue_type_id, status_id,
			title, description, priority, reporter_id, status_position
		) VALUES ($1, $2, $3, $4, $5, '', 'medium', $6, $7)
		RETURNING id
	`, seed.projectID, in.number, seed.issueTypeID, in.statusID, in.title, seed.reporterID, in.statusPosition)
	if err != nil {
		t.Fatalf("insert issue: %v", err)
	}
	return id
}

func fetchStatusOrder(t *testing.T, db *sqlx.DB, projectID, statusID string) []orderedIssue {
	t.Helper()
	var out []orderedIssue
	err := db.Select(&out, `
		SELECT id, status_position
		FROM issues
		WHERE project_id = $1 AND status_id = $2 AND archived_at IS NULL
		ORDER BY status_position ASC
	`, projectID, statusID)
	if err != nil {
		t.Fatalf("fetch status order: %v", err)
	}
	return out
}

func assertOrder(t *testing.T, got, want []orderedIssue) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("order length: got=%d want=%d", len(got), len(want))
	}
	for i := range got {
		if got[i].ID != want[i].ID || got[i].Pos != want[i].Pos {
			t.Fatalf("row[%d]: got=(%s,%d) want=(%s,%d)", i, got[i].ID, got[i].Pos, want[i].ID, want[i].Pos)
		}
	}
}

func assertContiguousPositions(t *testing.T, got []orderedIssue) {
	t.Helper()
	for i := range got {
		if got[i].Pos != i {
			t.Fatalf("positions not contiguous at idx=%d, got=%d", i, got[i].Pos)
		}
	}
}

func assertContainsSameIDs(t *testing.T, got []orderedIssue, wantIDs []string) {
	t.Helper()
	if len(got) != len(wantIDs) {
		t.Fatalf("id set size: got=%d want=%d", len(got), len(wantIDs))
	}
	wantSet := make(map[string]struct{}, len(wantIDs))
	for _, id := range wantIDs {
		wantSet[id] = struct{}{}
	}
	for _, row := range got {
		if _, ok := wantSet[row.ID]; !ok {
			t.Fatalf("unexpected issue id: %s", row.ID)
		}
		delete(wantSet, row.ID)
	}
	if len(wantSet) > 0 {
		t.Fatalf("missing %d issue ids after move", len(wantSet))
	}
}

func TestCreateIssue(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB, projectSeed) (CreateIssueParams, func(*testing.T))
		wantErr error
	}{
		{
			name: "creates issue with auto number and position",
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (CreateIssueParams, func(*testing.T)) {
				p := CreateIssueParams{
					ProjectID: seed.projectID, IssueTypeID: seed.issueTypeID,
					StatusID: seed.statusTodoID, Title: "First issue",
					ReporterID: seed.reporterID, Priority: "medium",
				}
				return p, func(t *testing.T) {}
			},
		},
		{
			name: "numbers increment per project",
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (CreateIssueParams, func(*testing.T)) {
				p := CreateIssueParams{
					ProjectID: seed.projectID, IssueTypeID: seed.issueTypeID,
					StatusID: seed.statusTodoID, Title: "Issue",
					ReporterID: seed.reporterID, Priority: "medium",
				}
				first, err := CreateIssue(context.Background(), db, p)
				if err != nil {
					t.Fatalf("create first: %v", err)
				}
				return p, func(t *testing.T) {
					second, err := CreateIssue(context.Background(), db, p)
					if err != nil {
						t.Fatalf("create second: %v", err)
					}
					if second.Number != first.Number+1 {
						t.Fatalf("number: got %d, want %d", second.Number, first.Number+1)
					}
					if second.StatusPosition != first.StatusPosition+1 {
						t.Fatalf("status_position: got %d, want %d", second.StatusPosition, first.StatusPosition+1)
					}
				}
			},
		},
		{
			name: "empty priority defaults to medium",
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (CreateIssueParams, func(*testing.T)) {
				p := CreateIssueParams{
					ProjectID: seed.projectID, IssueTypeID: seed.issueTypeID,
					StatusID: seed.statusTodoID, Title: "Issue",
					ReporterID: seed.reporterID,
				}
				return p, func(t *testing.T) {}
			},
		},
		{
			name: "creates issue with optional fields",
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (CreateIssueParams, func(*testing.T)) {
				due := time.Now().Add(48 * time.Hour)
				p := CreateIssueParams{
					ProjectID: seed.projectID, IssueTypeID: seed.issueTypeID,
					StatusID: seed.statusTodoID, Title: "Issue with extras",
					Description: "details", ReporterID: seed.reporterID,
					AssigneeID: seed.reporterID, Priority: "high",
					DueDate: &due,
				}
				return p, func(t *testing.T) {}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed := seedProject(t, db)
			p, check := tt.arrange(t, db, seed)
			got, err := CreateIssue(context.Background(), db, p)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateIssue() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil {
				if got.ID == "" {
					t.Fatal("expected non-empty id")
				}
				if got.Number <= 0 {
					t.Fatalf("number: got %d, want > 0", got.Number)
				}
				if got.StatusPosition < 0 {
					t.Fatalf("status_position: got %d, want >= 0", got.StatusPosition)
				}
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestGetIssue(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB, projectSeed) (projectID, issueID string, check func(*testing.T))
		wantErr error
	}{
		{
			name: "returns existing issue",
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (string, string, func(*testing.T)) {
				id := insertIssue(t, db, seed, issueSeed{number: 1, title: "A", statusID: seed.statusTodoID, statusPosition: 0})
				return seed.projectID, id, func(t *testing.T) {}
			},
		},
		{
			name:    "not found",
			wantErr: ErrIssueNotFound,
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (string, string, func(*testing.T)) {
				return seed.projectID, "00000000-0000-0000-0000-000000000000", nil
			},
		},
		{
			name:    "wrong project returns not found",
			wantErr: ErrIssueNotFound,
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (string, string, func(*testing.T)) {
				id := insertIssue(t, db, seed, issueSeed{number: 1, title: "A", statusID: seed.statusTodoID, statusPosition: 0})
				return "00000000-0000-0000-0000-000000000000", id, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed := seedProject(t, db)
			projID, issueID, check := tt.arrange(t, db, seed)
			got, err := GetIssue(context.Background(), db, projID, issueID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetIssue() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil && got.ID != issueID {
				t.Fatalf("id: got %q, want %q", got.ID, issueID)
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestListIssues(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB, projectSeed) (ListIssuesParams, func(*testing.T, []Issue))
		wantErr error
	}{
		{
			name: "returns active issues",
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (ListIssuesParams, func(*testing.T, []Issue)) {
				a := insertIssue(t, db, seed, issueSeed{number: 1, title: "A", statusID: seed.statusTodoID, statusPosition: 0})
				insertIssue(t, db, seed, issueSeed{number: 2, title: "B", statusID: seed.statusTodoID, statusPosition: 1})
				if err := ArchiveIssue(context.Background(), db, seed.projectID, a); err != nil {
					t.Fatalf("archive issue: %v", err)
				}
				return ListIssuesParams{ProjectID: seed.projectID}, func(t *testing.T, got []Issue) {
					if len(got) != 1 {
						t.Fatalf("len: got %d, want 1", len(got))
					}
				}
			},
		},
		{
			name: "filter by status",
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (ListIssuesParams, func(*testing.T, []Issue)) {
				insertIssue(t, db, seed, issueSeed{number: 1, title: "Todo", statusID: seed.statusTodoID, statusPosition: 0})
				insertIssue(t, db, seed, issueSeed{number: 2, title: "Doing", statusID: seed.statusDoingID, statusPosition: 0})
				return ListIssuesParams{ProjectID: seed.projectID, StatusID: seed.statusTodoID}, func(t *testing.T, got []Issue) {
					if len(got) != 1 {
						t.Fatalf("len: got %d, want 1", len(got))
					}
					if got[0].StatusID != seed.statusTodoID {
						t.Fatalf("status_id: got %q, want %q", got[0].StatusID, seed.statusTodoID)
					}
				}
			},
		},
		{
			name: "empty project returns empty slice",
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (ListIssuesParams, func(*testing.T, []Issue)) {
				return ListIssuesParams{ProjectID: seed.projectID}, func(t *testing.T, got []Issue) {
					if len(got) != 0 {
						t.Fatalf("len: got %d, want 0", len(got))
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed := seedProject(t, db)
			p, check := tt.arrange(t, db, seed)
			got, err := ListIssues(context.Background(), db, p)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ListIssues() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if check != nil {
				check(t, got)
			}
		})
	}
}

func TestUpdateIssue(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB, projectSeed) (UpdateIssueParams, func(*testing.T))
		wantErr error
	}{
		{
			name: "updates title and priority",
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (UpdateIssueParams, func(*testing.T)) {
				id := insertIssue(t, db, seed, issueSeed{number: 1, title: "Old", statusID: seed.statusTodoID, statusPosition: 0})
				p := UpdateIssueParams{IssueID: id, ProjectID: seed.projectID, Title: "New", Priority: "high"}
				return p, func(t *testing.T) {}
			},
		},
		{
			name: "clears assignee when nil",
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (UpdateIssueParams, func(*testing.T)) {
				id := insertIssue(t, db, seed, issueSeed{number: 1, title: "A", statusID: seed.statusTodoID, statusPosition: 0})
				p := UpdateIssueParams{IssueID: id, ProjectID: seed.projectID, Title: "A", Priority: "medium", AssigneeID: nil}
				return p, func(t *testing.T) {}
			},
		},
		{
			name:    "not found",
			wantErr: ErrIssueNotFound,
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (UpdateIssueParams, func(*testing.T)) {
				p := UpdateIssueParams{
					IssueID:   "00000000-0000-0000-0000-000000000000",
					ProjectID: seed.projectID, Title: "X", Priority: "low",
				}
				return p, nil
			},
		},
		{
			name:    "archived issue returns not found",
			wantErr: ErrIssueNotFound,
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (UpdateIssueParams, func(*testing.T)) {
				id := insertIssue(t, db, seed, issueSeed{number: 1, title: "A", statusID: seed.statusTodoID, statusPosition: 0})
				if err := ArchiveIssue(context.Background(), db, seed.projectID, id); err != nil {
					t.Fatalf("archive: %v", err)
				}
				return UpdateIssueParams{IssueID: id, ProjectID: seed.projectID, Title: "X", Priority: "low"}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed := seedProject(t, db)
			p, check := tt.arrange(t, db, seed)
			got, err := UpdateIssue(context.Background(), db, p)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("UpdateIssue() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil {
				if got.Title != p.Title {
					t.Fatalf("title: got %q, want %q", got.Title, p.Title)
				}
				if got.Priority != p.Priority {
					t.Fatalf("priority: got %q, want %q", got.Priority, p.Priority)
				}
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestArchiveIssue(t *testing.T) {
	db := openTestDB(t)
	ensureSchema(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB, projectSeed) (projectID, issueID string)
		wantErr error
	}{
		{
			name: "archives active issue",
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (string, string) {
				id := insertIssue(t, db, seed, issueSeed{number: 1, title: "A", statusID: seed.statusTodoID, statusPosition: 0})
				return seed.projectID, id
			},
		},
		{
			name:    "not found",
			wantErr: ErrIssueNotFound,
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (string, string) {
				return seed.projectID, "00000000-0000-0000-0000-000000000000"
			},
		},
		{
			name:    "already archived",
			wantErr: ErrIssueNotFound,
			arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (string, string) {
				id := insertIssue(t, db, seed, issueSeed{number: 1, title: "A", statusID: seed.statusTodoID, statusPosition: 0})
				if err := ArchiveIssue(context.Background(), db, seed.projectID, id); err != nil {
					t.Fatalf("first archive: %v", err)
				}
				return seed.projectID, id
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed := seedProject(t, db)
			projID, issueID := tt.arrange(t, db, seed)
			err := ArchiveIssue(context.Background(), db, projID, issueID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ArchiveIssue() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil {
				got, err := GetIssue(context.Background(), db, projID, issueID)
				if err != nil {
					t.Fatalf("get archived issue: %v", err)
				}
				if got.ArchivedAt == nil {
					t.Fatal("expected archived_at to be set")
				}
			}
		})
	}
}
