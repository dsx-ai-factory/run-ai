#!/usr/bin/env python3
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
"""Generate a component NOTICE file from its resolved dependency graph.

Usage: generate-notice.py <component-dir> [<component-dir> ...]

For each Go component this runs `go-licenses report` to map every third-party
package to its license, `go list -m all` to resolve exact module versions, and
`go-licenses save` to collect the upstream license texts verbatim. First-party
github.com/dsx-ai-factory/run-ai/... modules are omitted: they live in this
repository and are covered by the root LICENSE.
"""

import argparse
import csv
import hashlib
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
from collections import Counter, defaultdict

FIRST_PARTY = "github.com/dsx-ai-factory/run-ai/"
REPO_ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

SPDX_HEADER = (
    "# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. "
    "All rights reserved.\n"
    "# SPDX-License-Identifier: Apache-2.0\n\n"
)

HEADER = SPDX_HEADER + """NVIDIA Run:ai — {component}
Copyright (c) 2026 NVIDIA Corporation & Affiliates. All rights reserved.

This product includes software developed at NVIDIA Corporation & Affiliates,
licensed under the Apache License, Version 2.0. See the LICENSE file at the
root of this repository for the full license text.

This file inventories the third-party software distributed with this component
and the license under which each dependency is made available. It is generated
from the component's resolved dependency graph by
build/scripts/generate-notice.py — do not edit by hand.
"""


def go_licenses_bin():
    explicit = os.environ.get("GO_LICENSES")
    if explicit:
        return explicit
    found = shutil.which("go-licenses")
    if found:
        return found
    gopath = subprocess.run(
        ["go", "env", "GOPATH"], capture_output=True, text=True, check=True
    ).stdout.strip()
    candidate = os.path.join(gopath, "bin", "go-licenses")
    if not os.path.exists(candidate):
        sys.exit(
            "go-licenses not found. Install it with:\n"
            "  go install github.com/google/go-licenses@latest"
        )
    return candidate


def module_versions(component_dir):
    """Map module path -> version for everything in the component's build list."""
    out = subprocess.run(
        ["go", "list", "-m", "-e", "-f", "{{.Path}} {{.Version}}", "all"],
        cwd=component_dir,
        capture_output=True,
        text=True,
    )
    versions = {}
    for line in out.stdout.splitlines():
        parts = line.split()
        if len(parts) == 2:
            versions[parts[0]] = parts[1]
    return versions


def resolve_version(pkg, versions):
    """Longest module path that prefixes this package path wins."""
    best = ""
    for mod in versions:
        if (pkg == mod or pkg.startswith(mod + "/")) and len(mod) > len(best):
            best = mod
    return versions.get(best, "") if best else ""


UPSTREAM_PLACEHOLDER_NOTE = (
    "[Reproduced verbatim from upstream. This file itself contains unfilled license\n"
    "template fields (for example <COPYRIGHT HOLDER>). They are left exactly as the\n"
    "copyright holder published them: a third-party license text must not be edited.]"
)
PLACEHOLDER_TOKENS = (
    "<YEAR>",
    "<COPYRIGHT HOLDER>",
    "[yyyy]",
    "[name of copyright owner]",
    "<year>",
    "<owner>",
    "<OWNER>",
    "<name of author>",
)

APPENDIX_MARKER = "APPENDIX: How to apply the Apache License to your work"
APPENDIX_CUT = "END OF TERMS AND CONDITIONS"
APPENDIX_NOTE = (
    "END OF TERMS AND CONDITIONS\n\n"
    "[The upstream file continues with the Apache-2.0 appendix, a fill-in-the-blank\n"
    "template for applying the license to a new work. It is expressly not part of\n"
    "the license terms, and this copy of it was never filled in, so it is omitted\n"
    "here rather than reproduced with unresolved placeholders.]"
)


def strip_unfilled_appendix(text):
    """Drop the Apache-2.0 'how to apply' appendix when it was never filled in.

    Many projects complete the appendix with their real copyright line; those are
    left untouched. Only a copy still carrying the literal [yyyy] template is cut,
    since that is what reads as an unresolved placeholder in the NOTICE.
    """
    if APPENDIX_MARKER not in text or "[yyyy]" not in text:
        return text
    head, _sep, _tail = text.partition(APPENDIX_CUT)
    return head + APPENDIX_NOTE


