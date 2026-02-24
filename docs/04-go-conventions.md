# Convenciones Go del proyecto

## Principio base

Go no es un lenguaje OOP. No trasladar patrones de Java/Spring (Repository, Service, Manager, Factory) al código. El paquete es el namespace y el identificador — no el nombre del tipo.

---

## Estructura de paquetes

Paquetes planos por dominio bajo `internal/`:

```
internal/
  issues/
    issues.go               # tipos, errores, API pública
    store.go                # persistencia SQL (privado)
    store_integration_test.go
  projects/
    projects.go
    store.go
    store_integration_test.go
  boards/
    ...
```

**Reglas:**
- Un paquete por dominio, no por capa técnica
- No crear subdirectorios dentro de un paquete de dominio (`internal/issues/repository/` está mal)
- No crear un paquete `store` que mezcle todos los dominios

---

## Separación dominio / persistencia dentro del paquete

### `<dominio>.go` — dominio
Contiene:
- Tipos de dominio (`Issue`, `MoveIssueParams`, etc.)
- Errores del dominio (`ErrIssueNotFound`)
- Validaciones de negocio (`func (p MoveIssueParams) Validate() error`)
- API pública del paquete (`func MoveIssue(ctx, db, p)`)

### `store.go` — persistencia
Contiene:
- Funciones SQL **privadas** (`func moveIssue(ctx, db, p)`)
- Tipos internos de mapeo (`type issuePosition struct`)
- Constantes de implementación (`const reorderOffset`)

No contiene:
- Validaciones de negocio
- Tipos de dominio
- Funciones exportadas

---

## Funciones vs métodos

Preferir funciones libres que reciben sus dependencias como parámetros:

```go
// correcto
func MoveIssue(ctx context.Context, db *sqlx.DB, p MoveIssueParams) error

// incorrecto — struct innecesario solo para cargar db
type Store struct { db *sqlx.DB }
func (s *Store) MoveIssue(ctx context.Context, p MoveIssueParams) error
```

Usar un struct solo cuando se necesita transportar **estado mutable** entre múltiples operaciones o cuando hay más de una dependencia que se configura una sola vez (servidor HTTP, cliente externo, etc.).

---

## Nombres

| Patrón OOP (evitar) | Go idiomático |
|---|---|
| `NewIssueRepository(db)` | `issues.New(db)` o directamente `issues.MoveIssue(ctx, db, p)` |
| `type IssueRepository struct` | `type Store struct` o eliminar el struct |
| `type IssueService struct` | funciones en el paquete `issues` |
| `IssueManager`, `IssueHandler` | funciones con nombre descriptivo |

El constructor, si existe, se llama `New`. El tipo principal del paquete refleja qué es, no el patrón que implementa.

---

## Interfaces

Definir interfaces **solo cuando se necesitan**:
- Al menos dos implementaciones concretas, o
- Necesidad real de inyectar un mock en tests

No definir interfaces preventivas. En Go, las interfaces se definen del lado del consumidor, no del productor.

```go
// incorrecto — interface sin consumidor real
type IssueStorer interface {
    MoveIssue(ctx context.Context, p MoveIssueParams) error
}

// correcto — solo cuando hay necesidad concreta
```

---

## Persistencia

- SQL explícito: sin ORM pesado, usar `sqlx`
- Las funciones SQL son privadas al paquete
- La validación de inputs ocurre en el dominio antes de llegar a la persistencia

---

## Tests

### Archivos por paquete

```
internal/issues/
  issues_test.go              # unit tests — lógica de dominio, sin DB
  store_integration_test.go   # integration tests — PostgreSQL real
```

### Unit tests (`issues_test.go`)

- Prueban lógica de dominio pura: validaciones, reglas de negocio, tipos
- Sin base de datos, sin dependencias externas
- Siempre **table-driven**

```go
func TestMoveIssueParams_Validate(t *testing.T) {
    tests := []struct {
        name    string
        p       MoveIssueParams
        wantErr bool
    }{
        {"valid params",           MoveIssueParams{ProjectID: "p", IssueID: "i", TargetPosition: 0}, false},
        {"missing project_id",     MoveIssueParams{ProjectID: "", IssueID: "i"},                     true},
        {"negative target_position", MoveIssueParams{ProjectID: "p", IssueID: "i", TargetPosition: -1}, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.p.Validate()
            if (err != nil) != tt.wantErr {
                t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration tests (`store_integration_test.go`)

- Prueban persistencia contra PostgreSQL real
- Requieren `MINI_JIRA_TEST_DSN`; se saltan automáticamente si no está configurado (`t.Skip`)
- Los tests determinísticos van en un único `TestX` **table-driven**
- Los tests de concurrencia van como funciones separadas (goroutines y channels no encajan en table-driven)

#### Estructura table-driven para integración

```go
func TestMoveIssue(t *testing.T) {
    db := openTestDB(t)    // una sola vez por TestX
    ensureSchema(t, db)    // una sola vez por TestX

    tests := []struct {
        name    string
        arrange func(*testing.T, *sqlx.DB, projectSeed) (MoveIssueParams, func(*testing.T))
        wantErr error
    }{
        {
            name: "within same status",
            arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (MoveIssueParams, func(*testing.T)) {
                a := insertIssue(t, db, seed, ...)
                b := insertIssue(t, db, seed, ...)
                p := MoveIssueParams{..., IssueID: b, TargetPosition: 0}
                return p, func(t *testing.T) {
                    assertOrder(t, fetchStatusOrder(...), []orderedIssue{{ID: b, Pos: 0}, {ID: a, Pos: 1}})
                }
            },
        },
        {
            name:    "issue not found",
            wantErr: ErrIssueNotFound,
            arrange: func(t *testing.T, db *sqlx.DB, seed projectSeed) (MoveIssueParams, func(*testing.T)) {
                return MoveIssueParams{..., IssueID: "00000000-0000-0000-0000-000000000000"}, nil
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            seed := seedProject(t, db)  // proyecto fresco por caso, cleanup automático
            p, check := tt.arrange(t, db, seed)
            err := MoveIssue(context.Background(), db, p)
            if !errors.Is(err, tt.wantErr) {
                t.Fatalf("MoveIssue() error = %v, wantErr = %v", err, tt.wantErr)
            }
            if check != nil {
                check(t)
            }
        })
    }
}
```

**Claves del patrón:**
- `arrange` retorna los parámetros de la llamada **y** un closure de assert que captura los IDs insertados
- `seedProject` se llama dentro de cada subtest — datos aislados, limpieza via `t.Cleanup` + cascade delete
- `db` y `ensureSchema` se crean una sola vez fuera del loop para no reconectar en cada caso
- Si el caso solo verifica error, `check` es `nil`
