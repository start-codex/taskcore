# Product Scope

## Problem

Current project management tools tend to be:

- Too complex for small teams that just need to track work.
- Focused exclusively on software, or so generic they offer no real methodology support.
- Expensive to adapt — core workflow features locked behind enterprise plans.
- Split across two products: project tracking in one tool, documentation in another. Teams end up with decisions in Confluence and work in Jira, but the two never truly talk to each other.

## Vision

Taskcore is a self-hostable, open-source workflow platform where any team can manage ideas, projects, and work using methodologies that fit their area.

Software delivery is the first fully modeled workflow. The long-term goal is broader: a dental clinic, a law firm, a marketing agency, or an engineering team should all be able to track what needs to happen, what is happening, and what is done — without paying for enterprise plans or stitching together two separate tools.

## Core concept

```
workspace → project → work items
                        ↑
              boards are methodology-aware views over that work
```

Issues (the current implementation term) belong to the project, not to the board. A board is a view and configuration layer — it defines how work is visualized and in what order columns appear. The same work item can appear on multiple boards if needed.

The broader target-vision term is **work item**. As Taskcore grows to support cross-industry templates, "issue" will remain valid for software contexts while "work item" covers the general case.

## What makes Taskcore different

- **Documentation as the source of planning** — not a passive wiki attached as an afterthought, but the starting point for backlog creation, roadmap definition, sprint scope, and review artifacts.
- **No Jira/Confluence split** — project pages, decision records, and execution artifacts live in the same project, not in separate products.
- **Templates for any industry** — workflow presets for engineering, HR, legal, marketing, finance, design, and more.
- **No paywalled core features** — the workflows that matter are available to everyone.
- **Self-hosted and open source** — run it on your own infrastructure, contribute to it, or fork it.

## Primary users

Software teams:

- Developer
- Tech lead
- Product manager

## Secondary users

Any team that manages staged work:

- HR and recruiting
- Legal
- Marketing and content
- Operations
- Finance
- Design and UX

---

## Current state

What exists in the codebase today:

- Workspaces: create and manage team workspaces.
- Projects: create projects with a Kanban or Scrum template; templates preconfigure statuses and one default board.
- Boards: view for visualizing work by status columns.
- Statuses: per-project, with categories `todo`, `doing`, `done`.
- Issue types: per-project, configurable (e.g. Task, Bug, Epic, Story).
- Issues CRUD: create, read, list, update, archive.
- `MoveIssue`: backend and API layer complete; UI drag-and-drop is not yet wired.
- Internationalization: English and Spanish via Paraglide.

**Not yet shipped:**

- Drag-and-drop board UI (backend ready; frontend not wired).
- Cookie-based auth and authorization enforcement (`POST /api/auth/login` exists; session storage layer shipped with hashed tokens and archived-user rejection; middleware, cookie handling, and per-handler enforcement are not yet shipped).
- Wiki/documentation pages.
- Full methodology customization (custom fields per issue, configurable transition rules, richer workflow rules).

---

## Target vision

### Software workflow vision

In software teams, the biggest gap is not the task board — it is the broken connection between what the team knows and what the team builds.

Decisions get made in meetings and disappear. Business rules live in someone's head. Architecture choices are buried in old pull requests. Sprint planning happens without context.

Taskcore's target vision for software: **documentation is the source of planning**, not a passive wiki.

Teams document:

- Business rules
- Architectural decisions
- Application boundaries and context
- Implementation constraints and assumptions

From that documentation, teams **plan and refine**:

- Backlog creation and refinement grounded in documented rules and decisions
- Roadmap aligned with architectural context
- Sprint scope informed by documented implementation constraints
- Estimation sessions with shared context
- Planning sessions tied to documented decisions
- Review artifacts linked back to the decisions they validate

This relationship is **manual first**: documentation and execution artifacts are explicitly linked, with workflow support, before any form of automated generation is introduced.

---

## Business rules

- Every issue belongs to exactly one project.
- Every issue has exactly one status at any point in time.
- Statuses belong to the project, not to the board. A board maps its columns to one or more statuses.
- Board column order is defined in the board configuration.
- In the MVP, transitions between any two statuses are allowed (no restricted transition rules yet).
- Issue types and statuses are per-project. An issue's type and status must belong to the same project as the issue.
- Issue hierarchy: `parent_issue_id` exists in the schema; full hierarchy enforcement and UI are planned (Phase 2).
- Issue numbering is sequential per project, generated via `project_issue_counters` to avoid race conditions.

---

## Future extensions

- Issue hierarchy: Epic → Story → Task → Subtask, with domain rule enforcement.
- Sprint model: create, populate, start, and close sprints; backlog view.
- Cross-industry workflow templates.
- Custom fields per project.
- Rule-based automations (e.g. notify on status change).
- Audit trail and event log.
- Project documentation pages with decision records.
- Explicit links between documentation pages and work items.
