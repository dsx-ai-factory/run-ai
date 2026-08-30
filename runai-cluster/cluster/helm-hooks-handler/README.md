# helm-hooks-handler

`helm-hooks-handler` is the runai-cluster binary that runs the **Helm lifecycle
hooks** for the Run:AI cluster chart (`runai-chart`). It is not a long-running
service: the chart schedules it as short-lived `Job`s at defined points in the
install / upgrade / delete lifecycle (`pre-install`, `post-install`,
`pre-upgrade`, `pre-crd-upgrade`, `pre-resource-management-upgrade`,
`post-upgrade`, `pre-delete`) and for standalone config validation
(`validate-cluster-config`). A single flag selects which hook to run; the binary
builds the appropriate hook, executes its tasks against the cluster and the
control plane, and exits. It handles control-plane handshake, image-pull-secret
and custom-CA propagation, CRD/schema upgrades, version/queue validation, and
one-off migrations (engine→KAI, resource-management, department CRD migration).
It is owned by the **Griffin Team**.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)

## Overview

The binary selects exactly one hook (`internal/hooks/factory`) based on the
lifecycle flag it is invoked with, then runs that hook's `Call`. Hooks include:

- **`pre-install`** — prepares the cluster before first install: control-plane
  connectivity/handshake, custom-CA trust, image-pull secrets, base resources.
- **`post-install`** — finishes install-time setup after the operator is up.
- **`pre-upgrade`** — pre-checks and resource preparation before an upgrade.
- **`pre-crd-upgrade`** — applies/patches resources ahead of CRD changes; when
  `--engine-to-kai-migration-enabled` is set, tears down leftover engine
  `SchedulingShard`s (self-gating — no-op when none exist) so the deprecated
  engine scheduler never runs alongside the KAI scheduler.
- **`pre-resource-management-upgrade`** — migrates legacy departments
  (`scheduling.run.ai/v2alpha2`) to KAI departments (`kai.resources/v1alpha1`)
  when `--resource-management-migration-enabled` is set (self-gating on the
  legacy CRD being present).
- **`post-upgrade`** — post-upgrade validation and cleanup, including
  queue-version validation and (with `--resource-management-migration-cleanup`)
  deleting the legacy department CRD once migration is complete.
- **`pre-delete`** — cleanup run before the release is removed.
- **`validate-cluster-config`** — validates the supplied `clusterConfig` values
  and exits (no mutation).

Supporting packages implement the control-plane details download
(`internal/control-plane-details-downloader`), image-pull-secret actions
(`internal/image-pull-secret`), dynamic schema upgrades and RunaiConfig
patching (`internal/cluster-actions`), and Kubernetes client bootstrap
(`internal/clients`). Hooks that don't touch the cluster can run with
`--requires-kubernetes-access=false`.

## Repository layout

```text
.
├── cmd/
│   └── main.go                          # Entry point: parse options, build hook, Call()
├── internal/
│   ├── installation/util/               # Options parsing (flags + params.json configmap)
│   ├── clients/                         # Kubernetes + control-plane client bootstrap
│   ├── control-plane-details-downloader/# Fetch control-plane details/handshake
│   ├── image-pull-secret/               # Image-pull secret create/propagate actions
│   ├── hooks/
│   │   ├── factory/                     # HookFactory: flag → concrete hook
│   │   ├── pre-hooks/                   # pre-install/upgrade/crd-upgrade/delete/validate
│   │   └── post-hooks/                  # post-install/upgrade + migrations & cleanup
│   └── cluster-actions/                 # Dynamic schema upgrades, RunaiConfig patch/backup
├── Dockerfile
├── Makefile
└── go.mod
```

## Building and running

Built with Go (see [`go.mod`](go.mod) — module
`github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler`, Go 1.26.3)
and orchestrated through the [`Makefile`](Makefile), which includes
`../../../build/makefile/runai-cluster.mk` and `index.mk`. Setting `DEBUG=1`
builds with `-gcflags="all=-N -l"` and uses `Dockerfile.debug`.

| Target | Description |
|--------|-------------|
| `make build` | Builds the service (`build-go` → both amd64 and arm64 binaries into `bin/`). |
| `make build-go-amd` | Builds the amd64 binary (`bin/helm-hooks-handler-amd64`); honors `OS=`/`ARCH=`. |
| `make build-go-arm` | Builds the arm64 binary (`bin/helm-hooks-handler-arm64`); hardcodes `GOOS=linux GOARCH=arm64`. |
| `make format` | Formats the code (`format-go`). |
| `make lint` | Runs the Go linter (`lint-go`). |
| `make validate` | No-op (kept for parity across runai-cluster services). |
| `make local-deploy` | Builds (amd64) and pushes the image (`build-go docker-build-generic docker-push-generic`). |

### Local build (local-dev.sh)

`helm-hooks-handler` is a **runai-cluster Go service** (listed in
`RC_GO_SERVICES` in [`local-dev.sh`](../../../local-dev.sh)). `local-dev.sh` builds
amd64 only (Kind nodes are linux/amd64):

```bash
# amd64 — what local-dev.sh runs (build_go_service)
make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
docker build --build-arg TARGETARCH=amd64 -t runai-local/helm-hooks-handler:0.0.0-dev .
```

For an arm64 cluster, build the arm64 path manually (`build-go-arm` **hardcodes**
`GOOS=linux GOARCH=arm64` — do **not** pass `OS=`/`ARCH=`):

