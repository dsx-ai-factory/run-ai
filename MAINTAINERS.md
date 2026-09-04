# Maintainers

This document lists the maintainers of the NVIDIA Run:ai project and how the
role is granted and relinquished.

## Current maintainers

Maintainers are tracked as a GitHub team rather than a name list in this file,
so membership changes do not require a pull request:

- https://github.com/orgs/run-ai/teams/maintainers

Path-level review ownership is separate and is defined in
[`CODEOWNERS`](CODEOWNERS), which maps each area of the repository to the teams
that must approve changes to it. A maintainer is not automatically a code owner
for every path, and a code owner is not automatically a maintainer.

## Responsibilities

Maintainers are responsible for:

- Reviewing and merging pull requests, per the review rules in
  [`CONTRIBUTING.md`](CONTRIBUTING.md#code-review)
- Triaging incoming issues and applying labels
- Running the full CI pipeline on external contributions (the `/test` step
  described in [`CONTRIBUTING.md`](CONTRIBUTING.md#ci-for-external-contributors--what-runs-and-when))
- Applying `backport/v2.X` labels so fixes reach supported releases
- Making release decisions for the versions listed in [`SUPPORT.md`](SUPPORT.md)
- Responding to reports received through [`SECURITY.md`](SECURITY.md)
- Upholding the [Code of Conduct](CODE_OF_CONDUCT.md)

## Becoming a maintainer

Maintainers are selected on demonstrated contribution rather than affiliation or
tenure. The criteria are:

1. **Sustained contributions** — quality changes over a period of time, not a
   single large drop
2. **Review quality** — thoughtful, constructive reviews of others' work
3. **Community engagement** — helping contributors, triaging issues, answering
   questions
4. **Domain knowledge** — working understanding of the area they will own

Process: an existing maintainer nominates the candidate, and the nomination
carries on the approval of a simple majority of current maintainers. See
[`GOVERNANCE.md`](GOVERNANCE.md#adding-and-removing-maintainers)
for the full procedure, including how a maintainer steps down or is removed.

## Emeritus maintainers

Maintainers who are no longer active. In the usual case — stepping down, or
inactivity — emeritus status is not a judgement, and it is reversed by the
normal nomination process. Where emeritus status followed a Code of Conduct
removal, the additional precondition in
[`GOVERNANCE.md`](GOVERNANCE.md#adding-and-removing-maintainers) applies.

_None yet._

## Governance

Decision-making, tie-breaking, and the amendment process are documented in
[`GOVERNANCE.md`](GOVERNANCE.md).
