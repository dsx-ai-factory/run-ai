# catalog-service — Build Context

**Type:** container
**Team:** Viper
**Module:** `github.com/run-ai/runai/control-plane/catalog-service`

## Make targets

Targets come from the service `Makefile` plus the shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, and the `golang.mk` /
`openapi.mk` they pull in). The service sets `SERVICE_NAME=catalog-service` and
pins `OPENAPI_GENERATOR_VERSION_SERVER`/`_CLIENT_GO` to `v7.12.0`.

| Target | What it does |
|--------|-------------|
| `build` | `openapi-build` then `build-go` — builds the OpenAPI bundle + generated code, then compiles the Go binary for amd64 + arm64 |
| `build-go` | `build-go-amd` + `build-go-arm` — both arch binaries |
| `build-go-amd` | Compiles `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) inside the Go builder container → `bin/catalog-service-amd64` |
| `build-go-arm` | Same for `arm64` → `bin/catalog-service-arm64` |
| `test` | `test-ginkgo` — runs the Ginkgo unit-test suites |
| `lint` | `lint-go` — golangci-lint in the builder container |
| `purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `validate` | `openapi-validate` — validate the OpenAPI spec (yaml + generated-code up-to-date + public routes + service field) |
| `openapi-lint` | `openapi-lint-generic` — lint the swagger bundle |
| `local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` (exports `OS=linux ARCH=amd64`) |

## Key build notes

- Generated code: OpenAPI server stubs are generated into `internal/generated`
  (openapi-generator `v7.12.0`). `make build` runs `openapi-build` first; run
  OpenAPI generation before hand-building and never edit generated dirs.
- Runtime base image: final image is `FROM scratch`; a
  `registry.access.redhat.com/ubi9/ubi-minimal` stage supplies `/etc/passwd`,
  `/etc/group`, certs (`/etc/ssl`, `/etc/pki`) and a writable `/tmp`. Runs as
  user `65532:65532`. Bundles SQL migrations (`pkg/database/migrations/up|down`)
  and the swagger bundle (`swagger/_build`).
- Build args: `TARGETARCH` (selects `bin/catalog-service-$TARGETARCH`).
- Build command (what local-dev.sh actually runs for this service): it is in
  `CP_GO_SERVICES`, so `build_go_service` runs, in two steps:
  1. Compile the amd64 binary in the builder container:
     `make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=`
     (`DOCKERHUB_PROXY=` pulls the Go builder image from Docker Hub;
     `GITHUB_TOKEN=` is emptied in the make call — the private-module token is
     applied elsewhere.)
  2. Build + load the image:
     `docker buildx build --platform linux/amd64 --provenance=false --load --build-arg TARGETARCH=amd64 -t runai-local/catalog-service:0.0.0 .`
     Image ref: `runai-local/catalog-service:0.0.0` (tag `0.0.0` — the CP chart
     appVersion; any other tag breaks the local deploy).
- Configuration: env-var driven at runtime (ConfigMap/Secret from the CP chart);
  config keys are typed sections (server/services/NGC/database) and each key can
  be overridden by its uppercase env-var equivalent. No sample YAML ships with
  the service.
- No `pkg/client/` and no `integration-tests/` directory in this service.

## local-dev.sh iteration

This is a control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
</content>
