---
---

# Create a Tenant and Cluster

This page describes how to onboard a new tenant end to end: create the tenant, provision and register its Kubernetes cluster, install the cluster using the `runai-installer` CLI, and share access details with the tenant.

Onboarding a tenant requires an authenticated Bearer token - see [API Access Setup](api-access-setup.md) if you have not generated one yet.

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
* **Prerequisites and features** - The installer installs the required prerequisites (GPU Operator, Prometheus) and any features for that tenant's planned workload types automatically as part of installing the cluster - see [Prerequisites and Features](prerequisites-and-features.md) to determine which ones apply.

## Register the Cluster

After provisioning the tenant's Kubernetes cluster and completing its system requirements, register the cluster in the NVIDIA Run:ai platform. This step creates a unique cluster object and associates it with the correct tenant.

To register a cluster:

* Send a `POST` request to the `/api/v1/clusters` endpoint, including the tenant's `tenantId`. See [Clusters](api:multi-tenant-api:POST/api/v1/clusters) API for more details.
* Provide the cluster's **name**, **domain** ([FQDN](../../system-requirements/networking-setup.md#fqdn-and-dns)), and the target NVIDIA Run:ai version.
* Once registered, the platform returns a cluster `uuid`. This is used to complete the installation and connect the tenant's cluster to the NVIDIA Run:ai control plane.

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

### Retrieve the Cluster Bootstrap Token

Once the cluster is registered, retrieve a bootstrap token to install it with the `runai-installer` CLI:

* Send a `GET` request to the `/v1/clusters/{clusterUuid}/cluster-install-info?version=<version>` endpoint using the `uuid` returned when the cluster was registered.
* The response includes a `bootstrapToken`. The `bootstrapToken` is a short-lived token that expires 2 hours after it is issued. If installation takes longer, send the `GET` request again to retrieve a new `bootstrapToken` before running the installer.

**Example response:**

```bash
{
  "bootstrapToken": "ABC333DDD"
}
```

Save the `bootstrapToken` to a local file - the installer reads it from disk rather than accepting it inline, to avoid storing it in the configuration file.

## Install the Cluster

<Note>

**Configuration Templates**

The steps below build the minimum configuration to install a tenant cluster on standard Kubernetes. Depending on your Kubernetes distribution, workload routing, networking setup, and the prerequisites and workload features you need, you may require additional configuration.

* Once you understand the minimum configuration, see [Configuration Templates](tenant-cluster-templates.md) - if your use case matches one, use it as your starting point instead of building the configuration from scratch.
* See [Prerequisites and Features](prerequisites-and-features.md) to understand the components required, depending on the workload types that tenant plans to run.
* See [Configuration Reference](configuration-reference.md) for the full set of supported fields and descriptions, to build a configuration from scratch.

</Note>

Create a `cluster-config.yaml` file for the tenant's cluster. Use the domain, TLS certificates, and networking values you prepared in [Networking Setup](../../system-requirements/networking-setup.md) when completing the configuration below.

At a minimum, configure:

1. Define the installation:

   - `install.mode` - Set to `runai_cluster_only`.
   - `install.connectivity` - Set to `connected`, for an installation with outbound internet access.

   ```yaml
   install:
     mode: runai_cluster_only
     connectivity: connected
   ```

2. Configure the registry:

   - `registry.type` - Set to `ngc`.
   - `registry.tokenPath` - For NVIDIA Run:ai Support-Entitled Release installations, provide the path to your NGC Personal Key. Open Source installations do not require registry credentials.

   ```yaml
   registry:
     type: ngc
     tokenPath: /path/to/ngc-token # Support-Entitled Release only
   ```

