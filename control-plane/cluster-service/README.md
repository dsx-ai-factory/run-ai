# Cluster Service

Cluster Service is the control-plane service (owned by the **Griffin** team) that
manages the Run:ai *cluster* domain: the clusters registered against a tenant,
their control-plane configuration and installation/version metadata, and the
cluster-scoped resources that hang off them — nodes, node pools, network
topologies, and cluster metrics. It stores desired state in its own PostgreSQL
schema (`cluster_service`), reconciles/syncs live state pulled from each cluster,
and exposes a stable HTTP/REST API generated from an OpenAPI specification.

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

- **Clusters** — registration, listing (full and "minimal"), lifecycle
  (install, connect, graceful/forced delete), version and installation
  metadata, and per-cluster history tracking.
- **Cluster control-plane config** — serving the `ClusterControlPlaneConfig`
  each cluster fetches on boot (Helm registry coordinates, install/version
  bounds, image-pull secret, FIPS mode, domain wiring), gated by cluster
  access control (authn + authz).
- **Nodes** — node status is synced from each cluster, persisted, and served
  through the API (counts, per-node lookups, telemetry).
- **Node pools** — node-pool listing (full and minimal), settings, ancestors,
  notifications, and a background sync loop that reconciles node pools per
  tenant on an interval.
- **Network topologies** — CRUD and sync of cluster network-topology objects.
- **Metrics & settings** — cluster/node/nodepool metric endpoints and
  tenant-scoped settings (advanced GPU metrics, CPU resource quota, project
  manual over-quota).

Background workers include the **node-pool sync**, the **deleted-cluster
cleaner** (reaps clusters that exceeded the graceful-delete window), and a
**pub/sub** integration used to drive async cluster deletion and status
post-update handling.

The API surface is described in [`swagger/source/openapi.yaml`](swagger/source/openapi.yaml)
and bundled into [`swagger/_build/`](swagger/_build).

## Repository layout

```text
.
├── cmd/                        # Service entry point (main.go)
├── pkg/
│   ├── adapters/               # Adapters (e.g. network-topology sync)
│   ├── api/                    # API-level helper types (group-by, verbosity)
│   ├── business/               # Domain logic (cluster install, cluster-cp-config, cluster-infra)
│   ├── cleaners/               # Deleted-cluster cleaner (background reaper)
│   ├── client/                 # Versioned Go client (see pkg/client/README.md)
│   │   └── generated/          # OpenAPI-generated client (do not edit)
│   ├── clients/                # HTTP clients for upstream services (backend, workloads)
│   ├── common/                 # Shared errors
│   ├── config/                 # Configuration schema (registered config structs)
│   ├── database/               # DAO layer, models, and SQL migrations (up/down)
│   ├── pubsub/                 # Pub/sub integration
│   ├── router/                 # HTTP router + middleware wiring
│   ├── runaimetrics/           # Run:ai metric definitions (per cluster/node/nodepool/tenant)
│   ├── services/               # API service implementations
│   ├── settings/               # Typed setting definitions + managers
│   ├── sync/                   # Node-pool sync loop
│   ├── test/                   # DB/service integration-style Ginkgo tests
│   └── utils/
├── internal/
│   ├── generated/              # OpenAPI-generated server stubs (do not edit)
│   └── testutils/
├── swagger/                    # OpenAPI source and built bundle
├── mocks/                      # Generated mocks
├── Dockerfile
├── Makefile
└── go.mod
```

## Building and running

The project is built with Go (see [`go.mod`](go.mod) for the required version)
and orchestrated through the [`Makefile`](Makefile), which sets `SERVICE_NAME`
and includes the shared `../../build/makefile/control-plane.mk` and
`../../build/makefile/index.mk` (which pull in `golang.mk`, `openapi.mk`,
`versioning.mk`, etc.).

| Target | Description |
|--------|-------------|
| `make build` | `openapi-build` then `build-go` — bundle + generate OpenAPI, then compile the Go binary for amd64 + arm64 |
| `make build-go` | `build-go-amd` + `build-go-arm` |
| `make build-go-amd` | Compile `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) in the Go builder container → `bin/cluster-service-amd64` |
| `make build-go-arm` | Same for `arm64` → `bin/cluster-service-arm64` |
| `make lint` | `lint-go` — golangci-lint (shared config `../infra/lint/.golangci.yaml`) |
| `make purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `make validate` | `openapi-validate` — validate the OpenAPI spec + generated code |
| `make openapi-lint` | Lint the swagger bundle |
| `make local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` (OS=linux ARCH=amd64) |

### Docker image

[`Dockerfile`](Dockerfile) is a two-stage build. A
`registry.access.redhat.com/ubi9/ubi-minimal` **base** stage supplies
`/etc/passwd`, `/etc/group`, SSL certs (`/etc/ssl`, `/etc/pki`) and a writable
`/tmp`. The final image is `FROM scratch`, copies in the pre-built
`bin/cluster-service-$TARGETARCH` binary, the SQL migrations
(`pkg/database/migrations/up|down`), and the swagger bundle, and runs as user
`65532:65532`. It takes one build arg, `TARGETARCH`:

```bash
docker build --build-arg TARGETARCH=amd64 -t cluster-service:dev .
```

### Local build (local-dev.sh)

`cluster-service` is a control-plane Go service (listed in `CP_GO_SERVICES`), so
the repo-root [`local-dev.sh`](../../local-dev.sh) builds it in two steps:

1. Compile the amd64 binary in the Go builder container:
   ```bash
   make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
   ```
   (`DOCKERHUB_PROXY=` pulls the Go builder image straight from Docker Hub;
   `GITHUB_TOKEN=` is emptied in the make call — the private-module token is
   applied elsewhere.)
