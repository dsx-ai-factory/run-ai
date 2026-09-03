#!/usr/bin/env python3
"""Turn a run's result bundle into the resource-recommendation data file.

Reads the artifacts produced by resource-rec-run.yml and joins them with the product's
declared defaults, so a scale level can be written up without hand-copying numbers.

    ./build-recommendation.py <bundle-dir> [--monorepo ~/work/runai] [-o out.md]

<bundle-dir> is the extracted artifact directory containing vpa/ and run-metadata.json.

Measured values are copied from the source data and only unit-converted: CPU in cores,
memory in GB. Group sums are max x service-count, because a group's value is applied to
every service in it.

The recommendation is the VPA upper bound, read as a **request**. See "How to read the
recommendation" in the generated output for why, and why no limit is recommended.
"""
import argparse
import csv
import json
import os
import re
import sys

# VPA emits these when a container never accumulated a real sample: the first histogram
# bucket times the 1.15 margin. Instrument floors, not measurements.
CPU_FLOOR = "11m"
MEM_FLOOR = "11500k"

MEM_UNITS = {"": 1, "k": 1e3, "M": 1e6, "G": 1e9, "T": 1e12,
             "Ki": 1024, "Mi": 1024**2, "Gi": 1024**3, "Ti": 1024**4}

# The bar for flagging a service that has no declared default to compare to.
SIGNIFICANT_CPU = 0.5     # cores
SIGNIFICANT_MEM = 0.5e9   # bytes

MARK = "⚠️"

# Prometheus rows are a raw observed maximum: no safety margin and no confidence widening,
# unlike VPA's output. This stands in for both so the two sources are comparable.
PROM_MARGIN = 1.20
PROM_MARK = "*"

# Test scaffolding we install ourselves — never part of a customer deployment.
HARNESS_CONTAINERS = {"socat", "kwok", "kube-rbac-proxy", "manager-kwok"}
HARNESS_SUBSTRINGS = ("kwok", "etcd-metrics-proxy", "vpa-")

CLUSTER_ALIASES = {
    "runai-scheduler-nodepool-1": "runai-scheduler",
    "runai-scheduler-default": "runai-scheduler",
    "runai-agent": "agent",
    "podgroup-controller": "pod-group-controller",
    "runai-job-controller": "job-controller",
    "runai-admission-controller": "admission-controller",
    "runai-project-controller": "project-controller",
    "runai-node-exporter": "node-exporter",
    # camel_to_kebab("NodePoolController") yields "node-pool-controller", but the deployed
    # container is "nodepool-controller". Without this the service falls out of its group.
    "nodepool-controller": "node-pool-controller",
}

# Grouped by what the parameter controls: the system being measured, the load applied to it,
# then how the run itself was shaped. Bookkeeping that only identifies the run comes last.
SCALE_PARAMS = [
    # System scale
    ("np", "Node pool count", "Run:ai node pools created"),
    ("n", "Nodes per pool", "Simulated (KWOK) nodes in each pool"),
    ("g", "GPUs per node", "Simulated GPUs on each node"),
    ("w_cap", "Cluster GPU capacity (W_cap)", "Concurrent workloads the cluster can run"),
    # Load shape
    ("mode", "Submission mode", "rate = sustained arrival; burst = all at once"),
    ("w_total", "Total workloads (W_total)", "Capacity plus pending"),
    ("w_pending", "Pending workloads", "Workloads queued beyond cluster GPU capacity"),
    ("pr", "Pods per workload", "1 = Training workload; >1 = Distributed"),
    ("r", "Submission rate (per minute)", "Steady-state arrival rate"),
    ("workload_duration_seconds", "Workload duration (seconds)",
     "Derived as capacity/rate; how long each workload runs"),
    ("max_concurrent", "Max concurrent submissions", "Cap on in-flight submission calls"),
    ("api_exerciser_users", "API exerciser users", "Concurrent read-only API clients"),
    ("api_exerciser_think_time", "API exerciser think time (s)", "Pause between API calls"),
    # Run shaping
    ("auto_deletion_seconds", "Auto-deletion TTL (seconds)",
     "Delay before a completed workload is removed; 0 = never"),
    ("expected_resident_workloads", "Expected resident workloads",
     "Steady-state population = rate/60 x (duration + TTL)"),
    ("steady_state_duration", "Steady-state duration (minutes)", "Time held at the steady rate"),
    # Identifies the run rather than describing it
    ("scale_level", "Scale level", "Label for this measurement point"),
    ("cluster_name", "Cluster name", "Environment the run executed against"),
    ("run_id", "Run ID", "GitHub Actions run that produced the bundle"),
    ("run_start", "Run start (UTC)", "First workload submitted"),
    ("run_end", "Run end (UTC)", "Submitter exited"),
]


