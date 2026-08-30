# Documentation Agent Guide

Read this file before editing, creating, or reviewing any page under `docs/`.
For repository-wide context — products, sources of truth, branches, and the
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
2. **Ground content in the code.** Every field name, default value, behavior,
   CLI flag, and API endpoint must be verified against the source code or API
   spec. Do not invent or assume values.
3. **Document only what the changed source files in this PR directly
   implement.** Do not infer or extrapolate behavior from indirect signals. If
   you cannot point to a specific changed file that supports a claim, do not
   make it. If a feature or field is not present in the changed files, do not
   document it.
4. **Use the page type that matches the content.** Refer to the Page Types
   section below to select the right type, then read an existing page of that
   type as your reference pattern.

## What to Do by Change Type

Use this table to determine the documentation action before starting work.

| Change type | Action | Reference |
|---|---|---|
| New UI resource (object with add/edit/delete/metrics) | Load `.agents/skills/runai-docs-ui-how-to/SKILL.md` | — |
| New quick start (single feature or workload type walkthrough) | Load `.agents/skills/runai-docs-quickstart/SKILL.md` | `docs/workloads-in-nvidia-run-ai/using-training/quick-starts/standard-training-quickstart.md` |
| New feature guide (scheduler feature, GPU optimization, topology, NUMA) | Load `.agents/skills/runai-docs-feature-guide/SKILL.md` | `docs/platform-management/aiinitiatives/resources/topology-aware-scheduling.md` |
| New field or option on an existing UI resource | Find and update the existing page for that resource | Existing page for that resource |
| New or changed scheduler behavior | Update relevant page in `docs/platform-management/runai-scheduler/` | Similar scheduler page |
| New or changed policy or scheduling rule | Update relevant page in `docs/platform-management/policies/` | Similar policy page |
| New installation step, Helm value, or infrastructure change | Update relevant page in `docs/installation-and-configuration/` | Similar installation page |
| New CLI flag or command | Do not edit — CLI reference is auto-generated. Add `<!-- TODO: CLI ref update needed -->` to the related page and flag for engineering. | — |
| New or changed API endpoint or field | Do not edit — REST API reference is auto-generated. Flag for engineering. | — |
| New or changed MCP server tool | Load `.agents/skills/update-mcp-server-docs/SKILL.md` | `docs/installation-and-configuration/mcp-server.md` |
| New concept or platform-level behavior | Create a conceptual page in the relevant `docs/` subdirectory | Similar concept page |
| New error message or troubleshooting step | Find and update the relevant how-to or infrastructure page | Nearest existing page |

When no exact match exists, find the most similar existing page in `docs/` and follow its structure.

## Mapping Code Changes to Doc Pages

When you receive a set of changed source files, follow these steps to determine
which doc pages need updating.

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
4. **Check `docs/.docs-skip`.** If the affected path is listed in
   `docs/.docs-skip`, do not edit it. Add a `<!-- TODO: update needed -->` comment
   to the nearest related page and flag for engineering.
5. **Determine product scope.** Confirm which products (Self-Hosted, SaaS,
   Multi-Tenant) the change applies to. Register or update the page in every
   applicable nav file under `fern/products/`.
6. **Keep changes together.** If a single code change touches multiple doc
   areas, include all affected pages in one PR rather than splitting across
   multiple PRs.

## What's New and Release Notes

What's New and hotfix pages live under `docs/getting-started/whats-new/`.
Do not create new pages in this directory without confirming the version
file does not already exist.

| File pattern | Purpose |
|---|---|
| `whats-new-X-Y.md` | Features and enhancements for version X.Y |
| `hotfixes-X-Y.md` | Hotfix patches applied after the X.Y GA release |
| `deprecations.md` | Features scheduled for removal in a future version |

### Feature and Enhancement Entries

Add a new `##` section per significant feature or behavior change.
Lead with the user-visible impact, not the implementation detail.

- Use present tense: "The Scheduler now supports…"
- Name the UI element, CLI flag, or API field explicitly.
- Add a `<Badge>` for experimental or beta features.
- Link to the relevant how-to or feature guide page.

### Hotfix Entries

Write each hotfix entry in past tense:

> Fixed an issue where [specific behavior] [in what context].

