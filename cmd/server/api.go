// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package main

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/tookly/internal/boards"
	"github.com/start-codex/tookly/internal/emailverification"
	"github.com/start-codex/tookly/internal/instance"
	"github.com/start-codex/tookly/internal/invitations"
	"github.com/start-codex/tookly/internal/issues"
	"github.com/start-codex/tookly/internal/passwordreset"
	"github.com/start-codex/tookly/internal/issuetypes"
	"github.com/start-codex/tookly/internal/projects"
	"github.com/start-codex/tookly/internal/statuses"
	"github.com/start-codex/tookly/internal/users"
	"github.com/start-codex/tookly/internal/workspaces"
)

// newAPIHandler builds the API sub-mux with auth middleware and all domain routes.
func newAPIHandler(db *sqlx.DB) http.Handler {
	api := http.NewServeMux()
	instance.RegisterRoutes(api, db)
	users.RegisterRoutes(api, db)
	passwordreset.RegisterRoutes(api, db)
	emailverification.RegisterRoutes(api, db)
	workspaces.RegisterRoutes(api, db)
	invitations.RegisterRoutes(api, db)
	projects.RegisterRoutes(api, db)
	statuses.RegisterRoutes(api, db)
	issuetypes.RegisterRoutes(api, db)
	boards.RegisterRoutes(api, db)
	issues.RegisterRoutes(api, db)
	return withAuth(api, db)
}
