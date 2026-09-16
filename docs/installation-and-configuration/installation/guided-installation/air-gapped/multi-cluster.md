---
---

# Multi-Cluster Installation

This page describes how to install the NVIDIA Run:ai control plane on a dedicated Kubernetes cluster, and one or more NVIDIA Run:ai clusters on separate Kubernetes clusters, from an air-gapped bundle. To install the control plane and cluster together on one Kubernetes cluster instead, see [Same-Cluster Installation](same-cluster.md).

The `runai-installer` CLI runs once per Kubernetes cluster with a dedicated configuration file:

1. **Install the control plane** (`mode: runai_control_plane_only`) on the control plane cluster.
2. **Install the cluster component** (`mode: runai_cluster_only`) on each NVIDIA Run:ai cluster, and register it with the control plane - see the TODO note under [Install Each NVIDIA Run:ai Cluster](#install-each-nvidia-runai-cluster) below.

## Before You Start

Complete [Preparations](preparations.md) before continuing - system requirements for your Kubernetes cluster, registry credentials, downloading the `runai-installer` binary, and building your [Air-Gapped Bundle](preparations.md#air-gapped-bundle). The installer deploys and configures NVIDIA Run:ai itself; it does not provision this underlying infrastructure for you.

## Install the Control Plane

<Info>

**Configuration Templates**

The configuration shown below is the minimum required to install. If your use case matches one of the [Configuration Templates](templates.md), copy that template instead of building a configuration from scratch. For a complete list of configuration fields to build your own configuration from scratch, see the [Configuration Reference](../configuration-reference.md).

</Info>

Create a `cp-config.yaml` file for the control plane cluster. At a minimum, configure:

- `install.mode` - Set to `runai_control_plane_only` to install the control plane only. See [`install`](../configuration-reference.md#install).
- `install.connectivity` - `air_gapped`, for an installation with no outbound internet access.
- `registry` - Credentials for your own internal mirror registry (populated by loading the air-gapped bundle), not NGC. Uses basic authentication only. See [Registry Credentials](preparations.md#registry-credentials).
- `airgap` - The same `workingDir` and `valuesDir` you used with `airgap load`, plus `skipPrepare: true` - the bundle is already loaded, so `install` skips straight to the chart install. See [Air-Gapped Bundle](preparations.md#air-gapped-bundle) and [`airgap`](../configuration-reference.md#airgap) in the Configuration Reference.
- `controlPlane.domain` - The FQDN for the NVIDIA Run:ai UI and API. See [FQDN and DNS](../../../system-requirements/networking-setup.md#fqdn-and-dns).
- `controlPlane.tls` - A certificate and private key for the control plane's FQDN, as `certPath` and `keyPath`. See [TLS Certificates](../../../system-requirements/networking-setup.md#tls-certificates).
- `externalConnection` - Required on standard Kubernetes, even if workloads will not be exposed externally - the control plane UI/API and the cluster endpoint are routed through it. See [Networking Setup](../../../system-requirements/networking-setup.md).

<Info>

**Note**

`controlPlane.customHelmValues` merges Helm values into the control plane chart, for settings not otherwise exposed by the installer - see [`controlPlane`](../configuration-reference.md#controlplane) for how to set it. See [Advanced Control Plane Configurations](../../../configuration/control-plane-config.md) for the full set of keys it accepts.

</Info>

The following example installs the NVIDIA Run:ai control plane from an already-loaded air-gapped bundle:

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

Create a `cluster-config.yaml` file for each NVIDIA Run:ai cluster.

<Warning>

**TODO**

This section is stale - cluster registration is not fully automatic. `runai_cluster_only` now requires `cluster.controlPlaneUrl`, `cluster.uid`, and `cluster.bootstrapTokenPath` in the cluster config, obtained from the control plane. How the user obtains these values is still being decided by engineering (as of 2026-09). Update this note, the field list, and the example config below once that's settled. The control plane must already be installed and reachable at the cluster's domain.

</Warning>

At a minimum, configure:

- `install.mode` - Set to `runai_cluster_only` to install the cluster component only. See [`install`](../configuration-reference.md#install).
- `install.connectivity` - `air_gapped`, for an installation with no outbound internet access.
- `registry` - Credentials for your own internal mirror registry (populated by loading the air-gapped bundle), not NGC. Uses basic authentication only. See [Registry Credentials](preparations.md#registry-credentials).
- `airgap` - The same `workingDir` and `valuesDir` you used with `airgap load`, plus `skipPrepare: true` - the bundle is already loaded, so `install` skips straight to the chart install. See [Air-Gapped Bundle](preparations.md#air-gapped-bundle) and [`airgap`](../configuration-reference.md#airgap) in the Configuration Reference.
- `cluster.name` - The name of the NVIDIA Run:ai cluster as it appears in the control plane
- `cluster.domain` - The FQDN for this cluster. Required when the cluster component is installed on a different Kubernetes cluster than the control plane. The installer also uses this domain to reach the control plane when registering the cluster.
- `cluster.tls` - A certificate and private key for the cluster's FQDN, as `certPath` and `keyPath`. See [TLS Certificates](../../../system-requirements/networking-setup.md#tls-certificates).
- `externalConnection` - Required on standard Kubernetes, even if workloads will not be exposed externally - the control plane UI/API and the cluster endpoint are routed through it. See [Networking Setup](../../../system-requirements/networking-setup.md).

<Info>

**Note**

- `cluster.customHelmValues` merges Helm values into the cluster chart, for settings not otherwise exposed by the installer - see [`cluster`](../configuration-reference.md#cluster) for how to set it. See [Advanced Cluster Configurations](../../../configuration/cluster-config.md) for the full set of keys it accepts.
- GPU Operator and Prometheus are required for every installation, but the installer does not install them in air-gapped environments - they are skipped by default. Install both yourself beforehand, using each vendor's own air-gapped installation instructions. See [Prerequisites and Features](preparations.md#prerequisites-and-features).

</Info>

The following example installs the NVIDIA Run:ai cluster component from an already-loaded air-gapped bundle:

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
  tls:
    certPath: /path/to/cluster-a-cert.pem
    keyPath: /path/to/cluster-a-key.pem
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
