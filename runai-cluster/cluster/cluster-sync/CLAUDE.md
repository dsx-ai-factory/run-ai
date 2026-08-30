# cluster-sync — Build Context

**Type:** container (runai-cluster Go service)
**Team:** Suricata Team (`artifact-metadata.json` → `jira.team_name`; Slack `@suricata`)
**Module:** `github.com/run-ai/runai/runai-cluster/cluster/cluster-sync` (Go 1.26)

## What it is

controller-runtime reconcilers (one per resource kind — nodes, pods, projects,
departments, workloads, PVCs, secrets, config maps, policies, data volumes, etc.)
that watch cluster objects and push their state to the Run.ai backend over
authenticated HTTP using generated service clients. Outbound half of the
cluster↔backend sync (inbound is `agent`).

## Make targets

| Target | What it does |
|--------|-------------|
| build  | `build-go` → builds both amd64 and arm64 binaries into `bin/` |
| build-go-amd | amd64 binary (honors `OS=`/`ARCH=`) → `bin/cluster-sync-amd64` |
| build-go-arm | arm64 binary (hardcodes `GOOS=linux GOARCH=arm64`) → `bin/cluster-sync-arm64` |
| test   | `test-go` (containerized unit tests) |
| test-dev | `go test -v ./... -coverprofile cover.out` on host |
| lint   | `lint-go` |
| format | `format-go` |
| validate | no-op |
| local-deploy | `build-go docker-build-generic docker-push-generic` (amd64) |
| copy-clients | Copies generated swagger clients from `runai-common-packages/clients` into `pkg/generated/` |

Includes: `../../../build/makefile/runai-cluster.mk`, `index.mk`.

## Key build notes

- **runai-cluster Go service.** local-dev.sh (`RC_GO_SERVICES` → `build_go_service`) runs:
  ```bash
  make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
  docker build --build-arg TARGETARCH=amd64 -t runai-local/cluster-sync:0.0.0-dev .
  ```
- **Image tag: `0.0.0-dev`** (RC tag, hardcoded in the runai-cluster chart values). Ref: `runai-local/cluster-sync:0.0.0-dev`.
- **Both arches:** amd64 (`build-go-amd`, reads `OS=`/`ARCH=`) and arm64 (`build-go-arm`, hardcodes `GOOS=linux GOARCH=arm64` — do NOT pass `OS=`/`ARCH=`). Same tag; only `--build-arg TARGETARCH` differs. local-dev.sh builds amd64 only (Kind nodes are linux/amd64).
- **Generated code:** `pkg/generated/` swagger clients are generated — run `make copy-clients` (copies from `runai-common-packages/clients`; current `CLIENTS` = `k8s-objects-tracker-client policy-service-client`). k8s clientset generated via `tools.go` code-generator. Do not edit generated dirs by hand.
- **Base image:** Dockerfile is `ubi9/ubi-minimal` (for certs/passwd/tmp) → final `FROM scratch`, non-root `USER 65532:65532`, `ENTRYPOINT ["./cluster-sync"]`. Build arg `TARGETARCH` selects `bin/cluster-sync-$TARGETARCH`.
- **Artifact caching:** `artifact-metadata.json` sets `cache_artifacts: true`, `path_to_artifacts: bin/`.
- **Config:** loaded via `runai-common-packages/config` (mapstructure env tags in `pkg/config/config.go`); secrets (`CLUSTER_ID`, `CLIENT_NAME`, `CLIENT_SECRET`) from mounted secret. Out-of-cluster fallback reads `KUBE_CONFIG_FILE_PATH`.
- **Tests:** unit `*_test.go` + Ginkgo service suites under `test/` (one per resource kind, plus `fakecachedclient`). No docker-compose needed.

## local-dev.sh iteration

runai-cluster service:
```bash
./local-dev.sh --skip-build-cp --skip-cluster
```
