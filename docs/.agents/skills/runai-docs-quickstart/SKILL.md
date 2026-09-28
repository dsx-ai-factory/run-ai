---
name: runai-docs-quickstart
description: Generate a quick start page for a single NVIDIA Run:ai platform capability — step-by-step walkthrough with UI/CLI/API tabs, fixed boilerplate sections, and verified code examples. Use when adding a new quick start for a workload type or scheduler feature.
---

# NVIDIA Run:ai Quick Start Documentation

Use when adding a new quick start page — a step-by-step walkthrough that
demonstrates one platform capability using a concrete example.

## On Activation

Before drafting anything:

1. Ask: "What capability does this quick start demonstrate? What workload type
   or feature does it cover? What are the prerequisites?"
2. Read the API spec for the relevant workload type from the swagger source
   under `control-plane/<service>/swagger/source`.
3. Read these reference pages in full:
   - `docs/workloads-in-nvidia-run-ai/using-training/quick-starts/standard-training-quickstart.md`
   - `docs/platform-management/runai-scheduler/resource-optimization/quick-starts/gpu-fractions-quickstart.md`
4. Read `docs/AGENTS.md` for writing rules, Fern components, and terminology.

Document only what the code and API spec confirm. Anything you cannot verify
needs verification before it's included.

## Page to Create

One page at `docs/.../quick-starts/[feature-name]-quickstart.md` or
`docs/.../quick-starts/[workload-type]-quickstart.md`.

### Title

Two forms — use whichever matches the content:
- Workload type: `# Run Your First [Workload Type]`
- Feature/capability: `# Launching Workloads with [Feature Name]`

### Opening Paragraph

1–2 sentences defining what the walkthrough demonstrates. No steps yet.

### Prerequisites

```markdown
## Prerequisites

Before you start, make sure:

* You have created a [project](../../../platform-management/aiinitiatives/organization/projects.md) or have one created for you.
* The project has an assigned quota of at least [N] GPU.
* [Any feature-specific prerequisite — confirm in code before including]
```

Follow with this exact `<Note>` block:

```markdown
<Note>

**Note**

Flexible workload submission is enabled by default. If unavailable, contact
your administrator to enable it under **General settings** → Workloads →
Flexible workload submission.

</Note>
```

### Step 1: Logging In

Copy this block verbatim — it is identical across all quick starts:

```markdown
## Step 1: Logging In

<Tabs>

<Tab title="UI">

Browse to the provided NVIDIA Run:ai user interface and log in with your credentials.

</Tab>

<Tab title="CLI v2">

Run the below --help command to obtain the login options and log in according to your setup:

```sh
runai login --help
```

</Tab>

<Tab title="API">

To use the API, you will need to obtain a token as shown in [API authentication](/rest-api/how-to-authenticate-to-the-api).

</Tab>

</Tabs>
```

### Step 2: Submitting the Workload

Four tabs: `UI - Flexible`, `UI - Original`, `CLI v2`, `API`.

- **UI - Flexible** and **UI - Original**: numbered steps using bold UI element
  names. Start from **Workload manager** → **Workloads** → **+NEW WORKLOAD**.
- **CLI v2**: one `runai` command with `project-name` and `workload-name`
  placeholders. Link to the specific command page under
  `docs/reference/cli/runai/` — pages are named `runai_<noun>_<verb>.md` (e.g.
  `runai_workload_submit.md`) — find the actual file for the command rather
  than guessing the filename.
- **API**: `curl` command with `<COMPANY-URL>`, `<TOKEN>`, `<PROJECT-ID>`,
  `<CLUSTER-UUID>` placeholders. Follow with bullet explanations of each
  placeholder and links to the API calls that return those values.

End the API tab with this exact `<Note>` block:

```markdown
<Note>

**Note**

The above API snippet runs with NVIDIA Run:ai clusters of 2.18 and above only.

</Note>
```

For feature quick starts (not workload type): Step 2 may be split into multiple
steps (e.g., "Submit workload with feature enabled" + "Verify feature behavior").
Follow the gpu-fractions-quickstart.md structure for these.

### Verifying Each Tab's Example

A quick start covers three surfaces, and each one has its own source of
truth — verify each tab against that surface, not against one of the other
tabs:

- **UI tab**: verify the labels, button names, and flow against the current
  product UI or its frontend source. Do not infer UI steps from the CLI or
  API tab.
- **CLI v2 tab**: verify the command and flags against the CLI source under
  `control-plane/runai-cli`, not against `docs/reference/cli/runai/` — that
  directory is generated output, not the source.
- **API tab**: verify the endpoint, fields, and payload against the swagger
  source under `control-plane/<service>/swagger/source`, not against
  `fern/apis/` — those OpenAPI specs are themselves generated output copied
  from the swagger build, and not against the CLI command that happens to
  call the same operation.

### Next Steps

```markdown
## Next Steps

* Manage and monitor your newly created workload using the [Workloads](../../workloads.md) table.
* [Link to the next logical step in the user journey — e.g., inference after training]
```

## Nav Registration

Place the page in the `quick-starts/` subdirectory under the relevant section.
Add a nav entry in `fern/products/self-hosted.yml` under the appropriate
parent section.

## Quality Checks

Before finishing:

- [ ] Step 1 Login block is verbatim — do not paraphrase
- [ ] API `<Note>` about 2.18+ is present on the API tab
- [ ] All placeholder values (`<COMPANY-URL>`, `<TOKEN>`, etc.) have bullet explanations
- [ ] UI steps (labels, buttons, flow) verified against the product UI, not inferred from the CLI or API tab
- [ ] API call verified against the swagger source under `control-plane/<service>/swagger/source`, and links to the correct versioned API reference
- [ ] CLI command verified against the CLI source under `control-plane/runai-cli`
- [ ] Nav entry added to self-hosted.yml
- [ ] `fern docs dev` runs without errors
