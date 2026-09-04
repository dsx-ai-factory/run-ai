# policy-service — Build Context

**Type:** container (control-plane Go service)
**Team:** Viper Team (`artifact-metadata.json` → `jira.team_name`; Slack `@viper`)
**Module:** `github.com/dsx-ai-factory/run-ai/control-plane/policy-service` (Go 1.26)

## What it is

Owns Run.ai policies: stores policies per scope (tenant/department/project),
computes the effective policy by merging the hierarchy tree, validates workloads
for compliance, and syncs effective policies to clusters. REST API (chi,
OpenAPI-defined) over PostgreSQL. Applies its own DB migrations on startup.

## Make targets

| Target | What it does |
|--------|-------------|
| build  | `openapi-build` (bundle spec + generate server/go-client/js-client) then `build-go` (both arches → `bin/`) |
| test   | `test-ginkgo` (Ginkgo suites; DB suites use throwaway Postgres) |
| coverage | `test-go-acc` (ignores `clients/assets`) |
| lint   | `lint-go` |
| openapi-lint | `openapi-lint-generic` |
| validate | `openapi-validate` (yaml + generated-code-up-to-date + public routes + service-field) |
| purify | `purify-go` |
| local-deploy | `openapi-generate-code build-go docker-build-generic docker-push-generic` |
| copy-to-backend-js | Legacy: copy built swagger into backend-js tree |

Includes: `../../build/makefile/control-plane.mk`, `index.mk`.

## Key build notes

- **Control-plane Go service.** local-dev.sh (`CP_GO_SERVICES` → `build_go_service`) runs:
  ```bash
  make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
  docker build --build-arg TARGETARCH=amd64 -t runai-local/policy-service:0.0.0 .
  ```
- **Image tag: `0.0.0`** (CP tag, must match chart `appVersion`). Ref: `runai-local/policy-service:0.0.0`.
- **Both arches:** amd64 (`build-go-amd`, reads `OS=`/`ARCH=`) and arm64 (`build-go-arm`, hardcodes `GOOS=linux GOARCH=arm64` — no `OS=`/`ARCH=`). Same tag; only `--build-arg TARGETARCH` differs. local-dev builds amd64 only.
- **Generated code (run before build if spec changed):** OpenAPI server + Go/JS clients under `generated/{api,pkg}`, from `swagger/source/openapi.yaml` via `openapi-generator` v7.12.0. The service's `make build` (`openapi-build build-go`) regenerates first; the shared `build-go-amd` that local-dev calls does NOT regenerate. Do not edit `generated/` by hand. `make validate` fails if generated code/public routes are stale.
- **Base image:** Dockerfile is `ubi9/ubi-minimal` → final `FROM scratch`, non-root `USER 65532:65532`, `ENTRYPOINT ["./policy-service"]`. Build arg `TARGETARCH`. Image also ships `pkg/database/migrations/{up,down}` and `swagger/_build/*` — migrations run on startup.
- **Artifact caching:** `artifact-metadata.json` sets `cache_artifacts: true`, `path_to_artifacts: bin/`.
- **Config:** env-var driven. App settings from `ServerConfig`/`ServicesConfig` in `pkg/config/config.go` (env-var name = uppercase of the field key); DB from shared `dao-utils` `DB_*`. Key vars: `SERVICE_PORT` (8080), `LOG_LEVEL` (info), `ASSETS_SERVICE_URL`/`CLUSTER_SERVICE_URL`/`ORG_UNIT_SERVICE_URL`, `DB_HOST/NAME/SCHEMA/USER/PASSWORD/SSL_MODE`, `MIGRATION_FILES`, `PUBSUB_ENABLED`, `SECRETS_MOUNT_PATH`. `config.MockConfig()` wires test env. (Old README's `AUTH_SERVICE_URL`/`AUTHORIZATION_URL`/`BACKEND_SERVICE_URL`/`DB_SSL_ENABLED` are obsolete.)
- **Go client:** no in-tree `pkg/client`; the generated Go client module is `runai-clients/policy-service-client`.

## local-dev.sh iteration

Control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
