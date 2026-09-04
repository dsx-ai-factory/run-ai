# agent — Build Context

**Type:** container (runai-cluster Go service)
**Team:** Suricata Team (`artifact-metadata.json` → `jira.team_name`; Slack `@suricata`)
**Module:** `github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent` (Go 1.26)

## What it is

Applies Run.ai backend state INTO the cluster: pluggable resource handlers
(departments, projects, access rules, data volumes, priority classes, workloads
of all kinds, assets, network topologies, applications, KAI node pools) reconcile
backend state on an interval. Also hosts a TLS validating-webhook server
(`/validate-project`, `/validate-secrets-pvc`). Inbound half of the cluster↔backend
sync; outbound is `cluster-sync`.

## Make targets

| Target | What it does |
|--------|-------------|
| build  | `build-go` → builds both amd64 and arm64 binaries into `bin/` |
| build-go-amd | amd64 binary (honors `OS=`/`ARCH=`) → `bin/agent-amd64` |
| build-go-arm | arm64 binary (hardcodes `GOOS=linux GOARCH=arm64`) → `bin/agent-arm64` |
| test   | `test-go` (containerized unit tests, Ginkgo) |
| test-dev | `go test -v ./... -coverprofile cover.out` on host |
| lint   | `lint-go` |
| format | `format-go` |
| validate | no-op |
| local-deploy | `build-go docker-build-generic docker-push-generic` (amd64) |
| generate | Regenerate all OpenAPI clients (`generate-projects/backend/authorization/datavolumes`) |
| generate-{projects,backend,authorization,datavolumes} | Regenerate one client from its `swagger/` spec |

Includes: `../../../build/makefile/runai-cluster.mk`, `index.mk`.

## Key build notes

- **runai-cluster Go service.** local-dev.sh (`RC_GO_SERVICES` → `build_go_service`) runs:
  ```bash
  make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
  docker build --build-arg TARGETARCH=amd64 -t runai-local/agent:0.0.0-dev .
  ```
- **Image tag: `0.0.0-dev`** (RC tag, hardcoded in the runai-cluster chart values). Ref: `runai-local/agent:0.0.0-dev`.
- **Both arches:** amd64 (`build-go-amd`, reads `OS=`/`ARCH=`) and arm64 (`build-go-arm`, hardcodes `GOOS=linux GOARCH=arm64` — do NOT pass `OS=`/`ARCH=`). Same tag; only `--build-arg TARGETARCH` differs. local-dev builds amd64 only.
- **Generated code:** `pkg/generated/{projects,backend,auth,datavolumes}_openapi` are generated from `swagger/*.yaml` via `openapi-generator-cli` v6.0.1 (`onlyInterfaces=true`, `outputAsLibrary=true`). Run `make generate` (or a single `generate-*`) after updating a spec. Generated `go.mod`/`go.sum` are stripped so packages build within the module. Do not edit generated dirs by hand.
- **Base image:** Dockerfile is `ubi9/ubi-minimal` (for certs/passwd/tmp) → final `FROM scratch`, non-root `USER 65532:65532`, `ENTRYPOINT ["./agent"]`. Build arg `TARGETARCH` selects `bin/agent-$TARGETARCH`.
- **Artifact caching:** `artifact-metadata.json` sets `cache_artifacts: true`, `path_to_artifacts: bin/`.
- **Config:** env-var driven via `runai-common-packages/config` (uppercase `mapstructure` tags in `pkg/config/config.go`); secrets (`CLUSTER_ID`, `CLIENT_NAME`, `CLIENT_SECRET`) from the mounted secret file `/etc/secret/secret.json` (`constants.SecretMountPath`). `CLUSTER_VERSION` also read via `os.Getenv` in `main.go`. Defaults in `SetDefaults()`.
- **Webhook:** TLS validating-webhook server on `WEBHOOK_PORT` (9443), cert from `CERT_DIR` (`/etc/webhook/certs`, provisioned by init-ca); paths `/validate-project`, `/validate-secrets-pvc`.
- **Adding a handler:** implement `ObjectInspector` + `ResourceHandler` in `pkg/handlers`, register in `handlerInitFuncs` in `pkg/runai_agent.go`, add the API version to the scheme in `cmd/main.go` + `pkg/tests/suite_test.go`.

## local-dev.sh iteration

runai-cluster service:
```bash
./local-dev.sh --skip-build-cp --skip-cluster
```
