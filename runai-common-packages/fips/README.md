# fips

`fips` is a tiny, import-for-side-effect Go library in the Run.ai monorepo. Adding
it to a service's import graph logs the process's FIPS 140-3 cryptographic-module
status at startup, giving operators a single, consistent line confirming whether
the Go FIPS 140 module is enabled and enforced for that binary. It is consumed as
a versioned Go module — there is no binary and no container image.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Importing the library](#importing-the-library)
- [Building and testing](#building-and-testing)

## Overview

The package does exactly one thing, in a package-level `init()`:

- Initializes the shared [`logging`](../logging) package (`logging.Init(nil)`).
- Reads the Go runtime's FIPS 140 state via the standard-library
  [`crypto/fips140`](https://pkg.go.dev/crypto/fips140) package and logs it at
  `Info` level, for example:

  ```text
  FIPS 140-3 status: enabled=true enforced=false version="v1.0.0"
  ```

The reported values come from `fips140.Enabled()`, `fips140.Enforced()`, and
`fips140.Version()`. Whether FIPS is actually enabled/enforced is controlled by
how the binary is built and run (the `GOFIPS140` build setting — the shared Go
build targets in this monorepo compile with `GOFIPS140=v1.0.0`) and the
`GODEBUG=fips140=...` runtime setting, not by this package. `fips` only *observes
and reports* that state.

Because the behavior lives entirely in `init()`, the package is meant to be
imported for its side effect:

```go
import _ "github.com/run-ai/runai/runai-common-packages/fips"
```

## Repository layout

```text
.
├── fips.go      # init() — logs crypto/fips140 status via the logging package
├── Makefile
└── go.mod
```

## Importing the library

The module path comes from [`go.mod`](go.mod). Add a `require` pinned to a
`fips/vX.Y.Z` tag and take a blank import from `main` (or an early-loaded
package) so the status is logged as the process starts. The package's only
runtime dependency is the sibling [`logging`](../logging) module.

## Building and testing

This is a library, so `make build` and `make validate` are intentionally no-ops
(no binary, no image). Everything is driven through the [`Makefile`](Makefile),
which includes the shared `../../build/makefile/index.mk`.

| Target | Description |
|--------|-------------|
| `make test` | Run unit tests via the shared `test-go` target (runs `go test` inside the stock golang container) |
| `make lint` | Run the Go linter (golangci-lint) against `../.golangci.yaml` |
| `make purify` | `go fmt ./...`, `go vet ./...`, and `go mod tidy` |
| `make build` | No-op (library) |
| `make validate` | No-op recipe, but the shared makefile adds package import/replace checks as prerequisites |
| `make package-bump-version` | Tag a production release `fips/vX.Y.Z` |
| `make package-bump-version-dev` | Tag a dev release `fips/vX.Y.Z-dev.N` |

Releases are cut by tagging the repo `fips/vX.Y.Z`; see the
[repository README](../README.md) for the versioning workflow.
