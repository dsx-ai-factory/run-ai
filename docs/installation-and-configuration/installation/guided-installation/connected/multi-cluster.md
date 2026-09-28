---
---

# Multi-Cluster Installation

This page describes how to install the NVIDIA Run:ai control plane on a dedicated Kubernetes cluster, and one or more NVIDIA Run:ai clusters on separate Kubernetes clusters, over an internet-connected environment. To install the control plane and cluster together on one Kubernetes cluster instead, see [Same-Cluster Installation](same-cluster.md).

The `runai-installer` CLI runs once per Kubernetes cluster with a dedicated configuration file:

1. **Install the control plane** (`mode: runai_control_plane_only`) on the control plane cluster.
2. **Install the cluster component** (`mode: runai_cluster_only`) on each NVIDIA Run:ai cluster, and register it with the control plane.

## Before You Start

Complete [Preparations](preparations.md) before continuing, including the system requirements for your Kubernetes cluster, any required registry credentials, and downloading the `runai-installer` binary. The installer deploys and configures NVIDIA Run:ai itself; it does not provision this underlying infrastructure for you.

## Install the Control Plane

<Note>

**Configuration Templates**

The steps below build the minimum configuration for a multi-cluster installation on standard Kubernetes - one file for the control plane, one for each NVIDIA Run:ai cluster. Depending on your Kubernetes distribution, workload routing, networking setup, and the prerequisites and workload features you need, you may require additional configuration.

If your use case matches one of the [Configuration Templates](templates.md), use that template instead of building the configuration from scratch. See [Prerequisites and Features](../prerequisites-and-features.md) for the components required by different workload types, and the [Configuration Reference](../configuration-reference.md) for all supported fields and options.

</Note>

Create a `cp-config.yaml` file for the control plane cluster. Use the domains, TLS certificates, and networking values you prepared in [Networking Setup](../../../system-requirements/networking-setup.md) when completing the configuration below.

At a minimum, configure:

1. Define the installation:

   - `install.mode` - Set to `runai_control_plane_only` to install the control plane only.
   - `install.connectivity` - Set to `connected`, for an installation with outbound internet access.

   ```yaml
   install:
     mode: runai_control_plane_only
     connectivity: connected
   ```

