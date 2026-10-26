# Prerequisites and Features

Prerequisites and features are additional infrastructure required by the NVIDIA Run:ai cluster. Prerequisites are components every NVIDIA Run:ai deployment depends on, such as the GPU Operator. Features extend NVIDIA Run:ai to support specific workload types, such as distributed training or inference, and only need to be installed for the workloads you plan to run. This page describes how to install them manually using Helm, along with any Helm-specific handling. For detailed version compatibility, refer to the [Support matrix](../../support-matrix.md).

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

## Before You Start

Complete [Preparations](preparations.md) before starting here. The operators and frameworks described on this page must be installed before you [install the NVIDIA Run:ai cluster](helm-install.md), on the Kubernetes cluster where it will run - the same Kubernetes cluster as the control plane in a same-cluster deployment, or a separate Kubernetes cluster per NVIDIA Run:ai cluster in a multi-cluster deployment. See [Deployment Model](../../system-requirements/deployment-model.md).

## Prerequisites

### NVIDIA GPU Operator

The [NVIDIA GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/index.html) provisions and manages GPUs on worker nodes, and is required for the NVIDIA Run:ai cluster to run AI workloads.

<Note>

**Note**

For [Multi-Node NVLink](../../../platform-management/aiinitiatives/resources/using-gb200.md) support (e.g. GB200), GPU Operator 25.3 and above is required.

</Note>

<Note>

**Note**

To use GPU fractions (memory fractions, dynamic memory fractions, or compute modes), the following are required:

* GPU Operator 26.7.1 or later is required.
* CUDA driver r615 or later is required. GPU Operator 26.7.1 does not default to it, so set the driver version explicitly rather than relying on the default - with `--set driver.version=<version>` on Kubernetes, or by setting `driver.version` in the `ClusterPolicy` spec on OpenShift, where the GPU Operator is installed from OperatorHub.
* The NRI plugin must be disabled. Install the GPU Operator with `cdi.nriPluginEnabled` set to `false` (`--set cdi.nriPluginEnabled=false` on Kubernetes, or `cdi.nriPluginEnabled: false` in the `ClusterPolicy` spec on OpenShift). GPU memory fractions do not function while the NRI plugin is enabled.

