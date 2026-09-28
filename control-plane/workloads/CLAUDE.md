# workloads — Build Context

**Type:** container
**Team:** Suricata
**Module:** `github.com/dsx-ai-factory/run-ai/control-plane/workloads`

## Make targets

Targets come from the service `Makefile` plus the shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, and the `golang.mk` /
`openapi.mk` they pull in).

| Target | What it does |
|--------|-------------|
| `build` | `openapi-build` then `build-go` — builds the OpenAPI bundle + generated code, then compiles the Go binary for amd64 + arm64 |
| `build-go` | `build-go-amd` + `build-go-arm` — both arch binaries |
| `build-go-amd` | Compiles `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) in the Go builder container → `bin/workloads-amd64` |
| `build-go-arm` | Same for `arm64` → `bin/workloads-arm64` |
| `test` | `test-ginkgo` — runs the Ginkgo unit-test suite |
| `coverage` | `test-go-acc` — go-acc coverage (ignores `pkg/clients`, `pkg/kube_client`) |
| `lint` | `lint-go` — golangci-lint in the builder container |
| `purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `validate` | `openapi-validate` — validate the OpenAPI spec (yaml + generated code + public-route checks) |
| `openapi-lint` | `openapi-lint-generic` — lint the swagger bundle |
| `local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` (with `OS=linux ARCH=amd64`) |

## Key build notes

- Generated code: OpenAPI server stubs are generated (openapi-generator `v7.12.0`) into `internal/generated`. `make build` runs `openapi-build` first; run OpenAPI generation before hand-building and never edit generated dirs. The public Go client is a separate module in `pkg/client`, generated from the same spec.
- Runtime base image: final image is `FROM scratch`; a `registry.access.redhat.com/ubi9/ubi-minimal` stage supplies `/etc/passwd`, certs (`/etc/ssl`, `/etc/pki`) and a writable `/tmp`. Runs as user `65532:65532`. Bundles SQL migrations (`pkg/database/migrations/up|down`) and the swagger bundle.
- Build args: `TARGETARCH` (selects `bin/workloads-$TARGETARCH`).
- Dual run mode: one binary; `job-mode` arg selects the background "workloads-helper" worker (cleaner/watchdog/notify/metrics loops) instead of the HTTP server.
- Build command (what local-dev.sh actually runs for this service): it is in `CP_GO_SERVICES`, so `build_go_service` runs, in two steps:
  1. Compile the amd64 binary in the builder container:
     `make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=`
     (`DOCKERHUB_PROXY=` pulls the Go builder image from Docker Hub; `GITHUB_TOKEN=` is emptied in the make call — the private-module token is applied elsewhere.)
  2. Build + load the image: `docker buildx build --platform linux/amd64 --provenance=false --load --build-arg TARGETARCH=amd64 -t runai-local/workloads:0.0.0 .`
     Image ref: `runai-local/workloads:0.0.0` (tag `0.0.0` — the CP chart appVersion; any other tag breaks the local deploy).
- Configuration: env-var driven. Config structs (`pkg/config/config.go`) carry defaults; the chart ConfigMap/Secret project every value as an uppercase env var that overrides the corresponding config key. No YAML config file is mounted.
- Tests: Ginkgo unit tests only (no docker-compose integration suite).

## local-dev.sh iteration

This is a control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
