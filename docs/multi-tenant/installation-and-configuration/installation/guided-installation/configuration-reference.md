---
---

# Configuration Reference

The `runai-installer` CLI reads a configuration YAML file specified with the `--config` flag. This reference describes every supported configuration field, for both installing the control plane in multi-tenant mode and installing a tenant's cluster.

Fields marked **Required** must be specified. All other fields are optional.

## Configuration Overview

**Core Installation Settings**

- [Installation Settings](#installation-settings) - Installation mode and network connectivity model.
- [Registry Configuration](#registry-configuration) - Credentials for pulling NVIDIA Run:ai images and Helm charts: token authentication (NGC), basic authentication (Harbor and other OCI registries), or no authentication.
- [Default TLS Configuration](#default-tls-configuration) - Default TLS certificate for the control plane or cluster. The component-level `tls` block overrides this default.

**NVIDIA Run:ai Components**

- [Control Plane Configuration](#control-plane-configuration) - Control plane deployment: whether to install it, its wildcard domain, version, TLS, multi-tenant mode, and custom Helm values.
- [Cluster Configuration](#cluster-configuration) - Tenant cluster settings, including how to register it with its tenant using an identity obtained from the control plane's Management API.

**Prerequisites and Features**

- [Prerequisite Configuration](#prerequisite-configuration) - Infrastructure components the installer can set up on a tenant cluster before NVIDIA Run:ai: GPU Operator, Prometheus, and Network Operator, each independently set to install or skip.
- [Feature Configuration](#feature-configuration) - Components that extend a tenant cluster for specific workload types: Knative for standard inference, LWS for distributed inference, Kubeflow and MPI Operator for distributed training, plus NIM Operator, Grove, and Dynamo.
- [Common Prerequisite and Feature Fields](#common-prerequisite-and-feature-fields) - Fields shared by every prerequisite and feature, such as `mode`, `version`, and `customHelmValues`.

**Networking and Nodes**

- [External Connection](#external-connection) - How external traffic reaches the control plane or a tenant cluster: Kubernetes Ingress or Gateway API. Omitted entirely on OpenShift, where Routes are configured automatically.
- [System Nodes](#system-nodes) - Dedicating specific nodes to control plane or cluster system components, with a preferred or required scheduling mode and automatic node labeling.

**OpenShift**

- [OpenShift Configuration](#openshift-configuration) - Installer-wide OpenShift settings for tenant clusters, such as a mirrored OperatorHub catalog source.
- [Per-Component OpenShift Fields](#per-component-openshift-fields) - OpenShift-specific fields available under individual prerequisites and features.

---

## Core Installation Settings

### Installation Settings

Controls the installation type and network connectivity model.

```yaml
install:
  mode: runai_control_plane_only
  connectivity: connected
  kubernetesDistribution: ""
```

| Field | Type | Required | Description |
|---|---|---|---|
| `mode` | string | No | Installation mode. Set to `runai_control_plane_only` when installing the control plane, or `runai_cluster_only` when installing a tenant cluster - each uses its own configuration file. |
| `connectivity` | string | Yes | Network connectivity model. Set to `connected`. |
| `kubernetesDistribution` | string | No | Kubernetes distribution. Set to `openshift` when installing a tenant cluster on Red Hat OpenShift - this is the only recognized non-empty value. Omit for every other supported distribution (Upstream Kubernetes or RKE2). Not applicable to the control plane, which does not support OpenShift. |

---

### Registry Configuration

Configures the registry credentials used to pull NVIDIA Run:ai images and Helm charts. Which method you need depends on your installation:

- **Open Source installation** (`type: ngc`, no credentials) - Pulls NVIDIA Run:ai's publicly accessible images from NGC without credentials. `host` can be omitted; it defaults to NGC's registry (`nvcr.io`).
- **Token authentication** (`type` with `token` or `tokenPath`) - For an NVIDIA Run:ai Support-Entitled Release installation, pulling from NVIDIA's own NGC registry (`type: ngc`). For any other registry secured by a base64-encoded `.dockerconfigjson` pull secret instead of NGC credentials, use `type: dockerconfig`.
- **Basic authentication** (`username` with `password` or `passwordPath`) - For a private registry you control, such as Harbor or another OCI-compliant registry.
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
```

No authentication:

```yaml
registry:
  host: <REGISTRY_HOST>
```

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | Conditional | Registry type. Required when using token authentication, and for an Open Source installation's NGC pull. Supported values: `ngc` (NVIDIA GPU Cloud) and `dockerconfig` (a base64-encoded `.dockerconfigjson` pull secret for any other registry). |
| `host` | string | Conditional | Registry hostname. Optional when a token is set, or when `type` is `ngc` (even without a token - falls back to `nvcr.io`). Required for basic authentication and any other unauthenticated registry. |
| `token` | string | No | Registry token specified inline. Mutually exclusive with `tokenPath`. For production environments, use `tokenPath` to avoid storing secrets in the configuration file. |
| `tokenPath` | string | No | Path to a file containing the registry token. Mutually exclusive with `token`. |
| `username` | string | No | Username for basic authentication. Requires `password` or `passwordPath`. |
| `password` | string | No | Password specified inline. Mutually exclusive with `passwordPath`. |
| `passwordPath` | string | No | Path to a file containing the password. Mutually exclusive with `password`. |
| `chartRepository` | string | No | Overrides the default NGC Helm chart repository URL. |
| `insecure` | bool | No | Connects to the registry over HTTP instead of HTTPS. |
| `insecureSkipVerify` | bool | No | Connects over HTTPS without verifying the server's TLS certificate. |

---

### Default TLS Configuration

Configures the default TLS settings used by the control plane. A configuration that defines its own `controlPlane.tls` block uses that block instead of this one; a configuration that omits `controlPlane.tls` falls back to these top-level settings.

```yaml
tls:
  certPath: /path/to/cert.pem
  keyPath: /path/to/key.pem
  generateSelfSigned: false
```

| Field | Type | Required | Description |
|---|---|---|---|
| `certPath` | string | No | Absolute path to the TLS certificate file in PEM format. Must be a wildcard certificate matching the control plane domain. |
| `keyPath` | string | No | Absolute path to the TLS private key file in PEM format. |
| `generateSelfSigned` | bool | No | Generates a self-signed TLS certificate automatically. Use only for non-production environments. |

---

## NVIDIA Run:ai Components

### Control Plane Configuration

Configures the NVIDIA Run:ai control plane.

```yaml
controlPlane:
  enabled: true
  version: ""
  multiTenant: true
  domain: <WILDCARD_DOMAIN>
  tls:
    certPath: /path/to/cert.pem
    keyPath: /path/to/key.pem
  customHelmValues: {}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `enabled` | bool | Yes | Whether to install the control plane. |
| `version` | string | No | Control plane version. Omit to install the latest supported version. |
| `multiTenant` | bool | No | Set to `true` to install the control plane in multi-tenant mode. |
| `domain` | string | Yes | Wildcard fully qualified domain name (FQDN) where the NVIDIA Run:ai UI and API are available for every tenant (for example, `*.runai.hostorg.com`). |
| `tls` | [`tls`](#default-tls-configuration) | No | Control plane TLS settings. Overrides the top-level `tls` configuration. Must be a wildcard certificate matching `domain`. |
| `customHelmValues` | map | No | Custom Helm values merged into the control plane chart. Use this field to configure settings that are not exposed directly by the installer, including the control plane administrator credentials and the external metrics store connection - see [Advanced Control Plane Configurations](../../../../installation-and-configuration/configuration/control-plane-config.md) for the full set of keys. |

---

### Cluster Configuration

Configures a tenant's NVIDIA Run:ai cluster component. This section is applied once per tenant cluster, in its own configuration file. The cluster must already be [registered with its tenant](create-a-tenant-and-cluster.md#register-the-cluster) through the control plane's Management API before running the installer - the installer consumes that registration, it does not perform it.

```yaml
cluster:
  enabled: true
  version: ""
  name: <CLUSTER_NAME>
  domain: <CLUSTER_DOMAIN>
  workloadRouting: host
  tls:
    certPath: /path/to/cert.pem
    keyPath: /path/to/key.pem
  inference:
    wildcardDomain: "*.<INFERENCE_WILDCARD_DOMAIN>"
    wildcardTls:
      certPath: /path/to/wildcard-cert.pem
      keyPath: /path/to/wildcard-key.pem
  controlPlaneUrl: https://<CONTROL_PLANE_DOMAIN>
  uid: <CLUSTER_UUID>
  bootstrapTokenPath: /path/to/bootstrap-token
  customHelmValues: {}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `enabled` | bool | Yes | Whether to install the cluster component. |
| `version` | string | No | Cluster component version. Omit to install the latest supported version. |
| `name` | string | Yes | Cluster name as it appears in the NVIDIA Run:ai control plane. Must follow Kubernetes naming conventions (lowercase letters, numbers, and hyphens; maximum 63 characters). |
| `domain` | string | Yes | Fully qualified domain name (FQDN) for this tenant cluster. |
| `controlPlaneUrl` | string | Yes | URL of the NVIDIA Run:ai control plane, reachable over HTTPS from this cluster, so the installer can complete registration using the identity below. |
| `uid` | string | Yes | Unique identifier of the cluster, obtained by [registering it with its tenant](create-a-tenant-and-cluster.md#register-the-cluster) through the Management API. Used together with `bootstrapTokenPath` or `clientSecretPath`. |
| `bootstrapTokenPath` | string | Conditional | Path to a file containing the bootstrap token for the cluster identified by `uid`, obtained from the same Management API call. Mutually exclusive with `clientSecretPath`; exactly one is required. |
| `clientSecretPath` | string | Conditional | Path to a file containing the client secret for the cluster identified by `uid`, obtained from the Management API. Mutually exclusive with `bootstrapTokenPath`. |
| `workloadRouting` | string | No | Workload routing method. Supported values: `host` (subdomain-based routing, default) and `path` (path-based routing). |
| `tls` | [`tls`](#default-tls-configuration) | No | Cluster TLS settings. Overrides the top-level `tls` configuration. |
| `inference.wildcardDomain` | string | No | Wildcard FQDN for inference endpoints (for example, `*.runai-inference.tenant-a-corp.com`). Required when Knative is installed. |
| `inference.wildcardTls` | [`tls`](#default-tls-configuration) | No | TLS settings for the wildcard inference domain. |
| `customHelmValues` | map | No | Custom Helm values merged into the cluster chart. Use this field to configure settings that are not exposed directly by the installer - see [Advanced Cluster Configurations](../../../../installation-and-configuration/configuration/cluster-config.md) for the full set of keys. |

---

## Prerequisites and Features

### Prerequisite Configuration

Configures the infrastructure components that the installer can set up on a tenant cluster before deploying the NVIDIA Run:ai cluster component. Each of `gpuOperator`, `prometheus`, and `networkOperator` accepts the same set of fields - see [Common Prerequisite and Feature Fields](#common-prerequisite-and-feature-fields), below.

```yaml
prerequisites:
  skip: false
  gpuOperator:
    mode: install
  prometheus:
    mode: install
  networkOperator:
    mode: skip
```

| Field | Type | Required | Description |
|---|---|---|---|
| `skip` | bool | No | Skips installation of all prerequisites. Equivalent to setting `mode: skip` for every prerequisite. |
| `gpuOperator` | object | No | NVIDIA GPU Operator. Provisions and manages GPUs on worker nodes; required for a tenant cluster to run GPU workloads. Supports the additional `type` field (see below). |
| `prometheus` | object | No | Prometheus monitoring stack, required for NVIDIA Run:ai monitoring. |
| `networkOperator` | object | No | NVIDIA Network Operator. Required for RDMA or Multi-Node NVLink (MNNVL)-capable nodes, to enable high-performance networking features such as GPUDirect RDMA. |

### Feature Configuration

Configures components that extend a tenant cluster with support for specific workload types. Each component accepts the same fields as `prerequisites` - see [Common Prerequisite and Feature Fields](#common-prerequisite-and-feature-fields), below.

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
| `knative` | object | No | Knative Serving. Required for standard (single-node) inference workloads. Installing it requires `cluster.inference.wildcardDomain` to be set (plus a resolvable wildcard certificate, unless using Gateway API), or validation fails. |
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
```

| Field | Type | Required | Description |
|---|---|---|---|
| `mode` | string | No | Installation behavior. Supported values: `install` (install the component if it is not already installed) and `skip` (do not install or require the component). If omitted, defaults to `install` for prerequisites and `skip` for features. |
| `type` | string | No | `gpuOperator` only. GPU Operator variant: `real` (default) for physical GPUs, or `fake` for testing without GPU hardware. |
| `version` | string | No | Component version. Omit to install the latest version. |
| `customHelmValues` | map | No | Custom Helm values merged into the component's Helm chart. Not supported for `kubeflow` or `mpiOperator`, which install from pinned manifests rather than a Helm chart - setting it there is rejected. On OpenShift, components that install via OperatorHub (OLM) instead of Helm are configured with `openshift` instead (see below). |
| `openshift` | object | No | OpenShift-specific settings - see [Per-Component OpenShift Fields](#per-component-openshift-fields), below. Used only when `install.kubernetesDistribution` is `openshift`, and only for components that install via OLM on OpenShift. |

---

## Networking and Nodes

### External Connection

Configures how the control plane's or a tenant cluster's chart exposes its services - as Kubernetes `Ingress` objects or Gateway API `Gateway`/`HTTPRoute` objects. This assumes the matching controller is already running on the target cluster (an Ingress Controller such as HAProxy, or a Gateway API implementation such as KGateway) - see [Ingress Controller](../../system-requirements/networking-setup.md#ingress-controller) for setup. Omit this section entirely on OpenShift, where OpenShift Routes are configured automatically.

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

Configures dedicated nodes for the control plane's or a tenant cluster's system components.

By default, these components can be scheduled on any available node.

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

Applies to tenant clusters only, when `install.kubernetesDistribution: openshift`. The NVIDIA Run:ai control plane runs on Kubernetes. Skip this section for every other supported distribution.

### OpenShift Configuration

Configures OpenShift-specific installer settings for a tenant cluster.

```yaml
openshift:
  olmCatalog:
    source: my-mirrored-catalog
    sourceNamespace: openshift-marketplace
```

| Field | Type | Required | Description |
|---|---|---|---|
| `olmCatalog.source` | string | No | Name of the OperatorHub `CatalogSource` used for all OLM-based operator installations. Any OpenShift install can point at a custom catalog, such as a mirrored or vetted internal one. Omit to use the default public catalogs. |
| `olmCatalog.sourceNamespace` | string | No | Namespace containing the `CatalogSource`. |

### Per-Component OpenShift Fields

Set under any prerequisite's or feature's own `openshift` block (for example, `gpuOperator.openshift`) when that component installs via OperatorHub (OLM) instead of Helm. This is separate from the [OpenShift Configuration](#openshift-configuration) section, above, which determines the catalog source these operators are pulled from in the first place - these fields configure how each operator behaves once it is installed from that catalog.

| Field | Type | Required | Description |
|---|---|---|---|
| `installPlanApproval` | string | No | OLM upgrade strategy: `manual` (default; pins the operator to the installed version and requires manual approval to upgrade) or `automatic` (lets OLM upgrade the operator automatically when a new version is available in the channel). |
| `customResource` | map | No | Values merged into the operator's primary OpenShift custom resource spec. For example, `ClusterPolicy` for the GPU Operator, `NicClusterPolicy` for the Network Operator. Mirrors the approach an OpenShift admin uses to customize an OLM-installed operator by editing its CR rather than Helm values. |
| `subscriptionConfig.env` | map | No | Environment variables set on the operator pod via the OLM Subscription's `spec.config`. For settings that have no custom resource field, such as the NIM Operator's `ENABLE_WEBHOOKS`. |
