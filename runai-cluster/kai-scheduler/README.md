# KAI Scheduler dependency

Run:ai uses KAI Scheduler's Go APIs directly, and installs its Helm chart
indirectly: KAI arrives as a sub-dependency of `kai-resource-management` (KRM),
so KRM owns the chart pin and this repo no longer declares one.

The KAI chart's default queue, default shard, and config deployer are disabled.
The Run:ai operator owns `kai-config`, and the nodepool controller owns one
native KAI `SchedulingShard` per Run:ai node pool.

## Which KRM ships

`cluster/runai-chart/Chart.yaml` declares KRM as a version range (`~0.18.0`),
not an exact version, so every build takes the newest KRM patch of that minor.
A KRM patch therefore reaches a release without a pull request here, and nothing
in this repository records which one shipped.

What shipped is recorded on the release instead: each GitHub Release carries a
`subcharts.yaml` asset listing the resolved sub-chart versions, written from the
packaged chart at build time. That file is the source of truth for airgap
packaging and for answering "which KRM is in 3.0.20?".

Moving to the next minor means editing the range in `Chart.yaml` by hand.
Renovate deliberately does not touch this dependency: a range has no pin to
bump, and Renovate would rewrite it back to an exact version.

## Keeping the Go module in step

The chart floats, but the Go code cannot: cluster components compile against
KAI's Go module, and those types must not drift ahead of the scheduler that is
actually deployed.

```bash
./update-krm-version.sh v0.18.6
```

The script reads that KRM release's chart metadata, takes the `kai-scheduler`
version it bundles, and re-pins our KAI Go module to match across every module
that names it. It does not touch `Chart.yaml`. It needs `go`, `helm` and
mikefarah's `yq` v4, and it edits nothing until both the KRM chart and the KAI
Go module resolve.

Pass the newest KRM release the range covers, so the Go module matches what a
build would resolve today.

There is deliberately no way to bump KAI on its own: doing so would move our Go
types ahead of the scheduler KRM actually deploys, with nothing to catch it. To
take a newer KAI, release it in KRM first.
