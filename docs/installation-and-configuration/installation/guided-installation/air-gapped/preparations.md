---
---

# Preparations

This page describes how to prepare for installing NVIDIA Run:ai using the `runai-installer` CLI in an air-gapped environment, with no outbound internet access. For internet-connected environments, see [Connected Installation](../connected/preparations.md).

## System Requirements

Complete [System Requirements](../../../system-requirements/README.md) before running the installer, for each Kubernetes cluster you install onto - hardware and Kubernetes distribution, networking and DNS, TLS certificates, and the third-party services the installation deploys for you. The installer deploys and configures NVIDIA Run:ai itself; it does not provision this infrastructure.

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

NGC credentials are not needed to build the air-gapped bundle, for either Open Source or Support-Entitled Release - `airgap pack` pulls NVIDIA Run:ai container images and charts without authentication. Registry credentials are only needed for a registry you introduce yourself, such as your own registry hosting an `imageOverrides` source.

Once the bundle is loaded, the `registry` section in your configuration file authenticates to your own private mirror registry (populated from the bundle) using basic authentication (`username` and `password` or `passwordPath`). See [`registry`](../configuration-reference.md#registry) and [`airgap`](../configuration-reference.md#airgap) in the Configuration Reference.

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

## Air-Gapped Bundle

Getting images into your air-gapped network happens in two stages, on two different machines: **pack** the bundle on a connected machine, then **load** it once it reaches your air-gapped network.

### Pack the Bundle

Packing happens on a separate connected machine - any machine with outbound internet access.

On that connected machine, download the `runai-installer` binary and use its `airgap pack` command to build a single tarball containing every image your release needs - NVIDIA Run:ai platform images and third-party service images (PostgreSQL, Keycloak, and others) alike. See [Third-Party Services](../../../system-requirements/third-party-services.md#image-registry).

1. Download the `runai-installer` binary for your operating system and CPU architecture - see [Download the Installer](#download-the-installer) above.
2. Create a pack configuration file (`airgap-pack.yaml`) and run `airgap pack` against it:

   - `settings` - Required. Your target NVIDIA Run:ai version, architecture, and output path.
   - `imageOverrides` - Points specific third-party service images (for example, Keycloak) at your own source instead of the default one. Each entry also accepts `group`, required only when `name` exists in more than one image group - see [`imageOverrides`](../configuration-reference.md#airgap-pack) in the Configuration Reference.
   - `registryAuth` - Authenticates to your own registry hosting an `imageOverrides` source.

   ```bash
   cat > airgap-pack.yaml <<'EOF'
   settings:
     runaiVersion: <version>
     architecture: amd64
     output: /path/to/runai-airgap.tar

   imageOverrides:
     - name: keycloak
       source: <your-registry>/keycloak:<tag>

   registryAuth:
     - registry: <your-registry-host>
       username: <username>
       passwordPath: /path/to/registry-password
   EOF

   ./runai-installer airgap pack \
     --config airgap-pack.yaml
   ```

3. Transfer the resulting tarball into the air-gapped network.

### Load the Bundle

Loading happens once the tarball is in the air-gapped network, on any machine there with access to both the tarball and your internal mirror registry - this can be a different machine than the one you later run `install` from, as long as `install` can reach the same `workingDir` and `valuesDir` afterward.

1. Extract the `runai-installer` binary from the transferred bundle tarball.
2. Create a load configuration file (`airgap-load.yaml`) and run `airgap load` against it:

```bash
cat > airgap-load.yaml <<'EOF'
settings:
  bundlePath: /path/to/runai-airgap.tar
  workingDir: /data/runai-airgap
  valuesDir: /data/runai-airgap

registry:
  url: <your-registry-host>
  username: <username>
  passwordPath: /path/to/registry-password
EOF

./runai-installer airgap load \
  --config airgap-load.yaml
```

This extracts the bundle, pushes every image to your internal mirror registry, and writes `control-plane-airgap-values.yaml` and `cluster-airgap-values.yaml` into `valuesDir` - the values the installer applies to the control plane and cluster charts.

To run the install command, see [Same-Cluster Installation](same-cluster.md) or [Multi-Cluster Installation](multi-cluster.md).

## Prerequisites and Features

GPU Operator and Prometheus are required for every installation. Network Operator, and features such as Knative and Kubeflow, apply only if your environment or planned workload types need them. See [Prerequisites and Features](../prerequisites-and-features.md) for what each one does and when you need it.

The installer does not install prerequisites or features at all in air-gapped environments - you must install them yourself beforehand, directly on each target Kubernetes cluster, using each vendor's own air-gapped installation instructions.

## Next Steps

* [Same-Cluster Installation](same-cluster.md) - The minimum configuration to install the control plane and cluster together, on one Kubernetes cluster.
* [Multi-Cluster Installation](multi-cluster.md) - The minimum configuration to install the control plane and one or more clusters on separate Kubernetes clusters.
* [Configuration Templates](templates.md) - Complete, ready-to-use configurations for specific use cases - use one instead of the minimum configuration if it matches yours.
* [Configuration Reference](../configuration-reference.md) - All supported configuration fields, for building a configuration from scratch.
