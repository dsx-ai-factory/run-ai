# runai-cli

`runai-cli` is the Run:ai command-line tool (`runai`) — a researcher/admin CLI
for interacting with the Run:ai Control Plane and customer clusters (projects,
departments, workloads, assets, credentials, cluster operations, diagnostics,
self-upgrade). It is a Go binary built with [Cobra](https://github.com/spf13/cobra)
and [Viper](https://github.com/spf13/viper). It is **not** a deployed container
image — it cross-compiles to per-platform binaries that are distributed to users
by the [`cli-exposer`](../cli-exposer) service (see
the CLI is no longer
published as a standalone artifact). Owned by the **Viper Team**.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building](#building)
- [Testing](#testing)
- [Configuration](#configuration)
- [Getting a specific CLI version](#getting-a-specific-cli-version)
- [Developer notes](#developer-notes)

## Overview

The tool provides commands (grouped by Cobra) to:

- Authenticate against the control plane and manage credentials/tokens.
- Manage control-plane entities — projects, departments, nodes/nodepools,
  workloads, environments, compute, data sources, templates, applications.
- Operate against clusters — submit/list/describe workloads, exec, logs, port
  operations, cluster status, diagnostics/log collection.
- Self-upgrade and generate shell completion.

## Repository layout

```text
.
├── main.go                     # Entry point → cmd/runaicli.Execute()
├── cmd/
│   ├── runaicli/               # Cobra root command wiring
│   ├── common/
│   └── docgen/                 # Markdown docs generator
├── internal/                   # Command implementations and helpers
│   ├── controlplane/           # Control-plane API commands
│   ├── cluster/ k8s/ kubeconfig/ websocket/ network/
│   ├── auth/ credentials/ config/ cliflags/ flags/
│   ├── upgrade/ diagnostics/ completion/ autogenerate/mocks/
│   └── ...
├── pkg/
│   ├── config/
│   ├── logreader/
│   └── version/                # Version metadata (ldflags-injected)
├── docs/                       # Generated command docs (make build-docs-markdown)
├── scripts/                    # validations.sh, bump_version.sh, coverage.sh
├── version                     # RUNAI_CLI_MAJOR/MINOR/PATCH
├── .golangci.yml
├── Makefile
└── go.mod
```

## Building

Built with Go (see [`go.mod`](go.mod), Go 1.26) and orchestrated through the
[`Makefile`](Makefile). The version baked into the binary comes from the
[`version`](version) file (`RUNAI_CLI_MAJOR.MINOR.PATCH`, injected via
`-ldflags -X .../pkg/version.version`); builds are static (`CGO_ENABLED=0`,
`-extldflags "-static"`) and FIPS-enabled (`GOFIPS140=v1.0.0`).

| Target | Description |
|--------|-------------|
| `make dist` | `clean` + `generate` + `build-run` — build **all** platform binaries into `bin/dist/{os}/{arch}/runai`. This is the target `cli-exposer` and release pipelines use. |
| `make build` | **No-op**  — prints a message directing you to `make dist`. |
| `make build-run` | Creates the `bin/dist/*` layout and runs `build-all-platforms` in parallel (`-j$(CPU_AVAILABLE)`) |
| `make build-all-platforms` | Depends on all six per-platform targets below |
| `make cli-linux-amd64` | `GOOS=linux GOARCH=amd64` → `bin/dist/linux/amd64/runai` |
| `make cli-linux-arm64` | `GOOS=linux GOARCH=arm64` → `bin/dist/linux/arm64/runai` |
| `make cli-darwin-amd64` | `GOOS=darwin GOARCH=amd64` → `bin/dist/darwin/amd64/runai` |
| `make cli-darwin-arm64` | `GOOS=darwin GOARCH=arm64` → `bin/dist/darwin/arm64/runai` |
| `make cli-windows-amd64` | `GOOS=windows GOARCH=amd64` → `bin/dist/windows/amd64/runai` |
| `make cli-windows-arm64` | `GOOS=windows GOARCH=arm64` → `bin/dist/windows/arm64/runai` |
| `make generate` | `go generate` (run once before builds) |
| `make build-docs-markdown` | Builds the `docgen` tool and regenerates the command docs under `docs/` |
| `make clean` | `rm -rf bin/` |
| `make bump-version BUMP_TYPE={major\|minor\|patch}` | Bump the `version` file via `scripts/bump_version.sh` |

There is no Dockerfile and no deployed image for this service — the artifact is
the set of `runai` binaries. In a local `local-dev.sh` run, only the linux/amd64
binary is compiled (by `cli-exposer`'s build) and baked into the `cli-exposer`
image; producing all platforms is reserved for `make dist` / release pipelines.

## Testing

| Target | Description |
|--------|-------------|
| `make test` | `test-go` — run the unit tests in the Go builder container |
| `make test-local` | `go test ./... -v` on the host |
| `make test-dev` | Host tests with coverage profile (`cover.out`) |
| `make coverage` | Update coverage via `scripts/coverage.sh` |
| `make lint` | `lint-go` — golangci-lint (config `.golangci.yml`) |
| `make validate` | `validate-script` — runs `scripts/validations.sh` |
| `make generate-mockgen` | Install requirements and regenerate mockgen mocks into `internal/autogenerate/mocks` |

## Configuration

The CLI reads its configuration from a user config file (managed by Viper) and
from command-line flags; `runai config generate --config-path=<path>
--config-file=<file.yaml>` writes a config file populated with defaults.
Authentication state and credentials are stored per the `internal/config` and
`internal/credentials` packages.

## Getting a specific CLI version

The CLI is not published as a standalone artifact. To obtain a binary for a
specific version:

- **Build from source** — check out the tag and run `make dist` (or `make build`
  on older tags) in `control-plane/runai-cli/`; binaries land in
  `bin/dist/{os}/{arch}/runai`.
- **Pull from a running cluster** on that version:
  `curl https://<cluster>/api/v1/cli-exposer/dist/{os}/{arch}/runai`.
- **Extract from the `cli-exposer` image** for that version (binaries are baked
  at `/app/dist/{os}/{arch}/runai`):

  ```bash
  docker create --name tmp cli-exposer:<version>
  docker cp tmp:/app/dist/<os>/<arch>/runai ./runai && docker rm tmp
  ```

## Developer notes

- **Cobra** — the CLI is largely generated by `cobra-cli`. To add a command:
  `cobra-cli add <command>`, move it to the relevant directory under `cmd/` or
  `internal/`, register it in the `init` section, and wire up its function/flags.
  Keep `cmd/` limited to Cobra command definitions.
- **Linter** — run `make lint` before pushing. To bypass a single line, comment
  `//lint:<linter> // <reason>`; prefer updating `.golangci.yml` when justified.
- **Autocompletion** — debug dynamic completion with
  `__complete <sub command> <verb> "<prefix>"` (the trailing empty string is
  mandatory when there is no prefix).
- **Docs** — regenerate `docs/` with `make build-docs-markdown`.

[Design Document](https://runai.atlassian.net/wiki/spaces/EN/pages/2750513159/RunAI+CLI+Design)
</content>
