# keycloak-extensions

Custom [Keycloak](https://www.keycloak.org/) extensions for Run:ai — a set of
Java **SPI providers**, a custom login/email **theme**, and the OpenTelemetry
Java agent, packaged as a sidecar container image that copies its payload into a
running Keycloak instance. This is how the platform's identity layer
(fronted by `identity-manager`) customises Keycloak: token/claim mappers,
event listeners, and an extra admin REST endpoint. Owned by the **Sphinx** team.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [Upgrade instructions](#upgrade-instructions)
- [Further reading](#further-reading)

## Overview

The service is responsible for:

- **SPI providers** (`runai-custom-providers/`) — a single JAR
  (`runai-custom-providers.jar`) loaded by Keycloak at startup, providing:
  identity-provider mappers (claim validator, custom claim→attribute mapper),
  OIDC protocol mappers (custom groups, counter, user-app groups), event
  listeners (last-login, password-audit), and a custom admin REST resource
  (`runaienteties`). See
  [`runai-custom-providers/README.md`](runai-custom-providers/README.md) for
  the full per-provider reference.
- **Theme** (`runai-theme/`) — custom Keycloak login and email templates.
- **OpenTelemetry agent** — downloaded at build time and shipped in the image.
- **Sidecar delivery** — at runtime the container copies providers, theme, and
  agent into shared volumes (`/providers`, `/themes`, `/agent`) via the scripts
  in `scripts/`, then writes Keycloak config.

## Repository layout

```text
.
├── runai-custom-providers/     # Java SPI extensions (Gradle project)
│   ├── src/main/java/...       # Providers (mappers, event listeners, admin REST)
│   ├── src/test/java/...       # JUnit 5 tests
│   ├── build.gradle            # Keycloak 26.x compileOnly deps, Java 21 toolchain
│   ├── settings.gradle
│   └── README.md               # Per-provider documentation
├── runai-theme/                # Custom Keycloak login + email theme
│   ├── login/
│   └── email/
├── scripts/                    # Runtime copy + keycloak-config scripts
│   ├── copy-extensions.sh
│   └── write-keycloak-conf.sh
├── bin/                        # Build output (jars) — populated by `make build`
├── Dockerfile                  # Alpine sidecar image
└── Makefile
```

## Building and running

This is a **Java / Gradle** service, not a Go service. The Makefile drives Gradle
inside a `gradle:8.12.0-jdk21` Docker container (`GRADLE_IMAGE`).

| Target | Description |
|--------|-------------|
| `make build` | Runs Gradle in the container: clears the Gradle cache, downloads the OpenTelemetry Java agent (`v2.29.0`, sha256-verified) into `bin/`, runs `gradle build -p runai-custom-providers`, and copies `runai-custom-providers.jar` to `bin/`. Produces the artifacts the Dockerfile expects. |
| `make test` | Runs `gradle test -p runai-custom-providers` in the container, then parses the HTML test reports via `scripts/parse_html_report.py` |
| `make lint` | No-op |
| `make validate` | No-op |

The Gradle project (`runai-custom-providers/build.gradle`) targets the Java 21
toolchain and compiles against Keycloak `26.6.+` server SPI dependencies
(`compileOnly` — provided by the Keycloak server at runtime). Tests use
JUnit 5, Mockito, and AssertJ.

### Runtime image (Dockerfile)

The Dockerfile is a minimal `alpine:3.21` image. It creates the extensions
staging and shared target directories owned by `65532:65532`, copies in the
built `runai-custom-providers.jar`, the OpenTelemetry agent JAR, the
`runai-theme/`, and the `scripts/`. It runs as user `65532:65532`, and its
`CMD` runs `copy-extensions.sh` (copies the payload into the shared volumes)
then `write-keycloak-conf.sh`.

| Source (image) | Destination (shared volume) | Content |
|---|---|---|
| `/extensions/providers/` | `/providers/` | Provider JARs |
| `/extensions/themes/` | `/themes/` | Keycloak theme |
| `/extensions/agent/` | `/agent/` | OpenTelemetry Java agent |

### Local build (local-dev.sh)

`local-dev.sh` builds this service through its dedicated
`build_keycloak_extensions` function. It runs:

```bash
# Step 1 — Gradle build inside the gradle:8.12.0-jdk21 container:
make build DOCKERHUB_PROXY=
# Step 2 — build + load the sidecar image:
docker buildx build --platform linux/amd64 --provenance=false --load \
  -t runai-local/keycloak-extensions:0.0.0 .
```

`DOCKERHUB_PROXY=` (empty) pulls the `gradle` builder image from Docker Hub. No
`GITHUB_TOKEN` and no `TARGETARCH` build-arg are involved — it is a Java build
and the image just packages the resulting JARs. The image tag is **`0.0.0`**
(the control-plane chart appVersion); any other tag breaks the local deploy.
Image ref: `runai-local/keycloak-extensions:0.0.0`.

> The sibling `identity-manager` integration tests also build this image (via
> its `make build-keycloak` target).

## Testing

| Target | Description |
|--------|-------------|
| `make test` | Runs the JUnit 5 suite for `runai-custom-providers` in the Gradle container and parses the HTML reports |

Within the Gradle project you can also run `gradle test -p runai-custom-providers`
directly (Java 21 + Gradle 8 required).

## Configuration

The providers read a small number of environment variables at runtime (inside
the Keycloak process). See
[`runai-custom-providers/README.md`](runai-custom-providers/README.md) for the
full per-provider configuration; the service-wide integration variables are:

| Variable | Description | Default |
|----------|-------------|---------|
| `KC_RUNAI_AUDIT_SERVICE_URL` | Base URL of the audit service (password-audit event listener) | `http://runai-backend-audit-service:8080` |
| `KC_RUNAI_AUTHORIZATION_SERVICE_URL` | Base URL of the authorization service (group intersection in the groups mappers) | `http://runai-backend-authorization:8080` |

Per-provider settings (claim names, user attributes, tenant IDs, token claim
names) are configured through the Keycloak Admin Console / realm configuration,
not via environment variables.

## Upgrade instructions

1. Update the Keycloak version in `runai-custom-providers/build.gradle`
   (`keycloakVersion` / the `org.keycloak:*` versions).
2. Update the base image tag in the chart / Keycloak deployment to match the new
   Keycloak release (see the
   [Keycloak Quay.io page](https://quay.io/repository/keycloak/keycloak?tab=tags)).
3. Update the Keycloakx Helm chart version based on the
   [Keycloakx chart releases](https://github.com/codecentric/helm-charts/tree/master/charts/keycloakx).
4. Run `make build` and verify `make test` passes.

## Further reading

- [`runai-custom-providers/README.md`](runai-custom-providers/README.md) —
  full reference for every SPI provider (mappers, event listeners, admin REST
  endpoint) and their configuration.
</content>
