# Eval: New UI Resource — Departments

## Input

> A new UI resource called Departments has been added to the NVIDIA Run:ai
> platform. It has the following UI flows: create, edit, delete, viewing
> scheduling rules, applying access rules, and viewing the effective policy.
> Departments group multiple projects under a shared organizational scope.

## Expected Output

A single page at `docs/platform-management/aiinitiatives/organization/departments.md`
following this exact structure:

### Structure

- **H1**: `# Departments`
- **Opening paragraph**: 2–3 sentences explaining what departments are and what
  they enable (grouping projects, managing resources at scale, applying policies)
- **`## Departments Table`**: where to find it in the UI, what it lists; column
  definitions table; sub-tables for Node Pools, Projects, Subjects, Workloads;
  Customizing the Table View bullet list
- **`## Creating a Department`**: numbered steps through the full wizard —
  scope selection, naming, GPU quota, over-quota, CPU quota (opt-in), scheduling
  rules, access rules
- **`## Editing a Department`**: numbered steps with `<Info>` note about
  constraints
- **`## Viewing a Department's Policy`**: numbered steps with `<Info>` note
  about policy inheritance
- **`## Deleting a Department`**: numbered steps with `<Warning>` about
  cascading deletion
- **`## Using CLI`**: one line linking to `docs/reference/cli/runai/runai-department.md`
- **`## Using API`**: one line linking to the Departments API reference

### Pass Criteria

- [ ] One page produced — not two separate pages
- [ ] `<Warning>` used for the delete confirmation — never plain text
- [ ] `<Info>` used for all notes — never plain text
- [ ] All cross-references are markdown links — never plain text
- [ ] Using CLI appears before Using API
- [ ] Nav entry added to `fern/products/self-hosted.yml` and `fern/products/saas.yml`
- [ ] No invented field names or behaviors — everything grounded in source

## Reference

Compare output against:
`docs/platform-management/aiinitiatives/organization/departments.md`
