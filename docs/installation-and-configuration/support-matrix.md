# Support Matrix

The support matrix outlines the verified compatibility standards for NVIDIA Run:ai v3.0, and applies whether you install using the [Guided Installation](installation/guided-installation/README.md) or [Manual Helm](installation/install-using-helm/preparations.md) path. To ensure a stable and performant deployment, all infrastructure components, including Kubernetes/OpenShift distributions, NVIDIA Operators, and specialized frameworks, must align with the versions specified below. Use this matrix as a validation checklist prior to performing new installations or upgrades. Once you've confirmed your versions, continue to [System Requirements](system-requirements/README.md) to prepare your cluster infrastructure.

## Operator and Framework Versions <a href="#operator-and-framework-versions" id="operator-and-framework-versions"></a>

### Prerequisites

| Component                            | Supported Versions   |
| ------------------------------------- | -------------------- |
| NVIDIA GPU Operator<sup>1</sup>       | 25.10 - 26.7.1        |
| NVIDIA DRA driver<sup>2</sup>         | 25.8 - 25.12          |
| NVIDIA Network Operator               | 25.10 - 26.7         |
| Prometheus / Kube‑Prometheus Stack    | 3.13.1 / 88.6.5 and above |

<sup>1</sup> GPU Operator 26.7.1 is required to use GPU fractions, alongside CUDA driver r615 or later. GPU Operator 26.7.1 does not default to CUDA driver r615, so the driver version must be set explicitly. See [GPU Memory Fractions](../platform-management/runai-scheduler/resource-optimization/memory-fractions.md#prerequisites).

<sup>2</sup> Applies to a separately installed DRA driver, used when the GPU Operator runs in `ClusterPolicy` mode (the default). With GPU Operator 26.7 in `GPUCluster` mode, the GPU Operator deploys the DRA driver itself. See [NVIDIA Dynamic Resource Allocation (DRA) Driver](installation/guided-installation/prerequisites-and-features.md#nvidia-dynamic-resource-allocation-dra-driver).

### Features

| Component                            | Supported Versions   |
| ------------------------------------- | -------------------- |
| Knative Serving<sup>3</sup>                | 1.21 - 1.23          |
| Kubeflow Training Operator            | 1.9.4                |
| MPI Operator                          | 0.6.0 - 0.8.2         |
| Leader-Worker Set (LWS)               | 0.7.0 - 0.10.0       |
| NIM Operator                          | 3.1.x                |
| Dynamo Operator                        | 1.5.x                |

<sup>3</sup> On OpenShift, Knative is installed through the OpenShift Serverless operator, which supports Knative 1.17 only.

## Third-Party Service Versions

Each NVIDIA Run:ai release is tested and validated against a single minor version of each dependency.

| Component    | Deployed On                | Supported Version |
| ------------ | --------------------------- | ----------------- |
| Keycloak     | NVIDIA Run:ai control plane | 26.6              |
| PostgreSQL   | NVIDIA Run:ai control plane | 16                |
| Traefik      | NVIDIA Run:ai control plane | 3.7               |
| NATS         | NVIDIA Run:ai control plane | 2.11.8            |
| Thanos       | NVIDIA Run:ai control plane | 0.38.0            |
| Valkey       | NVIDIA Run:ai cluster       | 9.0.4             |
| oauth2-proxy | NVIDIA Run:ai cluster       | 7.15.4            |
| git-sync     | NVIDIA Run:ai cluster       | 4.7.0             |

## Supported NVIDIA GPUs

NVIDIA Run:ai v3.0 is compatible with all Data Center GPUs supported by the NVIDIA GPU Operator. Hardware compatibility is determined by the specific version of the GPU Operator deployed within your cluster.

* Supported operator range - NVIDIA Run:ai v3.0 supports GPU Operator versions 25.10 through 26.7.1.
* Hardware verification - To confirm if a specific GPU model is supported, please cross-reference your Operator version with the [Supported NVIDIA Data Center GPUs and Systems](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/platform-support.html#supported-nvidia-data-center-gpus-and-systems) documentation.

<Note>

**Note**

* NVIDIA DGX Spark, NVIDIA Jetson and workstations are not supported.
* vGPU is not supported. NVIDIA Run:ai currently supports GPU passthrough only.
* In addition to GPU Operator compatibility, support also depends on the framework being used (for example, CUDA, PyTorch, TensorFlow, or NVIDIA NIM). Before running a workload, verify that the selected framework version supports your target GPU architecture according to the relevant support matrix.

</Note>

## NVIDIA Run:ai Compatible Distributions

NVIDIA Run:ai v3.0 supports a wide range of Kubernetes distributions across on-premises, hybrid, and public cloud environments. Use the table below to verify version requirements for your specific platform.

<Note>

**Note**

For a per-version Kubernetes and OpenShift compatibility matrix, see [Kubernetes and OpenShift Compatibility](upgrade/plan-the-upgrade.md#kubernetes-and-openshift-compatibility). Review this before upgrading to ensure your Kubernetes or OpenShift distribution version is compatible with your target NVIDIA Run:ai version.

</Note>

| Orchestration Platform                 | Versions                         | Engine     | x86       | ARM       |
| -------------------------------------- | -------------------------------- | ---------- | --------- | --------- |
| Upstream Kubernetes                    | 1.35 - 1.37                      | Containerd | Supported | Supported |
| Red Hat OpenShift                      | 4.20 - 4.22                      | CRI-O      | Supported | Supported |
| Amazon Elastic Kubernetes Service (EKS) | Based on the upstream Kubernetes | Containerd | Supported | Supported |
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
| [OpenNebula Kubernetes Service (OneKS)](https://docs.opennebula.io/7.4/platform_services/oneks/catalog/nvidia_runai_platform/) | 2.25              | 1.34                    |
| [Rafay platform](https://docs.rafay.co/aiml/app_marketplace/helm_app/apps/runai/requirements/) | 2.23<br>2.24      | 1.33-1.34<br>1.33-1.35  |
| [Spectro Cloud](https://docs.palette-ai.com/docs/reference/featured-profile-bundles/runai/)    | 2.24              | 1.34                    |
| [vCluster](https://www.vcluster.com/docs/platform/integrations/certified-stacks/runai)         | 2.24              | 1.34                    |
| VMware vSphere Kubernetes Service (VKS)                                                        | 2.22<br>2.24      | 1.33<br>1.35            |
