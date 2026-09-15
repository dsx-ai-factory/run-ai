# GPU Memory Fractions

To submit a [workload](../../../workloads-in-nvidia-run-ai/workloads.md) with GPU resources in Kubernetes, you typically need to specify an integer number of GPUs. However, workloads often require diverse GPU memory and compute requirements or even use GPUs intermittently depending on the application (such as inference workloads, training workloads or notebooks at the model-creation phase). Additionally, GPUs are becoming increasingly powerful, offering more processing power and larger memory capacity for applications. Despite the increasing model sizes, the increasing capabilities of GPUs allow them to be effectively shared among multiple users or applications.

NVIDIA Run:ai's GPU memory fractions provide an agile and easy-to-use method to share a GPU or multiple GPUs across workloads. With GPU memory fractions, you can divide the GPU memory into smaller chunks and allocate each chunk to a different workload, resulting in higher GPU utilization and more efficient resource allocation.

This page covers how GPU **memory** is divided between workloads. How those workloads share the GPU **compute** is a separate setting — refer to [GPU compute modes](compute-modes.md).

## Benefits of GPU Memory Fractions

Utilizing GPU memory fractions to share GPU resources among multiple workloads provides numerous advantages for both platform administrators and practitioners, including improved efficiency, resource optimization, and enhanced user experience.

