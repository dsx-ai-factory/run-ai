# `runaidb` Package Documentation

The `runaidb` package provides a unified database layer for Run.ai services, built on top of [GORM](https://gorm.io) with PostgreSQL. It manages connection lifecycle, schema migrations (via [`golang-migrate`](https://github.com/golang-migrate/migrate)), per-service session users, multi-schema access, and ships with first-class testing helpers (typed mock + SQL mock).

---

## Table of Contents

- [Features and Public API](#features-and-public-api)
  - [Initialization (`Init`)](#initialization-init)
  - [DB](#db)
  - [NewDB (per-schema connection)](#newdb-per-schema-connection)
  - [Migrate](#migrate)
  - [SetLogger](#setlogger)
  - [Close](#close)
  - [SetMock](#setmock)
- [The `orm.DB` Interface](#the-ormdb-interface)
- [Configuration](#configuration)
- [Migrations](#migrations)
- [Service Session Users](#service-session-users)
- [Example Usage](#example-usage)
- [Testing and Mocking](#testing-and-mocking)
  - [SetOrmMock — interface-level mock](#setormmock--interface-level-mock)
  - [SetSQLMock — SQL-level mock](#setsqlmock--sql-level-mock)
  - [Argument matchers](#argument-matchers)
  - [Choosing between the two mocks](#choosing-between-the-two-mocks)

---

## Features and Public API

### Initialization (`Init`)

Initializes the database connection, optionally runs auto-migrations, and (by default) provisions a per-service database user used for the runtime session. This is a **blocking** operation that returns once the connection has been established (or the underlying connector returns a fatal error). `Init` is guarded by a `sync.Once` — calling it twice is a no-op until `Close` is called.

You can pass a configuration struct to `Init` to programmatically set options instead of (or in addition to) using environment variables.

> **Note:** As a general practice, it is usually preferred to allow the package to load the configuration by itself (using environment variables) rather than providing a config struct to the `Init` function. This ensures a consistent configuration standard across all services using this package.

**Example:**

```go
cfg := &dbconfig.Config{
    Connection: dbconfig.ConnectionConfig{
        Host:     "localhost",
        Port:     5432,
        Name:     "backend",
        Schema:   "public",
        User:     "postgres",
        Password: "postgres",
        SslMode:  dbconfig.SslModeDisable,
    },
    Migration: dbconfig.MigrationConfig{
        Auto: ptr.Bool(true),
    },
    ServiceName: "my-service",
}
err := runaidb.Init(cfg)
if err != nil {
    panic(err)
}
```

Pass `nil` to load defaults from environment variables:

```go
err := runaidb.Init(nil)
if err != nil {
    panic(err)
}
```

- **Blocking**: Yes
- **Idempotent**: Yes (subsequent calls are no-ops until `Close`)

---

### DB

Returns the global primary `orm.DB` handle established by `Init`. Returns `nil` if the package has not been initialized.

```go
db := runaidb.DB()
err := db.WithContext(ctx).Where("id = ?", id).First(&user)
```

---

### NewDB (per-schema connection)

Returns a dedicated `orm.DB` handle bound to a specific PostgreSQL **schema**. Useful for services that read/write across multiple schemas (e.g. tenant-per-schema). The handle is cached per schema name, so repeated calls for the same schema return the same connection.

- If the requested schema matches the schema of the primary connection, the primary handle is returned.
- Auto-migration is **disabled** for schema connections created via `NewDB` (only the primary connection runs migrations on startup).

```go
schemaDB, err := runaidb.NewDB("schema_name")
if err != nil {
    return err
}
err = schemaDB.WithContext(ctx).Find(&projects)
```

---

### Migrate

Runs migrations explicitly, outside of the auto-migrate path on `Init`. Useful for CLI tools and operators that need to step a database forward, roll it back, or `force` a version.

`Migrate` opens a fresh admin connection (it does **not** use the service-session user), runs the migration, and closes the connection — it does not affect the global `DB()` handle.

```go
err := runaidb.Migrate(&runaidb.MigrationParams{
    Direction:     dbconfig.MigrationDirectionUp,    // up | down | force
    TargetVersion: ptr.String("20251001120000"),     // optional
})
```

- **Blocking**: Yes
- **Requires:** `Init` to have been called first.

---

### SetLogger

Sets a custom logger used by the package and by the underlying GORM driver. Should be called **before** `Init`. If not set, a default no-op logger is used.

The logger must implement:

```go
type Logger interface {
    Infof(format string, args ...any)
    Errorf(format string, args ...any)
    Warnf(format string, args ...any)
    Debugf(format string, args ...any)
}
```

```go
runaidb.SetLogger(myLogger)
```

---

### Close

Gracefully closes the primary connection and any per-schema connections opened via `NewDB`, and resets the internal `sync.Once` so the package can be re-initialized. Returns a joined error if any underlying connection fails to close.

```go
err := runaidb.Close()
```

- **Blocking**: Yes

---

### SetMock

Replaces the global primary DB with the provided `orm.DB`. This is the low-level primitive used by the testing helpers below — most tests should use [`dbtesting.SetOrmMock`](#setormmock--interface-level-mock) or [`dbtesting.SetSQLMock`](#setsqlmock--sql-level-mock) rather than calling `SetMock` directly.

```go
runaidb.SetMock(mockDB)
```

---

## The `orm.DB` Interface

`orm.DB` is a thin, mockable wrapper around `*gorm.DB`. It exposes the GORM operations Run.ai services use, with two key differences:

1. **Write operations return `(rowsAffected int64, err error)`** instead of returning `*gorm.DB`. This makes the rows-affected check explicit and avoids accidental chaining bugs.
2. **`Actual() *gorm.DB`** is available as an escape hatch when you genuinely need the raw GORM handle (e.g. for an unsupported clause).

The interface is composed of:

- `Transactioner` — `Transaction`, `SavePoint`, `RollbackTo`
- `QueryBuilder` — `Model`, `Table`, `Where`, `Joins`, `Preload`, `Order`, `Limit`, `Clauses`, `Raw`, … &nbsp;_(chainable, return `DB`)_
- `QueryReader` — `First`, `Take`, `Last`, `Find`, `Count`, … &nbsp;_(terminators, return `error`; `Row`/`Rows` return the underlying `*sql.Row` / `(*sql.Rows, error)`)_
- `QueryWriter` — `Create`, `Save`, `Updates`, `Delete`, `Exec`, `Scan`, … &nbsp;_(terminators, return `(rowsAffected int64, err error)`)_

Plus `WithContext`, `Session`, and `Close`.

GORM clauses are re-exported under `runaidb/orm/clause` (e.g. `clause.Returning{}`, `clause.OnConflict{}`, `clause.Locking{}`) so callers don't need to import `gorm.io/gorm/clause` directly.

---

## Configuration

All configuration options can be set via environment variables or via the config struct passed to `Init`. Below is the full list of supported options.

| ENV Variable | Description | Default |
|---|---|---|
| `DB_HOST` | PostgreSQL host (**required**). Accepts a comma-separated multi-host list for client-side failover (e.g. `h1,h2,h3` or `h1:5432,h2:5433`); hosts without an inline `:port` use `DB_PORT` | — |
| `DB_PORT` | PostgreSQL port; applied to any `DB_HOST` entry that omits an inline port | `5432` |
| `DB_NAME` | Database name | `backend` |
| `DB_SCHEMA` | Schema name (**required**) | — |
| `DB_USER` | Admin user (used for connect, migrations, and creating the service session) | — |
| `DB_PASSWORD` | Admin password | — |
| `DB_SSL_MODE` | One of `disable`, `allow`, `prefer`, `require`, `verify-ca`, `verify-full` | `disable` |
| `DB_STATEMENT_TIMEOUT` | Per-statement timeout (Go duration string) | `2min` |
| `DB_TARGET_SESSION_ATTRS` | pgx multi-host session selection; one of `any`, `read-write`, `read-only`, `primary`, `standby`, `prefer-standby`. Use `read-write` to pick the primary from a multi-host `DB_HOST` | — |
| `DB_READ_HOST` | Read-replica host. Setting it enables a dedicated read connection (`ReadDB()`). Accepts a comma-separated multi-host list, same as `DB_HOST` | — |
| `DB_READ_PORT` | Read-replica port; applied to any `DB_READ_HOST` entry that omits an inline port | `5432` |
| `DB_READ_USER` | Read-replica user (**required when `DB_READ_HOST` is set**) | — |
| `DB_READ_PASSWORD` | Read-replica password (**required when `DB_READ_HOST` is set**) | — |
| `DB_READ_SSL_MODE` | Read-replica SSL mode; same values as `DB_SSL_MODE` | `disable` |
| `DB_READ_TARGET_SESSION_ATTRS` | Session selection for the read connection; same values as `DB_TARGET_SESSION_ATTRS`. Use `prefer-standby`/`read-only` to target replicas. If unset, the read path does **not** inherit `DB_TARGET_SESSION_ATTRS` | — |
| `DB_MAX_CONNECTIONS` | Connection pool size | `20` |
| `DB_MAX_IDLE_CONNECTIONS` | Max idle connections in the pool | `2` |
| `DB_MAX_IDLE_CONNECTIONS_TIME_DURATION` | Max idle time for a pooled connection | `1h` |
| `DB_MIGRATION_AUTO` | Run migrations automatically on `Init` | `true` |
| `DB_MIGRATION_FILES` | Source URL passed to `golang-migrate` (e.g. `file:///migrations`) | `file:///migrations` |
| `DB_MIGRATION_DIRECTION` | `up`, `down`, or `force` | `up` |
| `DB_MIGRATION_TARGET_VERSION` | Specific version to migrate to (optional) | — |
| `DB_MIGRATION_LOCK_RETRIES` | Retries when acquiring the migration advisory lock | `3` |
| `DB_USE_SERVICE_SESSION` | If `true`, create/use a per-service DB role for runtime queries (see below) | `true` |
| `DB_ENCRYPTION_KEY` | Optional AES key (must be empty, 16, 24, or 32 bytes) used by encryption helpers | — |
| `LOG_LEVEL` | One of `Error`, `Warn`, `Info`, `Debug` | `Info` |
| `SERVICE_NAME` | Service name (**required**); used as the service-session role name | — |

---

## Migrations

`runaidb` integrates [`golang-migrate`](https://github.com/golang-migrate/migrate) and reads SQL migrations from the source URL configured by `DB_MIGRATION_FILES` (typically `file:///migrations` in a Docker image).

- When `DB_MIGRATION_AUTO=true` (the default), migrations run during `Init` against the **admin** user — before any service-session user is created or used.
- For one-shot operations (e.g. CI jobs, rollbacks), call `runaidb.Migrate(...)` explicitly with the desired `Direction` (`up`, `down`, `force`) and an optional `TargetVersion`.
- An advisory lock is held during migration; concurrent migrators will retry up to `DB_MIGRATION_LOCK_RETRIES` times.

---

## Service Session Users

When `DB_USE_SERVICE_SESSION=true` (the default), `Init`:

1. Connects as the admin user (`DB_USER`) and runs migrations.
2. Creates (or reuses) a PostgreSQL role named after `SERVICE_NAME`, with the same password as the admin user, and grants it permissions on the configured schema.
3. Re-opens the connection as the service role, closes the admin connection, and stores the service-role connection as the global `DB()`.

This gives each service its own least-privilege identity in the database while keeping startup self-contained. If the admin user lacks privileges to create the role (e.g. in a managed environment where roles are pre-provisioned), `Init` logs a warning and falls back to the admin connection.

---

## Example Usage

```go
import (
    "context"

    "github.com/run-ai/runai/runai-common-packages/runaidb"
    dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"
)

type Project struct {
    ID        uint   `gorm:"primaryKey"`
    Name      string `gorm:"column:name"`
    ClusterID string `gorm:"column:cluster_id"`
}

func main() {
    runaidb.SetLogger(myLogger)

    if err := runaidb.Init(nil); err != nil {
        panic(err)
    }
    defer runaidb.Close()

    ctx := context.Background()
    db := runaidb.DB().WithContext(ctx)

    // Create
    p := &Project{Name: "alpha", ClusterID: "cluster-1"}
    if _, err := db.Create(p); err != nil {
        panic(err)
    }

    // Read
    var projects []Project
    if err := db.Where("cluster_id = ?", "cluster-1").Find(&projects); err != nil {
        panic(err)
    }

    // Update
    if _, err := db.Model(&Project{}).
        Where("id = ?", p.ID).
        Update("name", "beta"); err != nil {
        panic(err)
    }

    // Transaction
    err := db.Transaction(func(tx orm.DB) error {
        if _, err := tx.Create(&Project{Name: "gamma"}); err != nil {
            return err
        }
        if _, err := tx.Create(&Project{Name: "delta"}); err != nil {
            return err
        }
        return nil
    })
    if err != nil {
        panic(err)
    }

    // Run a one-off migration step
    if err := runaidb.Migrate(&runaidb.MigrationParams{
        Direction: dbconfig.MigrationDirectionUp,
    }); err != nil {
        panic(err)
    }
}
```

---

## Testing and Mocking

The `runaidb/testing` sub-package provides two complementary mocking strategies. Both swap out the global `runaidb.DB()` for the duration of a single `*testing.T`, and both auto-restore the original via `t.Cleanup`.

```go
import dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
```

### SetOrmMock — interface-level mock

`SetOrmMock` returns a [testify](https://github.com/stretchr/testify) `*mocks.DB` for `orm.DB`. You set expectations on the **interface methods** that are called (e.g. `Where`, `Find`), not on the SQL produced.

Use it when:

- The code under test calls a small, fixed set of `orm.DB` methods.
- You don't care about exact SQL; you care that the right method was invoked with the right Go-level arguments.
- You want to assert behavior in tightly-scoped unit tests (repository methods, service layer).

```go
func TestRepo_GetProject(t *testing.T) {
    dbMock := dbtesting.SetOrmMock(t)

    dbMock.On("WithContext", mock.Anything).Return(dbMock)
    dbMock.On("Where", "id = ?", 42).Return(dbMock)
    dbMock.On("First", mock.AnythingOfType("*Project"), mock.Anything).
        Run(func(args mock.Arguments) {
            *(args.Get(0).(*Project)) = Project{ID: 42, Name: "alpha"}
        }).
        Return(nil)

    p, err := repo.GetProject(context.Background(), 42)
    require.NoError(t, err)
    require.Equal(t, "alpha", p.Name)
}
```

### SetSQLMock — SQL-level mock

`SetSQLMock` wires a real `*gorm.DB` on top of [`go-sqlmock`](https://github.com/DATA-DOG/go-sqlmock) and exposes a `Mocker` you use to declare the **exact SQL queries** the test expects. The whole driver/dialect/clause-builder pipeline runs for real — only the underlying `database/sql` driver is faked.

Use it when:

- You want to verify the SQL your code produces (column lists, joins, ordering, soft-deletes).
- The code under test exercises GORM features that are awkward to mock at the interface level (e.g. `Clauses(clause.Returning{})`, `Unscoped`, savepoints).
- You want integration-test confidence without a real database.

The mocker provides four entry points:

| Method | Purpose |
|---|---|
| `ExpectQuery(sql, args...)` | A `SELECT` (or any statement that returns rows, including `UPDATE … RETURNING`). Supply the expected rows via `.Returns([]Model{...})`. |
| `ExpectExec(sql, args...)` | A statement that returns row count only (`INSERT`, `UPDATE`, `DELETE`). Supply rows-affected via `.Returns(n)`. |
| `ExpectTransaction()` | Returns a `TransactionExpectation` with `.Begin()`, `.Commit()`, `.Rollback()`. |
| `ExpectClose()` | Expects a connection close. |

Each `Expect…` returns an expectation handle; both `Returns(...)` and `ReturnsError(err)` are available.

The matcher is **whitespace- and quote-insensitive**: `"` characters and runs of whitespace are normalized before comparison, and `$N` placeholders in the actual SQL are converted to `?`. This means you can write expected SQL the way you'd read it, without GORM's identifier-quoting noise.

```go
import (
    "errors"
    "testing"

    "github.com/run-ai/runai/runai-common-packages/runaidb/orm"
    "github.com/run-ai/runai/runai-common-packages/runaidb/orm/clause"
    dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
    "github.com/stretchr/testify/require"
    "gorm.io/gorm"
)

type Project struct {
    gorm.Model
    Name      string `gorm:"column:name"`
    ClusterID string `gorm:"column:cluster_id"`
    Status    string `gorm:"column:status"`
}

func (Project) TableName() string { return "projects" }

func TestListProjects(t *testing.T) {
    mocker := dbtesting.SetSQLMock(t)

    mocker.ExpectQuery(
        "SELECT * FROM projects WHERE (projects.cluster_id = ? AND projects.status = ?) "+
            "AND projects.deleted_at IS NULL ORDER BY name ASC",
        "cluster-1", "active",
    ).Returns([]Project{
        {Name: "proj-a", ClusterID: "cluster-1", Status: "active"},
        {Name: "proj-b", ClusterID: "cluster-1", Status: "active"},
    })

    var got []Project
    err := runaidb.DB().Table("projects").
        Where(Project{ClusterID: "cluster-1", Status: "active"}).
        Order("name ASC").
        Find(&got)

    require.NoError(t, err)
    require.Len(t, got, 2)
    require.Equal(t, "proj-a", got[0].Name)
	require.Equal(t, "proj-b", got[1].Name)
}

func TestSoftDelete_InTransaction(t *testing.T) {
    mocker := dbtesting.SetSQLMock(t)

    // GORM wraps every write in an implicit BEGIN/COMMIT — expect them even
    // though the chained call doesn't show it (opt out via SkipDefaultTransaction).
    mocker.ExpectTransaction().Begin()
    mocker.ExpectExec(
        "UPDATE projects SET deleted_at=? WHERE projects.cluster_id = ? AND projects.deleted_at IS NULL",
        dbtesting.AnyTime(), "cluster-1",
    ).Returns(2)
    mocker.ExpectTransaction().Commit()

    rows, err := runaidb.DB().Table("projects").
        Where(Project{ClusterID: "cluster-1"}).
        Delete(&Project{})

    require.NoError(t, err)
    require.Equal(t, int64(2), rows)
}

func TestUpdateReturning(t *testing.T) {
    mocker := dbtesting.SetSQLMock(t)

    mocker.ExpectTransaction().Begin()
    mocker.ExpectQuery(
        "UPDATE projects SET status=? WHERE projects.cluster_id = ? "+
            "AND projects.deleted_at IS NULL RETURNING *",
        "archived", "cluster-1",
    ).Returns([]Project{{Name: "proj-a", ClusterID: "cluster-1", Status: "archived"}})
    mocker.ExpectTransaction().Commit()

    p := &Project{Status: "archived"}
    _, err := runaidb.DB().Table("projects").
        Where(Project{ClusterID: "cluster-1"}).
        Clauses(clause.Returning{}).
        Updates(p)

    require.NoError(t, err)
    require.Equal(t, "proj-a", p.Name)
}

func TestQuery_Error(t *testing.T) {
    mocker := dbtesting.SetSQLMock(t)

    mocker.ExpectQuery("SELECT * FROM projects WHERE projects.cluster_id = ? AND projects.deleted_at IS NULL", "cluster-1").
        ReturnsError(errors.New("db down"))

    var got []Project
    err := runaidb.DB().Table("projects").
        Where(Project{ClusterID: "cluster-1"}).
        Find(&got)

    require.EqualError(t, err, "db down")
}
```

**Notes on writing expected SQL**

- Column names follow GORM's snake-case naming strategy or the explicit `gorm:"column:..."` tag.
- For models that embed `gorm.Model`, soft-delete predicates (`deleted_at IS NULL`) appear automatically — include them in the expected SQL or use `Unscoped()` to skip them.
- Use `?` placeholders, not `$1`/`$2` — the matcher normalizes the actual SQL to `?` form before comparing.
- Any unmet expectation fails the test.

When using rows-returning helpers, `Returns([]Model{...})` reflects over the slice's element type to derive column names from `gorm:"column:..."` tags (falling back to the snake-case field name). Embedded structs (e.g. `gorm.Model`) are flattened. Pass an empty slice (`[]Model{}`) to return zero rows.

### Argument matchers

When the exact value of an argument doesn't matter (e.g. timestamps written by GORM, generated IDs), use the matchers exported by the testing package:

| Matcher | Matches |
|---|---|
| `dbtesting.AnyArg()` | Any value |
| `dbtesting.AnyTime()` | Any `time.Time` |

```go
mocker.ExpectExec(
    "UPDATE users SET name = ?, updated_at = ? WHERE id = ?",
    dbtesting.AnyArg(), dbtesting.AnyTime(), 1,
).Returns(1)
```

### Choosing between the two mocks

| Concern | `SetOrmMock` | `SetSQLMock` |
|---|---|---|
| What you assert on | Calls to `orm.DB` methods | Generated SQL + args |
| Test coupling | To your repository's API | To your repository's SQL output |
| Catches GORM-level mistakes (wrong clause, missing `Unscoped`, etc.)? | No | Yes |
| Best for | Service-layer unit tests where the DB is a collaborator | Repository-layer tests where SQL correctness is the contract |

A common pattern is to use `SetSQLMock` for repository tests (verify the SQL) and `SetOrmMock` for service/handler tests above the repository (verify the orchestration).
