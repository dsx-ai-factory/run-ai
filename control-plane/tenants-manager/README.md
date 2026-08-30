# Tenants Manager

A Go microservice that manages tenant lifecycle in a multi-tenant deployment. It handles tenant creation, deletion, and configuration, and exposes a REST API generated from an OpenAPI specification.

## Overview

- **REST API** — OpenAPI 3.0 spec in [`swagger/source/openapi.yaml`](./swagger/source/openapi.yaml); generated server stubs live in [`generated/`](./generated/)
- **Database** — PostgreSQL; migrations in [`pkg/database/migrations/`](./pkg/database/migrations/)
- **Auth integration** — delegates identity and realm management to an external identity service (Keycloak-compatible)
- **Pub/sub** — publishes tenant lifecycle events via NATS or Redis streams

## Prerequisites

- Go 1.26+
- Docker and Docker Compose (for integration tests)
- `openapi-generator-cli` v7.12.0 (for regenerating server stubs)

## Build

```sh
make build          # generates OpenAPI stubs + compiles the binary
make build-go       # compiles only (skips OpenAPI generation)
```

The compiled binary is written to `./bin/`.

### Local build (local-dev.sh)

This is a standard control-plane Go service; `local-dev.sh` builds it through
the generic `build_go_service` function, in two steps:

```bash
# Step 1 — compile the amd64 binary in the Go builder container:
make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
# Step 2 — build + load the image:
docker buildx build --platform linux/amd64 --provenance=false --load \
  --build-arg TARGETARCH=amd64 -t runai-local/tenants-manager:0.0.0 .
```

`DOCKERHUB_PROXY=` (empty) pulls the Go builder image from Docker Hub;
`GITHUB_TOKEN=` is emptied in the make call (the private-module token is applied
elsewhere). The image tag is **`0.0.0`** (the control-plane chart appVersion);
any other tag breaks the local deploy. Image ref:
`runai-local/tenants-manager:0.0.0`. For a fast CP-only re-iteration:
`./local-dev.sh --skip-build-rc --skip-cluster`.

## Run locally

Integration tests spin up all dependencies via Docker Compose:

```sh
make run-local
```

This builds the binary and starts the full stack (Postgres, NATS, identity service mock, BFF mock) using the `localdebug` Compose profile.

## Test

```sh
make test           # unit tests + integration tests
make unit-test      # unit tests only
make integ-tests    # integration tests only
```

## Lint and validate

```sh
make lint           # Go linter
make validate       # validates the OpenAPI spec bundle
```

## Configuration

The service is configured via environment variables. Key variables:

| Variable | Description | Default |
|---|---|---|
| `SERVICE_PORT` | HTTP port to listen on | `8080` |
| `DB_HOST` | Postgres host | — |
| `DB_PORT` | Postgres port | `5432` |
| `DB_NAME` | Database name | — |
| `DB_SCHEMA` | Schema name | — |
| `DB_USER` | Database user | — |
| `DB_PASSWORD` | Database password | — |
| `DB_SSL_MODE` | SSL mode (`disable`/`require`) | `require` |
| `MIGRATION_FILES` | Path to migration files | `file:///migrations` |
| `IDENTITY_MNG_URL` | URL of the identity management service | — |
| `AUTH_SERVICE_URL` | URL of the auth service | — |
| `TENANT_DOMAIN_TEMPLATE` | URL template for tenant domains | — |
| `PUBSUB_TYPE` | Pub/sub backend (`nats` or `redis`) | — |
| `PUBSUB_URL` | Pub/sub connection URL | — |
| `BFF_SERVICE_URL` | URL of the BFF service | — |
| `LOG_LEVEL` | Log level (`debug`/`info`/`warn`/`error`) | `info` |
| `TRACING_ENABLED` | Enable OpenTelemetry tracing | `false` |

## API

The full API spec is in [`swagger/source/openapi.yaml`](./swagger/source/openapi.yaml). After building, a bundled spec is also available at `swagger/_build/swagger.yaml`.

To regenerate server stubs after editing the spec:

```sh
make openapi-build
```

## Client package

Other services that need to call this service import the versioned Go client under [`pkg/client/`](./pkg/client/). See the [client README](./pkg/client/README.md) for versioning and release instructions.
