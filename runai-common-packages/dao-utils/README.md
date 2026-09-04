# dao-utils

`dao-utils` is a shared Go library in the Run.ai monorepo that provides the common
data-access-object (DAO) building blocks used by control-plane services that talk
to PostgreSQL through [GORM](https://gorm.io). It owns the pieces every service
would otherwise reimplement: connection management (writer/reader/admin roles),
schema migrations, embeddable audited base models, JSONB column types, API
filter/sort parsing, advisory schema locking, periodic cleanup jobs, and
test-container helpers. It is consumed as a versioned Go module — there is no
binary and no container image.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Importing the library](#importing-the-library)
- [Public API](#public-api)
- [Building and testing](#building-and-testing)
- [Configuration](#configuration)

## Overview

The package (Go package name `dao_utils`) is responsible for:

- **Connections** — `Init`, `DB`, `DBRead`, `ReconnectDb`, and a
  `ConnectionFactory` that opens role-scoped `*gorm.DB` handles
  (`OpenWriter`, `OpenReader`, `OpenAdmin`) from a single `DataBaseConfig`,
  including a separate read host/port for read replicas.
- **Migrations** — `Migrate` runs [golang-migrate](https://github.com/golang-migrate/migrate)
  file-based migrations, with a `MigrateLogger`, force-recovery on dirty/unknown
  versions, and optional per-service DB user creation.
- **Base models** — `BaseModel` and `DeleteBaseModel` embeddable structs that add
  a UUID primary key, `CreatedAt`/`UpdatedAt`/`DeletedAt`, and audit columns
  (`CreatedBy`/`UpdatedBy`/`DeletedBy`) with `BeforeCreate` hooks.
- **Schema locking** — `PGLock` implements PostgreSQL advisory locks
  (`TryAcquireSchemaLock`, `AcquireSchemaLock`, `ReleaseSchemaLock`) so only one
  instance runs migrations at a time.
- **Cleanup / cron** — `InitCleaner` and `CronJob` schedule recurring SQL
  maintenance queries, running through `pg_cron` when available and falling back
  to an in-process goroutine otherwise.
- **JSONB** — `JsonBMap` is a `map[string]interface{}` column type implementing
  `driver.Valuer`/`sql.Scanner` with `Clone`/`Unmarshal` helpers.
- **Filter & sort** — `filter_sort_utils.go` plus the `validators` sub-package
  parse REST API `filterBy`/`sortBy` query parameters into safe, parameterized
  GORM `WHERE`/`ORDER BY` clauses, with per-field validators (UUID, date, user
  input) and full-text search via `ApplyTSSearchParam`.
- **Testing helpers** — `testing_utils.go` spins up PostgreSQL (optionally with a
  read replica) via [testcontainers](https://golang.testcontainers.org/),
  provides `OpenPostgresForTesting`, `CleanTables`, and `SetMock`/`ResetMock`.

## Repository layout

```text
.
├── base_model.go            # BaseModel (UUID + audit + soft delete)
├── delete_base_model.go     # DeleteBaseModel (hard-delete variant)
├── config.go                # DataBaseConfig + DB_* env var names, SSL modes
├── connection_factory.go    # ConnectionFactory, OpenWriter/Reader/Admin
├── db.go                    # Init, DB, DBRead, Migrate, SetMock/ResetMock
├── cleaner.go               # InitCleaner — periodic table cleanup
├── cronjob.go               # CronJob — pg_cron or goroutine scheduling
├── pg_lock.go               # PGLock — advisory schema lock
├── json_bmap.go             # JsonBMap JSONB column type
├── filter_sort_utils.go     # API filter/sort → GORM clauses, TS search
├── utils.go                 # migration + transaction helpers
├── testing_utils.go         # testcontainers Postgres helpers
├── validators/              # field + filter/sort validators
├── Makefile
└── go.mod
```

## Importing the library

The module path comes from [`go.mod`](go.mod). Add a `require` (pinned to a
`dao-utils/vX.Y.Z` tag) and import the package:

```go
import dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"

cfg := &dao_utils.DataBaseConfig{ /* or load from env */ }
db := dao_utils.Init(cfg)
```

Versions are released by tagging the repo `dao-utils/vX.Y.Z`; see the
[repository README](../README.md) and `make package-bump-version`.

## Public API

Key entry points: `Init`, `DB`, `DBRead`, `Migrate`, `ConnectionFactory`,
`BaseModel`, `DeleteBaseModel`, `PGLock`, `CronJob`, `InitCleaner`, `JsonBMap`,
`ValidateAndParameterizeApiFilters`/`ApplyParameterizedFilters`,
`ApplySortByParam`, `ApplyTSSearchParam`, and the `validators` sub-package
(`CreateFieldConfigMap`, `NewUUIDFieldValidation`, `NewDateFieldValidation`, …).

## Building and testing

This is a library, so `make build` and `make validate` are intentionally no-ops
(no binary, no image). Everything is driven through the [`Makefile`](Makefile),
which includes the shared `../../build/makefile/index.mk`.

| Target | Description |
|--------|-------------|
| `make test` | Run the unit tests (`go test ./...`); some suites start PostgreSQL via testcontainers and need a working Docker daemon |
| `make lint` | Run the Go linter (golangci-lint) against `../.golangci.yaml` |
| `make purify` | `go fmt ./...`, `go vet ./...`, and `go mod tidy` |
| `make build` | No-op (library) |
| `make validate` | No-op recipe, but the shared makefile adds package import/replace checks as prerequisites |
| `make package-bump-version` | Tag a production release `dao-utils/vX.Y.Z` |
| `make package-bump-version-dev` | Tag a dev release `dao-utils/vX.Y.Z-dev.N` |

## Configuration

`DataBaseConfig` can be populated programmatically or loaded from environment
variables (keys are the uppercase equivalents of the struct's config keys). The
main variables (see `config.go`):

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` / `DB_PORT` | Primary (writer) host and port | — |
| `DB_READ_HOST` / `DB_READ_PORT` | Read-replica host and port | — |
| `DB_NAME` | Database name | — |
| `DB_SCHEMA` | Schema name | — |
| `DB_USER` / `DB_PASSWORD` | Credentials (loaded as secrets) | — |
| `DB_SSL_MODE` | One of `disable`, `allow`, `prefer`, `require`, `verify-ca`, `verify-full` | — |
| `SERVICE_NAME` | Owning service (used for the per-service DB session/user) | — |
| `DB_USE_SERVICE_SESSION` | Use the per-service DB user | — |
| `STATEMENT_TIMEOUT` | Postgres `statement_timeout` | — |
| `MAX_CONNECTIONS` / `MAX_IDLE_CONNECTIONS_TIME_DURATION` | Connection-pool tuning | — |
| `SKIP_DEFAULT_TRANSACTION` / `FULL_SAVE_ASSOCIATIONS` | GORM behavior flags | — |
| `MIGRATION_FILES` | golang-migrate source (e.g. `file:///migrations`) | — |
| `MIGRATION_COMMAND` | Migration command/version | — |
| `MIGRATION_LOCK_RETRIES` | Advisory-lock retry count | `3` |
| `SECRETS_MOUNT_PATH` | Directory secrets are mounted from | — |
