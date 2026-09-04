# testutils

`testutils` is a Go test-helper library in the Run:ai
`runai-application-packages` tree. It collects utilities that other packages'
test suites reuse: golden-file (input/expected) loading, readable map/YAML
comparison for Gomega, random test-data generation, and consistent random
tenant/cluster/project identifier sets. It is a **library package**: no binary,
no Dockerfile, consumed by other modules through `go.mod` (typically as a test
dependency).

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Public API](#public-api)
- [Importing](#importing)
- [Building and testing](#building-and-testing)

## Overview

The package (`package testutils`) provides four groups of helpers:

- **Golden-file loading (`file_based.go`)** — `LoadInputFile` /
  `LoadExpectedFile` (and their `*From` variants) read YAML fixtures from a
  test's `testing/inputs` and `testing/expected` directories. If an expected
  file is missing, it is auto-written from the actual output and the test fails
  so the new golden file can be reviewed. Failures are reported through
  Ginkgo/Gomega.
- **Map comparison (`maps.go`)** — `Expect(x).ToEqual(y)` marshals both operands
  to YAML and diffs them (via `google/go-cmp`) so mismatches print as a readable
  YAML diff instead of an opaque struct dump.
- **Random strings (`randomize.go`)** — `RandString(n)` builds a random
  alphabetic string; `RandStringTestMode` restricts the alphabet for
  deterministic, comparable output.
- **Random identity sets (`randoms.go`)** — `RandomValues` holds a coherent set
  of random tenant/cluster/department/project identifiers with `Clone*` builders
  (e.g. `CloneWithDifferentTenant`, `CloneWithSpecificProject`) and
  `Forbidden()/Allowed()` markers for authorization-style table tests.

## Repository layout

```text
.
├── file_based.go          # Golden input/expected YAML loaders
├── maps.go                # MapComparator: Expect(...).ToEqual(...)
├── randomize.go           # RandString / RandStringTestMode
├── randoms.go             # RandomValues + Clone* builders
├── artifact-metadata.json
├── Makefile
└── go.mod
```

## Public API

| Symbol | Description |
|--------|-------------|
| `LoadInputFile` / `LoadInputFileFrom` | Read a YAML input fixture by test name |
| `LoadExpectedFile` / `LoadExpectedFileFrom` | Read (or create) a golden expected YAML |
| `Expect(x).ToEqual(y)` (`MapComparator`) | Gomega-style readable YAML diff comparison |
| `RandString(n)` / `RandStringTestMode(letters)` | Random / deterministic random strings |
| `RandomValues` + `NewRandomValues()` + `Clone*` / `Forbidden()` / `Allowed()` | Coherent random tenant/cluster/project id sets |

## Importing

```go
import "github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

vals := testutils.NewRandomValues()
other := vals.CloneWithDifferentTenant()
testutils.Expect(gotMap).ToEqual(wantMap)
```

## Building and testing

The module only `include`s the shared makefiles under
[`build/makefile`](../../build/makefile); it is a library, so there is no image
to build (`make build` is a no-op) and it is consumed purely via `go.mod`. It
depends on Ginkgo/Gomega, `google/go-cmp`, and `sigs.k8s.io/yaml` — the helpers
are meant to be called from inside test code.

| Target | Description |
|--------|-------------|
| `make test` | Run unit tests (`test-go`) in the standard golang container |
| `make lint` | Run `golangci-lint` (`lint-go`) in a linter container |
| `make validate` | Package import/replace checks (no build step) |
| `make build` | No-op — this is a library, not a deployable artifact |

Tests run inside a Docker `golang:<version>` image (version derived from
[`go.mod`](go.mod)). To iterate without Docker, `go test ./...` works locally.
