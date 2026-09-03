# KAI Scheduler dependency

Run:ai installs the upstream KAI Scheduler Helm chart and uses its Go APIs
directly. All three dependency coordinates are intentionally pinned together in
`kai-version.yaml`:

- the OCI Helm chart in `cluster/runai-chart/Chart.yaml`;
- the Go module used by cluster components;
- the documented upstream KAI release.

The KAI chart's default queue, default shard, and config deployer are disabled.
The Run:ai operator owns `kai-config`, and the nodepool controller owns one
native KAI `SchedulingShard` per Run:ai node pool.

Use `./update-kai-version.sh v0.16.1` to pin a tagged release. Use
`./update-kai-version.sh main` to resolve the latest KAI main commit, its Go
pseudo-version, and the matching `0.0.0-<short-sha>` OCI chart tag. Pass a
12-character or full commit SHA as the same single argument to reproduce a
specific KAI build.
