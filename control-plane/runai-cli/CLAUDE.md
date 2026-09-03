# runai-cli — Build Context

**Type:** go (CLI binary — cross-platform, NOT a deployed image)
**Team:** Viper Team
**Module:** `github.com/dsx-ai-factory/run-ai/control-plane/runai-cli` (Go 1.26.3)

## Make targets

Targets come from the service `Makefile` plus the shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, and `golang.mk` /
`openapi.mk`).

| Target | What it does |
|--------|-------------|
| `dist` | `clean` + `generate` + `build-run` — build ALL platform binaries into `bin/dist/{os}/{arch}/runai`. Used by cli-exposer + release pipelines. |
| `build` | **No-op** ; prints a message pointing to `make dist` |
| `build-run` | Create `bin/dist/*` layout, run `build-all-platforms` with `-j$(CPU_AVAILABLE)` |
| `build-all-platforms` | All six per-platform targets |
| `cli-{linux,darwin,windows}-{amd64,arm64}` | `GOOS/GOARCH go build` (FIPS `GOFIPS140=v1.0.0`, static, ldflags-injected version) → `bin/dist/<os>/<arch>/runai` |
| `generate` | `go generate` (run before builds) |
| `build-docs-markdown` | Build `docgen`, regenerate `docs/` |
| `test` | `test-go` — unit tests in the Go builder container |
| `test-local` / `test-dev` | Host `go test` (with coverage for `test-dev`) |
| `coverage` | `scripts/coverage.sh` |
| `lint` | `lint-go` — golangci-lint (shared `../infra/lint/.golangci.yaml`) |
| `validate` | `validate-script` → `scripts/validations.sh` |
| `generate-mockgen` | Install requirements + regenerate mocks → `internal/autogenerate/mocks` |
| `bump-version BUMP_TYPE=...` | `scripts/bump_version.sh` |
| `clean` | `rm -rf bin/` |

## Key build notes

- **No Dockerfile / no deployed image.** The artifact is the set of `runai`
  binaries; there is no `runai-local/runai-cli:<tag>` image and no per-service
  `local-dev.sh` deploy step.
- Version: from the `version` file (`RUNAI_CLI_MAJOR.MINOR.PATCH`), injected via
  `-ldflags -X .../pkg/version.version`. Builds are static (`CGO_ENABLED=0`,
  `-extldflags "-static" -s -w`) and FIPS-enabled (`GOFIPS140=v1.0.0`).
- Generated code: mockgen mocks (`make generate-mockgen`) + `go generate`; don't
  edit generated dirs.
- Built with Cobra (`cobra-cli`) + Viper.
- **How it reaches a local cluster:** `local-dev.sh` does not build runai-cli as
  its own service. The `cli-exposer` build (`build_cli_exposer`) compiles the
  runai-cli **linux/amd64** binary in `library/golang:1.26.3` and bakes it into
  the cli-exposer image. Full multi-platform `make dist` is release-only.

## local-dev.sh iteration

runai-cli has no standalone deploy. To pick up CLI changes in a local cluster,
rebuild cli-exposer (its build hash covers `runai-cli/`):
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
</content>
