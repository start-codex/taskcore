<p align="center">
  <img src="front/src/lib/assets/tookly-logo.svg" alt="Tookly logo" width="180" />
</p>

# Tookly

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white" alt="Go 1.26" />
  <img src="https://img.shields.io/badge/SvelteKit-2-FF3E00?logo=svelte&logoColor=white" alt="SvelteKit 2" />
  <img src="https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql&logoColor=white" alt="PostgreSQL 16" />
  <img src="https://img.shields.io/badge/License-BSL%201.1-0F172A" alt="License BSL 1.1" />
</p>

Tookly is a source-available workflow platform for teams. Capture ideas, plan work, track execution, and follow through — regardless of your industry or methodology.

Software delivery is the first deeply defined workflow: documentation, decisions, and architecture feed the backlog, roadmap, and sprints. The long-term goal is broader: any team, any domain.

## Goals

- Software-first, not software-only — built for engineering teams today, extensible to any team tomorrow.
- Documentation and planning in one place — decisions, business rules, and architecture drive the backlog, not a separate wiki.
- Self-hostable — no vendor lock-in, no paywalled core features.
- Simple to start, flexible to grow — works for a two-person team and scales to an organization.

## Why Tookly

| | Tookly | Jira | Linear | Asana | Monday |
|---|---|---|---|---|---|
| Self-hosted | Yes | Paid / complex | No | No | No |
| Docs + planning together | Yes (target) | Separate (Confluence) | No | No | No |
| Core features | All open | Many behind paywall | Free tier limited | Free tier limited | Free tier limited |
| Cross-industry templates | Planned | Software only | Software only | General | General |
| Source-available | Yes (BSL 1.1) | No | No | No | No |

Tookly is not a clone of any existing tool — it is a workflow platform with its own identity.

## Current product baseline

What is shipped today:

- Workspaces and projects.
- Kanban and Scrum project templates (preconfigure statuses and one default board).
- Boards, statuses, issue types, issues CRUD.
- Board drag-and-drop: move issues between columns and reorder within columns.
- Issue detail page: view and edit title, description, priority, assignee, due date.
- Basic board filters: client-side filtering by assignee, priority, and issue type.
- Instance bootstrap: first-install setup wizard creates the initial global admin.
- Workspace invitations: admin invite page, accept page with registration, login redirect with `next`.
- Forgot/reset password pages with email token (1h TTL, one-time use, session invalidation).
- Change password with session invalidation on other devices.
- Instance admin settings page for SMTP configuration.
- SMTP email delivery with Mailpit for local development.
- Local email/password authentication with server-side sessions.
- Workspace and project membership enforcement with admin/owner roles.
- Internationalization: English and Spanish.

Not in the current baseline yet: SMTP delivery, password reset, user invitations, SSO/OIDC, or a first-install global admin bootstrap flow.

See [docs/05-roadmap.md](docs/05-roadmap.md) for what is in progress and planned.

## Tech stack

- **Backend:** Go 1.26 — monolith, `database/sql` + `sqlx`, explicit SQL queries.
- **Frontend:** SvelteKit 2 + Svelte 5 + Tailwind 4 + local shadcn-style components built on top of Bits UI primitives + Paraglide i18n.
- **Database:** PostgreSQL.
- **Deployment:** Single Docker image, `docker compose` for local dev.

## Roadmap

See [docs/05-roadmap.md](docs/05-roadmap.md) for the full phased roadmap.

Summary:
- **Phase 0 — Foundation** `[shipped]` — core backend, domains, templates, i18n.
- **Phase 1 — MVP hardening** `[shipped]` — full auth, membership enforcement, board UI, issue detail, board filters.
- **Phase 1.5 — Identity, onboarding, and instance admin** `[planned]` — SMTP, password reset, invitations, SSO/OIDC, first-install bootstrap.
- **Phase 2 — Software workflow depth** `[planned]` — issue hierarchy, sprints, backlog, planning board.
- **Phase 3 — Documentation-led planning** `[planned]` — project pages, decision records, doc↔work item links.
- **Phase 4 — Cross-industry templates** `[planned]` — workflow presets for HR, legal, marketing, sales, and more.
- **Phase 5 — Automation + reporting** `[planned]` — automations, metrics, burndown, velocity tracking.
- **Phase 6 — AI assistant and MCP** `[planned]` — provider-agnostic copilot, proposals, MCP connectors, assisted documentation.

## Getting started

```bash
git clone https://github.com/start-codex/tookly
cd tookly
docker compose up --build
```

App runs at `http://localhost:8080`.

Create your first user:

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email": "you@example.com", "name": "Your Name", "password": "yourpassword"}'
```

## Database management

The database uses a local bind mount at `.docker/postgres/`.

```bash
make db-up          # Start only the database
make db-down        # Stop containers
make db-reset       # Reset database (removes all data)
make db-clean       # Remove database folder only
make db-backup      # Create a backup of the database folder
make db-size        # Show database folder size
make db-shell       # Open PostgreSQL shell
```

## API

All responses follow the envelope format:

```json
{ "status": 200, "data": {} }
{ "status": 400, "error": "description" }
```

## Documentation

- Product scope: [docs/01-product-scope.md](docs/01-product-scope.md)
- Architecture: [docs/02-architecture.md](docs/02-architecture.md)
- Data model: [docs/03-data-model.md](docs/03-data-model.md)
- Go conventions: [docs/04-go-conventions.md](docs/04-go-conventions.md)
- Roadmap: [docs/05-roadmap.md](docs/05-roadmap.md)
- AI assistant: [docs/06-ai-assistant.md](docs/06-ai-assistant.md)
- Changelog: [CHANGELOG.md](CHANGELOG.md)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). All contributions require signing our [Contributor License Agreement](CLA.md).

## License

Business Source License 1.1 — see [LICENSE](LICENSE) for details.

- **Self-hosting**: permitted for your own internal business purposes.
- **Competing SaaS**: not permitted under BSL. Contact licensing@startcodex.com for commercial licensing.
- **Change date**: each version converts to Apache License 2.0 four years after release.

Tookly is source-available. [Tookly Cloud](https://tookly.com) is a commercial SaaS product maintained by Start Codex SAS.
