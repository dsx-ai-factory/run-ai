# advice-restart-controller

`advice-restart-controller` is the runai-cluster component that turns **Heimdall
"advice"** into concrete workload actions. Heimdall (NVIDIA's fault-detection
system) writes `Advice` custom resources (`heimdall.nvidia.com/v1alpha1`) that
describe a detected issue for a job (crash, hang, straggler, …) and a
recommended remediation. This controller watches those `Advice` objects and,
when the recommendation is to **restart** the job, locates the owning Run:AI
workload and creates a `RestartRequest` (`run.ai/v1alpha1`) that the rest of the
platform acts on — optionally carrying node-exclusion constraints so the job is
rescheduled away from faulty nodes. It is owned by the **Fox Team**.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)

## Overview

The service is responsible for:

- **Reconciling `Advice` into `RestartRequest`** — the `AdviceController`
  (`pkg/controller/advice_controller.go`) watches `Advice` CRs. For advice whose
  recommended state is `ADVICE_STATE_RESTARTED`, it resolves the **top-owner**
  workload of the referenced job (via `runai-cluster/common/top-owner`), checks
  eligibility, and creates a `RestartRequest` labeled back to the advice. Advice
  in non-restart states is ignored.
- **Idempotency and retry limits** — before acting, the controller lists
  existing `RestartRequest`s linked to an advice (label `LabelAdviceName`) and
  skips/requeues based on their phase (`Succeeded` → done, `InProgress`/`Pending`
  → requeue after 30s, `Failed` → counts toward `MAX_RETRIES_PER_ADVICE`).
- **Workload-type handlers** — restart behavior is pluggable per workload GVK
  (`internal/workloadtype/`); currently a handler is registered for the
  `WorkloadRunner` GVK. Node exclusions from the advice are merged into the
  restart request: for a `WorkloadRunner` they become a single `add` of
  `/spec/instructions/nodeAffinity`, unioned with the nodes that instruction
  already excludes.
- **`log` vs `action` mode** — `MODE=log` observes and emits events without
  creating restart requests; `MODE=action` performs the restart. Any other value
  panics at startup.
- **Events & health/metrics** — emits Kubernetes events on the workload, serves
  a metrics endpoint (`METRICS_BIND_ADDRESS`) and `healthz`/`readyz` probes.

## Repository layout

```text
.
├── cmd/
│   └── main.go                 # Entry point: config, scheme, manager, controller
├── pkg/
│   ├── config/                 # AdviceRestartControllerConfig (env-var driven)
│   └── controller/             # AdviceController: reconcile, events, constants
├── internal/
│   ├── nodeexclusion/          # Merge advice node exclusions into restart requests
│   ├── workloadtype/           # Per-workload-type restart handlers (GVK → factory)
│   └── tests/                  # envtest integration suite (testdata: CRDs, TLS certs)
├── Dockerfile
├── Makefile
└── go.mod
```

## Building and running

Built with Go (see [`go.mod`](go.mod) — module
`github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/advice-restart-controller`,
Go 1.26.3) and orchestrated through the [`Makefile`](Makefile), which includes
`../../../build/makefile/runai-cluster.mk` and `index.mk`. Build/test targets run
inside a golang Docker container.

| Target | Description |
|--------|-------------|
| `make build` | Builds the service (`build-go` → both amd64 and arm64 binaries into `bin/`). |
| `make build-go-amd` | Builds the amd64 binary (`bin/advice-restart-controller-amd64`); honors `OS=`/`ARCH=`. |
| `make build-go-arm` | Builds the arm64 binary (`bin/advice-restart-controller-arm64`); hardcodes `GOOS=linux GOARCH=arm64`. |
| `make format` | Formats the code (`format-go`). |
| `make lint` | Runs the Go linter (`lint-go`). |
| `make validate` | No-op (kept for parity across runai-cluster services). |
| `make local-deploy` | Builds (amd64) and pushes the image (`build-go docker-build-generic docker-push-generic`). |

### Local build (local-dev.sh)

