# Dynamic GPU Memory Fractions

Many workloads utilize GPU resources intermittently, with long periods of inactivity. These workloads typically need GPU resources when they are running AI applications or debugging a model in development. Other workloads such as inference may utilize GPUs at lower rates than requested, but may demand higher resource usage during peak utilization. The disparity between resource request and actual resource utilization often leads to inefficient utilization of GPUs. This usually occurs when multiple workloads request resources based on their peak demand, despite operating below those peaks for the majority of their runtime.

To address this challenge, NVIDIA Run:ai provides dynamic GPU memory fractions. This feature optimizes GPU utilization by enabling workloads to adjust their memory usage at runtime. Instead of a single value, a workload declares a memory **Request** that is guaranteed to it, and a higher memory **Limit** that it can grow into when the additional memory is not needed by others.

This page covers how GPU **memory** is divided between workloads. How those workloads share the GPU **compute** is a separate setting — refer to [GPU compute modes](compute-modes.md).

## Benefits of Dynamic GPU Memory Fractions

* **Higher GPU utilization** - Memory that a workload requested but is not using at the moment is available to other workloads on the same GPU.
* **Smaller quota requirements** - Workloads can be sized to their typical usage rather than their peak, so a project needs less quota for the same work.
* **Resilient interactive sessions** - When a workload must give back borrowed memory, NVIDIA Run:ai sends an OOMKill signal to the process that is accessing the GPU, not to the pod. For a Jupyter Notebook this means the notebook itself keeps running. The user can continue working and the notebook re-attaches to a GPU the next time it uses one, rather than losing the session and starting over.

## How Dynamic GPU Memory Fractions Work

With dynamic GPU memory fractions, you [submit workloads](../../../workloads-in-nvidia-run-ai/workloads.md) using a GPU memory Request and Limit, which is achieved by leveraging the Kubernetes Request and Limit notations. You can either:

* Request a GPU memory portion using a percentage of a GPU and specify a Limit
* Request a GPU memory size (GB, MB) and specify a Limit

When setting a GPU memory limit either as a GPU memory portion or a GPU memory size, the Limit must be equal to or greater than the memory Request. Both forms are translated into the actual requested memory size of the Request (guaranteed resources) and the Limit (burstable resources, not guaranteed).

For example, you can specify a workload with a GPU memory portion request of 0.25 GPU, and add a limit of up to 0.80 GPU. The Scheduler schedules the workload to a node that can provide the request (0.25), and then assigns the workload to a GPU. The workload is allowed to occupy memory between 0 and 0.80 of the GPU memory, where only 0.25 of the GPU memory is guaranteed to that workload. The rest of the memory, from 0.25 to 0.80, is loaned to the workload as long as it is not needed by other workloads.

NVIDIA Run:ai automatically manages the state changes between Request and Limit as well as the reverse, when the balance needs to be returned, updating the workloads' utilization compared to Request and Limit parameters in the [metrics pane for each workload](../../../workloads-in-nvidia-run-ai/workloads.md).

### What the Application Sees

An application sees its own share of GPU memory through NVML and `nvidia-smi`, rather than the memory of the whole device. The reported size depends on how the workload was submitted:

* **Request only** - The application sees a GPU memory space the size of the Request.
* **Request and Limit** - The application sees the memory size it can actually allocate at the moment it is bound to the GPU. This value is never below the Request and never above the Limit. Where it falls between them depends on how much additional GPU memory is free at that point in time.

### Reclaiming Loaned Memory

To guarantee fair quality of service between different workloads using the same GPU, NVIDIA Run:ai uses an extendable GPUOOMKiller (Out Of Memory Killer) component that guarantees quality of service using Kubernetes semantics of Request and Limit. When memory that was loaned to a workload is needed by another workload, the GPUOOMKiller sends an OOMKill signal to the process accessing the GPU. The pod itself is not terminated.

<Info>

**Note**

Dynamic GPU memory fractions is enabled by default in the [cluster](../../../installation-and-configuration/post-installation-configuration/cluster-config.md).

</Info>

## Multi-GPU Dynamic Memory Fractions

NVIDIA Run:ai also supports workload submission using multi-GPU dynamic memory fractions. Multi-GPU dynamic memory fractions work similarly to dynamic memory fractions on a single GPU workload, however, instead of a single GPU device, the NVIDIA Run:ai Scheduler selects the node and the GPUs, and allocates the same Request and Limit pair on multiple GPU devices within the same node. For example, if practitioners develop a new model that uses 8 GPUs and requires 40 GB of memory per GPU, but may want to burst out and consume up to the full GPU memory, they can allocate 8×40 GB with multi-GPU memory fractions and a limit of 80 GB (for example, an H100 GPU) instead of reserving the full memory of each GPU. This leaves 40 GB of GPU memory available on each of the 8 GPUs for other workloads within that node. This is useful during model development, where memory requirements are usually lower due to experimentation with smaller models or configurations.

This approach improves GPU utilization and availability, enabling more precise and often smaller quota requirements for the end user.

## Prerequisites

