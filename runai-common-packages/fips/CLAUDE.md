# fips — Build Context

**Type:** go (library — no binary, no image)
**Team:** runai-common-packages (shared infrastructure package)
**Module:** `github.com/run-ai/runai/runai-common-packages/fips`

## Make targets

| Target | What it does |
|--------|-------------|
| `test` | `test-go` — runs `go test` inside the stock golang container |
| `lint` | `lint-go` — golangci-lint against `../.golangci.yaml` |
| `purify` | `go fmt ./...` + `go vet ./...` + `go mod tidy` |
| `build` | No-op (library) |
| `validate` | No-op recipe; shared makefile adds package import/replace prerequisites |
| `package-bump-version` | Tag production release `fips/vX.Y.Z` |
| `package-bump-version-dev` | Tag dev release `fips/vX.Y.Z-dev.N` |

## Key build notes

- **Build command:** none — library consumed via `go.mod require` (pinned to a
  `fips/vX.Y.Z` tag). `make build` is a no-op; no image, no `cmd/main.go`, no
  Dockerfile, no `local-dev.sh` build path.
- The whole package is a single `init()` in `fips.go` that logs the
  `crypto/fips140` status through the sibling `logging` package. Import it for its
  side effect: `import _ ".../runai-common-packages/fips"`.
- Actual FIPS enablement/enforcement is decided by the build (`GOFIPS140`, which the
  shared Go build targets set to `v1.0.0`) and runtime `GODEBUG=fips140=...`, not by
  this package — it only observes and reports.
- Only runtime dependency is the sibling `logging` module.

## Releasing

Versioned by git tag `fips/vX.Y.Z`. Use `make package-bump-version` (production, from
master) or `make package-bump-version-dev` (any branch).
