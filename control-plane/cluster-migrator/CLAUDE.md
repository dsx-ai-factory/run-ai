# cluster-migrator — Build Context

**Type:** container (run-once DB migration **job**, not a long-running service — no HTTP API)
**Team:** Griffin
**Module:** `github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator`

## Make targets

Targets come from the service `Makefile` plus the shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, and the `golang.mk` they
pull in).

| Target | What it does |
|--------|-------------|
| `build` | `build-go` — compiles the Go binary for amd64 + arm64 |
| `build-go` | `build-go-amd` + `build-go-arm` |
| `build-go-amd` | Compiles `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) in the Go builder container → `bin/cluster-migrator-amd64` |
| `build-go-arm` | Same for `arm64` → `bin/cluster-migrator-arm64` |
| `test` | `test-ginkgo` — Ginkgo unit tests |
| `lint` | `lint-go` — golangci-lint in the builder container |
| `purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `validate` | no-op (kept for CP interface parity) |
| `local-deploy` | `build-go` + `docker-build-generic` + `docker-push-generic` (OS=linux, ARCH=amd64) |

## Key build notes

- No generated code, no OpenAPI/swagger, no `pkg/client`. Config is struct-tag
  driven (`config/config.go`) with env-var overrides; DB config comes from the
  shared `runaidb` package.
- Runtime base image: final image is `FROM scratch`; a
  `registry.access.redhat.com/ubi9/ubi-minimal` stage supplies `/etc/passwd`,
  `/etc/group`, certs (`/etc/ssl`, `/etc/pki`) and a writable `/tmp`. Runs as
  user `65532:65532`. Binary is copied in as `cluster_migrator` and is the
  entrypoint.
- Build args: `TARGETARCH` (selects `bin/cluster-migrator-$TARGETARCH`).
- Build command (what local-dev.sh actually runs): it is in `CP_GO_SERVICES`, so
  `build_go_service` runs, in two steps:
  1. `make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=`
     (`DOCKERHUB_PROXY=` pulls the Go builder image from Docker Hub;
     `GITHUB_TOKEN=` is emptied in the make call — private-module token applied
     elsewhere.)
  2. `docker buildx build --platform linux/amd64 --provenance=false --load --build-arg TARGETARCH=amd64 -t runai-local/cluster-migrator:0.0.0 .`
     Image ref: `runai-local/cluster-migrator:0.0.0` (tag `0.0.0` — the CP chart
     appVersion; any other tag breaks the local deploy).
- Runtime behavior: idempotent. The single migration (node-pool system
  settings) is gated by a control table
  (`NodepoolSystemSettingsMigrationControl`), auto-migrated into the target
  schema at startup. Re-running after success is a no-op; force a re-run by
  bumping `DB_MIGRATION_VERSION`.
- Runs in the install/upgrade flow as the `runai-backend-cluster-migrator` Job.

## local-dev.sh iteration

This is a control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
