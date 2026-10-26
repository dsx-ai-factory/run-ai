# Temp Backend Settings Service

`temp-backend-settings` is a control-plane HTTP service (owned by the **Suricata**
team) that serves tenant-level feature flags and configuration settings. It is a
**transitional** service: its logic was migrated from the legacy TypeScript
backend's `SettingController` and `InternalTenantController`, and it exists to
keep the original settings API alive while individual settings are moved into
their owning Go domain services. Other services — notably
[`bff-service`](../bff-service) — proxy settings reads/writes to it at
`http://runai-backend-temp-backend-settings:8080` (see `bff-service`'s
`BACKEND_SERVICE_URL`).

> **Do not add new public API endpoints here.** This service exists solely to
> support the migration of existing settings. New settings/APIs belong in the
> appropriate domain service. Once the product versions that depend on the
> original settings API reach end-of-support, this service will be decommissioned.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [HTTP API](#http-api)
- [Required backing services](#required-backing-services)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [OpenAPI and code generation](#openapi-and-code-generation)
- [Go client](#go-client)

## Overview

The service is responsible for:

- **Tenant settings storage** — persisting tenant-scoped settings (feature
  flags and configuration) in its own PostgreSQL `tenant_settings` table via the
  `TenantSettingsDAO`.
- **The legacy settings API** — serving the deprecated `GET`/`PUT`
  `/v1/k8s/setting` surface that older clients still call, plus per-key lookups.
- **Setting-key translation** — mapping the public setting names (e.g.
  `EnableModelCatalog`, `DisableAudit`, `EnableDataVolumes`) to their stored
  dotted DB keys (e.g. `tenant.enable_model_catalog`) — see
  [`pkg/types/tenant_setting_keys.go`](pkg/types/tenant_setting_keys.go).
- **Cross-service enrichment** — resolving effective setting values by calling
  peer control-plane services (cluster-service, tenants-manager,
  identity-manager, bff-service, org-unit-service, catalog-service,
  datavolumes-service, workloads, authorization) through their generated Go
  clients. The `pkg/business/providers` + `pkg/business/adapters` layers wrap
  those calls.
- **Internal all-tenant lookup** — an internal-only endpoint that returns a
  setting value for every tenant, consumed by other backend components.

The API surface is defined in
[`swagger/source/openapi.yaml`](swagger/source/openapi.yaml) and bundled into
[`swagger/_build/swagger.yaml`](swagger/_build/swagger.yaml).

## Repository layout

```text
.
├── cmd/                        # Service entry point (main.go): init config/db/clients, serve router
├── pkg/
│   ├── autorization/           # Authorization helper for protected routes
│   ├── business/
│   │   ├── adapters/           # Thin adapters over peer-service generated clients
│   │   ├── providers/          # Higher-level providers that resolve setting values
│   │   └── settings/           # Settings domain logic, validators, version utils
│   ├── client/                 # Versioned Go client (separate go.mod)
│   ├── config/                 # ServicesConfig (peer-service URLs, flags)
│   ├── dao/                    # tenant_settings DAO
│   ├── database/migrations/    # SQL migrations (up/ + down/), bundled into the image
│   ├── router/                 # chi router wiring (health + tenant + internal)
│   ├── services/               # API service implementations (tenant + internal)
│   └── types/                  # Public→DB setting-key lookup table
├── internal/
│   ├── generated/              # OpenAPI-generated server stubs + spec copy (do not edit)
│   └── services/               # Internal helpers (health API service)
├── swagger/                    # OpenAPI source and built bundle
├── Dockerfile
├── Makefile
└── go.mod
```

## HTTP API

| Method & path | Operation | Notes |
|---|---|---|
| `GET /health` | `health` | Unauthenticated liveness/readiness check |
| `GET /v1/k8s/setting` | `getTenantSettings` | **Deprecated.** All tenant settings; `excludeMigrated` query param |
| `PUT /v1/k8s/setting` | `updateTenantSetting` | **Deprecated.** Update a setting (`settings:update` authorization) |
| `GET /v1/k8s/setting/{settingKey}` | `getTenantSettingsByKey` | **Deprecated.** Single setting by key |
| `GET /internal/tenants/all-tenants/setting/{key}` | `getSettingByKeyForAllTenants` | Internal-only; value of one key across all tenants |

Tenant-scoped routes carry `x-scope: [system, tenant]`; the deprecated routes
point clients to the migrated `Settings` API for anything that has already moved
to its domain service.

## Required backing services

The service reaches peer control-plane services (defaults from
[`pkg/config/types.go`](pkg/config/types.go), all overridable — see
[Configuration](#configuration)) and a PostgreSQL database:

| Service | Default address | Config key |
|---|---|---|
| cluster-service | `http://runai-backend-cluster-service:8080` | `cluster_service_url` |
| tenants-manager | `http://runai-backend-tenants-manager:8080` | `tenant_manager_url` |
| identity-manager | `http://runai-backend-identity-manager:8080` | `identity_manager_url` |
| bff-service | `http://runai-backend-bff-service:8080` | `bff_service_url` |
| org-unit-service | `http://runai-backend-org-unit-service:8080` | `org_unit_service_url` |
| workloads | `http://runai-backend-workloads:8080` | `workloads_service_url` |
| catalog-service | `http://runai-backend-catalog-service:8080` | `catalog_service_url` |
| datavolumes-service | `http://runai-backend-datavolumes-service:8080` | `datavolumes_service_url` |
| authorization | `http://runai-backend-authorization:8080` | `authorization_service_url` |
| PostgreSQL | configured via `DATABASE_*` env vars | — |

## Building and running

The project is built with Go (see [`go.mod`](go.mod)) and orchestrated through
the [`Makefile`](Makefile), which includes the shared
`../../build/makefile/control-plane.mk` and `index.mk` (which pull in `golang.mk`
and `openapi.mk`).

| Target | Description |
|---|---|
| `make build` | `openapi-build` then `build-go` — build the OpenAPI bundle + generate code, then compile the amd64 + arm64 binaries |
| `make build-go` | `build-go-amd` + `build-go-arm` |
| `make build-go-amd` | Compile `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) in the Go builder container → `bin/temp-backend-settings-amd64` |
| `make build-go-arm` | Same for `arm64` → `bin/temp-backend-settings-arm64` |
| `make generate-mocks` / `make generate-mocks-local` | Regenerate test mocks from `.mockery.yml` with the shared mockery v3 setup (in Docker / without Docker) |
| `make generate-mocks-client` | Regenerate `pkg/client` mocks from `pkg/client/.mockery.yml` |
| `make purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `make validate` | `openapi-validate` — validate spec + generated code are up to date |
| `make openapi-lint` | Lint the swagger bundle |
| `make local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` |

### Local build (local-dev.sh)

`temp-backend-settings` is in `CP_GO_SERVICES` in
[`local-dev.sh`](../../local-dev.sh), so its `build_go_service` path runs:

```bash
# 1. Compile the amd64 binary in the Go builder container
make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=

# 2. Build + load the image into the local store / Kind nodes
docker buildx build --platform linux/amd64 --provenance=false --load \
  --build-arg TARGETARCH=amd64 -t runai-local/temp-backend-settings:0.0.0 .
```

`DOCKERHUB_PROXY=` (empty) pulls the golang builder image straight from Docker
Hub; `GITHUB_TOKEN=` is emptied in the make call (the private-module token is
applied elsewhere). The image tag **`0.0.0`** is the control-plane chart
`appVersion` — any other tag breaks the local deploy. Image ref:
`runai-local/temp-backend-settings:0.0.0`.

To re-iterate on control-plane only (reuse the running cluster):

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```

The [`Dockerfile`](Dockerfile) is a two-stage build: a
`registry.access.redhat.com/ubi9/ubi-minimal` `base` stage supplies `/etc/passwd`,
`/etc/group`, TLS certs (`/etc/ssl`, `/etc/pki`) and a writable `/tmp`; the final
image is `FROM scratch`, runs as user `65532:65532`, and bundles the compiled
binary, the SQL migrations (`pkg/database/migrations/up|down`), and the swagger
bundle. It takes a `TARGETARCH` build arg selecting `bin/temp-backend-settings-$TARGETARCH`.

## Testing

| Target | Description |
|---|---|
| `make test` | `build-go` + `test-go` — compile then run the Go unit tests in the builder container |
| `make lint` | `lint-go` — golangci-lint in the builder container |

Unit tests live next to the code (`*_test.go` under `pkg/business`,
`pkg/services`, `pkg/autorization`). There is no docker-compose integration
suite in this service.

## Configuration

Runtime configuration is loaded into the typed `ServicesConfig`
([`pkg/config/types.go`](pkg/config/types.go)); the fields are the peer-service
URLs listed under [Required backing services](#required-backing-services) plus:

| Config key | Description | Default |
|---|---|---|
| `show_alpha_flags` | Whether alpha feature flags are returned | `false` |

`SetDefaults()` provides the in-cluster defaults; any key may be overridden by
its uppercase environment-variable equivalent. Database connectivity and the
HTTP server come from the shared platform packages and are driven by their
standard environment variables:

| Variable | Description |
|---|---|
| `DATABASE_*` | PostgreSQL host / port / name / user / password (shared DB package) |

## OpenAPI and code generation

The HTTP API is defined in
[`swagger/source/openapi.yaml`](swagger/source/openapi.yaml), split into
`resources/` (tenant + internal), `parameters/`, and `schemas/`. `make build`
runs `openapi-build` first, which bundles the spec and generates:

- The **server stubs** in [`internal/generated`](internal/generated) — consumed
  by [`pkg/router`](pkg/router/router.go).
- The **Go client** in [`pkg/client/generated`](pkg/client/generated).

Generation uses [OpenAPI Generator](https://openapi-generator.tech) `v7.12.0`
(pinned in the [`Makefile`](Makefile)). Both `generated/` directories **must not
be edited by hand** — changes are overwritten on the next generation run.

## Go client

A versioned Go client (its own `go.mod`) lives at
[`pkg/client`](pkg/client) for other services that need to call this service.
The generated client sits under `pkg/client/generated`, wrapped by
`pkg/client/api.go`.
