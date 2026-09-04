# versions

`versions` is a small Go library in the Run:ai `runai-application-packages` tree
that answers a single question consistently across the platform: **is a given
version at least the minimum supported version?** It centralizes semantic
version comparison — including the special handling of `0.0.0`-prefixed
development/master builds — so services agree on compatibility checks. It is a
**library package**: no binary, no Dockerfile, consumed by other modules through
`go.mod`.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Public API](#public-api)
- [Importing](#importing)
- [Building and testing](#building-and-testing)

## Overview

The package (`package versions`) wraps `hashicorp/go-version` semver parsing and
adds Run:ai's conventions:

- **`IsVersionSupported(version, minSupportedVersion)`** — parses both operands
  as semver and returns whether `version >= minSupportedVersion`. Accepts
  `X.Y`, `X.Y.Z`, and `0.0.0-XXXX` forms.
- **`AreAllVersionsSupported(versions, minSupportedVersion)`** — the same check
  applied to a slice, short-circuiting on the first unsupported or unparseable
  entry.
- **Dev-version shortcut** — any version starting with the `0.0.0` prefix
  (`DEV_VERSION_PREFIX`, used for master/dev builds) is always treated as
  supported, so local and pre-release builds never fail the gate.

Parse failures return an error rather than silently passing, so callers can
distinguish "too old" from "unparseable".

## Repository layout

```text
.
├── versions.go            # IsVersionSupported, AreAllVersionsSupported, DEV_VERSION_PREFIX
├── artifact-metadata.json
├── Makefile
└── go.mod
```

## Public API

| Symbol | Description |
|--------|-------------|
| `DEV_VERSION_PREFIX` (`"0.0.0"`) | Prefix marking dev/master builds that always pass |
| `IsVersionSupported(version, minSupportedVersion string) (bool, error)` | Is a single version ≥ the minimum |
| `AreAllVersionsSupported(versions []string, minSupportedVersion string) (bool, error)` | Are all versions ≥ the minimum |

## Importing

```go
import "github.com/dsx-ai-factory/run-ai/runai-application-packages/versions"

ok, err := versions.IsVersionSupported("2.19.1", "2.18.0")
allOk, err := versions.AreAllVersionsSupported([]string{"2.19", "0.0.0-abc"}, "2.18.0")
```

## Building and testing

The module only `include`s the shared makefiles under
[`build/makefile`](../../build/makefile); it is a library, so there is no image
to build (`make build` is a no-op) and it is consumed purely via `go.mod`. Its
only dependency is `hashicorp/go-version`.

| Target | Description |
|--------|-------------|
| `make test` | Run unit tests (`test-go`) in the standard golang container |
| `make lint` | Run `golangci-lint` (`lint-go`) in a linter container |
| `make validate` | Package import/replace checks (no build step) |
| `make build` | No-op — this is a library, not a deployable artifact |

Tests run inside a Docker `golang:<version>` image (version derived from
[`go.mod`](go.mod)). To iterate without Docker, `go test ./...` works locally.
