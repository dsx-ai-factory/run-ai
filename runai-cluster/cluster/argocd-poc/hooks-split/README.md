<!--
SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# runai-cluster hooks as separate Argo CD Applications

A proof of concept for running the `runai-cluster` chart's lifecycle hooks apart
from the release. The hook-handler Jobs need the short-lived bootstrap token,
and Argo CD runs hooks on every sync, so any sync later than the token's
lifetime fails. Here the hooks run only when something asks for them.

The chart option is `hooks.render` (`all`, `none`, `only`) with `hooks.phase`
(`all`, `pre`, `post`, `delete`). See `values.yaml`. The default renders the same
manifests as before.

## Layout

| Path | What it is |
|---|---|
| `apps/runai-cluster-pre.yaml` | Pre hooks only. Sync wave -30. |
| `apps/runai-cluster.yaml` | The release with `hooks.render: none`. Sync wave -20. |
| `apps/runai-cluster-post.yaml` | Post hooks only. Sync wave -10. |
| `uninstall/runai-cluster-uninstall.yaml` | Pre-delete hook only. Applied at uninstall. |
| `hooks-run/*/hooks-run.yaml` | The ConfigMap each hooks Application tracks. |
| `values/runai-cluster.yaml` | Values shared by all four. No credentials. |

An Application that contains only hooks never goes OutOfSync, because Argo CD
does not track hook resources. Each hooks Application therefore also tracks a
`hooks-run` ConfigMap. Changing its `runId` starts a sync and runs the hooks.
Nothing else does, and self-heal is off for those Applications.

## Flows

**Install.** Create the cluster in the control plane and write the cluster uid
and bootstrap token to the Secret named in `values/runai-cluster.yaml`. Then
commit the three Applications with a new `runId`. Argo CD runs pre, the release,
then post, in wave order.

**Upgrade.** Refresh the Secret, then change `version` and `runId` in both
`hooks-run` files and `<CLUSTER_VERSION>` in the Applications. The release can
be synced or self-healed at any time without the token.

**Uninstall.** Apply `uninstall/runai-cluster-uninstall.yaml` with a new
`runId` and wait for it to succeed. Then delete the `runai-cluster` Application,
and remove the Secret and namespaces.

## Limits

- The chart has no separate install and upgrade mode, so a pre run renders both
  the pre-install and pre-upgrade Jobs, as Argo CD already does today.
- Hooks rendered by the `kai-resource-management` and `karta` subcharts are not
  controlled by `hooks.render`. They stay with the release Application. Disable
  both subcharts for the hooks Applications, as these templates do.
- The three Applications share one Helm release name but render disjoint sets
  of resources. The `hooks-params` ConfigMap goes to the post phase only, since
  only the post Jobs mount it.
- Not tested on a cluster.
