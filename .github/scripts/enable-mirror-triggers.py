#!/usr/bin/env python3
"""Re-enable auto-firing triggers on mirrored e2e workflows.

Used by the e2e drift-detector pipeline after `git filter-repo` produces a
history-preserving mirror of runai's e2e workflows into runai-e2e-internal.
The nightly e2e pipelines run in runai-e2e-internal now, so on the
mirror the triggers that runai keeps commented must be re-enabled.

For every workflow file passed (or every *.yml/*.yaml under a directory):
  - uncomment any commented `schedule:`, `push:` and `pull_request:` trigger
    block (the key line and its commented body), preserving indentation;
  - drop the sentinel `workflow_dispatch:` line runai injected to stay manually
    runnable while the trigger was commented (marked with `# mirror-dispatch`);
  - leave `workflow_call`, real `workflow_dispatch` and everything else as is.

The transform is deterministic and idempotent.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

ENABLE_KEYS = {"schedule", "push", "pull_request"}

_ON_RE = re.compile(r"^on\s*:(.*)$")
_KEY_RE = re.compile(r"^(\s*)([A-Za-z_][A-Za-z0-9_-]*)\s*:")
_COMMENT_KEY_RE = re.compile(r"^(\s*)#\s?([A-Za-z_][A-Za-z0-9_-]*)\s*:")
_SENTINEL_RE = re.compile(r"^\s*workflow_dispatch:\s*#\s*mirror-dispatch\b")


def _indent(line: str) -> int:
    return len(line) - len(line.lstrip(" "))


def _comment_pad(line: str) -> int:
    """Indent (spaces before the '#') of a commented line."""
    return len(line) - len(line.lstrip(" "))


def _uncomment(line: str) -> str:
    """Strip the first '# ' after the leading indentation."""
    return re.sub(r"^(\s*)#\s?", r"\1", line, count=1)


def transform(text: str) -> str:
    lines = text.split("\n")
    on_idx = next((i for i, l in enumerate(lines) if _ON_RE.match(l)), None)
    if on_idx is None:
        return text
    if _ON_RE.match(lines[on_idx]).group(1).strip():
        return text

    end = len(lines)
    for i in range(on_idx + 1, len(lines)):
        l = lines[i]
        if l.strip() == "" or l.lstrip().startswith("#"):
            continue
        if _indent(l) == 0:
            end = i
            break

    # Child indent = indent of the first child key, active or commented.
    child_indent = None
    for i in range(on_idx + 1, end):
        l = lines[i]
        if l.strip() == "":
            continue
        m = _KEY_RE.match(l)
        if m:
            child_indent = len(m.group(1))
            break
        cm = _COMMENT_KEY_RE.match(l)
        if cm:
            child_indent = len(cm.group(1))
            break
    if child_indent is None:
        return text

    # Uncomment commented trigger blocks at child indent.
    i = on_idx + 1
    while i < end:
        line = lines[i]
        cm = _COMMENT_KEY_RE.match(line)
        if cm and len(cm.group(1)) == child_indent and cm.group(2) in ENABLE_KEYS:
            lines[i] = _uncomment(line)
            j = i + 1
            while j < end:
                nxt = lines[j]
                if nxt.strip() == "":
                    j += 1
                    continue
                # commented body lines are indented past the key's '#'
                if nxt.lstrip().startswith("#") and _comment_pad(nxt) > child_indent:
                    lines[j] = _uncomment(nxt)
                    j += 1
                    continue
                break
            i = j
            continue
        i += 1

    # Drop the sentinel workflow_dispatch runai injected.
    new_lines = [l for l in lines if not _SENTINEL_RE.match(l)]
    return "\n".join(new_lines)


def _iter_files(args: list[str]):
    for a in args:
        p = Path(a)
        if p.is_dir():
            yield from sorted(p.glob("*.yml"))
            yield from sorted(p.glob("*.yaml"))
        elif p.is_file():
            yield p


def main(argv: list[str]) -> int:
    if not argv:
        print("usage: enable-mirror-triggers.py <file-or-dir> ...", file=sys.stderr)
        return 2
    changed = 0
    for path in _iter_files(argv):
        original = path.read_text()
        updated = transform(original)
        if updated != original:
            path.write_text(updated)
            changed += 1
            print(f"enabled triggers: {path}")
    print(f"transformed {changed} file(s)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
