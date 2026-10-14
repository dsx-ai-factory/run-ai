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
| [Gov-Ready Installation](#gov-ready-installation) | You are installing a Gov-ready environment on Kubernetes. |

## Before You Use a Template

Complete [Preparations](preparations.md) before using a template. This includes completing the applicable [System Requirements](../../../system-requirements/README.md) and having the required domains, TLS certificates, networking values, registry credentials, and other installation inputs ready.

* See [Same-Cluster Installation](same-cluster.md#create-a-configuration-file) or [Multi-Cluster Installation](multi-cluster.md), whichever matches your deployment model, to understand the minimum configuration needed before adapting a template.
* See [Prerequisites and Features](../prerequisites-and-features.md) to understand the components required, depending on the workload types you plan to enable, and how to install them yourself in air-gapped environments.
* See [Configuration Reference](../configuration-reference.md) for the full set of supported fields and descriptions.

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
  customHelmValues:
    supportEntitled: true # Support-Entitled Release only

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
  customHelmValues:
    supportEntitled: true # Support-Entitled Release only

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
  customHelmValues:
    supportEntitled: true # Support-Entitled Release only

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

controlPlane:
  enabled: true
  domain: <DOMAIN>
  customHelmValues:
    supportEntitled: true # Support-Entitled Release only

cluster:
  enabled: true
  name: <CLUSTER_NAME>
  customHelmValues: {}

prerequisites:
  gpuOperator:
    mode: skip
  prometheus:
    mode: skip
```

## Gov-Ready Installation

Use this template for a Gov-ready same-cluster installation on Kubernetes. It includes FIPS mode, an external PostgreSQL instance, and the Keycloak and NATS settings that the hardened images require. The image values for the third-party services are not part of the template, because you add the hardened images to the pack configuration and `airgap load` generates the image values. For what each value does, refer to [Gov-Ready Third-Party Services](../../../system-requirements/gov-ready-third-party-services.md). For the overall flow, refer to [Gov-Ready Installation](../../gov-ready-installation.md).

To use the template in a multi-cluster deployment, split it as described in [Use a Template for Multi-Cluster](#use-a-template-for-multi-cluster): the `controlPlane` values go in the control plane config and the `cluster` values go in the cluster config.

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
  customHelmValues:
    supportEntitled: true # Support-Entitled Release only
    global.fipsMode: only
    global.postgresql.auth.host: <POSTGRESQL_HOST>
    global.postgresql.auth.port: <POSTGRESQL_PORT>
    global.postgresql.auth.database: <DATABASE_NAME>
    global.postgresql.auth.username: <ROLE_NAME>
    global.postgresql.auth.password: <ROLE_PASSWORD> # Or use existingSecret
    global.postgresql.auth.postgresPassword: <ANY_VALUE> # Required, not used with an external instance
    global.postgresql.auth.dbSslMode: <SSL_MODE>
    postgresql.enabled: false
    keycloakx.securityContext.runAsUser: 65532
    keycloakx.command[0]: /bin/sh
    keycloakx.command[1]: -c
    keycloakx.command[2]: >-
      KC_BOOTSTRAP_ADMIN_USERNAME="$(cat /run/secrets/keycloak/admin-username)"
      KC_BOOTSTRAP_ADMIN_PASSWORD="$(cat /run/secrets/keycloak/admin-password)"
      exec /opt/keycloak/bin/kc.sh start
      --features=token-exchange\,admin-fine-grained-authz:v1\,fips
    nats.command[0]: /nats-server

cluster:
  enabled: true
  name: <CLUSTER_NAME>
  tls:
    certPath: /path/to/cluster-cert.pem
    keyPath: /path/to/cluster-key.pem
  customHelmValues:
    global.fipsMode: only

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
