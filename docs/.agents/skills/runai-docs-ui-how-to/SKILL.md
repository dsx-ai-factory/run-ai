---
name: runai-docs-ui-how-to
description: Generate documentation for a new NVIDIA Run:ai UI resource type — one page covering the intro, table, and all UI procedures (create, edit, delete, and any additional flows), plus nav registration. Use when a new object with UI flows is added to the platform.
---

# NVIDIA Run:ai UI Resource Documentation

Use when a new UI resource type is introduced — an object that has its own
table and create/edit/delete flows in the NVIDIA Run:ai platform.

Examples: Departments, Projects, Node Pools, Environments, Compute Resources,
Data Sources, Credentials, Workload Templates.

## On Activation

Before drafting anything:

1. Ask: "What is the name of the new resource? What UI flows does it have
   (create, edit, delete, metrics, policies, access rules)?"
2. Read the API spec for this resource from `fern/apis/` if it exists.
3. Read `docs/platform-management/aiinitiatives/organization/departments.md`
   in full — this is your structural reference.
4. Read `docs/AGENTS.md` for writing rules, Fern components, and terminology.

Document only what the code and API spec confirm. Anything you cannot verify
needs verification before it's included.

## Page to Create

One page at `docs/.../[resource-name].md` following the `departments.md`
pattern exactly:

- **H1**: bare noun — `# [Resource Name]`
- **Opening paragraph(s)**: 1-2 sentences explaining what the resource is and
  what it enables. No steps, no table yet.
- **[Resource] Table** section: where to find it in the UI, what it lists,
  then a column definitions table. Include a `<figure>` image placeholder if
  a screenshot exists in `docs/assets/`.
- **One section per UI flow**:
  - `## Creating a [Resource]` — numbered steps, bold UI elements
  - `## Editing a [Resource]`
  - `## Deleting a [Resource]` — include `<Warning>` for destructive action
  - Additional sections for any other flows (scheduling rules, access rules,
    policies, metrics)
- Use `<Warning>` for destructive actions and important caveats — never plain text
- Use `<Note>` for notes about defaults, limits, or constraints — never plain text
- Bold all UI element names: **+NEW RESOURCE**, **SAVE**, **DELETE**
- **`## Using CLI`** at the end: one line linking to the specific command page
  under `docs/reference/cli/runai/`, if a CLI command exists for this
  resource. Pages are named `runai_<noun>_<verb>.md` (e.g.
  `runai_department_list.md`) — find the actual file for the resource rather
  than guessing the filename, and link to that file, not to the directory.
  Only include if confirmed in the code.
- **`## Using API`** after Using CLI: one line linking to the API reference, if the resource has an API. Use Fern's `api:` link syntax (e.g. `api:[api-name]:[METHOD path]`), not a hardcoded domain — it resolves at build time regardless of URL pattern or version.

## Nav Registration

Register the page in the correct product nav files under `fern/products/`.
Add to every product the resource applies to (`self-hosted.yml`,
`multi-tenant.yml` as appropriate).

## Quality Checks

Before finishing:

- [ ] All field names, defaults, and behaviors verified against API spec or
  source code — no invented values
- [ ] No unverified claims — anything not confirmed by source needs verification first
- [ ] `<Warning>` used for every destructive action — never plain text
- [ ] `<Note>` used for every note — never plain text
- [ ] All cross-references are markdown links — never plain text
- [ ] Nav entries added to all applicable product files
- [ ] `fern docs dev` runs without errors
