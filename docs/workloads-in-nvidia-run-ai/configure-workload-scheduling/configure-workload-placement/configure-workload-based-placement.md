# Configure Workload-Based Placement

Override the automatic topology-aware scheduling for a distributed workload. For how topology-aware scheduling works and how administrators configure it, refer to [Network topology-aware scheduling](../../../platform-management/aiinitiatives/resources/topology-aware-scheduling/README.md) and [Workload-based topology placement](../../../platform-management/aiinitiatives/resources/topology-aware-scheduling/workload-level-topology.md).

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
