# k8s-objects-tracker

`k8s-objects-tracker` is the control-plane service that keeps a durable,
control-plane-side record of selected Kubernetes objects living in the connected
Run:ai clusters, so the rest of the platform can read and reason about them
without talking to each cluster directly. Cluster-side components report object
state up to the control plane; this service persists it in its own PostgreSQL
database and serves it back through a scoped HTTP/REST API (project / department /
cluster). It is owned by the **Viper** team.

Currently tracked object kinds: **PersistentVolumeClaims (PVCs)**, **Secrets**,
**ConfigMaps**, **StorageClasses** and their **StorageClassConfiguration**, plus
per-object **resource issues** and a per-cluster **objects sync status**.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [OpenAPI and code generation](#openapi-and-code-generation)

## Overview

The service is responsible for:

- **PVCs** — storing PVCs reported from each cluster and exposing them scoped by
  project, department and cluster, including data-sharing attributes.
- **Secrets** — tracking Secret existence/metadata (not payloads) and serving
  them scoped by cluster, project and department (`ClusterSecrets`,
  `ProjectSecrets`, `DepartmentSecrets`, and their by-name variants).
- **ConfigMaps** — persisting ConfigMap records and cleaning them up when their
  owning project/cluster/tenant is removed.
- **StorageClasses & StorageClassConfiguration** — the set of storage classes
  present in a cluster plus control-plane customization (permissions, defaults,
  customization) layered on top of them.
- **Cluster objects status** — a per-cluster view of how far each cluster's
  object sync has progressed (backed by the ConfigMaps accessor).
- **Resource issues** — issues attached to tracked objects (e.g. a PVC), stored
  in a dedicated table and cleaned up alongside their objects.

Object records enter the system in two ways: cluster components push object
state into the service's write endpoints, and the service subscribes to
platform pub/sub events to keep its store consistent —
`ProjectDeleted`, `ClusterDelete` and `TenantDelete` each trigger cascading
cleanup of the objects (and resource issues) belonging to that scope. Incoming
requests carry a JWT (project-scoped APIs preserve the caller token), and
authorization is enforced via the shared authorization client. A cluster
capabilities checker (backed by `cluster-service`) and in-process caches for
clusters and projects reduce cross-service chatter.

The API surface is defined in
[`swagger/source/openapi.yaml`](swagger/source/openapi.yaml) and bundled into
[`swagger/_build`](swagger/_build).

## Repository layout

```text
.
├── cmd/                        # Service entry point (main.go)
├── pkg/
│   ├── auth/                   # RBAC / permitted-scopes helpers
│   ├── clients/                # HTTP clients for authorization, clusters, org-unit
│   ├── common/                 # Context helpers (tenant id, etc.)
│   ├── config/                 # Typed configuration + defaults
│   ├── database/               # DAO layer, GORM models, SQL migrations (up/down)
│   ├── projects/               # Projects cache
│   ├── pubsub/                 # Event subscriptions (project/cluster/tenant delete)
│   └── services/               # API service implementations + chi router wiring
├── internal/
│   └── generated/              # OpenAPI-generated server stubs (do not edit)
├── swagger/                    # OpenAPI source, config and built bundle
├── Dockerfile
├── Makefile
└── go.mod
```

## Building and running

The project is built with Go (see [`go.mod`](go.mod)) and orchestrated through
the [`Makefile`](Makefile), which sets `SERVICE_NAME=k8s-objects-tracker` and
includes the shared `../../build/makefile/control-plane.mk` and `index.mk`
(which pull in `golang.mk` and `openapi.mk`).

| Target | Description |
|--------|-------------|
| `make build` | `openapi-build` then `build-go` — build the OpenAPI bundle + generated code, then compile the Go binary for amd64 + arm64 |
| `make build-go` | `build-go-amd` + `build-go-arm` |
| `make build-go-amd` | Compile `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) in the Go builder container → `bin/k8s-objects-tracker-amd64` |
| `make build-go-arm` | Same for `arm64` → `bin/k8s-objects-tracker-arm64` |
| `make lint` | `lint-go` — golangci-lint in the builder container |
| `make purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `make validate` | `openapi-validate` — validate the OpenAPI spec (schema, code up-to-date, public/service routes) |
| `make openapi-lint` | Lint the swagger bundle |

The runtime image is defined by [`Dockerfile`](Dockerfile): a
`registry.access.redhat.com/ubi9/ubi-minimal` stage supplies `/etc/passwd`,
certs (`/etc/ssl`, `/etc/pki`) and a writable `/tmp`, and the final image is
`FROM scratch`, running as user `65532:65532`. It bundles the compiled
`bin/k8s-objects-tracker-$TARGETARCH` binary, the SQL migrations
(`pkg/database/migrations/up|down`) and the swagger bundle. It takes one
build arg, `TARGETARCH`, which selects the binary to copy:

```bash
docker build --build-arg TARGETARCH=amd64 -t k8s-objects-tracker:dev .
```

### Local build (local-dev.sh)

`k8s-objects-tracker` is a control-plane Go service (listed in `CP_GO_SERVICES`
in [`local-dev.sh`](../../local-dev.sh)), so the `build_go_service` path runs two
steps:

1. Compile the amd64 binary in the builder container:
   ```bash
   make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
   ```
   (`DOCKERHUB_PROXY=` pulls the Go builder image from Docker Hub;
   `GITHUB_TOKEN=` is emptied in the make call — the private-module token is
   applied elsewhere.)
2. Build + load the image:
   ```bash
   docker buildx build --platform linux/amd64 --provenance=false --load \
     --build-arg TARGETARCH=amd64 -t runai-local/k8s-objects-tracker:0.0.0 .
   ```

The image ref is `runai-local/k8s-objects-tracker:0.0.0`. The tag **`0.0.0`** is
the control-plane chart `appVersion` — any other tag breaks the local deploy.

To rebuild and redeploy only the control plane against an existing cluster:

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```

## Testing

| Target | Description |
|--------|-------------|
| `make test` | `test-ginkgo` — run the Ginkgo unit-test suites |
| `make coverage` | `test-go-acc` — produce a coverage report (`GO_ACC_IGNORE=pkg/clients`) |

Tests use Ginkgo/Gomega; DAO suites exercise PostgreSQL via the shared
database-testing helpers and therefore require a running Docker daemon.

## Configuration

The service is **environment-variable driven**. Configuration is expressed as
typed sections with built-in defaults; every key can be overridden by its
uppercase environment-variable equivalent (e.g. `service_port` → `SERVICE_PORT`,
`log_level` → `LOG_LEVEL`, `cluster_service_url` → `CLUSTER_SERVICE_URL`). No
config file is shipped with the service — defaults apply unless overridden.

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_PORT` | HTTP port to listen on | `8080` |
| `LOG_LEVEL` | Log verbosity | `info` |
| `RETRY_MAX_INTERVAL_SEC` | Max backoff interval between retries | `10` |
| `RETRY_MAX_ELAPSED_SEC` | Max total retry window | `60` |
| `DAYS_LIMIT_FOR_FIRST_SYNC` | Look-back window (days) for a cluster's first object sync | `7` |
| `CLEANUP_REFRESH_RATE_HOURS` | How often the cleanup job runs | `24` |
| `CLEANUP_KEEP_ASSETS_DAYS` | How long deleted assets are retained | `30` |
| `KEY_DELIMITER` | Delimiter used in composite object keys | `_` |
| `K8S_CLIENT_CONFIG_QPS` | Kubernetes client QPS | `600` |
| `K8S_CLIENT_CONFIG_BURST` | Kubernetes client burst | `3000` |
| `CLUSTER_SERVICE_URL` | cluster-service base URL | `http://runai-backend-cluster-service:8080` |
| `ORG_UNIT_SERVICE_URL` | org-unit-service base URL | `http://runai-backend-org-unit-service:8080` |
| `AUTH_SERVICE_URL` | authorization service base URL | `http://runai-backend-authorization:8080` |
| `EXPIRATION_SEC` / `CLEANUP_SEC` | Cluster/projects cache TTL and cleanup interval | `600` / `1200` |
| `CACHE_PROJECTS_MAX_RECORD_SIZE_KB` | Max cached projects record size | `300` |
| `PUBSUB_ENABLED` | Enable/disable pub/sub event handling | — |
| `SECRETS_MOUNT_PATH` | Directory the config loader reads mounted secrets from | `/etc/secrets` |

The PostgreSQL connection (host, port, user, password, `sslmode`, statement
timeout, and schema — defaulting to `k8s_objects_tracker`) is provided through
the standard database configuration and is likewise overridable via environment
variables. SQL migrations are bundled in the image under
`pkg/database/migrations/up|down` and applied on startup.

## OpenAPI and code generation

The HTTP API is defined in
[`swagger/source/openapi.yaml`](swagger/source/openapi.yaml), split into
per-resource files under `swagger/source/resources` (PVCs, Secrets in cluster /
project / department scopes, StorageClasses, StorageClassConfiguration, …) and
shared schemas/parameters/securitySchemes. From this source `make build`
(via `openapi-build`) generates the **server stubs** in
[`internal/generated`](internal/generated) — consumed by the chi router in
`pkg/services`.

The generated code is produced by
[OpenAPI Generator](https://openapi-generator.tech) (`v7.12.0`, pinned in the
[`Makefile`](Makefile)) and **must not be edited by hand** — changes are
overwritten on the next generation run. `make validate` re-generates into a temp
location and diffs it against the committed spec and stubs to keep them in sync.
