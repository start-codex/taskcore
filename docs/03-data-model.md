# Data Model

## Design principle

- An issue belongs to the project (`project_id`), not to a board.
- A status also belongs to the project (`project_id`).
- A board is a view and configuration layer over issues — it is not the owner of work.

---

## Current entities

### Workspace

- `id` (UUID, PK)
- `name`
- `slug`
- `created_at`
- `updated_at`
- `archived_at` (nullable)

### User

- `id` (UUID, PK)
- `email` (unique)
- `name`
- `password_hash` (Argon2id)
- `created_at`
- `updated_at`
- `archived_at` (nullable)

### WorkspaceMember

- `workspace_id` (FK)
- `user_id` (FK)
- `role` (`owner`, `admin`, `member`)
- `created_at`
- `updated_at`
- `archived_at` (nullable)

### Session

- `id` (TEXT, PK — SHA-256 hash of the raw token)
- `user_id` (FK → User)
- `created_at`
- `expires_at`
- `last_used_at` (nullable — set on creation, not refreshed per request in current implementation)

Indexes: `user_id`, `expires_at`.

### Project

- `id` (UUID, PK)
- `workspace_id` (FK)
- `name`
- `key` (e.g. `ENG`, `MKT`) — short uppercase identifier used in issue keys
- `description`
- `created_at`
- `updated_at`
- `archived_at` (nullable)

### ProjectMember

- `project_id` (FK)
- `user_id` (FK)
- `role` (`admin`, `member`, `viewer`)
- `created_at`
- `updated_at`
- `archived_at` (nullable)

Note: `project_members` exists in the schema and is populated during project member management, but is not currently used as an authorization source. Authorization is based on `workspace_members` only (Phase 1).

### ProjectIssueCounter

- `project_id` (PK)
- `last_number`
- `created_at`
- `updated_at`

Used to generate sequential issue numbers per project without race conditions. See implementation note below.

### Status

- `id` (UUID, PK)
- `project_id` (FK)
- `name` (e.g. `To Do`, `In Progress`, `Done`)
- `category` (`todo`, `doing`, `done`)
- `position`
- `created_at`
- `updated_at`
- `archived_at` (nullable)

**Note:** when a project is created with a Kanban or Scrum template, statuses are preconfigured automatically. Board column mapping is **not auto-created** by the template.

### IssueType

- `id` (UUID, PK)
- `project_id` (FK)
- `name` (e.g. `Epic`, `Story`, `Task`, `Subtask`, `Bug`)
- `icon`
- `level` (hierarchy depth: 0 = top-level, higher = deeper child)
- `created_at`
- `updated_at`
- `archived_at` (nullable)

### Board

- `id` (UUID, PK)
- `project_id` (FK)
- `name`
- `type` (`kanban`, `scrum`)
- `filter_query` (board-level issue filter)
- `created_at`
- `updated_at`
- `archived_at` (nullable)

### BoardColumn

- `id` (UUID, PK)
- `board_id` (FK)
- `name`
- `position`
- `created_at`
- `updated_at`
- `archived_at` (nullable)

### BoardColumnStatus

- `board_column_id` (FK)
- `status_id` (FK)
- `created_at`

A column can map to one or more statuses. Both the column and the status must belong to the same project.

### Issue

- `id` (UUID, PK)
- `project_id` (FK)
- `number` (sequential per project)
- `issue_type_id` (FK)
- `status_id` (FK)
- `parent_issue_id` (nullable, FK → Issue) — hierarchy field
- `title`
- `description`
- `priority` (`low`, `medium`, `high`, `critical`)
- `assignee_id` (nullable, FK → User)
- `reporter_id` (FK → User)
- `due_date` (nullable)
- `status_position` — sort order within a status column
- `created_at`
- `updated_at`
- `archived_at` (nullable)

Recommended public key in UI/API: `PROJECT_KEY-NUMBER` (e.g. `ENG-123`).

### IssueEvent (audit log)

