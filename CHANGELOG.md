# Changelog

This file tracks notable changes to Taskcore.

The project does not use tagged releases yet. Until versioning starts, history is recorded using an `Unreleased` section for ongoing work, plus dated milestone entries for committed historical work.

Contributors should add ongoing changes to the `Unreleased` section. When a milestone or release happens, move those entries into a new dated section at the top of the history.

---

## [Unreleased]

### Added
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
- Added a root changelog to track notable project changes
- Added a README link to the changelog

### Changed
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
