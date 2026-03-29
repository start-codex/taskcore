package main

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/boards"
	"github.com/start-codex/taskcode/internal/issues"
	"github.com/start-codex/taskcode/internal/issuetypes"
	"github.com/start-codex/taskcode/internal/projects"
	"github.com/start-codex/taskcode/internal/statuses"
	"github.com/start-codex/taskcode/internal/users"
	"github.com/start-codex/taskcode/internal/workspaces"
)

// newAPIHandler builds the API sub-mux with auth middleware and all domain routes.
func newAPIHandler(db *sqlx.DB) http.Handler {
	api := http.NewServeMux()
	users.RegisterRoutes(api, db)
	workspaces.RegisterRoutes(api, db)
	projects.RegisterRoutes(api, db)
	statuses.RegisterRoutes(api, db)
	issuetypes.RegisterRoutes(api, db)
	boards.RegisterRoutes(api, db)
	issues.RegisterRoutes(api, db)
	return withAuth(api, db)
}
