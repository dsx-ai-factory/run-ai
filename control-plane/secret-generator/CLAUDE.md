# secret-generator — Build Context

**Type:** container
**Team:** Griffin
**Module:** `github.com/dsx-ai-factory/run-ai/control-plane/secret-generator`

## Make targets

Targets come from the service `Makefile` plus the shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, and the `golang.mk` / `base.mk` they pull
in). This service has **no** OpenAPI/swagger and **no** `pkg/client`, so there are no
`openapi-*`, `validate` (real), or `client-*` targets — `make validate` is a deliberate no-op.

| Target | What it does |
|--------|-------------|
| `build` | `build-go` — compiles the Go binary for amd64 + arm64 |
| `build-go` | `build-go-amd` + `build-go-arm` |
| `build-go-amd` | Compiles `./cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) in the Go builder container → `bin/secret-generator-amd64` |
| `build-go-arm` | Same for `arm64` → `bin/secret-generator-arm64` |
| `test` | `test-ginkgo` — Ginkgo unit-test suites (pure unit tests, fake k8s client, no backing services) |
| `test-go` | Plain `go test ./...` in the builder container |
| `lint` | `lint-go` — golangci-lint in the linter container |
| `purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `validate` | No-op (kept for a uniform target set across CP services) |
| `docker-build-generic` | Build the container image via `docker buildx` |
| `docker-push-generic` | `docker push ${DOCKER_IMAGE_NAME}` |
| `local-deploy` | `build-go` (OS=linux ARCH=amd64) → `docker-build-generic` → `docker-push-generic` |

## Key build notes

- No generated code: there is no swagger spec or generated client/server to regenerate before
  building. Just `make build`.
- Runtime base image: final image is `FROM scratch`; a
  `registry.access.redhat.com/ubi9/ubi-minimal` stage supplies `/etc/passwd`, `/etc/group`,
  certs (`/etc/ssl`, `/etc/pki`), `/etc/nsswitch.conf`, and a writable `/tmp`. Runs as user
  `65532:65532`. Bundles the compiled binary + `NOTICE`.
- Build args: `TARGETARCH` (selects `bin/secret-generator-$TARGETARCH`).
- Runtime: this is a **batch job / Helm hook**, not a long-running server. It uses the
  in-cluster ServiceAccount, loads a JSON secret-spec file from `CONFIG_FILE_PATH`, reconciles
  Secrets in the release namespace, and exits (5-minute overall timeout).
- Config: key-driven with defaults in `config/config.go`; each key is overridable by its
  uppercase env var (`NAMESPACE`, `CONFIG_FILE_PATH`, `FULLNAME_OVERRIDE`, `LOG_LEVEL`,
  `DRY_RUN`). The Helm chart injects these.
- Build command (what local-dev.sh actually runs for this service): it is in `CP_GO_SERVICES`,
  so `build_go_service` runs, in two steps:
  1. Compile the amd64 binary in the builder container:
     `make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=`
     (`DOCKERHUB_PROXY=` pulls the Go builder image from Docker Hub; `GITHUB_TOKEN=` is emptied
     in the make call — the private-module token is applied elsewhere.)
  2. Build + load the image:
     `docker buildx build --platform linux/amd64 --provenance=false --load --build-arg TARGETARCH=amd64 -t runai-local/secret-generator:0.0.0 .`
     Image ref: `runai-local/secret-generator:0.0.0` (tag `0.0.0` — the CP chart appVersion;
     any other tag breaks the local deploy).
- Tests: pure Go unit tests under `pkg/generator/`; no docker-compose or backing services.

## local-dev.sh iteration

This is a control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
</content>
