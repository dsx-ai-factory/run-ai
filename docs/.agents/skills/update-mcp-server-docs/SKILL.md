---
name: update-mcp-server-docs
description: Detect changes to MCP server tools (runai-mcp-server/internal/tools/) and agent skills (skills/) on the current branch and update docs/agents/tools-reference.md and docs/agents/agent-skills.md. Run with no arguments — detects changes automatically from git diff against the default branch. Finishes by creating a new branch and opening a PR for the user and the docs reviewer (Sherin) to approve.
allowed-tools: Bash, Read, Edit
---

Detect changes to MCP tools in `runai-mcp-server/internal/tools/` and to agent skills in `skills/` on the current branch, and keep these two pages in sync:

- `docs/agents/tools-reference.md` — MCP tools.
- `docs/agents/agent-skills.md` — agent skills.

**Important distinctions before starting:**

- The `Description` field in Go source (tools) and the `description` frontmatter in `SKILL.md` (skills) are **model-facing** — optimized for LLMs deciding when and how to call the tool or trigger the skill. They are not copied verbatim into the docs.
- The docs use **user-facing** descriptions — one sentence explaining what the tool or skill does in plain English for a human reader.
- `tools-reference.md` groups tools by **logical domain**, not by source directory. There is no separate Write Tools section — read and write tools appear together within each domain group, distinguished by the `Type` column (`Read` or `Write`).
- `agent-skills.md` lists skills in one `Available Skills` table.
- Not all registered tools or skills appear in the docs. New ones need a judgment call on whether they warrant a doc entry — show the user the tool or skill and ask.
- **A skill or tool is only documented for a release if it exists on that release's branch.** Check this in Step 1b before documenting anything for a release.

---

## Step 1 — Find the repo root and baseline

```bash
REPO_ROOT=$(git rev-parse --show-toplevel)
DEFAULT=$(git symbolic-ref --short refs/remotes/origin/HEAD | sed 's#origin/##')
BASE=$(git merge-base "origin/$DEFAULT" HEAD)
CHANGED_TOOLS=$(git diff --name-only "$BASE"...HEAD | grep 'runai-mcp-server/internal/tools/')
CHANGED_SKILLS=$(git diff --name-only "$BASE"...HEAD | grep '^skills/')
echo "$CHANGED_TOOLS"
echo "$CHANGED_SKILLS"
```

`$DEFAULT` is the repo's default branch (currently `main`). Do not hardcode `master`.

If both `CHANGED_TOOLS` and `CHANGED_SKILLS` are empty, tell the user "No MCP tool or agent skill changes detected against the default branch" and stop.

If the user names tools or skills that were merged earlier (nothing in the current diff), skip the diff and take the names they give you. For skills, compare the skills on the default branch against the rows already in `agent-skills.md`:

```bash
git ls-tree -d --name-only "origin/$DEFAULT" skills/ | sed 's#skills/##'
```

### Step 1b — Confirm the release branch has them

When the docs change targets a release (for example `release-3.0`), or the user asks whether something is "in 3.0":

```bash
git fetch origin
git ls-tree -d --name-only origin/release-<X.Y> skills/ | sed 's#skills/##'
git ls-tree -r --name-only origin/release-<X.Y> runai-mcp-server/internal/tools/ | grep register.go
```

Also check the latest tag for that release. If a skill or tool is **not** on the release branch, say so, do not add it to that release's docs, and ask the user whether to wait for a cherry-pick. A claim that something was cherry-picked is not evidence — verify on the branch or the tag.

---

## Step 2 — Categorize the changes

### Tools

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

### Skills

**Category D — Skill added or removed**
Signal: a `skills/<skill-name>/SKILL.md` file was added or deleted.
```bash
git diff --name-status "$BASE"...HEAD | grep -E '^[AD]\s+skills/[^/]+/SKILL\.md'
```

