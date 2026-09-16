---
---

# Configuration Templates

Ready-to-use configuration YAMLs for internet-connected environments. Each template is a complete, standalone `config.yaml` for a same-cluster installation (`install.mode: full_installation`) - copy the one that matches your use case. For a multi-cluster deployment, split the template into two configs - see [Splitting a Template for Multi-Cluster](#splitting-a-template-for-multi-cluster) below for a worked example.

<Note>

**Note**

See the [Configuration Reference](../configuration-reference.md) for what every field accepts.

</Note>

## Full Install (Same-Cluster)

Deploys the control plane and cluster on the same Kubernetes cluster with every prerequisite and feature enabled, routing external traffic through an Ingress controller (`haproxy`). See [Prerequisites and Features](../prerequisites-and-features.md). Replace the domain, TLS certificate paths, and (Support-Entitled only) registry credentials before applying.

```yaml
install:
  mode: full_installation
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

controlPlane:
  enabled: true
  version: ""
  domain: runai.mycorp.local
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues: {}

cluster:
  enabled: true
  version: ""
  name: my-cluster
  tls:
    certPath: /path/to/cluster-wildcard-cert.pem
    keyPath: /path/to/cluster-wildcard-key.pem
  inference:
    wildcardDomain: "*.runai-inference.mycorp.local"
    wildcardTls:
      certPath: /path/to/inference-wildcard-cert.pem
      keyPath: /path/to/inference-wildcard-key.pem
  customHelmValues: {}

prerequisites:
  gpuOperator:
    mode: install
    version: ""
    customHelmValues: {}
  prometheus:
    mode: install
    version: ""
    customHelmValues: {}
  networkOperator:
    mode: install
    version: ""
    customHelmValues: {}

features:
  kubeflow:
    mode: install
    version: ""
  mpiOperator:
    mode: install
    version: ""
  knative:
    mode: install
    version: ""
    customHelmValues: {}
  lws:
    mode: install
    version: ""
    customHelmValues: {}
  nimOperator:
    mode: install
    version: ""
    customHelmValues: {}
  grove:
    mode: install
    version: ""
    customHelmValues: {}
  dynamo:
    mode: install
    version: ""
    customHelmValues: {}

externalConnection:
  type: ingress
  ingress:
    className: haproxy
```

## Splitting a Template for Multi-Cluster

To use a template above in a multi-cluster deployment, split it into two config files instead of copying it as-is. The example below splits the [Full Install (Same-Cluster)](#full-install-same-cluster) template into a control plane config and a cluster config:

- The control plane config keeps `install`, `registry`, `controlPlane`, and `externalConnection`, and drops `cluster`, `prerequisites`, and `features` entirely - those belong to the cluster config.
- The cluster config keeps `install`, `registry`, `cluster`, `prerequisites`, `features`, and `externalConnection`. Its `cluster` block also needs `cluster.domain` and its own TLS certificate for that domain, since the cluster is now reachable at its own FQDN rather than sharing the control plane's Kubernetes cluster.

`cp-config.yaml`:

```yaml
install:
  mode: runai_control_plane_only
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

controlPlane:
  enabled: true
  version: ""
  domain: runai.mycorp.local
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues: {}

externalConnection:
  type: ingress
  ingress:
    className: haproxy
```

`cluster-config.yaml`:

```yaml
install:
  mode: runai_cluster_only
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

cluster:
  enabled: true
  version: ""
  name: my-cluster
  domain: runai-cluster.mycorp.local
  tls:
    certPath: /path/to/cluster-cert.pem
    keyPath: /path/to/cluster-key.pem
  inference:
    wildcardDomain: "*.runai-inference.mycorp.local"
    wildcardTls:
      certPath: /path/to/inference-wildcard-cert.pem
      keyPath: /path/to/inference-wildcard-key.pem
  customHelmValues: {}

prerequisites:
  gpuOperator:
    mode: install
    version: ""
    customHelmValues: {}
  prometheus:
    mode: install
    version: ""
    customHelmValues: {}
  networkOperator:
    mode: install
    version: ""
    customHelmValues: {}

features:
  kubeflow:
    mode: install
    version: ""
  mpiOperator:
    mode: install
    version: ""
  knative:
    mode: install
    version: ""
    customHelmValues: {}
  lws:
    mode: install
    version: ""
    customHelmValues: {}
  nimOperator:
    mode: install
    version: ""
    customHelmValues: {}
  grove:
    mode: install
    version: ""
    customHelmValues: {}
  dynamo:
    mode: install
    version: ""
    customHelmValues: {}

externalConnection:
  type: ingress
  ingress:
    className: haproxy
```

Apply the same split to any other template on this page - move `cluster`, `prerequisites`, and `features` into the cluster config, add `cluster.domain` and its TLS certificate, and keep `install`, `registry`, `controlPlane`, and `externalConnection` in the control plane config.

## Training Deployment Only

Enables the GPU Operator and Prometheus prerequisites and the Kubeflow and MPI Operator features, for training workloads only. See [Prerequisites and Features](../prerequisites-and-features.md). Replace the domain, TLS certificate paths, and (Support-Entitled only) registry credentials before applying.

```yaml
install:
  mode: full_installation
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

controlPlane:
  enabled: true
  version: ""
  domain: runai.mycorp.local
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues: {}

cluster:
  enabled: true
  version: ""
  name: my-cluster
  tls:
    certPath: /path/to/cluster-wildcard-cert.pem
    keyPath: /path/to/cluster-wildcard-key.pem
  customHelmValues: {}

prerequisites:
  gpuOperator:
    mode: install
    version: ""
    customHelmValues: {}
  prometheus:
    mode: install
    version: ""
    customHelmValues: {}

features:
  kubeflow:
    mode: install
    version: ""
  mpiOperator:
    mode: install
    version: ""

externalConnection:
  type: ingress
  ingress:
    className: haproxy
```

## Inference Deployment Only

Enables the GPU Operator and Prometheus prerequisites and the Knative, LWS, NIM Operator, Grove, and Dynamo features, for inference workloads only. See [Prerequisites and Features](../prerequisites-and-features.md). Replace the domain, TLS certificate paths, and (Support-Entitled only) registry credentials before applying.

```yaml
install:
  mode: full_installation
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

controlPlane:
  enabled: true
  version: ""
  domain: runai.mycorp.local
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues: {}

cluster:
  enabled: true
  version: ""
  name: my-cluster
  tls:
    certPath: /path/to/cluster-wildcard-cert.pem
    keyPath: /path/to/cluster-wildcard-key.pem
  inference:
    wildcardDomain: "*.runai-inference.mycorp.local"
    wildcardTls:
      certPath: /path/to/inference-wildcard-cert.pem
      keyPath: /path/to/inference-wildcard-key.pem
  customHelmValues: {}

prerequisites:
  gpuOperator:
    mode: install
    version: ""
    customHelmValues: {}
  prometheus:
    mode: install
    version: ""
    customHelmValues: {}

features:
  knative:
    mode: install
    version: ""
    customHelmValues: {}
  lws:
    mode: install
    version: ""
    customHelmValues: {}
  nimOperator:
    mode: install
    version: ""
    customHelmValues: {}
  grove:
    mode: install
    version: ""
    customHelmValues: {}
  dynamo:
    mode: install
    version: ""
    customHelmValues: {}

externalConnection:
  type: ingress
  ingress:
    className: haproxy
```

## Gateway API

Installs the GPU Operator and Prometheus prerequisites with no features enabled, routing external traffic through the Kubernetes Gateway API instead of Ingress. See [Prerequisites and Features](../prerequisites-and-features.md). Replace the domain, TLS certificate paths, Gateway name/namespace, and (Support-Entitled only) registry credentials before applying.

```yaml
install:
  mode: full_installation
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

controlPlane:
  enabled: true
  version: ""
  domain: runai.mycorp.local
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues: {}

cluster:
  enabled: true
  version: ""
  name: my-cluster
  tls:
    certPath: /path/to/cluster-wildcard-cert.pem
    keyPath: /path/to/cluster-wildcard-key.pem
  customHelmValues: {}

prerequisites:
  gpuOperator:
    mode: install
    version: ""
    customHelmValues: {}
  prometheus:
    mode: install
    version: ""
    customHelmValues: {}

externalConnection:
  type: gatewayApi
  gatewayApi:
    name: my-gateway
    namespace: gateway-system
```

## OpenShift

Deploys the control plane and cluster on the same OpenShift cluster (`install.kubernetesDistribution: openshift`), with the training and inference prerequisites and features enabled. Omits `externalConnection` - OpenShift Routes are configured automatically. Sets `prerequisites.prometheus.mode: skip`, since OpenShift's built-in Cluster Monitoring Operator replaces the installer-managed Prometheus stack. Install the [Node Feature Discovery (NFD) Operator](../prerequisites-and-features.md#nvidia-gpu-operator) manually through OperatorHub first - a GPU Operator prerequisite on OpenShift not installed by `runai-installer`. Replace the domain, TLS certificate paths, and (Support-Entitled only) registry credentials before applying.

```yaml
install:
  mode: full_installation
  connectivity: connected
  kubernetesDistribution: openshift

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

controlPlane:
  enabled: true
  version: ""
  domain: runai.mycorp.local
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem

cluster:
  enabled: true
  version: ""
  name: my-cluster
  tls:
    certPath: /path/to/cluster-wildcard-cert.pem
    keyPath: /path/to/cluster-wildcard-key.pem
  inference:
    wildcardDomain: "*.runai-inference.mycorp.local"
    wildcardTls:
      certPath: /path/to/inference-wildcard-cert.pem
      keyPath: /path/to/inference-wildcard-key.pem

prerequisites:
  gpuOperator:
    mode: install
    version: ""
  prometheus:
    mode: skip
  networkOperator:
    mode: install
    version: ""

features:
  kubeflow:
    mode: install
    version: ""
  mpiOperator:
    mode: install
    version: ""
  knative:
    mode: install
    version: ""
  lws:
    mode: install
    version: ""
  nimOperator:
    mode: install
    version: ""
  grove:
    mode: install
    version: ""
  dynamo:
    mode: install
    version: ""
```

