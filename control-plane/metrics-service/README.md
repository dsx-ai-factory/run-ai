# Metrics Service

Metrics Service is the control-plane service (owned by the **Phoenix** team) that answers
time-series metric queries for Run:AI and produces scheduled usage reports. It sits between
the platform's UI/API callers and the per-tenant metric stores (Thanos, Mimir, or Grafana
Cloud/Grafana Labs), translating logical metric names into backend PromQL, injecting the
correct filters/ranges/quantiles, and picking the right query variant for the cluster's Run:AI
version. It also owns the lifecycle of **metric stores** (the tenant→datasource mapping) and
runs a background **reports** pipeline that collects metrics on a schedule and stores CSV
reports in S3. The HTTP/REST API is generated from an OpenAPI specification.

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

- **Metric queries** — resolving logical metric names into backend PromQL and executing them
  against the tenant's datasource. A query pipeline injects simple filters, aggregation ranges,
  and quantiles (`FILTERS`, `RANGE`, `QUANTILE` placeholders), and a **query-version picker**
  selects the query variant that matches the target cluster's Run:AI version.
- **Metric stores** — managing the per-tenant mapping from tenant to datasource. A datasource
  has a `type` (`thanos`, `mimir`, or `grafanalabs`) and read/write endpoints; the manager
  subscribes to tenant-delete events (via pub/sub) to clean up stores when a tenant is removed.
- **Datasources & connections** — a connection manager builds and caches clients to the
  configured metric backends and issues range/instant queries against them.
- **Reports** — a background **processor** collects metrics on a fixed interval, a **collector**
  builds the report data, results are written to **S3** as CSV (via `gocsv`), and a **deleter**
  removes reports past their TTL. Report state and retries are tracked in the service's own
  PostgreSQL database and driven by a cron-jobs configuration.
- **Cleaner** — an optional background job that purges old report rows from the database.

It depends on **cluster-service** (to resolve cluster versions/names) and **tenants-manager**
(tenant metadata), calls **audit-service** for audit events, and can run entirely against mocks
(`use_mocks: true`) for local development. The API surface is described in
[`swagger/source/openapi.yaml`](swagger/source/openapi.yaml) and bundled into
[`swagger/_build/`](swagger/_build).

## Repository layout

```text
.
├── cmd/                        # Service entry point (main.go)
├── internal/
│   ├── clusters/               # cluster-service client (versions, names)
│   ├── config/                 # Configuration loading and registration
│   ├── connections/            # Datasource connection manager
│   ├── database/               # DAO layer, SQL migrations, cleaner + cron jobs
│   ├── datasources/            # Datasource client (range/instant queries)
│   ├── errors/                 # Typed errors
│   ├── generated/              # OpenAPI-generated server stubs (do not edit)
│   ├── grafana/                # Grafana / Grafana Labs integration
│   ├── metrics/                # Query builder, injectors, version/query pickers
│   ├── metricstores/           # Metric-store manager, datasource types
│   ├── providers/              # Datasource provider (static / dynamic)
│   ├── reports/                # Reports processor, collector, deleter, S3 client
│   ├── routers/                # HTTP router wiring
│   ├── services/               # API service implementations
│   └── tenants/                # tenants-manager client
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

The project is built with Go (see [`go.mod`](go.mod) for the required version) and orchestrated
through the [`Makefile`](Makefile), which includes the shared control-plane makefiles
(`../../build/makefile/control-plane.mk` and `index.mk`).

| Target | Description |
|--------|-------------|
| `make build` | `openapi-build` then `build-go` — build the OpenAPI bundle + generate code, then compile the Go binaries for amd64 and arm64 |
| `make build-go` | `build-go-amd` + `build-go-arm` — both arch binaries |
| `make build-go-amd` | Compile `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) in the Go builder container → `bin/metrics-service-amd64` |
| `make build-go-arm` | Same for `arm64` → `bin/metrics-service-arm64` |
| `make lint` | `lint-go` — run golangci-lint in the builder container |
| `make purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `make validate` | `openapi-validate` — validate the OpenAPI spec (yaml + generated code up-to-date + public routes) |
| `make openapi-lint` | Lint the swagger bundle |
| `make generate-mocks` / `make generate-mocks-local` | Regenerate test mocks with mockery v3 from `.mockery.yml` (in the Go builder container / without Docker) |
| `make generate-mocks-client` | Regenerate `pkg/client` mocks from `pkg/client/.mockery.yml` |
| `make local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` |

### Local build (local-dev.sh)

`metrics-service` is a control-plane Go service (listed in `CP_GO_SERVICES`), so
`$REPO_ROOT/local-dev.sh` builds it in two steps:

```bash
# 1) Compile the amd64 binary inside the Go builder container
make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=

# 2) Build + load the image
docker buildx build --platform linux/amd64 --provenance=false --load \
  --build-arg TARGETARCH=amd64 -t runai-local/metrics-service:0.0.0 .
