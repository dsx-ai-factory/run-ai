# cli-exposer

The **cli-exposer** is the control-plane service that distributes the Run:ai
CLI (`runai`) to users. It bakes the cross-platform `runai` binaries, their
generated docs, and version metadata into its own container image and serves
them (plus a generated install script) over HTTP, so clusters can hand out the
correct CLI build for their control-plane version. As of
it is the single
distribution point for the CLI — the CLI is no longer published as a standalone
artifact. Owned by the **Viper Team**.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [OpenAPI and code generation](#openapi-and-code-generation)

## Overview

The service is responsible for:

- **CLI distribution** — serving the `runai` binaries for every supported
  `{os}/{arch}` (baked into the image at `/app/dist/{os}/{arch}/runai`), e.g.
  `GET /api/v1/cli-exposer/dist/{os}/{arch}/runai`.
- **Install script** — rendering a platform install script from
  `installer/unix/installer_template.sh`.
- **CLI docs & version** — exposing the generated CLI command docs and the
  current CLI version.

It embeds the sibling [`runai-cli`](../runai-cli) build artifacts at image-build
time; see that service for the CLI itself.

### Authorization

The service assumes authentication/authorization is performed upstream (by the
gateway) and passed as request headers:

| Header | Purpose |
|--------|---------|
| `X-Runai-Subject-Id` | The user or app making the request |
| `X-Runai-Tenant-Id` | The tenant ID of the requester |

## Repository layout

```text
.
├── cmd/                        # Service entry point (main.go)
├── pkg/
│   ├── config/                 # Configuration (env-overridable)
│   ├── consts/
│   ├── document/               # CLI docs serving
│   ├── installer/              # Install-script rendering
│   ├── router/                 # HTTP router wiring
│   ├── services/               # API service implementations
│   └── version/                # Version metadata (ldflags-injected)
├── internal/
│   ├── generated/              # OpenAPI-generated server stubs (do not edit)
│   └── static/                 # Static assets
├── installer/                  # Install-script templates (unix/, windows/)
├── swagger/                    # OpenAPI source and built bundle
├── bin/                        # Build output: server binary + embedded CLI dist/docs/version
├── Dockerfile
├── Makefile
└── go.mod
```

## Building and running

The project is built with Go (see [`go.mod`](go.mod)) and orchestrated through
the [`Makefile`](Makefile), which includes the shared `control-plane.mk` /
`index.mk` (and the `golang.mk` / `openapi.mk` they pull in).

| Target | Description |
|--------|-------------|
| `make build` | `openapi-build` + `build-cli` + `build-go` — build the OpenAPI bundle, build all CLI platform binaries (`make -C ../runai-cli dist`) and copy them + docs + version into `bin/`, then compile the server for amd64 + arm64 |
| `make build-cli` | Cleans `bin/`, runs `make -C ../runai-cli dist`, and copies the CLI `dist/`, `docs/`, and `version` into `bin/` |
| `make build-go` | `build-go-amd` + `build-go-arm` — compile the server binary for both arches (FIPS `GOFIPS140=v1.0.0`) → `bin/cli-exposer-<arch>` |
| `make run-local` | Build CLI artifacts, `go mod vendor`, then build & run the server locally with dev ldflags |
| `make lint` | `lint-go` — golangci-lint in the builder container |
| `make purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `make validate` | `openapi-validate` — validate the OpenAPI spec |
| `make openapi-lint` | Lint the swagger bundle |
| `make local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` (OS=linux ARCH=amd64) |
| `make coverage` | Run tests with coverage, filtering generated code, and print the report |

### Runtime image (Dockerfile)

Multi-stage: a `registry.access.redhat.com/ubi9/ubi-minimal` `base` stage
stages `/etc/passwd`, certs (`/etc/ssl`, `/etc/pki`) and a writable `/tmp`; the
final image is `FROM scratch`. It copies in `bin/cli-exposer-$TARGETARCH`, the
embedded CLI `dist/`, `docs/`, `version`, the unix installer template, the
swagger bundle, and `internal/static`. Runs as `65532:65532`, exposes ports
**8080** and **9090**, entrypoint `./cli-exposer`. Build arg: `TARGETARCH`.

### Local build (local-dev.sh)

`local-dev.sh` builds this service through a **dedicated `build_cli_exposer`
function** (not the generic `build_go_service`), because it must also compile
the CLI so host developers don't need Go installed. Its hash is computed over
**both** `cli-exposer/` and `runai-cli/`. The steps are:

```bash
# Step 1 — compile the cli-exposer server binary (amd64) in the Go builder:
make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=

# Step 2 — compile the runai-cli binary for linux/amd64 inside golang:1.26.3
#          (only linux/amd64 is needed for the Kind cluster; full `make dist`
#           across all platforms is reserved for release pipelines):
docker run --rm -w /local/control-plane/runai-cli -v <repo>:/local \
  -e GOFIPS140=v1.0.0 -e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=0 \
  library/golang:1.26.3 \
  go build -ldflags "-X .../pkg/version.version=<ver>-local -extldflags -static -s -w" \
  -o bin/dist/linux/amd64/runai .

# Step 3 — copy runai-cli bin/dist, docs, version into cli-exposer/bin/
# Step 4 — build + load the image:
docker buildx build --platform linux/amd64 --provenance=false --load \
  --build-arg TARGETARCH=amd64 -t runai-local/cli-exposer:0.0.0 .
```

`DOCKERHUB_PROXY=` (empty) pulls the Go builder image from Docker Hub;
`GITHUB_TOKEN=` is emptied in the make call (the private-module token is applied
elsewhere). The image tag is **`0.0.0`** (the CP chart appVersion); any other
tag breaks the local deploy. Image ref: `runai-local/cli-exposer:0.0.0`.

## Testing

| Target | Description |
|--------|-------------|
| `make test` | `test-ginkgo` — runs the Ginkgo unit test suite |
| `make coverage` | Run tests with coverage (excludes `internal/generated` and `cmd/main.go`) and print a function-level report |

## Configuration

The service is configured via environment variables (config keys can be set by
their uppercase env-var equivalents; sensible defaults are built in):

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_PORT` | HTTP port to listen on | `8080` |
| `RETRY_MAX_INTERVAL_SEC` | Max backoff interval for retries | `10` |
| `RETRY_MAX_ELAPSED_SEC` | Max total elapsed retry time | `60` |
| `ROUTER_READ_TIME_SEC` | HTTP server read timeout | — |
| `ROUTER_WRITE_TIME_SEC` | HTTP server write timeout | — |
| `ROUTER_IDLE_TIME_SEC` | HTTP server idle timeout | — |

## OpenAPI and code generation

The HTTP API is defined under [`swagger/`](swagger); `make build` runs
`openapi-build` first, generating server stubs into
[`internal/generated`](internal/generated) via OpenAPI Generator. The generated
directory must not be edited by hand — changes are overwritten on regeneration.
</content>
