# Dynamic GPU Memory Fractions

Many workloads utilize GPU resources intermittently, with long periods of inactivity. These workloads typically need GPU resources when they are running AI applications or debugging a model in development. Other workloads such as inference may utilize GPUs at lower rates than requested, but may demand higher resource usage during peak utilization. The disparity between resource request and actual resource utilization often leads to inefficient utilization of GPUs. This usually occurs when multiple workloads request resources based on their peak demand, despite operating below those peaks for the majority of their runtime.

To address this challenge, NVIDIA Run:ai provides dynamic GPU memory fractions. This feature optimizes GPU utilization by enabling workloads to adjust their memory usage at runtime. Instead of a single value, a workload declares a memory **Request** that is guaranteed to it, and a higher memory **Limit** that it can grow into when the additional memory is not needed by others. Dynamic GPU memory fractions rely on the standard NVIDIA CUDA implementation to enforce the memory allocation on each GPU.

<Note>
**Note**

This page covers how GPU **memory** is divided between workloads. How those workloads share GPU **compute** is a separate setting. Refer to [GPU compute modes](compute-modes.md).
</Note>

## Benefits of Dynamic GPU Memory Fractions

* **Higher GPU utilization** - Memory that a workload requested but is not using at the moment is available to other workloads on the same GPU.
* **Smaller quota requirements** - Workloads can be sized to their typical usage rather than their peak, so a project needs less quota for the same work.
* **Resilient interactive sessions** - When a workload must give back borrowed memory, the OOMKill signal is sent to the process that is accessing the GPU rather than to the container as a whole. Where that process is not the root process of the container, as in most interactive notebooks, the session survives and the user can use the GPU again without losing their work.

## Quota Planning with Dynamic GPU Memory Fractions

When planning the quota distribution for your [projects](../../aiinitiatives/organization/projects.md) and [departments](../../aiinitiatives/organization/departments.md), dynamic memory fractions give the platform administrator the ability to allocate more precise quota per project and department, assuming the usage of GPU memory fractions, or enforcing it with [pre-defined policies](../../policies/policy-yaml-reference.md) or [compute resource](../../../workloads-in-nvidia-run-ai/workload-assets/compute-resources.md) templates. The Limit lets a workload burst beyond its allocated quota when the GPU has memory to spare, without provisioning additional quota for that burst capacity.

### Model Development

Consider researchers developing and testing models in notebooks:

* Sizing the Request to a workload's typical memory footprint during experimentation keeps quota consumption low, while a higher Limit lets the workload grow during occasional peaks, for example loading a larger checkpoint to validate a change.
* Because quota is based on the Request, more researchers can be granted access to the same node pool than static fractions would allow, with bursts handled opportunistically rather than provisioned for upfront.

### Serial Workload Chains

Consider a processing pipeline where multiple stages run in sequence on the same GPU, for example a data-preprocessing stage followed by a training stage:

* Each stage requests only the memory it is guaranteed as its Request, and grows into its Limit as memory becomes available when a prior stage in the chain completes and frees its GPU memory.
* Quota reflects the sum of each stage's Request rather than the peak memory used by any single stage, since memory moves between stages as demand shifts instead of each one holding its peak for its entire lifetime.

## How Dynamic GPU Memory Fractions Work

A dynamic GPU memory fraction is submitted with a Request and a Limit, using the Kubernetes Request and Limit notations. NVIDIA Run:ai supports this both on a single GPU and across multiple GPUs on the same node.

### Single-GPU Dynamic Memory Fractions

With dynamic GPU memory fractions, you [submit workloads](../../../workloads-in-nvidia-run-ai/workloads.md) using a GPU memory Request and Limit. You can either:

* Request a GPU memory portion using a percentage of a GPU and specify a Limit
* Request a GPU memory size (GB, MB) and specify a Limit

When setting a GPU memory limit either as a GPU memory portion or a GPU memory size, the Limit must be equal to or greater than the memory Request. Both forms are translated into the actual requested memory size of the Request (guaranteed resources) and the Limit (burstable resources, not guaranteed).