```

`DOCKERHUB_PROXY=` (empty) pulls the golang builder image straight from Docker Hub;
`GITHUB_TOKEN=` is emptied in the make call (the private-module token is applied elsewhere).
The image tag **must** be `0.0.0` — it is the control-plane chart `appVersion`, and any other
tag breaks the local deploy. Image ref: `runai-local/metrics-service:0.0.0`.

The Dockerfile is a two-stage build: a `registry.access.redhat.com/ubi9/ubi-minimal` stage
supplies `/etc/passwd`, `/etc/group`, TLS certs (`/etc/ssl`, `/etc/pki`) and a writable `/tmp`;
the final image is `FROM scratch`, runs as user `65532:65532`, and bundles the compiled binary,
the SQL migrations (`internal/database/migrations/up|down`), the swagger bundle, and `NOTICE`.
It takes a single `--build-arg TARGETARCH` (selects `bin/metrics-service-$TARGETARCH`).

## Testing

| Target | Description |
|--------|-------------|
| `make test` | `build` + `test-go` (unit tests) + `integration-tests` |
| `make integration-tests` | Build the promtool agent + integration-test binary, then run the full docker-compose suite (abort-on-container-exit, torn down afterwards) |
| `make integration-tests-debug` | Bring up only the compose `run-setup` profile so tests can be run from an IDE |
| `make build-integration-tests` | Compile the `integration-tests/tests` test binary and `docker compose build` |

The integration suite (see [`integration-tests/`](integration-tests)) uses docker-compose to
stand up a Prometheus instance (with a bundled promtool agent), PostgreSQL, and supporting
services; the compose file runs with `AGENT_PORT=8090`.

## Configuration

The service is **config-file driven**: `cmd/main.go` loads a YAML file whose path is given by
the `CONF_PATH` environment variable, and reads Kubernetes secrets from a secrets folder. The
struct tags in `internal/config` and the sub-packages are **config-file keys**, not env vars.
Any key can be overridden by its uppercase environment-variable equivalent. The env-var table
below therefore holds only bootstrap variables; database, S3, pub/sub, tracing, and metrics
settings are supplied by the shared runtime packages via the same config file / secrets.

| Variable | Description | Default |
|----------|-------------|---------|
| `CONF_PATH` | Path to the YAML config file to load | — |

A representative (trimmed) config — see [`integration-tests/config.yaml`](integration-tests/config.yaml)
as the source of truth:

```yaml
log_level: debug            # service log level
use_mocks: true             # run against mocked cluster/tenant clients (local/dev)
reports:                    # background reports pipeline
  processor_run_interval_sec: 2
  s3_config:
    bucket: metrics-service # S3 bucket for generated CSV reports
datasources:
  datasource_url: http://prometheus:9090   # default read datasource (Thanos Query in prod)
metrics:                    # logical metric name → PromQL query variants
  example_metric:
    cross_version_query: sum(example_metric_new{FILTERS}) by(clusterId) or sum(example_metric_old{FILTERS}) by(clusterId)
    queries:
      2.12.0: sum(example_metric_new{FILTERS}) by(clusterId)   # per-cluster-version query
      2.11.0: sum(example_metric_old{FILTERS}) by(clusterId)
cleaner:                    # background DB cleanup of old reports
  is_cleanup_enabled: true
  cleaner_interval_hours: 12
  delete_report_ttl_days: 90
cron_jobs:                  # report retry/stuck-detection scheduling
  reports:
    interval_seconds: 2
    max_retry_count: 1
    retry_failed_reports_min_age_minutes: 0
    retry_failed_reports_max_age_minutes: 30
    fail_stuck_processing_reports_min_age_minutes: 30
```

Top-level sections:

- **`log_level` / `use_mocks`** — service-wide settings (`ServiceConfiguration`).
- **`metrics`** — a map keyed by logical metric name; each entry declares a `cross_version_query`
  and a `queries` map keyed by Run:AI version. `FILTERS`, `RANGE`, and `QUANTILE` placeholders
  are substituted by the injector chain at query time.
- **`datasources`** — the default datasource URL used to build read connections. In production
  this points at Thanos Query (`http://runai-backend-thanos-query:9090`); a datasource may be
  of type `thanos`, `mimir`, or `grafanalabs`.
- **`clusters` / `tenants`** — upstream service URLs (default the in-cluster
  `runai-backend-cluster-service:8080` and `runai-backend-tenants-manager:8080`) plus cache TTLs.
- **`reports`** — reports pipeline (fetch/concurrency limits, run period, S3 bucket).
- **`cleaner`** — optional DB cleanup of expired report rows.
- **`cron_jobs`** — report retry, max-retry, and stuck-processing detection timings.

## OpenAPI and code generation

The HTTP API is defined in [`swagger/source/openapi.yaml`](swagger/source/openapi.yaml), split
into `parameters`, `requests`, `schemas`, and `securitySchemes` under `swagger/source`. From
this source the build (`make build` → `openapi-build`) generates:

- The **server stubs** in [`internal/generated`](internal/generated) — consumed by `cmd/main.go`
  and the router (metrics, reports, metric-stores, and health controllers).
- The **Go client** in [`pkg/client/generated`](pkg/client/generated), wrapped by hand-written
  helpers under [`pkg/client`](pkg/client).

Generation uses [OpenAPI Generator](https://openapi-generator.tech) `v7.12.0` with
`VERSIONED_CLIENT=true`. Both `generated/` directories **must not be edited by hand** — changes
are overwritten on the next generation run.

## Go client

Other Run:AI services consume Metrics Service through the versioned Go client at
[`pkg/client`](pkg/client). See [`pkg/client/README.md`](pkg/client/README.md) for how to import
the versioned package, the metrics / metric-stores / report-creator clients, and the
`make client-bump-version` / `make client-bump-version-dev` workflow.

## Further reading

- [`pkg/client/README.md`](pkg/client/README.md) — client usage and versioning policy
