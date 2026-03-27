# Technical Architecture

## Technical goals

- Single binary, easy to deploy.
- Low memory and CPU footprint.
- Maintainable and extensible by domain.
- Modern UX without a heavy SPA.

---

## Current stack

**Backend**

- Go 1.26
- HTTP router: `net/http` stdlib (Go 1.22+ supports method routing and path params natively)
- Database: PostgreSQL
- SQL layer: `database/sql` + `sqlx`
- Migrations: `golang-migrate`

**Frontend**

- SvelteKit 2 + Svelte 5
- Vite
- Tailwind CSS 4
- Local shadcn-style components built on top of Bits UI primitives (no external shadcn-svelte library dependency)
- Paraglide for i18n (EN + ES)

The frontend is compiled and embedded into the Go binary at build time. The Go server serves the SvelteKit app and the API from a single process.

---

## Application style

- Modular monolith (no microservices in MVP).
- API REST mounted under `/api` (no version prefix in current routes).
- Jira-like model: issue is owned by project and status; board is a view (filter + columns) over issues.

---

## Actual directory structure

```
cmd/
  server/
    main.go           # entrypoint: configures DB, router, starts server
    middleware.go     # withRequestID, withLogger, withRecover (private to cmd)

internal/
  boards/
    boards.go         # types, errors, public API
    handler.go        # HTTP handlers + RegisterRoutes
    store.go          # private SQL persistence
    boards_test.go
    store_integration_test.go

  issues/
    issues.go
    handler.go
    store.go
    issues_test.go
    store_integration_test.go

  issuetypes/
    issuetypes.go
    handler.go
    store.go

  projects/
    projects.go
    handler.go
    store.go
    projects_test.go
    store_integration_test.go

  statuses/
    statuses.go
    handler.go
    store.go

  users/
    users.go
    handler.go
    store.go
    password.go
    users_test.go
    store_integration_test.go

  workspaces/
    workspaces.go
    handler.go
    store.go
    workspaces_test.go
    store_integration_test.go

  respond/
    respond.go        # shared HTTP utilities (respond.JSON, respond.Error, respond.Decode)
                      # does not import any domain package

migrations/
  *.up.sql
  *.down.sql

front/
  src/
  package.json
  ...
```

---

## Handler pattern

Handlers are thin by design: parse the request, call the domain function, write the response.

Each domain package exposes a single `RegisterRoutes(mux *http.ServeMux, db *sqlx.DB)` function. `cmd/server/main.go` calls them in order. All handler functions are private (`handleCreate`, not `HandleCreate`).

Domain packages register paths **without** the `/api/` prefix. `cmd/server/main.go` mounts them on a sub-mux with `http.StripPrefix("/api", api)`, so the full public URL becomes `/api/projects/...`.

```go
func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB) {
    mux.HandleFunc("POST /projects/{projectID}/issues", handleCreateIssue(db))
    mux.HandleFunc("GET /projects/{projectID}/issues/{issueID}", handleGetIssue(db))
}

func handleCreateIssue(db *sqlx.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        projectID := r.PathValue("projectID")
        var p issues.CreateIssueParams
        if err := respond.Decode(r, &p); err != nil {
            respond.Error(w, http.StatusBadRequest, err)
            return
        }
        p.ProjectID = projectID
        issue, err := issues.CreateIssue(r.Context(), db, p)
        if err != nil {
            fail(w, err)
            return
        }
        respond.JSON(w, http.StatusCreated, issue)
    }
}
```

Each `handler.go` defines a local `fail(w, err)` function that maps domain sentinel errors to HTTP codes. Unknown errors → 500.

---

## Authentication

### Current state

- `POST /api/auth/login` exists and returns a user payload on success.
- The frontend currently stores the returned user object in **local storage**.
- Session storage layer shipped: `internal/sessions` with `Create`, `Validate`, `Delete`. Tokens are SHA-256 hashed before storage; archived users are rejected on validation.
- Cookie-based session middleware, auth endpoints (`/auth/me`, `/auth/logout`), and per-handler membership enforcement are **not yet implemented** (PRs 2–6 of Phase 1).

### Target

- Server-side sessions with secure cookies: `HttpOnly`, `Secure`, `SameSite=Strict`.
- Session middleware validates the cookie on every request and injects the authenticated user into the context.
- Workspace and project membership enforced per handler using the context user.
- Logout endpoint invalidates the server-side session.

---

## Design rules

- One package per domain, not per technical layer.
- Do not create `internal/domain`, `internal/app`, or `internal/store` globals.
- Do not use OOP subdirectory patterns inside a domain (`repository/`, `service/`, `manager/`).
- Domain and persistence coexist in the same package:
  - `<domain>.go` — types, errors, validation, public API.
  - `store.go` — private SQL functions and persistence details.
- Prefer free functions with explicit dependencies (e.g. `func MoveIssue(ctx, db, p)`).
- Explicit SQL, testable against real PostgreSQL in integration tests.
- Interfaces only when there is a concrete need (not preventive).

---

## Observability

- Structured JSON logs.
- Request ID per request (middleware in `cmd/server/middleware.go`).
- Minimal metrics target: latency, error rate, throughput.

---

## Security

- Input validation in the handler (before calling the domain function) and again inside the domain (`Validate()` as second line of defense).
- CSRF protection for state-changing endpoints — planned as part of full session-based authentication.
- Cookies with `HttpOnly`, `Secure`, `SameSite` — planned, pending auth completion.
- Access control per workspace and project — planned, pending auth completion.

---

## Target architecture — documentation-led planning

In the documentation-led planning phase (Phase 3), the architecture extends to support project documentation alongside execution:

- Documentation pages belong to projects, stored in the same database.
- Pages and work items share explicit link records — no implicit coupling.
- Planning workflows (backlog refinement, sprint planning, reviews) reference documented context directly.
- The initial implementation is **user-driven and manual**: users create links, not the system.
- Automated inference from documentation is deferred to Phase 5 or later.

This means no separate "wiki service" or external documentation product. Documentation is a first-class domain in the same monolith.

---

## Planned additions

- Session middleware and server-side session store.
- Workspace and project membership enforcement in all handlers.
- Sprint and backlog endpoints (`internal/sprints/`).
- Project templates API (extend `internal/projects/` or add `internal/templates/`).
- Notification domain (`internal/notifications/`).
- Project documentation pages domain (`internal/pages/`).
