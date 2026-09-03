---
---

# Multi-Cluster Installation

This page describes how to install NVIDIA Run:ai in a [multi-cluster deployment](../../README.md#deployment-model), where the NVIDIA Run:ai control plane runs on a dedicated Kubernetes cluster and one or more NVIDIA Run:ai clusters run on separate Kubernetes clusters. To install the control plane and cluster on the same Kubernetes cluster, see [Same-Cluster Installation](same-cluster-installation.md).

In a multi-cluster deployment, the `runai-installer` CLI runs once per Kubernetes cluster with a dedicated configuration file:

1. **Install the control plane** (`mode: runai_control_plane_only`) on the control plane cluster.
2. **Install the cluster component** (`mode: runai_cluster_only`) on each NVIDIA Run:ai cluster. The installer registers each cluster with the control plane automatically.

The configuration shown here is the minimum required to install. For complete, ready-to-use configurations for specific use cases, see [Connected Templates](connected.md) or [Air-Gapped Templates](air-gapped.md). For a complete list of configuration fields, see the [Configuration Reference](configuration-reference.md).

## Before You Start

Complete [System Requirements](../../system-requirements/README.md) before running the installer, for **each** Kubernetes cluster - networking, DNS, TLS certificates, and other infrastructure the installation depends on. The installer deploys and configures NVIDIA Run:ai itself; it does not provision this underlying infrastructure for you.

Have the following at hand for the installation:

| Requirement | Details |
|---|---|
| Kubernetes access | `kubectl` configured with cluster-admin or equivalent permissions on the control plane cluster and on each target cluster |
| Control plane access | A kubeconfig with access to the control plane's Kubernetes cluster - the installer uses it to obtain the registration token that connects each cluster to the control plane |
| Network connectivity | Outbound HTTPS (port 443) from the nodes of each NVIDIA Run:ai cluster to the control plane FQDN. See [Connectivity](../../system-requirements/deployment-model.md#connectivity) |
| NGC API key | Required for NVIDIA Run:ai Support-Entitled Release installations only, to pull NVIDIA Run:ai images and Helm charts from NGC. See [Registry Credentials](#registry-credentials) below |
| FQDNs | A resolvable domain name for the NVIDIA Run:ai control plane, and per NVIDIA Run:ai cluster: a cluster domain name and a wildcard FQDN for externally exposed workloads. See [FQDN and DNS](../../system-requirements/networking-setup.md#fqdn-and-dns) |
| TLS certificates | A certificate and private key for each of your FQDNs, or set `tls.generateSelfSigned: true` to have the installer generate self-signed certificates automatically (non-production environments only). See [TLS Certificates](../../system-requirements/networking-setup.md#tls-certificates) |
| Prerequisites and features | GPU Operator and Prometheus are required for every installation. Network Operator, and features such as Knative and Kubeflow, apply only if your environment or planned workload types need them. See [Prerequisites and Features](prerequisites-and-features.md) for what each one does and when you need it. |

### Registry Credentials

NVIDIA Run:ai Support-Entitled Release installations require an NGC Personal Key to authenticate to the NVIDIA Run:ai container registry (`nvcr.io`) and Helm registry (`helm.ngc.nvidia.com/nvidia/runai`) for pulling NVIDIA Run:ai container images and charts. Open Source installations pull without credentials and can skip this step. Generate a Personal Key from the [NGC Personal Keys](https://org.ngc.nvidia.com/setup/personal-keys) page - the same key is also used to download the air-gapped installation package, if applicable.

Personal Keys begin with `nvapi-`. NGC authenticates using the literal username `$oauthtoken` with the key as the password.

For an air-gapped install, this NGC Personal Key is only used to download the bundle - it is not part of the `registry` section in your configuration file. That section instead authenticates to your own private mirror registry (populated from the bundle) using basic authentication (`username` and `password` or `passwordPath`); NGC token authentication is rejected for air-gapped installs. See [`registry`](configuration-reference.md#registry) and [`airgap`](configuration-reference.md#airgap) in the Configuration Reference.

## Download the Installer

Download the binary for your CPU architecture on each machine you run the installer from:

```bash
# AMD64 (x86_64)
curl -LO <release-url>/runai-installer-amd64
chmod +x runai-installer-amd64

# ARM64
curl -LO <release-url>/runai-installer-arm64
chmod +x runai-installer-arm64
```

## Step 1: Install the Control Plane

Create a `cp-config.yaml` file for the control plane cluster. At a minimum, configure:

- `install.mode` - Set to `runai_control_plane_only` to install the control plane only. See [`install`](configuration-reference.md#install)
- `install.connectivity` - `connected` or `air_gapped`
- `registry` - Your NGC credentials, required only for NVIDIA Run:ai Support-Entitled Release installations. Open Source installations still set `registry.type: ngc`, but without credentials
- `controlPlane.domain` - The FQDN for the NVIDIA Run:ai UI and API
- `controlPlane.customHelmValues` - Merges Helm values into the control plane chart, for settings not otherwise exposed by the installer. See [Advanced Control Plane Configurations](../../configuration/control-plane-config.md) for the full set of keys it accepts.

The following example installs the NVIDIA Run:ai control plane with a CA-signed TLS certificate:

```yaml
install:
  mode: runai_control_plane_only
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token

controlPlane:
  enabled: true
  domain: runai.mycorp.local
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues: {}

prerequisites:
  gpuOperator:
    mode: skip

externalConnection:
  type: ingress
```

The GPU Operator is set to `skip` because the control plane cluster does not run GPU workloads.

Before applying any changes, validate your configuration against the control plane cluster with `--preflight-only`:

```bash
./runai-installer install \
  --config cp-config.yaml \
  --kubeconfig /path/to/cp-kubeconfig \
  --preflight-only
```

To go further and preview the full installation plan without applying it, run a dry run instead:

```bash
./runai-installer install \
  --config cp-config.yaml \
  --kubeconfig /path/to/cp-kubeconfig \
  --dry-run
```

Once you're ready, run the installer against the control plane cluster:

```bash
./runai-installer install \
  --config cp-config.yaml \
  --kubeconfig /path/to/cp-kubeconfig \
  --logs-dir ./logs
```

When the installation completes, verify that the control plane is reachable at its FQDN (for example, `https://runai.mycorp.local`) before continuing.

## Step 2: Install Each NVIDIA Run:ai Cluster

Create a `cluster-config.yaml` file for each NVIDIA Run:ai cluster. At a minimum, configure:

- `install.mode` - Set to `runai_cluster_only` to install the cluster component only. See [`install`](configuration-reference.md#install)
- `install.connectivity` - `connected` or `air_gapped`
- `registry` - Your NGC credentials, required only for NVIDIA Run:ai Support-Entitled Release installations. Open Source installations still set `registry.type: ngc`, but without credentials
- `cluster.name` - The name of the NVIDIA Run:ai cluster as it appears in the control plane
- `cluster.domain` - The FQDN for this cluster. Required when the cluster component is installed on a different Kubernetes cluster than the control plane. The installer also uses this domain to reach the control plane when registering the cluster
- `cluster.customHelmValues` - Merges Helm values into the cluster chart, for settings not otherwise exposed by the installer. See [Advanced Cluster Configurations](../../configuration/cluster-config.md) for the full set of keys it accepts.

<Info>

**Note**

Cluster registration is automatic - the installer obtains the registration token from the control plane. The control plane must already be installed and reachable at the cluster's domain.

</Info>

The following example installs the NVIDIA Run:ai cluster component on a separate Kubernetes cluster and registers it with the control plane:

```yaml
install:
  mode: runai_cluster_only
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token

cluster:
  enabled: true
  name: cluster-a
  domain: runai-a.mycorp.local
  tls:
    certPath: /path/to/cluster-a-cert.pem
    keyPath: /path/to/cluster-a-key.pem
  customHelmValues: {}

prerequisites:
  gpuOperator:
    mode: install

externalConnection:
  type: ingress
```

Before applying any changes, validate your configuration against the target cluster with `--preflight-only`:

```bash
./runai-installer install \
  --config cluster-config.yaml \
  --kubeconfig /path/to/cluster-a-kubeconfig \
  --preflight-only
```

To go further and preview the full installation plan without applying it, run a dry run instead:

```bash
./runai-installer install \
  --config cluster-config.yaml \
  --kubeconfig /path/to/cluster-a-kubeconfig \
  --dry-run
```

Once you're ready, run the installer against the target cluster:

```bash
./runai-installer install \
  --config cluster-config.yaml \
  --kubeconfig /path/to/cluster-a-kubeconfig \
  --logs-dir ./logs
```

Repeat this step for each additional NVIDIA Run:ai cluster with its own configuration file - each cluster requires a unique `cluster.name`, its own `cluster.domain`, and a matching TLS certificate.

## Verify the Installation

After each installer run completes successfully, verify that the NVIDIA Run:ai pods are running:

```bash
# On the control plane cluster
kubectl get pods -n runai-backend

# On each NVIDIA Run:ai cluster
kubectl get pods -n runai
```

All pods should reach `Running` or `Completed` status.

Then log-in to the NVIDIA Run:ai UI at the control plane domain and verify that each installed cluster appears under its configured `cluster.name` with a **Connected** status. If a cluster does not show as **Connected**, see [Troubleshooting scenarios](../operations/clusters.md#troubleshooting-scenarios).

## Command Reference

The installer commands and flags are identical for all installation modes - run each command against the relevant cluster with its own `--config` and `--kubeconfig`.

### Global Flags

These flags apply to all `runai-installer` subcommands.

| Flag             | Description                                     | Default           |
| ---------------- | ----------------------------------------------- | ----------------- |
| `--logs-dir`     | Directory where installer log files are written | Current directory |
| `--step-timeout` | Maximum time allowed for each installation step | `7m`              |
| `--dev`          | Development mode                                | `false`           |

### `install` Command

The `install` command installs NVIDIA Run:ai from a configuration file.

```bash
./runai-installer install [flags]
```

| Flag | Description | Default |
|---|---|---|
| `--config` | Path to the installation configuration file **(required)** | - |
| `--dry-run` | Plan the installation without applying any changes | `false` |
| `--preflight-only` | Run preflight checks and exit without installing | `false` |
| `--skip-preflight` | Skip preflight checks and proceed with installation | `false` |
| `-y`, `--yes` | Run non-interactively without confirmation prompts | `false` |
| `--output` | Output format: `text` or `json` | `text` on a TTY, `json` otherwise |
| `--events-file` | Write NDJSON installation events to a file | - |
| `--status-file` | Write the final installation summary to a file | - |
| `--on-failure` | Failure policy: `abort` or `skip` | `abort` |
| `--kubeconfig` | Path to a kubeconfig file | - |
| `--env` | Run:ai environment selecting chart repos: `prod`, `staging`, `lab`, or `test` | `prod` |
| `--ngc-token-path` | Path to an NGC API key file (overrides `registry` in the configuration file) | - |
| `--dev-mode` | Relax input validation for development and testing | `false` |
| `--execution-timeout` | Maximum time allowed for the installation | `50m` |

## Troubleshooting

### Cluster Cannot Reach the Control Plane

If cluster registration fails or a cluster does not show as **Connected**, verify that the NVIDIA Run:ai cluster's nodes can reach the control plane FQDN over HTTPS (port 443):

```bash
curl -v https://<control-plane-domain>
```

### Installation Step Timed Out

Increase the per-step timeout for slow networks or large image pulls:

```bash
./runai-installer install \
  --config config.yaml \
  --step-timeout 15m
```

### View Installer Logs

```bash
# Stream logs in real time
tail -f ./logs/runai-installer.log

# Show only errors
grep -i error ./logs/runai-installer.log

# Filter Helm operations
grep "\[Helm\]" ./logs/runai-installer.log

# Filter Kubernetes API calls
grep "\[K8s\]" ./logs/runai-installer.log
```

### Registry Authentication Failed

Verify that the NGC API key file is valid and readable:

```bash
cat /path/to/ngc-token
```

The file must contain only the API key. The `runai-installer` CLI automatically trims leading and trailing whitespace.
