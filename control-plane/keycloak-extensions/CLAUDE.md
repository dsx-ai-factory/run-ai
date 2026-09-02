# keycloak-extensions — Build Context

**Type:** container (Java / Gradle — Keycloak SPI plugins + theme)
**Team:** Sphinx
**Gradle project:** `runai-custom-providers` (group `com.runai.keycloak`)

## Make targets

Targets come from the service `Makefile` (which includes the shared
`control-plane.mk` / `index.mk`). This is a Gradle-based service — targets run
Gradle inside a `gradle:8.12.0-jdk21` Docker container (`GRADLE_IMAGE`), not the
Go builder.

| Target | What it does |
|--------|-------------|
| `build` | In the Gradle container: clear Gradle cache, download the OpenTelemetry Java agent (`v2.29.0`, sha256-verified) into `bin/`, run `gradle build -p runai-custom-providers`, copy `runai-custom-providers.jar` → `bin/` |
| `test` | `gradle test -p runai-custom-providers` in the container, then parse HTML reports via `scripts/parse_html_report.py` |
| `lint` | No-op |
| `validate` | No-op |

## Key build notes

- Language/toolchain: Java 21, Gradle 8.12; compiles against Keycloak `26.6.+`
  server SPI deps (`compileOnly`, provided by the server at runtime). Tests:
  JUnit 5 + Mockito + AssertJ.
- Generated code: none.
- Runtime base image: `alpine:3.21` sidecar. Creates dirs owned by
  `65532:65532`, copies in the provider JAR, OTel agent JAR, `runai-theme/`,
  and `scripts/`. Runs as `65532:65532`; CMD runs `copy-extensions.sh` then
  `write-keycloak-conf.sh` to populate `/providers`, `/themes`, `/agent`.
- Build args: none (`container_build` for this service is called without
  `--build-arg TARGETARCH`).
- Build command (what local-dev.sh actually runs): the dedicated
  `build_keycloak_extensions` function, in two steps:
  1. Gradle build in the container: `make build DOCKERHUB_PROXY=`
     (`DOCKERHUB_PROXY=` pulls the `gradle` builder image from Docker Hub. No
     `GITHUB_TOKEN` — Java build, no private Go modules.)
  2. Build + load the image: `docker buildx build --platform linux/amd64
     --provenance=false --load -t runai-local/keycloak-extensions:0.0.0 .`
     Image ref: `runai-local/keycloak-extensions:0.0.0` (tag `0.0.0` — the CP
     chart appVersion; any other tag breaks the local deploy).
- Also built by `identity-manager`'s integration tests (`make build-keycloak`).

## local-dev.sh iteration

This is a control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
</content>
