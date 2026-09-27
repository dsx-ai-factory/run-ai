# Create and Install Clusters

Once system requirements are installed on the tenant's Kubernetes environment, you must create the cluster and initiate the installation process.

This step involves:

* **Registering the cluster** using the `/api/v1/clusters` API. You’ll provide a cluster name, the associated fully qualified domain name (FQDN), and the required NVIDIA Run:ai version.
* **Installing the NVIDIA Run:ai cluster component** using the `runai-installer` CLI, which connects the tenant’s cluster to the platform.
* **Establishing connectivity** between the tenant’s environment and the control plane to enable monitoring, workload submission, and scheduling.

The cluster domain must match the FQDN configured during the [system requirements](system-requirements.md) step. After completing this step, the tenant environment becomes fully operational within the NVIDIA Run:ai platform.

## Create and Register the Cluster

After installing the system requirements on the tenant’s Kubernetes environment, you must register the cluster in the NVIDIA Run:ai platform. This step creates a unique cluster object and associates it with the correct tenant using the provided domain.

To create a cluster:

* Send a `POST` request to the `/api/v1/clusters` endpoint. See [Clusters](api:multi-tenant-api:POST/api/v1/clusters) API for more details.
* Provide the cluster’s **name**, **domain** ([FQDN](system-requirements.md#fully-qualified-domain-name-fqdn)) and the target NVIDIA Run:ai [version](system-requirements.md#kubernetes-distribution).
* Once registered, the platform will return a cluster UUID. This UUID will be used to complete the installation and connect the tenant’s cluster to the NVIDIA Run:ai control plane.

This registration step does **not** install anything on the tenant’s cluster yet. It only prepares the cluster for installation and establishes the tenant-cluster association.

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
  "chartRepoURL": "https://helm.ngc.nvidia.com/nvidia/runai",
  "bootstrapToken": "ABC333DDD"
}
```

## Install the Cluster

Run the returned `installationStr` in the terminal on the tenant's Kubernetes cluster. This installs the NVIDIA Run:ai cluster component and links the cluster to the tenant in the control plane.

## Upgrade the Cluster

To upgrade a tenant's cluster to a newer version of the NVIDIA Run:ai platform, repeat the installation process using the updated target [version](system-requirements.md#kubernetes-distribution).
