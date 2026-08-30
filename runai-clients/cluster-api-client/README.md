# cluster-api-client

Generated Go client for the Run:ai **cluster-api** API ("An API for proxying k8s commands on workloads"). It is the typed, versioned Go module that other services import to talk to the cluster-api service over HTTP — proxying Kubernetes object operations and workload status/commands. The package is produced entirely by [OpenAPI Generator](https://openapi-generator.tech) from the cluster-api OpenAPI spec and **must not be edited by hand**.

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

- **`generated/`** — the OpenAPI-generated Go client: `APIClient`, per-tag API services (`ClusterApiAPI`, `K8sObjectAPI`, `StatusAPI`, `HealthAPI`), model structs (`model_object.go`, `model_workload_status.go`, `model_error.go`), `Configuration`, and generated markdown docs.
- **`mocks/`** — generated testify mocks of the API services for consumers' unit tests.

## Repository layout

```text
.
├── generated/              # OpenAPI-generated Go client (DO NOT EDIT)
│   ├── api_*.go            # Per-tag API services (cluster_api, k8s_object, status, health)
│   ├── model_*.go          # Request/response model structs
│   ├── client.go           # Generated APIClient
│   ├── configuration.go    # Servers / auth / HTTP configuration
│   ├── docs/               # Generated markdown docs
│   └── README.md           # Generator-produced usage notes
├── mocks/                  # Generated testify mocks of the API services
├── Makefile
├── go.mod                  # module github.com/run-ai/runai/runai-clients/cluster-api-client
└── go.sum
```

## Which API this wraps

| | |
|---|---|
| Upstream service | `runai-cluster/cluster/cluster-api` |
| OpenAPI source (source of truth) | [`runai-cluster/cluster/cluster-api/swagger/source/openapi.yaml`](../../runai-cluster/cluster/cluster-api/swagger/source/openapi.yaml) |
| API version | 1.0.0 |
| Generator | OpenAPI Generator `7.7.0` (Go client) |
| Languages / artifacts generated | Go |

## Importing the client

There is no wrapper package; consumers import the generated package directly:

```go
import openapi "github.com/run-ai/runai/runai-clients/cluster-api-client/generated"

cfg := openapi.NewConfiguration()
cfg.Servers = openapi.ServerConfigurations{{URL: baseURL}}
client := openapi.NewAPIClient(cfg)

resp, httpResp, err := client.K8sObjectAPI.
	// ...call the desired operation
	Execute()
```

Pin the module in your `go.mod` to a published version tag of the form `runai-clients/cluster-api-client/vX.Y.Z`.

## Regenerating the client

This module is **not** regenerated from its own directory. The generated code is produced from the owning service, `runai-cluster/cluster/cluster-api`, which points the shared OpenAPI tooling at this directory (`SHARED_PACKAGES_PATH/cluster-api-client/generated`). From `runai-cluster/cluster/cluster-api` run the Go-client generation target:

```bash
# from runai-cluster/cluster/cluster-api
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

Note: the real build/validation of the generated code happens in the owning service (`runai-cluster/cluster/cluster-api`), which also diffs freshly generated client code against this checked-in copy to detect drift.
