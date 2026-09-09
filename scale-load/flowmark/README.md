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

## Scenarios

A scenario is a yaml file: a `suite` key naming a registered suite, plus
knobs the suite understands:

```yaml
suite: selfcheck
message: hello from flowmark
```