For example, you can specify a workload with a GPU memory portion request of 0.25 GPU, and add a limit of up to 0.80 GPU. The Scheduler schedules the workload to a node that can provide the request (0.25), and then assigns the workload to a GPU. The workload is allowed to occupy memory between 0 and 0.80 of the GPU memory, where only 0.25 of the GPU memory is guaranteed to that workload. The rest of the memory, from 0.25 to 0.80, is loaned to the workload as long as it is not needed by other workloads.

NVIDIA Run:ai automatically manages the transition between Request and Limit, and the reverse when the borrowed memory needs to be returned. You can view a workload's utilization against its Request and Limit in the [metrics pane for each workload](../../../workloads-in-nvidia-run-ai/workloads.md).

### Multi-GPU Dynamic Memory Fractions

NVIDIA Run:ai also supports workload submission using multi-GPU dynamic memory fractions. Multi-GPU dynamic memory fractions work similarly to dynamic memory fractions on a single GPU workload, however, instead of a single GPU device, the NVIDIA Run:ai Scheduler selects the node and the GPUs, and allocates the same Request and Limit pair on multiple GPU devices within the same node. For example, if practitioners develop a new model that uses 8 GPUs and requires 40 GB of memory per GPU, but may want to burst out and consume up to the full GPU memory, they can allocate 8×40 GB with multi-GPU memory fractions and a limit of 80 GB (for example, an H100 GPU) instead of reserving the full memory of each GPU. This leaves 40 GB of GPU memory available on each of the 8 GPUs for other workloads within that node.

This approach improves GPU utilization and availability, enabling more precise and often smaller quota requirements for the end user.

### What the GPU Application Sees

A GPU application sees its own share of GPU memory through the NVIDIA Management Library (NVML) and `nvidia-smi`, rather than the memory of the whole device.

A GPU memory fraction submitted without a Limit has a Limit equal to its Request. The GPU application therefore sees a GPU memory space the size of the Request, and that size does not change while the workload runs.

For a workload submitted with a Request and a Limit, the free memory reported to the GPU application always covers the rest of its Request, and grows beyond that when the GPU has memory to spare - never above the Limit. In other words, the Request is always available to the workload, and the memory between the Request and the Limit is available only while no one else needs it.

For example, take a workload with a 4 GB Request and a 12 GB Limit that is currently using 1 GB of GPU memory:

* If 9 GB is free on the GPU, the GPU application sees 9 GB free and can grow toward its 12 GB Limit.
* If the GPU is busy and only 1 GB is free, the GPU application still sees 3 GB free - the remainder of its guaranteed 4 GB.

### Reclaiming Loaned Memory

To guarantee fair quality of service between different workloads using the same GPU, NVIDIA Run:ai uses an extendable GPUOOMKiller (Out Of Memory Killer) component that guarantees quality of service using Kubernetes semantics of Request and Limit. When memory that was loaned to a workload is needed by another workload, the GPUOOMKiller sends an OOMKill signal to the process that is accessing the GPU.

What happens to the workload depends on how the container is built:

* If the process accessing the GPU is the root process of the container, terminating it also terminates the container.
* If the process accessing the GPU is a child of the root process, the container keeps running. The GPU application can start a new process and use the GPU again.

Interactive notebooks are typically built the second way. The notebook server runs as the root process and the GPU work runs in a kernel process that the server starts. An OOMKill therefore ends the GPU work but leaves the notebook session running, and the user can run the cell again.

A memory reclaim is targeted at a single process and affects only the container that process belongs to. This is true in both [GPU compute modes](compute-modes.md), including SM-sharing, where workloads otherwise share an error domain.

## Prerequisites

<Note>

**Note**

NVIDIA Run:ai installs and operates normally without these prerequisites in place, and non-fractional workloads are unaffected. Fractional workloads, however, are not scheduled onto a node until that node meets every prerequisite. If you are upgrading from a release that used the proprietary GPU sharing implementation, refer to [Migrate GPU Fractions](../../../../installation-and-configuration/upgrade/migrate-gpu-fractions.md) for per-node verification steps.