def module_dirs(component_dir):
    """Map module path -> (version, local source dir) from the module cache."""
    out = subprocess.run(
        ["go", "list", "-m", "-e", "-f", "{{.Path}}|{{.Version}}|{{.Dir}}", "all"],
        cwd=component_dir,
        capture_output=True,
        text=True,
    )
    dirs = {}
    for line in out.stdout.splitlines():
        parts = line.split("|")
        if len(parts) == 3 and parts[2]:
            dirs[parts[0]] = (parts[1], parts[2])
    return dirs


def module_license_fallback(pkg, dirs):
    """Resolve a package go-licenses could not classify from its module root.

    go-licenses reports Unknown when a subpackage has no license file of its own
    and the detector cannot name the module-root license (MIT-0, for one). The
    module root's LICENSE is still the governing one, so read it directly.
    """
    best = ""
    for mod in dirs:
        if (pkg == mod or pkg.startswith(mod + "/")) and len(mod) > len(best):
            best = mod
    if not best:
        return None
    version, path = dirs[best]
    for name in ("LICENSE", "LICENSE.txt", "LICENSE.md", "COPYING", "LICENCE"):
        candidate = os.path.join(path, name)
        if os.path.exists(candidate):
            text = open(candidate, encoding="utf-8", errors="replace").read()
            title = classify_license(text)
            url = f"https://pkg.go.dev/{best}@{version}?tab=licenses"
            return title, url, text
    return None


def classify_license(text):
    """Name a license from its text using distinctive, non-overlapping phrases.

    Deliberately conservative: anything not confidently matched comes back as
    "Unknown" so the validator forces a human to look, rather than guessing a
    name onto a license an OSS review board will rely on.
    """
    flat = " ".join(text.split())
    if "Apache License" in flat and "Version 2.0" in flat:
        return "Apache-2.0"
    # Upstream files vary: "Mozilla Public License Version 2.0" and
    # "Mozilla Public License, version 2.0" both occur. Must be tested before
    # any GPL check: the MPL text names the GPL as a compatible secondary
    # license, which a naive substring match reads as a GPL license.
    if re.search(r"Mozilla Public License,? version 2\.0", flat, re.I):
        return "MPL-2.0"
    if "Permission is hereby granted, free of charge" in flat:
        if "The above copyright notice and this permission notice" in flat:
            return "MIT"
        return "MIT-0"
    if "Permission to use, copy, modify, and/or distribute this software" in flat:
        return "ISC"
    if "Redistribution and use in source and binary forms" in flat:
        if "Neither the name of" in flat or "nor the names of its" in flat:
            return "BSD-3-Clause"
        return "BSD-2-Clause"
    # Require the FSF preamble, not a passing mention of the GPL by name.
    if (
        "GNU GENERAL PUBLIC LICENSE" in flat.upper()
        and "Everyone is permitted to copy and distribute verbatim copies" in flat
    ):
        return "GPL (review required)"
    if "CC0 1.0 Universal" in flat:
        return "CC0-1.0"
    return "Unknown"


def has_requirements(component_dir):
    """True if go.mod declares any dependency at all."""
    gomod = os.path.join(component_dir, "go.mod")
    if not os.path.exists(gomod):
        return False
    return "require" in open(gomod, encoding="utf-8", errors="replace").read()