**Category E — Skill name or description changed**
Signal: `skills/<skill-name>/SKILL.md` was modified (frontmatter `name` or `description`, or the skill's purpose changed).
```bash
git diff --name-status "$BASE"...HEAD | grep -E '^M\s+skills/[^/]+/SKILL\.md'
```

Changes only under `skills/<skill-name>/evals/` or `references/` do not need a docs change.

Tell the user which categories triggered before proceeding.

---

## Step 3 — Extract metadata from source

### Tools

For each affected tool, read its `.go` file and extract:
- `Name` — the string value of the `Name:` field in `runaimcp.Tool{...}`
- `ReadOnly` — `true` means read-only; `false` means write tool
- `Destructive` — `true` if present (used to flag especially risky write tools)

Do **not** copy the source `Description` — it is model-facing and not appropriate for the doc. You will draft a user-facing description in Step 5.

For removed tools, get the tool `Name` from the old version:
```bash
git show "$BASE":runai-mcp-server/internal/tools/<group>/<tool>.go | grep 'Name:'
```

### Skills

For each affected skill, read `skills/<skill-name>/SKILL.md` and extract:
- `name` — the frontmatter `name` (matches the directory name)
- The frontmatter `description` and its "Use for ..." trigger phrases — use these to draft an example prompt
- Whether the skill only reads data or can also change state. Skills that call write tools are `Write`; all others are `Read`.

Do **not** copy the frontmatter `description` — it is model-facing (trigger-rich, with "Not for ..." exclusions). You will draft a user-facing description in Step 5.

For removed skills, get the `name` from the old version:
```bash
git show "$BASE":skills/<skill-name>/SKILL.md | grep '^name:'
```

---

## Step 4 — Read the current docs

### Tools — `docs/agents/tools-reference.md`

Read `$REPO_ROOT/docs/agents/tools-reference.md`.

The page has `## <Group>` subsections, each with a three-column table:

```
| Tool | Type | Description |
|---|---|---|
| `tool_name` | Read | Description. |
| `tool_name` | Write | Description. |
```

`Type` is `Read` when `ReadOnly: true` in source, `Write` when `ReadOnly: false`. There is no separate Write Tools section — all tools appear within their domain group.

Note which tools are currently documented and in which group.

### Skills — `docs/agents/agent-skills.md`

Read `$REPO_ROOT/docs/agents/agent-skills.md`.

The `## Available Skills` section has one four-column table:

```
| Skill | Type | Description | Example Prompt |
|---|---|---|---|
| [`skill-name`](<link to the skill directory>) | Read | One-sentence description. | "Example prompt." |
```

Note which skills are currently documented. Also check that the existing skill links point at the repo's default branch (`main`), not a branch that no longer exists. If they point elsewhere, tell the user and offer to fix them in the same PR.

---

## Step 5 — Determine and propose changes

### Tools

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

### Skills

**Skill removed:**
- Check if the skill appears in `agent-skills.md`.
- If yes: show the row and ask the user to confirm removal.
- If no: tell the user the skill was removed from source but was not in the doc — no action needed.

**Skill added:**
- Check if the skill already appears in the doc.
- If already there: no action needed.
- If not, and Step 1b confirmed it is on the target branch: draft a one-sentence user-facing description (plain English, describing what the skill produces for the user) and one example prompt taken from the skill's "Use for ..." phrases. Show the user:
  - The skill name
  - Its `Read` / `Write` type
  - Your proposed one-sentence description and example prompt
  - The proposed table row
  Wait for approval before writing.

**Skill name changed:**
- Find the old name in the doc and propose a rename, including the link. Show before/after. Wait for approval.

**Skill description or purpose changed:**
- Ask the user: "The model-facing description for `<skill-name>` changed. Does the user-facing doc description or example prompt also need updating?" Show the current doc entry. If yes, draft a revised one and show it for approval.

---

## Step 6 — Apply approved changes

### Tools

Insert or update rows using this three-column format:
```
| `tool_name` | Read | One-sentence user-facing description. |
| `tool_name` | Write | One-sentence user-facing description. |
```

- **Read tools** (`ReadOnly: true`): `Read` in the Type column.
- **Write tools** (`ReadOnly: false`): `Write` in the Type column. If `Destructive: true`, note it in the description (e.g. "Deleting a project permanently deletes all its workloads.").
- Insert new rows in alphabetical order by tool name within the correct domain section.
- Write tools go in the same domain section as their read counterparts — not in a separate section.

For **new group** (Category C): add a new `## <Group Name>` subsection with the three-column table header. Ask the user what the logical group name should be — do not assume the directory name maps directly to a doc section name.

### Skills

Insert or update rows in the `## Available Skills` table using this four-column format:
```
| [`skill-name`](https://github.com/dsx-ai-factory/run-ai/tree/main/skills/skill-name) | Read | One-sentence user-facing description. | "Example prompt." |
```

- Link each skill to its directory on the default branch (`main`).
- Keep existing rows in their current order and append new skills at the end, unless the user asks for another order.
- Keep the description to one sentence and the example prompt to one natural-language question or request.

---

## Step 7 — Create a branch and open a PR

Do **not** commit to `main` or to a release branch directly. Once the user has approved the changes, put them on a new branch and open a PR for the user and the docs reviewer (Sherin) to approve.

1. Create a new branch from the right base: `origin/main` for the default-branch docs; for a release (for example `release-3.0`), branch from `origin/release-<X.Y>` and target that branch. Prefer a `git worktree` so the user's working copy is untouched:
   ```bash
   git fetch origin
   git worktree add -b docs/update-agent-docs-<short-topic> <path> origin/<base-branch>
   ```
   Apply the approved edits in the new branch, not in the original checkout.
2. Commit only the docs files you changed (`docs/agents/tools-reference.md` and/or `docs/agents/agent-skills.md`) with a clear, self-contained message. Add DCO sign-off (`git commit -s`) if the repo requires it for the user's account.
3. Push the branch and open a **draft** PR against the base branch:
   ```bash
   git push -u origin docs/update-agent-docs-<short-topic>
   gh pr create --draft --base <base-branch> --title "<clear, self-contained title>" --body-file <body.md>
   ```
   Fill in the repo's PR template (`.github/pull_request_template.md`): why the change is needed, the issue reference (`Relates to #<n>` when there is one), and the checklist. List the tools and skills that were added, removed or renamed.
4. Request review from Sherin (`SherinDaher-Runai`):
   ```bash
   gh pr edit <pr-number> --add-reviewer SherinDaher-Runai
   ```
   The user approves the PR together with Sherin. Do not merge it yourself.
5. Never push to `main` or a release branch directly, never force-push, and never merge.

If the user wants to review the edits before a PR exists, stop after Step 6, list the files changed, and run this step only when they say so.

---

## Step 8 — Report

Summarize:
- Tools added, removed, or renamed in `tools-reference.md`
- Skills added, removed, or renamed in `agent-skills.md`
- Tools or skills whose source description changed but whose doc entry was left unchanged (with a note to revisit)
- Any tools or skills in source that were not added to the doc (per user's decision), and any left out because they are not yet on the target release branch
- The branch name, the PR link, and who was asked to review
