# Support Matrix

The support matrix outlines the verified compatibility standards for NVIDIA Run:ai SaaS. To ensure a stable and performant deployment, all infrastructure components, including Kubernetes/OpenShift distributions, NVIDIA Operators, and specialized frameworks, must align with the versions specified below. Use this matrix as a validation checklist prior to performing new installations or upgrades.

## Operator and Framework Versions

<Tabs>
  <Tab title="v3.0 (latest)">
| Component                                                                                                    | Supported Versions   |
| ------------------------------------------------------------------------------------------------------------ | -------------------- |
| [NVIDIA GPU Operator](../installation-and-configuration/installation/install-using-helm/system-requirements.md#nvidia-gpu-operator)                         | 25.10-26.3           |
| [NVIDIA Network Operator](../installation-and-configuration/installation/install-using-helm/system-requirements.md#nvidia-network-operator)                 | 25.10-26.1           |
| [NVIDIA DRA driver](../installation-and-configuration/installation/install-using-helm/system-requirements.md#nvidia-dynamic-resource-allocation-dra-driver) | 25.8 - 25.12         |
| [Prometheus / Kube‑Prometheus Stack](../installation-and-configuration/installation/install-using-helm/system-requirements.md#prometheus)                   | 3.5 / 76.0 and above |
| [Kubeflow Training Operator](../installation-and-configuration/installation/install-using-helm/system-requirements.md#distributed-training)                 | 1.9.2                |
| [MPI Operator](../installation-and-configuration/installation/install-using-helm/system-requirements.md#distributed-training)                               | 0.6.0 or later       |
| [Knative Serving](../installation-and-configuration/installation/install-using-helm/system-requirements.md#inference)                                       | 1.19 – 1.21          |
| [Leader-Worker Set (LWS)](../installation-and-configuration/installation/install-using-helm/system-requirements.md#distributed-inference)                   | 0.7.0 or higher      |
  </Tab>
  <Tab title="v2.26">
| Component                                                                                                    | Supported Versions   |
| ------------------------------------------------------------------------------------------------------------ | -------------------- |
| [NVIDIA GPU Operator](../installation-and-configuration/installation/install-using-helm/system-requirements.md#nvidia-gpu-operator)                         | 25.10-26.3           |
| [NVIDIA Network Operator](../installation-and-configuration/installation/install-using-helm/system-requirements.md#nvidia-network-operator)                 | 25.10-26.1           |
| [NVIDIA DRA driver](../installation-and-configuration/installation/install-using-helm/system-requirements.md#nvidia-dynamic-resource-allocation-dra-driver) | 25.8 - 25.12         |
| [Prometheus / Kube‑Prometheus Stack](../installation-and-configuration/installation/install-using-helm/system-requirements.md#prometheus)                   | 3.5 / 76.0 and above |
| [Kubeflow Training Operator](../installation-and-configuration/installation/install-using-helm/system-requirements.md#distributed-training)                 | 1.9.2                |
| [MPI Operator](../installation-and-configuration/installation/install-using-helm/system-requirements.md#distributed-training)                               | 0.6.0 or later       |
| [Knative Serving](../installation-and-configuration/installation/install-using-helm/system-requirements.md#inference)                                       | 1.19 – 1.21          |
| [Leader-Worker Set (LWS)](../installation-and-configuration/installation/install-using-helm/system-requirements.md#distributed-inference)                   | 0.7.0 or higher      |
  </Tab>
  <Tab title="v2.25">
| Component                                                                                                    | Supported Versions   |
| ------------------------------------------------------------------------------------------------------------ | -------------------- |
| [NVIDIA GPU Operator](../installation-and-configuration/installation/install-using-helm/system-requirements.md#nvidia-gpu-operator)                         | 25.10-26.3           |
| [NVIDIA Network Operator](../installation-and-configuration/installation/install-using-helm/system-requirements.md#nvidia-network-operator)                 | 25.10-26.1           |
| [NVIDIA DRA driver](../installation-and-configuration/installation/install-using-helm/system-requirements.md#nvidia-dynamic-resource-allocation-dra-driver) | 25.8 - 25.12         |
| [Prometheus / Kube‑Prometheus Stack](../installation-and-configuration/installation/install-using-helm/system-requirements.md#prometheus)                   | 3.5 / 76.0 and above |
| [Kubeflow Training Operator](../installation-and-configuration/installation/install-using-helm/system-requirements.md#distributed-training)                 | 1.9.2                |
| [MPI Operator](../installation-and-configuration/installation/install-using-helm/system-requirements.md#distributed-training)                               | 0.6.0 or later       |
| [Knative Serving](../installation-and-configuration/installation/install-using-helm/system-requirements.md#inference)                                       | 1.19 – 1.21          |
| [Leader-Worker Set (LWS)](../installation-and-configuration/installation/install-using-helm/system-requirements.md#distributed-inference)                   | 0.7.0 or higher      |
  </Tab>
  <Tab title="v2.24">
| Component                                                                                                    | Supported Versions   |
| ------------------------------------------------------------------------------------------------------------ | -------------------- |
| [NVIDIA GPU Operator](../installation-and-configuration/installation/install-using-helm/system-requirements.md#nvidia-gpu-operator)                         | 25.3-25.10           |
| [NVIDIA Network Operator](../installation-and-configuration/installation/install-using-helm/system-requirements.md#nvidia-network-operator)                 | 24.4-25.10           |
| [NVIDIA DRA driver](../installation-and-configuration/installation/install-using-helm/system-requirements.md#nvidia-dynamic-resource-allocation-dra-driver) | 25.3-25.8            |
| [Prometheus / Kube‑Prometheus Stack](../installation-and-configuration/installation/install-using-helm/system-requirements.md#prometheus)                   | 3.5 / 76.0 and above |
| [Kubeflow Training Operator](../installation-and-configuration/installation/install-using-helm/system-requirements.md#distributed-training)                 | 1.9.2                |
| [MPI Operator](../installation-and-configuration/installation/install-using-helm/system-requirements.md#distributed-training)                               | 0.6.0 or later       |
| [Knative Serving](../installation-and-configuration/installation/install-using-helm/system-requirements.md#inference)                                       | 1.11 – 1.18          |
| [Leader-Worker Set (LWS)](../installation-and-configuration/installation/install-using-helm/system-requirements.md#distributed-inference)                   | 0.7.0 or higher      |
  </Tab>

</Tabs>

## Supported NVIDIA GPUs

NVIDIA Run:ai is compatible with all Data Center GPUs supported by the NVIDIA GPU Operator. Hardware compatibility is determined by the specific version of the GPU Operator deployed within your cluster.

* Supported operator range - See the [Operator and framework versions](support-matrix.md#operator-and-framework-versions) section to determine the supported GPU Operator version per NVIDIA Run:ai cluster.
* Hardware verification - To confirm if a specific GPU model is supported, please cross-reference your Operator version with the [Supported NVIDIA Data Center GPUs and Systems](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/platform-support.html#supported-nvidia-data-center-gpus-and-systems) documentation.

<Callout intent="info">
* NVIDIA DGX Spark, NVIDIA Jetson and workstations are not supported.
* vGPU is not supported. NVIDIA Run:ai currently supports GPU passthrough only.
* In addition to GPU Operator compatibility, support also depends on the framework being used (for example, CUDA, PyTorch, TensorFlow, or NVIDIA NIM). Before running a workload, verify that the selected framework version supports your target GPU architecture according to the relevant support matrix.
</Callout>

## NVIDIA Run:ai Compatible Distributions

NVIDIA Run:ai supports a wide range of Kubernetes distributions across on-premises, hybrid, and public cloud environments. Use the table below to verify version requirements for your specific platform.

<Tabs>
  <Tab title="v3.0 (latest)">
| Orchestration Platform                 | Versions                         | Engine     | x86       | ARM       |
| -------------------------------------- | -------------------------------- | ---------- | --------- | --------- |
| Upstream Kubernetes                    | 1.34-1.36                        | Containerd | Supported | Supported |
| Red Hat OpenShift                      | 4.18-4.21                        | CRI-O      | Supported | Supported |
| Amazon Elastic Kubernetes Engine (EKS) | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Google Kubernetes Engine (GKE)         | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Azure Kubernetes Service (AKS)         | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Oracle Kubernetes Engine (OKE)         | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Rancher Kubernetes Engine 2 (RKE2)     | Based on the upstream Kubernetes | Containerd | Supported | Supported |
  </Tab>
  <Tab title="v2.26">
| Orchestration Platform                 | Versions                         | Engine     | x86       | ARM       |
| -------------------------------------- | -------------------------------- | ---------- | --------- | --------- |
| Upstream Kubernetes                    | 1.34-1.36                        | Containerd | Supported | Supported |
| Red Hat OpenShift                      | 4.18-4.21                        | CRI-O      | Supported | Supported |
| Amazon Elastic Kubernetes Engine (EKS) | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Google Kubernetes Engine (GKE)         | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Azure Kubernetes Service (AKS)         | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Oracle Kubernetes Engine (OKE)         | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Rancher Kubernetes Engine 2 (RKE2)     | Based on the upstream Kubernetes | Containerd | Supported | Supported |
  </Tab>
  <Tab title="v2.25">
| Orchestration Platform                 | Versions                         | Engine     | x86       | ARM       |
| -------------------------------------- | -------------------------------- | ---------- | --------- | --------- |
| Upstream Kubernetes                    | 1.33-1.35                        | Containerd | Supported | Supported |
| Red Hat OpenShift                      | 4.18-4.21                        | CRI-O      | Supported | Supported |
| Amazon Elastic Kubernetes Engine (EKS) | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Google Kubernetes Engine (GKE)         | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Azure Kubernetes Service (AKS)         | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Oracle Kubernetes Engine (OKE)         | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Rancher Kubernetes Engine 2 (RKE2)     | Based on the upstream Kubernetes | Containerd | Supported | Supported |
  </Tab>
  <Tab title="v2.24">
| Orchestration Platform                 | Versions                         | Engine     | x86       | ARM       |
| -------------------------------------- | -------------------------------- | ---------- | --------- | --------- |
| Upstream Kubernetes                    | 1.33-1.35                        | Containerd | Supported | Supported |
| Red Hat OpenShift                      | 4.17-4.20                        | CRI-O      | Supported | Supported |
| Amazon Elastic Kubernetes Engine (EKS) | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Google Kubernetes Engine (GKE)         | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Azure Kubernetes Service (AKS)         | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Oracle Kubernetes Engine (OKE)         | Based on the upstream Kubernetes | Containerd | Supported | Supported |
| Rancher Kubernetes Engine 2 (RKE2)     | Based on the upstream Kubernetes | Containerd | Supported | Supported |
  </Tab>

</Tabs>

For existing Kubernetes clusters, see the following Kubernetes version support matrix for the latest NVIDIA Run:ai cluster releases:

| NVIDIA Run:ai version | Supported Kubernetes versions | Supported OpenShift versions |
| --------------------- | ----------------------------- | ---------------------------- |
| 2.26 (latest)         | 1.34 to 1.36                  | 4.18 to 4.21                 |
| 2.25                  | 1.33 to 1.35                  | 4.18 to 4.21                 |
| 2.24                  | 1.33 to 1.35                  | 4.17 to 4.20                 |
| 2.23                  | 1.31 to 1.34                  | 4.16 to 4.19                 |
| 2.22                  | 1.31 to 1.33                  | 4.15 to 4.19                 |

For information on supported versions of managed Kubernetes, it's important to consult the release notes provided by your Kubernetes service provider. There, you can confirm the specific version of the underlying Kubernetes platform supported by the provider, ensuring compatibility with NVIDIA Run:ai. For an up-to-date end-of-life statement see [Kubernetes Release History](https://kubernetes.io/releases/) or [OpenShift Container Platform Life Cycle Policy](https://access.redhat.com/support/policy/updates/openshift).

## Partner-Compatible Distributions

The following Kubernetes distributions are **partner-compatible**. They are tested and validated by the partner, who is responsible for maintaining compatibility with NVIDIA Run:ai:

| Kubernetes distribution                                                                        | NVIDIA Run:ai version               | Supported Kubernetes versions                     |
| ---------------------------------------------------------------------------------------------- | ----------------------------------- | ------------------------------------------------- |
| Crusoe Managed Kubernetes (CMK)                                                                | 2.22                                | 1.33                                              |
| [Mirantis k0rdent](https://catalog.k0rdent.io/v1.7.0/apps/runai-cp/)                           | <ul><li>2.22</li><li>2.23</li></ul> | <ul><li>1.32-1.33</li><li>1.33-1.34</li></ul>     |
| [Rafay platform](https://docs.rafay.co/aiml/app_marketplace/helm_app/apps/runai/requirements/) | <ul><li>2.23</li><li>2.24</li></ul> | <ul><li>1.33 - 1.34</li><li>1.33 - 1.35</li></ul> |
| [vCluster](https://www.vcluster.com/docs/platform/integrations/certified-stacks/runai)         | 2.24                                | 1.34                                              |
| VMware vSphere Kubernetes Service (VKS)                                                        | <ul><li>2.22</li><li>2.24</li></ul> | <ul><li>1.33</li><li>1.35</li></ul>               |