Dynamic GPU memory fractions require the same cluster prerequisites as GPU memory fractions — CUDA 13.4 or later, CUDA driver r615 or later, and NVIDIA GPU Operator 26.7 or later on every GPU node. Refer to [GPU memory fractions](memory-fractions.md#prerequisites) for the full list.

## Setting Dynamic GPU Memory Fractions

Using the [compute resources](../../../workloads-in-nvidia-run-ai/workload-assets/compute-resources.md) asset, you can define the compute requirements by specifying your requested GPU memory portion or GPU memory size, and set a Limit. You can then use the compute resource with any of the [NVIDIA Run:ai workload types](../../../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md) for single-GPU and multi-GPU dynamic memory fractions. In addition, you can view the workloads' utilization compared to Request and Limit parameters in the [metrics pane for each workload](../../../workloads-in-nvidia-run-ai/workloads.md).

* **Single-GPU dynamic memory fractions** - Define the compute requirement to run 1 GPU device, by specifying either a portion (percentage) of the overall memory or a memory request (GB, MB) with a Limit.
* **Multi-GPU dynamic memory fractions** - Define the compute requirement to run multiple GPU devices, by specifying either a portion (percentage) of the overall memory or a memory request (GB, MB) with a Limit.

<Info>

**Note**

Dynamic GPU memory fractions is disabled by default in the NVIDIA Run:ai UI only. To set a memory Limit through the UI, it must be enabled by your administrator under **General settings** → **Resources** → **GPU resource optimization**. The CLI, the APIs, and the Scheduler support dynamic GPU memory fractions by default, and are not affected by this setting.

</Info>

## Setting Dynamic GPU Memory Fractions via YAML

<Warning>

**Breaking change from version 3.0:**

The `RUNAI_GPU_MEMORY_LIMIT` environment variable is no longer honored. The memory Limit is now set with an annotation.

A workload that still sets `RUNAI_GPU_MEMORY_LIMIT` starts and runs normally, but with no memory ceiling — it is treated as a Request without a Limit. No error is raised and no event is generated. Audit your manifests for the variable rather than waiting for a symptom.

</Warning>

Dynamic GPU memory fractions extend the [GPU memory fractions annotations](memory-fractions.md#annotation-reference) with a memory Limit. The Request is set exactly as it is for a static memory fraction.

| Annotation | Scope | Input format | Purpose |
|---|---|---|---|
| `nvidia.com/container.<container-name>.gpu-memory.limit` | Container | Memory size in Kubernetes resource quantity format. For example, `80G`. | Sets the memory Limit for a request made as an absolute memory size. Replaces `RUNAI_GPU_MEMORY_LIMIT` for absolute values. |
| `kai.scheduler/container.<container-name>.gpu-memory.portion.limit` | Container | A portion of GPU memory as a double-precision floating-point number. For example, `0.8`. | Sets the memory Limit for a request made as a GPU memory portion. Replaces `RUNAI_GPU_MEMORY_LIMIT` for portion values. |

<Info>

**Note**

* Pair an absolute memory Request with `nvidia.com/container.<container-name>.gpu-memory.limit`, and a memory portion Request with `kai.scheduler/container.<container-name>.gpu-memory.portion.limit`. Pairing a Request with the Limit of the other form causes the Limit to be interpreted as the wrong kind of value.
* Substitute `<container-name>` with the name of the container receiving the fraction. This is the same container named by `gpu-fraction-container-name`, or the first container in the pod if that annotation is omitted. If both are used, the names must be identical or the request is rejected.
* The Limit does not affect the Scheduler. The Scheduler places the workload according to its Request. The Limit governs only how far the workload can then grow at runtime.

</Info>

### Migrating from the Environment Variable

The following example shows a workload requesting an absolute memory size before and after the upgrade to v3.0.

Before, up to v2.26:

```yaml
spec:
  containers:
  - image: gcr.io/run-ai-demo/quickstart-cuda
    name: app
    env:
    - name: RUNAI_GPU_MEMORY_LIMIT   # <-- remove: no longer honored in v3.0
      value: "8000M"
```

After, from v3.0:

```yaml
metadata:
  annotations:
    gpu-memory: "4000"   # MiB
    nvidia.com/container.app.gpu-memory.limit: "8000M"
spec:
  containers:
  - image: gcr.io/run-ai-demo/quickstart-cuda
    name: app
```

### Examples

The following example requests an absolute amount of GPU memory with a Limit, on 4 GPU devices:

```yaml
metadata:
  annotations:
    gpu-memory: "4000"   # MiB
    # or, new-style: nvidia.com/container.app.gpu-memory.request: "40G"
    nvidia.com/container.app.gpu-memory.limit: "80G"
    # "time-slicing" (the default, may be omitted) or "sm-sharing"
    nvidia.com/container.app.gpu-compute.mode: "sm-sharing"
    # optional; 1 if omitted — number of fractional GPU devices
    gpu-fraction-num-devices: "4"
spec:
  containers:
    - name: app          # the <container-name> used in the limit annotation
```

The following example requests a portion of GPU memory with a Limit, on 2 GPU devices:

```yaml
metadata:
  annotations:
    gpu-fraction: "0.5"
    kai.scheduler/container.app.gpu-memory.portion.limit: "0.8"
    # "time-slicing" (the default, may be omitted) or "sm-sharing"
    nvidia.com/container.app.gpu-compute.mode: "time-slicing"
    # optional; 1 if omitted — number of fractional GPU devices
    gpu-fraction-num-devices: "2"
spec:
  containers:
    - name: app          # the <container-name> used in the limit annotation
```

## Deployment Considerations

* The Limit must be equal to or greater than the memory Request.
* The Limit is enforced at runtime only and is never read by the Scheduler. Raising a Limit does not change where a workload is placed.
* Memory loaned above the Request is not guaranteed. A workload operating above its Request can have that memory reclaimed at any time.

## Using CLI

To view the available actions, go to the [CLI v2 reference](../../../reference/cli/runai/) and run according to your workload.

## Using API

To view the available actions, go to the [API reference](https://run-ai-docs.nvidia.com/api/2.26/) and run according to your workload.
