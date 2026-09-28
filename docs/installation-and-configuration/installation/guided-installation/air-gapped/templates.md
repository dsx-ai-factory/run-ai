---
---

# Configuration Templates

Use these templates to start with a complete `config.yaml` instead of building one from scratch. Choose the template that best matches your Kubernetes environment, workloads, and networking setup, then replace the example values with those prepared for your environment.

The templates below are for same-cluster installations. For a multi-cluster installation, see [Use a Template for Multi-Cluster](#use-a-template-for-multi-cluster).

| Template | Use this when |
|---|---|
| [Full Installation](#full-installation) | You want both training and inference capabilities and all supported prerequisites/features included. |
| [Gateway API Installation](#gateway-api-installation) | Your environment uses Kubernetes Gateway API instead of an Ingress controller. |
| [OpenShift Installation](#openshift-installation) | You are installing NVIDIA Run:ai on OpenShift. |

## Before You Use a Template

Complete [Preparations](preparations.md) before using a template. This includes completing the applicable [System Requirements](../../../system-requirements/README.md) and having the required domains, TLS certificates, networking values, registry credentials, and other installation inputs ready.

<Note>

**Note**

See the [Configuration Reference](../configuration-reference.md) for what every field accepts.

</Note>

## Full Installation

```yaml
install:
  mode: full_installation
  connectivity: air_gapped

registry:
  host: <REGISTRY_HOST>
  username: <REGISTRY_USERNAME>
  passwordPath: /path/to/registry-password

airgap:
  workingDir: /data/runai-airgap
  valuesDir: /data/runai-airgap
  skipPrepare: true

controlPlane:
  enabled: true
  domain: <DOMAIN>
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues: {}

cluster:
  enabled: true
  name: <CLUSTER_NAME>
  tls:
    certPath: /path/to/cluster-cert.pem
    keyPath: /path/to/cluster-key.pem
  customHelmValues: {}

prerequisites:
  gpuOperator:
    mode: skip
    customHelmValues: {}
  prometheus:
    mode: skip
    customHelmValues: {}
  networkOperator:
    mode: skip
    customHelmValues: {}

externalConnection:
  type: ingress
  ingress:
    className: <YOUR_INGRESS_CLASS>
```

## Use a Template for Multi-Cluster

To use a template in a multi-cluster deployment, split it into two config files instead of copying it as-is. The example below splits the [Full Installation](#full-installation) template into a control plane config and a cluster config:

`cp-config.yaml`:

```yaml
install:
  mode: runai_control_plane_only
  connectivity: air_gapped

registry:
  host: <REGISTRY_HOST>
  username: <REGISTRY_USERNAME>
  passwordPath: /path/to/registry-password

airgap:
  workingDir: /data/runai-airgap
  valuesDir: /data/runai-airgap
  skipPrepare: true

controlPlane:
  enabled: true
  domain: <DOMAIN>
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues: {}

externalConnection:
  type: ingress
  ingress:
    className: <YOUR_INGRESS_CLASS>
```

`cluster-config.yaml`:

```yaml
install:
  mode: runai_cluster_only
  connectivity: air_gapped

registry:
  host: <REGISTRY_HOST>
  username: <REGISTRY_USERNAME>
  passwordPath: /path/to/registry-password

airgap:
  workingDir: /data/runai-airgap
  valuesDir: /data/runai-airgap
  skipPrepare: true

cluster:
  enabled: true
  name: <CLUSTER_NAME>
  domain: <CLUSTER_DOMAIN>
  controlPlaneUrl: https://<DOMAIN>
  controlPlaneCredentials:
    username: <ADMIN_EMAIL>
    passwordPath: /path/to/cp-password
  tls:
    certPath: /path/to/cluster-cert.pem
    keyPath: /path/to/cluster-key.pem
  customHelmValues: {}

externalConnection:
  type: ingress
  ingress:
    className: <YOUR_INGRESS_CLASS>
```

## Gateway API Installation

```yaml
install:
  mode: full_installation
  connectivity: air_gapped

registry:
  host: <REGISTRY_HOST>
  username: <REGISTRY_USERNAME>
  passwordPath: /path/to/registry-password

airgap:
  workingDir: /data/runai-airgap
  valuesDir: /data/runai-airgap
  skipPrepare: true

controlPlane:
  enabled: true
  domain: <DOMAIN>
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues: {}

cluster:
  enabled: true
  name: <CLUSTER_NAME>
  tls:
    certPath: /path/to/cluster-cert.pem
    keyPath: /path/to/cluster-key.pem
  customHelmValues: {}

prerequisites:
  gpuOperator:
    mode: skip
    customHelmValues: {}
  prometheus:
    mode: skip
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
  mode: full_installation
  connectivity: air_gapped
  kubernetesDistribution: openshift

registry:
  host: <REGISTRY_HOST>
  username: <REGISTRY_USERNAME>
  passwordPath: /path/to/registry-password

airgap:
  workingDir: /data/runai-airgap
  valuesDir: /data/runai-airgap
  skipPrepare: true

openshift:
  olmCatalog:
    source: <OPERATORHUB_CATALOG_SOURCE>
    sourceNamespace: openshift-marketplace

controlPlane:
  enabled: true
  domain: <DOMAIN>
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues: {}

cluster:
  enabled: true
  name: <CLUSTER_NAME>
  tls:
    certPath: /path/to/cluster-cert.pem
    keyPath: /path/to/cluster-key.pem
  customHelmValues: {}

prerequisites:
  gpuOperator:
    mode: skip
  prometheus:
    mode: skip
```

