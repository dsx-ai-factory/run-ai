# cluster-migrator

`cluster-migrator` is a **run-once database migration job** in the Run:ai
control plane, owned by the **Griffin** team. It is not a long-running service
and exposes **no HTTP API**. It copies each tenant's node-pool system settings out
of the legacy `backend` schema and onto that tenant's node pools in the
`cluster_service` schema. The source data is only read — it is never modified —
so the job is safe to re-run and to roll back to.

The job is executed by the control-plane Helm chart during install/upgrade (as a
Kubernetes `Job`, `runai-backend-cluster-migrator`). It runs after the target
schemas exist and before the services that own those schemas begin serving, so
that upgraded tenants come up with their node-pool system settings already
present.

## Contents

- [Overview](#overview)
- [What it migrates](#what-it-migrates)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [Re-running the job](#re-running-the-job)

## Overview

The job connects to a single PostgreSQL database and opens three GORM
connections against different schemas: the **source** schema (`backend`, the
legacy data), the **target** schema (`cluster_service`, the destination),
and a schema-less "global" connection. On start it auto-migrates its own
bookkeeping table (`NodepoolSystemSettingsMigrationControl`) in the target
schema; it records which tenants have already been migrated so the job is
idempotent.

`cmd/main.go` then runs the migration. It first checks the version-control
marker and exits early if the migration is already done, so re-running the job
is a no-op once it has succeeded.

## What it migrates

- **Node-pool system settings** (`nodepoolsystemsettings/migration`) — reads
  each tenant's `project.quota.cpu_resources` and
  `project.quota.manual_over_quota` from `backend.tenant_settings` and writes
  them to the `allow_cpu_quota` and `allow_adjustable_weights` columns of every
  node pool the tenant has. Those columns come from cluster-service's 2.24.1
  schema migration, so the job first waits up to 10 minutes for them to exist.
  A failure here is fatal so the job surfaces incomplete migrations loudly.

Progress is tracked per tenant in a control table, so tenants already migrated
are skipped on the next run.

The 2.15 → 2.16 migrations that copied clusters and node pools out of the
`backend` schema were removed once those versions were no longer supported.

## Repository layout

```text
.
├── cmd/                        # Job entry point (main.go)
├── config/                     # ConfigData struct + defaults (env/config driven)
├── database/                   # GORM connections, DAOs, and models
│   └── models/                 # Tenant, tenant-setting and migration-control models
├── nodepoolsystemsettings/
│   └── migration/              # Node-pool system-settings migration logic (+ tests)
├── Dockerfile
├── Makefile
└── go.mod
```

## Building and running

The project is built with Go (see [`go.mod`](go.mod)) and orchestrated through
the [`Makefile`](Makefile), which includes the shared
`../../build/makefile/control-plane.mk` and `../../build/makefile/index.mk`
(which pull in `golang.mk`).

| Target | Description |
|--------|-------------|
| `make build` | `build-go` — compile the amd64 + arm64 binaries in the Go builder container |
| `make test` | `test-ginkgo` — run the Ginkgo unit tests |
| `make lint` | `lint-go` — run golangci-lint in the builder container |
| `make purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `make validate` | no-op (kept for interface parity across CP services) |
| `make local-deploy` | `build-go` + `docker-build-generic` + `docker-push-generic` (OS=linux, ARCH=amd64) |

Under the hood `build-go` runs `build-go-amd` + `build-go-arm`, each compiling
`cmd/main.go` for `GOOS=linux` with FIPS mode (`GOFIPS140=v1.0.0`) inside the
`golang` builder container, producing `bin/cluster-migrator-<arch>`.

The [`Dockerfile`](Dockerfile) is a two-stage build: a
`registry.access.redhat.com/ubi9/ubi-minimal` stage supplies `/etc/passwd`,
`/etc/group`, certs and a writable `/tmp`; the final image is `FROM scratch`,
copies in `bin/cluster-migrator-$TARGETARCH` as `cluster_migrator`, runs as user
`65532:65532`, and sets it as the entrypoint. It takes one build arg,
`TARGETARCH`, which selects the binary to copy.

```bash
docker build --build-arg TARGETARCH=amd64 -t cluster-migrator:dev .
```

### Local build (local-dev.sh)

`cluster-migrator` is listed in `CP_GO_SERVICES` in
[`local-dev.sh`](../../local-dev.sh), so the local build runs `build_go_service`,
which does exactly two steps:

1. Compile the amd64 binary in the builder container:
   ```bash
   make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
   ```
   (`DOCKERHUB_PROXY=` pulls the Go builder image straight from Docker Hub;
   `GITHUB_TOKEN=` is emptied in the make call — the private-module token is
   applied elsewhere.)
2. Build and load the image:
   ```bash
   docker buildx build --platform linux/amd64 --provenance=false --load \
     --build-arg TARGETARCH=amd64 -t runai-local/cluster-migrator:0.0.0 .
   ```

The tag **must** be `0.0.0` (the control-plane chart appVersion the Helm chart
references) — any other tag breaks the local deploy.

To iterate on just the control plane and reuse an existing cluster:

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```

## Testing

| Target | Description |
|--------|-------------|
| `make test` | Runs the Ginkgo suite under `nodepoolsystemsettings/migration` |
| `make lint` | golangci-lint |

The suite starts a throwaway Postgres container through `integdb`
(testcontainers) and applies cluster-service's migrations to it, so Docker must
be running.

## Configuration

The job is configuration-driven: it reads a set of keys (with sensible
defaults) that can each be overridden by their uppercase environment-variable
equivalent (e.g. the `db_migration_version` key is set via `DB_MIGRATION_VERSION`).
Database connection settings come from the shared DAO config.

**Migration keys** (from [`config/config.go`](config/config.go)):

| Key / env var | Description | Default |
|---|---|---|
| `db_migration_version` / `DB_MIGRATION_VERSION` | Migration version marker; bump to force a re-run (e.g. `v2`) | `v1` |
| `db_schema_source_backend` / `DB_SCHEMA_SOURCE_BACKEND` | Legacy source schema | `backend` |
| `db_schema_target` / `DB_SCHEMA_TARGET` | Destination schema | `cluster_service` |
| `should_mark_completed` / `SHOULD_MARK_COMPLETED` | After the run, record a global completed marker so later runs skip the migration | `false` |
| `tenants_to_migrate` / `TENANTS_TO_MIGRATE` | Restrict to these tenant ids | `[]` (all) |
| `tenants_to_not_migrate` / `TENANTS_TO_NOT_MIGRATE` | Exclude these tenant ids | `[]` |

**Database keys** (shared DAO config) — `DB_HOST`, `DB_PORT`, `DB_NAME`,
`DB_USER`, `DB_PASSWORD`, `DB_SSL_MODE`, and related tuning keys. All three
connections (source, target, global) target the same database and differ only by
schema `search_path`.

## Re-running the job

Because each phase is idempotent, simply re-applying the job after a successful
run does nothing. To force a re-run, delete the old job and bump the version:

```bash
kubectl delete job runai-backend-cluster-migrator
```

Set `DB_MIGRATION_VERSION` to a new value (e.g. `v2`) in the job manifest and
re-apply it. The new version marker will not match the recorded completion
marker, so the migration will run again.
