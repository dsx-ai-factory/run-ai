# runai-crd-upgrader

`runai-crd-upgrader` is the runai-cluster job that installs and upgrades the
Run:AI **CustomResourceDefinitions** during a chart install/upgrade. It is a
one-shot binary (not a controller): it reads a directory of CRD YAML manifests
baked into its image, applies each one to the cluster with a server-side apply,
and then waits for every CRD to reach the `Established` condition before exiting.
Running CRD application as a dedicated job — rather than as ordinary chart
templates — lets the chart guarantee CRDs exist and are served before the
operator and other components that depend on them start. It is owned by the
**Griffin Team**.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)

## Overview

The service is responsible for:

- **Applying CRDs** — `pkg/upgrader` loads every YAML document under the
  `--crds-dir` directory (default `/crds`, populated from the image's baked-in
  `crds/`) and applies each `CustomResourceDefinition` via a server-side apply
  using a stable field manager (`runai-crd-upgrader`, persisted in
  `managedFields` and not to be renamed). Applies use a bounded exponential
  backoff to ride out transient API errors.
- **Waiting for `Established`** — after applying, it polls each CRD (every
  `--poll-interval`, default 1s) until the `Established` condition is true,
  giving up after `--established-timeout` (default 5m). This ensures dependent
  workloads only start once the CRDs are actually served by the API server.
- **Tuned client** — it raises the client-go rate limits above the 5 QPS / 10
  burst default (`QPS=50`, `Burst=100`) so the apply + poll loop is not
  throttled client-side (API Priority & Fairness still protects the server), and
  handles `SIGINT`/`SIGTERM` for clean shutdown.

The CRD manifests it ships (`crds/`) cover the Run:AI cluster APIs — clusters,
projects, access rules, applications, external workloads, node pools,
RunaiConfig, the installation `Cluster` CRD, and the optimization
`ResourceInterface` CRD.

## Repository layout

```text
.
├── cmd/
│   └── main.go            # Entry point: flags, in-cluster client, run upgrader
├── pkg/
│   └── upgrader/          # Load YAML → server-side apply → wait for Established
├── crds/                  # CRD YAML manifests baked into the image (applied at runtime)
├── Dockerfile
├── Makefile
└── go.mod
```

## Building and running

Built with Go (see [`go.mod`](go.mod) — module
`github.com/run-ai/runai/runai-cluster/cluster/runai-crd-upgrader`, Go 1.26.3)
and orchestrated through the [`Makefile`](Makefile), which includes
`../../../build/makefile/runai-cluster.mk` and `index.mk`. The Makefile sets
`GO_LDFLAGS='-s -w'` to strip the binary.

| Target | Description |
|--------|-------------|
| `make build` | Builds the service (`build-go` → both amd64 and arm64 binaries into `bin/`). |
| `make build-go-amd` | Builds the amd64 binary (`bin/runai-crd-upgrader-amd64`); honors `OS=`/`ARCH=`. |
| `make build-go-arm` | Builds the arm64 binary (`bin/runai-crd-upgrader-arm64`); hardcodes `GOOS=linux GOARCH=arm64`. |
| `make format` | Formats the code (`format-go`). |
| `make lint` | Runs the Go linter (`lint-go`). |
| `make validate` | No-op (kept for parity across runai-cluster services). |
| `make test` | Runs unit tests (`test-go`). |

### Local build (local-dev.sh)

`runai-crd-upgrader` is a **runai-cluster Go service** (listed in
`RC_GO_SERVICES` in [`local-dev.sh`](../../../local-dev.sh)). `local-dev.sh` builds
amd64 only (Kind nodes are linux/amd64):

```bash
# amd64 — what local-dev.sh runs (build_go_service)
make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
docker build --build-arg TARGETARCH=amd64 -t runai-local/runai-crd-upgrader:0.0.0-dev .
```

For an arm64 cluster, build the arm64 path manually (`build-go-arm` **hardcodes**
`GOOS=linux GOARCH=arm64` — do **not** pass `OS=`/`ARCH=`):

```bash
# arm64 — manual
make build-go-arm DOCKERHUB_PROXY= GITHUB_TOKEN=
docker build --build-arg TARGETARCH=arm64 -t runai-local/runai-crd-upgrader:0.0.0-dev .
```

- The image tag is **`0.0.0-dev`** (the `RC_IMAGE_TAG`, hardcoded by the
  runai-cluster chart via `crdUpgrader.image.tag`); any other tag breaks the
  local deploy. Image ref: `runai-local/runai-crd-upgrader:0.0.0-dev`.
- `DOCKERHUB_PROXY=` (empty) pulls the golang builder image directly from Docker
  Hub; `GITHUB_TOKEN=` is empty in the make call (the private-module token is
  applied elsewhere).
- To iterate locally, rebuild and re-deploy the RC chart:
  ```bash
  ./local-dev.sh --skip-build-cp --skip-cluster
  ```

### Dockerfile

Single-stage `FROM scratch`. It copies the pre-built
`bin/runai-crd-upgrader-$TARGETARCH` binary to `/usr/bin/runai-crd-upgrader`, the
`crds/` directory to `/crds`, and the `NOTICE`. Runs as non-root
`USER 65532:65532`, `ENTRYPOINT ["/usr/bin/runai-crd-upgrader"]`.

Build arg: `TARGETARCH` (amd64/arm64) — selects which pre-built binary is copied
in. (Unlike the other runai-cluster services, this image is not built from a
`ubi9/ubi-minimal` base — it needs no certs/DNS files.)

## Testing

| Target | Description |
|--------|-------------|
| `make test` | Runs unit tests via `test-go` (containerized). |

Unit tests live in `pkg/upgrader/upgrader_test.go`. No docker-compose/backing
services required. Note that `runai-crd-upgrader/crds` is also consumed by other
services' env-tests (see `build/makefile/testenv.mk`, which applies these CRDs
to the kind cluster).

## Configuration

Configuration is **command-line flag driven** (`cmd/main.go`). The Kubernetes
client is built from the in-cluster config. There is no env-var or YAML config.

| Flag | Description | Default |
|------|-------------|---------|
| `--crds-dir` | Directory containing the CRD YAML files to apply | `/crds` |
| `--established-timeout` | Maximum time to wait for all CRDs to become `Established` | `5m` |
| `--poll-interval` | Interval between `Established` checks | `1s` |

The chart runs this image as a hook/init job (`crdUpgrader` values in
[`runai-chart/values.yaml`](../runai-chart/values.yaml): `enabled`,
`establishedTimeout`, image, resources).
