# s3 — Build Context

**Type:** go (library — no binary, no image)
**Team:** runai-common-packages (shared infrastructure package)
**Module:** `github.com/dsx-ai-factory/run-ai/runai-common-packages/s3`

## Make targets

| Target | What it does |
|--------|-------------|
| `test` | `test-go` — runs `go test` inside the stock golang container |
| `lint` | `lint-go` — golangci-lint against `../.golangci.yaml` |
| `purify` | `go fmt ./...` + `go vet ./...` + `go mod tidy` |
| `build` | No-op (library) |
| `validate` | No-op recipe; shared makefile adds package import/replace prerequisites |
| `package-bump-version` | Tag production release `s3/vX.Y.Z` |
| `package-bump-version-dev` | Tag dev release `s3/vX.Y.Z-dev.N` |

## Key build notes

- **Build command:** none — library consumed via `go.mod require` (pinned to an
  `s3/vX.Y.Z` tag). `make build` is a no-op; no image, no `cmd/main.go`, no
  Dockerfile, no `local-dev.sh` build path.
- Wraps the MinIO Go SDK. `Init(nil)` builds one global `*minio.Client` under
  `sync.Once`; `Client()` returns it (error if not initialized).
- Credential mode is chosen from config: empty `ACCESS_KEY_ID` → IAM role creds;
  otherwise static access-key/secret. `USE_SSL` toggles TLS. `ACCESS_KEY` is a
  secret-tagged field loaded from the mounted secrets folder via `InitConfig`.
- Config keys: `ENDPOINT`, `ACCESS_KEY_ID`, `ACCESS_KEY`, `USE_SSL`.
- Depends on `minio/minio-go/v7` and the sibling `config`/`logging` modules.

## Releasing

Versioned by git tag `s3/vX.Y.Z`. Use `make package-bump-version` (production, from
master) or `make package-bump-version-dev` (any branch).
