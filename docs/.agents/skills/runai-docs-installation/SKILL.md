---
name: runai-docs-installation
description: Route an installation-doc change to the correct file(s) — the shared system-requirements tier, NVIDIA Run:ai's two install paths (Guided Installation and Manual Helm), and the connected/air-gapped variants within Guided Installation — keeping mirrored and cross-path sections in sync. Use when a change affects installation, system requirements, prerequisites, or the installer configuration reference under docs/installation-and-configuration/.
---

# NVIDIA Run:ai Installation Documentation

Use when a change affects the installation flow itself, system requirements,
prerequisites or features, or the installer's configuration reference. Not
for upgrades or post-install configuration — see
`docs/installation-and-configuration/upgrade/` and
`docs/installation-and-configuration/configuration/` for those.

## The Two Install Paths

| Path | Location | When it's used |
|---|---|---|
| Guided Installation | `installation/guided-installation/` | Recommended — the `runai-installer` CLI, driven by a `config.yaml` |
| Manual Helm | `installation/install-using-helm/` | GitOps/Argo CD environments, or admins who manage Helm releases directly |

Guided Installation further splits into a 2×2 matrix: `connected/` vs
`air-gapped/`, each containing the same four files — `preparations.md`,
`same-cluster.md`, `multi-cluster.md`, `templates.md` — with the same H2
section skeleton in both variants (content differs; structure doesn't).
Manual Helm has no separate connected/air-gapped files — the same
distinction is handled as `<Accordion title="Connected">` /
`<Accordion title="Air-gapped">` blocks nested inside its Kubernetes/OpenShift
subsections, in the same file.

Both paths share the same `runai-installer` binary for air-gapped bundle
building (`airgap pack` / `airgap load`) — Manual Helm's own docs instruct
users to download and run it before their Helm commands. Do not assume a
concept is Guided-only just because it's the installer's binary; check
whether Manual Helm's docs already reference it (see Verifying Against
Source below).

`installation/bcm-install.md` (Base Command Manager) is a third, legacy path
that only redirects to older-version instructions — out of scope for this
skill.

**Being linked to from other docs.** Guided Installation is the default link
target from anywhere outside `installation/` — a feature doc mentioning a
prerequisite (e.g. Knative) links to
`guided-installation/prerequisites-and-features.md`, not the Helm version,
unless the content is genuinely Helm-specific. See Links in
`docs/WRITING.md`.

## The System Requirements Tier

Not every product-level fact belongs inside `installation/` at all. A third
location sits above both paths:
`docs/installation-and-configuration/system-requirements/` (`README.md`,
`deployment-model.md`, `infrastructure-requirements.md`,
`networking-setup.md`, `external-access-to-critical-services.md`,
`third-party-services.md`) — the single, tool-agnostic baseline both Guided
Installation and Manual Helm require. Nearly every file under
`installation/` links back into it rather than restating it.

Check this before applying Verifying Against Source below — it decides
whether a fact belongs in either path's docs at all:

- **No per-path mechanism** — a fact about the environment or infrastructure
  that's true before a tool is even chosen (supported Kubernetes versions,
  hardware/GPU requirements, storage, the same-cluster vs multi-cluster
  deployment-model decision, which FQDNs and certificate types are needed in
  the abstract) → belongs in `system-requirements/` only. Neither path's own
  docs restate it — they link to it, typically from a "Before You Start"
  section. Do not duplicate it into `prerequisites-and-features.md` or
  `preparations.md` on either side.
