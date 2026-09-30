# Workloads Service

The Workloads service is the control-plane service that owns the platform's
record of every Run:ai-managed **workload** (workspaces, trainings, distributed
trainings, inference services, and AI-application deployments). It is owned by
the **Suricata** team. The service plays two closely-related roles: it is the
**system of record** that the cluster agent syncs workload state into, and it is
the **submission/CRUD API** the rest of the platform uses to create, patch and
delete workloads. It also serves the AI-application Helm-chart upload feature and
exposes workload metrics. Its HTTP API is generated from an OpenAPI specification
under [`swagger/source/openapi.yaml`](swagger/source/openapi.yaml).

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Run modes](#run-modes)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [OpenAPI and code generation](#openapi-and-code-generation)

## Overview

The service is responsible for:

- **Workload tracking** — persisting the state of all Run:ai-managed workload
  components reported from the cluster. The agent bulk-syncs cluster workloads
  and applications into the service's PostgreSQL database; the service serves
  read/list/count/history views over that data (including materialized-view
  refresh for the workloads list).
- **Workload CRUD / submission** — creating, updating and deleting workloads
  (workspaces, trainings, distributed, inference, and "any-workload" generic
  submissions). It resolves policy defaults and rules, org-unit scoping, and
  asset references before persisting the desired workload spec that the agent
  then reconciles into the cluster.
- **AI applications** — managing AI-application definitions and the local Helm
  **chart upload** feature (uploaded charts are served to the cluster from a
  publicly reachable base URL, with a size limit and retention-based cleanup).
- **Background maintenance** — a helper "job mode" runs periodic goroutines: a
  DB cleaner (aging out deleted/finished/stuck-creating workloads), a
  notification client, a watchdog for stuck-creating workloads, a category
  migrator, and metrics writer/exporter loops.

Note that **workloads** and **workloads-manager** are distinct services.
Workloads owns the workload records, submission logic and tracking; it calls the
separate workloads-manager service (via `WORKLOADS_MANAGER_SERVICE_URL`) for the
functionality that service owns.

## Repository layout

```text
.
├── cmd/                          # Service entry point (main.go); dual server/job mode
├── pkg/
│   ├── api/                      # API-layer helpers
│   ├── batch/                    # Batch workload operations
│   ├── clients/                  # HTTP clients for upstream services (settings, …)
│   ├── config/                   # Configuration structs and registration
│   ├── database/
│   │   ├── dao/                  # Tracking DAO layer
│   │   ├── dao_crud/             # CRUD DAO layer
│   │   ├── models/ models_crud/  # DB models
│   │   ├── go_migrations/        # Programmatic migrations run at startup
│   │   └── migrations/{up,down}/ # SQL migrations bundled into the image
│   ├── helper-job/               # "job-mode" detection for the background worker
│   ├── metrics_exporter/         # Workload metrics export loop
│   ├── metrics_writer/           # Per-tenant metrics writer loop
│   ├── notifiy_client/           # Workload notification dispatch
│   ├── pubsub/                   # Pub/sub registration
│   ├── router/                   # HTTP router wiring (main + helper routers)
│   ├── services/                 # Tracking + AI-application domain logic
│   ├── services_crud/            # Workload submission / CRUD logic
│   ├── services_integration*/    # Glue between tracking and CRUD paths
│   ├── settings/                 # Typed workload setting definitions
│   ├── watchdog/                 # Stuck-workload watchdog
│   └── utils/
├── internal/
│   └── generated/                # OpenAPI-generated server stubs (do not edit)
├── swagger/                      # OpenAPI source and built bundle
├── mocks/
├── Dockerfile
├── Makefile
└── go.mod
```

## Run modes

A single binary runs in two modes, selected by the `job-mode` argument:

- **Server mode** (default) — serves the workloads HTTP API via the main router.
- **Job mode** (`job-mode` arg present) — the "workloads-helper" deployment;
  starts the background goroutines (cleaner, notify client, watchdog, metrics
  writer/exporter, category migrator) and a minimal helper router.

## Building and running

The project is built with Go (see [`go.mod`](go.mod) for the required version)
and orchestrated through the [`Makefile`](Makefile), which pulls in the shared
`build/makefile/control-plane.mk` and `index.mk` includes.

| Target | Description |
|---|---|
| `make build` | `openapi-build` then `build-go` — build the OpenAPI bundle + generated code, then compile the binary for amd64 and arm64 |
| `make test` | `test-ginkgo` — run the Ginkgo unit-test suite |
| `make lint` | `lint-go` — run golangci-lint in the builder container |
| `make purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `make validate` | `openapi-validate` — validate the OpenAPI spec (yaml + generated code + public-route checks) |
| `make openapi-lint` | Lint the swagger bundle |
| `make coverage` | `test-go-acc` — coverage report (ignores `pkg/clients`, `pkg/kube_client`) |
| `make local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` |

Under the hood `build-go` runs `build-go-amd` + `build-go-arm`, each compiling
`cmd/main.go` inside the golang builder container with `GOFIPS140=v1.0.0` to
`bin/workloads-<arch>`.

The [`Dockerfile`](Dockerfile) is a two-stage build: a
`registry.access.redhat.com/ubi9/ubi-minimal` stage supplies `/etc/passwd`,
certs and a writable `/tmp`; the final image is `FROM scratch`, runs as user
`65532:65532`, and copies in the compiled `bin/workloads-$TARGETARCH` binary, the
SQL migrations (`migrations/up`, `migrations/down`) and the swagger bundle. The
image takes a single `--build-arg TARGETARCH` (amd64/arm64):

```bash
docker build --build-arg TARGETARCH=amd64 -t workloads:dev .
```

### Local build (local-dev.sh)

`workloads` is a control-plane Go service (`CP_GO_SERVICES`), so
[`local-dev.sh`](../../local-dev.sh) builds it in two steps. First it compiles the
amd64 binary in the builder container:

```bash
make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
```

(`DOCKERHUB_PROXY=` pulls the Go builder image from Docker Hub; `GITHUB_TOKEN=`
is emptied in the make call — the private-module token is applied elsewhere.)
Then it builds and loads the image:

```bash
docker buildx build --platform linux/amd64 --provenance=false --load \
  --build-arg TARGETARCH=amd64 -t runai-local/workloads:0.0.0 .
```

The image ref is `runai-local/workloads:0.0.0` — the tag `0.0.0` matches the
control-plane chart `appVersion`; any other tag breaks the local deploy.

For fast control-plane iteration (rebuild + redeploy CP only, reuse the cluster):

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```

## Testing

| Target | Description |
|---|---|
| `make test` | Run the Ginkgo unit-test suite (`test-ginkgo`) |
| `make coverage` | Build a coverage report via `go-acc` (`test-go-acc`) |

Unit tests run inside the golang builder container; there is no docker-compose
integration suite in this service.

## Configuration

The service is **environment-variable driven** in deployment: the control-plane
chart's workloads ConfigMap and Secret project all settings as env vars. The Go
config structs (in [`pkg/config/config.go`](pkg/config/config.go)) carry sane
built-in defaults, and **any config key can be overridden by its uppercase
env-var equivalent** — which is exactly how the chart supplies values. No YAML
config file is mounted. The authoritative env-var list lives in the chart at
`control-plane/control-plane-chart/templates/workloads-service/configmap.yaml`
(and `secret.yaml` for `DB_USER`).

Key variables:

| Variable | Description | Default |
|---|---|---|
| `SERVICE_PORT` | HTTP port to listen on | from chart |
| `LOG_LEVEL` | Log verbosity | `info` |
| `SERVER_READ_TIMEOUT` / `SERVER_WRITE_TIMEOUT` / `SERVER_IDLE_TIMEOUT` | HTTP server timeouts | from chart |
| `DB_HOST` / `DB_PORT` / `DB_NAME` / `DB_USER` / `DB_SCHEMA` / `DB_SSL_MODE` | PostgreSQL connection (write) | — |
| `DB_READ_HOST` / `DB_READ_PORT` | Optional read-replica endpoint | — |
| `VIEW_REFRESH_RATE_MILLI_SECONDS` | Workloads materialized-view refresh rate | `0` |
| `CLEANER_INTERVAL_HOURS` | DB cleaner run interval | `24` |
| `DELETED_WORKLOADS_DAYS_TO_LIVE` | Retention for deleted workloads | `14` |
| `FINISHED_WORKLOADS_DAYS_TO_LIVE` | Retention for finished workloads | `28` |
| `CREATING_WORKLOADS_MINUTES_TO_LIVE` | Retention for stuck-creating workloads | `60` |
| `WORKLOADS_BATCH_SIZE` | Sync/CRUD batch size | `50` |
| `CHART_PUBLIC_BASE_URL` / `CHART_MAX_UPLOAD_BYTES` / `CHART_RETENTION_MINUTES` | AI-application chart-upload feature | 10 MiB / 24 h |
| `AUDIT_ENABLED` / `AUDIT_SERVICE_URL` | Audit integration | — |
| `TRACING_ENABLED` | OpenTelemetry tracing | — |
| `CACHE_ENABLED` / `CACHE_URL` / `CACHE_TYPE` | Optional distributed cache | `false` |
| `PUBSUB_URL` / `PUBSUB_TYPE` | Pub/sub backend | from chart |
| Upstream URLs — `ASSETS_SERVICE_URL`, `POLICY_SERVICE_URL`, `CLUSTER_SERVICE_URL`, `K8S_OBJECTS_TRACKER_URL`, `ORG_UNIT_URL`, `DATA_VOLUME_URL`, `TENANTS_MANAGER_URL`, `TEMP_BACKEND_SETTINGS_URL`, `WORKLOADS_MANAGER_SERVICE_URL` | Addresses of the control-plane services this one calls | in-cluster service DNS |

## OpenAPI and code generation

The HTTP API is defined in [`swagger/source/openapi.yaml`](swagger/source/openapi.yaml),
split into per-resource files under `swagger/source/resources` and shared
schemas/parameters. From this source the build (`make build` → `openapi-build`)
bundles the spec into `swagger/_build/` and generates the **server stubs** in
[`internal/generated`](internal/generated) (openapi-generator `v7.12.0`,
consumed by the router). The generated directory **must not be edited by hand** —
changes are overwritten on the next generation run. The public Go client for this
service is a separate module under [`pkg/client`](pkg/client).
