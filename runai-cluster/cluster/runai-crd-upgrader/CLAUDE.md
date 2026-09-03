# runai-crd-upgrader — Build Context

**Type:** container (runai-cluster Go service — one-shot CRD job)
**Team:** Griffin Team (`artifact-metadata.json` → `jira.team_name`; Slack `@griffin`)
**Module:** `github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runai-crd-upgrader` (Go 1.26.3)

## What it is

One-shot job that installs/upgrades the Run:AI CRDs. Reads YAMLs from
`--crds-dir` (default `/crds`, baked into the image from `crds/`), server-side
applies each CRD with field manager `runai-crd-upgrader` (bounded exponential
backoff), then polls until each CRD is `Established` before exiting. Runs before
the operator so dependent components find their CRDs served. In-cluster client
tuned to QPS 50 / Burst 100.

## Make targets

| Target | What it does |
|--------|-------------|
| build  | `build-go` → builds both amd64 and arm64 binaries into `bin/` |
| build-go-amd | amd64 binary (honors `OS=`/`ARCH=`) → `bin/runai-crd-upgrader-amd64` |
| build-go-arm | arm64 binary (hardcodes `GOOS=linux GOARCH=arm64`) → `bin/runai-crd-upgrader-arm64` |
| test   | `test-go` (containerized unit tests; `pkg/upgrader`) |
| lint   | `lint-go` |
| format | `format-go` |
| validate | no-op |

Includes: `../../../build/makefile/runai-cluster.mk`, `index.mk`. Sets
`GO_LDFLAGS='-s -w'` (stripped binary). No `local-deploy` target defined.

## Key build notes

- **runai-cluster Go service.** local-dev.sh (`RC_GO_SERVICES` → `build_go_service`) runs:
  ```bash
  make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
  docker build --build-arg TARGETARCH=amd64 -t runai-local/runai-crd-upgrader:0.0.0-dev .
  ```
- **Image tag: `0.0.0-dev`** (RC tag; chart value `crdUpgrader.image.tag`). Ref: `runai-local/runai-crd-upgrader:0.0.0-dev`.
- **Both arches:** amd64 (`build-go-amd`, reads `OS=`/`ARCH=`) and arm64 (`build-go-arm`, hardcodes `GOOS=linux GOARCH=arm64` — do NOT pass `OS=`/`ARCH=`). Same tag; only `--build-arg TARGETARCH` differs. local-dev builds amd64 only.
- **Base image:** Dockerfile is single-stage `FROM scratch` (no ubi9 base — needs no certs/DNS). Copies binary to `/usr/bin/runai-crd-upgrader` and `crds/` to `/crds`, non-root `USER 65532:65532`, `ENTRYPOINT ["/usr/bin/runai-crd-upgrader"]`. Build arg `TARGETARCH` selects `bin/runai-crd-upgrader-$TARGETARCH`.
- **Artifact caching:** `artifact-metadata.json` sets `cache_artifacts: true`, `path_to_artifacts: bin/`; `ngc.guest_access: true`.
- **Config:** flags only — `--crds-dir` (/crds), `--established-timeout` (5m), `--poll-interval` (1s). In-cluster K8s config; client QPS=50/Burst=100.
- **CRDs:** the `crds/` dir is the source of truth for cluster CRDs and is ALSO applied by other services' env-tests (`build/makefile/testenv.mk` → `kubectl create -f ../runai-crd-upgrader/crds`). Keep in sync with `runai-chart/charts/crds/crds`.
- **Deps:** minimal — `k8s.io/apiextensions-apiserver`, `apimachinery`, `client-go` (v0.35.4). No cluster SDK.

## local-dev.sh iteration

runai-cluster service (rebuilds the CRD job image; re-runs on chart upgrade):
```bash
./local-dev.sh --skip-build-cp --skip-cluster
```
