---
---

# Preparations

This page describes how to prepare for installing NVIDIA Run:ai using the `runai-installer` CLI over an internet-connected environment. For environments without outbound internet access, see [Air-Gapped Installation](../air-gapped/preparations.md).

## System Requirements

Complete the [System Requirements](../../../system-requirements/README.md) section before starting. It covers:

* [Infrastructure Requirements](../../../system-requirements/infrastructure-requirements.md) - Hardware, operating system, Kubernetes distribution, storage, and permissions required in your Kubernetes environment.
* [Networking Setup](../../../system-requirements/networking-setup.md) - Load balancer and ingress requirements, FQDNs, TLS certificates and private keys, and a local CA chain if your certificates are privately signed. The exact requirements depend on your deployment model, routing configuration, and the workload types you plan to support.
* [Third-Party Services](../../../system-requirements/third-party-services.md) - The third-party services deployed as part of the installation.
* [External Access to Critical Services](../../../system-requirements/external-access-to-critical-services.md) - The inbound and outbound connectivity required to reach those services and NVIDIA registries.

Your [deployment model](../../../system-requirements/deployment-model.md) determines how the installation steps are applied. In a same-cluster deployment, the control plane and NVIDIA Run:ai cluster are installed on the same Kubernetes cluster. In a multi-cluster deployment, control plane setup is completed once, and cluster setup is repeated for each NVIDIA Run:ai cluster.

## Domains and TLS Certificates

Each domain and certificate you prepared in [FQDN and DNS](../../../system-requirements/networking-setup.md#fqdn-and-dns) and [TLS Certificates](../../../system-requirements/networking-setup.md#tls-certificates) maps to a specific `config.yaml` field under `controlPlane` or `cluster`. Which ones you need depends on your deployment model (same-cluster or multi-cluster), your workload routing, and the workload types you plan to support - use this section as a reference for where each one goes. You supply the actual values when you write your `config.yaml` in [Same-Cluster Installation](same-cluster.md) or [Multi-Cluster Installation](multi-cluster.md).

### Control Plane Domain and Certificate

This certificate secures the NVIDIA Run:ai control plane FQDN you prepared in [FQDN and DNS](../../../system-requirements/networking-setup.md#fqdn-and-dns), protecting access to the NVIDIA Run:ai UI and API. See [TLS Certificates](../../../system-requirements/networking-setup.md#tls-certificates) to confirm which certificate you need and why.

```yaml
controlPlane:
  domain: runai.mycorp.local
  tls:
    certPath: /path/to/fullchain.pem
    keyPath: /path/to/private.pem
```

### Cluster Domain and Certificate

This certificate secures communication between the NVIDIA Run:ai control plane and the cluster, associated with the cluster's domain. In a same-cluster deployment, omit `cluster.domain` - the cluster reuses `controlPlane.domain`, since both run on the same Kubernetes cluster. `cluster.tls` is still required for both same-cluster and multi-cluster deployments.

```yaml
cluster:
  domain: runai-a.mycorp.local # Multi-cluster only; omit for same-cluster
  tls:
    certPath: /path/to/fullchain.pem
    keyPath: /path/to/private.pem
```

### Workspaces, Training, and Distributed Inference Wildcard Certificate

<Note>

**Note**

For path-based routing, this certificate is not required.

</Note>

To allow secure access to workspace, training, and distributed inference workloads via subdomains, configure a wildcard TLS certificate that matches the cluster domain - `*.runai.mycorp.local` in a same-cluster deployment, or `*.runai-a.mycorp.local` for cluster A in a multi-cluster deployment. Required for both same-cluster and multi-cluster deployments.

```yaml
cluster:
  domain: runai-a.mycorp.local # Multi-cluster only; omit for same-cluster - see Cluster Domain and Certificate above
  tls:
    certPath: /path/to/wildcard-fullchain.pem # matches *.runai.mycorp.local (same-cluster) or *.runai-a.mycorp.local (multi-cluster, cluster A)
    keyPath: /path/to/wildcard-private.pem
```

### Standard Inference Wildcard Domain and Certificate

<Note>

**Note**

The inference wildcard certificate is only required for standard inference (Knative, single-node) workloads exposed externally via wildcard FQDN. Other inference workloads (for example, distributed inference using serving frameworks such as LWS, Dynamo, KServe, or others) do not require this certificate. For cluster-local standard inference, skip this section.

</Note>

To securely expose inference services over HTTPS, configure a wildcard TLS certificate for the inference domain (e.g. `*.runai-inference.mycorp.local`).

```yaml
cluster:
  inference:
    wildcardDomain: "*.runai-inference.mycorp.local"
    wildcardTls:
      certPath: /path/to/wildcard-fullchain.pem
      keyPath: /path/to/wildcard-private.pem
```

### Local Certificate Authority (CA)

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
