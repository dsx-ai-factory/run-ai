# Cluster System Requirements

After provisioning the tenant’s Kubernetes cluster, the next step is to install the required system components to support the NVIDIA Run:ai platform.

The NVIDIA Run:ai cluster is deployed as a Kubernetes application. This section outlines the minimum hardware and software requirements that must be installed and configured on each tenant cluster before deploying the NVIDIA Run:ai components.

## Hardware Requirements

The following hardware requirements are for the Kubernetes cluster nodes. By default, all NVIDIA Run:ai cluster services run on all available nodes. For production deployments, you may want to set [node roles](../../infrastructure-setup/advanced-setup/node-roles.md), to separate between system and worker nodes, reduce downtime and save CPU cycles on expensive GPU Machines.

### Architecture

* **x86** - Supported for Kubernetes and OpenShift.
* **ARM** - Supported for Kubernetes and OpenShift.

### NVIDIA Run:ai Cluster - System Nodes

This configuration is the minimum requirement you need to install and use NVIDIA Run:ai cluster.

| Component  | Required Capacity |
| ---------- | ----------------- |
| CPU        | 10 cores          |
| Memory     | 20GB              |
| Disk space | 50GB              |

<Info>

**Note**

To designate nodes to NVIDIA Run:ai system services, follow the instructions as described in [System nodes](../../infrastructure-setup/advanced-setup/node-roles.md#system-nodes).

</Info>

### NVIDIA Run:ai Cluster - Worker Nodes

The NVIDIA Run:ai cluster supports x86 and ARM CPUs, and any NVIDIA GPUs supported by the NVIDIA GPU Operator. The list of supported GPUs depends on the version of the NVIDIA GPU Operator installed in the cluster. NVIDIA Run:ai supports GPU Operator versions 25.10 to 26.3.

For the list of supported GPUs, see [Supported NVIDIA Data Center GPUs and Systems](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/platform-support.html#supported-nvidia-data-center-gpus-and-systems). To install the GPU Operator, see [NVIDIA GPU Operator](system-requirements.md#nvidia-gpu-operator).

<Info>

**Note**

* NVIDIA DGX Spark, NVIDIA Jetson and workstations are not supported.
* vGPU is not supported. NVIDIA Run:ai currently supports GPU passthrough only.

</Info>

The following configuration represents the minimum hardware requirements for installing and operating the NVIDIA Run:ai cluster on worker nodes. Each node must meet these specifications:

| Component | Required Capacity |
| --------- | ----------------- |
| CPU       | 2 cores           |
| Memory    | 4GB               |

<Info>

**Note**

To designate nodes to NVIDIA Run:ai workloads, follow the instructions as described in [Worker nodes](../../infrastructure-setup/advanced-setup/node-roles.md#worker-nodes).

</Info>

## Software Requirements

The following software requirements must be fulfilled on the Kubernetes cluster.

### Operating System

* Any **Linux** operating system supported by both Kubernetes and NVIDIA GPU Operator
* NVIDIA Run:ai cluster on Google Kubernetes Engine (GKE) supports both Ubuntu and Container Optimized OS (COS). COS is supported only with NVIDIA GPU Operator 24.6 or newer, and NVIDIA Run:ai cluster version 2.19 or newer.
* NVIDIA Run:ai cluster on Elastic Kubernetes Service (EKS) does not support Bottlerocket or Amazon Linux.
* NVIDIA Run:ai cluster on Oracle Kubernetes Engine (OKE) supports only Ubuntu.
* Internal tests are being performed on **Ubuntu 22.04** and **CoreOS** for OpenShift.

### Kubernetes Distribution

NVIDIA Run:ai cluster requires Kubernetes. The following Kubernetes distributions are supported:

* Vanilla Kubernetes
* OpenShift Container Platform (OCP)
* NVIDIA Base Command Manager (BCM)
* Rancher Kubernetes Engine 2 (RKE2)

<Info>

**Note**

* The latest release of the NVIDIA Run:ai cluster supports **Kubernetes 1.34 to 1.36** and **OpenShift 4.18 to 4.21**.
* For [Multi-Node NVLink](https://app.gitbook.com/s/fvuZQSF7VCx1Azh0LF0H/platform-management/aiinitiatives/resources/using-gb200) support (e.g. GB200), Kubernetes 1.32 and above is required.

</Info>

For existing Kubernetes clusters, see the following Kubernetes version support matrix for the latest NVIDIA Run:ai cluster releases:

| NVIDIA Run:ai version | Supported Kubernetes versions | Supported OpenShift versions |
| --------------------- | ----------------------------- | ---------------------------- |
| v2.26 (latest)        | 1.34 to 1.36                  | 4.18 to 4.21                 |
| v2.25                 | 1.33 to 1.35                  | 4.18 to 4.21                 |
| v2.24                 | 1.33 to 1.35                  | 4.17 to 4.20                 |
| v2.23                 | 1.31 to 1.34                  | 4.16 to 4.19                 |
| v2.22                 | 1.31 to 1.33                  | 4.15 to 4.19                 |

For information on supported versions of managed Kubernetes, it's important to consult the release notes provided by your Kubernetes service provider. There, you can confirm the specific version of the underlying Kubernetes platform supported by the provider, ensuring compatibility with NVIDIA Run:ai. For an up-to-date end-of-life statement see [Kubernetes Release History](https://kubernetes.io/releases/) or [OpenShift Container Platform Life Cycle Policy](https://access.redhat.com/support/policy/updates/openshift).

### Container Runtime

NVIDIA Run:ai supports the following [container runtimes](https://kubernetes.io/docs/setup/production-environment/container-runtimes/). Make sure your Kubernetes cluster is configured with one of these runtimes:

* [Containerd](https://kubernetes.io/docs/setup/production-environment/container-runtimes/#containerd) (default in Kubernetes)
* [CRI-O](https://cri-o.io/) (default in OpenShift)

### Kubernetes Pod Security Admission

NVIDIA Run:ai supports `restricted` policy for [Pod Security Admission](https://kubernetes.io/docs/concepts/security/pod-security-admission/) (PSA) on OpenShift only. Other Kubernetes distributions are only supported with `privileged` policy.

For NVIDIA Run:ai on OpenShift to run with PSA `restricted` policy:

* Label the `runai` namespace as described in [Pod Security Admission](https://kubernetes.io/docs/concepts/security/pod-security-admission/) with the following labels:

```
pod-security.kubernetes.io/audit=privileged
pod-security.kubernetes.io/enforce=privileged
pod-security.kubernetes.io/warn=privileged
```

* The workloads submitted through NVIDIA Run:ai should comply with the restrictions of PSA restricted policy. This can be enforced using Policies.

### NVIDIA Run:ai Namespace

The NVIDIA Run:ai must be installed in a namespace or project (OpenShift) called `runai`. Use the following to create the namespace/project:

<Tabs>

<Tab title="Kubernetes">

```bash
kubectl create ns runai
```

</Tab>

<Tab title="OpenShift">

```bash
oc new-project runai
```

</Tab>

</Tabs>

### Kubernetes Load Balancer

In Kubernetes, services of type `LoadBalancer` are used to expose applications outside the cluster through a single, stable IP address, providing a consistent entry point for external traffic. In managed cloud environments this capability is built-in, while in self-hosted and on-premise deployments it must be provided explicitly.

MetalLB fulfills this role by allocating external IP addresses from a predefined pool and advertising them on the external network, enabling access to services running inside the cluster.

In NVIDIA Run:ai, this is required to support north-south traffic, including access to the NVIDIA Run:ai control plane, APIs, UI, inference endpoints, and externally exposed development workspaces and training workloads.

1. Reserve a range of IP addresses (recommended a full 32 subnet) for example: `172.20.10.0-172.20.10.255`
2.  Install MetalLB:

    ```bash
    helm repo add metallb https://metallb.github.io/metallb
    helm repo update
    helm install metallb metallb/metallb --version 0.15.3 --namespace metallb-system --create-namespace
    ```
3.  Create a YAML file named `metalLB-config.yaml` and replace `<IPADDRESS-RANGE-START>-<IPADDRESS-RANGE-END>` with the reserved range of IP addresses:

    ```yaml
    apiVersion: metallb.io/v1beta1
    kind: IPAddressPool
    metadata:
      name: runai-ip-pool
      namespace: metallb-system
    spec:
      addresses:
      - <IPADDRESS-RANGE-START>-<IPADDRESS-RANGE-END> # for example 172.20.10.0-172.20.10.255
    ---
    apiVersion: metallb.io/v1beta1
    kind: L2Advertisement
    metadata:
      name: runai-l2-advertisement
      namespace: metallb-system
    spec:
      ipAddressPools:
      - runai-ip-pool
    ```
4.  Apply the YAML:

    ```bash
    kubectl apply -f metalLB-config.yaml
    ```

### Kubernetes Ingress Controller

NVIDIA Run:ai cluster requires [Kubernetes Ingress Controller](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/) to be installed on the Kubernetes cluster.

* OpenShift and RKE2 come with a pre-installed ingress controller.
* Make sure that a default ingress controller, `global.ingress.ingressClass` is set. For more details, see [Advanced cluster configurations](../../infrastructure-setup/advanced-setup/cluster-config.md).

There are many ways to install and configure different ingress controllers. The following provides a simple example to install and configure the [HAProxy Kubernetes Ingress Controller](https://www.haproxy.com/documentation/kubernetes-ingress/) using [helm](https://helm.sh/):

```bash
helm repo add haproxytech https://haproxytech.github.io/helm-charts
helm repo update
helm install haproxy-kubernetes-ingress haproxytech/kubernetes-ingress \
--create-namespace \
--namespace haproxy-controller \
--set controller.kind=DaemonSet \
--set controller.service.type=NodePort
```

### NVIDIA GPU Operator

The NVIDIA Run:ai cluster requires NVIDIA GPU Operator to be installed on the Kubernetes cluster. GPU Operator versions 25.10 to 26.3 are supported.

<Info>

**Note**

For [Multi-Node NVLink](https://app.gitbook.com/s/fvuZQSF7VCx1Azh0LF0H/platform-management/aiinitiatives/resources/using-gb200) support (e.g. GB200), GPU Operator 25.3 and above is required.

</Info>

See the [Installing the NVIDIA GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/getting-started.html), followed by notes below:

* Use the default `gpu-operator` namespace . Otherwise, you must specify the target namespace using the flag `runai-operator.config.nvidiaDcgmExporter.namespace` as described in customized cluster installation.
* NVIDIA drivers may already be installed on the nodes. In such cases, use the NVIDIA GPU Operator flags `--set driver.enabled=false`. [DGX OS](https://docs.nvidia.com/dgx/dgx-os-6-user-guide/release_notes.html) is one such example as it comes bundled with NVIDIA Drivers.
* For distribution-specific additional instructions see below:

<Accordion title="OpenShift Container Platform (OCP)">

**OpenShift Container Platform (OCP)**

The Node Feature Discovery (NFD) Operator is a prerequisite for the NVIDIA GPU Operator in OpenShift. Install the NFD Operator using the Red Hat OperatorHub catalog in the OpenShift Container Platform web console. For more information, see [Installing the Node Feature Discovery (NFD) Operator](https://docs.nvidia.com/datacenter/cloud-native/openshift/latest/install-nfd.html).

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

For troubleshooting information, see the [NVIDIA GPU Operator Troubleshooting Guide](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/troubleshooting.html).

### NVIDIA Network Operator

When deploying on clusters with RDMA or Multi-Node NVLink‑capable nodes (e.g. B200, GB200), the NVIDIA Network Operator is required to enable high-performance networking features such as GPUDirect RDMA in Kubernetes. Network Operator versions 25.10 to 26.1 are supported.

The Network Operator works alongside the NVIDIA GPU Operator to provide:

* NVIDIA networking drivers for advanced network capabilities.
* Kubernetes device plugins to expose high‑speed network hardware to workloads.
* Secondary network components to support network‑intensive applications.

The Network Operator must be installed and configured as follows:

1. Install the network operator as detailed in [Network Operator Deployment on Vanilla Kubernetes Cluster](https://docs.nvidia.com/networking/display/kubernetes2440/getting-started-kubernetes.html#network-operator-deployment-on-vanilla-kubernetes-cluster).
2. Configure SR-IOV InfiniBand support as detailed in [Network Operator Deployment with an SR-IOV InfiniBand Network](https://docs.nvidia.com/networking/display/kubernetes2440/getting-started-kubernetes.html#network-operator-deployment-with-an-sr-iov-infiniband-network).

### NVIDIA Dynamic Resource Allocation (DRA) Driver

When deploying on clusters with Multi-Node NVLink (e.g. GB200), the NVIDIA DRA driver is essential to enable Dynamic Resource Allocation at the Kubernetes level. To install, follow the instructions in [Configure and Helm-install the driver](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/dra-intro-install.html#configure-and-helm-install-the-driver). DRA driver versions 25.8 to 25.12 are supported.

After the DRA driver is installed, update `runaiconfig` using the `GPUNetworkAccelerationEnabled=True` flag to enable GPU network acceleration. This triggers an update of the NVIDIA Run:ai workload-controller deployment and restarts the controller. See [Advanced cluster configurations](../../infrastructure-setup/advanced-setup/cluster-config.md) for more details.

### Prometheus

<Info>

**Note**

Installing Prometheus applies for Kubernetes only.

</Info>

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

## Routing Traffic to and from NVIDIA Run:ai Services

This section describes how to route traffic to and from NVIDIA Run:ai services. Proper traffic routing is required to enable secure access to the NVIDIA Run:ai control plane (via the UI or API), as well as external access to development workspaces, training workloads, and inference workloads.

NVIDIA Run:ai supports two routing approaches for exposing services: host-based routing and path-based routing. While path-based routing exposes multiple services under a shared domain using URL paths, host-based routing assigns each service its own subdomain. Since many development tools and applications expect to run at the root path, host-based routing avoids common compatibility issues and is therefore used by default in NVIDIA Run:ai.

NVIDIA Run:ai uses host-based routing by default, which relies on DNS and TLS configuration to securely expose services. To support this, three key components must be configured:

* A [Fully qualified domain name (FQDN)](system-requirements.md#fully-qualified-domain-name-fqdn) structure
* [TLS certificates](system-requirements.md#tls-certificate) for secure communication
* [Host-based routing](system-requirements.md#host-based-routing-default) for workload exposure

Together, these components ensure that traffic is routed correctly and securely across all NVIDIA Run:ai services.

<Info>

**Note**

* NVIDIA Run:ai also supports path-based routing. If this approach better fits your environment, you can use it instead of the default host-based routing. In this case, the [workspace and training wildcard certificate](system-requirements.md#workspaces-and-training-workload-certificate) is not required.
* To use path-based routing, disable host-based routing by setting `clusterConfig.global.subdomainSupport: false` during Helm installation. See [Advanced cluster configurations](../../infrastructure-setup/advanced-setup/cluster-config.md).

</Info>

### Fully Qualified Domain Name (FQDN)

<Info>

**Note**

Fully Qualified Domain Name applies to Kubernetes only.

</Info>

NVIDIA Run:ai services rely on Fully Qualified Domain Names (FQDNs) to route traffic between system components and to expose workloads externally. In NVIDIA Run:ai, the FQDN settings are needed for:

* Enabling communication between the control plane and the cluster
* Exposing development workspaces and training workloads via subdomains
* Exposing inference workloads via dedicated subdomains

You must configure domain names for each of the following communication types:

* Control plane ↔ cluster communication\
  Example: `runai.mycorp.local`\
  The IP address of this domain must be resolvable within the organization’s private network.
* Workspace and training workloads (external access)\
  Example: `*.runai.mycorp.local`
* Inference workloads (external access)\
  Example: `*.runai-inference.mycorp.local`

<Info>

**Note**

The inference wildcard (`*.runai-inference.mycorp.local`) is only required if you plan to expose inference workloads outside the cluster. For cluster-local inference, skip the inference wildcard record below and the matching [Inference wildcard certificate](system-requirements.md#inference-wildcard-certificate). See [Inference](system-requirements.md#inference).

</Info>

Since NVIDIA Run:ai uses host-based routing, wildcard DNS records must be configured to enable external access to workloads.

Configure the following DNS records. Both records should resolve to the same cluster public IP address, or to the cluster’s load balancer IP in on-prem environments. This ensures that each workspace or workload is assigned a unique subdomain under the wildcard domains:

```bash
*.runai.mycorp.local → <cluster IP>
*.runai-inference.mycorp.local → <cluster IP>
```

### NVIDIA Run:ai TLS Certificates

TLS certificates secure communication between NVIDIA Run:ai components and enable HTTPS access to exposed services.

NVIDIA Run:ai requires three TLS certificates, each aligned with a specific domain:

* Cluster domain certificate
* Workspaces and training workload certificate
* Inference certificate (wildcard)

#### Cluster Domain Certificates (Single-Domain)

*   **Kubernetes** - To enable secure communication between the NVIDIA Run:ai control plane and the cluster, configure a TLS certificate associated with the cluster’s main domain (e.g. `runai.mycorp.local`). This certificate should be stored as a secret named `runai-cluster-domain-tls-secret` in the `runai` namespace.

    * Replace `/path/to/fullchain.pem` with the actual path to your TLS certificate.
    * Replace `/path/to/private.pem` with the actual path to your private key.

    ```bash
    kubectl create secret tls runai-cluster-domain-tls-secret -n runai \
      --cert /path/to/fullchain.pem \
      --key /path/to/private.pem
    ```
* **OpenShift** - NVIDIA Run:ai uses the OpenShift default Ingress router for serving. The TLS certificate configured for this router must be issued by a trusted CA. For more details, see the OpenShift documentation on [configuring certificates](https://docs.redhat.com/en/documentation/openshift_container_platform/4.18/html/security_and_compliance/configuring-certificates#replacing-default-ingress).

#### Workspaces & Training Workload Wildcard Certificate

<Info>

**Note**

For path-based routing, ignore this configuration and move to the next step.

</Info>

* **Kubernetes** - To allow secure access to workspace and training workloads via subdomains, configure a wildcard TLS certificate that matches the cluster domain (e.g. `*.runai.mycorp.local`). This certificate should be stored as a secret named `runai-cluster-domain-star-tls-secret` in the `runai` namespace.
  * Replace `/path/to/fullchain.pem` with the actual path to your TLS certificate.
  *   Replace `/path/to/private.pem` with the actual path to your private key.

      ```bash
      kubectl create secret tls runai-cluster-domain-star-tls-secret -n runai \
        --cert /path/to/fullchain.pem \
        --key /path/to/private.pem
      ```
* **OpenShift** - A wildcard TLS certificate for Workspace and Training workloads is not required. OpenShift Routes handle TLS termination for inference endpoints using the platform’s built-in routing and certificate management.

#### Inference Wildcard Certificate

<Info>

**Note**

The inference wildcard certificate is only required if inference services are exposed externally via wildcard FQDN. For cluster-local inference, skip this subsection. See [Inference](system-requirements.md#inference).

</Info>

*   **Kubernetes** - To securely expose inference services over HTTPS, configure a wildcard TLS certificate for the inference domain (e.g. `*.runai-inference.mycorp.local`). This certificate should be stored as a secret named `runai-cluster-inference-tls-secret` in the `knative-serving` namespace.

    * Replace `/path/to/fullchain.pem` with the actual path to your TLS certificate.
    * Replace `/path/to/private.pem` with the actual path to your private key.

    ```bash
    kubectl create secret tls runai-cluster-inference-tls-secret -n knative-serving \
        --cert /path/to/fullchain.pem \
        --key /path/to/private.pem
    ```
* **OpenShift** - A wildcard TLS certificate for Inference workloads is not required. OpenShift Routes handle TLS termination for inference endpoints using the platform’s built-in routing and certificate management.

### Host-Based Routing (Default)

<Info>

**Note**

* The following steps are required for Kubernetes only. For OpenShift, no additional configuration is required.
* NVIDIA Run:ai also support path-based routing. If you prefer to use it instead of the default host-based routing, disable host-based routing by setting `clusterConfig.global.subdomainSupport: false` during the Helm installation. See [Advanced cluster configurations](../../infrastructure-setup/advanced-setup/cluster-config.md).
* If you choose path-based routing, skip the below steps.

</Info>

Host-based routing binds together the configured domains (FQDN), TLS certificates, and ingress rules to expose workloads externally. With NVIDIA Run:ai host-based routing, workloads are exposed using subdomains, so each workload is assigned its own URL. For example:

```bash
https://<project>-<workload>.<CLUSTER_URL>
```

Host-based routing relies on:

* The FQDN structure defined earlier
* The TLS certificates configured for those domains

This section describes how to connect these components to enable workload exposure.

1. Ensure that:
   * Wildcard DNS records are configured (see [FQDN](system-requirements.md#fully-qualified-domain-name-fqdn) section)
   * TLS certificates are created and stored as Kubernetes secrets (see [TLS certificates](system-requirements.md#nvidia-run-ai-tls-certificates) section)
2.  Create the ingress resource

    ```yaml
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: runai-cluster-domain-star-ingress
      namespace: runai
    spec:
      rules:
      - host: '*.<CLUSTER_URL>'
      tls:
      - hosts:
        - '*.<CLUSTER_URL>'
        secretName: runai-cluster-domain-star-tls-secret
    ```
3.  Run the following:

    ```bash
    kubectl apply -f <filename>
    ```

## Additional Software Requirements

Additional NVIDIA Run:ai capabilities, Distributed Training and Inference require additional Kubernetes applications (frameworks) to be installed on the cluster.

### Distributed Training

Distributed training enables training of AI models over multiple nodes. This requires installing a distributed training framework on the cluster. The following frameworks are supported:

* [TensorFlow](https://www.tensorflow.org/)
* [PyTorch](https://pytorch.org/)
* [XGBoost](https://xgboost.readthedocs.io/)
* [MPI v2](https://docs.open-mpi.org/)
* [JAX](https://docs.jax.dev/en/latest/index.html)

There are several ways to install each framework. A simple method of installation example is the [Kubeflow Training Operator](https://www.kubeflow.org/docs/components/training/installation/) which includes TensorFlow, PyTorch, XGBoost and JAX.

It is recommended to use **Kubeflow Training Operator v1.9.2**, and **MPI Operator v0.6.0 or later** for compatibility with advanced workload capabilities, such as Stopping a workload and Scheduling rules.

* To install the Kubeflow Training Operator for TensorFlow, PyTorch, XGBoost and JAX frameworks, run the following command:

```bash
kubectl apply -k "github.com/kubeflow/training-operator.git/manifests/overlays/standalone?ref=v1.9.2"
```

* To install the MPI Operator for MPI v2, run the following command:

```bash
kubectl apply --server-side -f https://raw.githubusercontent.com/kubeflow/mpi-operator/v0.6.0/deploy/v2beta1/mpi-operator.yaml
```

<Info>

**Note**

If you require both the MPI Operator and Kubeflow Training Operator, follow the steps below:

* Install the Kubeflow Training Operator as described above.
* Disable and delete MPI v1 in the Kubeflow Training Operator by running:

```bash
kubectl patch deployment training-operator -n kubeflow --type='json' -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args", "value": ["--enable-scheme=tfjob", "--enable-scheme=pytorchjob", "--enable-scheme=xgboostjob", "--enable-scheme=jaxjob"]}]'
kubectl delete crd mpijobs.kubeflow.org
```

* Install the MPI Operator as described above.

</Info>

### Inference

Inference enables serving of AI models. This requires the [Knative Serving](https://knative.dev/docs/serving/) framework to be installed on the cluster and supports Knative versions 1.19 to 1.21.

To configure inference:

* **Inference request metrics (OTLP)** - For Knative Serving 1.19 or later, add the `observability` block to your `KnativeServing` YAML to export request metrics to the NVIDIA Run:ai API and UI. The OTLP receiver is enabled by default in NVIDIA Run:ai 2.25. See [Inference request metrics (OTLP)](system-requirements.md#inference-request-metrics-otlp) below.
* **Internal or external access** - Choose how inference services are reached: externally via a wildcard FQDN (the default) or cluster-local. See the matching section below.

#### Inference Request Metrics (OTLP)

To surface inference request metrics in the NVIDIA Run:ai API and UI, configure Knative Serving 1.19 or later to export request metrics over OpenTelemetry Protocol (OTLP) to the NVIDIA Run:ai Prometheus instance. The `observability` block included in the `KnativeServing` YAML examples below sets the required `config-observability` keys (`metrics-protocol`, `request-metrics-protocol`, `request-metrics-endpoint`). For supported fields, see [Enabling metric collection](https://knative.dev/docs/serving/observability/metrics/collecting-metrics/#enabling-metric-collection) in the Knative documentation.

The OTLP receiver on the NVIDIA Run:ai Prometheus instance is enabled by default.

To disable it, set `spec.prometheus.config.enableOTLPReceiver: false`. See [Advanced cluster configurations](../../infrastructure-setup/advanced-setup/cluster-config.md) for more details.

#### External-Access Inference

With external access, inference services are reachable from outside the cluster. Use this mode when inference workloads must be reachable by clients outside the cluster.

<Tabs>

<Tab title="Kubernetes">

On Kubernetes, external access exposes inference services through a wildcard FQDN (for example, `*.runai-inference.mycorp.local`). This requires the wildcard DNS record (see [FQDN](system-requirements.md#fully-qualified-domain-name-fqdn)), the [Inference wildcard certificate](system-requirements.md#inference-wildcard-certificate), and an HAProxy ingress to handle TLS termination.

Follow the [Installing Knative](https://knative.dev/v1.18-docs/install/operator/knative-with-operators/) instructions or run:

```bash
helm repo add knative-operator https://knative.github.io/operator
helm install knative-operator --create-namespace --namespace knativeoperator --version 1.18.2 knative-operator/knative-operator
```

Once installed, follow the below steps:

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

Follow the [Installing Knative](https://knative.dev/docs/install/operator/knative-with-operators/) instructions or run:

```bash
helm repo add knative-operator https://knative.github.io/operator
helm install knative-operator --create-namespace --namespace knativeoperator --version {VERSION} knative-operator/knative-operator
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

### Autoscaling

NVIDIA Run:ai allows for autoscaling a deployment according to the below metrics:

* Latency (milliseconds)
* Throughput (requests/sec)
* Concurrency (requests)

Using a custom metric (for example, Latency) requires installing the [Kubernetes Horizontal Pod Autoscaler (HPA)](https://knative.dev/docs/install/yaml-install/serving/install-serving-with-yaml/#install-optional-serving-extensions). Use the following command to install. Make sure to update the {VERSION} in the below command with a [supported Knative version](system-requirements.md#inference).

```bash
kubectl apply -f https://github.com/knative/serving/releases/download/knative-{VERSION}/serving-hpa.yaml
```

### Distributed Inference

NVIDIA Run:ai supports distributed inference (multi-node) deployments using the Leader Worker Set (LWS). To enable this capability, you must install the [LWS Helm chart](https://lws.sigs.k8s.io/docs/installation/#install-by-helm) in version 0.7.0 or higher on your cluster:

```bash
CHART_VERSION=0.7.0
helm install lws oci://registry.k8s.io/lws/charts/lws \
  --version=$CHART_VERSION \
  --namespace lws-system \
  --create-namespace \
  --wait --timeout 300s
```
