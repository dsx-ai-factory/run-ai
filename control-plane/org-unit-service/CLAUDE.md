# org-unit-service — Build Context

**Type:** container
**Team:** Suricata Team (`@suricata`)
**Module:** `github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service`

Control-plane Go/REST service that owns the organizational-unit hierarchy
(departments, projects), their resource quotas / node-pool allocations, over-time
usage metrics, reports and settings. Backed by PostgreSQL schema `org_unit_service`.

## Make targets

| Target | What it does |
|--------|-------------|
| `build` | `openapi-build` + `build-go` (validate/bundle spec, regen server code, compile amd64 **and** arm64) |
| `build-go-amd` | Compile linux/amd64 binary in Docker → `bin/org-unit-service-amd64` |
| `build-go-arm` | Compile linux/arm64 binary in Docker → `bin/org-unit-service-arm64` |
| `openapi-build` | Bundle spec, build docs, publish public spec, regenerate server code |
| `validate` | `openapi-validate` — validate spec + generated-code-up-to-date + public routes |
| `openapi-lint` | Lint the swagger bundle |
| `lint` | `lint-go` (golangci-lint; shared config `../infra/lint/.golangci.yaml`) |
| `purify` | `purify-go` — auto-fix Go formatting/imports |
| `test` | `test-ginkgo` — run Ginkgo/Gomega unit suites (recursive, randomized, 5m) |
| `local-deploy` | Generate code, build amd64 binary, `docker-build` + `docker-push` |

Shared targets come from `../../build/makefile/control-plane.mk` and
`../../build/makefile/index.mk` (base.mk, golang.mk, openapi.mk, …).

## Key build notes

- **Generated code:** `make build` runs `openapi-build` first — server stubs land
  in `generated/` (OpenAPI Generator server `v7.12.0`). Never hand-edit
  `generated/`; `make validate` fails if it drifts from the spec.
- **Base image:** final stage is `scratch`; a `registry.access.redhat.com/ubi9/ubi-minimal`
  base stage supplies passwd/group, CA/PKI trust, `nsswitch.conf` and a writable
  `/tmp`. Runs as non-root `65532:65532`; entrypoint `./org-unit-service`.
- **Build args:** `TARGETARCH` (selects `bin/org-unit-service-$TARGETARCH`).
- **Image contents:** binary + `NOTICE` + `migrations/{up,down}` + bundled swagger.
- **Build command (local-dev.sh):** from the service dir it runs
  `make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=`, then
  `docker buildx build --build-arg TARGETARCH=amd64 -t runai-local/org-unit-service:0.0.0 .`
  Only amd64 is built locally. Image tag is the fixed CP version **`0.0.0`**
  (CP Helm chart `appVersion`); any other tag breaks the local deploy.
  `DOCKERHUB_PROXY=` empty → builder image from Docker Hub; `GITHUB_TOKEN=` empty
  in the make call (token applied elsewhere).
- **Config model:** config-file driven (registered typed sections with per-section
  defaults). Keys are flat; any key is overridable by its uppercase env-var name.
  No checked-in sample config; DB keys come from shared `runaidb`. See README.
- **Client:** versioned Go client is the separate `control-plane/org-unit-service/pkg/client`
  module, not `pkg/client/`.

## local-dev.sh iteration

Control-plane service — rebuild CP only, reuse the running cluster:

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
