# authorization-migrator — Build Context

**Type:** container (run-once DB migration job, not a long-running service)
**Team:** Suricata Team
**Module:** `github.com/dsx-ai-factory/run-ai/control-plane/authorization-migrator`

## What it is

A one-shot Kubernetes Job run by the control-plane Helm chart during
install/upgrade. It migrates legacy `2.11` authorization role-assignment data
from the `backend`/`tenants_manager` source schemas into the modern
`authorization` target schema (2.12+). Reads-only from the sources; never mutates
legacy data. Idempotent via a `MigrationControl` version row in the target
schema (`DB_MIGRATION_VERSION`). No HTTP API.

## Make targets

The service `Makefile` sets `SERVICE_NAME=authorization-migrator` and includes
`../../build/makefile/control-plane.mk` + `index.mk` (which pulls in
`golang.mk`). Local targets delegate to the shared Go recipes:

| Target | What it does |
|--------|-------------|
| `build` | `build-go` = `build-go-amd` + `build-go-arm` — compile amd64 + arm64 binaries (`bin/authorization-migrator-<arch>`) in the Go builder container, FIPS `GOFIPS140=v1.0.0` |
| `test` | `test-go` — Go unit tests in the builder container |
| `lint` | `lint-go` — golangci-lint in the builder container |
| `purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `validate` | no-op (interface parity across CP services) |

## Key build notes

- Generated code: none (no OpenAPI/swagger). It depends on the sibling
  `control-plane/authorization` module for the target schema's DAOs and generated
  API types.
- Runtime base image: final image is `FROM scratch`; a
  `registry.access.redhat.com/ubi9/ubi-minimal` `base` stage supplies
  `/etc/passwd`, `/etc/group`, certs (`/etc/ssl`, `/etc/pki`) and a writable
  `/tmp`. Runs as user `65532:65532`; `ENTRYPOINT ["./authorization_migrator"]`.
- Build args: `TARGETARCH` (selects `bin/authorization-migrator-$TARGETARCH`).
- Build command (what local-dev.sh runs — it is in `CP_GO_SERVICES`, so
  `build_go_service` runs, in two steps):
  1. Compile the amd64 binary in the builder container:
     `make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=`
     (`DOCKERHUB_PROXY=` pulls the Go builder image from Docker Hub;
     `GITHUB_TOKEN=` is emptied in the make call — the private-module token is
     applied elsewhere.)
  2. Build + load the image:
     `docker buildx build --platform linux/amd64 --provenance=false --load --build-arg TARGETARCH=amd64 -t runai-local/authorization-migrator:0.0.0 .`
     Image ref: `runai-local/authorization-migrator:0.0.0` (tag `0.0.0` — the CP
     chart appVersion; any other tag breaks the local deploy).
- Config: env-var driven (`DB_*`, `KEYCLOAK_*`); secret keys also read as files
  under `config/secrets/`. No config YAML file.

## local-dev.sh iteration

This is a control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
