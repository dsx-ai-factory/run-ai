---
---

# Preparations

This page describes how to prepare for installing NVIDIA Run:ai using the `runai-installer` CLI over an internet-connected environment. For environments without outbound internet access, see [Air-Gapped Installation](../air-gapped/preparations.md).

NVIDIA Run:ai supports two deployment models - see [Same-Cluster Installation](same-cluster.md) to install the control plane and cluster together on one Kubernetes cluster, or [Multi-Cluster Installation](multi-cluster.md) to install the control plane and one or more clusters on separate Kubernetes clusters. For a complete list of configuration fields, see the [Configuration Reference](../configuration-reference.md).

## System Requirements

Complete [System Requirements](../../../system-requirements/README.md) before running the installer, for each Kubernetes cluster you install onto. It covers the infrastructure you must provision yourself beforehand - hardware and Kubernetes distribution, networking and DNS, TLS certificates, and outbound firewall/registry access - as well as the third-party services the installation deploys for you. The installer deploys and configures NVIDIA Run:ai itself; it does not provision this underlying infrastructure for you.

Have the following at hand for the installation:

| Requirement | Details |
|---|---|
| Kubernetes access | `kubectl` configured with cluster-admin or equivalent permissions on the target cluster - on the control plane cluster and on each NVIDIA Run:ai cluster, in a multi-cluster deployment |
| FQDNs | A resolvable domain name for the NVIDIA Run:ai control plane, and a wildcard FQDN for externally exposed workloads - per cluster, in a multi-cluster deployment. See [FQDN and DNS](../../../system-requirements/networking-setup.md#fqdn-and-dns) |
| TLS certificates | A certificate and private key for each of your FQDNs, or set `tls.generateSelfSigned: true` to have the installer generate self-signed certificates automatically (non-production environments only). See [TLS Certificates](../../../system-requirements/networking-setup.md#tls-certificates) |

Multi-cluster deployments also need:

| Requirement | Details |
|---|---|
| Control plane access | The control plane's FQDN (`cluster.controlPlaneUrl`), reachable over HTTPS - the installer registers each cluster by calling the control plane's API directly, authenticated with credentials it reads from a Kubernetes Secret |
| Network connectivity | Outbound HTTPS (port 443) from the nodes of each NVIDIA Run:ai cluster to the control plane FQDN. See [Control Plane and Cluster Sync](../../../system-requirements/external-access-to-critical-services.md#control-plane-and-cluster-sync) |

<Warning>

**TODO**

Unconfirmed how the installer obtains the control plane's admin credentials when registering a cluster in a genuine multi-cluster deployment. The registration step reads these credentials from a Kubernetes Secret via the same Kubernetes client used for the current run - which, for `runai_cluster_only` (run with `--kubeconfig` pointed at the target NVIDIA Run:ai cluster, not the control plane), would be the target cluster's client, not the control plane's. Need to confirm with engineering whether a separate credential-passing mechanism exists for this case, or whether this is a real gap in the current flow.

</Warning>

## Registry Credentials

Open Source installations set `registry.type: ngc` in the configuration, with no credentials needed. NVIDIA Run:ai Support-Entitled Release installations set the same field, plus an NGC Personal Key to authenticate to the NVIDIA Run:ai container registry (`nvcr.io`) and Helm registry (`helm.ngc.nvidia.com/nvidia/runai`) for pulling NVIDIA Run:ai container images and charts. Generate a Personal Key from the [NGC Personal Keys](https://org.ngc.nvidia.com/setup/personal-keys) page.

Personal Keys begin with `nvapi-`. NGC authenticates using the literal username `$oauthtoken` with the key as the password.

## Download the Installer

<Warning>

**TODO**

`<release-url>` is a placeholder - the real download URL for the `runai-installer` binary is not yet finalized. Replace it once the release location is confirmed.

</Warning>

Download the `runai-installer` binary for your operating system and CPU architecture, on each machine you run the installer from:

```bash
# Linux, AMD64 (x86_64)
curl -LO <release-url>/runai-installer-linux-amd64
chmod +x runai-installer-linux-amd64

# Linux, ARM64
curl -LO <release-url>/runai-installer-linux-arm64
chmod +x runai-installer-linux-arm64

# macOS, AMD64 (Intel)
curl -LO <release-url>/runai-installer-darwin-amd64
chmod +x runai-installer-darwin-amd64

# macOS, ARM64 (Apple Silicon)
curl -LO <release-url>/runai-installer-darwin-arm64
chmod +x runai-installer-darwin-arm64
```

## Next Steps

* [Same-Cluster Installation](same-cluster.md) - Install the control plane and cluster together, on one Kubernetes cluster.
* [Multi-Cluster Installation](multi-cluster.md) - Install the control plane and one or more clusters on separate Kubernetes clusters.
* [Configuration Templates](templates.md) - Complete, ready-to-use configurations for specific use cases.
