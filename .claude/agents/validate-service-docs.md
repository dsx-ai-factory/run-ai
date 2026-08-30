---
name: validate-service-docs
description: Validate a service README against its Makefile targets. For a single service provide its path; say "all" to check every undocumented service listed in CLAUDE.md.
tools:
  - Bash
  - Read
---

## Finding the repo root
```bash
REPO_ROOT=$(git rev-parse --show-toplevel)
```
All paths are relative to `$REPO_ROOT`. Never use absolute user-specific paths.

## Modes

**Single service:** user provides e.g. `control-plane/authorization`. Validate that one.

**Bulk / "all":** Read the "Missing documentation" section from `CLAUDE.md` — those are already known to have no README. Report them as MISSING immediately. Then optionally spot-check a sample of the documented ones for content quality.

## What you check per service

1. **README exists** — if not: `❌ MISSING`
2. **Makefile targets covered:**
   ```bash
   # Check local Makefile AND shared includes (build/makefile/*.mk)
   grep -E "^[a-zA-Z][a-zA-Z0-9_-]*:" "$REPO_ROOT/<path>/Makefile" "$REPO_ROOT"/build/makefile/*.mk 2>/dev/null | cut -d: -f2 | sort -u
   ```
   For each of `build`, `test`, `lint`, `validate` that appears in either the service Makefile or its shared includes: check if mentioned in README using:
   ```bash
   grep -qE "make $t|\`$t\`" "$DIR/README.md"
   ```
3. **README length** — flag as TOO SHORT if under 200 characters
4. **Team name** — if in `artifact-metadata.json` but not in README, suggest adding it
5. **CLAUDE.md exists** — note if missing (informational only)

## Output format

Single service:
```
## control-plane/authorization

✅ README exists (847 chars)
❌ `make lint` target exists but not mentioned in README
⚠️  Team "Suricata Team" not in README
ℹ️  No CLAUDE.md — run /write-service-docs to generate
```

Bulk:
```
### ❌ No README (N services)
- control-plane/frontend
...

### ⚠️ README exists but has issues
| Service | Issue |
|---------|-------|
| control-plane/bff-service | `make lint` not documented |
```

## Rules
- Read-only. Never write or edit files.
- Use only paths from `CLAUDE.md` — do not scan the filesystem.
- Tell the user to run `/write-service-docs <path>` for anything that needs fixing.
