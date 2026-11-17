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

For a workload submitted with a Request and a Limit, the free memory reported to the GPU application always covers the rest of its Request, and grows beyond that when the GPU has memory to spare - never above the Limit. In other words, the Request is always available to the workload, and the memory between the Request and the Limit is available only while no one else needs it. Refer to [Known limitations with CUDA 13.4 and driver r615](#known-limitations-with-cuda-134-and-driver-r615) for cases where this guarantee does not hold.

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

NVIDIA Run:ai installs and operates normally without these prerequisites in place, and workloads not using GPU fractions are unaffected. Workloads using GPU fractions, however, are not scheduled onto a node until that node meets every prerequisite. If you are upgrading from a release that used the proprietary GPU sharing implementation, refer to [Migrate GPU Fractions](../../../installation-and-configuration/upgrade/migrate-gpu-fractions.md) for per-node verification steps.

</Note>

Dynamic GPU memory fractions require the following on every GPU node in the cluster. GPU nodes must also run Linux with cgroup v2 mounted at `/sys/fs/cgroup`, and since [GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/index.html) 26.7.1 does not default to CUDA driver r615, set the driver version explicitly rather than relying on the default.

| Component | Required version |
|---|---|
| CUDA | 13.4 or later |
| CUDA driver | r615 or later |
| NVIDIA GPU Operator | 26.7.1 or later |
| NVIDIA GPU Operator NRI plugin | Disabled. Install the NVIDIA GPU Operator with `cdi.nriPluginEnabled` set to `false` (`--set cdi.nriPluginEnabled=false`). GPU memory fractions do not function while the NRI plugin is enabled. |

## Using Dynamic GPU Memory Fractions in Workloads

<Note>

**Note**

If you are upgrading from a release that used the proprietary GPU sharing implementation, the UI, CLI, and APIs for submitting workloads are unchanged. Workloads are submitted exactly as they were before, and existing automation, pipelines, and saved templates continue to work without modification.

</Note>

<Warning>

**Warning**

If you are upgrading from a release that used the `RUNAI_GPU_MEMORY_LIMIT` environment variable, that variable is no longer honored; the memory Limit is now set with an annotation. A workload that still sets `RUNAI_GPU_MEMORY_LIMIT` starts and runs normally, but with no memory ceiling, and no error or event is raised. Refer to [Migrate GPU Fractions](../../../installation-and-configuration/upgrade/migrate-gpu-fractions.md) for the annotation replacement and a before/after example.

</Warning>

When you submit a workload, you can add a memory Limit to the memory Request, so that the workload can use more memory than its Request while the GPU has memory to spare. Setting the Limit through the UI or CLI requires an administrator to enable it first. For the steps, the annotation reference, and examples, refer to [Request GPU memory fractions](../../../workloads-in-nvidia-run-ai/configure-workload-scheduling/configure-gpu-sharing/request-gpu-memory-fractions.md#dynamic-gpu-memory-fractions).

## Enforcing Dynamic GPU Memory Fractions with a Policy

Administrators can also control dynamic GPU memory fractions centrally with a [workload policy](../../policies/policies-and-rules.md). A policy can supply a default Request and Limit for workloads that do not set one, or enforce specific values so that users cannot change them. This is available for both [native workloads](../../policies/native-workload-policies.md) and [supported workload types](../../policies/supported-workload-type-policies.md).

<Note>

**Note**

Policy enforcement does not reach [externally submitted workloads](../../../workloads-in-nvidia-run-ai/workload-types/supported-features.md#externally-submitted-kubernetes-workloads).

</Note>

## Deployment Considerations

* Selecting a GPU portion using percentages as units for the Request does not guarantee the exact memory size. This means 50% of an A-100-40GB is 20 GB while 50% of an A-100-80GB is 40 GB. To have better control over the exact allocated memory, specify the exact memory size, for example 40 GB.
* The Limit must be equal to or greater than the memory Request.
* The Limit is enforced at runtime only and is never read by the Scheduler. Raising a Limit does not change where a workload is placed.
* Memory loaned above the Request is not guaranteed. A workload operating above its Request can have that memory reclaimed at any time.
* To keep a workload running through a memory reclaim, build the container so that the process accessing the GPU is not the root process of the container. A workload whose root process accesses the GPU directly is terminated when its loaned memory is reclaimed.

### Known Limitations with CUDA 13.4 and Driver r615

The following limitations apply to dynamic GPU memory fractions when running on CUDA 13.4 with driver r615. Both issues will be resolved in a future release of CUDA.

* **Memory is reclaimed from one workload at a time.** When several workloads on a GPU are using memory above their Request, and another workload on that GPU allocates up to its own Request, only one of the bursting workloads receives an OOMKill signal. If that single reclaim does not free enough memory, the allocating workload does not receive the full memory of its Request.
* **A workload allocating within its Request can be terminated.** In some scenarios, a workload allocating memory up to its Request receives a SIGKILL, instead of memory being reclaimed from the workloads operating above their own Request. Administrators can address this using the misc cgroup fallback. Refer to [MPS memory partitioning](https://docs.nvidia.com/deploy/mps/latest/mpsv3-memory-partitioning.html) for the affected scenarios and the configuration.

## Using CLI

To view the available actions, go to the [CLI v2 reference](../../../reference/cli/runai/runai.md) and run according to your workload.

## Using API

To view the available actions, go to the [API reference](../../../rest-api/about-the-rest-api.md) and run according to your workload.
