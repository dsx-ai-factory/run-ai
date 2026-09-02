# fielder

`fielder` is a **Python code-generation tool** that builds Run:ai's workload
"field spec" files from a single declarative source, [`config.yaml`](config.yaml).
It is not a service and produces no binary or container image — developers run it
locally to regenerate the generated YAML specs, and CI runs it in `--validate`
mode to ensure the committed files are still in sync with `config.yaml`.

The workload field specs describe every field of a workload across three facets —
**Spec**, **Defaults** and **Rules** — for the various workload target files
(e.g. distributed inference, policy defaults/rules). Keeping these hand-editable
YAML files consistent by hand is error-prone, so `fielder` derives them all from
one config plus a set of skeletons.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Configuration](#configuration)
- [Excel output](#excel-output)

## Overview

The tool is responsible for:

- **Targets** — each entry under `targets:` in `config.yaml` is a file `fielder`
  generates. A target declares its `type` (`spec` / `defaults` / `rules`), output
  `path` (a `.yaml` file, resolved relative to the parent `control-plane/`
  directory), an optional `skeleton`, and field-extraction/renaming rules.
- **Fields** — each entry under `fields:` describes a field's content. A field can
  carry multiple **variants** so a single logical field can render differently in
  two targets (used for "exotic" fields, e.g. some DistributedInference fields).
- **Skeletons** — the base YAML fragments in [`skeletons/`](skeletons) that
  targets are built on top of (e.g. `DistributedInferenceCommonSpec.yaml`,
  `InferencePolicyRulesV2.yaml`, `Empty.yaml`).
- **Build vs validate** — in build mode it writes the target files and exports an
  Excel view; in validate mode it only diffs and fails if any target is out of sync.

## Repository layout

```text
.
├── __main__.py          # Package entry point → build.main()
├── build.py             # CLI: parse --validate, load config.yaml, drive targets
├── config.py            # Config dataclass (fields + targets) + Excel export
├── target.py            # Target dataclass, WorkingMode (BUILD/VALIDATE), file build/diff
├── field.py             # Field / Variant dataclasses
├── config.yaml          # THE source of truth: all targets + fields (~85 KB)
├── skeletons/           # Base YAML fragments targets are generated from
├── requirements.txt     # pyyaml, openpyxl
├── Makefile
└── artifact-metadata.json
```

## Building and running

Built through the [`Makefile`](Makefile), which includes
`../../build/makefile/control-plane.mk` and `index.mk`. All targets run the tool
inside a `python:3.13` container (the shared `DOCKER_PYTHON_COMMAND`, which
`pip install -r requirements.txt` then runs `python .`). `PYTHON_REQ_PATH=.` and
the package's `__main__.py` invokes `build.main()`.

| Target | Description |
|--------|-------------|
| `make rebuild` | Run the tool in **build** mode: regenerate every target file from `config.yaml` and export `config.xlsx`. This is what a developer runs, committing the regenerated files. |
| `make validate` | Run `rebuild` with `PYTHON_SCRIPT_FILE_NAME=". --validate"` — **validate** mode: diff targets against `config.yaml` and exit non-zero if any file is out of sync (CI uses this). |
| `make build` | No-op — building the generated files is a developer action, CI only validates. |
| `make test` / `make lint` / `make purify` | No-ops (kept for the standard target contract). |

Locally, without the container:

```bash
cd control-plane/fielder
make rebuild        # regenerate the field spec files + config.xlsx
make validate       # check the committed files match config.yaml
```

There is **no image** for `fielder`, so it is not part of `local-dev.sh`'s build
loop and has no `runai-local/fielder` tag. Its output is checked-in YAML consumed
by the rest of the control plane, not a deployed artifact.

## Configuration

`fielder` has no runtime env-var configuration — everything is driven by
[`config.yaml`](config.yaml), which has two top-level sections:

- **`targets`** — a map keyed by target id. Each target has: `type`
  (`spec`/`defaults`/`rules`), `path` (output `.yaml`, must end in `.yaml`),
  optional `skeleton` (a file under `skeletons/`), `comment`, `nullable`, and
  `extract_fields` / name-change rules.
- **`fields`** — a map of field definitions, each optionally with multiple
  `variants` so one field can render differently per target.

Output paths are resolved relative to `BASE_DIR = ".."` (the `control-plane/`
directory), e.g. under `infra/openapi/docs/schemas/workloads_spec`.

## Excel output

Each `make rebuild` also generates `config.xlsx`, a spreadsheet view of every
target's content. In the sheet:

- **S** = Spec
- **D** = Defaults
- **R** = Rules
- **X** = Special targets

Installing the IntelliJ Excel plugin lets you view `config.xlsx` alongside the
code in a PR. `*.xlsx` is git-ignored.
