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

`basic-flow` runs against a real Run:ai environment: it creates a project,
submits workloads, measures how long each one takes to reach the cluster,
and deletes only what it created. It needs `--cp-url`, `--kubeconfig` and
credentials in `FLOWMARK_CP_USER` / `FLOWMARK_CP_PASSWORD` (or
`FLOWMARK_CP_TOKEN`):

```
FLOWMARK_CP_USER=... FLOWMARK_CP_PASSWORD=... ./bin/flowmark \
  --scenarios ./scenarios --output ./out \
  --cp-url https://my-control-plane.example.com \
  --kubeconfig ~/.kube/config
```

Knobs of `scenarios/basic-two-workloads.yaml`:

| knob | default | meaning |
|---|---|---|
| `projects` | 1 | projects to create; workloads are spread across them |
| `workloads` | 2 | workloads to submit |
| `workloadType` | `Training` | only `Training` for now |
| `timeout` | `5m` | project readiness and workload arrival limit |
| `cluster` | *(single cluster)* | cluster name when the control plane has several |

Timing comes from a Kubernetes watch opened before anything is submitted:
each workload's duration is the gap, on FlowMark's own clock, between the
control plane accepting it and its `WorkloadRunner` appearing on the
cluster. The section reports `sync_to_cluster_ms` per workload plus
`p50_ms`, `p90_ms`, `p99_ms`, `mean_ms`, `max_ms`, `makespan_ms`,
`throughput_per_s`, `success_rate`, `sample_size`, `pending_count` and
`failed_count`. Workloads that never reached the cluster within the timeout
are counted in `pending_count`, never as zero latency, and make the section
fail; when nothing arrived the latency keys are omitted entirely.

`external-workload` is the first PRD scenario: it burst-creates plain
Kubernetes Deployments in a Run:ai project namespace and measures how long
Run:ai's external-workload-integrator takes to create the matching
`ExternalWorkload` for each one, for a growing number of Deployments. It
needs the same flags and credentials as `basic-flow`.

Knobs of `scenarios/external-workload.yaml`:

| knob | default | meaning |
|---|---|---|
| `workloads` | `[100, 250, 500]` | round sizes, each unique; each round creates that many Deployments |
| `workloadType` | `Deployment` | only `Deployment` for now |
| `podsPerWorkload` | 1 | replicas per Deployment, 0 to 10 |
| `timeout` | `15m` | how long a round waits for its ExternalWorkloads |
| `setupTimeout` | `5m` | project readiness, quiesce between rounds, cleanup |
| `settle` | `90s` | extra quiet time after a round's objects are gone, before the next round |
| `slo` | *(none)* | mapping of whole-number milliseconds copied into the section; `create_to_ew_p90_ms` also appears as `create_to_ew_slo_ms` |
| `cluster` | *(single cluster)* | cluster name when the control plane has several |

Rounds are isolated: the watch is opened before the first create, the
ExternalWorkload is matched to its Deployment by the UID the API returned,
and after each round the Deployments are deleted and the suite waits until
the namespace holds no ExternalWorkloads and no pods of that round, then
pauses for `settle`, before the next round starts. The pause exists because
an empty namespace is not yet a quiet cluster right after hundreds of pods
were torn down. Per round the section reports
`create_to_ew_p50_ms@N`, `p90`, `p99`, `max`, `throughput_per_sec@N`,
`ew_makespan_ms@N`, `submit_makespan_ms@N`, `quiesce_ms@N`,
`sample_size@N`, `pending_count@N`, `failed_count@N` and
`success_rate@N`. Deployments whose ExternalWorkload never appeared count
as pending, never as zero, and fail the section; a round that cannot
quiesce stops the run.
