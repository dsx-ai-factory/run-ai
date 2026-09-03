#!/usr/bin/env python3
"""Compare several run bundles side by side, one row per scale level.

    ./compare-scale-levels.py <bundle-dir>... [--monorepo ~/work/runai] [-o out.md]

Answers "what grows with scale, and how fast" — build-recommendation.py answers "what should
this one level be sized at". Parsing, group membership and margins are imported from it so the
two cannot drift apart.

Levels are ordered by cluster GPU capacity, not by the label, so a mislabelled bundle still
sorts where its numbers put it.
"""
import argparse
import importlib.util
import os
import sys

_spec = importlib.util.spec_from_file_location(
    "build_recommendation", os.path.join(os.path.dirname(os.path.abspath(__file__)),
                                         "build-recommendation.py"))
br = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(br)

GROUPS = ("scheduling", "sync", "workload")
TOP_N = 3


def load(bundle, monorepo):
    """Reduce one bundle to the numbers the comparison needs."""
    vpa = [r for r in br.read_csv(os.path.join(bundle, "vpa", "vpa-recommendations.csv"))
           if not br.is_harness(r["container"])]
    cm = [r for r in br.read_csv(os.path.join(bundle, "vpa", "container-metrics.csv"))
          if not br.is_harness(r["container"])]
    if not vpa:
        sys.exit(f"no vpa/vpa-recommendations.csv under {bundle}")

    meta_path = os.path.join(bundle, "run-metadata.json")
    meta = {}
    if os.path.exists(meta_path):
        import json
        meta = br.derive_scale_params(json.load(open(meta_path)))

    groups, gm_path = br.load_group_mapping(monorepo)
    if not groups:
        sys.exit(f"could not read ServiceGroupMapping from {gm_path}")
    group_of, declared_key, deployed_name = br.build_classifier(groups)
    grouped, ungrouped = br.split_by_group(vpa, group_of)

    # Services VPA never reported fall back to Prometheus, same rule as the per-level report:
    # only when they belong to no group, so a grouped service is never counted twice.
    prom = {r["container"]: r for r in cm if r["namespace"] == "runai"}
    seen = {r["container"] for r in vpa if r["namespace"] == "runai"}
    extra = [r for cont, r in sorted(prom.items())
             if cont not in seen and group_of(cont) is None]

    try:
        cap = (int(meta["np"]) * int(meta["n"]) * int(meta["g"])
               // max(1, int(meta.get("pr", 1))))
    except (KeyError, ValueError, TypeError):
        cap = 0

    return {
        "label": meta.get("scale_level", os.path.basename(bundle)),
        "meta": meta, "cap": cap, "groups": groups, "grouped": grouped,
        "declared_key": declared_key, "deployed_name": deployed_name,
        "ungrouped": ungrouped, "extra": extra,
        "cp": [r for r in vpa if r["namespace"] == "runai-backend"],
        "kube": [r for r in cm if r["namespace"] == "kube-system"],
        "etcd": br.read_csv(os.path.join(bundle, "vpa", "etcd-db-size.csv")),
    }


def rec(row, margin=1.0):
    """The recommended request pair for a row: VPA upper bound, or Prometheus max plus margin."""
    return ((br.parse_cpu(row["cpu_upper_bound"]) or 0) * margin,
            (br.parse_mem(row["memory_upper_bound"]) or 0) * margin)


def totals(sets):
    """Sum recommendations over (rows, margin) pairs, dropping floor values per dimension."""
    cpu = sum(rec(r, m)[0] for rows, m in sets for r in rows if not br.is_cpu_floor(r))
    mem = sum(rec(r, m)[1] for rows, m in sets for r in rows
              if not br.is_mem_floor(r) and br.has_mem(r))
    return cpu, mem


def top_consumers(sets, idx, fmt, floor_test):
    """The TOP_N largest contributors, rendered as `name (value)`."""
    scored = [(rec(r, m)[idx], r["container"]) for rows, m in sets for r in rows
              if not floor_test(r)]
    scored.sort(reverse=True)
    if not scored:
        return "-"
    return ", ".join(f"{name} ({fmt(v)})" for v, name in scored[:TOP_N])


def group_instances(level, grp):
    """Declared members, and how many instances each contributes (scheduler runs per pool)."""
    declared = sorted(br.camel_to_kebab(f) for f, g in level["groups"].items() if g == grp)
    counts = {}
    for r in level["grouped"].get(grp, []):
        k = level["declared_key"](r["container"])
        counts[k] = counts.get(k, 0) + 1
    return declared, sum(max(1, counts.get(d, 0)) for d in declared), counts


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("bundles", nargs="+")
    ap.add_argument("--monorepo", default=os.path.expanduser("~/work/runai"))
    ap.add_argument("-o", "--output")
    args = ap.parse_args()

    levels = sorted((load(b, args.monorepo) for b in args.bundles),
                    key=lambda x: (x["cap"], x["label"]))

    out = []
    w = out.append
    w("# Resource recommendation — scale level comparison\n")
    w("CPU in cores, memory in GB. Every figure is the recommended **request**: VPA's upper "
      "bound, or a Prometheus maximum plus a "
      f"{br.PROM_MARGIN - 1:.0%} margin where VPA produced no recommendation. Values at VPA's "
      "histogram floor are excluded per dimension. See a single-level report for the full "
      "explanation.\n")

    # Parameters as rows and levels as columns: there are many more parameters than levels, and
    # this way each parameter's progression reads left to right.
    w("\n## Run parameters\n")
    w("| Parameter | " + " | ".join(L["label"] for L in levels) + " |")
    w("|---" * (len(levels) + 1) + "|")
    for key, name, _ in br.SCALE_PARAMS:
        if not any(key in L["meta"] for L in levels):
            continue
        vals = " | ".join(f"`{L['meta'][key]}`" if key in L["meta"] else "-" for L in levels)
        w(f"| {name} | {vals} |")

    # --- 1. service groups, one table per group
    w("\n## Service groups\n")
    w("A group's value is applied to every member, so the total is the largest member's "
      "recommendation times the instance count.\n")
    w("**Instance count** is how many pods the group's value is actually applied to, which is "
      "not the same as the number of services: the scheduler runs one instance per node pool "
      "plus one for the default pool, so it contributes more than one.\n")
    w("**Utilized** is the sum of the members' own recommendations — what the group would need "
      "if each service were sized individually — and its percentage of the group total.\n")
    for grp in GROUPS:
        if not any(L["grouped"].get(grp) for L in levels):
            continue
        declared, _, _ = group_instances(levels[-1], grp)
        shown = ", ".join(f"`{levels[-1]['deployed_name'].get(d, d)}`" for d in declared)
        w(f"\n### {grp} ({len(declared)} services)\n")
        w(f"{shown}\n")
        w("| Scale level | Instance count | highest CPU | top CPU consumer | highest Mem (GB) "
          "| top Mem consumer | total CPU | total Mem (GB) | utilized CPU | utilized Mem (GB) |")
        w("|---|---|---|---|---|---|---|---|---|---|")
        for L in levels:
            all_rows = L["grouped"].get(grp, [])
            rows = [r for r in all_rows if not br.is_floor(r)]
            if not rows:
                w(f"| {L['label']} | - | - | - | - | - | - | - | - | - |")
                continue
            _, n, _ = group_instances(L, grp)
            tc = max(rows, key=lambda r: rec(r)[0])
            tm = max(rows, key=lambda r: rec(r)[1])
            mc, mm = rec(tc)[0], rec(tm)[1]
            # What the members would need sized individually. Floor and unobserved members
            # contribute nothing, so this understates utilization slightly.
            uc = sum(rec(r)[0] for r in all_rows if not br.is_cpu_floor(r))
            um = sum(rec(r)[1] for r in all_rows if not br.is_mem_floor(r))
            w(f"| {L['label']} | {n} | {br.c(mc)} | {tc['container']} | {br.g(mm)} "
              f"| {tm['container']} | {br.c(mc * n)} | {br.g(mm * n)} "
              f"| {br.c(uc)} ({uc / (mc * n) * 100:.0f}%) "
              f"| {br.g(um)} ({um / (mm * n) * 100:.0f}%) |")

    # --- 2. ungrouped cluster services
    ungrouped_sets = lambda L: [(L["ungrouped"], 1.0), (L["extra"], br.PROM_MARGIN)]
    w("\n## Run:ai cluster services not in any group\n")
    w("Each is configured individually, so these are plain sums of every service's own "
      "recommendation.\n")
    w(members_line(levels, ungrouped_sets))
    section_table(w, levels, ungrouped_sets)

    # --- 3. control plane
    w("\n## Control plane services\n")
    w("Configured per service in the control-plane Helm chart; there is no group mechanism.\n")
    w(members_line(levels, lambda L: [(L["cp"], 1.0)]))
    section_table(w, levels, lambda L: [(L["cp"], 1.0)])

    # --- 4. kube-system
    w("\n## kube-system services\n")
    w("Not part of Run:ai, and shown for infrastructure sizing rather than configuration. "
      "**These are maximum measured values with no VPA involved** — no recommendation model, no "
      "safety margin, no confidence widening. Test-harness containers are excluded.\n")
    w(members_line(levels, lambda L: [(L["kube"], 1.0)]))
    w("| Scale level | total CPU | total Mem (GB) | top CPU consumers | top Mem consumers |")
    w("|---|---|---|---|---|")
    for L in levels:
        sets = [(L["kube"], 1.0)]
        cpu = sum(rec(r)[0] for r in L["kube"])
        mem = sum(rec(r)[1] for r in L["kube"])
        w(f"| {L['label']} | {br.c(cpu)} | {br.g(mem)} "
          f"| {top_consumers(sets, 0, br.c, lambda r: False)} "
          f"| {top_consumers(sets, 1, br.g, lambda r: False)} |")

    # --- 5. grand total
    w("\n## Grand total\n")
    w("Service groups counted as the group value times instance count; every other section as "
      "the sum of its services.\n")
    w("| Scale level | groups CPU | groups Mem (GB) | ungrouped CPU | ungrouped Mem (GB) "
      "| control plane CPU | control plane Mem (GB) | kube-system CPU | kube-system Mem (GB) "
      "| total CPU | total Mem (GB) |")
    w("|---|---|---|---|---|---|---|---|---|---|---|")
    for L in levels:
        gc = gm = 0.0
        for grp in GROUPS:
            rows = [r for r in L["grouped"].get(grp, []) if not br.is_floor(r)]
            if not rows:
                continue
            _, n, _ = group_instances(L, grp)
            gc += max(rec(r)[0] for r in rows) * n
            gm += max(rec(r)[1] for r in rows) * n
        uc, um = totals([(L["ungrouped"], 1.0), (L["extra"], br.PROM_MARGIN)])
        pc, pm = totals([(L["cp"], 1.0)])
        kc = sum(rec(r)[0] for r in L["kube"])
        km = sum(rec(r)[1] for r in L["kube"])
        w(f"| {L['label']} | {br.c(gc)} | {br.g(gm)} | {br.c(uc)} | {br.g(um)} "
          f"| {br.c(pc)} | {br.g(pm)} | {br.c(kc)} | {br.g(km)} "
          f"| {br.c(gc + uc + pc + kc)} | {br.g(gm + um + pm + km)} |")

    # --- 6. etcd
    w("\n## etcd database size\n")
    w("Total size on disk. etcd does not shrink without a defrag, so the maximum is the "
      "end-of-run size rather than a peak.\n")
    w("| Scale level | min (GB) | p95 (GB) | max (GB) |")
    w("|---|---|---|---|")
    for L in levels:
        row = next((r for r in L["etcd"] if "total" in r.get("metric", "")), None)
        if not row:
            w(f"| {L['label']} | - | - | - |")
            continue
        w(f"| {L['label']} | {br.g(float(row['min_bytes']))} | {br.g(float(row['p95_bytes']))} "
          f"| {br.g(float(row['max_bytes']))} |")

    text = "\n".join(out) + "\n"
    if args.output:
        open(args.output, "w").write(text)
        print(f"wrote {args.output}", file=sys.stderr)
    else:
        print(text)


def members_line(levels, sets_for):
    """The services a section covers, unioned across levels so one level missing a service
    does not drop it from the list."""
    names = sorted({br.service_label(r) for L in levels for rows, _ in sets_for(L) for r in rows})
    if not names:
        return "_No services in this section._\n"
    return f"{len(names)} services: " + ", ".join(f"`{n}`" for n in names) + "\n"


def section_table(w, levels, sets_for):
    """Per-level sum plus its largest contributors, used by the two per-service sections."""
    w(f"| Scale level | total CPU | total Mem (GB) | top {TOP_N} CPU consumers "
      f"| top {TOP_N} Mem consumers |")
    w("|---|---|---|---|---|")
    for L in levels:
        sets = sets_for(L)
        cpu, mem = totals(sets)
        w(f"| {L['label']} | {br.c(cpu)} | {br.g(mem)} "
          f"| {top_consumers(sets, 0, br.c, br.is_cpu_floor)} "
          f"| {top_consumers(sets, 1, br.g, br.is_mem_floor)} |")


if __name__ == "__main__":
    main()