Rules:
- No "why" clauses or root-cause speculation.
- Use "submitted" not "launched" for workload actions.
- Consolidate multiple entries that share the same CVE or issue ID into one line.

### What Not to Include

Do not add What's New entries for:
- API reference doc updates
- Internal telemetry or logging changes
- Test fixes
- Changes marked "no docs needed"
- SaaS-only fixes on Self-Hosted What's New pages, and vice versa

## Writing Style Guide

Apply these rules to documentation, examples, headings, UI text, and release
notes that you create or edit.

- Write in a professional, active, conversational, and engaging voice.
- Use active voice whenever possible. Use present tense for product behavior.
  Address the reader in second person as "you."
- Keep sentences concise. Prefer sentences with fewer than 30 words.
- Use plain English and precise technical terms. Avoid jargon, filler,
  colloquialisms, and flowery marketing claims.
- Avoid contractions in technical documentation. Write "do not," "cannot,"
  and "it is."
- Write "NVIDIA" in all caps and use "an NVIDIA," not "a NVIDIA."
- Spell out uncommon abbreviations on first use. Spell out LLM, RAG, SLM, VLM,
  and MoE on first use.
- Use NVIDIA spellings such as data center, dataset, open source, pretrained,
  startup, webpage, website, and Wi-Fi.
- Replace Latinisms with plain English. Use "for example," "that is," "and so
  on," "through," and "compared to."
- Use "refer to" instead of "see," "can" instead of "may" for possibility,
  and "after" instead of "once" for time.
- Do not use "please" in technical instructions.
- Use numerals for specific values, parameters, measurements, and values of 10
  or more. Spell out zero through nine in general prose.
- Include a space between a number and its unit. Use a comma in numbers with
  four or more digits.
- Use title case for headings. Do not style headings with code, bold, italics,
  quotation marks, ampersands, or exclamation marks.
- Use the Oxford comma. Put periods inside quotation marks in U.S. style.
- Use hyphens only for compound modifiers before nouns. Do not hyphenate an
  adverb that ends in "ly."
- Format commands, code, filenames, paths, and API identifiers as code. Use
  bold for UI elements and the greater-than sign for UI navigation.
- Introduce lists, tables, code examples, and images with a complete sentence.
  Use parallel construction in lists.
- Use descriptive link text. Do not use raw URLs in running text or generic
  link text such as "click here" or "read more."
- Write dates as Month DD, YYYY. Omit the year when it matches the publication
  year. Write time with a 12-hour clock and include minutes only when needed.
- Do not rewrite quoted UI labels, API field names, or audience role labels in
  tables to enforce second person.
- Provide useful alt text and preserve a logical heading hierarchy.
- Verify commands, flags, API names, defaults, and technical claims against
  source code or another checked-in source of truth.
- Do not rewrite literal code, identifiers, commands, URLs, or quoted terminal
  and API output to satisfy prose rules.
- Apply rules to improve clarity. Do not make mechanical changes that reduce
  technical accuracy or readability.
- Put the condition before the instruction: "To enable GPU fractions, set
  `gpuFractionEnabled: true`" — not the reverse.
- One action per numbered step.
- Remove filler words: "simply," "just," "in order to," "it is important to
  note."
- Remove marketing language: "seamless," "powerful," "cutting-edge," "unlock,"
  "leverage," "comprehensive."
- Name the thing. Write the flag, field, or menu path — not "configure the
  appropriate settings."

## Terminology

Always write "NVIDIA Run:ai" in prose — never "Run:ai" alone. Use `runai`
only for CLI commands, API flags, and inline code.

| Term | Rule |
|---|---|
| Product name | **NVIDIA Run:ai** in prose — never "Run:ai" alone |
| In code | `runai` — CLI commands, API flags, inline code only |
| Scheduler | **NVIDIA Run:ai Scheduler** on first mention; "the Scheduler" after |
| GPU / CPU | Always uppercase |
| Kubernetes | Never "k8s" in prose |
| CLI version | **CLI v2** — not "CLI v2.0" or "new CLI" |
| Deployment types | **Self-Hosted**, **SaaS**, **Multi-Tenant** — always capitalized |
| workload | Not "job" (except in CLI reference) |
| workspace | Not "interactive workload" |
| organization | Not "tenant" or "customer" |
| node pool | Not "node group" |
| GPU fraction | Not "fractional GPU" |

