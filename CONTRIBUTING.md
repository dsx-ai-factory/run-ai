# Contributing to Run:ai

Thanks for your interest in contributing. This guide covers how to file issues, propose changes, and submit pull requests.

By participating, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md). To report a security vulnerability, follow the private process in [Security](SECURITY.md): **do not open a public issue**.

---

## Ways to Contribute

- **You want to report a bug, feature idea, or documentation issue** - File a [GitHub issue](https://github.com/dsx-ai-factory/run-ai/issues/new/choose) using the appropriate template. Maintainers will triage and label it.
- **You want to ask a usage question** - Start a [Discussion](https://github.com/dsx-ai-factory/run-ai/discussions). Questions are not filed as issues, so the issue tracker stays a list of defects and requests.
- **You want to propose a new feature and implement it** - File a feature-request issue describing your intended feature, and discuss the design there. Once the plan looks good, implement it following the Contribution Workflow below.
- **You want to implement a feature or bug fix for an outstanding issue** - Pick up an issue labelled `good first issue` or `help wanted`, comment to claim it, then follow the Contribution Workflow below.

### How Issue Priority Is Decided

Maintainers communicate priority through the **Priority field** on the issue, not through a custom label, so the field is the single place to look, and it stays accurate when an issue moves between projects and milestones.

Priority is set by maintainers during triage (see [Maintainers](MAINTAINERS.md)) and weighs:

- **Severity and blast radius** - Does it break a supported release, cause data loss, or degrade scheduling for every tenant, versus affecting one optional path?
- **Whether a workaround exists** - And how costly that workaround is.
- **Affected versions** - Regressions on a currently supported release (see [Support](SUPPORT.md)) outrank issues on end-of-life branches, which are not accepted at all except for critical security issues.
- **Breadth of impact** - How many users or clusters hit it.
- **Cost and risk of the fix** - Weighed against the above.

**If you think an issue is mis-prioritised, say so in a comment on that issue.** Context that changes the calculus is the most useful thing you can add: your NVIDIA Run:ai and Kubernetes versions, how many clusters or users are affected, whether you have a workaround and what it costs you, and a reproduction if the issue lacks one. Maintainers re-triage on new information: a clear impact report is far more effective than a "+1". Priority is a maintainer decision and is not guaranteed to change, but the reasoning will be stated on the issue.

---

## Getting Started

1. **Clone the repository.**

   ```bash
   git clone https://github.com/dsx-ai-factory/run-ai.git
   cd run-ai
   ```

2. **Build and run the full platform.** `local-dev.sh` builds the control plane and cluster services from source and deploys them to a local Kubernetes (Kind) cluster. See [Local Development Guide](local-dev.md) for prerequisites and all available flags.

   ```bash
   ./local-dev.sh
   ```

3. **Run tests.** Tests run per-service, not from the repo root. Each service directory under `control-plane/` and `runai-cluster/` has its own `README.md` and `Makefile` describing how to build, test, and validate it.

   ```bash
   cd <service-dir>
   make test
   ```

---

## Contribution Workflow

Push your changes to a branch on your own fork of run-ai, then open a pull request when the code is ready. The steps below walk through the full process.

### Step 1: Fork the Repository

Fork the repository to your own GitHub account. If you already cloned `dsx-ai-factory/run-ai` directly in [Getting Started](#getting-started), point that same clone at your fork instead of re-cloning:

```bash
git remote rename origin upstream
git remote add origin https://github.com/<your-username>/run-ai.git
```

### Step 2: Create a Branch

Branch from `main`, or from the live release branch you are targeting (see [Branching Policy](#branching-policy)). There's no enforced naming scheme, but a descriptive prefix helps. A few real examples from this repo:

```
docs/onboarding-fixes
fix/airgap-registry-secret-upgrade
```

### Step 3: Make Your Changes

- **One concern per PR.** Don't bundle unrelated fixes or refactors together; reviewers can approve a focused change much faster than a mixed one.
- **Add tests for new behavior.** Tests live alongside the code they cover, inside the service you're changing.
- **Update docs if user-facing.** Documentation source lives under [`docs/`](docs/); see [Documentation Contributions](#documentation-contributions) below.

### Step 4: Test Locally

Run the changed service's own Makefile targets before opening the PR: this mirrors what CI's `pr-fast.yaml` runs against your diff:

```bash
cd control-plane/assets-service   # whichever service(s) you touched
make build test lint validate
```

### Step 5: Sign Off Your Commits

Every commit needs a DCO sign-off (see [DCO](#developer-certificate-of-origin-dco)):

```bash
git commit -s -m "Fix node-pool selector when foo is bar"
```

### Step 6: Open a Pull Request

```bash
git push origin your-branch-name
```

- Open the PR against `main` (preferred) or the release branch (`release-3.0`).
- Fill in the PR template.
- Reference the issue with `Fixes #123`. The `PR links an approved issue` check requires exactly one linked issue that is open and labelled `approved` by a maintainer; for a PR into a release branch, link it from the issue's "Development" panel, as GitHub ignores closing keywords there.

### Step 7: Wait for CI

CI runs automatically once your PR is open: Layer 1 runs immediately, Layer 2 (the full build and end-to-end pipeline) starts once a maintainer types `/ok-to-test`. See [CI for Contributors](#ci-for-contributors--what-runs-and-when) below for what each layer covers.

### Step 8: Address Review Feedback

A maintainer approval plus a green `Validate Successful Run` status are both required for merge; see [Code Review](#code-review) for how reviewers get assigned.

---

## Documentation Contributions

- Documentation source lives under [`docs/`](docs/), published via the Fern navigation config in [`fern/`](fern/).
- Read [Documentation Agent Guide](docs/AGENTS.md) before editing, creating, or reviewing any page.
- Run `fern docs dev` before opening the PR, and `fern check` if you changed an OpenAPI spec. Run `tools/materialize-fern-versions.sh --fetch` first: older supported versions are built from their release branch (see [Versioned Docs](fern/VERSIONING.md)).

---

## Branching Policy

Run:ai accepts pull requests targeting `main` or the live release branch (today: `release-3.0`).

- **Main-first is preferred** - Open your PR against `main` so the fix flows to the supported release via the `backport/release-3.0` label that a maintainer applies after merge, rather than landing only on `main` and getting forgotten on the release branch.
- **Direct release-branch PRs are accepted** - When you can only reproduce or test on the version you're running, open the PR against `release-3.0`; during review, a maintainer will help you decide whether the fix should also be cherry-picked to `main`.

---

## CI for Contributors — What Runs and When

We split CI into two layers:

- **Layer 1: Fast, automatic, no secrets.**
  - Runs on every fork PR as soon as it is opened: DCO check, lint (Go, Helm, CODEOWNERS, workflows), CodeQL diff scan, govulncheck.
  - These checks post results back to your PR within a few minutes.
  - They are informational: they help you fix obvious issues before a maintainer looks at the PR, but they do not block merge by themselves.
- **Layer 2: Full build and end-to-end tests, has secrets.**
  - This is the pipeline that produces the required `Validate Successful Run` status check, and it only runs after a maintainer reviews your PR and types `/ok-to-test` in a comment.
  - End-to-end tests take 30–90 minutes.
  - A failed Layer 2 run blocks merge for everyone, including maintainers.
  - If a maintainer hasn't reviewed your PR within a few business days and you'd like the full pipeline to run, feel free to ping in the PR.
- **If Layer 2 fails, logs aren't directly visible.**
  - Look at the PR labels (`pipeline/build-failed`, `pipeline/e2e-failed`, `pipeline/jira-failed`) for a hint about which stage broke.
  - A maintainer will help you triage from there.

### Running the Checks Locally

Layer 1 is fully reproducible locally, and running it before you push is the fastest way to a green PR. There is deliberately **no single repo-wide gate command**, since this is a monorepo of independently buildable services. Layer 2 (full build and end-to-end tests) needs secrets and a cluster, so it cannot run locally at all.

**Start with the service you changed.** The closest thing to a local gate is each service's own Makefile:

```bash
cd <service-dir>            # e.g. runai-cluster/cluster/anyworkload-controller
make build test lint validate
```

That is exactly what CI runs for you: the `targets` job in [`pr-fast.yaml`](.github/workflows/pr-fast.yaml) discovers which services your PR touches and runs all four targets in each one, using that service's Makefile.

Then run whichever language-level checks apply. These mirror the corresponding jobs in [`pr-checks.yaml`](.github/workflows/pr-checks.yaml) and the standalone workflows, at the versions CI pins:

| You changed | Run | Mirrors |
|---|---|---|
| Any service | `cd <service-dir> && make build test lint validate` | `targets` job in `pr-fast.yaml` |
| Go code | `golangci-lint run` (v2.9.0)<br>`govulncheck ./...` | `go-lint`, `govulncheck` in `pr-checks.yaml` |
| Python | `ruff check <files>` (0.15.15) | `py-lint` in `pr-checks.yaml` |
| JS/TS | `npx eslint <files>` | `js-lint` in `pr-checks.yaml` |
| Helm charts | `helm dependency update control-plane/control-plane-chart`<br>`helm lint control-plane/control-plane-chart`<br>`helm template control-plane/control-plane-chart > /dev/null` | `helm` in `pr-checks.yaml` |
| GitHub Actions workflows | `actionlint` | `actionlint.yaml` |
| `docs/` or `fern/` | `tools/materialize-fern-versions.sh --fetch`<br>`fern check`<br>`python3 ./tools/check-fern-nav.py fern` | `fern-docs-ci.yaml` |
| `CODEOWNERS` | `make validate` (from the repo root) | — not run in CI; local check only |

The root `Makefile`'s `validate` target is real: it runs [`validate-codeowners.py`](.github/scripts/validate-codeowners.py) against `CODEOWNERS`. Its `build`, `test`, and `lint` targets are placeholders that only echo `NA`; run those per-service instead, as shown above.

`fern docs dev` starts a local preview server and does **not** validate anything. Use `fern check` to reproduce the CI check.

---

## Extending the Project

Most new capability lands through one of these extension points rather than by modifying core services. Follow the existing implementations in each directory; they are the reference for structure, naming, and tests.

- **New workload type** - NVIDIA Run:ai interprets workloads declaratively through Karta CRs. Adding support for a new Kubernetes workload kind means adding its interpretation rather than writing a bespoke controller. See [NVIDIA Run:ai CRD Reference](docs/developer/crds.md) and the `karta` chart dependency in [`runai-cluster/cluster/runai-chart`](runai-cluster/cluster/runai-chart).

- **New anyworkload-controller plugin** - Behaviour attached to a workload's lifecycle is implemented as a plugin under [`runai-cluster/cluster/anyworkload-controller/plugins/`](runai-cluster/cluster/anyworkload-controller/plugins). Use `networking`, `personalcreds`, or `restartrules` as templates: implement the hook interfaces your feature needs, add a `New(ctx, deps)` builder, and register it in `cmd/main.go`.

- **New cluster service or controller** - Add a directory under [`runai-cluster/cluster/`](runai-cluster/cluster) with an `artifact-metadata.json`, a Makefile including the shared `build/makefile/*.mk`, and a `README.md` plus `CLAUDE.md`. Wire its RBAC and deployment into [`runai-cluster/cluster/runai-chart`](runai-cluster/cluster/runai-chart).

- **New control-plane service** - Same shape under [`control-plane/`](control-plane), wired into [`control-plane/control-plane-chart`](control-plane/control-plane-chart). Generated API clients live in [`runai-clients/`](runai-clients).

- **New MCP tool** - Exposing an operation to AI assistants is done in [`runai-mcp-server/`](runai-mcp-server).

Start with an issue describing the extension point you intend to use, so maintainers can confirm the approach before you invest in the implementation. See [Architecture and Components](docs/developer/architecture.md) for how the pieces fit together.

---

## Code Review

- Branch protection requires **at least one approval** and a green `Validate Successful Run` check before merge.
- A `CODEOWNERS` file at the root automatically requests review from the right team for each affected path.
- For non-trivial changes, expect more than one reviewer to weigh in: this is a team norm, not a tool gate. Be patient and responsive.

---

## AI-Assisted Contributions

Using AI coding assistants (e.g. Claude Code, GitHub Copilot) to help write a contribution is fine: the same review bar applies regardless of how the change was produced.

- You are responsible for every line you submit. Read and understand the diff before opening a PR; don't submit AI output you haven't verified.
- The DCO sign-off (below) still certifies that you have the right to submit the change under this project's license, exactly as it would for hand-written code.
- Don't paste proprietary, confidential, or export-controlled material into third-party AI tools as part of preparing a contribution.
- PRs that are largely unreviewed AI output (inconsistent with the surrounding code, containing invented APIs, or with commit messages that don't reflect the actual change) will be asked to be revised or closed, same as any other low-quality submission.

---

## Developer Certificate of Origin (DCO)

External contributions to this repository require a Developer Certificate of Origin sign-off on every commit. Signing off certifies that the contribution is your own work, or that you have the right to submit it under the project's license. The full text of the DCO follows:

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.

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

---

## Commit Messages and PR Titles

Write clear, present-tense messages explaining **why** the change is needed, not just what the change does. A good PR title is a single sentence under ~70 characters.

We do not enforce Conventional Commits format. Release notes are generated automatically from PR titles and labels; clear titles make better release notes.

---

## Community Standards

Participation in this project is governed by our [Code of Conduct](CODE_OF_CONDUCT.md).

### Reporting a Code of Conduct Concern

Report unacceptable behaviour to **<GitHub_Conduct@nvidia.com>**. Reports are handled privately, and community leaders are obligated to respect the privacy and security of the reporter.

### What to Expect After You Report

- **Acknowledgement** - Within 3 business days of your report
- **Initial assessment** - Within 10 business days, including whether the report falls within the scope of this Code of Conduct
- **Status updates** - At least every 14 days while the report is open
- **Outcome** - Communicated to the reporter once the review concludes

Where a report is escalated to NVIDIA internal processes, resolution may take longer than the timelines above. You will be told if that happens.

Actions taken follow the four-tier Enforcement Guidelines in [Code of Conduct](CODE_OF_CONDUCT.md#enforcement-guidelines): correction, warning, temporary ban, permanent ban.

### What Is Out of Scope

The Code of Conduct governs conduct in this project's community spaces (see [Scope](CODE_OF_CONDUCT.md#scope)). The following are **not** handled through the Code of Conduct process; use the route listed instead:

| Out of scope | Use instead |
|---|---|
| Software defects, crashes, or unexpected behaviour | [GitHub Issues](https://github.com/dsx-ai-factory/run-ai/issues) |
| Security vulnerabilities | [Security](SECURITY.md) - do not open a public issue |
| Disagreement with a technical or design decision, or with a maintainer's review | Discuss on the issue or pull request; see [Code review](#code-review) |
| A maintainer declining, deprioritising, or closing your contribution | Not misconduct on its own - see [How Issue Priority Is Decided](#how-issue-priority-is-decided) |
| Commercial, licensing, or support-contract disputes | Contact NVIDIA via [Support](SUPPORT.md) |
| Conduct in venues unrelated to this project, where no one is representing the project | Outside this project's remit |

Reports made in bad faith, for example to retaliate against, harass, or silence another participant, are themselves a Code of Conduct violation.

---

## Getting Help

- [GitHub Discussions](https://github.com/dsx-ai-factory/run-ai/discussions) - Usage questions and design conversations
- [GitHub Issues](https://github.com/dsx-ai-factory/run-ai/issues) - Bugs, feature requests, documentation issues
- [Security](SECURITY.md) - Private vulnerability reports
- [Support](SUPPORT.md) - Which versions are supported

Thank you for contributing.
