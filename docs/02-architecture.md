# Arquitectura Técnica (propuesta estilo Gitea/Forgejo)

## Objetivos técnicos

- Binario único fácil de desplegar.
- Bajo consumo de memoria/CPU.
- Mantenible y extensible en módulos.
- UX moderna sin forzar SPA pesada.

## Stack recomendado

- Backend principal: Go.
- Router HTTP: Chi o Gin (preferencia: Chi por simpleza).
- Render UI: templates server-side (Go templates) + HTMX/Alpine.js opcional.
- DB: PostgreSQL (SQLite opcional para desarrollo).
- SQL layer: database/sql + sqlx (preferencia: sqlx).
- Auth: sesiones con cookies seguras + OAuth opcional.
- Migrations: golang-migrate.

## Estilo de aplicación

- Monolito modular (no microservicios en MVP).
- API REST versionada (`/api/v1`) para integraciones.
- Modelo Jira-like: issue centrado en proyecto/status; board como vista (filtro + columnas).
- UI web renderizada en servidor para rendimiento y simplicidad.

## Estructura sugerida

- `cmd/server`: entrypoint.
- `internal/http`: handlers, middlewares, routing.
- `internal/domain`: entidades y reglas de negocio.
- `internal/app`: casos de uso (services).
- `internal/store`: acceso a datos.
- `internal/auth`: autenticación/autorización.
- `web/templates`: HTML templates.
- `web/static`: CSS/JS/assets.
- `migrations`: SQL migrations.

## Módulos de dominio

- IAM: usuarios, roles, membresías.
- Projects: proyectos y miembros.
- Boards: tableros, columnas, orden.
- Issues: tareas, tipos, relaciones.
- Audit: eventos de cambios.

## Reglas de diseño

- Dominio primero: reglas en `internal/domain` + `internal/app`.
- Handlers delgados: solo parsean request y llaman casos de uso.
- SQL explícito y testeable.
- Compatibilidad progresiva: UI y API sobre el mismo dominio.

## Observabilidad mínima

- Logs estructurados (JSON).
- Request ID por petición.
- Métricas básicas (latencia, errores, throughput).

## Seguridad mínima

- Control de acceso por workspace/proyecto.
- Validación estricta de input en handler + caso de uso.
- Protección CSRF en formularios.
- Cookies `HttpOnly`, `Secure`, `SameSite`.
