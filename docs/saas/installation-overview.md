# Installation and Configuration

NVIDIA Run:ai SaaS is a cloud-hosted deployment of the NVIDIA Run:ai platform. NVIDIA hosts, manages, and operates the control plane on your behalf — including updates, availability, and infrastructure. Your responsibility is to install the NVIDIA Run:ai cluster component on each Kubernetes cluster where AI workloads will run, and connect it to the NVIDIA-managed control plane.

## System Components

NVIDIA Run:ai is made up of two components:

* **NVIDIA Run:ai control plane** — Provides centralized resource management, workload submission, and cluster monitoring and analytics. In SaaS, this is hosted and operated by NVIDIA. You do not install or manage it.
* **NVIDIA Run:ai cluster** — Provides enhanced scheduling and workload management on each Kubernetes cluster where workloads run. This is installed and managed by you, on your own infrastructure.

You can connect one or more NVIDIA Run:ai clusters to your NVIDIA-managed control plane. Each cluster you connect must meet the prerequisites described in [Before You Install](../installation-and-configuration/installation/before-you-install.md).

## Shared Responsibility

| Category | Components | Owner |
| -------- | ---------- | ----- |
| Control plane software and infrastructure | NVIDIA Run:ai control plane, hosted services | NVIDIA |
| Cluster software | NVIDIA Run:ai cluster images and Helm charts | NVIDIA |
| Hardware | Cluster system nodes, worker nodes | Customer |
| Kubernetes infrastructure | OS, Kubernetes distribution, container runtime, storage class, load balancer, ingress controller | Customer |
| Networking | FQDN, DNS, TLS certificates for cluster endpoints | Customer |
| NVIDIA prerequisites | GPU Operator, Network Operator, DRA Driver | Customer |
| Monitoring | Prometheus | Customer |

## What the Cluster Installation Includes

Installing the NVIDIA Run:ai cluster on your Kubernetes infrastructure involves the following:

* **Infrastructure you provide** — the hardware, OS, Kubernetes distribution, networking, and storage the cluster component runs on. These must be in place before installation begins. See [Before You Install](../installation-and-configuration/installation/before-you-install.md).

* **Bundled third-party services** — a set of services deployed as part of the NVIDIA Run:ai cluster Helm chart. NVIDIA Run:ai installs and manages these services. See [Third-Party Services](../installation-and-configuration/installation/third-party-services.md).

* **NVIDIA components** — GPU Operator is required to provision and manage GPUs on cluster worker nodes. Network Operator is required for advanced networking configurations such as Multi-Node NVLink (MNNVL). These are separate NVIDIA products that NVIDIA Run:ai depends on but does not bundle.

* **Monitoring** — Prometheus is required for metrics collection. It is customer-owned and not managed by NVIDIA Run:ai.

* **Feature-specific components** — some NVIDIA Run:ai features require additional components to be present in the cluster.

  | Component | Required for |
  | --------- | ------------ |
  | KNative Serving | Inference autoscaling |
  | Kubeflow Training Operator | Distributed training |
  | MPI Operator | MPI-based distributed training |
  | LeaderWorkerSet (LWS) | Distributed inference |
  | NIM Operator | NVIDIA Inference Microservices |
  | Grove | Dynamic GPU repartitioning |
  | Dynamo | Inference pipeline optimization |

## Next Steps

* Review the [Support Matrix](../installation-and-configuration/support-matrix.md) to confirm your Kubernetes distribution, OS, and GPU Operator version are supported.
* Complete [Before You Install](../installation-and-configuration/installation/before-you-install.md) to prepare your cluster infrastructure.
* Review [Third-Party Services](../installation-and-configuration/installation/third-party-services.md) to understand the bundled services.
* [Install the cluster](../installation-and-configuration/installation/install-using-helm/helm-install.md) using Helm.
