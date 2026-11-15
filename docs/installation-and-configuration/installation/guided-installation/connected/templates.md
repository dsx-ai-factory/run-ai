---
---

# Configuration Templates

Use these templates to start with a complete `config.yaml` instead of building one from scratch. Choose the template that best matches your Kubernetes environment, workloads, and networking setup, then replace the example values with those prepared for your environment.

The templates below are for same-cluster installations. For a multi-cluster installation, see [Use a Template for Multi-Cluster](#use-a-template-for-multi-cluster).

| Template | Use this when |
|---|---|
| [Full Installation](#full-installation) | You want both training and inference capabilities and all supported prerequisites/features included. |
| [Training Installation](#training-installation) | You plan to run training workloads and do not need inference-specific features. |
| [Inference Installation](#inference-installation) | You plan to run inference workloads and do not need training-specific features. |
| [Gateway API Installation](#gateway-api-installation) | Your environment uses Kubernetes Gateway API instead of an Ingress controller. |
| [OpenShift Installation](#openshift-installation) | You are installing NVIDIA Run:ai on OpenShift. |
| [Gov-Ready Installation](#gov-ready-installation) | You are installing a Gov-ready environment on Kubernetes. |

## Before You Use a Template

Complete [Preparations](preparations.md) before using a template. This includes completing the applicable [System Requirements](../../../system-requirements/README.md) and having the required domains, TLS certificates, networking values, registry credentials, and other installation inputs ready.

* See [Same-Cluster Installation](same-cluster.md#create-a-configuration-file) or [Multi-Cluster Installation](multi-cluster.md), whichever matches your deployment model, to understand the minimum configuration needed before adapting a template.
* See [Prerequisites and Features](../prerequisites-and-features.md) to understand the components required, depending on the workload types you plan to enable - each template installs a fixed set of prerequisites and features, and depending on which one you choose, you may need to adjust versions or add `customHelmValues` for your environment (for example, an explicit CUDA driver version if you plan to use GPU fractions).
* See [Configuration Reference](../configuration-reference.md) for the full set of supported fields and descriptions.

## Full Installation

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
  domain: <DOMAIN>
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues:
    supportEntitled: true # Support-Entitled Release only

cluster:
  enabled: true
  version: ""
  name: <CLUSTER_NAME>
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

## Use a Template for Multi-Cluster

To use a template in a multi-cluster deployment, split it into two config files instead of copying it as-is. The example below splits the [Full Installation](#full-installation) template into a control plane config and a cluster config:

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
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

cluster:
  enabled: true
  version: ""
  name: <CLUSTER_NAME>
  domain: <CLUSTER_DOMAIN>
  controlPlaneUrl: https://<DOMAIN>
  controlPlaneCredentials:
    username: <ADMIN_EMAIL>
    passwordPath: /path/to/cp-password
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
  mode: full_installation
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

controlPlane:
  enabled: true
  version: ""
  domain: <DOMAIN>
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues:
    supportEntitled: true # Support-Entitled Release only

cluster:
  enabled: true
  version: ""
  name: <CLUSTER_NAME>
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
  mode: full_installation
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

controlPlane:
  enabled: true
  version: ""
  domain: <DOMAIN>
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues:
    supportEntitled: true # Support-Entitled Release only

cluster:
  enabled: true
  version: ""
  name: <CLUSTER_NAME>
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
  mode: full_installation
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

controlPlane:
  enabled: true
  version: ""
  domain: <DOMAIN>
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues:
    supportEntitled: true # Support-Entitled Release only

cluster:
  enabled: true
  version: ""
  name: <CLUSTER_NAME>
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
  mode: full_installation
  connectivity: connected
  kubernetesDistribution: openshift

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

controlPlane:
  enabled: true
  version: ""
  domain: <DOMAIN>
  customHelmValues:
    supportEntitled: true # Support-Entitled Release only

cluster:
  enabled: true
  version: ""
  name: <CLUSTER_NAME>
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

## Gov-Ready Installation

Use this template for a Gov-ready same-cluster installation on Kubernetes. It includes FIPS mode, the Docker Hub image pull secret, an external PostgreSQL instance, and the hardened images of the third-party services. For what each value does, and how to get access to the images, refer to [Gov-Ready Third-Party Services](../../../system-requirements/gov-ready-third-party-services.md). For the overall flow, refer to [Gov-Ready Installation](../../gov-ready-installation.md).

To use the template in a multi-cluster deployment, split it as described in [Use a Template for Multi-Cluster](#use-a-template-for-multi-cluster): the `controlPlane` values go in the control plane config and the `cluster` values go in the cluster config.

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
  domain: <DOMAIN>
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  customHelmValues:
    supportEntitled: true # Support-Entitled Release only
    global.fipsMode: only
    global.imagePullSecrets[0].name: runai-reg-creds
    global.imagePullSecrets[1].name: <dhi-secret-name>
    global.postgresql.auth.host: <POSTGRESQL_HOST>
    global.postgresql.auth.port: <POSTGRESQL_PORT>
    global.postgresql.auth.database: <DATABASE_NAME>
    global.postgresql.auth.username: <ROLE_NAME>
    global.postgresql.auth.password: <ROLE_PASSWORD> # Or use existingSecret
    global.postgresql.auth.postgresPassword: <ANY_VALUE> # Required, not used with an external instance
    global.postgresql.auth.dbSslMode: <SSL_MODE>
    global.postgresImage.registry: docker.io
    global.postgresImage.repository: <organization>/dhi-postgres
    global.postgresImage.tag: <fips-ready-tag>
    postgresql.enabled: false
    keycloakx.image.repository: docker.io/<organization>/dhi-keycloak
    keycloakx.image.tag: <fips-ready-tag>
    keycloakx.imagePullSecrets[0].name: runai-reg-creds
    keycloakx.imagePullSecrets[1].name: <dhi-secret-name>
    keycloakx.securityContext.runAsUser: 65532
    keycloakx.command[0]: /bin/sh
    keycloakx.command[1]: -c
    keycloakx.command[2]: >-
      KC_BOOTSTRAP_ADMIN_USERNAME="$(cat /run/secrets/keycloak/admin-username)"
      KC_BOOTSTRAP_ADMIN_PASSWORD="$(cat /run/secrets/keycloak/admin-password)"
      exec /opt/keycloak/bin/kc.sh start
      --features=token-exchange\,admin-fine-grained-authz:v1\,fips
    traefik.image.registry: docker.io
    traefik.image.repository: <organization>/dhi-traefik
    traefik.image.tag: <fips-ready-tag>
    nats.image.registry: docker.io
    nats.image.repository: <organization>/dhi-nats
    nats.image.tag: <fips-ready-tag>
    nats.command[0]: /nats-server
    thanos.image.registry: docker.io
    thanos.image.repository: <organization>/dhi-thanos
    thanos.image.tag: <fips-ready-tag>

cluster:
  enabled: true
  version: ""
  name: <CLUSTER_NAME>
  tls:
    certPath: /path/to/cluster-cert.pem
    keyPath: /path/to/cluster-key.pem
  inference:
    wildcardDomain: "*.<INFERENCE_WILDCARD_DOMAIN>"
    wildcardTls:
      certPath: /path/to/inference-wildcard-cert.pem
      keyPath: /path/to/inference-wildcard-key.pem
  customHelmValues:
    global.fipsMode: only
    global.additionalImagePullSecrets[0]: runai-reg-creds
    global.additionalImagePullSecrets[1]: <dhi-secret-name>
    thirdPartyImages.valkey.registry: docker.io
    thirdPartyImages.valkey.repository: <organization>/dhi-valkey
    thirdPartyImages.valkey.tag: <fips-ready-tag>
    thirdPartyImages.git-sync.registry: docker.io
    thirdPartyImages.git-sync.repository: <organization>/dhi-git-sync
    thirdPartyImages.git-sync.tag: <fips-ready-tag>

prerequisites:
  gpuOperator:
    mode: install
    version: ""
    customHelmValues: {} # Gov-ready: add the GPU Operator Helm values
  prometheus:
    mode: install
    version: ""
    customHelmValues: {}
  networkOperator:
    mode: install
    version: ""
    customHelmValues: {} # Gov-ready: add the Network Operator Helm values

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
