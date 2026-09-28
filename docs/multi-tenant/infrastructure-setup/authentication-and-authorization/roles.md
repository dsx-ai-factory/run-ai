# Roles

A role defines a set of permissions that can be assigned to a subject in a scope, determining which actions the subject is allowed to perform on NVIDIA Run:ai entities such as projects, workloads, and users. Each permission represents an allowed action including view, create, edit, and delete on a specific entity.

Related permissions are grouped into permission sets, which bundle actions for a particular functional area (for example, managing workloads). These permission sets act as reusable building blocks that simplify role definition and ensure consistent access control across the platform.

NVIDIA Run:ai provides [predefined roles](roles.md#roles-in-nvidia-run-ai) for common use cases and allows administrators to create [custom roles](roles.md#custom-roles) to meet specific organizational requirements.

## Roles in NVIDIA Run:ai

NVIDIA Run:ai provides a set of predefined roles and their permissions. Under each role is a detailed list of the actions that the role assignee is authorized to perform for each entity. See [Roles](api:multi-tenant-api:GET/api/v2/authorization/roles) API for more details.

## Custom Roles

Administrators can create and manage custom roles to meet specific organizational access requirements using the [Roles](api:multi-tenant-api:POST/api/v2/authorization/roles) API.

A custom role defines access by combining one or more **permission sets**, which together determine the actions that can be performed across NVIDIA Run:ai resources. By selecting the appropriate permission sets, administrators can tailor roles to match their organization’s access model while maintaining consistent and controlled authorization across the platform.

<Note>

**Note**

Custom roles can be created and managed via API only. Once created, they are available for viewing in the Roles grid and can be selected when assigning access rules in the UI and API.

</Note>

### Permission Sets

Permission sets are the building blocks of both predefined and custom roles. They represent the complete set of privileges required for a role to perform specific operations across NVIDIA Run:ai resources.

A permission set is constructed from a collection of permissions. Each permission within a permission set defines:

* a **resource type**, and
* the allowed **actions** on that resource: `create`, `read`, `update`, or `delete`

The Permission sets API provides a catalog of all available permission sets in the NVIDIA Run:ai platform. Permission sets may overlap in the resource types or actions they include. For more details, see the [Permissions](api:multi-tenant-api:GET/api/v1/authorization/permissions) API.

#### Permission Set Naming and Scope

Permission sets are designed to group actions logically, following a clear pattern to define scope. To view the complete list of permission sets and understand how each is defined, see the [Permissions](api:multi-tenant-api:GET/api/v1/authorization/permissions) API:

* Edit Access (`EditAccess`) - These sets grant full CRUD (Create, Read, Update, Delete) permissions over the specified resource type, along with Read access to any associated data or dependencies required for management. For example, **workspaceEditAccess** grants full permissions to `create`, `read`, `update`, and `delete` workspaces, and `read` access all related data and resources required for their management, such as:
  * `workloads`
  * `workload-properties`
  * `nodepools-minimal`
  * `clusters-minimal`
  * `projects`
  * `templates`
  * `environments`
  * `compute-resources`
  * `pvc-assets`
* Read Access (`ReadAccess`) - These sets grant only view permissions for the specific resource and its related components. For example, **workloadReadAccess** grants permissions to `read` workloads, and `read` access all related data and resources associated with them, such as:
  * `workspaces`
  * `trainings`
  * `inferences`
  * `workload-properties`
  * `nodepools-minimal`
  * `clusters-minimal`
  * `projects`

### Creating a Custom Role

#### Requirements for Custom Roles

Custom roles must meet the following requirements:

* **UI access permissions** - To allow a custom role to access the NVIDIA Run:ai user interface, include the following permission sets:
  * `settingsReadAccess`
  * `accountReadAccess`
  * `brandingSettingsReadAccess`
  * `securitySettingsReadAccess`
  * `notificationSettingsReadAccess`
* **Kubernetes (cluster) permissions** - Kubernetes permissions are **required only** if the role needs direct access to the Kubernetes cluster. If cluster access is required, the custom role must include a `kubernetesPermissions` object. This ensures the role inherits Kubernetes cluster-level permissions from an existing predefined NVIDIA Run:ai role.

#### Scope Behavior

When creating a custom role, the `scopeType` and associated scope identifier, `scopeId`, determine where the role is available and who can use it in a multi-tenant environment:

* `scopeType = system` - Creates a global role that is available across all customer tenants.\
  This is typically used by platform owner (host organization) to define standardized roles (for example, an AI developer role) that should be reused by all customer tenants.
* `scopeType = tenant` with a customer tenant ID - Creates a role that is scoped to a specific customer tenant only using `scopeId`. The role is visible and assignable only within that tenant.
* `scopeType = tenant` with the Console (system) tenant ID - Creates a role scoped to the platform owner (host organization) tenant using `scopeId`.

#### Steps for Creating a Custom Role

1. Identify the required permission sets:
   * Determine the capabilities the new role needs (e.g., managing training jobs, viewing clusters, etc.).
   * Retrieve the full catalog of available permission set identifiers (`name` and `id`) using the `GET /v1/api/permission-sets` endpoint. This provides the list of all valid IDs (e.g., `inferenceEditAccess`, `workloadReadAccess`) to populate the `permissionSets` array in your request body.
   * Include `settingsReadAccess`, `accountReadAccess`, `brandingSettingsReadAccess` and `securitySettingsReadAccess` to allow the custom role access the NVIDIA Run:ai user interface.
2. (Optional) Identify the Kubernetes predefined role ID. The custom role must inherit its underlying Kubernetes cluster permissions from an existing NVIDIA Run:ai predefined role if the role needs to access the Kubernetes cluster directly:
   * Retrieve the list of all available predefined roles using the `GET /v2/authorization/roles` endpoint.
   * Select the unique ID (a string identifier, e.g., "12") of the predefined role that grants the appropriate level of cluster access for your custom role.
   * This ID will populate the `predefinedRole` field under the `kubernetesPermissions` object.
3. Construct and send the API request. Use the `POST /v2/authorization/roles` endpoint to create the role.

**Example: MLOps Role**

This example demonstrates a role that can fully manage inference workloads and access related monitoring information.

```json
{
  "name": "MLOps",
  "description": "Fully manage inference workloads, and access other relevant information for monitoring",
  "scopeType" = "system"
  "scopeId" = "<SYSTEM-ID>"
  "permissionSets": [
    { "name": "inferenceEditAccess", "id": "<permissionSetId>" },
    { "name": "templatesEditAccess", "id": "<permissionSetId>" },
    { "name": "dataAndStorageReadAccess", "id": "<permissionSetId>" },
    { "name": "nodesReadAccess", "id": "<permissionSetId>" },
    { "name": "clustersReadAccess", "id": "<permissionSetId>" },
    { "name": "settingsReadAccess", "id": "<permissionSetId>" },
    { "name": "accountReadAccess", "id": "<permissionSetId>" }
  ],
  "kubernetesPermissions": {
    "predefinedRole": "12" 
  }
}
```

### Enabling / Disabling Custom Roles

Administrators can control whether it is available for use by enabling or disabling it:

* **Enabled** - Can be assigned to users and used in access rules.
* **Disabled** - Remain defined in the system but cannot be assigned to users or used in access rules.

## Troubleshooting Common Issues

<Accordion title="Can't delete role">

**Can't delete role**

**Description:** The role cannot be deleted.

**Mitigation:**

1. Verify that the role is not assigned to any users.
2. If the role is currently in use, disable the role first, and then delete it.

</Accordion>

<Accordion title="Permissions not taking effect">

**Permissions not taking effect**

**Description:** The selected permission sets are not taking effect.

**Mitigation:**

1. Ensure that the role is **enabled**.
2. Verify that the user has been **assigned the role**.
3. Check for **conflicting permissions** granted by other roles assigned to the user.

</Accordion>

<Accordion title="Cluster permissions not working">

**Cluster permissions not working**

**Description:** Cluster-level permissions are not taking effect.

**Mitigation:**

1. Verify that the required ClusterRole exists in all relevant clusters.
2. Confirm that the cluster version is 2.21 or later.
3. Ensure that the cluster is properly connected to the tenant.

</Accordion>
