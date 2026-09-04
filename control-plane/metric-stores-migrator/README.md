# metric-stores-migrator

`metric-stores-migrator` is a **run-once migration job** (not a long-running
service) owned by the **Phoenix** team. It copies each tenant's *metric store*
configuration out of `tenants-manager` and re-registers it against the newer
`metrics-service` metric-stores API, translating the data from the legacy
tenants-manager schema to the new metrics-service schema along the way. It reads
the old data but never modifies it, so the migration is safe to re-run.

The job is executed once by the control-plane Helm chart (as a Kubernetes
`Job`) during install/upgrade. It has **no HTTP API and no database of its
own** — it is purely an ETL step between two other control-plane services.

## Contents

- [Overview](#overview)
- [What it migrates](#what-it-migrates)
- [When it runs](#when-it-runs)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)

## Overview

A *metric store* describes where a tenant's metrics are read from and written to
(the datasource id, the store `type`, and the read/write endpoints with their
URLs, headers and optional basic-auth credentials). Historically this
configuration lived inside `tenants-manager`; it now belongs to
`metrics-service`. This job bridges that move.

For every tenant it:

- **Reads tenants** — calls the internal `tenants-manager` API to list all
  tenants.
- **Reads each tenant's metric store** — fetches the tenant's extended metric
  store from `tenants-manager`. If a tenant has no metric store configured, the
  job creates a default *provisioned* metric store for that tenant id instead.
- **Translates the schema** — maps the tenants-manager metric-store shape
  (type, datasource id, read/write endpoints, endpoint auth/basic credentials)
  onto the `metrics-service` `MetricStore` model.
- **Writes to metrics-service** — calls `CreateMetricStore` on the
  metrics-service client. A `409 Conflict` response is treated as *already
  migrated* and counted, not failed.

At the end the job logs a summary (`succeed`, `already migrated`, `failed`
counts). It handles `SIGINT` gracefully, cancelling in-flight work before
exiting.

## What it migrates

| Source (tenants-manager) | Destination (metrics-service) |
|---|---|
| Tenant list (internal API) | — (iteration only) |
| `ExtendedMetricStore` (type, datasource id, read/write endpoints, endpoint headers, basic auth) | `MetricStore` (per tenant id) via `CreateMetricStore` |
| *no metric store for tenant* | default provisioned `MetricStore` for the tenant id |

No changes are made to the old tenants-manager data — the migration is read-only
on the source side and idempotent on the destination side (re-runs conflict and
are skipped).

## When it runs

The job is deployed and executed **once** by the control-plane Helm chart during
install/upgrade. Because it is a Kubernetes `Job`, it will not re-run
automatically on a subsequent `helm upgrade`. To force a re-run, delete the
completed job first:

```bash
kubectl delete job runai-backend-metric-stores-migrator
```

Since the write side returns `409 Conflict` for tenants that already have a
metric store, re-running is safe and simply reports those tenants as *already
migrated*.

## Repository layout

```text
.
├── cmd/
│   └── main.go             # Entry point: registers config, builds clients, runs the migrator
├── internal/
│   ├── migrator.go         # Migrate loop, per-tenant translation, status counters
│   ├── tenants_client.go   # tenants-manager client + TenantsConfiguration
│   ├── migrator_test.go    # Unit tests
│   └── mocks/              # mockery-generated test doubles (do not edit)
├── .mockery.yml            # mockery config for the interfaces mocked in tests
├── Dockerfile
├── Makefile
└── go.mod
```

## Building and running

The project is built with Go (see [`go.mod`](go.mod)) and orchestrated through
the [`Makefile`](Makefile), which pulls in the shared control-plane build
includes (`../../build/makefile/control-plane.mk`, `index.mk`, and the
`golang.mk` / `base.mk` they include).

| Target | Description |
|--------|-------------|
| `make build` | `build-go` — compile the amd64 + arm64 binaries |
| `make build-go` | `build-go-amd` + `build-go-arm` |
| `make build-go-amd` | Compile `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) in the Go builder container → `bin/metric-stores-migrator-amd64` |
| `make build-go-arm` | Same for `arm64` → `bin/metric-stores-migrator-arm64` |
| `make test` | `build` + `test-go` (unit tests) |
| `make lint` | `lint-go` — golangci-lint in the builder container |
| `make purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `make validate` | No-op (this job has no OpenAPI spec to validate) |
| `make generate-mocks` | Regenerate the mockery mocks under `internal/mocks` |
| `make local-deploy` | `build-go` + `docker-build-generic` + `docker-push-generic` (amd64) |
| `make docker-build-generic` | Build the container image |
| `make docker-push-generic` | Push the container image |

The [`Dockerfile`](Dockerfile) is a two-stage build. A
`registry.access.redhat.com/ubi9/ubi-minimal` stage supplies `/etc/passwd`,
`/etc/group`, TLS certs (`/etc/ssl`, `/etc/pki`), `nsswitch.conf`, and a writable
`/tmp`. The final stage is `FROM scratch`, copies in the
`bin/metric-stores-migrator-$TARGETARCH` binary (renamed `metric_stores_migrator`)
and `NOTICE`, runs as user `65532:65532`, and sets the binary as `ENTRYPOINT`.
`TARGETARCH` is the only build arg.

```bash
docker build --build-arg TARGETARCH=amd64 -t metric-stores-migrator:dev .
```

### Local build (local-dev.sh)

This service is in `CP_GO_SERVICES`, so `local-dev.sh` builds it with the
standard `build_go_service` flow:

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
     --build-arg TARGETARCH=amd64 -t runai-local/metric-stores-migrator:0.0.0 .
   ```

Image ref: `runai-local/metric-stores-migrator:0.0.0` (tag `0.0.0` is the
control-plane chart appVersion — any other tag breaks the local deploy).

To iterate on control-plane only:

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```

## Testing

| Target | Description |
|--------|-------------|
| `make test` | `build` then `test-go` — compile and run the Go unit tests |
| `make test-go` | Run the Go unit tests in the builder container |
| `make generate-mocks` | Regenerate mockery mocks used by the tests |

Unit tests cover the migration loop and schema translation using the
mockery-generated `TenantsClienter` and `MetricStoresClienter` mocks — no live
`tenants-manager` or `metrics-service` is required.

## Configuration

The job is **environment-variable driven**. Configuration keys are declared as
struct fields with defaults; each key can be overridden by its uppercase
environment-variable equivalent.

| Variable | Description | Default |
|----------|-------------|---------|
| `TENANT_MANAGER_URL` | Base URL of the `tenants-manager` service (source of tenants + legacy metric stores) | `http://runai-backend-tenants-manager:8080` |
| `METRICS_SERVER_HOST` | Host (`host:port`) of the `metrics-service` metric-stores API (write target) | `runai-backend-metrics-service:8080` |
| `METRICS_SERVER_SCHEME` | Scheme used to reach `metrics-service` | `http` |

All work is scoped under tenant context id `1`. There are no ports to expose and
no persistent storage — the process starts, migrates every tenant, logs a
summary, and exits.
