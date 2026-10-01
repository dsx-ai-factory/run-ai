# Authorization Service

The Authorization service is the control-plane authority for Run:AI's
role-based access control (RBAC). It owns access rules, roles, permission sets,
and the authorization-decision layer that other services call to decide whether
a subject may perform an action on a resource. Owned by the **Suricata** team,
it exposes a stable HTTP/REST API generated from an OpenAPI specification and a
versioned Go client that the rest of the platform embeds to enforce access.

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

- **Access rules** — creating, listing, batch-creating, syncing, and hard/soft
  deleting the bindings that grant a subject (user, group, application, SSO
  group) a role over a scope (tenant, cluster, department, project). Includes
  agent-sync and cluster-scoped/subject-scoped/me-scoped read variants.
- **Roles** — managing tenant roles and the built-in static roles (including
  legacy static roles), and enabling/disabling roles.
- **Permission sets & permissions** — resolving the permissions granted by a
  role, and managing permission sets.
- **Authorization decisions** — evaluating whether a request is authorized for a
  given resource-type / action / resource-id (the `Evaluation` and
  `PrivilegeEscalationEvaluate` surfaces), plus computing a subject's permitted
  scopes and "me" access rules.
- **Settings** — tenant-scoped authorization settings (e.g. scoped-authorization
  enablement, CSP-managed tenant), migrated on startup.
- **Scope synchronization** — background goroutines that keep access rules in
  line with the organizational hierarchy (tenants, org units, clusters) by
  calling tenants-manager, org-unit-service, and cluster-service, and a cache of
  clusters pending deletion.

The API surface is described in
[`swagger/source/openapi.yaml`](swagger/source/openapi.yaml) (title
_"Authorization"_) and bundled into
[`swagger/_build/swagger.yaml`](swagger/_build/swagger.yaml).

## Repository layout

```text
.
├── cmd/                        # Service entry point (main.go)
├── pkg/
│   ├── cache/                  # In-process caches (org hierarchy, clusters pending deletion)
│   ├── client/                 # Versioned Go client (see pkg/client/README.md)
│   │   └── generated/          # OpenAPI-generated client (do not edit)
│   ├── clients/                # HTTP clients for upstream services
│   ├── config/                 # Configuration structs and registration
│   ├── database/
│   │   ├── dao/                # DAO layer + static roles / permission-set seeds
│   │   ├── migrations/         # SQL migrations (up/down)
│   │   └── models/             # Persistence models
│   ├── pubsub/                 # Pub/sub integration
│   ├── router/                 # HTTP router wiring
│   ├── scopes_sync/            # Background scope-synchronization goroutines
│   ├── services/               # API service implementations (access rules, roles, permissions, me, settings, …)
│   └── settings/               # Typed authorization settings
├── generated/                  # OpenAPI-generated server stubs (do not edit)
├── swagger/                    # OpenAPI source and built bundle
├── Dockerfile
├── Makefile
├── NOTICE
└── go.mod
```

## Building and running

The project is built with Go (see [`go.mod`](go.mod) for the required version)
and orchestrated through the [`Makefile`](Makefile), which includes the shared
`../../build/makefile/control-plane.mk` and `index.mk` (which in turn pull in
`golang.mk` and `openapi.mk`).

| Target | Description |
|--------|-------------|
| `make build` | `openapi-build` then `build-go` — build the OpenAPI bundle + generate code, then compile the Go binary for amd64 + arm64 |
| `make build-go` | `build-go-amd` + `build-go-arm` — both arch binaries |
| `make build-go-amd` | Compile `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) in the Go builder container → `bin/authorization-amd64` |
| `make build-go-arm` | Same for `arm64` → `bin/authorization-arm64` |
| `make lint` | `lint-go` — run golangci-lint in the builder container |
| `make purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `make validate` | `openapi-validate` — validate the OpenAPI spec (yaml + generated code up-to-date + public routes + x-service field) |
| `make openapi-build` | Bundle the spec, emit JSON/HTML, copy the central spec, build the public spec, and generate server/client code |
| `make openapi-lint` | Lint the bundled swagger |
| `make docker-build-generic` | Build the service container image |
| `make docker-push-generic` | Push the container image to the registry |
| `make local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` (OS=linux, ARCH=amd64) |

The Dockerfile is a two-stage build: a
`registry.access.redhat.com/ubi9/ubi-minimal` **base** stage supplies
`/etc/passwd`, `/etc/group`, TLS certs (`/etc/ssl`, `/etc/pki`) and a writable
`/tmp`; the final image is `FROM scratch` and copies the compiled
`bin/authorization-$TARGETARCH` binary, the SQL migrations
(`pkg/database/migrations/up|down`), the swagger bundle, and `NOTICE`. It runs
as user `65532:65532` with entrypoint `./authorization`. The image takes one
build arg, `TARGETARCH`, selecting which arch binary to copy:

```bash
docker build --build-arg TARGETARCH=amd64 -t authorization:dev .
```

### Local build (local-dev.sh)

