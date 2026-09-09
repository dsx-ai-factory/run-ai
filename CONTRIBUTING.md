# Contributing to Run:ai

Thanks for your interest in contributing. This guide covers how to file issues, propose changes, and submit pull requests.

By participating, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md). To report a security vulnerability, follow the private process in [`SECURITY.md`](SECURITY.md) — **do not open a public issue**.

---

## Ways to contribute

1. **Report a bug, request a feature, or flag a documentation issue.** Open a [GitHub issue](https://github.com/dsx-ai-factory/run-ai/issues/new/choose) using the appropriate template. Maintainers will triage and label it.
2. **Discuss an idea before you build it.** For non-trivial changes, open a feature issue first so the design can be agreed on before you write code. For support and usage questions, open an issue.
3. **Implement a fix or feature.** Pick up an issue labelled `good first issue` or `help wanted`, comment to claim it, then follow the workflow below.

---

## Getting started

```bash
# Clone
git clone https://github.com/dsx-ai-factory/run-ai.git
cd runai

# Build (Go modules, no private dependencies — fork-safe)
make build

# Run unit tests
make test
```

Component-specific setup lives in each component's `README.md`. For end-to-end testing on a real cluster, see the per-component docs.

---

## Pull request workflow

1. **Fork** the repository to your own GitHub account.
2. **Branch** from `master` (or from the live release branch you are targeting — see branching policy below). Use a descriptive name.
3. **Make focused changes.** One concern per PR. Add tests for new behavior. Update docs if user-facing.
4. **Sign off every commit** with `git commit -s` (see [DCO](#developer-certificate-of-origin-dco) below). This applies to **fork PRs only**; internal Run:ai contributors are not DCO-checked.
5. **Open the PR** against `master` (preferred) or the relevant `v2.X` release branch. Fill in the PR template, reference the issue with `Fixes #123`, and add a `/kind` line.
6. **Wait for CI.** Layer 1 checks (lint, DCO for forks, CodeQL diff, govulncheck) run automatically. The full build pipeline and end-to-end tests run after a maintainer reviews and types `/ok-to-test` on the PR.
7. **Address review feedback.** A maintainer approval plus a green `Validate Successful Run` status are both required for merge.

---

## Branching policy

Run:ai accepts pull requests targeting `master` or any live release branch (today: `v2.22`, `v2.23`, `v2.24`, `v2.25`, `v2.26`).

**Master-first is preferred.** Open your PR against `master` so the fix flows to every supported release via the `backport/v2.X` label that a maintainer applies after merge. This avoids "fixed in v2.24 but forgotten on v2.26" regressions.

**Direct release-branch PRs are accepted** when you can only reproduce or test on the version you're running. Open the PR against that release branch (for example `v2.26`); during review, a maintainer will help you decide whether the fix should also be cherry-picked to `master` and to other live releases.

**Don't** open PRs against EOL branches (`v2.21` and earlier). They are not maintained.

**How backports work:** when a maintainer applies a `backport/v2.X` label to your merged PR, an automated workflow opens a follow-up PR cherry-picking the merge commit onto each labelled release branch. If the cherry-pick conflicts, the bot leaves a `needs-manual-backport/v2.X` label and a maintainer takes it from there. You don't need to open backport PRs yourself.

---

## CI for external contributors — what runs and when

We split CI into two layers:

- **Layer 1 — fast, automatic, no secrets.** Runs on every fork PR as soon as it is opened: DCO check, lint (Go, Helm, CODEOWNERS, workflows), CodeQL diff scan, govulncheck. These checks post results back to your PR within a few minutes. They are informational — they help you fix obvious issues before a maintainer looks at the PR, but they do not block merge by themselves.
- **Layer 2 — full build and end-to-end tests, has secrets.** This is the pipeline that produces the required `Validate Successful Run` status check, and it only runs after a maintainer reviews your PR and types `/ok-to-test` in a comment. End-to-end tests take 30–90 minutes. A failed Layer 2 run blocks merge — for everyone, including internal contributors. If a maintainer hasn't reviewed your PR within a few business days and you'd like the full pipeline to run, feel free to ping in the PR.

If a Layer 2 step fails, the workflow logs from `runai-actions` (the private repository where the full pipeline lives) are not directly visible to external contributors. Look at the PR labels (`pipeline/build-failed`, `pipeline/e2e-failed`, `pipeline/jira-failed`) for a hint about which stage broke, and a maintainer will help you triage from there.

### Running the checks locally

Layer 1 is the part you can fully reproduce on your machine, and doing so before you push is the fastest way to a green PR. There is deliberately **no single repo-wide gate command**: this is a monorepo of independent modules, the root `Makefile` targets are placeholders, and Layer 2 (full build and end-to-end tests) needs secrets and a cluster, so it cannot run locally at all.

Run the checks matching what you changed — these mirror the corresponding jobs in [`.github/workflows/pr-checks.yaml`](.github/workflows/pr-checks.yaml), at the versions CI pins:

| You changed | Run | Mirrors CI job |
|---|---|---|
| Go code | `cd <module> && go build ./... && go test ./... && gofmt -l .`<br>`golangci-lint run` (v2.9.0) | `go-lint`, `govulncheck` |
| Python | `ruff check <files>` (0.15.15) | `py-lint` |
| JS/TS | `npx eslint <files>` | `js-lint` |
| Helm charts | `helm dependency update control-plane/control-plane-chart`<br>`helm lint control-plane/control-plane-chart`<br>`helm template control-plane/control-plane-chart > /dev/null` | `helm` |
| GitHub Actions workflows | `actionlint` | `actionlint.yaml` |
| `CODEOWNERS` | `make validate` | `validate-codeowners` |
| `docs/` or `fern/` | `fern docs dev` | `fern-docs-ci.yaml` |

Each service directory also has its own Makefile with `build`, `test`, `lint`, and `validate` targets scoped to that service — use those when working inside a single component.

---

## Code review

- Branch protection requires **at least one approval** and a green `Validate Successful Run` check before merge.
- `CODEOWNERS` files at the root and under `.github/` and `docs/` automatically request review from the right team for each affected path.
- For non-trivial changes, expect more than one reviewer to weigh in — this is a team norm, not a tool gate. Be patient and responsive.

---

## AI-assisted contributions

Using AI coding assistants (e.g. Claude Code, GitHub Copilot) to help write a
contribution is fine — the same review bar applies regardless of how the
change was produced:

- You are responsible for every line you submit. Read and understand the
  diff before opening a PR; don't submit AI output you haven't verified.
- The DCO sign-off (below) still certifies that you have the right to submit
  the change under this project's license, exactly as it would for
  hand-written code.
- Don't paste proprietary, confidential, or export-controlled material into
  third-party AI tools as part of preparing a contribution.
- PRs that are largely unreviewed AI output — inconsistent with the
  surrounding code, containing invented APIs, or with commit messages that
  don't reflect the actual change — will be asked to be revised or closed,
  same as any other low-quality submission.

---

## Developer Certificate of Origin (DCO)

External contributions to this repository require a Developer Certificate of Origin sign-off on every commit. Signing off certifies that the contribution is your own work, or that you have the right to submit it under the project's license. The full text of the DCO follows:

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
1 Letterman Drive
Suite D4700
San Francisco, CA, 94129

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.

Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

To sign off, add the `-s` flag to your commit:

```bash
git commit -s -m "Fix node-pool selector when foo is bar"
```

This appends a line to your commit message:

```
Signed-off-by: Your Name <your.email@example.com>
```

The name and email must match the Git author identity on the commit. If you missed the sign-off, amend with:

```bash
git commit --amend -s
git push --force-with-lease
```

A `dco` check runs on every fork PR and lists any unsigned commits inline.

**Internal contributors:** the DCO check does not run on commits authored from `@run.ai` and `@nvidia.com` accounts. Internal contributions are governed by the existing employee IP agreement.

---

## Commit messages and PR titles

Write clear, present-tense messages explaining **why** the change is needed, not just what it does. A good PR title is a single sentence under ~70 characters.

We do not enforce Conventional Commits format. Release notes are generated from PR titles and labels via GitHub's auto-generated release notes — clear titles make better release notes.

---

## Getting help

- [GitHub Issues](https://github.com/dsx-ai-factory/run-ai/issues) — bugs, feature requests, documentation issues, usage questions
- [SECURITY.md](SECURITY.md) — private vulnerability reports
- [SUPPORT.md](SUPPORT.md) — which versions are supported

Thank you for contributing.