2. Configure the registry:

   - `registry.type` - Set to `ngc`.
   - `registry.tokenPath` - For NVIDIA Run:ai Support-Entitled Release installations, provide the path to your NGC Personal Key. Open Source installations do not require registry credentials. See [Registry Credentials](preparations.md#registry-credentials).

   ```yaml
   registry:
     type: ngc
     tokenPath: /path/to/ngc-token # Support-Entitled Release only
   ```

3. Define the control plane:

   - `controlPlane.domain` - The FQDN for the NVIDIA Run:ai UI and API. See [FQDN and DNS](../../../system-requirements/networking-setup.md#fqdn-and-dns).
   - `controlPlane.tls` - The certificate and private key for the control plane FQDN, as `certPath` and `keyPath`. See [TLS Certificates](../../../system-requirements/networking-setup.md#tls-certificates).
   - `controlPlane.version` - The control plane version to install. Omit it to install the latest supported version.
   - `controlPlane.customHelmValues` - Merges Helm values into the control plane chart, for settings not otherwise exposed by the installer. See [Advanced Control Plane Configurations](../../../configuration/control-plane-config.md) for the full set of keys.
   - `controlPlane.customHelmValues.supportEntitled` - For NVIDIA Run:ai Support-Entitled Release, set to `true`.

   ```yaml
   controlPlane:
     enabled: true
     domain: <DOMAIN>
     tls:
       certPath: /path/to/cp-cert.pem
       keyPath: /path/to/cp-key.pem
     version: "" # Omit to install the latest supported version
     customHelmValues:
       supportEntitled: true # Support-Entitled Release only
   ```

4. Configure the networking:

   - `externalConnection` - Configures how traffic reaches NVIDIA Run:ai's own services - the control plane UI/API and the cluster endpoint - through a Kubernetes Ingress Controller or Gateway API implementation already running on your cluster. See [Networking Setup](../../../system-requirements/networking-setup.md).

   ```yaml
   externalConnection:
     type: ingress
     ingress:
       className: <YOUR_INGRESS_CLASS>
   ```

The following example combines these settings to install the NVIDIA Run:ai control plane:

```yaml
install:
  mode: runai_control_plane_only
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

controlPlane:
  enabled: true
  domain: <DOMAIN>
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  version: ""
  customHelmValues:
    supportEntitled: true # Support-Entitled Release only

externalConnection:
  type: ingress
  ingress:
    className: <YOUR_INGRESS_CLASS>
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

- **Run the installer** - Once you are ready, run it against the control plane cluster:

  ```bash
  ./runai-installer install \
    --config cp-config.yaml \
    --kubeconfig /path/to/cp-kubeconfig \
    --logs-dir ./logs
  ```

When the installation completes, verify that the control plane is reachable at its FQDN (for example, `https://runai.mycorp.local`) before continuing.

## Enable Telemetry Service (Support-Entitled)

NVIDIA Run:ai includes telemetry for Support-Entitled Release installations, reporting initial metrics including platform usage and adoption signals, GPU utilization and allocation, and cluster health signals to support proactive troubleshooting and continued product improvement. By default, telemetry is not associated with your organization. You can opt in to identified telemetry by providing an [NGC Service API Key](https://docs.nvidia.com/ngc/latest/ngc-catalog-user-guide.html#generating-a-service-api-key) (SAK) in your configuration file, which associates the telemetry with your NCA ID.

Configure the SAK through `controlPlane.customHelmValues`, using one of the following approaches:

<Accordion title="Pass the SAK Value Directly">

**Pass the SAK Value Directly**

Set the SAK as a plaintext value in the configuration file. The chart creates and manages the Kubernetes Secret for you.

```yaml
controlPlane:
  ...
  customHelmValues:
    supportEntitled: true
    telemetryService:
      telemetryAuth:
        token: <SAK_TOKEN>
```

<Warning>

**Warning**

The SAK value is stored in plaintext in the configuration file. Protect the file accordingly, or use the pre-created Secret approach below.

</Warning>

Re-run the installer against the control plane cluster to apply the configuration.

</Accordion>

<Accordion title="Reference a Pre-Created Secret">

**Reference a Pre-Created Secret**

Create the Secret containing the SAK yourself before running the installer:

```bash
kubectl create secret generic runai-telemetry-sak \
  --namespace runai-backend \
  --from-literal=TELEMETRY_TOKEN=<SAK_TOKEN>
```

Then reference it by name in the configuration file - the installer does not create or manage this Secret. Always set the secret key name explicitly, even if it matches the default (`TELEMETRY_TOKEN`):

```yaml
controlPlane:
  ...
  customHelmValues:
    supportEntitled: true
    telemetryService:
      telemetryAuth:
        existingSecret: runai-telemetry-sak
        secretKey: TELEMETRY_TOKEN
```

Re-run the installer against the control plane cluster to apply the configuration.

</Accordion>

## Retrieve Control Plane Credentials

Retrieve the generated control plane administrator username and password - use these when installing NVIDIA Run:ai clusters below:

```bash
kubectl get configmap runai-backend-tenants-manager -n runai-backend -o jsonpath='{.data.ADMIN_USERNAME}'
kubectl get secret runai-backend-gen-tenants-manager -n runai-backend -o jsonpath='{.data.ADMIN_PASSWORD}' | base64 -d
```

## Install Each NVIDIA Run:ai Cluster

Create a `cluster-config.yaml` file for each NVIDIA Run:ai cluster. The control plane must already be installed and reachable at its FQDN. Use the domains, TLS certificates, and networking values you prepared in [Networking Setup](../../../system-requirements/networking-setup.md) when completing the configuration below.

At a minimum, configure:

1. Define the installation:

   - `install.mode` - Set to `runai_cluster_only` to install the cluster component only.
   - `install.connectivity` - Set to `connected`, for an installation with outbound internet access.

   ```yaml
   install:
     mode: runai_cluster_only
     connectivity: connected
   ```

2. Configure the registry:

   - `registry.type` - Set to `ngc`.
   - `registry.tokenPath` - For NVIDIA Run:ai Support-Entitled Release installations, provide the path to your NGC Personal Key. Open Source installations do not require registry credentials. See [Registry Credentials](preparations.md#registry-credentials).

   ```yaml
   registry:
     type: ngc
     tokenPath: /path/to/ngc-token # Support-Entitled Release only
   ```

3. Define the cluster:

   - `cluster.name` - The name of the NVIDIA Run:ai cluster as it appears in the control plane.
   - `cluster.domain` - The FQDN for this cluster. See [FQDN and DNS](../../../system-requirements/networking-setup.md#fqdn-and-dns). Required when the cluster component is installed on a different Kubernetes cluster than the control plane. The installer also uses this domain to reach the control plane when registering the cluster.
   - `cluster.controlPlaneUrl` - The control plane's FQDN, reachable over HTTPS from this cluster.
   - `cluster.controlPlaneCredentials` - Identifies this cluster to the control plane. Use the control plane administrator username and password generated during control plane installation - the installer uses these to register the cluster and fetch its bootstrap token automatically, in the same run. See [Retrieve Control Plane Credentials](#retrieve-control-plane-credentials).
   - `cluster.workloadRouting` - `host` (the default) or `path`. This determines which certificate you need for `cluster.tls` below.
   - `cluster.tls` - A certificate and private key for the cluster domain, as `certPath` and `keyPath`. Which certificate you need depends on `workloadRouting` and the workload types you plan to support - see [TLS Certificates](../../../system-requirements/networking-setup.md#tls-certificates).
   - `cluster.version` - The cluster version to install. Omit it to install the latest supported version.
   - `cluster.customHelmValues` - Merges Helm values into the cluster chart, for settings not otherwise exposed by the installer. See [Advanced Cluster Configurations](../../../configuration/cluster-config.md) for the full set of keys.

   ```yaml
   cluster:
     enabled: true
     name: <CLUSTER_NAME>
     domain: <CLUSTER_DOMAIN>
     workloadRouting: host # Default; determines the certificate cluster.tls needs
     controlPlaneUrl: https://<DOMAIN>
     controlPlaneCredentials:
       username: <ADMIN_EMAIL>
       passwordPath: /path/to/cp-password
     tls:
       certPath: /path/to/cluster-a-cert.pem
       keyPath: /path/to/cluster-a-key.pem
     version: "" # Omit to install the latest supported version
     customHelmValues: {}
   ```

4. Enable the prerequisites:

   - `prerequisites.gpuOperator.mode` / `prerequisites.prometheus.mode` - Set both to `install`; required for every installation.
   - `version` - Each also accepts its own `version` field. Omit it to install the latest version, or see [Support Matrix](../../../support-matrix.md) for the supported version range if you want to set it explicitly. See [Prerequisites and Features](../prerequisites-and-features.md#prerequisites).
   - `customHelmValues` - Merges Helm values into the component's own Helm chart, for settings not otherwise exposed by the installer. The available keys are determined by that component, not NVIDIA Run:ai - see [Prerequisites and Features](../prerequisites-and-features.md#prerequisites) for a link to each component's own documentation.

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

   The minimum configuration does not include features such as Knative, Kubeflow, MPI Operator, LWS, NIM Operator, Dynamo, or Grove. These are skipped by default.

   If your workloads require any of these features, add the corresponding `features` configuration. See [Prerequisites and Features](../prerequisites-and-features.md#features).

   Some features require additional networking configuration. For example, externally exposed standard inference with Knative requires an additional inference wildcard FQDN and TLS certificate. See [Networking Setup](../../../system-requirements/networking-setup.md).

   </Note>

5. Configure the networking:

   - `externalConnection` - Configures how traffic reaches NVIDIA Run:ai's own services - the control plane UI/API and the cluster endpoint - through a Kubernetes Ingress Controller or Gateway API implementation already running on your cluster. This is separate from `cluster.workloadRouting`, which controls how the workloads you submit are exposed. See [Networking Setup](../../../system-requirements/networking-setup.md).

   ```yaml
   externalConnection:
     type: ingress
     ingress:
       className: <YOUR_INGRESS_CLASS>
   ```

The following example combines these settings to install the NVIDIA Run:ai cluster component on a separate Kubernetes cluster, and has the installer register it with the control plane using the control plane administrator username and password:

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
  controlPlaneUrl: https://<DOMAIN>
  controlPlaneCredentials:
    username: <ADMIN_EMAIL>
    passwordPath: /path/to/cp-password
  tls:
    certPath: /path/to/cluster-a-cert.pem
    keyPath: /path/to/cluster-a-key.pem
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

- **Run the installer** - Once you are ready, run it against the target cluster:

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

## Connect to NVIDIA Run:ai User Interface

<Note>

**Note**

After installation, it may take a few minutes for the UI to become accessible (up to 15 minutes). Accessing it too early may result in a "server cannot be reached" error.

</Note>

1. Open your browser and go to the control plane domain (for example, `https://runai.mycorp.local`).
2. Log in using the control plane administrator username and password retrieved earlier. It is recommended to change the password after the first login.
3. Verify that each installed cluster appears under its configured `cluster.name` with a **Connected** status. If a cluster does not show as **Connected**, see [Troubleshooting scenarios](../../../operations/clusters.md#troubleshooting-scenarios).

## Command Reference and Troubleshooting

See [Command Reference and Troubleshooting](../command-reference.md) for installer flags and common troubleshooting scenarios.
