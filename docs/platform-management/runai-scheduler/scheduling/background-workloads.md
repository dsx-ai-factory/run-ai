# Background Workloads

Clusters run infrastructure workloads alongside the AI workloads that use them, such as health checks, diagnostics, monitoring agents, and similar maintenance tasks. These workloads occupy node resources, so the nodes they run on appear partially full, and the capacity available to AI workloads is reduced.

Background workloads are workloads that the NVIDIA Run:ai Scheduler treats as invisible when it plans resources. The Scheduler places AI workloads as though the background workloads were not running, and evicts them on demand when their resources are needed. This keeps the full capacity of the cluster available to AI workloads, while still allowing infrastructure workloads to use resources that would otherwise sit idle.

## Benefits of Background Workloads

* **Full cluster capacity for AI workloads** - Infrastructure workloads do not reduce the resources the Scheduler can offer to AI workloads.
* **Useful work from idle resources** - Background workloads run on resources that no AI workload currently needs, instead of leaving them unused.
* **No effect on quota or fairness** - Background workloads are excluded from quota and fairshare calculations, so they never trigger over quota usage or resource reclaim between projects.

## How Background Workloads Work

At the start of each scheduling cycle, the Scheduler virtually evicts all background workloads and plans resources as if they were not running. Only after placing the AI workloads does the Scheduler offer the remaining resources back to the background workloads.

The outcome for each background workload depends on the result of that planning:

* A background workload that still fits on its original node keeps running and is not disturbed.
* A background workload whose resources are now needed by an AI workload is evicted from the cluster.

When a background workload is evicted, its controller (for example, a DaemonSet or Deployment) recreates the pod immediately. If the cluster has no spare capacity, the recreated pod remains in `Pending` state until resources free up. This is the expected steady state for background workloads, not an error.

Each eviction emits an event on the workload with the action `backgroundpods`, which identifies the eviction as a background workload eviction rather than preemption or resource reclaim. Use these events to distinguish expected background workload churn from scheduling problems.

Because background workloads are excluded from quota and fairshare calculations, they cannot push a project over quota, and they cannot cause resource reclaim between projects or departments. Refer to [The NVIDIA Run:ai Scheduler: Concepts and Principles](concepts-and-principles.md#quota) for how quota and fairshare are calculated for AI workloads.

## Setting Up Background Workloads

Background workloads require a dedicated project with no quota, and a label on each workload that identifies it to the Scheduler.

<Note>

**Note**

The background workloads capability is enabled by default and requires no configuration to start using it. Administrators can change its default behavior per node pool through the KAI Scheduler shard configuration. Refer to the [KAI Scheduler background pods documentation](https://github.com/kai-scheduler/KAI-Scheduler/blob/main/docs/background-pods/README.md) for the available settings.

</Note>

### Create a Project for Background Workloads

Create a dedicated [project](../../aiinitiatives/organization/projects.md) for your background workloads, and assign it the following quota settings on every node pool where background workloads run:

| Setting | Value | Purpose |
| --- | --- | --- |
| GPU quota | `0` | The project deserves no guaranteed resources. |
| CPU quota | `0` | The project deserves no guaranteed resources. |
| CPU memory quota | `0` | The project deserves no guaranteed resources. |
| Limit | Unlimited | The project can use any amount of spare capacity when it is available. |
| Over quota weight | `0` | The project receives no share of unused resources through the over quota mechanism. |

These settings mean the project is never entitled to resources of its own, but can use whatever is currently unused. Refer to [Projects](../../aiinitiatives/organization/projects.md) for how to create a project and assign its quota.

### Label the Workloads

Add the `kai.scheduler/background` label to each workload that should be treated as a background workload, and submit it under the project created for background workloads.

| Field | Value | Notes |
| --- | --- | --- |
| `kai.scheduler/background` | `"true"` | Identifies the workload as a background workload. Set on the workload. |
| Termination grace period (`terminationGracePeriodSeconds`) | `5` | Recommended. Controls how long the workload has to shut down when evicted. The Kubernetes default is 30 seconds. |

The following are handled for you and do not need to be set manually:

* **Project** - Derived from the project the workload is submitted under. There is no need to set a queue name.
* **Scheduler name** - Set automatically by the NVIDIA Run:ai platform.

#### NVIDIA Run:ai Native Workloads

For [native NVIDIA Run:ai workloads](../../../workloads-in-nvidia-run-ai/workload-types/native-workloads.md), the label can be set during workload submission using one of the following methods:

* **UI** - Under **General** settings, click **+LABEL** in **Set label(s)**. In **Name**, enter `kai.scheduler/background`, and in **Value**, enter `true`
* **API** - Set using the `labels` field, with `kai.scheduler/background` as the `name` and `true` as the `value`
* **CLI** - Set using the `--label` flag, for example `--label kai.scheduler/background=true`

#### Supported Workload Types

For [supported workload types](../../../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md) submitted with a YAML manifest, add the label to your YAML manifest under the `metadata.labels` section of your workload definition:

```yaml
metadata:
  labels:
    kai.scheduler/background: "true"
spec:
  terminationGracePeriodSeconds: 5
```

## Deployment Considerations

* **Resource types** - Background workloads can use CPU or GPU resources. GPU background workloads must request whole GPUs.
* **Preemptibility** - Background workloads must be preemptible. Any [workload priority](workload-priority-control.md) can be used, as long as the workload's preemptibility is set to preemptible. Non-preemptible and semi-preemptible workloads require in-quota resources, and a background workload project has no quota, so those workloads are never scheduled.
* **Termination grace period** - Set a short termination grace period. A background workload is evicted whenever an AI workload needs its resources. A long grace period delays the AI workload while the background workload shuts down. A short period, such as 5 seconds, keeps that delay small.
* **Pending state** - Expect background workloads to sit in `Pending` state. When the cluster is fully used by AI workloads, background workloads remain pending until resources free up. This is the intended behavior.
* **Controller** - Use a controller that recreates pods. Any workload type can be a background workload. Because background workloads are evicted as a normal part of operation, submit them through a controller, such as a DaemonSet or Deployment, that recreates the pod after eviction.

## Using CLI

To view the available actions, go to the [CLI v2 reference](../../../reference/cli/runai/runai.md) and run according to your workload.

## Using API

To view the available actions, go to the [API reference](../../../rest-api/about-the-rest-api.md) and run according to your workload.
