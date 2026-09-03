# assets-service-client

Generated Go client for the Run:ai **assets-service** API ("An API for managing assets"). It is the typed, versioned Go module that other control-plane and cluster services import to talk to assets-service over HTTP — listing and creating compute, environment, model, datasource, credentials and workload-template assets, and driving the internal cluster-sync / events endpoints. The bulk of the package is produced by [OpenAPI Generator](https://openapi-generator.tech) from the assets-service OpenAPI spec and **must not be edited by hand**; a thin hand-written wrapper adds an ergonomic, interface-based facade on top of the generated code.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Which API this wraps](#which-api-this-wraps)
- [Importing the client](#importing-the-client)
- [Regenerating the client](#regenerating-the-client)
- [Building and running](#building-and-running)
- [Testing](#testing)

## Overview

The package has two layers:

- **Generated layer** (`generated/`) — the raw OpenAPI-generated Go client: `APIClient`, per-tag API services (`ComputeAPI`, `EnvironmentAPI`, `ModelAPI`, `CredentialsAPI`, `DatasourcesAPI`, `SynchronizationAPI`, …), request/response model structs and `Configuration`. This directory is machine-generated and regenerated from the spec.
- **Wrapper layer** (`api.go`, `internal/`) — a hand-maintained facade in package `assets_service_client` that exposes the `IAssetsClient` / `IAssetsExternalClient` interfaces and constructors (`NewAssetsClient`, `NewAssetsExternalClient`, `NewExternalClientWithConfig`). The constructors resolve internal vs external base URLs via `client-utils` and wrap `generated.APIClient`. `mocks/` holds generated testify mocks of these interfaces for consumers' unit tests.

## Repository layout

```text
.
├── api.go                  # Wrapper facade: IAssetsClient interfaces + constructors (hand-written)
├── internal/
│   └── client.go           # Wrapper implementation delegating to the generated APIClient
├── generated/              # OpenAPI-generated Go client (DO NOT EDIT)
│   ├── api_*.go            # Per-tag API services (compute, environment, model, …)
│   ├── model_*.go          # Request/response model structs
│   ├── client.go           # Generated APIClient
│   ├── configuration.go    # Servers / auth / HTTP configuration
│   ├── docs/               # Generated markdown docs
│   └── README.md           # Generator-produced usage notes
├── mocks/                  # Generated testify mocks of the wrapper interfaces
├── Makefile
├── go.mod                  # module github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client
└── go.sum
```

## Which API this wraps

| | |
|---|---|
| Upstream service | `control-plane/assets-service` |
| OpenAPI source (source of truth) | [`control-plane/assets-service/swagger/source/openapi.yaml`](../../control-plane/assets-service/swagger/source/openapi.yaml) |
| API version | 2.0.0 |
| Generator | OpenAPI Generator `v7.12.0` (Go client) |
| Languages / artifacts generated | Go |

## Importing the client

Consumers depend on this as a standalone Go module and import the wrapper package:

```go
import assets_service_client "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"

client := assets_service_client.NewAssetsClient(baseURL)
resp, httpResp, err := client.ListComputeAssets(ctx, name)
```

Model types are referenced from the generated subpackage:

```go
import openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
```

Pin the module in your `go.mod` to a published version tag of the form `runai-clients/assets-service-client/vX.Y.Z`.

## Regenerating the client

This module is **not** regenerated from its own directory. The generated code is produced from the owning service, `control-plane/assets-service`, which points the shared OpenAPI tooling at this directory (`SHARED_PACKAGES_PATH/assets-service-client/generated`). From `control-plane/assets-service` run the Go-client generation target:

```bash
# from control-plane/assets-service
make openapi-generate-code-client-go     # regenerate only the Go client
make openapi-build                       # bundle spec + generate server, Go client and JS client
```

The generation runs the pinned `openapitools/openapi-generator-cli` image (`-g go`, `generateInterfaces=true`, `packageVersion=1.17`, `REMOVE_X_INTERNAL=true`) and overwrites `generated/`. After regenerating, only hand-edit `api.go` / `internal/` — never files under `generated/`.

## Building and running

This is a library module, not a runnable binary, so the local Makefile targets are intentionally no-ops except for tests. It only `include`s the shared makefiles (`build/makefile/control-plane.mk`, `build/makefile/index.mk`).

| Target | Description |
|--------|-------------|
| `make build` | No-op (nothing to compile; the module is consumed as a package) |
| `make lint` | No-op |
| `make validate` | No-op |

## Testing

| Target | Description |
|--------|-------------|
| `make test` | Runs `test-go` — the shared Go unit-test target, executing the package tests inside the pinned Go builder image |

Note: the real build/validation of the generated code happens in the owning service (`control-plane/assets-service`), which also diffs freshly generated client code against this checked-in copy to detect drift.
