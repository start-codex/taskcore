# Modelo de Datos Inicial (Jira-like)

## Principio de diseño

- El issue pertenece al proyecto (`project_id`).
- El estado (`status`) también pertenece al proyecto.
- El board es una vista/configuración sobre issues (no dueño del issue).

## Entidades principales

### Workspace

- `id`
- `name`
- `slug`
- `created_at`

### User

- `id`
- `email`
- `name`
- `created_at`

### WorkspaceMember

- `workspace_id`
- `user_id`
- `role` (`owner`, `admin`, `member`)

### Project

- `id`
- `workspace_id`
- `name`
- `key` (ej: `ENG`, `MKT`)
- `description`
- `created_at`

### ProjectMember

- `project_id`
- `user_id`
- `role` (`admin`, `member`, `viewer`)

### ProjectIssueCounter

- `project_id` (PK)
- `last_number`
- `updated_at`

### Status

- `id`
- `project_id`
- `name` (ej: `Por hacer`, `En curso`, `Finalizado`)
- `category` (`todo`, `doing`, `done`)
- `position`

### IssueType

- `id`
- `project_id`
- `name` (ej: `Epic`, `HU`, `Task`, `Subtask`, `Bug`)
- `icon`
- `level` (jerarquía: 0..n)

### Board

- `id`
- `project_id`
- `name`
- `type` (`kanban`, `scrum`)
- `filter_query` (filtro del board)

### BoardColumn

- `id`
- `board_id`
- `name`
- `position`

### BoardColumnStatus

- `board_column_id`
- `status_id`

Nota: una columna puede mapear uno o varios status.

### Issue

- `id`
- `project_id`
- `number` (secuencial por proyecto)
- `issue_type_id`
- `status_id`
- `parent_issue_id` (nullable)
- `title`
- `description`
- `priority` (`low`, `medium`, `high`, `critical`)
- `assignee_id` (nullable)
- `reporter_id`
- `due_date` (nullable)
- `status_position`
- `created_at`
- `updated_at`

Clave pública recomendada en UI/API: `PROJECT_KEY-NUMBER` (ej: `ENG-123`).

### IssueEvent (auditoría)

- `id`
- `issue_id`
- `actor_id`
- `event_type` (`created`, `updated`, `moved`, `commented`)
- `payload_json`
- `created_at`

## Reglas de integridad

- `Issue.issue_type_id` debe pertenecer al mismo `project_id` del issue.
- `Issue.status_id` debe pertenecer al mismo `project_id` del issue.
- `Issue.parent_issue_id` debe pertenecer al mismo proyecto.
- Jerarquía por nivel: el hijo debe tener `level` mayor que el padre.
- Anti-ciclo: no se permite crear ciclos en `parent_issue_id`.
- `BoardColumnStatus`: columna y status deben pertenecer al mismo proyecto.
- Posición en columna: `status_position` es única por (`project_id`, `status_id`) para issues activos (`archived_at IS NULL`).

## Índices sugeridos

- `Issue(project_id, status_id, status_position)`
- `Issue(assignee_id)`
- `Issue(parent_issue_id)`
- `Status(project_id, position)`
- `BoardColumn(board_id, position)`

## Implementación SQL

- Fuente de verdad: migraciones SQL en `migrations/`.
- Acceso desde Go: `database/sql` + `sqlx` con repositorios por módulo.
- Evitar ORM pesado para mantener control y rendimiento.

## Generación de `Issue.number`

- Estrategia recomendada: contador por proyecto (`project_issue_counters`), no `MAX(number)+1`.
- En la misma transacción de creación de issue:
```sql
INSERT INTO project_issue_counters (project_id, last_number)
VALUES ($1, 1)
ON CONFLICT (project_id)
DO UPDATE SET last_number = project_issue_counters.last_number + 1
RETURNING last_number;
```
- Usar `last_number` como `issues.number`.
