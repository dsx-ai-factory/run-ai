# metric-stores-migrator — Build Context

**Type:** container (run-once migration job — no HTTP API, no DB)
**Team:** Phoenix
**Module:** `github.com/run-ai/runai/control-plane/metric-stores-migrator`

## What it is

A Kubernetes `Job` run once by the control-plane Helm chart. It reads each
tenant's metric store from `tenants-manager`, translates it to the
`metrics-service` schema, and calls `CreateMetricStore` on metrics-service.
Read-only on the source; idempotent on the target (`409 Conflict` = already
migrated). Re-run by deleting the job:
`kubectl delete job runai-backend-metric-stores-migrator`.

## Make targets

Targets come from the service `Makefile` plus the shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, and the `golang.mk` /
`base.mk` they pull in).

| Target | What it does |
|--------|-------------|
| `build` | `build-go` — compile amd64 + arm64 binaries |
| `build-go` | `build-go-amd` + `build-go-arm` |
| `build-go-amd` | Compiles `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) in the Go builder container → `bin/metric-stores-migrator-amd64` |
| `build-go-arm` | Same for `arm64` → `bin/metric-stores-migrator-arm64` |
| `test` | `build` + `test-go` |
| `test-go` | Runs the Go unit tests in the builder container |
| `generate-mocks` | Regenerates mockery mocks under `internal/mocks` (`.mockery.yml` present) |
| `lint` | `lint-go` — golangci-lint in the builder container |
| `purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `validate` | No-op (no OpenAPI spec) |
| `local-deploy` | `build-go` + `docker-build-generic` + `docker-push-generic` (amd64) |
| `docker-build-generic` / `docker-push-generic` | Build / push the container image |

## Key build notes

- No generated OpenAPI code here; only mockery mocks (`make generate-mocks`, mockery `v3.2.5`). Never edit `internal/mocks` by hand.
- Runtime base image: final image is `FROM scratch`; a `registry.access.redhat.com/ubi9/ubi-minimal` stage supplies `/etc/passwd`, `/etc/group`, certs (`/etc/ssl`, `/etc/pki`), `nsswitch.conf`, and a writable `/tmp`. Runs as user `65532:65532`. Binary copied in as `metric_stores_migrator` and set as `ENTRYPOINT`.
- Build args: `TARGETARCH` (selects `bin/metric-stores-migrator-$TARGETARCH`).
- Build command (what local-dev.sh actually runs for this service): it is in `CP_GO_SERVICES`, so `build_go_service` runs, in two steps:
  1. Compile the amd64 binary in the builder container:
     `make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=`
     (`DOCKERHUB_PROXY=` pulls the Go builder image from Docker Hub; `GITHUB_TOKEN=` is emptied in the make call — the private-module token is applied elsewhere.)
  2. Build + load the image: `docker buildx build --platform linux/amd64 --provenance=false --load --build-arg TARGETARCH=amd64 -t runai-local/metric-stores-migrator:0.0.0 .`
     Image ref: `runai-local/metric-stores-migrator:0.0.0` (tag `0.0.0` — the CP chart appVersion; any other tag breaks the local deploy).
- Configuration is env-var driven (no config file): `TENANT_MANAGER_URL` (default `http://runai-backend-tenants-manager:8080`), `METRICS_SERVER_HOST` (default `runai-backend-metrics-service:8080`), `METRICS_SERVER_SCHEME` (default `http`). Each key can be overridden by its uppercase env-var equivalent.
- No integration tests / docker-compose; unit tests use mockery mocks only.

## local-dev.sh iteration

This is a control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
