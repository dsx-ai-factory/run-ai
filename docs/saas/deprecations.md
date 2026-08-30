# Deprecation Notifications

Features and capabilities that are deprecated or scheduled for removal in NVIDIA Run:ai SaaS.

<Callout intent="info">
Deprecated features, APIs, and capabilities remain available for **six months** from the time of the deprecation notice, after which they may be removed.
</Callout>

### July 2026

#### GPU Memory SWAP

GPU Memory SWAP will be deprecated and removed from future releases. GPU Memory SWAP is being replaced by Snapshot, an improved NVIDIA-native capability built on CUDA Checkpoint / Restore. Snapshot captures the full state of a Kubernetes Pod, releases idle GPU and CPU memory, and reduces cold-start time when workloads resume. No immediate action is required as part of this release. GPU Memory SWAP remains fully supported in NVIDIA Run:ai cluster v2.26.

#### TensorFlow, XG Boost, and JAX

The TensorFlow, XG Boost, and JAX distributed frameworks are deprecated and will be removed in a future release. These frameworks are based on Kubeflow Training Operator v1 CRDs, which are deprecated upstream. As a result, submitting workloads using these frameworks as native NVIDIA Run:ai distributed workload types is also deprecated. See [Distributed training](../installation-and-configuration/installation/install-using-helm/system-requirements.md#distributed-training) for supported frameworks.

#### Resource Interfaces to Kartas

The Resource Interface feature has been renamed to Karta across NVIDIA Run:ai. The `resourceInterfaces` API field is deprecated and will be removed in a future release. See [Extending workload support with Karta](../workloads-in-nvidia-run-ai/workload-types/extending-workload-support/) for more details.

#### Node Type

The Node Type feature is deprecated and will be removed in a future release. To control workload placement on specific nodes, use workload policies with node affinity instead of the `nodeType` field in organization unit or workload definitions. See [NVIDIA Run:ai native workload policies](../platform-management/policies/native-workload-policies.md) for more details.

#### Legacy Workload Submission

The legacy workload submission form is deprecated and will be removed in a future release. Submitting workloads through the legacy form is being replaced by flexible workload submission, which lets you select an existing setup or start from scratch, review existing setups, and understand policy definitions. We recommend transitioning to flexible workload submission, which is enabled by default. If it has been disabled, re-enable it under **General settings** → Workloads → Flexible workload submission.

#### API Deprecation Notifications <a href="#api-deprecation-notifications-june-2026" id="api-deprecation-notifications-june-2026"></a>

**Deprecated Endpoints**

| Deprecated Endpoint          | Replacement Endpoint             |
| ---------------------------- | -------------------------------- |
| `/api/v1/logo`               | `/api/v1/branding/settings/logo` |
| `/api/v1/org-unit/node-type` | N/A                              |

**Deprecated Parameters**

| Endpoint                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | Deprecated Parameter                           | Replacement Parameter                                                                                       |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| `/api/v2/workloads`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | `resourceInterfaces`                           | `kartas`                                                                                                    |
| `/api/v1/node-pools`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | `schedulingConfiguration.placementStrategy`    | <p><code>gpuPodsPlacementStrategy</code> + <code>cpuPodsPlacementStrategy</code><br></p>                    |
| `/api/v1/node-pools`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | `gpuResourceOptimization`                      | N/A                                                                                                         |
| <ul><li><code>/api/v1/workloads/distributed</code></li><li><code>/api/v2/policy/distributed</code></li><li><code>/api/v1/workload-templates/distributed</code></li></ul>                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | `DistributedFramework`: `TF`, `XGBoost`, `JAX` | N/A                                                                                                         |
| <ul><li><code>/api/v1/asset/workload-template</code></li><li><code>/api/v1/org-unit/departments</code></li><li><code>/api/v1/org-unit/projects</code></li><li><code>/api/v1/workload-templates</code></li><li><code>/api/v1/workload-templates/distributed</code></li><li><code>/api/v1/workload-templates/inferences</code></li><li><code>/api/v1/workload-templates/trainings</code></li><li><code>/api/v1/workload-templates/workspaces</code></li><li><code>/api/v1/workloads/distributed</code></li><li><code>/api/v1/workloads/distributed-inferences</code></li><li><code>/api/v1/workloads/inferences</code></li><li><code>/api/v1/workloads/trainings</code></li><li><code>/api/v1/workloads/workspaces</code></li><li><code>/api/v2/policy/distributed</code></li><li><code>/api/v2/policy/distributed-inferences</code></li><li><code>/api/v2/policy/inferences</code></li><li><code>/api/v2/policy/trainings</code></li><li><code>/api/v2/policy/workspaces</code></li></ul> | `nodeType`                                     | Use a workload policy node selector to target node types instead - `nodeAffinityRequired.nodeSelectorTerms` |

### April 2026

#### JFrog Artifacts

Support for JFrog as an artifact registry is now deprecated. JFrog will continue to be supported for existing deployments, but it will be removed in a future release within approximately one year. Customers are encouraged to migrate to NVIDIA NGC, which is now the recommended artifact registry for NVIDIA Run:ai.

#### TGI Server - Hugging Face

The Text Generation Inference (TGI) server used in the Hugging Face flow is deprecated. TGI is in maintenance mode and expected to reach end-of-life. Existing deployments will continue to work, but users should transition to other supported inference options.

#### Administrator CLI

The Administrator CLI (`runai-adm`) is deprecated. The two tasks it supported are now handled as follows:

* **Log collection** - Use `runai diagnostics collect-logs` to collect diagnostic logs from the Kubernetes cluster for troubleshooting or sharing with NVIDIA Run:ai Support. See [CLI command reference](../reference/cli/runai/runai-diagnostics.md) for more details.
* **Node role management** - Use `kubectl` to configure node roles, which is already the recommended approach for node management. See [Node roles](../installation-and-configuration/post-installation-configuration/node-roles.md) for more details.

#### Node Level Scheduler

The Node-level Scheduler feature is deprecated. This allows NVIDIA Run:ai to focus on more scalable and flexible GPU resource optimization capabilities. Existing deployments will continue to work, but the feature will be removed in a future release.

#### Department Administrator Role

The previous Department administrator role has been deprecated and renamed to Department administrator legacy. A new Department administrator role replaces it with the same permissions, except it no longer includes Read Clusters access. Existing role assignments are not affected, as roles are referenced by ID in all APIs.

#### API Deprecation Notifications <a href="#api-deprecation-notifications" id="api-deprecation-notifications"></a>

**Deprecated Endpoints**

| Deprecated Endpoint                                   | Replacement Endpoint                                                                           |
| ----------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| `/api/v1/token`                                       | `/api/v2/token`                                                                                |
| `POST /api/v1/notification-channels/slack/create-app` | `POST /api/v1/slack/app`                                                                       |
| `GET /api/v1/notification-channels/slack/add-app`     | `POST /api/v1/slack/workspace`                                                                 |
| `POST /api/v1/notification-channels/slack/auth-app`   | `GET /api/v1/slack/auth`                                                                       |
| `GET /api/v1/notification-state`                      | <p><code>GET /api/v1/email/state</code><br><code>GET /api/v1/slack/state</code></p>            |
| `PATCH /api/v1/notification-state`                    | <p><code>PATCH /api/v1/slack/state</code></p><p><code>PATCH /api/v1/email/state</code></p>     |
| `GET /api/v1/notification-state/detailed`             | `GET /api/v1/slack/state/detailed`                                                             |
| `POST /api/v1/validate-notification-channel`          | <p><code>POST /api/v1/slack/validate</code></p><p><code>POST /api/v1/email/validate</code></p> |
| `/api/v1/clusters/{clusterUuid}/nodes`                | `/api/v1/nodes`                                                                                |
| `/api/v1/org-unit/priorities`                         | `/api/v1/org-unit/ranks`                                                                       |
| `/api/v1/administration/user-applications`            | `/api/v1/administration/access-keys`                                                           |
| `/api/v1/administration/user-applications/{appId}`    | `/api/v1/administration/access-keys{accessKeyId}`                                              |

**Deprecated Parameters**

| Endpoint             | Deprecated Parameter    | Replacement Parameter                       |
| -------------------- | ----------------------- | ------------------------------------------- |
| `/api/v1/node-pools` | `overProvisioningRatio` | N/A                                         |
| `/api/v1/node-pools` | `placementStrategy`     | `schedulingConfiguration.placementStrategy` |
| `/api/v1/workloads`  | `externalConnections`   | `endpoints`                                 |
| `/api/v1/workloads`  | `urls`                  | `endpoints`                                 |

### January 2026

#### NVIDIA Run:ai Predefined Roles

The following predefined roles are deprecated in the UI and API. Review the new predefined roles to determine whether they meet your requirements, or create a custom role using the API. See [Roles](../infrastructure-setup/authentication/roles.md) for more details:

* Compute resource administrator
* Credentials administrator
* Data source administrator
* Data volume administrator
* Environment administrator
* L1 researcher
* L2 researcher
* ML engineer
* Research manager
* Template administrator

During the deprecation period, the following predefined roles will be updated with minimal access to cluster and node pool data:

* Both cluster and node pool access are planned to transition to Clusters minimal and Node pools minimal - L1 researcher, L2 researcher, Research manager
* Cluster access is planned to transition to Clusters minimal while node pools access remains unchanged - ML engineer
* Cluster access is planned to transition to Clusters minimal - Compute resource administrator, Credentials administrator, Data source administrator, Data volume administrator, Environment administrator, Template administrator

#### Models Catalog

The Models catalog page is deprecated. Previously, the Models catalog provided a quick start experience for deploying a curated set of Hugging Face models. The same capability is now available through the Hugging Face inference workload flow, which integrates directly with Hugging Face and allows you to browse, select, and deploy any supported model from an open list. To deploy Hugging Face models, use the [Hugging Face inference workload](../workloads-in-nvidia-run-ai/using-inference/hugging-face-inference.md) flow.

#### API Deprecation Notifications

**Deprecated Endpoints**

| Deprecated Endpoint                  | Replacement Endpoint          |
| ------------------------------------ | ----------------------------- |
| <p><code>/api/v1/apps</code><br></p> | `/api/v1/service-accounts`    |
| `/api/v1/user-applications`          | `/api/v1/access-keys`         |
| `/api/v1/authorization/roles`        | `/api/v2/authorization/roles` |

**Deprecated Parameters**

| Endpoint                                                                                                                                                           | Deprecated Parameter       | Replacement Parameter           |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------- | ------------------------------- |
| `api/v1/authorization/access-rules`                                                                                                                                | `subjectType: app` (enum)  | `subjectType: service-account`  |
| <ul><li><code>/api/v2/authorization/roles</code></li><li><code>/api/v1/authorization/roles</code></li><li><code>/api/v1/authorization/permissions</code></li></ul> | `resourceType: app` (enum) | `resourceType: service-account` |
| <ul><li><code>/api/v1/org-unit/projects</code></li><li><code>/api/v1/org-unit/departments</code></li></ul>                                                         | `resources: priority`      | `resources: rank`               |

#### CLI Deprecation Notifications

| CLI Command                   | Deprecated Parameter | Replacement Parameter    |
| ----------------------------- | -------------------- | ------------------------ |
| `runai login`                 | `application`        | `access-key`             |
| `runai login`                 | `user`               | `password`               |
| `runai [workloadtype] submit` | `--environment`      | `--environment-variable` |

### September 2025

#### Grafana Dashboards

The legacy Grafana dashboards - **Overview** and **Analytics** - are deprecated and will be removed in a future release. We recommend transitioning to the new dashboards available in the NVIDIA Run:ai UI, which are powered by NVIDIA Run:ai APIs. These dashboards provide improved visibility with drill-down capabilities and more flexibility for analyzing usage and performance.

<Callout intent="info">
The Consumption dashboard was deprecated in version [July 2025](#july-2025) and replaced with Reports.
</Callout>

#### CLI v1

CLI v1 was deprecated in [January 2025](#january-2025) and has now been fully removed from the platform. All command-line interactions should be performed using [CLI v2](../reference/cli/runai/).

<Callout intent="info">
CLI v1 will still be available for clusters below v2.18.
</Callout>

#### Jobs

The Jobs workload type was deprecated in [January 2025](#january-2025) and has now been fully removed from the platform. This means the Jobs option no longer appears in the **General settings** or in the Workloads page.

### July 2025

#### Consumption Dashboard <a href="#consumption-dashboard" id="consumption-dashboard"></a>

The Consumption dashboard is deprecated and replaced with [Reports](../platform-management/monitor-performance/reports.md). Consumption reports provide improved visibility into resource usage with enhanced filtering and export capabilities. We recommend transitioning to consumption reports for the most up-to-date insights.

#### Templates <a href="#templates" id="templates"></a>

The Templates feature is deprecated. We recommend transitioning to [flexible workload templates](../workloads-in-nvidia-run-ai/workload-templates/), which offer enhanced functionality and support for flexible workload types - including workspace, standard training, and distributed training.

### April 2025

#### Cluster API for Workload Submission <a href="#cluster-api-for-workload-submission" id="cluster-api-for-workload-submission"></a>

Using the Cluster API to submit NVIDIA Run:ai workloads via YAML was [deprecated](https://docs.run.ai/v2.18/home/whats-new-2-18/#cluster-api-deprecation) starting from NVIDIA Run:ai version 2.18. For cluster version 2.18 and above, use the NVIDIA Run:ai REST API to submit workloads. The Cluster API documentation has also been removed.

### January 2025

#### Ongoing Dynamic MIG Deprecation Process <a href="#ongoing-dynamic-mig-deprecation-process" id="ongoing-dynamic-mig-deprecation-process"></a>

The Dynamic MIG deprecation process started in version 2.19. NVIDIA Run:ai supports standard MIG profiles as detailed in [Configuring NVIDIA MIG profiles](../platform-management/aiinitiatives/resources/mig-profiles.md).

* Before upgrading to version 2.20, workloads submitted with Dynamic MIG and their associated node configurations must be removed
* In version 2.20, MIG was removed from the NVIDIA Run:ai UI under compute resources.
* In Q2/25 all ‘Dynamic MIG’ APIs and CLI commands will be fully deprecated. (it will fail)

#### CLI v1 Deprecation <a href="#cli-v1-deprecation" id="cli-v1-deprecation"></a>

CLI V1 is deprecated and no new features will be developed for it. It will remain available for use for the next two releases to ensure a smooth transition for all users. We recommend switching to **CLI v2**, which provides feature parity, backwards compatibility, and ongoing support for new enhancements. CLI v2 is designed to deliver a more robust, efficient, and user-friendly experience.

#### Legacy Jobs View Deprecation <a href="#legacy-jobs-view-deprecation" id="legacy-jobs-view-deprecation"></a>

The legacy **Jobs view** will be discontinued in favor of the more advanced **Workloads view**. The legacy submission form will still be accessible via the Workload manager view for a smoother transition.

#### appID and appSecret Deprecation <a href="#appid-and-appsecret-deprecation" id="appid-and-appsecret-deprecation"></a>

Deprecating appID and appSecret parameters used for [requesting an API token](https://run-ai-docs.nvidia.com/api/getting-started/how-to-authenticate-to-the-api). It will remain available for use for the next two releases. To create application tokens, use your client credentials - Client ID and Client secret.
