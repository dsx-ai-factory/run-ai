---
---

# Create a Tenant and Cluster

This page describes how to onboard a new tenant end to end: create the tenant, provision and register its Kubernetes cluster, install the cluster using the Helm command returned by the platform, and share access details with the tenant.

Onboarding a tenant requires an authenticated Bearer token - see [API Access Setup](../guided-installation/api-access-setup.md) if you have not generated one yet.

## Create a Tenant

### Prepare Tenant Information

Prepare the required details for the tenant you want to create:

* **Tenant name** - A unique identifier for the tenant.
* **Administrator email** - The primary contact who will receive access credentials and manage the tenant's environment.

Refer to the [Tenant](api:multi-tenant-api:POST/api/v1/tenants) API endpoint for more details.

### Create the Tenant

To register a new tenant in the system, send a `POST` request to the `/api/v1/tenants` endpoint with the tenant information. This creates the tenant and assigns it a unique domain under your managed wildcard DNS (e.g., `tenant-name.runai.hostorg-domain.com`) and exposes tenant-facing services.

**Example request:**

```bash
curl -L 'https://console.<DOMAIN>/api/v1/tenants' \
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer <TOKEN>' \
{
  "name": "mytenant",
  "email": "administrator@mytenant.com"
}
```

After the tenant is successfully created, the API response includes a `domain` field. This is the fully qualified domain name (FQDN) that the tenant will use to access NVIDIA Run:ai services, either through a web browser or programmatic integration. You'll [share](#share-tenant-access-details) this domain with the tenant's administrators once the cluster is installed and connected.

**Example response:**

```bash
tenant": {
    "id": 1,
    "name": "example",
    "displayName": "Example",
    "status": "Ready",
    "tenantId": 123,
    "eulaSigningUser": {
      "email": "user@tenant",
      "date": "2020-01-01T00:00:00Z",
      "ip": "192.168.10.210"
    },
    "created_at": "2020-01-01T00:00:00Z",
    "updated_at": "2020-01-02T00:00:00Z",
    "deleted_at": null
  },
  "additionalData": {
    "tenantDomain": "mytenant.<globaldomain>",
    "testUser": {
      "tenant-test@run.ai": "Abc!23"
    },
    "customerUser": {
      "administrator@mytenant.com": "<password>"
    }
  }
}
```

### View Existing Tenants

To retrieve a list of tenants registered in the system, send a `GET` request to the `/api/v1/tenants` endpoint. This returns details for each tenant, including the name, status, and domain information. You can use this endpoint to verify whether a tenant already exists or to audit tenant states across environments.

**Example request:**

```bash
curl -L 'https://console.<DOMAIN>/api/v1/tenants' \
-H 'Authorization: Bearer <TOKEN>' \
-H 'Accept: application/json'
```

**Example response:**

```bash
[
  {
    "id": 1,
    "name": "example",
    "displayName": "Example",
    "status": "Ready",
    "tenantId": 123,
    "createdAt": "2020-01-01T00:00:00Z",
    "updatedAt": "2020-01-02T00:00:00Z",
    "deletedAt": null
  }
]
```

**Tenant status values:**

The `status` field indicates the current lifecycle state of the tenant. Possible values include:

* `Creating` - The tenant is being provisioned.
* `Updating` - The tenant is being reconfigured.
* `Ready` - The tenant is active and fully operational.
* `Deleting` - The tenant is being decommissioned.
* `Deleted` - The tenant has been removed.
* `Failed` - Tenant creation or update has failed.

Use these values to track the readiness of tenant environments or detect errors during provisioning.

## Provision the Tenant's Kubernetes Cluster

As part of onboarding a new tenant, the host organization is responsible for setting up a dedicated Kubernetes environment that will host that tenant's NVIDIA Run:ai cluster.

### Deployment Scenarios

You can deploy the cluster using your preferred **Infrastructure as a Service (IaaS)** provider. Supported environments include:

* **On-premise** - Rancher, OpenShift, BCM, and OpenStack (recommended)
* **Public cloud** - GCP, AWS, OCI, and Azure

Select the platform that best fits your infrastructure and operational requirements.

### Required Steps