See [GPU Memory Fractions](../../../platform-management/runai-scheduler/resource-optimization/memory-fractions.md#prerequisites) for the full requirement set.

</Note>

* For installation instructions, see [Installing the NVIDIA GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/getting-started.html).
* For air-gapped installation, follow the instructions in [Install NVIDIA GPU Operator in Air-Gapped Environments](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/install-gpu-operator-air-gapped.html).
* For Gov-ready installations, follow the instructions in [Install the Gov-Ready GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/install-gpu-operator-gov-ready.html).
* Use the default `gpu-operator` namespace. Otherwise, you must specify the target namespace using the flag `clusterConfig.global.nvidiaDcgmExporter.namespace` as described in [Advanced cluster configurations](../../configuration/cluster-config.md).
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

### NVIDIA Network Operator

When deploying on clusters with RDMA or Multi-Node NVLink-capable nodes (for example B200, GB200), the NVIDIA Network Operator is required to enable high-performance networking features such as GPUDirect RDMA in Kubernetes.

The Network Operator works alongside the NVIDIA GPU Operator to provide:

* NVIDIA networking drivers for advanced network capabilities.
* Kubernetes device plugins to expose high-speed network hardware to workloads.
* Secondary network components to support network-intensive applications.

The Network Operator must be installed and configured as follows:

* Install the Network Operator as detailed in [Network Operator Deployment on Vanilla Kubernetes Cluster](https://docs.nvidia.com/networking/display/kubernetes2440/getting-started-kubernetes.html#network-operator-deployment-on-vanilla-kubernetes-cluster).
* Configure SR-IOV InfiniBand support as detailed in [Network Operator Deployment with an SR-IOV InfiniBand Network](https://docs.nvidia.com/networking/display/kubernetes2440/getting-started-kubernetes.html#network-operator-deployment-with-an-sr-iov-infiniband-network).
* For air-gapped installation, follow the instructions in [Network Operator Deployment in an Air-gapped Environment](https://docs.nvidia.com/networking/display/kubernetes2670/advanced/proxy-airgapped.html#network-operator-deployment-in-an-air-gapped-environment).
* For Gov-ready installations, follow the instructions in [Install the Gov-Ready Network Operator](https://docs.nvidia.com/networking/display/kubernetes2670/install-network-operator-gov-ready.html).

### NVIDIA Dynamic Resource Allocation (DRA) Driver

When deploying on clusters with Multi-Node NVLink (e.g. GB200), the [NVIDIA DRA driver](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/26.7/dra-intro-install.html) is required. For air-gapped installations, follow the air-gapped instructions in the [NVIDIA GPU Operator](#nvidia-gpu-operator) section.

How the DRA driver is installed depends on the GPU Operator mode. A cluster runs in one mode only:

* **`ClusterPolicy` mode (default)** - GPUs are exposed through the NVIDIA device plugin. Install the DRA driver separately. Refer to [Support Matrix](../../support-matrix.md) for the supported DRA driver versions.
* **`GPUCluster` mode (GPU Operator 26.7, new installations only)** - The GPU Operator deploys the DRA driver, which replaces the device plugin.

For details, refer to [NVIDIA DRA driver](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/26.7/dra-intro-install.html) and, on OpenShift, [Installing the NVIDIA GPU Operator with DRA on OpenShift](https://docs.nvidia.com/datacenter/cloud-native/openshift/26.7/dra-gpu-ocp.html).

<Warning>

**Important**

NVIDIA Run:ai limitations in `GPUCluster` mode:

* GPU fractions are not supported.
* NVIDIA Run:ai shows each node's GPU model, count, and memory only when GPU Feature Discovery is deployed separately. See [Deploying gpu-feature-discovery in standalone mode](https://github.com/NVIDIA/k8s-device-plugin#deploying-gpu-feature-discovery-in-standalone-mode).
* Workloads that request `nvidia.com/gpu` require [extended-resource translation](../../../platform-management/runai-scheduler/scheduling/concepts-and-principles.md#extended-resource-translation), which requires Kubernetes 1.37 or later.

</Warning>

After the DRA driver is installed, update the cluster configuration using the `anyworkload-controller.GPUNetworkAccelerationEnabled` flag to enable GPU network acceleration. This triggers an update of the NVIDIA Run:ai `anyworkload-controller` deployment and restarts the controller. For details on how to configure this value using Helm or `runaiconfig`, see [Advanced cluster configurations](../../configuration/cluster-config.md).

### Prometheus

<Note>

**Note**

Installing Prometheus applies to Kubernetes only.

</Note>

NVIDIA Run:ai requires Prometheus on the Kubernetes cluster. The setup has two parts:

* **Prerequisite: Kube-Prometheus Stack.** On Kubernetes, install the Kube-Prometheus Stack to provide the Prometheus Operator and the Prometheus Custom Resource Definitions (CRDs). OpenShift includes equivalent components through its built-in [Cluster Monitoring Operator](https://docs.redhat.com/en/documentation/openshift_container_platform/4.18/html-single/monitoring/index#about-ocp-monitoring), so no install is required.
* **NVIDIA Run:ai Prometheus instance.** During cluster installation, the NVIDIA Run:ai cluster operator creates a dedicated Prometheus Custom Resource (CR) in the `runai` namespace. This step applies to both Kubernetes and OpenShift, and produces the Prometheus instance that NVIDIA Run:ai uses for monitoring.

For RKE2, see [Enable Monitoring](https://ranchermanager.docs.rancher.com/how-to-guides/advanced-user-guides/monitoring-alerting-guides/enable-monitoring) for instructions on installing Prometheus.

The following example installs the community [Kube-Prometheus Stack](https://artifacthub.io/packages/helm/prometheus-community/kube-prometheus-stack) using [helm](https://helm.sh/):

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install prometheus prometheus-community/kube-prometheus-stack \
    -n monitoring --create-namespace --set grafana.enabled=false
```

## Features

To enable NVIDIA Run:ai capabilities such as Distributed Training and Inference, additional Kubernetes applications (frameworks) must be installed on the cluster. Only install what you need for the workload types you plan to run.

### Distributed Training

NVIDIA Run:ai supports distributed training - training AI models across multiple nodes - using the [Kubeflow Training Operator](https://www.kubeflow.org/docs/components/training/installation/) for [TensorFlow](https://www.tensorflow.org/), [PyTorch](https://pytorch.org/), [XGBoost](https://xgboost.readthedocs.io/), and [JAX](https://docs.jax.dev/en/latest/index.html) workloads, and the [MPI Operator](https://github.com/kubeflow/mpi-operator) for [MPI v2](https://docs.open-mpi.org/)-based workloads. Both are required for compatibility with advanced workload capabilities such as [Stopping a workload](../../../workloads-in-nvidia-run-ai/workloads.md) and [Scheduling rules](../../../platform-management/policies/scheduling-rules.md).

*   To install the Kubeflow Training Operator for TensorFlow, PyTorch, XGBoost and JAX frameworks, run the following command:

    ```bash
    kubectl apply --server-side -k "github.com/kubeflow/training-operator.git/manifests/overlays/standalone?ref=v1.9.2"
    ```
*   To install the MPI Operator for MPI v2, run the following command:

    ```bash
    kubectl apply --server-side -f https://raw.githubusercontent.com/kubeflow/mpi-operator/v0.6.0/deploy/v2beta1/mpi-operator.yaml
    ```

<Note>

**Note**

If you require both the MPI Operator and Kubeflow Training Operator, follow the steps below:

* Install the Kubeflow Training Operator as described above.
*   Disable and delete MPI v1 in the Kubeflow Training Operator by running:

    ```bash
    kubectl patch deployment training-operator -n kubeflow --type='json' -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args", "value": ["--enable-scheme=tfjob", "--enable-scheme=pytorchjob", "--enable-scheme=xgboostjob", "--enable-scheme=jaxjob"]}]'
    kubectl delete crd mpijobs.kubeflow.org
    ```
* Install the MPI Operator as described above.

</Note>

### Standard Inference

NVIDIA Run:ai uses [Knative Serving](https://knative.dev/docs/serving/) to enable standard (single-node) inference. Knative Serving lets you serve AI models as production workloads, using NVIDIA NIM, vLLM, or custom inference servers. It provides serverless, dynamic autoscaling (including scale to zero when idle), rolling updates without disrupting served traffic, and endpoints reachable via external access (a wildcard FQDN) or internal, cluster-local access. For an overview of the request flow, see [Workload Architecture](../../../workloads-in-nvidia-run-ai/using-inference/nvidia-run-ai-inference-overview.md#workload-architecture).

<Note>

**Note**

* On OpenShift, Knative is installed through the OpenShift Serverless operator, which supports Knative 1.17 only.
* By default, Knative pulls its `queue-proxy` sidecar image from the public registry. In air-gapped environments, update the `queue-sidecar-image` field in the `config-deployment` ConfigMap (in the `knative-serving` namespace) to point to your local registry instead. See the [Knative `config-deployment` ConfigMap](https://knative.dev/docs/serving/configuration/deployment/) for the field definition.

</Note>

To configure inference:

* **Inference request metrics (OTLP)** - For Knative Serving 1.19 or later, add the `observability` block to your `KnativeServing` YAML to export request metrics to the NVIDIA Run:ai API and UI. See [Inference request metrics (OTLP)](prerequisites-and-features.md#inference-request-metrics-otlp) below.
* **Internal or external access** - Choose how inference services are reached: externally via a wildcard FQDN (the default) or cluster-local. See the matching section below.

#### Inference Request Metrics (OTLP)

To surface inference request metrics in the NVIDIA Run:ai API and UI, configure Knative Serving 1.19 or later to export request metrics over OpenTelemetry Protocol (OTLP) to the NVIDIA Run:ai Prometheus instance. The `observability` block included in the `KnativeServing` YAML examples below sets the required `config-observability` keys (`metrics-protocol`, `request-metrics-protocol`, `request-metrics-endpoint`). For supported fields, see [Enabling metric collection](https://knative.dev/docs/serving/observability/metrics/collecting-metrics/#enabling-metric-collection) in the Knative documentation.

The OTLP receiver on the NVIDIA Run:ai Prometheus instance is enabled by default.

To disable it, set `spec.prometheus.config.enableOTLPReceiver: false`. See [Advanced cluster configurations](../../configuration/cluster-config.md) for more details.

#### External-Access Inference

With external access, inference services are reachable from outside the cluster. Use this mode when inference workloads must be reachable by clients outside the cluster.

<Tabs>

<Tab title="Kubernetes">

On Kubernetes, external access exposes inference services through a wildcard FQDN (for example, `*.runai-inference.mycorp.local`). This requires the wildcard DNS record (see [FQDN and DNS](../../system-requirements/networking-setup.md#fqdn-and-dns)), the standard inference wildcard certificate (see [Cluster TLS Certificates](preparations.md#cluster-tls-certificates)), and an HAProxy ingress to handle TLS termination.

Follow the [Installing Knative](https://knative.dev/docs/install/operator/knative-with-operators/) instructions or run the following, updating `{VERSION}` with a `knative-operator` release that supports the Knative Serving version you are using:

```bash
helm repo add knative-operator https://knative.github.io/operator
helm install knative-operator --create-namespace --namespace knative-operator --version {VERSION} knative-operator/knative-operator
```

Once installed, follow the steps below:

1.  Create the `knative-serving` namespace:

    ```bash
    kubectl create ns knative-serving
    ```
2.  Create a YAML file named `knative-serving.yaml` and replace the placeholder FQDN with your wildcard inference domain (for example, `runai-inference.mycorp.local`).

    ```yaml
    apiVersion: operator.knative.dev/v1beta1
    kind: KnativeServing
    metadata:
      name: knative-serving
      namespace: knative-serving
    spec:
      config:
        config-autoscaler:
          enable-scale-to-zero: "true"
        config-features:
          kubernetes.podspec-affinity: enabled
          kubernetes.podspec-init-containers: enabled
          kubernetes.podspec-persistent-volume-claim: enabled
          kubernetes.podspec-persistent-volume-write: enabled
          kubernetes.podspec-schedulername: enabled
          kubernetes.podspec-securitycontext: enabled
          kubernetes.podspec-tolerations: enabled
          kubernetes.podspec-volumes-emptydir: enabled
          kubernetes.podspec-fieldref: enabled
          kubernetes.containerspec-addcapabilities: enabled
          kubernetes.podspec-nodeselector: enabled
          multi-container: enabled
          kubernetes.podspec-hostipc: enabled
          kubernetes.podspec-hostnetwork: enabled
        domain:
          runai-inference.mycorp.local: "" # replace with the wildcard FQDN for Inference
        network:
          domainTemplate: '{{.Name}}-{{.Namespace}}.{{.Domain}}'
          ingress-class: kourier.ingress.networking.knative.dev
          default-external-scheme: https
        observability:
          metrics-protocol: "prometheus"
          request-metrics-protocol: "http/protobuf"
          request-metrics-endpoint: "http://prometheus-operated.runai.svc.cluster.local:9090/api/v1/otlp/v1/metrics"
      high-availability:
        replicas: 2
      ingress:
        kourier:
          enabled: true
    ```
3.  Apply the changes:

    ```bash
    kubectl apply -f knative-serving.yaml
    ```
4.  Configure HAProxy to proxy requests to Kourier / Knative and handle TLS termination using the wildcard certificate. Create a YAML file named `knative-ingress.yaml` and replace the FQDN placeholders with your wildcard inference domain:

    ```yaml
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: knative-serving
      namespace: knative-serving
    spec:
      ingressClassName: haproxy
      rules:
      - host: '*.runai-inference.mycorp.local' # replace with the wildcard FQDN for Inference
        http:
          paths:
          - backend:
              service:
                name: kourier
                port:
                  number: 80
            path: /
            pathType: Prefix
      tls:
      - hosts:
        - '*.runai-inference.mycorp.local' # replace with the wildcard FQDN for Inference
        secretName: runai-cluster-inference-tls-secret
    ```
5.  Apply the changes:

    ```bash
    kubectl apply -f knative-ingress.yaml
    ```

</Tab>

<Tab title="OpenShift">

On OpenShift Serverless, external access is the platform default: each KnativeService automatically gets an OpenShift Route.

Follow the [Installing the OpenShift Serverless Operator](https://docs.redhat.com/en/documentation/red_hat_openshift_serverless/1.37/html/installing_openshift_serverless/install-serverless-operator) instructions. Once installed, follow the steps below:

1.  Create the `knative-serving` project:

    ```bash
    oc new-project knative-serving
    ```
2.  Create a YAML file named `knative-serving.yaml`:

    ```yaml
    apiVersion: operator.knative.dev/v1beta1
    kind: KnativeServing
    metadata:
      finalizers:
        - knative-serving-openshift
        - knativeservings.operator.knative.dev
      name: knative-serving
      namespace: knative-serving
    spec:
      config:
        config-features:
          kubernetes.podspec-tolerations: enabled
          kubernetes.podspec-volumes-emptydir: enabled
          kubernetes.podspec-persistent-volume-claim: enabled
          multi-container: enabled
          kubernetes.podspec-persistent-volume-write: enabled
          kubernetes.podspec-fieldref: enabled
          kubernetes.podspec-schedulername: enabled
          kubernetes.podspec-nodeselector: enabled
          kubernetes.podspec-init-containers: enabled
          kubernetes.podspec-securitycontext: enabled
          kubernetes.podspec-affinity: enabled
          kubernetes.containerspec-addcapabilities: enabled
        observability:
          metrics-protocol: "prometheus"
          request-metrics-protocol: "http/protobuf"
          request-metrics-endpoint: "http://prometheus-operated.runai.svc.cluster.local:9090/api/v1/otlp/v1/metrics"
      controller-custom-certs:
        name: ''
        type: ''
      registry: {}
    ```
3.  Apply the changes:

    ```bash
    oc apply -f knative-serving.yaml
    ```

</Tab>

</Tabs>

#### Cluster-Local Inference

With cluster-local access, inference services are reachable only from inside the cluster through the cluster-internal domain (`<service>.<namespace>.svc.cluster.local`). This mode is recommended when inference workloads are consumed by other applications running in the same cluster.

On OpenShift Serverless, external access is the platform default; cluster-local must be enabled per inference workload. See the OpenShift tab below.

<Tabs>

<Tab title="Kubernetes">

On Kubernetes, cluster-local access does not require an external DNS record, a wildcard certificate, or an HAProxy ingress because traffic does not leave the cluster.

Follow the [Installing Knative](https://knative.dev/docs/install/operator/knative-with-operators/) instructions or run the following, updating `{VERSION}` with a `knative-operator` release that supports the Knative Serving version you are using:

```bash
helm repo add knative-operator https://knative.github.io/operator
helm install knative-operator --create-namespace --namespace knative-operator --version {VERSION} knative-operator/knative-operator
```

Once installed, follow the steps below:

1.  Create the `knative-serving` namespace:

    ```bash
    kubectl create ns knative-serving
    ```
2.  Create a YAML file named `knative-serving.yaml`:

    ```yaml
    apiVersion: operator.knative.dev/v1beta1
    kind: KnativeServing
    metadata:
      name: knative-serving
      namespace: knative-serving
    spec:
      config:
        config-autoscaler:
          enable-scale-to-zero: "true"
        config-features:
          kubernetes.podspec-affinity: enabled
          kubernetes.podspec-init-containers: enabled
          kubernetes.podspec-persistent-volume-claim: enabled
          kubernetes.podspec-persistent-volume-write: enabled
          kubernetes.podspec-schedulername: enabled
          kubernetes.podspec-securitycontext: enabled
          kubernetes.podspec-tolerations: enabled
          kubernetes.podspec-volumes-emptydir: enabled
          kubernetes.podspec-fieldref: enabled
          kubernetes.containerspec-addcapabilities: enabled
          kubernetes.podspec-nodeselector: enabled
          multi-container: enabled
          kubernetes.podspec-hostipc: enabled
          kubernetes.podspec-hostnetwork: enabled
        domain:
          svc.cluster.local: ""
        network:
          domainTemplate: '{{.Name}}.{{.Namespace}}.{{.Domain}}'
          ingress-class: kourier.ingress.networking.knative.dev
          default-external-scheme: http
        observability:
          metrics-protocol: "prometheus"
          request-metrics-protocol: "http/protobuf"
          request-metrics-endpoint: "http://prometheus-operated.runai.svc.cluster.local:9090/api/v1/otlp/v1/metrics"
      high-availability:
        replicas: 2
      ingress:
        kourier:
          enabled: true
          service-type: ClusterIP
    ```
3.  Apply the changes:

    ```bash
    kubectl apply -f knative-serving.yaml
    ```

</Tab>

<Tab title="OpenShift">

On OpenShift Serverless, external access is the platform default: KnativeServices automatically get an OpenShift Route. To make an inference workload cluster-local, configure the workload's routing settings so that the `networking.knative.dev/visibility=cluster-local` label is applied to the underlying KnativeService.

Follow the [Installing the OpenShift Serverless Operator](https://docs.redhat.com/en/documentation/red_hat_openshift_serverless/1.37/html/installing_openshift_serverless/install-serverless-operator) instructions. Once installed, follow the steps below:

1.  Create the `knative-serving` project:

    ```bash
    oc new-project knative-serving
    ```
2.  Create a YAML file named `knative-serving.yaml`.

    ```yaml
    apiVersion: operator.knative.dev/v1beta1
    kind: KnativeServing
    metadata:
      finalizers:
        - knative-serving-openshift
        - knativeservings.operator.knative.dev
      name: knative-serving
      namespace: knative-serving
    spec:
      config:
        config-features:
          kubernetes.podspec-tolerations: enabled
          kubernetes.podspec-volumes-emptydir: enabled
          kubernetes.podspec-persistent-volume-claim: enabled
          multi-container: enabled
          kubernetes.podspec-persistent-volume-write: enabled
          kubernetes.podspec-fieldref: enabled
          kubernetes.podspec-schedulername: enabled
          kubernetes.podspec-nodeselector: enabled
          kubernetes.podspec-init-containers: enabled
          kubernetes.podspec-securitycontext: enabled
          kubernetes.podspec-affinity: enabled
          kubernetes.containerspec-addcapabilities: enabled
        observability:
          metrics-protocol: "prometheus"
          request-metrics-protocol: "http/protobuf"
          request-metrics-endpoint: "http://prometheus-operated.runai.svc.cluster.local:9090/api/v1/otlp/v1/metrics"
      controller-custom-certs:
        name: ''
        type: ''
      registry: {}
    ```
3.  Apply the changes:

    ```bash
    oc apply -f knative-serving.yaml
    ```

</Tab>

</Tabs>

#### Autoscaling

NVIDIA Run:ai allows for autoscaling a deployment according to the metrics below:

* Latency (milliseconds)
* Throughput (requests/sec)
* Concurrency (requests)

Using a custom metric (for example, Latency) requires installing the [Kubernetes Horizontal Pod Autoscaler (HPA)](https://knative.dev/docs/install/yaml-install/serving/install-serving-with-yaml/#install-optional-serving-extensions). Use the following command to install. Make sure to update the {VERSION} in the below command with a [supported Knative version](prerequisites-and-features.md#standard-inference).

```bash
kubectl apply -f https://github.com/knative/serving/releases/download/knative-{VERSION}/serving-hpa.yaml
```

### Distributed Inference

NVIDIA Run:ai uses [LeaderWorkerSet (LWS)](https://lws.sigs.k8s.io/) to enable distributed (multi-node) inference - serving large language models that do not fit on a single node by distributing them across multiple nodes.

To enable this capability, install LWS on your cluster:

<Tabs>

<Tab title="Kubernetes">

Install the [LWS Helm chart](https://lws.sigs.k8s.io/docs/installation/#install-by-helm), updating `{VERSION}` with a [supported LWS version](../../support-matrix.md):

```bash
CHART_VERSION={VERSION}
helm install lws oci://registry.k8s.io/lws/charts/lws \
  --version=$CHART_VERSION \
  --namespace lws-system \
  --create-namespace \
  --wait --timeout 300s
```

</Tab>

<Tab title="OpenShift">

1. Install the Leader Worker Set Operator by following the [Installing the Leader Worker Set Operator](https://docs.redhat.com/en/documentation/openshift_container_platform/4.19/html/ai_workloads/leader-worker-set-operator#lws-install-operator_lws-managing) instructions in the OpenShift documentation. Only the operator installation step is required; the remaining sections can be ignored.
2. Configure the LWS namespace so NVIDIA Run:ai can detect it. The OpenShift LWS Operator installs into the `openshift-lws-operator` namespace by default. Choose one of the following methods:
   * **Helm flag:** Set `clusterConfig.lws.namespace=openshift-lws-operator` during cluster installation. For more details, see [Advanced cluster configurations](../../configuration/cluster-config.md).
   *   **runaiconfig patch:** Apply the setting directly:

       ```bash
       oc patch runaiconfig runai -n runai --type merge \
         -p '{"spec":{"lws":{"namespace":"openshift-lws-operator"}}}'
       ```

</Tab>

</Tabs>

### NIM Operator

The [NVIDIA NIM Operator](https://docs.nvidia.com/nim-operator/latest/) enables Kubernetes cluster administrators to operate the software components and services necessary to deploy NVIDIA NIMs and NVIDIA NeMo microservices in Kubernetes. NVIDIA Run:ai provides out-of-the-box support for submitting NIM Operator workloads [via YAML](../../../workloads-in-nvidia-run-ai/submit-via-yaml.md).

For air-gapped installation, follow the instructions in [Air-Gapped Environments - NVIDIA NIM Operator](https://docs.nvidia.com/nim-operator/latest/air-gap.html).

### Grove

NVIDIA Run:ai uses [NVIDIA Grove](https://developer.nvidia.com/grove) to orchestrate multi-component, gang-scheduled workloads, such as disaggregated NVIDIA Dynamo inference deployments where a single workload is split into multiple functions (for example, a router, prefill, and decode component), each made up of one or more leader-worker replica sets. Grove acts as the Kubernetes API (CRD) and controller for these pods, and is required alongside the Dynamo Operator for disaggregated Dynamo deployments.

For air-gapped installation instructions, see [Installation](https://github.com/ai-dynamo/grove/blob/main/docs/installation.md) in the Grove repository.

### Dynamo

The [Dynamo Operator](https://docs.nvidia.com/dynamo) is a Kubernetes operator that simplifies the deployment, configuration, and lifecycle management of DynamoGraphs. NVIDIA Run:ai provides out-of-the-box support for submitting Dynamo workloads [via YAML](../../../workloads-in-nvidia-run-ai/submit-via-yaml.md).

For air-gapped installation instructions, see the [Installation Guide for Dynamo Kubernetes Platform](https://docs.nvidia.com/dynamo/kubernetes/getting-started/introduction).

## Next Steps

With these prerequisites in place, continue to [Install the Control Plane](install-control-plane.md). The NVIDIA Run:ai cluster cannot be installed until the control plane is running, since installing the cluster connects it to the control plane.
