---
---

# Multi-Cluster Installation

This page describes how to install the NVIDIA Run:ai control plane on a dedicated Kubernetes cluster, and one or more NVIDIA Run:ai clusters on separate Kubernetes clusters, from an air-gapped bundle. To install the control plane and cluster together on one Kubernetes cluster instead, see [Same-Cluster Installation](same-cluster.md).

The `runai-installer` CLI runs once per Kubernetes cluster with a dedicated configuration file:

1. **Install the control plane** (`mode: runai_control_plane_only`) on the control plane cluster.
2. **Install the cluster component** (`mode: runai_cluster_only`) on each NVIDIA Run:ai cluster, and register it with the control plane.

## Before You Start

Complete [Preparations](preparations.md) before continuing, including the system requirements for your Kubernetes cluster, registry credentials, downloading the `runai-installer` binary, and building your [Air-Gapped Bundle](preparations.md#air-gapped-bundle). The installer deploys and configures NVIDIA Run:ai itself; it does not provision this underlying infrastructure for you.

## Install the Control Plane

<Note>

**Configuration Templates**

The steps below build the minimum configuration for a multi-cluster installation on standard Kubernetes, from an air-gapped bundle - one file for the control plane, one for each NVIDIA Run:ai cluster. Depending on your Kubernetes distribution, workload routing, networking setup, and the prerequisites and workload features you need, you may require additional configuration.

If your use case matches one of the [Configuration Templates](templates.md), use that template instead of building the configuration from scratch. See [Prerequisites and Features](../prerequisites-and-features.md) for the components required by different workload types, and the [Configuration Reference](../configuration-reference.md) for all supported fields and options.

</Note>

Create a `cp-config.yaml` file for the control plane cluster. At a minimum, configure:

1. Define the installation:

   - `install.mode` - Set to `runai_control_plane_only` to install the control plane only.
   - `install.connectivity` - Set to `air_gapped`, for an installation with no outbound internet access.

   ```yaml
   install:
     mode: runai_control_plane_only
     connectivity: air_gapped
   ```

2. Configure the registry:

   - `registry.host` - The hostname of your own internal mirror registry (populated by loading the air-gapped bundle), not NGC.
   - `registry.username` - The username to authenticate to that registry.
   - `registry.passwordPath` - The path to a file containing the password for that user. See [Registry Credentials](preparations.md#registry-credentials).

   ```yaml
   registry:
     host: registry.mycorp.local
     username: admin
     passwordPath: /path/to/registry-password
   ```

3. Configure the air-gapped bundle:

   - `airgap` - The same `workingDir` and `valuesDir` you used with `airgap load`, plus `skipPrepare: true` - the bundle is already loaded, so `install` skips straight to the chart install. See [Air-Gapped Bundle](preparations.md#air-gapped-bundle) and [`airgap`](../configuration-reference.md#airgap) in the Configuration Reference.

   ```yaml
   airgap:
     workingDir: /data/runai-airgap
     valuesDir: /data/runai-airgap
     skipPrepare: true
   ```

4. Define the control plane:

   - `controlPlane.domain` - The FQDN for the NVIDIA Run:ai UI and API. See [FQDN and DNS](../../../system-requirements/networking-setup.md#fqdn-and-dns).
   - `controlPlane.tls` - A certificate and private key for that FQDN, as `certPath` and `keyPath`. See [TLS Certificates](../../../system-requirements/networking-setup.md#tls-certificates). If a single certificate covers every FQDN you're using (control plane, cluster, and inference), you can set it once at the root-level `tls` field instead of repeating it under each component - see [`tls`](../configuration-reference.md#tls).
   - `controlPlane.version` - The control plane version to install. Omit it to install the latest supported version.
   - `controlPlane.customHelmValues` - Merges Helm values into the control plane chart, for settings not otherwise exposed by the installer. See [Advanced Control Plane Configurations](../../../configuration/control-plane-config.md) for the full set of keys.
   - `controlPlane.customHelmValues.supportEntitled` - For NVIDIA Run:ai Support-Entitled Release, set to `true`.

   ```yaml
   controlPlane:
     enabled: true
     domain: runai.mycorp.local
     tls:
       certPath: /path/to/cp-cert.pem
       keyPath: /path/to/cp-key.pem
     version: "" # Omit to install the latest supported version
     customHelmValues:
       supportEntitled: true # Support-Entitled Release only
   ```

5. Configure the networking:

   - `externalConnection` - Configures how traffic reaches NVIDIA Run:ai's own services - the control plane UI/API and the cluster endpoint - through a Kubernetes Ingress Controller or Gateway API implementation already running on your cluster. Required on standard Kubernetes; omit this section entirely on OpenShift, where Routes are configured automatically. See [Networking Setup](../../../system-requirements/networking-setup.md).

   ```yaml
   externalConnection:
     type: ingress
     ingress:
       className: <your-ingress-class>
   ```

The following example combines these settings to install the NVIDIA Run:ai control plane from an already-loaded air-gapped bundle:

```yaml
install:
  mode: runai_control_plane_only
  connectivity: air_gapped

registry:
  host: registry.mycorp.local
  username: admin
  passwordPath: /path/to/registry-password

airgap:
  workingDir: /data/runai-airgap
  valuesDir: /data/runai-airgap
  skipPrepare: true

controlPlane:
  enabled: true
  domain: runai.mycorp.local
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  version: ""
  customHelmValues:
    supportEntitled: true # Support-Entitled Release only

externalConnection:
  type: ingress
  ingress:
    className: <your-ingress-class>
```

- **Validate first** - Before applying any changes, check your configuration against the control plane cluster:

  ```bash
  ./runai-installer install \
    --config cp-config.yaml \
    --kubeconfig /path/to/cp-kubeconfig \
    --preflight-only
  ```

- **Preview the plan** - Go further and preview the full installation plan without applying it:

  ```bash
  ./runai-installer install \
    --config cp-config.yaml \
    --kubeconfig /path/to/cp-kubeconfig \
    --dry-run
  ```

- **Run the installer** - Once you're ready, run it against the control plane cluster:

  ```bash
  ./runai-installer install \
    --config cp-config.yaml \
    --kubeconfig /path/to/cp-kubeconfig \
    --logs-dir ./logs
  ```

When the installation completes, verify that the control plane is reachable at its FQDN (for example, `https://runai.mycorp.local`) before continuing.

## Install Each NVIDIA Run:ai Cluster

Create a `cluster-config.yaml` file for each NVIDIA Run:ai cluster. The control plane must already be installed and reachable at its FQDN.

At a minimum, configure:

1. Define the installation:

   - `install.mode` - Set to `runai_cluster_only` to install the cluster component only.
   - `install.connectivity` - Set to `air_gapped`, for an installation with no outbound internet access.

   ```yaml
   install:
     mode: runai_cluster_only
     connectivity: air_gapped
   ```

2. Configure the registry:

   - `registry.host` - The hostname of your own internal mirror registry (populated by loading the air-gapped bundle), not NGC.
   - `registry.username` - The username to authenticate to that registry.
   - `registry.passwordPath` - The path to a file containing the password for that user. See [Registry Credentials](preparations.md#registry-credentials).

   ```yaml
   registry:
     host: registry.mycorp.local
     username: admin
     passwordPath: /path/to/registry-password
   ```

3. Configure the air-gapped bundle:

   - `airgap` - The same `workingDir` and `valuesDir` you used with `airgap load`, plus `skipPrepare: true` - the bundle is already loaded, so `install` skips straight to the chart install. See [Air-Gapped Bundle](preparations.md#air-gapped-bundle) and [`airgap`](../configuration-reference.md#airgap) in the Configuration Reference.

   ```yaml
   airgap:
     workingDir: /data/runai-airgap-cluster-a
     valuesDir: /data/runai-airgap-cluster-a
     skipPrepare: true
   ```

4. Define the cluster:

   - `cluster.name` - The name of the NVIDIA Run:ai cluster as it appears in the control plane.
   - `cluster.domain` - The FQDN for this cluster. Required when the cluster component is installed on a different Kubernetes cluster than the control plane. The installer also uses this domain to reach the control plane when registering the cluster.
   - `cluster.controlPlaneUrl` - The control plane's FQDN, reachable over HTTPS from this cluster.
   - `cluster.uid` plus one of `cluster.bootstrapTokenPath`/`cluster.clientSecretPath`, or `cluster.controlPlaneCredentials` - Identifies this cluster to the control plane. See [Preparations](../connected/preparations.md#system-requirements) for the two available methods.
   - `cluster.workloadRouting` - `host` (the default) or `path`. This determines which certificate you need for `cluster.tls`, below.
   - `cluster.tls` - A certificate and private key for the cluster domain, as `certPath` and `keyPath`. Which certificate you need depends on `workloadRouting` and the workload types you plan to support - see [TLS Certificates](../../../system-requirements/networking-setup.md#tls-certificates).
   - `cluster.version` - The cluster version to install. Omit it to install the latest supported version.
   - `cluster.customHelmValues` - Merges Helm values into the cluster chart, for settings not otherwise exposed by the installer. See [Advanced Cluster Configurations](../../../configuration/cluster-config.md) for the full set of keys.

   ```yaml
   cluster:
     enabled: true
     name: cluster-a
     domain: runai-a.mycorp.local
     workloadRouting: host # Default; determines the certificate cluster.tls needs
     controlPlaneUrl: https://runai.mycorp.local
     controlPlaneCredentials:
       tokenPath: /path/to/cp-token
     tls:
       certPath: /path/to/cluster-a-cert.pem
       keyPath: /path/to/cluster-a-key.pem
     version: "" # Omit to install the latest supported version
     customHelmValues: {}
   ```

5. Configure the networking:

   - `externalConnection` - Configures how traffic reaches NVIDIA Run:ai's own services - the control plane UI/API and the cluster endpoint - through a Kubernetes Ingress Controller or Gateway API implementation already running on your cluster. This is separate from `cluster.workloadRouting`, which controls how the workloads you submit are exposed. Required on standard Kubernetes; omit this section entirely on OpenShift, where Routes are configured automatically. See [Networking Setup](../../../system-requirements/networking-setup.md).

   ```yaml
   externalConnection:
     type: ingress
     ingress:
       className: <your-ingress-class>
   ```

<Note>

**Note**

GPU Operator and Prometheus are required for every installation, but the installer does not install them in air-gapped environments - they are skipped by default. Install both yourself beforehand, using each vendor's own air-gapped installation instructions. Optional features such as Knative, Kubeflow, MPI Operator, LWS, NIM Operator, Dynamo, and Grove are also skipped by default and follow the same pattern - install what your workloads require, then add the corresponding `features` configuration. See [Prerequisites and Features](preparations.md#prerequisites-and-features).

</Note>

The following example combines these settings to install the NVIDIA Run:ai cluster component from an already-loaded air-gapped bundle, and has the installer register it with the control plane using a control plane token:

```yaml
install:
  mode: runai_cluster_only
  connectivity: air_gapped

registry:
  host: registry.mycorp.local
  username: admin
  passwordPath: /path/to/registry-password

airgap:
  workingDir: /data/runai-airgap-cluster-a
  valuesDir: /data/runai-airgap-cluster-a
  skipPrepare: true

cluster:
  enabled: true
  name: cluster-a
  domain: runai-a.mycorp.local
  workloadRouting: host
  controlPlaneUrl: https://runai.mycorp.local
  controlPlaneCredentials:
    tokenPath: /path/to/cp-token
  tls:
    certPath: /path/to/cluster-a-cert.pem
    keyPath: /path/to/cluster-a-key.pem
  version: ""
  customHelmValues: {}

externalConnection:
  type: ingress
  ingress:
    className: <your-ingress-class>
```

- **Validate first** - Before applying any changes, check your configuration against the target cluster:

  ```bash
  ./runai-installer install \
    --config cluster-config.yaml \
    --kubeconfig /path/to/cluster-a-kubeconfig \
    --preflight-only
  ```

- **Preview the plan** - Go further and preview the full installation plan without applying it:

  ```bash
  ./runai-installer install \
    --config cluster-config.yaml \
    --kubeconfig /path/to/cluster-a-kubeconfig \
    --dry-run
  ```

- **Run the installer** - Once you're ready, run it against the target cluster:

  ```bash
  ./runai-installer install \
    --config cluster-config.yaml \
    --kubeconfig /path/to/cluster-a-kubeconfig \
    --logs-dir ./logs
  ```

Repeat this step for each additional NVIDIA Run:ai cluster with its own configuration file - each cluster requires a unique `cluster.name` and its own `cluster.domain`.

## Verify the Installation

After each installer run completes successfully, verify that the NVIDIA Run:ai pods are running:

```bash
# On the control plane cluster
kubectl get pods -n runai-backend

# On each NVIDIA Run:ai cluster
kubectl get pods -n runai
```

All pods should reach `Running` or `Completed` status.

Then log-in to the NVIDIA Run:ai UI at the control plane domain and verify that each installed cluster appears under its configured `cluster.name` with a **Connected** status. If a cluster does not show as **Connected**, see [Troubleshooting scenarios](../../../operations/clusters.md#troubleshooting-scenarios).

## Command Reference and Troubleshooting

See [Command Reference and Troubleshooting](../command-reference.md) for installer flags and common troubleshooting scenarios.
