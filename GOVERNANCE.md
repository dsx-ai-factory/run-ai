# Governance

This document describes how decisions are made in the NVIDIA Run:ai project.
It covers roles, decision-making, and how the maintainer set changes.

It deliberately does not restate the contribution mechanics — how to open a pull
request, what CI runs, how review and backports work — which live in
[`CONTRIBUTING.md`](CONTRIBUTING.md).

## Project Status

This is an NVIDIA open-source project. The source is public and contributions
are welcome under the terms in [`CONTRIBUTING.md`](CONTRIBUTING.md), including
the Developer Certificate of Origin sign-off requirement. Roadmap and release
scheduling are set by NVIDIA.

## Roles

- **Contributors** - Anyone who opens an issue or pull request. No prior
  association with the project is required.
- **Code owners** - The teams listed in [`CODEOWNERS`](CODEOWNERS). Approval
  from the relevant code owners is required to merge a change to the paths
  they own. This is enforced by GitHub, not by convention.
- **Maintainers** - Listed in [`MAINTAINERS.md`](MAINTAINERS.md). Maintainers
  can merge pull requests, run the full CI pipeline on external
  contributions, cut releases, and change this document.

## Decision-Making

Most decisions are made through the ordinary review process: a change is
proposed as a pull request and merged once it has the approvals required by
[`CODEOWNERS`](CODEOWNERS) and a green pipeline.

For decisions that are not a single pull request — architectural direction,
adding or removing a supported release, changing the contribution or governance
rules — the project uses **lazy consensus**. A maintainer proposes the change in
a GitHub issue or discussion, and if no maintainer objects within five working
days, the proposal carries. Any maintainer may call for an explicit vote
instead, in which case a simple majority of maintainers decides.

Objections should be technical and should come with an alternative. An objection
pauses lazy consensus but does not veto the proposal: if it is not resolved
within a further ten working days, any maintainer may call the vote described
above, and the simple majority decides. A proposal is never blocked
indefinitely by a single objection.

### Tie-Breaking

If maintainers cannot reach agreement, the decision escalates to the NVIDIA
engineering leadership responsible for the project. This is a fallback, not a
routine step.

## Adding and Removing Maintainers

**Adding.** Any maintainer may nominate a contributor who meets the criteria in
[`MAINTAINERS.md`](MAINTAINERS.md#becoming-a-maintainer). The nomination is
raised in a GitHub issue and carries on the approval of a simple majority of
maintainers. An objection must be stated with reasons and is resolved by that
vote rather than by veto, following the same model as
[Decision-Making](#decision-making). On approval, the new maintainer is added to
the maintainers GitHub team and, where appropriate, to the relevant entries in
[`CODEOWNERS`](CODEOWNERS).

**Stepping down.** A maintainer may step down at any time by opening a pull
request moving themselves to the Emeritus section of
[`MAINTAINERS.md`](MAINTAINERS.md#emeritus-maintainers). No approval is needed.

**Removal.** A maintainer who has been inactive for six months, or who
repeatedly acts against the [Code of Conduct](CODE_OF_CONDUCT.md), may be moved
to Emeritus by a majority vote of the other maintainers. The affected maintainer
is notified before the vote and may respond.

The two grounds for removal differ in how they can be reversed. Removal for
**inactivity** carries no stigma and is reversed by the normal nomination
process above. Removal for **Code of Conduct** reasons is not reversed the
same way: renomination requires that the conduct matter has been reviewed and
closed by the enforcement contact named in
[`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md), and that review is a precondition
for a nomination being raised at all.

## Security

Vulnerability reports follow the private disclosure process in
[`SECURITY.md`](SECURITY.md) and are handled by maintainers together with NVIDIA
Product Security. Security fixes may be developed in private and merged without
the usual public review window; they are disclosed once a fix is available for
the supported releases listed in [`SUPPORT.md`](SUPPORT.md).

## Code of Conduct

All participants are bound by the [Code of Conduct](CODE_OF_CONDUCT.md).
Enforcement is handled by the maintainers through the reporting channel named in
that document.

## Amending This Document

Changes to this document follow the same lazy-consensus process described in
[Decision-Making](#decision-making), proposed as a pull request.