def collect_go_modules(component_dir):
    """Inventory third-party Go modules straight from the build graph.

    Used when go-licenses cannot run: it exits fatally if any dependency lacks a
    discoverable license file, which the repo's own first-party modules do. This
    reads `go list -deps` instead, so it reports at module rather than package
    granularity, and reads each license from the module cache.
    """
    proc = subprocess.run(
        ["go", "list", "-deps", "-e", "-json", "./..."],
        cwd=component_dir,
        capture_output=True,
        text=True,
    )
    # The component's own module has no version and is not a dependency of
    # itself; forks vendored under an upstream module path (goofys) would
    # otherwise appear as their own third-party dependency.
    main_module = subprocess.run(
        ["go", "list", "-m"], cwd=component_dir, capture_output=True, text=True
    ).stdout.strip().split("\n")[0]

    decoder = json.JSONDecoder()
    payload = proc.stdout
    mods = {}
    index = 0
    while index < len(payload):
        while index < len(payload) and payload[index] in " \n\t\r":
            index += 1
        if index >= len(payload):
            break
        try:
            obj, index = decoder.raw_decode(payload, index)
        except json.JSONDecodeError:
            break
        module = obj.get("Module")
        if not module or not module.get("Path"):
            continue
        path = module["Path"]
        if path.startswith(FIRST_PARTY) or path == FIRST_PARTY.rstrip("/"):
            continue
        if main_module and path == main_module:
            continue
        mods[path] = (module.get("Version") or "", module.get("Dir") or "")

    rows, versions, texts = [], {}, {}
    for path, (version, directory) in sorted(mods.items()):
        text = None
        if directory:
            for name in LICENSE_FILENAMES:
                candidate = os.path.join(directory, name)
                if os.path.exists(candidate):
                    text = open(candidate, encoding="utf-8", errors="replace").read()
                    break
        license_name = classify_license(text) if text else "Unknown"
        url = f"https://pkg.go.dev/{path}@{version}?tab=licenses" if version else path
        rows.append((path, url, license_name))
        versions[path] = version
        if text:
            texts[path] = text
    return rows, versions, texts


def run_report(bin_path, component_dir):
    proc = subprocess.run(
        [bin_path, "report", "./..."],
        cwd=component_dir,
        capture_output=True,
        text=True,
    )
    rows = []
    for row in csv.reader(proc.stdout.splitlines()):
        if len(row) != 3:
            continue
        pkg, url, license_name = row
        if pkg.startswith(FIRST_PARTY):
            continue
        rows.append((pkg, url, license_name))
    return sorted(set(rows))


def license_file_rank(name):
    """Rank a saved filename by how likely it is to be THE license text.

    Lower sorts first; None means the file is not a license file at all.
    """
    lowered = name.lower()
    for rank, candidate in enumerate(
        ("license", "licence", "copying", "notice", "license-mit", "license-apache")
    ):
        if lowered == candidate or lowered.startswith(candidate + "."):
            return rank
    if lowered.startswith(("license", "licence", "copying")):
        return 90
    return None


def run_save(bin_path, component_dir):
    """Collect upstream license texts; returns package path -> license text.

    `go-licenses save` copies the full source tree for reciprocal licenses such
    as MPL-2.0, to satisfy their source-redistribution terms. Only the license
    file itself belongs in a NOTICE, so non-license files are ignored here; for
    those dependencies the NOTICE points at upstream for the source.
    """
    tmp = tempfile.mkdtemp(prefix="notice-save-")
    subprocess.run(
        [bin_path, "save", "./...", "--save_path", tmp, "--force"],
        cwd=component_dir,
        capture_output=True,
        text=True,
    )
    candidates = defaultdict(list)
    for root, _dirs, files in os.walk(tmp):
        pkg = os.path.relpath(root, tmp).replace(os.sep, "/")
        if pkg.startswith(FIRST_PARTY):
            continue
        for name in files:
            rank = license_file_rank(name)
            if rank is not None:
                candidates[pkg].append((rank, os.path.join(root, name)))

    texts = {}
    for pkg, entries in candidates.items():
        for _rank, path in sorted(entries):
            try:
                texts[pkg] = open(path, encoding="utf-8", errors="replace").read()
                break
            except OSError:
                continue
    shutil.rmtree(tmp, ignore_errors=True)
    return texts


LICENSE_FILENAMES = (
    "LICENSE",
    "LICENSE.txt",
    "LICENSE.md",
    "LICENCE",
    "LICENSE-MIT",
    "LICENSE-APACHE",
    "COPYING",
)