2. Build and load the image:
   ```bash
   docker buildx build --platform linux/amd64 --provenance=false --load \
     --build-arg TARGETARCH=amd64 -t runai-local/cluster-service:0.0.0 .
   ```

The tag **`0.0.0`** matches the control-plane chart `appVersion`; any other tag
breaks the local deploy. Image ref: `runai-local/cluster-service:0.0.0`.

To iterate on just the control-plane (rebuild + redeploy, reuse the cluster):

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```

## Testing

| Target | Description |
|--------|-------------|
| `make test` | `test-ginkgo` — run the Ginkgo unit/DB test suites (`GINKGO_PROCS=4`, `GINKGO_COMPILERS=4`) |
| `make coverage` | Open the HTML coverage report (`go tool cover`) |

Tests are Ginkgo suites; the DB-backed suites under `pkg/test` and
`pkg/database` spin up test containers, so a working container runtime is
required to run them locally.

## Configuration

Cluster Service is **config-driven**: at startup it loads a set of registered
configuration structs (see [`pkg/config`](pkg/config)) from a mounted config /
secrets source. Each field maps to a config key (the `mapstructure` name), and
**any key can be overridden by its uppercase environment-variable equivalent**
(e.g. the `service_port` key is settable as `SERVICE_PORT`, `app_host` as
`APP_HOST`). Fields tagged as secrets (image-pull secret, metric-store
credentials) are read from the mounted secrets folder.

The registered configuration sections and their notable keys/defaults:

| Section | Key(s) | Default |
|---------|--------|---------|
| Service | `max_wait_for_cluster_remote_config_status_update` | `5m` |
| Service | `max_wait_for_cluster_graceful_delete` | `30m` |
| Service | `cluster_cleaner_run_timing` | `1h` |
| Server | `service_port` | `8080` |
| Auth | `identity_mng_client_url` | `runai-backend-identity-manager:8080` |
| Auth | `keycloak_external_endpoint` | — |
| Backend | `temp_backend_settings_url` | `http://runai-backend-temp-backend-settings:8080` |
| App | `app_host`, `cluster_host_template`, `runai_cluster_domain_suffix` | — (must be provided) |
| App | `cluster_connectivity_threshold` | `1m` |
| App | `org_unit_svc_url` | `http://runai-backend-org-unit-service:8080` |
| App | `tenants_manager_svc_url` | `http://runai-backend-tenants-manager:8080` |
| App | `workloads_svc_url` | `http://runai-backend-workloads:8080` |
| Installation | `version`, `cluster_installation_min_version`, `cluster_installation_max_version`, `cluster_installation_excluded_versions` | `version=development` |
| Installation | `cluster_installation_helm_registry_url` / `_username` / `_password_placeholder` | — |
| Installation | `installation_type` (`selfhosted`/`launchpad`/`airgap`/`saas`) | `selfhosted` |
| Installation | `cluster_installation_fips_mode` (`off`/`on`/`only`) | `off` |
| Environment | `environment_identifier`, `environment_release_name`, `fips_mode` | `environment=production`, `fips_mode=off` |
| CloudFormation | `aws_elb_suffix` | `lb.example.com` |
| Sync | `nodepool_sync_interval_minutes` | `180` |
| Sync | `nodepool_sync_tenants_chunk_size` | `2` |
| Sync | `nodepool_sync_tenant_interval_seconds` | `5` |
| AsyncClusterDelete | `async_cluster_delete_enabled` | `true` |
| ApplicationMetricStore | `application_metric_store_write_url` / `_username` / `_password` / `_headers` / `application_metric_store_label` | `label=applicative_runai` |
| ImagePullSecret | `image_pull_secret_enabled`, `static_image_pull_secret` (secret) | — |

The service also registers the shared **database** config (from the `dao-utils`
package, schema `cluster_service`) and the **authorization client** config.
Database connection settings therefore follow the shared `dao-utils` env
conventions rather than being declared in this repo.

## OpenAPI and code generation

The HTTP API is defined in [`swagger/source/openapi.yaml`](swagger/source/openapi.yaml),
split into per-resource files under `swagger/source/resources`, with shared
`schemas`, `parameters`, and `securitySchemes`. From this source the build
generates:

- The **server stubs** in [`internal/generated`](internal/generated) (consumed
  by `cmd/main.go` and the router).
- The **Go client** in [`pkg/client/generated`](pkg/client/generated), wrapped
  by hand-written helpers under [`pkg/client`](pkg/client).

Generation uses [OpenAPI Generator](https://openapi-generator.tech) pinned to
`v7.12.0` (server and Go client), with `VERSIONED_CLIENT=true`. `make build`
runs `openapi-build` (bundle + generate) before compiling. Both `generated/`
directories **must not be edited by hand** — changes are overwritten on the next
generation run. `make validate` re-generates into a temp dir and diffs it
against the committed code to ensure the spec and generated code stay in sync.

## Go client

Other Run:ai services consume Cluster Service through the versioned Go client at
[`pkg/client`](pkg/client), imported as a versioned package
(`control-plane/cluster-service/pkg/client/v1.0.0`). See
[`pkg/client/README.md`](pkg/client/README.md) for the full usage and versioning
policy, including the `make client-bump-version` (production) and
`make client-bump-version-dev` (feature-branch) workflow. Dev-versioned
references may not be merged to `master` — CI enforces this.

## Further reading

- [`pkg/client/README.md`](pkg/client/README.md) — client usage and versioning policy
- [`internal/generated/README.md`](internal/generated/README.md) — generated server reference (auto-generated)
- [`pkg/client/generated`](pkg/client/generated) — generated client (auto-generated)
