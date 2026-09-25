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
| FQDNs | The FQDNs and DNS records required for your deployment. Requirements depend on your deployment model, workload routing, and enabled workload features. See [FQDN and DNS](../../../system-requirements/networking-setup.md#fqdn-and-dns) and [Domains and TLS Certificates](#domains-and-tls-certificates) below |
| TLS certificates | The TLS certificates and private keys required for your configured domains. See [TLS Certificates](../../../system-requirements/networking-setup.md#tls-certificates) and [Domains and TLS Certificates](#domains-and-tls-certificates) below |

Multi-cluster deployments also need:

| Requirement | Details |
|---|---|
| Control plane access | The control plane's FQDN (`cluster.controlPlaneUrl`), reachable over HTTPS from each NVIDIA Run:ai cluster, and a way to identify the cluster to the control plane - see [Install Each NVIDIA Run:ai Cluster](multi-cluster.md#install-each-nvidia-runai-cluster) |
| Network connectivity | Outbound HTTPS (port 443) from the nodes of each NVIDIA Run:ai cluster to the control plane FQDN. See [Control Plane and Cluster Sync](../../../system-requirements/external-access-to-critical-services.md#control-plane-and-cluster-sync) |

### Domains and TLS Certificates

Create the DNS records and obtain the certificate and private key files described in [FQDN and DNS](../../../system-requirements/networking-setup.md#fqdn-and-dns) and [TLS Certificates](../../../system-requirements/networking-setup.md#tls-certificates) before installation, then supply them in your `config.yaml` as follows.

#### Control Plane Domain and Certificate

Secures access to the NVIDIA Run:ai UI and API:

```yaml
controlPlane:
  domain: runai.mycorp.local
  tls:
    certPath: /path/to/fullchain.pem
    keyPath: /path/to/private.pem
```

#### Cluster Domain and Certificate

Only needed when the NVIDIA Run:ai cluster is installed on a different Kubernetes cluster than the control plane (multi-cluster deployment). In a same-cluster deployment, omit `cluster.domain` and `cluster.tls` - the cluster uses the control plane's domain and certificate instead. See [Workspaces, Training, and Distributed Inference Wildcard Certificate](#workspaces-training-and-distributed-inference-wildcard-certificate) below for the host-based routing wildcard certificate.

```yaml
cluster:
  domain: runai-a.mycorp.local
  tls:
    certPath: /path/to/fullchain.pem
    keyPath: /path/to/private.pem
```

#### Workspaces, Training, and Distributed Inference Wildcard Certificate

Required for the default host-based routing, to expose workspace, training, and distributed inference workloads via subdomains. Not required for path-based routing (`cluster.workloadRouting: path`).

<Warning>

**TODO**

It is not yet confirmed whether this wildcard certificate is set via `cluster.tls` (the same field as the Cluster Domain and Certificate above) or a separate field. Confirm with engineering before publishing.

</Warning>

#### Standard Inference Wildcard Domain and Certificate

Required when the `knative` feature is installed, to expose standard inference workloads externally. Not required for cluster-local standard inference, or for other inference workloads such as distributed inference using LWS, Dynamo, or KServe.

```yaml
cluster:
  inference:
    wildcardDomain: "*.runai-inference.mycorp.local"
    wildcardTls:
      certPath: /path/to/wildcard-fullchain.pem
      keyPath: /path/to/wildcard-private.pem
```

#### Local Certificate Authority (CA)

If your certificates are signed by a private or local CA rather than a publicly trusted one, provide the CA chain. Set this at the top level, or override it per-component under `controlPlane.tls.caChainPath` / `cluster.tls.caChainPath` if different components use different CAs:

```yaml
tls:
  caChainPath: /path/to/ca-chain.pem
```

See [`tls`](../configuration-reference.md#tls), [`controlPlane`](../configuration-reference.md#controlplane), and [`cluster`](../configuration-reference.md#cluster) in the Configuration Reference for all fields.

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
