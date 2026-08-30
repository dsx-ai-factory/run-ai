# client-utils

`client-utils` is a small, dependency-light Go library in the Run:ai
`runai-application-packages` tree. It provides helpers that HTTP/API clients
across the platform reuse when building request query strings and normalizing
service base URLs. It is a **library package**: it has no binary, no Dockerfile,
and is not deployed — other modules consume it through `go.mod`.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Public API](#public-api)
- [Importing](#importing)
- [Building and testing](#building-and-testing)

## Overview

The package (`package client_utils`) offers two independent utilities:

- **Filter query building** — a fluent `FilterBuilder` that assembles the
  parameterized filter strings the platform's list endpoints understand. Each
  filter is `<field><operator><value>`, where the operator is one of the
  supported comparison tokens (`==`, `!=`, `<=`, `>=`, `=@` contains,
  `!@` not-contains, `=^` starts-with, `=$` ends-with). Multiple values for one
  field are OR-joined with `||`.
- **URL protocol normalization** — `GetInternalUrl` / `GetExternalUrl` ensure a
  service base URL carries a scheme, defaulting bare hosts to `http://`
  (internal) or `https://` (external) respectively.

## Repository layout

```text
.
├── filters.go            # FilterOperator constants, Filter, FilterBuilder
├── filters_test.go
├── protocols.go          # GetInternalUrl / GetExternalUrl scheme helpers
├── protocols_test.go
├── artifact-metadata.json
├── Makefile
└── go.mod
```

## Public API

| Symbol | Description |
|--------|-------------|
| `FilterOperator` + constants (`EqualOp`, `NotEqualOp`, `ContainsOp`, …) | Supported filter comparison operators |
| `NewFilterBuilder()` | Create a filter builder |
| `(*FilterBuilder).AddFilter(field, op, value)` | Append a single filter |
| `(*FilterBuilder).AddFilters(field, op, values)` | Append a filter with OR-joined values |
| `(*FilterBuilder).AddAFilter(*Filter)` | Append a pre-built `Filter` |
| `(*FilterBuilder).Get() []string` | Return the assembled filter strings |
| `GetInternalUrl(baseURL) string` | Prefix `http://` if no scheme present |
| `GetExternalUrl(baseURL) string` | Prefix `https://` if no scheme present |

## Importing

```go
import client_utils "github.com/run-ai/runai/runai-application-packages/client-utils"

url := client_utils.GetInternalUrl("my-service:8080")
filters := client_utils.NewFilterBuilder().
    AddFilter("status", client_utils.EqualOp, "Running").
    AddFilters("name", client_utils.ContainsOp, []string{"foo", "bar"}).
    Get()
```

## Building and testing

The module only `include`s the shared makefiles under
[`build/makefile`](../../build/makefile); it is a library, so there is no image
to build (`make build` is a no-op) and it is consumed purely via `go.mod`.

| Target | Description |
|--------|-------------|
| `make test` | Run unit tests (`test-go`) in the standard golang container |
| `make lint` | Run `golangci-lint` (`lint-go`) in a linter container |
| `make validate` | Package import/replace checks (no build step) |
| `make build` | No-op — this is a library, not a deployable artifact |

Tests run inside a Docker `golang:<version>` image (version derived from
[`go.mod`](go.mod)); `git-config-command` applies `GITHUB_TOKEN` for private
modules. To iterate without Docker, `go test ./...` works locally.
