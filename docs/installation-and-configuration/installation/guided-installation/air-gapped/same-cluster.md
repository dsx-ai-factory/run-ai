---
---

# Same-Cluster Installation

This page describes how to install the NVIDIA Run:ai control plane and cluster together, on one Kubernetes cluster (`mode: full_installation`), from an air-gapped bundle. To install the control plane and cluster on separate Kubernetes clusters instead, see [Multi-Cluster Installation](multi-cluster.md).

## Before You Start

Complete [Preparations](preparations.md) before continuing, including the system requirements for your Kubernetes cluster, registry credentials, downloading the `runai-installer` binary, and building your [Air-Gapped Bundle](preparations.md#air-gapped-bundle). The installer deploys and configures NVIDIA Run:ai itself; it does not provision this underlying infrastructure for you.

## Create a Configuration File

<Note>

**Configuration Templates**

The steps below build the minimum configuration for a same-cluster installation on standard Kubernetes, from an air-gapped bundle. Depending on your Kubernetes distribution, workload routing, networking setup, and the prerequisites and workload features you need, you may require additional configuration.

If your use case matches one of the [Configuration Templates](templates.md), use that template instead of building the configuration from scratch. See [Prerequisites and Features](../prerequisites-and-features.md) for the components required by different workload types, and the [Configuration Reference](../configuration-reference.md) for all supported fields and options.

</Note>

Create a `config.yaml` file describing your installation. At a minimum, configure:

1. Define the installation:

   - `install.mode` - Set to `full_installation` to install the control plane and cluster on the same Kubernetes cluster.
   - `install.connectivity` - Set to `air_gapped`, for an installation with no outbound internet access.

   ```yaml
   install:
     mode: full_installation
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

5. Define the cluster:

   - `cluster.name` - The name of the NVIDIA Run:ai cluster.
   - `cluster.domain` - Omit it for a same-cluster installation - the cluster reuses `controlPlane.domain`, since both run on the same Kubernetes cluster.
   - `cluster.workloadRouting` - `host` (the default) or `path`. This determines which certificate you need for `cluster.tls`, below.
   - `cluster.tls` - A certificate and private key for the cluster domain, as `certPath` and `keyPath`. Which certificate you need depends on `workloadRouting` and the workload types you plan to support - see [TLS Certificates](../../../system-requirements/networking-setup.md#tls-certificates).
   - `cluster.version` - The cluster version to install. Omit it to install the latest supported version.
   - `cluster.customHelmValues` - Merges Helm values into the cluster chart, for settings not otherwise exposed by the installer. See [Advanced Cluster Configurations](../../../configuration/cluster-config.md) for the full set of keys.

   ```yaml
   cluster:
     enabled: true
     name: my-cluster
     workloadRouting: host # Default; determines the certificate cluster.tls needs
     tls:
       certPath: /path/to/cluster-wildcard-cert.pem
       keyPath: /path/to/cluster-wildcard-key.pem
     version: "" # Omit to install the latest supported version
     customHelmValues: {}
   ```

6. Configure the networking:

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

The following example combines these settings to install the NVIDIA Run:ai control plane and cluster on the same Kubernetes cluster from an already-loaded air-gapped bundle:

```yaml
install:
  mode: full_installation
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

cluster:
  enabled: true
  name: my-cluster
  workloadRouting: host
  tls:
    certPath: /path/to/cluster-wildcard-cert.pem
    keyPath: /path/to/cluster-wildcard-key.pem
  version: ""
  customHelmValues: {}

externalConnection:
  type: ingress
  ingress:
    className: <your-ingress-class>
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

- **Run the installer** - Once you're ready:

  ```bash
  ./runai-installer install \
    --config /path/to/config.yaml \
    --logs-dir ./logs
  ```

The `runai-installer` CLI:

1. Loads and validates the configuration file.
2. Connects to the Kubernetes cluster and runs preflight checks.
3. Prompts for confirmation before applying any changes.
4. Deploys the NVIDIA Run:ai control plane and cluster components.
5. Reports the status and elapsed time for each step.

## Verify the Installation

After installation completes successfully, verify that the NVIDIA Run:ai pods are running:

```bash
# Control plane
kubectl get pods -n runai-backend

# Cluster
kubectl get pods -n runai
```

All pods should reach `Running` or `Completed` status.

Then log-in to the NVIDIA Run:ai UI at the control plane domain you configured (for example, `https://runai.mycorp.local`) and verify that the cluster appears with a **Connected** status. If the cluster does not show as **Connected**, see [Troubleshooting scenarios](../../../operations/clusters.md#troubleshooting-scenarios).

## Command Reference and Troubleshooting

See [Command Reference and Troubleshooting](../command-reference.md) for installer flags and common troubleshooting scenarios.
