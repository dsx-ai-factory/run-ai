# queryapi — Build Context

**Type:** go (library package)
**Team:** runai-application-packages
**Module:** `github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi`
(integration tests are a separate module:
`.../queryapi/integration-tests`)

## Make targets

| Target | What it does |
|--------|-------------|
| test | Runs `test-go` (unit) **and** `integration-tests` |
| integration-tests | `cd integration-tests && go test -v ./...` (DB-backed) |
| lint | Runs `lint-go` — `golangci-lint run` in a linter container |
| validate | Package import/replace checks (via `packages.mk`) |
| build | No-op — library package, no binary/image |

## Key build notes

- **Build command:** none. Library consumed via `go.mod`; `make build` is a
  no-op. No Dockerfile, no `cmd/main.go`, no `local-dev.sh` build function.
- Go version derived from `go.mod` (`go 1.26.3`).
- Depends on `runai-common-packages/runaidb` (`orm`, `orm/clause`) for the
  `dbenrich` enrichers.
- `integration-tests/` is a distinct Go module with its own `go.mod` and SQL
  migrations (`integration-tests/migrations/up`); it needs a database.
- Unit-test recipe (`test-go`) runs in a Docker `golang:<ver>` container; the
  `integration-tests` recipe runs `go test` directly on the host.
- Local iteration without Docker: `go test ./...`.

## No local-dev.sh iteration

This package is not built or deployed by `local-dev.sh`. Consumers pick it up
through their own `go.mod` require directives.