def collect_npm(component_dir):
    """Inventory the production npm dependency tree from package-lock.json.

    The lockfile is authoritative about which packages are dev-only: `npm ls
    --omit=dev` still emits dev packages nested under production ones, which
    would attribute software that is never redistributed. devDependencies are
    build and test tooling and do not belong in a NOTICE.
    """
    lockfile = os.path.join(component_dir, "package-lock.json")
    if not os.path.exists(lockfile):
        return [], {}, {}
    try:
        lock = json.load(open(lockfile, encoding="utf-8"))
    except (json.JSONDecodeError, OSError):
        return [], {}, {}

    rows, versions, texts = [], {}, {}
    for key, entry in sorted(lock.get("packages", {}).items()):
        if not key.startswith("node_modules/"):
            continue
        if entry.get("dev") or entry.get("devOptional"):
            continue
        name = key.split("node_modules/")[-1]
        version = entry.get("version", "")
        if not version:
            continue

        license_name = entry.get("license") or "Unknown"
        if isinstance(license_name, list):
            license_name = " OR ".join(str(item) for item in license_name)

        path = os.path.join(component_dir, key)
        if license_name == "Unknown":
            manifest = os.path.join(path, "package.json")
            if os.path.exists(manifest):
                try:
                    meta = json.load(open(manifest, encoding="utf-8"))
                    declared = meta.get("license") or meta.get("licenses")
                    if isinstance(declared, list):
                        declared = " OR ".join(
                            d.get("type", "") for d in declared if isinstance(d, dict)
                        )
                    elif isinstance(declared, dict):
                        declared = declared.get("type", "")
                    license_name = declared or "Unknown"
                except (json.JSONDecodeError, OSError):
                    pass

        for filename in LICENSE_FILENAMES:
            candidate = os.path.join(path, filename)
            if os.path.exists(candidate):
                texts[name] = open(
                    candidate, encoding="utf-8", errors="replace"
                ).read()
                break

        rows.append((name, f"https://www.npmjs.com/package/{name}/v/{version}", license_name))
        versions[name] = version
    return rows, versions, texts


def read_preamble(component_dir):
    """Optional NOTICE.preamble, preserved across regeneration.

    For components whose own code is not NVIDIA-authored (a vendored fork, for
    one), the generated header alone would misstate provenance.
    """
    path = os.path.join(component_dir, "NOTICE.preamble")
    return open(path, encoding="utf-8").read() if os.path.exists(path) else None


