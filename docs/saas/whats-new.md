# What's New

The what's new provides transparency into the latest changes and improvements to NVIDIA Run:ai’s SaaS platform. The updates include new features, optimizations, and fixes aimed at improving performance and user experience.

<Callout intent="info">
For a complete list of deprecations, see [Deprecation notifications](#deprecation-notifications). Deprecated features, APIs, and capabilities remain available for **six months** from the time of the deprecation notice, after which they may be removed.
</Callout>

## Gradual Rollout

SaaS features and bug fixes are gradually rolled out to customers to ensure a smooth transition and minimize any potential disruption. SaaS releases follow a scheduled rollout cadence, typically every two weeks, allowing us to introduce new functionalities in a controlled and predictable manner. **All customers receive the changes within 10 days of the initial release.**

In contrast, hotfixes are deployed as needed to address urgent issues and are released immediately to ensure the stability and security of the service.

## DGX Cloud

Certain features are first made available in fully managed cloud-based deployments provisioned through DGX Cloud. These features are labeled as `DGX Cloud only` and will become available to all customers in future releases.

## Feature Life Cycle

NVIDIA Run:ai uses life cycle labels to indicate the maturity and stability of features across releases:

* <Badge intent="success">`Experimental`</Badge> - This feature is in early development. It may not be stable and could be removed or changed significantly in future versions. Use with caution.
* <Badge intent="warning">`Beta`</Badge> - This feature is still being developed for official release in a future version and may have some limitations. Use with caution.
* `Legacy` - This feature is scheduled to be removed in future versions. We recommend using alternatives if available. Use only if necessary.

## July 2026 Releases

### July 05

#### Product Enhancements

* **Enhanced workload structure and inspection** - Supported workload types include enhanced visibility into workload structure and details directly from NVIDIA Run:ai. `From cluster v2.26 onward`
  * **Workload element YAML** - Displays the live Kubernetes custom resource for a workload, workload element, or pod, including the current spec, status, and metadata. The YAML can be copied or downloaded for further investigation. See [Structure](../workloads-in-nvidia-run-ai/workloads.md#structure) for more details.
  * **Workload details for supported workload types** - Extends the workload Details experience to supported workload types, providing access to configuration, runtime information, resources, and workload structure. See [Show/Hide Details](../workloads-in-nvidia-run-ai/workloads.md#show-hide-details) for more details.
* **Private NGC repository support for AI applications** - AI applications support private NGC Helm repositories. When a private repository is selected as the chart source, users are prompted to provide an NGC API key credential to authenticate with the repository before browsing and selecting charts. See [AI Applications](../ai-applications/ai-applications.md) for more details. `From cluster v2.26 onward`
* **Workloads v2 API enhancements** - The Workloads v2 API supports PersistentVolumeClaims and user credentials as part of the workload specification, allowing these resources to be created and managed alongside workloads. See [Workloads v2](https://run-ai-docs.nvidia.com/api/workloads/workloads-v2) API for more details. `From cluster v2.26 onward`
  * **PersistentVolumeClaims (PVCs)** - Create PersistentVolumeClaims as part of the workload definition, enabling workloads to provision and use storage without requiring pre-existing PVCs.
  * **User credentials** - Reference user credentials directly from workload definitions without exposing secret values in the workload manifest.
* **Suspend and resume workloads** - NVIDIA Run:ai supports suspend and resume operations for any supported workload type that defines suspend behavior, extending workload lifecycle management beyond native NVIDIA Run:ai workloads. Users can suspend and resume supported workloads through both the UI and API, providing a consistent management experience across workload types. Supported workload types include CronJob, Job, PyTorchJob, RayJob, and TFJob. See [Workloads](../workloads-in-nvidia-run-ai/workloads.md) for more details. `From cluster v2.26 onward`

<Callout intent="warning">
**Known Kubernetes Issue**

A known Kubernetes issue can cause batch job workload types (such as Job, PyTorchJob, and RayJob) to get stuck in the **Resuming** state. The issue is fixed in Kubernetes v1.32.9, v1.33.6, v1.34.2, and v1.35.0 (or newer). Upgrade to a fixed Kubernetes release to resolve this.
</Callout>

* **NUMA-aware scheduling for node pools** - Administrators can now enable NUMA-aware scheduling per node pool. When enabled, the Scheduler places workloads so that GPU, CPU, and CPU memory are allocated within the same NUMA node where possible, minimizing cross-NUMA fragmentation and improving performance on multi-NUMA servers such as DGX H100. When disabled, the Scheduler places workloads using aggregate node capacity only, without considering NUMA topology. Requires NFD deployed with the topology updater enabled and a Topology Manager policy other than `none` on the pool's nodes. Disabled by default. See [Configuring NUMA-Aware Scheduling](../platform-management/aiinitiatives/resources/numa-aware-scheduling.md) for more details. `From cluster v2.26 onward`
* **GPU device-level placement strategy for node pools** - Node pools support a separate device placement strategy for GPU pods in addition to the node placement strategy. The device placement strategy (Bin-pack or Spread) determines how GPU workloads are distributed across GPU devices within each node. This helps fractional distributed workloads - for example, bin-pack at the node level to use fewer nodes while spreading at the device level so pods land on different GPUs within a node. See [Node pools](../platform-management/aiinitiatives/resources/node-pools.md#adding-a-new-node-pool) for more details. `From cluster v2.26 onward`
* **Grove topology integration (API)** - NVIDIA Run:ai network topologies can be linked to a corresponding Grove ClusterTopology. Once linked, NVIDIA Run:ai surfaces the Grove topology name and levels through the [Network Topologies](https://run-ai-docs.nvidia.com/api/organizations/network-topologies) API, allowing users to identify the correct topology for their target node pool without needing to understand the underlying Grove topology configuration. See [Accelerating workloads with network topology-aware scheduling](../platform-management/aiinitiatives/resources/topology-aware-scheduling.md#topology-aware-scheduling-for-dynamo-over-grove) for more details. `From cluster v2.26 onward`
* **Workload policies for any supported workload type** - AWorkload policies support any supported workload type, extending policy governance beyond native NVIDIA Run:ai workloads. NVIDIA Run:ai provides built-in policy mappings for a wide range of supported workload types, so administrators can create policies immediately without additional configuration. Policies can define workload requirements, constraints, defaults, and enforcement actions. See [Supported workload types policies](../platform-management/policies/supported-workload-type-policies.md) for more details. To extend policy support to additional workload types introduced to the platform, administrators can define mappings using the Policy Mappings API. See [Workload types policy mapping](../platform-management/policies/workload-type-policy-mapping.md) for more details. <Badge intent="success">`Experimental`</Badge> `From cluster v2.26 onward`
* **Kubernetes 1.36 support** - NVIDIA Run:ai now supports Kubernetes version 1.36. `From cluster v2.26 onward`

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue where the applications list endpoint returned an empty creation date. |
| Fixed an issue where the distributed workload submission form incorrectly set GPU fields for the master replica when the MPI framework was selected, resulting in a submission error. |
| Fixed an issue where the Users minimal read API returned incorrect search results by including results matched on the creator field instead of filtering by the user's name or username. |
| Fixed an issue where a workload policy with <code>canAdd: false</code> for all storage data source types still displayed all storage options in the data source dropdown. |
| Fixed an issue where the <code>runai nodepool list</code> command crashed when the cluster had nodes excluded using the <code>managedNodePoolNames</code> configuration. |
| Fixed an issue where deleting a DataVolume left stale labels on the origin PVC, preventing it from being re-created. |
| Fixed an issue where NIMService workloads using fractional GPU annotations did not display GPU allocation in the NVIDIA Run:ai UI. |
| Fixed an issue where the NodePort address was not correctly displayed for workloads configured with the NodePort networking option. |
| Fixed an issue where the access rules dialog did not display existing users and service accounts in the selection component on the first click. |
| Fixed a security vulnerability related to GHSA-x744-4wpc-v9h2 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2026-42151 with severity HIGH. |
| Fixed security vulnerabilities related to CVE-2026-42586 and CVE-2026-42579 with severity HIGH. |

## June 2026 Releases

### June 09

#### Product Enhancements

* **Full name display for users** - Users now appear by their full name next to their email, in the format **Full Name (email)**, wherever they show up in the platform. First and last names can be set for local users when they are created or edited; names for SSO users are sourced from the identity provider, and the email is shown on its own when no name is available. A new User Minimal Read Access permission (`users-minimal`) provides the minimal access needed to display user names. It is included by default in some predefined roles and can be added to custom roles. See [Users](../infrastructure-setup/authentication/users.md) and [Roles](../infrastructure-setup/authentication/roles.md) for more details.
* **Workload connection endpoints in the Structure tab** - The Structure tab in the workload Details pane now displays a workload's connection endpoints, making it easier to find and access the tools and services a workload exposes. See [Structure](../workloads-in-nvidia-run-ai/workloads.md#structure) for more details. `From cluster v2.25 onward`

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue where the workload Connections dialog displayed the node IP instead of the user-defined tool name. |
| Fixed an issue where a missing runai-go-operator-tls-secret prevented Helm upgrades from completing. |
| Fixed an issue where the CLI accepted unsupported storage flags for distributed inference submissions. |
| Fixed a security vulnerability related to CVE-2026-42945 with severity CRITICAL. |
| Fixed a security vulnerability related to CVE-2026-42154 with severity HIGH. |
| Fixed security vulnerabilities related to CVE-2026-39832, CVE-2026-46595, CVE-2026-39831, CVE-2026-39828, CVE-2026-39833, and CVE-2026-42508 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2025-14459 with severity HIGH. |
| Fixed a security vulnerability related to GHSA-389r-gv7p-r3rp with severity HIGH. |
| Fixed a security vulnerability related to CVE-2026-42501 with severity HIGH. |
| Fixed a security vulnerability related to GHSA-fqw6-gf59-qr4w with severity HIGH. |
| Fixed a security vulnerability related to CVE-2026-39852 with severity HIGH. |

## May 2026 Releases

### May 25

#### Product Enhancements

* **Access rule transition for legacy roles** - Due to the deprecation of some roles, referred to as legacy roles, access rules using them should be transitioned to supported roles. The Access Rules page now surfaces a banner when rules still use legacy roles, and a new dialog lets admins map each legacy role to its replacement. Selected rules are recreated under the new role with the same subject and scope, and the legacy rules are deleted. See [Access Rules](../infrastructure-setup/authentication/accessrules.md) for more details. `From cluster v2.24 onward`
* **Workload element metrics** - The Structure tab now lets you view metrics per workload element. In the expanded view, selecting a workload element scopes the Metrics tab to that element. See [Structure](../workloads-in-nvidia-run-ai/workloads.md#structure) for more details. `From cluster v2.25 onward`

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue where some filters on the Clusters table did not work. |
| Fixed an issue where the GPU profiling metrics toggle in General Settings > Analytics displayed without a label. |
| Fixed a security vulnerability related to CVE-2026-41080 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2026-33811 with severity HIGH. |
| Fixed a security vulnerability related to GHSA-78h2-9frx-2jm8 with severity HIGH. |

### May 14

#### Product Enhancements

* **Actual topology placement per workload element (API)** - For multi-component workloads (such as Dynamo over Grove), the Workloads API now returns the actual topology placement of each workload element, so you can verify where each element's pods landed in the network hierarchy after scheduling. Workload-level actual placement is also now returned for any distributed workload whose node pool is attached to a topology, regardless of whether topology constraints were applied. See [Accelerating Workloads with Network Topology-Aware Scheduling](../platform-management/aiinitiatives/resources/topology-aware-scheduling.md#topology-constraints-visibility) for more details. `From cluster v2.25 onward`
* **Permission-aware General settings** - The General settings now reflects the user's settings permissions. The page is visible only when the user has at least one settings permission with CRUD (Create, Read, Update, Delete) access; if all of the user's settings permissions are read-only, the page is not visible (matching the existing behavior). See [General Settings](../settings/general-settings/) and [Roles](../infrastructure-setup/authentication/roles.md) for more details.
* **Workload structure expanded view** - The Structure tab in the workload Details pane now supports an expanded view, giving you a larger canvas to inspect a workload's hierarchy of workload elements, resource allocations, and pod status. This makes it easier to trace dependencies in workloads with many child elements - for example, multi-service Dynamo or LeaderWorkerSet deployments. See [Structure](../workloads-in-nvidia-run-ai/workloads.md#structure) for more details. `From cluster v2.25 onward`

#### Resolved Bugs

| Description |
| --- |
| Fixed a security vulnerability related to GHSA-xw7x-h9fj-p2c7 with severity CRITICAL. |
| Fixed a security vulnerability related to GHSA-78h2-9frx-2jm8 with severity HIGH. |
| Fixed a security vulnerability related to GHSA-pc3f-x583-g7j2 with severity HIGH. |
| Fixed a security vulnerability related to GHSA-vmg3-7v43-9g23 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2026-4878 with severity HIGH. |
| Fixed a security vulnerability related to GHSA-5jv8-h7qh-rf5p with severity HIGH. |
| Fixed a security vulnerability related to GHSA-mh2q-q3fh-2475 with severity HIGH. |
| Fixed an issue where NIMService external URLs were intermittently unavailable under load. |
| Fixed a security vulnerability related to CVE-2026-22016 with severity HIGH. |
| Fixed an issue where the Event history "Download to CSV/JSON" exported only the current page instead of all records. |
| Fixed an issue where the Event history "Subject Type" filter did not always return matching results. |
| Fixed an issue where the Terminal tab blocked connections to a running pod when the workload phase was not Running. |
| Fixed an issue where the Project telemetry endpoint returned a 500 error for project-scoped users. |

## April 2026 Releases

### April 28

* **Deploy AI applications directly from NGC catalog** - Blueprints (and other Helm charts) from the NVIDIA NGC catalog or via direct URL can be deployed as AI applications through the UI and API, without requiring direct cluster access. Deploying them as AI applications enables fast assembly of the building blocks that power agentic pipelines in a single workflow. You can override Helm values before deployment (for example, request GPU fractions), enabling flexible configuration and faster deployment of complex, multi-component AI applications. See [AI applications](../ai-applications/ai-applications.md) for more details. <Badge intent="success">`Experimental`</Badge> `From cluster v2.25 onward`
* **Topology placement visibility for workloads** - NVIDIA Run:ai exposes the actual topology placement of a workload, making it easy to validate scheduling decisions and troubleshoot performance issues for multi-node workloads. For each workload, you can see the topology name, the actual placement (topology level and value), the requested constraints, and whether those constraints were met. This information is available for both native and supported workload types via the [Workloads](https://run-ai-docs.nvidia.com/api/workloads/workloads) API, and for native workloads in the workload [Details](../workloads-in-nvidia-run-ai/workloads.md#show-hide-details) view in the UI. See [Accelerating workloads with network topology-aware scheduling](../platform-management/aiinitiatives/resources/topology-aware-scheduling.md) for more details. `From cluster v2.25 onward`
* **Enhanced workload structure visibility** - NVIDIA Run:ai now surfaces the full internal structure of [supported workload types](../workloads-in-nvidia-run-ai/submit-via-yaml.md), allowing users to navigate from the top-level workload down to individual elements and pods. For each element, you can inspect its name, type, and parent-child relationships both via the API and in the UI through a new hierarchical **Structure** tab in the workload [Details](../workloads-in-nvidia-run-ai/workloads.md#show-hide-details) panel. Administrators can use Karta to define the structure of additional workload types. This enables newly introduced workloads to expose the same hierarchy and relationships, making them visible and consistent across the API and UI. See [Defining a Karta](../workloads-in-nvidia-run-ai/workload-types/extending-workload-support/defining-a-karta.md) for more details. `From cluster v2.25 onward`
* **DRA support for GPU devices** - NVIDIA Run:ai supports scheduling supported workload types submitted [via YAML](../workloads-in-nvidia-run-ai/submit-via-yaml.md) using Kubernetes Dynamic Resource Allocation (DRA) ResourceClaims for GPU devices, in addition to the existing extended resources method (`nvidia.com/gpus`). This enables more flexible and expressive GPU resource requests, improving scheduling accuracy and laying the foundation for advanced capabilities. DRA-based workloads are fully supported across the NVIDIA Run:ai UI, API, and CLI. To avoid conflicts when mixing DRA and extended resources, it is recommended to use separate node pools for DRA workloads. See [Dynamic resource allocation (DRA)](../platform-management/runai-scheduler/scheduling/concepts-and-principles.md#dynamic-resource-allocation-dra) for more details. `From cluster v2.25 onward`
* **Automatic external endpoint discovery for workloads and AI applications** - After deploying a workload, its externally reachable endpoints are automatically discovered and displayed in the **Connections** column of the [Workloads](../workloads-in-nvidia-run-ai/workloads.md#connections-associated-with-the-workload) and [AI Applications](../ai-applications/ai-applications.md#connections-associated-with-the-ai-application) tables. Endpoints are also exposed as part of the workload object via the [Workloads](https://run-ai-docs.nvidia.com/api/workloads/workloads) API. For AI applications composed of multiple workloads, endpoints from all workloads are aggregated and can be viewed in the UI or queried at the application level via the [AI Applications](https://run-ai-docs.nvidia.com/api/ai-applications/ai-applications) API. Endpoint state is kept in sync automatically as networking resources are created, updated, or deleted. `From cluster v2.25 onward`
* **Hierarchical topology-aware scheduling for Dynamo workloads** - Dynamo workloads support hierarchical, multi-component topology-aware scheduling using Grove topology definitions. Administrators define and link Grove and NVIDIA Run:ai topologies, which are then exposed and used to apply topology constraints at both the workload (Deployment) and component (Service) levels. The NVIDIA Run:ai scheduler enforces these constraints, enabling precise placement across clusters with multiple topologies. This improves performance and avoids over-constrained scheduling compared to previous workload-level topology constraints. See [Accelerating workloads with network topology-aware scheduling](../platform-management/aiinitiatives/resources/topology-aware-scheduling.md#topology-aware-scheduling-for-dynamo-over-grove) for more details. `From cluster v2.25 onward`
* **Endpoint access control for NIM services API** - NVIDIA NIM services deployed via the NIM Operator support endpoint access control for NVIDIA Run:ai users, groups, and service accounts through the API. This enables secure access to inference endpoints and aligns NIM services with enterprise security requirements. See [NVIDIA NIM](https://run-ai-docs.nvidia.com/api/workloads/nvidia-nim) API for more details. `From cluster v2.25 onward`
* **CLI support for NVIDIA NIM services** - NVIDIA NIM services deployed via the NIM Operator can be managed directly from the NVIDIA Run:ai CLI, bringing a full command-line experience without requiring API access. See [CLI command reference](../reference/cli/runai/runai-inference-nim.md) for more details. `From cluster v2.25 onward`
  * Supports the full service lifecycle: submit, update, list, delete, exec, logs, port-forward, and describe.
  * Applies to existing NIM capabilities including autoscaling, fractional GPU, multi-node deployments, and Multi-LLM workloads.
* **Notifications API and Slack integration updates** - The notifications API was redesigned to provide a clearer and more structured model for managing notification state and channels. Email and Slack integrations are handled through dedicated APIs, replacing the previous unified notification channel endpoints. Slack integration is now fully supported through the UI across all deployment types, with the platform managing app creation, connection, and permissions. Legacy notification channel and Slack-specific endpoints are [deprecated](#api-deprecation-notifications) as part of this update. `From cluster v2.25 onward`
* **NVIDIA Run:ai artifacts now available on NGC** - NVIDIA Run:ai artifacts are now published to NVIDIA NGC, including Helm charts, container images, CLI packages, air-gapped packages, and documentation. This provides a unified experience with other NVIDIA products, allowing customers to access and manage NVIDIA Run:ai software using the same NGC tools and credentials. JFrog support is now [deprecated](#jfrog-artifacts). Existing deployments will continue to work during the deprecation period, and customers are encouraged to migrate to NGC. See [Installation](../installation-and-configuration/installation/) for more details. `From cluster v2.25 onward`
* **Host-based routing is now the default** - To support tools such as RStudio and Visual Studio Code without requiring additional configuration after installation, host-based routing is enabled by default. Workload URLs are exposed as subdomains, allowing workloads to run at the root path and avoiding file path issues. Kubernetes clusters require additional setup as part of system requirements configuration. OpenShift clusters require no additional configuration. Clusters upgrading from path-based routing are not affected. For setup instructions, see [System requirements](../installation-and-configuration/installation/install-using-helm/system-requirements.md#host-based-routing-default). For more details, see [External access to containers](../infrastructure-setup/advanced-setup/container-access/external-access-to-containers.md). `From cluster v2.25 onward`
* **Kubernetes Gateway API support** - NVIDIA Run:ai supports the Kubernetes Gateway API as a modern routing infrastructure, providing an alternative to Kubernetes Ingress. The feature is available as an opt-in option in v2.25 and can run alongside Ingress to enable zero-downtime migration. Ingress remains the default, with Gateway API expected to become the default in a future release. See [Kubernetes Gateway API](../infrastructure-setup/advanced-setup/kubernetes-gateway-api.md) for more details. `From cluster v2.25 onward`
* **Vertical Pod Autoscaling (VPA) for cluster services** - NVIDIA Run:ai now supports configuring Vertical Pod Autoscaling (VPA) on cluster-level services to automatically right-size CPU and memory resources based on observed usage. VPA can be configured globally, per scaling group, or per component via `runaiconfig`, with multiple update modes including `Off` (recommendations only), `Initial`, `Auto`, and `InPlaceOrRecreate`. See [Vertical pod autoscaling](../infrastructure-setup/advanced-setup/vpa.md) for more details. `From cluster v2.25 onward`
* **Component version updates** - NVIDIA Run:ai now supports the following. See the [Support matrix](../installation-and-configuration/support-matrix.md) for more details. `From cluster v2.25 onward`
  * OpenShift version 4.21
  * GPU Operator version 26.3
  * Network Operator 26.1
  * DRA driver 25.12
  * Support for Kubernetes version 1.32 and OpenShift version 4.17 has been removed.

#### Resolved Bugs

| Description |
| --- |
| Fixed a security vulnerability related to GHSA-9jj7-4m8r-rfcm with severity HIGH. |
| Fixed a security vulnerability related to CVE-2026-27144 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2026-33810 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2026-32281 with severity HIGH. |
| Fixed a security vulnerability related to GHSA-6v2p-p543-phr9 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2026-32280 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2026-27654 with severity HIGH. |
| Fixed a security vulnerability related to GHSA-hfvc-g4fc-pqhx with severity HIGH. |
| Fixed a security vulnerability related to GHSA-p77j-4mvh-x3m3 with severity CRITICAL. |
| Fixed a security vulnerability related to CVE-2024-12797 with severity HIGH. |
| Fixed a security vulnerability in the OAuth/OpenID login flow. |
| Fixed an issue where the Admin > Event history page returned a 500 error when filtering by Event ID. |
| Fixed an issue where the CLI v2 <code>--preemptible</code> flag did not set preemptibility on workspace workloads. |
| Fixed an issue where the <code>GET /api/v1/workloads/pods</code> endpoint ignored project scope for user API tokens. |
| Fixed an issue where <code>ingressClass</code> was missing from the Minimal Cluster resource. |
| Fixed an issue where enabling <code>enableWorkloadOwnershipProtection</code> prevented the Workload Overseer from enforcing project scheduling rules such as Idle GPU timeout and Workload Duration limits. |
| Fixed an issue where the UI displayed impossible Free GPU values (allocated + free exceeding the node's total capacity) on nodes with fractional GPU allocations. |

### April 15

#### Product Enhancements

* **MNNVL acceleration for YAML-based workload submission** - When submitting YAML-based workloads, you can specify whether Multi-Node NVLink (MNNVL) acceleration is required on a per-workload basis. When required, workloads are scheduled only on MNNVL-capable nodes; otherwise, they can run on any compatible nodes. This provides more precise control over placement and performance for multi-GPU workloads. See [Submit supported workload types via YAML](../workloads-in-nvidia-run-ai/submit-via-yaml.md) for more details. `From cluster v2.24 onward`
* **Distributed inference deployment from the UI** - AI practitioners can now submit and manage distributed (multi-node) inference workloads directly from the NVIDIA Run:ai UI. This enables deployment of large models that exceed the capacity of a single node. Each distributed replica consists of a leader and multiple workers, both configured as part of the submission form. Distributed mode is supported for custom inference only. See [Deploy a distributed inference workload](../workloads-in-nvidia-run-ai/using-inference/distributed-inference.md) for more details. `From cluster v2.22 onward`
* **Deployment of models from Hugging Face** - Models from Hugging Face can now be deployed directly through the NIM serving path using Multi-LLM NIM — a single container that enables deployment of a broad range of models. Previously, only NGC catalog models with dedicated NIM images could use the NIM serving flow; Hugging Face models were restricted to generic serving options. This removes the one-image-per-model dependency and allows a wider range of models to be served through the NIM path, directly from the Run:ai UI. See [Deploy inference workloads with NVIDIA NIM](../workloads-in-nvidia-run-ai/using-inference/nim-inference.md) for more details.
* **Model descriptions in NIM model selection** - The NIM model selection dropdown displays a short description beneath each model name, sourced from the NGC model card. AI practitioners can evaluate available models directly in the NVIDIA Run:ai UI without needing to look them up externally.
* **Trending model indicators in the Hugging Face model catalog** - The Hugging Face model catalog highlights trending models directly in the model selection dropdown. The top 20 trending models are marked with a trending indicator, making it easier to discover popular and emerging models alongside the most downloaded ones.
* **User credentials support in the CLI** - Credentials configured in User settings (My Credentials) can now be managed through the CLI including creating, listing, and deleting credentials. Supported credential types include Generic, Docker registry, and NGC API keys. Credentials can also be used during workload submission as image pull secrets or environment variables, enabling a more streamlined and consistent submission workflow across the CLI, UI, and API. See [CLI command reference](../reference/cli/runai/runai-my-credential.md) for more details. `From cluster v2.22 onward`
* **Extended CLI support for YAML-based workloads -** The CLI supports additional operations for supported workload types submitted via YAML, including `runai workload-type list` and `runai workload-type describe`. This improves visibility and control over supported workload types without requiring access to the UI or direct interaction with Kubernetes resources. See [CLI command reference](../reference/cli/runai/runai-workload-type.md) for more details. `From cluster v2.23 onward`
* **Time-based fairshare UI parameters** - Administrators can now configure key time-based fairshare parameters directly in the UI, including Historical usage weight and Historical usage window. Previously, these settings were only available through the node pool API. This makes it easier to tune how resource fairness is calculated based on historical and current usage, improving control over fairshare behavior without requiring API-level configuration. See [Node pools](../platform-management/aiinitiatives/resources/node-pools.md#adding-a-new-node-pool) for more details. `From cluster v2.24 onward`

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue where workspaces using the TensorBoard environment preset failed to start due to the tensorboard executable not being found in $PATH. |
| Fixed a security vulnerability related to GHSA-23c5-xmqv-rm74 with severity HIGH. |
| Fixed a security vulnerability related to GHSA-p77j-4mvh-x3m3 with severity HIGH. |
| Fixed an issue where the Access Rules API accepted invalid <code>subjectType</code> values without returning an error. |
| Fixed an issue where the emptyDir storage option was not applied to inference workloads even when selected. |
| Fixed an issue where the workload count displayed in the General settings was incorrect when filtering by enabled or disabled. |
| Fixed an issue where the automatic topology constraint applied to a workload targeted the farthest topology level instead of the closest. |
| Fixed a security vulnerability related to CVE-2024-12797 with severity HIGH. |

## March 2026 Releases

### March 22

#### Product Enhancements

* **NGC SaaS tenant creation** - NVIDIA Run:ai SaaS tenants can now be provisioned directly from the NVIDIA NGC portal through a new workflow. Provisioning is initiated by NGC organization admins, with a direct link to the NVIDIA Run:ai tenant console provided upon completion. See [Create and access the control plane tenant](../installation-and-configuration/installation/install-using-helm/install-control-plane.md) for more details.
* **Scope-aware filtering for over-time widgets** - Over-time widgets in the Overview dashboard now support scope-aware filtering by department or project. Users without full cluster permissions can select a single department or project within their allowed scope, and the widgets update accordingly. If only one department or project is available to the user, it is selected automatically. The Overview dashboard is now fully aware of the user’s scope and displays only relevant data.

#### Resolved Bugs

| Description |
| --- |
| Fixed a security vulnerability related to GHSA-m297-3jv9-m927 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2025-61732 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2024-1013 with severity HIGH. |
| Fixed a security vulnerability related to GHSA-72hv-8253-57qq with severity HIGH. |
| Fixed a security vulnerability related to GHSA-5rq4-664w-9x2c with severity HIGH. |
| Fixed a security vulnerability related to GHSA-9h8m-3fm2-qjrq with severity HIGH. |
| Fixed a security vulnerability related to CVE-2026-27142 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2024-41110 with severity HIGH. |
| Fixed an issue where workspace workload submissions intermittently failed with a "Workload failed due to a Network issue" error. |
| Fixed an issue where updating a project with node type names in the payload did not return the node type names in the 200 success response. |

### March 08

#### Product Enhancements

* **Sort by pod name in the Pods modal** - The Pods modal on both the Nodes view and the Workloads view now supports sorting by pod name.
* **Assign access rules during user and service account creation** - When creating a local [user](../infrastructure-setup/authentication/users.md#creating-a-local-user) or [service account](../infrastructure-setup/authentication/service-accounts.md#creating-a-service-account), you can now optionally assign access rules directly within the creation flow. This step is only shown to users with permissions to manage access rules.
* **View logs from previous container instances** - The [Logs](../workloads-in-nvidia-run-ai/workloads.md#logs) view now includes a **Container's logs from** dropdown, allowing you to switch between logs from the current running container instance and the previous one. This makes it easier to investigate restarts and container failures without leaving the workload view.
* **Extended metrics time range for workloads** - The Metrics view now includes a **Since first run** option in the time range selector, displaying metrics from the first time a workload transitioned to the Running phase until the current time.
* **Early removal of legacy Grafana dashboards** - The legacy Grafana-based dashboards have been removed ahead of the originally planned deprecation timeline (April 2026). Security vulnerabilities (CVEs) were identified in the Grafana version in use that could not be remediated through an upgrade at this stage. The new dashboards provide equivalent functionality and are the recommended replacement.

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue where image strings that included a port number in the registry URL were not parsed correctly. |
| Fixed an issue where the NVLink total bytes per pod metric was labeled with GPU metrics labels instead of the expected pod labels. |
| Fixed an issue where the Analytics table displayed incorrect GPU Compute Utilization values for Training and Interactive workloads. |
| Fixed an issue where department data was not synced to the cluster, affecting both department creation and updates. |
| Fixed an issue where the Connect button did not open the workspace URL for workloads submitted through YAML. |
| Fixed a security vulnerability related to CVE-2024-56171 with severity HIGH. |
| Fixed an issue where the dashboard returned a 500 error instead of an informative error message. |
| Fixed an issue where NIM and HuggingFace inference templates failed to submit when a policy defined locked storage instances. |
| Fixed a security vulnerability related to CVE-2025-64756 with severity HIGH. |
| Fixed an issue where updating a project with node type names in the payload did not return the node type names in the 200 success response. |
| Fixed a security vulnerability related to GHSA-5vv4-hvf7-2h46 with severity HIGH. |

## February 2026 Releases

### February 23

#### Product Enhancements

* **New blocked rule for workload policies** - A new `blocked` rule was added to workload policies, allowing administrators to prevent AI practitioners from specifying a value for a field. This can be used to lock security-related configurations from user modification without enforcing a specific default value (for example, supplemental groups). See [Policy YAML reference](../platform-management/policies/policy-yaml-reference.md#rule-types) for more details.
* **UI inactivity timeout updates** - The **Session timeout** setting was renamed to **UI inactivity timeout**. If left blank, users are logged out after 24 hours of inactivity by default. CLI and API access remain unaffected. See [General settings](../settings/general-settings/#security) for more details.
* **Permission-based access to the Overview dashboard** - The Overview dashboard is now available to roles that have at least one of the relevant READ permissions listed below. Dashboard widgets and data are displayed according to each user’s permissions, and widgets that are not applicable to the user’s permissions are automatically hidden. This enhancement also supports custom roles with different permission sets.
  * Clusters READ
  * Node pools READ
  * Nodes READ
  * Projects READ
  * Departments READ
  * Workloads READ

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue where the "Allocation ratio by node pool" widget in the Overview dashboard aggregated unlimited quotas together with other quotas, resulting in incorrect data. |
| Fixed an issue where, after the <code>runai upgrade</code> command completed successfully, the CLI incorrectly prompted the user to run the upgrade again. |
| Fixed an issue where inference workload metrics were not being refreshed correctly. |
| Fixed an issue in the flexible workload submission form where image pull secret section would present shared credentials instead of shared secrets resulting in a failure to submit the workload. |
| Fixed a security vulnerability related to GHSA-jmp9-x22r-554x with severity HIGH. |
| Fixed an issue where department data was not synced to the cluster, affecting both department creation and updates. |
| Fixed an issue in Projects and Departments where GPU utilization/allocation metrics were not displayed if only partial data was available. |
| Fixed a security vulnerability related to GHSA-cv78-6m8q-ph82 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2025-14459 and CVE-2025-64324 with severity HIGH. |
| Fixed an issue where users with the appropriate permissions could not delete system templates in the UI. |
| Fixed an issue where, on rare occasions, "Allocation ratio by node pool" widget would show incorrect data. |
| Fixed an issue where a node pool that included nodes without required topology labels became stuck in Updating after a topology was attached. |
| Fixed an issue where the UI shows the wrong GPU quotas for node pools associated with the “Default” department. |
| Fixed an issue where, on rare occasions, there was a race condition in some of the metrics causing the average GPU utilization to be above 100%. |

### February 10

#### Product Enhancements

* **Redesigned Projects and Departments management** - NVIDIA Run:ai introduces an improved organization management experience that provides better visibility into resource distribution and clearer explainability for how resources are prioritized and allocated across the organization. This update simplifies large-scale organizational management while maintaining full compatibility with NVIDIA Run:ai’s advanced scheduling capabilities. See [Projects](../platform-management/aiinitiatives/organization/projects.md) and [Departments](../platform-management/aiinitiatives/organization/departments.md) for more details. `From cluster v2.20 onward`
  * **Improved organizational visibility** - A clearer, “big picture” view of projects and departments, making it easier to understand how GPU resources are distributed and prioritized.
  * **Bulk management operations** - Administrators can perform bulk actions across multiple organizational units directly from the UI and API, reducing operational overhead.
  * **Clearer resource explainability** - Improved transparency into resource contention and ordering, helping align scheduling behavior with business needs.
* **Increased initialization timeout for inference workloads** - The maximum initialization timeout for inference workloads and templates in the UI has been increased to 720 minutes, allowing workloads with longer startup times, such as large models, to complete successfully without premature failure.
* **Delete predefined environment assets** - Users can now delete predefined environment assets for inference workloads, `chatbot-ui`, `gpt2`, and `llm-server`, giving greater control over environment configuration.
* **UI adjustments for distributed training** - The distributed training workflow (workloads and templates) in the UI now includes a third step for mutual workload setup. The "Allow different setup for the master" toggle has been removed. By default, the master and workers use the same setup unless a policy defines different behavior. This applies to Flexible submission only. See [Train models using a distributed training workload](../workloads-in-nvidia-run-ai/using-training/distributed-training-models.md) for more details.
* **New guided tour for projects and departments** - A built-in tour guides administrators through the projects and departments experience, highlighting key areas and workflows to help them get started quickly.

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue where credentials assets were not displayed in the Credentials table. |
| Fixed an issue where navigating back to the root level in dashboard widgets caused the dashboard to crash. |
| Fixed an issue where workloads submitted with names longer than 63 characters failed to schedule. |
| Fixed a security vulnerability related to CVE-2026-0861 with severity HIGH. |
| Fixed an issue where, when CPU quota and Limit projects from exceeding department quota were both enabled, updating department or project memory quotas to very large values failed with incorrect validation errors, even though the values were valid. |
| Fixed an issue where providing an invalid admin password during installation caused the tenant to become permanently stuck. |
| Fixed an issue where the <code>workload describe</code> command did not display the master specification for distributed workloads. |
| Fixed an issue where an incorrect FQDN used during certificate generation caused errors. |
| Fixed an issue where AI practitioner role did not have read access to policies granted through workload submission permission sets (for example, <code>workspaceEditAccess</code>). |
| Fixed an issue where a race condition during webhook certificate generation caused failures. |
| Fixed a security vulnerability related to CVE-2025-68973 with severity HIGH. |
| Fixed an issue where the Projects/Departments table in the Overview dashboard sometimes showed fewer than 15 projects/departments when their workloads did not have allocated GPUs or were not in Running or Pending status. |
| Fixed an issue where distributed inference workloads could be submitted successfully with an invalid workers value |
| Fixed an issue in the Overview dashboard where the Node pool filter did not work for the Idle workloads table. |
| Fixed an issue where <code>runai template list</code> returned incorrect output when using <code>--page-size</code> and <code>--max-items</code> together. |

## January 2026 Releases

### January 26

#### Product Enhancements

**Pod logs and terminal access** - Accessing pod logs and interactive shells is now faster and more consistent across the Workloads experience. You can open logs or connect to running pods directly from multiple entry points, with pod selection and status kept in sync as you move between views. See [Workloads](../workloads-in-nvidia-run-ai/workloads.md#show-hide-details) for more details. `From cluster v2.24 onward`

* One-click access to logs and terminals from the Pods view and Logs view, with the selected pod opened automatically.
* New Terminal tab for interactive access to pods and containers, including automatic connection when launching from the Workload grid.
* Synchronized pod selection and status across Pods, Logs, and Terminal views, while preserving existing responsive pod name behavior.

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue where running <code>runai logout</code> returned 404 Not Found when the session token had already expired. The logout command now completes successfully and returns a clear message. |
| Fixed an issue where the <code>template describe</code> command did not display the master specification for distributed templates when the master and worker configurations differed. |
| Fixed an issue where image pull secrets marked with <code>exclude=true</code> were not excluded from the workload. |
| Fixed an issue where during password change, the wrong current password logged the user out and redirected them to the login page. |
| Fixed an issue on OpenShift clusters where missing permissions to manage finalizers caused all workloads to remain stuck in Creating state. |
| Fixed a security vulnerability related to CVE-2025-15284 with severity HIGH. |
| Fixed an issue where distributed training workloads were not blocked when master and worker roles used different node pools. |
| Fixed an issue where charts in the Overview dashboard did not render data after the node pool filter was changed. |
| Fixed a security vulnerability related to CVE-2025-32988 with severity HIGH. |
| Fixed an issue where enabling authentication and authorization prevented user metrics from being collected for inference workloads running on Knative and NIM. |

### January 15

#### Product Enhancements

* **Custom roles using permission sets (API)** - Administrators can now create custom roles by combining predefined permission sets using the [Roles](https://run-ai-docs.nvidia.com/api/authentication-and-authorization/roles) API. Permission sets are predefined, supported groupings of permissions that represent all required dependencies for a specific operation (for example, workload submission). Custom roles can then be assigned to users or groups through access rules in the UI or API, alongside the existing NVIDIA Run:ai predefined roles. This allows organizations to tailor access control to their operational needs while maintaining compatibility with the platform’s supported permission model. See [Roles](../infrastructure-setup/authentication/roles.md#custom-roles-api-only) for more details. `From cluster v2.21 onward`
* **NVIDIA NIM service API enhancements** - NVIDIA Run:ai expands support for deploying and managing NIMs through the NVIDIA NIM Operator, providing a standardized, operator-based deployment flow aligned with NIM-native configurations. NIM services are fully managed through the NVIDIA Run:ai API, with UI and CLI support planned for a future release. This capability does not replace the current [NIM deployment flow](../workloads-in-nvidia-run-ai/using-inference/nim-inference.md) and is available as an additional option. See [NVIDIA NIM](https://run-ai-docs.nvidia.com/api/workloads/nvidia-nim) API for more details. `From cluster v2.23 onward`
  * Autoscaling allowing NIM services to scale dynamically based on demand
  * Fractional GPU support, enabling NIM services to request and use partial GPUs for more efficient GPU utilization
  * Multi-node NIM deployments, enabling distributed NIM workloads across multiple nodes
  * Policy enforcement through a dedicated [NVIDIA NIM Policy](https://run-ai-docs.nvidia.com/api/policies/policy) API for consistent governance of NIM services
  * Partial updates via a new PATCH endpoint, allowing targeted changes without resubmitting the full specification
  * [NIM Cache](https://docs.nvidia.com/nim-operator/latest/cache.html) support for model stores, enabling caching of specific LLM or multi-LLM model artifacts to improve startup time and reuse across deployments
* **MNNVL acceleration for supported workload types** - NVIDIA Run:ai now enables running supported workload types on Multi-Node NVLink (MNNVL) domains, including GB200 NVL72 systems. NVIDIA Run:ai applies the appropriate compute domain configuration to ensure workloads are placed and scaled within the same NVLink domain. AI practitioners can submit supported workload types using the [Workloads V2](https://run-ai-docs.nvidia.com/api/workloads/workloads-v2) API and configure their MNNVL preference as part of the workload submission. `From cluster v2.24 onward`
* **Separate priority and preemptibility controls** - Workload priority and preemptibility are configured as two independent parameters across the UI, CLI, and API for native and supported workload types. If no preemptibility value is specified, the existing behavior based on priority is applied automatically. See [Workload priority and preemption](../platform-management/runai-scheduler/scheduling/workload-priority-control.md) for more details. `From cluster v2.24 onward`
* **Authenticated browsing for the NGC catalog** - Browse the NGC catalog and private NGC registries as an authenticated user by selecting your NGC API key credentials during [workload](../workloads-in-nvidia-run-ai/workloads.md#adding-a-new-workload) submission or [template](../workloads-in-nvidia-run-ai/workload-templates/#adding-a-new-template) creation. This provides access to models and containers that require authentication while preserving the option to browse the public container registry. Private NGC registries require administrator configuration in the [General settings](../settings/general-settings/#workloads). <Badge intent="warning">`Beta`</Badge> `From cluster v2.23 onward`
* **NGC API key support for NVIDIA NIM workloads** - NVIDIA Run:ai supports using an NGC API key when deploying NIM workloads to handle both image access and model runtime authentication. A single NGC API key is automatically applied for pulling NIM images from the NGC catalog and injected as a runtime environment variable required for downloading model weights. This streamlines NIM deployment by removing the need for separate pull secrets and runtime credentials while enabling full user self-service for authenticated NIM workloads. See See [Deploy inference workloads from NVIDIA NIM](../workloads-in-nvidia-run-ai/using-inference/nim-inference.md) for more details. `From cluster v2.23 onward`
* **Updated predefined roles** - Predefined roles in NVIDIA Run:ai have been updated to better align with common organizational responsibilities and workflows. See [Roles](../infrastructure-setup/authentication/roles.md#roles-in-nvidia-runai) for more details:
  * Added new predefined roles - AI practitioner, Data and storage administrator, and Project administrator
  * Some existing predefined roles have been deprecated. See [Deprecation notifications](#january-2026) for more details.
* **Reduced access to clusters and node pools** - New APIs and permissions are now available to support reduced access to clusters and node pools - Clusters minimal and Node pools minimal. These APIs allow roles to perform actions such as workload submission while exposing only the minimal required cluster and node pool information (for example, names and IDs), rather than full read access. This improves role design by aligning the visible data with what is actually required for the action being performed. Roles that rely on full read access remain unchanged. Some predefined roles are planned to transition to the new minimal access as described in the [Deprecation notifications](#january-2026).
* **Visibility into workload topology constraints** - Workloads now expose the topology constraints requested during scheduling, providing clear visibility into how network topology influences placement decisions. In the UI, NVIDIA Run:ai native workloads display the requested topology constraints in the workload [Details](../workloads-in-nvidia-run-ai/workloads.md#show-hide-details) view, while the [Workloads](https://run-ai-docs.nvidia.com/api/workloads/workloads) API exposes these fields across native and supported workload types. `From cluster v2.24 onward`
* **Control access scope for inference serving endpoints** - Set whether an inference serving endpoint is accessible externally or restricted to internal cluster traffic when submitting [workloads](../workloads-in-nvidia-run-ai/using-inference/) or creating [templates](../workloads-in-nvidia-run-ai/workload-templates/inference-templates/). Endpoints can be configured as External (public access), if your administrator has configured Knative to support external access, or Internal only, limiting access to in-cluster traffic. `From cluster v2.24 onward`
* **Asset-based workload submission in the CLI** - The NVIDIA Run:ai CLI supports submitting native workloads using workload assets, such as compute resources, environments, and data sources. This allows AI practitioners to reuse the same predefined configurations available in the UI and API, reducing the need for long, flag-heavy CLI commands. Assets can be browsed and inspected directly from the CLI to support consistent and reliable workload submission. See [CLI command reference](../reference/cli/runai/). `From cluster v2.23 onward`
* **Improved fractional GPU support for multi-container pods** - Fractional GPUs are no longer limited to the first container in a pod. You can explicitly specify which container should receive fractional GPU resources using an annotation. If no container is specified, fractional GPUs continue to be associated with the first container by default. See [GPU fractions](../platform-management/runai-scheduler/resource-optimization/fractions.md#setting-gpu-fractions-via-yaml) and [Dynamic GPU fractions](../platform-management/runai-scheduler/resource-optimization/dynamic-fractions.md#setting-dynamic-gpu-fractions-via-yaml) for more details. `From cluster v2.24 onward`
* **Support for elastic distributed workloads on NVLink domains** - Elastic distributed workloads, including auto-scaling and dynamically sized deployments, are fully supported on GB200 NVL72 and Multi-Node NVLink (MNNVL) domains using NVIDIA DRA driver version 25.8 and later. NVIDIA Run:ai automatically applies ComputeDomain configuration and topology-aware scheduling to ensure workloads scale within the same NVLink domain. See [Using GB200 NVL72 and Multi-Node NVLink domains](../platform-management/aiinitiatives/resources/using-gb200.md) for more details. `From cluster v2.24 onward`
* **Native Load Balancer support** - NVIDIA Run:ai exposes LoadBalancer connectivity directly in the UI and CLI when submitting [workloads](../workloads-in-nvidia-run-ai/workloads.md#adding-a-new-workload) or creating [templates](../workloads-in-nvidia-run-ai/workload-templates/#adding-a-new-template) (assuming a load balancer is already installed in the cluster). Configure service ports explicitly and view clearer port configuration and connectivity status. `From cluster v2.24 onward`
* **Time-based fairshare configuration per node pool** - NVIDIA Run:ai supports time-based fairshare to improve long-term fairness in over-quota resource allocation. Instead of relying only on momentary demand, the Scheduler factors in historical GPU usage over time, ensuring that projects with lower recent consumption are given fair access to resources. Usage is tracked continuously, and each project’s GPU-hour consumption is evaluated against its configured weight to balance resource distribution more effectively across projects. Time-based fairshare can be enabled and configured per node pool using the [Node pools](../platform-management/aiinitiatives/resources/node-pools.md#adding-a-new-node-pool) form, with advanced customization available through the [Node pools](https://run-ai-docs.nvidia.com/api/organizations/nodepools) API. `From cluster v2.24 onward`
* **Extended storage visibility in CLI describe commands** - The `describe` command for native workloads supports `--storage`, showing storage resources such as PVCs, ConfigMaps, and Secrets.
* **Default pod anti-affinity** - A new cluster configuration, `global.requireDefaultPodAntiAffinity`, applies a default pod anti-affinity rule to prevent pods from the same service from being scheduled on the same node when possible. This setting is enabled by default. See [Advanced cluster configurations](../installation-and-configuration/post-installation-configuration/cluster-config.md) for more details. `From cluster v2.24 onward`
* **Ingress controller recommendation update** - Due to an announced deprecation by the upstream [NGINX Ingress Controller](https://kubernetes.io/blog/2025/11/11/ingress-nginx-retirement/) project, NVIDIA Run:ai is updating its recommended ingress controller to HAProxy Ingress for supported environments. The Kubernetes Ingress standard remains fully supported. This change affects only the underlying ingress controller implementation and is intended to ensure long-term security, stability, and maintainability. For fresh installations, see [Installation](../installation-and-configuration/installation/install-using-helm/system-requirements.md#kubernetes-ingress-controller). To upgrade from earlier versions, see [Migrate from NGINX to HAProxy Ingress](../installation-and-configuration/installation/install-using-helm/upgrade.md#migrate-from-nginx-to-haproxy-ingress). `From cluster v2.24 onward`
* **Component version updates** - NVIDIA Run:ai now supports Kubernetes version 1.35, OpenShift version 4.20, and GPU Operator version 25.10. Support for Kubernetes version 1.31 and OpenShift version 4.16 has been removed. `From cluster v2.24 onward`
* **Rancher Kubernetes Engine (RKE1)** - RKE1 is no longer supported due to reaching end of life (EOL). RKE2 is the recommended Rancher distribution. See the Rancher [migration guide](https://support.scc.suse.com/s/kb/RKE-to-RKE2-replatforming-instructions-and-FAQs?language=en_US) for more details. `From cluster v2.24 onward`

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue where the <code>--working-dir</code> parameter was ignored for Knative-based inference workloads, causing containers to start in <code>/</code> instead of the specified directory. |
| Fixed an issue where the Fully free GPU devices column displayed <code>-</code> instead of <code>0</code> when no fully free GPU devices were available under fractional GPU allocations. |
| Fixed an issue where projects created or updated during node pool deletion could reference a non-existent node pool and remain NotReady. |
| Fixed issues where readiness probes did not work correctly with serving port authorization in single-node Knative inference workloads. |
| Fixed an issue where the <code>/v1/k8s/setting</code> endpoint returned a 500 error for tenants without clusters, causing the UI to hang instead of redirecting to cluster creation. |
| Fixed an issue where the Node column displayed a sort icon but did not actually sort results in the Running / Requested Pods modal. |
| Fixed an issue where image names longer than the display limit were truncated without providing access to the full name. |
| Fixed an issue causing a delay before newly created clusters could be deleted, leaving the Remove option temporarily unavailable after creation. |
| Fixed an issue where Overview widgets did not update correctly when navigating from departments to projects. |
| Fixed an issue where copying a workload that included node affinity settings caused the re-submission to fail. |
| Fixed a security vulnerability related to CVE-2024-25621 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2025-32988 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2025-65637 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2025-58183 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2025-64756 with severity HIGH. |
| Fixed an issue where, in rare cases, sessions could disconnect due to token refresh handling. |

### January 05

#### Product Enhancements

* **YAML-based workload submission in the UI** - Submit supported workload types defined in YAML directly from the UI. This brings YAML-based submission, previously available through the API, into interactive workflows, allowing you to submit existing Kubernetes or framework-specific manifests while still benefiting from NVIDIA Run:ai scheduling, resource management, and monitoring. See [Submit supported workload types via YAML](../workloads-in-nvidia-run-ai/submit-via-yaml.md) for more details. `From cluster v2.23 onward`
* **Automatic network topology acceleration for supported workloads** - Network topology–aware scheduling is applied automatically to supported distributed workloads submitted [via YAML](../workloads-in-nvidia-run-ai/submit-via-yaml.md). Once a topology is attached to a node pool, NVIDIA Run:ai automatically applies Preferred topology constraints at the lowest available level for the entire workload, optimizing pod placement without additional user configuration. This expands the topology acceleration beyond NVIDIA Run:ai native distributed workloads to additional workload types. See [Accelerating workloads with network topology-aware scheduling](../platform-management/aiinitiatives/resources/topology-aware-scheduling.md) for more details. `From cluster v2.23 onward`
* **AI application–based workload grouping** **in the UI** - NVIDIA Run:ai provides a dedicated AI applications view. This view automatically groups Kubernetes resources deployed via Helm charts into a single logical application, allowing you to list, sort, and filter AI applications. You can also inspect aggregated resource requests and allocations (GPU, CPU, memory) and view the underlying workloads through the Details pane, making it easier to understand and manage complex, multi-component solutions. See [AI applications](../ai-applications/ai-applications.md) for more details. `From cluster v2.23 onward`
*   **Dynamo as a supported workload type** - NVIDIA Run:ai supports Dynamo-based inference workloads through the DynamoGraphDeployment workload type. This allows Dynamo workloads to be deployed, scheduled, and monitored using the same platform capabilities and operational model as native workloads. See [Supported workload types](../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md) for more details. `From cluster v2.23 onward`

    Key capabilities include:

    * **YAML-based deployment and management** - Dynamo workloads can be submitted using [YAML](../workloads-in-nvidia-run-ai/submit-via-yaml.md) from the UI, API, or CLI, without requiring direct cluster access.
    * **Hierarchical gang scheduling** - NVIDIA Run:ai supports hierarchical (multi-level) gang scheduling for Dynamo workloads. Replica groups are scheduled together as sub-gangs, and the entire workload is then scheduled as a single unit. This ensures coordinated placement and execution across all components of the Dynamo workload.
    * **Topology-aware scheduling** - NVIDIA Run:ai applies [topology-aware scheduling](../platform-management/aiinitiatives/resources/topology-aware-scheduling.md) at the workload level to ensure Dynamo workload components are placed according to the underlying cluster topology, improving communication efficiency and execution consistency.
    * **Automatic discovery of Dynamo frontend endpoints** - NVIDIA Run:ai automatically detects Dynamo frontend endpoints and exposes them for access and monitoring.
    * **Unified workload lifecycle and status visibility** - Dynamo workloads are managed, monitored, and tracked with a unified lifecycle and status view.
* **Distributed inference support in the CLI** - Native distributed inference workloads can be submitted and managed directly from the NVIDIA Run:ai CLI. AI practitioners can use familiar NVIDIA Run:ai commands to work with distributed inference workloads, such as list, describe, logs, exec, port-forward, update, and delete. See [CLI command reference](../reference/cli/runai/runai-inference-distributed.md) for more details. `From cluster v2.23 onward`
* **Template-based workload submission in the CLI** - The NVIDIA Run:ai CLI now supports submitting native workloads using existing templates. This allows AI practitioners to reuse the same predefined configurations available in the UI and API, reducing the need for long, flag-heavy CLI commands. Templates can be browsed and inspected directly from the CLI to support consistent and reliable workload submission. See [CLI command reference](../reference/cli/runai/runai-template.md) for more details. `From cluster v2.23 onward`

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue where the Project GET API returned missing limit fields instead of an explicit unlimited value when CPU quotas were enabled. |
| Fixed an issue where the PVC API did not validate <code>claimName</code> uniqueness. |
| Fixed an issue where workloads using multiple GPU fractions were missing GPU utilization and memory metrics. |
| Fixed an issue where the identity manager failed to start when the notification service was disabled. |
| Fixed a security vulnerability related to CVE-2025-58183 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2025-58186 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2025-58187 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2025-61729 with severity HIGH. |
| Fixed an issue where setting a GPU memory limit caused workload creation to fail. |

## December 2025 Releases

### December 14

#### Product Enhancements

* **LeaderWorkerSet (LWS) as a new workload type** - LeaderWorkerSet is now available as a supported workload type. LWS workloads can be deployed and managed [YAML submission](../workloads-in-nvidia-run-ai/submit-via-yaml.md) from the UI, API, or CLI, providing a standardized way to run leader–worker and multi-process workloads across the platform without direct cluster access. See [Supported workload types](../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md) for more details. `From cluster v2.23 onward`
* **Submit workloads from YAML via the CLI** - The CLI now supports submitting workloads directly from a YAML definition using a new `runai workload submit -f` command. This enables declarative workload creation while still allowing key fields to be overridden at submission time. Workloads created from YAML can also be deleted through the CLI, providing a simple way to manage YAML-defined workloads. See [CLI commands reference](../reference/cli/runai/runai-workload-submit.md) for more details.
* **Workloads v2 API update** - The [Workloads v2](https://run-ai-docs.nvidia.com/api/workloads/workloads-v2) API now includes a new PUT endpoint for updating workloads. This endpoint requires submitting a complete workload manifest, which fully replaces the existing workload configuration. `From cluster v2.23 onward`
* **Update NVIDIA NIM services** **API** - The [NVIDIA NIM](https://run-ai-docs.nvidia.com/api/workloads/nvidia-nim) API now includes a new PATCH endpoint for modifying NIM service workloads. This endpoint supports partial updates, allowing you to update only the fields you need without submitting the full workload definition. `From cluster v2.23 onward`
* **New NVIDIA NIM performance histograms added to Metrics** - The Metrics pane now includes two new histograms for NVIDIA NIM metrics. See [NVIDIA NIM metrics](../workloads-in-nvidia-run-ai/workloads.md#nvidia-nim) for more details. `From cluster v2.23 onward`
  * **End-to-end request latency** - Displays request distribution across latency buckets, helping you identify performance patterns and outliers over time.
  * **Time to first token (TTFT)** - Shows the distribution of TTFT across requests, enabling faster detection of model responsiveness issues.
* **Updated** `runai login` **CLI command** - The `runai login` CLI command has been updated to streamline authentication options. These changes align the CLI login modes with the current authentication model and improve clarity in how each method is used. See [CLI commands reference](../reference/cli/runai/runai-login-access-key.md) for more details.
  * `application` login is deprecated and replaced with `access-key`, which is now the supported method for logging in with a service account or user access key.
  * The previous `user` login mode is deprecated and renamed to `password` to more accurately reflect username-and-password authentication.
* **Email invitations for local users** - Administrators can now choose to automatically send an email invitation when creating a local user.
* **New EmptyDir data source for ephemeral storage** - A new EmptyDir data source is now available for one-time configuration during [workload](../workloads-in-nvidia-run-ai/workloads.md#adding-a-new-workload) submission or [template](../workloads-in-nvidia-run-ai/workload-templates/#adding-a-new-template) creation. EmptyDir provides temporary, node-local storage that exists only for the lifetime of the workload.

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue where the PVC API did not validate <code>claimName</code> uniqueness. |
| Fixed an issue so each access rule created or deleted in a batch action is now audited in the events history. |
| Fixed an issue where containers ran as root instead of a non-privileged user. |
| Fixed a permissions issue that allowed users with write-settings permissions to edit a centralized channel. |
| Fixed an issue where inference workload URLs were generated as <code>http</code> instead of <code>https</code>. |
| Fixed an issue where users with L1 Researcher and L2 Researcher roles could not list node pools using the NVIDIA Run:ai CLI. |
| Fixed an issue where the NGC API key asset <code>getById</code> API response was missing the <code>status</code> field. |
| Fixed an issue where users with the correct project permissions could create templates but were blocked from saving edits due to incorrect permission checks. |

### December 03

#### Product Enhancements

* **Service accounts replacing applications in the UI** - The Applications feature has been renamed to Service accounts throughout the UI. All existing functionality remains the same, and existing application records continue to appear unchanged. See [Service accounts](../infrastructure-setup/authentication/service-accounts.md) for more details.
* **Connections column enabled by default in the Workloads grid** - The Connections column is now selected by default in the [Workloads](../workloads-in-nvidia-run-ai/workloads.md#workloads-table) table. When a workload has a single connection its URL is displayed directly, with long URLs automatically shortened using an ellipsis. The URL is clickable and opens the Connections dialog. When multiple connections exist, the table displays the total count.
* **Enhanced workload details view** - The Workload Details tab now provides an enriched and clearer view of workload configuration data. The updated design improves readability and makes it easier to understand how a workload was submitted and configured. Key enhancements include:
  * **Improved layout and data presentation** - Configuration fields are now grouped and displayed more intuitively, helping users quickly find the information they need.
  * **Specification selector** - When a workload contains multiple specs, a new dropdown allows you to easily switch between them.
* **Overview dashboard enhancements** - We’ve made several improvements to the Overview dashboard to strengthen visibility and better support key monitoring workflows. These enhancements also support deprecating the legacy Grafana dashboards. See [Deprecation notifications](#september-2025) for more details:
  * **Enhancements to the Projects/Departments tables** - Timeframe controls for GPU allocation, utilization, and memory utilization are now located within each column. The tables now display up to 15 entries, include GPU quota, separate pending and running workloads, and provide direct links to each project or department.
  * **New pending-time widget** - Introduced a new widget that displays pending workloads count by pending time, helping admins understand how much time their workloads are waiting and also identify the projects/departments experiencing extended pending times.
  * **New guided tour for the Overview dashboard** - A built-in tour now walks administrators and AI practitioners through the key areas of the Overview dashboard, helping them navigate the interface and become familiar with the functionalities that enable them to get the most out of the dashboard.
  * **Additional dashboard improvements:**
    * Added a top-stats **Failed workloads** widget.
    * Added numeric counters to bar graphs, displaying values directly on each bar rather than only in tooltips.
    * Updated the **Workloads by category/type** widget to count **running** workloads only.
    * Added an **Idle time** column to the idle workloads table.
    * Updated widget ordering to separate **current time widgets** from **over time widgets**, improving the analysis process from identifying issue to over time investigation.
* **Consumption report enhancements for GPU hour breakdown** - The Consumption report now includes two new columns, GPU deserved quota hours and GPU over-quota hours. These metrics fully support all existing grouping options, including cluster, node pool, department, and project. This change also supports deprecating the legacy Consumption dashboard. See [Deprecation notifications](#september-2025) for more details.
* **Network topology visibility in clusters and node pools** - The Network topologies modal in the [Clusters](../infrastructure-setup/procedures/clusters.md#network-topologies-associated-with-the-cluster) page displays a new column showing which node pools each topology is associated with. This information is also available in the [Network topologies](https://run-ai-docs.nvidia.com/api/organizations/network-topologies) API. In addition, the node pools list command in the CLI now includes a network topology column, showing the name of the topology assigned to each node pool. `From cluster v2.23 onward`
* **Policy-aware behavior for templates and assets** - Templates and assets that do not fully comply with policy are no longer blocked outright when submitting a workload. Instead, NVIDIA Run:ai now evaluates non-compliance on a case-by-case basis:
  * **Fixable non-compliance** - If compliance can be achieved by adjusting settings during submission, the template or asset can be loaded. The UI highlights what needs to be updated to meet policy requirements.
  * **Non-fixable non-compliance** - If the non-compliant configuration cannot be changed, the template or asset cannot be used, and the relevant policy is displayed to explain the restriction.
  * **Quick workload submit behavior** - Templates with any non-compliance can now be loaded, but quick submit is automatically blocked. The full workload submission flow opens by default, where the UI highlights what needs to be updated to meet policy requirements.
* **Authenticated browsing for the NGC catalog** - You can now browse the NGC catalog as an authenticated user by selecting your own NGC API key credentials during workload submission or template creation. This enables access to models and containers that require authentication while preserving the existing option to browse the public container registry. <Badge intent="warning">`Beta`</Badge> `From cluster v2.23 onward`
* **New PVC events** - NVIDIA Run:ai now emits new PVC asset lifecycle events - Creating, Deleting, and Syncing. These events appear in the PVC’s Event history, extending the visibility introduced in previous releases and giving administrators clearer insight into PVC asset changes and activity over time.

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue that caused charts to remain in a loading state on every data refresh instead of only during the initial load. |
| Fixed an issue where the workloads service could enter a CrashLoopBackOff during upgrade. |
| Fixed a security vulnerability related to CVE-2025-47907 with severity HIGH. |
| Fixed an issue that caused session disconnections. |
| Fixed an issue that caused distributed inference workloads to become unsynchronized. |
| Fixed an issue where the external-workload-integrator on OpenShift entered a constant reconcile loop, causing high CPU utilization. |
| Fixed missing validations for CPU resources when the CPU quota feature flag was disabled, which caused project and department updates to skip required CPU checks. |
| Fixed an issue that could cause the operator to crash during installation due to a race condition in ingress initialization. |
| Fixed an issue where the UI incorrectly prevented creating templates with the same name across different scopes. |
| Fixed an issue where idle GPU timeout rules were incorrectly applied to preemptible workspaces. |

## November 2025 Releases

### November 18

#### Product Enhancements

* **Service accounts replacing applications** - The applications feature has been renamed to service accounts in the API. Service accounts provide the same functionality for programmatic authentication and management but with updated terminology. The deprecation of applications begins with version 2.24 and will continue for two additional releases before removal. Existing application records and endpoints will remain functional during this period to ensure backward compatibility. See the [Applications](https://run-ai-docs.nvidia.com/api/authentication-and-authorization/applications) API (`/api/v1/apps`) for more details.
* **Distributed inference templates (API)** - Distributed inference templates allow you to save workload configurations that can be reused across distributed inference submissions. These templates simplify the submission process and promote standardization across distributed inference workloads. `From cluster v2.22 onward`
* **Policy API for NIM services (API)** - A new Policy API is now available for NVIDIA NIM services, enabling administrators to define and enforce policies that control the behavior of NIM service workloads. These policies help ensure consistent configurations across deployments, improve governance, and simplify management of NIM service workloads. `From cluster v2.23 onward`
* **Autoscaling support in NVIDIA NIM service API** - The NVIDIA NIM service API now supports autoscaling for inference workloads deployed through the NIM Operator. When enabled, NIM services automatically adjust the number of active replicas based on defined metrics, allowing deployments to scale up or down dynamically as traffic changes. `From cluster v2.23 onward`
* **Multi-node NIM support in the NVIDIA NIM service API** - The NVIDIA NIM service API now supports deploying multi-node NVIDIA NIM workloads. `From cluster v2.23 onward`
* **Hugging Face model catalog browsing** - You can now browse and search the Hugging Face model catalog directly from the NVIDIA Run:ai UI and API when creating inference workloads. The live catalog view displays model details such as download count and gated status. For gated models, the platform prompts you to provide a Hugging Face token for access, while open models can be selected without authentication. See [Deploy inference workloads from Hugging Face](../workloads-in-nvidia-run-ai/using-inference/hugging-face-inference.md) for more details.

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue where cluster authentication didn’t use the tenant URL. |
| Fixed an issue where the DCGM metric chart was displayed even when the cluster did not support DCGM metrics. |
| Fixed an issue where resource name validation failed for hugepage resources by enhancing validation rules to properly support hugepages. |
| Fixed an issue where switching between workloads in the workload Details drawer displayed incorrect data, particularly the workload lifespan value. |
| Fixed an issue where the master spec was not inherited when creating a distributed workload from a template. |
| Fixed an issue where policies allowed <code>canEdit: false</code> under <code>attributes</code> without specifying a default value, which incorrectly passed validation. |
| Fixed an issue where the log viewer for distributed workloads displayed only a partial and unsorted list of pods. |
| Fixed an issue where the metric <code>gpu_memory_utilization_avg</code> returned a NaN value. |
| Fixed a security vulnerability related to CVE-2025-62156 with severity HIGH. |
| Fixed an issue where workload submission in the CLI failed when commands contained special characters. |
| Fixed an issue where a mismatch between Helm schema validation and pre-hooks runtime validation code caused <code>clusterConfig.binder.resources</code> errors during upgrades |
| Fixed an issue where workloads logs initially loaded older logs instead of the most recent ones. |
| Fixed an issue where creating or updating a policy failed with an ‘asset Id not found’ error when specifying an imposedAsset. |
| Fixed an issue where the workload controller could delete all running workloads when <code>init-ca</code> generated a new certificate (every 30 days). |
| Fixed an issue where users running Red Hat OpenShift Serverless experienced “Down” status alerts in OpenShift monitoring due to NVIDIA Run:ai Knative ServiceMonitors. |
| Fixed an issue where logs were not displayed in the UI for workloads submitted using the Workloads v2 submission API. |
| Fixed an issue where inference workload metrics did not allow selecting a specific pod for viewing metrics. |
| Fixed an issue where the UI displayed an incorrect access rule status for users with Cloud Operator roles. |
| Fixed an issue where the <code>RunaiAgentPullRateLow</code> and <code>RunaiAgentClusterInfoPushRateLow</code> Prometheus alerts were firing incorrectly without cause. |
| Fixed an issue where a race condition between the NVIDIA Run:ai operator and upgrade/install post hooks caused the upgrade to fail. |
| Fixed an issue where GPU fraction requests were not applied when submitting distributed workloads. |
| Fixed an issue where the NVIDIA Run:ai operator experienced unusually high CPU utilization after upgrade. |
| Fixed an issue where PVCs appeared with the status “Issues found” after upgrading to version 2.22. |

### November 02

#### Product Enhancements

* **Updated credential creation in the UI** - The Credentials page has been redesigned for improved usability. The Access key and Username & password credential types have been consolidated under Generic secret, where each secret format now opens a dedicated form with context-specific input fields. In addition, a dedicated SSH key format has been added under Generic secret for easier configuration of SSH-based authentication. This change simplifies the UI and provides a more streamlined experience for managing credentials. See [Credentials](../workloads-in-nvidia-run-ai/workload-assets/credentials.md#adding-new-credentials) for more details.
* **Min/max worker configuration for PyTorch distributed training** - You can now define the minimum and maximum number of workers directly from the UI when submitting PyTorch distributed training workloads. This provides greater flexibility and control over resource allocation. See [Train models using a distributed training workload](../workloads-in-nvidia-run-ai/using-training/distributed-training-models.md#setting-up-compute-resources) for more details.
* **Audit logging for password resets** - Audit logs now capture all password reset events, including administrator-initiated resets, user-initiated resets, and password-recovery (“forgot password”) actions. This enhancement improves traceability and security visibility across user management workflows.
* **Access keys replacing user applications** - The User applications feature has been renamed to Access keys across the UI and API (`/api/v1/user-applications`). Access keys provide the same functionality for programmatic authentication and management but with updated terminology. The deprecation of User applications begins with version 2.24 and will continue for two additional releases before removal. Existing User application records and endpoints will remain functional during this period to ensure backward compatibility. See [Access keys](../settings/user-settings/user-access-keys.md) for more details.

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue where selecting an environment asset template in the flexible workload form would not present the the capabilities field correctly. |
| Fixed an issue where the API allowed creating a PVC asset without a <code>claimName</code> when <code>existingPVC=false</code>. |
| Fixed an issue where users without permission to create data source assets were blocked from adding one-time data sources during workload submission. |
| Fixed an issue where the NGC API key validation did not allow special characters (<code>-</code>, <code>_</code>, <code>.</code>). Validation now supports these characters as expected. |
| Fixed an issue where removing the logo in Branding settings displayed an empty square. |
| Fixed an issue where pagination in the Node Pool page did not respond. |
| Fixed an issue where department administrators could not include cluster-scope templates in workloads due to incorrect validation of permitted scopes. |
| Fixed an issue where the grace period preemption field in the UI was limited to 5 minutes, even when the workload policy allowed longer durations. |
| Fixed an issue in the CLI installer where the PATH was not configured for all shells. The installer now correctly configures PATH for both zsh and bash. |
| Fixed an issue where policies were not applied when submitting a workload using a template. |
| Fixed an issue where the filterBy department option in the consumption report did not work as expected. |
| Fixed an issue where stale department were not properly removed after deleting a cluster. |
| Fixed an issue that caused conflicts with additional built-in Prometheus Operator deployments in OpenShift. |
| Fixed an issue where running a NIM inference workload on a fractional GPU prevented the Triton server from starting, causing inference endpoint requests to fail. |
| Fixed an issue where incorrect average GPU utilization per project and workload type was displayed in the Projects view charts and tables. |
| Fixed an issue where the <code>updatedBy</code> field of a policy did not show the latest user who updated it. |
| Fixed an issue where the Quota management dashboard occasionally displayed incorrect GPU quota values. |

## October 2025 Releases

### October 19

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue where setting <code>uid</code> or <code>gid</code> to <code>0</code> during environment creation was not allowed. |
| Fixed an issue where users with expired refresh tokens (after 24 hours) could not log in, as the token endpoint returned a 400 error. |
| Fixed an issue where certain policy calls failed when at least one unconfigured cluster existed in the system. |

### October 08

#### Product Enhancements

**Cluster diagnostics collection command** - Added a new CLI command, `runai diagnostics collect-logs`, which gathers diagnostic logs from the Kubernetes cluster for troubleshooting or sharing with NVIDIA Run:ai support. You can collect logs from all or specific namespaces, specify an output directory, and choose whether to include previous pod logs, simplifying cluster debugging and support workflows. See [runai diagnostics](../reference/cli/runai/runai-diagnostics.md) command for more details.

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue where credentials that were not yet synced to the cluster appeared in the credential selection dropdown in Hugging Face and NIM inference workloads. |
| Fixed an issue where YAML submitted workloads were not supported in batch deletion. |
| Fixed a security vulnerability related to CVE-2025-58754 with severity HIGH. |
| Fixed an issue where deleting a project did not remove access rules scoped to that project |

## September 2025 Releases

### September 28

#### Product Enhancements

* **Guided onboarding for first-time admins** - A new onboarding flow helps system and platform administrators quickly get started by walking through cluster installation, setting up SSO and onboarding the first research team, reducing setup complexity and accelerating time to adoption.
* **Guided onboarding experience for new researchers** - On their first login, all new researchers are directed to the Workloads page and guided through creating their first Jupyter Notebook workspace with a short tour. A template is available for immediate launch, helping users get started quickly. The guided tour remains available anytime from the Help menu.
* **Workload extensibility with Karta** - Karta enables organizations to extend NVIDIA Run:ai with new workload types from any ML framework, tool, or Kubernetes resource using a no-code configuration through the Workload Types API. This allows organizations to incorporate emerging AI/ML tools or custom resources without platform updates or code changes. These workloads become immediately available across the organization, empowering teams to innovate and collaborate while benefiting from advanced scheduling and monitoring. See [Extending workload support with Karta](../workloads-in-nvidia-run-ai/workload-types/extending-workload-support/) for more details. <Badge intent="success">`Experimental`</Badge> `From cluster v2.23 onward`
  * No-code onboarding - Register new workload types instantly via the [Workload Types](https://run-ai-docs.nvidia.com/api/workloads/workload-properties#post-api-v1-workload-types) API.
  * Seamless researcher experience - Submit and run workloads using a standard YAML manifest via the [Workloads v2](https://run-ai-docs.nvidia.com/api/workloads/workloads-v2) API.
  * Unified management - Newly added workloads are available to all teams and benefit from the same orchestration and monitoring as native types.
  * Karta-powered integration - Defines how each workload is interpreted and optimized, enabling consistent support for scaling, dependencies, and advanced scheduling.
  * Newly added [workload types](../workloads-in-nvidia-run-ai/workload-types/extending-workload-support/#supported-workload-types) - NIM Services, KServe, and JobSet.
* **New workload template capabilities** - The new templates simplify the workload submission experience by allowing you to launch a workload in a single click, without modifying any settings. In addition, several supporting capabilities have been introduced. See [Workload templates](../workloads-in-nvidia-run-ai/workload-templates/) for more details. `From cluster v2.23 onward`
  * Preset templates - A set of ready-to-use workload templates for NeMo, BioNeMo and PyTorch is now available, enabling you to launch workloads quickly.
  * Linked assets - Templates can now be linked to assets such as environments and compute resources. Any changes to these assets are automatically reflected in the template, ensuring consistency across workloads.
  * Migrating legacy templates - Existing legacy templates can now be migrated into the new workload templates format, allowing teams to retain their saved configurations while taking advantage of new features. This capability is available when the Flexible workload templates setting is toggled on. You will not lose your existing templates - all legacy templates remain available.
* **NGC public registry support for environment images** - Environment images and tags can now be selected directly from the NGC public registry when creating [workloads](../workloads-in-nvidia-run-ai/workloads.md), [environment](../workloads-in-nvidia-run-ai/workload-assets/environments.md) assets and [templates](../workloads-in-nvidia-run-ai/workload-templates/). This provides a streamlined way to access trusted NVIDIA containers without manually entering image URLs. <Badge intent="warning">`Beta`</Badge> `From cluster v2.23 onward`
* **Enhanced logging with per-container support** - Workload logs can be viewed at the container level within each pod through the UI, API and CLI, giving researchers and administrators finer control when monitoring and debugging workloads. In addition, downloaded logs are saved with unique file names that include the workload, pod, container, and timestamp, making it easier to organize and analyze logs from distributed workloads. See [Workloads](../workloads-in-nvidia-run-ai/workloads.md#logs) for more details.
* **Networking metrics** - A new metric, NVLink bandwidth total, has been added to [Nodes](../platform-management/aiinitiatives/resources/nodes.md#resource-utilization) and [Workloads](../workloads-in-nvidia-run-ai/workloads.md#resource-utilization) views in the UI and is also available through the [Nodes](https://run-ai-docs.nvidia.com/api/organizations/nodes) and [Pods](https://run-ai-docs.nvidia.com/api/workloads/pods) APIs. This improves visibility into network utilization, giving teams deeper insight into consumption patterns and resource allocations. `From cluster v2.23 onward`
* **Enhanced Git credential management** - Git data sources can now be configured with Generic secret credentials through the UI or API, with support for SSH private keys. This provides a consistent and secure way to authenticate to repositories, simplifying setup for administrators and enabling users to connect to Git-based workflows more easily. See [Credentials](../workloads-in-nvidia-run-ai/workload-assets/credentials.md#generic-secret) for more details. `From cluster v2.23 onward`
* **Customize your CLI list views** - The new `--columns` flag allows you to tailor the output of `runai list` commands to display only the fields you need, giving you complete control over table views. See [CLI commands reference](../reference/cli/runai/) for more details.
  * Select and order columns - Define exactly which columns to display and in what order.
  * Discover more data - Show useful fields that are not part of the default output.
  * Autocompletion support - Use tab completion to discover and select all available columns for any list command.
* **Distributed inference API enhancements** - The inference API has been extended with support for multi-node deployments, adding autoscaling and rolling updates. These enhancements improve the robustness, scalability, and manageability of distributed inference workloads. See [Distributed inferences](https://run-ai-docs.nvidia.com/api/workloads/distributed-inferences) API for more details. `From cluster v2.22 onward`
* **Distributed inference support for GB200 and MNNVL** - Distributed inference workloads can now take advantage of NVIDIA GB200 NVL72 and other Multi-Node NVLink systems. This enables automatic infrastructure detection, domain labeling, and optimized cross-node communication for high-bandwidth, performance-optimized inference execution. See [Using GB200 NVL72 and Multi-Node NVLink domains](../platform-management/aiinitiatives/resources/using-gb200.md) for more details. `From cluster v2.23 onward`
* **NVIDIA NIM service deployment API** - A new API is available for deploying NVIDIA NIM services, allowing programmatic creation and management of NIM service workloads for easier automation and integration. See [NVIDIA NIM](https://run-ai-docs.nvidia.com/api/workloads/nvidia-nim) API for more details. `From cluster v2.23 onward`
* **Support for Dynamo inference workloads** - Multi-node inference workloads deployed with the NVIDIA Dynamo framework can now be scheduled efficiently using gang scheduling and topology-aware scheduling. This ensures fast startup, low latency, and better resource utilization for disaggregated inference pipelines. <Badge intent="success">`Experimental`</Badge> `From cluster v2.23 onward`
* **Network topology-aware scheduling for distributed workloads** - NVIDIA Run:ai now supports topology-aware scheduling to optimize placement of distributed workloads across data center nodes. By leveraging Kubernetes node labels, the Scheduler can co-locate pods on nodes that are “closer” to each other in the network. This reduces communication overhead, improves workload efficiency, and helps maximize GPU utilization. Once administrators configure the network topology and associate it with node pools, scheduling is applied automatically for distributed workloads submitted through the platform. See [Accelerating workloads with network topology-aware scheduling](../platform-management/aiinitiatives/resources/topology-aware-scheduling.md) for more details. `From cluster v2.23 onward`
* **Scoped access rules** - Users with permissions restricted to a specific scope are now limited to access rules within that scope. This capability is enabled via a tenant setting (`enable_scoped_authorization`) in the [Settings](https://run-ai-docs.nvidia.com/api/authentication-and-authorization/settings) API. Once enabled, the [Access rules](https://run-ai-docs.nvidia.com/api/authentication-and-authorization/access-rules) API returns only the rules within the viewer’s scope (or narrower), and the same scope filtering is applied when viewing access rules in the UI. This ensures access control is aligned with scope boundaries and prevents users from seeing or modifying rules outside their domain.
* **Cluster configuration via Helm values** - Cluster configurations can now be managed directly through the Helm values interface (`clusterConfig`). At runtime, `runaiconfig` is the actual source of truth, representing what is actively running in the cluster. When a Helm upgrade is performed, the Helm values overwrite the existing `runaiconfig`, ensuring alignment with the chart. As a result, clusters configured through a Helm chart should always be managed through Helm. This keeps configurations consistent and predictable across deployments and upgrades. See [Advanced cluster configurations](../installation-and-configuration/post-installation-configuration/cluster-config.md) for more details. `From cluster v2.23 onward`
* **Component version updates** – NVIDIA Run:ai now supports Kubernetes version 1.34. Support for OpenShift version 4.15 has been removed. `From cluster v2.23 onward`
* **Support for ARM on OpenShift** - NVIDIA Run:ai now supports running on ARM-based nodes in OpenShift clusters, expanding deployment flexibility, and allowing organizations to leverage ARM architectures alongside existing x86 infrastructure within their OpenShift environments. `From cluster v2.23 onward`
* **Deleted workloads visible by default** - Deleted workloads are displayed by default in the UI under Workload manager. The toggle to enable this view has been removed, simplifying the experience and making it easier for users to track and review deleted workloads without extra configuration.
* **Direct tool connection** - When a workload has only one configured tool, clicking **Connect** opens the connection directly, without showing a selection menu. If multiple tools are configured, the selection menu will still appear.
* **Custom logo branding** - You can now upload a custom logo to appear in the top-right corner of the NVIDIA Run:ai platform interface. This allows organizations to personalize the platform UI with their own branding. Logos can be uploaded in SVG or PNG format (up to 128 KB) directly from the Branding settings.

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue where external token exchange failed because the API was incompatible with <code>access_tokens</code>. |
| Fixed an issue where updating project resources created through the deprecated Projects API did not work correctly. |
| Fixed an issue where inference workloads failed when using user credentials as an image pull secret. |
| Fixed an issue where, in certain edge cases, removing an inference workload without deleting its revision caused the cluster to panic during revision sync. |
| Fixed an issue where mappers could not be updated in identity providers (IdPs). |
| Fixed a security vulnerability related to CVE-2025-22868 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2025-7425 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2025-49794 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2025-53547 with severity HIGH. |
| Fixed an issue where email notifications configured through User settings were still sent after selecting and then immediately de-selecting all notification types. |
| Fixed an issue where the search and filter logic for NIM models retrieved from the NGC catalog produced inconsistent results, causing some models to appear in unexpected positions in the list. |
| Fixed an issue in the distributed inference policy API where some error messages displayed field names twice. |
| Fixed an issue in CLI v2 where the <code>--master-extended-resource</code> flag had no effect in MPI training workloads. |
| Fixed a security vulnerability related to CVE-2025-22874 with severity HIGH. |

### September 16

#### Product Enhancements

* **AI Application-based workload grouping** - NVIDIA Run:ai now automatically groups related workloads into a single logical application for any workloads deployed via Helm charts. This provides a unified view of complex solutions. Using the API, you can track aggregated resource requests and allocations (GPU, CPU, memory) and monitor the overall application status. In the UI, you can filter the Workloads page by application name to easily see all components of a solution together. See [AI Applications](https://run-ai-docs.nvidia.com/api/ai-applications) API for more details.
* **Flexible inference workload templates** - Flexible workload templates allow you to save workload configurations that can be reused across workload submissions. You can create templates from scratch or base them on existing assets - environments, compute resources, or data sources. These templates simplify the submission process and promote standardization across users and teams. See [Inference templates](../workloads-in-nvidia-run-ai/workload-templates/inference-templates/) for more details.
* **Application access for inference serving endpoints** - All inference workloads - [custom](../workloads-in-nvidia-run-ai/using-inference/custom-inference.md), [Hugging Face](../workloads-in-nvidia-run-ai/using-inference/hugging-face-inference.md) and [NVIDIA NIM](../workloads-in-nvidia-run-ai/using-inference/nim-inference.md), support authorizing applications (in addition to users and groups) when connecting to inference serving endpoints. This enables secure, programmatic access to inference endpoints when accessed externally from the cluster. To use this capability, configure the serving endpoint, authenticate using a token granted by an application, and use the token in API requests to the endpoint.
* **Credential creation during NIM and Hugging Face submissions** - You can now create My credentials of type Generic secret directly in the [NVIDIA NIM](../workloads-in-nvidia-run-ai/using-inference/nim-inference.md) and [Hugging Face](../workloads-in-nvidia-run-ai/using-inference/hugging-face-inference.md) inference workloads submission, avoiding the need to leave the flow to configure authentication.
* **NVIDIA NIM observability metrics** - Observability metrics are now available for NVIDIA NIM inference workloads via the UI and [Workloads](https://run-ai-docs.nvidia.com/api/workloads/workloads) / [Pods](https://run-ai-docs.nvidia.com/api/workloads/pods) APIs, giving teams better visibility into the performance of large language model (LLM) deployments. These metrics can be collected when deploying NIM through NVIDIA Run:ai, NIM operator, Helm chart, or directly via container images (with `run.ai/nim-workload: "true"` label). This enhancement enables more effective monitoring and troubleshooting of NIM-based inference workloads. See [Workloads](../workloads-in-nvidia-run-ai/workloads.md#metrics) and [NIM observability metrics via API](https://run-ai-docs.nvidia.com/api/api-guides/nim-observability-metrics-via-api) for more details. `From cluster v2.23 onward`
* **Application access for workload tools** - Added support for authorizing applications (in addition to users and groups) when connecting to tools. This makes it easier to integrate external systems or services that need direct access to workload tools, providing more flexibility in how connections are managed.
* **PVC details view in data sources** - A new details pane is available when selecting a PVC data source from the Data sources table. The pane shows **Event History** for cluster events, as well as **Details** such as scope, request settings, and partial storage class information. This enhancement gives administrators and AI practitioners greater visibility into PVC usage history and configuration, improving monitoring and debugging. See [Data sources](../workloads-in-nvidia-run-ai/workload-assets/overview.md) for more details. `From cluster v2.23 onward`
* **System policies for workload governance** - By default, every NVIDIA Run:ai account is governed by system policies that establish foundational security controls across all workloads, scopes, and interfaces (UI, API and CLI). These policies ensure consistent workload behavior and prevent unauthorized escalation, and can be viewed as part of the effective policy for any scope. Administrators can create new policies to update these defaults at any desired scope. This flexibility allows easing certain API restrictions when needed, while ensuring every change is explicit and auditable. See [System policies](../platform-management/policies/policies-and-rules.md#system-policies) for more details.
  * **Privileged parameter** - Set to `false` by default and not editable (`canEdit: false`), preventing containers from running with full host access unless explicitly enabled by an administrator.
  * **Grace period** - Defines how long a workload can continue running after a preemption request before termination. The default grace period is 30 seconds, with a system-enforced maximum of 5 minutes across UI, API and CLI submissions. This value can be updated at any scope within the policy hierarchy.
* **Policy synchronization changes** - Starting in version 2.23, control plane policies are no longer synchronized with the cluster. Policies are now stored and enforced only in the control plane, preventing conflicts with outdated cluster policies. See [Workload policies](../platform-management/policies/native-workload-policies.md) for more details. `From cluster v2.23 onward`
* **Keyboard shortcuts for dialogs and forms** - Common keyboard actions across most UI screens and dialogs. Press **Enter** to confirm actions and **Esc** to cancel, making it quicker and easier to navigate workflows.
* **Updated General settings toggles** - The following options are now enabled by default - Flexible workload submission, Flexible workload templates, Data volumes, and Policies.
* **Metrics view updates** - The metrics view has been reorganized with new naming and grouping:
  * Renamed **Default** metrics view to **Resource utilization**
  * Renamed **Advanced** metrics view to **GPU profiling**
  * Inference metrics are shown in a dedicated **Inference** dropdown, available for all inference workloads

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue where the selected node pool was not preserved when switching sections within the workload submission form for all workloads. |
| Fixed an issue where exported CSVs had misaligned columns, causing values (e.g., scope, workload type, creation time, cluster) to shift into incorrect fields. |
| Fixed a security vulnerability related to CVE-2025-5914 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2025-53547 with severity HIGH. |

## August 2025 Releases

### August 31

#### Product Enhancements

* **Expanded cluster role permissions** - Cluster roles have been updated to include `watch` permissions for all supported workload Custom Resource Definitions (CRDs) wherever `get` and `list` permissions were already present. This change ensures compatibility with Kubernetes operators that require `get`, `list`, and `watch` access for proper monitoring and integration with NVIDIA Run:ai workloads. `From cluster v2.21 onward`
* **Policy API for distributed inference** - A dedicated policy API is available for distributed inference enabling fine-grained control over distributed inference workloads. Administrators can define and enforce policies that govern scheduling, scaling, and update behavior, ensuring workloads adhere to organizational requirements and operate consistently across environments. See [Policy](https://run-ai-docs.nvidia.com/api/policies/policy) API for more details.
* **Removed General settings toggles** - The following options have been removed from the General settings page: Job submission, MPI distributed training, Weights & Biases SWEEP integration, and Docker image registry.

#### Resolved Bugs

| Description |
| --- |
| Fixed a security vulnerability related to CVE-2025-47907 with severity HIGH. |
| Fixed a bug which presented the value of the CPU memory in the wrong unit. |

### August 17

#### Product Enhancements

* **Workloads by category over time** - Added a widget to the Overview dashboard that shows the number of workloads per category (e.g., Train, Build, Deploy) over time. This visualization helps identify usage trends, compare activity across categories, and track changes over specific periods. This feature is also supported in the [API](https://run-ai-docs.nvidia.com/api/workloads/workloads#get-api-v1-workloads-telemetry). `From cluster v2.22 onward`
* **Minimum guaranteed runtime for preemptible workloads** - You can now configure a minimum guaranteed runtime for preemptible workloads in node pools via the UI and [API](https://run-ai-docs.nvidia.com/api/organizations/nodepools#post-api-v1-node-pools). This setting specifies the minimum time a preemptible workload will run once scheduled and bound to a node before becoming eligible for preemption. This reduces unexpected interruptions and makes workload execution more predictable. See [Node pools](../platform-management/aiinitiatives/resources/node-pools.md) for more details. `From cluster v2.23 onward`
* **Cluster filter enhancements for Nodes page** - The Nodes page now includes an “All” option in the clusters filter to make it easier to view and manage nodes across multiple clusters at once. When multiple clusters are selected, a Cluster column is displayed by default, showing each node’s associated cluster. Available in both the UI and API.
* **Separate admin toggles for Hugging Face and NVIDIA NIM models** - Previously, enabling Hugging Face and NVIDIA NIM models was managed through a single Models toggle in the Admin [settings](../settings/general-settings/). These options are now separated into distinct toggles, allowing administrators to enable or disable Hugging Face and NIM models independently for finer control over inference model availability.

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue where creating a workspace /training workload returned the error "terminationGracePeriod is not supported in this cluster. |
| Fixed an issue where the non-preemptible priority over-quota warning text was missing from the inference workload creation page. |
| Fixed an issue in the CLI documentation where the <code>--new-pvc</code> description did not clearly indicate that creating a new pvc means creating a new volume that is used only for the duration of the workload's lifecycle. |
| Fixed an issue where the "Get Role by ID" API returned an "insufficient permissions" error for system administrator. |
| Fixed a security vulnerability related to CVE-2025-22868 with severity HIGH. |
| Fixed a security vulnerability related to CVE-2025-7425 with severity HIGH. |

### August 03

#### Product Enhancements

* **Flexible submission form for NVIDIA NIM and Hugging Face workloads** - The flexible submission form is now supported for [NVIDIA NIM](../workloads-in-nvidia-run-ai/using-inference/nim-inference.md) and [Hugging Face](../workloads-in-nvidia-run-ai/using-inference/hugging-face-inference.md) inference workloads. This form allows users to submit workloads using an existing setup or provide custom settings for one-time use, enabling faster, more consistent submissions aligned with organizational policies.
* **Advanced setup form for NVIDIA NIM and Hugging Face workloads** - You can now access advanced configuration options when submitting [NVIDIA NIM](../workloads-in-nvidia-run-ai/using-inference/nim-inference.md) and [Hugging Face](../workloads-in-nvidia-run-ai/using-inference/hugging-face-inference.md) inference workloads, including editing the image and tag, modifying or adding environment variables, and setting workload priority. This provides greater flexibility for adapting workload configurations to specific requirements.
* **Dynamic NVIDIA NIM model list from NGC catalog** - The platform now retrieves the list of available NVIDIA NIM models directly from the NGC catalog using an API call. This ensures the model list remains current and reflects the latest offerings.

#### Resolved Bugs

| Description |
| --- |
| Fixed an issue where the audit logs page filter converted strings to lowercase, causing filtration to fail. |
| Fixed an issue where templates did not appear in the templates table. |
| Fixed an issue where upgrades failed due to changes in the OpenShift monitoring stack. |
| Fixed an issue where the workload flexible submission form did not load the correct default node pools for a project. |
| Fixed an issue where workloads created via CLI could not be cloned in the UI when flexible submission was disabled. |
| Fixed an issue where workloads could not be scheduled if the combined length of the project name and node pool name was excessively long. |
| Fixed an issue where, in OpenShift environments, certain container failures caused workloads to remain in the "Pending" phase instead of transitioning to "Failed". |
| Fixed an issue where enabling <code>enableWorkloadOwnershipProtection</code> for inference workloads caused newly submitted workloads to get stuck. |
| Fixed an issue where the <code>terminationGracePeriodSeconds</code> field accepted values greater than 300 seconds when submitted via the API. |
| Fixed an issue where setting defaults for <code>servingPort</code> fields failed and incorrectly required the container port default as well. |
| Fixed a security vulnerability related to CVE-2025-30749 with severity HIGH. |
| Fixed an issue where the UI logs view called an unsupported API. The fix added a cluster version check to ensure the correct API is used. |
| Fixed an issue where the <code>createdAt</code> timestamp was not updated when a policy was recreated, causing the timestamp to incorrectly reflect the original creation time. |
| Fixed a base image security vulnerability in <code>libxml2</code> related to CVE-2025-49796 with severity HIGH. |
| Fixed an issue where some services were missing from cluster service groups. |
| Fixed an issue where the SAML metadata XML redirect URL was invalid. |

## July 2025 Releases

### July 20

#### Product Enhancements

**Improved status messaging for node pools with undrained nodes** - When creating a node pool or labeling nodes to add to the node pool, nodes that are not fully drained (i.e., still have running workloads) will now trigger clearer status messages in the API and UI. These messages indicate that the node pool cannot include the affected nodes until they are drained and reach a "Ready" state. This helps administrators better understand node pool readiness and identify which nodes are still in transition. `From cluster v2.23 onward`

#### Resolved Bugs

| Description |
| --- |
| Removed Groups resource type from the available permissions in NVIDIA Run:ai roles. |
| Fixed an issue where the Inference Policy View option was missing from the Project and Department pages. |
| Fixed a security vulnerability in <code>runai-container-runtime-installer</code> and <code>runai-container-toolkit</code> related to CVE-2025-49794 with severity HIGH. |
| Fixed an issue where the validation for the number of workers in a policy was not applied correctly. |
| Fixed an issue where negative values were allowed for GPU resource optimization swap size in node pool API. |

## Transition Notices

### July 2026

#### GPU Fractions: Transition to CUDA-Based GPU Sharing

NVIDIA Run:ai is transitioning GPU fractions from its current proprietary sharing engine to technology built directly into NVIDIA CUDA. This moves GPU sharing onto the standard NVIDIA stack for broader industry compatibility and long-term NVIDIA maintenance. The NVIDIA Run:ai experience, APIs, YAML definitions, and workload management workflows are not changing.

* **Current release (cluster v2.26)** - Uses the existing GPU fractions engine. No changes, no action required.
* **Future releases** - CUDA-based fractions will replace the legacy engine. This upgrade will require updated CUDA driver and GPU Operator versions on all GPU nodes. GPU fractions will be disabled until both requirements are in place; all other workloads will continue to run normally. Once the stack requirements are met, fractions re-enable automatically with no further Run:ai upgrade needed.

**What to do now**

* No action is required today; cluster v2.26 is fully supported.
* Review your current CUDA driver and GPU Operator versions to prepare for upcoming stack requirements.
* Plan a maintenance window for the future stack upgrade - GPU fractions will need to be suspended during the upgrade.
* For questions, contact your NVIDIA Run:ai account team.
