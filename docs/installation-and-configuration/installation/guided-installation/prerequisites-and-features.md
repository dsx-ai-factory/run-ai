---
---

# Prerequisites and Features

Prerequisites and features are additional infrastructure required by the NVIDIA Run:ai cluster. Prerequisites are components every NVIDIA Run:ai deployment depends on, such as the GPU Operator. Features extend NVIDIA Run:ai to support specific workload types, such as distributed training or inference, and only need to be enabled for the workloads you plan to run. The `runai-installer` CLI installs and configures these for you, alongside NVIDIA Run:ai itself.

<Note>

**Air-Gapped Installations**

The installer does not support installing prerequisites or features in air-gapped mode. Install each component yourself beforehand, following the instructions below.

</Note>

| Component | Type | Required for |
|---|---|---|
| GPU Operator | Prerequisite | Every installation |
| Network Operator | Prerequisite | RDMA or Multi-Node NVLink (MNNVL)-capable nodes |
| DRA Driver | Prerequisite | Multi-Node NVLink (e.g. GB200) |
| Prometheus | Prerequisite | Every installation |
| Kubeflow Training Operator | Feature | Distributed training |
| MPI Operator | Feature | MPI-based distributed training |
| Knative Serving | Feature | Standard (single-node) inference |
| LeaderWorkerSet (LWS) | Feature | Distributed (multi-node) inference |
| NIM Operator | Feature | NVIDIA Inference Microservices |
| Grove | Feature | Multi-component, gang-scheduled workload orchestration (for example, disaggregated Dynamo deployments) |
| Dynamo | Feature | Inference pipeline optimization |

## Custom Helm Values

For settings not exposed directly by the installer, use `customHelmValues` to merge custom Helm values into the relevant chart. `controlPlane.customHelmValues` and `cluster.customHelmValues` accept any key documented in [Advanced Control Plane Configurations](../../configuration/control-plane-config.md) and [Advanced Cluster Configurations](../../configuration/cluster-config.md) respectively - for example, `GPUNetworkAccelerationEnabled` for Multi-Node NVLink.

Each prerequisite or feature's `customHelmValues` accepts keys from that component's own upstream Helm chart (GPU Operator, Prometheus, Knative, and so on) - see that component's own chart documentation, linked in its section below, for available values.

<Note>

**Note**

- `kubeflow` and `mpiOperator` install from pinned manifests rather than a Helm chart, so `customHelmValues` is rejected for both.
- On OpenShift, `customHelmValues` is rejected for any component that installs via OperatorHub (OLM) instead of Helm.

</Note>

## Prerequisites

Each prerequisite is configured under `prerequisites` in the configuration file, with a `mode` field controlling how the installer handles it.

* `install` (default) - Installs and configures the component for you.
* `skip` - Does not install or check for the component. Use this when the component is already installed on your cluster, or when the prerequisite does not apply to your environment (for example, the Network Operator on a cluster without RDMA or Multi-Node NVLink nodes).

<Note>

**Note**

For air-gapped installations, prerequisites default to `mode: install`, so set each one to `mode: skip` (or `prerequisites.skip`) explicitly - validation fails otherwise.

</Note>

Each prerequisite also accepts its own `version` field - omit it to install the latest version, or see [Support Matrix](../../support-matrix.md) for the supported version range if you want to set it explicitly.

### NVIDIA GPU Operator

The [NVIDIA GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/index.html) provisions and manages GPUs on worker nodes, and is required for the NVIDIA Run:ai cluster to run AI workloads.

<Note>

**Note**

For [Multi-Node NVLink](../../../platform-management/aiinitiatives/resources/using-gb200.md) support (e.g. GB200), GPU Operator 25.3 and above is required.

</Note>

<Note>

**Note**

To use GPU fractions (memory fractions, dynamic memory fractions, or compute modes), the following are required:

* GPU Operator 26.7.1 or later is required. The installer installs this version automatically, so `prerequisites.gpuOperator.version` does not need to be set.
* CUDA driver r615 or later is required. GPU Operator 26.7.1 does not default to it, so set the driver version explicitly rather than relying on the default - via `prerequisites.gpuOperator.customHelmValues.driver.version` on Kubernetes, or `prerequisites.gpuOperator.openshift.customResource.driver.version` on OpenShift.
* The NRI plugin must be disabled. Set `cdi.nriPluginEnabled` to `false` - via `prerequisites.gpuOperator.customHelmValues.cdi.nriPluginEnabled` on Kubernetes, or `prerequisites.gpuOperator.openshift.customResource.cdi.nriPluginEnabled` on OpenShift. GPU memory fractions do not function while the NRI plugin is enabled.

