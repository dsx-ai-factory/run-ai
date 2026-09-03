# agent

`agent` is the runai-cluster component that applies Run.ai control-plane ("runai-backend") state **into** the Kubernetes cluster. It authenticates to the backend, periodically pulls the desired state of org and workload resources, and reconciles each kind into the cluster through a set of pluggable **resource handlers** — departments, projects, access rules, data volumes, priority classes, workload-type configs, workloads (interactive / training / distributed / inference / distributed-inference / v2), credential and datasource assets, network topologies, applications, and KAI node pools. It also hosts a TLS **validating webhook** server for project/department and secret-PVC admission. It is owned by the **Suricata Team**. This is the inbound half of the cluster↔backend data path; the outbound half (cluster → backend) is handled by `cluster-sync`.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [OpenAPI and code generation](#openapi-and-code-generation)
- [Extending: adding a resource handler](#extending-adding-a-resource-handler)

## Overview

The service is responsible for:

- **Syncing backend state into the cluster** — on a configurable interval (`AGENT_SYNC_INTERVAL_MILLIS`), each registered handler (`pkg/handlers/*_handler.go`) fetches its resource from the backend and reconciles the corresponding cluster objects. Handlers are registered in `handlerInitFuncs` in `pkg/runai_agent.go`.
- **Authenticating to the backend** — exchanges `CLIENT_NAME` / `CLIENT_SECRET` for a token via the identity manager and calls the Run.ai API at `ENDPOINT` on behalf of cluster `CLUSTER_ID`.
- **Serving validating webhooks** — a controller-runtime TLS webhook server (cert from `CERT_DIR`, port `WEBHOOK_PORT`) registers `/validate-project` and `/validate-secrets-pvc` admission handlers; `ALLOWED_SERVICE_ACCOUNTS` gates which service accounts may perform project/department operations.
- **Optional policy sync** — when `POLICY_SYNC_ENABLED` is set, interactive/training policy handlers are added to the handler set.
- **KAI integration** — optionally syncs node pools and org units to the KAI scheduler CRDs (`ENABLE_KAI_NODEPOOL_CRD`).
- **Exposing metrics** — Prometheus metrics on `METRICS_PORT`/`METRICS_PATH`; supports leader election (`LEADER_ELECT`).

## Repository layout

```text
.
├── cmd/
│   └── main.go             # Entry point: config, manager, webhooks, agent loop
├── pkg/
│   ├── runai_agent.go      # Agent core: handler registration + sync loop
│   ├── handlers/           # One handler per resource kind (+ resource_handler_template)
│   ├── adapters/           # Resource adapters (e.g. network topology)
│   ├── backend-client/     # Authenticated client to the Run.ai backend
│   ├── config/             # AgentConfig struct, env mapping and defaults
│   ├── config-maps/        # runai config map handling
│   ├── common/             # Shared helpers
│   ├── metrics/            # Prometheus metrics exporter
│   ├── validator/          # Webhook validation logic
│   ├── generated/          # Generated OpenAPI clients (do not edit)
│   │   ├── projects_openapi/
│   │   ├── backend_openapi/
│   │   ├── auth_openapi/
│   │   └── datavolumes_openapi/
│   └── tests/              # Ginkgo suite wiring
├── swagger/                # Backend OpenAPI source specs (input to codegen)
├── scripts/                # deploy.sh, mirrord config
├── Dockerfile
├── Makefile
└── go.mod
```

## Building and running

Built with Go (see [`go.mod`](go.mod) — module `github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent`, Go 1.26) and orchestrated through the [`Makefile`](Makefile), which includes `../../../build/makefile/runai-cluster.mk` and `index.mk`. Build/test targets run inside a golang Docker container.

| Target | Description |
|--------|-------------|
| `make build` | Builds the service (`build-go` → both amd64 and arm64 binaries into `bin/`). |
| `make build-go-amd` | Builds the amd64 binary (`bin/agent-amd64`); honors `OS=`/`ARCH=`. |
| `make build-go-arm` | Builds the arm64 binary (`bin/agent-arm64`); hardcodes `GOOS=linux GOARCH=arm64`. |
| `make format` | Formats the code (`format-go`). |
| `make lint` | Runs the Go linter (`lint-go`). |
| `make validate` | No-op (kept for parity across runai-cluster services). |
| `make local-deploy` | Builds (amd64) and pushes the image (`build-go docker-build-generic docker-push-generic`). |
| `make generate` | Regenerates all OpenAPI clients (see [OpenAPI and code generation](#openapi-and-code-generation)). |
| `make generate-projects` / `generate-backend` / `generate-authorization` / `generate-datavolumes` | Regenerate a single generated client from its `swagger/` spec. |

### Local build (local-dev.sh)

`agent` is a **runai-cluster Go service** (listed in `RC_GO_SERVICES` in [`local-dev.sh`](../../../local-dev.sh)). `local-dev.sh` builds amd64 only (Kind nodes are linux/amd64):

```bash
# amd64 — what local-dev.sh runs (build_go_service)
make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
docker build --build-arg TARGETARCH=amd64 -t runai-local/agent:0.0.0-dev .
```

For an arm64 cluster, build the arm64 path manually (`build-go-arm` **hardcodes** `GOOS=linux GOARCH=arm64` — do **not** pass `OS=`/`ARCH=`):

```bash
# arm64 — manual
make build-go-arm DOCKERHUB_PROXY= GITHUB_TOKEN=
docker build --build-arg TARGETARCH=arm64 -t runai-local/agent:0.0.0-dev .
```

- The image tag is **`0.0.0-dev`** (the `RC_IMAGE_TAG`, hardcoded by the runai-cluster chart values); any other tag breaks the local deploy. Image ref: `runai-local/agent:0.0.0-dev`.
- `DOCKERHUB_PROXY=` (empty) pulls the golang builder image directly from Docker Hub; `GITHUB_TOKEN=` is empty in the make call (the private-module token is applied elsewhere).
- The tag is the same for both arches — only `--build-arg TARGETARCH` changes.

### Dockerfile

Multi-stage, producing a minimal `scratch` image:

- **`base`** — `registry.access.redhat.com/ubi9/ubi-minimal`; used only to source `/etc/passwd`, `/etc/group`, SSL certs (`/etc/ssl/certs`, `/etc/pki`), `nsswitch.conf`, and a pre-permissioned writable `/tmp` (owned `65532:0`, mode `770`).
- **`scratch` (final)** — copies those identity/TLS/DNS files and the pre-built `bin/agent-$TARGETARCH` binary (renamed `agent`), plus the `NOTICE`. Runs as non-root `USER 65532:65532`, `ENTRYPOINT ["./agent"]`.

Build arg: `TARGETARCH` (amd64/arm64) — selects which pre-built binary is copied in.

## Testing

| Target | Description |
|--------|-------------|
| `make test` | Runs unit tests via `test-go` (containerized). |
| `make test-dev` | Runs `go test -v ./... -coverprofile cover.out` directly on the host. |

Tests are [Ginkgo](https://onsi.github.io/ginkgo/) suites colocated with the handlers (`pkg/handlers/*_test.go`, `pkg/tests/`). No docker-compose/backing services required.

## Configuration

Configuration is **environment-variable driven** (the uppercase `mapstructure` tags in [`pkg/config/config.go`](pkg/config/config.go) are the env-var names). Defaults come from `SetDefaults()`. Values marked `config:"secret"` (`CLUSTER_ID`, `CLIENT_NAME`, `CLIENT_SECRET`) are instead read from a mounted secret file — a JSON file at `/etc/secret/secret.json` (`constants.SecretMountPath`). There is no YAML app-config file for this service.

| Variable | Description | Default |
|----------|-------------|---------|
| `CLUSTER_ID` | Cluster UUID (secret) | — |
| `CLIENT_NAME` | Application client name for backend auth (secret) | — |
| `CLIENT_SECRET` | Application client secret (secret) | — |
| `CLUSTER_VERSION` | Cluster version (used by the workload composer; also read directly in `main.go`) | — |
| `ENDPOINT` | Base URL of the Run.ai backend | — |
| `POLICY_SYNC_ENABLED` | Sync interactive/training policies from the backend | `false` |
| `TLS_REJECT_UNAUTHORIZED` | Reject unauthorized TLS certs when calling the backend | `false` |
| `AGENT_SYNC_INTERVAL_MILLIS` | Interval between sync passes (ms) | `10000` |
| `USE_DEBUG_LOG_LEVEL` | Enable debug logging | `false` |
| `METRICS_PORT` | Prometheus metrics port | `9400` |
| `METRICS_PATH` | Prometheus metrics path | `/metrics` |
| `PROJECTS_LIMIT` | Max projects fetched per sync | `200` |
| `DEPARTMENTS_LIMIT` | Max departments fetched per sync | `400` |
| `ACCESS_RULES_LIMIT` | Max access rules fetched per sync | `400` |
| `K8S_CLIENT_CONFIG_QPS` | Kubernetes client QPS | `50` |
| `K8S_CLIENT_CONFIG_BURST` | Kubernetes client burst | `300` |
| `LEADER_ELECT` | Enable leader election | `false` |
| `ENABLE_KAI_NODEPOOL_CRD` | Sync node pools to the KAI `NodePool` CRD | `false` |
| `CERT_DIR` | Directory holding the webhook server TLS cert (from init-ca) | `/etc/webhook/certs` |
| `WEBHOOK_PORT` | Validating-webhook server port | `9443` |
| `ALLOWED_SERVICE_ACCOUNTS` | Comma-separated service-account patterns allowed to perform project/department operations | `""` |

The Dockerfile sets `TARGETARCH` (build arg → env) to select the binary; it defines no runtime `ENV` configuration beyond that.

## OpenAPI and code generation

The agent consumes the backend over generated OpenAPI clients under `pkg/generated/` — **these directories are generated and must not be edited by hand.** The backend contract specs live in [`swagger/`](swagger/), and each generated client is produced from one spec with `openapi-generator-cli` (`v6.0.1`, `-g go`, `--additional-properties onlyInterfaces=true outputAsLibrary=true enumClassPrefix=true`):

| `make` target | Spec (`swagger/`) | Generated package (`pkg/generated/`) |
|---------------|-------------------|--------------------------------------|
| `generate-projects` | `projectsDepartmentsApi.yaml` | `projects_openapi` |
| `generate-backend` | `clusterApi.yaml` | `backend_openapi` |
| `generate-authorization` | `swagger.yaml` | `auth_openapi` |
| `generate-datavolumes` | `datavolumesApi.yaml` | `datavolumes_openapi` |
| `generate` | all of the above | all of the above |

To refresh a client, copy the latest backend spec into `swagger/` and run the matching target (or `make generate` for all). The generated `go.mod`/`go.sum` files are removed so the packages compile as part of this module.

## Extending: adding a resource handler

To sync a new resource kind:

1. Start from `pkg/handlers/resource_handler_template`.
2. Add `<resource>_handler.go` in `pkg/handlers`, implementing the `handlers.ObjectInspector` and `handlers.ResourceHandler` interfaces (a `NewCommonHandler` built from your type + an `ObjectInspector` covers most cases).
3. Register the handler's init function in `handlerInitFuncs` in `pkg/runai_agent.go`.
4. Ensure the scheme includes the resource's API version (in `cmd/main.go` and `pkg/tests/suite_test.go`).
5. Add tests for the new resource.
