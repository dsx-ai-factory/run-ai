# audit-service — Build Context

**Type:** container
**Team:** Phoenix
**Module:** `github.com/dsx-ai-factory/run-ai/control-plane/audit-service`

## Make targets

Targets come from the service `Makefile` plus the shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, and the `golang.mk` /
`openapi.mk` they pull in).

| Target | What it does |
|--------|-------------|
| `build` | `openapi-build` then `build-go` — build the OpenAPI bundle + generate code, then compile the Go binary for amd64 + arm64 |
| `build-go` | `build-go-amd` + `build-go-arm` — both arch binaries |
| `build-go-amd` | Compiles `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) in the Go builder container → `bin/audit-service-amd64` |
| `build-go-arm` | Same for `arm64` → `bin/audit-service-arm64` |
| `test` | `build` + `test-go` + `integration-tests` |
| `test-go` | Runs the Go unit tests in the builder container |
| `build-integration-tests` | Compiles the `integration`-tagged test binary and `docker compose build`s the test images |
| `integration-tests` | Brings up the compose stack and runs the suite (abort-on-container-exit), tearing volumes down first |
| `integration-tests-debug` | Compose stack without the tests container (run tests from an IDE) |
| `generate-mocks` / `generate-mocks-local` / `generate-mocks-client` | Shared targets: regenerate mocks with mockery v3 from `.mockery.yml` / `pkg/client/.mockery.yml` (`-local` runs without Docker) |
| `lint` | `lint-go` — golangci-lint in the builder container |
| `purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `validate` | `openapi-validate` — validate the OpenAPI spec |
| `openapi-lint` | Lint the built swagger bundle |

## Key build notes

- Generated code: OpenAPI server stubs (`internal/generated`) + Go client (`pkg/client/generated`) are generated (openapi-generator `v7.12.0`, `VERSIONED_CLIENT=true`). `make build` runs `openapi-build` first; run generation before hand-building and never edit generated dirs.
- Runtime base image: final image is `FROM scratch`; a `registry.access.redhat.com/ubi9/ubi-minimal` stage supplies `/etc/passwd`, `/etc/group`, certs (`/etc/ssl`, `/etc/pki`) and a writable `/tmp`. Runs as user `65532:65532`. Bundles SQL migrations (`internal/database/migrations/up|down`), the swagger bundle, and `NOTICE`.
- Build args: `TARGETARCH` (selects `bin/audit-service-$TARGETARCH`).
- Build command (what local-dev.sh actually runs for this service): it is in `CP_GO_SERVICES`, so `build_go_service` runs, in two steps:
  1. Compile the amd64 binary in the builder container:
     `make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=`
     (`DOCKERHUB_PROXY=` pulls the Go builder image from Docker Hub; `GITHUB_TOKEN=` is emptied in the make call — the private-module token is applied elsewhere.)
  2. Build + load the image: `docker buildx build --platform linux/amd64 --provenance=false --load --build-arg TARGETARCH=amd64 -t runai-local/audit-service:0.0.0 .`
     Image ref: `runai-local/audit-service:0.0.0` (tag `0.0.0` — the CP chart appVersion; any other tag breaks the local deploy).
- Config model: environment-variable driven. Typed config blocks in `internal/config` set defaults; keys are overridden by their uppercase env-var equivalents. DB/pub-sub/HTTP/logging come from shared Run:ai packages; DB creds are read from files under `SECRETS_MOUNT_PATH`.
- Integration tests: require docker-compose + backing services (PostgreSQL + a Redis-backed pub/sub).

## local-dev.sh iteration

This is a control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