def build_notice(component, rows, versions, texts, preamble=None):
    counts = Counter(license_name for _pkg, _url, license_name in rows)
    lines = [HEADER.format(component=component)]
    if preamble:
        lines.append(preamble.rstrip())
        lines.append("")

    lines.append("=" * 80)
    lines.append("Summary")
    lines.append("=" * 80)
    lines.append("")
    lines.append(f"{len(rows)} third-party packages:")
    lines.append("")
    for name, count in sorted(counts.items(), key=lambda kv: (-kv[1], kv[0])):
        lines.append(f"  - {name}: {count}")
    lines.append("")

    lines.append("=" * 80)
    lines.append("Dependencies")
    lines.append("=" * 80)
    lines.append("")
    lines.append("| Package | Version | License | License source |")
    lines.append("|---------|---------|---------|----------------|")
    for pkg, url, license_name in rows:
        version = resolve_version(pkg, versions) or "—"
        lines.append(f"| `{pkg}` | {version} | {license_name} | {url} |")
    lines.append("")

    # One copy of each distinct license text, listing every package it covers.
    by_text = defaultdict(list)
    for pkg, _url, _license_name in rows:
        text = texts.get(pkg)
        if text:
            by_text[hashlib.sha256(text.encode()).hexdigest()].append((pkg, text))

    no_text = sorted(pkg for pkg, _url, _name in rows if not texts.get(pkg))
    if no_text:
        lines.append("=" * 80)
        lines.append("Declared license, no license file distributed upstream")
        lines.append("=" * 80)
        lines.append("")
        lines.append(
            "These packages declare a license in their published metadata but ship no\n"
            "license file in their distributed artifact. The declared license in the\n"
            "table above governs; no text is reproduced here because upstream provides\n"
            "none to reproduce."
        )
        lines.append("")
        for pkg in no_text:
            lines.append(f"  - {pkg}")
        lines.append("")

    lines.append("=" * 80)
    lines.append("License texts")
    lines.append("=" * 80)
    lines.append("")
    lines.append(
        "The full text of each distinct dependency license follows, reproduced "
        "verbatim from the\nupstream project together with its copyright notice."
    )
    lines.append("")
    for digest in sorted(by_text, key=lambda d: by_text[d][0][0]):
        entries = by_text[digest]
        text = entries[0][1]
        lines.append("-" * 80)
        lines.append("Applies to:")
        for pkg, _ in sorted(entries):
            lines.append(f"  - {pkg}")
        if any(token in text for token in PLACEHOLDER_TOKENS):
            lines.append("")
            lines.append(UPSTREAM_PLACEHOLDER_NOTE)
        lines.append("-" * 80)
        lines.append("")
        lines.append(text.rstrip())
        lines.append("")
    return "\n".join(lines) + "\n"


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("components", nargs="+")
    args = parser.parse_args()

    bin_path = go_licenses_bin()
    for component_dir in args.components:
        abs_dir = os.path.abspath(component_dir)
        component = os.path.relpath(abs_dir, REPO_ROOT)
        is_go = os.path.exists(os.path.join(abs_dir, "go.mod"))
        is_npm = os.path.exists(os.path.join(abs_dir, "package.json"))
        if not is_go and not is_npm:
            print(f"SKIP  {component} (no go.mod or package.json)", flush=True)
            continue
        print(f"---   {component}", flush=True)

        if is_npm and not is_go:
            rows, versions, texts = collect_npm(abs_dir)
            if not rows:
                print(f"FAIL  {component}: npm ls returned no packages", flush=True)
                continue
            texts = {pkg: strip_unfilled_appendix(t) for pkg, t in texts.items()}
            notice = build_notice(
                component, rows, versions, texts, read_preamble(abs_dir)
            )
            with open(os.path.join(abs_dir, "NOTICE"), "w", encoding="utf-8") as handle:
                handle.write(notice)
            missing = sum(1 for pkg, _u, _l in rows if pkg not in texts)
            print(
                f"OK    {component}: {len(rows)} packages, "
                f"{len(set(l for _p, _u, l in rows))} licenses, "
                f"{missing} without text",
                flush=True,
            )
            continue

        rows = run_report(bin_path, abs_dir)
        # A fork published under an upstream module path (goofys) is the main
        # module, not a dependency of itself.
        main_module = subprocess.run(
            ["go", "list", "-m"], cwd=abs_dir, capture_output=True, text=True
        ).stdout.strip().split("\n")[0]
        if main_module:
            rows = [
                r
                for r in rows
                if r[0] != main_module and not r[0].startswith(main_module + "/")
            ]
        if rows:
            versions = module_versions(abs_dir)
            texts = run_save(bin_path, abs_dir)
        else:
            print(f"      {component}: go-licenses unusable, reading build graph")
            rows, versions, texts = collect_go_modules(abs_dir)
            if not rows and has_requirements(abs_dir):
                print(f"FAIL  {component}: no third-party modules resolved", flush=True)
                continue
            if not rows:
                print(f"      {component}: module declares no dependencies")

        dirs = module_dirs(abs_dir)
        resolved = []
        for pkg, url, license_name in rows:
            if license_name == "Unknown":
                fallback = module_license_fallback(pkg, dirs)
                if fallback:
                    license_name, url, text = fallback
                    texts.setdefault(pkg, text)
            resolved.append((pkg, url, license_name))
        rows = resolved
        texts = {pkg: strip_unfilled_appendix(text) for pkg, text in texts.items()}
        unversioned = [p for p, _u, _l in rows if not resolve_version(p, versions)]
        if unversioned:
            print(
                f"FAIL  {component}: {len(unversioned)} package(s) with no resolved "
                f"version (e.g. {unversioned[0]}) - NOTICE not written",
                flush=True,
            )
            continue

        notice = build_notice(component, rows, versions, texts, read_preamble(abs_dir))
        with open(os.path.join(abs_dir, "NOTICE"), "w", encoding="utf-8") as handle:
            handle.write(notice)
        missing = sum(1 for pkg, _u, _l in rows if pkg not in texts)
        print(
            f"OK    {component}: {len(rows)} packages, "
            f"{len(set(l for _p, _u, l in rows))} licenses, {missing} without text",
            flush=True,
        )


if __name__ == "__main__":
    main()
