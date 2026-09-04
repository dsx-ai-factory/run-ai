# logging — Build Context

**Type:** go (library — no binary, no image)
**Team:** runai-common-packages (shared infrastructure package)
**Module:** `github.com/dsx-ai-factory/run-ai/runai-common-packages/logging` (+ sub-package `.../logging/http`)

## Make targets

| Target | What it does |
|--------|-------------|
| `test` | `test-go` — runs `go test` inside the stock golang container |
| `lint` | `lint-go` — golangci-lint against `../.golangci.yaml` |
| `build` | No-op (library) |
| `validate` | No-op recipe; shared makefile adds package import/replace prerequisites |
| `package-bump-version` | Tag production release `logging/vX.Y.Z` |
| `package-bump-version-dev` | Tag dev release `logging/vX.Y.Z-dev.N` |

(No `purify` target in this package's Makefile.)

## Key build notes

- **Build command:** none — library consumed via `go.mod require` (pinned to a
  `logging/vX.Y.Z` tag). `make build` is a no-op; no image, no `cmd/main.go`, no
  Dockerfile, no `local-dev.sh` build path.
- Thin wrapper over zerolog. `Init(nil)` loads `Config` from env (`LOG_LEVEL`,
  `LOG_PRETTY`); package-level `Infof`/etc. log through the global logger and are
  no-ops until `Init` is called (`Fatalf` still exits).
- Every message records its format string under a `log_pattern` field for aggregation.
- `WithCtx(CtxLoggerHook)` enriches logs from `context.Context`; pair with
  `runaictx.LoggerHook(ctx)`. `logging/http.HttpMiddleware` logs each HTTP request.
- Depends on the sibling `config` and `runaictx` modules and `rs/zerolog`.

## Releasing

Versioned by git tag `logging/vX.Y.Z`. Use `make package-bump-version` (production,
from master) or `make package-bump-version-dev` (any branch).
