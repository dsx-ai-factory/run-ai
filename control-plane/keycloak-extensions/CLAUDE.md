# keycloak-extensions — Build Context

**Type:** container (Java / Gradle — Keycloak SPI plugins + theme — **and** a
small Go init binary)
**Team:** Sphinx
**Gradle project:** `runai-custom-providers` (group `com.runai.keycloak`)
**Go module:** `github.com/dsx-ai-factory/run-ai/control-plane/keycloak-extensions`

## Make targets

Targets come from the service `Makefile` (which includes the shared
`control-plane.mk` / `index.mk`, and — via `index.mk` — `golang.mk`). This
service builds two things: the Gradle-based SPI providers/theme payload, and
a small Go binary that stages that payload and generates `keycloak.conf`
inside the runtime image (see "Runtime base image" below).

| Target | What it does |
|--------|-------------|
| `build` | `build-gradle` then `build-go` |
| `build-gradle` | In the Gradle container: clear Gradle cache, download the OpenTelemetry Java agent (`v2.29.0`, sha256-verified) into `bin/`, run `gradle build -p runai-custom-providers`, copy `runai-custom-providers.jar` → `bin/` |
| `build-go` | `build-go-amd` + `build-go-arm` — compiles `cmd/main.go` for amd64 + arm64 into `bin/keycloak-extensions-$ARCH` |
| `test` | `test-gradle` then `test-go` |
| `test-gradle` | `gradle test -p runai-custom-providers` in the container, then parses HTML reports via `scripts/parse_html_report.py` |
| `test-go` | Plain `go test ./...` for the `internal/keycloakinit` package (table-driven unit tests, no Ginkgo — this is a one-shot init binary, not a controller/webhook) |
| `lint` | `lint-go` — golangci-lint against the Go code (the Gradle side has no lint target) |
| `validate` | No-op |

## Key build notes

- Language/toolchain: Java 21, Gradle 8.12 (SPI providers); Go 1.27 (init
  binary, stdlib only — no third-party dependencies). Java compiles against
  Keycloak `26.6.+` server SPI deps (`compileOnly`, provided by the server at
  runtime). Java tests: JUnit 5 + Mockito + AssertJ.
- Generated code: none.
- Runtime base image: **`scratch`**, multi-stage (a
  `registry.access.redhat.com/ubi9/ubi-minimal` stage supplies `/etc/passwd`,
  `/etc/group`, and pre-owned/writable `/providers`, `/themes`, `/agent`
  mount points — no CA certs needed, the binary makes no outbound network
  calls). Copies in the provider JAR, OTel agent JAR, `runai-theme/`, and the
  compiled `keycloak-extensions-$TARGETARCH` binary as `ENTRYPOINT`. Runs as
  `65532:65532`.
- The Go binary (`cmd/main.go` → `internal/keycloakinit`) replaces the old
  `scripts/copy-extensions.sh` + `scripts/write-keycloak-conf.sh` shell
  scripts, which can't run in a `scratch` image (no shell). It performs the
  same two steps in one run-to-completion process: copy
  `/extensions/{providers,themes,agent}` into `/providers`, `/themes`,
  `/agent`, then read `/opt/keycloak/conf/secrets/*` and write
  `/conf/keycloak.conf` (`KC_*` names → kebab-case keys, `KEYCLOAK_ADMIN*`
  skipped, everything else passed through). A missing secrets dir is
  tolerated (produces an empty conf file), matching the original script's
  behavior when nothing is mounted there.
- Build args: `TARGETARCH` (selects `bin/keycloak-extensions-$TARGETARCH`;
  auto-populated by `docker buildx build --platform ...`).
- Build command (what local-dev.sh actually runs): the dedicated
  `build_keycloak_extensions` function, in three steps:
  1. Gradle build in the container: `make build-gradle DOCKERHUB_PROXY=`
     (`DOCKERHUB_PROXY=` pulls the `gradle` builder image from Docker Hub. No
     `GITHUB_TOKEN` — Java build, no private Go modules.)
  2. Go build (amd64 only, matching `build_go_service`'s local-dev pattern):
     `make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=`
  3. Build + load the image: `docker buildx build --platform linux/amd64
     --provenance=false --load -t runai-local/keycloak-extensions:0.0.0 .`
     (`ARG TARGETARCH` is auto-populated by `--platform`.)
     Image ref: `runai-local/keycloak-extensions:0.0.0` (tag `0.0.0` — the CP
     chart appVersion; any other tag breaks the local deploy).
- Also built by `identity-manager`'s integration tests (`make build-keycloak`
  → plain `make build`, both arches). The integration-test
  `docker-compose.yaml` builds this image with an explicit
  `--build-arg TARGETARCH=amd64` (compose doesn't auto-populate it the way
  `buildx --platform` does) and — since `scratch` has no shell — runs the
  image's default `ENTRYPOINT` rather than a `command:` override; the
  dependent `keycloak` service waits on `condition:
  service_completed_successfully` instead of a shell-based healthcheck.

## local-dev.sh iteration

This is a control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
