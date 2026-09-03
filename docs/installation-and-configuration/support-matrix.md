# Support Matrix

The support matrix outlines the verified compatibility standards for NVIDIA Run:ai v3.0, and applies whether you install using the [Guided Installation](installation/guided-installation/README.md) or [Manual Helm](installation/install-using-helm/install-control-plane.md) path. To ensure a stable and performant deployment, all infrastructure components, including Kubernetes/OpenShift distributions, NVIDIA Operators, and specialized frameworks, must align with the versions specified below. Use this matrix as a validation checklist prior to performing new installations or upgrades. Once you've confirmed your versions, continue to [System Requirements](system-requirements/README.md) to prepare your cluster infrastructure.

## Operator and Framework Versions <a href="#operator-and-framework-versions" id="operator-and-framework-versions"></a>

### Prerequisites

| Component                            | Supported Versions   |
| ------------------------------------- | -------------------- |
| NVIDIA GPU Operator                   | 25.10 - 26.3         |
| NVIDIA DRA driver                     | 25.8 - 25.12         |
| NVIDIA Network Operator               | 25.10 - 26.1         |
| Prometheus / Kube‑Prometheus Stack    | 3.5 / 76.0 and above |

### Features

| Component                            | Supported Versions   |
| ------------------------------------- | -------------------- |
| Knative Serving                       | 1.19 - 1.21          |
| Kubeflow Training Operator            | 1.9.2                |
| MPI Operator                          | 0.6.0 or later       |
| Leader-Worker Set (LWS)               | 0.7.0 or higher      |
| NIM Operator                          | 3.0.x                |
| Grove                                  | TBD                  |
| Dynamo Operator                        | 1.2.0                |

## Third-Party Service Versions

Each NVIDIA Run:ai release is tested and validated against a single minor version of each dependency.

| Component    | Deployed On   | Supported Version |
| ------------ | ------------- | ----------------- |
| Keycloak     | Control plane | 26.6              |
| PostgreSQL   | Control plane | 16                |
| Traefik      | Control plane | 3.7               |
| NATS         | Control plane | 2.11.8            |
| Thanos       | Control plane | 0.38.0            |
| Valkey       | Cluster       | 9.1.1             |
| oauth2-proxy | Cluster       | v7.15.3           |
| git-sync     | Cluster       | v4.7.1            |

## Supported NVIDIA GPUs

NVIDIA Run:ai v3.0 is compatible with all Data Center GPUs supported by the NVIDIA GPU Operator. Hardware compatibility is determined by the specific version of the GPU Operator deployed within your cluster.

* Supported operator range - NVIDIA Run:ai v3.0 supports GPU Operator versions 25.10 through 26.3.
* Hardware verification - To confirm if a specific GPU model is supported, please cross-reference your Operator version with the [Supported NVIDIA Data Center GPUs and Systems](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/platform-support.html#supported-nvidia-data-center-gpus-and-systems) documentation.

<Info>

**Note**

* NVIDIA DGX Spark, NVIDIA Jetson and workstations are not supported.
* vGPU is not supported. NVIDIA Run:ai currently supports GPU passthrough only.
* In addition to GPU Operator compatibility, support also depends on the framework being used (for example, CUDA, PyTorch, TensorFlow, or NVIDIA NIM). Before running a workload, verify that the selected framework version supports your target GPU architecture according to the relevant support matrix.

</Info>

## NVIDIA Run:ai Compatible Distributions

NVIDIA Run:ai v3.0 supports a wide range of Kubernetes distributions across on-premises, hybrid, and public cloud environments. Use the table below to verify version requirements for your specific platform.

<Info>

**Note**

For a per-version Kubernetes and OpenShift compatibility matrix, see [Kubernetes and OpenShift Version Alignment](upgrade/helm.md#kubernetes-and-openshift-version-alignment). Review this before upgrading to ensure your Kubernetes or OpenShift distribution version is compatible with your target NVIDIA Run:ai version.

</Info>

| Orchestration Platform                 | Versions                         | Engine     | x86       | ARM       |
| -------------------------------------- | -------------------------------- | ---------- | --------- | --------- |
| Upstream Kubernetes                    | 1.34 - 1.36                      | Containerd | Supported | Supported |
| Red Hat OpenShift                      | 4.18 - 4.21                      | CRI-O      | Supported | Supported |
| Amazon Elastic Kubernetes Engine (EKS) | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Google Kubernetes Engine (GKE)         | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Azure Kubernetes Service (AKS)         | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Oracle Kubernetes Engine (OKE)         | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Rancher Kubernetes Engine 2 (RKE2)     | Based on the upstream Kubernetes | Containerd | Supported | Supported |

## Partner-Compatible Distributions

The following Kubernetes distributions are **partner-compatible**. They are tested and validated by the partner, who is responsible for maintaining compatibility with NVIDIA Run:ai:

| Kubernetes distribution                                                                        | NVIDIA Run:ai version               | Supported Kubernetes versions                     |
| ---------------------------------------------------------------------------------------------- | ----------------------------------- | ------------------------------------------------- |
| Crusoe Managed Kubernetes (CMK)                                                                | 2.22              | 1.33                    |
| [Mirantis k0rdent](https://catalog.k0rdent.io/v1.7.0/apps/runai-cp/)                           | 2.22<br>2.23      | 1.32-1.33<br>1.33-1.34  |
| [Rafay platform](https://docs.rafay.co/aiml/app_marketplace/helm_app/apps/runai/requirements/) | 2.23<br>2.24      | 1.33-1.34<br>1.33-1.35  |
| [Spectro Cloud](https://docs.palette-ai.com/docs/reference/featured-profile-bundles/runai/)    | 2.24              | 1.34                    |
| [vCluster](https://www.vcluster.com/docs/platform/integrations/certified-stacks/runai)         | 2.24              | 1.34                    |
| VMware vSphere Kubernetes Service (VKS)                                                        | 2.22<br>2.24      | 1.33<br>1.35            |
