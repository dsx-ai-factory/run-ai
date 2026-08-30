# org-units-migrator — Build Context

**Type:** container (run-once DB migration Job)
**Team:** Suricata
**Module:** `github.com/run-ai/runai/control-plane/org-units-migrator`

This is a **migration job**, not a long-running HTTP service — it has no OpenAPI
spec and no server. It reads legacy org-unit data from the `backend` schema and
writes it into the `org_unit_service` schema, per tenant, then exits.

## Make targets

Targets come from the service `Makefile` plus the shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, and the `golang.mk` they
pull in).

| Target | What it does |
|--------|-------------|
| `build` | `build-go` — compiles the Go binary for amd64 + arm64 |
| `build-go` | `build-go-amd` + `build-go-arm` |
| `build-go-amd` | Compiles `cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) in the Go builder container → `bin/org-units-migrator-amd64` |
| `build-go-arm` | Same for `arm64` → `bin/org-units-migrator-arm64` |
| `test` | `test-ginkgo` — Ginkgo suites (`ginkgo -r`, timeout 5m) over `migration/` + `database/` |
| `lint` | `lint-go` — golangci-lint in the builder container |
| `purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `validate` | No-op (no OpenAPI spec) |
| `local-deploy` | `build-go` + `docker-build-generic` + `docker-push-generic` (linux/amd64) |

## Key build notes

- No generated code, no OpenAPI. `make validate` is a no-op.
- Runtime base image: final image is `FROM scratch`; a
  `registry.access.redhat.com/ubi9/ubi-minimal` stage supplies `/etc/passwd`,
  `/etc/group`, certs (`/etc/ssl`, `/etc/pki`), `nsswitch.conf` and a writable
  `/tmp`. Runs as user `65532:65532`. Binary copied in as `/org_units_migrator`;
  that is the entrypoint.
- Build args: `TARGETARCH` (selects `bin/org-units-migrator-$TARGETARCH`).
- Build command (what local-dev.sh actually runs for this service): it is in
  `CP_GO_SERVICES`, so `build_go_service` runs, in two steps:
  1. Compile the amd64 binary in the builder container:
     `make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=`
     (`DOCKERHUB_PROXY=` pulls the Go builder image from Docker Hub; `GITHUB_TOKEN=`
     is emptied in the make call — the private-module token is applied elsewhere.)
  2. Build + load the image: `docker buildx build --platform linux/amd64
     --provenance=false --load --build-arg TARGETARCH=amd64 -t
     runai-local/org-units-migrator:0.0.0 .`
     Image ref: `runai-local/org-units-migrator:0.0.0` (tag `0.0.0` — the CP chart
     appVersion; any other tag breaks the local deploy).
- Configuration is env-var driven (no config file). Migration behavior:
  `DB_SCHEMA_SOURCE` (default `backend`), `DB_SCHEMA_TARGET` (default
  `org_unit_service`), `DB_MIGRATION_VERSION` (default `v1`),
  `SHOULD_MARK_COMPLETED`, `TENANTS_TO_MIGRATE`, `TENANTS_TO_NOT_MIGRATE`. DB
  connection via standard `DB_*` vars. Idempotent per-tenant via a
  `migration_control` table.

## local-dev.sh iteration

This is a control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
