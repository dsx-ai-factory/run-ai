# catalog-service-client

Generated Go client for the Run:ai **catalog-service** API ("An API for managing catalogs"). It is the typed, versioned Go module that other control-plane and cluster services import to talk to catalog-service over HTTP — managing Helm repositories and charts, container images and tags, registries, inference entries, NGC-integration settings and related catalog resources. The package is produced entirely by [OpenAPI Generator](https://openapi-generator.tech) from the catalog-service OpenAPI spec and **must not be edited by hand**.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Which API this wraps](#which-api-this-wraps)
- [Importing the client](#importing-the-client)
- [Regenerating the client](#regenerating-the-client)
- [Building and running](#building-and-running)
- [Testing](#testing)

## Overview

This module is a pure generated client (there is no hand-written wrapper layer). It contains:

- **`generated/`** — the OpenAPI-generated Go client: `APIClient`, per-tag API services (`HealthAPI`, `HelmReposAPI`, `ImagesAPI`, `RegistriesAPI`, `InferencesAPI`, `SettingsAPI`), request/response model structs (`model_*.go`), `Configuration`, and generated markdown docs.
- **`mocks/`** — generated testify mocks (`HealthAPI`, `ImagesAPI`, `RegistriesAPI`) for consumers' unit tests.

## Repository layout

```text
.
├── generated/              # OpenAPI-generated Go client (DO NOT EDIT)
│   ├── api_*.go            # Per-tag API services (helm_repos, images, registries, …)
│   ├── model_*.go          # Request/response model structs
│   ├── client.go           # Generated APIClient
│   ├── configuration.go    # Servers / auth / HTTP configuration
│   ├── docs/               # Generated markdown docs
│   └── README.md           # Generator-produced usage notes
├── mocks/                  # Generated testify mocks of the API services
├── Makefile
├── go.mod                  # module github.com/run-ai/runai/runai-clients/catalog-service-client
└── go.sum
```

## Which API this wraps

| | |
|---|---|
| Upstream service | `control-plane/catalog-service` |
| OpenAPI source (source of truth) | [`control-plane/catalog-service/swagger/source/openapi.yaml`](../../control-plane/catalog-service/swagger/source/openapi.yaml) |
| API version | 2.0.0 |
| Generator | OpenAPI Generator `v7.12.0` (Go client) |
| Languages / artifacts generated | Go |

## Importing the client

There is no wrapper package; consumers import the generated package directly:

```go
import openapi "github.com/run-ai/runai/runai-clients/catalog-service-client/generated"

cfg := openapi.NewConfiguration()
cfg.Servers = openapi.ServerConfigurations{{URL: baseURL}}
client := openapi.NewAPIClient(cfg)

resp, httpResp, err := client.ImagesAPI.ListImages(ctx).Execute()
```

Pin the module in your `go.mod` to a published version tag of the form `runai-clients/catalog-service-client/vX.Y.Z`.

## Regenerating the client

This module is **not** regenerated from its own directory. The generated code is produced from the owning service, `control-plane/catalog-service`, which points the shared OpenAPI tooling at this directory (`SHARED_PACKAGES_PATH/catalog-service-client/generated`). From `control-plane/catalog-service` run the Go-client generation target:

```bash
# from control-plane/catalog-service
make openapi-generate-code-client-go     # regenerate only the Go client
make openapi-build                       # bundle spec + generate server, Go client and JS client
```

The generation runs the pinned `openapitools/openapi-generator-cli` image (`-g go`, `generateInterfaces=true`, `packageVersion=1.17`, `REMOVE_X_INTERNAL=true`) and overwrites `generated/`. Everything under `generated/` is machine-produced — never edit it by hand.

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

Note: the real build/validation of the generated code happens in the owning service (`control-plane/catalog-service`), which also diffs freshly generated client code against this checked-in copy to detect drift.
