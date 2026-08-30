# Supported Features

This page compares feature support across different workload types in NVIDIA Run:ai. Use it to understand which scheduling, resource management, and platform capabilities are available for each workload type before selecting a workload model or submission method.

* [Native workloads](native-workloads.md) - Fully integrated into the platform - Workspace, Training and Inference.
* [Supported workload types](supported-workload-types.md) - A broad range of workloads from the ML and Kubernetes ecosystem enabled through Karta.
* [Externally submitted Kubernetes workloads](supported-features.md#externally-submitted-kubernetes-workloads) - Workloads submitted outside of NVIDIA Run:ai. These workloads receive only minimal scheduling and platform capabilities.

Feature availability may vary across NVIDIA Run:ai versions and cluster deployments. Refer to this page and the linked documentation for the most up-to-date support details.

## Workload Submission <a href="#workload-submission-methods" id="workload-submission-methods"></a>

|   | Workspace | Standard Training | Distributed Training | Inference | Distributed Inference | Supported workload types |
| --- | --- | --- | --- | --- | --- | --- |
| UI | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ |
| UI (via YAML) | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ |
| API (Workloads v1) | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ |
| API (Workloads v2) | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ |
| CLI | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

## Scheduling and Resource Management

| Functionality | Workspace | Standard Training | Distributed Training | Inference | Distributed Inference | Supported workload types |
| --- | --- | --- | --- | --- | --- | --- |
| [Fairness](../../platform-management/runai-scheduler/scheduling/concepts-and-principles.md#fairness-fair-resource-distribution) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| [Priority and preemption](../../platform-management/runai-scheduler/scheduling/concepts-and-principles.md#priority-and-preemption) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| [Over quota](../../platform-management/runai-scheduler/scheduling/concepts-and-principles.md#over-quota) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| [Node pools](../../platform-management/aiinitiatives/resources/node-pools.md) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| [Bin packing / Spread](../../platform-management/runai-scheduler/scheduling/concepts-and-principles.md#placement-strategy-bin-pack-and-spread) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| [Multi-GPU fractions](../../platform-management/runai-scheduler/resource-optimization/fractions.md) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| [Multi-GPU dynamic fractions](../../platform-management/runai-scheduler/resource-optimization/dynamic-fractions.md) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| [Node level scheduler](../../platform-management/runai-scheduler/resource-optimization/node-level-scheduler.md) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| [Multi-GPU memory swap](../../platform-management/runai-scheduler/resource-optimization/memory-swap.md) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Elastic scaling | ✗ | ✗ | ✓ | ✓ | ✓ | ✓ |
| [Gang scheduling](../../platform-management/runai-scheduler/scheduling/concepts-and-principles.md#gang-scheduling) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| [Network topology-aware scheduling](../../platform-management/aiinitiatives/resources/topology-aware-scheduling.md) | ✗ | ✗ | ✓ | ✗ | ✓ | ✓ |
| [GB200 NVL72 and Multi-Node NVLink domains (MNNVL)](../../platform-management/aiinitiatives/resources/using-gb200.md) | ✗ | ✗ | ✓ | ✗ | ✓ | ✓ |
| [Scheduling rules](../../platform-management/policies/scheduling-rules.md) | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ |

## Operational and Platform Features

| Functionality | Workspace | Standard Training | Distributed Training | Inference | Distributed Inference | Supported workload types |
| --- | --- | --- | --- | --- | --- | --- |
| [Monitoring](../../infrastructure-setup/procedures/system-monitoring.md) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Workload awareness | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| [RBAC](../../infrastructure-setup/authentication/overview.md#role-based-access-control-rbac-in-run-ai) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| [Workload actions (stop/run)](../workloads.md) | ✓ | ✓ | ✓ | ✗ | ✗ | ✓ |
| [Rolling updates](../using-inference/custom-inference.md) | ✗ | ✗ | ✗ | ✓ | ✗ | ✗ |
| [Workload templates](../workload-templates/) | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ |
| [Workload assets](../workload-assets/overview.md) | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ |
| [Workload Policies](../../platform-management/policies/native-workload-policies.md) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

<Info>

**Workload awareness**

Specific workload-aware visibility, so that different pods are identified and treated as a single workload (for example GPU utilization, workload view, dashboards).

</Info>

## Externally Submitted Kubernetes Workloads

Kubernetes workloads can be submitted outside of NVIDIA Run:ai, for example by using `kubectl` directly or Helm charts as part of an [AI application](../../ai-applications/ai-applications.md). These workloads are scheduled by NVIDIA Run:ai and receive full monitoring support, along with a subset of scheduling capabilities.

* Supported scheduling capabilities include:
  * [Fairness](../../platform-management/runai-scheduler/scheduling/concepts-and-principles.md#fairness-fair-resource-distribution)
  * [Priority and preemption](../../platform-management/runai-scheduler/scheduling/concepts-and-principles.md#priority-and-preemption)
  * [Over quota](../../platform-management/runai-scheduler/scheduling/concepts-and-principles.md#over-quota)
  * [Node pools](../../platform-management/aiinitiatives/resources/node-pools.md)
  * [Bin packing / Spread](../../platform-management/runai-scheduler/scheduling/concepts-and-principles.md#placement-strategy-bin-pack-and-spread)
  * [Multi-GPU fractions](../../platform-management/runai-scheduler/resource-optimization/fractions.md)
  * [Multi-GPU dynamic fractions](../../platform-management/runai-scheduler/resource-optimization/dynamic-fractions.md)
  * [Node level scheduler](../../platform-management/runai-scheduler/resource-optimization/node-level-scheduler.md)
  * [Multi-GPU memory swap](../../platform-management/runai-scheduler/resource-optimization/memory-swap.md)
  * [Gang scheduling](../../platform-management/runai-scheduler/scheduling/concepts-and-principles.md#gang-scheduling)
* All [monitoring](../workloads.md#show-hide-details) capabilities are supported including event history, metrics and logs.
