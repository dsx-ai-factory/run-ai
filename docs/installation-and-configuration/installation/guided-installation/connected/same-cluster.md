---
---

# Same-Cluster Installation

This page describes how to install the NVIDIA Run:ai control plane and cluster together, on one Kubernetes cluster (`mode: full_installation`), over an internet-connected environment. To install the control plane and cluster on separate Kubernetes clusters instead, see [Multi-Cluster Installation](multi-cluster.md).

## Before You Start

Complete [Preparations](preparations.md) before continuing, including the system requirements for your Kubernetes cluster, any required registry credentials, and downloading the `runai-installer` binary. The installer deploys and configures NVIDIA Run:ai itself; it does not provision this underlying infrastructure for you.

## Create a Configuration File

<Note>

**Configuration Templates**

The steps below build the minimum configuration for a same-cluster installation on Kubernetes or OpenShift. Depending on your Kubernetes distribution, workload routing, networking setup, and the prerequisites and workload features you need, you may require additional configuration.

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

   - `controlPlane.domain` - The FQDN for the NVIDIA Run:ai UI and API. See [Domains and TLS Certificates](preparations.md#domains-and-tls-certificates).
   - `controlPlane.tls` - The certificate and private key for the control plane FQDN, as `certPath` and `keyPath`. See [Domains and TLS Certificates](preparations.md#domains-and-tls-certificates).
   - `controlPlane.version` - The control plane version to install. Omit it to install the latest supported version.
   - `controlPlane.customHelmValues` - Merges Helm values into the control plane chart, for settings not otherwise exposed by the installer. See [Advanced Control Plane Configurations](../../../configuration/control-plane-config.md) for the full set of keys.
     - `supportEntitled` - For NVIDIA Run:ai Support-Entitled Release, must be set to `true`.
     - `telemetryService.enabled` - For NVIDIA Run:ai Support-Entitled Release, must be set to `true` to enable telemetry.
     - `telemetryService.telemetryAuth` - Optional, Support-Entitled Release only. Associates NVIDIA Run:ai's telemetry with your NCA ID - by default, telemetry is not associated with your organization. See [Registry Credentials](preparations.md#registry-credentials) for how to generate the NGC Service API Key. Provide it one of two ways:
       - `token: <SERVICE_API_KEY>` - Sets the Service API Key directly; the chart creates and manages the Kubernetes secret for you. Stored in plaintext in the configuration file, so protect it accordingly, or use `existingSecret` instead.
       - `existingSecret: <SECRET_NAME>` / `secretKey: <KEY_NAME>` - References a secret you create yourself beforehand - the installer does not create or manage it.
     - Third-party services - Add the image pull secret, the external PostgreSQL connection, and the hardened images and settings of the control plane third-party services. See [Gov-Ready Third-Party Services](../../../system-requirements/gov-ready-third-party-services.md) for the values. <Badge intent="info" minimal outlined>Gov-ready</Badge>
     - `global.fipsMode` - Enables FIPS mode, so that the NVIDIA Run:ai services use only FIPS-validated cryptography. Set to `only`. <Badge intent="info" minimal outlined>Gov-ready</Badge>

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
       telemetryService.enabled: true # Support-Entitled Release only
       telemetryService.telemetryAuth.token: <SERVICE_API_KEY> # Optional
       # Gov-ready: add the image pull secret, external PostgreSQL, and hardened control plane values, see Gov-Ready Third-Party Services
       global.fipsMode: only # Gov-ready
   ```

4. Define the cluster:

   - `cluster.name` - The name of the NVIDIA Run:ai cluster.
   - `cluster.domain` - Omit it for a same-cluster installation - the cluster reuses `controlPlane.domain`, since both run on the same Kubernetes cluster.
   - `cluster.workloadRouting` - `host` (the default) or `path`. This determines which certificate you need for `cluster.tls` below.
   - `cluster.tls` - A certificate and private key for the cluster domain, as `certPath` and `keyPath`. See [Domains and TLS Certificates](preparations.md#domains-and-tls-certificates) for which certificate you need.
   - `cluster.version` - The cluster version to install. Omit it to install the latest supported version.
   - `cluster.customHelmValues` - Merges Helm values into the cluster chart, for settings not otherwise exposed by the installer. See [Advanced Cluster Configurations](../../../configuration/cluster-config.md) for the full set of keys.
     - Third-party services - Add the image pull secret and the hardened images of the cluster third-party services. See [Gov-Ready Third-Party Services](../../../system-requirements/gov-ready-third-party-services.md) for the values. <Badge intent="info" minimal outlined>Gov-ready</Badge>
     - `global.fipsMode` - Enables FIPS mode, so that the NVIDIA Run:ai services use only FIPS-validated cryptography. Set to `only`. <Badge intent="info" minimal outlined>Gov-ready</Badge>

   ```yaml
   cluster:
     enabled: true
     name: <CLUSTER_NAME>
     workloadRouting: host # Default; determines the certificate cluster.tls needs
     tls:
       certPath: /path/to/cluster-cert.pem
       keyPath: /path/to/cluster-key.pem
     version: "" # Omit to install the latest supported version
     customHelmValues:
       # Gov-ready: add the image pull secret and hardened cluster values, see Gov-Ready Third-Party Services
       global.fipsMode: only # Gov-ready
   ```

5. Enable the prerequisites:

   - `prerequisites.gpuOperator` - Required for every installation. See [Prerequisites and Features](../prerequisites-and-features.md#prerequisites) for more details.
     - `mode` - Set to `install`.
     - `version` - Omit to install the latest version, or see [Support Matrix](../../../support-matrix.md) for the supported range if you want to set it explicitly. GPU fractions (memory fractions, dynamic memory fractions, or compute modes) require 26.7.1 or later, installed automatically when omitted.
     - `customHelmValues` - Merges Helm values into the GPU Operator's own Helm chart. Set `driver.version` to r615 or later if you plan to use GPU fractions, since GPU Operator 26.7.1 does not default to this driver version. Set `cdi.nriPluginEnabled` to `false`, since GPU memory fractions do not function while the NRI plugin is enabled. See [GPU Memory Fractions](../../../../platform-management/runai-scheduler/resource-optimization/memory-fractions.md#prerequisites).
     - GPU Operator values - Add the Helm values required for a Gov-ready GPU Operator to `customHelmValues`. See [Install the Gov-Ready GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/install-gpu-operator-gov-ready.html) for the values. <Badge intent="info" minimal outlined>Gov-ready</Badge>
   - `prerequisites.prometheus` - Required for every installation. See [Prerequisites and Features](../prerequisites-and-features.md#prerequisites) for more details.
     - `mode` - Set to `install`.
     - `version` - Omit to install the latest version, or see [Support Matrix](../../../support-matrix.md) for the supported range.
     - `customHelmValues` - Merges Helm values into Prometheus's own Helm chart.

   ```yaml
   prerequisites:
     gpuOperator:
       mode: install
       version: ""
       customHelmValues: {} # Gov-ready: add the GPU Operator Helm values
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

6. Configure external access to the control plane and cluster:

   - `externalConnection` - Configures how traffic reaches the NVIDIA Run:ai control plane UI/API and cluster endpoint from outside the cluster, through a Kubernetes Ingress Controller or Gateway API implementation already running on your cluster. This is separate from `cluster.workloadRouting`, which controls how the workloads you submit are exposed. See [Networking Setup](../../../system-requirements/networking-setup.md).

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
    # Gov-ready: add the image pull secret, external PostgreSQL, and hardened control plane values, see Gov-Ready Third-Party Services
    global.fipsMode: only # Gov-ready

cluster:
  enabled: true
  name: <CLUSTER_NAME>
  workloadRouting: host
  tls:
    certPath: /path/to/cluster-cert.pem
    keyPath: /path/to/cluster-key.pem
  version: ""
  customHelmValues:
    # Gov-ready: add the image pull secret and hardened cluster values, see Gov-Ready Third-Party Services
    global.fipsMode: only # Gov-ready

prerequisites:
  gpuOperator:
    mode: install
    version: ""
    customHelmValues: {} # Gov-ready: add the GPU Operator Helm values
  prometheus:
    mode: install
    version: ""

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
   - `install.connectivity` - Set to `connected`, for an installation with outbound internet access.
   - `install.kubernetesDistribution` - Set to `openshift`.

   ```yaml
   install:
     mode: full_installation
     connectivity: connected
     kubernetesDistribution: openshift
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

   - `controlPlane.domain` - The FQDN for the NVIDIA Run:ai UI and API. See [Domains and TLS Certificates](preparations.md#domains-and-tls-certificates).
   - `controlPlane.version` - The control plane version to install. Omit it to install the latest supported version.
   - `controlPlane.customHelmValues` - Merges Helm values into the control plane chart, for settings not otherwise exposed by the installer. See [Advanced Control Plane Configurations](../../../configuration/control-plane-config.md) for the full set of keys.
     - `supportEntitled` - For NVIDIA Run:ai Support-Entitled Release, must be set to `true`.
     - `telemetryService.enabled` - For NVIDIA Run:ai Support-Entitled Release, must be set to `true` to enable telemetry.
     - `telemetryService.telemetryAuth` - Optional, Support-Entitled Release only. Associates NVIDIA Run:ai's telemetry with your NCA ID - by default, telemetry is not associated with your organization. See [Registry Credentials](preparations.md#registry-credentials) for how to generate the NGC Service API Key. Provide it one of two ways:
       - `token: <SERVICE_API_KEY>` - Sets the Service API Key directly; the chart creates and manages the Kubernetes secret for you. Stored in plaintext in the configuration file, so protect it accordingly, or use `existingSecret` instead.
       - `existingSecret: <SECRET_NAME>` / `secretKey: <KEY_NAME>` - References a secret you create yourself beforehand - the installer does not create or manage it.
     - Third-party services - Add the image pull secret, the external PostgreSQL connection, and the hardened images and settings of the control plane third-party services. See [Gov-Ready Third-Party Services](../../../system-requirements/gov-ready-third-party-services.md) for the values. <Badge intent="info" minimal outlined>Gov-ready</Badge>
     - `global.fipsMode` - Enables FIPS mode, so that the NVIDIA Run:ai services use only FIPS-validated cryptography. Set to `only`. <Badge intent="info" minimal outlined>Gov-ready</Badge>

   ```yaml
   controlPlane:
     enabled: true
     domain: <DOMAIN>
     version: "" # Omit to install the latest supported version
     customHelmValues:
       supportEntitled: true # Support-Entitled Release only
       telemetryService.enabled: true # Support-Entitled Release only
       telemetryService.telemetryAuth.token: <SERVICE_API_KEY> # Optional
       # Gov-ready: add the image pull secret, external PostgreSQL, and hardened control plane values, see Gov-Ready Third-Party Services
       global.fipsMode: only # Gov-ready
   ```

4. Define the cluster:

   - `cluster.name` - The name of the NVIDIA Run:ai cluster.
   - `cluster.domain` - Omit it for a same-cluster installation - the cluster reuses `controlPlane.domain`, since both run on the same OpenShift cluster.
   - `cluster.version` - The cluster version to install. Omit it to install the latest supported version.
   - `cluster.customHelmValues` - Merges Helm values into the cluster chart, for settings not otherwise exposed by the installer. See [Advanced Cluster Configurations](../../../configuration/cluster-config.md) for the full set of keys.
     - Third-party services - Add the image pull secret and the hardened images of the cluster third-party services. See [Gov-Ready Third-Party Services](../../../system-requirements/gov-ready-third-party-services.md) for the values. <Badge intent="info" minimal outlined>Gov-ready</Badge>
     - `global.fipsMode` - Enables FIPS mode, so that the NVIDIA Run:ai services use only FIPS-validated cryptography. Set to `only`. <Badge intent="info" minimal outlined>Gov-ready</Badge>

   ```yaml
   cluster:
     enabled: true
     name: <CLUSTER_NAME>
     version: "" # Omit to install the latest supported version
     customHelmValues:
       # Gov-ready: add the image pull secret and hardened cluster values, see Gov-Ready Third-Party Services
       global.fipsMode: only # Gov-ready
   ```

5. Enable the prerequisites:

   - `prerequisites.gpuOperator` - Required for every installation. See [Prerequisites and Features](../prerequisites-and-features.md#prerequisites) for more details.
     - `mode` - Set to `install`. The installer installs the GPU Operator through OperatorHub (OLM).
     - `version` - Omit to install the latest version, or see [Support Matrix](../../../support-matrix.md) for the supported range if you want to set it explicitly. GPU fractions (memory fractions, dynamic memory fractions, or compute modes) require 26.7.1 or later, installed automatically when omitted.
     - `openshift.customResource` - Merges values into the GPU Operator's `ClusterPolicy`. `customHelmValues` is not supported on OpenShift. Set `driver.version` to r615 or later if you plan to use GPU fractions, since GPU Operator 26.7.1 does not default to this driver version. See [GPU Memory Fractions](../../../../platform-management/runai-scheduler/resource-optimization/memory-fractions.md#prerequisites) and [Per-Component OpenShift Fields](../configuration-reference.md#per-component-openshift-fields).
     - GPU Operator values - Add the values required for a Gov-ready GPU Operator to `openshift.customResource`. See [Install the Gov-Ready GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/install-gpu-operator-gov-ready.html) for the values. <Badge intent="info" minimal outlined>Gov-ready</Badge>
   - `prerequisites.prometheus` - See [Prerequisites and Features](../prerequisites-and-features.md#prerequisites) for more details.
     - `mode` - Set to `skip`, to use the cluster's built-in monitoring stack. `install` is rejected on OpenShift.

   ```yaml
   prerequisites:
     gpuOperator:
       mode: install
       version: ""
       openshift:
         customResource: # Gov-ready: add the GPU Operator values
           driver:
             version: <DRIVER_VERSION>
     prometheus:
       mode: skip
   ```

   <Note>

   **Features**

   The minimum configuration does not include features such as Knative, Kubeflow, MPI Operator, LWS, NIM Operator, Dynamo, or Grove. These are skipped by default.

   If your workloads require any of these features, add the corresponding `features` configuration. See [Prerequisites and Features](../prerequisites-and-features.md#features).

   Some features need OpenShift-specific configuration. For example, Knative, required for standard inference, installs through the OpenShift Serverless operator and requires `cluster.inference.wildcardDomain`, but no wildcard TLS certificate. Features that install through OperatorHub, such as Knative and LWS, are customized through `openshift.customResource` rather than `customHelmValues`. See [Per-Component OpenShift Fields](../configuration-reference.md#per-component-openshift-fields).

   </Note>

The following example combines these settings to install the NVIDIA Run:ai control plane and cluster on the same OpenShift cluster:

```yaml
install:
  mode: full_installation
  connectivity: connected
  kubernetesDistribution: openshift

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

controlPlane:
  enabled: true
  domain: <DOMAIN>
  version: ""
  customHelmValues:
    supportEntitled: true # Support-Entitled Release only
    # Gov-ready: add the image pull secret, external PostgreSQL, and hardened control plane values, see Gov-Ready Third-Party Services
    global.fipsMode: only # Gov-ready

cluster:
  enabled: true
  name: <CLUSTER_NAME>
  version: ""
  customHelmValues:
    # Gov-ready: add the image pull secret and hardened cluster values, see Gov-Ready Third-Party Services
    global.fipsMode: only # Gov-ready

prerequisites:
  gpuOperator:
    mode: install
    version: ""
    openshift:
      customResource: # Gov-ready: add the GPU Operator values
        driver:
          version: <DRIVER_VERSION>
  prometheus:
    mode: skip
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
