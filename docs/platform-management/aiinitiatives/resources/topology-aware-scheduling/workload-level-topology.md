# Workload-Based Topology Placement

This page describes topology-aware scheduling applied on a workload-based scope - the default scope, in which NVIDIA Run:ai treats the entire distributed workload as a single unit for placement. All pods in the workload are subject to the same topology constraints. To further co-locate workers in smaller, independently placed groups within the workload (for example, tensor-parallel training jobs), see [Segment-Based Topology Placement](segment-based-topology-placement.md) - segment-based constraints can be combined with workload-based constraints on the same workload; see [Combining Workload-Based and Segment-Based Placement](segment-based-topology-placement.md#combining-workload-based-and-segment-based-placement).

## Automatic Topology-Aware Scheduling

When a distributed workload is submitted, the platform automatically applies topology-aware scheduling based on the network topology configured on the target node pool. This behavior ensures that distributed workloads benefit from improved performance without additional user configuration.

By default, topology-aware scheduling in NVIDIA Run:ai is applied at the **workload level**. This means the Scheduler considers the entire distributed workload as a single unit and places all of its pods according to the same topology constraints.

NVIDIA Run:ai automatically applies a **Preferred** topology constraint at the lowest defined topology level. This co-locates pods as close as possible in the network hierarchy, reducing communication overhead. If the Scheduler cannot place pods at this level, it automatically escalates placement by moving up through the topology hierarchy (for example: from node → rack → block → zone), always seeking the closest available level to minimize latency.

This behavior applies to distributed [native workloads](../../../../workloads-in-nvidia-run-ai/workload-types/native-workloads.md), [supported workload types](../../../../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md), and workloads submitted via `kubectl`. To override the default behavior and define your own topology-related annotations when submitting a distributed workload, see [Fine-tuning Topology-Aware Scheduling per Workload](#fine-tuning-topology-aware-scheduling-per-workload) below.

## Fine-tuning Topology-Aware Scheduling per Workload

You can customize how NVIDIA Run:ai applies topology labels for each distributed workload, allowing you to override the default scheduling behavior when needed.

Use the following topology-aware scheduling annotations when submitting a distributed workload:

* `kai.scheduler/topology` - Specifies the name of the topology to use.\
  The value must match the name of a topology entity defined for the node pool.
* `kai.scheduler/topology-required-placement` and/or `kai.scheduler/topology-preferred-placement` - Define placement constraints using a topology label key (for example, `"rack"`).
  * Required - Enforces strict placement. All pods must be scheduled within the specified topology level.
  * Preferred - Expresses a soft preference. The Scheduler attempts to place pods within the specified topology level when possible.

You can use Required, Preferred, or both together for the same topology tree. When combining them, keep in mind that network topologies are hierarchical (tree-structured):

* A Preferred constraint defined at the same level as, or higher than, a Required constraint has no effect.
* A Preferred constraint is effective only when it is defined at a lower (more specific) topology level than the Required constraint.
* In this case, the Scheduler enforces the Required constraint while attempting to further group pods according to the Preferred constraint.

**How to apply these annotations:**

* **Native distributed workloads** - Define the topology annotations in the workload’s annotations field through the NVIDIA Run:ai UI, API, or CLI.
* **Supported workload types submitted via YAML** - Specify the topology-related annotations directly in the YAML manifest. NVIDIA Run:ai respects the user-defined annotations and does not override them.
* **Workloads submitted via `kubectl`** - Add annotations to the workload manifest by specifying the topology name and constraint type to ensure the Scheduler respects the desired topology.

For example:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: topology-aware-job
  annotations:
    kai.scheduler/topology: "network"
    kai.scheduler/topology-preferred-placement: "rack"
    kai.scheduler/topology-required-placement: "zone"
```

## LeaderWorkerSet (LWS) Behavior

LeaderWorkerSet (LWS) is a supported distributed workload type, receiving the same automatic topology-aware scheduling behavior described above. For LWS workloads, topology-aware scheduling is applied **per replica**:

* Each replica of a LeaderWorkerSet (as defined by `spec.replicas`) is gang scheduled.
* The same topology constraints are applied to each replica instead of the entire workload. This ensures that the leader and worker pods that belong to the same replica are placed according to the configured network topology (for example, within the same rack, block, or NVLink domain).

Example:

```yaml
apiVersion: leaderworkerset.x-k8s.io/v1
kind: LeaderWorkerSet
metadata:
  annotations:
    kai.scheduler/topology: "cluster-topology"
    kai.scheduler/topology-required-placement: "cloud.provider.com/topology-block"
    kai.scheduler/topology-preferred-placement: "cloud.provider.com/topology-rack"
  labels:
    runai/queue: test
  namespace: runai-test
  name: vllm
spec:
  replicas: 2
  leaderWorkerTemplate:
    size: 2
```

## Topology-Aware Scheduling for Dynamo over Grove

Dynamo disaggregated inference workloads require per-component topology-aware scheduling, since applying a single topology constraint across all pods can lead to suboptimal placement. NVIDIA Run:ai supports hierarchical, multi-component topology-aware scheduling for Dynamo over Grove workloads, enforcing constraints at both the workload (Deployment) and component (Service) levels.

<Note>

**Note**

Support for multiple topologies in the same cluster requires Dynamo 1.1.0 and Grove v0.1.0-alpha.9 or later. Starting from these versions, administrators can configure matching topologies in both NVIDIA Run:ai and Grove for each node pool, ensuring workloads are scheduled against the correct topology for the node pool where they run.

In earlier versions, only a single cluster-wide topology is supported - using multiple topologies in those versions may result in scheduling mismatches.

</Note>

Grove manages topologies independently and exposes them as the mechanism for requesting topology constraints in the Dynamo workload spec.

* **Administrator setup** - For each topology defined in NVIDIA Run:ai, the administrator creates a matching topology in Grove and references it to its corresponding KAI topology. Once linked, NVIDIA Run:ai surfaces the Grove topology name and levels to users through the [Network Topologies](api:main-api:GET/api/v1/network-topologies) API, without requiring them to understand the underlying Grove topology configuration.
* **Workload submission** - When submitting a Dynamo workload, the user selects the topology that matches the node pool on which the workload will run. If only a single topology exists in the cluster, it is always used. The Grove topology is surfaced through the [Network Topologies](api:main-api:GET/api/v1/network-topologies) API, allowing users to identify the correct topology for their target node pool.
* **Scheduling** - The NVIDIA Run:ai Scheduler enforces the topology constraints at both the workload (Deployment) and component (Service) levels.

If there is no matching topology for the required node pool, a mismatch can occur between what the user requests, what the Scheduler places, and what is visible on the workload. See [Topology constraints visibility](README.md#topology-constraints-visibility) for more details.
