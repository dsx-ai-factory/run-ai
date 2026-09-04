---
---

# Prerequisites and Features

The `runai-installer` CLI can install and configure additional infrastructure components alongside NVIDIA Run:ai. Prerequisites are components every NVIDIA Run:ai deployment depends on, such as the GPU Operator. Features extend NVIDIA Run:ai to support specific workload types, such as distributed training or inference, and only need to be enabled for the workloads you plan to run. Both are installed on the NVIDIA Run:ai cluster(s) only, not the control plane:

| Component | Type | Required for |
|---|---|---|
| [GPU Operator](#gpu-operator) | Prerequisite | Every installation |
| [Prometheus](#prometheus) | Prerequisite | Every installation |
| [Network Operator](#network-operator) | Prerequisite | RDMA or Multi-Node NVLink (MNNVL)-capable nodes |
| [Knative Serving](#knative-serving) | Feature | Standard (single-node) inference |
| [Kubeflow Training Operator](#kubeflow-training-operator) | Feature | Distributed training |
| [MPI Operator](#mpi-operator) | Feature | MPI-based distributed training |
| [LeaderWorkerSet (LWS)](#leaderworkerset-lws) | Feature | Distributed (multi-node) inference |
| [NIM Operator](#nim-operator) | Feature | NVIDIA Inference Microservices |
| [Grove](#grove) | Feature | Multi-component, gang-scheduled workload orchestration (for example, disaggregated Dynamo deployments) |
| [Dynamo](#dynamo) | Feature | Inference pipeline optimization |

## Custom Helm Values

For settings not exposed directly by the installer, use `customHelmValues` to merge custom Helm values into the relevant chart. `controlPlane.customHelmValues` and `cluster.customHelmValues` accept any key documented in [Advanced Control Plane Configurations](../../configuration/control-plane-config.md) and [Advanced Cluster Configurations](../../configuration/cluster-config.md) respectively - for example, `GPUNetworkAccelerationEnabled` for Multi-Node NVLink.

Every prerequisite and feature entry except `kubeflow` and `mpiOperator` (`gpuOperator`, `prometheus`, `networkOperator`, `knative`, `lws`, `nimOperator`, `grove`, `dynamo`) also accepts `customHelmValues`, merged into that component's own Helm chart. `kubeflow` and `mpiOperator` install from pinned manifests rather than a Helm chart, so `customHelmValues` is rejected for both. On OpenShift, `customHelmValues` is also rejected for any component that installs via OperatorHub (OLM) instead of Helm.

<Warning>

**TODO**

Unlike `controlPlane`/`cluster` - documented in [Advanced Control Plane Configurations](../../configuration/control-plane-config.md) and [Advanced Cluster Configurations](../../configuration/cluster-config.md) - there's no page in this repo documenting what values each prerequisite/feature's `customHelmValues` actually accepts - these are values for the component's own chart (for example, GPU Operator's own chart, which is external to NVIDIA Run:ai), not NVIDIA Run:ai's software. Need to decide: do we document real examples per component (and where do we source them - the component's own upstream docs?), or point readers to each component's own documentation instead?

</Warning>

## Prerequisites

Each prerequisite is configured under `prerequisites` in the configuration file, with a `mode` field controlling how the installer handles it:

* `install` (the default for most prerequisites) - The installer installs and configures the component for you.
* `skip` - The installer does not install or check for the component. Use this when the component is already installed on your cluster, or when the prerequisite doesn't apply to your environment (for example, the Network Operator on a cluster without RDMA or Multi-Node NVLink nodes).

See [Support Matrix](../../support-matrix.md) for supported versions of each.

### GPU Operator

The [NVIDIA GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/index.html) provisions and manages GPUs on worker nodes, and is required for the NVIDIA Run:ai cluster to run AI workloads.

<Info>

**Note**

For [Multi-Node NVLink](../../../platform-management/aiinitiatives/resources/using-gb200.md) support (e.g. GB200), GPU Operator 25.3 and above is required.

</Info>

* Use the default `gpu-operator` namespace. Otherwise, you must specify the target namespace using the flag `runai-operator.config.nvidiaDcgmExporter.namespace` as described in [Advanced cluster configurations](../../configuration/cluster-config.md).
* NVIDIA drivers may already be installed on the nodes. In such cases, use the NVIDIA GPU Operator flags `--set driver.enabled=false`. [DGX OS](https://docs.nvidia.com/dgx/dgx-os-6-user-guide/release_notes.html) is one such example as it comes bundled with NVIDIA Drivers.
* For distribution-specific requirements and additional instructions, see the sections below:

<Accordion title="OpenShift Container Platform (OCP)">

**OpenShift Container Platform (OCP)**

The Node Feature Discovery (NFD) Operator is a prerequisite for the NVIDIA GPU Operator in OpenShift. Install the NFD Operator using the Red Hat OperatorHub catalog in the OpenShift Container Platform web console. For more information, see [Installing the Node Feature Discovery (NFD) Operator](https://docs.nvidia.com/datacenter/cloud-native/openshift/latest/install-nfd.html).

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

### NVIDIA Dynamic Resource Allocation (DRA) Driver

When deploying on clusters with Multi-Node NVLink (e.g. GB200), the [NVIDIA DRA driver](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/dra-intro-install.html) is required to enable Dynamic Resource Allocation at the Kubernetes level. The DRA driver is installed with the GPU Operator, for both connected and air-gapped installations.

After the DRA driver is installed, update the cluster configuration using the `GPUNetworkAccelerationEnabled` flag to enable GPU network acceleration. This triggers an update of the NVIDIA Run:ai workload controller deployment and restarts the controller. Set it via `cluster.customHelmValues` in the installer's configuration file - see [Custom Helm Values](#custom-helm-values) above - or, outside the guided installer, via Helm or `runaiconfig` as described in [Advanced cluster configurations](../../configuration/cluster-config.md).

### Network Operator

When deploying on clusters with RDMA or Multi-Node NVLink-capable nodes (for example B200, GB200), the NVIDIA Network Operator is required to enable high-performance networking features such as GPUDirect RDMA in Kubernetes.

The Network Operator works alongside the NVIDIA GPU Operator to provide:

* NVIDIA networking drivers for advanced network capabilities.
* Kubernetes device plugins to expose high-speed network hardware to workloads.
* Secondary network components to support network-intensive applications.

To configure SR-IOV InfiniBand support, set the required values via `customHelmValues` - see [Network Operator Deployment with an SR-IOV InfiniBand Network](https://docs.nvidia.com/networking/display/kubernetes2440/getting-started-kubernetes.html#network-operator-deployment-with-an-sr-iov-infiniband-network) for which values to set.

### Prometheus

<Info>

**Note**

Installing Prometheus applies to Kubernetes only.

</Info>

NVIDIA Run:ai requires Prometheus on the Kubernetes cluster. The setup has two parts:

- **Prerequisite: Kube-Prometheus Stack** - On Kubernetes, install the Kube-Prometheus Stack to provide the Prometheus Operator and the Prometheus Custom Resource Definitions (CRDs). OpenShift includes equivalent components through its built-in Cluster Monitoring Operator, so no install is required. For RKE2, see [Enable Monitoring](https://ranchermanager.docs.rancher.com/how-to-guides/advanced-user-guides/monitoring-alerting-guides/enable-monitoring) instead.
- **NVIDIA Run:ai Prometheus instance** - During cluster installation, the NVIDIA Run:ai cluster operator creates a dedicated Prometheus Custom Resource (CR) in the `runai` namespace. This step applies to both Kubernetes and OpenShift, and produces the Prometheus instance that NVIDIA Run:ai uses for monitoring.

## Features

Each feature is configured under `features` in the configuration file, and only needs to be enabled for the workload types you plan to run, with a `mode` field controlling how the installer handles it:

* `install` - The installer installs and configures the component for you.
* `skip` (the default) - The installer does not install or check for the component. Use this when the component is already installed on your cluster, or when the feature doesn't apply to the workload types you plan to run.

See [Support Matrix](../../support-matrix.md) for supported versions of each.

### Knative Serving

NVIDIA Run:ai uses [Knative Serving](https://knative.dev/docs/serving/) to enable standard (single-node) inference. Setting `features.knative.mode: install` installs and configures it automatically, including exporting inference request metrics to the NVIDIA Run:ai Prometheus instance via OTLP, so they surface in the NVIDIA Run:ai API and UI.

Knative Serving lets you serve AI models as production workloads, using NVIDIA NIM, vLLM, or custom inference servers. It provides serverless, dynamic autoscaling (including scale to zero when idle), rolling updates without disrupting served traffic, and endpoints reachable via external access (a wildcard FQDN) or internal, cluster-local access. For distributed (multi-node) inference, see [LeaderWorkerSet (LWS)](#leaderworkerset-lws) below.

<Warning>

**TODO**

NVIDIA Run:ai also supports autoscaling on Latency, Throughput, or Concurrency metrics. Using a custom metric like Latency requires separately installing the Knative Horizontal Pod Autoscaler (HPA) extension. Per the 2026-08-30 sync, this is planned to become a native installer capability rather than a manual step - not yet implemented, so this section will need updating once that lands.

</Warning>

### Kubeflow Training Operator

NVIDIA Run:ai uses the [Kubeflow Training Operator](https://www.kubeflow.org/docs/components/training/installation/) to enable distributed training - training AI models across multiple nodes - for [TensorFlow](https://www.tensorflow.org/), [PyTorch](https://pytorch.org/), [XGBoost](https://xgboost.readthedocs.io/), and [JAX](https://docs.jax.dev/en/latest/index.html) workloads. It is required for compatibility with advanced workload capabilities such as [stopping a workload](../../../workloads-in-nvidia-run-ai/workloads.md) and [scheduling rules](../../../platform-management/policies/scheduling-rules.md).

### MPI Operator

NVIDIA Run:ai uses the [MPI Operator](https://github.com/kubeflow/mpi-operator) to enable distributed training for [MPI v2](https://docs.open-mpi.org/)-based workloads, alongside the frameworks the Kubeflow Training Operator provides. It is required for compatibility with advanced workload capabilities such as [stopping a workload](../../../workloads-in-nvidia-run-ai/workloads.md) and [scheduling rules](../../../platform-management/policies/scheduling-rules.md).

The Kubeflow Training Operator ships with MPI v1 support built in, which conflicts with the separate MPI Operator (MPI v2). When both `features.kubeflow.mode: install` and `features.mpiOperator.mode: install` are set, the installer handles this automatically - disabling MPI v1 support and removing its CRD before installing the MPI Operator.

### LeaderWorkerSet (LWS)

NVIDIA Run:ai uses [LeaderWorkerSet (LWS)](https://lws.sigs.k8s.io/) to enable distributed (multi-node) inference - serving large language models that don't fit on a single node by distributing them across multiple nodes. For standard (single-node) inference, see [Knative Serving](#knative-serving) above.

<Warning>

**TODO**

The Manual Helm path notes that on OpenShift, LWS installs via the OpenShift LWS Operator into the `openshift-lws-operator` namespace by default - not the namespace NVIDIA Run:ai expects - so `clusterConfig.lws.namespace` must be set for NVIDIA Run:ai to detect it. See [Distributed Inference](https://run-ai-docs.nvidia.com/self-hosted/getting-started/installation/install-using-helm/system-requirements#distributed-inference). Need to confirm whether the guided installer sets this automatically when `features.lws.mode: install` on OpenShift, or whether users still need to set it themselves (and if so, through which field).

</Warning>

### NIM Operator

The NVIDIA NIM Operator enables Kubernetes cluster administrators to operate the software components and services necessary to deploy NVIDIA NIMs and NVIDIA NeMo microservices in Kubernetes. NVIDIA Run:ai provides out-of-the-box support for submitting NIM Operator workloads [via YAML](../../../workloads-in-nvidia-run-ai/submit-via-yaml.md).

### Grove

NVIDIA Run:ai uses [NVIDIA Grove](https://developer.nvidia.com/grove) to orchestrate multi-component, gang-scheduled workloads, such as disaggregated NVIDIA Dynamo inference deployments where a single workload is split into multiple functions (for example, a router, prefill, and decode component), each made up of one or more leader-worker replica sets. Grove acts as the Kubernetes API (CRD) and controller for these pods, and is required alongside the Dynamo Operator for disaggregated Dynamo deployments.

<Warning>

**TODO**

Version TBD - unconfirmed which Grove version(s) this installer supports. See [Support Matrix](../../support-matrix.md).

</Warning>

### Dynamo

The Dynamo Operator is a Kubernetes operator that simplifies the deployment, configuration, and lifecycle management of DynamoGraphs. NVIDIA Run:ai provides out-of-the-box support for submitting Dynamo workloads [via YAML](../../../workloads-in-nvidia-run-ai/submit-via-yaml.md).
