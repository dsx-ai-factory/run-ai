# Optimize Workload Scheduling and Placement

The other Scheduling pages explain how the NVIDIA Run:ai Scheduler works and how administrators configure scheduling at the platform level. The pages in this section show how to apply these capabilities to an individual workload, for example when you submit it.

Each page links to the concept page that explains the capability, and each concept page links back to the page here.

Availability depends on the workload type and the submission method. For a comparison of the capabilities that each workload type supports, see [Supported features](../../../workloads-in-nvidia-run-ai/workload-types/supported-features.md).

| To do this | Use this page | Learn how it works |
| --- | --- | --- |
| Set the priority and preemptibility of a workload | [Configure Priority and Preemption](configure-priority-and-preemption.md) | [Workload priority and preemption](../scheduling/workload-priority-control.md) |
| Run a workload as a background workload | [Label a Workload as Background](label-a-workload-as-background.md) | [Background workloads](../scheduling/background-workloads.md) |
| Request CPU compute and memory, and set limits | [Configure CPU and Memory Requests](configure-cpu-and-memory-requests.md) | [CPU compute and memory allocation](../resource-optimization/allocation-of-cpu-and-memory.md) |
| Request a portion of GPU memory | [Request GPU Memory Fractions](gpu-sharing/request-gpu-memory-fractions.md) | [GPU memory fractions](../resource-optimization/memory-fractions.md), [Dynamic GPU memory fractions](../resource-optimization/dynamic-memory-fractions.md) |
| Set how a workload shares GPU compute | [Configure GPU Compute Mode](gpu-sharing/configure-gpu-compute-mode.md) | [GPU compute modes](../resource-optimization/compute-modes.md) |
| Place a distributed workload in a GB200 NVL72 node pool | [Configure GB200 NVL72 Placement](workload-placement/configure-gb200-nvl72-placement.md) | [Using GB200 NVL72 and Multi-Node NVLink domains](../../aiinitiatives/resources/using-gb200.md) |
| Set topology constraints for a distributed workload | [Configure Workload-Based Placement](workload-placement/configure-workload-based-placement.md) | [Workload-based topology placement](../../aiinitiatives/resources/topology-aware-scheduling/workload-level-topology.md) |
| Set topology constraints for groups of workers | [Configure Segment-Based Placement](workload-placement/configure-segment-based-placement.md) | [Segment-based topology placement](../../aiinitiatives/resources/topology-aware-scheduling/segment-based-topology-placement.md) |
| Enable Spectrum-X networking for a workload | [Enabling Spectrum-X Networking for NVIDIA Run:ai Workloads](spectrum-x-networking.md) | - |
