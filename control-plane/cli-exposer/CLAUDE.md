# cli-exposer — Build Context

**Type:** container
**Team:** Viper Team
**Module:** `github.com/run-ai/runai/control-plane/cli-exposer` (Go 1.26.3)

## Make targets

Targets come from the service `Makefile` plus the shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, and the `golang.mk` /
`openapi.mk` they pull in).

| Target | What it does |
|--------|-------------|
| `build` | `openapi-build` + `build-cli` + `build-go` — OpenAPI bundle, embed CLI artifacts, compile server for amd64 + arm64 |
| `build-cli` | `make -C ../runai-cli dist`, then copy CLI `dist/`, `docs/`, `version` into `bin/` |
| `build-go` | `build-go-amd` + `build-go-arm` — server binary (FIPS `GOFIPS140=v1.0.0`) → `bin/cli-exposer-<arch>` |
| `test` | `test-ginkgo` — Ginkgo unit test suite |
| `run-local` | build CLI artifacts + `go mod vendor` + build & run the server with dev ldflags |
| `lint` | `lint-go` — golangci-lint in the builder container |
| `purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `validate` | `openapi-validate` |
| `openapi-lint` | Lint the swagger bundle |
| `local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` (OS=linux ARCH=amd64) |
| `coverage` | Coverage report (excludes generated + `cmd/main.go`) |

## Key build notes

- Generated code: OpenAPI server stubs in `internal/generated` (`make
  openapi-build`, run before hand-building; never edit generated dirs).
- Runtime base image: final image is `FROM scratch`; a `ubi9/ubi-minimal` stage
  supplies `/etc/passwd`, certs (`/etc/ssl`, `/etc/pki`) and a writable `/tmp`.
  Runs as `65532:65532`, exposes `8080` + `9090`, entrypoint `./cli-exposer`.
  Embeds the runai-cli `dist/`, `docs/`, `version`, the unix installer template,
  the swagger bundle, and `internal/static`.
- Build args: `TARGETARCH` (selects `bin/cli-exposer-$TARGETARCH`).
- Build command (what local-dev.sh actually runs): a **dedicated
  `build_cli_exposer` function** (NOT `build_go_service`), because it also
  compiles the CLI. Its cache hash covers both `cli-exposer/` and `runai-cli/`.
  Steps:
  1. Compile the server binary: `make build-go-amd OS=linux ARCH=amd64
     DOCKERHUB_PROXY= GITHUB_TOKEN=` (`DOCKERHUB_PROXY=` pulls the Go builder
     from Docker Hub; `GITHUB_TOKEN=` emptied — token applied elsewhere).
  2. Compile runai-cli for linux/amd64 in `library/golang:1.26.3`
     (`GOFIPS140=v1.0.0`, `-o bin/dist/linux/amd64/runai`) — only linux/amd64;
     full `make dist` is release-only.
  3. Copy runai-cli `bin/dist`, `docs`, `version` into `cli-exposer/bin/`.
  4. Build + load: `docker buildx build --platform linux/amd64
     --provenance=false --load --build-arg TARGETARCH=amd64
     -t runai-local/cli-exposer:0.0.0 .`
     Image ref: `runai-local/cli-exposer:0.0.0` (tag `0.0.0` — CP chart
     appVersion; any other tag breaks the local deploy).

## local-dev.sh iteration

This is a control-plane service (rebuilding it also picks up runai-cli changes):
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
</content>
