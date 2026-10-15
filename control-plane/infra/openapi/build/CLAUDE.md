# openapi/build — Build Context

**Type:** none (`artifact_type: "none"` — no image, no binary)
**Team:** Sphinx (`artifact-metadata.json`)
**Service name:** openapi/build
**Artifact:** processed spec files under `_build/` (`cache_artifacts: true`, `path_to_artifacts: _build/`)

`openapi/build` builds Run:ai's **public OpenAPI docs spec** for local/CI use.
It is NOT a runtime service and does NOT build a container image, and it does
NOT publish anywhere externally. Node.js 22 + the merge/redocly tooling and
the Python pipeline all run inside a `python:3.13` container invoked by the
shared makefile targets.

## Make targets

Targets come from the service `Makefile` plus shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, `openapi.mk`).

| Target | What it does |
|--------|-------------|
| `build` | `get_latest_tag` + `openapi-build-public-spec` (regenerate merged public + system-scoped specs into `_build/`), copy `_build/runaiApi.swagger.yaml` → `_build/runai-public.yaml`, then run `scripts/main_build.py` in the python container |
| `prepare-specs` | `openapi-build-public-spec` then copy public + system-scoped specs into `_build/` |
| `get_latest_tag` | Resolve latest release tag → `LATEST_TAG` |
| `test` | Run the pytest suite `scripts/tests/run_all_tests.py` in the python container |
| `lint` | No-op |
| `validate` | `validate-openapi-public-spec` + `validate-tags` + `validate-docs-links` + `validate-openapi-spec-python` + `validate-openapi-backwards-compatibility` (all `continue_on_error` per metadata) |

## Key build notes

- **No image is built.** There is no Dockerfile and no `runai-local/openapi-build`
  tag; this service is not in `local-dev.sh`'s `CP_GO_SERVICES` and is not
  deployed as a workload. The deployed control plane serves the OpenAPI yaml
  spec download at `/api/docs` (from the frontend).
- **Build == spec processing.** `build` merges service swagger via
  `openapi-merge-cli`, bundles/filters with redocly (`config/redocly*.yaml` +
  `plugins/internal-filters.js`, `plugins/system-scoped-filters.js`) to strip
  `x-internal` endpoints and internal enum values, then post-processes:
  documentation link rewriting (`core/link_replacer.py`), version stamping,
  and branding normalization (`core/file_processor.py`).
- **No external publish step.** GitBook publishing (the old `make extra`
  target, `scripts/gitbook/`, `core/gitbook_api.py`, `scripts/main_extra.py`)
  was removed (RUN-42271) — this module only produces local build artifacts
  now.
- **Source specs:** `SPEC_FILE_SRC=../infra/openapi/specs/runai-public.yaml` and
  `SPEC_FILE_SRC_SYSTEM=../infra/openapi/specs/runai-system-scoped.yaml`.
- **Validation gates:** no `x-internal`/`x-authorization` tags in the built spec;
  all docs links carry `/latest/`; spec generates a Python client; `oasdiff`
  backwards-compatibility check against the base branch.
- **Python deps:** `scripts/requirements.txt` (`pytest`, `requests`); Node deps in
  `scripts/cli/package.json` (openapi-merge-cli / redocly).

## local-dev.sh iteration

This service is not built or deployed by `local-dev.sh`. Run its build/tests directly:
```bash
cd control-plane/infra/openapi/build && make build && make test
```
The general CP iteration flag remains:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
