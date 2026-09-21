---
---

# Multi-Cluster Installation

This page describes how to install the NVIDIA Run:ai control plane on a dedicated Kubernetes cluster, and one or more NVIDIA Run:ai clusters on separate Kubernetes clusters, over an internet-connected environment. To install the control plane and cluster together on one Kubernetes cluster instead, see [Same-Cluster Installation](same-cluster.md).

The `runai-installer` CLI runs once per Kubernetes cluster with a dedicated configuration file:

1. **Install the control plane** (`mode: runai_control_plane_only`) on the control plane cluster.
2. **Install the cluster component** (`mode: runai_cluster_only`) on each NVIDIA Run:ai cluster, and register it with the control plane.

## Before You Start

Complete [Preparations](preparations.md) before continuing - system requirements for your Kubernetes cluster, registry credentials if you need them, and downloading the `runai-installer` binary. The installer deploys and configures NVIDIA Run:ai itself; it does not provision this underlying infrastructure for you.

## Install the Control Plane

<Note>

**Configuration Templates**

The configuration shown below is the minimum required to install. If your use case matches one of the [Configuration Templates](templates.md), copy that template instead of building a configuration from scratch. For a complete list of configuration fields to build your own configuration from scratch, see the [Configuration Reference](../configuration-reference.md).

</Note>

Create a `cp-config.yaml` file for the control plane cluster. At a minimum, configure:

- `install.mode` - Set to `runai_control_plane_only` to install the control plane only. See [`install`](../configuration-reference.md#install).
- `install.connectivity` - `connected`, for an installation with outbound internet access.
- `registry` - Open Source installations set `registry.type: ngc` with no credentials needed. NVIDIA Run:ai Support-Entitled Release installations set the same field, plus your NGC credentials. See [Registry Credentials](preparations.md#registry-credentials).
- `controlPlane.domain` - The FQDN for the NVIDIA Run:ai UI and API. See [FQDN and DNS](../../../system-requirements/networking-setup.md#fqdn-and-dns).
- `controlPlane.tls` - A certificate and private key for the control plane's FQDN, as `certPath` and `keyPath`. See [TLS Certificates](../../../system-requirements/networking-setup.md#tls-certificates).
- `controlPlane.version` - The control plane version to install. Omit it to install the latest supported version. See [`controlPlane`](../configuration-reference.md#controlplane).
- `externalConnection` - Required on standard Kubernetes, even if workloads will not be exposed externally - the control plane UI/API and the cluster endpoint are routed through it. See [Networking Setup](../../../system-requirements/networking-setup.md).

<Note>

**Note**

`controlPlane.customHelmValues` merges Helm values into the control plane chart, for settings not otherwise exposed by the installer - see [`controlPlane`](../configuration-reference.md#controlplane) for how to set it. See [Advanced Control Plane Configurations](../../../configuration/control-plane-config.md) for the full set of keys it accepts.

</Note>

The following example installs the NVIDIA Run:ai control plane:

```yaml
install:
  mode: runai_control_plane_only
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

controlPlane:
  enabled: true
  domain: runai.mycorp.local
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues: {}

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

You can also generate this configuration from the NVIDIA Run:ai UI instead of writing it by hand - the onboarding wizard or the Clusters form walks through the same choices and generates a starter `config.yaml` and `runai-installer` command for you. See [Clusters](../../../operations/clusters.md).

At a minimum, configure:

- `install.mode` - Set to `runai_cluster_only` to install the cluster component only. See [`install`](../configuration-reference.md#install).
- `install.connectivity` - `connected`, for an installation with outbound internet access.
- `registry` - Open Source installations set `registry.type: ngc` with no credentials needed. NVIDIA Run:ai Support-Entitled Release installations set the same field, plus your NGC credentials. See [Registry Credentials](preparations.md#registry-credentials).
- `cluster.name` - The name of the NVIDIA Run:ai cluster as it appears in the control plane
- `cluster.domain` - The FQDN for this cluster. Required when the cluster component is installed on a different Kubernetes cluster than the control plane. The installer also uses this domain to reach the control plane when registering the cluster.
- `cluster.controlPlaneUrl` - The control plane's FQDN, reachable over HTTPS from this cluster.
- `cluster.uid` plus one of `cluster.bootstrapTokenPath`/`cluster.clientSecretPath`, or `cluster.controlPlaneCredentials` - Identifies this cluster to the control plane. See [Preparations](preparations.md#system-requirements) for the two available methods.
- `cluster.tls` - A certificate and private key for the cluster's FQDN, as `certPath` and `keyPath`. See [TLS Certificates](../../../system-requirements/networking-setup.md#tls-certificates).
- `cluster.version` - The cluster version to install. Omit it to install the latest supported version. See [`cluster`](../configuration-reference.md#cluster).
- `prerequisites` - GPU Operator and Prometheus are required for every installation. Set both to `mode: install`. Each also accepts its own `version` field - omit it to install the latest version, or see [Support Matrix](../../../support-matrix.md) for the supported version range if you want to set it explicitly. See [Prerequisites and Features](../prerequisites-and-features.md#prerequisites).
- `externalConnection` - Required on standard Kubernetes, even if workloads will not be exposed externally - the control plane UI/API and the cluster endpoint are routed through it. See [Networking Setup](../../../system-requirements/networking-setup.md).

<Note>

**Note**

`cluster.customHelmValues` merges Helm values into the cluster chart, for settings not otherwise exposed by the installer - see [`cluster`](../configuration-reference.md#cluster) for how to set it. See [Advanced Cluster Configurations](../../../configuration/cluster-config.md) for the full set of keys it accepts.

</Note>

The following example installs the NVIDIA Run:ai cluster component on a separate Kubernetes cluster, and has the installer register it with the control plane using a control plane token:

```yaml
install:
  mode: runai_cluster_only
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

cluster:
  enabled: true
  name: cluster-a
  domain: runai-a.mycorp.local
  controlPlaneUrl: https://runai.mycorp.local
  controlPlaneCredentials:
    tokenPath: /path/to/cp-token
  tls:
    certPath: /path/to/cluster-a-cert.pem
    keyPath: /path/to/cluster-a-key.pem
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
