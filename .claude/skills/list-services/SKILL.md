---
name: list-services
description: List all buildable Run.ai services grouped by type. Reads from CLAUDE.md — no repo scan needed.
allowed-tools: Read
---

The root `CLAUDE.md` already contains the full service list. Read it and present:
1. A summary count: total services, breakdown by type (container / go / cli / helm)
2. The "Missing documentation" section — services with no README
3. The full documented services list grouped by subsystem

Read `CLAUDE.md` from the repo root (find it with: the file is at the root of the current git repo — use `git rev-parse --show-toplevel` if needed to locate it).

Present the output cleanly. Do not scan the filesystem.
