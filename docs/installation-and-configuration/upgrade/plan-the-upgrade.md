# Plan the Upgrade

This page helps you plan an existing NVIDIA Run:ai deployment upgrade. Review the upgrade requirements below and identify which changes affect your environment before you begin.

## Supported Upgrade Paths

NVIDIA Run:ai supports a **maximum gap of four versions** per upgrade; attempting to skip more than four versions in a single upgrade may fail. If your target version is more than four versions ahead of your current version, perform multiple upgrades. Each upgrade can span a maximum of four versions, and each upgrade must complete successfully before you begin the next.

<Note>

**Note**

For Support-Entitled Releases, upgrades are officially supported and tested only for versions within the current [Supported Versions](../../getting-started/product-support-policy.md#supported-versions).

</Note>

To ensure a smooth and supported upgrade process:

* **Align control plane and cluster versions** - For best results, upgrade the control plane and cluster components to the same NVIDIA Run:ai version during the same maintenance window. Keeping versions aligned helps avoid unexpected behavior caused by version mismatches and ensures full compatibility across platform components.
* **Upgrade order** - When performing an upgrade:
  * Upgrade the control plane Helm chart first
  * Upgrade the cluster Helm chart only after the control plane upgrade completes successfully

## Kubernetes and OpenShift Compatibility

Before upgrading NVIDIA Run:ai, verify that your Kubernetes or OpenShift distribution version is compatible with the target NVIDIA Run:ai version. If your current Kubernetes or OpenShift distribution release falls outside the supported range, upgrade your distribution first.

| NVIDIA Run:ai version | Supported Kubernetes versions | Supported OpenShift versions |
| --------------------- | ----------------------------- | ---------------------------- |
| 3.0 (latest)          | 1.35 to 1.37                  | 4.20 to 4.22                 |
| 2.26                  | 1.34 to 1.36                  | 4.18 to 4.22                 |
| 2.25                  | 1.33 to 1.35                  | 4.18 to 4.22                 |
| 2.24                  | 1.33 to 1.35                  | 4.17 to 4.22                 |

For information on supported versions of managed Kubernetes, consult the release notes provided by your Kubernetes service provider. For an up-to-date end-of-life statement see [Kubernetes Release History](https://kubernetes.io/releases/) or [OpenShift Container Platform Life Cycle Policy](https://access.redhat.com/support/policy/updates/openshift).

## Changes to Plan for in v3.0

The following v3.0 changes can affect your upgrade. Review each one and determine whether it applies to your environment.

### NVIDIA CUDA-Based GPU Fractions

GPU fractions remain supported in v3.0, now powered by NVIDIA CUDA instead of the previous proprietary engine. This affects every cluster running fractional GPU workloads:

* Running and pending workloads using GPU fractions must be suspended (native workload types) or deleted (supported workload types) before the NVIDIA Run:ai upgrade begins.
* CUDA 13.4 or later, CUDA driver r615 or later, and GPU Operator 26.7.1 or later are required for fractions to work on v3.0.
* GPU Operator 26.7.1 does not default to CUDA driver r615. Set the driver version explicitly.
* The GPU Operator must be installed with `cdi.nriPluginEnabled` set to `false` (`--set cdi.nriPluginEnabled=false`). GPU memory fractions do not function while the NRI plugin is enabled.
* For [supported workload types](../../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md), Kubernetes manifests using dynamic fractions (a fraction request with a memory limit) require migration to a new annotation. [Native workload types](../../workloads-in-nvidia-run-ai/workload-types/native-workloads.md) keep the same UI, API, and CLI fields - NVIDIA Run:ai converts them to the new annotation format automatically.
* You can choose whether to upgrade the CUDA driver and GPU Operator before or after upgrading NVIDIA Run:ai - the choice affects only how long fractions stay unavailable.

See [Migrate GPU Fractions](migrate-gpu-fractions.md) for the full procedure.

### Unified Controller for All Workload Types

The controllers and CRDs behind native workload types (workspace, training, distributed training, standard inference, distributed inference) are removed in v3.0 and replaced by a single, unified controller and CRD.

Running and pending native workloads that were submitted through a legacy workload type before upgrading are affected by the upgrade, and some capabilities are unavailable until the workload is resubmitted. Identify running and pending workloads on legacy types now so you can plan for resubmission after upgrading.

See [Migrate Legacy Workloads](migrate-legacy-workloads.md) for the affected capabilities and the resubmission procedure.

### GPU Memory Swap Disabled

The node-level scheduler was deprecated in v2.25 and GPU memory swap in v2.26. In v3.0, both are disabled:

* The `swapEnabled` and `nodeLevelSchedulerEnabled` API fields accept only `false`. Setting either to `true` returns an API error.
* Upgrading to v3.0 forces both fields to `false`, even if they were previously `true`. Neither can be re-enabled afterward.
* Other `gpuResourceOptimization` fields, such as `cpuSwapMemorySize` and `reservedGpuMemoryForSwapOperations`, can still be set through the API, but their values are ignored and not sent to the cluster.
* Node pools no longer shows the **Resource utilization optimization** section.

Before upgrading, identify any workloads or node pools relying on GPU memory swap or the node-level scheduler, since neither is available after the upgrade and there is no post-upgrade action that restores them:

* **Workloads** - GPU memory swap workloads are dynamic fractions workloads under the hood. After upgrading, they continue running as standard dynamic fractions workloads, but without swap they no longer serve their original purpose. Delete these workloads before upgrading.
* **Node pools** - No action is required. Swap-related node pool configuration becomes inactive after the upgrade.

### Dynamic Resource Allocation (DRA) Driver

Upgrading the GPU Operator does not install or change the [NVIDIA DRA driver](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/26.7/dra-intro-install.html), and does not change the GPU Operator mode. If your cluster uses a separately installed DRA driver (for example, Multi-Node NVLink deployments such as GB200), keep it when upgrading. Refer to the [Support Matrix](../support-matrix.md) for the supported DRA driver versions. `GPUCluster` mode, in which GPU Operator 26.7 deploys the DRA driver itself, is supported for new installations only.

## Upgrade Workflow

1. [Back up your deployment](backup-before-upgrading.md), including cluster and control plane configuration and any third-party services such as PostgreSQL and metrics storage.
2. If you use GPU fractions, suspend or delete running and pending workloads using GPU fractions and migrate any dynamic fractions YAML as described in [Migrate GPU Fractions](migrate-gpu-fractions.md).
3. [Upgrade NVIDIA Run:ai](upgrade.md): prepare the software artifacts, then upgrade the control plane followed by the cluster.
4. [Migrate legacy workloads](migrate-legacy-workloads.md) if they apply to your environment, including resubmitting running and pending workloads on legacy types and resubmitting any GPU fraction workloads suspended or deleted in step 2.
5. Verify affected functionality - confirm the cluster is **Connected**, GPU fractions are enforcing their limits, and resubmitted workloads behave as expected.
