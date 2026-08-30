# cluster-api-client — Build Context

**Type:** go (generated API client package)
**Team / owner service:** runai-cluster/cluster/cluster-api
**Module:** github.com/run-ai/runai/runai-clients/cluster-api-client

## What it is

Pure OpenAPI-generated Go client for the cluster-api API (OpenAPI v1.0.0) — proxying k8s object
operations and workload status. No hand-written wrapper — everything under `generated/` is
machine-produced and must not be hand-edited. `mocks/` are testify mocks of the generated API
services. Consumers import `.../cluster-api-client/generated` directly.

## Make targets

| Target | What it does |
|--------|-------------|
| build | No-op (library module; nothing to compile) |
| lint | No-op |
| validate | No-op |
| test | Runs `test-go` (shared Go unit tests in the Go builder image) |

The Makefile only sets `SERVICE_NAME` and `include`s `../../build/makefile/control-plane.mk` and
`../../build/makefile/index.mk`.

## Key build notes

- Generated code lives in `generated/`, produced by OpenAPI Generator `7.7.0` (`-g go`). Do not edit.
- Regeneration is driven from the OWNING service, not here. From `runai-cluster/cluster/cluster-api`:
  `make openapi-generate-code-client-go` (or `make openapi-build` / `make openapi-generate-code`).
  It writes to `../../runai-clients/cluster-api-client/generated`.
- Spec source of truth: `runai-cluster/cluster/cluster-api/swagger/source/openapi.yaml`.
- No Dockerfile, no binary, no local-dev.sh deploy step — consumed Go module, versioned via
  `runai-clients/cluster-api-client/vX.Y.Z` tags and pinned in consumers' go.mod.
