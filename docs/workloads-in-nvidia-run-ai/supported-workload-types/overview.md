# Overview

NVIDIA Run:ai supports a broad range of workloads from the ML and Kubernetes ecosystem, in addition to its native workloads. These supported workload types are already registered in the platform and ready to use. NVIDIA Run:ai manages them through Karta, so they receive the same scheduling, orchestration, and monitoring capabilities as native workloads. Administrators can also register new workload types.

## Use Cases

* **Run existing ML and Kubernetes workloads** - Submit workloads from the tools your teams already use, such as Kubeflow, Ray, and Tekton, without moving them to a native workload type.
* **Deploy NVIDIA inference frameworks** - Run NVIDIA NIM Services and NVIDIA Dynamo deployments under the same scheduler as the rest of your workloads.
* **Run distributed workloads** - Run workloads such as MPIJob, RayCluster, and LeaderWorkerSet on shared GPU infrastructure.
* **Bring in workloads submitted outside the platform** - Workloads submitted directly to the cluster, for example with `kubectl` or Helm charts, are scheduled and monitored by NVIDIA Run:ai.
* **Support new frameworks** - Register workload types that are not yet supported, without platform updates or code changes.

## Key Features and Capabilities

Feature availability differs by workload type. For a breakdown of the capabilities that each workload type supports, refer to [Supported features](../workload-types/supported-features.md).

* **Pre-registered workload types** - Workload types from NVIDIA, Kubernetes, Kubeflow, Ray, Tekton, Slinky, and other frameworks are available without additional registration. Refer to [Supported workload types](list-of-supported-workload-types.md) for the full list.
* **Scheduling and monitoring** - Supported workload types are scheduled by the NVIDIA Run:ai Scheduler, appear in the Workloads table, and are monitored in the same way as native workloads.
* **Policies and access control** - Workload policies and role-based access control apply to supported workload types. Workload templates and workload assets are not available for them.
* **Default categories** - Each workload type belongs to a Build, Train, or Deploy category by default. The category determines the default priority and preemptibility. Administrators can change the default mapping. Refer to [Workload priority and preemption](../../platform-management/runai-scheduler/scheduling/workload-priority-control.md).
* **Submission with YAML** - Submit a Kubernetes YAML manifest from the NVIDIA Run:ai UI, or submit it outside of NVIDIA Run:ai, for example with `kubectl`.
* **Extensibility with Karta** - Administrators register new workload types by defining how NVIDIA Run:ai interprets the workload structure, maps its status for monitoring, and applies optimization instructions such as gang scheduling.

## Optimize Scheduling and Placement

Supported workload types are configured in the YAML manifest. The following pages describe the fields and annotations that apply:

* [Configure priority and preemption](../../platform-management/runai-scheduler/optimize-workload-scheduling-and-placement/configure-priority-and-preemption.md) - Override the default priority and preemptibility of the workload type.
* [Request GPU memory fractions](../../platform-management/runai-scheduler/optimize-workload-scheduling-and-placement/gpu-sharing/request-gpu-memory-fractions.md) and [Configure GPU compute mode](../../platform-management/runai-scheduler/optimize-workload-scheduling-and-placement/gpu-sharing/configure-gpu-compute-mode.md) - Request a portion of GPU memory and set how workloads share GPU compute.
* [Configure workload-based placement](../../platform-management/runai-scheduler/optimize-workload-scheduling-and-placement/workload-placement/configure-workload-based-placement.md) - Override the topology-aware scheduling of a distributed workload.
* [Configure segment-based placement](../../platform-management/runai-scheduler/optimize-workload-scheduling-and-placement/workload-placement/configure-segment-based-placement.md) - Group workers into segments that are placed together. Supported for PyTorchJob and LeaderWorkerSet.
* [Configure GB200 NVL72 placement](../../platform-management/runai-scheduler/optimize-workload-scheduling-and-placement/workload-placement/configure-gb200-nvl72-placement.md) - Set the placement and the Multi-Node NVLink behavior in GB200 NVL72 node pools.
* [Enabling Spectrum-X networking](../../platform-management/runai-scheduler/optimize-workload-scheduling-and-placement/spectrum-x-networking.md) - Attach Spectrum-X networks to the workload.

For all workload-level scheduling and placement pages, see [Optimize workload scheduling and placement](../../platform-management/runai-scheduler/optimize-workload-scheduling-and-placement/README.md).

## Getting Started

To learn about the supported workload types and their defaults, refer to [Supported workload types](list-of-supported-workload-types.md).

To submit a supported workload type, refer to [Submit supported workload types via YAML](../submit-via-yaml.md).

Administrators can register additional workload types. Refer to [Extending workload support with Karta](extending-workload-support/README.md).

For end-to-end examples, refer to the following tutorials:

* [NVIDIA Dynamo Aggregated Inference Deployment](../tutorials/inference-tutorials/aggregated-dynamo.md)
* [NVIDIA Dynamo Disaggregated Inference Deployment](../tutorials/inference-tutorials/disaggregated-dynamo.md)
* [Model-Free NIM Service Deployment](../tutorials/inference-tutorials/nim-service.md)
