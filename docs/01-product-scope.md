# Alcance de Producto

## Problema

Las herramientas actuales de gestión suelen ser:
- Demasiado complejas para equipos pequeños.
- Muy centradas solo en software o demasiado genéricas.
- Difíciles de adaptar sin pagar planes enterprise.

## Propuesta

Construir una alternativa open source tipo Jira:
- Moderna y mantenible.
- Con curva de aprendizaje rápida.
- Adaptable a devs y a cualquier negocio.

## Usuario inicial

- Equipos de desarrollo pequeños y medianos.
- Roles: developer, tech lead, PM.

## Usuario secundario

- Equipos no técnicos que también gestionan trabajo por etapas.

## MVP funcional

### Entidades

- Workspace (equipo/empresa)
- Proyecto
- Tablero
- Estado/columna
- Issue/tarea
- Usuario

### Funcionalidades mínimas

- Crear proyecto.
- Crear tablero.
- Ver tablero Kanban.
- Crear/editar/eliminar issue.
- Mover issue entre columnas.
- Asignar responsable.
- Filtro por estado/responsable.

### Flujo por defecto

- Por hacer
- En curso
- Finalizado

## Reglas de negocio iniciales

- Todo issue pertenece a un proyecto y un tablero.
- Todo issue tiene exactamente un estado actual.
- El orden de columnas se define en configuración del tablero.
- Las transiciones permitidas se permiten entre cualquier estado (MVP).

## Extensiones previstas

- Limitar transiciones permitidas por estado.
- Campos personalizados por proyecto.
- Tipos de issue y jerarquía.
- Backlog y sprints.
