---
---

# Tenant Cluster Installation

This page describes how to install a tenant's NVIDIA Run:ai cluster using the `runai-installer` CLI, after [creating the tenant](../organization-onboarding/create-a-tenant.md) and provisioning its Kubernetes cluster.

The `runai-installer` CLI runs once per tenant cluster with a dedicated configuration file, using a cluster identity obtained from the control plane's Management API - it does not register the cluster with a tenant itself.

## Before You Start

Complete the following before continuing:

* [Install the control plane](installation.md) in multi-tenant mode, if you haven't already.
* [Create the tenant](../organization-onboarding/create-a-tenant.md) this cluster belongs to.
* [Provision the tenant's Kubernetes cluster](../organization-onboarding/provision-kubernetes-clusters.md) and complete its [system requirements](../organization-onboarding/system-requirements.md).

## Register the Cluster

Register the cluster with the tenant it belongs to, using the [Clusters](api:multi-tenant-api:POST/api/v1/clusters) API:

* Send a `POST` request to the `/api/v1/clusters` endpoint, including the tenant's `tenantId`.
* The response includes a `uuid` identifying the cluster.
* Send a `GET` request to `/v1/clusters/{clusterUuid}/cluster-install-info?version=<version>` using that `uuid` to retrieve a `bootstrapToken`.

Save the `bootstrapToken` to a local file - the installer reads it from disk rather than accepting it inline, to avoid storing it in the configuration file.

## Install the Cluster

<Note>

**Configuration Templates**

The steps below build the minimum configuration to install a tenant cluster on standard Kubernetes. Depending on your Kubernetes distribution, workload routing, networking setup, and the prerequisites and workload features you need, you may require additional configuration. See [Prerequisites and Features](prerequisites-and-features.md) for the components required by different workload types, and the [Configuration Reference](configuration-reference.md) for all supported fields and options.

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
   - `cluster.domain` - The tenant cluster's own FQDN. See [FQDN and DNS](../../system-requirements/networking-setup.md#fqdn-and-dns).
   - `cluster.controlPlaneUrl` - The control plane's domain, reachable over HTTPS from this cluster.
   - `cluster.uid` - The cluster `uuid` returned when you [registered the cluster](#register-the-cluster).
   - `cluster.bootstrapTokenPath` - Path to the `bootstrapToken` file you saved when you [registered the cluster](#register-the-cluster). Mutually exclusive with `cluster.controlPlaneCredentials` - the installer does not register tenant clusters itself, since doing so requires a `tenantId` that only the Management API accepts.
   - `cluster.workloadRouting` - `host` (the default) or `path`. This determines which certificate you need for `cluster.tls` below.
   - `cluster.tls` - A certificate and private key for the cluster domain, as `certPath` and `keyPath`. Which certificate you need depends on `workloadRouting` and the workload types you plan to support - see [TLS Certificates](../../system-requirements/networking-setup.md#tls-certificates).
   - `cluster.version` - The cluster version to install. Omit it to install the latest supported version.
   - `cluster.customHelmValues` - Merges Helm values into the cluster chart, for settings not otherwise exposed by the installer. See [Advanced Cluster Configurations](../../infrastructure-setup/advanced-setup/cluster-config.md) for the full set of keys.

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

   - `prerequisites.gpuOperator.mode` / `prerequisites.prometheus.mode` - Set both to `install`; required for every tenant cluster. See [NVIDIA GPU Operator](prerequisites-and-features.md#nvidia-gpu-operator) and [Prometheus](prerequisites-and-features.md#prometheus).
   - `version` - Each also accepts its own `version` field. Omit it to install the latest version.
   - `customHelmValues` - Merges Helm values into the component's own Helm chart, for settings not otherwise exposed by the installer.

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

   The minimum configuration does not include features such as Knative, Kubeflow, MPI Operator, or LWS. These are skipped by default. If this tenant's workloads require any of these features, add the corresponding `features` configuration - see [Features](prerequisites-and-features.md#features).

   </Note>

5. Configure the networking:

   - `externalConnection` - Configures how traffic reaches this cluster's endpoints through a Kubernetes Ingress Controller or Gateway API implementation already running on the cluster. This is separate from `cluster.workloadRouting`, which controls how the workloads you submit are exposed. See [Networking Setup](../../system-requirements/networking-setup.md).

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

## Run the Installer

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

- **Run the installer** - Once you're ready, run it against the tenant's cluster:

  ```bash
  ./runai-installer install \
    --config cluster-config.yaml \
    --kubeconfig /path/to/tenant-a-kubeconfig \
    --logs-dir ./logs
  ```

Repeat this step for each tenant cluster, with its own configuration file, `uid`, and `bootstrapToken`.

## Verify the Installation

After the installer run completes successfully, verify that the NVIDIA Run:ai pods are running:

```bash
kubectl get pods -n runai
```

All pods should reach `Running` or `Completed` status. Querying the cluster through the [Clusters](api:multi-tenant-api:POST/api/v1/clusters) API should show a **Connected** status under its configured `cluster.name`.

## Command Reference and Troubleshooting

See [Command Reference and Troubleshooting](command-reference.md) for installer flags and common troubleshooting scenarios.
