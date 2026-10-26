---
name: runai-docs-feature-guide
description: Generate a feature guide page for an NVIDIA Run:ai scheduling or infrastructure feature — concept explanation, benefits, how it works, prerequisites, enabling/configuring steps, and Using API. Use for features like GPU optimization modes, topology-aware scheduling, NUMA scheduling, and similar platform capabilities.
---

# NVIDIA Run:ai Feature Guide Documentation

Use when documenting a new platform feature that requires both explanation
and configuration steps — not a pure UI resource page, not an installation
procedure, but a capability that administrators enable and configure.

Examples: dynamic GPU fractions, topology-aware scheduling, NUMA-aware
scheduling, time-slicing, new scheduling policies.

## On Activation

Before drafting anything:

1. Ask: "What is the feature? What problem does it solve? Who configures it
   (infrastructure admin, platform admin)? How is it enabled (Helm, UI,
   kubectl)?"
2. Read the source code or Helm chart for the feature to verify enabling flags,
   default values, and configuration parameters.
3. Read these reference pages in full — they are your structural templates:
   - `docs/platform-management/aiinitiatives/resources/topology-aware-scheduling/README.md`
   - `docs/platform-management/aiinitiatives/resources/numa-aware-scheduling.md`
4. Read `docs/AGENTS.md` for writing rules, Fern components, and terminology.

Document only what the code confirms. Anything you cannot verify needs
verification before it's included.

## Page to Create

One page at the most relevant path under `docs/platform-management/` or
`docs/installation-and-configuration/`. File name: kebab-case, descriptive of the
feature (`numa-aware-scheduling.md`).

### Title

Use a gerund or benefit phrase:
- `# Configuring [Feature Name]`
- `# Accelerating Workloads with [Feature Name]`
- `# [Feature Name]` (for shorter, well-known feature names)

### Opening Paragraph

2–3 sentences explaining what the feature does and what problem it solves.
No steps, no bullet lists yet. Ground every claim in verified source behavior.

### Benefits Section

```markdown
## Benefits of [Feature Name]

* **[Benefit 1]** — [one sentence explanation]
* **[Benefit 2]** — [one sentence explanation]
* **[Benefit 3]** — [one sentence explanation]
```

Include 3–5 benefits. If the feature benefits both AI practitioners and
platform administrators differently, split into sub-bullets by persona.

### How It Works

```markdown
## How [Feature Name] Works
```

Explain the mechanism. Include diagrams or code examples only when verified
from source. For multi-GPU variants of a feature, add:

```markdown
## Multi-GPU [Feature Name]
```

### Prerequisites

Include when the feature requires specific infrastructure setup before
enabling. List concrete technical conditions — Kubernetes operator versions,
node labels, kubelet configurations. Link to external documentation where
relevant.

### Enabling and Configuring

```markdown
## Enabling and Configuring [Feature Name]
```

Provide the exact steps to enable the feature. Common patterns:

- **Helm**: show the exact Helm values block verified from the chart source
- **General Settings UI**: bold the path: **General settings** → **Resources**
  → **[Setting name]**
- **kubectl patch**: show the exact command

Use `<Tabs>` only when the feature has meaningfully different enabling steps
per environment (Kubernetes vs. OpenShift). Do not use tabs just for style.

For sub-configurations within the feature, use H3 sections:

```markdown
### [Sub-configuration Name]
```

### Deployment Considerations

Include when the feature has known limitations, interactions with other
features, or performance trade-offs the administrator should know before
enabling. Common examples: dependency on another feature being enabled first,
memory overhead, incompatibility with certain workload types.

### Using CLI

Include one line linking to the specific command page under
`docs/reference/cli/runai/`, if a CLI command exists for this feature. Pages
are named `runai_<noun>_<verb>.md` (e.g. `runai_department_list.md`) — find
the actual file for the resource rather than guessing the filename, and link
to that file, not to the bare directory. Only include if confirmed in the
code.

```markdown
## Using CLI

To view the available actions on [feature name], see the [feature name] [CLI v2 reference](../../reference/cli/runai/runai_<noun>.md).
```

### Using API

Use Fern's `api:` link syntax, not a hardcoded domain. It resolves to the
correct endpoint page at build time regardless of URL pattern or version.
Point at the specific `METHOD path` that matches the action described (for
example, the `POST` endpoint if the text is about creating a resource, the
`GET` list endpoint if it is about viewing available actions) — never link to
just the API root or a bare tag name:

```markdown
## Using API

To view the available actions, go to the [Feature Name](api:[api-name]:[METHOD path]) API reference.
```

### Using MCP

Include one line naming the relevant MCP tool from
`docs/agents/tools-reference.md`, if the feature has a real, confirmed tool
there — check the domain tables (Identity, Credentials, Workloads, Cluster
and Infrastructure, Resources and Organization) rather than guessing a tool
name. MCP tools are not individually addressable pages, so link to the
domain section anchor (e.g. `../../agents/tools-reference.md#workloads`), not
a per-tool anchor, and name the tool itself in backticks. Only include if
confirmed in `tools-reference.md`.

```markdown
## Using MCP

The NVIDIA Run:ai MCP server exposes the [`tool_name`](../../agents/tools-reference.md#domain-section) tool to [do X] via natural language. See [AI Agents](../../agents/overview.md) for setup.
```

## Nav Registration

Add a nav entry in `fern/products/self-hosted.yml` under the most relevant
section. For scheduler optimization features, place under
`Managing Your Resources` or the relevant scheduler section.

## Quality Checks

Before finishing:

- [ ] All Helm values, kubectl commands, and UI paths verified against source
- [ ] No invented defaults — anything unconfirmed needs verification first
- [ ] `<Warning>` used for any destructive or irreversible configuration
- [ ] `<Note>` used for notes and tips — never plain text
- [ ] Deployment Considerations covers known interactions with other features
- [ ] Nav entry added to applicable product files
- [ ] `fern docs dev` runs without errors
