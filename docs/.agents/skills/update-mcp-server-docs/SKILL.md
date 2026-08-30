---
name: update-mcp-server-docs
description: Detect changes to MCP server tools on the current branch and update the Available Tools section in docs/installation-and-configuration/mcp-server.md. Run with no arguments — detects changes automatically from git diff against master.
allowed-tools: Bash, Read, Edit
---

Detect changes to tools in `runai-mcp-server/internal/tools/` on the current branch and keep the `## Available Tools` section of `docs/installation-and-configuration/mcp-server.md` in sync.

**Important distinctions before starting:**

- The `Description` field in Go source is **model-facing** — optimized for LLMs deciding when and how to call the tool. It is not copied verbatim into the doc.
- The doc uses **user-facing** descriptions — one sentence explaining what the tool does in plain English for a human reader.
- The doc groups tools by **logical domain**, not by source directory. There is no separate Write Tools section — read and write tools appear together within each domain group, distinguished by the `Type` column (`Read` or `Write`).
- Not all registered tools appear in the doc. New tools need a judgment call on whether they warrant a doc entry — show the user the tool and ask.

---

## Step 1 — Find the repo root and baseline

```bash
REPO_ROOT=$(git rev-parse --show-toplevel)
BASE=$(git merge-base origin/master HEAD)
CHANGED=$(git diff --name-only "$BASE"...HEAD | grep 'runai-mcp-server/internal/tools/')
echo "$CHANGED"
```

If `CHANGED` is empty, tell the user "No MCP tool changes detected against master" and stop.

---

## Step 2 — Categorize the changed files

**Category A — Tool added or removed**
Signal: a `register.go` file changed.
```bash
git diff --name-status "$BASE"...HEAD | grep 'runai-mcp-server/internal/tools/.*/register\.go'
```
For each changed `register.go`, diff it to see which `register<Name>` calls were added or removed:
```bash
git diff "$BASE"...HEAD -- runai-mcp-server/internal/tools/<group>/register.go
```

**Category B — Tool name or description changed**
Signal: a non-register, non-test `.go` file changed.
```bash
git diff --name-status "$BASE"...HEAD \
  | grep -E 'runai-mcp-server/internal/tools/.*\.go' \
  | grep -v 'register\.go\|_test\.go'
```

**Category C — New tool group**
Signal: `runai-mcp-server/internal/tools/tools.go` changed.

Tell the user which categories triggered before proceeding.

---

## Step 3 — Extract tool metadata from source

For each affected tool, read its `.go` file and extract:
- `Name` — the string value of the `Name:` field in `runaimcp.Tool{...}`
- `ReadOnly` — `true` means read-only; `false` means write tool
- `Destructive` — `true` if present (used to flag especially risky write tools)

Do **not** copy the source `Description` — it is model-facing and not appropriate for the doc. You will draft a user-facing description in Step 5.

For removed tools, get the tool `Name` from the old version:
```bash
git show "$BASE":runai-mcp-server/internal/tools/<group>/<tool>.go | grep 'Name:'
```

---

## Step 4 — Read the current doc

Read `$REPO_ROOT/docs/installation-and-configuration/mcp-server.md` and locate the `## Available Tools` section.

The section has `### <Group>` subsections, each with a three-column table:

```
| Tool | Type | Description |
|---|---|---|
| `tool_name` | Read | Description. |
| `tool_name` | Write | Description. |
```

`Type` is `Read` when `ReadOnly: true` in source, `Write` when `ReadOnly: false`. There is no separate Write Tools section — all tools appear within their domain group.

Note which tools are currently documented and in which group.

---

## Step 5 — Determine and propose changes

**Tool removed from `register.go`:**
- Check if the tool appears in the doc.
- If yes: show the row and ask the user to confirm removal.
- If no: tell the user the tool was removed from source but was not in the doc — no action needed.

**Tool added to `register.go`:**
- Check if the tool already appears in the doc (it may have been added manually).
- If already there: no action needed.
- If not: read the source file to understand what the tool does. Draft a one-sentence user-facing description (plain English, not the model-facing source string). Show the user:
  - The tool name
  - Its `ReadOnly` / `Destructive` flags
  - Your proposed one-sentence description
  - Which doc section you'd put it in (based on logical domain; ask if uncertain)
  - The proposed table row
  Wait for approval before writing.

**Tool name changed:**
- Find the old name in the doc and propose a rename. Show before/after. Wait for approval.

**Tool description changed in source:**
- The source description changed but it's model-facing. Ask the user: "The model-facing description for `<tool_name>` changed. Does the user-facing doc description also need updating?" Show the current doc entry. If yes, draft a revised user-facing description and show it for approval.

---

## Step 6 — Apply approved changes

Insert or update rows using this three-column format:
```
| `tool_name` | Read | One-sentence user-facing description. |
| `tool_name` | Write | One-sentence user-facing description. |
```

- **Read tools** (`ReadOnly: true`): `Read` in the Type column.
- **Write tools** (`ReadOnly: false`): `Write` in the Type column. If `Destructive: true`, note it in the description (e.g. "Deleting a project permanently deletes all its workloads.").
- Insert new rows in alphabetical order by tool name within the correct domain section.
- Write tools go in the same domain section as their read counterparts — not in a separate section.

For **new group** (Category C): add a new `### <Group Name>` subsection with the three-column table header. Ask the user what the logical group name should be — do not assume the directory name maps directly to a doc section name.

Do not commit. Leave that to the user.

---

## Step 7 — Report

Summarize:
- Tools added, removed, or renamed in the doc
- Tools whose source description changed but whose doc entry was left unchanged (with a note to revisit)
- Any tools added to source that were not added to the doc (per user's decision)
