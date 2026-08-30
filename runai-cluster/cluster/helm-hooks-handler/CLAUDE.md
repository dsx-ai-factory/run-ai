# helm-hooks-handler — Build Context

**Type:** container (runai-cluster Go service — Helm hook job)
**Team:** Griffin Team (`artifact-metadata.json` → `jira.team_name`; Slack `@griffin`)
**Module:** `github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler` (Go 1.26.3)

## What it is

Runs the `runai-chart` Helm lifecycle hooks as short-lived Jobs. A single flag
picks the hook (`pre-install`, `post-install`, `pre-upgrade`, `pre-crd-upgrade`,
`pre-resource-management-upgrade`, `post-upgrade`, `pre-delete`,
`validate-cluster-config`); the factory builds it, `Call()` executes, binary
exits. Handles CP handshake, image-pull-secret + custom-CA propagation,
CRD/schema upgrades, queue/version validation, and one-off migrations
(engine→KAI, resource-management/department CRD). Not a controller.

## Make targets

| Target | What it does |
|--------|-------------|
| build  | `build-go` → builds both amd64 and arm64 binaries into `bin/` |
| build-go-amd | amd64 binary (honors `OS=`/`ARCH=`) → `bin/helm-hooks-handler-amd64` |
| build-go-arm | arm64 binary (hardcodes `GOOS=linux GOARCH=arm64`) → `bin/helm-hooks-handler-arm64` |
| test   | `test-go` (containerized unit tests) |
| test-dev | `go test -v ./... -coverprofile cover.out` on host |
| lint   | `lint-go` |
| format | `format-go` |
| validate | no-op |
| local-deploy | `build-go docker-build-generic docker-push-generic` (amd64) |

Includes: `../../../build/makefile/runai-cluster.mk`, `index.mk`.

## Key build notes

- **runai-cluster Go service.** local-dev.sh (`RC_GO_SERVICES` → `build_go_service`) runs:
  ```bash
  make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
  docker build --build-arg TARGETARCH=amd64 -t runai-local/helm-hooks-handler:0.0.0-dev .
  ```
- **Image tag: `0.0.0-dev`** (RC tag; chart value `helmHooksHandler.image.tag`). Ref: `runai-local/helm-hooks-handler:0.0.0-dev`.
- **Both arches:** amd64 (`build-go-amd`, reads `OS=`/`ARCH=`) and arm64 (`build-go-arm`, hardcodes `GOOS=linux GOARCH=arm64` — do NOT pass `OS=`/`ARCH=`). Same tag; only `--build-arg TARGETARCH` differs. local-dev builds amd64 only.
- **DEBUG mode:** `DEBUG=1` → `-gcflags="all=-N -l"` and `Dockerfile.debug`.
- **Base image:** Dockerfile is `ubi9/ubi-minimal` (for certs/passwd/tmp) → final `FROM scratch`, non-root `USER 65532:65532`, `ENTRYPOINT ["./helm-hooks-handler"]`. Build arg `TARGETARCH` selects `bin/helm-hooks-handler-$TARGETARCH`.
- **Artifact caching:** `artifact-metadata.json` sets `cache_artifacts: true`, `path_to_artifacts: bin/`; `ngc.guest_access: true`.
- **Config:** command-line flags (`internal/installation/util/options.go`) — one boolean per hook + CP/cluster/CA/NMC/migration flags. Also loads a params configmap JSON via `--cm-file-path` (default `/runai/config/params.json`) into `CMConfigParams` (affinity, tolerations, mgmtPrometheusIngress, clusterConfig merge-patch). `--control-plane-client-secret` accepts `@/path` file syntax.
- **Chart wiring:** invoked by `runai-chart/templates/hooks/*-job.yaml` (each hook Job passes the matching flag); RBAC under `templates/hooks/*-rbac.yaml`. Migrations self-gate: `engineToKaiMigration`, `resourceManagementMigration` chart toggles.
- **Deps:** imports KAI-scheduler and the cluster `sdk` APIs; talks to the control plane via `internal/control-plane-details-downloader`.

## local-dev.sh iteration

runai-cluster service (rebuilds the hook image; hooks re-run on chart upgrade):
```bash
./local-dev.sh --skip-build-cp --skip-cluster
```