- `id` (UUID, PK)
- `issue_id` (FK)
- `actor_id` (FK → User)
- `event_type` (`created`, `updated`, `moved`, `commented`)
- `payload_json`
- `created_at`

Note: the `issue_events` table exists in the schema but does not have domain functions or API endpoints yet. It is reserved for future audit trail and activity feed features.

---

## Integrity rules

These constraints are enforced at the database level:

- `Issue.issue_type_id` must belong to the same `project_id` as the issue.
- `Issue.status_id` must belong to the same `project_id` as the issue.
- `Issue.parent_issue_id` must belong to the same project.
- Hierarchy by level: a child issue must have a `level` greater than its parent.
- Anti-cycle: `parent_issue_id` chains must not form cycles.
- `BoardColumnStatus`: column and status must belong to the same project.
- `status_position` is unique per (`project_id`, `status_id`) for active issues (`archived_at IS NULL`).

Note: `parent_issue_id` and `issue_type.level` integrity rules exist in the database. Full hierarchy domain rules (cycle prevention beyond DB constraints, level validation in the API) and the corresponding UI are planned in Phase 2.

---

## Suggested indexes

- `Issue(project_id, status_id, status_position)`
- `Issue(assignee_id)`
- `Issue(parent_issue_id)`
- `Status(project_id, position)`
- `BoardColumn(board_id, position)`
- `Session(user_id)`
- `Session(expires_at)`

---

## Implementation notes

**SQL as source of truth.** Migrations live in `migrations/`. All access from Go uses `database/sql` + `sqlx` with private SQL functions per domain package. No ORM.

**Issue number generation.** Uses a per-project counter in `project_issue_counters`, not `MAX(number)+1`, to avoid race conditions under concurrent inserts:

```sql
INSERT INTO project_issue_counters (project_id, last_number)
VALUES ($1, 1)
ON CONFLICT (project_id)
DO UPDATE SET last_number = project_issue_counters.last_number + 1
RETURNING last_number;
```

The returned `last_number` is used as `issues.number` in the same transaction.

---

## Future model directions

These entities and capabilities are planned for future phases. Field-level design is intentionally deferred until implementation planning begins.

**Phase 1.5 — Identity, onboarding, and instance admin**
- Invitation records: pending invites with expiration, role assignment, and acceptance lifecycle.
- Password reset / recovery tokens: expiring tokens for account recovery flows.
- SMTP configuration: instance-level settings for transactional email delivery.
- Federated identity: OIDC provider configuration, account linking, JIT provisioning metadata.
- Instance configuration: global admin records, system-wide settings, bootstrap state.

**Phase 2 — Software workflow depth**
- **Sprint / SprintIssue** — sprint planning and execution; links issues to time-boxed iterations.
- **Issue.start_date** — actual work-start date for teams that track when implementation begins.
- **Issue.story_points** — relative-effort field for sprint planning and velocity.
- **Estimation model** — configurable team-level approach for time-, effort-, or point-based estimation; deferred until its semantics are defined.
- **Comment** — threaded comments on issues.
- **Attachment** — files attached to issues.
- **CustomField** — per-project extensible fields on issues.

**Phase 3 — Documentation-led planning**
- **ProjectPage** — documentation pages that belong to a project; Markdown content, page hierarchy (parent/child), and decision records.
- **Page–WorkItem link** — an explicit link record between a documentation page and a work item, enabling manual traceability between documented decisions and execution artifacts.

**Phase 4 — Cross-industry templates**
- **ProjectTemplate** — reusable workflow preset that bundles statuses, issue types, board layout, and optionally a documentation structure.

**Phase 5 — Automation + reporting**
- **Notification** — event-driven alerts for status changes, assignments, mentions.

**Phase 6 — AI assistant and MCP**
- Provider configuration records: endpoint, API key, model, scoped per workspace or instance.
- Proposal records: suggested changes from AI or human origin, with target entity and payload.
- Assistant interaction metadata for context and history.
