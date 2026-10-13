---
---

# Configuration Reference

The `runai-installer` CLI reads a configuration YAML file specified with the `--config` flag. This reference describes every supported configuration field.

Fields marked **Required** must be specified. All other fields are optional.

## Configuration Overview

**Core Installation Settings**

- [Installation Settings](#installation-settings) - Installation mode (full, control-plane-only, or cluster-only) and network connectivity model (connected or air-gapped); determines which other sections apply.
- [Registry Configuration](#registry-configuration) - Credentials for pulling NVIDIA Run:ai images and Helm charts: token authentication (NGC), basic authentication (Harbor and other OCI registries), or no authentication.
- [Default TLS Configuration](#default-tls-configuration) - Default TLS certificates for the control plane and cluster: certificate, key, CA chain, or a self-signed option for non-production use. Component-level `tls` blocks override these defaults.

**NVIDIA Run:ai Components**

- [Control Plane Configuration](#control-plane-configuration) - Control plane deployment: whether to install it, its domain, version, TLS, and custom Helm values.
- [Cluster Configuration](#cluster-configuration) - Common cluster settings plus multi-cluster registration fields.

**Prerequisites and Features**

- [Prerequisite Configuration](#prerequisite-configuration) - Infrastructure components the installer can set up before NVIDIA Run:ai: GPU Operator, Prometheus, and Network Operator, each independently set to install or skip.
- [Feature Configuration](#feature-configuration) - Components that extend NVIDIA Run:ai for specific workload types: Knative for standard inference, LWS for distributed inference, Kubeflow and MPI Operator for distributed training, plus NIM Operator, Grove, and Dynamo.
- [Common Prerequisite and Feature Fields](#common-prerequisite-and-feature-fields) - Fields shared by every prerequisite and feature, such as `mode`, `version`, and `customHelmValues`.

**Networking and Nodes**

- [External Connection](#external-connection) - How external traffic reaches NVIDIA Run:ai services: Kubernetes Ingress or Gateway API. Omitted entirely on OpenShift, where Routes are configured automatically.
- [System Nodes](#system-nodes) - Dedicating specific nodes to NVIDIA Run:ai system components, with a preferred or required scheduling mode and automatic node labeling.

**OpenShift**

- [OpenShift Configuration](#openshift-configuration) - Installer-wide OpenShift settings, such as a mirrored OperatorHub catalog source for air-gapped OpenShift environments.
- [Per-Component OpenShift Fields](#per-component-openshift-fields) - OpenShift-specific fields available under individual prerequisites and features.

**Air-Gapped Installation**

- [Air-Gapped Installation Configuration](#air-gapped-installation-configuration) - Bundle path, working directory, and extraction behavior for an air-gapped installation. Required when `install.connectivity` is `air_gapped`.
- [Pack an Air-Gapped Bundle](#pack-an-air-gapped-bundle) - Configuration for the separate `airgap pack` command, which builds an air-gapped bundle from a release's artifacts.
- [Load an Air-Gapped Bundle](#load-an-air-gapped-bundle) - Configuration for the separate `airgap load` command, which pushes a bundle's images to your registry and writes the Helm values `install` consumes.

---

## Core Installation Settings

### Installation Settings

Controls the installation type and network connectivity model. `mode` is based on your [deployment model](../../system-requirements/deployment-model.md) - same-cluster or multi-cluster:

- `full_installation` - Same-cluster deployments.
- `runai_control_plane_only` and `runai_cluster_only` - Multi-cluster deployments, one per component.

```yaml
install:
  mode: full_installation
  connectivity: connected
  kubernetesDistribution: ""
```

| Field | Type | Required | Description |
|---|---|---|---|
| `mode` | string | No | <p>Installation mode, based on your deployment model:</p><ul><li><code>full_installation</code> - Same-cluster deployment: the NVIDIA Run:ai control plane and cluster together. Default when omitted.</li><li><code>runai_control_plane_only</code> - Multi-cluster deployment: the NVIDIA Run:ai control plane.</li><li><code>runai_cluster_only</code> - Multi-cluster deployment: the NVIDIA Run:ai cluster.</li></ul> |
| `connectivity` | string | Yes | <p>Network connectivity model:</p><ul><li><code>connected</code> - Internet-connected environments.</li><li><code>air_gapped</code> - Air-gapped environments.</li></ul> |
| `kubernetesDistribution` | string | No | Kubernetes distribution. Set to `openshift` when installing on Red Hat OpenShift - this is the only recognized non-empty value. Omit for every other supported distribution (Upstream Kubernetes, EKS, GKE, AKS, OKE, or RKE2). |

---

### Registry Configuration

Configures the registry credentials used to pull NVIDIA Run:ai images and Helm charts. Which method you need depends on your installation:

- **Open Source installation** (`type: ngc`, no credentials) - Pulls NVIDIA Run:ai's publicly accessible images from NGC without credentials. `host` can be omitted; it defaults to NGC's registry (`nvcr.io`).
- **Token authentication** (`type` with `token` or `tokenPath`) - For an NVIDIA Run:ai Support-Entitled Release installation, pulling from NVIDIA's own NGC registry (`type: ngc`). For any other registry secured by a base64-encoded `.dockerconfigjson` pull secret instead of NGC credentials, use `type: dockerconfig`.
- **Basic authentication** (`username` with `password` or `passwordPath`) - For a private registry you control, such as Harbor or another OCI-compliant registry - for example, an internal mirror populated for an air-gapped install, which always uses basic authentication rather than a token.
- **No authentication, custom registry** (`host` only, no `type`) - For any other unauthenticated registry. `host` is required in this case, since there's no default to fall back to.

These methods are mutually exclusive. The `--ngc-token-path` and `--registry-token-path` command-line flags override the corresponding value in this section - use `--ngc-token-path` for NGC token authentication, and `--registry-token-path` for a `.dockerconfigjson` token.

Token authentication (NGC):

```yaml
registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only
```

Token authentication (dockerconfig):

```yaml
registry:
  type: dockerconfig
  tokenPath: /path/to/dockerconfigjson-token
```

Basic authentication:

```yaml
registry:
  host: <REGISTRY_HOST>
  username: runai
  passwordPath: /path/to/password
  caChainPath: /path/to/ca-chain.pem
```

No authentication:

```yaml
registry:
  host: <REGISTRY_HOST>
```

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | Conditional | Registry type. Required when using token authentication, and for an Open Source installation's NGC pull. Supported values: `ngc` (NVIDIA GPU Cloud) and `dockerconfig` (a base64-encoded `.dockerconfigjson` pull secret for any other registry). |
| `host` | string | Conditional | Registry hostname. Optional when a token is set, or when `type` is `ngc` (even without a token - falls back to `nvcr.io`). Required for basic authentication, air-gapped installs, and any other unauthenticated registry. |
| `token` | string | No | Registry token specified inline. Mutually exclusive with `tokenPath`. For production environments, use `tokenPath` to avoid storing secrets in the configuration file. |
| `tokenPath` | string | No | Path to a file containing the registry token. Mutually exclusive with `token`. |
| `username` | string | No | Username for basic authentication. Requires `password` or `passwordPath`. |
| `password` | string | No | Password specified inline. Mutually exclusive with `passwordPath`. |
| `passwordPath` | string | No | Path to a file containing the password. Mutually exclusive with `password`. |
| `caChainPath` | string | No | Path to the CA certificate chain. Required when the registry uses a private CA. |
| `chartRepository` | string | No | Overrides the default NGC Helm chart repository URL. |
| `insecure` | bool | No | Connects to the registry over HTTP instead of HTTPS. |
| `insecureSkipVerify` | bool | No | Connects over HTTPS without verifying the server's TLS certificate. |

---

### Default TLS Configuration

Configures the default TLS settings used by the control plane and cluster components. A component that defines its own `tls` block (`controlPlane.tls`, `cluster.tls`, or `cluster.inference.wildcardTls`) uses that block instead of this one; a component that omits `tls` falls back to these top-level settings.

```yaml
tls:
  certPath: /path/to/cert.pem
  keyPath: /path/to/key.pem
  caChainPath: /path/to/ca-chain.pem
  generateSelfSigned: false
```

| Field | Type | Required | Description |
|---|---|---|---|
| `certPath` | string | No | Absolute path to the TLS certificate file in PEM format. |
| `keyPath` | string | No | Absolute path to the TLS private key file in PEM format. |
| `caChainPath` | string | No | Absolute path to the CA certificate chain. Required when the certificate is signed by a private CA. |
| `generateSelfSigned` | bool | No | Generates a self-signed TLS certificate automatically. Use only for non-production environments. |

---

## NVIDIA Run:ai Components

### Control Plane Configuration

Configures the NVIDIA Run:ai control plane. The NVIDIA Run:ai control plane is installed once, regardless of deployment model - in a same-cluster deployment it runs alongside the NVIDIA Run:ai cluster component; in a multi-cluster deployment it is installed on its own, separately from each NVIDIA Run:ai cluster.

```yaml
controlPlane:
  enabled: true
  version: ""
  multiTenant: false
  domain: <DOMAIN>
  tls:
    certPath: /path/to/cert.pem
    keyPath: /path/to/key.pem
  customHelmValues: {}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `enabled` | bool | Yes | Whether to install the control plane. |
| `version` | string | No | Control plane version. Omit to install the latest supported version. |
| `multiTenant` | bool | No | Enables multi-tenant mode. |
| `domain` | string | Yes | Fully qualified domain name (FQDN) where the NVIDIA Run:ai UI and API are available (for example, `runai.mycorp.local`). |
| `tls` | [`tls`](#default-tls-configuration) | No | Control plane TLS settings. Overrides the top-level `tls` configuration. |
| `customHelmValues` | map | No | Custom Helm values merged into the control plane chart. Use this field to configure settings that are not exposed directly by the installer - see [Advanced Control Plane Configurations](../../configuration/control-plane-config.md) for the full set of keys. |

---

### Cluster Configuration

Configures the NVIDIA Run:ai cluster component. In a multi-cluster deployment, this section is applied once per NVIDIA Run:ai cluster.

**Multi-cluster only:** `domain`, `controlPlaneUrl`, and the cluster registration fields (`uid`, `bootstrapTokenPath`, `clientSecretPath`, or `controlPlaneCredentials`) are used when the NVIDIA Run:ai cluster component is installed on a different Kubernetes cluster than the NVIDIA Run:ai control plane. Omit them for same-cluster installations unless otherwise noted.

```yaml
cluster:
  enabled: true
  version: ""
  name: <CLUSTER_NAME>
  domain: ""
  workloadRouting: path
  tls:
    certPath: /path/to/cert.pem
    keyPath: /path/to/key.pem
  inference:
    wildcardDomain: "*.<INFERENCE_WILDCARD_DOMAIN>"
    wildcardTls:
      certPath: /path/to/wildcard-cert.pem
      keyPath: /path/to/wildcard-key.pem
  controlPlaneUrl: ""
  uid: "" # alternative to controlPlaneCredentials, together with bootstrapTokenPath or clientSecretPath
  bootstrapTokenPath: ""
  controlPlaneCredentials:
    tokenPath: ""
  customHelmValues: {}
```

#### Common Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `enabled` | bool | Yes | Whether to install the cluster component. |
| `version` | string | No | Cluster component version. Omit to install the latest supported version. |
| `name` | string | Yes | Cluster name as it appears in the NVIDIA Run:ai control plane. Must follow Kubernetes naming conventions (lowercase letters, numbers, and hyphens; maximum 63 characters). |
| `workloadRouting` | string | No | Workload routing method. Supported values: `host` (subdomain-based routing, default) and `path` (path-based routing). |
| `tls` | [`tls`](#default-tls-configuration) | No | Cluster TLS settings. Overrides the top-level `tls` configuration. This one certificate secures both the cluster's own domain and every workspace, training, and distributed inference workload exposed on a subdomain beneath it. With `workloadRouting: host` (default), it must cover both the exact cluster domain and its wildcard as Subject Alternative Names (for example, both `runai.mycorp.local` and `*.runai.mycorp.local`). With `workloadRouting: path`, a certificate for the exact domain alone is sufficient. The matching wildcard DNS record is not a config field - create it yourself in your DNS provider. |
| `inference.wildcardDomain` | string | No | Wildcard FQDN for inference endpoints (for example, `*.runai-inference.mycorp.local`). Required when Knative is installed. |
| `inference.wildcardTls` | [`tls`](#default-tls-configuration) | No | TLS settings for the wildcard inference domain. |
| `customHelmValues` | map | No | Custom Helm values merged into the cluster chart. Use this field to configure settings that are not exposed directly by the installer - see [Advanced Cluster Configurations](../../configuration/cluster-config.md) for the full set of keys. |

#### Multi-Cluster Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `domain` | string | No | Fully qualified domain name (FQDN) for the NVIDIA Run:ai cluster component when it is installed on a different Kubernetes cluster than the NVIDIA Run:ai control plane. Omit for a same-cluster (`mode: full_installation`) installation. |
| `controlPlaneUrl` | string | No | URL of the NVIDIA Run:ai control plane, reachable over HTTPS. Required when the NVIDIA Run:ai cluster component is installed on a different Kubernetes cluster than the NVIDIA Run:ai control plane (`mode: runai_cluster_only`), so the installer can reach the control plane to register the cluster. |
| `uid` | string | No | Unique identifier of an existing cluster in the control plane, obtained from the control plane UI or API. Used together with exactly one of `bootstrapTokenPath` or `clientSecretPath` to register with an existing cluster, as an alternative to `controlPlaneCredentials`. |
| `bootstrapTokenPath` | string | No | Path to a file containing the bootstrap token for the existing cluster identified by `uid`, obtained from the control plane UI or API. Mutually exclusive with `clientSecretPath`. |
| `clientSecretPath` | string | No | Path to a file containing the client secret for the existing cluster identified by `uid`, obtained from the control plane UI or API. Mutually exclusive with `bootstrapTokenPath`. |
| `controlPlaneCredentials` | object | No | Control plane login used to create the cluster in the control plane if it does not already exist, or reuse it if it does, as an alternative to `uid` plus `bootstrapTokenPath`/`clientSecretPath`. Accepts either a username and password (`username`, and `passwordPath` or inline `password`) or a token (`tokenPath` or inline `token`), but not both. |

---

## Prerequisites and Features

### Prerequisite Configuration

Configures the infrastructure components that the installer can set up before deploying NVIDIA Run:ai. Each of `gpuOperator`, `prometheus`, and `networkOperator` accepts the same set of fields - see [Common Prerequisite and Feature Fields](#common-prerequisite-and-feature-fields), below. For what each component does and when you need it, see [Prerequisites](prerequisites-and-features.md#prerequisites) in Prerequisites and Features.

```yaml
prerequisites:
  skip: false
  gpuOperator:
    mode: install
  prometheus:
    mode: skip
  networkOperator:
    mode: skip
```

| Field | Type | Required | Description |
|---|---|---|---|
| `skip` | bool | No | Skips installation of all prerequisites. Equivalent to setting `mode: skip` for every prerequisite. |
| `gpuOperator` | object | No | NVIDIA GPU Operator. Provisions and manages GPUs on worker nodes; required for the NVIDIA Run:ai cluster to run GPU workloads. Supports the additional `type` field (see below). |
| `prometheus` | object | No | Prometheus monitoring stack, required for NVIDIA Run:ai monitoring. On OpenShift, `mode: install` is rejected - use the cluster's built-in monitoring stack instead by setting `mode: skip`. |
| `networkOperator` | object | No | NVIDIA Network Operator. Required for RDMA or Multi-Node NVLink (MNNVL)-capable nodes, to enable high-performance networking features such as GPUDirect RDMA. |

### Feature Configuration

Configures components that extend NVIDIA Run:ai with support for specific workload types. Each component accepts the same fields as `prerequisites` - see [Common Prerequisite and Feature Fields](#common-prerequisite-and-feature-fields), below. For what each component does and when you need it, see [Features](prerequisites-and-features.md#features) in Prerequisites and Features.

The features you enable determine which workload types the cluster supports:

| Workload Type | Recommended Features |
|---|---|
| **Distributed training** | `kubeflow`, `mpiOperator` |
| **Standard (single-node) inference** | `knative` |
| **Distributed (multi-node) inference** | `lws` |
| **NIM inference** | `nimOperator` |

```yaml
features:
  skip: false
  knative:
    mode: install
  kubeflow:
    mode: install
  mpiOperator:
    mode: install
  lws:
    mode: install
  nimOperator:
    mode: skip
  grove:
    mode: skip
  dynamo:
    mode: skip
```

| Field | Type | Required | Description |
|---|---|---|---|
| `skip` | bool | No | Skips installation of all features. |
| `knative` | object | No | Knative Serving. Required for standard (single-node) inference workloads. Installing it requires `cluster.inference.wildcardDomain` to be set (plus a resolvable wildcard certificate, unless using Gateway API or OpenShift), or validation fails. |
| `kubeflow` | object | No | Kubeflow Training Operator. Required for distributed training workloads (TensorFlow, PyTorch, XGBoost, and JAX). Installs from pinned manifests - `customHelmValues` is not supported. |
| `mpiOperator` | object | No | MPI Operator. Required for MPI v2-based distributed training, alongside the frameworks the Kubeflow Training Operator provides. Installs from pinned manifests - `customHelmValues` is not supported. |
| `lws` | object | No | LeaderWorkerSet (LWS). Required for distributed (multi-node) inference workloads. |
| `nimOperator` | object | No | NVIDIA NIM Operator. Enables Kubernetes cluster administrators to operate the software components and services necessary to deploy NVIDIA NIMs and NVIDIA NeMo microservices. |
| `grove` | object | No | NVIDIA Grove. Orchestrates multi-component, gang-scheduled workloads, such as disaggregated NVIDIA Dynamo inference deployments. |
| `dynamo` | object | No | Dynamo Operator. Simplifies the deployment, configuration, and lifecycle management of DynamoGraphs. Installing it (`mode: install`) requires `grove.mode: install` as well; validation fails otherwise. |

### Common Prerequisite and Feature Fields

`gpuOperator`, `prometheus`, `networkOperator` (prerequisites) and `knative`, `kubeflow`, `mpiOperator`, `lws`, `nimOperator`, `grove`, `dynamo` (features) all accept the following fields:

```yaml
gpuOperator:
  mode: install
  type: real
  version: ""
  customHelmValues: {}
  openshift:
    installPlanApproval: manual
```

| Field | Type | Required | Description |
|---|---|---|---|
| `mode` | string | No | Installation behavior. Supported values: `install` (install the component if it is not already installed) and `skip` (do not install or require the component). If omitted, defaults to `install` for prerequisites and `skip` for features. |
| `type` | string | No | `gpuOperator` only. GPU Operator variant: `real` (default) for physical GPUs, or `fake` for testing without GPU hardware. |
| `version` | string | No | Component version. Omit to install the latest version. To set it explicitly, see [Support Matrix](../../support-matrix.md) for the supported version range. |
| `customHelmValues` | map | No | Custom Helm values merged into the component's Helm chart. Not supported for `kubeflow` or `mpiOperator`, which install from pinned manifests rather than a Helm chart - setting it there is rejected. On OpenShift, also rejected for components that install via OLM instead of Helm (see `openshift`, below). |
| `openshift` | object | No | OpenShift-specific settings - see [Per-Component OpenShift Fields](#per-component-openshift-fields), below. Used only when `install.kubernetesDistribution` is `openshift`, and only for components that install via OLM on OpenShift. Setting it for a component that does not is rejected. |

---

## Networking and Nodes

### External Connection

Configures how NVIDIA Run:ai's own charts expose their services - as Kubernetes `Ingress` objects or Gateway API `Gateway`/`HTTPRoute` objects. This assumes the matching controller is already running on your cluster (an Ingress Controller such as HAProxy, or a Gateway API implementation such as KGateway) - see [Ingress Controller](../../system-requirements/networking-setup.md#ingress-controller) for setup. Switching between the two after an initial install requires a manual migration; see [Kubernetes Gateway API](../../configuration/networking/kubernetes-gateway-api.md).

Required on standard Kubernetes clusters - set `type` to `ingress` or `gatewayApi`. Omit this section entirely on OpenShift, where OpenShift Routes are configured automatically; on OpenShift you may still set `type: gatewayApi` if you also want Gateway API routing alongside the Route, but `type: ingress` is rejected there. `route` is never a value you set explicitly - even on OpenShift, setting `type: route` is rejected; Routes are only ever configured automatically when the section is omitted.

```yaml
externalConnection:
  type: ingress
  ingress:
    className: haproxy
```

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | Conditional | External traffic routing method. Supported values: `ingress` (Kubernetes Ingress) and `gatewayApi` (Gateway API). Required unless the section is omitted on OpenShift. |
| `ingress.className` | string | Conditional | Name of the `IngressClass` resource to use. Required when `type` is `ingress`. Mutually exclusive with `gatewayApi`. |
| `gatewayApi.name` | string | Conditional | Name of the Gateway resource. Required when `type` is `gatewayApi`. |
| `gatewayApi.namespace` | string | Conditional | Namespace of the Gateway resource. Required when `type` is `gatewayApi`. Mutually exclusive with `ingress`. |

---

### System Nodes

Configures dedicated nodes for NVIDIA Run:ai system components, including the control plane services and cluster agents.

By default, NVIDIA Run:ai system components can be scheduled on any available node.

```yaml
systemNodes:
  enabled: true
  mode: preferred
  label: runai.ai/system-node
  manageLabels: true
  nodes:
    - node-1
    - node-2
```

| Field | Type | Required | Description |
|---|---|---|---|
| `enabled` | bool | Yes | Whether to enable system node scheduling. |
| `mode` | string | No | Scheduling behavior. Supported values: `preferred` (schedule system components on system nodes when available, but allow scheduling on other nodes if necessary) and `required` (schedule system components only on system nodes). Default: `preferred`. |
| `label` | string | No | Node label used to identify system nodes. Defaults to `node-role.kubernetes.io/runai-system` if omitted. |
| `manageLabels` | bool | No | Whether the installer automatically applies the system node label to the specified nodes. When `true`, exactly one of `nodes` or `selector` must be set - both or neither is rejected. |
| `nodes` | list | No | List of node names to designate as system nodes. Mutually exclusive with `selector` when `manageLabels` is `true`. |
| `selector` | map | No | Label selector used to identify system nodes instead of specifying individual node names. Mutually exclusive with `nodes` when `manageLabels` is `true`. |

---

## OpenShift

Applies only when `install.kubernetesDistribution: openshift`. Skip this section for every other supported distribution.

### OpenShift Configuration

Configures OpenShift-specific installer settings.

```yaml
openshift:
  olmCatalog:
    source: my-mirrored-catalog
    sourceNamespace: openshift-marketplace
```

| Field | Type | Required | Description |
|---|---|---|---|
| `olmCatalog.source` | string | No | Name of the OperatorHub `CatalogSource` used for all OLM-based operator installations. Most commonly a mirrored catalog for air-gapped OpenShift environments, but not limited to that case - any OpenShift install can point at a custom catalog, such as a vetted internal one. Omit to use the default public catalogs. |
| `olmCatalog.sourceNamespace` | string | No | Namespace containing the `CatalogSource`. |

### Per-Component OpenShift Fields

Set under any prerequisite's or feature's own `openshift` block (for example, `gpuOperator.openshift`) when that component installs via OperatorHub (OLM) instead of Helm. This is separate from the [OpenShift Configuration](#openshift-configuration) section, above, which determines the catalog source these operators are pulled from in the first place - these fields configure how each operator behaves once it is installed from that catalog.

| Field | Type | Required | Description |
|---|---|---|---|
| `installPlanApproval` | string | No | OLM upgrade strategy: `manual` (default; pins the operator to the installed version and requires manual approval to upgrade) or `automatic` (lets OLM upgrade the operator automatically when a new version is available in the channel). |
| `customResource` | map | No | Values merged into the operator's primary OpenShift custom resource spec. For example, `ClusterPolicy` for the GPU Operator, `NicClusterPolicy` for the Network Operator. Mirrors the approach an OpenShift admin uses to customize an OLM-installed operator by editing its CR rather than Helm values. |
| `subscriptionConfig.env` | map | No | Environment variables set on the operator pod via the OLM Subscription's `spec.config`. For settings that have no custom resource field, such as the NIM Operator's `ENABLE_WEBHOOKS`. |

---

## Air-Gapped Installation

Air-gapped is a connectivity model (`install.connectivity: air_gapped`), not a separate deployment model - the same `mode` values (`full_installation`, `runai_control_plane_only`, `runai_cluster_only`) still apply.

### Air-Gapped Installation Configuration

Configures the air-gapped installation bundle. Required when `install.connectivity` is set to `air_gapped`. Air-gapped installs authenticate to the registry with `registry.username`/`registry.password` (or `passwordPath`) only - token authentication (`registry.type`/`token`/`tokenPath`) is rejected. Prerequisites and features generally cannot be installed from the offline bundle, since it only contains the NVIDIA Run:ai charts; set `prerequisites.skip`/`features.skip` (or `mode: skip` per component) for anything not already present on the cluster.

```yaml
airgap:
  bundlePath: /data/runai-airgap-bundle.tar.gz
  workingDir: /data/runai-airgap
  valuesDir: /data/runai-airgap
  skipPrepare: false
```

| Field | Type | Required | Description |
|---|---|---|---|
| `bundlePath` | string | Conditional | Absolute path to the air-gapped bundle archive. Required unless `skipPrepare` is `true`. |
| `workingDir` | string | Yes | Directory where the installer extracts the bundle. The directory must have sufficient free disk space. The installer does not create temporary directories automatically. |
| `valuesDir` | string | No | Directory where the installer writes the generated Helm values files for the control plane and cluster charts, pointing them at your internal registry. Defaults to `workingDir` if not set. |
| `skipPrepare` | bool | No | Skips bundle extraction and image import into the registry. Use when rerunning an installation after the bundle has already been extracted and the images are already available in the registry. |

---

### Air-Gapped Bundle Commands

`airgap pack` and `airgap load` are separate commands from `install`, each with its own configuration file passed via its own `--config` flag - unlike `airgap` above, which is part of the `install` configuration.

#### Pack an Air-Gapped Bundle

Configures the `airgap pack` command, which builds an air-gapped bundle from a release's artifacts on a connected machine. See [Air-Gapped Bundle](air-gapped/preparations.md#air-gapped-bundle) for the full procedure.

```yaml
settings:
  runaiVersion: 3.0.2
  architecture: amd64
  output: /path/to/runai-airgap.tar

imageOverrides:
  - name: keycloak
    source: <your-registry>/keycloak:<tag>

registryAuth:
  - registry: <your-registry-host>
    username: <username>
    passwordPath: /path/to/registry-password
```

| Field | Type | Required | Description |
|---|---|---|---|
| `settings.runaiVersion` | string | No | Target NVIDIA Run:ai version to pack. |
| `settings.architecture` | string | No | CPU architecture to pack. Supported values: `amd64`, `arm64`. |
| `settings.output` | string | No | Path to write the resulting bundle archive. Defaults to `runai-<version>-airgap.tar` in the current directory. |
| `imageOverrides` | list | No | Substitutes or adds images sourced from your own registry instead of the release's defaults - for example, a third-party service image such as Keycloak. |
| `imageOverrides[].group` | string | Conditional | Image group the override applies to. Required only when `name` exists in more than one group. |
| `imageOverrides[].name` | string | Yes (per entry) | Name of the image to override. |
| `imageOverrides[].source` | string | Yes (per entry) | Full image reference to pull instead, including an explicit tag or digest. |
| `registryAuth` | list | No | Credentials for registries referenced by `imageOverrides` sources, or for a Support-Entitled Release's own registry. Each entry authenticates to one registry. |
| `registryAuth[].registry` | string | Yes (per entry) | Registry hostname this entry authenticates to. |
| `registryAuth[].username` / `password` / `passwordPath` / `passwordEnv` | string | No | Basic authentication. `password`, `passwordPath`, and `passwordEnv` are mutually exclusive. |
| `registryAuth[].token` / `tokenPath` / `tokenEnv` | string | No | Token authentication. Mutually exclusive with each other. |
| `registryAuth[].bearerToken` / `bearerTokenPath` / `bearerTokenEnv` | string | No | Bearer token authentication. Mutually exclusive with each other. |
| `registryAuth[].caChainPath` | string | No | Path to the CA certificate chain. Required when the registry uses a private CA. |
| `registryAuth[].insecure` | bool | No | Connects to the registry over HTTP instead of HTTPS. |
| `registryAuth[].insecureSkipVerify` | bool | No | Connects over HTTPS without verifying the server's TLS certificate. |

---

#### Load an Air-Gapped Bundle

Configures the `airgap load` command, which extracts an air-gapped bundle, pushes its images to your internal registry, and writes the Helm values files consumed by `install`'s own [`airgap`](#air-gapped-installation-configuration) configuration. See [Air-Gapped Bundle](air-gapped/preparations.md#air-gapped-bundle) for the full procedure.

```yaml
settings:
  bundlePath: /path/to/runai-airgap.tar
  workingDir: /data/runai-airgap
  valuesDir: /data/runai-airgap

registry:
  url: <your-registry-host>
  username: <username>
  passwordPath: /path/to/registry-password
```

| Field | Type | Required | Description |
|---|---|---|---|
| `settings.bundlePath` | string | Yes | Path to the bundle archive produced by `airgap pack`. Must exist and be a file, not a directory. |
| `settings.workingDir` | string | Yes | Directory where the bundle is extracted. |
| `settings.valuesDir` | string | No | Directory where the generated Helm values files (`control-plane-airgap-values.yaml`, `cluster-airgap-values.yaml`) are written. Defaults to `workingDir` if not set. |
| `registry.url` | string | Yes | Destination registry for the pushed images, as a fully qualified hostname, `localhost`, or `host:port`, with an optional path. No scheme, and no tag or digest. |
| `registry.username` / `password` / `passwordPath` / `passwordEnv` | string | No | Basic authentication. `password`, `passwordPath`, and `passwordEnv` are mutually exclusive. |
| `registry.token` / `tokenPath` / `tokenEnv` | string | No | Token authentication. Mutually exclusive with each other. |
| `registry.caChainPath` | string | No | Path to the CA certificate chain. Required when the registry uses a private CA. |
| `registry.insecure` | bool | No | Connects to the registry over HTTP instead of HTTPS. |
| `registry.insecureSkipVerify` | bool | No | Connects over HTTPS without verifying the server's TLS certificate. |
