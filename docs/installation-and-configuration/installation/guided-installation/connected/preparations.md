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

<Note>

**Gov-ready**

For a Gov-ready installation, also review [Gov-Ready Third-Party Services](../../../system-requirements/gov-ready-third-party-services.md) and the [Gov-ready outbound access](../../../system-requirements/external-access-to-critical-services.md#gov-ready) to Docker Hub, which replace the standard third-party services images.

</Note>

Your [deployment model](../../../system-requirements/deployment-model.md) determines how the installation steps are applied. In a same-cluster deployment, the control plane and NVIDIA Run:ai cluster are installed on the same Kubernetes cluster. In a multi-cluster deployment, control plane setup is completed once, and cluster setup is repeated for each NVIDIA Run:ai cluster.

## Prerequisites and Features

GPU Operator and Prometheus are required for every installation. Network Operator, and features such as Knative and Kubeflow, apply only if your environment or planned workload types need them. Your planned workload types may also affect the networking configuration required for the installation, including wildcard domains and TLS certificates. See [Prerequisites and Features](../prerequisites-and-features.md) to determine which prerequisites and features you need and how to configure them.

In connected environments, the installer can set up supported prerequisites and features for you. By default, it uses the latest supported version; to use a different supported version, specify the `version` field for that component according to the versions listed in the [Support Matrix](../../../support-matrix.md).

<Note>

**Gov-ready**

For a Gov-ready installation, use the Gov-ready versions of the GPU Operator and Network Operator. Refer to the Gov-ready documentation of the [NVIDIA GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/install-gpu-operator-gov-ready.html) and the [NVIDIA Network Operator](https://docs.nvidia.com/networking/display/kubernetes2670/install-network-operator-gov-ready.html) for the preparation they require, and pass the values they require under the operator's `customHelmValues`.

</Note>

## Domains and TLS Certificates

Each domain and certificate you prepared in [FQDN and DNS](../../../system-requirements/networking-setup.md#fqdn-and-dns) and [TLS Certificates](../../../system-requirements/networking-setup.md#tls-certificates) maps to a specific `config.yaml` field under `controlPlane` or `cluster`. Which ones you need depends on your deployment model (same-cluster or multi-cluster), your workload routing, and the workload types you plan to support - use this section as a reference for where each one goes. You supply the actual values when you write your `config.yaml` in [Same-Cluster Installation](same-cluster.md) or [Multi-Cluster Installation](multi-cluster.md).

| Domain | Config Fields | Required When | Certificate |
| --- | --- | --- | --- |
| Control plane domain (for example, `runai.mycorp.local`) | `controlPlane.domain`, `controlPlane.tls` | Always | The exact domain |
| Cluster domain (for example, `runai-a.mycorp.local`) | `cluster.domain`, `cluster.tls` | `cluster.domain`: multi-cluster only. In same-cluster, omit it and the cluster reuses the control plane domain. `cluster.tls`: always. | Covered by `cluster.tls`, see the next row |
| Workspaces, training, and distributed inference (for example, `*.runai-a.mycorp.local`) | None. Do not set a separate domain field. The wildcard is `*.` plus your cluster domain. | Host-based routing (the default). Not needed with path-based routing. | The same `cluster.tls` certificate, which must list both `runai-a.mycorp.local` and `*.runai-a.mycorp.local` as Subject Alternative Names |
| Standard inference (Knative), for example `*.runai-inference.mycorp.local` | `cluster.inference.wildcardDomain`, `cluster.inference.wildcardTls` | Standard inference exposed externally only | The wildcard, in `cluster.inference.wildcardTls` |
| Local Certificate Authority (CA) | `tls.caChainPath` (or `controlPlane.tls.caChainPath` / `cluster.tls.caChainPath` per component) | If your certificates are signed by a private or local CA | Set once at the top level, or override per component if different components use different CAs |

## Registry Credentials

Open Source installations set `registry.type: ngc` in the configuration, with no credentials needed. NVIDIA Run:ai Support-Entitled Release installations set the same field, plus an NGC Personal Key to authenticate to the NVIDIA Run:ai container registry (`nvcr.io`) for pulling NVIDIA Run:ai container images. Generate a Personal Key from the [NGC Personal Keys](https://org.ngc.nvidia.com/setup/personal-keys) page.

Personal Keys begin with `nvapi-`. NGC authenticates using the literal username `$oauthtoken` with the key as the password. The Personal Key is used during installation, to pull the NVIDIA Run:ai container images.

For connected installs, NVIDIA Run:ai's telemetry for Support-Entitled Release installations is not associated with your organization by default. To associate it with your NCA ID, generate an [NGC Service API Key](https://docs.nvidia.com/ngc/latest/ngc-catalog-user-guide.html#generating-a-service-api-key). In the **Service** field, select both **NVIDIA Run:ai Self-Hosted** and **NGC Catalog**. The Personal Key is used during installation. The Service API Key is used at runtime, after installation, by the NVIDIA Run:ai control plane to report telemetry. Provide the resulting key in `telemetryService.telemetryAuth` during installation, either directly as `token`, or by referencing a secret you create yourself beforehand:

```bash
kubectl create secret generic <SECRET_NAME> \
  --namespace runai-backend \
  --from-literal=<KEY_NAME>=<SERVICE_API_KEY>
```

<Note>

**Gov-ready**

A Gov-ready installation also pulls the hardened third-party images from Docker Hub, which requires a separate Docker Hub access token and image pull secret. Refer to [Docker Hub Credentials](../../../system-requirements/gov-ready-third-party-services.md#docker-hub-credentials) for how to create them.

</Note>

## Download the Installer

Download the `runai-installer` binary for your operating system and CPU architecture, on each machine you run the installer from. Replace `<VERSION>` with the NVIDIA Run:ai version you are installing, for example `3.0.17`:

```bash
# Linux, AMD64 (x86_64)
curl -LO https://github.com/dsx-ai-factory/run-ai/releases/download/<VERSION>/runai-installer-linux-amd64
chmod +x runai-installer-linux-amd64

# Linux, ARM64
curl -LO https://github.com/dsx-ai-factory/run-ai/releases/download/<VERSION>/runai-installer-linux-arm64
chmod +x runai-installer-linux-arm64

# macOS, AMD64 (Intel)
curl -LO https://github.com/dsx-ai-factory/run-ai/releases/download/<VERSION>/runai-installer-darwin-amd64
chmod +x runai-installer-darwin-amd64

# macOS, ARM64 (Apple Silicon)
curl -LO https://github.com/dsx-ai-factory/run-ai/releases/download/<VERSION>/runai-installer-darwin-arm64
chmod +x runai-installer-darwin-arm64
```

## Next Steps

* [Same-Cluster Installation](same-cluster.md) - The minimum configuration to install the control plane and cluster together, on one Kubernetes cluster.
* [Multi-Cluster Installation](multi-cluster.md) - The minimum configuration to install the control plane and one or more clusters on separate Kubernetes clusters.
* [Configuration Templates](templates.md) - Complete, ready-to-use configurations for specific use cases - use one instead of the minimum configuration if it matches yours.
* [Configuration Reference](../configuration-reference.md) - All supported configuration fields, for building a configuration from scratch.
