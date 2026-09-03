# runaictx

`runaictx` is the shared request-context library for the Run.ai monorepo. It
defines the canonical set of Run.ai context fields (tenant id, subject id, groups,
trace/span ids, client id, …) and the helpers that move them between a Go
`context.Context`, HTTP headers, structured log fields, and plain maps. It gives
every service one consistent vocabulary for propagating identity and tracing
information across service boundaries. It is consumed as a versioned Go module —
there is no binary and no container image.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Importing the library](#importing-the-library)
- [Public API](#public-api)
- [Building and testing](#building-and-testing)

## Overview

The package is responsible for:

- **Context field vocabulary** — the `ContextField` type and its constants
  (`CfTenantId`, `CfSubjectID`, `CfSubjectType`, `CfGroups`,
  `CfSupplementalGroups`, `CfClientId`, `CfClientType`, `CfTraceId`, `CfSpanId`,
  `CfApplicationId`, `CfIsManagementTenant`, `CfAccept`, `CfGId`, `CfUId`,
  `CfSId`, …) used as `context.Context` keys everywhere in the platform.
- **Accessors** — `GetField`, `GetString`, and typed extractors:
  `ParseTenantId`, `ParseSubjectId`, `ParseGroups`, `ParseAuthData`, and
  `IsManagementTenant`.
- **Serialization** — `ToMap`/`FromMap` convert the well-known fields to and from
  a `map[string]interface{}` (using stable snake_case identifiers such as
  `tenant_id`, `subject_id`, `trace_id`), used when passing context across
  boundaries that aren't HTTP (e.g. pub/sub messages).
- **HTTP propagation** (`runaictx/http` sub-package):
  - `ContextMiddleware` — server middleware that reads the `X-Runai-*` request
    headers (tenant, subject, groups, client, trace/span, …) into the context and
    echoes the trace id back on the response.
  - `SetHeaders`/`GetHeaders` — populate outgoing `http.Header`s from the context
    so downstream calls carry the same identity.
- **Log enrichment** — `LoggerHook(ctx)` returns a hook consumable by the
  [`logging`](../logging) package's `WithCtx`, so the context fields appear as
  structured log fields.

## Repository layout

```text
.
├── context.go            # ContextField constants, GetField/GetString, ToMap/FromMap
├── context_actions.go    # ctx-field ↔ log/map identifier mapping + value coercion
├── extractors.go         # ParseTenantId/SubjectId/Groups/AuthData, IsManagementTenant
├── logger_hook.go        # LoggerHook — bridge to logging.WithCtx
├── http/
│   ├── middleware.go     # ContextMiddleware — X-Runai-* headers ↔ context
│   └── headers.go        # SetHeaders/GetHeaders — context → outgoing headers
├── Makefile
└── go.mod
```

## Importing the library

The module path comes from [`go.mod`](go.mod). Add a `require` pinned to a
`runaictx/vX.Y.Z` tag:

```go
import (
    "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
    runaictxhttp "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
)

// Server: extract X-Runai-* headers into the request context
router.Use(runaictxhttp.ContextMiddleware)

// Anywhere: read identity
tenantID, err := runaictx.ParseTenantId(ctx)
```

## Public API

- **Context:** `ContextField` + constants, `GetField`, `GetString`, `ToMap`,
  `FromMap`.
- **Extractors:** `IsManagementTenant`, `ParseTenantId`, `ParseSubjectId`,
  `ParseGroups`, `ParseAuthData`.
- **Logging bridge:** `LoggerHook`.
- **HTTP:** `http.ContextMiddleware`, `http.SetHeaders`, `http.GetHeaders`.

## Building and testing

This is a library, so `make build` and `make validate` are intentionally no-ops
(no binary, no image). Everything is driven through the [`Makefile`](Makefile),
which includes the shared `../../build/makefile/index.mk`.

| Target | Description |
|--------|-------------|
| `make test` | Run unit tests via the shared `test-go` target (runs `go test` inside the stock golang container) |
| `make lint` | Run the Go linter (golangci-lint) against `../.golangci.yaml` |
| `make build` | No-op (library) |
| `make validate` | No-op recipe, but the shared makefile adds package import/replace checks as prerequisites |
| `make package-bump-version` | Tag a production release `runaictx/vX.Y.Z` |
| `make package-bump-version-dev` | Tag a dev release `runaictx/vX.Y.Z-dev.N` |

(There is no `purify` target in this package's Makefile.)

Releases are cut by tagging the repo `runaictx/vX.Y.Z`; see the
[repository README](../README.md) for the versioning workflow.
