# Mini Jira OSS

Mini Jira OSS es una plataforma open source para gestionar trabajo en tableros tipo Kanban/Scrum.

Objetivo:
- Empezar simple para equipos de desarrollo.
- Mantener flexibilidad para cualquier negocio (marketing, operaciones, soporte, etc.).
- Ser mantenible, moderno y extensible.

## Principios

- Open source primero: contribuciones claras, documentación y arquitectura modular.
- Configurable por proyecto: cada proyecto puede definir su propio flujo.
- MVP mínimo: `Por hacer`, `En curso`, `Finalizado`.
- Escalable: tipos de trabajo (Epic, HU, Task, Subtask) como configuración, no hardcode.

## Dirección técnica

- Enfoque estilo Gitea/Forgejo: monolito en Go, liviano y fácil de desplegar.
- UI server-rendered (templates) con mejoras progresivas.
- API REST para integraciones y clientes externos.
- Acceso a datos con database/sql + sqlx y queries SQL explícitas.

## MVP (fase 1)

- Gestión de organizaciones/equipos.
- Proyectos.
- Tableros por proyecto.
- Columnas/estados base:
  - Por hacer
  - En curso
  - Finalizado
- Tareas con:
  - Título
  - Descripción
  - Estado
  - Prioridad
  - Responsable
  - Fecha límite
- Drag & drop de tareas entre columnas.
- Historial de cambios básico (auditoría mínima).

## Fase 2 (dev-centric)

- Tipos de issue configurables por proyecto:
  - Epic
  - Historia de usuario (HU)
  - Task
  - Subtask
- Relación jerárquica entre issues (Epic -> HU -> Task -> Subtask).
- Sprint planning básico.
- Backlog.

## Fase 3 (cross-industry)

- Plantillas por industria (dev, marketing, soporte, legal, operaciones).
- Automatizaciones básicas (ej. al pasar a Finalizado, notificar).
- Métricas e informes.

## Documentación

- Alcance de producto: [docs/01-product-scope.md](docs/01-product-scope.md)
- Arquitectura técnica: [docs/02-architecture.md](docs/02-architecture.md)
- Modelo de datos inicial: [docs/03-data-model.md](docs/03-data-model.md)

## Decisiones clave

1. El flujo de estados es configurable por proyecto.
2. Los tipos de issue son configurables por proyecto.
3. Se empieza con defaults simples para acelerar adopción.
4. Stack base en Go (estilo Gitea/Forgejo), no Node como dependencia principal.
5. El issue pertenece al proyecto; los boards son vistas configurables (estilo Jira).

## Licencia

Sugerida: AGPL-3.0 o MIT.
- AGPL-3.0 si quieres proteger mejoras en despliegues SaaS.
- MIT si priorizas adopción máxima.
