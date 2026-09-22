# What's New in Version 3.0

The NVIDIA Run:ai v3.0 what's new provides a detailed summary of the latest features, enhancements, and updates introduced in this version. It serves as a guide to help you understand the new capabilities and how to leverage them for improved workload management, resource optimization, and more.

<Note>

**Note**

For a complete list of deprecated features, APIs, and capabilities, see [Deprecation Notifications](deprecation-notifications.md).

</Note>

## Feature Life Cycle

NVIDIA Run:ai uses life cycle labels to indicate the maturity and stability of features across releases:

* <Badge intent="success">Experimental</Badge> - This feature is in early development. It may not be stable and could be removed or changed significantly in future versions. Use with caution.
* <Badge intent="warning">Beta</Badge> - This feature is still being developed for official release in a future version and may have some limitations. Use with caution.
* <Badge intent="note">Legacy</Badge> - This feature is scheduled to be removed in future versions. We recommend using alternatives if available. Use only if necessary.

## Open Source

NVIDIA Run:ai v3.0 is now available as an open source platform, alongside NVIDIA Run:ai Support-Entitled, the supported version of the same platform. Both provide the same platform functionality, while Support-Entitled includes NVIDIA support SLAs and other entitlements.

## Agent Skills and MCP Server

* **Agent Skills** - NVIDIA Run:ai now provides Agent Skills, pre-built and evaluated workflows that teach AI agents how to perform common Run:ai operational tasks using the NVIDIA Run:ai MCP server. Version 3.0 includes two skills: `runai-diagnose-failed-workload`, which diagnoses failed, crashed, or CrashLoopBackOff workloads, and `runai-audit-assigned-quota`, which audits GPU quota fairness across a cluster and identifies reclaimable GPUs per node pool. Refer to [Agent Skills](../../agents/agent-skills.md) for more details. <Badge intent="tip" minimal outlined>From cluster v2.25 onward</Badge>
* **MCP Server** - The NVIDIA Run:ai MCP server expands from 16 to 27 tools, including six new write tools that let AI agents perform actions in Run:ai in addition to retrieving and analyzing platform information. New write capabilities include tools such as `manage_credentials`, `submit_workload`, and `manage_node_pool`. The server works with MCP-compatible clients such as Claude Code, Cursor, VS Code/GitHub Copilot, and Codex. Refer to [Tools Reference](../../agents/tools-reference.md) for the full list. <Badge intent="tip" minimal outlined>From cluster v2.25 onward</Badge>

## Workloads and Scheduling

