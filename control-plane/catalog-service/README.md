# catalog-service

Catalog Service is the control-plane service (owned by the **Viper** team) that
powers Run:AI's catalog of runnable content: container **images**, container
**registries**, **Helm repositories** and their charts, curated **inference**
presets, and **Hugging Face** / **NGC** model discovery. It aggregates data
from external providers (NVIDIA NGC, Hugging Face, and user-configured
registries and Helm repos) behind a stable HTTP/REST API generated from an
OpenAPI specification, and persists user-configured registries, Helm repos and
feature settings in its own PostgreSQL database.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [OpenAPI and code generation](#openapi-and-code-generation)

## Overview

The service is responsible for:

- **Images** — browsing container images and image tags. It queries NVIDIA NGC
  (via an internal NGC client with a configurable request timeout and response
  caching) and cross-references the tenant's assets through the assets-service
  client.
- **Registries** — CRUD for container registries the tenant has configured, plus
  a set of built-in registry presets (`InitRegistriesPresets`).
- **Helm repositories** — CRUD for Helm repos and browsing the charts they
  expose, seeded with built-in Helm-repo presets (`InitHelmRepoPresets`). Repos
  can be marked private.
- **Inferences** — serving curated inference-catalog entries.
- **Hugging Face** — model/catalog lookup against Hugging Face.
- **Settings** — the tenant-scoped `EnableNgcIntegration` feature setting,
  managed through the shared setting framework (backed by
  `temp-backend-settings`) and migrated on startup.
- **Health** — a health/readiness endpoint.

The HTTP layer is a `chi` router (`pkg/router`) wrapping the OpenAPI-generated
controllers in `internal/generated`, with shared middleware for tracing,
request context, metrics, structured logging, and JWT propagation. Domain logic
lives in `pkg/services`, external calls in `pkg/clients`, and persistence in
`pkg/database`. The API surface is described in
[`swagger/source/openapi.yaml`](swagger/source/openapi.yaml) and bundled into
[`swagger/_build/`](swagger/_build).

## Repository layout

```text
.
├── cmd/                        # Service entry point (main.go)
├── pkg/
│   ├── clients/                # Upstream clients: NGC, assets-service, temp-backend-settings
│   │   └── mocks/              # Generated client mocks
│   ├── config/                 # Configuration structs and registration
│   ├── constants/              # Shared constants
│   ├── database/
│   │   ├── dao/                # DAO layer
│   │   ├── migrations/         # SQL migrations (up/ and down/)
│   │   └── models/             # DB models
│   ├── router/                 # HTTP router wiring (chi + middleware)
│   ├── services/               # API service implementations (images, registries,
│   │   │                       #   helm-repos, inferences, settings, health)
│   │   └── huggingface/        # Hugging Face catalog lookup
│   ├── settings/               # Typed setting definitions (EnableNgcIntegration)
│   └── test/                   # Package-level API tests
├── internal/
│   └── generated/              # OpenAPI-generated server stubs (do not edit)
├── swagger/                    # OpenAPI source and built bundle
├── Dockerfile
├── Makefile
└── go.mod
```

## Building and running

The project is built with Go (see [`go.mod`](go.mod) for the required version)
and orchestrated through the [`Makefile`](Makefile), which includes the shared
`../../build/makefile/control-plane.mk` and `../../build/makefile/index.mk`.

| Target | Description |
|--------|-------------|
| `make build` | `openapi-build` then `build-go` — build the OpenAPI bundle + generated code, then compile the Go binary for amd64 + arm64 |
| `make test` | `test-ginkgo` — run the Ginkgo unit-test suites |
| `make lint` | `lint-go` — run golangci-lint in the builder container |
| `make purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `make validate` | `openapi-validate` — validate the OpenAPI spec (yaml + generated-code up-to-date + public routes + service field) |
| `make openapi-lint` | Lint the swagger bundle |
| `make local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` (forces `OS=linux ARCH=amd64`) |

Under the hood `build-go` runs `build-go-amd` + `build-go-arm`, each compiling
`cmd/main.go` inside the golang builder container with `GOFIPS140=v1.0.0` to
`bin/catalog-service-<arch>`.

### Local build (local-dev.sh)

`catalog-service` is a control-plane Go service (listed in `CP_GO_SERVICES`), so
`local-dev.sh` builds it via `build_go_service` in two steps:

1. Compile the amd64 binary in the builder container:
   ```bash
   make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
   ```
   (`DOCKERHUB_PROXY=` pulls the golang builder image from Docker Hub;
   `GITHUB_TOKEN=` is emptied in the make call — the private-module token is
   applied elsewhere.)
2. Build + load the image:
   ```bash
   docker buildx build --platform linux/amd64 --provenance=false --load \
     --build-arg TARGETARCH=amd64 -t runai-local/catalog-service:0.0.0 .
   ```
   Image ref: `runai-local/catalog-service:0.0.0` — tag `0.0.0` is the CP chart
   appVersion; any other tag breaks the local deploy.

For a fast control-plane-only iteration loop:

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```

The [`Dockerfile`](Dockerfile) is a multi-stage build: a
`registry.access.redhat.com/ubi9/ubi-minimal` `base` stage supplies
`/etc/passwd`, `/etc/group`, certs (`/etc/ssl`, `/etc/pki`) and a writable
`/tmp`; the final image is `FROM scratch`, copies in the
`bin/catalog-service-$TARGETARCH` binary, the SQL migrations
(`migrations/up`, `migrations/down`) and the swagger bundle, runs as user
`65532:65532`, and takes `TARGETARCH` as a `--build-arg`.

```bash
docker build --build-arg TARGETARCH=amd64 -t catalog-service:dev .
```

## Testing

| Target | Description |
|--------|-------------|
| `make test` | Run the Ginkgo unit-test suites (`test-ginkgo`) |
| `make lint` | Run golangci-lint |

Unit tests live next to the code they cover (`pkg/services/*_test.go`) and in
`pkg/test/`. Client mocks under `pkg/clients/mocks/` are generated with mockery.

## Configuration

At runtime the service is configured through **environment variables** (supplied
as a ConfigMap/Secret by the control-plane Helm chart). Configuration values are
grouped into typed sections — server, upstream services, NGC, and database — and
each key can equivalently be provided through a loaded config file; any config
key can be overridden by its uppercase environment-variable equivalent. Defaults
below come from the code's `SetDefaults()` methods.

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_PORT` | HTTP port to listen on | `8080` |
| `AUTHORIZATION_SERVICE_URL` | authorization service base URL | `http://runai-backend-authorization:8080` |
| `ASSETS_SERVICE_URL` | assets-service base URL | `http://runai-backend-assets-service:8080` |
| `TEMP_BACKEND_SETTINGS_URL` | temp-backend-settings base URL (settings backend) | `http://runai-backend-temp-backend-settings:8080` |
| `REQUEST_TIMEOUT_SEC` | NGC request timeout, in seconds | `30` |
| `DB_HOST` / `DB_PORT` | PostgreSQL host / port | — |
| `DB_NAME` / `DB_SCHEMA` | database name / schema | — |
| `DB_USER` / `DB_PASSWORD` | database credentials (secret) | — |
| `DB_SSL_MODE` | PostgreSQL SSL mode | — |
| `DB_VERBOSE_LOGGING` | verbose DB query logging | — |
| `MIGRATION_FILES` | path to SQL migration files | — |
| `SERVICE_NAME` | service name reported by the DB layer | `catalog-service` |
| `SERVER_READ_TIMEOUT` / `SERVER_WRITE_TIMEOUT` / `SERVER_IDLE_TIMEOUT` | HTTP server timeouts | — |
| `CACHE_ENABLED` | enable response caching | `false` |
| `CACHE_MAX_RECORD_SIZE_KB` / `CACHE_MAX_STORE_SIZE_MB` / `CACHE_TTL_SECONDS` | cache tuning (when caching is enabled) | `10000` / `10000` / `600` |
| `TRACING_ENABLED` | enable distributed tracing | — |
| `AUDIT_ENABLED` / `AUDIT_ENABLE_PRINT_TO_STDOUT` | enable audit logging / mirror to stdout | — |

The DB-related keys are consumed by the shared `runaidb` database layer; the
service, server and NGC keys are defined in [`pkg/config`](pkg/config/config.go).
The authoritative deploy-time defaults live in the control-plane chart's
`catalog-service` ConfigMap.

## OpenAPI and code generation

The HTTP API is defined in
[`swagger/source/openapi.yaml`](swagger/source/openapi.yaml), split into
per-resource files under `swagger/source/resources` (Images, ImageTags,
Registries, HelmRepos, HelmRepoCharts, HuggingfaceCatalog, Settings, …) and
shared `schemas` / `parameters` / `securitySchemes`. From this source the build
generates the **server stubs** in
[`internal/generated`](internal/generated) — consumed by the router in
`pkg/router`. The bundle is written to `swagger/_build/` and copied to the
central control-plane specs directory.

Generation uses [OpenAPI Generator](https://openapi-generator.tech) (server +
Go-client version pinned to `v7.12.0` in the [`Makefile`](Makefile)). Run
`make build` (or `make openapi-generate-code`) to regenerate; the
`internal/generated` directory **must not be edited by hand** — changes are
overwritten on the next generation run.
</content>
