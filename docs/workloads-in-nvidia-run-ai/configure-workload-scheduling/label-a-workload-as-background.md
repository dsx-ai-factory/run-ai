# Label a Workload as Background

Label a workload so that the NVIDIA Run:ai Scheduler treats it as a background workload. For what background workloads are and how to create the project that runs them, refer to [Background workloads](../../platform-management/runai-scheduler/scheduling/background-workloads.md).

## Label the Workloads

Add the `kai.scheduler/background` label to each workload that should be treated as a background workload, and submit it under the project created for background workloads.

| Field | Value | Notes |
| --- | --- | --- |
| `kai.scheduler/background` | `"true"` | Identifies the workload as a background workload. Set on the workload. |
| Termination grace period (`terminationGracePeriodSeconds`) | `5` | Recommended. Controls how long the workload has to shut down when evicted. The Kubernetes default is 30 seconds. |

The following are handled for you and do not need to be set manually:

* **Project** - Derived from the project the workload is submitted under. There is no need to set a queue name.
* **Scheduler name** - Set automatically by the NVIDIA Run:ai platform.

### NVIDIA Run:ai Native Workloads

For [native NVIDIA Run:ai workloads](../workload-types/workload-types.md), the label can be set during workload submission using one of the following methods:

* **UI** - Under **General** settings, click **+LABEL** in **Set label(s)**. In **Name**, enter `kai.scheduler/background`, and in **Value**, enter `true`
* **API** - Set using the `labels` field, with `kai.scheduler/background` as the `name` and `true` as the `value`
* **CLI** - Set using the `--label` flag, for example `--label kai.scheduler/background=true`

### Supported Workload Types

For [supported workload types](../supported-workload-types/list-of-supported-workload-types.md) submitted with a YAML manifest, add the label to your YAML manifest under the `metadata.labels` section of your workload definition:

```yaml
metadata:
  labels:
    kai.scheduler/background: "true"
spec:
  terminationGracePeriodSeconds: 5
```

## Deployment Considerations

* **Resource types** - Background workloads can use CPU or GPU resources. GPU background workloads must request whole GPUs.
* **Preemptibility** - Background workloads must be preemptible. Any [workload priority](../../platform-management/runai-scheduler/scheduling/workload-priority-control.md) can be used, as long as the workload's preemptibility is set to preemptible. Non-preemptible and semi-preemptible workloads require in-quota resources, and a background workload project has no quota, so those workloads are never scheduled.
* **Termination grace period** - Set a short termination grace period. A background workload is evicted whenever an AI workload needs its resources. A long grace period delays the AI workload while the background workload shuts down. A short period, such as 5 seconds, keeps that delay small.
* **Controller** - Use a controller that recreates pods. Any workload type can be a background workload. Because background workloads are evicted as a normal part of operation, submit them through a controller, such as a DaemonSet or Deployment, that recreates the pod after eviction.
