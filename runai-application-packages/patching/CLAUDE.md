# patching — Build Context

**Type:** go (library package)
**Team:** runai-application-packages
**Module:** `github.com/dsx-ai-factory/run-ai/runai-application-packages/patching`

## Make targets

| Target | What it does |
|--------|-------------|
| test | Runs `test-go` — `go test` inside a `golang:<ver>` container (Ginkgo suite) |
| lint | Runs `lint-go` — `golangci-lint run` in a linter container |
| validate | Package import/replace checks (via `packages.mk`) |
| build | No-op — library package, no binary/image |

## Key build notes

- **Build command:** none. Library consumed via `go.mod`; `make build` is a
  no-op. No Dockerfile, no `cmd/main.go`, no `local-dev.sh` build function.
- Go version derived from `go.mod` (`go 1.26.3`).
- Depends on sibling packages `convutils` and `testutils`, plus
  `PaesslerAG/jsonpath` and Ginkgo/Gomega.
- Tests read golden YAML from `input/` and `expected/`; a missing expected file
  is auto-created and the run fails so it can be reviewed.
- Real recipes live in `build/makefile/golang.mk` and `build/makefile/packages.mk`.
- Local iteration without Docker: `go test ./...`.

## No local-dev.sh iteration

This package is not built or deployed by `local-dev.sh`. Consumers pick it up
through their own `go.mod` require directives.
