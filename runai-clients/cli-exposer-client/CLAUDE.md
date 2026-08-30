# cli-exposer-client — Build Context

**Type:** go (generated API client package)
**Team / owner service:** control-plane/cli-exposer
**Module:** github.com/run-ai/runai/runai-clients/cli-exposer-client

## What it is

Pure OpenAPI-generated Go client for the cli-exposer API (OpenAPI v2.0.0) — the RunAI researcher CLI
surface. No hand-written wrapper — everything under `generated/` is machine-produced and must not be
hand-edited. `mocks/` are testify mocks of the generated API services. Consumers import
`.../cli-exposer-client/generated` directly.

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
- Regeneration is driven from the OWNING service, not here. From `control-plane/cli-exposer`:
  `make openapi-generate-code-client-go` (or `make openapi-build` / `make openapi-generate-code`).
  It writes to `../../runai-clients/cli-exposer-client/generated`.
- Spec source of truth: `control-plane/cli-exposer/swagger/source/openapi.yaml`.
- No Dockerfile, no binary, no local-dev.sh deploy step — consumed Go module, versioned via
  `runai-clients/cli-exposer-client/vX.Y.Z` tags and pinned in consumers' go.mod.
