# temp-backend-settings — Build Context

**Type:** container
**Team:** Suricata Team
**Module:** `github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings`

Transitional HTTP service serving legacy tenant settings/feature-flags migrated
from the TypeScript backend's `SettingController`/`InternalTenantController`.
Long-running service (not a one-shot job); `bff-service` proxies to it.

## Make targets

Targets come from the service `Makefile` plus the shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, and the `golang.mk` /
`openapi.mk` they pull in).

| Target | What it does |
|--------|-------------|
| `build` | `openapi-build` then `build-go` — bundle+generate OpenAPI, then compile amd64 + arm64 binaries |
| `build-go` | `build-go-amd` + `build-go-arm` |
| `build-go-amd` | Compiles `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) in the builder container → `bin/temp-backend-settings-amd64` |
| `build-go-arm` | Same for `arm64` → `bin/temp-backend-settings-arm64` |
| `test` | `build-go` + `test-go` — compile then run Go unit tests in the builder container |
| `generate-mocks` / `generate-mocks-local` | Regenerate test mocks from `.mockery.yml` with the shared mockery v3 setup (in Docker / without Docker) |
| `generate-mocks-client` | Regenerate `pkg/client` mocks from `pkg/client/.mockery.yml` |
| `lint` | `lint-go` — golangci-lint in the builder container |
| `purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `validate` | `openapi-validate` — validate spec + generated code up to date |
| `openapi-lint` | `openapi-lint-generic` — lint the swagger bundle |
| `openapi-build` | Bundle spec (yaml/json/html), copy central spec, build public spec, generate code |
| `local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` |

## Key build notes

- Generated code: OpenAPI server stubs (`internal/generated`) + Go client (`pkg/client/generated`) via openapi-generator `v7.12.0`. `make build` runs `openapi-build` first; run generation before hand-building and never edit generated dirs. Test mocks via `make generate-mocks` or `make generate-mocks-local` (`.mockery.yml`), and `make generate-mocks-client` for `pkg/client` (`pkg/client/.mockery.yml`).
- Runtime base image: final image is `FROM scratch`; a `registry.access.redhat.com/ubi9/ubi-minimal` stage supplies `/etc/passwd`, `/etc/group`, certs (`/etc/ssl`, `/etc/pki`) and a writable `/tmp`. Runs as user `65532:65532`. Bundles SQL migrations (`pkg/database/migrations/up|down`) and the swagger bundle.
- Build args: `TARGETARCH` (selects `bin/temp-backend-settings-$TARGETARCH`).
- Build command (what local-dev.sh actually runs for this service): it is in `CP_GO_SERVICES`, so `build_go_service` runs, in two steps:
  1. Compile the amd64 binary in the builder container:
     `make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=`
     (`DOCKERHUB_PROXY=` pulls the Go builder image from Docker Hub; `GITHUB_TOKEN=` is emptied in the make call — the private-module token is applied elsewhere.)
  2. Build + load the image: `docker buildx build --platform linux/amd64 --provenance=false --load --build-arg TARGETARCH=amd64 -t runai-local/temp-backend-settings:0.0.0 .`
     Image ref: `runai-local/temp-backend-settings:0.0.0` (tag `0.0.0` — the CP chart appVersion; any other tag breaks the local deploy).
- Configuration: typed `ServicesConfig` (peer-service URLs + `show_alpha_flags`) with in-cluster defaults; any key overridable by its uppercase env-var equivalent. DB via shared package `DATABASE_*` env vars.
- No docker-compose integration suite in this service.

## local-dev.sh iteration

This is a control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
