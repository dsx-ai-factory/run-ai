# openapi/build — Build Context

**Type:** none (`artifact_type: "none"` — no image, no binary)
**Service name:** openapi/build
**Artifact:** processed spec files under `_build/` (`cache_artifacts: true`, `path_to_artifacts: _build/`)

`openapi/build` builds Run:ai's **public OpenAPI docs spec** for local/CI use.
It is NOT a runtime service and does NOT build a container image, and it does
NOT publish anywhere externally. Node.js 22 + the merge/redocly tooling run inside a `python:3.13` container invoked by the
shared makefile targets.

## Make targets

Targets come from the service `Makefile` plus shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, `openapi.mk`).

| Target | What it does |
|--------|-------------|
| `build` | `openapi-build-public-spec` (regenerate merged public + system-scoped specs into `_build/`), then copy `_build/runaiApi.swagger.yaml` → `_build/runai-public.yaml` |
| `prepare-specs` | `openapi-build-public-spec` then copy public + system-scoped specs into `_build/` |
| `test` | No-op |
| `lint` | No-op |
| `validate` | `validate-openapi-public-spec` + `validate-tags` + `validate-openapi-spec-python` + `validate-openapi-backwards-compatibility` (all `continue_on_error` per metadata) |

## Key build notes

- **No image is built.** There is no Dockerfile and no `runai-local/openapi-build`
  tag; this service is not in `local-dev.sh`'s `CP_GO_SERVICES` and is not
  deployed as a workload. The deployed control plane serves the OpenAPI yaml
  spec download at `/api/docs` (from the frontend).
- **Build == spec merging.** `build` merges service swagger via
  `openapi-merge-cli`, bundles/filters with redocly (`config/redocly*.yaml` +
  `plugins/internal-filters.js`, `plugins/system-scoped-filters.js`) to strip
  `x-internal` endpoints and internal enum values; there is no further post-processing.
- **No external publish step.** GitBook publishing (the old `make extra`
  target and `scripts/gitbook/`)
  was removed (RUN-42271) — this module only produces local build artifacts
  now.
- **Source specs:** `SPEC_FILE_SRC=../infra/openapi/specs/runai-public.yaml` and
  `SPEC_FILE_SRC_SYSTEM=../infra/openapi/specs/runai-system-scoped.yaml`.
- **Validation gates:** no `x-internal`/`x-authorization` tags in the built spec;
  spec generates a Python client; `oasdiff`
  backwards-compatibility check against the base branch.
- **Node deps:** `scripts/cli/package.json` (openapi-merge-cli / redocly).

## local-dev.sh iteration

This service is not built or deployed by `local-dev.sh`. Run its build/tests directly:
```bash
cd control-plane/infra/openapi/build && make build && make test
```
The general CP iteration flag remains:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
