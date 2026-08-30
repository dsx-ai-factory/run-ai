# integdb — Build Context

**Type:** go (test-support library — no service binary, no published image)
**Team:** runai-common-packages (shared infrastructure package)
**Module:** `github.com/run-ai/runai/runai-common-packages/integdb`

## Make targets

| Target | What it does |
|--------|-------------|
| `test` | `test-go` — `go test` in the stock golang container; **requires Docker** (starts a Postgres testcontainer) |
| `lint` | `lint-go` — golangci-lint against `../.golangci.yaml` |
| `purify` | `go fmt ./...` + `go vet ./...` + `go mod tidy` |
| `build` | No-op (library) |
| `validate` | No-op recipe; shared makefile adds package import/replace prerequisites |
| `package-bump-version` | Tag production release `integdb/vX.Y.Z` |
| `package-bump-version-dev` | Tag dev release `integdb/vX.Y.Z-dev.N` |

## Key build notes

- **Build command:** none — library consumed via `go.mod require` (pinned to an
  `integdb/vX.Y.Z` tag). `make build` is a no-op; there is no `cmd/main.go` and no
  `local-dev.sh` build path.
- **Dockerfile:** present but NOT a service image. It is a testcontainers build
  context — `FROM postgres:${POSTGRES_VERSION}` plus `postgresql-${POSTGRES_VERSION_MAJOR}-cron`
  (build args `POSTGRES_VERSION`, `POSTGRES_VERSION_MAJOR`). `SetupPostgresDB` builds
  it on the fly so `pg_cron` is available in the test database.
- Single entry point `SetupPostgresDB` returns a `runaidb/config.Config` + cleanup func.
- Depends on `testcontainers-go` (+ its postgres module), `config`, and `runaidb`.
- Tests need a running Docker daemon.

## Releasing

Versioned by git tag `integdb/vX.Y.Z`. Use `make package-bump-version` (production,
from master) or `make package-bump-version-dev` (any branch).
