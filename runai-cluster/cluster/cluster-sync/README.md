# cluster-sync

`cluster-sync` is the runai-cluster component that keeps the Run.ai control plane ("runai-backend") in sync with the live state of a Kubernetes cluster. It runs a large set of [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) reconcilers — one per resource kind (nodes, node pools, pods, projects, departments, workloads, PVCs, secrets, config maps, storage/priority classes, access rules, policies, data volumes, network topologies, and more) — watches those objects in the cluster, and pushes their state to the backend over authenticated HTTP using the generated service clients. It is owned by the **Suricata Team**. This is the outbound half of the cluster↔backend data path (the inbound half is handled by the `agent` service).

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [Code generation](#code-generation)

## Overview

The service is responsible for:

- **Reconciling cluster resources to the backend** — each `pkg/controllers/*_controller.go` watches one resource kind and enqueues changes; the matching `pkg/handlers/*_handler.go` transforms the Kubernetes object into the backend API payload and sends it via the generated client.
- **Authenticating to the backend** — exchanges `CLIENT_NAME` / `CLIENT_SECRET` (an application credential) for a token against the identity manager, then calls the Run.ai API at `ENDPOINT` on behalf of cluster `CLUSTER_ID`.
- **Batched and rate-limited syncing** — requests are chunked (`MAX_NUM_OF_*_PER_REQUEST`) and dispatched through two work-queue rate limiters: a "workloads" limiter (pods/workloads) and a "regular" limiter (projects, node pools, etc.), each with its own delay and worker count.
- **Periodic full re-sync and liveness** — a full cluster-info sync runs on an interval (`SYNC_INTERVAL_MINUTES` / `CLUSTER_INFO_SYNC_INTERVAL_SECONDS`), and a liveness loop reports health on `LIVENESS_INTERVAL_SECONDS`.
- **NVIDIA GPU node readiness** — optionally reports node GPU readiness based on NVIDIA component labels (container-toolkit, DCGM exporter).
- **Metrics and profiling** — exposes Prometheus metrics (`METRICS_PORT`/`METRICS_PATH`) and an optional pprof profiler.

## Repository layout

```text
.
├── cmd/
│   ├── main.go             # Entry point: sets up scheme, k8s client, controllers, backend client
│   └── profiling/          # Optional pprof profiler wiring
├── pkg/
│   ├── backend-client/     # Wrapper around the generated backend HTTP clients + auth
│   ├── cachedclient/       # Cached Kubernetes client used by controllers
│   ├── config/             # ClusterSyncConfig struct, env var mapping and defaults
│   ├── constants/          # Shared constants (labels, mount paths)
│   ├── controllers/        # One reconciler per resource kind (nodes, pods, projects, …)
│   ├── handlers/           # Per-kind translation to backend API payloads (+ parsers/)
│   ├── generated/          # Generated swagger clients (do not edit — see copy-clients)
│   ├── logger/             # zap logger setup
│   ├── metrics/            # Prometheus metrics
│   └── util/               # Shared helpers
├── test/                   # Ginkgo service/integration tests (one *_test.go per kind)
├── scripts/deploy.sh       # Deploy helper
├── assets/                 # Docs images (run/debug config)
├── tools.go                # Tool dependency pins
├── Dockerfile
├── Makefile
└── go.mod
```

## Building and running

Built with Go (see [`go.mod`](go.mod) — module `github.com/run-ai/runai/runai-cluster/cluster/cluster-sync`, Go 1.26) and orchestrated through the [`Makefile`](Makefile), which includes `../../../build/makefile/runai-cluster.mk` and `index.mk`. Build/test targets run inside a golang Docker container.

| Target | Description |
|--------|-------------|
| `make build` | Builds the service (`build-go` → both amd64 and arm64 binaries into `bin/`). |
| `make build-go-amd` | Builds the amd64 binary (`bin/cluster-sync-amd64`); honors `OS=`/`ARCH=`. |
| `make build-go-arm` | Builds the arm64 binary (`bin/cluster-sync-arm64`); hardcodes `GOOS=linux GOARCH=arm64`. |
| `make format` | Formats the code (`format-go`). |
| `make lint` | Runs the Go linter (`lint-go`). |
| `make validate` | No-op (kept for target parity across runai-cluster services). |
| `make local-deploy` | Builds (amd64) and pushes the image (`build-go docker-build-generic docker-push-generic`). |
| `make copy-clients` | Refreshes the generated swagger clients from `runai-common-packages/clients` (see [Code generation](#code-generation)). |

### Local build (local-dev.sh)

`cluster-sync` is a **runai-cluster Go service** (listed in `RC_GO_SERVICES` in [`local-dev.sh`](../../../local-dev.sh)). `local-dev.sh` builds amd64 only (Kind nodes are linux/amd64):

```bash
# amd64 — what local-dev.sh runs (build_go_service)
make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
docker build --build-arg TARGETARCH=amd64 -t runai-local/cluster-sync:0.0.0-dev .
```

For an arm64 cluster, build the arm64 path manually (note: `build-go-arm` **hardcodes** `GOOS=linux GOARCH=arm64` — do **not** pass `OS=`/`ARCH=`):

```bash
# arm64 — manual
make build-go-arm DOCKERHUB_PROXY= GITHUB_TOKEN=
docker build --build-arg TARGETARCH=arm64 -t runai-local/cluster-sync:0.0.0-dev .
```

- The image tag is **`0.0.0-dev`** (the `RC_IMAGE_TAG`, hardcoded by the runai-cluster chart values); any other tag breaks the local deploy. Image ref: `runai-local/cluster-sync:0.0.0-dev`.
- `DOCKERHUB_PROXY=` (empty) pulls the golang builder image directly from Docker Hub; `GITHUB_TOKEN=` is empty in the make call (the private-module token is applied elsewhere).
- The tag is the same for both arches — only `--build-arg TARGETARCH` changes.

### Dockerfile

Multi-stage, producing a minimal `scratch` image:

- **`base`** — `registry.access.redhat.com/ubi9/ubi-minimal`; used only to source `/etc/passwd`, `/etc/group`, SSL certs (`/etc/ssl/certs`, `/etc/pki`), `nsswitch.conf`, and a pre-permissioned writable `/tmp` (owned `65532:0`, mode `770`).
- **`scratch` (final)** — copies those identity/TLS/DNS files and the pre-built `bin/cluster-sync-$TARGETARCH` binary (renamed `cluster-sync`), plus the `NOTICE`. Runs as non-root `USER 65532:65532`, `ENTRYPOINT ["./cluster-sync"]`.

Build arg: `TARGETARCH` (amd64/arm64) — selects which pre-built binary is copied in.

### Running locally (out-of-cluster)

When `rest.InClusterConfig()` fails, `main.go` falls back to the kubeconfig at `KUBE_CONFIG_FILE_PATH`. Set the backend credentials/endpoint from your cluster's runai secret:

```bash
export KUBE_CONFIG_FILE_PATH=~/.kube/config
export CLUSTER_ID=... CLIENT_NAME=... CLIENT_SECRET=... ENDPOINT=https://<backend>
```

Pass `--log-level -1` for debug logging.

## Testing

| Target | Description |
|--------|-------------|
| `make test` | Runs unit tests via `test-go` (containerized). |
| `make test-dev` | Runs `go test -v ./... -coverprofile cover.out` directly on the host. |

Unit tests live next to the code (`*_test.go`). The `test/` directory holds [Ginkgo](https://onsi.github.io/ginkgo/) service tests — one suite per synced resource kind (nodes, projects, workloads, PVCs, secrets, policies, etc.) plus fakes (`fakecachedclient`) and shared helpers (`backend_testing_utils.go`, `util.go`). Individual suites can be run with the `ginkgo` CLI.

## Configuration

Configuration is **environment-variable driven** (the uppercase `mapstructure` tags in [`pkg/config/config.go`](pkg/config/config.go) are the env-var names). Defaults come from `SetDefaults()`. Values marked `config:"secret"` (`CLUSTER_ID`, `CLIENT_NAME`, `CLIENT_SECRET`) are instead read from a mounted secret file — a JSON file at `/etc/secret/secret.json` (`constants.SecretMountPath`). There is no YAML app-config file for this service.

| Variable | Description | Default |
|----------|-------------|---------|
| `CLUSTER_ID` | Cluster UUID (secret) | — |
| `CLUSTER_DOMAIN` | Cluster domain | — |
| `CLIENT_NAME` | Application client name for backend auth (secret) | — |
| `CLIENT_SECRET` | Application client secret (secret) | — |
| `ENDPOINT` | Base URL of the Run.ai backend | — |
| `KUBE_CONFIG_FILE_PATH` | Kubeconfig path for out-of-cluster runs (read in `main.go`) | — |
| `MAX_NUM_OF_EVENTS_PER_REQUEST` | Event batch size | `1000` |
| `MAX_NUM_OF_PODS_PER_REQUEST` | Pod batch size | `1800` |
| `MAX_NUM_OF_NODES_PER_REQUEST` | Node batch size | `1000` |
| `MAX_NUM_OF_WORKLOADS_PER_REQUEST` | Workload batch size | `1000` |
| `MAX_NUM_OF_DATAVOLUMES_PER_REQUEST` | Data volume batch size | `2000` |
| `SYNC_EVENTS_INTERVAL_SECONDS` | Event sync interval | `10` |
| `LIVENESS_INTERVAL_SECONDS` | Liveness report interval | `10` |
| `CLUSTER_INFO_SYNC_INTERVAL_SECONDS` | Cluster-info sync interval | `10` |
| `MAX_NUMBER_OF_RETRIES` | Retries for token retrieval / liveness budget | `10` |
| `RETRIES_INTERVAL_SECONDS` | Retry interval | `10` |
| `SYNC_INTERVAL_MINUTES` | Full re-sync interval | `180` |
| `EXISTING_SECRET_TO_SYNC_LABEL` | Label selecting secrets to sync | `RunaiResourceLabel` |
| `ASSETS_SECRET_TO_SYNC_LABEL` | Label selecting asset secrets to sync | `RunaiAssetLabel` |
| `CONFIG_MAP_TO_SYNC_LABEL` | Label selecting config maps to sync | `RunaiResourceLabel` |
| `SSL_CERTIFICATE_VERIFICATION` | Verify backend TLS certificate | `true` |
| `METRICS_PORT` | Prometheus metrics port | `9400` |
| `METRICS_PATH` | Prometheus metrics path | `/metrics` |
| `ENABLE_PROFILER` | Enable pprof profiler | `false` |
| `PROFILER_API_PORT` | Profiler port | `8182` |
| `K8S_CLIENT_CONFIG_QPS` | Kubernetes client QPS | `50` |
| `K8S_CLIENT_CONFIG_BURST` | Kubernetes client burst | `300` |
| `NVIDIA_GPU_LABEL` | Node label marking a GPU node | `feature.node.kubernetes.io/pci-10de.present` |
| `NODE_LABEL_NVIDIA_CONTAINER_TOOLKIT` | Node label for NVIDIA container toolkit readiness | `nvidia.com/gpu.deploy.container-toolkit` |
| `NODE_LABEL_NVIDIA_DCGM_EXPORTER` | Node label for DCGM exporter readiness | `nvidia.com/gpu.deploy.dcgm-exporter` |
| `ENABLE_NODE_NVIDIA_COMPONENTS_READINESS` | Report NVIDIA component readiness | `false` |
| `SYNC_PODS_WITH_WORKLOADS` | Sync pods together with workloads | `false` |
| `RATE_LIMITER_WORKLOADS_DELAY_SECONDS` | Workloads limiter delay | `10` |
| `RATE_LIMITER_WORKLOADS_NUM_OF_WORKERS` | Workloads limiter workers | `3` |
| `RATE_LIMITER_REGULAR_DELAY_SECONDS` | Regular limiter delay | `5` |
| `RATE_LIMITER_REGULAR_NUM_OF_WORKERS` | Regular limiter workers | `1` |
| `LEADER_ELECT` | Enable leader election | `false` |

The Dockerfile sets `TARGETARCH` (build arg → env) to select the binary; it defines no runtime `ENV` configuration beyond that.

## Code generation

`cluster-sync` talks to the backend through generated OpenAPI clients under `pkg/generated/` — **these directories are generated and must not be edited by hand.** Rather than regenerating in place, the clients are copied from the canonical generated sources in `runai-common-packages/clients`:

```bash
make copy-clients
```

The `copy-clients` target copies `runai-common-packages/clients/<client>/generated/*` into `pkg/generated/swagger-runai-<client>/`. The active `CLIENTS` set in the Makefile is currently `k8s-objects-tracker-client` and `policy-service-client` (other clients — assets/authorization/workloads/etc. — are consumed as Go module dependencies via `go.mod`; see the `TODO` note in the Makefile about migrating the remaining ones to the new generator).

The Kubernetes clientset under `pkg/` is likewise generated (via the code-generator tooling pinned in `tools.go`); regenerate it with the update-client script rather than editing by hand.
