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
  load/                     steady submission rates; arrives with the load-ramp suite
```

Knobs of the `scale-burst` suite:

| knob | default | meaning |
|---|---|---|
| `workloads` | `[100, 250, 500]` | round sizes; each round creates that many workloads |
| `workloadType` | `Deployment` | `Deployment` or `Training` |
| `podsPerWorkload` | 1 | replicas per Deployment; must be 1 for `Training` |
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

## Results

Every run writes `reports/results-<timestamp>.json` and adds it to
`reports/manifest.json`, the ordered list of runs. The file holds run
metadata (start time, provider, cluster type, versions, clock offset) and one
section per scenario