* For the AI practitioner:
  * **Reduced wait time** - Workloads with smaller GPU requests are more likely to be scheduled quickly, minimizing delays in accessing resources.
  * **Increased workload capacity** - More workloads can be run using the same admin-defined GPU [quota](../scheduling/concepts-and-principles.md#quota) and available unused resources - [over quota](../scheduling/concepts-and-principles.md#over-quota).
* For the platform administrator:
  * **Improved GPU utilization** - Sharing GPUs across workloads increases the utilization of individual GPUs, resulting in better overall platform efficiency.
  * **Higher resource availability** - More users gain access to GPU resources, ensuring better distribution.
  * **Enhanced workload throughput** - More workloads can be served per GPU, ensuring maximum output from existing hardware.
  * **Optimized scheduling** - Smaller and dynamic resource allocations gives the [Scheduler](../scheduling/concepts-and-principles.md) a higher chance of finding GPU resources for incoming workloads.

## Quota Planning with GPU Memory Fractions

When planning the quota distribution for your [projects](../../aiinitiatives/organization/projects.md) and [departments](../../aiinitiatives/organization/departments.md), using memory fractions gives the platform administrator the ability to allocate more precise quota per project and department, assuming the usage of GPU memory fractions or enforcing it with [pre-defined policies](../../policies/policy-yaml-reference.md) or [compute resource](../../../workloads-in-nvidia-run-ai/workload-assets/compute-resources.md) templates.

### Interactive Development

Consider an organization with a department budgeted for **two nodes of 8×H100 GPUs** and a team of 32 researchers:

* Allocating 0.5 GPU per researcher ensures all researchers have access to GPU resources.
* Using memory fractions enables researchers to run smaller workloads intermittently within their quota or go over their quota by using temporary over quota resources with higher resource demanding workloads.
* Using GPUs for notebook-based model development, where GPUs are not continuously active and can be shared among multiple users.

### Inference Serving

Consider a team serving a small language model across a single node of **8×H100 80 GB GPUs**. A model such as Qwen3-8B occupies approximately 16.4 GB of GPU memory in BF16 precision, leaving most of each GPU unused when the model is given a whole device.

* Requesting 38 GB of GPU memory per model instance fits two instances on each GPU, for a total of 16 concurrent model servers on the node instead of 8.
* The remaining memory on each GPU accommodates the CUDA context and driver overhead, so the allocation holds under production load.
* Quota can be expressed in the memory each model actually needs, rather than rounded up to whole devices.

For more details on mapping your organization and resources, refer to [Adapting AI initiatives to your organization](../../aiinitiatives/adapting-ai-initiatives.md).

## How GPU Memory Fractions Work

From NVIDIA Run:ai v3.0, GPU memory fractions use the standard NVIDIA CUDA implementation. The NVIDIA Run:ai proprietary GPU sharing implementation is replaced by memory partitioning built into CUDA itself.

When a workload is submitted, the Scheduler finds a node with a GPU that can satisfy the requested GPU portion or GPU memory, then it schedules the pod to that node. CUDA then enforces the requested memory size on the selected GPU. **Each container uses its own separate GPU virtual memory address space.** An allocation that exceeds the requested memory size fails, exactly as it would on a physically smaller GPU, so no workload can use more than it requested and no workload can run over another workload's memory. This gives users the experience of a logical GPU per workload.

Each container also sees only its own share of GPU memory through NVML and `nvidia-smi`, rather than the memory of the whole device. A workload that requests 10 GB on an 80 GB GPU reports 10 GB of total GPU memory.

While [MIG](../../aiinitiatives/resources/mig-profiles.md) requires administrative work to configure every MIG slice, where a slice is a fixed chunk of memory, GPU memory fractions allow dynamic and fully flexible allocation of GPU memory chunks. NVIDIA Run:ai GPU memory fractions are agile and dynamic, allowing you to allocate and free GPU memory fractions during the runtime of the system, at any size between zero and the maximum GPU portion (100%) or memory size (up to the maximum memory size of a GPU).

The NVIDIA Run:ai Scheduler can work alongside other schedulers. To avoid collisions with other schedulers, the NVIDIA Run:ai Scheduler creates special reservation pods. After a workload is submitted requesting a fraction of a GPU, NVIDIA Run:ai creates a pod in a dedicated `runai-reservation` namespace with the full GPU as a resource, allowing other schedulers to understand that the GPU is reserved.

<Info>

**Note**

* Splitting a GPU into memory fractions may generate some fragmentation of the GPU memory. The [Scheduler](../scheduling/concepts-and-principles.md) will try to consolidate GPU resources where feasible (that is, preemptible workloads).
* Using [bin-pack](../../aiinitiatives/resources/node-pools.md#adding-a-new-node-pool) as a scheduling placement strategy can also reduce GPU fragmentation.
* Using [dynamic GPU memory fractions](dynamic-memory-fractions.md) ensures that even small unused fragments of GPU memory are utilized by workloads.

</Info>

## Multi-GPU Memory Fractions

NVIDIA Run:ai also supports workload submission using multi-GPU memory fractions. Multi-GPU memory fractions work similarly to single-GPU memory fractions, however, the NVIDIA Run:ai Scheduler selects the node and the GPUs, and allocates the same fraction size on multiple GPU devices within the same node. For example, if practitioners develop a new model that uses 8 GPUs and requires 40 GB of memory per GPU, they can allocate 8×40 GB with multi-GPU memory fractions instead of reserving the full memory of each GPU (for example, 80 GB). This leaves 40 GB of GPU memory available on each of the 8 GPUs for other workloads within that node.

Sharing a single GPU between multiple workloads remains unchanged. It now serves multiple workloads using multi-GPUs per workload, single-GPU per workload, or a mix of both.

## Prerequisites

GPU memory fractions require the following on every GPU node in the cluster:

| Component | Required version |
|---|---|
| CUDA | 13.4 or later |
| CUDA driver | r615 or later |
| NVIDIA GPU Operator | 26.7 or later |

In addition, GPU nodes must run Linux with cgroup v2 mounted at `/sys/fs/cgroup`.

<Info>

**Note**

NVIDIA Run:ai v3.0 installs and operates normally without these prerequisites in place, and non-fractional workloads are unaffected. Fractional workloads, however, are not scheduled onto a node until that node meets every prerequisite. Because the prerequisites are satisfied per node, a node pool with mixed driver versions runs fractional workloads on the upgraded nodes while workloads landing on the remaining nodes stay pending. Verify the driver and GPU Operator versions node by node rather than cluster-wide.

</Info>

## Setting GPU Memory Fractions

The NVIDIA Run:ai UI, CLI, and APIs are unchanged in v3.0. Workloads are submitted exactly as they were before, and existing automation, pipelines, and saved templates continue to work without modification.

Using the [compute resources](../../../workloads-in-nvidia-run-ai/workload-assets/compute-resources.md) asset, you can define the compute requirements by specifying your requested GPU portion or GPU memory, and use it with any of the [NVIDIA Run:ai workload types](../../../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md) for single-GPU and multi-GPU memory fractions.

* **Single-GPU memory fractions** - Define the compute requirement to run 1 GPU device, by specifying either a fraction (percentage) of the overall memory or a memory request (GB, MB).
* **Multi-GPU memory fractions** - Define the compute requirement to run multiple GPU devices, by specifying either a fraction (percentage) of the overall memory or a memory request (GB, MB).

## Setting GPU Memory Fractions via YAML

To enable GPU memory fractions for workloads submitted via Kubernetes YAML, use the annotations described below.

NVIDIA Run:ai v3.0 maintains backward compatibility with the annotations used in earlier versions. The legacy annotations continue to work unchanged. Alongside them, v3.0 introduces official NVIDIA annotations prefixed with `nvidia.com/`, which provide the new v3.0 functionality and an NVIDIA-native alternative to the legacy request annotations.

<Info>

**Note**

Make sure the [default scheduler](../scheduling/default-scheduler.md) is set to `runai-scheduler`.

</Info>

### Annotation Prefixes

Two prefixes are used, and the prefix indicates which component consumes the annotation:

* `nvidia.com/` - The annotation reaches the node and is applied to CUDA.
* `kai.scheduler/` - The annotation is consumed by the Scheduler only and is never passed to the node. GPU memory portions are a Scheduler convention, because CUDA expresses memory only as an absolute size. The Scheduler resolves the portion against the GPU it selected and passes an absolute size to the node.

### Annotation Reference

The following annotations configure GPU memory fractions. Annotations that carry a `container.<container-name>` element apply to a single container. Annotations without that element apply to the pod.

| Annotation | Scope | Input format | Purpose |
|---|---|---|---|
| `gpu-fraction` | Pod | A portion of GPU memory as a double-precision floating-point number. For example, `0.25`, `0.75`. | Requests a portion of GPU memory. |
| `gpu-memory` | Pod | Memory size in MiB. For example, `2500`, `4096`. | Requests an absolute amount of GPU memory. |
| `gpu-fraction-num-devices` | Pod | The number of GPU devices to allocate using the specified `gpu-fraction` or `gpu-memory` value. Defaults to 1 if omitted. | Requests multiple GPU devices. |
| `gpu-fraction-container-name` | Pod | The name of a container in the pod. Defaults to the first container (index 0) if omitted. | Selects which container receives the GPU allocation when the request is made with `gpu-fraction` or `gpu-memory`. |
| `nvidia.com/container.<container-name>.gpu-memory.request` | Container | Memory size in Kubernetes resource quantity format. For example, `40G`. | Requests an absolute amount of GPU memory. This is the official NVIDIA alternative to `gpu-memory`. |
| `nvidia.com/container.<container-name>.gpu-compute.mode` | Container | `time-slicing` or `sm-sharing`. Defaults to `time-slicing` if omitted. | Sets how the workload shares GPU compute. Refer to [GPU compute modes](compute-modes.md). |

<Info>

**Note**

* GPU memory fractions can be assigned to only one container in the pod.
* The number of GPU devices is a pod-level property, so `gpu-fraction-num-devices` has no container-scoped equivalent.
* There is no container-scoped equivalent for a GPU memory portion request. Use `gpu-fraction`.
* If `gpu-fraction-container-name` is used together with a container-scoped annotation, the container names must be identical or the request is rejected.

</Info>

Dynamic GPU memory fractions add a memory limit to the request. Refer to [dynamic GPU memory fractions](dynamic-memory-fractions.md) for those annotations.

### Examples

The following example requests an absolute amount of GPU memory on 4 GPU devices:

```yaml
metadata:
  annotations:
    gpu-memory: "4000"   # MiB
    # or, new-style: nvidia.com/container.app.gpu-memory.request: "40G"
    # "time-slicing" (the default, may be omitted) or "sm-sharing"
    nvidia.com/container.app.gpu-compute.mode: "sm-sharing"
    # optional; 1 if omitted — number of fractional GPU devices
    gpu-fraction-num-devices: "4"
spec:
  containers:
    - name: app
```

The following example requests a portion of GPU memory on 2 GPU devices:

```yaml
metadata:
  annotations:
    gpu-fraction: "0.5"
    # "time-slicing" (the default, may be omitted) or "sm-sharing"
    nvidia.com/container.app.gpu-compute.mode: "time-slicing"
    # optional; 1 if omitted — number of fractional GPU devices
    gpu-fraction-num-devices: "2"
spec:
  containers:
    - name: app
```

The following complete example creates a pod that requests 2 GPU devices, each requesting 50% of memory:

```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    user: test
    gpu-fraction: "0.5"
    gpu-fraction-num-devices: "2"
    # Specify which container should receive the GPU memory fraction allocation
    # By default, the first container (index 0) receives the GPU allocation
    gpu-fraction-container-name: "app"
  labels:
    runai/queue: test
  name: multi-fractional-pod-job
  namespace: test
spec:
  containers:
  - image: gcr.io/run-ai-demo/quickstart-cuda
    imagePullPolicy: Always
    name: app
    resources:
      limits:
        cpu: 200m
        memory: 200Mi
      requests:
        cpu: 100m
        memory: 100Mi
    securityContext:
      capabilities:
        drop: ["ALL"]
  schedulerName: runai-scheduler
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 5
```

## Deployment Considerations

* Selecting a GPU portion using percentages as units does not guarantee the exact memory size. This means 50% of an A-100-40GB is 20 GB while 50% of an A-100-80GB is 40 GB. To have better control over the exact allocated memory, specify the exact memory size, for example 40 GB.
* GPU memory fractions control the memory split. Each workload's share of GPU compute is controlled separately by its compute mode — refer to [GPU compute modes](compute-modes.md).
* An application sees its own requested GPU memory size through NVML and `nvidia-smi`, not the memory of the whole device. Tools that report total GPU memory inside the container report the fraction.

## Using CLI

To view the available actions, go to the [CLI v2 reference](../../../reference/cli/runai/) and run according to your workload.

## Using API

To view the available actions, go to the [API reference](https://run-ai-docs.nvidia.com/api/2.26/) and run according to your workload.
