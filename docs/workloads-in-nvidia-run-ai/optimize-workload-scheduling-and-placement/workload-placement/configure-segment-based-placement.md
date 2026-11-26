# Configure Segment-Based Placement

Divide the workers of a distributed workload into segments and set the topology constraint for each segment. For how segment-based placement works, refer to [Segment-based topology placement](../../../platform-management/aiinitiatives/resources/topology-aware-scheduling/segment-based-topology-placement.md).

## Fine-tuning Topology-Aware Scheduling per Segment

Use the following annotations to define segment-based topology constraints:

* `kai.scheduler/topology` - Name of the topology to use. This only needs to be set once - if it is set at the workload root and omitted from the segment's annotations, the segment inherits the root's topology name. If it is set in both places, the segment-based (PodTemplate) value takes precedence.
* `kai.scheduler/segment-size` - Number of pods per segment. Must be greater than 1 and not greater than the replica count. For native distributed training workloads, set this through the workers-per-segment field when submitting the workload via the UI, API, or CLI instead of this annotation.
* `kai.scheduler/segment-topology-required-placement` - Topology level all pods in a segment must share (hard constraint).
* `kai.scheduler/segment-topology-preferred-placement` - Topology level pods in a segment should share when possible (soft constraint).

**How to apply these annotations:**

* **Native distributed training workloads** - Supported for the PyTorch framework only. Set the number of workers per segment when submitting the [distributed training workload](../../using-training/distributed-training-models.md).
* **Supported workload types submitted via YAML** - Supported for PyTorchJob and LeaderWorkerSet (LWS). Add the annotations above to the `Worker` template (PyTorchJob) or the `workerTemplate` (LeaderWorkerSet).
* **Workloads submitted via `kubectl`** - Topology-aware scheduling is not applied automatically by NVIDIA Run:ai for workloads submitted this way. Add the same annotations directly to the `Worker` template (PyTorchJob) or `workerTemplate` (LeaderWorkerSet) in the workload manifest to ensure the Scheduler respects the desired segment-based placement.

For example:

```yaml
apiVersion: kubeflow.org/v1
kind: PyTorchJob
metadata:
  name: pytorch-tp4
spec:
  pytorchReplicaSpecs:
    Master:
      replicas: 1
      template:
        spec:
          containers:
            - name: pytorch
              image: pytorch/pytorch:latest
    Worker:
      replicas: 16
      template:
        metadata:
          annotations:
            kai.scheduler/topology: "cluster-topology"
            kai.scheduler/segment-size: "4"
            kai.scheduler/segment-topology-required-placement: "cloud.provider.com/topology-rack"
        spec:
          containers:
            - name: pytorch
              image: pytorch/pytorch:latest
```

## Combining Workload-Based and Segment-Based Placement

Workload-based and segment-based placement are not mutually exclusive - you can apply both to the same workload. A common pattern is to enforce a broader [workload-based](../../../platform-management/aiinitiatives/resources/topology-aware-scheduling/workload-level-topology.md) constraint that keeps the entire workload within one topology level (for example, a zone), while a tighter segment-based constraint co-locates each segment within a more specific level (for example, a rack) inside that zone.

For example:

```yaml
apiVersion: kubeflow.org/v1
kind: PyTorchJob
metadata:
  name: pytorch-tp4
  annotations:
    kai.scheduler/topology: "cluster-topology"
    kai.scheduler/topology-required-placement: "topology.kubernetes.io/zone"
spec:
  pytorchReplicaSpecs:
    Master:
      replicas: 1
      template:
        spec:
          containers:
            - name: pytorch
              image: pytorch/pytorch:latest
    Worker:
      replicas: 16
      template:
        metadata:
          annotations:
            kai.scheduler/segment-size: "4"
            kai.scheduler/segment-topology-required-placement: "cloud.provider.com/topology-rack"
        spec:
          containers:
            - name: pytorch
              image: pytorch/pytorch:latest
```

Here, the workload-based `kai.scheduler/topology-required-placement` pins the entire job to a single zone, while the segment-based `kai.scheduler/segment-topology-required-placement` further co-locates each 4-worker segment within the same rack. The `kai.scheduler/topology` name is set once at the workload root and inherited by the Worker segment.

## LeaderWorkerSet (LWS) Behavior

Segment size for LWS can be set either through the `kai.scheduler/segment-size` annotation or through the LWS-native `spec.leaderWorkerTemplate.subGroupPolicy.subGroupSize` field. If both are set, the annotation takes precedence.

The LWS `subGroupPolicyType` determines how the leader pod is grouped relative to segments:

* `LeaderWorker` - The leader pod groups together with the workers in the first segment.
* `LeaderExcluded` - The leader pod is excluded from segmentation. `(size - 1)` must be evenly divisible by `subGroupSize`.

For example:

```yaml
apiVersion: leaderworkerset.x-k8s.io/v1
kind: LeaderWorkerSet
metadata:
  name: lws-segmented
spec:
  replicas: 1
  leaderWorkerTemplate:
    size: 5
    subGroupPolicy:
      subGroupSize: 2
      subGroupPolicyType: LeaderWorker
    leaderTemplate:
      spec:
        containers:
          - name: leader
            image: my/image:latest
    workerTemplate:
      metadata:
        annotations:
          kai.scheduler/topology: "cluster-topology"
          kai.scheduler/segment-topology-required-placement: "cloud.provider.com/topology-rack"
      spec:
        containers:
          - name: worker
            image: my/image:latest
```
