---
---

# Same-Cluster Installation

This page describes how to install the NVIDIA Run:ai control plane and cluster together, on one Kubernetes cluster (`mode: full_installation`), from an air-gapped bundle. To install the control plane and cluster on separate Kubernetes clusters instead, see [Multi-Cluster Installation](multi-cluster.md).

## Before You Start

Complete [Preparations](preparations.md) before continuing, including the system requirements for your Kubernetes cluster, registry credentials, downloading the `runai-installer` binary, and building your [Air-Gapped Bundle](preparations.md#air-gapped-bundle). The installer deploys and configures NVIDIA Run:ai itself; it does not provision this underlying infrastructure for you.

<Note>

**Note**

The installer does not install prerequisites or features in air-gapped environments. Install each required component yourself beforehand, on each target Kubernetes cluster, using that component's air-gapped installation instructions. See [Prerequisites and Features](preparations.md#prerequisites-and-features).

Some features require additional networking configuration. For example, externally exposed standard inference with Knative requires an additional inference wildcard FQDN and TLS certificate. See [Networking Setup](../../../system-requirements/networking-setup.md).

</Note>

## Create a Configuration File

<Note>

**Configuration Templates**

The steps below build the minimum configuration for a same-cluster installation on Kubernetes or OpenShift, from an air-gapped bundle. Depending on your Kubernetes distribution, workload routing, networking setup, and the prerequisites and workload features you need, you may require additional configuration.

* Once you understand the minimum configuration, see [Configuration Templates](templates.md) - if your use case matches one, use it as your starting point instead of building the configuration from scratch.
* See [Prerequisites and Features](../prerequisites-and-features.md) to understand the components required, depending on the workload types you plan to enable.
* See [Configuration Reference](../configuration-reference.md) for the full set of supported fields and descriptions, to build a configuration from scratch.

</Note>

Create a `config.yaml` file describing your installation. Use the domains, TLS certificates, and networking values you prepared in [Networking Setup](../../../system-requirements/networking-setup.md) when completing the configuration below.

Expand the section for your platform:

<Accordion title="Kubernetes">

For a same-cluster installation on Kubernetes, configure:

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
     host: <REGISTRY_HOST>
     username: <REGISTRY_USERNAME>
     passwordPath: /path/to/registry-password
   ```

3. Configure the air-gapped bundle:

   - `airgap` - The same `workingDir` and `valuesDir` you used with `airgap load`, plus `skipPrepare: true` - the bundle is already loaded, so `install` skips straight to the chart install. See [Air-Gapped Bundle](preparations.md#air-gapped-bundle) and [`airgap`](../configuration-reference.md#air-gapped-installation-configuration) in the Configuration Reference.

   ```yaml
   airgap:
     workingDir: /data/runai-airgap
     valuesDir: /data/runai-airgap
     skipPrepare: true
   ```

4. Define the control plane:

   - `controlPlane.domain` - The FQDN for the NVIDIA Run:ai UI and API. See [Domains and TLS Certificates](preparations.md#domains-and-tls-certificates).
   - `controlPlane.tls` - The certificate and private key for the control plane FQDN, as `certPath` and `keyPath`. See [Domains and TLS Certificates](preparations.md#domains-and-tls-certificates).
   - `controlPlane.customHelmValues` - Merges Helm values into the control plane chart, for settings not otherwise exposed by the installer. See [Advanced Control Plane Configurations](../../../configuration/control-plane-config.md) for the full set of keys.
     - `supportEntitled` - For NVIDIA Run:ai Support-Entitled Release, must be set to `true`.
     - Third-party services - Add the external PostgreSQL connection and the Keycloak and NATS settings of the control plane third-party services. Do not set image values, because `airgap load` generates them. See [Gov-Ready Third-Party Services](../../../system-requirements/gov-ready-third-party-services.md) for the values. <Badge intent="info" minimal outlined>Gov-ready</Badge>
     - `global.fipsMode` - Enables FIPS mode, so that the NVIDIA Run:ai services use only FIPS-validated cryptography. Set to `only`. <Badge intent="info" minimal outlined>Gov-ready</Badge>

   ```yaml
   controlPlane:
     enabled: true
     domain: <DOMAIN>
     tls:
       certPath: /path/to/cp-cert.pem
       keyPath: /path/to/cp-key.pem
     customHelmValues:
       supportEntitled: true # Support-Entitled Release only
       # Gov-ready: add the external PostgreSQL, Keycloak, and NATS values, see Gov-Ready Third-Party Services
       global.fipsMode: only # Gov-ready
   ```

5. Define the cluster:

   - `cluster.name` - The name of the NVIDIA Run:ai cluster.
   - `cluster.domain` - Omit it for a same-cluster installation - the cluster reuses `controlPlane.domain`, since both run on the same Kubernetes cluster.
   - `cluster.workloadRouting` - `host` (the default) or `path`. This determines which certificate you need for `cluster.tls` below.
   - `cluster.tls` - A certificate and private key for the cluster domain, as `certPath` and `keyPath`. See [Domains and TLS Certificates](preparations.md#domains-and-tls-certificates) for which certificate you need.
   - `cluster.customHelmValues` - Merges Helm values into the cluster chart, for settings not otherwise exposed by the installer. See [Advanced Cluster Configurations](../../../configuration/cluster-config.md) for the full set of keys.
     - `global.fipsMode` - Enables FIPS mode, so that the NVIDIA Run:ai services use only FIPS-validated cryptography. Set to `only`. <Badge intent="info" minimal outlined>Gov-ready</Badge>

   ```yaml
   cluster:
     enabled: true
     name: <CLUSTER_NAME>
     workloadRouting: host # Default; determines the certificate cluster.tls needs
     tls:
       certPath: /path/to/cluster-cert.pem
       keyPath: /path/to/cluster-key.pem
     customHelmValues:
       global.fipsMode: only # Gov-ready
   ```

6. Skip the prerequisites:

   - `prerequisites.skip` - Set to `true`. The air-gapped bundle contains only the NVIDIA Run:ai charts, so the installer cannot install prerequisites such as the GPU Operator and Prometheus. Make sure the prerequisites your workloads need are already installed on the cluster. See [Prerequisites and Features](../prerequisites-and-features.md#prerequisites).

   ```yaml
   prerequisites:
     skip: true
   ```

7. Configure external access to the control plane and cluster:

   - `externalConnection` - Configures how traffic reaches the NVIDIA Run:ai control plane UI/API and cluster endpoint from outside the cluster, through a Kubernetes Ingress Controller or Gateway API implementation already running on your cluster. This is separate from `cluster.workloadRouting`, which controls how the workloads you submit are exposed. See [Networking Setup](../../../system-requirements/networking-setup.md).

   ```yaml
   externalConnection:
     type: ingress
     ingress:
       className: <YOUR_INGRESS_CLASS>
   ```

The following example combines these settings to install the NVIDIA Run:ai control plane and cluster on the same Kubernetes cluster from an already-loaded air-gapped bundle:

```yaml
install:
  mode: full_installation
  connectivity: air_gapped

registry:
  host: <REGISTRY_HOST>
  username: <REGISTRY_USERNAME>
  passwordPath: /path/to/registry-password

airgap:
  workingDir: /data/runai-airgap
  valuesDir: /data/runai-airgap
  skipPrepare: true

controlPlane:
  enabled: true
  domain: <DOMAIN>
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues:
    supportEntitled: true # Support-Entitled Release only
    # Gov-ready: add the external PostgreSQL, Keycloak, and NATS values, see Gov-Ready Third-Party Services
    global.fipsMode: only # Gov-ready

cluster:
  enabled: true
  name: <CLUSTER_NAME>
  workloadRouting: host
  tls:
    certPath: /path/to/cluster-cert.pem
    keyPath: /path/to/cluster-key.pem
  customHelmValues:
    global.fipsMode: only # Gov-ready

prerequisites:
  skip: true

externalConnection:
  type: ingress
  ingress:
    className: <YOUR_INGRESS_CLASS>
```

</Accordion>

<Accordion title="OpenShift">

For a same-cluster installation on OpenShift, configure:

1. Define the installation:

   - `install.mode` - Set to `full_installation` to install the control plane and cluster on the same OpenShift cluster.
   - `install.connectivity` - Set to `air_gapped`, for an installation with no outbound internet access.
   - `install.kubernetesDistribution` - Set to `openshift`.

   ```yaml
   install:
     mode: full_installation
     connectivity: air_gapped
     kubernetesDistribution: openshift
   ```

2. Configure the registry:

   - `registry.host` - The hostname of your own internal mirror registry (populated by loading the air-gapped bundle), not NGC.
   - `registry.username` - The username to authenticate to that registry.
   - `registry.passwordPath` - The path to a file containing the password for that user. See [Registry Credentials](preparations.md#registry-credentials).

   ```yaml
   registry:
     host: <REGISTRY_HOST>
     username: <REGISTRY_USERNAME>
     passwordPath: /path/to/registry-password
   ```

3. Configure the air-gapped bundle:

   - `airgap` - The same `workingDir` and `valuesDir` you used with `airgap load`, plus `skipPrepare: true` - the bundle is already loaded, so `install` skips straight to the chart install. See [Air-Gapped Bundle](preparations.md#air-gapped-bundle) and [`airgap`](../configuration-reference.md#air-gapped-installation-configuration) in the Configuration Reference.

   ```yaml
   airgap:
     workingDir: /data/runai-airgap
     valuesDir: /data/runai-airgap
     skipPrepare: true
   ```

4. Define the control plane:

   - `controlPlane.domain` - The FQDN for the NVIDIA Run:ai UI and API. See [Domains and TLS Certificates](preparations.md#domains-and-tls-certificates).
   - `controlPlane.customHelmValues` - Merges Helm values into the control plane chart, for settings not otherwise exposed by the installer. See [Advanced Control Plane Configurations](../../../configuration/control-plane-config.md) for the full set of keys.
     - `supportEntitled` - For NVIDIA Run:ai Support-Entitled Release, must be set to `true`.
     - Third-party services - Add the external PostgreSQL connection and the Keycloak and NATS settings of the control plane third-party services. Do not set image values, because `airgap load` generates them. See [Gov-Ready Third-Party Services](../../../system-requirements/gov-ready-third-party-services.md) for the values. <Badge intent="info" minimal outlined>Gov-ready</Badge>
     - `global.fipsMode` - Enables FIPS mode, so that the NVIDIA Run:ai services use only FIPS-validated cryptography. Set to `only`. <Badge intent="info" minimal outlined>Gov-ready</Badge>

   ```yaml
   controlPlane:
     enabled: true
     domain: <DOMAIN>
     customHelmValues:
       supportEntitled: true # Support-Entitled Release only
       # Gov-ready: add the external PostgreSQL, Keycloak, and NATS values, see Gov-Ready Third-Party Services
       global.fipsMode: only # Gov-ready
   ```

5. Define the cluster:

   - `cluster.name` - The name of the NVIDIA Run:ai cluster.
   - `cluster.domain` - Omit it for a same-cluster installation - the cluster reuses `controlPlane.domain`, since both run on the same OpenShift cluster.
   - `cluster.customHelmValues` - Merges Helm values into the cluster chart, for settings not otherwise exposed by the installer. See [Advanced Cluster Configurations](../../../configuration/cluster-config.md) for the full set of keys.
     - `global.fipsMode` - Enables FIPS mode, so that the NVIDIA Run:ai services use only FIPS-validated cryptography. Set to `only`. <Badge intent="info" minimal outlined>Gov-ready</Badge>

   ```yaml
   cluster:
     enabled: true
     name: <CLUSTER_NAME>
     customHelmValues:
       global.fipsMode: only # Gov-ready
   ```

6. Skip the prerequisites:

   - `prerequisites.skip` - Set to `true`. The air-gapped bundle contains only the NVIDIA Run:ai charts, so the installer cannot install prerequisites such as the GPU Operator and Prometheus. Make sure the prerequisites your workloads need are already installed on the cluster. See [Prerequisites and Features](../prerequisites-and-features.md#prerequisites).

   ```yaml
   prerequisites:
     skip: true
   ```

The following example combines these settings to install the NVIDIA Run:ai control plane and cluster on the same OpenShift cluster from an already-loaded air-gapped bundle:

```yaml
install:
  mode: full_installation
  connectivity: air_gapped
  kubernetesDistribution: openshift

registry:
  host: <REGISTRY_HOST>
  username: <REGISTRY_USERNAME>
  passwordPath: /path/to/registry-password

airgap:
  workingDir: /data/runai-airgap
  valuesDir: /data/runai-airgap
  skipPrepare: true

controlPlane:
  enabled: true
  domain: <DOMAIN>
  customHelmValues:
    supportEntitled: true # Support-Entitled Release only
    # Gov-ready: add the external PostgreSQL, Keycloak, and NATS values, see Gov-Ready Third-Party Services
    global.fipsMode: only # Gov-ready

cluster:
  enabled: true
  name: <CLUSTER_NAME>
  customHelmValues:
    global.fipsMode: only # Gov-ready

prerequisites:
  skip: true
```

</Accordion>

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
