# openapi/build

`openapi/build` is the control-plane component that **builds Run:ai's public
OpenAPI documentation spec**. It merges the individual service swagger files
into a single public spec, and filters out internal endpoints/fields
for local/CI use.

This module does **not** build or ship a container image, and it does not
serve docs at runtime. Its `artifact_type` is `none`: the build produces
processed spec files under `_build/`. In a deployed control plane the frontend
serves the OpenAPI yaml spec download at `/api/docs`.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Validation](#validation)
- [Configuration](#configuration)
- [How to add a new service](#how-to-add-a-new-service)

## Overview

The module is responsible for:

- **Merging** individual service swagger files into a unified public spec
  (driven by `config/merge-config.json`, with a system-scoped variant in
  `config/merge-config-system-system-scoped.json`).
- **Filtering** the merged spec so internal detail never reaches customers —
  the redocly bundle applies the decorators in `plugins/internal-filters.js`
  and `plugins/system-scoped-filters.js` (removing `x-internal` endpoints and
  stripping internal enum values), configured by `config/redocly.yaml`.

## Repository layout

```text
.
├── config/                 # openapi-merge + redocly configs (public + system-scoped)
│   ├── merge-config.json
│   ├── merge-config-system-system-scoped.json
│   ├── redocly.yaml
│   ├── redocly-system-scoped.yaml
│   └── lint.yaml
├── plugins/                # redocly decorator plugins (internal + system-scoped filters)
├── swagger/                # rootApi.yaml / rootApiSystemScoped.yaml / rootApiTags.yaml
├── scripts/
│   └── cli/                # Node deps (openapi-merge-cli / redocly) via package.json
├── _build/                 # Build output: runaiApi.swagger.yaml, runai-public.yaml, runai-system-scoped.yaml
├── NOTICE
├── Makefile
└── artifact-metadata.json
```

## Building and running

Built through the [`Makefile`](Makefile), which includes
`../../build/makefile/control-plane.mk` and `index.mk`. There is **no Go binary
and no Docker image** — the "build" is spec merging and filtering. Node.js 22 and the merge
/ redocly tooling run inside a `python:3.13` container invoked by the shared
makefile targets.

| Target | Description |
|--------|-------------|
| `make build` | `openapi-build-public-spec` (regenerate the merged public + system-scoped specs into `_build/`), then copy `_build/runaiApi.swagger.yaml` → `_build/runai-public.yaml` |
| `make prepare-specs` | `openapi-build-public-spec`, then copy the public and system-scoped specs into `_build/` |

No local image is produced, so `openapi/build` is not part of `local-dev.sh`'s
`CP_GO_SERVICES` build loop and no `runai-local/openapi-build` tag exists. The
build is run in CI (and manually) to regenerate the docs spec.

## Testing

| Target | Description |
|--------|-------------|
| `make test` | No-op (kept for the standard target contract) |
| `make lint` | No-op (kept for the standard target contract) |

## Validation

`make validate` runs the full documentation-quality gate (any step may continue
on error per `artifact-metadata.json`):

| Target | Description |
|--------|-------------|
| `make validate` | `validate-openapi-public-spec` + `validate-tags` + `validate-openapi-spec-python` + `validate-openapi-backwards-compatibility` |
| `validate-tags` | Fail if the built spec still contains `x-internal` or `x-authorization` tags |
| `validate-openapi-spec-python` | Generate a Python client from the spec via OpenAPI Generator as a smoke test |
| `validate-openapi-backwards-compatibility` | Run `oasdiff` against the base branch to catch breaking public-API changes |

## Configuration

`openapi/build` is a build tool, not a runtime service, so it has no service
env-var config. Its behaviour is driven by makefile variables and files:

| Variable / input | Description |
|------------------|-------------|
| `SPEC_FILE_SRC` | Source merged public spec (`../infra/openapi/specs/runai-public.yaml`) |
| `SPEC_FILE_SRC_SYSTEM` | Source system-scoped spec (`../infra/openapi/specs/runai-system-scoped.yaml`) |
| `config/merge-config*.json` | Which service swagger files are merged and in what order |
| `config/redocly*.yaml` + `plugins/` | Bundle/filter rules (internal endpoint and enum stripping) |
| `swagger/rootApi*.yaml` | Root API doc |

## How to add a new service

- Update `config/merge-config.json` with the new swagger file path.
- Run `make build` to verify the merged output, then `make validate`.
