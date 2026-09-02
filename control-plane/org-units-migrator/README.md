# org-units-migrator

`org-units-migrator` is a **run-once database migration job** (not a long-running
service) owned by the **Suricata** team. It back-fills organizational-unit data
from the legacy `backend` schema into the `org_unit_service` schema introduced in
Run:AI 2.18, so that the new `org-unit-service` becomes the source of truth for
departments, projects and their node-pool/quota settings. It reads from the old
schema, translates each record into the new model, and writes it into the target
schema **without mutating the source data**. It runs per-tenant and is idempotent:
completed tenants are skipped on re-runs.

## Contents

- [Overview](#overview)
- [What it migrates](#what-it-migrates)
- [When it runs](#when-it-runs)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [Re-running the job](#re-running-the-job)

## Overview

Historically (≤ 2.17) departments and projects lived in the monolithic `backend`
database schema. Starting in 2.18 these "org units" are owned by the dedicated
`org-unit-service`, which stores them in its own `org_unit_service` schema. This
job performs the one-time data move during upgrade so that no organizational data
is lost when the new service takes over.

The entry point ([`cmd/main.go`](cmd/main.go)) does the following:

1. Initializes logging, configuration, and three database connections — one to the
   **source** schema (`backend`), one to the **target** schema (`org_unit_service`),
   and one schema-less **global** connection.
2. Short-circuits if the migration was already marked completed globally.
3. Resolves the set of tenants to migrate (all tenants, minus/limited by the
   allow/deny lists in configuration).
4. Runs the migration for each tenant, tracking progress in a `migration_control`
   table (`STARTED` → `COMPLETED` / `FAILED`) so partial runs are resumable.
5. Marks the migration globally completed (only when `should_mark_completed` is set)
   and exits `0`.

## What it migrates

For every eligible tenant, [`migration/migrate.go`](migration/migrate.go) reads
the tenant's data from the source schema and recreates it in the target schema, in
order:

- **Settings** — the tenant's org-unit-related settings map, used to shape the
  translated departments and projects.
- **Departments (v1)** — legacy department records → new department model.
- **Node types** — allowed node-type definitions attached to org units.
- **Projects (v1)** — legacy project records → new project model.

Per-tenant status is written to `migration_control` keyed by
`(version, tenant_id)`. A tenant already at `COMPLETED` for the configured
`db_migration_version` is skipped, making the whole job safe to re-run.

## When it runs

The migration is executed as a Kubernetes **Job** deployed by the control-plane
Helm chart during install/upgrade (job name `runai-backend-org-units-migrator`).
It runs once per upgrade to 2.18+, translating each tenant's data before the new
`org-unit-service` serves traffic. Because it only reads the legacy schema and
writes the new one, the old data remains intact as a fallback.

## Repository layout

```text
.
├── cmd/                    # Entry point (main.go) — the migration job
├── config/                 # ConfigData struct + defaults (env-driven)
├── database/               # GORM DAOs for source/target schemas
│   ├── db.go               # Opens source/target/global DB connections
│   ├── departments_dao.go  # Read legacy departments
│   ├── projects_dao.go     # Read legacy projects
│   ├── node_types_dao.go   # Read node types
│   ├── settings_dao.go     # Read tenant settings
│   ├── tenants_dao.go      # Resolve tenants to migrate
│   ├── version_dao.go      # migration_control read/write (status tracking)
│   └── models/             # Source & target row models
├── migration/              # Translation logic (departments, projects, node types)
├── keycloak/               # Keycloak-related helpers
├── version_control/
├── Dockerfile
├── Makefile
└── go.mod
```

## Building and running

The project is built with Go (see [`go.mod`](go.mod)) and orchestrated through the
[`Makefile`](Makefile), which includes the shared control-plane build logic from
`../../build/makefile/control-plane.mk` and `../../build/makefile/index.mk`
(pulling in `golang.mk`).

| Target | Description |
|--------|-------------|
| `make build` | `build-go` — compile the amd64 + arm64 binaries |
| `make build-go` | `build-go-amd` + `build-go-arm` |
| `make build-go-amd` | Compile `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) in the Go builder container → `bin/org-units-migrator-amd64` |
| `make build-go-arm` | Same for `arm64` → `bin/org-units-migrator-arm64` |
| `make test` | `test-ginkgo` — run the Ginkgo unit/DAO tests |
| `make lint` | `lint-go` — golangci-lint in the builder container |
| `make purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `make validate` | No-op (this service has no OpenAPI spec) |
| `make local-deploy` | `build-go` + `docker-build-generic` + `docker-push-generic` (linux/amd64) |

The [`Dockerfile`](Dockerfile) is a two-stage build: a
`registry.access.redhat.com/ubi9/ubi-minimal` `base` stage supplies `/etc/passwd`,
`/etc/group`, TLS certs (`/etc/ssl`, `/etc/pki`), `nsswitch.conf` and a writable
`/tmp`; the final image is `FROM scratch`, copies the arch-selected binary
(`bin/org-units-migrator-$TARGETARCH`) as `/org_units_migrator`, runs as user
`65532:65532`, and sets it as the entrypoint. Build arg: `TARGETARCH`.

```bash
docker build --build-arg TARGETARCH=amd64 -t org-units-migrator:dev .
```

### Local build (local-dev.sh)

This service is listed in `CP_GO_SERVICES`, so `local-dev.sh` builds it via
`build_go_service` in two steps:

1. Compile the amd64 binary in the builder container:
   ```bash
   make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
   ```
   (`DOCKERHUB_PROXY=` pulls the Go builder image from Docker Hub; `GITHUB_TOKEN=`
   is emptied in the make call — the private-module token is applied elsewhere.)
2. Build + load the image:
   ```bash
   docker buildx build --platform linux/amd64 --provenance=false --load \
     --build-arg TARGETARCH=amd64 -t runai-local/org-units-migrator:0.0.0 .
   ```

Image ref: `runai-local/org-units-migrator:0.0.0` (tag `0.0.0` — the CP chart
appVersion; any other tag breaks the local deploy). To rebuild the control plane
and redeploy against an existing cluster:

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```

## Testing

| Target | Description |
|--------|-------------|
| `make test` | Runs the Ginkgo suites under `migration/` and `database/` (`test-ginkgo`) |
| `make lint` | golangci-lint |

The tests use Ginkgo/Gomega (`ginkgo -r`) and exercise the per-resource migration
and DAO logic (`departments_test.go`, `projects_test.go`, `node_types_test.go`,
`settings_test.go`, `tenants_test.go`, `version_test.go`, `migrate_test.go`).

## Configuration

Configuration is **environment-variable driven** — the config keys are the
uppercase form of the fields below, and the database connection is configured with
the standard `DB_*` variables. The job takes no config file.

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_SCHEMA_SOURCE` | Source schema to read legacy org-unit data from | `backend` |
| `DB_SCHEMA_TARGET` | Target schema to write the new model into | `org_unit_service` |
| `DB_MIGRATION_VERSION` | Migration version recorded in `migration_control` | `v1` |
| `SHOULD_MARK_COMPLETED` | Whether to mark the migration globally completed after success | `false` |
| `TENANTS_TO_MIGRATE` | Explicit allow-list of tenants (empty = all) | `[]` |
| `TENANTS_TO_NOT_MIGRATE` | Deny-list of tenants to skip | `[]` |
| `DB_HOST` / `DB_PORT` | PostgreSQL host / port | — |
| `DB_NAME` | Database name | — |
| `DB_USER` / `DB_PASSWORD` | Credentials (provided via mounted secret) | — |
| `DB_SSL_MODE` | PostgreSQL `sslmode` | — |

## Re-running the job

The job is idempotent — completed tenants are skipped. To force a fresh run, delete
the existing Job and re-apply it. To advance to a new migration version, bump the
`DB_MIGRATION_VERSION` environment variable (e.g. to `v2`) before re-applying:

```bash
kubectl delete job runai-backend-org-units-migrator
# set DB_MIGRATION_VERSION=v2 in the job manifest, then apply
```
