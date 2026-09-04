---
---

# Same-Cluster Installation

This page describes how to install NVIDIA Run:ai in a [same-cluster deployment](../../README.md#deployment-model), where the NVIDIA Run:ai control plane and the NVIDIA Run:ai cluster run on the same Kubernetes cluster (`mode: full_installation`). To install the control plane and clusters on separate Kubernetes clusters, see [Multi-Cluster Installation](multi-cluster-installation.md).

The configuration shown here is the minimum required to install. For complete, ready-to-use configurations for specific use cases, see [Connected Templates](connected.md) or [Air-Gapped Templates](air-gapped.md). For a complete list of configuration fields, see the [Configuration Reference](configuration-reference.md).

## Before You Start

Complete [System Requirements](../../system-requirements/README.md) before running the installer - your Kubernetes cluster, networking, DNS, TLS certificates, and other infrastructure the installation depends on. The installer deploys and configures NVIDIA Run:ai itself; it does not provision this underlying infrastructure for you.

Have the following at hand for the installation:

| Requirement | Details |
|---|---|
| Kubernetes access | `kubectl` configured with cluster-admin or equivalent permissions on the target cluster |
| NGC API key | Required for NVIDIA Run:ai Support-Entitled Release installations only, to pull NVIDIA Run:ai images and Helm charts from NGC. See [Registry Credentials](#registry-credentials) below |
| FQDNs | A resolvable domain name for the NVIDIA Run:ai control plane, and a wildcard FQDN for externally exposed workloads. See [FQDN and DNS](../../system-requirements/networking-setup.md#fqdn-and-dns) |
| TLS certificates | A certificate and private key for each of your FQDNs, or set `tls.generateSelfSigned: true` to have the installer generate self-signed certificates automatically (non-production environments only). See [TLS Certificates](../../system-requirements/networking-setup.md#tls-certificates) |
| Prerequisites and features | GPU Operator and Prometheus are required for every installation. Network Operator, and features such as Knative and Kubeflow, apply only if your environment or planned workload types need them. See [Prerequisites and Features](prerequisites-and-features.md) for what each one does and when you need it. |

### Registry Credentials

NVIDIA Run:ai Support-Entitled Release installations require an NGC Personal Key to authenticate to the NVIDIA Run:ai container registry (`nvcr.io`) and Helm registry (`helm.ngc.nvidia.com/nvidia/runai`) for pulling NVIDIA Run:ai container images and charts. Open Source installations pull without credentials and can skip this step. Generate a Personal Key from the [NGC Personal Keys](https://org.ngc.nvidia.com/setup/personal-keys) page - the same key is also used to download the air-gapped installation package, if applicable.

Personal Keys begin with `nvapi-`. NGC authenticates using the literal username `$oauthtoken` with the key as the password.

For an air-gapped install, this NGC Personal Key is only used to download the bundle - it is not part of the `registry` section in your configuration file. That section instead authenticates to your own private mirror registry (populated from the bundle) using basic authentication (`username` and `password` or `passwordPath`); NGC token authentication is rejected for air-gapped installs. See [`registry`](configuration-reference.md#registry) and [`airgap`](configuration-reference.md#airgap) in the Configuration Reference.

## Download the Installer

Download the binary for your CPU architecture:

```bash
# AMD64 (x86_64)
curl -LO <release-url>/runai-installer-amd64
chmod +x runai-installer-amd64

# ARM64
curl -LO <release-url>/runai-installer-arm64
chmod +x runai-installer-arm64
```

## Create a Configuration File

Create a `config.yaml` file describing your installation. At a minimum, configure:

- `install.mode` - Set to `full_installation` to install the control plane and cluster on the same Kubernetes cluster. For the other modes, see [`install`](configuration-reference.md#install)
- `install.connectivity` - `connected` or `air_gapped`
- `registry` - Your NGC credentials, required only for NVIDIA Run:ai Support-Entitled Release installations. Open Source installations still set `registry.type: ngc`, but without credentials
- `controlPlane.domain` - The FQDN for the NVIDIA Run:ai UI and API
- `cluster.name` - The name of the NVIDIA Run:ai cluster
- `controlPlane.customHelmValues` / `cluster.customHelmValues` - Merges Helm values into the control plane or cluster chart, for settings not otherwise exposed by the installer. See [Advanced Control Plane Configurations](../../configuration/control-plane-config.md) and [Advanced Cluster Configurations](../../configuration/cluster-config.md) for the full set of keys each accepts.

The following example installs the NVIDIA Run:ai control plane and cluster on the same Kubernetes cluster using a self-signed TLS certificate:

```yaml
install:
  mode: full_installation
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token

tls:
  generateSelfSigned: true

controlPlane:
  enabled: true
  domain: runai.mycorp.local
  customHelmValues: {}

cluster:
  enabled: true
  name: my-cluster
  customHelmValues: {}

prerequisites:
  gpuOperator:
    mode: install

externalConnection:
  type: ingress
```

## Run the Installer

Before applying any changes, validate your configuration against the target cluster with `--preflight-only`:

```bash
./runai-installer install \
  --config /path/to/config.yaml \
  --preflight-only
```

To go further and preview the full installation plan without applying it, run a dry run instead:

```bash
./runai-installer install \
  --config /path/to/config.yaml \
  --dry-run
```

Once you're ready, run the installer:

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

Then log-in to the NVIDIA Run:ai UI at the control plane domain you configured (for example, `https://runai.mycorp.local`) and verify that the cluster appears with a **Connected** status. If the cluster does not show as **Connected**, see [Troubleshooting scenarios](../operations/clusters.md#troubleshooting-scenarios).

## Command Reference

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
