# Roadmap

Tookly is a workflow platform. Software delivery is the first deeply defined methodology. Broader cross-industry adoption comes through templates and methodology-aware configuration. Documentation-led planning is a major differentiator.

Each phase has a status label:
- `[shipped]` — exists in the codebase today.
- `[in progress]` — partially implemented; actively being developed.
- `[planned]` — not yet started; described at the intent and scope level.

---

## Phase 0 — Foundation `[shipped]`

The core platform infrastructure and first working end-to-end flow.

**Backend**
- Go 1.26 monolith with domain-per-package structure under `internal/`.
- PostgreSQL, explicit SQL, `sqlx`; no ORM.
- Domain packages: `users`, `workspaces`, `projects`, `boards`, `statuses`, `issuetypes`, `issues`.
- Shared HTTP utilities in `internal/respond/`.
- Per-package `RegisterRoutes` pattern; handlers are private.
- `MoveIssue` domain function and API endpoint — backend and API layer complete.
- Docker Compose local dev setup.
- Makefile commands: `db-up`, `db-down`, `db-reset`, `db-clean`, `db-backup`, `db-size`, `db-shell`.

**Frontend**
- SvelteKit 2 + Svelte 5 + Tailwind 4.
- Local shadcn-style components built on top of Bits UI primitives (no external shadcn-svelte library).
- Paraglide i18n: English and Spanish.
- Frontend embedded and served by the Go binary.

**Project templates**
- Kanban template: preconfigures `To Do`, `In Progress`, `Done` statuses + one default board.
- Scrum template: preconfigures `Backlog`, `To Do`, `In Progress`, `In Review`, `Done` statuses + one default board.
- Board column/status mapping is **not auto-created** by the template.

**What is not yet shipped in this phase**
- UI drag-and-drop: `MoveIssue` backend is ready; the frontend is not wired yet.

---

## Phase 1 — MVP hardening `[shipped]`

Close the gap between what the backend supports and what the UI delivers. Deliver a fully usable, secure baseline.

**Authentication and authorization** (6-PR delivery plan):
- PR 1 `[shipped]` — Session storage foundation: `internal/sessions` with `Create`, `Validate`, `Delete`. SHA-256 hashed tokens. Archived-user rejection. Migration `0003_create_sessions`.
- PR 2 `[shipped]` — Auth middleware (`withAuth`) and endpoints: `POST /auth/login` (cookie), `GET /auth/me`, `POST /auth/logout`. Self-only `GET /users/{userID}`.
- PR 3 `[shipped]` — Membership authorization: `internal/authz` with context helpers and workspace membership enforcement on all API routes (read and write). Consolidated `internal/authctx` into `internal/authz`.
- PR 4 `[shipped]` — Remove client-controlled identity: drop `owner_id`, `reporter_id`, `user_id` from API contracts; derive from session.
- PR 5 `[shipped]` — Admin/owner authorization for workspace and project administration.
- PR 6 `[shipped]` — Frontend session migration (replace auth localStorage with `/auth/me`) and workflow configuration admin enforcement.

**Remaining UI work** (4-PR delivery plan):
- PR 21 `[shipped]` — Add `due_date` to issue update and create API contracts.
- PR 22 `[shipped]` — Issue detail page: view and edit title, description, priority, assignee, due date.
- PR 23 `[shipped]` — Board drag-and-drop: move issues between columns with optimistic updates.
- PR 24 `[shipped]` — Basic board filters: client-side filtering by assignee, priority, and issue type.

---

## Phase 1.5 — Identity, onboarding, and instance admin `[in progress]`

Make self-hosted deployments operable beyond basic local login. This phase covers the missing identity and bootstrap capabilities that sit between MVP auth and broader platform workflows.