3. Define the cluster:

   - `cluster.name` - The name of the cluster as it appears in the control plane.
   - `cluster.domain` - The tenant cluster's own FQDN. See [Domains and TLS Certificates](preparations.md#domains-and-tls-certificates).
   - `cluster.controlPlaneUrl` - The control plane's domain, reachable over HTTPS from this cluster.
   - `cluster.uid` - The cluster `uuid` returned when you [registered the cluster](#register-the-cluster).
   - `cluster.bootstrapTokenPath` - Path to the `bootstrapToken` file you saved when you [registered the cluster](#register-the-cluster). Mutually exclusive with `cluster.controlPlaneCredentials` - the installer does not register tenant clusters itself, since doing so requires a `tenantId` that only the Management API accepts.
   - `cluster.workloadRouting` - `host` (the default) or `path`. This determines which certificate you need for `cluster.tls` below.
   - `cluster.tls` - A certificate and private key for the cluster domain, as `certPath` and `keyPath`. See [Domains and TLS Certificates](preparations.md#domains-and-tls-certificates) for which certificate you need.
   - `cluster.version` - The cluster version to install. Omit it to install the latest supported version.
   - `cluster.customHelmValues` - Merges Helm values into the cluster chart, for settings not otherwise exposed by the installer. See [Advanced Cluster Configurations](../../../../installation-and-configuration/configuration/cluster-config.md) for the full set of keys.

   ```yaml
   cluster:
     enabled: true
     name: <CLUSTER_NAME>
     domain: <CLUSTER_DOMAIN>
     workloadRouting: host # Default; determines the certificate cluster.tls needs
     controlPlaneUrl: https://<CONTROL_PLANE_DOMAIN>
     uid: <CLUSTER_UUID>
     bootstrapTokenPath: /path/to/bootstrap-token
     tls:
       certPath: /path/to/cluster-cert.pem
       keyPath: /path/to/cluster-key.pem
     version: "" # Omit to install the latest supported version
     customHelmValues: {}
   ```

4. Enable the prerequisites:

   - `prerequisites.gpuOperator` - Required for every tenant cluster. See [Prerequisites and Features](prerequisites-and-features.md#prerequisites) for more details.
     - `mode` - Set to `install`.
     - `version` - Omit to install the latest version, or see [Support Matrix](../../support-matrix.md) for the supported range if you want to set it explicitly. GPU fractions (memory fractions, dynamic memory fractions, or compute modes) require 26.7.1 or later, installed automatically when omitted.
     - `customHelmValues` - Merges Helm values into the GPU Operator's own Helm chart. Set `driver.version` to r615 or later if this tenant plans to use GPU fractions, since GPU Operator 26.7.1 does not default to this driver version. See [GPU Memory Fractions](../../../../platform-management/runai-scheduler/resource-optimization/memory-fractions.md#prerequisites).
   - `prerequisites.prometheus` - Required for every tenant cluster. See [Prerequisites and Features](prerequisites-and-features.md#prerequisites) for more details.
     - `mode` - Set to `install`. On OpenShift, `install` is rejected - set to `skip` instead and use the cluster's built-in monitoring stack.
     - `version` - Omit to install the latest version, or see [Support Matrix](../../support-matrix.md) for the supported range.
     - `customHelmValues` - Merges Helm values into Prometheus's own Helm chart.

   ```yaml
   prerequisites:
     gpuOperator:
       mode: install
       version: ""
       customHelmValues: {}
     prometheus:
       mode: install
       version: ""
       customHelmValues: {}
   ```

   <Note>

   **Features**

   The minimum configuration does not include features such as Knative, Kubeflow, MPI Operator, LWS, NIM Operator, Grove, or Dynamo. These are skipped by default.

   If this tenant's workloads require any of these features, add the corresponding `features` configuration. See [Features](prerequisites-and-features.md#features).

   Some features require additional networking configuration. For example, externally exposed standard inference with Knative requires an additional inference wildcard FQDN and TLS certificate. See [Networking Setup](../../system-requirements/networking-setup.md).

   </Note>

5. Configure external access to the cluster:

   - `externalConnection` - Configures how traffic reaches this cluster's endpoints from outside the cluster, through a Kubernetes Ingress Controller or Gateway API implementation already running on the cluster. This is separate from `cluster.workloadRouting`, which controls how the workloads you submit are exposed. See [Networking Setup](../../system-requirements/networking-setup.md).

   ```yaml
   externalConnection:
     type: ingress
     ingress:
       className: <YOUR_INGRESS_CLASS>
   ```

The following example combines these settings to install a tenant's NVIDIA Run:ai cluster:

```yaml
install:
  mode: runai_cluster_only
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

cluster:
  enabled: true
  name: <CLUSTER_NAME>
  domain: <CLUSTER_DOMAIN>
  workloadRouting: host
  controlPlaneUrl: https://<CONTROL_PLANE_DOMAIN>
  uid: <CLUSTER_UUID>
  bootstrapTokenPath: /path/to/bootstrap-token
  tls:
    certPath: /path/to/cluster-cert.pem
    keyPath: /path/to/cluster-key.pem
  version: ""
  customHelmValues: {}

prerequisites:
  gpuOperator:
    mode: install
    version: ""
    customHelmValues: {}
  prometheus:
    mode: install
    version: ""
    customHelmValues: {}

externalConnection:
  type: ingress
  ingress:
    className: <YOUR_INGRESS_CLASS>
```

### Run the Installer

- **Validate first** - Before applying any changes, check your configuration against the target cluster:

  ```bash
  ./runai-installer install \
    --config cluster-config.yaml \
    --kubeconfig /path/to/tenant-a-kubeconfig \
    --preflight-only
  ```

- **Preview the plan** - Go further and preview the full installation plan without applying it:

  ```bash
  ./runai-installer install \
    --config cluster-config.yaml \
    --kubeconfig /path/to/tenant-a-kubeconfig \
    --dry-run
  ```

- **Run the installer** - Once you are ready, run it against the tenant's cluster:

  ```bash
  ./runai-installer install \
    --config cluster-config.yaml \
    --kubeconfig /path/to/tenant-a-kubeconfig \
    --logs-dir ./logs
  ```

Repeat this step for each tenant cluster, with its own configuration file, `uid`, and `bootstrapToken`.

### Verify the Installation

After the installer run completes successfully, verify that the NVIDIA Run:ai pods are running:

```bash
kubectl get pods -n runai
```

All pods should reach `Running` or `Completed` status. Querying the cluster through the [Clusters](api:multi-tenant-api:POST/api/v1/clusters) API should show a **Connected** status under its configured `cluster.name`.

## Share Tenant Access Details

Once the cluster is installed and connected, provide the tenant with:

* **NVIDIA Run:ai access URL** - The fully qualified domain name (FQDN) for their environment, e.g.: `https://tenant-a.runai.hostorg.com`. This is used to access the UI, API, and CLI.
* **Administrator credentials** - The username and password you configured during tenant creation. These are used to authenticate in all interfaces (UI/API/CLI).

## Command Reference and Troubleshooting

See [Command Reference and Troubleshooting](command-reference.md) for installer flags and common troubleshooting scenarios.
