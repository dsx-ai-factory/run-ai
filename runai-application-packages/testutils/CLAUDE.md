# testutils — Build Context

**Type:** go (library package)
**Team:** runai-application-packages
**Module:** `github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils`

## Make targets

| Target | What it does |
|--------|-------------|
| test | Runs `test-go` — `go test` inside a `golang:<ver>` container |
| lint | Runs `lint-go` — `golangci-lint run` in a linter container |
| validate | Package import/replace checks (via `packages.mk`) |
| build | No-op — library package, no binary/image |

## Key build notes

- **Build command:** none. Library consumed via `go.mod` (usually as a test
  dependency); `make build` is a no-op. No Dockerfile, no `cmd/main.go`, no
  `local-dev.sh` build function.
- Go version derived from `go.mod` (`go 1.26.3`).
- Depends on Ginkgo/Gomega, `google/go-cmp`, `google/uuid`, and
  `sigs.k8s.io/yaml`; helpers assert via Ginkgo/Gomega, so they run inside test
  code.
- Golden-file helpers expect fixtures under the caller's `testing/inputs` and
  `testing/expected` directories; a missing expected file is auto-created and
  the test fails for review.
- Real recipes live in `build/makefile/golang.mk` and `build/makefile/packages.mk`.
- Local iteration without Docker: `go test ./...`.

## No local-dev.sh iteration

This package is not built or deployed by `local-dev.sh`. Consumers pick it up
through their own `go.mod` require directives.
