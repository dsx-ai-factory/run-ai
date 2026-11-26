# Segment-Based Topology Placement

For distributed training workloads, you can divide worker pods into fixed-size topology-aware segments, with each segment independently co-located within the same topology level. This is useful for workloads where groups of workers communicate more frequently within the group than across groups - for example, tensor-parallel training jobs where intra-group communication benefits from tight locality. Segment-based placement can be combined with [workload-based placement](workload-level-topology.md) on the same workload; see [Combining Workload-Based and Segment-Based Placement](../../../runai-scheduler/optimize-workload-scheduling-and-placement/workload-placement/configure-segment-based-placement.md#combining-workload-based-and-segment-based-placement) below.

<Note>

**Note**

Segment-based placement is currently supported for the PyTorch framework (native distributed training workloads), PyTorchJob, and LeaderWorkerSet (LWS) ([supported workload types](../../../../workloads-in-nvidia-run-ai/supported-workload-types/list-of-supported-workload-types.md)).

</Note>

## Automatic Segment-Based Placement

When a segment size is configured for a distributed workload, either through the `kai.scheduler/segment-size` annotation or the native workload's workers-per-segment field, NVIDIA Run:ai automatically applies topology-aware scheduling to each segment. This behavior ensures segments benefit from improved intra-group performance without additional user configuration beyond setting the segment size.

By default, NVIDIA Run:ai automatically applies a **Preferred** topology constraint at the lowest defined topology level, scoped to each individual segment. This is applied in addition to the existing workload-level topology automation. This co-locates the pods within a segment as close as possible in the network hierarchy, reducing communication overhead between workers that belong to the same segment. If the Scheduler cannot place a segment's pods at this level, it automatically escalates placement by moving up through the topology hierarchy (for example: from node → rack → block → zone), always seeking the closest available level to minimize latency for that segment.

This automatic behavior applies to distributed native training workloads and PyTorchJob. It does not apply to LeaderWorkerSet (LWS) workloads - see [LeaderWorkerSet (LWS) Behavior](../../../runai-scheduler/optimize-workload-scheduling-and-placement/workload-placement/configure-segment-based-placement.md#leaderworkerset-lws-behavior) below. To override the default behavior and define your own segment-based topology annotations when submitting a distributed workload, see [Fine-tuning Topology-Aware Scheduling per Segment](#fine-tuning-topology-aware-scheduling-per-segment) below.

This automatic behavior does not apply to workloads submitted via `kubectl`. For those workloads, you must define the segment-based topology annotations yourself - see [Fine-tuning Topology-Aware Scheduling per Segment](#fine-tuning-topology-aware-scheduling-per-segment) below.

## Fine-tuning Topology-Aware Scheduling per Segment

You can set segment-based topology constraints for each distributed workload with annotations when you submit it. This includes combining them with workload-based constraints and the behavior for LeaderWorkerSet workloads. For the annotations and examples, refer to [Configure segment-based placement](../../../runai-scheduler/optimize-workload-scheduling-and-placement/workload-placement/configure-segment-based-placement.md).
