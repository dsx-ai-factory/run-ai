# Makefile build infrastructure

Shared, dockerized `make` targets that every control-plane service consumes via
a one-line `include`. The goal is reproducible builds: every `make build`,
`make test`, `make lint` runs inside a pinned container, so output is identical
regardless of the host.

This document is for service authors and infra maintainers. It is the reference
that follow-up PRs will reconcile against.

## Status: two forked copies in the monorepo

The same idea is implemented twice. They have drifted; consolidating them
is the goal. The forks today:

| Path | Owner / scope | Files |
|------|---------------|-------|
| `control-plane/infra/makefile/` | Control-plane services (~33 services) | `base.mk`, `golang.mk`, `javascript.mk`, `openapi.mk`, `versioning.mk`, `mirrord.mk`, `no-op.mk`, `index.mk`, `Makefile` |
| `runai-cluster/common/build/makefile/` | runai-cluster services | `base.mk`, `golang.mk`, `javascript.mk`, `testenv.mk`, `mirrord.mk`, `index.mk` |

A service-by-service drift audit is in [§ Drift audit](#drift-audit) below.

## How a service consumes this

A service-level `Makefile` typically looks like:

```makefile
SERVICE_NAME=my-service

include ../infra/makefile/index.mk

# Optional: service-specific overrides
build: build-go
test: test-go
lint: lint-go
docker-build: docker-build-generic
docker-push: docker-push-generic
```

The `include` pulls in `base.mk`, `golang.mk`, `openapi.mk`, `javascript.mk`,
`mirrord.mk`, and `versioning.mk`. Services typically wrap the
`-generic` targets to give them simpler names.

Required variables:

| Variable | Set by | Notes |
|----------|--------|-------|
| `SERVICE_NAME` | Service Makefile | Used in image names, console logs, paths inside containers. |

`SERVICE_NAME` is mandatory. `index.mk` resolves its sibling `.mk` files from
its own location via `MAKEFILE_LIST`, so no path variable is required from
consumers regardless of where the service sits in the tree.

## File reference (control-plane)

### `index.mk`

Entry point. Includes `base.mk`, `golang.mk`, `openapi.mk`, `javascript.mk`,
`mirrord.mk`, and `versioning.mk` in that order.

### `base.mk`

Console colors, OS/arch detection, generic Docker commands, image tag logic,
git config setup. Provides:

- Container invocation primitives — `DOCKER_COMMAND`, `DOCKER_COMMAND_WORK_DIR`,
  `DOCKER_COMMAND_PRIVILEGED`, `DOCKER_LINUX_COMMAND`, `DOCKER_CURL_COMMAND`,
  `DOCKER_YQ_COMMAND`, `DOCKER_PYTHON_COMMAND`, `DOCKER_COOKIECUTTER_COMMAND`.
- Targets — `help`, `git-config-command`, `docker-build-generic`,
  `docker-push-generic`.

Override knobs (all `?=`):

| Variable | Default | Purpose |
|----------|---------|---------|
| `DOCKER_TAG` | `0.0.0-dev` | Image tag. |
| `DOCKER_REPO_BASE` | `runai.jfrog.io/cp-containers-lab` | Image registry prefix. |
| `DOCKERHUB_PROXY` | `runai.jfrog.io/dockerhub/` | Public-image proxy. Set empty to pull straight from Docker Hub on environments without `runai.jfrog.io` auth. |
| `DOCKER_WORK_DIR` | `/local/control-plane/${SERVICE_NAME}` | Container working directory. |
| `ARCH` | `arm64` on Apple Silicon, else `amd64` | Build architecture. |
| `OS` | `linux` or `darwin` | Build OS. |

### `golang.mk`

Dockerized Go build, test, lint, format, and mock generation.

- Targets — `build-go` (`build-go-amd` + `build-go-arm`), `test-go`,
  `test-ginkgo` (auto-selects `-local` or `-privileged` based on OS),
  `test-go-acc` (local-only; CI no-op), `lint-go`, `format-go`, `purify-go`,
  `generate-mocks`, `generate-mocks-client`, `gocache`,
  `build-integration-tests-generic`.

Override knobs:

| Variable | Default | Purpose |
|----------|---------|---------|
| `BUILD_IN_PATH` | `cmd/main.go` | Entry point. |
| `GO_VERSION_DEFAULT` | `1.26.6` | Fallback Go version if `go.mod` lookup fails. |
| `GO_VERSION` | Derived from `go.mod` `go X.Y.Z` directive, else `GO_VERSION_DEFAULT` | Resolved Go image tag. |
| `GOLANG_LINTER_CONFIG_PATH` | `../infra/lint/.golangci.yaml` | Lint config. |
| `GO_ACC_IGNORE` | `""` | Comma-separated paths excluded from go-acc coverage. |
| `GO_LDFLAGS` | `''` | Passed to `go build -ldflags`. |
| `INTEG_TESTS_PACKAGE` | `./integration-tests/tests/` | Package compiled by `build-integration-tests-generic`. |
| `INTEG_TESTS_BIN` | `bin/integ-tests` | Output path of the integration-test binary (relative to the service dir). |
| `INTEG_TESTS_TAGS` | `integration` | Build tags for the integration-test binary. |

Fixed (non-overridable) versions: `MOCKERY_VERSION=v2.53.3`,
`GOLANGCI_LINT_VERSION=v2.9.0`, ginkgo `v2.13.2`, go-acc `v0.2.8`.

### `javascript.mk`

Dockerized Node/npm build, test, lint.

- Targets — `build-js`, `test-js`, `lint-js`.

Override knobs:

| Variable | Default | Purpose |
|----------|---------|---------|
| `NODEJS_VERSION` | `20-alpine3.17` | Node image tag. |
| `REDOCLY_AUTHORIZATION` | `""` | Forwarded into the container as `REDOCLY_AUTHORIZATION` env var. |

All three targets run `npm i --save-dev && npm run <stage>` inside a Node container.

### `openapi.mk`

Bundling, code generation, validation, and publishing of OpenAPI specs.
Control-plane only — runai-cluster has no equivalent.

Provides ~25 targets in three groups:

- **Bundle/build** — `openapi-bundle-generic`, `openapi-bundle-json-generic`,
  `openapi-html-doc-generic`, `openapi-bundle`, `openapi-build`,
  `openapi-copy-central-spec`.
- **Code generation** — `openapi-generate-code-server-go`,
  `openapi-generate-code-client-go`, `openapi-generate-code-client-js`,
  `openapi-generate-code`.
- **Validation** — `openapi-lint-generic`, `openapi-validate-yaml`,
  `openapi-validate-code`, `openapi-validate-auth-exists`,
  `openapi-check-service-field`, `openapi-validate-public-service-routes-published`,
  `openapi-validate`.
- **Public specs** — `openapi-generate-public-spec`, `openapi-cleanup-spec`,
  `openapi-publish-spec`, `openapi-build-public-spec`,
  `openapi-build-system-scoped-spec`, `validate-openapi-public-spec`.

Override knobs of note:

| Variable | Default | Purpose |
|----------|---------|---------|
| `VERSIONED_CLIENT` | `false` | Enables versioned-client flow (see `versioning.mk`). |
| `REDOCLY_CLI_VERSION` | `1.12.0` | redocly/cli image tag. |
| `OPENAPI_GENERATOR_VERSION_SERVER` | `v7.7.0` | openapitools/openapi-generator-cli tag for server code. |
| `OPENAPI_GENERATOR_VERSION_CLIENT_GO` | `v7.7.0` | …for Go client code. |
| `OPENAPI_GENERATOR_VERSION_CLIENT_TS` | `v6.0.1` | …for TS client code. |
| `OPENAPI_GENERATOR_SERVER_CODE_PATH` | `internal/generated` | Where server code lands. |
| `OPENAPI_GENERATOR_INTERNAL_CODE_PATH` | `pkg/api` | Inner sub-path under the generator output. |
| `OPENAPI_CLIENT_NAME_GO` | `${SERVICE_NAME}-client` | Go client module name. |
| `OPENAPI_CLIENT_NAME_TS` | `${SERVICE_NAME}-client` | TS client module name. |
| `OPENAPI_CLIENT_SUB_FOLDER` | `generated` | Sub-folder under `runai-clients/<name>/`. |
| `OPENAPI_GENERATOR_DEBUG` | `false` | Add `--global-property debug…` flags. |
| `OPENAPI_GENERATOR_ADDITIONAL_PROPERTY_ADD_RESPONSE_HEADERS` | `false` | Toggles `addResponseHeaders` generator property. |

### `versioning.mk`

Git tag bump and validation for versioned Go modules. `TAG_FORMAT` is the tag
prefix (`$(TAG_FORMAT)/vX.Y.Z` prod, `-dev.N` dev).

- `client-bump-version(-dev)` — legacy patch-only bump for generated clients.
- `package-bump-version(-dev)` — `runai-common-packages/*` and
  `runai-application-packages/*` only; auto patch/minor via `apidiff`,
  confirms before tagging.
- `check-breaking-changes` — preview the `apidiff` verdict, no tagging.
- `validate` / `validate-imports` — go.mod replace / `-dev` client checks.

No `?=` knobs; everything derives from `SERVICE_NAME` and git tag state.

### `mirrord.mk`

Local debugging via [mirrord]. Single target `debug` runs `mirrord exec`
against `.mirrord/mirrord.json`. Press Ctrl+C once to reload, twice to stop.

[mirrord]: https://mirrord.dev/

Override knobs:

| Variable | Default | Purpose |
|----------|---------|---------|
| `RUN_CMD` | `go run ./cmd/main.go` | Command to run under mirrord. |
| `CMD_ARGS` | (empty) | Args appended to `RUN_CMD`. |
| `MIRRORD_ENV_VARS` | (empty) | Env vars prepended to the `mirrord exec` invocation. |

### `no-op.mk`

Empty `test`, `lint`, `validate`, `build` targets. Used by services that opt
out of one or more standard stages (e.g. `redoc`).

### `Makefile`

Top-level `Makefile` of this directory: `include ./no-op.mk`. Exists so that
`make` invoked inside `control-plane/infra/makefile/` itself is well-defined.

## Common pitfalls

- **`DOCKER_COMMAND` mounts `${PWD}/../../`** — assumes the service is two levels
  below the repo root (i.e. `control-plane/<service>/`). Services at a deeper
  path won't see their files. Don't rely on `DOCKER_COMMAND` from outside that
  layout without overriding `-v`.
- **`docker-build-generic` is `--platform=linux/amd64`-only** — even on Apple
  Silicon. Acceptable for production images; surprising for local debugging.
- **`test-go-acc` silently skips on Linux** — by design (testcontainers don't
  work in the CI runner), but the skip is silent rather than an explicit opt-out.
- **`go install <tool>@<version>` happens at runtime** in `ginkgo`, `go-acc`,
  and `mockery` targets. A network outage or yanked module breaks the build.
- **`-v ${HOME}/.gitconfig:…` mounts host state** into the container. Two
  developers' machines can produce subtly different builds; this is the cost
  of supporting private GitHub modules under `GOPRIVATE=github.com/run-ai/*`.

These are not bugs to fix in this PR — they are background for the
hardening PRs.

## Drift audit

Below: where each shared concept has diverged between the two copies. Cells
that match between copies are omitted from the comparison. Use this as the
reference for the reconciliation PRs (2..5).

### `base.mk`

| Concern | `control-plane/infra/makefile/` | `runai-cluster/common/build/makefile/` |
|---------|----------------------------------|-----------------------------------------|
| Image registry | `DOCKER_REPO_BASE?=runai.jfrog.io/cp-containers-lab` | `DOCKER_REPO_BASE?=runai.jfrog.io/op-containers-${ENVIRONMENT}` (with `ENVIRONMENT?=lab`) |
| Public-image proxy | `DOCKERHUB_PROXY?=runai.jfrog.io/dockerhub/`, threaded into every Docker image reference | Not set; references like `node:${NODEJS_VERSION}`, `debian:12.1-slim` resolve directly |
| Container working dir | `DOCKER_WORK_DIR?=/local/control-plane/${SERVICE_NAME}` | `DOCKER_WORK_DIR?=/local/runai-cluster/${SUB_DIR}/${SERVICE_NAME}` (requires `SUB_DIR`) |
| Workspace mount | `-v "${PWD}/../../:/local"` | `-v "${PWD}/../../../:/local:z"` (3 levels up, with `:z` SELinux label) |
| Privileged command | `DOCKER_COMMAND_PRIVILEGED` defined | `DOCKER_COMMAND_PRIVILEGED` defined (same shape, with `:z` mount + `runai-cluster` path) |
| Extra commands | `DOCKER_CURL_COMMAND`, `DOCKER_YQ_COMMAND`, `DOCKER_COOKIECUTTER_COMMAND`, `DOCKER_PYTHON_COMMAND` | Only `DOCKER_LINUX_COMMAND` |
| `docker-build-generic` | `docker build … -f ./Dockerfile -t … --platform=linux/amd64` | `docker build ${DOCKER_BUILD_ADDITIONAL_ARGS} … -f ${DOCKERFILE_PATH} -t … --platform=linux/amd64` (overridable Dockerfile path + extra args) |
| `VERSION` alias | Not defined | `VERSION=${DOCKER_TAG}` |
| `git-config-command` | Identical | Identical |
| Console colors / OS detection / help target | Identical | Identical |

### `golang.mk`

| Concern | `control-plane/infra/makefile/` | `runai-cluster/common/build/makefile/` |
|---------|----------------------------------|------------------------------------------|
| Go version | Derived from `go.mod` `go X.Y.Z`, fallback `GO_VERSION_DEFAULT?=1.26.3`; image is `library/golang:${GO_VERSION}` | Hard-coded `GO_VERSION=1.26.3-bookworm`; image is `golang:${GO_VERSION}` |
| `GOPRIVATE` | Set on every Docker invocation: `-e GOPRIVATE=github.com/run-ai/*` | Not set; relies on `~/.gitconfig` + `~/.netrc` mounts |
| Mockery | v2.53.3, installed inside the Docker `golang` image at runtime, then `go generate ./...` | v3.2.5, installed locally to `$(BUILD_DIR_ABS)/mockery` on the host; invoked as `$(MOCKERY) --all --keeptree …` |
| Ginkgo / Gomega versions | Ginkgo `v2.13.2`, Gomega unpinned | Ginkgo `v2.13.2` (variable `GINKGO_VERSION`), Gomega `v1.30.0` |
| Test timeout | `--timeout=15m` | `--timeout=5m` |
| `lint-go` | Always runs | Always runs; uses `--new-from-rev=$(LINT_REV)` (only flag lines new on this branch) |
| `test-go` | `go test ./... -short -coverprofile ./cover.out` | `${GO_TEST_COMMAND}` = `go test ./... -coverprofile ./cover.out` (no `-short`). Depends on `go-list` target which validates `go mod`. |
| `test-go-acc` | Defined; Linux skips, Darwin runs | Not defined |
| FIPS / build env | `go build -ldflags ${GO_LDFLAGS} -o … ${BUILD_IN_PATH}` | `env GOFIPS140=v1.0.0 go build ${GO_BUILD_ADDITIONAL_FLAGS} -o … ${BUILD_IN_PATH}` |
| CGO commands | Not provided | `DOCKER_CGO_BASE_COMMAND` and `DOCKER_CGO_COMMAND` exposed for services that need CGO |
| `go-list` target | Not defined | Defined; runs `go list -m all` to surface broken module replaces (gated by `GO_LIST_ENABLED`) |
| `validate` | Not defined here (lives in `versioning.mk`) | Defined; same `go.mod` import validation as the CP `versioning.mk` `validate-imports` target |
| Cache mounts | Same shape on both sides: `~/.cache/go-build-docker-gocache` → `/tmp/.cache`, `~/.cache/go-build-docker-gopath` → `/go`. `runai-cluster` adds `:z` SELinux labels. | |
| `purify-go`, `format-go`, `build-go-amd`, `build-go-arm`, `gocache` | Same shape on both sides. | |

### `javascript.mk`

| Concern | `control-plane/infra/makefile/` | `runai-cluster/common/build/makefile/` |
|---------|----------------------------------|------------------------------------------|
| Docker command flags | `-i` (interactive), `REDOCLY_AUTHORIZATION` env passthrough | Neither `-i` nor `REDOCLY_AUTHORIZATION` |
| Image source | `${DOCKERHUB_PROXY}library/node:${NODEJS_VERSION}` | `node:${NODEJS_VERSION}` |
| Build/test/lint commands | `npm i --save-dev && npm run <stage>` | Same |

### `mirrord.mk`

| Concern | `control-plane/infra/makefile/` | `runai-cluster/common/build/makefile/` |
|---------|----------------------------------|------------------------------------------|
| `mirrord -V` check | `mirrord -V` (output visible) | `@mirrord -V > /dev/null` (silent) |
| Debug target info banner | Single `$(info)` line | Three `$(info)` lines with `-----------` separators |
| Otherwise | Identical | Identical |

### Files unique to one side

| File | Where | Why |
|------|-------|-----|
| `openapi.mk` | control-plane only | runai-cluster services don't expose OpenAPI specs. |
| `versioning.mk` | control-plane, runai-common-packages, runai-application-packages | The versioned-module tag-bump flow (control-plane clients and `runai-common-packages`) isn't used by runai-cluster services. |
| `no-op.mk` | control-plane only | Service-level opt-out for `test`/`lint`/`build`/`validate`. (runai-cluster services tend to wire those targets directly into their own Makefile.) |
| `testenv.mk` | runai-cluster only | Kubernetes envtest fixtures (Kind, kubebuilder, helm). Control-plane services don't use envtest. |
| `Makefile` | control-plane only | Default `make` inside the dir → no-op. |
