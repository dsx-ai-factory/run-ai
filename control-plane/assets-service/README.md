# Assets Service

Assets Service is the control-plane service (owned by the **Viper** team) that
manages **workload assets** for Run:ai — the reusable building blocks users
reference when they submit workloads. It owns environments, compute profiles,
data sources, credentials, models, inference/training/workspace definitions and
workload templates, persists them in its own PostgreSQL schema, and keeps the
control-plane copy in sync with what actually exists on the clusters. It exposes
a stable HTTP/REST API generated from an OpenAPI specification.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [Authorization](#authorization)
- [OpenAPI and code generation](#openapi-and-code-generation)
- [Go client](#go-client)

## Overview

The service is responsible for:

- **Environments** — container image + runtime definitions (command, args,
  ports, environment variables) reused across workloads.
- **Compute** — compute resource profiles (CPU / GPU / memory requests and
  limits).
- **Data sources** — NFS, PVC, Git, S3, host-path and ConfigMap volume
  definitions, plus data-volume references.
- **Credentials** — access keys, passwords, docker-registry secrets, generic
  secrets and NGC API keys. Secret material is encrypted (see `pkg/encryption`)
  and stored in Kubernetes Secrets in the configured namespace.
- **Models & inference** — model assets, Hugging Face model info lookups, and
  the v2 inference / training / workspace merged asset definitions.
- **Workload templates** — reusable templates that bundle the above assets.
- **Flat workloads** — the denormalised (`flat_workloads`) view used for fast
  workload listing and querying.
- **Synchronization** — the internal `Synchronization` API and agent-sync flow
  that reconcile assets between the clusters and the control plane, plus a
  background cleanup job (`pkg/database/dbcleaner`) that prunes stale assets.

The API surface is described in
[`swagger/source/openapi.yaml`](swagger/source/openapi.yaml) (title *Assets*,
version 2.0.0) and bundled into `swagger/_build/`.

## Repository layout

```text
.
├── cmd/                     # Service entry point (main.go): init clients, start HTTP + cleanup job
├── pkg/
│   ├── clients/             # HTTP clients for upstream services (workloads-service, …)
│   ├── config/              # Configuration structs and loading (env-var driven)
│   ├── database/            # DAO layer, dbmodel, migrations (up/down), dbcleaner job
│   ├── datavolumes/         # Data-volume asset accessor
│   ├── encryption/          # Secret encryption helpers
│   ├── flat_workloads/      # Denormalised workload view logic
│   ├── kube_client/         # Kubernetes client (Secrets, etc.)
│   ├── projects/            # Project resolution helpers
│   ├── pubsub/              # Pub/sub registration
│   ├── query_types/         # Query/filter types
│   ├── services/            # API service implementations (one file per asset kind) + router
│   ├── testing/             # Test helpers
│   └── utils/
├── generated/               # OpenAPI-generated server stubs (do not edit)
├── swagger/                 # OpenAPI source (source/) and built bundle (_build/)
├── third_party/
├── Dockerfile
├── Makefile
└── go.mod
```

The Go client is **not** in this repo — it is generated into the separate
`runai-clients/assets-service-client` module (see [Go client](#go-client)).

## Building and running

The project is built with Go (see [`go.mod`](go.mod) for the required version)
and orchestrated through the [`Makefile`](Makefile), which sets `SERVICE_NAME`
and includes the shared `../../build/makefile/control-plane.mk` and
`../../build/makefile/index.mk` (which in turn pull in `golang.mk` and
`openapi.mk`). OpenAPI generation is pinned to `v7.12.0`.

| Target | Description |
|--------|-------------|
| `make build` | `openapi-build` then `build-go` — build the OpenAPI bundle + generate code, then compile the Go binary for amd64 + arm64 |
| `make build-go` | `build-go-amd` + `build-go-arm` |
| `make build-go-amd` | Compile `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) inside the Go builder container → `bin/assets-service-amd64` |
| `make build-go-arm` | Same for `arm64` → `bin/assets-service-arm64` |
| `make lint` | `lint-go` — run golangci-lint in the builder container |
| `make purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `make validate` | `openapi-validate` — validate the OpenAPI spec (yaml + generated code up-to-date + public routes + `x-service` field) |
| `make openapi-lint` | Lint the built swagger bundle |
| `make local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` (forces `OS=linux ARCH=amd64`) |

The Dockerfile is a multi-stage build. A
`registry.access.redhat.com/ubi9/ubi-minimal` `base` stage supplies
`/etc/passwd`, `/etc/group`, SSL certs (`/etc/ssl`, `/etc/pki`), `nsswitch.conf`
and a writable `/tmp`; the final image is `FROM scratch`, runs as user
`65532:65532`, and copies in the `assets-service-$TARGETARCH` binary, the
`NOTICE` file, the SQL migrations (`pkg/database/migrations/up|down`) and the
built swagger bundle. It takes a single `--build-arg TARGETARCH` (`amd64` /
`arm64`) to select which binary to package:

```bash
docker build --build-arg TARGETARCH=amd64 -t assets-service:dev .
```

### Local build (local-dev.sh)

`assets-service` is a control-plane Go service (listed in `CP_GO_SERVICES`), so
the top-level [`local-dev.sh`](../../local-dev.sh) builds it via its
`build_go_service` function in two steps:

1. Compile the amd64 binary in the builder container:
   ```bash
   make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
   ```
   (`DOCKERHUB_PROXY=` pulls the Go builder image straight from Docker Hub;
   `GITHUB_TOKEN=` is emptied in the make call — the private-module token is
   applied elsewhere.)
2. Build + load the image:
   ```bash
   docker buildx build --platform linux/amd64 --provenance=false --load \
     --build-arg TARGETARCH=amd64 -t runai-local/assets-service:0.0.0 .
   ```

The image is tagged `runai-local/assets-service:0.0.0` — tag `0.0.0` is the
control-plane chart `appVersion`; any other tag breaks the local deploy.

To iterate on the control plane only (rebuild CP images, reuse the cluster):

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```

## Testing

| Target | Description |
|--------|-------------|
| `make test` | `test-ginkgo` — run the Ginkgo suites recursively (`--skip-package=third_party`, 15m timeout) |
| `make coverage` | `test-go-acc` — produce a go-acc coverage report (ignores `pkg/clients`, `pkg/kube_client`, `cmd`) |

The unit/integration tests are Ginkgo-based and run in-process; test helpers
live under `pkg/testing`. Tests that touch the database use the mock config in
`pkg/config` (see `MockConfig`).

## Configuration

Assets Service is **environment-variable driven**. Configuration structs live in
[`pkg/config/config.go`](pkg/config/config.go); each field's key maps to its
uppercase env-var equivalent (e.g. `service_port` → `SERVICE_PORT`). Defaults
come from each struct's `SetDefaults`. Database connection settings are loaded
by the shared `runaidb` config from the `DB_*` env vars.

**Server**

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_PORT` | HTTP port to listen on | `8080` |
| `RETRY_MAX_INTERVAL_SEC` | Max backoff interval for retries | `10` |
| `RETRY_MAX_ELAPSED_SEC` | Max total retry time | `60` |
| `DAYS_LIMIT_FOR_FIRST_SYNC` | Look-back window for the first cluster sync | `7` |
| `CLEANUP_REFRESH_RATE_HOURS` | Cleanup job interval | `24` |
| `CLEANUP_KEEP_ASSETS_DAYS` | How long to keep deleted assets | `30` |
| `LOG_LEVEL` | Log level | `info` |
| `KEY_DELIMITER` | Config key delimiter | `_` |
| `POD_NAMESPACE` | Kubernetes namespace used for Secrets | `runai-backend` |
| `SECRETS_MOUNT_PATH` | Path to the mounted secrets folder | — |
| `PUBSUB_ENABLED` | Enable pub/sub | — |

**Upstream services**

| Variable | Description |
|----------|-------------|
| `POLICY_SERVICE_URL` | policy-service base URL |
| `AUTH_SERVICE_URL` | authorization service base URL |
| `K8S_OBJECTS_TRACKER_URL` | k8s-objects-tracker base URL |
| `WORKLOADS_SERVICE_URL` | workloads service base URL |
| `CLUSTER_SERVICE_URL` | cluster-service base URL |
| `DATA_VOLUMES_SERVICE_URL` | datavolumes-service base URL |
| `ORG_UNIT_SERVICE_URL` | org-unit-service base URL |
| `DB_BACKEND_SCHEMA` | Schema of the legacy backend DB (`db_backend_schema`, default `backend`) |
| `SYNC_BULK_SIZE` | Batch size for asset sync (default `50`) |

**Database** (loaded by the shared `runaidb` config)

| Variable | Description |
|----------|-------------|
| `DB_HOST` | PostgreSQL host |
| `DB_NAME` | Database name |
| `DB_USER` | Database user |
| `DB_PASSWORD` | Database password |
| `DB_SCHEMA` | Assets schema |
| `DB_SSL_MODE` | SSL mode (`disable`, …) |
| `MIGRATION_FILES` | `file://` URL to the migrations directory |

**Kubernetes client**

| Variable | Description | Default |
|----------|-------------|---------|
| `KUBECONFIG` | Optional kubeconfig for running outside a pod | — |
| `K8S_CLIENT_CONFIG_QPS` | k8s client QPS | `600` |
| `K8S_CLIENT_CONFIG_BURST` | k8s client burst | `3000` |

**Secret data keys** — `DATA_KEY_USER` (`user`), `DATA_KEY_PASSWORD`
(`password`), `DATA_KEY_ACCESS_KEY_ID` (`accessKeyId`),
`DATA_KEY_SECRET_ACCESS_KEY` (`secretAccessKey`).

**Hugging Face** — `MAX_CONCURRENT_HUGGING_FACE_REQUESTS` (`8`),
`HUGGING_FACE_MODEL_INFO_URL`, `HUGGING_FACE_MODEL_FILE_URL`,
`HUGGING_FACE_CALCULATOR_JOIN_URL`, `HUGGING_FACE_CALCULATOR_DATA_URL`.

**Caches** — the tenant-keys, cluster, org-unit and projects caches each accept
`EXPIRATION_SEC` (`600`) and `CLEANUP_SEC` (`1200`); the projects cache also
takes `CACHE_PROJECTS_MAX_RECORD_SIZE_KB` (`150`).

<details>
<summary>Local run env + port-forwarding</summary>

Run the server directly:

```bash
go run cmd/main.go
```

Example env (IntelliJ run configuration):

```
ORG_UNIT_SERVICE_URL=http://localhost:9091;CLUSTER_SERVICE_URL=http://localhost:5500;AUTH_SERVICE_URL=http://localhost:4000;DB_HOST=localhost;DB_NAME=backend;DB_PASSWORD=password;DB_SCHEMA=assets;DB_SSL_MODE=disable;DB_USER=user;K8S_OBJECTS_TRACKER_URL=http://localhost:6000;KUBECONFIG=~/.kube/config;MIGRATION_FILES=file:///<repo>/control-plane/assets-service/pkg/database/migrations;POLICY_SERVICE_URL=http://localhost:8000;SERVICE_PORT=8080;WORKLOADS_SERVICE_URL=http://localhost:9000;DATA_VOLUMES_SERVICE_URL=http://localhost:9090
```

Port-forward the upstream services:

```bash
kubectl -n runai-backend port-forward service/runai-backend-postgresql 5432:5432 &
kubectl -n runai-backend port-forward service/runai-backend-authorization 4000:8080 &
kubectl -n runai-backend port-forward service/runai-backend-k8s-objects-tracker 6000:8080 &
kubectl -n runai-backend port-forward service/runai-backend-policy-service 8000:8080 &
kubectl -n runai-backend port-forward service/runai-backend-workloads 9000:8080 &
kubectl -n runai-backend port-forward service/runai-backend-cluster-service 5500:8080 &
kubectl -n runai-backend port-forward service/runai-backend-datavolumes 9090:8080 &
kubectl -n runai-backend port-forward service/runai-backend-org-unit-service 9091:8080 &
```

</details>

## Authorization

The service assumes all authentication/authorization is performed upstream (by
the gateway) and delivered as request headers:

| Header | Purpose |
|--------|---------|
| `X-Runai-Subject-Id` | The user or application making the request |
| `X-Runai-Tenant-Id` | The tenant ID of the requester |

## OpenAPI and code generation

The HTTP API is defined in
[`swagger/source/openapi.yaml`](swagger/source/openapi.yaml), split into
per-resource files under `swagger/source/resources`, shared `parameters` and
`schemas`. From this source the build (`make build` → `openapi-build`)
generates:

- The **server stubs** in [`generated/`](generated)
  (`OPENAPI_GENERATOR_SERVER_CODE_PATH=generated`), consumed by `pkg/services`.
- The **Go client** into the separate `runai-clients/assets-service-client`
  module.

Both are produced by [OpenAPI Generator](https://openapi-generator.tech)
(version `v7.12.0`, pinned in the `Makefile`) and **must not be edited by
hand** — changes are overwritten on the next generation run. Run
`make openapi-generate-code` (or `make build`) to regenerate.

## Go client

Other Run:ai services consume Assets Service through the generated Go client in
the [`runai-clients/assets-service-client`](../../runai-clients/assets-service-client)
module (its own Go module, versioned independently). There is no in-repo
`pkg/client` for this service — regenerate the client from this service's spec
via the OpenAPI targets above.
</content>
