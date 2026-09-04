---
---

# Guided Installation

The `runai-installer` CLI is the recommended way to install NVIDIA Run:ai. It reads a configuration YAML file, validates your configuration against the Kubernetes cluster, and installs the required components in the correct order.

## When to Use the Installer

Use the guided installation path when:

- You are deploying a new NVIDIA Run:ai environment.
- You want built-in preflight validation before any changes are applied.
- You prefer a single tool that validates your configuration, runs preflight checks, installs prerequisites, and deploys the NVIDIA Run:ai control plane and cluster.

For environments with an existing GitOps pipeline (Argo CD) or where administrators manage Helm releases directly, see [Manual Installation (Helm)](../install-using-helm/helm-install.md).

## Installer Flow

The `runai-installer` CLI installs NVIDIA Run:ai from a configuration YAML file. Before applying any changes, it validates your configuration, runs preflight checks, installs any required prerequisites, and deploys the NVIDIA Run:ai control plane and cluster components. The flow is the same for same-cluster and multi-cluster installations - in a multi-cluster deployment, the installer runs once per Kubernetes cluster.

The following diagram shows the installer flow - from downloading the `runai-installer` CLI to a verified, workload-ready cluster:

<img src="../../../assets/guided-installation-flow.png" alt="Guided installation flow: download the runai-installer CLI, prepare the config.yaml configuration file, run the install-new command, pass preflight checks, install the prerequisites, control plane, and cluster components, then verify the installation" />

The installation is built from two artifacts: the `runai-installer` binary (available for AMD64 and ARM64) and a declarative `config.yaml` file - the single input that describes what to install and how. At run time, the installer writes detailed logs to a local directory and persists session state in the cluster, so an interrupted installation can be resumed without re-entering configuration values.

The installer isn't a one-time step - rerunning it against an already-installed cluster with an updated `config.yaml` adds any prerequisites or features that were skipped initially, installing only what's missing and leaving the rest of the deployment untouched.

## What the Installer Does

The installer's configuration file covers every deployment combination - same-cluster or multi-cluster, connected or air-gapped, standard Kubernetes or OpenShift - all from one configuration file.

From there, the installer:

* **Configures networking** - Routing (Ingress, Route, or Gateway API) and TLS (your certificates, or self-signed ones it generates) - based on the FQDNs, DNS records, and TLS certificates you prepare beforehand. See [Networking Setup](../../system-requirements/networking-setup.md).
* **Installs prerequisites** - Components NVIDIA Run:ai depends on, such as the GPU Operator. See [Prerequisites and Features](prerequisites-and-features.md#prerequisites).
* **Installs features** - Components that add support for specific workload types, such as distributed training or inference. See [Prerequisites and Features](prerequisites-and-features.md#features).
* **Deploys the NVIDIA Run:ai control plane and/or cluster Helm charts** - Along with the third-party services they depend on, deployed as part of the same chart. Reaching those registries, and NVIDIA's, requires the outbound access you prepare beforehand. See [Third-Party Services](../../system-requirements/third-party-services.md) and [External Access to Critical Services](../../system-requirements/external-access-to-critical-services.md).
* **Handles registry authentication** for pulling NVIDIA Run:ai images and charts - Required for NVIDIA Run:ai Support-Entitled Release installations only. Open Source installations pull without credentials. See [Registry Credentials](same-cluster-installation.md#registry-credentials).

## Next Steps

* [Same-Cluster Installation](same-cluster-installation.md) - Install the control plane and cluster together, on one Kubernetes cluster.
* [Multi-Cluster Installation](multi-cluster-installation.md) - Install the control plane and one or more clusters on separate Kubernetes clusters.
* [Configuration Reference](configuration-reference.md) - Reference for all configuration YAML fields and their accepted values - build your own configuration from scratch.
* [Connected Templates](connected.md) or [Air-Gapped Templates](air-gapped.md) - Complete, ready-to-use configurations for specific use cases.
