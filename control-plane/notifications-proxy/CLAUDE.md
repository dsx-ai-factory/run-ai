# notifications-proxy — Build Context

**Type:** container
**Team:** Phoenix
**Module:** `github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy`

## Make targets

Targets come from the service `Makefile` plus the shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, and the `golang.mk` /
`openapi.mk` they pull in).

| Target | What it does |
|--------|-------------|
| `build` | `openapi-build` then `build-go` — builds the OpenAPI bundle, then compiles the Go binary for amd64 + arm64 |
| `build-go` | `build-go-amd` + `build-go-arm` — both arch binaries |
| `build-go-amd` | Compiles `./cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) inside the Go builder container → `bin/notifications-proxy-amd64` |
| `build-go-arm` | Same for `arm64` → `bin/notifications-proxy-arm64` |
| `test` | `build` + `test-go` + `integration-tests` |
| `test-go` | Runs the Go unit tests in the builder container |
| `generate-mocks` / `generate-mocks-local` | Shared targets: regenerate test mocks from `.mockery.yml` with mockery v3 (in Docker / without Docker) |
| `generate-integration-test-smtp-openapi-client` | Regenerates the MailHog SMTP test client from its OpenAPI spec |
| `build-notifications-service-bin` | Compiles the sibling `notifications-service` binary so the compose stack builds locally |
| `build-integration-tests` | Builds the `-tags integration` test binary + `docker compose build` |
| `integration-tests` | Full compose stack (all profiles), abort-on-container-exit |
| `integration-tests-debug` | Compose stack without the test container (run tests from IDE) |
| `lint` | `lint-go` — golangci-lint in the builder container |
| `purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `validate` | `openapi-validate` — validate the OpenAPI spec |
| `openapi-build` / `openapi-lint` | Build / lint the swagger bundle |
| `local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` (linux/amd64) |

## Key build notes

- Generated code: OpenAPI server stubs (`internal/generated`) + Go client (`pkg/client/generated`) via openapi-generator `v7.12.0`, `VERSIONED_CLIENT=true`; plus mockery v3 mocks from `.mockery.yml` (`make generate-mocks`, or `make generate-mocks-local` without Docker). `make build` runs `openapi-build` first; run generation before hand-building and never edit generated dirs.
- Runtime base image: final image is `FROM scratch`; a `registry.access.redhat.com/ubi9/ubi-minimal` stage supplies `/etc/passwd`, `/etc/group`, certs (`/etc/ssl`, `/etc/pki`) and a writable `/tmp`. Runs as user `65532:65532`. Bundles SQL migrations (`internal/database/migrations/up|down`) and the swagger bundle.
- Build args: `TARGETARCH` (selects `bin/notifications-proxy-$TARGETARCH`).
- Build command (what local-dev.sh actually runs for this service): it is in `CP_GO_SERVICES`, so `build_go_service` runs, in two steps:
  1. Compile the amd64 binary in the builder container:
     `make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=`
     (`DOCKERHUB_PROXY=` pulls the Go builder image from Docker Hub; `GITHUB_TOKEN=` is emptied in the make call — the private-module token is applied elsewhere.)
  2. Build + load the image: `docker buildx build --platform linux/amd64 --provenance=false --load --build-arg TARGETARCH=amd64 -t runai-local/notifications-proxy:0.0.0 .`
     Image ref: `runai-local/notifications-proxy:0.0.0` (tag `0.0.0` — the CP chart appVersion; any other tag breaks the local deploy).
- Config: env-var driven (keys have uppercase env-var equivalents); secrets (`DB_USER`, `DB_PASSWORD`, `DB_ENCRYPTION_KEY`, `SLACK_CLIENT_ID`, `SLACK_CLIENT_SECRET`) read from files under `SECRETS_MOUNT_PATH`. Requires PostgreSQL + pub/sub (redis) + upstream notifications-service.
- Integration tests: require docker-compose + backing services (PostgreSQL, Redis, notifications-service, MailHog SMTP, Slack mock).

## local-dev.sh iteration

This is a control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
</content>
