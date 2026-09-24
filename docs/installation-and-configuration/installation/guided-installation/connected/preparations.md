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
| Control plane access | The control plane's FQDN (`cluster.controlPlaneUrl`), reachable over HTTPS from each NVIDIA Run:ai cluster, and a way to identify the cluster to the control plane - see below |
| Network connectivity | Outbound HTTPS (port 443) from the nodes of each NVIDIA Run:ai cluster to the control plane FQDN. See [Control Plane and Cluster Sync](../../../system-requirements/external-access-to-critical-services.md#control-plane-and-cluster-sync) |

Identify the NVIDIA Run:ai cluster to the control plane using one of two methods:

- Let the installer add the cluster - The default for a new installation. Provide a control plane login, and the installer creates the cluster in the control plane for you (or reuses it, if a cluster with that name already exists there). Set `cluster.controlPlaneCredentials` to either a username and password (`username`, plus `passwordPath` or inline `password`) or a token (`tokenPath` or inline `token`). The control plane's own admin login works here - the installer prints it after installing the control plane. See [Install the Control Plane](multi-cluster.md#install-the-control-plane).
- Cluster already added in the control plane - Use this method if someone already added the cluster in the control plane UI or API beforehand, where it now shows status **Waiting to connect**. Provide the identity the control plane issued for it: `cluster.uid`, plus one of `cluster.bootstrapTokenPath` or `cluster.clientSecretPath`.

## Registry Credentials

Open Source installations set `registry.type: ngc` in the configuration, with no credentials needed. NVIDIA Run:ai Support-Entitled Release installations set the same field, plus an NGC Personal Key to authenticate to the NVIDIA Run:ai container registry (`nvcr.io`) and Helm registry (`helm.ngc.nvidia.com/nvidia/runai`) for pulling NVIDIA Run:ai container images and charts. Generate a Personal Key from the [NGC Personal Keys](https://org.ngc.nvidia.com/setup/personal-keys) page.

Personal Keys begin with `nvapi-`. NGC authenticates using the literal username `$oauthtoken` with the key as the password.

<Note>

**Note**

The installer does not yet expose a dedicated field for Support-Entitled Release branding (the UI branding and support contact link shown in the control plane). Set it yourself via `controlPlane.customHelmValues.supportEntitled: true`. See [`customHelmValues`](../configuration-reference.md#controlplane) in the Configuration Reference.

</Note>

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
