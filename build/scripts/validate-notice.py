#!/usr/bin/env python3
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
"""Check NOTICE files for the defects an OSS review rejects them for.

Usage: validate-notice.py [<path> ...]     (defaults to every NOTICE in the repo)

Exits non-zero if any file fails. Checks, per file:
  - no unresolved license-template placeholders (<YEAR>, [yyyy], <owner>, ...)
  - no dependency left as an unresolved/unknown license
  - no dependency left without a version
  - no source code embedded in place of a license text
  - every dependency row is backed by a license text or declared as having none
"""

import os
import re
import sys

PLACEHOLDERS = re.compile(
    r"<YEAR>|<COPYRIGHT HOLDER>|\[yyyy\]|\[year\]|<year>|<owner>|<OWNER>"
    r"|\[name of copyright owner\]|\[fullname\]|<name of author>|<NAME>"
)

# Lines that only appear in source files, not in license texts.
CODE = re.compile(
    r"^(package [a-z][\w/]*$|import \(|func [A-Z(]|type \w+ (struct|interface)\{?$"
    r"|#include <|module\.exports|require\(['\"])",
    re.MULTILINE,
)

UPSTREAM_NOTE = "Reproduced verbatim from upstream."

ROW = re.compile(r"^\| `([^`]+)` \| ([^|]*?) \| ([^|]*?) \| ", re.MULTILINE)
APPLIES = re.compile(r"^  - (\S+)$", re.MULTILINE)


def find_notices(paths):
    if paths:
        return paths
    found = []
    for root, dirs, files in os.walk("."):
        dirs[:] = [d for d in dirs if d not in (".git", "node_modules")]
        if "NOTICE" in files:
            found.append(os.path.join(root, "NOTICE"))
    return sorted(found)


def validate(path):
    text = open(path, encoding="utf-8", errors="replace").read()
    problems = []

    # A placeholder inside a third-party license text reproduced verbatim is the
    # upstream author's, not ours, and must not be edited out. The generator
    # annotates those blocks; anything outside such a block is our defect.
    unannotated = []
    for block in text.split("Applies to:"):
        if UPSTREAM_NOTE in block:
            continue
        unannotated.extend(PLACEHOLDERS.findall(block))
    if unannotated:
        counts = {}
        for token in unannotated:
            counts[token] = counts.get(token, 0) + 1
        shown = ", ".join(f"{k} x{v}" for k, v in sorted(counts.items()))
        problems.append(f"{len(unannotated)} unresolved placeholder(s): {shown}")

    code = CODE.findall(text)
    if code:
        problems.append(f"{len(code)} line(s) of embedded source code")

    rows = ROW.findall(text)
    if rows:
        unknown = [pkg for pkg, _v, lic in rows if lic.strip().lower() == "unknown"]
        if unknown:
            problems.append(
                f"{len(unknown)} dependency/ies with unresolved license "
                f"(e.g. {unknown[0]})"
            )
        unversioned = [
            pkg for pkg, ver, _l in rows if ver.strip() in ("", "—", "-", "Unknown")
        ]
        if unversioned:
            problems.append(
                f"{len(unversioned)} dependency/ies with no version "
                f"(e.g. {unversioned[0]})"
            )

        # Every row must be covered: a reproduced license text, or an explicit
        # declaration that upstream ships no license file.
        covered = set(APPLIES.findall(text))
        missing = [pkg for pkg, _v, _l in rows if pkg not in covered]
        if missing:
            problems.append(
                f"{len(missing)} dependency/ies with no license text and not "
                f"declared as text-less (e.g. {missing[0]})"
            )

    head = "\n".join(text.split("\n")[:6])
    if "NVIDIA Run:ai" not in head:
        problems.append("missing standard NVIDIA Run:ai header")

    return rows, problems


def main():
    notices = find_notices(sys.argv[1:])
    failed = 0
    for path in notices:
        rows, problems = validate(path)
        label = path[2:] if path.startswith("./") else path
        if problems:
            failed += 1
            print(f"FAIL  {label}")
            for problem in problems:
                print(f"        - {problem}")
        else:
            print(f"PASS  {label} ({len(rows)} dependencies)")
    print(f"\n{len(notices) - failed}/{len(notices)} passed")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
