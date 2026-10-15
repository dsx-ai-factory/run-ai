# openapi/build

`openapi/build` is the control-plane component that **builds Run:ai's public
OpenAPI documentation spec**. It merges the individual service swagger files
into a single public spec, filters out internal endpoints/fields, and
post-processes the result (version stamping, branding, documentation link
rewriting) for local/CI use. It is owned by the **Sphinx** team.

This module does **not** build or ship a container image, and it does not
serve docs at runtime. Its `artifact_type` is `none`: the build produces
processed spec files under `_build/`. In a deployed control plane the public
docs are served externally — on SaaS `/api/docs` redirects to the docs site
configured via the chart value `redoc.docsSite.url`; on self-hosted no
API-docs endpoint is served.

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
- **Post-processing** links and versions via the Python pipeline in
  `scripts/` (doc-link rewriting to `/latest/`, version stamping, branding
  normalization).

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
├── scripts/                # Python spec build pipeline
│   ├── main_build.py       # process the built spec for local use
│   ├── core/               # processors, link replacer, version manager, file processor
│   ├── cli/                # Node deps (openapi-merge-cli / redocly) via package.json
│   ├── tests/               # pytest suite (run_all_tests.py)
│   └── requirements.txt    # pytest, requests
├── _build/                 # Build output: runaiApi.swagger.yaml, runai-public.yaml, runai-system-scoped.yaml
├── NOTICE
├── Makefile
└── artifact-metadata.json
```

## Building and running

Built through the [`Makefile`](Makefile), which includes
`../../build/makefile/control-plane.mk` and `index.mk`. There is **no Go binary
and no Docker image** — the "build" is spec processing. Node.js 22 and the merge
/ redocly tooling run inside a `python:3.13` container invoked by the shared
makefile targets.

| Target | Description |
|--------|-------------|
| `make build` | `get_latest_tag` + `openapi-build-public-spec` (regenerate the merged public + system-scoped specs into `_build/`), then copy `_build/runaiApi.swagger.yaml` → `_build/runai-public.yaml` and run the modular Python build script (`scripts/main_build.py`) inside the python container |
| `make prepare-specs` | `openapi-build-public-spec`, then copy the public and system-scoped specs into `_build/` |
| `make get_latest_tag` | Fetch git tags and resolve the latest release tag (`LATEST_TAG`) used to stamp doc versions |

No local image is produced, so `openapi/build` is not part of `local-dev.sh`'s
`CP_GO_SERVICES` build loop and no `runai-local/openapi-build` tag exists. The
build is run in CI (and manually) to regenerate the docs spec.

## Testing

| Target | Description |
|--------|-------------|
| `make test` | Run the spec-processing pytest suite: installs `scripts/requirements.txt` and runs `scripts/tests/run_all_tests.py` inside the privileged python container |
| `make lint` | No-op (kept for the standard target contract) |

The pytest suite under `scripts/tests/` covers the config loader, file
processor, link replacer, and version manager.

## Validation

`make validate` runs the full documentation-quality gate (any step may continue
on error per `artifact-metadata.json`):

| Target | Description |
|--------|-------------|
| `make validate` | `validate-openapi-public-spec` + `validate-tags` + `validate-docs-links` + `validate-openapi-spec-python` + `validate-openapi-backwards-compatibility` |
| `validate-tags` | Fail if the built spec still contains `x-internal` or `x-authorization` tags |
| `validate-docs-links` | Fail if any `run-ai-docs.nvidia.com/self-hosted/...` link is missing the `/latest/` segment |
| `validate-openapi-spec-python` | Generate a Python client from the spec via OpenAPI Generator as a smoke test |
| `validate-openapi-backwards-compatibility` | Run `oasdiff` against the base branch to catch breaking public-API changes |

## Configuration

`openapi/build` is a build tool, not a runtime service, so it has no service
env-var config. Its behaviour is driven by makefile variables and files:

| Variable / input | Description |
|------------------|-------------|
| `SPEC_FILE_SRC` | Source merged public spec (`../infra/openapi/specs/runai-public.yaml`) |
| `SPEC_FILE_SRC_SYSTEM` | Source system-scoped spec (`../infra/openapi/specs/runai-system-scoped.yaml`) |
| `VERSION` / `LATEST_TAG` / `ENVIRONMENT` | Version + environment stamped into the processed spec |
| `config/merge-config*.json` | Which service swagger files are merged and in what order |
| `config/redocly*.yaml` + `plugins/` | Bundle/filter rules (internal endpoint and enum stripping) |
| `swagger/rootApi*.yaml` | Root API doc |

The chart value `redoc.docsSite.url` (in the control-plane chart) points the
deployed control plane at the external docs site; it is not consumed here.

## How to add a new service

- Update `config/merge-config.json` with the new swagger file path.
- Run `make build` to verify the merged output, then `make validate`.
