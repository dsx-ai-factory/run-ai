# KAI Scheduler dependency

Run:ai uses KAI Scheduler's Go APIs directly, and installs its Helm chart
indirectly: KAI arrives as a sub-dependency of `kai-resource-management` (KRM),
so KRM owns the chart pin and this repo no longer declares one. What
`kai-version.yaml` pins here is:

- the Go module used by cluster components;
- the documented upstream KAI release;
- the KRM chart version in `cluster/runai-chart/Chart.yaml`, under `krm:`.

The KAI chart's default queue, default shard, and config deployer are disabled.
The Run:ai operator owns `kai-config`, and the nodepool controller owns one
native KAI `SchedulingShard` per Run:ai node pool.

Use `./update-krm-version.sh v0.1.0` to pin a KRM release. The script reads
that release's chart metadata, takes the `kai-scheduler` version it bundles, and
re-pins our KAI Go module to match, so the two cannot drift apart. Modules that
reach KAI only through a local `replace` are tidied as well. It needs `go`,
`helm` and mikefarah's `yq` v4, and it edits nothing until both the KRM chart
and the KAI Go module resolve.

There is deliberately no way to bump KAI on its own: doing so would move our Go
types ahead of the scheduler KRM actually deploys, with nothing to catch it. To
take a newer KAI, release it in KRM first.

The nightly `Nightly KRM Update` workflow runs the same script against the
latest published KRM release.
