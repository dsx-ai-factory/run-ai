# Deprecation Notifications

This page lists features, APIs, and capabilities that are deprecated as of this release. Deprecated items remain available for a limited time before they are removed; see each entry for details on timing and recommended alternatives.

<Note>

**Note**

For details on how long deprecated features remain supported, see [Support, Lifecycle, and Deprecation Policy](../product-support-policy.md).

</Note>

## Deprecated in Version 3.0

No features, APIs, or capabilities are newly deprecated in version 3.0.

## Removed in Version 3.0

The following features and APIs are removed in version 3.0.

* **Applications and user applications APIs** - The `/api/v1/apps` and `/api/v1/user-applications` endpoints, and the `/api/v1/administration/user-applications` endpoints, are removed. Use service accounts (`/api/v1/service-accounts`) and user access keys (`/api/v1/administration/access-keys`) instead.
* **Cluster nodes endpoint** - The `/api/v1/clusters/{clusterUuid}/nodes` endpoint is removed. Use `/api/v1/nodes` instead.
* **Org unit priorities endpoint** - The `/api/v1/org-unit/priorities` endpoint is removed. Use `/api/v1/org-unit/ranks` instead.
* **`tenantIsCspManaged` field removed** - The `tenantIsCspManaged` field has been removed from the `/api/v1/authorization/settings` endpoints.
* **Roles v1 endpoints** - The `GET /api/v1/authorization/roles` and `GET /api/v1/authorization/roles/{roleIdPath}` endpoints, deprecated since v2.24, are removed. Use `/api/v2/authorization/roles` and `/api/v2/authorization/roles/{roleIdPath}` instead.
* **`priority` field in project and department resources** - The `resources[].priority` field is removed from the projects and departments APIs. Responses no longer return it, and requests that still send it are rejected. Use `rank` instead.
* **JFrog as a software artifact source** - JFrog is no longer supported as a source for NVIDIA Run:ai software artifacts, following a deprecation that started in v2.25. NGC is the only supported path for installing and upgrading NVIDIA Run:ai. See [Upgrade NVIDIA Run:ai](../../installation-and-configuration/upgrade/upgrade.md) for the upgrade procedure.
* **Node-level scheduler and GPU memory swap** - The node-level scheduler (deprecated in v2.25) and GPU memory swap (deprecated in v2.26) are turned off when upgrading to v3.0 and cannot be re-enabled. See [Plan the Upgrade](../../installation-and-configuration/upgrade/plan-the-upgrade.md) for what to check before upgrading.
