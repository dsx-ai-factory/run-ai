# identity-manager — Build Context

**Type:** container
**Team:** Sphinx
**Module:** `github.com/dsx-ai-factory/run-ai/control-plane/identity-manager`

## Make targets

Targets come from the service `Makefile` plus the shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, and the `golang.mk` /
`openapi.mk` they pull in).

| Target | What it does |
|--------|-------------|
| `build` | `openapi-build` then `build-go` — builds the OpenAPI bundle, then compiles the Go binary for amd64 + arm64 |
| `build-go` | `build-go-amd` + `build-go-arm` — both arch binaries |
| `build-go-amd` | Compiles `./cmd` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) inside the Go builder container → `bin/identity-manager-amd64` |
| `build-go-arm` | Same for `arm64` → `bin/identity-manager-arm64` |
| `unit-test` | `build` + `test-go` — compile then run unit tests |
| `test` | `build-go` + `test-go` + docker-compose integration tests (tears the compose stack down afterwards) |
| `test-go` | Runs the Go unit tests in the builder container |
| `run-local` | `build-go` then brings up the `localdebug` docker-compose profile |
| `build-integration-tests` | Builds `integration-tests/tests` as a test binary + `docker compose build` (depends on `build-keycloak`) |
| `build-keycloak` | Builds the sibling `../keycloak-extensions` image (integration-test dependency) |
| `docker-compose-integration-tests-up` | Runs the full compose integration suite, exiting on the `identity-manager-tests` container |
| `docker-compose-integration-tests-debug` | Same suite without the failure handler |
| `coverage` | Opens the HTML coverage report (`go tool cover -html=c.out`) |
| `lint` | `lint-go` — golangci-lint in the builder container |
| `purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `validate` | `openapi-validate` — validate the OpenAPI spec |
| `openapi-build` / `openapi-lint` | Build / lint the swagger bundle |
| `local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` |

## Key build notes

- Generated code: OpenAPI server stubs + Go client are generated (openapi-generator `v7.12.0`, `VERSIONED_CLIENT=true`). `make build` runs `openapi-build` first; run OpenAPI generation before hand-building and never edit generated dirs.
- Runtime base image: final image is `FROM scratch`; a `registry.access.redhat.com/ubi9/ubi-minimal` stage supplies `/etc/passwd`, certs (`/etc/ssl`, `/etc/pki`) and a writable `/tmp`. Runs as user `65532:65532`. Bundles SQL migrations (`pkg/database/migrations/up|down`) and the swagger bundle.
- Build args: `TARGETARCH` (selects `bin/identity-manager-$TARGETARCH`).
- Build command (what local-dev.sh actually runs for this service): it is in `CP_GO_SERVICES`, so `build_go_service` runs, in two steps:
  1. Compile the amd64 binary in the builder container:
     `make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=`
     (`DOCKERHUB_PROXY=` pulls the Go builder image from Docker Hub; `GITHUB_TOKEN=` is emptied in the make call — the private-module token is applied elsewhere.)
  2. Build + load the image: `docker buildx build --platform linux/amd64 --provenance=false --load --build-arg TARGETARCH=amd64 -t runai-local/identity-manager:0.0.0 .`
     Image ref: `runai-local/identity-manager:0.0.0` (tag `0.0.0` — the CP chart appVersion; any other tag breaks the local deploy).
- Integration tests: require docker-compose + backing services (and build the sibling keycloak-extensions image).

## local-dev.sh iteration

This is a control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