* **Semi-preemptible workloads** - NVIDIA Run:ai adds a third preemptibility mode alongside preemptible and non-preemptible workloads. A semi-preemptible workload guarantees that its minimum required replicas continue running, while replicas above that minimum behave as preemptible replicas and can be reclaimed when needed. Set this from the workload submission flow, in a workload policy, or directly in YAML. Refer to [Workload Priority and Preemption](../../platform-management/runai-scheduler/scheduling/workload-priority-control.md) for more details. <Badge intent="tip" minimal outlined>From cluster v3.0 onward</Badge>
* **Unified DRA and extended-resource scheduling** - NVIDIA Run:ai previously supported scheduling supported workload types submitted via YAML using Kubernetes Dynamic Resource Allocation (DRA) `ResourceClaim` requests, in addition to the existing extended-resource method. To avoid conflicts, DRA and extended-resource workloads were recommended to use separate node pools. In v3.0, the Scheduler adds a translation layer that converts extended-resource requests into DRA `ResourceClaim` requests, removing this limitation and allowing both request types to share the same node pool.
* **Segment-based topology placement** - Distributed training workloads can now be divided into fixed-size segments so workers within each segment are automatically co-located within the same topology domain. By default, NVIDIA Run:ai applies a preferred topology placement for each segment. Native workload support is available for PyTorch, while supported workload types include PyTorchJob and LeaderWorkerSet (LWS). Refer to [Segment-Based Topology Placement](../../platform-management/aiinitiatives/resources/topology-aware-scheduling/segment-based-topology-placement.md) for more details. <Badge intent="tip" minimal outlined>From cluster v3.0 onward</Badge>
* **GPU fractions on NVIDIA CUDA** - GPU fractions are now powered by NVIDIA CUDA instead of the previous proprietary sharing engine. Version 3.0 also introduces a new `sm-sharing` compute mode for fractional workloads, while `time-slicing` remains the default. Upgrading existing GPU fraction workloads may require additional steps depending on how fractions are configured. See [Upgrade to 3.0](#upgrade-to-30) for upgrade considerations, and refer to [GPU Compute Modes](../../platform-management/runai-scheduler/resource-optimization/compute-modes.md) and [GPU Memory Fractions](../../platform-management/runai-scheduler/resource-optimization/memory-fractions.md) for more details. <Badge intent="tip" minimal outlined>From cluster v3.0 onward</Badge>

## AI Applications

* **Upload a local Helm chart** - When creating an AI application, you can now upload a packaged Helm chart (`.tgz`, up to 10 MB) directly from your local machine, in addition to using an NGC catalog entry or custom URL. Local chart upload is available through the UI and API. Refer to [AI Applications](../../ai-applications/ai-applications.md) for more details. <Badge intent="tip" minimal outlined>From cluster v2.25 onward</Badge>
* **AI application structure diagram** - The AI application details view now includes a structure diagram that shows how workloads connect to the Kubernetes resources they use, including Secrets, PersistentVolumeClaims, and ConfigMaps. You can filter and search the diagram to find specific workloads or resources in more complex applications. Refer to [AI Applications](../../ai-applications/ai-applications.md) for more details. <Badge intent="tip" minimal outlined>From cluster v2.23 onward</Badge>

## Inference

* **vLLM metrics for LLM inference workloads** - NVIDIA Run:ai exposes vLLM metrics for inference workloads running vLLM, including request concurrency, token throughput, latency percentiles, and GPU KV-cache utilization, matching the existing NVIDIA NIM metrics pattern. Refer to [Metrics and Telemetry](../../platform-management/monitor-performance/metrics.md) and [Workloads](../../workloads-in-nvidia-run-ai/workloads.md#metrics) for more details. <Badge intent="tip" minimal outlined>From cluster v3.0 onward</Badge>

## Node Pools and Nodes

* **NUMA-aware scheduling visibility** - NVIDIA Run:ai now surfaces NUMA-aware scheduling prerequisites and per-NUMA-node resource information directly in the Nodes and Node Pools views. You can see the Topology Manager policy and total and allocated GPUs, CPU cores, and CPU memory by NUMA node, making it easier to verify that NUMA-aware scheduling is configured correctly and troubleshoot workloads that cannot be placed because of NUMA fragmentation. Nodes and node pools show a Missing Prerequisites status when the required topology information is unavailable or invalid. Refer to [Configuring NUMA-Aware Scheduling](../../platform-management/aiinitiatives/resources/numa-aware-scheduling.md) for more details. <Badge intent="tip" minimal outlined>From cluster v2.26 onward</Badge>
* **GPU device placement strategy** - The Node Pools table has a new opt-in "GPU device placement strategy" column, showing the Bin-pack or Spread strategy applied at the GPU device level, distinct from the existing node-level placement strategy. Refer to [Node Pools](../../platform-management/aiinitiatives/resources/node-pools.md) for more details.

## Policies

* **Multi-scope policies for supported workload types** - Policies for supported workload types can now be applied at tenant, cluster, department, and project scopes. Each policy rule can specify whether a lower scope is allowed to override it. By default, override is disabled, so a conflicting rule defined at a lower scope does not take effect unless the higher-scope rule explicitly allows it. The policy view also shows the source scope for each effective rule. Refer to [Supported Workload Types Policies](../../platform-management/policies/supported-workload-type-policies.md) for more details. <Badge intent="tip" minimal outlined>From cluster v2.26 onward</Badge>

## Authentication and Authorization

* **Access rules for all authenticated users and service accounts** - Access rules can now use a built-in subject type that applies to all authenticated users and service accounts. This makes it possible to define default access that automatically applies to new users and service accounts at the tenant, department, or project scope. Refer to [Access Rules](../../platform-management/authentication/accessrules.md) for more details. <Badge intent="tip" minimal outlined>From cluster v3.0 onward</Badge>
* **Granular asset permission sets** - New permission sets provide more granular read and edit access to workload templates, environments, compute resources, and credentials. This makes it possible to create custom roles with access to specific asset types without granting the broader asset-management permissions. Refer to [Permissions Reference](../../platform-management/authentication/permissions.md) for more details.

## General Enhancements

* **Bulk report deletion** - You can now select and delete multiple reports from the Reports table in a single action. A confirmation dialog shows how many reports will be deleted, and if part of the operation fails, NVIDIA Run:ai reports the first failed report and the total number of failures. CSV download remains a single-report action. Refer to [Reports](../../platform-management/monitor-performance/reports.md) for more details.
* **Email notification recipient verification** - The Email Notifications verification flow in General Settings now includes an optional recipient email address. This lets you verify that the configured SMTP server accepts mail for a specific recipient without sending an actual email. Refer to [Notifications](../../settings/general-settings/notifications.md) for more details.

## Installation and Deployment

* **`runai-installer` CLI** - NVIDIA Run:ai introduces `runai-installer`, the recommended installation method for both Open Source and Support-Entitled Release installations. The installer uses a single declarative configuration file to validate the installation, run preflight checks, install required prerequisites and features, and deploy the NVIDIA Run:ai control plane and cluster components. It supports connected and air-gapped environments and both same-cluster and multi-cluster deployments. Upgrades are not yet supported by `runai-installer`. Refer to [Guided Installation](../../installation-and-configuration/installation/guided-installation/README.md) for more details.
* **Open Source and Support-Entitled Release editions** - NVIDIA Run:ai v3.0 is available as an Open Source release and as NVIDIA Run:ai Support-Entitled, a supported version of the same platform. Both provide the same platform functionality, use the same installation model, and support connected and air-gapped environments. Open Source installations access NVIDIA Run:ai platform images and Helm charts from NGC without credentials, while Support-Entitled installations require an NGC API key and include NVIDIA support SLAs and other entitlements. Refer to [Installation Options](../../installation-and-configuration/overview.md#installation-options) for more details.
* **NVIDIA Run:ai telemetry** - NVIDIA Run:ai v3.0 introduces an optional telemetry service for Support-Entitled deployments. The service reports platform usage and adoption signals, GPU utilization and allocation, and cluster health information to NVIDIA to support proactive troubleshooting and continued product improvement. Telemetry is available only with Support-Entitled installations. Refer to [Install the Control Plane](../../installation-and-configuration/installation/install-using-helm/install-control-plane.md) for configuration details.
* **Third-party service images from upstream registries** - Connected installations now pull third-party service images directly from their official upstream registries, including Docker Hub, Quay, and `registry.k8s.io`. This applies to both Guided Installation and Manual Helm, for Open Source and Support-Entitled installations. No registry credentials are required for these upstream images. For air-gapped installations, third-party service images are included in the same image bundle as the NVIDIA Run:ai platform images and loaded into the internal registry. Refer to [Third-Party Services](../../installation-and-configuration/system-requirements/third-party-services.md) for more details.
* **New air-gapped image workflow** - Air-gapped installations now use a single bundle containing all required NVIDIA Run:ai platform and third-party service images. The bundle is created on a connected machine using `runai-installer airgap pack`, then transferred into the air-gapped environment, where `airgap load` pushes the images to the internal mirror registry and generates the Helm values used for installation. This workflow applies to both Guided Installation and Manual Helm. Refer to [Air-Gapped Installation](../../installation-and-configuration/installation/guided-installation/air-gapped/preparations.md) for more details.

## Configuration

* **Consolidated `anyworkload-controller` configuration** - The `workload-controller` and `inference-workload-controller` are consolidated under a single `anyworkload-controller` in v3.0, and the cluster configuration has been updated accordingly so settings that were previously defined separately by workload type are now configured under `anyworkload-controller`. Refer to [Cluster Configuration](../../installation-and-configuration/configuration/cluster-config.md) for more details. Existing native workloads submitted before the upgrade may be affected until they are resubmitted. See [Upgrade to 3.0](#upgrade-to-30) for upgrade considerations. <Badge intent="tip" minimal outlined>From cluster v3.0 onward</Badge>
* **OpenShift Route TLS passthrough** - On OpenShift, NVIDIA Run:ai exposes workload URLs through Route objects. By default, Routes use edge termination, where TLS terminates at the OpenShift Router and traffic to the workload pod is plain HTTP. Version 3.0 adds a passthrough termination option, which terminates TLS inside the pod and preserves end-to-end encryption from the client to the workload. The setting applies cluster-wide to all workload Routes created by NVIDIA Run:ai. Refer to [External Access to Containers](../../installation-and-configuration/configuration/networking/external-access-to-containers.md#openshift-route-tls-termination) for configuration requirements and impact on existing workloads. <Badge intent="tip" minimal outlined>From cluster v3.0 onward</Badge>

## Upgrade to 3.0

Upgrading to v3.0 requires administrators to perform additional steps beyond the base upgrade procedure. See [Plan the Upgrade](../../installation-and-configuration/upgrade/plan-the-upgrade.md) for the specific changes to plan for, and [Upgrade NVIDIA Run:ai](../../installation-and-configuration/upgrade/upgrade.md) for the base upgrade procedure.

### Upstream Third-Party Images

NVIDIA Run:ai v3.0 changes how third-party service images are sourced. For connected environments, third-party service images are now pulled directly from their official upstream registries, including Docker Hub, Quay, and `registry.k8s.io`. Before upgrading, ensure that the required registries are accessible from the cluster and back up your NVIDIA Run:ai configurations and third-party services. See [Back Up Before Upgrading](../../installation-and-configuration/upgrade/backup-before-upgrading.md) and [Third-Party Services](../../installation-and-configuration/system-requirements/third-party-services.md) for more details.

### NVIDIA CUDA-Based GPU Fractions

GPU fractions remain supported in v3.0, now powered by NVIDIA CUDA instead of the previous proprietary engine. This makes the capability part of the standard NVIDIA stack, with broader compatibility and long-term maintenance by NVIDIA. Native workload types use the same UI, API, and CLI fields as before; supported workload types require a manifest change only if their manifest sets a fraction limit. See [Migrate GPU Fractions](../../installation-and-configuration/upgrade/migrate-gpu-fractions.md) for prerequisites and the full procedure.

### Unified Controller for All Workload Types

The controllers and CRDs behind native workload types (workspace, training, distributed training, standard inference, distributed inference) are removed in v3.0 and replaced by a single, unified controller and CRD.

Native workloads submitted before upgrading through a legacy workload type - workspace, training, distributed training, standard inference, or distributed inference - continue running after the upgrade, but lose some capabilities until resubmitted. See [Migrate Legacy Workloads](../../installation-and-configuration/upgrade/migrate-legacy-workloads.md) for the affected capabilities and the resubmission procedure.

### Dynamic Resource Allocation (DRA) Driver

A fresh GPU Operator installation installs the NVIDIA DRA driver automatically, but upgrading an existing GPU Operator deployment does not. If your cluster requires the DRA driver (for example, Multi-Node NVLink deployments such as GB200), install it manually after upgrading the GPU Operator. See [Plan the Upgrade](../../installation-and-configuration/upgrade/plan-the-upgrade.md) for details.

### GPU Memory Swap Disabled

Upgrading to v3.0 forces GPU memory swap and the node-level scheduler off, even if they were previously enabled, and neither can be re-enabled afterward. The node-level scheduler was deprecated in v2.25 and GPU memory swap in v2.26. Snapshot, the planned replacement capability, is targeted for v3.1 and is not available in v3.0. See [Plan the Upgrade](../../installation-and-configuration/upgrade/plan-the-upgrade.md) for what to check before upgrading.

### New Air-Gapped Upgrade Workflow

Air-gapped upgrades in v3.0 use the new image-bundle workflow. `runai-installer airgap pack` creates a single bundle on a connected machine containing the NVIDIA Run:ai platform and third-party service images required for the target release. After the bundle is transferred into the air-gapped environment, `airgap load` pushes the images to the internal mirror registry and generates the Helm values required for the upgrade. See [Upgrade NVIDIA Run:ai](../../installation-and-configuration/upgrade/upgrade.md) for the full upgrade procedure.

## Support Matrix

NVIDIA Run:ai v3.0 updates supported versions across Kubernetes distributions, NVIDIA operators, and framework operators:

* **NVIDIA GPU Operator 26.7** - The maximum supported version is now 26.7. With GPU Operator 26.7, the DRA driver is installed by the GPU Operator instead of separately.
* **Prometheus / Kube-Prometheus Stack 3.13.1 / 88.6.5** - The minimum supported versions are now 3.13.1 and 88.6.5.
* **Knative Serving 1.23** - The maximum supported version is now 1.23.
* **Kubeflow Training Operator 1.9.4** - Version 1.9.4 is now supported.
* **MPI Operator 0.8.2** - The maximum supported version is now 0.8.2.
* **LeaderWorkerSet (LWS) 0.10.0** - The maximum supported version is now 0.10.0.
* **Upstream Kubernetes 1.35–1.37** - The supported range is now Kubernetes 1.35 through 1.37. The same range applies to supported managed Kubernetes distributions that track upstream Kubernetes.
* **Red Hat OpenShift 4.20–4.22** - The supported range is now OpenShift 4.20 through 4.22.

Refer to [Support Matrix](../../installation-and-configuration/support-matrix.md) for the complete list of supported versions.

## Documentation Changes

The [Installation and Configuration](../../installation-and-configuration/overview.md) documentation has been restructured to consolidate setup content that was previously spread across Getting Started, Advanced Setup, and Infrastructure Procedures.

* **Unified Installation and Configuration section** - Installation content that previously lived under Getting Started, configuration content that was under Advanced Setup, and ongoing operational procedures that were under Infrastructure Procedures are now consolidated under a single Installation and Configuration section.
* **System Requirements is now installation-path independent** - With v3.0 supporting both Guided Installation and Manual Helm, preparation of the Kubernetes cluster and surrounding infrastructure is no longer documented inside the Helm installation flow. The planning content is now split into focused pages:
  * **Deployment Model** - Same-cluster and multi-cluster deployment planning.
  * **Infrastructure Requirements** - Hardware, operating system, Kubernetes distribution, storage, permissions, and other infrastructure requirements that were previously spread across control plane and cluster system requirements.
  * **Networking Setup** - Load balancer, ingress, FQDNs, DNS, TLS certificates, and related networking preparation that previously appeared in cluster system requirements.
  * **Third-Party Services** - Services deployed with NVIDIA Run:ai and options for using externally managed instances.
  * **External Access to Critical Services** - The former Network Requirements content has been reorganized under this page. It covers the external connectivity NVIDIA Run:ai requires, including access to NGC, upstream registries, control plane and cluster communication, and other external services.
* **Manual Helm installation flow reorganized** - The Manual Helm installation now starts with a single Preparations page that brings together setup tasks that were previously spread across preparation, control plane system requirements, and cluster system requirements. This includes NVIDIA Run:ai registry credentials, TLS certificates, namespaces, and other installation prerequisites. The flow then continues to Prerequisites and Features, followed by the control plane and cluster installation procedures. Upgrade content has been moved out of the Helm installation flow into a separate Upgrade section.
* **Prerequisites and Features are now separate from System Requirements** - Workload- and environment-specific components that previously appeared under cluster system requirements are now documented under Prerequisites and Features, including the GPU Operator, Prometheus, Network Operator, and components required for supported workload types.
* **Aligned to the Day 0, Day 1, and Day 2 lifecycle** - The new structure follows the setup lifecycle:
  * **Day 0: Planning** - Plan the deployment and prepare the Kubernetes cluster and surrounding infrastructure required before installing NVIDIA Run:ai.
  * **Day 1: Installation and Configuration** - Install NVIDIA Run:ai using Guided Installation or Manual Helm, then configure the platform for the environment.
  * **Day 2: Operations** - Perform ongoing operational tasks such as cluster management, scaling, high availability, node maintenance, backup and restore, monitoring, and logging.

Refer to [Installation and Configuration](../../installation-and-configuration/overview.md) for the updated navigation.
