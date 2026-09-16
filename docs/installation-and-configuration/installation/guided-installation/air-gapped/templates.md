---
---

# Configuration Templates

Ready-to-use configuration YAMLs for air-gapped environments. Each template is a complete, standalone `config.yaml` for a same-cluster installation (`install.mode: full_installation`) - copy the one that matches your use case. For a multi-cluster deployment, split the template into two configs - see [Splitting a Template for Multi-Cluster](#splitting-a-template-for-multi-cluster) below for a worked example.

<Info>

**Note**

See the [Configuration Reference](../configuration-reference.md) for what every field accepts.

</Info>

## Full Install (Same-Cluster)

Deploys the control plane and cluster together on the same Kubernetes cluster, pulling images from your private mirror registry and routing external traffic through an Ingress controller (`haproxy`). Before applying, replace the domain, TLS certificate paths, private mirror registry, and air-gapped bundle working directory with your own values.

```yaml
install:
  mode: full_installation
  connectivity: air_gapped

registry:
  host: registry.mycorp.local
  username: admin
  passwordPath: /path/to/registry-password

airgap:
  workingDir: /data/runai-airgap
  valuesDir: /data/runai-airgap
  skipPrepare: true

controlPlane:
  enabled: true
  domain: runai.mycorp.local
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues: {}

cluster:
  enabled: true
  name: my-cluster
  tls:
    certPath: /path/to/cluster-wildcard-cert.pem
    keyPath: /path/to/cluster-wildcard-key.pem
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
    className: haproxy
```

## Splitting a Template for Multi-Cluster

To use a template above in a multi-cluster deployment, split it into two config files instead of copying it as-is. The example below splits the [Full Install (Same-Cluster)](#full-install-same-cluster) template into a control plane config and a cluster config:

- The control plane config keeps `install`, `registry`, `airgap`, `controlPlane`, and `externalConnection`, and drops `cluster` and `prerequisites` entirely - `cluster` belongs to the cluster config, and prerequisites are installed manually in air-gapped environments rather than through either config.
- The cluster config keeps `install`, `registry`, `airgap`, and `cluster`. Its `cluster` block also needs `cluster.domain` and its own TLS certificate for that domain, since the cluster is now reachable at its own FQDN rather than sharing the control plane's Kubernetes cluster. Both configs point at the same air-gapped bundle.

`cp-config.yaml`:

```yaml
install:
  mode: runai_control_plane_only
  connectivity: air_gapped

registry:
  host: registry.mycorp.local
  username: admin
  passwordPath: /path/to/registry-password

airgap:
  workingDir: /data/runai-airgap
  valuesDir: /data/runai-airgap
  skipPrepare: true

controlPlane:
  enabled: true
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
  connectivity: air_gapped

registry:
  host: registry.mycorp.local
  username: admin
  passwordPath: /path/to/registry-password

airgap:
  workingDir: /data/runai-airgap
  valuesDir: /data/runai-airgap
  skipPrepare: true

cluster:
  enabled: true
  name: my-cluster
  domain: runai-cluster.mycorp.local
  tls:
    certPath: /path/to/cluster-cert.pem
    keyPath: /path/to/cluster-key.pem
  customHelmValues: {}

externalConnection:
  type: ingress
  ingress:
    className: haproxy
```

Apply the same split to the Gateway API template below - move `cluster` into the cluster config, add `cluster.domain` and its TLS certificate, and keep `install`, `registry`, `airgap`, `controlPlane`, and `externalConnection` in the control plane config.

## Gateway API

Deploys the control plane and cluster together on the same Kubernetes cluster, pulling images from your private mirror registry and routing external traffic through the Kubernetes Gateway API instead of Ingress. Before applying, replace the domain, TLS certificate paths, private mirror registry, and air-gapped bundle working directory with your own values.

```yaml
install:
  mode: full_installation
  connectivity: air_gapped

registry:
  host: registry.mycorp.local
  username: admin
  passwordPath: /path/to/registry-password

airgap:
  workingDir: /data/runai-airgap
  valuesDir: /data/runai-airgap
  skipPrepare: true

controlPlane:
  enabled: true
  domain: runai.mycorp.local
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues: {}

cluster:
  enabled: true
  name: my-cluster
  tls:
    certPath: /path/to/cluster-wildcard-cert.pem
    keyPath: /path/to/cluster-wildcard-key.pem
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
    name: my-gateway
    namespace: gateway-system
```

## OpenShift

Deploys the control plane and cluster together on the same OpenShift cluster (`install.kubernetesDistribution: openshift`) from an air-gapped bundle. Unlike the templates above, this one omits `externalConnection` entirely - OpenShift Routes are configured automatically. It also sets `openshift.olmCatalog` to point OLM-based operator installs at your mirrored OperatorHub catalog, since the public catalogs aren't reachable in an air-gapped environment. GPU Operator and Prometheus are installed manually, using each vendor's own air-gapped OpenShift instructions - see [Prerequisites and Features](preparations.md#prerequisites-and-features) - and the Node Feature Discovery (NFD) Operator, a prerequisite for the GPU Operator on OpenShift, is also installed manually through your mirrored catalog. Before applying, replace the domain, TLS certificate paths, private mirror registry, air-gapped bundle working directory, and OperatorHub catalog source with your own values.

```yaml
install:
  mode: full_installation
  connectivity: air_gapped
  kubernetesDistribution: openshift

registry:
  host: registry.mycorp.local
  username: admin
  passwordPath: /path/to/registry-password

airgap:
  workingDir: /data/runai-airgap
  valuesDir: /data/runai-airgap
  skipPrepare: true

openshift:
  olmCatalog:
    source: my-mirrored-catalog
    sourceNamespace: openshift-marketplace

controlPlane:
  enabled: true
  domain: runai.mycorp.local
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues: {}

cluster:
  enabled: true
  name: my-cluster
  tls:
    certPath: /path/to/cluster-wildcard-cert.pem
    keyPath: /path/to/cluster-wildcard-key.pem
  customHelmValues: {}

prerequisites:
  gpuOperator:
    mode: skip
  prometheus:
    mode: skip
```

