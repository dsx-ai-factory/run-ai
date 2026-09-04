---
name: write-service-docs
description: Generate or update a detailed README.md and CLAUDE.md for a Run.ai monorepo service, matching the quality of identity-manager and bff-service. Provide the service path relative to repo root. When invoked without a path, detects shared makefile changes in the current PR and regenerates docs for all affected services.
tools:
  - Bash
  - Read
  - Write
  - Edit
---

## Two modes

**With a service path** (e.g. `/write-service-docs control-plane/identity-manager`): generate or update
`README.md` and `CLAUDE.md` for that single service. Skip to **"Finding the repo root"** below.

**Without a service path**: detect shared makefile changes in the current PR and regenerate docs
for every affected service. Follow the steps below.

### Shared-makefile mode

#### Step 1 — find changed shared makefiles

```bash
REPO_ROOT=$(git rev-parse --show-toplevel)
BASE=$(git merge-base origin/master HEAD)
git diff --name-only "$BASE"...HEAD | grep '^build/makefile/.*\.mk$'
```

If the output is empty, tell the user "No shared makefile changes detected" and stop.

#### Step 2 — check whether targets actually changed

Only regenerate if Make targets were added, removed, or renamed. Internal variable changes
(`:=`, `?=`, `+=` lines) do not require doc updates.

```bash
BASE=$(git merge-base origin/master HEAD)
for MK_FILE in $(git diff --name-only "$BASE"...HEAD | grep '^build/makefile/.*\.mk$'); do
  git diff "$BASE"...HEAD -- "$MK_FILE" \
    | grep -E '^[+-][a-zA-Z_][a-zA-Z0-9_-]*[[:space:]]*:([^=]|$)' \
    | grep -v '^---\|^+++'
done
```

If the output is empty, tell the user "Shared makefiles changed but no target additions/removals
detected — doc regeneration not needed" and stop.

#### Step 3 — find all affected services

All services chain through `index.mk`, `control-plane.mk`, or `runai-cluster.mk`, so any change
to a leaf `.mk` file (e.g. `golang.mk`, `base.mk`) affects every service. Only non-comment
include lines are matched to avoid false positives from commented-out includes.

```bash
REPO_ROOT=$(git rev-parse --show-toplevel)
find "$REPO_ROOT" -name "Makefile" -not -path "*/.git/*" | while read -r makefile; do
  if grep -E "^[^#]*include[[:space:]].*(index|control-plane|runai-cluster)\.mk" "$makefile" \
      > /dev/null 2>&1; then
    svc_dir=$(dirname "$makefile")
    if [ -f "$svc_dir/artifact-metadata.json" ]; then
      echo "${svc_dir#$REPO_ROOT/}"
    fi
  fi
done | sort -u
```

If the output is empty, tell the user "No service Makefiles include the changed shared makefile"
and stop. Otherwise tell the user how many services were found and list them. Ask for confirmation
before proceeding if the count exceeds 20.

#### Step 4 — regenerate docs for each affected service

For each service path, run the single-service flow below (starting at "Finding the repo root"),
with that service path as the argument. Process them one by one, confirming each write before
moving on.

#### Step 5 — report

Summarise: which `.mk` files changed, which targets were affected, how many services were
regenerated, and whether any were skipped.

---

Generate thorough technical documentation for a single service in the Run.ai monorepo.
Match the quality of `control-plane/identity-manager/README.md` and `control-plane/bff-service/README.md` — those are the reference standard.

## Finding the repo root
```bash
REPO_ROOT=$(git rev-parse --show-toplevel)
SVC="$REPO_ROOT/<service_path>"
```
Never use absolute user-specific paths.

## MANDATORY: cross-reference local-dev.sh for the REAL build parameters

The Makefile `?=` defaults are NOT what developers actually use locally. `local-dev.sh` is the
source of truth for how each service is really built and tagged. **Read `$REPO_ROOT/local-dev.sh`**
and find the build function for this service's type. Document the ACTUAL invocation, not the theoretical one.

The canonical facts from local-dev.sh (verify against the current script — do not hardcode blindly):

- **Go services:** `make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=`
  then `docker build --build-arg TARGETARCH=amd64`