* **Cluster provisioning** - Create and allocate a dedicated Kubernetes cluster for the tenant to ensure workload isolation and compliance.
* **Kubernetes installation** - Set up and configure Kubernetes on the provisioned infrastructure to prepare for NVIDIA Run:ai deployment.
* **System requirements** - Complete [Preparations](preparations.md) and [Prerequisites and Features](prerequisites-and-features.md) on the provisioned cluster before continuing.

## Register the Cluster

After provisioning the tenant's Kubernetes cluster and completing its system requirements, register the cluster in the NVIDIA Run:ai platform. This step creates a unique cluster object and associates it with the correct tenant.

To register a cluster:

* Send a `POST` request to the `/api/v1/clusters` endpoint, including the tenant's `tenantId`. See [Clusters](api:multi-tenant-api:POST/api/v1/clusters) API for more details.
* Provide the cluster's **name**, **domain** ([FQDN](../../system-requirements/networking-setup.md#fqdn-and-dns)), and the target NVIDIA Run:ai version.
* Once registered, the platform returns a cluster `uuid`.

This registration step does **not** install anything on the tenant's cluster yet. It only prepares the cluster for installation and establishes the tenant-cluster association.

**Example request:**

```bash
curl -L 'https://console.<DOMAIN>/api/v1/clusters' \
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer <TOKEN>' \
{
  "name": "string",
  "domain": "string",
  "version": "string",
  "tenantId": 1001
}
```

**Example response:**

```bash
{
  "uuid": "A0EEBC99-9C0B-4EF8-BB6D-6BB9BD380A11",
  "name": "example",
  "tenantId": 1001,
  "domain": "my.company.com",
  "status": { },
  "createdAt": "2020-01-01T00:00:00Z",
  "updatedAt": "2020-01-02T00:00:00Z",
  "lastLiveness": "2020-01-02T00:00:00Z",
  "version": "2.15.0"
}
```

## Retrieve Cluster Installation Instructions

Once the cluster is registered, the next step is to install the NVIDIA Run:ai cluster component on the tenant's Kubernetes environment. This is done using the Helm command provided by the platform.

To retrieve the installation command and required parameters:

* Send a `GET` request to the `/v1/clusters/{clusterUuid}/cluster-install-info?version=<version>` endpoint using the UUID returned when the cluster was created.
* The response includes a preconfigured `installationStr`, Helm repository details, and a `bootstrapToken` to be used in the install command. The `bootstrapToken` is a short-lived token that expires 2 hours after it is issued. If installation takes longer, send the `GET` request again to retrieve a new `installationStr` and `bootstrapToken` before running the command.

You will run the returned `installationStr` on the tenant's cluster to complete the installation and connect it to the NVIDIA Run:ai platform.

**Example response:**

```bash
{
  "installationStr": "helm upgrade -i runai-cluster -n runai runai/runai-cluster --set cluster.uid=<CLUSTER_UUID> --set controlPlane.url=<CONTROL_PLANE_DOMAIN> --set controlPlane.bootstrapToken=<BOOTSTRAP_TOKEN>",
  "repositoryName": "runai",
  "chartRepoURL": "https://helm.ngc.nvidia.com/nvidia/runai-oss",
  "bootstrapToken": "ABC333DDD"
}
```

## Install the Cluster

Run the returned `installationStr` in the terminal on the tenant's Kubernetes cluster. This installs the NVIDIA Run:ai cluster component and links the cluster to the tenant in the control plane.

### Verify the Installation

After the installation completes successfully, verify that the NVIDIA Run:ai pods are running:

```bash
kubectl get pods -n runai
```

All pods should reach `Running` or `Completed` status. Querying the cluster through the [Clusters](api:multi-tenant-api:POST/api/v1/clusters) API should show a **Connected** status under its configured name.

## Share Tenant Access Details

Once the cluster is installed and connected, provide the tenant with:

* **NVIDIA Run:ai access URL** - The fully qualified domain name (FQDN) for their environment, e.g.: `https://tenant-a.runai.hostorg.com`. This is used to access the UI, API, and CLI.
* **Administrator credentials** - The username and password you configured during tenant creation. These are used to authenticate in all interfaces (UI/API/CLI).
