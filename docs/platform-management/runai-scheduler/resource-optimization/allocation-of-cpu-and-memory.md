# CPU Compute and Memory Allocation

When allocating compute resources for workloads, GPUs are often seen as the main bottleneck. However, CPU compute and CPU memory are equally important:

* **CPU compute** - Essential for tasks such as data preprocessing and post-processing during training.
* **CPU memory** - Directly affects batch sizes and the volume of data your training run can handle efficiently.

Modern GPU servers typically include substantial CPU compute and memory to support these needs.

## Requesting CPU Compute and Memory

When submitting a [workload](../../../workloads-in-nvidia-run-ai/workloads.md) or creating [compute resources](../../../workloads-in-nvidia-run-ai/workload-assets/compute-resources.md), you can explicitly request both CPU compute and memory resources. The system guarantees that, if the workload is scheduled, the requested resources will be available for that workload.

The number of CPUs and amount of memory your workload will receive is guaranteed to be the number requested. In practice, however, you may receive more resources than requested:

* If the workload you submitted is the only workload running on a node, it can utilize all available CPU compute and memory on that node until another workload is scheduled.
* When another workload is submitted, each workload will receive a number of compute and memory proportional to the number requested. For example, if the first workload requests 1 CPU and the second requests 3 CPUs on a node with 40 CPUs, the workloads will receive 10 and 30 CPUs respectively.
* If CPU compute and memory are not explicitly requested, resources are assigned from the [cluster default](allocation-of-cpu-and-memory.md#cluster-default-resource-settings).

<Note>

**Note**

If your workload temporarily uses more memory than requested and new workloads are scheduled, it may be forced to release memory, potentially resulting in an out of memory (OOM) error.

</Note>

## Setting CPU Compute and Memory Limits

You can further control CPU compute and memory resource usage by specifying limits. The system will ensure your workload does not consume more than the limits set for compute or memory.

* If your workload exceeds its memory limit, it will get an out of memory error and may be terminated.
* The limit must be equal to or greater than the request.

## Cluster Default Resource Settings

If CPU compute and memory resource requests and/or limits are not explicitly provided, the system applies cluster-wide defaults.

### CPU Compute Requests

* **If GPUs are requested** - The default CPU allocation is determined by a set ratio of CPUs per GPU.\
  For example, with a default ratio of 1:6 and a request for 2 GPUs, your job will be assigned 12 CPUs (2 GPUs × 6 CPUs each).
* **If no GPUs are requested** - The default CPU allocation is determined by a ratio of CPUs per CPU limit. For example, with a default ratio of 1:0.2 and a CPU limit of 10, your job will be assigned 2 CPUs (10 × 0.2).
* **System defaults** - The out-of-the-box default is 1:1 (1 CPU per GPU requested) and 1:0.1 (0.1 CPUs per CPU limit if no GPUs are requested). These ratios can be modified in the cluster settings.

### CPU Memory Requests

* **If GPUs are requested** - The default memory allocation is set as a specific amount per GPU.\
  For example, if the default is 100MiB per GPU and your job requests 4 GPUs, it will be assigned 400MiB of memory.
* **If no GPUs are requested** - The default memory allocation is determined by a ratio of CPU memory limit to CPU memory request. By default, this ratio is 1:0.1 (your memory request will be 10% of the memory limit you specify).
* **System defaults** - The defaults are 100MiB per GPU, and 0.1 (10%) for memory requests when no GPUs are involved. These defaults can be customized in the cluster settings.

### CPU Compute and Memory Limits

By default, NVIDIA Run:ai sets the limit to Auto, meaning the workload can use up to the node’s maximum available resources unless you explicitly set a limit. Administrators can configure default limits using the cluster configuration.

## Setting Cluster Defaults

Administrators can change cluster-wide defaults for CPU and memory requests and/or limits. See [Advanced cluster configurations](../../../installation-and-configuration/configuration/cluster-config.md) for more details:

* **Using Helm** - Set the values under `clusterConfig` in your `values.yaml` file and upgrade the chart.
* **Using runaiconfig at runtime** - Edit the `runaiconfig` Custom Resource under `spec`.

The example below shows:

* A CPU request with a default ratio of 2:1 CPUs to GPUs.
* A CPU memory request with a default ratio of 200MB per GPU.
* A CPU limit with a default ratio of 4:1 CPU to GPU.
* A memory limit with a default ratio of 2GB per GPU.
* A CPU request with a default ratio of 0.1 CPUs per 1 CPU limit.
* A CPU memory request with a default ratio of 0.1:1 request per CPU memory limit.

```yaml
  clusterConfig:
    limitRange:
      cpuDefaultRequestGpuFactor: 2
      memoryDefaultRequestGpuFactor: 200Mi
      cpuDefaultLimitGpuFactor: 4
      memoryDefaultLimitGpuFactor: 2Gi
      cpuDefaultRequestCpuLimitFactorNoGpu: 0.1
      memoryDefaultRequestMemoryLimitFactorNoGpu: 0.1
```

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
