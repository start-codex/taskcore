// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package main

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/trazawork/internal/boards"
	"github.com/start-codex/trazawork/internal/issues"
	"github.com/start-codex/trazawork/internal/issuetypes"
	"github.com/start-codex/trazawork/internal/projects"
	"github.com/start-codex/trazawork/internal/statuses"
	"github.com/start-codex/trazawork/internal/users"
	"github.com/start-codex/trazawork/internal/workspaces"
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
