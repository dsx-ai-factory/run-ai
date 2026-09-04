# runaihttp

`runaihttp` is the shared HTTP-server and response-helper library for the Run.ai
monorepo. It provides the boilerplate that generated OpenAPI servers and
control-plane services need on top of the [chi](https://github.com/go-chi/chi)
router: a configurable HTTP server with graceful shutdown and a built-in
health-check mode, PostgreSQL-error-to-HTTP-status mapping, request field
validation helpers, and typed response writers (plain text, HTML, file download,
CSV). It is consumed as a versioned Go module — there is no binary and no container
image.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Importing the library](#importing-the-library)
- [Public API](#public-api)
- [Building and testing](#building-and-testing)
- [Configuration](#configuration)

## Overview

The package is responsible for:

- **HTTP server** — `ServeHTTP(chi.Router, *ServerConfig)` starts an
  `http.Server` with sane read/write/idle timeouts, binds to the configured
  address, and shuts down gracefully on `SIGINT` within `ShutdownTimeout`. An
  `init()` also handles a `--health` CLI flag: when set, the process performs an
  HTTP health check against the configured host/port/endpoint and exits `0`/`1`
  accordingly (used as a container health probe). `ServerConfig` loads from
  environment when `nil` is passed.
- **Error mapping** — `LoadErrorDetails` inspects an error and, for PostgreSQL
  driver errors (`pgconn.PgError`), maps SQL states to HTTP status codes:
  duplicate-key → `409`, invalid-syntax → `400`, foreign-key → `412`,
  trigger-raised → `409`, and GORM "record not found" → `404`. It deliberately
  avoids leaking raw Postgres detail for unmapped errors.
- **Validation helpers** — `AssertMinItems`/`AssertMaxItems`,
  `AssertMinLength`/`AssertMaxLength`, and `AssertPattern` implement the
  OpenAPI-style field constraints used by generated request validators.
- **Response writers** — `PlainTextResponse`, `HtmlResponse`, `TextResponse`,
  `DownloadFileResponse`, and `CreateCsvResponse`/`CsvFile` write non-JSON
  responses with the right `Content-Type`/`Content-Disposition` headers.

## Repository layout

```text
.
├── server.go            # ServeHTTP, ServerConfig, --health probe, graceful shutdown
├── error_handling.go    # LoadErrorDetails — Postgres error → HTTP status
├── items.go             # AssertMinItems / AssertMaxItems
├── length.go            # AssertMinLength / AssertMaxLength
├── pattern.go           # AssertPattern (+ generic getValuePtr)
├── csv_support.go       # PlainText/Html/Text/DownloadFile/Csv response writers
├── Makefile
└── go.mod
```

## Importing the library

The module path comes from [`go.mod`](go.mod). Add a `require` pinned to a
`runaihttp/vX.Y.Z` tag:

```go
import (
    "github.com/go-chi/chi/v5"
    "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"
)

func main() {
    r := chi.NewRouter()
    // … register routes …
    runaihttp.ServeHTTP(r, nil) // nil → load ServerConfig from env, blocks until shutdown
}
```

## Public API

- **Server:** `ServeHTTP`, `ServerConfig`.
- **Errors:** `LoadErrorDetails` and the exported SQL-state constants
  (`ErrForeignKey`, `ErrDuplicateKey`, `ErrInvalidSyntax`, `ErrTriggerFailed`,
  `ErrTextNotFound`).
- **Validation:** `AssertMinItems`, `AssertMaxItems`, `AssertMinLength`,
  `AssertMaxLength`, `AssertPattern`.
- **Responses:** `PlainTextResponse`, `HtmlResponse`, `TextResponse`,
  `DownloadFileResponse`, `CreateCsvResponse`, `CsvFile`.

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
| `make package-bump-version` | Tag a production release `runaihttp/vX.Y.Z` |
| `make package-bump-version-dev` | Tag a dev release `runaihttp/vX.Y.Z-dev.N` |

## Configuration

`ServerConfig` is loaded from environment variables (uppercase equivalents of the
config keys) or passed directly to `ServeHTTP`. All fields have defaults:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_PORT` | Port to listen on | `8080` |
| `SERVER_IP_ADDRESS` | Bind address | `0.0.0.0` |
| `SERVER_WRITE_TIMEOUT` | Max response-write duration | `15s` |
| `SERVER_READ_TIMEOUT` | Max request-read duration | `15s` |
| `SERVER_IDLE_TIMEOUT` | Keep-alive idle timeout | `60s` |
| `SERVER_SHUTDOWN_TIMEOUT` | Graceful-shutdown deadline | `15s` |
| `SERVER_HEALTH_HOST` | Host used by the `--health` probe | `localhost` |
| `SERVER_HEALTH_ENDPOINT` | Path used by the `--health` probe | `/` |

Releases are cut by tagging the repo `runaihttp/vX.Y.Z`; see the
[repository README](../README.md) for the versioning workflow.
