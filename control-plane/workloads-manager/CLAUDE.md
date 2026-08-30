# workloads-manager — Build Context

**Type:** container
**Team:** Viper
**Module:** `github.com/run-ai/runai/control-plane/workloads-manager`

## Make targets

Targets come from the service `Makefile` plus the shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, and the `golang.mk` /
`openapi.mk` they pull in).

| Target | What it does |
|--------|-------------|
| `build` | `openapi-build` then `build-go` — builds the OpenAPI bundle + generated code, then compiles the Go binary for amd64 + arm64 |
| `build-go` | `build-go-amd` + `build-go-arm` — both arch binaries |
| `build-go-amd` | Compiles `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) inside the Go builder container → `bin/workloads-manager-amd64` |
| `build-go-arm` | Same for `arm64` → `bin/workloads-manager-arm64` |
| `test` | `test-ginkgo` — runs the Ginkgo unit-test suites (randomized) in the Go container |
| `coverage` | `test-go-acc` — coverage run, ignoring `pkg/clients` and `pkg/kube_client`, into `cover.out` |
| `lint` | `lint-go` — golangci-lint in the builder container |
| `purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `validate` | `openapi-validate` — validate the OpenAPI spec (bundle/code up-to-date + bearer-auth checks) |
| `openapi-lint` | `openapi-lint-generic` — lint the built swagger bundle |
| `local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` (exports OS=linux, ARCH=amd64) |

Note: no `mockery`, `run-local`, `unit-test`, or docker-compose
integration-test targets exist for this service — `test` is Ginkgo-only and
there is no `integration-tests/` directory.

## Key build notes

- Generated code: OpenAPI server stubs live in `internal/generated` (openapi-generator `v7.12.0`, server + go-client pinned in the Makefile). `make build` runs `openapi-build` first; run generation before hand-building and never edit the generated dir.
- Runtime base image: final image is `FROM scratch`; a `registry.access.redhat.com/ubi9/ubi-minimal` stage supplies `/etc/passwd`, `/etc/group`, certs (`/etc/ssl`, `/etc/pki`), `nsswitch.conf`, and a writable `/tmp`. Runs as user `65532:65532`. Bundles SQL migrations (`pkg/database/migrations/up|down`) and the swagger bundle.
- Build args: `TARGETARCH` (selects `bin/workloads-manager-$TARGETARCH`).
- Build command (what local-dev.sh actually runs for this service): it is in `CP_GO_SERVICES`, so `build_go_service` runs, in two steps:
  1. Compile the amd64 binary in the builder container:
     `make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=`
     (`DOCKERHUB_PROXY=` pulls the Go builder image from Docker Hub; `GITHUB_TOKEN=` is emptied in the make call — the private-module token is applied elsewhere.)
  2. Build + load the image: `docker buildx build --platform linux/amd64 --provenance=false --load --build-arg TARGETARCH=amd64 -t runai-local/workloads-manager:0.0.0 .`
     Image ref: `runai-local/workloads-manager:0.0.0` (tag `0.0.0` — the CP chart appVersion; any other tag breaks the local deploy).
- Config: environment-variable driven via the shared config package — typed structs (`Services`, `Server`, shared DB config) with `SetDefaults()`; each key is overridable by its uppercase env-var equivalent (e.g. `CLUSTER_SERVICE_URL`, `SERVICE_PORT`, `DB_HOST`).
- Distinct from the `workloads` service: `workloads-manager` is the v2 submission/orchestration + manifest-generation + NIM/GVK-plugin layer; `workloads` is the workload record/storage service it calls.

## local-dev.sh iteration

This is a control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
