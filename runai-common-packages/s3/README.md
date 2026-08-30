# s3

`s3` is a small shared Go library in the Run.ai monorepo that provides a single,
process-wide S3-compatible object-storage client built on
[MinIO's Go SDK](https://github.com/minio/minio-go). It centralizes how services
connect to S3/MinIO — endpoint, credentials (static or IAM), and TLS — so every
consumer shares one configured client instead of wiring up MinIO directly. It is
consumed as a versioned Go module — there is no binary and no container image.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Importing the library](#importing-the-library)
- [Public API](#public-api)
- [Building and testing](#building-and-testing)
- [Configuration](#configuration)

## Overview

The package is responsible for:

- **Client initialization** — `Init(*S3Config)` builds a `*minio.Client` exactly
  once (guarded by `sync.Once`) and stores it as a package-global. When a config
  is passed it is applied; when `nil` is passed the configuration is loaded from
  the environment via the shared config package. If no access key is provided, the
  client uses **IAM role credentials**; otherwise it uses **static
  access-key/secret** credentials. `UseSSL` controls TLS. Initialization failures
  are logged (warning) rather than panicking.
- **Client access** — `Client()` returns the initialized `minio.Client`, or an
  error if `Init` has not run (or failed).
- **Config loading** — `InitConfig` registers the `S3Config` schema (including the
  Kubernetes secrets folder for the secret-tagged fields) so credentials can come
  from mounted secrets.

## Repository layout

```text
.
├── client.go      # Init (sync.Once), Client(), IAM vs static credential selection
├── config.go      # S3Config, SetDefaults, InitConfig (schema + secrets registration)
├── Makefile
└── go.mod
```

## Importing the library

The module path comes from [`go.mod`](go.mod). Add a `require` pinned to an
`s3/vX.Y.Z` tag:

```go
import "github.com/run-ai/runai/runai-common-packages/s3"

func main() {
    s3.Init(nil) // nil → load S3Config from env/secrets

    client, err := s3.Client()
    if err != nil {
        // client not initialized
    }
    _ = client // use the minio.Client
}
```

## Public API

- **`Init(cfg *S3Config)`** — initialize the global client (idempotent). Pass a
  struct to configure programmatically, or `nil` to load from the environment.
- **`Client() (minio.Client, error)`** — return the initialized client.
- **`InitConfig() error`** — register the `S3Config` schema and secrets folder.
- **`S3Config`** — the configuration struct.

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
| `make package-bump-version` | Tag a production release `s3/vX.Y.Z` |
| `make package-bump-version-dev` | Tag a dev release `s3/vX.Y.Z-dev.N` |

## Configuration

`S3Config` is populated programmatically or loaded from environment variables /
mounted secrets (the uppercase equivalents of the config keys). `AccessKey` is a
secret-tagged field, so it can be sourced from the mounted secrets folder:

| Variable | Description | Default |
|----------|-------------|---------|
| `ENDPOINT` | S3/MinIO endpoint (host:port) | — |
| `ACCESS_KEY_ID` | Access key id; if empty, IAM role credentials are used | — |
| `ACCESS_KEY` | Secret access key (secret-tagged) | — |
| `USE_SSL` | Connect over TLS | `false` |

Releases are cut by tagging the repo `s3/vX.Y.Z`; see the
[repository README](../README.md) for the versioning workflow.
