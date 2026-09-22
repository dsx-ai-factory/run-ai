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

Runs every `*.yaml` scenario in the scenarios directory and writes
`out/results.json`. Exit code 0 means every scenario passed;
1 means any error: a failed scenario (the results file is still written),
invalid scenarios or scenario files, or a failure writing the results.

The file has three top-level keys: `schema` (currently `v1`), `metadata`
(`started_at`, `finished_at`, and `clock_offset_ms` when a kubeconfig was
given), and `scenarios`, one section per scenario file. `clock_offset_ms`
is how far the cluster clock is ahead of the machine running FlowMark,
measured once at startup with a probe ConfigMap in the `default`
namespace. It is reported for context only and never enters a latency. If
the probe fails, the run continues and `clock_offset_error` holds the reason.

`./bin/flowmark --help` lists all flags.

Suites that talk to a Run:ai environment take the cluster from
`--kubeconfig`, the control plane from `--cp-url`, and credentials from
environment variables: `FLOWMARK_CP_USER` and `FLOWMARK_CP_PASSWORD`, or
a pre-issued `FLOWMARK_CP_TOKEN`. Credentials are never accepted as
flags and never appear in output. The selfcheck scenario needs none of
these.

## Scenarios

A scenario is a yaml file: a `suite` key naming a registered suite, plus
knobs the suite understands:

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
| `timeout` | `15m` | how long a round waits for its workloads to reach the cluster |
| `setupTimeout` | `5m` | project readiness, quiesce between rounds, cleanup |
| `settle` | `90s` | extra quiet time after a round's objects are gone, before the next round |
| `slo` | *(none)* | mapping of whole-number milliseconds copied into the section; `p90_ms` also appears as `slo_p90_ms` |
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
`max_ms@N`, `throughput_per_sec@N`, `arrival_makespan_ms@N`,
`submit_makespan_ms@N`, `quiesce_ms@N`, `sample_size@N`, `pending_count@N`,
`failed_count@N` and `success_rate@N`, plus `flow`, `rounds`,
`workload_type`, `pods_per_workload`, `settle_ms` and the copied `slo`.
Workloads that never reached the cluster within the timeout are counted in
`pending_count`, never as zero latency, and fail the section; when nothing
arrived the latency keys are omitted entirely; a round that cannot quiesce
stops the run. Submission to the control plane is paced at 25 requests per
second and to the Kubernetes API at 50, so `submit_makespan_ms@N` shows how
much of each round's window was FlowMark's own submission.