</Note>

Dynamic GPU memory fractions require the following on every GPU node in the cluster. GPU nodes must also run Linux with cgroup v2 mounted at `/sys/fs/cgroup`, and since [GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/index.html) 26.7.1 does not default to CUDA driver r615, set the driver version explicitly rather than relying on the default.

| Component | Required version |
|---|---|
| CUDA | 13.4 or later |
| CUDA driver | r615 or later |
| NVIDIA GPU Operator | 26.7.1 or later |

## Setting Dynamic GPU Memory Fractions

<Note>

**Note**

If you are upgrading from a release that used the proprietary GPU sharing implementation, the UI, CLI, and APIs for submitting workloads are unchanged. Workloads are submitted exactly as they were before, and existing automation, pipelines, and saved templates continue to work without modification.

</Note>

Using the [compute resources](../../../workloads-in-nvidia-run-ai/workload-assets/compute-resources.md) asset, you can define the compute requirements by specifying your requested GPU memory portion or GPU memory size, and set a Limit. You can then use the compute resource with [NVIDIA Run:ai native workloads](../../../workloads-in-nvidia-run-ai/workload-types/native-workloads.md) for single-GPU and multi-GPU dynamic memory fractions. In addition, you can view the workloads' utilization compared to Request and Limit parameters in the [metrics pane for each workload](../../../workloads-in-nvidia-run-ai/workloads.md).

* **Single-GPU dynamic memory fractions** - Define the compute requirement to run 1 GPU device, by specifying either a portion (percentage) of the overall memory or a memory request (GB, MB) with a Limit.
* **Multi-GPU dynamic memory fractions** - Define the compute requirement to run multiple GPU devices, by specifying either a portion (percentage) of the overall memory or a memory request (GB, MB) with a Limit.

<Note>

**Note**

Dynamic GPU memory fractions is disabled by default. To set a memory Limit through the NVIDIA Run:ai UI or CLI, it must be enabled by your administrator under **General settings** → **Resources** → **GPU resource optimization**. While the setting is disabled, the UI does not offer a Limit field on a compute resource, and the CLI rejects a submission that specifies one.

The setting does not apply to supported workload types submitted via YAML or to externally submitted workloads, since their Limit is set directly with a Kubernetes annotation rather than through the UI or CLI. Do not rely on this setting to prevent the use of memory Limits across a cluster.

</Note>

## Setting Dynamic GPU Memory Fractions via YAML