- **Has a per-path mechanism** — the requirement lives in
  `system-requirements/`, but *satisfying* it differs by tool → the
  requirement itself stays a single edit there; the per-path *how* is
  handled separately per Verifying Against Source and Mapping Guided
  Installation to Manual Helm below. Example: which FQDNs and certificate
  types are needed lives in `system-requirements/networking-setup.md`; how
  each path actually provisions those certificates is per-path (Guided's
  `tls.*` config.yaml field, Manual Helm's manual secret creation) — those
  are two separate edits, neither of which touches
  `networking-setup.md` itself.

If it's unclear which tier a change belongs in, grep `system-requirements/`
for the topic first. If it's already covered there, that page is the
requirement's home — any further doc change stays scoped to *how* each
install path executes it, not to restating the requirement.

Several `system-requirements/` pages already mention connected vs.
air-gapped inline, as prose within one page — this is not the same as the
per-path connectivity split under `installation/`. There are no separate
`system-requirements/connected/` and `system-requirements/air-gapped/`
files to mirror; a connectivity-dependent fact here just needs to stay
accurate for both cases within the same paragraph or list.

## On Activation

Before drafting anything:

1. Ask what's changing — the install flow itself, a system requirement, a
   prerequisite/feature, or a config.yaml field/installer command. Do not
   ask the human which tier or path(s) it applies to — determine that in
   order: The System Requirements Tier first, then Verifying Against Source
   for path applicability. Only ask if both leave it genuinely ambiguous
   (e.g. a brand-new capability with no existing analog).
2. Read `docs/AGENTS.md` for writing rules, Fern components, and terminology.
3. Read the reference page(s) matched by the routing decision below, in
   full, before editing.

## Verifying Against Source

Path applicability is a fact to verify, not a guess from folder structure or
from where you first noticed the change. Apply this test to any change,
including ones with no precedent in the skill:

**1. Is it a fact about what the product requires or does, or a fact about
one tool's own mechanism?** Both paths deploy the same control plane and
cluster Helm charts with the same underlying requirements — TLS, registry
auth, prerequisites, features, networking. They differ only in tooling:
Guided Installation wraps chart deployment in `config.yaml` +
`runai-installer` automation; Manual Helm runs `helm install`/`upgrade`
directly, by hand.

- If the change is a fact about the product (a new TLS requirement, a new
  prerequisite, a new supported registry auth method, a new feature) — it is
  real in **both paths by default**. Verify the requirement itself against
  product source (the relevant Helm chart's `values.yaml`/templates, or the
  controller/service that enforces it) — not against either doc, and not
  against installer source, which only tells you how Guided automates it.
- If the change has no equivalent in the other tool at all — an installer
  preflight check that exists only because the installer automates a step a
  human does manually elsewhere, or a GitOps/Argo CD consideration specific
  to Manual Helm — it's genuinely single-path. Treat this as the exception:
  confirm it by checking whether the other path's docs already cover the
  same requirement in different words before concluding it's single-path.

**2. For each path the change is real in, verify that path's own mechanism
independently** — never transfer one path's syntax into the other's docs:
- Guided's realization (a config.yaml field, a CLI flag, an installer
  validation step) — verify against `runai-installer/installer/internal/`
  (`config`, `cli`, `inputcheck`, `step`, `prereq`, etc.).
- Manual Helm's realization (a `--set` value, a manual Kubernetes step, a
  cert/secret created by hand) — verify against the relevant Helm chart's
  `values.yaml`/templates, or the existing manual instructions.

**3. Before editing, grep both doc trees for the exact term** — command
name, flag, field, or product name — under
`docs/installation-and-configuration/installation/guided-installation/` and
`.../install-using-helm/`. Do this even for a change you've already
classified as single-path: the other path may still *reference* it (a
cross-link, a shared prerequisite mention) and need a smaller update even
without a new section.

### Worked Examples

- **`airgap pack`/`airgap load` fix**: a fact about tooling, but the tooling
  (the `runai-installer` binary) is shared — Manual Helm's own docs instruct
  users to run it. Grepping `install-using-helm/` for `airgap` confirms this
  (finds it in `preparations.md`). Both paths need the fix.
