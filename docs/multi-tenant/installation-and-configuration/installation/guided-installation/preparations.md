---
---

# Preparations

This page describes how to prepare for installing NVIDIA Run:ai in multi-tenant mode using the `runai-installer` CLI.

## System Requirements

Complete the [System Requirements](../../system-requirements/README.md) section before starting. It covers:

* [Infrastructure Requirements](../../system-requirements/infrastructure-requirements.md) - Hardware, operating system, Kubernetes distribution, storage, and permissions required for the control plane's Kubernetes cluster.
* [Networking Setup](../../system-requirements/networking-setup.md) - Load balancer and ingress requirements, the control plane's wildcard FQDN, and its wildcard TLS certificate and private key. The exact requirements depend on your routing configuration.
* [Third-Party Services](../../system-requirements/third-party-services.md) - The third-party services deployed as part of the installation.
* [Metrics Store Requirements](../../system-requirements/metrics-store-requirements/README.md) - The external metrics store the control plane requires.
* [External Access to Critical Services](../../system-requirements/external-access-to-critical-services.md) - The inbound and outbound connectivity required to reach those services and NVIDIA registries.

Your [deployment model](../../system-requirements/deployment-model.md) is always multi-cluster: control plane setup is completed once, and cluster setup is repeated for each tenant cluster.

## Prerequisites and Features

GPU Operator and Prometheus are required for every tenant cluster. Network Operator, and features such as Knative and Kubeflow, apply only if that tenant's environment or planned workload types need them. Their planned workload types may also affect the networking configuration required for the installation, including wildcard domains and TLS certificates. See [Prerequisites and Features](prerequisites-and-features.md) to determine which prerequisites and features a tenant cluster needs and how to configure them.

The installer can set up supported prerequisites and features for you. By default, it uses the latest supported version; to use a different supported version, specify the `version` field for that component according to the versions listed in the [Support Matrix](../../support-matrix.md).

## Domains and TLS Certificates

The domains and certificates you prepared in [FQDN and DNS](../../system-requirements/networking-setup.md#fqdn-and-dns) and [TLS Certificates](../../system-requirements/networking-setup.md#tls-certificates) map to specific `config.yaml` fields. You supply the control plane values when you write your `config.yaml` in [Control Plane Installation](installation.md), and each tenant cluster's values when you [install that cluster](create-a-tenant-and-cluster.md#install-the-cluster).

| Domain | Config Fields | Required When | Certificate |
| --- | --- | --- | --- |
| Control plane domain (wildcard, for example `*.runai.hostorg.com`) | `controlPlane.domain`, `controlPlane.tls` | Always | A wildcard certificate matching the control plane domain |
| Tenant cluster domain (for example, `runai.tenant-a-corp.com`) | `cluster.domain`, `cluster.tls` | Always | Covered by `cluster.tls`, see the next row |
| Workspaces, training, and distributed inference (for example, `*.runai.tenant-a-corp.com`) | None. Do not set a separate domain field. The wildcard is `*.` plus the tenant cluster domain. | Host-based routing (the default). Not needed with path-based routing. | The same `cluster.tls` certificate, which must list both `runai.tenant-a-corp.com` and `*.runai.tenant-a-corp.com` as Subject Alternative Names |
| Standard inference (Knative), for example `*.runai-inference.tenant-a-corp.com` | `cluster.inference.wildcardDomain`, `cluster.inference.wildcardTls` | Standard inference exposed externally only | The wildcard, in `cluster.inference.wildcardTls` |

## Registry Credentials

Open Source installations set `registry.type: ngc` in the configuration, with no credentials needed. NVIDIA Run:ai Support-Entitled Release installations set the same field, plus an NGC Personal Key to authenticate to the NVIDIA Run:ai container registry (`nvcr.io`) for pulling NVIDIA Run:ai container images. Generate a Personal Key from the [NGC Personal Keys](https://org.ngc.nvidia.com/setup/personal-keys) page.

Personal Keys begin with `nvapi-`. NGC authenticates using the literal username `$oauthtoken` with the key as the password. The Personal Key is used during installation, to pull the NVIDIA Run:ai container images.

For connected installs, NVIDIA Run:ai's telemetry for Support-Entitled Release installations is not associated with your organization by default. To associate it with your NCA ID, generate an [NGC Service API Key](https://docs.nvidia.com/ngc/latest/ngc-catalog-user-guide.html#generating-a-service-api-key). In the **Service** field, select both **NVIDIA Run:ai Self-Hosted** and **NGC Catalog**. The Personal Key is used during installation. The Service API Key is used at runtime, after installation, by the NVIDIA Run:ai control plane to report telemetry. Provide the resulting key in `telemetryService.telemetryAuth` during installation, either directly as `token`, or by referencing a secret you create yourself beforehand:

```bash
kubectl create secret generic <SECRET_NAME> \
  --namespace runai-backend \
  --from-literal=<KEY_NAME>=<SERVICE_API_KEY>
```

## Download the Installer

Download the `runai-installer` binary for your operating system and CPU architecture, on the machine you run the installer from. Replace `<VERSION>` with the NVIDIA Run:ai version you are installing, for example `3.0.17`:

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

* [Control Plane Installation](installation.md) - The minimum configuration to install the control plane.
* [API Access Setup](api-access-setup.md) - Configure programmatic API access using service account credentials.
* [Configuration Templates](templates.md) - Complete, ready-to-use configurations for specific use cases - use one instead of the minimum configuration if it matches yours.
* [Configuration Reference](configuration-reference.md) - All supported configuration fields, for building a configuration from scratch.