`authorization` is a control-plane Go service, so the repo-root `local-dev.sh`
builds it via its `build_go_service` path in two steps:

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
     --build-arg TARGETARCH=amd64 -t runai-local/authorization:0.0.0 .
   ```

The image ref is `runai-local/authorization:0.0.0` — tag **`0.0.0`** matches the
control-plane chart appVersion; any other tag breaks the local deploy. For a
control-plane-only iteration:

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```

## Testing

| Target | Description |
|--------|-------------|
| `make test` | `test-ginkgo` — run the Ginkgo unit/spec suites (the `pkg/services/*_test.go`, DAO tests, etc.) in the builder container |

The suites run against the code directly (Ginkgo, randomized), so no
docker-compose backing stack is required for the default `make test`.

## Configuration

Configuration is **config-object driven**: the service registers typed config
structs (`Server`, `SyncJob`) that are populated from built-in
defaults, an optional config file, and environment-variable overrides. Any
config key can be overridden by its uppercase environment-variable equivalent
(nested keys joined by the configured delimiter). A representative configuration,
trimmed to the registered structs and their defaults, is:

```yaml
server:
  service_port: 8080                       # HTTP listen port
  workloads_url: http://runai-backend-workloads:8080
  org_unit_service_url: http://runai-backend-org-unit-service:8080
  cluster_service_url: http://runai-backend-cluster-service:8080
  tenant_service_url: http://runai-backend-tenants-manager:8080
  datavolumes_service_url: http://runai-backend-datavolumes:8080
  temp_backend_settings_url: http://runai-backend-temp-backend-settings:8080
  is_csp: false

sync_job:
  sync_job_interval_hours: 168             # full scope-sync interval
  sync_job_interval_tenant_seconds: 5      # per-tenant sync interval
```

- **`server`** — the HTTP port and the URLs of every upstream control-plane
  service the authorizer calls (workloads, org-unit-service, cluster-service,
  tenants-manager, datavolumes, temp-backend-settings), plus the `is_csp` flag.
- **`sync_job`** — the intervals that drive the background scope-sync goroutines.

**Bootstrap and infrastructure env vars.** The config loader and the shared
database/secret layers read these directly:

| Variable | Description | Default |
|----------|-------------|---------|
| `CONFIG_FILE_PATH` | Path to an additional config file to load | — |
| `CONFIG_SECRETS_FILE_PATH` | Path to a secrets config file (watched) | — |
| `SECRETS_MOUNT_PATH` | Folder of mounted secret files (e.g. K8s secrets) | `/etc/secrets` |
| `CONFIG_KEY_DELIMITER` | Delimiter used to map nested keys to env-var names | `.` |
| `DB_HOST` / `DB_PORT` | PostgreSQL host / port | — |
| `DB_READ_HOST` / `DB_READ_PORT` | Read-replica host / port | — |
| `DB_NAME` | Database name | — |
| `DB_SCHEMA` | DB schema (service uses `authorization`) | `authorization` |
| `DB_USER` / `DB_PASSWORD` | DB credentials (secret) | — |
| `DB_SSL_MODE` | PostgreSQL SSL mode | — |

The service also embeds the standard logging/metrics/tracing/audit/pub-sub
packages, which read their own environment configuration.

## OpenAPI and code generation

The HTTP API is defined in
[`swagger/source/openapi.yaml`](swagger/source/openapi.yaml) and split into
per-resource files under `swagger/source/resources` and shared schemas under
`swagger/source/schemas` (plus `parameters` and `securitySchemes`). From this
source the build generates:

- The **server stubs** in [`generated/`](generated) (`OPENAPI_GENERATOR_SERVER_CODE_PATH=generated`),
  consumed by `cmd/main.go` and the router.
- The **Go client** in [`pkg/client/generated`](pkg/client/generated), wrapped by
  hand-written helpers under [`pkg/client`](pkg/client).

Both are produced by [OpenAPI Generator](https://openapi-generator.tech)
(version `v7.12.0`, pinned in the [`Makefile`](Makefile)) and **must not be
edited by hand** — changes are overwritten on the next generation run.
`make build` runs `openapi-build` (which includes code generation) first;
`make validate` checks that the checked-in bundle and generated code are
up-to-date.

## Go client

Other Run:AI services consume Authorization through the versioned Go client at
[`pkg/client`](pkg/client), imported as `control-plane/authorization/pkg/client/v<version>`.
See [`pkg/client/README.md`](pkg/client/README.md) for:

- How to import and wire the versioned client (`authorization_client.Init(...)`)
- Marking endpoints authorized/unauthorized via the `x-authorize` /
  `x-no-authorization` OpenAPI extensions
- The `make client-bump-version` and `make client-bump-version-dev` workflow

## Further reading

- [`pkg/client/README.md`](pkg/client/README.md) — client usage and versioning policy
- [`generated/README.md`](generated/README.md) — generated server reference (auto-generated)
- [`pkg/client/generated`](pkg/client/generated) — generated client reference (auto-generated)
