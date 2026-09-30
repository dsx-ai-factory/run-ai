---
---

# Same-Cluster Installation

This page describes how to install the NVIDIA Run:ai control plane and cluster together, on one Kubernetes cluster (`mode: full_installation`), over an internet-connected environment. To install the control plane and cluster on separate Kubernetes clusters instead, see [Multi-Cluster Installation](multi-cluster.md).

## Before You Start

Complete [Preparations](preparations.md) before continuing, including the system requirements for your Kubernetes cluster, any required registry credentials, and downloading the `runai-installer` binary. The installer deploys and configures NVIDIA Run:ai itself; it does not provision this underlying infrastructure for you.

## Create a Configuration File

<Note>

**Configuration Templates**

The steps below build the minimum configuration for a same-cluster installation on standard Kubernetes. Depending on your Kubernetes distribution, workload routing, networking setup, and the prerequisites and workload features you need, you may require additional configuration.

If your use case matches one of the [Configuration Templates](templates.md), use that template instead of building the configuration from scratch. See [Prerequisites and Features](../prerequisites-and-features.md) for the components required by different workload types, and the [Configuration Reference](../configuration-reference.md) for all supported fields and options.

</Note>

Create a `config.yaml` file describing your installation. Use the domains, TLS certificates, and networking values you prepared in [Networking Setup](../../../system-requirements/networking-setup.md) when completing the configuration below.

At a minimum, configure:

1. Define the installation:

   - `install.mode` - Set to `full_installation` to install the control plane and cluster on the same Kubernetes cluster.
   - `install.connectivity` - Set to `connected`, for an installation with outbound internet access.

   ```yaml
   install:
     mode: full_installation
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

4. Define the cluster:

   - `cluster.name` - The name of the NVIDIA Run:ai cluster.
   - `cluster.domain` - Omit it for a same-cluster installation - the cluster reuses `controlPlane.domain`, since both run on the same Kubernetes cluster.
   - `cluster.workloadRouting` - `host` (the default) or `path`. This determines which certificate you need for `cluster.tls` below.
   - `cluster.tls` - A certificate and private key for the cluster domain, as `certPath` and `keyPath`. Which certificate you need depends on `workloadRouting` and the workload types you plan to support - see [TLS Certificates](../../../system-requirements/networking-setup.md#tls-certificates).
   - `cluster.version` - The cluster version to install. Omit it to install the latest supported version.
   - `cluster.customHelmValues` - Merges Helm values into the cluster chart, for settings not otherwise exposed by the installer. See [Advanced Cluster Configurations](../../../configuration/cluster-config.md) for the full set of keys.

   ```yaml
   cluster:
     enabled: true
     name: <CLUSTER_NAME>
     workloadRouting: host # Default; determines the certificate cluster.tls needs
     tls:
       certPath: /path/to/cluster-cert.pem
       keyPath: /path/to/cluster-key.pem
     version: "" # Omit to install the latest supported version
     customHelmValues: {}
   ```

5. Enable the prerequisites:

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

6. Configure the networking:

   - `externalConnection` - Configures how traffic reaches NVIDIA Run:ai's own services - the control plane UI/API and the cluster endpoint - through a Kubernetes Ingress Controller or Gateway API implementation already running on your cluster. This is separate from `cluster.workloadRouting`, which controls how the workloads you submit are exposed. See [Networking Setup](../../../system-requirements/networking-setup.md).

   ```yaml
   externalConnection:
     type: ingress
     ingress:
       className: <YOUR_INGRESS_CLASS>
   ```

The following example combines these settings to install the NVIDIA Run:ai control plane and cluster on the same Kubernetes cluster:

```yaml
install:
  mode: full_installation
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

cluster:
  enabled: true
  name: <CLUSTER_NAME>
  workloadRouting: host
  tls:
    certPath: /path/to/cluster-cert.pem
    keyPath: /path/to/cluster-key.pem
  version: ""
  customHelmValues: {}

prerequisites:
  gpuOperator:
    mode: install
    version: ""
  prometheus:
    mode: install
    version: ""

externalConnection:
  type: ingress
  ingress:
    className: <YOUR_INGRESS_CLASS>
```

## Run the Installer

- **Validate first** - Before applying any changes, check your configuration against the target cluster:

  ```bash
  ./runai-installer install \
    --config /path/to/config.yaml \
    --preflight-only
  ```

- **Preview the plan** - Go further and preview the full installation plan without applying it:

  ```bash
  ./runai-installer install \
    --config /path/to/config.yaml \
    --dry-run
  ```

- **Run the installer** - Once you are ready:

  ```bash
  ./runai-installer install \
    --config /path/to/config.yaml \
    --logs-dir ./logs
  ```

The `runai-installer` CLI:

1. Loads and validates the configuration file.
2. Connects to the Kubernetes cluster and runs preflight checks.
3. Prompts for confirmation before applying any changes.
4. Installs required prerequisites.
5. Deploys the NVIDIA Run:ai control plane and cluster components.
6. Reports the status and elapsed time for each step.

## Verify the Installation

After installation completes successfully, verify that the NVIDIA Run:ai pods are running:

```bash
# Control plane
kubectl get pods -n runai-backend

# Cluster
kubectl get pods -n runai
```

All pods should reach `Running` or `Completed` status.

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

Re-run the installer as described in [Run the Installer](#run-the-installer) to apply the configuration.

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

Re-run the installer as described in [Run the Installer](#run-the-installer) to apply the configuration.

</Accordion>

## Connect to NVIDIA Run:ai User Interface

<Note>

**Note**

After installation, it may take a few minutes for the UI to become accessible (up to 15 minutes). Accessing it too early may result in a "server cannot be reached" error.

</Note>

1. Retrieve the generated control plane administrator username and password:

   ```bash
   kubectl get configmap runai-backend-tenants-manager -n runai-backend -o jsonpath='{.data.ADMIN_USERNAME}'
   kubectl get secret runai-backend-gen-tenants-manager -n runai-backend -o jsonpath='{.data.ADMIN_PASSWORD}' | base64 -d
   ```

2. Open your browser and go to the control plane domain you configured (for example, `https://runai.mycorp.local`).
3. Log in using the credentials retrieved above. It is recommended to change the password after the first login.
4. Verify that the cluster appears with a **Connected** status. If the cluster does not show as **Connected**, see [Troubleshooting scenarios](../../../operations/clusters.md#troubleshooting-scenarios).

## Command Reference and Troubleshooting

See [Command Reference and Troubleshooting](../command-reference.md) for installer flags and common troubleshooting scenarios.