- **New prerequisite/operator**: a fact about the product (something the
  cluster needs), so real in both paths by rule 1. Whether Guided documents
  it as installer-automatable depends on rule 2 — check
  `runai-installer/installer/internal/prereq/registry.go` (and its `ids`/
  `components` packages); if it's registered, use `mode: install`/
  `mode: skip` language matching GPU Operator's existing entry, otherwise
  document it as always `mode: skip` with manual installation, the same way
  unregistered prerequisites already are. Manual Helm documents it as manual
  installation either way, independent of what the installer registry says.
- **New TLS requirement**: a fact about the product (both paths deploy
  ingress/TLS on the same charts), so real in both by rule 1 — verify the
  requirement itself against the chart or the service enforcing it. Then
  verify each path's syntax separately by rule 2: Guided's `tls.*`
  config.yaml fields against `runai-installer/installer/internal/step/` (see
  `tls_secret_step.go`) and `internal/inputcheck/`; Manual Helm's certificate
  handling against its existing `Control Plane Certificate`/`Cluster TLS
  Certificates` sections in `install-using-helm/preparations.md`, which
  already documents the manual, by-hand equivalent. Finding those exact
  locations in each path is not structural — see Mapping Guided to Manual
  Helm below.

## Mapping Guided Installation to Manual Helm

Once Verifying Against Source says a change is real in both paths, this
answers *where* in each, and what should read the same vs. different once
you're there. Unlike `connected/` vs `air-gapped/` (identical files,
identical section skeletons — see Mirroring below), Guided Installation and
Manual Helm are **not** structural mirrors. Guided organizes files by
installation concern (`preparations.md`, `same-cluster.md`,
`multi-cluster.md`, `templates.md`); Manual Helm organizes by installed
component (`preparations.md`, `install-control-plane.md`, `helm-install.md`).
There is no heading-to-heading match — find the corresponding content by
topic, not by structure:

1. **Locate Guided's presentation.** Check the field/flag definition in
   `configuration-reference.md` or `command-reference.md` first — that's the
   canonical spec. Then grep the topic keyword across the
   `connected/`/`air-gapped/` `preparations.md`, `same-cluster.md`, and
   `multi-cluster.md` files — a requirement can be walked through in prose in
   more than one of them.
2. **Locate Manual Helm's presentation** the same way: grep the topic
   keyword across every file under `install-using-helm/`. There's no single
   canonical file to check — Manual Helm weaves a given requirement into
   whichever step it's relevant to (TLS, for example, shows up in
   `preparations.md`'s certificate sections *and* as inline steps in
   `install-control-plane.md`/`helm-install.md`).
3. **What must stay identical at every location found**: the underlying
   requirement or fact itself — the actual values, constraints, minimum
   versions, supported formats established by the source change. If the
   requirement changed, every location documenting it needs the same updated
   fact, in both paths.
