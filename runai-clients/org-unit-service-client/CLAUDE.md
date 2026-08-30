# org-unit-service-client — Build Context

**Type:** go (generated API client library)
**Module:** github.com/run-ai/runai/runai-clients/org-unit-service-client
**Wraps:** control-plane/org-unit-service (`Org unit` API — departments, projects, node types, notifications, reports)
**Spec source:** control-plane/org-unit-service/swagger/source/openapi.yaml

## Make targets

| Target | What it does |
|--------|-------------|
| build | No-op (generated library; nothing to compile) |
| lint | No-op |
| validate | No-op |
| test | Runs `test-go` — Go unit tests in the shared build container |
| generate-mocks | Regenerates `mocks/` via mockery |
| client-bump-version | Tag a prod release `runai-clients/org-unit-service-client/vX.Y.Z` (from synced master) |
| client-bump-version-dev | Tag a `-dev` release from a feature branch |

## Key build notes

- **Not a service/binary** — no Dockerfile, no `local-dev.sh` deploy step. It is a Go library
  consumed by control-plane services.
- **Generated code:** `generated/` is produced by OpenAPI Generator (Go client, v7.12.0; version
  pinned in `build/makefile/openapi.mk`). **Do not hand-edit it.**
- **Regeneration is driven from the owning service**, not from here:
  `cd control-plane/org-unit-service && make openapi-generate-code-client-go` (or the full
  `make openapi-generate-code`). Output lands in `runai-clients/org-unit-service-client/generated`.
- **Hand-written layer:** `api.go` (interfaces + constructor), `internal/client.go` (implementation),
  `structs/structs.go` (helper types). These are the only files to edit by hand.
- **Consumption:** module required at `v0.0.0` with a `replace => ../../runai-clients/org-unit-service-client`
  directive in each consumer's go.mod.

## local-dev.sh iteration

This package has no deploy step. To pick up changes, rebuild the consuming control-plane service:

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
