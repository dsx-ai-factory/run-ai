# patching

`patching` is a Go library in the Run:ai `runai-application-packages` tree that
implements the reflection- and JSON-path-based building blocks for PATCH-style
partial updates. Services that expose PATCH APIs use it to merge a sparse
"patch" struct onto existing data and to strip explicitly-cleared fields out of
a spec. It is a **library package**: no binary, no Dockerfile, consumed by other
modules through `go.mod`.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Public API](#public-api)
- [Importing](#importing)
- [Building and testing](#building-and-testing)

## Overview

The package (`package patching`) provides two complementary mechanisms:

- **Struct patching (`patching.go`)** — works on structs whose fields are
  pointers. `ApplyPatch(data, patch)` recursively copies every non-nil field of
  `patch` onto `data`; nil patch fields are left untouched (so "absent" means
  "no change"). Nested struct pointers are recursed into, and a nested struct
  that ends up entirely nil is collapsed back to nil. `CheckIfAllMembersAreNil`
  reports whether a pointer-to-struct has all pointer/collection members unset,
  and the generic `PatchField[T](**field, *value)` overwrites a field only when a
  value is supplied.
- **Field clearing (`clear_fields.go`)** — `ClearFields(spec, jsonPaths, opts)`
  removes entries from a decoded JSON/YAML map (`map[string]interface{}`) by a
  list of JSON paths. It supports flat fields (`spec.image`), nested fields
  (`spec.compute.gpuDevicesRequest`), and array elements
  (`spec.compute.extendedResources[0]`). Array removals are marked and compacted
  in a second pass so that clearing multiple indices on the same array does not
  shift the wrong items. `ClearFieldsOptions.IgnoreNonExisting` toggles whether
  missing paths error or are silently skipped.

## Repository layout

```text
.
├── patching.go            # ApplyPatch, CheckIfAllMembersAreNil, PatchField
├── patching_test.go
├── clear_fields.go        # ClearFields + JSON-path parsing (uses PaesslerAG/jsonpath)
├── clear_fields_test.go
├── suite_test.go          # Ginkgo suite bootstrap
├── input/                 # Test input YAML payloads
├── expected/              # Golden expected YAML payloads
├── artifact-metadata.json
├── Makefile
└── go.mod
```

## Public API

| Symbol | Description |
|--------|-------------|
| `ApplyPatch(data, patch interface{})` | Recursively apply non-nil pointer fields of `patch` onto `data` |
| `CheckIfAllMembersAreNil(input interface{}) bool` | True if all members of a pointer-to-struct are nil |
| `PatchField[T comparable](**T, *T)` | Set a field pointer only if a value is provided |
| `ClearFields(spec map[string]interface{}, jsonPaths []string, ClearFieldsOptions) error` | Remove fields/array elements by JSON path |
| `ClearFieldsOptions{IgnoreNonExisting bool}` | Options for `ClearFields` |

## Importing

```go
import "github.com/run-ai/runai/runai-application-packages/patching"

patching.ApplyPatch(existing, patchRequest)
err := patching.ClearFields(spec,
    []string{"spec.image", "spec.env[0]"},
    patching.ClearFieldsOptions{IgnoreNonExisting: true})
```

## Building and testing

The module only `include`s the shared makefiles under
[`build/makefile`](../../build/makefile). It is a library, so there is no image
to build (`make build` is a no-op). Tests are written with
[Ginkgo](https://onsi.github.io/ginkgo/)/Gomega and read the golden files under
`input/` and `expected/`.

| Target | Description |
|--------|-------------|
| `make test` | Run unit tests (`test-go`) in the standard golang container |
| `make lint` | Run `golangci-lint` (`lint-go`) in a linter container |
| `make validate` | Package import/replace checks (no build step) |
| `make build` | No-op — this is a library, not a deployable artifact |

Tests run inside a Docker `golang:<version>` image (version derived from
[`go.mod`](go.mod)). To iterate without Docker, `go test ./...` works locally.
