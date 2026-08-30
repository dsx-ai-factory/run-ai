# org-unit-service-client

Generated Go client for the **org-unit-service** control-plane service. It wraps the `Org unit` HTTP
API ("An API for managing RunAi organizational units") so other Run:ai backend services can read and
manage organizational units — departments and projects (with v1 variants), node types, notifications
and reports. The package is owned by the same team that owns the
[org-unit-service](../../control-plane/org-unit-service) service; the generated code plus a
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

- **Generated OpenAPI client** ([`generated/`](generated)) — low-level, per-tag API structs and models
  produced by [OpenAPI Generator](https://openapi-generator.tech) from the service's OpenAPI spec
  (`DepartmentsAPI`, `DepartmentsV1API`, `ProjectsAPI`, `ProjectsV1API`, `OrgUnitAPI`, `NodeTypesAPI`,
  `NotificationsAPI`, `ReportsAPI`, `HealthAPI`). **This directory must not be edited by hand.**
- **Hand-written wrapper** — [`api.go`](api.go) declares the public client interface(s) and
  constructor, [`internal/client.go`](internal/client.go) implements them on top of the generated
  client, and [`structs/`](structs) holds hand-written request/response helper types.
  [`mocks/`](mocks) contains [mockery](https://github.com/vektra/mockery)-generated test doubles.

Consumers program against the wrapper interfaces, not the generated code directly.

## Repository layout

```text
.
├── api.go                  # Public client interface(s) + constructor (hand-written)
├── structs/
│   └── structs.go          # Hand-written helper types
├── internal/
│   └── client.go           # Interface implementation over the generated client
├── generated/              # OpenAPI-generated client + models (DO NOT EDIT)
├── mocks/                  # mockery-generated mocks of the client interfaces
├── artifact-metadata.json  # artifact_type: go
├── Makefile
├── go.mod                  # module github.com/run-ai/runai/runai-clients/org-unit-service-client
└── go.sum
```

## Importing the client

The module path is `github.com/run-ai/runai/runai-clients/org-unit-service-client`. Inside the
monorepo, consuming services require it at `v0.0.0` and point at the local copy with a `replace`
directive in their `go.mod`:

```go
require github.com/run-ai/runai/runai-clients/org-unit-service-client v0.0.0

replace github.com/run-ai/runai/runai-clients/org-unit-service-client => ../../runai-clients/org-unit-service-client
```

Then construct the client with the service base URL and call the interface methods.

## OpenAPI and code generation

The API contract lives with the **server**, not here:
[`control-plane/org-unit-service/swagger/source/openapi.yaml`](../../control-plane/org-unit-service/swagger/source/openapi.yaml).
The Go client in [`generated/`](generated) is produced by OpenAPI Generator (Go client generator,
version pinned in [`build/makefile/openapi.mk`](../../build/makefile/openapi.mk)) as part of the
**owning service's** build. Regenerate it from the service directory, not from here:

```bash
cd ../../control-plane/org-unit-service
make openapi-generate-code            # server stubs + Go client + JS client
# or just the Go client:
make openapi-generate-code-client-go
```

That target writes into `runai-clients/org-unit-service-client/generated`, runs `gofmt`, drops the
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
`runai-clients/org-unit-service-client/vX.Y.Z`:

| Target | Description |
|--------|-------------|
| `make client-bump-version` | Cut a production client tag (only from a synced `master`) |
| `make client-bump-version-dev` | Cut a `-dev` client tag from a feature branch |

Within the monorepo consumers use the `v0.0.0` + `replace` workspace pattern above, so a tag bump is
only needed when a version must be referenced outside the workspace. Do not merge a PR that requires a
`-dev` client version — CI rejects it.
