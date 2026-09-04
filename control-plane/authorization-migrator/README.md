# authorization-migrator

`authorization-migrator` is a **run-once database migration job**, not a
long-running service. It moves all authorization-related data from the legacy
`2.11` layout (role assignments stored in the `backend` schema) into the modern
`authorization` schema used by `2.12` and later. The job reads from the source
schemas, translates each legacy role assignment into the new authorization
model, and writes the result into the target schema. **It never mutates the old
data** — the legacy tables are read-only sources, so re-running against an
already-migrated database is safe. It is owned by the **Suricata** team.

The job is deployed by the control-plane Helm chart and executes once as part of
the install/upgrade flow. On completion it records a row in a
`migration_control` table in the target schema so subsequent runs short-circuit
(see [Idempotency](#idempotency)).

## Contents

- [Overview](#overview)
- [What it migrates](#what-it-migrates)
- [Idempotency](#idempotency)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)

## Overview

The job is responsible for:

- **Reading legacy role assignments** — it runs a single wide `SELECT` over the
  legacy `backend` schema (`auth_user`, `roles`, `department_admins`,
  `departments`, `projects`, `projects_users_permissions`, `cluster_permission`,
  `clusters`, `tenants`, …), joining users to their roles and to the department,
  project and cluster scopes each role applies to.
- **Resolving tenants** — it reads the `tenants_manager` schema to map each
  legacy tenant to its realm/auth information.
- **Translating to the new model** — each legacy row is converted into the
  `authorization` schema's role-assignment representation (using the shared
  `authorization` service's static-roles / static-legacy-roles DAOs and its
  generated API types), then written to the target schema.
- **Keycloak coordination** — a Keycloak admin client is initialised (host,
  admin username/password) and a refresh-token ticker keeps the admin token
  alive for the duration of the migration.
- **Recording completion** — after a successful pass the job marks the migration
  version as done in the target schema.

If the source tables do not exist (Postgres error `42P01`, "undefined table"),
the job logs that migration is unnecessary and exits cleanly — this is the
expected path on a fresh install with no `2.11` data.

## What it migrates

The unit of migration is a **role assignment**: which user holds which role over
which scope (all-clusters, a specific cluster, a department, or a project) within
a tenant. The legacy representation spread this across many tables; the new
`authorization` schema stores it in a normalized form. The migrator's `SELECT`
flattens the legacy graph (one row per user × role × scope) and the
`migration` package rewrites each flattened row into the target schema.

## Idempotency

The `version_control` package `AutoMigrate`s a `MigrationControl` table
(primary key `version`) in the **target** schema. On startup the job checks
whether a row for the configured `DB_MIGRATION_VERSION` already exists:

- present → `Migrate done already, finishing.` and the job exits without touching
  data;
- absent → the migration runs and, on success, inserts the version row.

To **force a re-run** (as the operations runbook describes): delete the job
(`kubectl delete job runai-backend-authorization-migrator`), bump the
`DB_MIGRATION_VERSION` value (e.g. `v1` → `v2`) so the new version is treated as
not-yet-done, and re-apply the manifest.

## Repository layout

```text
.
├── cmd/                    # Job entry point (main.go) — init + Migrate + mark-done
├── config/                 # Config struct, defaults, and mounted secrets/
│   └── secrets/            # Secret keys read as files (DB_USER, DB_PASSWORD, KEYCLOAK_ADMIN_*)
├── database/               # db.go — opens a *gorm.DB scoped to a schema name
├── keycloak/               # client.go — Keycloak admin client + refresh-token ticker
├── migration/              # migration.go + utils.go — the legacy→new SQL and row translation
├── version_control/        # control.go — MigrationControl table (idempotency marker)
├── Dockerfile
├── Makefile
└── go.mod
```

## Building and running

The job is built with Go (see [`go.mod`](go.mod)) and orchestrated through the
[`Makefile`](Makefile), which only sets `SERVICE_NAME=authorization-migrator`
and includes the shared `../../build/makefile/control-plane.mk` and
`index.mk` (which pulls in `golang.mk`). The service `Makefile` maps its
top-level targets onto the shared Go recipes:

| Target | Description |
|--------|-------------|
| `make build` | `build-go` — compile the amd64 + arm64 binaries (`bin/authorization-migrator-<arch>`) in the Go builder container (FIPS `GOFIPS140=v1.0.0`) |
| `make test` | `test-go` — run Go unit tests in the builder container |
| `make lint` | `lint-go` — golangci-lint in the builder container |
| `make purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `make validate` | no-op (kept for interface parity across CP services) |

Shared Go targets available from the includes include `build-go-amd`,
`build-go-arm`, `format-go` and `coverage`.

The [`Dockerfile`](Dockerfile) is a two-stage build: a
`registry.access.redhat.com/ubi9/ubi-minimal` `base` stage supplies
`/etc/passwd`, `/etc/group`, TLS certs (`/etc/ssl`, `/etc/pki`) and a writable
`/tmp`; the final image is `FROM scratch`, copies in
`bin/authorization-migrator-$TARGETARCH` (build arg `TARGETARCH`) as
`authorization_migrator`, runs as user `65532:65532`, and sets the binary as the
`ENTRYPOINT`.

```bash
docker build --build-arg TARGETARCH=amd64 -t authorization-migrator:dev .
```

### Local build (local-dev.sh)

`authorization-migrator` is in `CP_GO_SERVICES`, so `local-dev.sh` builds it via
its `build_go_service` path:

1. Compile the amd64 binary in the builder container:
   ```bash
   make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
   ```
   (`DOCKERHUB_PROXY=` pulls the Go builder image from Docker Hub;
   `GITHUB_TOKEN=` is emptied here — the private-module token is applied
   elsewhere.)
2. Build + load the image:
   ```bash
   docker buildx build --platform linux/amd64 --provenance=false --load \
     --build-arg TARGETARCH=amd64 -t runai-local/authorization-migrator:0.0.0 .
   ```

Image ref: `runai-local/authorization-migrator:0.0.0` — tag `0.0.0` is the
control-plane chart appVersion; any other tag breaks the local deploy. To iterate
on control-plane only:

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```

## Testing

| Target | Description |
|--------|-------------|
| `make test` | `test-go` — build and run the Go unit tests (e.g. `config/config_test.go`) |

## Configuration

The job is **environment-variable driven**. Configuration keys map one-to-one to
uppercase environment variables (e.g. `db_host` → `DB_HOST`); secret keys are
additionally read from files mounted under `config/secrets/` (one file per key).
Values are validated as *required* at startup — a missing required key is fatal.

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | — |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_NAME` | Database name | — |
| `DB_SCHEMA_SOURCE_BACKEND` | Legacy source schema (role data) | `backend` |
| `DB_SCHEMA_SOURCE_TENANTS` | Legacy tenants source schema | `tenants_manager` |
| `DB_SCHEMA_TARGET` | Target schema to write into | `authorization` |
| `DB_MIGRATION_VERSION` | Idempotency marker; bump to force a re-run | `v1` |
| `DB_SSL_MODE` | Postgres SSL mode (falls back to `disable` if invalid) | `disable` |
| `DB_USER` | Postgres user (secret; file under `config/secrets/`) | — |
| `DB_PASSWORD` | Postgres password (secret) | — |
| `KEYCLOAK_HOST` | Keycloak base URL | `http://keycloak-http:80` |
| `KEYCLOAK_ADMIN_USERNAME` | Keycloak admin user (secret) | — |
| `KEYCLOAK_ADMIN_PASSWORD` | Keycloak admin password (secret) | — |
