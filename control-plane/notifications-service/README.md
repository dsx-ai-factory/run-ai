# notifications-service

`notifications-service` is the control-plane service that turns platform events into delivered notifications. It ingests events from several sources (an HTTP events API, inbound SMTP, Kubernetes events, and Prometheus/Alertmanager alerts), routes them through configurable filters, enriches them, renders them with per-channel templates, and dispatches them to **sinks** — email (SMTP), Slack, and a Postgres-backed notification **history**. It is owned by the **Phoenix** team. Configuration is file-driven (YAML) with two providers — `basic` and `advanced` (the advanced provider additionally exposes an HTTP API to manage listener/route/sink config at runtime).

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

- **Listening for events** — pluggable listeners under `internal/listen/`: an HTTP events endpoint (`http`), an inbound SMTP server (`smtp`), a Kubernetes events watcher (`kevent`), and a Prometheus alerts receiver (`prom`).
- **Routing and filtering** — `internal/route/` and `internal/filter/` match events against configured routes (matchers on labels/reasons/namespaces/objects) and decide which sinks receive them.
- **Enriching** — `internal/enrich/` augments events with additional metadata before rendering.
- **Rendering per channel** — `sinks_templates/` holds Go templates per sink type (`email/`, `slack/`) for generic notifications, workloads alerts, and Keycloak welcome/reset emails.
- **Dispatching to sinks** — `internal/sinks/`: `email` (over SMTP), `slack`, `postgres`, and `history`. HTML email bodies are sanitized (`bluemonday`) and can be converted from HTML to Markdown for Slack.
- **Persisting history** — a Postgres-backed store (`internal/postgres/`, `internal/history/`) with its own HTTP API to query past notifications.
- **Runtime config management** — the `advanced` provider (`internal/config/advanced/`) serves an OpenAPI-defined HTTP API to create/update listeners, routes, and sinks without a restart.

## Repository layout

```text
.
├── cmd/
│   └── main.go             # Entry point: selects config provider, builds app, starts listeners
├── internal/
│   ├── app/                # App wiring and lifecycle (Start/Stop, event routers)
│   ├── config/             # Config providers
│   │   ├── basic/          # File-only config provider
│   │   ├── advanced/       # File + HTTP-managed config (has its own OpenAPI server + docs)
│   │   └── service/        # Aggregated config loader
│   ├── listen/             # Event listeners: http, smtp, kevent, prom
│   ├── route/              # Event → sink routing
│   ├── filter/             # Route matchers / filtering
│   ├── enrich/             # Event enrichment
│   ├── events/             # Event model
│   ├── notify/             # Notification building/dispatch
│   ├── sinks/              # Sink implementations: email, slack, postgres, history, smtp
│   ├── history/            # Notification history store + OpenAPI server
│   ├── postgres/           # Postgres access layer
│   └── util/
├── pkg/
│   └── client/             # Public Go client, standalone versioned module
│       ├── generated/      # generated notify/config/history clients
│       └── types/          # shared exported types
├── sinks_templates/        # Per-sink Go templates (email/, slack/)
├── integration_tests/      # Docker-compose e2e tests (service + postgres + mailhog SMTP + slack)
├── bin/                    # Build output (notifications-service-<arch>, mockery)
├── .mockery.yml            # Mock generation config (mockery v3)
├── Dockerfile
├── Makefile
└── go.mod
```

## Building and running

Built with Go (see [`go.mod`](go.mod) — module `github.com/dsx-ai-factory/run-ai/control-plane/notifications-service`, Go 1.26) and orchestrated through the [`Makefile`](Makefile), which includes `../../build/makefile/control-plane.mk` and `index.mk`. **`make build` regenerates the OpenAPI server/client code first, then compiles both architectures** — the generated directories are a build product, not hand-written.

| Target | Description |
|--------|-------------|
| `make build` | Runs all `generate-*-openapi-*` targets + `build-integration-tests`, then builds `bin/notifications-service-{arm64,amd64}` (`build_arm` + `build_amd`). |
| `make build_amd` | Generates OpenAPI code, then `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build` → `bin/notifications-service-amd64`. |
| `make build_arm` | Same for `GOOS=linux GOARCH=arm64` → `bin/notifications-service-arm64`. |
| `make lint` | Runs the Go linter (`lint-go`). |
| `make openapi-lint` | Lints the OpenAPI bundle (`openapi-lint-generic`). |
| `make validate` | No-op (kept for control-plane target parity). |
| `make purify` | Auto-fix formatting and imports (`purify-go`). |
| `make generate-mocks-local` / `make generate-mocks` | (Re)generate mocks per `.mockery.yml` with the shared mockery v3 targets (locally / in Docker); `make generate-mocks-client` does the same for `pkg/client/.mockery.yml`. |
| `make local-deploy` | `openapi-generate-code build-go docker-build-generic docker-push-generic` (amd64) — build and push the image. |

