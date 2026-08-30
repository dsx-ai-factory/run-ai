# runaihttp — Build Context

**Type:** go (library — no binary, no image)
**Team:** runai-common-packages (shared infrastructure package)
**Module:** `github.com/run-ai/runai/runai-common-packages/runaihttp`

## Make targets

| Target | What it does |
|--------|-------------|
| `test` | `test-go` — runs `go test` inside the stock golang container |
| `lint` | `lint-go` — golangci-lint against `../.golangci.yaml` |
| `purify` | `go fmt ./...` + `go vet ./...` + `go mod tidy` |
| `build` | No-op (library) |
| `validate` | No-op recipe; shared makefile adds package import/replace prerequisites |
| `package-bump-version` | Tag production release `runaihttp/vX.Y.Z` |
| `package-bump-version-dev` | Tag dev release `runaihttp/vX.Y.Z-dev.N` |

## Key build notes

- **Build command:** none — library consumed via `go.mod require` (pinned to a
  `runaihttp/vX.Y.Z` tag). `make build` is a no-op; no image, no `cmd/main.go`, no
  Dockerfile, no `local-dev.sh` build path.
- `ServeHTTP(chi.Router, *ServerConfig)` runs a chi-based server with graceful
  shutdown; `ServerConfig` loads from env when nil (`SERVICE_PORT`, `SERVER_*`).
  An `init()` implements a `--health` CLI probe used as a container health check.
- `LoadErrorDetails` maps Postgres `pgconn.PgError` SQL states to HTTP statuses.
- Assert* helpers implement OpenAPI field constraints; csv_support.go has non-JSON
  response writers (text/html/file/csv) — note csv_support.go contains mustache
  (`{{...}}`) template fragments consumed by the OpenAPI generator.
- Depends on `go-chi/chi/v5`, `jackc/pgx/v5`, and the sibling `config`/`logging` modules.

## Releasing

Versioned by git tag `runaihttp/vX.Y.Z`. Use `make package-bump-version` (production,
from master) or `make package-bump-version-dev` (any branch).