- **Frontend (JS):** `make build-js DOCKERHUB_PROXY=`
- **keycloak-extensions (Gradle):** `make build DOCKERHUB_PROXY=`
- **`DOCKERHUB_PROXY=` (empty)** — pulls builder images straight from Docker Hub, bypassing the internal proxy
- **`GITHUB_TOKEN=` (empty in the make call)** — the private-module token is applied elsewhere
- **Image tag = the "VERSION"** — FIXED, not `DOCKER_TAG=dev`:
  - Control-plane services → **`0.0.0`** (must match the CP chart `appVersion`)
  - runai-cluster services → **`0.0.0-dev`** (RC chart hardcodes this)
  - Image ref format: `runai-local/<service_name>:<tag>`
- These tags are hardcoded because the Helm charts reference them; using any other tag breaks the local deploy.

Every README **Building and running** section MUST include a "Local build (local-dev.sh)" subsection
showing the exact command above for this service's type, plus the correct image tag (`0.0.0` for CP,
`0.0.0-dev` for RC). Every CLAUDE.md MUST include these in its parameters/build-notes.

### Both architectures (amd64 AND arm64) — for Go services
For every Go service, document BOTH arch build paths, because they differ:
- **amd64** (what local-dev.sh runs): `make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=`
  then `docker build --build-arg TARGETARCH=amd64 -t runai-local/<svc>:<tag> .`
- **arm64** (manual; for arm64 clusters): `make build-go-arm DOCKERHUB_PROXY= GITHUB_TOKEN=`
  then `docker build --build-arg TARGETARCH=arm64 -t runai-local/<svc>:<tag> .`

Key difference to state explicitly (verify in `build/makefile/golang.mk`):
- `build-go-amd` reads `OS=`/`ARCH=` variables — that's why local-dev.sh passes `OS=linux ARCH=amd64`.
- `build-go-arm` **hardcodes** `GOOS=linux GOARCH=arm64` — do NOT pass `OS=`/`ARCH=` to it.
- The image tag is the SAME (`0.0.0` CP / `0.0.0-dev` RC) for both arches; only `TARGETARCH` changes.
- Note that `local-dev.sh` itself only builds amd64 (Kind nodes are linux/amd64); arm64 is a manual path.

## What you read (read ALL of these before writing anything)

0. **`$REPO_ROOT/local-dev.sh`** — the build function for this service type (see MANDATORY section above)
1. **`artifact-metadata.json`** — artifact_type, team name, service name
2. **`Makefile`** — read the full content. Many services only `include` shared files; if so, also read:
   - `$REPO_ROOT/build/makefile/*.mk` — shared targets
   - Any other included `.mk` files listed in the Makefile
   Extract: every target name + its `## description` comment. Group into build targets vs test targets.
3. **`Dockerfile`** — base image, build args (`ARG`), `ENV` vars, exposed ports, entrypoint/cmd, multi-stage structure
4. **`go.mod`** (first 20 lines) — Go module path, Go version, key dependencies
5. **`package.json`** — if present: scripts section
6. **Directory layout** — run `ls cmd/ pkg/ internal/ integration-tests/ swagger/ 2>/dev/null` and one level deeper for key dirs: `ls pkg/*/` etc. Use this to build the repo layout tree.
7. **`cmd/main.go`** or **`internal/config/`** — look for env var definitions, `os.Getenv`, config structs. This is how you find configuration.
8. **`integration-tests/`** — check if it exists and what it contains (docker-compose files, README)
9. **`swagger/source/openapi.yaml`** — if present, note the API exists and where the spec lives
10. **Existing `README.md`** — if present, update rather than replace; keep any sections that are accurate

## README.md structure to write

Follow this structure exactly, only omitting sections that genuinely don't apply:

```markdown
# <service-name>

One paragraph: what this service does, its role in the platform, and which team owns it.
Be specific — name the things it manages (e.g. "owns X, Y, Z") and how it fits in (e.g. "thin layer on top of Keycloak").

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [OpenAPI and code generation](#openapi-and-code-generation)   ← only if swagger/ exists
- [Go client](#go-client)                                        ← only if pkg/client/ exists
- [Further reading](#further-reading)                           ← only if nested READMEs exist

## Overview

The service is responsible for:

- **Thing 1** — explanation
- **Thing 2** — explanation

(Be specific. Don't write "handles X" — write what X actually is.)

## Repository layout

Show a tree of the top-level directories with one-line descriptions:
​```text
.
├── cmd/                    # Service entry point (main.go)
├── pkg/
│   ├── business/           # Domain logic
│   ├── client/             # Versioned Go client
│   └── database/           # DAO layer and SQL migrations
├── internal/
│   ├── config/             # Configuration loading
│   └── generated/          # OpenAPI-generated server stubs (do not edit)
├── integration-tests/      # Docker-compose-driven integration tests
├── swagger/                # OpenAPI source and built bundle
├── Dockerfile
├── Makefile
└── go.mod
​```

## Building and running

The project is built with Go (see [`go.mod`](go.mod)) and orchestrated through the [`Makefile`](Makefile).

| Target | Description |
|--------|-------------|
| `make build` | <what it actually does, from ## comment or by reading the target> |
| `make lint` | Run the Go linter (golangci-lint) |
| `make purify` | Auto-fix formatting and imports |
| `make validate` | <description> |
| `make run-local` | <description, if target exists> |
| `make docker-build` | Build the container image |

List ALL build-related targets. Include the exact `make <target>` form.
If `make build` runs OpenAPI generation first, say so explicitly.

For Dockerfile: describe the multi-stage build (stage names, base images, what each stage produces).
If there are `--build-arg` parameters, list them:
​```bash
docker build --build-arg TARGETARCH=amd64 -t <name>:dev .
​```

## Testing

| Target | Description |
|--------|-------------|
| `make unit-test` | Build and run unit tests |
| `make test` | Unit tests + integration tests |
| `make docker-compose-integration-tests-up` | Run the full docker-compose test suite |
| `make docker-compose-integration-tests-debug` | Same, without the failure handler |
| `make coverage` | Open coverage report |

List ALL test-related targets. If integration tests need backing services, note what they need.

## Configuration

List every environment variable the service reads. Find them in:
- `internal/config/`, `pkg/config/`, `cmd/main.go`
- `Dockerfile` ENV lines
- `integration-tests/config/`

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_PORT` | Port to listen on | `8080` |
| `DATABASE_HOST` | PostgreSQL host | — |

## OpenAPI and code generation   (omit section if no swagger/)

Where the spec lives, what gets generated, and the `make` target to regenerate.
Always note: "generated directories must not be edited by hand".

## Go client   (omit section if no pkg/client/)

Point to `pkg/client/README.md`. Mention `make client-bump-version` if that target exists.

## Further reading   (omit if no nested READMEs)

- [`pkg/client/README.md`](pkg/client/README.md) — client usage and versioning
- (any other nested READMEs discovered)
```

## CLAUDE.md to write alongside the README

```markdown
# <service-name> — Build Context

**Type:** <artifact_type>
**Team:** <team from artifact-metadata.json>

## Make targets

| Target | What it does |
|--------|-------------|
| build  | ... |
| test   | ... |
| lint   | ... |

## Key build notes

- Generated code: run `make <generate-target>` before `make build` (generated dirs must not be edited)
- Base image: <from Dockerfile>
- Build args: TARGETARCH (amd64 / arm64)
- Integration tests: require docker-compose + backing services

## local-dev.sh iteration

For control-plane services:
​```bash
./local-dev.sh --skip-build-rc --skip-cluster
​```
For runai-cluster services:
​```bash
./local-dev.sh --skip-build-cp --skip-cluster
​```
```

## Quality rules

- Every make target that exists in the Makefile (or its shared includes) must appear in a table.
- Every env var you find in config/ or Dockerfile must appear in the Configuration table.
- Directory layout must reflect what actually exists — run `ls` to confirm before writing.
- Do not invent information. If you cannot determine what a target does, read its recipe in the Makefile.
- Minimum length: 400 words. If shorter, you missed something — go back and read more files.

## After writing

Tell the user: "Wrote README.md and CLAUDE.md to <relative_path>/" and list the sections you included.
