# Validity Gate Monitor

Runs alongside a resource-recommendation scale test to assert that the cluster infrastructure remained healthy throughout the run. Emits a JSON report and exits non-zero if any validity gate fails.

## What it checks

**Validity Gates** — failure means the run result cannot be trusted (sets `OverallStatus=FAIL`, `InfraLimited=true`):
1. API Server Request Latency P50 — threshold 500ms
2. etcd WAL Fsync P99 — threshold 100ms
3. etcd Proposals Pending — threshold: sustained > 0 for 30s
4. etcd DB Size Ratio — threshold 70%
5. Pod Restarts (runai + runai-backend) — threshold 0
6. Node NotReady Count — threshold 0
7. System Pods Unhealthy — threshold 0
8. Prometheus Scrape Gaps — threshold 0

**Hardware Sufficiency** — failure means the cluster was resource-constrained (also sets `OverallStatus=FAIL`, `InfraLimited=true`):
9. Node CPU Utilization Peak — threshold 80%
10. Container CPU Peak Fraction of Allocatable — threshold 50%

**Degraded Metrics** — breach sets `Degraded=true` but does NOT fail the run; infra is stable, data is usable:
11. Scheduler Allocate Cycle Latency
12. Reconciliation Latency P99 by Controller
13. Work Queue Depth Peak by Controller
14. Workload Submission API Latency P50/P99
15. Workload List API Latency P95 — threshold 2s

## Prerequisites

- A Run:AI cluster with Prometheus Operator installed
- `kubectl` configured to reach the cluster (`KUBECONFIG` env var or `~/.kube/config`)
- Docker (for local Grafana)
- Go 1.22+ (to build/run the monitor binary)

## Cluster setup — run once per env

```bash
./scripts/setup-monitoring.sh ~/.kube/<env-name>
```

Idempotent. It:

1. **Applies `manifests/etcd-metrics-proxy.yaml`** — DaemonSet (socat) + Service + ServiceMonitor
   tunnelling etcd's localhost metrics port (2381) to a scrapeable port (2382) on the control-plane
   node. etcd only listens on loopback, so without this proxy Prometheus cannot scrape it and gates
   2–4 are permanently `UNAVAILABLE`. The DaemonSet runs with `hostNetwork` to reach that loopback
   port; the listener binds to the node's primary IP (the pod IP under `hostNetwork`) rather than
   `0.0.0.0`, which is the address Prometheus scrapes via the headless Service.
2. **Applies `manifests/runai-backend-servicemonitor.yaml`** — tells Prometheus to scrape every
   `runai-backend` service exposing a `metrics` port. Without it the workload submission and
   workload list latency histograms do not exist.
3. **Enables the Prometheus admin API**, which the TSDB snapshot endpoint requires. This is patched
   onto `runaiconfig.spec.prometheus.spec.enableAdminAPI` — *not* onto the Prometheus CR, which the
   operator overwrites wholesale from RunaiConfig on every reconcile. The patch rolls the Prometheus
   StatefulSet, so run this before any measurement: the restart would otherwise trip validity gate 5.
4. **Verifies every scrape target the gates depend on is up** — the five jobs in
   `CriticalMetricSources` plus `runai-backend`. Failing here costs seconds; discovering a dead
   `kube-state-metrics` mid-run costs the run.

## Local Grafana dashboard

```bash
./scripts/monitor-live.sh ~/.kube/<env-name>
```

Port-forwards Prometheus to **9091**, starts Grafana on **3001**, prints the dashboard URLs, and
tears both down on Ctrl-C. No login required (anonymous admin); the datasource is pre-configured to
reach Prometheus at `host.docker.internal:9091`.

- http://localhost:3001/d/validity-gate-monitor
- http://localhost:3001/d/workload-activity

### Effect of Grafana on scale test measurements

Keeping Grafana open and auto-refreshing (every 10 s by default) during a scale test
contaminates the resource measurements. Each refresh fires all panels as concurrent PromQL
queries against Prometheus. The expensive ones — histogram quantiles, high-cardinality
`sum by (pod, container)` aggregations — consume Prometheus CPU, which shows up directly
in the container CPU/memory panels you are trying to get a clean read on. Prometheus also
competes with its own ingestion loop under query load, which can introduce scrape delays.

**Recommended two-iteration approach:**

1. **Health-check run (Grafana open):** Run with Grafana open and auto-refresh enabled.
   Goal: verify the environment is stable and hardware is sufficient — etcd, API server
   latency, pod restarts, node memory/CPU all within thresholds. The resource numbers from
   this run are dirty (include Grafana + Prometheus query overhead) and should not be used
   as final results.

2. **Measurement run (Grafana closed):** Close the Grafana tab (or set dashboard refresh
   to "Off") before starting the test. Prometheus continues scraping and storing everything
   — no data is lost. After the run, open Grafana, set the time range to the test window,
   and replay all panels over the stored data with zero interference on the run itself.

## Not an environment readiness check

Gate 5 (`pod_restarts`) has a threshold of zero, but `remove-limits` and the Prometheus roll restart
pods by design — so this monitor either fails a good env or passes one that is still stabilising.
Use `cmd/env-readiness` to decide whether an env is ready to test against.

## Running the monitor

From `unified_tests/`:
```bash
KUBECONFIG=~/.kube/<env-name> go run ./cmd/validity-gate-monitor/... \
  --heartbeat_interval=5m \
  --output_dir=/tmp/validity-gate \
  --run_duration=2h        # omit to run until Ctrl-C / SIGTERM
```

The monitor logs a heartbeat every `heartbeat_interval` with current metric values, then on exit writes a structured JSON report to `output_dir/validity-gate-report.json`.

Crash logs (if any pod restarted) are collected at each heartbeat into `output_dir/crash-logs/<ns>-<pod>-<container>.log`.

### Config options (env vars / flags via viper)

| Key | Default | Description |
|---|---|---|
| `heartbeat_interval` | `5m` | How often to log live metric values |
| `run_duration` | `` | Total run duration; empty = run until signal |
| `output_dir` | `/tmp/validity-gate` | Where to write the JSON report and crash logs |
| `prometheus_pod_name` | `prometheus-runai-0` | Pod to port-forward for queries |
| `prometheus_namespace` | `runai` | Namespace of the Prometheus pod |
| `prometheus_port` | `9090` | Remote port on the Prometheus pod |
| `target_namespaces` | `[runai, runai-backend]` | Namespaces checked for pod restarts / unhealthy pods |
| `etcd_proposal_sustained_secs` | `30` | Seconds etcd proposals must be > 0 to count as a failure |

## Reading the report

```json
{
  "overall_status": "PASS",        // PASS or FAIL
  "infra_limited": false,          // true if any validity gate or hardware sufficiency check failed
  "degraded": false,               // true if any degraded metric breached its threshold
  "validity_gates": [...],
  "hardware_sufficiency": [...],
  "degraded_metrics": [...]
}
```

Each metric entry:
```json
{
  "metric": "api_server_request_latency_p50",
  "value": "12ms",
  "threshold": "500ms",
  "headroom": "97.6%",
  "status": "PASS"        // PASS | FAIL | DEGRADED | UNAVAILABLE
}
```

`UNAVAILABLE` means the metric had no data (e.g., etcd metrics before `etcd-metrics-proxy` is applied, or workload list latency when no list traffic occurred). It does not count as a failure.
