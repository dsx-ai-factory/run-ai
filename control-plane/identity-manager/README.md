# Identity Manager

Identity Manager is the control-plane service that owns identity and access
concerns for Run:AI: realms, identity providers, applications, service
accounts, access keys, users, and the OAuth2 / OIDC token endpoints used by the
rest of the platform. It is a thin, opinionated layer on top of
[Keycloak](https://www.keycloak.org/) (accessed via [gocloak](https://github.com/Nerzal/gocloak))
that exposes a stable HTTP/REST API generated from an OpenAPI specification.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [OpenAPI and code generation](#openapi-and-code-generation)
- [Go client](#go-client)
- [Further reading](#further-reading)

## Overview

The service is responsible for:

- **Realms** — creating, reconciling and deleting per-tenant Keycloak realms.
- **Identity providers** — configuring external IdPs (OIDC, SAML, OCP) and
  their attribute mappers.
- **Applications & service accounts** — managing OAuth2 clients, application
  tokens, and service-account credentials.
- **Users & access keys** — local user management, password lifecycle, and
  long-lived access keys.
- **Settings** — tenant-scoped security settings (auto-redirect SSO, browser
  session timeout, logout redirect URI, …).
- **Tokens** — issuing access tokens (v1 and v2 grants), application tokens,
  refresh, and logout.

It also runs **reconcilers** that bring Keycloak state in line with the desired
state stored in the service's own database (realms, clients, apps, internal
apps, user apps, workload auth).

The API surface is described in [`swagger/source/openapi.yaml`](swagger/source/openapi.yaml)
and bundled into [`swagger/_build/swagger.yaml`](swagger/_build/swagger.yaml).

## Repository layout

```
.
├── cmd/                       # Service entry point (main.go)
├── pkg/
│   ├── business/              # Domain logic
│   ├── client/                # Versioned Go client (see pkg/client/README.md)
│   │   └── generated/         # OpenAPI-generated client (do not edit)
│   ├── config/                # Configuration loading
│   ├── database/              # DAO layer and SQL migrations
│   ├── models/                # Domain models
│   ├── pubsub/                # Pub/sub integration
│   ├── realm/                 # Realm-related abstractions
│   ├── reconciler/            # Reconcilers (realms, clients, apps, …)
│   ├── router/                # HTTP router wiring
│   ├── services/              # API service implementations
│   ├── settings/              # Typed setting definitions
│   └── utils/
├── internal/
│   ├── auth/                  # Auth helpers
│   ├── cache/                 # In-process caches
│   ├── generated/             # OpenAPI-generated server stubs (do not edit)
│   ├── iam/                   # IAM (Keycloak) integration
│   │   └── gocloak/           # Generated gocloak interface (see its README)
│   └── notification/
├── integration-tests/         # Docker-compose-driven integration tests
├── swagger/                   # OpenAPI source and built bundle
├── Dockerfile
├── Makefile
└── go.mod
```

## Building and running

The project is built with Go (see [`go.mod`](go.mod) for the required version)
and orchestrated through the top-level [`Makefile`](Makefile), which inherits
from the shared `infra/makefile/index.mk`.

Common targets:

| Target                | Description                                         |
| --------------------- | --------------------------------------------------- |
| `make build`          | Validate OpenAPI, generate code, build the binary   |
| `make run-local`      | Build and run locally via `docker-compose`          |
| `make docker-build`   | Build the service container image                   |
| `make docker-push`    | Push the container image to the registry            |
| `make lint`           | Run the Go linter                                   |
| `make purify`         | Auto-fix formatting and imports                     |
| `make validate`       | Validate the OpenAPI specification                  |
| `make all`            | `validate` + `lint` + `build` + `test`              |

The runtime binary is produced as `bin/identity-manager-<arch>` and packaged
into the image defined by [`Dockerfile`](Dockerfile).

## Testing

| Target                                       | Description                              |
| -------------------------------------------- | ---------------------------------------- |
| `make unit-test`                             | Build and run unit tests                 |
| `make test`                                  | Unit tests + integration tests           |
| `make test-integration-iam-pkg`              | Integration tests for the IAM package    |
| `make docker-compose-integration-tests-up`   | Run the full docker-compose test suite   |
| `make docker-compose-integration-tests-debug`| Same, but without the failure handler    |
| `make coverage`                              | Open `c.out` as an HTML coverage report  |

## OpenAPI and code generation

The HTTP API is defined in [`swagger/source/openapi.yaml`](swagger/source/openapi.yaml)
and split into per-resource files under `swagger/source/resources` and shared
schemas under `swagger/source/schemas`. From this source the build generates:

- The **server stubs** in [`internal/generated`](internal/generated) (consumed by
  `cmd/main.go` and the routers).
- The **Go client** in [`pkg/client/generated`](pkg/client/generated), wrapped
  by hand-written helpers under [`pkg/client`](pkg/client).

Both `generated/` directories are produced by
[OpenAPI Generator](https://openapi-generator.tech) (version pinned in the
`Makefile`) and **must not be edited by hand** — changes will be overwritten on
the next generation run.

## Go client

Other Run:AI services consume Identity Manager through the versioned Go client
at [`pkg/client`](pkg/client). See [`pkg/client/README.md`](pkg/client/README.md)
for:

- How to import the versioned client package
- The `AuthnManager` (OAuth2 client-credentials, auto-refresh, retry on 401)
- The `make client-bump-version` and `make client-bump-version-dev` workflow

## Further reading

- [`pkg/client/README.md`](pkg/client/README.md) — client usage and versioning policy
- [`internal/iam/gocloak/README.md`](internal/iam/gocloak/README.md) — how the gocloak interface is regenerated
- [`pkg/client/generated/README.md`](pkg/client/generated/README.md) — generated client reference (auto-generated)
- [`internal/generated/README.md`](internal/generated/README.md) — generated server reference (auto-generated)
