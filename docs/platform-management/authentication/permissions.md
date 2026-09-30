# Permissions

This page provides a reference of the permissions available in the NVIDIA Run:ai platform. Use it when building [custom roles](roles.md) to understand what each permission controls and how permission names in the API map to the names shown in the user interface.

## About Permissions

A permission is the actions (View, Edit, Create and Delete) that the user can perform over a NVIDIA Run:ai entity (resource type). These actions map to the API actions `read`, `update`, `create`, and `delete` respectively. Related permissions are grouped into [permission sets](#permission-sets), and a [role](roles.md) is defined as a collection of permission sets assigned to a subject in a scope. NVIDIA Run:ai provides predefined roles that are system-defined and cannot be edited or deleted. Administrators can also create custom roles via the API.

When you build a custom role through the API, permissions are referenced by their **API name** (for example, `compute-resources`). The NVIDIA Run:ai user interface presents the same permission under a **display name** (for example, "Compute resources"). The tables below map the two and describe what each permission controls.

<Note>

**Note**

* Some permissions are system-internal and are not shown in the Roles view in the user interface. These are noted in the tables below.
* **Deprecated:** The `apps` API name is deprecated. Use `service-account` for managing service accounts.

</Note>

## Permission Reference

The following table lists all permissions, mapping each permission's API name to its user interface display name and describing what it controls.

| API name | Display name | Description |
| --- | --- | --- |
| `access-keys` | [Access Keys](../../settings/user-settings/user-access-keys.md) | Access keys are used for API integrations with NVIDIA Run:ai. An access key contains a client ID and a client secret. This permission is not shown in the Roles view. |
| `access_rules` | [Access rules](accessrules.md) | Access rules provide users, groups, or service accounts privileges to system entities. An access rule is the assignment of a role to a subject in a scope. |
| `tenant` | Account | The top-level organizational node headed by the account, comprised of clusters, departments, and projects. |
| `applications` | [AI Applications](../../ai-applications/ai-applications.md) | An AI application represents a high-level logical grouping of all Kubernetes resources that together deliver a functional AI solution. |
| `apps` | [Applications](service-accounts.md) | Deprecated. Superseded by `service-account`. Distinct from `applications` (AI Applications). |
| `branding-settings` | [Branding settings](../../settings/general-settings/README.md#branding) | Custom logo displayed in the top-right corner of the NVIDIA Run:ai platform interface. |
| `cluster-config` | [Advanced cluster configuration](../../installation-and-configuration/configuration/cluster-config.md) | Governs synchronization of cluster configuration between a cluster and the control plane. This permission is system-internal: it is not shown in the Roles view and is not held by tenant administrators. Do not include `cluster-config` in a custom role. A role that includes it cannot be assigned, and access rule creation fails with `403 Insufficient permissions`. |
| `cluster` | [Clusters](../../installation-and-configuration/operations/clusters.md) | A registered Kubernetes cluster that the NVIDIA Run:ai control plane manages, schedules workloads on, and monitors. |
| `clusters-minimal` | [Clusters minimal](../../installation-and-configuration/operations/clusters.md) | Provides read-only access to a lightweight subset of cluster metadata (name, ID, status). Automatically included by the platform when a role grants cluster read access, to support resource display in the user interface. |
| `compute-resources` | [Compute resources](../../workloads-in-nvidia-run-ai/workload-assets/compute-resources.md) | A compute resource is a workload asset template that simplifies how workloads are submitted and can be used by AI practitioners when they submit their workloads. |
| `credentials` | [Credentials](../../workloads-in-nvidia-run-ai/workload-assets/credentials.md) | Credentials are workload assets that mask sensitive access information, such as passwords, tokens, and access keys, necessary for accessing protected resources. |
| `pvc-assets` | [Data sources](../../workloads-in-nvidia-run-ai/workload-assets/data-volumes.md) | A data source asset backed by a Persistent Volume Claim (PVC), a Kubernetes concept for managing cluster storage provisioned by an administrator or dynamically using a StorageClass. |
| `git-assets` | [Data sources](../../workloads-in-nvidia-run-ai/workload-assets/overview.md) | A data source asset that copies code from a Git branch into a dedicated folder in the container, providing the workload with the latest code repository. |
| `nfs-assets` | [Data sources](../../workloads-in-nvidia-run-ai/workload-assets/overview.md) | A data source asset backed by a Network File System (NFS), used for sharing storage among pods in the cluster; content is preserved outside the lifecycle of a single pod. |
| `s3-assets` | [Data sources](../../workloads-in-nvidia-run-ai/workload-assets/overview.md) | A data source asset that maps a remote S3 bucket into the workload's file system, accessible across different workload executions. |
| `host-path-assets` | [Data sources](../../workloads-in-nvidia-run-ai/workload-assets/overview.md) | A data source asset that mounts a host path file or directory on the workload's file system; data persists across workloads. |
| `cm-volume-assets` | [Data sources](../../workloads-in-nvidia-run-ai/workload-assets/overview.md) | A data source asset backed by a Kubernetes ConfigMap, used to store non-confidential data such as environment variables and command-line arguments. |
| `secret-volume-assets` | [Data sources](../../workloads-in-nvidia-run-ai/workload-assets/overview.md) | A data source asset that maps a credential into the workload's file system to provide sensitive access information such as passwords, tokens, and access keys. |
| `datavolumes` | [Data volumes](../../workloads-in-nvidia-run-ai/workload-assets/data-volumes.md) | Data volumes are workload assets that offer a solution for storing, managing, and sharing AI training data across scopes and workloads. |
| `department` | [Departments](../../platform-management/aiinitiatives/organization/departments.md) | Departments group multiple projects under a shared organizational scope for quota management, policy enforcement, and asset sharing. |
| `environments` | [Environments](../../workloads-in-nvidia-run-ai/workload-assets/environments.md) | An environment is a workload asset consisting of a configuration that simplifies how workloads are submitted and can be used by AI practitioners. |
| `events-history` | [Event history](../../installation-and-configuration/operations/observability/event-history.md) | An audit log of all changes to business objects (clusters, projects, assets, and more) in the NVIDIA Run:ai platform, accessible to system administrators with tenant-wide permissions. |
| `inferences` | [Inferences](../../workloads-in-nvidia-run-ai/using-inference/custom-inference.md) | Inference workloads deploy trained models to make real-time or batch predictions in a production environment. |
| `network-topologies` | [Network topologies](../../installation-and-configuration/operations/clusters.md) | Named configurations of node labels that describe the physical network structure of a cluster (racks, blocks, NVLink domains), used by the scheduler to minimize latency for distributed workloads. |
| `nodepools` | [Node pools](../../platform-management/aiinitiatives/resources/node-pools.md) | A node pool is a NVIDIA Run:ai construct representing a set of nodes grouped into a bucket of resources using a predefined or administrator-defined node label. |
| `nodepools-minimal` | [Node pools minimal](../../platform-management/aiinitiatives/resources/node-pools.md) | Provides read-only access to a lightweight subset of node pool metadata. Automatically included by the platform when a role grants node pool, department, or project read access. |
| `nodes` | [Nodes](../../platform-management/aiinitiatives/resources/nodes.md) | Nodes are Kubernetes elements automatically discovered by the NVIDIA Run:ai platform and registered for scheduling and administrator visibility. |
| `notification-settings` | [Notification settings](../../settings/general-settings/notifications.md) | Configuration for notification options that keep administrators and users informed of important events and updates. |
| `policies` | [Policies](../../platform-management/policies/native-workload-policies.md) | Workload policies allow administrators to set best practices, enforce limitations, and standardize workload submission processes within their organization. |
| `project` | [Projects](../../platform-management/aiinitiatives/organization/projects.md) | Projects are the primary organization management unit in NVIDIA Run:ai, used to control how resources are allocated and prioritized across teams and initiatives. |
| `registries` | [Registries](../../workloads-in-nvidia-run-ai/workload-assets/environments.md#adding-a-new-environment) | Stored connection configurations for private container image registries, used when selecting images in environment assets. |
| `roles` | [Roles](roles.md) | A role defines a set of permissions that can be assigned to a subject in a scope, determining which actions the subject is allowed to perform on NVIDIA Run:ai entities. |
| `security-settings` | [Security settings](../../settings/general-settings/README.md#security) | Platform-level identity provider (SSO) configuration and session security settings. |
| `service-account` | [Service accounts](service-accounts.md) | Service accounts are used for API integrations with NVIDIA Run:ai. A service account contains a client ID and a client secret. |
| `settings` | [Settings](../../settings/general-settings/README.md) | Centralized controls for enabling or disabling key NVIDIA Run:ai platform features across resources, workload types, and security. |
| `storage-class-configuration` | Storage class configuration | Kubernetes storage classes define how storage is provisioned for AI workloads, used when creating PVC data sources and data volumes. |
| `templates` | [Templates](../../workloads-in-nvidia-run-ai/workload-templates/README.md) | A workload template is a reusable setup that defines all the required configuration fields for submitting a workload, allowing users to predefine settings. |
| `trainings` | [Trainings](../../workloads-in-nvidia-run-ai/using-training/train-models.md) | Training workloads run resource-intensive, batch-mode model development jobs, often requiring distributed computing across multiple nodes. |
| `users` | [Users](users.md) | Users can be managed locally or via an identity provider (IdP) and are assigned permissions through access rules. |
| `users-minimal` | [Users minimal](users.md) | Provides read-only access to a minimal subset of user data, such as user IDs and display names. Used by system roles to resolve user identities without exposing full user management permissions. |
| `workload-integration-metrics` | Workload integration metrics | System-internal. Used by the platform to collect workload metrics from integrations. Not intended to be assigned when building custom roles. |
| `workload-properties` | [Workload properties](../../workloads-in-nvidia-run-ai/workloads.md#managing-workload-properties) | The priority, preemptibility, and workload category assigned to each workload type by default. |
| `workloads` | [Workloads](../../workloads-in-nvidia-run-ai/workloads.md) | Workloads are the fundamental building blocks for consuming resources, enabling AI practitioners to support the entire lifecycle of an AI initiative. |
| `workspaces` | [Workspaces](../../workloads-in-nvidia-run-ai/using-workspaces/running-workspace.md) | The workspace is where data scientists conduct initial research, experiment with data sets, and test algorithms — the most flexible stage in the ML lifecycle. |

<Note>

**Note**

The seven `*-assets` API names above all appear under the single **Data sources** display name in the user interface. When building a custom role via the API, grant the specific asset types your users need using the relevant permission sets.

</Note>

## Permission Sets

Permissions are bundled into **permission sets** — reusable building blocks that group the actions for a particular functional area (for example, viewing workloads or managing clusters). Permission sets are used through the [Roles](roles.md) API to define both predefined and custom roles, and are referenced by the API names below. Most permission sets come in a **read** variant (view only) and an **edit** variant (create, read, update, and delete). The following permission sets are available:

| API name | Description |
| --- | --- |
| `workspaceEditAccess` | Grants full permissions to create, read, update, and delete workspaces, and to access all related data and resources required for their management. |
| `trainingEditAccess` | Grants full permissions to create, read, update, and delete training workloads, and to access all related data and resources required for their management. |
| `inferenceEditAccess` | Grants full permissions to create, read, update, and delete inference workloads, and to access all related data and resources required for their management. |
| `yamlWorkloadEditAccess` | Grants full permissions to create, read, update, and delete workloads defined in YAML, and to access all related data and resources required for their management. |
| `workloadReadAccess` | Grants permissions to view workloads, including all related data and resources associated with them. |
| `workloadPropertiesEditAccess` | Grants full permissions to create, read, update, and delete workload properties, and to access all related data and resources required for their management. |
| `workloadPropertiesReadAccess` | Grants permissions to view workload properties, including all related data and resources associated with them. |
| `templatesEditAccess` | Grants full permissions to create, read, update, and delete workload templates, and to access all related data and resources required for their management. |
| `templatesReadAccess` | Grants permissions to view workload templates, including all related data and resources associated with them. |
| `templatesOnlyEditAccess` | Grants permissions to create, read, update, and delete workload templates. |
| `templatesOnlyReadAccess` | Grants permissions to view workload templates. |
| `environmentsOnlyEditAccess` | Grants permissions to create, read, update, and delete environments. |
| `environmentsOnlyReadAccess` | Grants permissions to view environments. |
| `computeResourcesOnlyEditAccess` | Grants permissions to create, read, update, and delete compute resources. |
| `computeResourcesOnlyReadAccess` | Grants permissions to view compute resources. |
| `credentialsOnlyEditAccess` | Grants permissions to create, read, update, and delete credentials. |
| `credentialsOnlyReadAccess` | Grants permissions to view credentials. |
| `dataAndStorageEditAccess` | Grants full permissions to create, read, update, and delete data and storage, and to access all related data and resources required for their management. |
| `dataAndStorageReadAccess` | Grants permissions to view data and storage, including all related data and resources associated with them. |
| `policiesEditAccess` | Grants full permissions to create, read, update, and delete policies, and to access all related data and resources required for their management. |
| `policiesReadAccess` | Grants permissions to view policies, including all related data and resources associated with them. |
| `projectsEditAccess` | Grants full permissions to create, read, update, and delete projects, and to access all related data and resources required for their management. |
| `projectsReadAccess` | Grants permissions to view projects, including all related data and resources associated with them. |
| `departmentsEditAccess` | Grants full permissions to create, read, update, and delete departments, and to access all related data and resources required for their management. |
| `departmentsReadAccess` | Grants permissions to view departments, including all related data and resources associated with them. |
| `clustersEditAccess` | Grants permissions to create, read, update, and delete clusters. |
| `clustersReadAccess` | Grants permissions to view clusters. |
| `networkTopologyEditAccess` | Grants full permissions to create, read, update, and delete network topologies, and to access all related data and resources required for their management. |
| `networkTopologyReadAccess` | Grants permissions to view network topologies, including all related data and resources associated with them. |
| `nodesReadAccess` | Grants permissions to view nodes, including all related data and resources associated with them. |
| `nodePoolsEditAccess` | Grants full permissions to create, read, update, and delete node pools, and to access all related data and resources required for their management. |
| `nodePoolsReadAccess` | Grants permissions to view node pools, including all related data and resources associated with them. |
| `quotaManagementDashboardReadAccess` | Grants permissions to view the quota management dashboard, including all related data and resources associated with them. |
| `usersEditAccess` | Grants permissions to create, read, update, and delete users. |
| `usersReadAccess` | Grants permissions to view users. |
| `serviceAccountsEditAccess` | Grants permissions to create, read, update, and delete service accounts. |
| `serviceAccountsReadAccess` | Grants permissions to view service accounts. |
| `accessRulesEditAccess` | Grants full permissions to create, read, update, and delete access rules, and to access all related data and resources required for their management. |
| `accessRulesReadAccess` | Grants permissions to view access rules, including all related data and resources associated with them. |
| `rolesEditAccess` | Grants permissions to create, read, update, and delete roles. |
| `rolesReadAccess` | Grants permissions to view roles. |
| `settingsEditAccess` | Grants permissions to create, read, update, and delete settings. |
| `settingsReadAccess` | Grants permissions to view settings. |
| `securitySettingsEditAccess` | Grants permissions to create, read, update, and delete security settings. |
| `securitySettingsReadAccess` | Grants permissions to view security settings. |
| `brandingSettingsEditAccess` | Grants permissions to create, read, update, and delete branding settings. |
| `brandingSettingsReadAccess` | Grants permissions to view branding settings. |
| `eventHistoryReadAccess` | Grants permissions to view event history. |
| `accountEditAccess` | Grants permissions to create, read, update, and delete accounts. |
| `accountReadAccess` | Grants permissions to view accounts. |
| `accessKeysEditAccess` | Grants full permissions to create, read, update, and delete user access keys. |
| `storageClassConfigurationEditAccess` | Grants full permissions to create, read, update, and delete storage class configurations used by volume and PVC assets. |