See [GPU Memory Fractions](../../../platform-management/runai-scheduler/resource-optimization/memory-fractions.md#prerequisites) for the full requirement set.

</Note>

* For air-gapped installation, follow the instructions in [Install NVIDIA GPU Operator in Air-Gapped Environments](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/install-gpu-operator-air-gapped.html).
* For Gov-ready installations, follow the instructions in [Install the Gov-Ready GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/install-gpu-operator-gov-ready.html).
* The installer uses the default namespace for your platform - `gpu-operator` on Kubernetes, `nvidia-gpu-operator` on OpenShift. Otherwise, you must specify the target namespace using the flag `clusterConfig.global.nvidiaDcgmExporter.namespace` as described in [Advanced cluster configurations](../../configuration/cluster-config.md).
* NVIDIA drivers may already be installed on the nodes. In such cases, use the NVIDIA GPU Operator flags `--set driver.enabled=false`. [DGX OS](https://docs.nvidia.com/dgx/dgx-os-6-user-guide/release_notes.html) is one such example as it comes bundled with NVIDIA Drivers.
* For distribution-specific requirements and additional instructions, see the sections below:

<Accordion title="OpenShift Container Platform (OCP)">

**OpenShift Container Platform (OCP)**

The Node Feature Discovery (NFD) Operator is a prerequisite for the NVIDIA GPU Operator in OpenShift. For connected installs, `runai-installer` installs the NFD Operator automatically as part of the GPU Operator prerequisite. For air-gapped installs, install the NFD Operator yourself beforehand, using the Red Hat OperatorHub catalog in the OpenShift Container Platform web console. For more information, see [Installing the Node Feature Discovery (NFD) Operator](https://docs.nvidia.com/datacenter/cloud-native/openshift/latest/install-nfd.html).

The installer installs the GPU Operator into the `nvidia-gpu-operator` namespace via OperatorHub (OLM), rather than the `gpu-operator` namespace used on Kubernetes.

The installer then creates a `ClusterPolicy`, so the GPU Operator runs in `ClusterPolicy` mode and does not deploy the DRA driver. For Multi-Node NVLink, refer to [NVIDIA Dynamic Resource Allocation (DRA) Driver](#nvidia-dynamic-resource-allocation-dra-driver).

</Accordion>

<Accordion title="Elastic Kubernetes Service (EKS)">

**Elastic Kubernetes Service (EKS)**

* When setting up the cluster, do **not** install the NVIDIA device plug-in (we want the NVIDIA GPU Operator to install it instead).
* When using the [eksctl](https://eksctl.io/) tool to create a cluster, use the flag `--install-nvidia-plugin=false` to disable the installation.

For GPU nodes, EKS uses an AMI which already contains the NVIDIA drivers. As such, you must use the GPU Operator flags: `--set driver.enabled=false.`

</Accordion>

<Accordion title="Google Kubernetes Engine (GKE)">

**Google Kubernetes Engine (GKE)**

Before installing the GPU Operator:

1.  Create the `gpu-operator` namespace by running:

    ```bash
    kubectl create ns gpu-operator
    ```
2.  Create the following file:

    ```yaml
    #resourcequota.yaml

    apiVersion: v1
    kind: ResourceQuota
    metadata:
    name: gcp-critical-pods
    namespace: gpu-operator
    spec:
    scopeSelector:
        matchExpressions:
        - operator: In
        scopeName: PriorityClass
        values:
        - system-node-critical
        - system-cluster-critical
    ```
3.  Run:

    ```bash
    kubectl apply -f resourcequota.yaml
    ```

</Accordion>

<Accordion title="Rancher Kubernetes Engine 2 (RKE2)">

**Rancher Kubernetes Engine 2 (RKE2)**

Before installing the GPU Operator, verify the [host OS requirements](https://docs.rke2.io/add-ons/gpu_operators?GPUoperator=v25.3.x#host-os-requirements) are met. Then, install the [operator](https://docs.rke2.io/add-ons/gpu_operators#operator-installation).

When installing GPU Operator v25.3, update the Helm values file as follows:

```yaml
apiVersion: helm.cattle.io/v1
kind: HelmChart
metadata:
  name: gpu-operator
  namespace: kube-system
spec:
  repo: https://helm.ngc.nvidia.com/nvidia
  chart: gpu-operator
  version: v25.3.4
  targetNamespace: gpu-operator
  createNamespace: true
  valuesContent: |-
    toolkit:
      env:
      - name: CONTAINERD_SOCKET
        value: /run/k3s/containerd/containerd.sock
```

</Accordion>

<Accordion title="Oracle Kubernetes Engine (OKE)">

**Oracle Kubernetes Engine (OKE)**

* During cluster setup, [create a node pool](https://docs.oracle.com/en-us/iaas/tools/python/latest/api/container_engine/models/oci.container_engine.models.NodePool.html#oci.container_engine.models.NodePool.initial_node_labels), and set `initial_node_labels` to include `oci.oraclecloud.com/disable-gpu-device-plugin=true` which disables the NVIDIA GPU device plugin.
* For GPU nodes, OKE defaults to Oracle Linux, which is incompatible with NVIDIA drivers. To resolve this, use a custom Ubuntu image instead.

</Accordion>

For troubleshooting information, see the [NVIDIA GPU Operator Troubleshooting Guide](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/troubleshooting.html).

### NVIDIA Network Operator

When deploying on clusters with RDMA or Multi-Node NVLink-capable nodes (for example B200, GB200), the NVIDIA Network Operator is required to enable high-performance networking features such as GPUDirect RDMA in Kubernetes.

The Network Operator works alongside the NVIDIA GPU Operator to provide:

* NVIDIA networking drivers for advanced network capabilities.
* Kubernetes device plugins to expose high-speed network hardware to workloads.
* Secondary network components to support network-intensive applications.

With `prerequisites.networkOperator.mode: install`, the installer does the following:

- **Kubernetes** - Installs the Network Operator Helm chart.
- **OpenShift** - Installs the Network Operator via OperatorHub (OLM), along with the SR-IOV Network Operator and the Node Feature Discovery (NFD) Operator, if not already present.

When installing the Network Operator:

* To configure SR-IOV InfiniBand support on Kubernetes, set the required values via `customHelmValues` - see [Network Operator Deployment with an SR-IOV InfiniBand Network](https://docs.nvidia.com/networking/display/kubernetes2440/getting-started-kubernetes.html#network-operator-deployment-with-an-sr-iov-infiniband-network) for which values to set. This does not apply to OpenShift - configure SR-IOV support via the SR-IOV Network Operator directly.
* For air-gapped installation, follow the instructions in [Network Operator Deployment in an Air-gapped Environment](https://docs.nvidia.com/networking/display/kubernetes2670/advanced/proxy-airgapped.html#network-operator-deployment-in-an-air-gapped-environment).
* For Gov-ready installations, follow the instructions in [Install the Gov-Ready Network Operator](https://docs.nvidia.com/networking/display/kubernetes2670/install-network-operator-gov-ready.html).

### NVIDIA Dynamic Resource Allocation (DRA) Driver

When deploying on clusters with Multi-Node NVLink (e.g. GB200), the [NVIDIA DRA driver](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/26.7/dra-intro-install.html) is required. For air-gapped installations, follow the air-gapped instructions in the [NVIDIA GPU Operator](#nvidia-gpu-operator) section.

How the DRA driver is installed depends on the GPU Operator mode. A cluster runs in one mode only:

* **`ClusterPolicy` mode (default)** - GPUs are exposed through the NVIDIA device plugin. Install the DRA driver separately. Refer to [Support Matrix](../../support-matrix.md) for the supported DRA driver versions.
* **`GPUCluster` mode (GPU Operator 26.7, new installations only)** - The GPU Operator deploys the DRA driver, which replaces the device plugin.

For details, refer to [NVIDIA DRA driver](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/26.7/dra-intro-install.html) and, on OpenShift, [Installing the NVIDIA GPU Operator with DRA on OpenShift](https://docs.nvidia.com/datacenter/cloud-native/openshift/26.7/dra-gpu-ocp.html).

By default, the installer installs the GPU Operator in `ClusterPolicy` mode and does not install the DRA driver. To install the GPU Operator in `GPUCluster` mode:

* **Kubernetes** - Set the following values in `prerequisites.gpuOperator.customHelmValues`. Use flat dotted keys, as shown - nested YAML is not supported. If the NVIDIA driver is pre-installed on the nodes, replace `driver.nvidiaDriverCRD.enabled: true` with `driver.enabled: false`.

  ```yaml
  prerequisites:
    gpuOperator:
      mode: install
      customHelmValues:
        gpuCluster.deployCR: true
        clusterPolicy.deployCR: false
        driver.nvidiaDriverCRD.enabled: true
  ```

* **OpenShift** - The installer always creates a `ClusterPolicy`. Set `prerequisites.gpuOperator.mode` to `skip` and install the GPU Operator yourself, following [Installing the NVIDIA GPU Operator with DRA on OpenShift](https://docs.nvidia.com/datacenter/cloud-native/openshift/26.7/dra-gpu-ocp.html).

<Warning>

**Important**

NVIDIA Run:ai limitations in `GPUCluster` mode:

* GPU fractions are not supported.
* NVIDIA Run:ai shows each node's GPU model, count, and memory only when GPU Feature Discovery is deployed separately. See [Deploying gpu-feature-discovery in standalone mode](https://github.com/NVIDIA/k8s-device-plugin#deploying-gpu-feature-discovery-in-standalone-mode).
* Workloads that request `nvidia.com/gpu` require [extended-resource translation](../../../platform-management/runai-scheduler/scheduling/concepts-and-principles.md#extended-resource-translation), which requires Kubernetes 1.37 or later.

</Warning>

After the DRA driver is installed, update the cluster configuration using the `anyworkload-controller.GPUNetworkAccelerationEnabled` flag to enable GPU network acceleration. This triggers an update of the NVIDIA Run:ai `anyworkload-controller` deployment and restarts the controller. Set it via `cluster.customHelmValues` in the installer's configuration file - see [Custom Helm Values](#custom-helm-values) above - or, outside the guided installer, via Helm or `runaiconfig` as described in [Advanced cluster configurations](../../configuration/cluster-config.md).

### Prometheus

<Note>

**Note**

Installing Prometheus applies to Kubernetes only.

</Note>

NVIDIA Run:ai requires Prometheus on the Kubernetes cluster. The setup has two parts:

- **Prerequisite: Kube-Prometheus Stack** - On Kubernetes, install the Kube-Prometheus Stack to provide the Prometheus Operator and the Prometheus Custom Resource Definitions (CRDs). OpenShift includes equivalent components through its built-in Cluster Monitoring Operator, so no install is required. For RKE2, see [Enable Monitoring](https://ranchermanager.docs.rancher.com/how-to-guides/advanced-user-guides/monitoring-alerting-guides/enable-monitoring) instead.
- **NVIDIA Run:ai Prometheus instance** - During cluster installation, the NVIDIA Run:ai cluster operator creates a dedicated Prometheus Custom Resource (CR) in the `runai` namespace. This step applies to both Kubernetes and OpenShift, and produces the Prometheus instance that NVIDIA Run:ai uses for monitoring.

For air-gapped installation instructions, see [Prometheus](../install-using-helm/prerequisites-and-features.md#prometheus) in the Manual Helm path.

## Features

Each feature is configured under `features` in the configuration file, and only needs to be enabled for the workload types you plan to run, with a `mode` field controlling how the installer handles it.

* `install` - Installs and configures the component for you.
* `skip` (default) - Does not install or check for the component. Use this when the component is already installed on your cluster, or when the feature does not apply to the workload types you plan to run.

<Note>

**Note**

Features already default to `mode: skip`, so air-gapped installations need no changes here.

</Note>

Each feature also accepts its own `version` field - omit it to install the latest version, or see [Support Matrix](../../support-matrix.md) for the supported version range if you want to set it explicitly.

### Distributed Training

NVIDIA Run:ai supports distributed training - training AI models across multiple nodes - using the [Kubeflow Training Operator](https://www.kubeflow.org/docs/components/training/installation/) for [TensorFlow](https://www.tensorflow.org/), [PyTorch](https://pytorch.org/), [XGBoost](https://xgboost.readthedocs.io/), and [JAX](https://docs.jax.dev/en/latest/index.html) workloads, and the [MPI Operator](https://github.com/kubeflow/mpi-operator) for [MPI v2](https://docs.open-mpi.org/)-based workloads. Both are required for compatibility with advanced workload capabilities such as [stopping a workload](../../../workloads-in-nvidia-run-ai/workloads.md) and [scheduling rules](../../../platform-management/policies/scheduling-rules.md).

The Kubeflow Training Operator ships with MPI v1 support built in, which conflicts with the separate MPI Operator (MPI v2). When both `features.kubeflow.mode: install` and `features.mpiOperator.mode: install` are set, the installer handles this automatically - disabling MPI v1 support and removing its CRD before installing the MPI Operator.

For air-gapped installation instructions, see [Distributed Training](../install-using-helm/prerequisites-and-features.md#distributed-training) in the Manual Helm path.

### Standard Inference

NVIDIA Run:ai uses [Knative Serving](https://knative.dev/docs/serving/) to enable standard (single-node) inference. Knative Serving lets you serve AI models as production workloads, using NVIDIA NIM, vLLM, or custom inference servers. It provides serverless, dynamic autoscaling (including scale to zero when idle), rolling updates without disrupting served traffic, and endpoints reachable via external access (a wildcard FQDN) or internal, cluster-local access. For an overview of the request flow, see [Workload Architecture](../../../workloads-in-nvidia-run-ai/using-inference/nvidia-run-ai-inference-overview.md#workload-architecture).

Setting `features.knative.mode: install` installs the Knative operator and applies a `KnativeServing` custom resource:

- **Kubernetes** - Installs Knative Serving via the Knative Operator, and configures scale-to-zero autoscaling, the wildcard inference domain, Kourier as the ingress class, and the Kubernetes pod-spec features NVIDIA Run:ai workloads require.
- **OpenShift** - Installs Knative Serving via the OpenShift Serverless operator (OperatorHub), and configures the wildcard inference domain and the Kubernetes pod-spec features NVIDIA Run:ai workloads require. External access goes through OpenShift Routes instead of Kourier, and scale-to-zero uses the operator's default.

For air-gapped installation instructions, see [Standard Inference](../install-using-helm/prerequisites-and-features.md#standard-inference) in the Manual Helm path.

#### Autoscaling

NVIDIA Run:ai allows for autoscaling a deployment according to the metrics below:

* Latency (milliseconds)
* Throughput (requests/sec)
* Concurrency (requests)

Using a custom metric (for example, Latency) requires the Knative Horizontal Pod Autoscaler (HPA) extension. Setting `features.knative.mode: install` installs this extension automatically alongside Knative Serving, so no additional step is needed.

For air-gapped installation instructions, see [Autoscaling](../install-using-helm/prerequisites-and-features.md#autoscaling) in the Manual Helm path.

### Distributed Inference

NVIDIA Run:ai uses [LeaderWorkerSet (LWS)](https://lws.sigs.k8s.io/) to enable distributed (multi-node) inference - serving large language models that do not fit on a single node by distributing them across multiple nodes.

Setting `features.lws.mode: install` installs LWS:

- **Kubernetes** - Installs the LWS Helm chart into the `lws-system` namespace.
- **OpenShift** - Installs the OpenShift LWS Operator (via OperatorHub) into the `openshift-lws-operator` namespace.

Either way, the installer automatically points NVIDIA Run:ai (`clusterConfig.lws.namespace`) at the namespace LWS runs in, so no manual configuration is needed for NVIDIA Run:ai to detect it.

For air-gapped installation instructions, see [Distributed Inference](../install-using-helm/prerequisites-and-features.md#distributed-inference) in the Manual Helm path.

### NIM Operator

The [NVIDIA NIM Operator](https://docs.nvidia.com/nim-operator/latest/) enables Kubernetes cluster administrators to operate the software components and services necessary to deploy NVIDIA NIMs and NVIDIA NeMo microservices in Kubernetes. NVIDIA Run:ai provides out-of-the-box support for submitting NIM Operator workloads [via YAML](../../../workloads-in-nvidia-run-ai/submit-via-yaml.md).

For air-gapped installation, follow the instructions in [Air-Gapped Environments - NVIDIA NIM Operator](https://docs.nvidia.com/nim-operator/latest/air-gap.html).

### Grove

NVIDIA Run:ai uses [NVIDIA Grove](https://developer.nvidia.com/grove) to orchestrate multi-component, gang-scheduled workloads, such as disaggregated NVIDIA Dynamo inference deployments where a single workload is split into multiple functions (for example, a router, prefill, and decode component), each made up of one or more leader-worker replica sets. Grove acts as the Kubernetes API (CRD) and controller for these pods, and is required alongside the Dynamo Operator for disaggregated Dynamo deployments.

For air-gapped installation instructions, see [Installation](https://github.com/ai-dynamo/grove/blob/main/docs/installation.md) in the Grove repository.

### Dynamo

The [Dynamo Operator](https://docs.nvidia.com/dynamo) is a Kubernetes operator that simplifies the deployment, configuration, and lifecycle management of DynamoGraphs. NVIDIA Run:ai provides out-of-the-box support for submitting Dynamo workloads [via YAML](../../../workloads-in-nvidia-run-ai/submit-via-yaml.md).

For air-gapped installation instructions, see the [Installation Guide for Dynamo Kubernetes Platform](https://docs.nvidia.com/dynamo/kubernetes/getting-started/introduction).
