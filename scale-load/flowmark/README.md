# FlowMark

Performance and scale benchmark for Run:ai flows. Runs scenario files (yaml)
against a ready Run:ai environment, measures flow latencies from watch events,
and publishes results as versioned JSON to a public dashboard.

## Build

```
make build
```

The binary lands in `bin/flowmark`.

## Run

```
./bin/flowmark --scenarios ./scenarios --output ./out
```

Runs every `*.yaml` scenario in the scenarios directory, then writes the run
into `--output`: one `results-<timestamp>.json` per run plus `manifest.json`,
the ordered list of all runs in that directory. After writing, FlowMark reads
both files back and compares them; a difference or a missing manifest entry
is an error, so a run whose numbers did not land fails like any other failed
run. Exit code 0 means every scenario passed and the results were written and
verified; 1 means any error: a failed scenario (the results are still
written), invalid scenarios or scenario files, or a failed write or
verification, in which case the message names the failed step, e.g.
`report: write results results-20260928T010000Z.json: ...`.

Uploading to S3 is not the binary's job. As in the KAI-Scheduler scale tests,
the nightly workflow publishes the output directory with the AWS CLI after
the run and maintains the bucket's `manifest.json` the same way; that workflow
arrives with the nightly task. The binary stays fully offline.

`./bin/flowmark --help` lists all flags.

Suites that talk to a Run:ai environment take the cluster from
`--kubeconfig`, the control plane from `--cp-url`, and credentials from
environment variables: `FLOWMARK_CP_USER` and `FLOWMARK_CP_PASSWORD`, or
a pre-issued `FLOWMARK_CP_TOKEN`. Credentials are never accepted as
flags and never appear in output. The selfcheck scenario needs none of
these.

### The results file

```
schema: "v1"
regression_detected: false     # always false until the regression task lands
metadata:
  date, finished_at            # UTC; date is the run start and names the file
  provider, cluster_type       # --provider / --cluster-type, default "local"; CI passes the real ones
  control_plane_version, cluster_version   # --cp-version / --cluster-version, omitted when unknown
  clock_offset_ms              # cluster clock minus ours, measured once at startup when a kubeconfig is given
sections:                      # one per scenario file: category, metrics, failed, error
```

These are the fields the dashboard draws from: `category` picks the Scale or
Load tab, each section's `p90_ms` and `slo_p90_ms` give the line and its SLO,
`metadata.date` is the x axis, `metadata.control_plane_version` labels every
dot and marks version changes, and `regression_detected` colours a run red.
`clock_offset_ms` is context only and never enters a latency; if the probe
fails the run continues and `clock_offset_error` holds the reason. The file
name carries the start time to the second, so two runs on the same day never
overwrite each other. The manifest is `{"runs": [...]}`, appended in run
order, and re-publishing the same run is a no-op. One writer per destination
is assumed, locally and in the bucket: two nightly writers updating the
manifest at the same moment could lose an entry.

## Scenarios

A scenario is a yaml file: a `suite` key naming a registered suite, plus
knobs the suite understands. `workload-burst` rejects any knob it does not
know, so a misspelled key fails the scenario before anything is created
instead of silently falling back to a default; `selfcheck` accepts anything
because echoing its knobs is its job:

```yaml
suite: selfcheck
message: hello from flowmark
```

## Suites

`selfcheck` touches nothing and needs no environment. It stays the
zero-dependency check that the yaml → runner → suite → results wiring works.

`workload-burst` is the scale suite. It creates one Run:ai project, then for
each round size N in the yaml it opens a Kubernetes watch, burst-creates N
workloads, measures how long each one takes to appear on the cluster, and
tears the round down before the next one starts. The yaml decides which flow
is measured:

| `workloadType` | what is created | what the watch waits for | `flow` in the section |
|---|---|---|---|
| `Deployment` | plain Deployments through the Kubernetes API | the `ExternalWorkload` Run:ai creates for each one | `deployment_to_external_workload` |
| `Training` | Trainings through the Run:ai control-plane API | the `WorkloadRunner` that reaches the cluster | `training_to_workload_runner` |

