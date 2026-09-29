# workloads-manager-client

Generated Go client for the **workloads-manager** control-plane service. It wraps the
`Workloads V2 Service` HTTP API so other Run:ai backend services can create, read, update and batch
V2 workloads. The package is owned by the same team that owns the
[workloads-manager](../../control-plane/workloads-manager) service; the generated code plus a
hand-written wrapper live here in `runai-clients/` so they can be versioned and imported
independently of the server.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Importing the client](#importing-the-client)
- [OpenAPI and code generation](#openapi-and-code-generation)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Versioning](#versioning)

## Overview

This package is split into two parts:

- **Generated OpenAPI client** ([`generated/`](generated)) — low-level API structs and models produced
  by [OpenAPI Generator](https://openapi-generator.tech) from the service's OpenAPI spec (V2 workload
  create/get/update, batch operations and health). **This directory must not be edited by hand.**
- **Hand-written wrapper** — [`api.go`](api.go) declares the `IWorkloadsManagerClient` /
  `IWorkloadsManagerExternalClient` interfaces and the `NewWorkloadsManagerClient` constructor, and
  [`internal/client.go`](internal/client.go) implements them on top of the generated client.
  [`mocks/`](mocks) contains [mockery](https://github.com/vektra/mockery)-generated test doubles.

Consumers program against the wrapper interfaces, not the generated code directly. The exported
methods cover `CreateWorkloadV2`, `GetWorkloadV2ById`, `UpdateWorkloadV2` and `BatchWorkloads`.

## Repository layout

```text
.
├── api.go                  # IWorkloadsManagerClient interfaces + NewWorkloadsManagerClient (hand-written)
├── internal/
│   └── client.go           # Interface implementation over the generated client
├── generated/              # OpenAPI-generated client + models (DO NOT EDIT)
├── mocks/                  # mockery-generated mocks of the client interfaces
├── artifact-metadata.json  # artifact_type: go
├── Makefile
├── go.mod                  # module github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client
└── go.sum
```

## Importing the client

The module path is `github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client`. Inside the
monorepo, consuming services require it at `v0.0.0` and point at the local copy with a `replace`
directive in their `go.mod`:

```go
require github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client v0.0.0

replace github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client => ../../runai-clients/workloads-manager-client
```

Then construct the client with `NewWorkloadsManagerClient(baseURL)` and call the interface methods.

## OpenAPI and code generation

The API contract lives with the **server**, not here:
[`control-plane/workloads-manager/swagger/source/openapi.yaml`](../../control-plane/workloads-manager/swagger/source/openapi.yaml).
The Go client in [`generated/`](generated) is produced by OpenAPI Generator (Go client generator,
version pinned in [`build/makefile/openapi.mk`](../../build/makefile/openapi.mk)) as part of the
**owning service's** build. Regenerate it from the service directory, not from here:

```bash
cd ../../control-plane/workloads-manager
make openapi-generate-code            # server stubs + Go client + JS client
# or just the Go client:
make openapi-generate-code-client-go
```

That target writes into `runai-clients/workloads-manager-client/generated`, runs `gofmt`, drops the
generator's test files and tidies this module. **Never hand-edit `generated/`** — changes are
overwritten on the next generation.

## Building and running

This is a library, not a binary — there is nothing to compile into an executable. The package
[`Makefile`](Makefile) includes the shared control-plane makefiles, and deliberately makes the
following targets no-ops:

| Target | Description |
|--------|-------------|
| `make build` | No-op (nothing to build; it is a generated library) |
| `make lint` | No-op |
| `make validate` | No-op |

| Target | Description |
|--------|-------------|
| `make generate-mocks` | Regenerate the interface mocks in `mocks/` via mockery |

## Testing

| Target | Description |
|--------|-------------|
| `make test` | Runs `test-go` — the Go unit tests for this module in the shared build container |

## Versioning

Released client versions are cut with the shared versioning targets (see
[`build/makefile/versioning.mk`](../../build/makefile/versioning.mk)), which tag under
`runai-clients/workloads-manager-client/vX.Y.Z`:

| Target | Description |
|--------|-------------|
| `make client-bump-version` | Cut a production client tag (only from a synced `main`) |
| `make client-bump-version-dev` | Cut a `-dev` client tag from a feature branch |

Within the monorepo consumers use the `v0.0.0` + `replace` workspace pattern above, so a tag bump is
only needed when a version must be referenced outside the workspace. Do not merge a PR that requires a
`-dev` client version — CI rejects it.
