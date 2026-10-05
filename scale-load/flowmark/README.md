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
`--only` and `--workloads` are what `SCENARIO` and `WORKLOADS` set.

## Scenarios

A scenario is one yaml file under `scenarios/`, sorted into a folder per
test category. `suite` names what runs, the other keys are knobs of that
suite. An unknown knob fails the scenario, and so does a file in the wrong
folder: every suite declares its category and the runner checks the two match.

```
scenarios/
  selfcheck.yaml            checks the wiring, needs no environment
  scale/                    how latency grows with the number of workloads
    deployment-burst.yaml   Deployment created through Kubernetes → ExternalWorkload appears
    training-burst.yaml     Training created through the control plane → WorkloadRunner appears
    workspace-burst.yaml    Workspace created through the control plane → WorkloadRunner appears
  load/                     how latency behaves under a steady stream of submissions
    workspace-ttj.yaml      Workspaces submitted at a fixed rate → time to job, split into spans
```

Knobs of the `scale-burst` suite:

| knob | default | meaning |
|---|---|---|
| `workloads` | `[100, 250, 500]` | round sizes; each round creates that many workloads |
| `workloadType` | `Deployment` | `Deployment`, `Training` or `Workspace` |
| `podsPerWorkload` | 1 | replicas per Deployment; must be 1 for `Training` and `Workspace` |
| `image` | `registry.k8s.io/pause:3.9` | container image; point it at a mirror on air-gapped clusters |
| `submitConcurrency` | 16 | parallel create and delete calls |
| `kubeRequestsPerSecond` | 50 | Kubernetes client rate limit |
| `controlPlaneRequestsPerSecond` | 25 | control plane client rate limit |
| `timeout` | `15m` | how long a round waits for all workloads to arrive |
| `setupTimeout` | `5m` | project creation, quiesce between rounds, cleanup |
| `settle` | `90s` | quiet time between rounds |
| `slo` | none | e.g. `p90_ms: 5000`; copied into the results, does not fail the run |
| `cluster` | single cluster | cluster name when the control plane has several |

Each run creates its own project, runs the rounds, deletes everything, and
deletes the project. Latency per workload is the time from the create
response to the first watch event of its cluster object.

Knobs of the `load-ttj` suite:

| knob | default | meaning |
|---|---|---|
| `users` | 10 | parallel submitters; the most requests in flight at once |
| `rate` | 2 | target submissions per second; may not exceed `controlPlaneRequestsPerSecond` |
| `duration` | `5m` | how long submissions keep coming |
| `workloadType` | `Workspace` | `Workspace` or `Training`, both submitted through the control plane |
| `image` | `registry.k8s.io/pause:3.9` | container image of every workload |
| `timeout` | `10m` | how long to wait for arrivals after the last submission |
| `setupTimeout` | `5m` | project creation, each create call, cleanup |
| `kubeRequestsPerSecond`, `controlPlaneRequestsPerSecond`, `cluster`, `slo` | as above | shared with `scale-burst` |

Each accepted Workspace is timed in three legs on one clock: `api_accepted`
from the request to the 200, `cluster_synced` from the 200 to its
WorkloadRunner on the cluster, `pod_created` from the WorkloadRunner to its
first pod. `ttj`, the time to job, runs from the 200 to the first pod. A
rejected submission lowers `success_rate` and enters no latency. The pacer
never fires missed ticks later, so `actual_rate_per_sec` next to
`target_rate_per_sec` shows whether the environment kept up with the load.

## Results

Every run writes `reports/results-<timestamp>.json` and adds it to
`reports/manifest.json`, the ordered list of runs. The file holds run
metadata (start time, provider, cluster type, versions) and one
section per scenario. A `scale-burst` section has `p50_ms@N`, `p90_ms@N`,
`p99_ms@N`, `throughput_per_sec@N`, `never_arrived_count@N`, `failed_count@N` and
the knobs used, `N` being the round size; `makespan_ms@N` runs from the first
create response to the last arrival and `throughput_per_sec@N` is the number
of arrived workloads divided by it. A `load-ttj` section has `<span>_p50_ms`,
`<span>_p90_ms` and `<span>_p99_ms` for `api_accepted`, `cluster_synced`,
`pod_created` and `ttj`, plus `submitted`, `success_rate`,
`target_rate_per_sec` and `actual_rate_per_sec`. Workloads that never arrived
are counted in `never_arrived_count` and never as zero latency. The health of
the measurement itself is not a metric: when the cluster watch had to be
reopened, or a WorkloadRunner or pod was seen more than a second before the
step that precedes it, the section fails and `error` explains why, and the
log carries the same line.