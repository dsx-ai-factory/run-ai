# FlowMark

Scale benchmark for Run:ai flows. It creates workloads in bursts against a
real Run:ai environment, measures how long each one takes to reach the
cluster, and writes the numbers as JSON for a trend dashboard.

## Run a test

1. Get a Run:ai environment (control plane plus cluster) and its kubeconfig.
2. Create `.env` in this directory (git-ignored):

   ```
   FLOWMARK_CP_URL=https://my-control-plane.example.com
   FLOWMARK_CP_USER='username'
   FLOWMARK_CP_PASSWORD='...'
   KUBECONFIG=/path/to/kubeconfig
   ```

   `FLOWMARK_CP_TOKEN=...` can replace user and password.

3. Run a scenario:

   ```
   make run SCENARIO=training-burst WORKLOADS=100   # one round of 100
   make run SCENARIO=training-burst                 # rounds as written in the file
   make run                                        # every scenario
   ```

Progress is logged to stderr. Results land in `./reports/`. Exit code 0
means every scenario passed and the results were written; 1 means something
failed, and the last line says what.

Without make: `make build`, then `./bin/flowmark --help` lists every flag.
`--only` and `--workloads` are what `SCENARIO` and `WORKLOADS` set;
`--workloads` reaches only the scenarios whose suite has that knob, so
`make run WORKLOADS=20` leaves the load scenarios untouched. Every selected
scenario file is checked before the first one starts.
`--regression-window` (7) and `--regression-threshold` (0.15) tune the
regression check described under Results.

## Scenarios

A scenario is one yaml file under `scenarios/`, sorted into a folder per
test category. `suite` names what runs, the other keys are knobs of that
suite. An unknown knob fails the scenario, and so does a file in the wrong
folder: every suite declares its category and the runner checks the two match.

```
scenarios/
  scale/                    how latency grows with the number of workloads
    training-burst.yaml     Training created through the control plane → WorkloadRunner appears
    workspace-burst.yaml    Workspace created through the control plane → WorkloadRunner appears
  load/                     how latency behaves under a steady stream of submissions
    workspace-ttj.yaml      Workspaces submitted at a fixed rate → time to job, split into spans
```

The selfcheck scenario is not a test and stays out of this tree. It lives
next to its suite as `internal/suites/selfcheck/selfcheck.yaml`; run it with
`--scenarios internal/suites/selfcheck` to check the wiring without an
environment.

Knobs of the `scale-burst` suite:

| knob | default | meaning |
|---|---|---|
| `workloads` | `[100, 250, 500]` | round sizes; each round creates that many workloads. Before round 1 one workload is created and removed to warm up the new project; it is not measured |
| `workloadType` | `Training` | `Training` or `Workspace`, both created through the control plane |
| `image` | `registry.k8s.io/pause:3.9` | container image; point it at a mirror on air-gapped clusters |
| `submitConcurrency` | 16 | parallel create and delete calls |
| `kubeRequestsPerSecond` | 50 | Kubernetes client rate limit |
| `controlPlaneRequestsPerSecond` | 25 | control plane client rate limit |
| `timeout` | `15m` | how long a round waits for all workloads to arrive |
| `setupTimeout` | `5m` | project creation, quiesce between rounds, cleanup |
| `settle` | `90s` | quiet time between rounds |
| `slo` | none | upper bounds on `p50_ms`, `p90_ms` or `p99_ms`, e.g. `p90_ms: 5000`, and a lower bound on `throughput_per_sec`, e.g. `throughput_per_sec: 4`; add `@<round size>` for one round, e.g. `p90_ms@500: 60000`, otherwise the plain key covers every round; per-round keys for rounds that are not in the run, for example after `--workloads 20`, are ignored and left out of the report; a round above its limit is reported as a regression, the run still passes |
| `cluster` | single cluster | cluster name when the control plane has several |

Each run creates its own project, runs the rounds, deletes everything, and
deletes the project. Latency per workload is the time from the create
response to the first watch event of its cluster object.

Knobs of the `load-ttj` suite:

