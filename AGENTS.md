# AGENTS.md

This repository contains the NVIDIA Run:ai documentation.

For contribution mechanics — cloning, the pull request workflow, DCO sign-off,
branching, and code review — see `CONTRIBUTING.md`. This file covers
documentation-specific guidance only; for coding conventions across the rest
of the monorepo (services, Helm chart, CLI, etc.), see `CLAUDE.md` at the
repository root and the per-service `CLAUDE.md` files it indexes.

## Secrets and credentials

Never commit secrets, API keys, tokens, or credentials — real or
placeholder-looking-real — to this repository, including in `docs/` examples.
When a doc needs to show a credential-shaped value (a token, a password, a
connection string), use an obviously-fake placeholder (e.g. `<your-api-key>`
or `sk-...redacted...`), not a plausible-looking fabricated one. If you
discover a secret already committed, do not open a public PR referencing it —
report it per [`SECURITY.md`](SECURITY.md) instead, since removing it from a
file doesn't remove it from git history.

## Verify before opening a PR

There is no single repo-wide gate command — the root `Makefile`'s `build`,
`test`, and `lint` targets only echo `NA` (its `validate` target is real), and
the full pipeline runs remotely, reporting back on the PR as the
`Validate Successful Run` check.

Start with the service you changed. This is what CI runs for you — the
`targets` job in `.github/workflows/pr-fast.yaml` discovers the services a PR
touches and runs all four targets in each, using that service's own Makefile:

```bash
cd <service-dir>            # e.g. runai-cluster/cluster/anyworkload-controller
make build test lint validate
```

Then run whichever language-level checks apply, matching the jobs in
`.github/workflows/pr-checks.yaml` at the versions CI pins:

```bash
# Go
golangci-lint run            # CI pins v2.9.0; matches the go-lint job
govulncheck ./...            # matches the govulncheck job
gofmt -l .                   # formatting is not optional in this repo

# Python
ruff check <changed-files>   # CI pins ruff 0.15.15

# JS/TS — from the package directory
npx eslint <changed-files>

# Helm chart changes
helm dependency update control-plane/control-plane-chart
helm lint control-plane/control-plane-chart
helm template control-plane/control-plane-chart > /dev/null

# GitHub Actions workflow changes
actionlint

# Docs (docs/ or fern/) — `fern docs dev` only starts a preview server
# and validates nothing; `fern check` is the actual CI check.
fern check
python3 ./tools/check-fern-nav.py fern

# CODEOWNERS changes — real target, but no CI job runs it; local check only
make validate                # from the repository root
```

Every commit must be signed off (`git commit -s`) — the DCO check gates fork
PRs. See `CONTRIBUTING.md`.

## Coding conventions

These apply monorepo-wide; a service's own `CLAUDE.md` wins where it differs.

- **Match the surrounding code.** Mirror the file's existing naming, comment
  density, error-handling style, and test idiom rather than importing
  conventions from elsewhere.
- **Go:** standard `gofmt` formatting, no exceptions. Wrap errors with
  `fmt.Errorf("...: %w", err)` — never discard an error silently. Controller
  and webhook tests in this repo use Ginkgo/Gomega; follow the existing suite
  structure in the package you are editing.
- **Every source file carries the SPDX header** used across the repo:
  `SPDX-FileCopyrightText` and `SPDX-License-Identifier: Apache-2.0`.
- **Comments explain why, not what.** Do not add comments that restate the
  code, and do not leave commented-out code behind.
- **No new dependencies without a reason.** Prefer the standard library or an
  already-vendored package; a new module in `go.mod` needs justification in the
  PR description.

### Preferred and deprecated patterns

```go
// Preferred — error is wrapped with context the caller can act on.
if err := r.Get(ctx, key, &workload); err != nil {
    return fmt.Errorf("get workload %s: %w", key, err)
}

// Deprecated — the original error is destroyed, so callers cannot use
// errors.Is/As and the log loses the cause.
if err := r.Get(ctx, key, &workload); err != nil {
    return errors.New("failed to get workload")
}
```

```go
// Preferred — table-driven test, each case named and independently reported.
for _, tc := range []struct {
    name string
    in   string
    want bool
}{
    {name: "empty input", in: "", want: false},
    {name: "valid name", in: "runai-workload", want: true},
} {
    t.Run(tc.name, func(t *testing.T) { ... })
}

// Deprecated — a single test with several unnamed assertions. The first
// failure hides the rest and the output does not say which case broke.
func TestValidate(t *testing.T) {
    assert(validate(""))
    assert(validate("runai-workload"))
}
```

## Products in Scope

| Product | Slug |
|---|---|
| Self-hosted | `self-hosted` |
| SaaS | `saas` |
| Multi-tenant | `multi-tenant` |

## Repository Layout

| Path | Purpose |
|---|---|
| `docs/` | Markdown source pages |
| `fern/` | Fern navigation config (`docs.yml`, product files, components) |
| `fern/apis/` | OpenAPI specs for REST API reference |
| `docs/.agents/skills/` | Repository-local documentation skills |

## Sources of Truth

- **Product behavior, CLI flags, API fields, and defaults**: Run:ai engineering
  source code. Verify all technical claims against source before documenting.
  Do not invent values.
- **Navigation structure**: `fern/docs.yml` and `fern/products/`.
- **Writing standards and repo style**: `docs/AGENTS.md`.
- **Terminology**: `docs/terminology.md`.
- **Auto-generated content**: CLI reference (`docs/reference/cli/`) and REST API
  reference are generated by CI pipelines from source elsewhere in this repo.
  Do not hand-edit the generated output.

## Validated Commands

Preview the documentation site locally:

```bash
fern docs dev
```

## Documentation

Do not proactively edit anything under `docs/` (the user-facing product
documentation) as a side effect of a code change, even if the change affects
a user-visible surface. Only update `docs/` when explicitly asked to in the
task.

When explicitly asked to update `docs/`:

1. Read `docs/AGENTS.md` before editing.
2. Locate and update the affected pages under `docs/`, and run `fern docs dev`
   to validate the result.

### Document Only What Has Shipped

Document behavior only when you can point to it as finished and shipped in
the current source code or a released artifact. An accepted GitHub issue or
merged design decision describes intent, not shipped state — use it to
identify what to verify against the source, not as evidence on its own that
behavior is final. Do not document behavior that is planned, in progress, or
hidden behind a disabled-by-default flag. If a code path looks unfinished or
you cannot tell whether a behavior is final, flag it in the PR description
rather than documenting it as current product behavior.

Documentation changes go through the same review as any other change in this
repository — see Code review in `CONTRIBUTING.md`.