### Local build (local-dev.sh)

`notifications-service` is a **control-plane Go service** (listed in `CP_GO_SERVICES` in [`local-dev.sh`](../../local-dev.sh)). `local-dev.sh` builds amd64 only (Kind nodes are linux/amd64) via the shared `build-go-amd` target:

```bash
# amd64 — what local-dev.sh runs
make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
docker build --build-arg TARGETARCH=amd64 -t runai-local/notifications-service:0.0.0 .
```

For an arm64 cluster, build the arm64 path manually (`build-go-arm` **hardcodes** `GOOS=linux GOARCH=arm64` — do **not** pass `OS=`/`ARCH=`):

```bash
# arm64 — manual
make build-go-arm DOCKERHUB_PROXY= GITHUB_TOKEN=
docker build --build-arg TARGETARCH=arm64 -t runai-local/notifications-service:0.0.0 .
```

- The image tag is **`0.0.0`** (the CP `appVersion`); any other tag breaks the local deploy. Image ref: `runai-local/notifications-service:0.0.0`.
- `DOCKERHUB_PROXY=` (empty) pulls the golang builder from Docker Hub; `GITHUB_TOKEN=` is empty in the make call (the private-module token is applied elsewhere).
- Same tag for both arches — only `--build-arg TARGETARCH` changes.
- **Note:** the shared `build-go-amd` compiles the binary but does not run OpenAPI generation; the service's own `make build` (`build_amd`/`build_arm`) is the target that regenerates the OpenAPI code first. Run `make build` (or the relevant `generate-*` targets) if the specs changed.

### Dockerfile

Multi-stage with two selectable final targets, both non-root (`USER 65532:65532`, `WORKDIR /app`, `ENTRYPOINT ["./notifications-service"]`):

- **`base`** — `registry.access.redhat.com/ubi9/ubi-minimal`; sources `/etc/passwd`, `/etc/group`, SSL certs (`/etc/ssl/certs`, `/etc/pki`), `nsswitch.conf`, and a pre-permissioned writable `/tmp`.
- **`integration`** — ubi-minimal runtime used by the integration-test compose stack; copies the binary, the generated Swagger docs (`/swagger/`), and `sinks_templates/`.
- **`prod`** (default deploy) — `FROM scratch` with the identity/TLS/DNS files from `base`, the pre-built `bin/notifications-service-$TARGETARCH` binary, the Swagger docs, and `sinks_templates/`.

Build arg: `TARGETARCH` (amd64/arm64) — selects which pre-built binary is copied in.

## Testing

| Target | Description |
|--------|-------------|
| `make test` | Full suite: `build` + unit tests (`test-go`) + integration tests. |
| `make integration-tests` | Builds the integration binaries and runs the full docker-compose stack (all profiles) to completion (`--abort-on-container-exit`). |
| `make integration-tests-debug` | Brings up only the backing services (profiles `service`, `postgres`, `smtp`, `slack`) so the service/tests can be run from an IDE. |
| `make build-integration-tests` | Compiles the tagged integration test binary (`-tags integration`) and builds the compose images. |
| `make generate-integration-test-smtp-openapi-client` | Regenerates the MailHog SMTP client used by the tests. |