def duration_minutes(s):
    """Go duration string to a bare number of minutes, so the unit can live in the label."""
    m = re.match(r"^([0-9.]+)\s*([a-z]*)$", str(s or "").strip())
    if not m:
        return s
    n, unit = float(m.group(1)), m.group(2)
    factor = {"h": 60, "m": 1, "s": 1 / 60, "": 1}.get(unit)
    return s if factor is None else f"{n * factor:g}"


def derive_scale_params(meta):
    """Add the capacity figures implied by the topology, and normalise units that the
    parameter labels carry instead."""
    if "steady_state_duration" in meta:
        meta["steady_state_duration"] = duration_minutes(meta["steady_state_duration"])
    try:
        cap = int(meta["np"]) * int(meta["n"]) * int(meta["g"]) // max(1, int(meta.get("pr", 1)))
    except (KeyError, ValueError, TypeError):
        return meta
    meta["w_cap"] = cap
    meta["w_total"] = cap + int(meta.get("w_pending", 0))
    return meta


def parse_cpu(s):
    # Helm values may carry bare numbers (cpu: 1), so coerce before parsing.
    s = str(s if s is not None else "").strip()
    if not s:
        return None
    try:
        return float(s[:-1]) / 1000 if s.endswith("m") else float(s)
    except ValueError:
        return None


def parse_mem(s):
    s = str(s if s is not None else "").strip()
    if not s:
        return None
    m = re.match(r"^([0-9.]+)([A-Za-z]*)$", s)
    if not m:
        return None
    return float(m.group(1)) * MEM_UNITS.get(m.group(2), 1)


def num(s):
    """Parse a CSV number, tolerating the blanks a truncated file leaves behind."""
    try:
        return float(s)
    except (TypeError, ValueError):
        return None


def c(v):
    """CPU in cores."""
    return "-" if v is None else f"{v:.3f}"


def g(v):
    """Memory in GB."""
    return "-" if v is None else f"{v / 1e9:.2f}"


# Floor is per dimension: a service can idle on CPU while using real memory, and reporting
# only the both-at-floor case hides that its CPU number is an instrument reading.
def is_cpu_floor(row):
    return row.get("cpu_target") == CPU_FLOOR


def is_mem_floor(row):
    return row.get("memory_target") == MEM_FLOOR


# A blank memory column means the extractor found no memory series for that container.
# Counted with the floors so it is excluded from totals rather than summed as zero.
def has_mem(row):
    return parse_mem(row.get("memory_upper_bound")) is not None


def is_floor(row):
    return is_cpu_floor(row) and is_mem_floor(row)


def is_harness(name):
    n = name.lower()
    return n in HARNESS_CONTAINERS or any(s in n for s in HARNESS_SUBSTRINGS)


def read_csv(path):
    return list(csv.DictReader(open(path))) if os.path.exists(path) else []