4. **What must differ, by design, at every location**: the mechanism and the
   voice describing it. Guided's copy describes what the installer does with
   a config.yaml field ("the installer accepts...", "set
   `tls.generateSelfSigned` to..."); Manual Helm's copy describes what the
   administrator does by hand ("create a Kubernetes secret containing...",
   "generate a certificate using..."). Never copy Guided's field-and-value
   phrasing into Manual Helm's manual-step phrasing, or vice versa — match
   each path's own existing prose pattern for a similar requirement, not the
   other path's prose.

This mapping happens in addition to, not instead of, the intra-path
mirroring below — a cross-path change landing in a connectivity-dependent
section still needs the connected/air-gapped (or accordion) mirror checked
within each path separately.

## Routing the Change

Work through these in order — stop at the first one that matches:

1. **A config.yaml field, installer flag, or installer command** — the field
   definition itself has one source of truth:
   - `guided-installation/configuration-reference.md` — config.yaml fields
   - `guided-installation/command-reference.md` — installer CLI flags,
     troubleshooting
   But the grep-both-trees check from Verifying Against Source still applies
   to *usage* of that field or command — `airgap pack`/`airgap load` are
   config.yaml/CLI concepts defined here, yet also walked through as usage
   steps in `guided-installation/air-gapped/preparations.md` and
   `install-using-helm/preparations.md`. Fixing the definition doesn't fix
   the usage steps; check both.
2. **A prerequisite or feature** (what it does, when it's needed, Helm
   values) — a single page per path, not split further by connectivity:
   - `guided-installation/prerequisites-and-features.md` (Guided)
   - `install-using-helm/prerequisites-and-features.md` (Manual Helm)
   These are separate pages with separate content — route which one(s) get
   the change using the registry check in Verifying Against Source, not a
   guess.
3. **A step in the install flow, or anything that differs by connectivity**
   (registry auth, bundle building, cluster-count logic) — belongs in the
   per-variant pages (`preparations.md`, `same-cluster.md`,
   `multi-cluster.md`, `templates.md`) under `connected/` or `air-gapped/`
   for Guided, or the matching `<Accordion>` block for Manual Helm. This is
   where the mirroring rule below applies — for Guided across
   `connected/`/`air-gapped/`, and for Manual Helm across its
   Connected/Air-gapped accordions in the same file.
4. **A step specific to Manual Helm's flow** —
   `install-using-helm/preparations.md`, `install-control-plane.md`, or
   `helm-install.md`, matched by which part of the flow it's in.

If a change is real in both Guided Installation and Manual Helm, treat each
path independently — never point one path's doc at the other's page; the
flows, terminology, and artifacts differ.

## Mirroring Connected and Air-Gapped

Guided Installation mirrors `connected/` against `air-gapped/` file-for-file
and section-for-section. Manual Helm mirrors the same distinction as
`<Accordion title="Connected">` against `<Accordion title="Air-gapped">`
blocks inside the same file. Before finishing an edit to any connectivity-
dependent content:

1. Guided: open the same file under the other connectivity directory. Manual
   Helm: find the sibling accordion in the same file.
2. Find the matching section (same heading text, or the closest equivalent —
   e.g. air-gapped's `## Air-Gapped Bundle` has no connected equivalent).
3. If the change is connectivity-agnostic, apply the same edit there.
4. If the change is connectivity-specific (e.g. registry credential format
   differs by design), do not copy it verbatim — write the correct
   equivalent for that variant instead.

Do not finish the task until this check has been done for every mirrored
section touched — in every path the change reached — even when the answer is
"no matching section, nothing to mirror."

## Nav Registration

Register new pages in `fern/products/self-hosted.yml`. Match the nesting of
the location edited: `system-requirements/`, Guided Installation's
`connected/` or `air-gapped/`, or Manual Helm.

## Quality Checks

Before finishing:

- [ ] Checked The System Requirements Tier first — a pure environment/
  infrastructure fact with no per-path mechanism was not duplicated into
  either path's own docs, only linked from them
- [ ] Path applicability (Guided, Manual Helm, or both) determined by the
  Verifying Against Source check — grepping both trees for an existing
  fix, or checking `runai-installer/installer/internal/prereq/registry.go`
  for a new prerequisite/feature — not guessed or asked of the human
- [ ] If real in both paths: every location found via Mapping Guided
  Installation to Manual Helm was updated with the same underlying fact, each
  written in that path's own mechanism and voice — not copy-pasted across
- [ ] If connectivity-specific: the matching connected/air-gapped section was
  checked and mirrored (per-file for Guided, per-accordion for Manual Helm),
  or confirmed not applicable
- [ ] Shared reference pages (`configuration-reference.md`,
  `command-reference.md`) edited once, not duplicated into the per-variant
  pages
- [ ] A new prerequisite/feature: installer-automation status verified
  against the `prereq` registry, not assumed from either doc's existing
  wording
- [ ] `<Warning>` and `<Note>` used per `docs/WRITING.md` — never plain text
- [ ] All field names, defaults, and flags verified against source — nothing
  unconfirmed included
- [ ] `fern docs dev` runs without errors
