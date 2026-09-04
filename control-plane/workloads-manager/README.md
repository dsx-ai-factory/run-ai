# workloads-manager

`workloads-manager` is the control-plane service that owns the **v2 workload
submission pipeline** for Run:AI. It is the business-logic brain that takes a
user's workload request, validates and enriches it, applies policy, translates
it into the concrete Kubernetes manifests a cluster will run, and coordinates
with the other control-plane services before the workload is persisted and
dispatched. It is owned by the **Viper** team. It is a distinct service from
`workloads`: `workloads` is the storage/record service for workloads, whereas
`workloads-manager` is the higher-level orchestration and conversion layer that
sits in front of it and produces the manifests.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Required backing services](#required-backing-services)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [OpenAPI and code generation](#openapi-and-code-generation)

## Overview

The service is responsible for:

- **Workloads v2 API** — the `WorkloadsV2` endpoints (create, get by id, list,
  internal-create, batch) that accept a declarative workload request and turn it
  into a runnable workload. Requests can be submitted in the native v2 shape or
  converted from the v1 shape.
- **Manifest generation and dry-run** — building the concrete Kubernetes
  manifests for a workload from its request, including a **dry-run** mode that
  returns the would-be result and validation/submission errors without creating
  anything.
- **GVK plugins** — a plugin factory keyed by GroupVersionKind that applies
  resource-type-specific processing. This includes the NVIDIA **NIM** /
  Dynamo graph-deployment path and a no-op default, so new workload kinds are
  handled by registering a plugin rather than branching in core code.
- **NVIDIA NIM services** — the `NVIDIANIM` endpoints for NIM-backed inference
  services, including NIM-specific networking and multi-node NVLink (**MNNVL**)
  and segment-topology handling.
- **Policy application** — resolving and applying policy from the
  policy-service against the submitted workload before it is accepted.
- **Settings** — tenant-scoped workload settings, notably the
  **YAML submission** toggle that governs whether workload types may be
  submitted as standard Kubernetes YAML.
- **Persistence & events** — its own PostgreSQL database (schema managed by the
  bundled SQL migrations) and pub/sub event publication for workload changes.

Incoming requests are authenticated and authorized (the JWT is preserved and
forwarded), traced, and metered by shared HTTP middleware before reaching the
generated API controllers.

The API surface is described in
[`swagger/source/openapi.yaml`](swagger/source/openapi.yaml) and bundled into
[`swagger/_build/swagger.yaml`](swagger/_build/swagger.yaml).

## Repository layout

```text
.
├── cmd/                        # Service entry point (main.go)
├── pkg/
│   ├── clients/                # HTTP clients for upstream services
│   │   ├── clusters/           #   cluster-service
│   │   ├── org_unit/           #   org-unit-service
│   │   └── workloads/          #   workloads service
│   ├── config/                 # Configuration structs + registration
│   ├── constants/
│   ├── database/               # DAO layer, GORM models and SQL migrations
│   ├── middleware/             # Request-context middleware
│   ├── plugins/                # GVK plugin factory (NIM / Dynamo, no-op)
│   ├── pubsub/                 # Pub/sub registration
│   ├── router/                 # HTTP router wiring
│   ├── services/               # API implementations (workloads v2, NIM, settings, health)
│   └── settings/               # Typed setting definitions (YAML submission)
├── internal/
│   └── generated/              # OpenAPI-generated server stubs (do not edit)
├── swagger/                    # OpenAPI source and built bundle
├── Dockerfile
├── Makefile
└── go.mod
```

## Required backing services

Defaults are set in [`pkg/config/config.go`](pkg/config/config.go) and can be
overridden via environment variables (see [Configuration](#configuration)).

| Service | Default address |
|---|---|
| authorization | `http://runai-backend-authorization:8080` |
| org-unit-service | `http://runai-backend-org-unit-service:8080` |
| cluster-service | `http://runai-backend-cluster-service:8080` |
| workloads | `http://runai-backend-workloads:8080` |
| policy-service | `http://runai-backend-policy-service:8080` |
| PostgreSQL | configured via `DB_*` env vars |

## Building and running

The project is built with Go (see [`go.mod`](go.mod) for the required version)
and orchestrated through the [`Makefile`](Makefile), which includes the shared
`build/makefile/control-plane.mk` and `index.mk` (which in turn pull in
`golang.mk` and `openapi.mk`).

| Target | Description |
|---|---|
| `make build` | `openapi-build` then `build-go` — build the OpenAPI bundle and generated code, then compile the Go binary for amd64 + arm64 |
| `make build-go` | `build-go-amd` + `build-go-arm` — both arch binaries |
| `make build-go-amd` | Compile `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) in the Go builder container → `bin/workloads-manager-amd64` |
| `make build-go-arm` | Same for `arm64` → `bin/workloads-manager-arm64` |
| `make validate` | `openapi-validate` — validate the OpenAPI spec (bundle/code up-to-date + auth checks) |
| `make openapi-lint` | Lint the built swagger bundle |
| `make lint` | `lint-go` — run golangci-lint in the builder container |
| `make purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `make local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` (OS=linux, ARCH=amd64) |

The [`Dockerfile`](Dockerfile) is a two-stage build: a
`registry.access.redhat.com/ubi9/ubi-minimal` `base` stage supplies
`/etc/passwd`, `/etc/group`, CA certificates (`/etc/ssl`, `/etc/pki`),
`nsswitch.conf`, and a writable `/tmp`; the final image is `FROM scratch`. It
copies in `bin/workloads-manager-$TARGETARCH`, the SQL migrations
(`pkg/database/migrations/up|down`), and the built swagger bundle, and runs as
user `65532:65532`. `TARGETARCH` is the only build arg.

```bash
docker build --build-arg TARGETARCH=amd64 -t workloads-manager:dev .
```

### Local build (local-dev.sh)

`workloads-manager` is in `CP_GO_SERVICES`, so the repo-root `local-dev.sh`
builds it via its `build_go_service` function in two steps:

1. Compile the amd64 binary inside the Go builder container:

   ```bash
   make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
   ```

   (`DOCKERHUB_PROXY=` pulls the `library/golang` builder image straight from
   Docker Hub; `GITHUB_TOKEN=` is emptied in the make call — the private-module
   token is applied elsewhere.)

2. Build and load the image:

   ```bash
   docker buildx build --platform linux/amd64 --provenance=false --load \
     --build-arg TARGETARCH=amd64 -t runai-local/workloads-manager:0.0.0 .
   ```

   Image ref: `runai-local/workloads-manager:0.0.0`. The tag **must** be `0.0.0`
   — it matches the control-plane chart `appVersion`; any other tag breaks the
   local deploy.

To iterate on this service against an existing local cluster:

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```

## Testing

| Target | Description |
|---|---|
| `make test` | `test-ginkgo` — run the Ginkgo unit-test suites (randomized, in the Go container) |
| `make coverage` | `test-go-acc` — run coverage (ignoring `pkg/clients`, `pkg/kube_client`) and produce `cover.out` |

The tests are Ginkgo suites that live alongside the code they cover (for example
`pkg/services/*_test.go` and `pkg/plugins/*_test.go`); there is no
docker-compose integration-test suite in this service.

## Configuration

Configuration is **environment-variable driven**. The service registers typed
config structs (`Services`, `Server`, and the shared database config) at
startup; each field has a built-in default and is populated from an environment
variable whose name is the upper-cased form of the key shown below. Any key can
therefore be overridden by setting its uppercase env-var equivalent.

**Service endpoints and behavior** (from [`pkg/config/config.go`](pkg/config/config.go)):

| Env var | Description | Default |
|---|---|---|
| `AUTHORIZATION_SERVICE_URL` | authorization service base URL | `http://runai-backend-authorization:8080` |
| `ORG_UNIT_URL` | org-unit-service base URL | `http://runai-backend-org-unit-service:8080` |
| `CLUSTER_SERVICE_URL` | cluster-service base URL | `http://runai-backend-cluster-service:8080` |
| `WORKLOADS_SERVICE_URL` | workloads service base URL | `http://runai-backend-workloads:8080` |
| `POLICY_SERVICE_URL` | policy-service base URL | `http://runai-backend-policy-service:8080` |
| `CLUSTER_CACHE_EXPIRATION_SEC` | cluster-capabilities cache TTL | `600` |
| `CLUSTER_CACHE_CLEANUP_SEC` | cluster-capabilities cache cleanup interval | `1200` |
| `SUSPEND_DEF_CACHE_TTL_SEC` | suspend-definition cache TTL | `300` |
| `SERVICE_PORT` | HTTP port to listen on | `8080` |
| `RETRY_MAX_INTERVAL_SEC` | max backoff interval for retries | `10` |
| `RETRY_MAX_ELAPSED_SEC` | max total retry time | `60` |

**Database** (`runaidb`; a subset shown):

| Env var | Description |
|---|---|
| `DB_HOST` / `DB_PORT` | PostgreSQL host / port |
| `DB_READ_HOST` / `DB_READ_PORT` | read-replica host / port |
| `DB_NAME` / `DB_SCHEMA` | database name / schema |
| `DB_USER` / `DB_PASSWORD` | credentials (secret) |
| `DB_SSL_MODE` | PostgreSQL SSL mode |
| `DB_MIGRATION_FILES` | migration source dir |
| `SERVICE_NAME` | service-session DB role name |

Logging, metrics, tracing and pub/sub are initialized from their own shared
config (e.g. `LOG_LEVEL`) at startup.

## OpenAPI and code generation

The HTTP API is defined in
[`swagger/source/openapi.yaml`](swagger/source/openapi.yaml), split into
per-resource files under `swagger/source/resources` and shared schemas under
`swagger/source/schemas`. `make build` runs `openapi-build` first, which bundles
the spec into `swagger/_build/` and generates the **server stubs** in
[`internal/generated`](internal/generated) (consumed by
[`pkg/router/router.go`](pkg/router/router.go)). The generator is
[OpenAPI Generator](https://openapi-generator.tech) `v7.12.0` (pinned in the
`Makefile`). The `generated/` directory **must not be edited by hand** — changes
are overwritten on the next generation run. `make validate` checks that the
committed bundle and generated code are up to date and that the spec declares
bearer authentication.
