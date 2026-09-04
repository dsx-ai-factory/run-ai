# assets-service-client — Build Context

**Type:** go (generated API client package)
**Team / owner service:** control-plane/assets-service
**Module:** github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client

## What it is

Versioned Go client for the assets-service API (OpenAPI v2.0.0). `generated/` is OpenAPI-generated
and must not be hand-edited; `api.go` + `internal/client.go` are a hand-written wrapper exposing
`IAssetsClient` / `IAssetsExternalClient` and constructors (`NewAssetsClient`, `NewAssetsExternalClient`,
`NewExternalClientWithConfig`). `mocks/` are testify mocks of those interfaces.

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

- Generated code lives in `generated/` and is produced by OpenAPI Generator `v7.12.0` (`-g go`).
  Do not edit it by hand.
- Regeneration is driven from the OWNING service, not here. From `control-plane/assets-service`:
  `make openapi-generate-code-client-go` (or the umbrella `make openapi-build` /
  `make openapi-generate-code`). It writes to `../../runai-clients/assets-service-client/generated`.
- Spec source of truth: `control-plane/assets-service/swagger/source/openapi.yaml`.
- No Dockerfile, no binary, no local-dev.sh deploy step — this is a consumed Go module, versioned
  via `runai-clients/assets-service-client/vX.Y.Z` tags and pinned in consumers' go.mod.
- Depends on `runai-application-packages/client-utils` for internal/external URL resolution.
