---
---

# Same-Cluster Installation

This page describes how to install the NVIDIA Run:ai control plane and cluster together, on one Kubernetes cluster (`mode: full_installation`), over an internet-connected environment. To install the control plane and cluster on separate Kubernetes clusters instead, see [Multi-Cluster Installation](multi-cluster.md).

## Before You Start

Complete [Preparations](preparations.md) before continuing - system requirements for your Kubernetes cluster, registry credentials if you need them, and downloading the `runai-installer` binary. The installer deploys and configures NVIDIA Run:ai itself; it does not provision this underlying infrastructure for you.

## Create a Configuration File

<Note>

**Configuration Templates**

The configuration shown below is the minimum required to install. If your use case matches one of the [Configuration Templates](templates.md), copy that template instead of building a configuration from scratch. For a complete list of configuration fields to build your own configuration from scratch, see the [Configuration Reference](../configuration-reference.md).

</Note>

Create a `config.yaml` file describing your installation. At a minimum, configure:

- `install.mode` - Set to `full_installation` to install the control plane and cluster on the same Kubernetes cluster. For the other modes, see [`install`](../configuration-reference.md#install).
- `install.connectivity` - `connected`, for an installation with outbound internet access.
- `registry` - Open Source installations set `registry.type: ngc` with no credentials needed. NVIDIA Run:ai Support-Entitled Release installations set the same field, plus your NGC credentials. See [Registry Credentials](preparations.md#registry-credentials).
- `controlPlane.domain` - The FQDN for the NVIDIA Run:ai UI and API. See [FQDN and DNS](../../../system-requirements/networking-setup.md#fqdn-and-dns).
- `controlPlane.tls` / `cluster.tls` - A certificate and private key for each FQDN, as `certPath` and `keyPath`. See [TLS Certificates](../../../system-requirements/networking-setup.md#tls-certificates).
- `cluster.name` - The name of the NVIDIA Run:ai cluster
- `controlPlane.version` / `cluster.version` - The control plane or cluster version to install. Omit either to install the latest supported version. See [`controlPlane`](../configuration-reference.md#controlplane) and [`cluster`](../configuration-reference.md#cluster).
- `prerequisites` - GPU Operator and Prometheus are required for every installation. Set both to `mode: install`. Each also accepts its own `version` field - omit it to install the latest version, or see [Support Matrix](../../../support-matrix.md) for the supported version range if you want to set it explicitly. See [Prerequisites and Features](../prerequisites-and-features.md#prerequisites).
- `externalConnection` - Required on standard Kubernetes, even if workloads will not be exposed externally - the control plane UI/API and the cluster endpoint are routed through it. See [Networking Setup](../../../system-requirements/networking-setup.md).

<Note>

**Note**

`controlPlane.customHelmValues` / `cluster.customHelmValues` merge Helm values into the control plane or cluster chart, for settings not otherwise exposed by the installer - see [`controlPlane`](../configuration-reference.md#controlplane) and [`cluster`](../configuration-reference.md#cluster) for how to set it. See [Advanced Control Plane Configurations](../../../configuration/control-plane-config.md) and [Advanced Cluster Configurations](../../../configuration/cluster-config.md) for the full set of keys each accepts.

</Note>

The following example installs the NVIDIA Run:ai control plane and cluster on the same Kubernetes cluster:

```yaml
install:
  mode: full_installation
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

cluster:
  enabled: true
  name: my-cluster
  tls:
    certPath: /path/to/cluster-wildcard-cert.pem
    keyPath: /path/to/cluster-wildcard-key.pem
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

Then log-in to the NVIDIA Run:ai UI at the control plane domain you configured (for example, `https://runai.mycorp.local`) and verify that the cluster appears with a **Connected** status. If the cluster does not show as **Connected**, see [Troubleshooting scenarios](../../../operations/clusters.md#troubleshooting-scenarios).

## Command Reference and Troubleshooting

See [Command Reference and Troubleshooting](../command-reference.md) for installer flags and common troubleshooting scenarios.
