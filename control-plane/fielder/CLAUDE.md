# fielder — Build Context

**Type:** (unset in `artifact-metadata.json`) — effectively a Python code-generation tool, no image, no binary
**Service name:** fielder
**Language:** Python 3.13 (deps: `pyyaml`, `openpyxl`)

`fielder` generates Run:ai's workload field-spec YAML files (Spec / Defaults /
Rules) from `config.yaml` + `skeletons/`. Developers run it to regenerate the
files; CI runs it in validate mode to confirm they are in sync.

## Make targets

Targets come from the service `Makefile` plus shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, `base.mk`).

| Target | What it does |
|--------|-------------|
| `rebuild` | Run `DOCKER_PYTHON_COMMAND` (python:3.13 container: `pip install -r requirements.txt` then `python .`) in **build** mode — regenerate all targets + export `config.xlsx` |
| `validate` | `make rebuild PYTHON_SCRIPT_FILE_NAME=". --validate"` — validate mode, fails if generated files differ from `config.yaml` |
| `build` | No-op (CI only validates; devs run `rebuild`) |
| `test` / `lint` / `purify` | No-ops |

## Key build notes

- **No container image, no Go binary.** `artifact_type` is empty; `fielder` is not
  in `local-dev.sh`'s `CP_GO_SERVICES` and is not deployed. There is no
  `runai-local/fielder` tag.
- **Entry point:** `python .` runs `__main__.py` → `build.main()`. `--validate`
  switches `WorkingMode.BUILD` → `WorkingMode.VALIDATE` (diff-only, no writes).
- **Source of truth:** `config.yaml` (~85 KB), two sections `targets:` and
  `fields:`. Targets = files to generate (`type` spec/defaults/rules, `.yaml`
  `path`, optional `skeleton`, extract/rename rules). Fields may declare multiple
  `variants` to render differently per target.
- **Output location:** paths resolve relative to `BASE_DIR = ".."` (the
  `control-plane/` dir), e.g. `infra/openapi/docs/schemas/workloads_spec`.
- **Excel:** `config.xlsx` regenerated on every `rebuild` (git-ignored);
  S=Spec, D=Defaults, R=Rules, X=Special.
- **Container runner:** `DOCKER_PYTHON_COMMAND` from `build/makefile/base.mk`
  (`DOCKER_PYTHON_IMG=${DOCKERHUB_PROXY}library/python:3.13`, `PYTHON_REQ_PATH=.`).

## local-dev.sh iteration

`fielder` is not built or deployed by `local-dev.sh`. Run it directly:
```bash
cd control-plane/fielder && make rebuild   # or: make validate
```
