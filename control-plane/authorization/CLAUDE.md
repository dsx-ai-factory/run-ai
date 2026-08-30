# authorization — Build Context

**Type:** container
**Team:** Suricata
**Module:** `github.com/run-ai/runai/control-plane/authorization`

## Make targets

Targets come from the service `Makefile` plus the shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, and the `golang.mk` /
`openapi.mk` they pull in).

| Target | What it does |
|--------|-------------|
| `build` | `openapi-build` then `build-go` — builds the OpenAPI bundle + generated code, then compiles the Go binary for amd64 + arm64 |
| `build-go` | `build-go-amd` + `build-go-arm` — both arch binaries |
| `build-go-amd` | Compiles `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) inside the Go builder container → `bin/authorization-amd64` |
| `build-go-arm` | Same for `arm64` → `bin/authorization-arm64` |
| `test` | `test-ginkgo` — runs the Ginkgo unit/spec suites in the builder container |
| `lint` | `lint-go` — golangci-lint in the builder container |
| `purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `validate` | `openapi-validate` — validate the OpenAPI spec (yaml + generated code up-to-date + public routes + x-service field) |
| `openapi-build` | Bundle spec, emit JSON/HTML, copy central spec, build public spec, generate server + client code |
| `openapi-lint` | `openapi-lint-generic` — lint the bundled swagger |
| `docker-build-generic` / `docker-push-generic` | Build / push the container image |
| `local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` (OS=linux, ARCH=amd64) |
| `client-bump-version` / `client-bump-version-dev` | Tag a new / dev version of the Go client (from `versioning.mk`) |

Note: this service has **no** `unit-test`, `run-local`, `coverage`, or
docker-compose integration targets — `make test` runs Ginkgo directly and there
is no `integration-tests/` directory.

## Key build notes

- Generated code: OpenAPI server stubs live in `generated/`
  (`OPENAPI_GENERATOR_SERVER_CODE_PATH=generated`) and the Go client in
  `pkg/client/generated` (openapi-generator `v7.12.0`, `VERSIONED_CLIENT=true`,
  `OPENAPI_CLIENT_NAME_GO=client`). `make build` runs `openapi-build` first; run
  OpenAPI generation before hand-building and never edit generated dirs.
- Runtime base image: final image is `FROM scratch`; a
  `registry.access.redhat.com/ubi9/ubi-minimal` stage supplies `/etc/passwd`,
  `/etc/group`, certs (`/etc/ssl`, `/etc/pki`) and a writable `/tmp`. Runs as
  user `65532:65532`. Bundles SQL migrations
  (`pkg/database/migrations/up|down`), the swagger bundle, and `NOTICE`.
- Build args: `TARGETARCH` (selects `bin/authorization-$TARGETARCH`).
- Config model: config-object driven (`General`, `Server`, `SyncJob` structs
  with defaults; any key overridable by its uppercase env-var equivalent). DB
  via `DB_*` env vars; schema `authorization`.
- Build command (what local-dev.sh actually runs for this service): it is in
  `CP_GO_SERVICES`, so `build_go_service` runs, in two steps:
  1. Compile the amd64 binary in the builder container:
     `make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=`
     (`DOCKERHUB_PROXY=` pulls the Go builder image from Docker Hub;
     `GITHUB_TOKEN=` is emptied in the make call — the private-module token is
     applied elsewhere.)
  2. Build + load the image:
     `docker buildx build --platform linux/amd64 --provenance=false --load --build-arg TARGETARCH=amd64 -t runai-local/authorization:0.0.0 .`
     Image ref: `runai-local/authorization:0.0.0` (tag `0.0.0` — the CP chart
     appVersion; any other tag breaks the local deploy).

## local-dev.sh iteration

This is a control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