**Instance configuration, bootstrap, and email** (PRs #25–#28 `[shipped]`)
- `instance_config` key-value table for instance-level settings.
- `internal/instance` package: `GetConfig`, `SetConfig`, `IsInitialized`, `Bootstrap`.
- `POST /instance/bootstrap`: atomic first-install flow creates global admin.
- `GET /instance/status`: check initialization state.
- `is_instance_admin` field on users; `RequireInstanceAdmin` in authz.
- `POST /users` blocked before bootstrap, public after.

**Transactional email foundation**
- Instance-level SMTP configuration.
- Email templates and delivery pipeline for system emails.
- Delivery status, retry handling, and safe tokenized links for user-facing flows.

**Account lifecycle**
- Forgot-password / reset-password flow with expiring tokens.
- Change-password flow for authenticated users.
- Optional email verification for newly created accounts.

**Invitations**
- Invite user by email into the instance or a workspace.
- Pending invitation model with resend, revoke, and expiration.
- Invitation acceptance flow that creates or links the user account and assigns the intended role.

**Federated identity**
- OpenID Connect (OIDC) / SSO login with external identity providers.
- Account linking strategy for existing local users.
- Just-in-time provisioning policy and post-login membership mapping.

**Instance bootstrap**
- First-install flow to create the initial global administrator.
- Initial system configuration, including system-wide legal/terms text when the deployment requires it.
- Separation between instance-wide administration and workspace/project administration.

---

## Phase 2 — Software workflow depth `[planned]`

Software delivery is the first deeply modeled workflow in Tookly. This phase brings sprint-based and hierarchy-based planning.

Note: software is the first vertical, not the only one.

- **Issue hierarchy**: Epic → Story → Task → Subtask. Schema fields (`parent_issue_id`, `issue_type.level`) and basic DB integrity (same-project, level ordering, anti-cycle) already exist; domain-level validation in the API and the hierarchy UI are pending.
- **Backlog view**: list of issues not assigned to any sprint; drag issues into a sprint.
- **Sprint model**: create sprint, add issues from backlog, start sprint, close sprint.
- **Sprint planning board**: board scoped to a single active sprint.
- **Planning fields for software teams**: `start_date` to capture when work actually began, and `story_points` to capture relative effort for sprint planning and velocity.
- **Estimation model**: remains configurable by team and is deferred until the platform defines how time-based and effort-based estimates coexist.
- **Retrospective notes**: free-text notes attached to a closed sprint.
- **Issue key display**: `ENG-42` format in UI and API responses.

---

## Phase 3 — Documentation-led planning `[planned]`

Documentation is the **source of planning context**, not a passive wiki attached as an afterthought.

This phase establishes the core differentiator of Tookly: teams document their decisions, rules, and architecture, and from that documentation they **plan and refine** their backlog, roadmap, sprint scope, estimations, and reviews.

The initial implementation is **manual first**: links between documentation and work items are created explicitly by users. No automated generation.

**Documentation**
- Project documentation pages with a Markdown editor.
- Page hierarchy: parent and child pages within a project.
- Decision records: a structured page type for capturing architectural and business decisions.
- Decision records attached to projects, not floating as standalone documents.

**Linking docs to work items**
- Explicit link between a documentation page and one or more work items.
- Link visible from both sides: page shows linked issues; issue shows linked pages.
- No automated derivation — users create links intentionally.

**Planning grounded in documentation**
- Backlog planning and refinement informed by documented decisions and business rules.
- Sprint planning sessions reference documented context.
- Review artifacts linked back to the decisions they validate.
- Roadmap views aligned with architectural context documented in pages.

---

## Phase 4 — Cross-industry templates `[planned]`

Templates are reusable workflow presets. They are not isolated products — they are bundles of configuration that any project can start from.

Each template bundles:
- Statuses with categories and positions.
- Work-item types with names and hierarchy levels.
- Default board layout (columns mapped to statuses).
- Optional planning structure (e.g. sprint-aware or backlog-aware).
- Optional documentation structure (e.g. starter pages for decision records or runbooks).

**Initial template catalog**

| Template | Category |
|---|---|
| Software / Kanban | Engineering |
| Software / Scrum | Engineering |
| Bug Tracking | Engineering |
| General Task Tracking | Any |
| HR / Recruiting | HR |
| Legal / Document Management | Legal |
| Marketing / Campaign | Marketing |
| Support / IT Service | Operations |
| Sales Pipeline | Sales |
| Finance / Budget | Finance |
| Operations / Process | Operations |
| Design / UX | Design |

Templates can be selected at project creation time. Custom templates can be created and saved from an existing project.

---

## Phase 5 — Automation + reporting `[planned]`

Add intelligence and visibility on top of the workflow.

**Automations**
- Rule-based triggers: "when status changes to Done → notify assignee".
- Initial rules: status change, assignment change, due date approaching.
- Rules are configured per project.
- Email notifications for workflow events use the transactional email channel introduced in Phase 1.5; webhooks remain available for external systems.

**Reporting and metrics**
- Project dashboard: throughput, cycle time, open items by type and priority.
- Sprint burndown chart.
- Velocity tracking across sprints.

**Export and integrations**
- CSV and JSON export of issues per project.
- Webhook outbound notifications for external integrations.

**Assisted transformations (deferred)**
- Assisted derivation of backlog items or sprint scope from documentation content.
- Only introduced here, after the manual-first foundation of Phase 3 is established.

---

## Phase 6 — AI assistant and MCP `[planned]`

Workflow-oriented assistant that helps teams query, draft, structure and execute within Tookly. See [docs/06-ai-assistant.md](06-ai-assistant.md) for full details.

**Provider configuration**
- Provider-agnostic: configurable via API per workspace or instance.
- Supported providers: OpenAI/GPT, Anthropic/Claude, Google/Gemini, Ollama, or any compatible API.
- Configuration: endpoint URL, API key, model name.
- Without a configured provider, AI features are unavailable but Tookly works normally.

**Assistant / Copilot**
- Chat interface for querying workspace data (issues, boards, projects, members).
- Assisted drafting and structuring of documentation pages.
- Execution of Tookly operations under the authenticated user's session and permissions.
- No AI superuser — if the user lacks permissions, the operation fails as in UI.

**Proposals**
- Unified `Proposal` model for changes suggested by AI or by a human.
- Same shape: origin (human/ai), author, target entity, payload.
- Same execution flow and permission model regardless of origin.

**MCP integration**
- Connectors to read and act on external systems (Git, CI, messaging, etc.).
- Scoped by the authenticated user's session and permissions.
- Extensible connector model for self-hosted environments.

**Documentation**
- Project pages persisted as editable Markdown.
- Predefined templates/forms to structure initial content.
- Free editing after creation — no semantic graph or canonical types.
