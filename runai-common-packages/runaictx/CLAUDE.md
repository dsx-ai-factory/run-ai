# runaictx — Build Context

**Type:** go (library — no binary, no image)
**Team:** runai-common-packages (shared infrastructure package)
**Module:** `github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx` (+ sub-package `.../runaictx/http`)

## Make targets

| Target | What it does |
|--------|-------------|
| `test` | `test-go` — runs `go test` inside the stock golang container |
| `lint` | `lint-go` — golangci-lint against `../.golangci.yaml` |
| `build` | No-op (library) |
| `validate` | No-op recipe; shared makefile adds package import/replace prerequisites |
| `package-bump-version` | Tag production release `runaictx/vX.Y.Z` |
| `package-bump-version-dev` | Tag dev release `runaictx/vX.Y.Z-dev.N` |

(No `purify` target in this package's Makefile.)

## Key build notes

- **Build command:** none — library consumed via `go.mod require` (pinned to a
  `runaictx/vX.Y.Z` tag). `make build` is a no-op; no image, no `cmd/main.go`, no
  Dockerfile, no `local-dev.sh` build path.
- Defines the canonical Run.ai `ContextField` keys (tenant/subject/groups/trace/…)
  and moves them between `context.Context`, `X-Runai-*` HTTP headers, log fields,
  and maps.
- Server side: `runaictx/http.ContextMiddleware` reads request headers into ctx.
  Client side: `http.SetHeaders`/`GetHeaders` write ctx onto outgoing requests.
- `LoggerHook(ctx)` bridges to `logging.WithCtx` for structured log enrichment.
- Depends on the sibling `logging` module and `rs/zerolog`.

## Releasing

Versioned by git tag `runaictx/vX.Y.Z`. Use `make package-bump-version` (production,
from main) or `make package-bump-version-dev` (any branch).
