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

Runs every `*.yaml` scenario in the scenarios directory and writes all
sections to `out/results.json`. Exit code 0 means every scenario passed;
1 means any error: a failed scenario (the results file is still written),
invalid scenarios or scenario files, or a failure writing the results.

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

Its section reports `sync_to_cluster_ms` per workload, `sync_to_cluster_max_ms`,
and `pending`: workloads that never reached the cluster within the timeout are
counted as pending, never as zero, and make the section fail.
