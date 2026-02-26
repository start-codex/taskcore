# Arquitectura Técnica

## Objetivos técnicos

- Binario único fácil de desplegar.
- Bajo consumo de memoria/CPU.
- Mantenible y extensible por dominios.
- UX moderna sin forzar SPA pesada.

## Stack

- Backend: Go 1.23.
- Router HTTP: `net/http` stdlib (Go 1.22+ soporta method routing y path params nativos).
- DB: PostgreSQL.
- SQL layer: `database/sql` + `sqlx`.
- Auth: sesiones con cookies seguras + OAuth opcional.
- Migrations: golang-migrate.

## Estilo de aplicación

- Monolito modular (no microservicios en MVP).
- API REST versionada (`/api/v1`) para integraciones.
- Modelo Jira-like: issue centrado en proyecto/status; board como vista (filtro + columnas).

## Estructura de directorios

```
cmd/
  server/
    main.go                   # entrypoint: configura DB, router, arranca servidor

internal/
  api/
    router.go                 # net/http mux + middlewares globales (logger, requestID, recover)
    issues.go                 # handlers de issues
    projects.go               # handlers de proyectos
    boards.go                 # handlers de tableros
    workspaces.go             # handlers de workspaces
    users.go                  # handlers de usuarios

  issues/
    issues.go                 # tipos, errores y API pública
    store.go                  # persistencia SQL privada del paquete
    issues_test.go            # unit tests (sin DB)
    store_integration_test.go

  projects/
    projects.go
    store.go
    projects_test.go
    store_integration_test.go

  boards/
    boards.go
    store.go
    boards_test.go
    store_integration_test.go

  workspaces/
    workspaces.go
    store.go
    workspaces_test.go
    store_integration_test.go

  users/
    users.go
    store.go
    users_test.go
    store_integration_test.go

migrations/
  0001_init.up.sql
  0001_init.down.sql
```

## Reglas de diseño

- Un paquete por dominio, no por capa técnica.
- No crear `internal/domain`, `internal/app` o `internal/store` globales.
- No usar subdirectorios por patrón OOP (`repository/`, `service/`, `manager/`) dentro de cada dominio.
- Dominio y persistencia conviven en el mismo paquete:
  - `<dominio>.go`: tipos, errores, validaciones, API pública.
  - `store.go`: SQL privado y detalles de persistencia.
- Preferir funciones libres con dependencias explícitas (ej: `func MoveIssue(ctx, db, p)`).
- SQL explícito y testeable con PostgreSQL real en integración.
- Interfaces solo cuando exista una necesidad concreta (no preventivas).

Ver detalle completo en [docs/04-go-conventions.md](04-go-conventions.md).

## Handlers

Delgados por diseño: solo parsean el request, llaman la función de dominio y escriben la respuesta.

```go
mux.HandleFunc("POST /api/v1/projects/{projectID}/issues", handleCreateIssue(db))
mux.HandleFunc("GET /api/v1/projects/{projectID}/issues/{issueID}", handleGetIssue(db))

func handleCreateIssue(db *sqlx.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        projectID := r.PathValue("projectID")
        var p issues.CreateIssueParams
        if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        p.ProjectID = projectID
        issue, err := issues.CreateIssue(r.Context(), db, p)
        // ...
    }
}
```

## Orden de implementación (MVP)

1. `internal/projects` — base de casi todo lo demás.
2. `internal/workspaces` — contexto de multi-tenancy.
3. `internal/boards` — vista del tablero.
4. `internal/issues` CRUD — crear, leer, listar, actualizar, archivar (`MoveIssue` ya existe).
5. `internal/users` — membresías y asignación.
6. `internal/api` — router HTTP + handlers por dominio.
7. Auth — sesiones con cookies.

## Observabilidad mínima

- Logs estructurados (JSON).
- Request ID por petición (middleware).
- Métricas básicas: latencia, errores, throughput.

## Seguridad mínima

- Control de acceso por workspace/proyecto.
- Validación de input en handler + dominio.
- Protección CSRF en formularios.
- Cookies `HttpOnly`, `Secure`, `SameSite`.
