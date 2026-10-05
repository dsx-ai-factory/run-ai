# Notifications Proxy

Notifications Proxy is the control-plane service that sits between the Run:ai
frontend and the internal `notifications-service`, exposing a stable, versioned
HTTP/REST API for **notification configuration**: notification types,
per-user/per-tenant subscriptions, notification channels (email and Slack),
notification state, and the Slack OAuth "app" installation flow. It is owned by
the **Phoenix** team. Where the frontend needs to configure how and where users
are notified, it talks to this proxy; the proxy validates and persists that
configuration (in its own PostgreSQL schema) and forwards the parts it does not
own to the upstream `notifications-service` config endpoint.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [OpenAPI and code generation](#openapi-and-code-generation)
- [Go client](#go-client)
- [Further reading](#further-reading)

## Overview

The service is responsible for:

- **Notification types** — exposing the catalogue of notification kinds the
  platform can emit (`GET /api/v1/notification-types`).
- **Notification state** — reading the enabled/disabled state of a user's
  notifications, including a `detailed` view (`/api/v1/notification-state`,
  `/api/v1/notification-state/detailed`).
- **Email channels** — configuring, validating and inspecting the email
  notification channel and its state (`/api/v1/email`, `/api/v1/email/state`,
  `/api/v1/email/validate`).
- **Slack channels** — the full Slack integration surface: channel config and
  state (`/api/v1/slack`, `/api/v1/slack/state`, `/api/v1/slack/state/detailed`),
  workspace/app installation and OAuth (`/api/v1/slack/app`,
  `/api/v1/slack/workspace`, `/api/v1/slack/auth`, `/api/v1/slack/validate`,
  and the `notification-channels/slack/{auth-app,add-app,create-app}` routes).
- **Notification channels** — generic channel CRUD and validation
  (`/api/v1/notification-channels`, `/api/v1/notification-channels/{name}`,
  `/api/v1/validate-notification-channel`).
- **Subscriptions** — user/tenant subscriptions to notification types
  (`/api/v1/subscriptions`, `/api/v1/subscriptions/{id}`).
- **Health** — an unauthenticated `GET /` health check.

It also **listens on pub/sub** for tenant-deletion events (via
`internal/events_listeners`) to clean up channel/subscription state, initializes
a **notify client** to the upstream service, enforces access control through the
shared authorization client, and supports **multi-tenant** and single-tenant
modes. Sensitive values (DB credentials, Slack client id/secret) are read from
mounted secret files rather than plain env vars.

The API surface is described in
[`swagger/source/openapi.yaml`](swagger/source/openapi.yaml) and bundled into
[`swagger/_build/swagger.yaml`](swagger/_build/swagger.yaml).

## Repository layout

```text
.
├── cmd/                         # Service entry point (main.go)
├── internal/
│   ├── authorization/           # Access-control authorizer (single/multi-tenant)
│   ├── channel/                 # Notification channel manager
│   ├── clients/                 # HTTP clients to upstream notifications-service
│   ├── config/                  # Service + Slack configuration structs
│   ├── database/                # DAO layer, models, and SQL migrations
│   ├── email/                   # Email channel manager
│   ├── events_listeners/        # Pub/sub listeners (tenant delete)
│   ├── generated/               # OpenAPI-generated server stubs (do not edit)
│   ├── notification_state/      # Notification-state manager
│   ├── notification_types/      # Notification-types manager
│   ├── router/                  # HTTP router wiring (chi)
│   ├── services/                # API service implementations (one per resource)
│   ├── slack/                   # Slack manager, OAuth session handling, config
│   └── subscription/            # Subscription manager
├── pkg/
│   └── client/                  # Versioned Go client (see pkg/client/README.md)
│       └── generated/           # OpenAPI-generated client (do not edit)
├── integration_tests/           # Docker-compose-driven integration tests
│   ├── mailhog/                 # MailHog SMTP fixture + generated client
│   ├── slack/                   # Slack mock fixture
│   ├── config/                  # Backing-service config for the compose stack
│   └── secrets/                 # Test secret files (DB, Slack)
├── swagger/                     # OpenAPI source and built bundle
├── Dockerfile
├── Makefile
└── go.mod
```

## Building and running

The project is built with Go (see [`go.mod`](go.mod) for the required version)
and orchestrated through the [`Makefile`](Makefile), which includes the shared
`../../build/makefile/control-plane.mk` and `index.mk` (which pull in the shared
`golang.mk` and `openapi.mk`).

| Target | Description |
|--------|-------------|
| `make build` | `openapi-build` then `build-go` — build the OpenAPI bundle, then compile the binary for amd64 + arm64 |
| `make build-go` | `build-go-amd` + `build-go-arm` — compile both arch binaries |
| `make build-go-amd` | Compile `./cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) in the Go builder container → `bin/notifications-proxy-amd64` |
| `make build-go-arm` | Same for `arm64` → `bin/notifications-proxy-arm64` |
| `make openapi-build` | Build the swagger bundle from `swagger/source/` |
| `make openapi-lint` | Lint the swagger bundle |
| `make validate` | `openapi-validate` — validate the OpenAPI spec |
| `make lint` | `lint-go` — golangci-lint in the builder container |
| `make purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `make generate-mocks` / `make generate-mocks-local` | Regenerate test mocks from `.mockery.yml` with mockery v3 (shared targets; in Docker / without Docker) |
| `make generate-integration-test-smtp-openapi-client` | Regenerate the MailHog SMTP test client from its OpenAPI spec |
| `make local-deploy` | `openapi-generate-code` + `build-go` + `docker-build-generic` + `docker-push-generic` (linux/amd64) |

### Local build (local-dev.sh)

`notifications-proxy` is listed in `CP_GO_SERVICES`, so `local-dev.sh` builds it
via `build_go_service` in two steps:

```bash
# 1. Compile the amd64 binary inside the Go builder container
make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=

# 2. Build + load the image into the local store / Kind nodes
docker buildx build --platform linux/amd64 --provenance=false --load \
  --build-arg TARGETARCH=amd64 -t runai-local/notifications-proxy:0.0.0 .
```

`DOCKERHUB_PROXY=` pulls the golang builder image straight from Docker Hub;
`GITHUB_TOKEN=` is emptied in the make call (the private-module token is applied
elsewhere). The image tag **must** be `0.0.0` — the control-plane chart's
appVersion — or the local deploy will not find the image.

For a fast control-plane-only iteration loop:

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```

The [`Dockerfile`](Dockerfile) is a two-stage build: a
`registry.access.redhat.com/ubi9/ubi-minimal` `base` stage supplies
`/etc/passwd`, `/etc/group`, TLS certs (`/etc/ssl`, `/etc/pki`) and a writable
`/tmp`; the final image is `FROM scratch`, copies in the
`notifications-proxy-$TARGETARCH` binary, the SQL migrations
(`internal/database/migrations/up|down`), the swagger bundle and `NOTICE`, and
runs as user `65532:65532`. `TARGETARCH` is the only build arg.

## Testing

| Target | Description |
|--------|-------------|
| `make test` | `build` + `test-go` + `integration-tests` |
| `make test-go` | Run the Go unit tests in the builder container |
| `make build-notifications-service-bin` | Compile the sibling `notifications-service` binary so the compose stack builds its image locally |
| `make build-integration-tests` | Build the `-tags integration` test binary and `docker compose build` the stack |
| `make integration-tests` | Bring the full compose stack up (all profiles) and abort on container exit |
| `make integration-tests-debug` | Bring up the compose stack (setup/smtp/slack/redis profiles) without the test container, to run tests from an IDE |

The integration suite needs docker-compose and its backing fixtures:
PostgreSQL, Redis (pub/sub), the sibling `notifications-service`, a **MailHog**
SMTP server, and a **Slack mock**.

## Configuration

The service is env-var driven. Each configuration key has an uppercase env-var
equivalent; sensitive values are read from files under `SECRETS_MOUNT_PATH`
rather than from env vars. Values shown below are taken from the integration
docker-compose stack and the service's config defaults.

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_PORT` | HTTP port to listen on | `8080` |
| `SERVICE_NAME` | Service identifier used in logging/metrics | — |
| `LOG_LEVEL` | Log verbosity (`debug`, `info`, …) | `info` |
| `RUNAI_ENV_DOMAIN` | Public Run:ai environment domain (**required**) | — |
| `MULTI_TENANCY_ENABLED` | Enable multi-tenant authorization mode | — |
| `NOTIFY_CONFIG_URL` | Upstream `notifications-service` config endpoint | `http://runai-backend-notifications-service:8080` |
| `SLACK_BASE_URL` | Slack API base URL (mockable in tests) | `https://slack.com` |
| `SECRETS_MOUNT_PATH` | Directory of mounted secret files | `/etc/secrets` |
| `USE_MOCKS` | Use the mock authorization client (tests only) | `false` |
| `DB_HOST` | PostgreSQL host | — |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_NAME` | PostgreSQL database name | — |
| `DB_SCHEMA` | PostgreSQL schema | `notifications_proxy` |
| `DB_SSL_MODE` | PostgreSQL SSL mode | — |
| `PUBSUB_TYPE` | Pub/sub backend type (e.g. `redis`) | — |
| `PUBSUB_URL` | Pub/sub connection URL | — |

**Secret files** (read from `SECRETS_MOUNT_PATH`, one value per file):
`DB_USER`, `DB_PASSWORD`, `DB_ENCRYPTION_KEY`, `SLACK_CLIENT_ID`,
`SLACK_CLIENT_SECRET`. See `integration_tests/secrets/` for the test fixtures
and `integration_tests/docker-compose.yaml` for the full runtime environment.

## OpenAPI and code generation

The HTTP API is defined in
[`swagger/source/openapi.yaml`](swagger/source/openapi.yaml), split into
`parameters/`, `requests/` and `schemas/` fragments. From this source the build
generates:

- The **server stubs** in [`internal/generated`](internal/generated) — consumed
  by [`internal/router/router.go`](internal/router/router.go).
- The **Go client** in [`pkg/client/generated`](pkg/client/generated), wrapped
  by hand-written helpers in [`pkg/client`](pkg/client).

Both are produced by [OpenAPI Generator](https://openapi-generator.tech)
(`v7.12.0`, `VERSIONED_CLIENT=true`) and **must not be edited by hand** — changes
are overwritten on the next generation run. The `make openapi-build` /
`openapi-validate` targets build and validate the bundle;
`make local-deploy` runs `openapi-generate-code` first.

## Go client

Other Run:ai services consume Notifications Proxy through the versioned Go client
at [`pkg/client`](pkg/client). See [`pkg/client/README.md`](pkg/client/README.md)
for how to import the versioned client and the
`make client-bump-version` / `make client-bump-version-dev` workflow.

## Further reading

- [`pkg/client/README.md`](pkg/client/README.md) — client usage and versioning
- [`internal/generated/README.md`](internal/generated/README.md) — generated server reference (auto-generated)
</content>
</invoke>
