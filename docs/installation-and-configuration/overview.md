# Overview

NVIDIA Run:ai is a Kubernetes-native platform for efficient GPU resource management and AI workload orchestration. It runs on customer-managed infrastructure: the hardware, Kubernetes cluster, and network that the organization owns and operates. NVIDIA Run:ai delivers the software that manages and orchestrates the GPUs; the customer provides everything it sits on.

This page explains the platform components, the setup lifecycle, and what to understand before installing - deployment models, shared responsibility, installation options, and installation paths. To prepare your environment and install NVIDIA Run:ai, skip ahead to [Next Steps](#next-steps).

## Platform Components

NVIDIA Run:ai is made up of two components, both installed on a Kubernetes cluster:

* **NVIDIA Run:ai control plane** - Provides centralized resource management, handles workload submission, and delivers cluster monitoring and analytics. The control plane is the management layer; it connects to one or more NVIDIA Run:ai clusters and gives administrators and users a single place to operate across them.
* **NVIDIA Run:ai cluster** - Set of cluster services that provide enhanced scheduling and workload management on each Kubernetes cluster where workloads run. The cluster component extends Kubernetes native scheduling with Run:ai's GPU-aware capabilities.

Every installation includes at least one NVIDIA Run:ai control plane and one NVIDIA Run:ai cluster. In a typical deployment, both are installed on the same Kubernetes cluster. As your environment grows, you can connect additional NVIDIA Run:ai clusters to the same control plane. See [Deployment Model](#deployment-model).

<figure><img src="../assets/self-hosted.png" alt="" width="375"/><figcaption></figcaption></figure>

## Setup Lifecycle

Setting up NVIDIA Run:ai follows a three-phase lifecycle that reflects the natural progression from initial installation through ongoing operations.

### Day 0: Installation

Day 0 gets NVIDIA Run:ai running on your Kubernetes cluster for the first time.

* Choose your deployment model, meet infrastructure requirements, set up networking, enable external access, and review third-party services
* Install NVIDIA Run:ai using the path that fits your environment - Guided Installer, Manual Helm, or Base Command Manager

### Day 1: Configuration

Day 1 covers the infrastructure configuration described in [Configuration](configuration/README.md) that makes NVIDIA Run:ai production-ready for your environment. Which configurations apply depends on your environment and workload needs. They are performed after installation and are typically done once, though they may be revisited as your environment evolves.

* Assign node roles and configure advanced cluster and control plane settings
* Set up cluster authentication
* Configure networking for external access, service mesh, and container access
* Apply security hardening to your cluster
* Set up shared storage
* Connect third-party integrations such as Karpenter

### Day 2: Operations

Day 2 covers the ongoing operational tasks described in [Operations](operations/README.md) that keep NVIDIA Run:ai healthy and your clusters running smoothly.

* View and manage your clusters
* Tune NVIDIA Run:ai system services for scale, high availability, and resource right-sizing as your environment grows
* Perform node maintenance and restore from backup when needed
* Monitor GPU and workload metrics, collect logs, and review event history

### Upgrades and Troubleshooting

Upgrades and troubleshooting are lifecycle events that can occur at any phase and are not bound to a specific day.

* To upgrade NVIDIA Run:ai, see [Plan the Upgrade](upgrade/plan-the-upgrade.md)
* To investigate deployment issues, see [Troubleshooting](troubleshooting/what-gets-deployed.md)

## Understanding Installation

### Deployment Model

The NVIDIA Run:ai control plane and NVIDIA Run:ai clusters can be deployed across your Kubernetes environment in one of two models. The two components can run either on the same Kubernetes cluster or on separate clusters; in both cases, network connectivity between them is required. See [Deployment Model](system-requirements/deployment-model.md) for more details.

* **Same-cluster deployment** - The NVIDIA Run:ai control plane and the first NVIDIA Run:ai cluster are installed on the same Kubernetes cluster. This is the most common deployment model. Infrastructure such as the load balancer, ingress controller, and DNS can be shared between the two components.

<img src="../assets/deployment-same-cluster.svg" alt="Same-cluster deployment: NVIDIA Run:ai Control Plane and Cluster installed on a single Kubernetes cluster" />

* **Multi-cluster deployment** - The NVIDIA Run:ai control plane runs on a dedicated Kubernetes cluster, while one or more NVIDIA Run:ai clusters run on separate Kubernetes clusters. Each cluster requires its own load balancer, ingress controller, FQDNs, and TLS certificates.

<img src="../assets/deployment-multi-cluster.svg" alt="Multi-cluster deployment: NVIDIA Run:ai Control Plane on Kubernetes Cluster 1 managing NVIDIA Run:ai Cluster Services on separate Kubernetes clusters" />

### Shared Responsibility

NVIDIA Run:ai operates under a shared responsibility model. NVIDIA provides and manages the NVIDIA Run:ai software, while the customer provides and operates the underlying infrastructure. This distinction defines which components the customer prepares and operates, and which are delivered and maintained as part of the NVIDIA Run:ai platform.

When using the recommended guided installer, several customer-owned items in this table, including the GPU Operator and Prometheus, are set up by the guided installer. When installing manually via Helm, all customer-owned items must be in place before the Helm charts are applied. Third-party services are the exception: they are deployed by the NVIDIA Run:ai Helm charts in both installation paths and do not need to be installed beforehand.

| Category                  | Components                                                                                                       | Owner    |
| ------------------------- | ---------------------------------------------------------------------------------------------------------------- | -------- |
| Platform software         | NVIDIA Run:ai images and Helm charts                                                                             | NVIDIA   |
| Third-party services      | Keycloak, Traefik, Valkey, NATS, and others                                                                      | Customer |
| Hardware                  | Control plane system nodes, cluster system nodes, worker nodes                                                   | Customer |
| Kubernetes infrastructure | OS, Kubernetes distribution, container runtime, storage class, shared storage                                   | Customer |
| Networking                | Load balancer, ingress controller, FQDN, DNS, TLS certificates                                                   | Customer |
| Prerequisites             | GPU Operator, Network Operator, Prometheus                                                                       | Customer |

Since you own the infrastructure, you can apply OS and Kubernetes security patches and hardening independently, without coordinating with NVIDIA, provided the environment stays within the [support matrix](support-matrix.md) and the permissions, connectivity, and configuration that NVIDIA Run:ai requires remain intact.

### Installation Options

NVIDIA Run:ai is open source software. If you have an NVIDIA Supported license, an NGC API key is required to pull platform images and Helm charts. Both options share the same installation model and support both connected and air-gapped environments.

| License          | Description                                        | NGC API Key Required |
| ---------------- | -------------------------------------------------- | -------------------- |
| Open Source      | Full platform functionality, community-supported.  | No                   |
| NVIDIA Supported | NVIDIA Supported version with compliance features. | Yes                  |

NVIDIA Supported includes a **Gov-ready** variant for customers with strict security and compliance requirements (FIPS-enabled, hardened images). Gov-ready uses the same installation model as standard NVIDIA Supported, but third-party service images are sourced from [Chainguard](https://www.chainguard.dev/) instead of Docker Hub, and customers are responsible for obtaining and providing those images. See [Third-Party Services](system-requirements/third-party-services.md#image-registry) for details.

#### Connected and Air-Gapped Environments

NVIDIA Run:ai can be deployed in either a connected or air-gapped environment, regardless of license. Both use the same [installation paths](#installation-paths) and tooling, but air-gapped environments require additional preparation before installation begins.

| Connectivity | Description                                                                                                                                                                            |
| ------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Connected    | The cluster can reach the internet to pull images and Helm charts directly from NGC and upstream registries during installation.                                                       |
| Air-gapped   | The cluster has no internet access. Before installation, a script or installer flag is used on a connected machine to download and pack all required artifacts, which are then transferred into the air-gapped environment and loaded into a local registry. |

#### Kubernetes and OpenShift

NVIDIA Run:ai runs on both Kubernetes and OpenShift, and the installation model is the same for both. Where they differ is in how certain infrastructure pieces work: OpenShift routes traffic through its built-in router rather than an ingress controller, installs operators like the GPU Operator through OLM rather than Helm, and manages TLS at the router level rather than through Kubernetes secrets. These differences are handled within each installation guide, and the overall flow remains the same.

### Installation Paths

NVIDIA Run:ai supports three installation paths. Choose based on your tooling, environment, and the balance you want between installation convenience and fine-grained control over configuration values.

| Path | Tooling | Best For |
| ---- | ------- | -------- |
| **[Guided Installation](installation/guided-installation/README.md)** (recommended) | Guided Installer CLI (`runai-installer`) with a configuration YAML file | Most deployments |
| **[Manual Helm](installation/install-using-helm/install-control-plane.md)** | NVIDIA Run:ai Helm charts, deployed directly | Organizations managing deployments through their own tooling (for example, Argo CD), or needing granular control over Helm values |
| **[BCM](installation/bcm-install.md)** (NVIDIA Supported only) | NVIDIA Base Command Manager | DGX SuperPOD environments managed by BCM |

### Artifact Sources

During installation, NVIDIA Run:ai pulls platform images, Helm charts, and third-party service images from different sources depending on your license and environment. Understanding these sources helps you plan network access, registry credentials, and air-gapped deployments.

* **NVIDIA Run:ai platform images and Helm charts** are published to NVIDIA NGC. Open Source installations access them without credentials. NVIDIA Supported installations require an NGC API key. Organizations that mirror artifacts to an internal registry can configure the installer or Helm charts to pull from that registry instead.

* **Third-party service images** (PostgreSQL, Keycloak, and others) are pulled from their official upstream registries on Docker Hub, Quay, and the Kubernetes registry (`registry.k8s.io`) for Open Source and NVIDIA Supported installations. For Gov-ready installations, customers are responsible for obtaining and providing hardened, FIPS-enabled images from [Chainguard](https://www.chainguard.dev/). See [Third-Party Services](system-requirements/third-party-services.md#image-registry) for the complete list.

<Warning>
**TODO (Nadav):** With the new approach there will be no pre-built air-gap bundle anymore. Instead, we will supply a script (or an installer flag) that downloads all artifacts from the different registries and packs them on a connected machine at the customer's site, for transfer into the air-gapped environment. This bullet (and the other air-gap-bundle references on this page and in Before You Install) will need updating once the new flow is final.
</Warning>

### What an Installation Includes

An NVIDIA Run:ai installation relies on several categories of software and infrastructure. Some components are installed and managed by NVIDIA Run:ai, while others are customer-provided prerequisites or optional components that enable specific platform capabilities. For a detailed breakdown of ownership responsibilities, see [Shared Responsibility](#shared-responsibility).

* **Infrastructure you provide** - The hardware, operating system, Kubernetes distribution, networking, and storage that host NVIDIA Run:ai. These components must be in place before installation begins. See [System Requirements](system-requirements/README.md).

* **Third-party services** - A set of supporting services deployed as part of the installation, including Keycloak for authentication and identity management, PostgreSQL as the centralized database, Traefik as the API gateway, NATS for asynchronous communication, and others. See [Third-Party Services](system-requirements/third-party-services.md).

* **Prerequisites** - Components that NVIDIA Run:ai requires: GPU Operator (provisions and manages GPUs on worker nodes), Network Operator (required for Multi-Node NVLink workloads), and Prometheus (required for metrics collection and monitoring).

* **Feature-specific components** - Optional components required only for specific NVIDIA Run:ai features. When using the guided installer, these components can be enabled in the configuration file and installed alongside NVIDIA Run:ai.

  | Component | Required for |
  | --------- | ------------ |
  | KNative Serving | Standard inference |
  | Kubeflow Training Operator | Distributed training |
  | MPI Operator | MPI-based distributed training |
  | LeaderWorkerSet (LWS) | Distributed inference |
  | NIM Operator | NVIDIA Inference Microservices |
  | Grove | Multi-component, gang-scheduled workload orchestration (for example, disaggregated Dynamo deployments) |
  | Dynamo | Inference pipeline optimization |

## Next Steps

<Info>

**BCM installations**

If you are deploying NVIDIA Run:ai through NVIDIA Base Command Manager (BCM), do not follow the steps below. BCM manages its own prerequisites, support matrix, and installation process. Refer directly to the [BCM installation instructions](installation/bcm-install.md) for your supported release.

</Info>

* Check the [Support Matrix](support-matrix.md) to know which Kubernetes/OpenShift versions, operating system, GPU and other operators, and third-party service versions NVIDIA Run:ai supports - this determines what to prepare and install in your environment.
* Complete [System Requirements](system-requirements/README.md) to plan your deployment model and prepare your infrastructure, networking, and more, before installation begins.
* Install using the path that fits your environment:
  * [Guided Installation](installation/guided-installation/README.md) - Recommended for most installations.
  * [Manual Helm](installation/install-using-helm/install-control-plane.md) - For organizations managing releases through their own tooling.
