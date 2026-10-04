---
---

# Configuration Templates

Use these templates to start with a complete tenant cluster `config.yaml` instead of building one from scratch. Choose the template that best matches your workload types and networking setup, then replace the example values with those prepared for that tenant.

| Template | Use this when |
|---|---|
| [Full Cluster Installation](#full-cluster-installation) | You want every prerequisite and feature installed, covering both training and inference workloads. |
| [Training Installation](#training-installation) | The tenant plans to run training workloads and does not need inference-specific features. |
| [Inference Installation](#inference-installation) | The tenant plans to run inference workloads and does not need training-specific features. |
| [Gateway API Installation](#gateway-api-installation) | The tenant cluster uses Kubernetes Gateway API instead of an Ingress controller. |
| [OpenShift Installation](#openshift-installation) | You are installing the tenant cluster on OpenShift. |

## Before You Use a Template

Complete [Preparations](preparations.md) before using a template. This includes completing the applicable [System Requirements](../../system-requirements/README.md), [registering the cluster](create-a-tenant-and-cluster.md#register-the-cluster), and having the required domain, TLS certificates, networking values, and registry credentials ready.

* See [Create a Tenant and Cluster](create-a-tenant-and-cluster.md#install-the-cluster) to understand the minimum configuration needed before adapting a template.
* See [Prerequisites and Features](prerequisites-and-features.md) to understand the components required, depending on the workload types that tenant plans to run - each template installs a fixed set of prerequisites and features, and depending on which one you choose, you may need to adjust versions or add `customHelmValues` for your environment (for example, an explicit CUDA driver version if you plan to use GPU fractions).
* See [Configuration Reference](configuration-reference.md) for the full set of supported fields and descriptions.

## Full Cluster Installation

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
  name: <CLUSTER_NAME>
  domain: <CLUSTER_DOMAIN>
  workloadRouting: host
  controlPlaneUrl: https://<CONTROL_PLANE_DOMAIN>
  uid: <CLUSTER_UUID>
  bootstrapTokenPath: /path/to/bootstrap-token
  tls:
    certPath: /path/to/cluster-cert.pem
    keyPath: /path/to/cluster-key.pem
  inference:
    wildcardDomain: "*.<INFERENCE_WILDCARD_DOMAIN>"
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
    className: <YOUR_INGRESS_CLASS>
```

## Training Installation

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
  name: <CLUSTER_NAME>
  domain: <CLUSTER_DOMAIN>
  workloadRouting: host
  controlPlaneUrl: https://<CONTROL_PLANE_DOMAIN>
  uid: <CLUSTER_UUID>
  bootstrapTokenPath: /path/to/bootstrap-token
  tls:
    certPath: /path/to/cluster-cert.pem
    keyPath: /path/to/cluster-key.pem
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
    className: <YOUR_INGRESS_CLASS>
```

## Inference Installation

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
  name: <CLUSTER_NAME>
  domain: <CLUSTER_DOMAIN>
  workloadRouting: host
  controlPlaneUrl: https://<CONTROL_PLANE_DOMAIN>
  uid: <CLUSTER_UUID>
  bootstrapTokenPath: /path/to/bootstrap-token
  tls:
    certPath: /path/to/cluster-cert.pem
    keyPath: /path/to/cluster-key.pem
  inference:
    wildcardDomain: "*.<INFERENCE_WILDCARD_DOMAIN>"
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
    className: <YOUR_INGRESS_CLASS>
```

## Gateway API Installation

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
  name: <CLUSTER_NAME>
  domain: <CLUSTER_DOMAIN>
  controlPlaneUrl: https://<CONTROL_PLANE_DOMAIN>
  uid: <CLUSTER_UUID>
  bootstrapTokenPath: /path/to/bootstrap-token
  tls:
    certPath: /path/to/cluster-cert.pem
    keyPath: /path/to/cluster-key.pem
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
    name: <GATEWAY_NAME>
    namespace: <GATEWAY_NAMESPACE>
```

## OpenShift Installation

```yaml
install:
  mode: runai_cluster_only
  connectivity: connected
  kubernetesDistribution: openshift

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

cluster:
  enabled: true
  version: ""
  name: <CLUSTER_NAME>
  domain: <CLUSTER_DOMAIN>
  controlPlaneUrl: https://<CONTROL_PLANE_DOMAIN>
  uid: <CLUSTER_UUID>
  bootstrapTokenPath: /path/to/bootstrap-token
  inference:
    wildcardDomain: "*.<INFERENCE_WILDCARD_DOMAIN>"
  customHelmValues: {}

prerequisites:
  skip: false
  gpuOperator:
    mode: install
    version: ""
  prometheus:
    mode: skip
  networkOperator:
    mode: install
    version: ""

features:
  skip: false
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
```
