# integdb

`integdb` is a shared Go test-support library in the Run.ai monorepo. It spins up a
disposable, migration-ready PostgreSQL instance for integration tests using
[testcontainers-go](https://golang.testcontainers.org/), and hands back a fully
populated [`runaidb`](../runaidb) `Config` plus a cleanup function. It exists so
that every service's integration suite gets an identical, `pg_cron`-enabled
Postgres without re-implementing container setup. It is consumed as a versioned Go
module — there is no service binary and no published container image (the bundled
`Dockerfile` is only a testcontainers build context).

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Importing the library](#importing-the-library)
- [Public API](#public-api)
- [Building and testing](#building-and-testing)
- [Configuration](#configuration)

## Overview

The package is responsible for:

- **Provisioning Postgres** — `SetupPostgresDB(*IntegdbConfig)` runs a
  `postgres:<version>` container (via the testcontainers `postgres` module) with
  `pg_cron` preloaded (`shared_preload_libraries=pg_cron`), waits until it is
  ready, and resolves the mapped host/port.
- **Building the DB config** — it returns a `*runaidb/config.Config` with the
  connection details, schema, disabled SSL, auto-migration enabled when a
  migration directory is provided, and a randomly generated encryption key — so
  callers can hand it straight to `runaidb`.
- **Cleanup** — returns a `func()` that terminates the container; call it via
  `defer` at the end of the test.
- **pg_cron image** — when a container spins up, `integdb` builds a custom image
  from the bundled [`Dockerfile`](Dockerfile), which extends the official
  `postgres` image with the `postgresql-<major>-cron` package so scheduled-job
  tests work.

## Repository layout

```text
.
├── integdb.go        # SetupPostgresDB, IntegdbConfig, container/image setup
├── integdb_test.go   # Tests
├── Dockerfile        # postgres + pg_cron (testcontainers build context)
├── Makefile
└── go.mod
```

## Importing the library

The module path comes from [`go.mod`](go.mod). Add a `require` pinned to an
`integdb/vX.Y.Z` tag and call it from a test's setup:

```go
import "github.com/dsx-ai-factory/run-ai/runai-common-packages/integdb"

dbConf, cleanup, err := integdb.SetupPostgresDB(nil) // nil → load defaults / env
if err != nil {
    t.Fatal(err)
}
defer cleanup()
// hand dbConf to runaidb, run migrations, run the suite …
```

## Public API

- **`SetupPostgresDB(conf *IntegdbConfig) (*dbconfig.Config, func(), error)`** —
  the single entry point. Pass `nil` to load `IntegdbConfig` from defaults/env, or
  pass a struct to override values.
- **`IntegdbConfig`** — the input struct (see [Configuration](#configuration)).

## Building and testing

This is a test-support library, so `make build` and `make validate` are
intentionally no-ops (no service binary, no published image). Everything is driven
through the [`Makefile`](Makefile), which includes the shared
`../../build/makefile/index.mk`.

| Target | Description |
|--------|-------------|
| `make test` | Run unit tests via the shared `test-go` target — requires a working Docker daemon (it starts a Postgres container) |
| `make lint` | Run the Go linter (golangci-lint) against `../.golangci.yaml` |
| `make purify` | `go fmt ./...`, `go vet ./...`, and `go mod tidy` |
| `make build` | No-op (library) |
| `make validate` | No-op recipe, but the shared makefile adds package import/replace checks as prerequisites |
| `make package-bump-version` | Tag a production release `integdb/vX.Y.Z` |
| `make package-bump-version-dev` | Tag a dev release `integdb/vX.Y.Z-dev.N` |

## Configuration

`IntegdbConfig` is populated programmatically or from environment variables (the
uppercase equivalents of the config keys). All fields have defaults, so `nil`
works out of the box:

| Variable | Description | Default |
|----------|-------------|---------|
| `INTEGDB_USER` | Postgres superuser | `postgres` |
| `INTEGDB_PASSWORD` | Postgres password | random 32-char hex, regenerated per run |
| `INTEGDB_NAME` | Database name | `backend` |
| `INTEGDB_SCHEMA` | Schema to use (required) | `test` |
| `INTEGDB_MIGRATION_DIR_PATH` | golang-migrate source; empty disables auto-migration | `file:///migrations` |
| `INTEGDB_IMAGE_VERSION` | Postgres image tag (major version drives the `pg_cron` package) | `16` |

Read the effective password from `dbConf.Connection.Password`, or set
`INTEGDB_PASSWORD` to pin it.

Releases are cut by tagging the repo `integdb/vX.Y.Z`; see the
[repository README](../README.md) for the versioning workflow.