def load_group_mapping(monorepo):
    p = os.path.join(monorepo, "runai-cluster/cluster/sdk/apis/run/v1/runaiconfig_types.go")
    if not os.path.exists(p):
        return {}, p
    src = open(p).read()
    return {f: grp.lower() for f, grp in
            re.findall(r'"(\w+)":\s+(\w+)ServicesGroup', src)}, p


def load_cluster_defaults(monorepo):
    base = os.path.join(monorepo, "runai-cluster/cluster/sdk/apis/run/v1")
    out = {}
    if not os.path.isdir(base):
        return out
    for d in sorted(os.listdir(base)):
        p = os.path.join(base, d, "types.go")
        if not os.path.exists(p):
            continue
        m = re.search(r"func \(r \*Resources\) SetDefaultsWhereNeeded\(\)\s*\{(.*?)\n\}",
                      open(p).read(), re.S)
        if not m:
            continue
        body = m.group(1)

        def grab(kind, res):
            mm = re.search(r"r\." + kind + r"\[v1\.Resource" + res +
                           r"\]\s*=\s*resource\.MustParse\(\"([^\"]+)\"\)", body)
            return mm.group(1) if mm else ""
        out[d] = {"req_cpu": grab("Requests", "CPU"), "req_mem": grab("Requests", "Memory"),
                  "lim_cpu": grab("Limits", "CPU"), "lim_mem": grab("Limits", "Memory")}
    return out


def camel_to_kebab(s):
    return re.sub(r"(?<!^)(?=[A-Z])", "-", s).lower()


def load_cp_defaults(monorepo):
    p = os.path.join(monorepo, "control-plane/control-plane-chart/values.yaml")
    if not os.path.exists(p):
        return {}
    try:
        import yaml
    except ImportError:
        # Silently returning {} would drop every control-plane default column and the report
        # would look as though nothing is declared for them.
        print("WARNING: PyYAML not installed — control-plane defaults will be blank",
              file=sys.stderr)
        return {}
    out = {}
    for key, val in (yaml.safe_load(open(p)) or {}).items():
        if not isinstance(val, dict) or not isinstance(val.get("resources"), dict):
            continue
        r = val["resources"]
        req, lim = r.get("requests") or {}, r.get("limits") or {}
        out[camel_to_kebab(key)] = {
            "req_cpu": req.get("cpu", ""), "req_mem": req.get("memory", ""),
            "lim_cpu": lim.get("cpu", ""), "lim_mem": lim.get("memory", ""),
        }
    return out


def build_classifier(groups):
    """Resolve deployed container names against the declared ServiceGroupMapping.

    Returns (group_of, declared_key, deployed_name). Shared with the scale-level comparison
    script so both read group membership the same way.
    """
    gbc = {camel_to_kebab(f): grp for f, grp in groups.items()}
    gbc["runai-scheduler-nodepool-1"] = gbc.get("run-a-i-scheduler", "scheduling")
    gbc["runai-scheduler-default"] = gbc["runai-scheduler-nodepool-1"]
    for cont, pkg in CLUSTER_ALIASES.items():
        if pkg in gbc:
            gbc.setdefault(cont, gbc[pkg])

    # Deployed container name -> the field name the mapping declares it under. The two differ
    # for aliased services, and for the scheduler, which runs one instance per node pool.
    def declared_key(container):
        if container.startswith("runai-scheduler"):
            return "run-a-i-scheduler"
        return CLUSTER_ALIASES.get(container, container)

    # The declared field name is not always what is deployed ("RunAIScheduler" kebabs to
    # "run-a-i-scheduler"), so show the name that appears in the cluster.
    deployed_name = {v: k for k, v in CLUSTER_ALIASES.items()}
    deployed_name["run-a-i-scheduler"] = "runai-scheduler"

    def group_of(container):
        if container in gbc:
            return gbc[container]
        alias = CLUSTER_ALIASES.get(container)
        return gbc.get(alias) if alias else None

    return group_of, declared_key, deployed_name


