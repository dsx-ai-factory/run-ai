# Backend for Frontend Service

The BFF (Backend for Frontend) service is the API gateway between the Run:ai frontend and the backend configuration services. It consolidates reads/writes for UI-facing settings so the frontend has a single, well-typed surface to talk to.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Settings model](#settings-model)
- [Required backing services](#required-backing-services)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [OpenAPI and code generation](#openapi-and-code-generation)
- [Go client](#go-client)

## Overview

The service is responsible for:

- **Frontend boot configuration** — exposing deployment-resolved, browser-safe UI metadata through the internal, unauthenticated, browser-accessible `GET /api/v1/boot-config` endpoint.
- **Settings** — reading and writing tenant-scoped UI settings (branding, notifications, workload defaults, appearance, system banners) stored in its own PostgreSQL database.
- **Proxying** — forwarding requests to `temp-backend-settings` and `tenants-manager` for settings it does not own directly.
- **Authorization** — enforcing access control on protected endpoints via the shared authorization package.

Boot configuration is deployment-global and intentionally unauthenticated because the frontend needs it before login and before application state exists. Only values safe to expose to every browser belong in this contract. Credentials, Kubernetes Secrets, tenant-specific values, and the deployment type itself must not be returned. The Unleash client key is a browser-public frontend token and is treated as ordinary deployment configuration.

The API surface is described in [`swagger/source/openapi.yaml`](swagger/source/openapi.yaml) and bundled into [`swagger/_build/swagger.yaml`](swagger/_build/swagger.yaml).

## Repository layout

```text
.
├── cmd/                        # Service entry point (main.go)
├── pkg/
│   ├── authorization/          # Authorization helpers
│   ├── client/                 # Versioned Go client (see pkg/client/README.md)
│   ├── clients/                # HTTP clients for upstream services
│   ├── database/               # DAO layer and SQL migrations
│   ├── services/               # Domain service implementations
│   ├── settings/               # Typed setting definitions (branding, notifications, …)
│   ├── system-notification/    # System-wide notification banner logic
│   └── utils/
├── internal/
│   ├── auth/                   # Auth middleware
│   ├── config/                 # Configuration loading and defaults
│   ├── generated/              # OpenAPI-generated server stubs (do not edit)
│   ├── router/                 # HTTP router wiring
│   └── services/               # Internal service helpers
├── integration-tests/          # Docker-compose-driven integration tests
├── swagger/                    # OpenAPI source and built bundle
├── Dockerfile
├── Makefile
└── go.mod
```

## Settings model

The BFF owns five settings categories, each stored in its own database table and exposed via the OpenAPI spec:

| Category | Description |
|---|---|
| `branding` | Logo, colour theme, and white-labelling overrides |
| `notifications` | Email / alert notification preferences |
| `ui-workloads` | Workload-list UI defaults (columns, filters, pagination) |
| `ui-appearance` | General UI appearance preferences |
| `system-notification` | System-wide banners displayed to all users |

## Required backing services

| Service | Default address | Override env var |
|---|---|---|
| temp-backend-settings | `http://runai-backend-temp-backend-settings:8080` | `BACKEND_SERVICE_URL` |
| tenants-manager | `http://runai-backend-tenants-manager:8080` | `TENANT_MANAGER_URL` |
| PostgreSQL | configured via `DATABASE_*` env vars | see `integration-tests/config/` |

## Building and running

The project is built with Go (see [`go.mod`](go.mod) for the required version) and orchestrated through the [`Makefile`](Makefile).

| Target | Description |
|---|---|
| `make build` | Validate OpenAPI, generate code, build the binary |
| `make run-local` | Build and run locally via `docker-compose` (mocks enabled) |
| `make lint` | Run the Go linter |
| `make purify` | Auto-fix formatting and imports |
| `make validate` | Validate the OpenAPI specification |

## Testing

| Target | Description |
|---|---|
| `make unit-test` | Build and run unit tests |
| `make test` | Unit tests + integration tests |
| `make docker-compose-integration-tests-up` | Run the full docker-compose test suite |
| `make docker-compose-integration-tests-debug` | Same, but without the failure handler |
| `make coverage` | Open `c.out` as an HTML coverage report |

## OpenAPI and code generation

The HTTP API is defined in [`swagger/source/openapi.yaml`](swagger/source/openapi.yaml). From this source the build generates:

- The **server stubs** in [`internal/generated`](internal/generated) — consumed by the router.
- The **Go client** in [`pkg/client`](pkg/client).

Both `generated/` directories are produced by [OpenAPI Generator](https://openapi-generator.tech) (version pinned in the `Makefile`) and **must not be edited by hand**.

## Go client

Other Run:ai services consume the BFF through the versioned Go client at [`pkg/client`](pkg/client). See [`pkg/client/README.md`](pkg/client/README.md) for:

- How to import the versioned client package
- How to create development and production version tags
- The `make client-bump-version` and `make client-bump-version-dev` workflow
