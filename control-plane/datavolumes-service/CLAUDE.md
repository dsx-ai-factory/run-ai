# datavolumes-service — Build Context

**Type:** container
**Team:** Viper
**Module:** `github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service`

## Make targets

Targets come from the service `Makefile` plus the shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, and the `golang.mk` /
`openapi.mk` they pull in).

| Target | What it does |
|--------|-------------|
| `build` | `openapi-build` then `build-go` — build the swagger bundle + generated code, then compile the Go binary for amd64 + arm64 |
| `build-go` | `build-go-amd` + `build-go-arm` — both arch binaries |
| `build-go-amd` | Compiles `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) in the Go builder container → `bin/datavolumes-service-amd64` |
| `build-go-arm` | Same for `arm64` → `bin/datavolumes-service-arm64` |
| `test` | `test-ginkgo` — Ginkgo/Gomega unit suites in the builder container |
| `validate` | `openapi-validate` — validate the OpenAPI spec (yaml up-to-date + generated code + public routes) |
| `openapi-lint` | `openapi-lint-generic` — lint the swagger bundle |
| `lint` | `lint-go` — golangci-lint in the builder container |
| `purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` (forces `OS=linux ARCH=amd64`) |

## Key build notes

- Generated code: OpenAPI server stubs live in `internal/generated` (openapi-generator server `v7.12.0`, pinned in the Makefile). `make build` runs `openapi-build` first; run OpenAPI generation before hand-building and never edit `internal/generated`.
- Runtime base image: final image is `FROM scratch`; a `registry.access.redhat.com/ubi9/ubi-minimal` stage supplies `/etc/passwd`, `/etc/group`, certs (`/etc/ssl`, `/etc/pki`), `nsswitch.conf`, and a writable `/tmp`. Runs as user `65532:65532`. Bundles SQL migrations (`pkg/database/migrations/up|down`) and the swagger bundle.
- Build args: `TARGETARCH` (selects `bin/datavolumes-service-$TARGETARCH`).
- Build command (what local-dev.sh actually runs for this service): it is in `CP_GO_SERVICES`, so `build_go_service` runs, in two steps:
  1. Compile the amd64 binary in the builder container:
     `make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=`
     (`DOCKERHUB_PROXY=` pulls the Go builder image from Docker Hub; `GITHUB_TOKEN=` is emptied in the make call — the private-module token is applied elsewhere.)
  2. Build + load the image: `docker buildx build --platform linux/amd64 --provenance=false --load --build-arg TARGETARCH=amd64 -t runai-local/datavolumes-service:0.0.0 .`
     Image ref: `runai-local/datavolumes-service:0.0.0` (tag `0.0.0` — the CP chart appVersion; any other tag breaks the local deploy).
- Config: env-var driven (typed structs with in-code defaults; keys overridable by their uppercase env-var equivalent). Key vars: `SERVICE_PORT`, `BACKEND_URL`, `CLUSTER_SERVICE_URL`, `ORG_UNIT_SERVICE_URL`, `AUTH_SERVICE_URL`, plus shared `LOG_LEVEL` / `DATABASE_*`.
- Depends on: `temp-backend-settings` (settings), `cluster-service` (capabilities), `org-unit-service` (hierarchy), authorization service, audit-service client, and NATS pub/sub.
- Tests: Ginkgo unit suites only; no docker-compose integration suite in this service.

## local-dev.sh iteration

This is a control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