```bash
# arm64 — manual
make build-go-arm DOCKERHUB_PROXY= GITHUB_TOKEN=
docker build --build-arg TARGETARCH=arm64 -t runai-local/helm-hooks-handler:0.0.0-dev .
```

- The image tag is **`0.0.0-dev`** (the `RC_IMAGE_TAG`, hardcoded by the
  runai-cluster chart via `helmHooksHandler.image.tag`); any other tag breaks the
  local deploy. Image ref: `runai-local/helm-hooks-handler:0.0.0-dev`.
- `DOCKERHUB_PROXY=` (empty) pulls the golang builder image directly from Docker
  Hub; `GITHUB_TOKEN=` is empty in the make call (the private-module token is
  applied elsewhere).
- The chart runs this image as hook `Job`s — see
  [`runai-chart/templates/hooks`](../runai-chart/templates/hooks). To iterate
  locally, rebuild and re-deploy the RC chart:
  ```bash
  ./local-dev.sh --skip-build-cp --skip-cluster
  ```

### Dockerfile

Multi-stage, producing a minimal `scratch` image:

- **`base`** — `registry.access.redhat.com/ubi9/ubi-minimal`; sources
  `/etc/passwd`, `/etc/group`, SSL certs (`/etc/ssl/certs`, `/etc/pki`),
  `nsswitch.conf`, and a pre-permissioned writable `/tmp` (owned `65532:0`, mode
  `770`).
- **`scratch` (final)** — copies those identity/TLS/DNS files and the pre-built
  `bin/helm-hooks-handler-$TARGETARCH` binary (renamed `helm-hooks-handler`),
  plus the `NOTICE`. Runs as non-root `USER 65532:65532`,
  `ENTRYPOINT ["./helm-hooks-handler"]`.

Build arg: `TARGETARCH` (amd64/arm64) — selects which pre-built binary is copied in.

## Testing

| Target | Description |
|--------|-------------|
| `make test` | Runs unit tests via `test-go` (containerized). |
| `make test-dev` | Runs `go test -v ./... -coverprofile cover.out` directly on the host. |

Tests are colocated with the hooks and helpers (`internal/**/*_test.go`,
including Ginkgo suites). No docker-compose/backing services required.

## Configuration

Configuration is **command-line flag driven** (`internal/installation/util/options.go`);
each hook is selected by its own boolean flag. Optional runtime parameters are
also loaded from a JSON configmap file (`--cm-file-path`, default
`/runai/config/params.json`) into `CMConfigParams` (affinity, tolerations,
management Prometheus ingress, and the `clusterConfig` merge-patch). The chart
mounts and passes these values.

**Hook-selection flags** (exactly one is set per invocation): `--pre-install`,
`--post-install`, `--pre-upgrade`, `--pre-crd-upgrade`,
`--pre-resource-management-upgrade`, `--post-upgrade`, `--pre-delete`,
`--validate-cluster-config`.

**Key configuration flags:**

| Flag | Description | Default |
|------|-------------|---------|
| `--control-plane-url` | Control-plane URL to connect the cluster to | `""` |
| `--control-plane-client-id` | Client ID for control-plane auth | `""` |
| `--control-plane-client-secret` | Client secret; accepts `@/path/to/file` to read from a file | `""` |
| `--cluster-uid` | UID of the cluster to install/upgrade | `""` |
| `--cluster-url` | URL of the cluster | `""` |
| `--cluster-version` | Version of the cluster to install/upgrade | `""` |
| `--custom-ca-enabled` / `--custom-ca-secret-name` / `--custom-ca-secret-key` / `--custom-ca-secret-optional` | Trust a custom CA certificate from a secret | `false` / `""` / `""` / `false` |
| `--image-registry` | Image registry to use across runai-operator services | `""` |
| `--additional-image-pull-secrets` | Additional image-pull secret names (repeatable) | — |
| `--cm-file-path` | Path to the params configmap JSON | `/runai/config/params.json` |
| `--nmc-enabled` / `--nmc-remote-cluster-api-server-url` / `--nmc-service-account-name` / `--nmc-allowed-service-accounts` | NMC (node management) install options | `false` / `""` / `""` / — |
| `--development-headless` | Install without control-plane communication | `false` |
| `--development-skip-dependency-checks` | Skip dependency checks | `false` |
| `--disable-infrastructure-metrics-remote-write` | Disable infra-metrics remote write | `false` |
| `--queue-validation-timeout-millis` | Queue validation timeout | `600000` |
| `--queue-validation-interval-millis` | Queue validation poll interval | `3000` |
| `--operand-down-interval-millis` | Operand-down validation interval | `5000` |
| `--engine-to-kai-migration-enabled` | Tear down leftover engine SchedulingShards before KAI (self-gating) | `false` |
| `--resource-management-migration-enabled` | Migrate legacy departments → KAI departments (self-gating) | `false` |
| `--resource-management-migration-cleanup` | Delete legacy department CRD after migration (self-gating) | `false` |
| `--enable-rbac-escalation` | Allow operands to escalate their RBAC | `true` |
| `--requires-kubernetes-access` | Whether the hook needs cluster access | `true` |
| `--fips-mode` | Go FIPS 140-3 runtime mode (`off`/`on`/`only`) | `off` |
| `--log-level` | Logging verbosity | `0` |

`params.json` (loaded via `--cm-file-path`) carries: `affinity`, `tolerations`,
`mgmtPrometheusIngress`, and `clusterConfig` (RFC 7396 JSON Merge Patch applied
to the RunaiConfig).
