---
---

# Preparations

This page describes how to prepare for installing NVIDIA Run:ai using the `runai-installer` CLI over an internet-connected environment. For environments without outbound internet access, see [Air-Gapped Installation](../air-gapped/preparations.md).

## System Requirements

Complete [System Requirements](../../../system-requirements/README.md) before running the installer, for each Kubernetes cluster you install onto - hardware and Kubernetes distribution, networking and DNS, TLS certificates, outbound firewall/registry access, and the third-party services the installation deploys for you. The installer deploys and configures NVIDIA Run:ai itself; it does not provision this infrastructure.

Have the following at hand for the installation. Make sure you have the domains, TLS certificates, and other values required for your deployment - these depend on your deployment model, workload routing, and the workload features you plan to enable. See [Networking Setup](../../../system-requirements/networking-setup.md) and [Prerequisites and Features](../prerequisites-and-features.md).

| Requirement | Details |
|---|---|
| Kubernetes access | `kubectl` configured with cluster-admin or equivalent permissions on the target cluster - on the control plane cluster and on each NVIDIA Run:ai cluster, in a multi-cluster deployment |
| FQDNs | The FQDNs and DNS records required for your deployment. Requirements depend on your deployment model, workload routing, and enabled workload features. See [FQDN and DNS](../../../system-requirements/networking-setup.md#fqdn-and-dns) |
| TLS certificates | The TLS certificates and private keys required for your configured domains. See [TLS Certificates](../../../system-requirements/networking-setup.md#tls-certificates) |

Multi-cluster deployments also need:

| Requirement | Details |
|---|---|
| Control plane access | The control plane's FQDN (`cluster.controlPlaneUrl`), reachable over HTTPS from each NVIDIA Run:ai cluster, and a way to identify the cluster to the control plane - see [Install Each NVIDIA Run:ai Cluster](multi-cluster.md#install-each-nvidia-runai-cluster) |
| Network connectivity | Outbound HTTPS (port 443) from the nodes of each NVIDIA Run:ai cluster to the control plane FQDN. See [Control Plane and Cluster Sync](../../../system-requirements/external-access-to-critical-services.md#control-plane-and-cluster-sync) |

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

* [Same-Cluster Installation](same-cluster.md) - The minimum configuration to install the control plane and cluster together, on one Kubernetes cluster.
* [Multi-Cluster Installation](multi-cluster.md) - The minimum configuration to install the control plane and one or more clusters on separate Kubernetes clusters.
* [Configuration Templates](templates.md) - Complete, ready-to-use configurations for specific use cases - use one instead of the minimum configuration if it matches yours.
* [Configuration Reference](../configuration-reference.md) - All supported configuration fields, for building a configuration from scratch.
