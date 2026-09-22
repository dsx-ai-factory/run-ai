---
name: runai-docs-feature-guide
description: Generate a feature guide page for an NVIDIA Run:ai scheduling or infrastructure feature — concept explanation, benefits, how it works, prerequisites, enabling/configuring steps, and Using API. Use for features like GPU optimization modes, topology-aware scheduling, NUMA scheduling, and similar platform capabilities.
---

# NVIDIA Run:ai Feature Guide Documentation

Use when documenting a new platform feature that requires both explanation
and configuration steps — not a pure UI resource page, not an installation
procedure, but a capability that administrators enable and configure.

Examples: GPU memory swap, dynamic GPU fractions, topology-aware scheduling,
NUMA-aware scheduling, time-slicing, new scheduling policies.

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

Document only what the code confirms. Mark anything unverifiable as
`<!-- TODO: verify with engineering -->`.

## Page to Create

One page at the most relevant path under `docs/platform-management/` or
`docs/installation-and-configuration/`. File name: kebab-case, descriptive of the
feature (`gpu-memory-swap.md`, `numa-aware-scheduling.md`).

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

### Using API

```markdown
## Using API

To view the available actions, go to the [Feature Name](https://run-ai-docs.nvidia.com/api/2.26/[path]) API reference.
```

<!-- TBD: Confirm OSS API reference URL pattern — current links use
run-ai-docs.nvidia.com/api/2.26/ but OSS URL is not yet confirmed. Use a
placeholder and mark with TODO until resolved. -->

## Nav Registration

Add a nav entry in `fern/products/self-hosted.yml` under the most relevant
section. For infrastructure features, also add to `fern/products/saas.yml`
if applicable. For scheduler optimization features, place under
`Managing Your Resources` or the relevant scheduler section.

## Quality Checks

Before finishing:

- [ ] All Helm values, kubectl commands, and UI paths verified against source
- [ ] No invented defaults — `<!-- TODO -->` for anything unverified
- [ ] `<Warning>` used for any destructive or irreversible configuration
- [ ] `<Info>` used for notes and tips — never plain text
- [ ] Deployment Considerations covers known interactions with other features
- [ ] Nav entry added to applicable product files
- [ ] `fern docs dev` runs without errors