It needs `--cp-url`, `--kubeconfig` and credentials in `FLOWMARK_CP_USER` /
`FLOWMARK_CP_PASSWORD` (or `FLOWMARK_CP_TOKEN`):

```
FLOWMARK_CP_USER=... FLOWMARK_CP_PASSWORD=... ./bin/flowmark \
  --scenarios ./scenarios --output ./out \
  --cp-url https://my-control-plane.example.com \
  --kubeconfig ~/.kube/config
```

Two scenarios ship with it: `scenarios/external-workload.yaml` (Deployments)
and `scenarios/training-sync.yaml` (Trainings). Knobs:

| knob | default | meaning |
|---|---|---|
| `workloads` | `[100, 250, 500]` | round sizes, each unique; each round creates that many workloads |
| `workloadType` | `Deployment` | `Deployment` or `Training` |
| `podsPerWorkload` | 1 | replicas per Deployment, 0 to 10; must be 1 for `Training` |
| `image` | `registry.k8s.io/pause:3.9` | container image of every workload; point it at a mirror on air-gapped clusters |
| `submitConcurrency` | 16 | parallel create and delete calls, 1 to 64 |
| `kubeRequestsPerSecond` | 50 | Kubernetes client rate limit, burst is four seconds' worth |
| `controlPlaneRequestsPerSecond` | 25 | control-plane client rate limit, burst is four seconds' worth |
| `timeout` | `15m` | how long a round waits for its workloads to reach the cluster |
| `setupTimeout` | `5m` | project readiness, quiesce between rounds, cleanup |
| `settle` | `90s` | extra quiet time after a round's objects are gone, before the next round; not paid after the last round |
| `slo` | *(none)* | mapping of whole-number milliseconds copied into the section for the reader and for the later regression check; `p90_ms` also appears as `slo_p90_ms`. Informational: the suite never fails a round for missing it |
| `cluster` | *(single cluster)* | cluster name when the control plane has several |

Timing comes from a watch opened before anything is submitted: each
workload's duration is the gap, on FlowMark's own clock, between the create
response and the first event for its cluster object, matched by the
Deployment UID or the workload name. Deployments are deleted through the
Kubernetes API and Trainings through the control plane; the suite then waits
until the namespace holds no workload objects and no pods at all, pauses for
`settle`, and only then starts the next round. An empty namespace is not yet
a quiet cluster right after hundreds of pods were torn down.

Per round the section reports `p50_ms@N`, `p90_ms@N`, `p99_ms@N`,
`mean_ms@N`, `max_ms@N`, `throughput_per_sec@N`, `arrival_makespan_ms@N`,
`submit_makespan_ms@N`, `quiesce_ms@N`, `sample_size@N`, `pending_count@N`,
`failed_count@N`, `unmatched_arrivals@N` and `success_rate@N`, plus `flow`,
`rounds`, `workload_type`, `pods_per_workload`, `image`,
`submit_concurrency`, `kube_requests_per_sec`,
`control_plane_requests_per_sec`, `settle_ms` and the copied `slo`.
Workloads that never reached the cluster within the timeout are counted in
`pending_count`, never as zero latency, and fail the section; when nothing
arrived the latency keys are omitted entirely; a round that cannot quiesce
stops the run. `unmatched_arrivals@N` counts cluster objects that appeared
but matched no submitted workload, which is how a run that ends all pending
tells a missing correlation field from a slow cluster. Every object carries
`flowmark/run`, `flowmark/round` (the round index) and `flowmark/round-size`
labels, so a pod on the cluster can be tied to a metric key at a glance.
`submit_makespan_ms@N` shows how much of each round's window was FlowMark's
own submission, which the concurrency and rate knobs bound; raising
`submitConcurrency` alone changes nothing once the rate limits are saturated.