`advice-restart-controller` is a **runai-cluster Go service** (listed in
`RC_GO_SERVICES` in [`local-dev.sh`](../../../local-dev.sh)). `local-dev.sh` builds
amd64 only (Kind nodes are linux/amd64):

```bash
# amd64 — what local-dev.sh runs (build_go_service)
make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
docker build --build-arg TARGETARCH=amd64 -t runai-local/advice-restart-controller:0.0.0-dev .
```

For an arm64 cluster, build the arm64 path manually (`build-go-arm` **hardcodes**
`GOOS=linux GOARCH=arm64` — do **not** pass `OS=`/`ARCH=`):

```bash
# arm64 — manual
make build-go-arm DOCKERHUB_PROXY= GITHUB_TOKEN=
docker build --build-arg TARGETARCH=arm64 -t runai-local/advice-restart-controller:0.0.0-dev .
```

- The image tag is **`0.0.0-dev`** (the `RC_IMAGE_TAG`, hardcoded by the
  runai-cluster chart values); any other tag breaks the local deploy. Image ref:
  `runai-local/advice-restart-controller:0.0.0-dev`.
- `DOCKERHUB_PROXY=` (empty) pulls the golang builder image directly from Docker
  Hub; `GITHUB_TOKEN=` is empty in the make call (the private-module token is
  applied elsewhere).
- The tag is the same for both arches — only `--build-arg TARGETARCH` changes.

To iterate locally against a running cluster:

```bash
./local-dev.sh --skip-build-cp --skip-cluster
```

### Dockerfile

Multi-stage, producing a minimal `scratch` image:

- **`base`** — `registry.access.redhat.com/ubi9/ubi-minimal`; used only to source
  `/etc/passwd`, `/etc/group`, SSL certs (`/etc/ssl/certs`, `/etc/pki`),
  `nsswitch.conf`, and a pre-permissioned writable `/tmp` (owned `65532:0`, mode
  `770`).
- **`scratch` (final)** — copies those identity/TLS/DNS files and the pre-built
  `bin/advice-restart-controller-$TARGETARCH` binary (renamed
  `advice-restart-controller`), plus the `NOTICE`. Runs as non-root
  `USER 65532:65532`, `ENTRYPOINT ["./advice-restart-controller"]`.

Build arg: `TARGETARCH` (amd64/arm64) — selects which pre-built binary is copied in.

## Testing

| Target | Description |
|--------|-------------|
| `make test` | Runs unit + env-tests via `envtest-docker-go` (runs `envtest-go` inside a privileged golang container). |
| `make test-dev` | Runs `envtest-go` directly on the host. |

Tests are [Ginkgo](https://onsi.github.io/ginkgo/) suites (`pkg/**/*_test.go`,
`internal/**/*_test.go`). The `internal/tests` suite uses controller-runtime
[envtest](https://book.kubebuilder.io/reference/envtest.html): it downloads
Kubernetes control-plane binaries via `setup-envtest`
(`KUBEBUILDER_ASSETS`, K8s `1.33`) and loads test CRDs/certs from
`internal/tests/testdata/`.

## Configuration

Configuration is **environment-variable driven** (the uppercase `mapstructure`
tags in [`pkg/config/config.go`](pkg/config/config.go) are the env-var names).
Defaults come from `SetDefaults()`. There is no YAML app-config file for this
service.

| Variable | Description | Default |
|----------|-------------|---------|
| `MODE` | `action` (create restart requests) or `log` (observe only); any other value panics at startup | `action` |
| `RESTART_LIMIT` | Restart limit passed to the controller | `3` |
| `MAX_RETRIES_PER_ADVICE` | Max failed `RestartRequest`s tolerated per advice before giving up | `3` |
| `USE_DEBUG_LOG_LEVEL` | Enable debug logging | `false` |
| `K8S_CLIENT_CONFIG_QPS` | Kubernetes client QPS | `50` |
| `K8S_CLIENT_CONFIG_BURST` | Kubernetes client burst | `300` |
| `METRICS_BIND_ADDRESS` | Prometheus metrics bind address | `:8383` |

The Dockerfile sets `TARGETARCH` (build arg → env) to select the binary; it
defines no runtime `ENV` configuration beyond that.
