# tenants-manager — Build Context

**Type:** container
**Team:** Sphinx
**Module:** `github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager`

## Make targets

Targets come from the service `Makefile` plus the shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, and the `golang.mk` /
`openapi.mk` they pull in).

| Target | What it does |
|--------|-------------|
| `build` | `openapi-build` then `build-go` — build the OpenAPI bundle, then compile the Go binary for amd64 + arm64 |
| `build-go` | `build-go-amd` + `build-go-arm` — both arch binaries (FIPS `GOFIPS140=v1.0.0`) → `bin/tenants-manager-<arch>` |
| `build-go-amd` | Compiles the service for `GOOS=linux GOARCH=amd64` in the Go builder container → `bin/tenants-manager-amd64` |
| `build-go-arm` | Same for `arm64` → `bin/tenants-manager-arm64` |
| `unit-test` | `build` + `test-ginkgo` — compile then run unit tests (Ginkgo, skipping `integration-tests`) |
| `test` | `build` + `test-ginkgo` + `docker-compose-integration-tests-up` |
| `integ-tests` | `build-go` + `docker-compose-integration-tests-up` |
| `run-local` | `build-go` then bring up the `localdebug` docker-compose profile |
| `build-integration-tests` | Build `integration-tests/tests` as a test binary + `docker compose build` |
| `docker-compose-integration-tests-up` | Run the full compose integration suite (abort-on-container-exit) |
| `docker-compose-integration-tests-debug` | Same suite without the failure handler |
| `lint` | `lint-go` — golangci-lint in the builder container |
| `purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `validate` | `openapi-validate` — validate the OpenAPI spec |
| `openapi-lint` | Lint the swagger bundle |

## Key build notes

- Generated code: OpenAPI server stubs in `generated/`
  (openapi-generator `v7.12.0`, `VERSIONED_CLIENT=true`) plus mockery mocks
  (`.mockery.yml` → `internal/mocks`). `make build` runs `openapi-build` first;
  run generation before hand-building and never edit generated dirs.
- Runtime base image: final image is `FROM scratch`; a `ubi9/ubi-minimal` stage
  supplies `/etc/passwd`, certs (`/etc/ssl`, `/etc/pki`) and a writable `/tmp`.
  Runs as `65532:65532`. Bundles SQL migrations
  (`pkg/database/migrations/up|down`) and the swagger bundle.
- Build args: `TARGETARCH` (selects `bin/tenants-manager-$TARGETARCH`).
- Build command (what local-dev.sh actually runs): it is in `CP_GO_SERVICES`, so
  the generic `build_go_service` runs, in two steps:
  1. Compile the amd64 binary: `make build-go-amd OS=linux ARCH=amd64
     DOCKERHUB_PROXY= GITHUB_TOKEN=` (`DOCKERHUB_PROXY=` pulls the Go builder
     from Docker Hub; `GITHUB_TOKEN=` emptied — token applied elsewhere).
  2. Build + load: `docker buildx build --platform linux/amd64
     --provenance=false --load --build-arg TARGETARCH=amd64
     -t runai-local/tenants-manager:0.0.0 .`
     Image ref: `runai-local/tenants-manager:0.0.0` (tag `0.0.0` — the CP chart
     appVersion; any other tag breaks the local deploy).
- Integration tests: require docker-compose + backing services (Postgres, NATS,
  identity/BFF mocks).

## local-dev.sh iteration

This is a control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
</content>
