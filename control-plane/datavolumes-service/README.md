# datavolumes-service

`datavolumes-service` is the control-plane service that owns **data volumes** —
named, reusable references to existing Kubernetes `PersistentVolumeClaim`s that
let Run:AI users share the same underlying storage across scopes (project,
department, cluster) and across workloads. It is owned by the **Viper** team.
The service exposes a stable HTTP/REST API generated from an OpenAPI
specification, persists data-volume records in its own PostgreSQL database, and
enforces access control and org-hierarchy scoping on every request.

A data volume wraps an origin PVC (`origin_pvc_name`) and adds Run:AI-level
metadata: name, description, owning project/department/cluster, scope, sharing
rules (which other scopes may consume it), and a lifecycle status (`phase`,
conditions, the materialised PVC/PV names). This lets platform admins and
researchers reference one dataset by name instead of copying storage per
project.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [OpenAPI and code generation](#openapi-and-code-generation)

## Overview

The service is responsible for:

- **Data-volume lifecycle** — create, read, list (with filtering, sorting and
  pagination), and delete data volumes. Each volume is uniquely constrained per
  cluster by name, and per (cluster, project/department/cluster) by origin PVC
  so the same PVC is not registered twice within one scope.
- **Scope & sharing** — a volume is created at a project, department, or cluster
  scope and can be shared with additional scopes (`SharedScopes`). The service
  resolves the org hierarchy (project → department → cluster → tenant) to decide
  where a volume is usable, including a "usable in project" query.
- **Authorization** — every route is authorized through the shared
  authorization client using the `datavolumes` resource type (read/create/delete
  actions), scoped by the caller's JWT.
- **Settings** — owns the tenant-scoped `EnableDataVolumes` setting (GA), which
  gates whether data-volume creation is allowed. The setting is registered and
  migrated at startup and backed by `temp-backend-settings`.
- **Cluster capabilities** — checks cluster capabilities (via the
  cluster-capabilities checker against `cluster-service`) so the API only offers
  data volumes on clusters that support them.
- **Events** — publishes/consumes asynchronous pub/sub events so cluster-side
  components can reconcile the actual PVC/PV and report status back into the
  volume's `phase`.
- **Health** — an internal, unauthenticated `GET /` health check.

The API surface is described in
[`swagger/source/openapi.yaml`](swagger/source/openapi.yaml) (title
`datavolumes-service`, version `2.0.0`) and bundled into
[`swagger/_build/swagger.yaml`](swagger/_build/swagger.yaml).

## Repository layout

```text
.
├── cmd/                       # Service entry point (main.go)
├── pkg/
│   ├── cache/                 # Org-hierarchy manager / in-process caches
│   ├── clients/               # Upstream HTTP clients (temp-backend-settings)
│   ├── config/                # Configuration structs + defaults
│   ├── database/
│   │   ├── dao/               # Data-access layer for data volumes
│   │   ├── dbtables/          # Table definitions
│   │   ├── models/            # GORM models (Datavolume, SharedScope, …)
│   │   └── migrations/        # SQL migrations (up/ and down/)
│   ├── pubsub/                # Async event registration/handlers
│   ├── router/                # HTTP router + middleware wiring
│   ├── services/              # API service implementations (datavolumes, settings, health)
│   ├── services_utils/        # Auth-data parsing and shared helpers
│   ├── settings/              # Typed settings (EnableDataVolumes)
│   └── test/
├── internal/
│   └── generated/             # OpenAPI-generated server stubs (do not edit)
├── swagger/                   # OpenAPI source and built bundle
├── Dockerfile
├── Makefile
└── go.mod
```

## Building and running

The project is built with Go (see [`go.mod`](go.mod) for the required version)
and orchestrated through the [`Makefile`](Makefile), which includes the shared
`build/makefile/control-plane.mk` and `index.mk` (which pull in `golang.mk` and
`openapi.mk`).

| Target | Description |
|--------|-------------|
| `make build` | `openapi-build` then `build-go` — build the OpenAPI bundle + generated code, then compile the Go binary (amd64 + arm64) |
| `make build-go` | `build-go-amd` + `build-go-arm` — both arch binaries |
| `make build-go-amd` | Compile `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) in the Go builder container → `bin/datavolumes-service-amd64` |
| `make build-go-arm` | Same for `arm64` → `bin/datavolumes-service-arm64` |
| `make validate` | `openapi-validate` — validate the OpenAPI spec (yaml + generated code up-to-date + public routes) |
| `make openapi-lint` | Lint the swagger bundle |
| `make lint` | `lint-go` — golangci-lint in the builder container |
| `make purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `make local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` (forces `OS=linux ARCH=amd64`) |

The Dockerfile is a two-stage build: a
`registry.access.redhat.com/ubi9/ubi-minimal` `base` stage supplies
`/etc/passwd`, `/etc/group`, TLS certs (`/etc/ssl`, `/etc/pki`),
`nsswitch.conf`, and a writable `/tmp`; the final image is `FROM scratch` and
copies in the compiled `bin/datavolumes-service-$TARGETARCH` binary, the SQL
migrations (`pkg/database/migrations/up|down`), and the swagger bundle. It runs
as user `65532:65532` with entrypoint `./datavolumes-service`. The only build
arg is `TARGETARCH`:

```bash
docker build --build-arg TARGETARCH=amd64 -t datavolumes-service:dev .
```

### Local build (local-dev.sh)

`datavolumes-service` is a control-plane Go service (`CP_GO_SERVICES`), so
`local-dev.sh` builds it in two steps and tags the image with the CP chart
`appVersion` **`0.0.0`** (any other tag breaks the local deploy):

```bash
# 1) compile the amd64 binary in the Go builder container
make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=

# 2) build + load the image
docker buildx build --platform linux/amd64 --provenance=false --load \
  --build-arg TARGETARCH=amd64 -t runai-local/datavolumes-service:0.0.0 .
```

`DOCKERHUB_PROXY=` pulls the Go builder image straight from Docker Hub;
`GITHUB_TOKEN=` is emptied in the make call (the private-module token is applied
elsewhere). Image ref: `runai-local/datavolumes-service:0.0.0`.

To iterate on the control plane locally (rebuild CP only, reuse the cluster):

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```

## Testing

| Target | Description |
|--------|-------------|
| `make test` | `test-ginkgo` — run the Ginkgo unit-test suites in the Go builder container |
| `make coverage` | (from shared includes) open the HTML coverage report |

Unit tests are written with Ginkgo/Gomega and run against the standard shared
Go test tooling; there is no docker-compose integration-test suite in this
service directory.

## Configuration

The service is **environment-variable driven**. Configuration is defined as
typed structs (`pkg/config/config.go`) that ship with in-code defaults; each key
can be overridden by its uppercase environment-variable equivalent. Backing
concerns (database, logging, cache, tracing, pub/sub) are configured through the
shared Run:AI packages initialised in `cmd/main.go`.

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_PORT` | HTTP port the service listens on | `8080` |
| `BACKEND_URL` | `temp-backend-settings` base URL (settings backend) | `http://runai-backend-temp-backend-settings:8080` |
| `CLUSTER_SERVICE_URL` | `cluster-service` base URL (cluster-capabilities checks) | `http://runai-backend-cluster-service:8080` |
| `ORG_UNIT_SERVICE_URL` | `org-unit-service` base URL (org hierarchy) | `http://runai-backend-org-unit-service:8080` |
| `AUTH_SERVICE_URL` | Authorization service base URL (authorization client) | — |
| `LOG_LEVEL` | Log verbosity (shared logging package) | `info` |
| `DATABASE_*` | PostgreSQL connection settings, consumed by the shared data-access layer (`dao-utils`, DB name `datavolumes`) | — |

Pub/sub (NATS) and tracing are initialised from their shared-package defaults.

## OpenAPI and code generation

The HTTP API is defined in
[`swagger/source/openapi.yaml`](swagger/source/openapi.yaml), split into
per-resource files under `swagger/source/resources` and shared
`parameters`/`schemas`/`securitySchemes`. From this source the build generates
the **server stubs** in [`internal/generated`](internal/generated) (consumed by
`pkg/router` and the service implementations). Generation is performed by
[OpenAPI Generator](https://openapi-generator.tech) (server generator pinned to
`v7.12.0` in the [`Makefile`](Makefile)); `make build` runs `openapi-build`
first. The `internal/generated` directory **must not be edited by hand** — it is
overwritten on the next generation run.
