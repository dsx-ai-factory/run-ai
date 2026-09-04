# logging

`logging` is the shared structured-logging library for the Run.ai monorepo. It is a
thin, opinionated wrapper around [zerolog](https://github.com/rs/zerolog) that gives
every service the same global logger, log format, and context-enrichment behavior.
It exposes package-level `Debugf`/`Infof`/`Warnf`/`Errorf`/`Fatalf` helpers, a
pluggable `Logger` interface, request-context enrichment, and an HTTP
request-logging middleware. It is consumed as a versioned Go module — there is no
binary and no container image.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Importing the library](#importing-the-library)
- [Public API](#public-api)
- [Building and testing](#building-and-testing)
- [Configuration](#configuration)

## Overview

The package is responsible for:

- **Global logger** — `Init(*Config)` builds a zerolog logger (JSON by default,
  or a human-readable console writer when pretty mode is on) and stores it as the
  process-wide `globalLogger`. Passing `nil` loads configuration from the
  environment. The package-level `Debugf`/`Infof`/`Warnf`/`Errorf`/`Fatalf`
  functions log through it (and are no-ops until `Init` runs, except `Fatalf`
  which still exits).
- **Log pattern field** — every message records its format string under a
  `log_pattern` field, so log lines can be grouped/aggregated by template even
  when their interpolated values differ.
- **Context enrichment** — `WithCtx(CtxLoggerHook)` returns a child logger that
  copies selected `context.Context` values into every log line via a zerolog
  hook. `CtxLoggerHook` supplies the context-key → log-field mapping (the
  [`runaictx`](../runaictx) package provides a ready-made hook).
- **HTTP middleware** — the `logging/http` sub-package's `HttpMiddleware` wraps an
  `http.Handler`, logging method, path, user-agent, status code and duration for
  each request (and the response body for `>= 400` responses), enriched with the
  request context.
- **Test helpers** — `GetDebugHookLogger` and `GetDebugWriter` return a logger
  plus a hook/writer that capture emitted events for assertions in unit tests.

## Repository layout

```text
.
├── global_functions.go   # Init + package-level Debugf/Infof/Warnf/Errorf/Fatalf, WithCtx
├── logger.go             # Logger interface, ConsoleLogger, log_pattern field
├── hooks.go              # ctx-enrichment hook adapter + debug test hooks/writers
├── config.go             # Config (log_level, log_pretty) + defaults
├── http/
│   └── middleware.go     # HttpMiddleware — per-request logging
├── Makefile
└── go.mod
```

## Importing the library

The module path comes from [`go.mod`](go.mod). Add a `require` pinned to a
`logging/vX.Y.Z` tag:

```go
import "github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"

func main() {
    logging.Init(nil)           // load config from env
    logging.Infof("starting %s", name)
}
```

Enrich logs with request context:

```go
import "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("handled request")
```

## Public API

- **Global:** `Init`, `Debugf`, `Infof`, `Warnf`, `Errorf`, `Fatalf`, `Actual`,
  `WithCtx`.
- **Types:** `Logger` (interface), `ConsoleLogger`, `Config`, `CtxLoggerHook`.
- **HTTP:** `http.HttpMiddleware`.
- **Testing:** `GetDebugHookLogger`, `GetDebugWriter`, `NewLogDebugHook`.

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
| `make package-bump-version` | Tag a production release `logging/vX.Y.Z` |
| `make package-bump-version-dev` | Tag a dev release `logging/vX.Y.Z-dev.N` |

(There is no `purify` target in this package's Makefile.)

## Configuration

`Config` is loaded from environment variables (uppercase equivalents of the config
keys) or passed directly to `Init`:

| Variable | Description | Default |
|----------|-------------|---------|
| `LOG_LEVEL` | zerolog level name (`Debug`, `Info`, `Warn`, `Error`, …); invalid values fall back to `Info` | `Info` |
| `LOG_PRETTY` | `true` for human-readable console output; otherwise structured JSON | `false` |

Releases are cut by tagging the repo `logging/vX.Y.Z`; see the
[repository README](../README.md) for the versioning workflow.
