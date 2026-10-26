# advice-restart-controller — Build Context

**Type:** container (runai-cluster Go service)
**Team:** Fox Team (`artifact-metadata.json` → `jira.team_name`; Slack `@fox`)
**Module:** `github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/advice-restart-controller` (Go 1.26.3)

## What it is

Watches Heimdall `Advice` CRDs (`heimdall.nvidia.com/v1alpha1`). When an advice
recommends `ADVICE_STATE_RESTARTED`, resolves the top-owner Run:AI workload and
creates a `RestartRequest` (`run.ai/v1alpha1`), merging in any node exclusions.
Idempotent/retry-bounded via `RestartRequest` phase + `MAX_RETRIES_PER_ADVICE`.
`MODE=action` performs restarts; `MODE=log` only observes.

## Make targets

| Target | What it does |
|--------|-------------|
| build  | `build-go` → builds both amd64 and arm64 binaries into `bin/` |
| build-go-amd | amd64 binary (honors `OS=`/`ARCH=`) → `bin/advice-restart-controller-amd64` |
| build-go-arm | arm64 binary (hardcodes `GOOS=linux GOARCH=arm64`) → `bin/advice-restart-controller-arm64` |
| test   | `envtest-docker-go` (runs `envtest-go` in a privileged golang container; Ginkgo + envtest) |
| test-dev | `envtest-go` on host (downloads `setup-envtest` KUBEBUILDER_ASSETS, K8s 1.33) |
| lint   | `lint-go` |
| format | `format-go` |
| validate | no-op |
| local-deploy | `build-go docker-build-generic docker-push-generic` (amd64) |

Includes: `../../../build/makefile/runai-cluster.mk`, `index.mk`.

## Key build notes

- **runai-cluster Go service.** local-dev.sh (`RC_GO_SERVICES` → `build_go_service`) runs:
  ```bash
  make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
  docker build --build-arg TARGETARCH=amd64 -t runai-local/advice-restart-controller:0.0.0-dev .
  ```
- **Image tag: `0.0.0-dev`** (RC tag, hardcoded in the runai-cluster chart values). Ref: `runai-local/advice-restart-controller:0.0.0-dev`.
- **Both arches:** amd64 (`build-go-amd`, reads `OS=`/`ARCH=`) and arm64 (`build-go-arm`, hardcodes `GOOS=linux GOARCH=arm64` — do NOT pass `OS=`/`ARCH=`). Same tag; only `--build-arg TARGETARCH` differs. local-dev builds amd64 only.
- **Base image:** Dockerfile is `ubi9/ubi-minimal` (for certs/passwd/tmp) → final `FROM scratch`, non-root `USER 65532:65532`, `ENTRYPOINT ["./advice-restart-controller"]`. Build arg `TARGETARCH` selects `bin/advice-restart-controller-$TARGETARCH`.
- **Artifact caching:** `artifact-metadata.json` sets `cache_artifacts: true`, `path_to_artifacts: bin/`.
- **Config:** env-var driven via `runai-common-packages/config` (uppercase `mapstructure` tags in `pkg/config/config.go`); defaults in `SetDefaults()`. Key knobs: `MODE`, `RESTART_LIMIT`, `MAX_RETRIES_PER_ADVICE`, `METRICS_BIND_ADDRESS` (:8383).
- **Testing:** uses controller-runtime envtest (`setup-envtest`, K8s 1.33) + Ginkgo; test CRDs/certs under `internal/tests/testdata/`. `test` runs privileged in Docker.
- **Scheme/CRDs:** registers Heimdall `Advice` (imported from `github.com/dsx-ai-factory/nvidia-recovery/advice-operator/api/v1alpha1`; envtest loads the Advice CRD from that module's `config/crd/bases` via `go list -m`, so bumping the dependency also updates the test CRD), `run.ai/v1alpha1` (RestartRequest), distributed-workload `v2alpha2`, and kubeflow types in `cmd/main.go`.
- **Chart wiring:** RBAC template at `runai-chart/templates/clusterroles/advice-restart-controller.yaml`.

## local-dev.sh iteration

runai-cluster service:
```bash
./local-dev.sh --skip-build-cp --skip-cluster
```
