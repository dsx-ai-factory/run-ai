# assets-service — Build Context

**Type:** container
**Team:** Viper (`@viper`)
**Module:** `github.com/dsx-ai-factory/run-ai/control-plane/assets-service`

## Make targets

Targets come from the service `Makefile` plus the shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, and the `golang.mk` /
`openapi.mk` they pull in). OpenAPI generator is pinned to `v7.12.0`.

| Target | What it does |
|--------|-------------|
| `build` | `openapi-build` then `build-go` — builds the OpenAPI bundle + generated code, then compiles the Go binary for amd64 + arm64 |
| `build-go` | `build-go-amd` + `build-go-arm` |
| `build-go-amd` | Compiles `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) inside the Go builder container → `bin/assets-service-amd64` |
| `build-go-arm` | Same for `arm64` → `bin/assets-service-arm64` |
| `test` | `test-ginkgo` — runs Ginkgo suites recursively (`--skip-package=third_party`, 15m timeout) |
| `coverage` | `test-go-acc` — go-acc coverage report (ignores `pkg/clients,pkg/kube_client,cmd`) |
| `lint` | `lint-go` — golangci-lint in the builder container |
| `purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `validate` | `openapi-validate` — validate the OpenAPI spec (yaml + generated code up-to-date + public routes + `x-service` field) |
| `openapi-lint` | `openapi-lint-generic` — lint the built swagger bundle |
| `local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` (forces `OS=linux ARCH=amd64`) |

Note: `SERVICE_NAME=assets-service` and `OPENAPI_GENERATOR_SERVER_CODE_PATH=generated`
(server stubs land in `generated/`, not `internal/generated`).

## Key build notes

- Generated code: OpenAPI server stubs go to `generated/`; the Go client is
  generated into the separate `runai-clients/assets-service-client` module.
  `make build` runs `openapi-build` first; run generation before hand-building
  and never edit generated dirs.
- Runtime base image: final image is `FROM scratch`; a
  `registry.access.redhat.com/ubi9/ubi-minimal` stage supplies `/etc/passwd`,
  `/etc/group`, certs (`/etc/ssl`, `/etc/pki`), `nsswitch.conf` and a writable
  `/tmp`. Runs as user `65532:65532`. Bundles SQL migrations
  (`pkg/database/migrations/up|down`), the `NOTICE` file and the swagger bundle.
- Build args: `TARGETARCH` (selects `bin/assets-service-$TARGETARCH`).
- Cross-repo: pulls private Go modules from `github.com/run-ai/*` (workloads,
  runai-operator, etc.) — needs `GITHUB_TOKEN`/git insteadOf for private access.
- Build command (what local-dev.sh actually runs for this service): it is in
  `CP_GO_SERVICES`, so `build_go_service` runs, in two steps:
  1. Compile the amd64 binary in the builder container:
     `make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=`
     (`DOCKERHUB_PROXY=` pulls the Go builder image from Docker Hub;
     `GITHUB_TOKEN=` is emptied in the make call — the private-module token is
     applied elsewhere.)
  2. Build + load the image:
     `docker buildx build --platform linux/amd64 --provenance=false --load --build-arg TARGETARCH=amd64 -t runai-local/assets-service:0.0.0 .`
     Image ref: `runai-local/assets-service:0.0.0` (tag `0.0.0` — the CP chart
     appVersion; any other tag breaks the local deploy).
- Config: environment-variable driven (`pkg/config/config.go`); struct keys map
  to uppercase env vars. DB settings come from `DB_*` / `MIGRATION_FILES`.

## local-dev.sh iteration

This is a control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
</content>
