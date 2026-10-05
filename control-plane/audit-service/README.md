# Audit Service

Audit Service is the control-plane service that records and serves the Run:ai
**audit trail** — the immutable log of security- and configuration-relevant
actions performed across the platform (who did what, to which resource, when,
and from where). It ingests audit events emitted by other services over the
shared pub/sub bus, batches and persists them to its own PostgreSQL store, and
exposes an authorized HTTP/REST API (generated from an OpenAPI specification)
for querying and exporting those records. The service is owned by the
**Phoenix** team.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [OpenAPI and code generation](#openapi-and-code-generation)
- [Go client](#go-client)
- [Further reading](#further-reading)

## Overview

The service is responsible for:

- **Ingesting audit events** — a pub/sub consumer subscribes to audit messages
  published by other services (`internal/ingest`, `internal/pubsub`). Incoming
  records flow through a handler chain into a **batcher** that accumulates them
  and flushes to the database sink on a size/interval trigger, so bursts of
  events become efficient bulk inserts.
- **Persisting audit logs** — the DAO layer (`internal/database`) writes records
  to the `audit_logs` table in the service's PostgreSQL schema, with SQL
  migrations bundled into the image.
- **Serving the audit trail** — an authorized REST API (`internal/services`,
  wired in `internal/router`) lets callers list and filter audit records with
  pagination, and **export** them as CSV (via `gocsv`). Authorization is
  enforced through the shared authorization client. A separate internal API is
  used by other services to submit audit records directly.
- **Retention / cleanup** — a scheduled DB cleaner (`internal/database/cleaner.go`)
  periodically deletes records older than a configurable TTL, so the audit table
  does not grow unbounded.
- **Tenant lifecycle** — a subscriber reacts to tenant-delete events and removes
  the corresponding tenant's audit data.

The API surface is described in
[`swagger/source/openapi.yaml`](swagger/source/openapi.yaml) and bundled into
[`swagger/_build`](swagger/_build).

## Repository layout

```text
.
├── cmd/                        # Service entry point (main.go)
├── internal/
│   ├── auth/                   # Authorization helpers / mock service
│   ├── config/                 # Typed config (service + batcher) with defaults
│   ├── database/               # Audit-log DAO, migrations (up/down), retention cleaner
│   ├── generated/              # OpenAPI-generated server stubs (do not edit)
│   ├── ingest/                 # Pub/sub consumer, handler chain, batcher, DB sink
│   ├── pubsub/                 # Event subscribers (e.g. tenant delete)
│   ├── router/                 # HTTP router wiring + middleware
│   └── services/               # API service implementations + CSV export
├── pkg/
│   └── client/                 # Versioned Go client (see pkg/client/README.md)
│       └── generated/          # OpenAPI-generated client (do not edit)
├── integration-tests/          # Docker-compose-driven integration tests
├── swagger/                    # OpenAPI source and built bundle
├── Dockerfile
├── Makefile
└── go.mod
```

## Building and running

The project is built with Go (see [`go.mod`](go.mod) for the required version)
and orchestrated through the [`Makefile`](Makefile), which inherits the shared
`build/makefile/control-plane.mk` and `index.mk` includes (which in turn pull in
`golang.mk` and `openapi.mk`).

| Target | Description |
|--------|-------------|
| `make build` | `openapi-build` then `build-go` — build the OpenAPI bundle + generate code, then compile the Go binary for amd64 + arm64 |
| `make test` | `build` + `test-go` + `integration-tests` |
| `make lint` | `lint-go` — run golangci-lint in the builder container |
| `make purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `make validate` | `openapi-validate` — validate the OpenAPI spec |
| `make openapi-lint` | Lint the built swagger bundle |
| `make generate-mocks` / `make generate-mocks-local` | Regenerate the service mocks from `.mockery.yml` with mockery v3 (in Docker / without Docker) |
| `make generate-mocks-client` | Regenerate the `pkg/client` mocks from `pkg/client/.mockery.yml` |
| `make build-integration-tests` | Compile the integration test binary + `docker compose build` |
| `make integration-tests` | Run the full docker-compose integration suite (abort-on-container-exit) |
| `make integration-tests-debug` | Bring up the compose stack without the tests container (run tests from the IDE) |

Under the hood `build-go` runs `build-go-amd` + `build-go-arm`, each compiling
`cmd/main.go` inside the Go builder container with `GOFIPS140=v1.0.0` to
`bin/audit-service-<arch>`.

### Dockerfile

The image is a two-stage build. A `registry.access.redhat.com/ubi9/ubi-minimal`
`base` stage supplies `/etc/passwd`, `/etc/group`, TLS certs (`/etc/ssl`,
`/etc/pki`) and a writable `/tmp`. The final stage is `FROM scratch` and copies
in the prebuilt binary `bin/audit-service-$TARGETARCH`, the SQL migrations
(`internal/database/migrations/up|down`), the swagger bundle, and `NOTICE`. It
runs as user `65532:65532`. The only build arg is `TARGETARCH` (selects the
architecture of the copied binary).

```bash
docker build --build-arg TARGETARCH=amd64 -t audit-service:dev .
```

### Local build (local-dev.sh)

`audit-service` is a control-plane Go service, so `local-dev.sh` builds it via
its `build_go_service` path in two steps:

1. Compile the amd64 binary in the builder container:
   ```bash
   make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
   ```
   (`DOCKERHUB_PROXY=` pulls the Go builder image straight from Docker Hub;
   `GITHUB_TOKEN=` is emptied in the make call — the private-module token is
   applied elsewhere.)
2. Build and load the image:
   ```bash
   docker buildx build --platform linux/amd64 --provenance=false --load \
     --build-arg TARGETARCH=amd64 -t runai-local/audit-service:0.0.0 .
   ```

The image ref is `runai-local/audit-service:0.0.0`. The tag **must** be `0.0.0`
(the control-plane chart appVersion) — any other tag breaks the local deploy.

To iterate on just the control plane against an existing cluster:

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```

## Testing

| Target | Description |
|--------|-------------|
| `make test` | `build` + `test-go` (Go unit tests) + `integration-tests` |
| `make build-integration-tests` | Build the `integration` test binary and `docker compose build` the test images |
| `make integration-tests` | Bring up the compose stack (PostgreSQL, pub/sub, the service) and run the suite, exiting on container exit |
| `make integration-tests-debug` | Same stack without the tests container, for debugging from an IDE |

Integration tests require Docker Compose; the stack stands up PostgreSQL and a
Redis-backed pub/sub as backing services (see
[`integration-tests/docker-compose.yaml`](integration-tests/docker-compose.yaml)).

## Configuration

Audit Service is **environment-variable driven**. Typed configuration blocks
(`internal/config/config.go`) declare sensible defaults, and each key can be
overridden by its **uppercase environment-variable equivalent** (e.g. the
`batch_max_size` key is set via `BATCH_MAX_SIZE`). Database, pub/sub, HTTP,
logging, metrics and tracing come from the shared Run:ai packages and read their
own environment variables. Database credentials are read from files under the
directory named by `SECRETS_MOUNT_PATH`.

| Variable | Description | Default |
|----------|-------------|---------|
| `LOG_LEVEL` | Log verbosity | `info` |
| `SERVICE_NAME` | Service identity (metrics/tracing) | `audit-service` |
| `DB_HOST` | PostgreSQL host | — |
| `DB_PORT` | PostgreSQL port | — |
| `DB_NAME` | Database name | — |
| `DB_SCHEMA` | Schema for audit tables | `audit_service` |
| `SECRETS_MOUNT_PATH` | Directory holding `DB_USER` / `DB_PASSWORD` secret files | — |
| `PUBSUB_TYPE` | Pub/sub backend (e.g. `redis`, `nats`) | — |
| `PUBSUB_URL` | Pub/sub connection URL | — |
| `PUBSUB_SUBSCRIBE_RETRIES` | Subscribe retry count | — |
| `BATCH_MAX_SIZE` | Max records per DB flush batch | `10` |
| `BATCH_INTERVAL_SEC` | Max seconds before a partial batch flushes | `10` |
| `AUDIT_RECORD_CLEANER_ENABLED` | Enable the retention cleaner | `true` |
| `AUDIT_RECORD_CLEANER_INTERVAL_HOURS` | Cleaner run interval | `24` |
| `AUDIT_RECORD_TTL_DAYS` | Delete audit records older than this | `90` |
| `USE_MOCKS` | Use a mock authorization client (tests/local) | `false` |

See [`integration-tests/config/config.yaml`](integration-tests/config/config.yaml)
for a working example of these values.

## OpenAPI and code generation

The HTTP API is defined in
[`swagger/source/openapi.yaml`](swagger/source/openapi.yaml) (with per-resource
schemas, parameters, and error definitions under `swagger/source`). From this
source the build generates:

- The **server stubs** in [`internal/generated`](internal/generated) — consumed
  by [`internal/router`](internal/router).
- The **Go client** in [`pkg/client/generated`](pkg/client/generated), wrapped
  by hand-written helpers under [`pkg/client`](pkg/client).

Both `generated/` directories are produced by
[OpenAPI Generator](https://openapi-generator.tech) (`v7.12.0`, with
`VERSIONED_CLIENT=true`) via `make build` (which runs `openapi-build` first) and
**must not be edited by hand** — changes are overwritten on the next generation
run.

## Go client

Other Run:ai services consume Audit Service through the versioned Go client at
[`pkg/client`](pkg/client). See [`pkg/client/README.md`](pkg/client/README.md)
for how to import the versioned client package and the
`make client-bump-version` / `make client-bump-version-dev` workflow.

## Further reading

- [`pkg/client/README.md`](pkg/client/README.md) — client usage and versioning policy