def split_by_group(vpa_rows, group_of):
    """Partition the runai-namespace rows into {group: rows} and the ungrouped remainder."""
    grouped, ungrouped = {}, []
    for r in vpa_rows:
        if r["namespace"] != "runai":
            continue
        grp = group_of(r["container"])
        if grp:
            grouped.setdefault(grp, []).append(r)
        else:
            ungrouped.append(r)
    return grouped, ungrouped


EMPTY = {"req_cpu": "", "req_mem": "", "lim_cpu": "", "lim_mem": ""}


def defaults_for(container, table, aliases=None):
    return table.get((aliases or {}).get(container, container), EMPTY)


def cells(default, measured, fmt, threshold):
    """Render the (default, measured) pair, marking whichever cell carries the problem.

    A default below the measurement is marked on the default. Where no default exists
    there is nothing to mark, so a significant consumer is marked on the measurement.
    """
    if default is None:
        if measured is not None and measured > threshold:
            return "-", f"**{fmt(measured)}** {MARK}"
        return "-", fmt(measured)
    if measured is not None and default < measured:
        return f"**{fmt(default)}** {MARK}", fmt(measured)
    return fmt(default), fmt(measured)


def limit_cell(default, measured, fmt):
    """The shipped limit, marked when the recommendation would exceed it."""
    if default is None:
        return "-"
    if measured is not None and default < measured:
        return f"**{fmt(default)}** {MARK}"
    return fmt(default)


def service_label(row):
    """Container name, qualified by its deployment only where that adds something.

    Several containers carry a generic name — kubebuilder defaults every controller's to
    `manager`, and Thanos ships `receive` and `query` — which says nothing about what the row
    is. The workload column names the deployment and does. But most deployments are just the
    container under a chart prefix, where repeating it is noise, so those are left bare.
    Prometheus-sourced rows aggregate the workload away and are never qualified.
    """
    workload, container = row.get("workload", ""), row["container"]
    if workload in ("", container):
        return container
    for prefix in ("runai-backend-", "runai-"):
        if workload.removeprefix(prefix) == container:
            return container
    return f"{container} ({workload})"


def compare_rows(rows, defaults, aliases, prom=False):
    """Emit shipped default (request and limit) vs recommended request, CPU then memory."""
    out = []
    for r in sorted(rows, key=lambda x: x["container"]):
        d = defaults_for(r["container"], defaults, aliases)
        rec_c, rec_m = parse_cpu(r["cpu_upper_bound"]), parse_mem(r["memory_upper_bound"])
        if prom:
            rec_c = rec_c * PROM_MARGIN if rec_c is not None else None
            rec_m = rec_m * PROM_MARGIN if rec_m is not None else None
        def_cpu, new_cpu = cells(parse_cpu(d["req_cpu"]), rec_c, c, SIGNIFICANT_CPU)
        def_mem, new_mem = cells(parse_mem(d["req_mem"]), rec_m, g, SIGNIFICANT_MEM)
        lim_cpu = limit_cell(parse_cpu(d["lim_cpu"]), rec_c, c)
        lim_mem = limit_cell(parse_mem(d["lim_mem"]), rec_m, g)
        if prom:
            new_cpu += PROM_MARK
            new_mem += PROM_MARK
        else:
            if is_cpu_floor(r):
                new_cpu += " `floor`"
            if is_mem_floor(r):
                new_mem += " `floor`"
        out.append(f"| {service_label(r)} | {def_cpu} | {lim_cpu} | {new_cpu} "
                   f"| {def_mem} | {lim_mem} | {new_mem} |")
    return out


COMPARE_HEADER = [
    "| Service | default request CPU | default limit CPU | recommended request CPU "
    "| default request Mem (GB) | default limit Mem (GB) | recommended request Mem (GB) |",
    "|---|---|---|---|---|---|---|",
]

