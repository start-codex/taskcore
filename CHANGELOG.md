# Changelog

This file tracks notable changes to Tookly.

The project does not use tagged releases yet. Until versioning starts, history is recorded using an `Unreleased` section for ongoing work, plus dated milestone entries for committed historical work.

Contributors should add ongoing changes to the `Unreleased` section. When a milestone or release happens, move those entries into a new dated section at the top of the history.

---

## [Unreleased]

### Added
- Added workspace invitations: create, list, revoke, resend, accept (authenticated or new user registration)
- Added `internal/invitations` package with token-based invitation flow (SHA-256, 7-day TTL)
- Added `GET/POST /invitations/accept` public endpoints for invitation acceptance
- Added forgot password page at `/forgot-password` and reset password page at `/reset-password`
- Added "Forgot password?" link to login form
- Added success banner on login after password reset
- Added forgot/reset password flow: `POST /auth/forgot-password` and `POST /auth/reset-password`
- Added `internal/passwordreset` package with token-based reset (SHA-256, 1h TTL, one-time use)
- Added `users.SetPassword` for password reset without current password verification
- Added change password flow: `POST /auth/change-password` with session invalidation
- Added `ChangePasswordForm` shared component for account settings
- Added minimum password length enforcement (8 characters)
- Added instance admin settings page at `/admin` with SMTP configuration form
- Added `is_instance_admin` to frontend User type
- Added "Admin" link in user dropdown for instance admins
- Added transactional email foundation with `internal/email` package and embedded HTML templates
- Added Mailpit to docker-compose for local SMTP testing
- Added SMTP configuration API: `GET/POST /instance/smtp` and `POST /instance/smtp/test` (instance admin only)
- Added setup wizard at `/setup` for first-install instance configuration
- Added instance status check guards on login and app layouts
- Added instance bootstrap flow: `POST /instance/bootstrap` creates first admin user atomically with `SELECT FOR UPDATE`
- Added `GET /instance/status` endpoint for checking initialization state
- Added `is_instance_admin` field to users for instance-level administration
- Added `RequireInstanceAdmin` to `internal/authz` for instance admin enforcement
- Added `POST /users` blocked (409) before instance bootstrap, public after
- Added `instance_config` table and `internal/instance` package for instance-level key-value configuration
- Added session storage foundation (`internal/sessions`): `Create`, `Validate`, `Delete`
- Added SHA-256 token hashing — raw tokens never stored in the database
- Added archived-user rejection in session validation (`ErrUserArchived`)
- Added migration `0003_create_sessions` with indexes on `user_id` and `expires_at`
- Added workspace-scoped URL routing (`/{workspace}/...`)
- Added Kanban and Scrum project templates with localized status names per user language
- Added project creation wizard: step 1 template selection, step 2 name and key
- Added empty states with action buttons across all pages (no workspace, no projects, no boards, no statuses)
- Added settings page with language switcher
- Added projects overview dashboard at `/{workspace}/`
- Added `withAuth` session middleware with allowlist for public routes
- Added `GET /auth/me` endpoint — always 200, distinguishes auth errors from internal errors
- Added `POST /auth/logout` endpoint — idempotent, clears cookie, returns 204
- Added `internal/authz` package with membership authorization helpers and context user ID helpers
- Added workspace membership enforcement on all API routes (read and write)
- Added project-member creation guard: target user must be a workspace member
- Added `sessions.IsAuthError` helper for centralized error classification
- Added basic board filters: client-side filtering by assignee, priority, and issue type
- Added board drag-and-drop: move issues between status columns and reorder within columns via `svelte-dnd-action`
- Added issue detail page at `/{workspace}/projects/{id}/issues/{issueID}` with edit for title, description, priority, assignee, due date
- Added clickable issue cards on board view linking to issue detail
- Added `due_date` field to issue create and update API contracts (`YYYY-MM-DD` or `null`)
- Added `RequireWorkspaceAdmin` to `internal/authz` for admin/owner role enforcement
- Added `ApiError` class to frontend API client for typed HTTP error handling
- Added `auth.me()` and `auth.logout()` to frontend API client
- Added `signIn()`, `restore()`, and `logout()` to frontend auth store
- Added BSL 1.1 license (replaces AGPL-3.0) with Apache 2.0 change license after 4 years per version
- Added Contributor License Agreement (CLA.md)
- Added Contributing guide (CONTRIBUTING.md)
- Added a root changelog to track notable project changes
- Added a README link to the changelog

### Changed
- Changed `POST /auth/login` to create session and set `HttpOnly` cookie with `SameSite=Strict`
- Changed `GET /users/{userID}` to enforce self-only access (403 on mismatch)
- Changed login to reject archived users before session creation
- Changed `internal/authctx` consolidated into `internal/authz`
- Changed `POST /workspaces` to derive owner from authenticated session (removed `owner_id` from request body)
- Changed `GET /workspaces` to derive user from authenticated session (removed `user_id` query parameter)
- Changed `POST /projects/{projectID}/issues` to derive reporter from authenticated session (removed `reporter_id` from request body)
- Changed workspace admin routes (`DELETE /workspaces/{id}`, member management) to require admin/owner role
- Changed project admin routes (`POST /workspaces/{id}/projects`, `DELETE /projects/{id}`, member management) to require workspace admin/owner role
- Changed workflow config routes (boards, columns, statuses, issue types) to require workspace admin/owner role
- Changed `GET /workspaces/{id}/members` from admin-only to member-level (read access for assignee selection)
- Changed `GET /projects/{id}/members` from admin-only to member-level (read access)
- Changed frontend auth from localStorage to session-based `/auth/me` validation
- Changed frontend auth store to in-memory only (removed localStorage persistence)
- Changed frontend logout to call backend `POST /auth/logout` before clearing state
- Changed authz resource resolution to allow archived projects, boards, and columns (domain handlers decide visibility)
- Changed project name from Taskcore to Traza Work, then to Tookly
- Changed license from AGPL-3.0 to BSL 1.1
- Changed workspace creation to add creator as owner member in a single transaction
- Changed sidebar workspace selection to sync via URL navigation instead of local state
- Changed board view to sync statuses on navigation via `$effect`

### Fixed
- Fixed Go nil slice serialization returning JSON `null` instead of `[]`
- Fixed board not updating when switching between projects

---

## 2026-03-04

### Added
- Added workspace-scoped project loading and standardized API responses
- Added English and Spanish internationalization (i18n)

### Changed
- Changed project name from Mini Jira OSS to Taskcore

---

## 2026-02-26

### Added
- Added password authentication and member management
- Added SvelteKit frontend integration in the Docker/app flow

---

## 2026-02-25

### Added
- Added Docker app service and migration support

### Changed
- Changed database initialization to application-managed migrations at startup
- Changed architecture and documentation toward domain-per-package structure and stdlib routing

---

## 2025-12-08 / 2025-12-09

### Added
- Added the initial project structure, Docker setup, and database migrations
- Added early application modules: issues, projects, middleware, and initial auth work
- Initial platform foundation established
