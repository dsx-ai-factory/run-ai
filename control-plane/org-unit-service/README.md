# org-unit-service

`org-unit-service` is the control-plane service that owns Run:AI's **organizational
unit** hierarchy — **departments** and **projects** — together with the resource
quotas, node-pool allocations, over-time usage metrics and settings attached to
them. It is a Go HTTP/REST service (owned by the **Suricata Team**) whose API is
generated from an OpenAPI specification, backed by its own PostgreSQL schema
(`org_unit_service`), and integrated with the rest of the platform via the shared
pub/sub bus and the clients of neighbouring services (tenants-manager,
cluster-service, workloads, datavolumes, temp-backend-settings, authorization,
audit).

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [OpenAPI and code generation](#openapi-and-code-generation)
- [Go client](#go-client)

## Overview

The service is responsible for:

- **Organizational units** — creating, updating, listing and deleting
  **departments** and **projects**, including the parent/child relationships
  between them and their per-node-pool resource allocations.
- **Resource quotas & node-pool resources** — tracking deserved / limit / over-
  subscription GPU, CPU and memory quotas per org unit and per node pool, and
  the database views that aggregate them.
- **Over-time metrics** — a background writer that periodically queries usage
  metrics over 24h / 7d / 30d ranges (rate-limited) and persists them for
  reporting.
- **Reports** — generating org-unit usage reports and uploading them to an S3
  bucket.
- **Settings** — org-unit-scoped settings such as the limit-quota
  over-subscription toggle, migrated on start-up through the shared setting
  package.
- **Resource sync & cleanup** — a helper job that periodically syncs org-unit
  resources against cluster state and a DB cleaner that hard-deletes
  soft-deleted rows after a retention window.

The process runs in two modes selected at runtime: the **HTTP server** (default)
and a **helper job** (`ShouldRunAsJob()`) that runs the over-time metrics writer,
the resource-sync goroutines and the DB cleaner instead of serving HTTP.

The API surface is described in [`swagger/source/openapi.yaml`](swagger/source/openapi.yaml)
and bundled into [`swagger/_build/swagger.yaml`](swagger/_build/swagger.yaml).

## Repository layout

```text
.
├── cmd/                        # Service entry point (main.go) — server + helper-job modes
├── pkg/
│   ├── apis/                   # API surface wiring
│   ├── clients/                # HTTP clients for upstream backend services
│   ├── config/                 # Typed configuration sections (registered on init)
│   ├── database/               # DAO layer, cleaner, and SQL migrations (up/ + down/)
│   ├── errors/                 # Domain error types
│   ├── helper_job/             # Helper-job mode detection
│   ├── overtime_metrics_writer/# Background over-time metrics writer
│   ├── pubsub/                 # Pub/sub integration
│   ├── resources_sync/         # Org-unit ↔ cluster resource sync goroutines
│   ├── router/                 # HTTP router wiring
│   ├── runaimetrics/           # Metrics helpers
│   ├── services/               # API service implementations (departments, projects, …)
│   ├── settings/               # Typed setting definitions
│   └── test/                   # Test helpers
├── generated/                  # OpenAPI-generated server stubs (do not edit)
│   ├── api/
│   └── pkg/
├── swagger/                    # OpenAPI source and built bundle
│   ├── source/                 # openapi.yaml + parameters/ requests/ schemas/ securitySchemes/
│   └── _build/                 # Bundled swagger.yaml / .json / .html
├── Dockerfile
├── Makefile
└── go.mod
```

The versioned Go **client** for this service lives in a separate module,
[`control-plane/org-unit-service/pkg/client`](pkg/client),
not under `pkg/client/`.

## Building and running

The project is built with Go (see [`go.mod`](go.mod); module
`github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service`) and orchestrated through
the [`Makefile`](Makefile), which includes the shared
`../../build/makefile/control-plane.mk` and `../../build/makefile/index.mk`.

| Target | Description |
|--------|-------------|
| `make build` | `openapi-build` then `build-go` — validate/bundle the OpenAPI spec, regenerate server code, then compile the binary for amd64 and arm64 |
| `make build-go-amd` | Compile the linux/amd64 binary in Docker → `bin/org-unit-service-amd64` (FIPS build) |
| `make build-go-arm` | Compile the linux/arm64 binary in Docker → `bin/org-unit-service-arm64` |
| `make openapi-build` | Bundle the spec, build HTML/JSON docs, publish the public spec and generate server code |
| `make validate` | `openapi-validate` — validate the OpenAPI spec, generated code and public routes |
| `make openapi-lint` | Lint the swagger bundle |
| `make lint` | Run the Go linter (golangci-lint, shared config in `../infra/lint/.golangci.yaml`) |
| `make purify` | Auto-fix Go formatting and imports |
| `make local-deploy` | Generate code, build the amd64 binary, `docker build` and `docker push` the image |

The container image is built from [`Dockerfile`](Dockerfile), a two-stage build:
a `registry.access.redhat.com/ubi9/ubi-minimal` **base** stage that provides the
`passwd`/`group` entries, CA/PKI trust store, `nsswitch.conf` and a writable
`/tmp` (mode 0770, owned `65532:0`); and a final `scratch` stage that copies those
plus the compiled `bin/org-unit-service-$TARGETARCH` binary, the `NOTICE` file, the
DB migrations (`migrations/up`, `migrations/down`) and the bundled swagger, runs as
non-root user `65532:65532`, and sets `ENTRYPOINT ["./org-unit-service"]`. The only
build arg is `TARGETARCH`:

```bash
docker build --build-arg TARGETARCH=amd64 -t org-unit-service:dev .
```

### Local build (local-dev.sh)

`local-dev.sh` builds this service through its generic control-plane Go path
(`build_go_service`). For each service it runs, from the service directory:

```bash
make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
```

then builds and loads the image:

```bash
docker buildx build --build-arg TARGETARCH=amd64 -t runai-local/org-unit-service:0.0.0 .
```

Notes:
- Only the **amd64** binary is compiled locally (`build-go-amd`, not the full
  `make build`); the image tag is the fixed control-plane version **`0.0.0`**
  (the CP Helm chart's `appVersion`), producing `runai-local/org-unit-service:0.0.0`.
  Any other tag breaks the local deploy.
- `DOCKERHUB_PROXY=` (empty) pulls the golang builder image straight from Docker
  Hub; `GITHUB_TOKEN=` is passed empty in the `make` call (the private-module
  token is applied elsewhere in the script).

To iterate on the control plane against an already-running local cluster:

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```

## Testing

| Target | Description |
|--------|-------------|
| `make test` | `test-ginkgo` — run the unit-test suites via Ginkgo (recursive, randomized, 5m timeout) |

Tests are written with [Ginkgo](https://onsi.github.io/ginkgo/) / Gomega; the
`test` target installs the pinned Ginkgo/Gomega versions, runs `go mod tidy` and
executes every suite in the module. There is no docker-compose integration-test
harness in this service.

## Configuration

The service is **config-file driven**. On start-up it registers a set of typed
configuration sections (`config.Register[...]`) whose fields are keys read from a
configuration file supplied at deploy time (the control-plane Helm chart mounts
it); each section defines its own defaults. **Any key can be overridden by its
uppercase environment-variable equivalent** (e.g. the `service_port` key is
overridden by `SERVICE_PORT`). Keys are flat (not nested under a section name).

There is no checked-in sample file in this service; the following representative
config reproduces the registered keys and their in-code defaults:

```yaml
# General
log_level: info

# Server — listen port and upstream service URLs
service_port: 8080
temp_backend_settings_url: http://runai-backend-temp-backend-settings:8080
tenants_manager_url:       http://runai-backend-tenants-manager:8080
cluster_service_url:       http://runai-backend-cluster-service:8080
workloads_service_url:     http://runai-backend-workloads:8080
datavolumes_service_url:   http://runai-backend-datavolumes:8080

# Reports — S3 bucket for generated usage reports (no default)
s3_bucket: ""

# OvertimeMetrics — background usage-metrics writer
query_24h_range_interval_seconds:      3600      # 1h
query_7d_range_interval_seconds:       21600     # 6h
query_30d_range_interval_seconds:      86400     # 24h
query_24h_range_initial_sleep_seconds: 10
query_7d_range_initial_sleep_seconds:  21
query_30d_range_initial_sleep_seconds: 33
rate_limiter_limit: 1
rate_limiter_burst: 10

# SyncJob — org-unit ↔ cluster resource sync
sync_job_interval_hours:          24
sync_job_interval_tenant_seconds: 5

# DBCleanerConfig — hard-delete of soft-deleted rows
cleaner_enabled:          true
cleaner_interval_hours:   24
soft_deleted_days_to_live: 30

# Database (shared runaidb keys; db_user / db_password are secrets)
db_host: ""
db_port: ""
db_name: ""
db_schema: org_unit_service
db_user: ""
db_password: ""
db_ssl_mode: ""
db_max_connections: 0
db_migration_files: ""
```

Top-level sections:

- **General** — `log_level`.
- **Server** — the HTTP listen port and the base URLs of every upstream service
  this service calls (temp-backend-settings, tenants-manager, cluster-service,
  workloads, datavolumes).
- **Reports** — the S3 bucket that generated reports are uploaded to.
- **OvertimeMetrics** — query intervals, initial sleeps and rate-limiter settings
  for the over-time metrics writer.
- **SyncJob** — intervals for the resource-sync helper job.
- **DBCleanerConfig** — whether the DB cleaner runs, how often, and the retention
  window before soft-deleted rows are hard-deleted.
- **Database** — connection, pool and migration settings, provided through the
  shared `runaidb` config (`db_host`, `db_port`, `db_name`, `db_schema`,
  `db_user`, `db_password`, `db_ssl_mode`, `db_max_connections`, …). `db_user` and
  `db_password` are treated as secrets. The service's schema is
  `org_unit_service` and migrations ship in `pkg/database/migrations/{up,down}`.

## OpenAPI and code generation

The HTTP API is defined in [`swagger/source/openapi.yaml`](swagger/source/openapi.yaml),
split across `swagger/source/parameters`, `requests`, `schemas` and
`securitySchemes`, and bundled into `swagger/_build/`. From this source the build
(`make openapi-build`, run as part of `make build`) generates the **server stubs**
into [`generated/`](generated) using
[OpenAPI Generator](https://openapi-generator.tech) (server version
`v7.12.0`, pinned in the [`Makefile`](Makefile)). The `generated/` directory
**must not be edited by hand** — changes are overwritten on the next generation
run. `make validate` includes a check that the committed generated code and
bundled spec are up to date.

## Go client

Other Run:AI services consume org-unit-service through the versioned Go client in
the separate [`control-plane/org-unit-service/pkg/client`](pkg/client)
module (generated from this service's OpenAPI spec). `cmd/main.go` itself imports
the sibling clients it depends on (tenants-manager, cluster-service, workloads,
datavolumes, authorization, audit).
