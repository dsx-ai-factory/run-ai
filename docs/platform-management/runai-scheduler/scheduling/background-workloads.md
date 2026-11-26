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

Add the `kai.scheduler/background` label to each background workload, and submit it under the project created for background workloads. For the label, the ways to set it, and the considerations for background workloads, refer to [Label a workload as background](../../../workloads-in-nvidia-run-ai/optimize-workload-scheduling-and-placement/label-a-workload-as-background.md).

## Deployment Considerations

* **Pending state** - Expect background workloads to sit in `Pending` state. When the cluster is fully used by AI workloads, background workloads remain pending until resources free up. This is the intended behavior.

## Using CLI

To view the available actions, go to the [CLI v2 reference](../../../reference/cli/runai/runai.md) and run according to your workload.

## Using API

To view the available actions, go to the [API reference](../../../rest-api/about-the-rest-api.md) and run according to your workload.
