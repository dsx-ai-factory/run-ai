# dao-utils — Build Context

**Type:** go (library — no binary, no image)
**Team:** runai-common-packages (shared infrastructure package)
**Module:** `github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils` (Go package `dao_utils`)

## Make targets

| Target | What it does |
|--------|-------------|
| `test` | `go test ./...` (some suites spin PostgreSQL via testcontainers — needs Docker) |
| `lint` | `lint-go` — golangci-lint against `../.golangci.yaml` |
| `purify` | `go fmt ./...` + `go vet ./...` + `go mod tidy` |
| `build` | No-op (library) |
| `validate` | No-op recipe; shared makefile adds `validate-package-imports/replaces` prerequisites |
| `package-bump-version` | Tag production release `dao-utils/vX.Y.Z` (must be on main, in sync with origin) |
| `package-bump-version-dev` | Tag dev release `dao-utils/vX.Y.Z-dev.N` |

## Key build notes

- **Build command:** none — this is a library consumed via `go.mod require` by other
  modules (pinned to a `dao-utils/vX.Y.Z` tag). `make build` is intentionally a no-op;
  no Docker image is produced and there is no `local-dev.sh` build path.
- **No Dockerfile / no cmd/main.go.**
- Makefile includes `../../build/makefile/index.mk`; `test`/`lint` recipes are defined
  locally (`test` uses plain `go test ./...`, not the shared `test-go` docker path).
- Depends on GORM (postgres + sqlite drivers), golang-migrate, testcontainers-go, and
  the sibling packages `config`, `logging`, `metrics`, `tracing`, plus `jsonutils`.
- Tests that use testcontainers require a running Docker daemon.

## Releasing

Versioned by git tag `dao-utils/vX.Y.Z`. Use `make package-bump-version` (production,
from main) or `make package-bump-version-dev` (any branch) — see the
[runai-common-packages README](../README.md).