Expand acronyms on first use per page: "Multi-Instance GPU (MIG),"
"persistent volume claim (PVC)."

For the complete terminology reference, see `docs/terminology.md`.

## Page Types

Each page serves one purpose. Match the type to the content:

| Type | H1 form | Use when |
|---|---|---|
| Conceptual | Bare noun: `Introduction to Workloads` | Background and behavior, no steps |
| How-to | Gerund + noun: `Running Workspaces` | Step-by-step procedure for a single task |
| Feature guide | Gerund or benefit phrase | Background is inseparable from configuration |
| Infrastructure guide | Technology noun | Infrastructure setup (Helm, Kubernetes manifests) |
| Tutorial | Descriptive workflow title | Multi-task end-to-end workflow |
| Feature quick start | Gerund + feature name | Single feature, end-to-end walkthrough |
| Persona quick start | `Quick Start for [Persona]` | Role-based platform onboarding |
| Reference | Noun + "Reference" | Exhaustive field, flag, or parameter listing |
| Overview | `Overview` or section noun | High-level entry point for a section |

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

Do not skip heading levels. Use `##` → `###` → `####` only.

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
| `layout:` | Home page (`docs/README.md`) only |

## Auto-Generated Content

Do not edit these sections directly — changes must be made in the source
engineering repositories:

| Section | Source location |
|---|---|
| CLI reference (`docs/reference/cli/`) | `control-plane/runai-cli/docs` in the Run:ai control-plane repo |
| REST API reference | Swagger source files at `control-plane/<service>/swagger/source` |

## Fern Components

Use Fern MDX components. Do not use raw HTML. Always leave a blank line before
and after opening and closing tags.

Never use plain text for callouts or warnings. Every note, tip, warning, or
important callout must use a Fern component — not bold text, blockquotes, or
any other markdown construct.

- `<Info>` — notes and supplemental information; open body with **Note**
- `<Warning>` — caveats, prerequisites, breaking changes; use **Important**,
  **Prerequisite**, **Warning**, or **Breaking change from version X.Y:**
- `<Callout intent="info">` — short single-sentence informational notes
- `<Accordion title="...">` — optional detail most readers can skip
- `<Tabs>` / `<Tab title="...">` — steps that differ by interface; standard
  titles are `UI`, `CLI v2`, `API`
- `<Cards>` / `<Card>` — landing and index pages only; not in content pages
- `<Badge>` — inline feature lifecycle: `Experimental` (`success`), `Beta`
  (`warning`), `Legacy` (`note`)

## Navigation

Register new pages in the correct product nav file. A page with no nav entry
is unreachable:

| Content applies to | Nav file |
|---|---|
| OSS / Self-Hosted | `fern/products/self-hosted.yml` |
| Self-Hosted + SaaS | `fern/products/self-hosted.yml` + `fern/products/saas.yml` |
| SaaS-specific | `fern/products/saas.yml` |
| Multi-Tenant | `fern/products/multi-tenant.yml` |

## Links

- Within `docs/`: use relative paths with the file extension.
- To the REST API reference: use the versioned API docs URL.
- To external content: use the full URL; link text must name the destination.
- UI navigation paths: bold element names and use → as the separator.

## Verification

Run these commands before opening a pull request:

| Command | When to run |
|---|---|
| `fern docs dev` | After any change to `docs/` or `fern/` — validates the full site locally |
| `fern check` | After adding or changing OpenAPI specs in `fern/apis/` |

If `fern docs dev` fails, resolve all errors before opening the PR. Do not open
a PR with a broken local build.

## Pre-PR Checklist

Before opening a pull request, verify:

- [ ] File is placed in the correct `docs/` subdirectory
- [ ] Nav entry added to every applicable product file under `fern/products/`
- [ ] `# H1` title present in the body
- [ ] All Fern components have blank lines before and after tags
- [ ] Links use relative paths with extensions (within `docs/`) or full URLs
- [ ] No internal references: no Jira IDs, internal hostnames, `TODO`/`FIXME`
- [ ] Terminology follows the table above
- [ ] Code blocks tagged with a language
- [ ] `fern docs dev` runs without errors
