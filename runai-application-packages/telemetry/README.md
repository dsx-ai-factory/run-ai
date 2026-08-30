# telemetry

`telemetry` is a small, zero-dependency Go library in the Run:ai
`runai-application-packages` tree. It holds shared types and helpers for the
platform's telemetry "group-by" reporting — translating the API-level grouping
fields a caller requests into the concrete database fields a metrics/telemetry
query should group on. It is a **library package**: no binary, no Dockerfile,
consumed by other modules through `go.mod`.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Public API](#public-api)
- [Importing](#importing)
- [Building and testing](#building-and-testing)

## Overview

The package (`package telemetry_utils`) provides:

- **`TelemetryGroupValue` (`telemetry_group.go`)** — a flat struct describing a
  single telemetry data point together with up to two grouping dimensions, each
  with a main and secondary name/value pair (`Group1MainName/Value`,
  `Group1SecondaryName/Value`, `Group2*`). It is the shared shape passed between
  producers and consumers of grouped telemetry.
- **`GetModelGroupBy` (`utils.go`)** — validates a requested list of group-by
  fields (max two) against a `map[string][]string` that maps each API field name
  to its underlying DB field(s), returning the resolved `[][]string` or an error
  listing the valid group names when an unknown field is requested.

## Repository layout

```text
.
├── telemetry_group.go     # TelemetryGroupValue struct
├── utils.go               # GetModelGroupBy validation/translation helper
├── artifact-metadata.json
├── Makefile
└── go.mod
```

## Public API

| Symbol | Description |
|--------|-------------|
| `TelemetryGroupValue` | Struct: a data point plus up to two group dimensions |
| `GetModelGroupBy(groupBy []string, modelFieldToDbFieldMap map[string][]string) ([][]string, error)` | Validate (max 2) and translate API group fields to DB fields |

## Importing

```go
import telemetry_utils "github.com/run-ai/runai/runai-application-packages/telemetry"

dbGroups, err := telemetry_utils.GetModelGroupBy(
    []string{"project", "cluster"},
    map[string][]string{
        "project": {"project_id"},
        "cluster": {"cluster_id"},
    },
)
```

## Building and testing

The module only `include`s the shared makefiles under
[`build/makefile`](../../build/makefile); it is a library, so there is no image
to build (`make build` is a no-op) and it is consumed purely via `go.mod`. The
package currently has no test files of its own.

| Target | Description |
|--------|-------------|
| `make test` | Run unit tests (`test-go`) in the standard golang container |
| `make lint` | Run `golangci-lint` (`lint-go`) in a linter container |
| `make validate` | Package import/replace checks (no build step) |
| `make build` | No-op — this is a library, not a deployable artifact |

Tests run inside a Docker `golang:<version>` image (version derived from
[`go.mod`](go.mod)). To iterate without Docker, `go test ./...` works locally.