PROM_FOOTNOTE = (
    f"\n`{PROM_MARK}` Prometheus-sourced, with the {PROM_MARGIN - 1:.0%} margin applied — see "
    "How to read the recommendation.\n")

COMPARE_LEGEND = (
    f"{MARK} on a **default** means the shipped value is below what was measured. "
    f"{MARK} on a **measurement** means no default is declared for it at all, and the service "
    f"consumed more than {SIGNIFICANT_CPU} cores / {SIGNIFICANT_MEM/1e9:g} GB.\n")


def sum_section(label, groups, w):
    """Emit a one-row total table. groups is a list of row lists."""
    s = {"cpu": 0.0, "mem": 0.0}
    total = cpu_n = mem_n = 0
    for rows, margin in groups:
        for r in rows:
            total += 1
            if not is_cpu_floor(r):
                cpu_n += 1
                s["cpu"] += (parse_cpu(r["cpu_upper_bound"]) or 0) * margin
            if not is_mem_floor(r) and has_mem(r):
                mem_n += 1
                s["mem"] += parse_mem(r["memory_upper_bound"]) * margin
    w(f"\n**{label}** — floor and unmeasured values excluded per dimension, since they are "
      f"instrument readings rather than measurements. CPU sums {cpu_n} of {total} services, "
      f"memory {mem_n} of {total}.\n")
    w("| total CPU | total Mem (GB) |")
    w("|---|---|")
    w(f"| {c(s['cpu'])} | {g(s['mem'])} |")
    return s


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("bundle")
    ap.add_argument("--monorepo", default=os.path.expanduser("~/work/runai"))
    ap.add_argument("-o", "--output")
    args = ap.parse_args()

    b = args.bundle
    vpa = [r for r in read_csv(os.path.join(b, "vpa", "vpa-recommendations.csv"))
           if not is_harness(r["container"])]
    cm = [r for r in read_csv(os.path.join(b, "vpa", "container-metrics.csv"))
          if not is_harness(r["container"])]
    etcd = read_csv(os.path.join(b, "vpa", "etcd-db-size.csv"))
    meta_path = os.path.join(b, "run-metadata.json")
    meta = derive_scale_params(
        json.load(open(meta_path)) if os.path.exists(meta_path) else {})
    if not vpa:
        path = os.path.join(b, "vpa", "vpa-recommendations.csv")
        sys.exit(f"{path} is missing or holds no non-harness rows" if os.path.exists(path)
                 else f"no {path}")

    groups, gm_path = load_group_mapping(args.monorepo)
    if not groups:
        sys.exit(f"could not read ServiceGroupMapping from {gm_path}")
    cluster_defaults = load_cluster_defaults(args.monorepo)
    cp_defaults = load_cp_defaults(args.monorepo)

    group_of, declared_key, deployed_name = build_classifier(groups)
    grouped, ungrouped = split_by_group(vpa, group_of)

    out = []
    w = out.append
    w(f"# Resource recommendation data — {meta.get('scale_level', 'unknown')} scale\n")
    w("CPU in cores, memory in GB. Measured values are unit-converted copies of the run's "
      "artifacts.\n")

    w("\n## How to read the recommendation\n")
    w("**The recommendation is a request, taken from VPA's upper bound. Limit recommendations are "
      "out of scope.** All three VPA figures recommend the *request*, not a request/limit pair; "
      "the upper bound is the one that covers observed peaks, which suits guaranteed operation "
      "rather than density.\n")
    w("VPA derives them from a decaying usage histogram — CPU as a percentile, memory as each "
      "window's peak — plus a 15% margin and a confidence multiplier that shrinks as history "
      "accumulates. At stock settings a few hours of history inflates the upper bound ~13x, so "
      "this run used `confidence-interval-cpu/memory=10m`, `memory-aggregation-interval=5m x 24` "
      "and `pod-recommendation-min-cpu-millicores=1` / `min-memory-mb=1`.\n")
    w(f"`{PROM_MARK}` — measured from Prometheus because VPA produced no recommendation. A "
      f"{PROM_MARGIN - 1:.0%} margin stands in for VPA's margin and widening, making it a weaker "
      "basis than the VPA rows.\n")
    w(f"`floor` — VPA's histogram floor (`{CPU_FLOOR}` CPU, `{MEM_FLOOR}` memory): the service "
      "consumed too little to measure, not that it needs that much. Marked per dimension, since a "
      "service can idle on CPU while using real memory, and excluded from sums.\n")

    # --- testing method
    w("\n## How this was measured\n")
    w("Nodes and GPUs are simulated with **KWOK**, and workload pods never execute: no container "
      "is created and no image is pulled — KWOK patches each pod's status to Succeeded once its "
      "duration elapses. Every workload is still submitted through the public API, so Run:ai "
      "services do the full real work of admission, scheduling, reconciliation, status "
      "propagation and sync. Not exercised: image pulls, real GPU allocation, kubelet-side work, "
      "workload logs, and storage provisioning.\n")
    w("**Usage telemetry is largely absent.** Fake GPUs emit no DCGM metrics and pods that never "
      "run emit no cAdvisor or node-level series, though object-count cardinality is still "
      "exercised. Monitoring-stack figures are therefore a floor for a real cluster of the same "
      "size.\n")
    if meta.get("mode") == "rate":
        w("**Rate-mode traffic.** Submission ramps up quickly until the target workload population "
          "is reached, then holds a steady arrival rate **R** for the rest of the run. That target "
          "deliberately exceeds the number of workloads the cluster's GPUs can run at once "
          "(**W_cap**), so the surplus stays pending and the run maintains a constant queue "
          "instead of draining it. Each workload is given a lifetime of **D = W_cap / R** — the "
          "time arrivals at rate R take to fill the cluster once — so a workload finishes roughly "
          "as its replacement is admitted and the cluster stays saturated throughout.\n")
    w("**The control plane sees only two API paths**: workload submission and workload listing. "
      "No UI sessions, authentication churn, policy or asset editing, other endpoints, or "
      "additional clusters. Read its figures as a lower bound for the workload path, not a "
      "production control-plane sizing.\n")
    w("**Self-hosted: control plane and cluster plane share one Kubernetes cluster**, so their "
      "totals combine here. A split deployment does not divide these figures cleanly — its "
      "control plane serves all of its clusters, not just this one.\n")

    # --- 1. service groups
    w("\n## Service groups\n")
    w("A group's value is copied to **every** service in it, so it must be sized to its largest "
      "member and the group total is that value times the number of running instances. The top "
      "consumer named "
      "below is therefore the service that actually sets the value, and every other member is "
      "provisioned to match it whether or not it needs to be.\n")
    w("| Group | Instances | group highest CPU | top CPU consumer | group highest Mem (GB) "
      "| top Mem consumer | total CPU (highest x instance count) "
      "| total Mem (GB) (highest x instance count) |")
    w("|---|---|---|---|---|---|---|---|")
    tot = {"cpu": 0.0, "mem": 0.0}
    members = []
    for grp in ("scheduling", "sync", "workload"):
        # The group value is a per-pod request, so the multiplier is the number of running
        # instances — not the number of declared services. They differ for the scheduler, which
        # runs one instance per node pool plus one for the default pool. A declared service with
        # no measurement still gets the value, so it counts as one.
        declared = sorted(camel_to_kebab(f) for f, gr in groups.items() if gr == grp)
        rows = [r for r in grouped.get(grp, []) if not is_floor(r)]
        if not rows:
            continue
        instances = {}
        for r in grouped.get(grp, []):
            instances[declared_key(r["container"])] = instances.get(
                declared_key(r["container"]), 0) + 1
        n = sum(max(1, instances.get(name, 0)) for name in declared)
        top_c = max(rows, key=lambda r: parse_cpu(r["cpu_upper_bound"]) or 0)
        mx_c = parse_cpu(top_c["cpu_upper_bound"]) or 0
        tot["cpu"] += mx_c * n
        # A group whose members all lack a memory series has no maximum to apply, so its
        # memory cells read "-" rather than contributing a zero to the grand total.
        mem_rows = [r for r in rows if has_mem(r)]
        top_m = max(mem_rows, key=lambda r: parse_mem(r["memory_upper_bound"])) if mem_rows else None
        mx_m = parse_mem(top_m["memory_upper_bound"]) if top_m else None
        if mx_m is not None:
            tot["mem"] += mx_m * n
        w(f"| {grp} | {n} | {c(mx_c)} | {top_c['container']} | {g(mx_m)} "
          f"| {top_m['container'] if top_m else '-'} "
          f"| {c(mx_c * n)} | {g(mx_m * n if mx_m is not None else None)} |")
        members.append((grp, declared, grouped.get(grp, [])))
    w(f"| **all groups** | | | | | | **{c(tot['cpu'])}** | **{g(tot['mem'])}** |")

    w("\nMembers of each group as declared in `ServiceGroupMapping`, with what each contributed. "
      "A declared service with no measurement still receives the group's value.\n")
    for grp, declared, rows in members:
        seen = {}
        for r in rows:
            seen.setdefault(declared_key(r["container"]), []).append(r)
        parts = []
        for name in declared:
            got = seen.get(name, [])
            shown = deployed_name.get(name, name)
            if not got:
                parts.append(f"`{shown}` (not observed)")
            elif all(is_floor(r) for r in got):
                parts.append(f"`{shown}` (floor)")
            elif len(got) > 1:
                parts.append(f"`{shown}` ({len(got)} instances)")
            else:
                parts.append(f"`{shown}`")
        w(f"- **{grp}** ({len(declared)}): {', '.join(parts)}")
    w("\n`floor` never accumulated a real sample; `not observed` produced no row at all — it is "
      "either not deployed in this configuration or was not exercised by this workload mix. "
      "Neither influences the group value, but both are sized by it.\n")
    w("**The scheduler runs one instance per node pool, plus one for the default pool**, and each "
      "instance requests the group value independently. The instance count above accounts for "
      "this.\n")

    # --- 2. control plane
    cp = [r for r in vpa if r["namespace"] == "runai-backend"]
    w("\n## Control plane services\n")
    w("Configured per service in the control-plane Helm chart; there is no group mechanism here.\n")
    w(COMPARE_LEGEND)
    out.extend(COMPARE_HEADER)
    out.extend(compare_rows(cp, cp_defaults, None))

    cp_tot = sum_section("Control plane total", [(cp, 1.0)], w)

    # --- 3. ungrouped cluster services
    prom = {r["container"]: r for r in cm if r["namespace"] == "runai"}
    w("\n## Run:ai cluster services not in any group\n")
    w("No group setting sizes these; each is configured individually.\n")
    w(COMPARE_LEGEND)
    out.extend(COMPARE_HEADER)
    out.extend(compare_rows(ungrouped, cluster_defaults, CLUSTER_ALIASES))
    # Fall back to Prometheus only for services VPA never reported at all, and only when
    # they belong to no group — otherwise grouped services would be counted twice.
    vpa_runai = {r["container"] for r in vpa if r["namespace"] == "runai"}
    extra = [r for cont, r in sorted(prom.items())
             if cont not in vpa_runai and group_of(cont) is None]
    out.extend(compare_rows(extra, cluster_defaults, CLUSTER_ALIASES, prom=True))
    if extra:
        w(PROM_FOOTNOTE)
    ung_tot = sum_section("Ungrouped cluster services total",
                          [(ungrouped, 1.0), (extra, PROM_MARGIN)], w)

    # --- 4. kube-system
    w("\n## kube-system services\n")
    w("Every kube-system service is listed, excluding the test harness we install ourselves. "
      "Not part of Run:ai, but they run on the same nodes and their consumption is driven by the "
      "same object churn, so they belong in the sizing picture.\n")
    w("These figures are for **infrastructure sizing, not configuration**: they say how much of a "
      "node these services occupy, not what should be requested for them. No margin is applied — "
      "the observed maximum is shown as measured.\n")
    w(f"{MARK} marks a consumer above {SIGNIFICANT_CPU} cores or {SIGNIFICANT_MEM/1e9:g} GB.\n")
    w("| Service | max CPU | max Mem (GB) |")
    w("|---|---|---|")
    ks = {"cpu": 0.0, "mem": 0.0}
    ks_rows = sorted([r for r in cm if r["namespace"] == "kube-system"],
                     key=lambda x: x["container"])
    mem_missing = 0
    for r in ks_rows:
        mc = parse_cpu(r["cpu_upper_bound"]) or 0
        mm = parse_mem(r["memory_upper_bound"])
        ks["cpu"] += mc
        if mm is None:
            mem_missing += 1
        else:
            ks["mem"] += mm
        cpu_cell = f"**{c(mc)}** {MARK}" if mc > SIGNIFICANT_CPU else c(mc)
        mem_cell = f"**{g(mm)}** {MARK}" if mm is not None and mm > SIGNIFICANT_MEM else g(mm)
        w(f"| {r['container']} | {cpu_cell} | {mem_cell} |")

    note = (f" Memory excludes {mem_missing} of {len(ks_rows)} services with no memory series."
            if mem_missing else "")
    w(f"\n**kube-system total** — observed maximum per service, summed.{note}\n")
    w("| total CPU | total Mem (GB) |")
    w("|---|---|")
    w(f"| {c(ks['cpu'])} | {g(ks['mem'])} |")

    # --- totals across every section
    sections = [
        ("Service groups (highest x instance count)", tot["cpu"], tot["mem"]),
        ("Control plane", cp_tot["cpu"], cp_tot["mem"]),
        ("Ungrouped cluster services", ung_tot["cpu"], ung_tot["mem"]),
        ("kube-system", ks["cpu"], ks["mem"]),
    ]
    w("\n## Totals\n")
    w("The service-groups row is what the group mechanism would **reserve** (its value applied to "
      "every member, including those that never needed it); the other rows sum each service's own "
      "recommendation.\n")
    w("| Section | total CPU | total Mem (GB) |")
    w("|---|---|---|")
    for name, cpu, mem in sections:
        w(f"| {name} | {c(cpu)} | {g(mem)} |")

    w("\n**Grand total** — every section combined.\n")
    w("| total CPU | total Mem (GB) |")
    w("|---|---|")
    w(f"| {c(sum(s[1] for s in sections))} | {g(sum(s[2] for s in sections))} |")

    # --- 5. etcd
    w("\n## etcd database size\n")
    if etcd:
        w("| Metric | min (GB) | p95 (GB) | max (GB) |")
        w("|---|---|---|---|")
        for r in etcd:
            w(f"| {r['metric']} | {g(num(r['min_bytes']))} | {g(num(r['p95_bytes']))} "
              f"| {g(num(r['max_bytes']))} |")
    else:
        w("_No etcd-db-size.csv in this bundle (run predates the etcd size table)._")

    # --- 6. scale parameters
    w("\n## Scale parameters\n")
    w("| Parameter | Value | Meaning |")
    w("|---|---|---|")
    for key, name, desc in SCALE_PARAMS:
        if key in meta:
            w(f"| {name} | `{meta[key]}` | {desc} |")

    text = "\n".join(out) + "\n"
    if args.output:
        open(args.output, "w").write(text)
        print(f"wrote {args.output}", file=sys.stderr)
    else:
        print(text)


if __name__ == "__main__":
    main()
