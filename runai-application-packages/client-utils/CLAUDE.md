# client-utils — Build Context

**Type:** go (library package)
**Team:** runai-application-packages
**Module:** `github.com/run-ai/runai/runai-application-packages/client-utils`

## Make targets

| Target | What it does |
|--------|-------------|
| test | Runs `test-go` — `go test` inside a `golang:<ver>` container |
| lint | Runs `lint-go` — `golangci-lint run` in a linter container |
| validate | Package import/replace checks (via `packages.mk`) |
| build | No-op — library package, no binary/image |

## Key build notes

- **Build command:** none. This is a library consumed by other modules via
  `go.mod`; `make build` and `make validate` produce no image. There is no
  Dockerfile, no `cmd/main.go`, and no `local-dev.sh` build function.
- Go version derived from `go.mod` (`go 1.26.3`).
- Make targets only `include ../../build/makefile/index.mk`; the real recipes
  live in `build/makefile/golang.mk` (`test-go`, `lint-go`) and
  `build/makefile/packages.mk` (`validate` prerequisites).
- Tests execute in Docker via `DOCKER_GO_COMMAND_LINUX`; `GITHUB_TOKEN` is wired
  through `git-config-command` for private `github.com/run-ai/*` modules.
- Local iteration without Docker: `go test ./...`.

## No local-dev.sh iteration

This package is not built or deployed by `local-dev.sh`. Consumers pick it up
through their own `go.mod` require directives.
