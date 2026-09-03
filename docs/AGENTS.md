# Documentation Agent Guide

Read this file before editing, creating, or reviewing any page under `docs/`.
For repository-wide context — products, sources of truth, and the
documentation-impact trigger — see the root `AGENTS.md`.

## How to Approach Documentation Work

Before drafting any page:

1. **Match existing pages.** Find the most similar existing page in `docs/` and
   read it in full. Match its structure, sections, Fern components, heading
   levels, and link style exactly. The existing pages are the living standard
   for structure and layout.

   When adding to an existing page — a new table row, a new config entry, a
   new step in a procedure — read at least 3 existing entries of the same type
   on that page before adding. Match the description length, column structure,
   value formatting, and level of detail exactly. Do not introduce a new
   pattern.

   Before adding anything to an existing page, read the surrounding content to
   understand every pattern in use — punctuation, sentence structure, how
   conditions and qualifiers are expressed. Replicate those patterns exactly.
   Do not introduce a new pattern for anything already handled consistently on
   the page.
2. **Ground every claim in something concrete.** Verify field names, default
   values, behavior, CLI flags, and API endpoints against the source code or
   an API spec — see Claim Accuracy in `docs/WRITING.md` for the full rules.
3. **Use the page type that matches the content.** Refer to the Page Types
   section below to select the right type, then read an existing page of that
   type as your reference pattern.
4. **Determine product scope.** Confirm which products (Self-hosted, SaaS,
   Multi-tenant) the page applies to. Register or update the page in every
   applicable nav file under `fern/products/`.

## What to Do by Change Type

Use this table to determine the documentation action before starting work.

| Change type | Action | Reference |
|---|---|---|
| New UI resource (object with add/edit/delete/metrics) | Load `docs/.agents/skills/runai-docs-ui-how-to/SKILL.md` | — |
| New quick start (single feature or workload type walkthrough) | Load `docs/.agents/skills/runai-docs-quickstart/SKILL.md` | `docs/workloads-in-nvidia-run-ai/using-training/quick-starts/standard-training-quickstart.md` |
| New feature guide (scheduler feature, GPU optimization, topology, NUMA) | Load `docs/.agents/skills/runai-docs-feature-guide/SKILL.md` | `docs/platform-management/aiinitiatives/resources/topology-aware-scheduling.md` |
| New CLI flag or command | Do not edit — CLI reference is auto-generated. Add `<!-- TODO: CLI ref update needed -->` to the related page and flag for engineering. | — |
| New or changed API endpoint or field | Do not edit — REST API reference is auto-generated. Flag for engineering. | — |
| New or changed MCP server tool | Load `docs/.agents/skills/update-mcp-server-docs/SKILL.md` | `docs/installation-and-configuration/mcp-server.md` |
| New feature, enhancement, or behavior change with no dedicated skill above | See How to Approach Documentation Work above | — |

When no exact match exists, find the most similar existing page in `docs/` and follow its structure.

## Mapping Code Changes to Doc Pages

Once documentation work has been explicitly requested for a code change,
follow these steps to determine which doc pages need updating.

1. **Read the diff summary.** Identify which source files changed and what the
   user-visible impact is. Focus on public API fields, CLI flags, UI behavior,
   Helm values, scheduler behavior, and default values. Ignore internal
   implementation details with no user-visible effect.
2. **Classify the change type.** Use the table above to determine the
   documentation action. If the change affects multiple areas, apply each row
   independently.
3. **Search for existing coverage.** Before creating a new page, search `docs/`
   for the feature, field, or behavior name. Update the existing page rather
   than creating a duplicate.
4. **Keep changes together.** If a single code change touches multiple doc
   areas, include all affected pages in one PR rather than splitting across
   multiple PRs.

## Writing Style Guide

Writing rules — voice, terminology, formatting, claim accuracy — live in
`docs/WRITING.md`. Read it before drafting or editing prose.

## Terminology

Terms, correct spelling, and casing live in `docs/terminology.md`. Read it
before drafting or editing prose — do not guess at product or feature names.

## Page Types

Each page serves one purpose. Match the type to the content:

| Type | Use when |
|---|---|
| Conceptual | Background and behavior, no steps |
| How-to | Step-by-step procedure for a single task |
| Feature guide | Background is inseparable from configuration |
| Infrastructure guide | Infrastructure setup (Helm, Kubernetes manifests) |
| Tutorial | Multi-task end-to-end workflow |
| Feature quick start | Single feature, end-to-end walkthrough |
| Persona quick start | Role-based platform onboarding |
| Reference | Exhaustive field, flag, or parameter listing |
| Overview | High-level entry point for a section |

Decision rules:
- Use a feature guide when background is inseparable from configuration. Use a
  how-to when the concept lives on a separate page and the page is purely
  procedural.
- Use an infrastructure guide for Kubernetes-level setup (Gateway API, Karpenter,
  node roles). Use a feature guide for Run:ai UI or platform configuration.
- A feature quick start demonstrates one feature end-to-end. A tutorial
  connects multiple features into a complete workflow.
- A persona quick start onboards a role to the platform. A feature quick start
  teaches one capability to someone already onboarded.

## Repository Structure

Place new files under `docs/` in the subdirectory that matches the topic area.
Use kebab-case filenames ending in `.md` — for example,
`configure-node-pools.md`.

Standard content pages have no frontmatter. Start the file with the `# H1`
title. Add frontmatter only when a Fern-specific key is needed:

| Key | When to use |
|---|---|
| `hidden: true` | Drafts or archived content not reachable from nav |
| `noIndex: true` | Combined with `hidden: true` to suppress search indexing |
| `description` | Section-index pages where a subtitle improves SEO |

## Auto-Generated Content

Do not edit these sections directly — changes must be made at the source
location elsewhere in this repo:

| Section | Source location |
|---|---|
| CLI reference (`docs/reference/cli/`) | `control-plane/runai-cli/docs` |
| REST API reference | Swagger source files at `control-plane/<service>/swagger/source` |

## Fern Components

Which MDX component to use for callouts, tabs, and other formatting lives in
`docs/WRITING.md`. Read it before adding a callout, warning, or similar.

## Navigation

Register new pages in the correct product nav file. A page with no nav entry
is unreachable:

| Content applies to | Nav file |
|---|---|
| Self-hosted + SaaS | `fern/products/self-hosted.yml` + `fern/products/saas.yml` |
| SaaS-specific | `fern/products/saas.yml` |
| Multi-tenant | `fern/products/multi-tenant.yml` |

## Links

Link formatting — internal vs. external, UI navigation paths, link text —
lives in `docs/WRITING.md`.

## Verification

Run these commands before opening a pull request:

| Command | When to run |
|---|---|
| `fern docs dev` | After any change to `docs/` or `fern/` — validates the full site locally |
| `fern check` | After adding or changing OpenAPI specs in `fern/apis/` |

If `fern docs dev` fails, resolve all errors before opening the PR. Do not open
a PR with a broken local build.