| knob | default | meaning |
|---|---|---|
| `users` | 10 | parallel submitters; the most requests in flight at once |
| `spawnRate` | 0 | users joining per second at the start; 0 starts every user at once. Submissions ramp with the active users, so `rate` is reached once all users are in |
| `rate` | 2 | target submissions per second; may not exceed `controlPlaneRequestsPerSecond` |
| `duration` | `3m` | how long submissions keep coming |
| `workloadType` | `Workspace` | `Workspace` or `Training`, both submitted through the control plane |
| `image` | `registry.k8s.io/pause:3.9` | container image of every workload |
| `timeout` | `10m` | how long to wait for arrivals after the last submission |
| `setupTimeout` | `5m` | project creation, each create call, cleanup |
| `slo` | none | upper bounds named `<span>_<stat>`, e.g. `ttj_p90_ms: 10000`, `cluster_synced_p99_ms: 12000`, and `failed_count`, the most rejected submissions allowed; `p90_ms` alone belongs to `scale-burst` |
| `kubeRequestsPerSecond`, `controlPlaneRequestsPerSecond`, `cluster` | as above | shared with `scale-burst` |

Each accepted Workspace is timed in three legs on one clock: `api_accepted`
from the request to the 200, `cluster_synced` from the 200 to its
WorkloadRunner on the cluster, `pod_created` from the WorkloadRunner to its
first pod. `ttj`, the time to job, runs from the 200 to the first pod. A
rejected submission counts in `failed_count` and enters no latency. The pacer
never fires missed ticks later; when fewer than 90% of the planned
submissions were made, the section fails and says so.

## Dashboard

The nightly results are shown at https://dsx-ai-factory.github.io/run-ai/flowmark/: a
table of runs and a trend chart per scenario and KPI, with the SLO line, the
flagged nights and the control plane build of every point. The page lives in
`dashboard/` and is described in [dashboard/README.md](dashboard/README.md).

## Results

Every run writes `reports/results-<timestamp>.json` and adds it to
`reports/manifest.json`, the ordered list of runs. The file holds run
metadata (start time, provider, cluster type, versions) and one
section per scenario. A `scale-burst` section has `p50_ms@N`, `p90_ms@N`,
`p99_ms@N` and `throughput_per_sec@N`, `N` being the
round size; `throughput_per_sec@N` is the number
of arrived workloads divided by the time from the first create response to
the last arrival. A `load-ttj` section has `<span>_p50_ms`,
`<span>_p90_ms` and `<span>_p99_ms` for `api_accepted`, `cluster_synced`,
`pod_created` and `ttj`, plus `failed_count`, the creates the control plane
rejected; the accepted ones are still timed. A workload that never reached
the cluster fails the section, is counted in `error` and never enters a
latency as zero. The health of
the measurement itself is not a metric: when the cluster watch had to be
reopened without a resume point, so events in the gap may have been missed,
or a WorkloadRunner or pod was seen more than a second before the step that
precedes it, the section fails and `error` explains why, and the log carries
the same line. A watch the API server closes routinely is resumed from the
last seen resourceVersion and is only logged.

The results are then judged in two ways. Against history: every p90 is
compared with the median of its last 7 values in earlier runs of the same
`cluster_type` found in the output directory, and more than 15% above that
median is a regression. Runs where the scenario did not run or its section
failed do not count, an earlier file that cannot be read is skipped and named
in the log, and with fewer than 7 values the comparison is skipped.
Against the SLO: the limits from the scenario file are copied into the
section's `slo` block, and every metric with a limit there, p90 or not, is a
regression when it misses it, above a latency limit or below a throughput
floor, regardless of history. A flagged metric gets
a sibling `<metric>_regression: true`, `regression_detected` at the top of the
file becomes true, and `regressions` lists every finding with a ready-made
message, for example "training-burst p90_ms@100: 5900 ms is +18% over the
median of 5000 ms from the last 7 runs". The same lines appear in the log. A
regression does not change the exit code. The median adapts: a slowdown that
lasts is reported for about four nights and then becomes the new baseline.