Integration tests use Docker Compose to stand up the service plus a Postgres database, a [MailHog](https://github.com/mailhog/MailHog) SMTP server (inbox viewable at `http://localhost:8025/`), and a Slack stub — see [`integration_tests/README.md`](integration_tests/README.md) for the debug env vars.

## Configuration

The runtime reads only a small set of environment variables; everything else comes from a YAML config file (individual config keys can also be overridden by their uppercase environment-variable equivalents).

| Variable | Description | Default |
|----------|-------------|---------|
| `CONF_PATH` | Path to the YAML config file | — |
| `ENV_CONFIG_PROVIDER` | Config provider: `advanced` (file + HTTP-managed) or anything else → `basic` (file only) | `basic` |
| `ENCRYPT_DECRYPT_KEY` | Key used to encrypt/decrypt secret config values (e.g. sink credentials) | — |
| `CONF_DB_HOST` | Postgres host override used by the config/history store (integration) | — |

The config file is a set of nested sections — `app`, `history`, `database`, `router`, `listeners`, `sinks`, and `routes`. `listeners`, `sinks`, and `routes` are **maps keyed by a user-chosen id**, each entry declaring a `type` and its type-specific fields; `routes` wire a listener to one or more sinks by id. The canonical, up-to-date shape is the sample files under [`integration_tests/config/`](integration_tests/config/) — `config.basic.yaml`, `config.advanced.yaml`, and `config.advanced.local.yaml`. A representative config:

```yaml
app:
  log_level: debug          # zerolog level
  api_port: 8085            # service API port

history:
  enabled: true
  cleanup:
    max_age_duration: 1m    # how long history rows are retained
    max_events_per_user: 5
    interval_duration: 5s   # cleanup sweep interval

database:                   # Postgres for history + advanced config store
  host: localhost
  port: "5432"
  database: runai
  schema: public
  user: runai
  password: admin
  ssl_mode: disable
  encrypt_decrypt_key: <key>   # encrypts secret sink fields (also settable via ENCRYPT_DECRYPT_KEY)

router:
  concurrent_limit: 50      # max in-flight event dispatches
  cached_events: 1000       # in-memory event cache size

listeners:                  # map of id → listener; type ∈ { http_listener, smtp_listener, kube_events, prometheus_alerts }
  http-runai:
    type: http_listener     # (basic provider nests fields under a `http_listener:` block instead)
    port: 5000
  smtp-runai:
    type: smtp_listener

sinks:                      # map of id → sink; type ∈ { email, slack, postgres, history }
  history:
    type: history
  email1:
    type: email
    user: smtp_user
    password: password
    smtp_host: smtp.mail.com
    smtp_port: 587
    smtp_tls_enabled: true

routes:                     # map of id → route: which listener feeds which sinks
  email_k8s_notifications:
    listener_id: kube_listener_runai
    notify:
      - email1              # sink id(s) from `sinks`
```

> The `basic` and `advanced` providers differ slightly in listener syntax (advanced uses `type: http_listener`; basic nests fields under a `http_listener:` block) — compare `config.basic.yaml` and `config.advanced.yaml`. Any config key can also be overridden by its uppercase environment-variable equivalent. Consult the sample files and the provider structs in `internal/config/` for the full field set.

## OpenAPI and code generation

The service exposes and consumes three OpenAPI contracts; server stubs and Go clients are generated with `openapi-generator-cli` (server `v7.12.0`, `--additional-properties router=chi`). **Generated directories must not be edited by hand** — regenerate via the targets below.

| Spec | Server stub target | Client target |
|------|--------------------|---------------|
| `internal/listen/http/openapi/notifications.yaml` (events API) | `generate-events-openapi-server` → `internal/listen/http/openapi/generated` | `generate-events-openapi-client` → `pkg/client/generated/notify` |
| `internal/config/advanced/http_listener/openapi/listener_api.yaml` (config-provider API) | `generate-config-provider-openapi-server` → `.../openapi/generated` | `generate-events-openapi-client-pg-provider` → `pkg/client/generated/config_client` |
| `internal/history/openapi/endpoints_config.yaml` (history API) | `generate-history-openapi-server` → `internal/history/openapi/generated` | `generate-history-openapi-client` → `pkg/client/generated/history` |

`generate-config-provider-openapi-docs` additionally renders HTML + `swagger.json`/`swagger.yaml` docs (copied into the image at `/swagger/`). `make fix-imports-on-autogenerated-code` runs `goimports` over generated output. All generation targets run automatically as prerequisites of `build_amd`/`build_arm`.

## Go client

`pkg/client/` is the public Go client other services use to send notifications and query config/history. It wraps the three generated clients (`generated/notify`, `generated/config_client`, `generated/history`) with higher-level builders (`notify_builder.go`, `notify_client.go`, `notify_event_client.go`). The client is versioned (`VERSIONED_CLIENT=true`). See [`pkg/client/README.md`](pkg/client/README.md).

## Further reading

- [`pkg/client/README.md`](pkg/client/README.md) — generated events/config/history client usage
- [`integration_tests/README.md`](integration_tests/README.md) — running and debugging the docker-compose e2e suite