This method applies to [supported workload types](../../../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md) and to [workloads submitted externally](../../../workloads-in-nvidia-run-ai/workload-types/supported-features.md#externally-submitted-kubernetes-workloads), outside of NVIDIA Run:ai. To enable dynamic GPU memory fractions for workloads submitted via Kubernetes YAML, use the annotations described below.

<Note>

**Note**

Make sure the [default scheduler](../scheduling/default-scheduler.md) is set to `runai-scheduler`.

</Note>

<Warning>

**Warning**

If you are upgrading from a release that used the `RUNAI_GPU_MEMORY_LIMIT` environment variable, that variable is no longer honored; the memory Limit is now set with an annotation. A workload that still sets `RUNAI_GPU_MEMORY_LIMIT` starts and runs normally, but with no memory ceiling, and no error or event is raised. Refer to [Migrate GPU Fractions](../../../../installation-and-configuration/upgrade/migrate-gpu-fractions.md) for the annotation replacement and a before/after example.

</Warning>

### Annotation Prefixes

Two prefixes are used, and the prefix indicates which component consumes the annotation:

* `nvidia.com/` - The annotation reaches the node and is applied to CUDA.
* `kai.scheduler/` - The annotation is consumed by the Scheduler only and is never passed to the node. GPU memory portions are a Scheduler convention, because CUDA expresses memory only as an absolute size. The Scheduler resolves the portion against the GPU it selected and passes an absolute size to the node.

### Annotation Reference

A dynamic GPU memory fraction is always allocated to a single container. NVIDIA Run:ai currently supports one fractional container per pod, so every annotation below ultimately applies to that one container. The Request sets the guaranteed memory, and the Limit sets the ceiling the workload can grow into. The annotations differ only in how the container is identified:

* Annotations that embed `container.<container-name>` in the key name the container directly.
* Annotations without that element apply to the container named by `gpu-fraction-container-name`, or to the first container in the pod when that annotation is omitted.

| Annotation | Input format | Purpose |
|---|---|---|
| `gpu-fraction` | A portion of GPU memory as a double-precision floating-point number. For example, `0.25`, `0.75`. | Requests a portion of GPU memory. |
| `gpu-memory` | Memory size in MiB. For example, `2500`, `4096`. | Requests an absolute amount of GPU memory. |
| `gpu-fraction-num-devices` | The number of GPU devices to allocate using the specified `gpu-fraction` or `gpu-memory` value. Defaults to 1 if omitted. | Requests multiple GPU devices for the fractional container. |
| `gpu-fraction-container-name` | The name of a container in the pod. Defaults to the first container (index 0) if omitted. | Selects which container receives the GPU allocation when the request is made with `gpu-fraction` or `gpu-memory`. |
| `nvidia.com/container.<container-name>.gpu-memory.request` | Memory size in Kubernetes resource quantity format. For example, `40G`. | Requests an absolute amount of GPU memory. This is the official NVIDIA alternative to `gpu-memory`. |
| `nvidia.com/container.<container-name>.gpu-memory.limit` | Memory size in Kubernetes resource quantity format. For example, `80G`. | Sets the memory Limit for a request made as an absolute memory size. |
| `kai.scheduler/container.<container-name>.gpu-memory.portion.limit` | A portion of GPU memory as a double-precision floating-point number. For example, `0.8`. | Sets the memory Limit for a request made as a GPU memory portion. |

<Note>

**Note**

* Pair an absolute memory Request with `nvidia.com/container.<container-name>.gpu-memory.limit`, and a memory portion Request with `kai.scheduler/container.<container-name>.gpu-memory.portion.limit`. Pairing a Request with the Limit of the other form causes the Limit to be interpreted as the wrong kind of value.
* Substitute `<container-name>` with the name of the container receiving the fraction. This is the same container named by `gpu-fraction-container-name`, or the first container in the pod if that annotation is omitted. If both are used, the names must be identical or the request is rejected.
* The Limit does not affect the Scheduler. The Scheduler places the workload according to its Request. The Limit governs only how far the workload can then grow at runtime.
* To set how a workload shares GPU compute, refer to the annotations described in [GPU compute modes](compute-modes.md).

</Note>

### Examples

The following example requests an absolute amount of GPU memory with a Limit, on 4 GPU devices:

```yaml
metadata:
  annotations:
    gpu-memory: "4000"   # MiB
    # or: nvidia.com/container.app.gpu-memory.request: "40G"
    nvidia.com/container.app.gpu-memory.limit: "80G"
    # optional; 1 if omitted - number of fractional GPU devices
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
    # optional; 1 if omitted - number of fractional GPU devices
    gpu-fraction-num-devices: "2"
spec:
  containers:
    - name: app          # the <container-name> used in the limit annotation
```

## Deployment Considerations

* The Limit must be equal to or greater than the memory Request.
* The Limit is enforced at runtime only and is never read by the Scheduler. Raising a Limit does not change where a workload is placed.
* Memory loaned above the Request is not guaranteed. A workload operating above its Request can have that memory reclaimed at any time.
* To keep a workload running through a memory reclaim, build the container so that the process accessing the GPU is not the root process of the container. A workload whose root process accesses the GPU directly is terminated when its loaned memory is reclaimed.

## Using CLI

To view the available actions, go to the [CLI v2 reference](../../../reference/cli/runai/) and run according to your workload.

## Using API

To view the available actions, go to the [API reference](https://run-ai-docs.nvidia.com/api/2.26/) and run according to your workload.
