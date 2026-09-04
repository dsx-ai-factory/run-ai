# policy-service

`policy-service` owns Run.ai **policies** — the rules that constrain and default workload specifications (allowed values, limits, defaults, imposed assets) across the org hierarchy. It stores policies per **scope** (tenant → department → project), computes the **effective policy** for any scope by merging the hierarchy tree, validates submitted workloads for **compliance** against the effective policy, and syncs the resulting effective policies out to clusters. It exposes a REST API (chi router, OpenAPI-defined) and persists to PostgreSQL. It is owned by the **Viper Team**.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [OpenAPI and code generation](#openapi-and-code-generation)
- [Go client](#go-client)
- [Running locally against a port-forwarded backend](#running-locally-against-a-port-forwarded-backend)

## Overview

The service is responsible for:

- **Policy storage & scoping** — persists policies keyed by scope (tenant/department/project) and workload type; `pkg/scopes/` derives the scope of an incoming request.
- **Effective-policy computation** — `pkg/tree/` builds the scope hierarchy and merges parent→child policies into a single effective policy per scope (`calculate_effective.go`, `merge.go`), including imposed assets and cluster-deletion approval.
- **Compliance validation** — `pkg/compliance/` validates a workload spec against the effective policy and reports violations.
- **Workload spec mapping** — `pkg/system_mapping/` maps Kubernetes workload kinds (Deployment, CronJob, Grove/DynamoGraphDeployment, …) into policy-addressable field groups so policies can target concrete spec fields.
- **Cluster sync** — `pkg/services/` applies policy changes and tracks `sync_status` per scope so cluster agents pick up updates; `pkg/pubsub/` publishes change events.
- **Talking to sibling services** — `pkg/clients/` calls assets-service, cluster-service, and org-unit-service for the data needed to resolve and validate policies.

## Repository layout

```text
.
├── cmd/
│   └── main.go             # Service entry point
├── pkg/
│   ├── clients/            # Clients to assets / cluster / org_unit services
│   ├── compliance/         # Workload-vs-policy validation
│   ├── config/             # Config registration (env-driven) + DB config
│   ├── database/           # DAO layer, models, tables, SQL migrations
│   │   └── migrations/{up,down}/   # golang-migrate SQL migrations
│   ├── pubsub/             # Change-event publishing
│   ├── scopes/             # Scope derivation from requests
│   ├── services/           # Apply, breakdown generation, hooks, health
│   ├── system_mapping/     # K8s workload-kind → policy field-group mapping
│   └── tree/               # Hierarchy tree + effective-policy merge
├── generated/              # OpenAPI-generated server + client code (do not edit)
│   ├── api/
│   └── pkg/
├── swagger/
│   ├── source/openapi.yaml # OpenAPI source ("Policies" API)
│   └── _build/             # Bundled swagger.yaml/json (build output)
├── Dockerfile
├── Makefile
└── go.mod
```

## Building and running

Built with Go (see [`go.mod`](go.mod) — module `github.com/dsx-ai-factory/run-ai/control-plane/policy-service`, Go 1.26) and orchestrated through the [`Makefile`](Makefile), which includes `../../build/makefile/control-plane.mk` and `index.mk`. **`make build` bundles the OpenAPI spec and regenerates server/client code first, then compiles** — the `generated/` directories are a build product.

| Target | Description |
|--------|-------------|
| `make build` | `openapi-build` (bundle spec + generate server/Go-client/JS-client code) then `build-go` (compile both amd64 and arm64 into `bin/`). |
| `make lint` | Runs the Go linter (`lint-go`). |
| `make openapi-lint` | Lints the OpenAPI bundle (`openapi-lint-generic`). |
| `make validate` | `openapi-validate` — validates the spec (yaml + generated code up-to-date + public routes published + service-field checks). |
| `make purify` | Auto-fix formatting and imports (`purify-go`). |
| `make coverage` | Coverage via `test-go-acc` (ignores `clients/assets`). |
| `make copy-to-backend-js` | Legacy: copies the built `swagger.yaml` into the backend-js tree. |
| `make local-deploy` | `openapi-generate-code build-go docker-build-generic docker-push-generic` — build and push the image. |

### Local build (local-dev.sh)

`policy-service` is a **control-plane Go service** (listed in `CP_GO_SERVICES` in [`local-dev.sh`](../../local-dev.sh)). `local-dev.sh` builds amd64 only (Kind nodes are linux/amd64):

```bash
# amd64 — what local-dev.sh runs
make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
docker build --build-arg TARGETARCH=amd64 -t runai-local/policy-service:0.0.0 .
```

For an arm64 cluster, build the arm64 path manually (`build-go-arm` **hardcodes** `GOOS=linux GOARCH=arm64` — do **not** pass `OS=`/`ARCH=`):

```bash
# arm64 — manual
make build-go-arm DOCKERHUB_PROXY= GITHUB_TOKEN=
docker build --build-arg TARGETARCH=arm64 -t runai-local/policy-service:0.0.0 .
```

- The image tag is **`0.0.0`** (the CP `appVersion`); any other tag breaks the local deploy. Image ref: `runai-local/policy-service:0.0.0`.
- `DOCKERHUB_PROXY=` (empty) pulls the golang builder from Docker Hub; `GITHUB_TOKEN=` is empty in the make call (the private-module token is applied elsewhere).
- Same tag for both arches — only `--build-arg TARGETARCH` changes.
- **Note:** the shared `build-go-amd` compiles but does not run OpenAPI generation; the service's own `make build` (`openapi-build build-go`) is what regenerates `generated/` first. Run `make build` (or `make openapi-generate-code`) if the spec changed.

### Dockerfile

Two-stage, producing a minimal `scratch` image:

- **`base`** — `registry.access.redhat.com/ubi9/ubi-minimal`; sources `/etc/passwd`, `/etc/group`, SSL certs (`/etc/ssl/certs`, `/etc/pki`), and a pre-permissioned writable `/tmp`.
- **`scratch` (final)** — copies those identity/TLS files, the pre-built `bin/policy-service-$TARGETARCH` binary (renamed `policy-service`), the `NOTICE`, and — importantly — the **DB migrations** (`pkg/database/migrations/{up,down}`) and the built swagger (`swagger/_build/*`). Runs as non-root `USER 65532:65532`, `ENTRYPOINT ["./policy-service"]`. The service applies its migrations on startup, which is why they ship in the image.

Build arg: `TARGETARCH` (amd64/arm64) — selects which pre-built binary is copied in.

## Testing

| Target | Description |
|--------|-------------|
| `make test` | Runs the [Ginkgo](https://onsi.github.io/ginkgo/) suites (`test-ginkgo`). |
| `make coverage` | Runs tests with coverage via `test-go-acc` (excludes `clients/assets`). |

Suites live next to the code (`pkg/**/*_test.go`, `suite_test.go`). DB-backed suites use a throwaway Postgres (`pkg/database/dbtesting`); `config.MockConfig()` wires the test DB/env.

## Configuration

Configuration is **environment-variable driven**. Application settings are registered from the `ServerConfig` / `ServicesConfig` structs in [`pkg/config/config.go`](pkg/config/config.go) (each field's env-var name is the uppercase form of its key); database settings come from the shared `dao-utils` `DB_*` variables. Defaults come from each struct's `SetDefaults()`.

**Application:**

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_PORT` | HTTP port to listen on | `8080` |
| `LOG_LEVEL` | Log level | `info` |
| `POLICIES_PER_BULK` | Policies processed per bulk operation | `50` |
| `SYNC_API_WORKERS` | Concurrent workers for the sync API | `20` |
| `RETRY_MAX_INTERVAL_SEC` | Max backoff interval for retries | `10` |
| `RETRY_MAX_ELAPSED_SEC` | Max total retry duration | `60` |
| `DAYS_LIMIT_FOR_FIRST_SYNC` | Look-back window (days) for the first cluster sync | `7` |
| `ASSETS_SERVICE_URL` | assets-service base URL | `http://runai-backend-assets-service:8080` |
| `CLUSTER_SERVICE_URL` | cluster-service base URL | `http://runai-backend-cluster-service:8080` |
| `ORG_UNIT_SERVICE_URL` | org-unit-service base URL | `http://runai-backend-org-unit-service:8080` |
| `PUBSUB_ENABLED` | Enable publishing change events | — |
| `SECRETS_MOUNT_PATH` | Directory holding mounted secret files | — |

**Database** (via `dao-utils`):

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | — |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_NAME` | Database name | — |
| `DB_SCHEMA` | Schema (e.g. `policy_service`) | — |
| `DB_USER` | Database user | — |
| `DB_PASSWORD` | Database password | — |
| `DB_SSL_MODE` | SSL mode (`disable`/`require`/…) | — |
| `MIGRATION_FILES` | Migration source URL (e.g. `file://.../migrations`) | — |
| `SERVICE_NAME` | Service name used by the DB layer | `policy-service` |

(`dao-utils` also honors optional `DB_READ_HOST`/`DB_READ_PORT`, `STATEMENT_TIMEOUT`, `MAX_CONNECTIONS`, and related tuning variables.)

## OpenAPI and code generation

The API is defined in [`swagger/source/openapi.yaml`](swagger/source/openapi.yaml) (the "Policies" API). `make build` runs `openapi-build`, which bundles the spec into `swagger/_build/` and then `openapi-generate-code` produces the **server stubs and clients** into `generated/` (server-go, client-go, client-js), using `openapi-generator` `v7.12.0`. **The `generated/` directories must not be edited by hand** — change the spec and regenerate. `make validate` (`openapi-validate`) fails the build if the generated code or public-route listing is out of date.

## Go client

There is no in-tree `pkg/client`; the generated Go client is published as its own module at [`runai-clients/policy-service-client`](../../runai-clients/policy-service-client) (produced from this service's spec). Consumers depend on that module rather than importing `generated/` directly.

## Running locally against a port-forwarded backend

To run the server against services in an existing cluster:

```bash
go run cmd/main.go
```

Set the DB and sibling-service URLs (IDE/`env` form; corrected to current config keys — the earlier `AUTH_SERVICE_URL`/`AUTHORIZATION_URL`/`BACKEND_SERVICE_URL`/`DB_SSL_ENABLED` variables are obsolete):

```
ASSETS_SERVICE_URL=http://localhost:8001;CLUSTER_SERVICE_URL=http://localhost:5500;ORG_UNIT_SERVICE_URL=http://localhost:6500;DB_HOST=localhost;DB_NAME=backend;DB_SCHEMA=policy_service;DB_USER=user;DB_PASSWORD=password;DB_SSL_MODE=disable;PUBSUB_ENABLED=false;MIGRATION_FILES=file://<repo>/control-plane/policy-service/pkg/database/migrations
```

Port-forward the backing services:

```bash
kubectl -n runai-backend port-forward service/runai-backend-postgresql 5432:5432 &
kubectl -n runai-backend port-forward service/runai-backend-policy-service 8000:8080 &
kubectl -n runai-backend port-forward service/runai-backend-org-unit-service 6500:8080 &
kubectl -n runai-backend port-forward service/runai-backend-cluster-service 5500:8080 &
kubectl -n runai-backend port-forward service/runai-backend-assets-service 8001:8080 &
```

Tear the forwards down:

```bash
kill -9 `lsof -i tcp:8000 -i tcp:6500 -i tcp:8001 -i tcp:5500 -i tcp:5432 | awk '{ print $2 }' | grep -v PID`
```

**Debug query** — inspect cluster↔policy sync state:

```sql
SELECT scope, p.workload_type, update_count,
       p.deleted_at  AS pol_del,
       ep.deleted_at AS ef_del,
       ss.sent_to_agent_apply  AS sent_upd,
       ss.sent_to_agent_delete AS sent_del,
       ss.processed_by_agent   AS ag_proc
FROM policies.policies AS p
LEFT JOIN policies.effective_policies AS ep ON ep.policy_id = p.id
LEFT JOIN policies.sync_status        AS ss ON ss.scope_type = p.scope
ORDER BY update_count;
```
