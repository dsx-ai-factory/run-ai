---
name: write-service-docs
description: Generate or update a detailed README.md and CLAUDE.md for a Run.ai monorepo service, matching the quality of identity-manager and bff-service. Provide the service path relative to repo root. When invoked without a path, detects shared makefile changes in the current PR and regenerates docs for all affected services.
allowed-tools: Bash, Read, Write, Edit
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
AFFECTED=""
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

## MANDATORY: DERIVE the real build commands — never hardcode target names

Build target names and recipes CHANGE over time. Do NOT assume any specific target (such as
`build-go-amd`/`build-go-arm`, `build-js`, etc.) — always **read the current repo at generation
time** and document exactly what you find. The Makefile `?=` defaults are also not what developers
actually run; `local-dev.sh` is the source of truth for the real local build.

Derive the build like this:

1. **Resolve the service's `Makefile` `build:` target.** Read the service `Makefile` and follow
   what `build:` depends on / invokes. Most services only `include ../../build/makefile/*.mk`, so
   read those includes and find the ACTUAL recipe(s): the exact `make` target name(s), the
   `go build` / compiler invocation, any `GOOS`/`GOARCH`/arch handling, and any `docker build` step.
   Document the real target names and commands — whatever they are today.
2. **Read `$REPO_ROOT/local-dev.sh`** and find the build function for THIS service's type
   (Go service / frontend-JS / image-only / Gradle-based / etc.). local-dev.sh is the source of
   truth for how the service is really built and tagged locally. Document the ACTUAL invocation it
   runs — the target it calls, the `make` variables it passes, and the `docker build` flags.
3. **Match reality on arch.** If the build has separate per-arch recipes/targets, document each; if
   it's a single target or single-arch, document just that. Do not assume two arches or any target name.

Cross-check these conventions AGAINST the current script (they may change — verify, don't assume):
- **`DOCKERHUB_PROXY=` (empty)** in the local-dev call — pulls builder images from Docker Hub.
- **`GITHUB_TOKEN=` (empty in the make call)** — the private-module token is applied elsewhere.
- **Image tag = a FIXED version the Helm charts reference** (commonly `0.0.0` for control-plane,
  `0.0.0-dev` for runai-cluster). Confirm the exact tag in local-dev.sh — any other tag breaks the
  local deploy. Image ref format: `runai-local/<service_name>:<tag>`.

Every README **Building and running** section MUST include a "Local build (local-dev.sh)" subsection
showing the EXACT command(s) local-dev.sh runs for this service (as derived above), plus the correct
image tag. Every CLAUDE.md MUST include these in its build-notes. If a shared makefile change ever
renames the build target, the docs stay correct because you read it fresh each time.

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
7. **Configuration — env vars AND config files.** Two models exist; determine which the service uses and document accordingly:
   - **Env-var driven:** `cmd/main.go`, `internal/config/`, `pkg/config/` — look for `os.Getenv`/`LookupEnv`, `env:`/`envconfig:` struct tags, `Dockerfile` `ENV` lines.
   - **Config-file driven** (detect via signals like a config library using `SetConfigFile` / `mapstructure` tags / a loaded `*.yaml`): the struct tags are config-*file* keys, NOT env vars. **You MUST read the actual sample config files** to document the real schema — do not reconstruct it from a grep of struct tags (that misses nesting, maps-keyed-by-id, and type blocks). Find and read samples under: `integration-tests/config/`, `integration_tests/config/`, `config/`, `testdata/`, `*.yaml`/`*.yml` next to the service, and any path passed to `SetConfigFile`. Reproduce a representative real config (trimmed) and describe each top-level section. If env vars can override config keys, say so in plain terms (e.g. "any key can be overridden by its uppercase env-var equivalent") — **do NOT name the config library or its internals (viper, `AutomaticEnv`, etc.) in the generated docs; those are implementation details.** Only `CONF_PATH`-style *bootstrap* env vars (which file to load, which provider) belong in the env-var table.
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

Document configuration for whichever model the service uses (see read-step 7). Most services are env-var driven; some are config-file (YAML) driven; some are both. Do not name the underlying config library or its internals (viper, `AutomaticEnv`, etc.) in the generated docs — describe behavior, not the library.

**Env vars** — list every one the service reads. Find them in `internal/config/`, `pkg/config/`, `cmd/main.go`, and `Dockerfile` ENV lines:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_PORT` | Port to listen on | `8080` |
| `DATABASE_HOST` | PostgreSQL host | — |

**Config file (only if the service loads a YAML config file)** — do NOT reduce this to a flat key list reconstructed from struct tags; that misses the real structure. Read the actual sample config files (`integration-tests/config/`, `config/`, `testdata/`, etc.), then:
- Show a representative real config as a fenced ```yaml block (trimmed), with a short inline comment per key.
- Describe each top-level section; call out sections that are **maps keyed by a user-chosen id** (e.g. listeners/sinks/routes each declare a `type` + type-specific fields) and how they reference each other.
- Point to the sample files by path as the source of truth, and if env vars can override config keys, note it in plain terms (not by naming the config library or `AutomaticEnv`).
- Keep only bootstrap env vars (`CONF_PATH`, provider selector, encryption key) in the env-var table above.

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
- Build args: <the actual `ARG`s from the Dockerfile — e.g. TARGETARCH, if present>
- Build command: <the real `make` target(s) local-dev.sh runs for this service, derived from the Makefile + local-dev.sh — do not assume a target name>
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
