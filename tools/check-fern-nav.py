#!/usr/bin/env python3
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
"""Verify every `path:` in the Fern navigation config resolves to a real file.

`fern check` validates configuration and OpenAPI specs; per docs/AGENTS.md it is
not the full-site validation (`fern docs dev` is). In particular it does not tell
you that a nav entry points at a doc that was moved, renamed, or never migrated.
That is the most common way a docs PR breaks the published site, so it gets its
own fast, dependency-free check.

Paths are resolved relative to the file that declares them, matching Fern.

Usage:
    tools/check-fern-nav.py [FERN_DIR]      # default: ./fern

Exits 1 if any referenced path is missing.
"""

from __future__ import annotations

import collections
import glob
import os
import re
import sys

# `path: <value>`, ignoring commented-out lines. Values are plain scalars in this
# workspace; strip optional quotes rather than pulling in a YAML dependency the
# runner may not have.
PATH_RE = re.compile(r"^\s*path:\s*(?P<value>\S+)")


def config_files(fern_dir: str) -> list[str]:
    """Every Fern config that can declare a nav `path:`."""
    found = sorted(glob.glob(os.path.join(fern_dir, "products", "*.yml")))
    for name in ("docs.yml", "archived-versions.yml"):
        candidate = os.path.join(fern_dir, name)
        if os.path.isfile(candidate):
            found.append(candidate)
    return found


def collect_refs(files: list[str]) -> dict[str, list[str]]:
    """Map each resolved path -> the config files that reference it."""
    refs: dict[str, list[str]] = collections.defaultdict(list)
    for config in files:
        base = os.path.dirname(config)
        with open(config, encoding="utf-8") as handle:
            for line in handle:
                if line.lstrip().startswith("#"):
                    continue
                match = PATH_RE.match(line)
                if not match:
                    continue
                value = match.group("value").strip("\"'")
                resolved = os.path.normpath(os.path.join(base, value))
                refs[resolved].append(config)
    return refs


def main(argv: list[str]) -> int:
    fern_dir = argv[1] if len(argv) > 1 else "fern"
    if not os.path.isdir(fern_dir):
        print(f"error: no such directory: {fern_dir}", file=sys.stderr)
        return 2

    files = config_files(fern_dir)
    if not files:
        print(f"error: no Fern config files found under {fern_dir}", file=sys.stderr)
        return 2

    refs = collect_refs(files)
    missing = {path: srcs for path, srcs in refs.items() if not os.path.exists(path)}

    print(f"scanned {len(files)} config file(s)")
    print(f"unique path refs : {len(refs)}")
    print(f"resolve OK       : {len(refs) - len(missing)}")
    print(f"MISSING          : {len(missing)}")

    if not missing:
        return 0

    # Group by top two path segments so a whole absent subtree reads as one
    # problem instead of hundreds of identical-looking lines.
    print("\nmissing, grouped by area:")
    grouped = collections.Counter("/".join(p.split(os.sep)[:2]) for p in missing)
    for area, count in grouped.most_common():
        print(f"  {count:5d}  {area}/")

    print("\nfirst 20 missing paths:")
    for path in sorted(missing)[:20]:
        declared_in = sorted({os.path.basename(s) for s in missing[path]})
        print(f"  {path}   (referenced by {', '.join(declared_in)})")

    print(
        "\nEvery nav entry must point at a file that exists. Either add the missing "
        "file, or remove/repoint the nav entry in the config(s) named above.",
        file=sys.stderr,
    )
    return 1


if __name__ == "__main__":
    sys.exit(main(sys.argv))
