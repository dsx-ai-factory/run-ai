# Configure CPU and Memory Requests

Request CPU compute and memory for a workload, and set limits. For how the allocation works and the cluster defaults, refer to [CPU compute and memory allocation](../resource-optimization/allocation-of-cpu-and-memory.md).

## Requesting CPU Compute and Memory

When submitting a [workload](../../../workloads-in-nvidia-run-ai/workloads.md) or creating [compute resources](../../../workloads-in-nvidia-run-ai/workload-assets/compute-resources.md), you can explicitly request both CPU compute and memory resources. The system guarantees that, if the workload is scheduled, the requested resources will be available for that workload.

## Setting CPU Compute and Memory Limits

You can further control CPU compute and memory resource usage by specifying limits. The system will ensure your workload does not consume more than the limits set for compute or memory.

* If your workload exceeds its memory limit, it will get an out of memory error and may be terminated.
* The limit must be equal to or greater than the request.

## Validating CPU Resource Allocations

Once a workload is submitted, its resource requests and limits are reflected in the underlying Kubernetes pod. Review pod specifications in Kubernetes to check the assigned CPU and memory.

1. Get the pod name by running: `runai workload describe WORKLOAD_NAME`
2. The pod will appear under the `PODS` category. Run: `kubectl describe pod <POD_NAME>`

The information will appear under `Requests` and `Limits`. For example:

```yaml
Limits:
    nvidia.com/gpu:  2
Requests:
    cpu:             1
    memory:          104857600
    nvidia.com/gpu:  2
```
