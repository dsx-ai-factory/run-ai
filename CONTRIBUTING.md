# Contributing to Run:ai

Thanks for your interest in contributing. This guide covers how to file issues, propose changes, and submit pull requests.

By participating, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md). To report a security vulnerability, follow the private process in [`SECURITY.md`](SECURITY.md) — **do not open a public issue**.

---

## Ways to Contribute

- **You want to report a bug, feature idea, or documentation issue** - File a [GitHub issue](https://github.com/dsx-ai-factory/run-ai/issues/new/choose) using the appropriate template. Maintainers will triage and label it.
- **You want to propose a new feature and implement it** - File a feature-request issue describing your intended feature, and discuss the design there. Once the plan looks good, implement it following the Pull Request Workflow below.
- **You want to implement a feature or bug fix for an outstanding issue** - Pick up an issue labelled `good first issue` or `help wanted`, comment to claim it, then follow the Pull Request Workflow below.

---

## Getting Started

1. **Clone the repository.**

   ```bash
   git clone https://github.com/dsx-ai-factory/run-ai.git
   cd runai
   ```

2. **Build.** Uses Go modules — no private dependencies, so it's fork-safe.

   ```bash
   make build
   ```

3. **Run unit tests.**

   ```bash
   make test
   ```

Setup and end-to-end testing on a real cluster are service-specific. Each service directory under `control-plane/` and `runai-cluster/` has its own `README.md` and `Makefile` describing how to build, test, and validate it.

---

## Pull Request Workflow

1. **Fork** the repository to your own GitHub account.
2. **Branch** from `main` (or from the live release branch you are targeting — see [Branching Policy](#branching-policy) below). Use a descriptive name.
3. **Make focused changes.** One concern per PR. Add tests for new behavior. Update docs if user-facing.
4. **Sign off every commit** with `git commit -s` (see [DCO](#developer-certificate-of-origin-dco) below). This applies to **fork PRs only**; internal Run:ai contributors are not DCO-checked.
5. **Open the PR** against `main` (preferred) or the release branch (`release-3.0`). Fill in the PR template, reference the issue with `Fixes #123`, and add a `/kind` line.
6. **Wait for CI.** Layer 1 checks (lint, DCO for forks, CodeQL diff, govulncheck) run automatically. The full build pipeline and end-to-end tests run after a maintainer reviews and types `/ok-to-test` on the PR.
7. **Address review feedback.** A maintainer approval plus a green `Validate Successful Run` status are both required for merge.

---

## Branching Policy

Run:ai accepts pull requests targeting `main` or the live release branch (today: `release-3.0`).

- **Main-first is preferred** - Open your PR against `main` so the fix flows to the supported release via the `backport/release-3.0` label that a maintainer applies after merge, rather than landing only on `main` and getting forgotten on the release branch.
- **Direct release-branch PRs are accepted** - When you can only reproduce or test on the version you're running, open the PR against `release-3.0`; during review, a maintainer will help you decide whether the fix should also be cherry-picked to `main`.
- **How backports work** - When a maintainer applies the `backport/release-3.0` label to your merged PR, an automated workflow opens a follow-up PR cherry-picking the merge commit onto the release branch. If the cherry-pick conflicts, the bot leaves a `needs-manual-backport/release-3.0` label and a maintainer takes it from there. You don't need to open backport PRs yourself.

---

## CI for External Contributors — What Runs and When

We split CI into two layers:

- **Layer 1 — fast, automatic, no secrets.**
  - Runs on every fork PR as soon as it is opened: DCO check, lint (Go, Helm, CODEOWNERS, workflows), CodeQL diff scan, govulncheck.
  - These checks post results back to your PR within a few minutes.
  - They are informational — they help you fix obvious issues before a maintainer looks at the PR, but they do not block merge by themselves.
- **Layer 2 — full build and end-to-end tests, has secrets.**
  - This is the pipeline that produces the required `Validate Successful Run` status check, and it only runs after a maintainer reviews your PR and types `/ok-to-test` in a comment.
  - End-to-end tests take 30–90 minutes.
  - A failed Layer 2 run blocks merge — for everyone, including internal contributors.
  - If a maintainer hasn't reviewed your PR within a few business days and you'd like the full pipeline to run, feel free to ping in the PR.
- **If Layer 2 fails — logs aren't directly visible.**
  - The workflow logs from `runai-actions` (the private repository where the full pipeline lives) are not directly visible to external contributors.
  - Look at the PR labels (`pipeline/build-failed`, `pipeline/e2e-failed`, `pipeline/jira-failed`) for a hint about which stage broke.
  - A maintainer will help you triage from there.

---

## Code Review

- Branch protection requires **at least one approval** and a green `Validate Successful Run` check before merge.
- `CODEOWNERS` files at the root and under `.github/` and `docs/` automatically request review from the right team for each affected path.
- For non-trivial changes, expect more than one reviewer to weigh in — this is a team norm, not a tool gate. Be patient and responsive.

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

**Internal contributors** - The DCO check does not run on commits authored from `@run.ai` and `@nvidia.com` accounts. Internal contributions are governed by the existing employee IP agreement.

---

## Commit Messages and PR Titles

Write clear, present-tense messages explaining **why** the change is needed, not just what the change does. A good PR title is a single sentence under ~70 characters.

We do not enforce Conventional Commits format. Release notes are generated automatically from PR titles and labels — clear titles make better release notes.

---

## Getting Help

- [GitHub Issues](https://github.com/dsx-ai-factory/run-ai/issues) - Bugs, feature requests, documentation issues, usage questions
- [SECURITY.md](SECURITY.md) - Private vulnerability reports
- [SUPPORT.md](SUPPORT.md) - Which versions are supported

Thank you for contributing.
