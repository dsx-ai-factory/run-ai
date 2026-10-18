# Kubernetes Gateway API

NVIDIA Run:ai supports the Kubernetes Gateway API as an alternative to Ingress for routing external traffic. Gateway API provides a flexible and extensible model for defining how traffic is exposed and routed within the cluster.

This page builds on the concepts described in [Routing Traffic to and from NVIDIA Run:ai Services](../../system-requirements/networking-setup.md#routing-traffic-to-and-from-nvidia-runai-services), including FQDN configuration, TLS certificates, and routing behavior.

<Note>

**Note**

Gateway API support in NVIDIA Run:ai is **optional**. If you are currently using HAProxy Ingress or other ingress controllers, no action is required. Customers who wish to adopt Gateway API may do so using the instructions on this page.

</Note>

## Choose Your Installation Path

The networking concepts and Gateway API requirements on this page apply to both installation paths. Where configuration differs, follow the instructions for your installation method:

* **Guided Installation** - Configure NVIDIA Run:ai using `runai-installer` and `config.yaml`.
* **Manual Helm** - Configure NVIDIA Run:ai directly using Helm values and Kubernetes resources.

## Scope and Prerequisites

This guide assumes a self-hosted deployment where the NVIDIA Run:ai control plane and cluster services are installed on the same Kubernetes cluster.

Before proceeding, ensure that the following prerequisites are already configured:

* Fully Qualified Domain Names (FQDN) for:
  * Control plane access
  * Development workspaces, training, and distributed inference workloads
  * Standard inference (Knative) workloads
* TLS certificates associated with the configured FQDNs

These configurations are described in the [Routing Traffic to and from NVIDIA Run:ai Services](../../system-requirements/networking-setup.md#routing-traffic-to-and-from-nvidia-runai-services) section.

## Routing Modes in NVIDIA Run:ai

NVIDIA Run:ai supports two routing approaches for exposing services: **host-based routing** and **path-based routing**.

Different NVIDIA Run:ai services use these routing approaches as follows:

| Service                         | Routing Mode                     | Example                                          |
| ------------------------------- | -------------------------------- | ------------------------------------------------ |
| Control plane                   | Host-based (single domain)       | `https://runai.mycorp.local`                     |
| Standard inference (Knative) workloads | Host-based (wildcard subdomains) | `https://<service>.runai-inference.mycorp.local` |
| Workspaces, training, and distributed inference workloads | Host-based or path-based         | See section below                                |

## Installing a Gateway Controller

NVIDIA Run:ai supports any [conformant Gateway API implementation](https://gateway-api.sigs.k8s.io/implementations/). The example below uses **KGateway**. If you are using a different conformant controller, follow its installation documentation and then proceed to the migration steps.

1.  Install the Gateway API CRDs:

    ```bash
    kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.4.0/standard-install.yaml
    ```
2.  Install the KGateway CRDs:

    ```bash
    helm upgrade -i --create-namespace \
      --namespace kgateway-system \
      --version v2.2.1 \
      kgateway-crds oci://cr.kgateway.dev/kgateway-dev/charts/kgateway-crds
    ```
3.  Install the KGateway controller:

    ```bash
    helm upgrade -i -n kgateway-system kgateway \
      oci://cr.kgateway.dev/kgateway-dev/charts/kgateway \
      --version v2.2.1
    ```

Ensure that the Gateway controller is running before proceeding.

## Routing Traffic for Workspaces, Training, and Distributed Inference Workloads

The following sections describe how development workspaces, training, and distributed inference workloads are exposed using Gateway API.

Depending on the selected routing mode, these workloads are accessed differently:

*   In **host-based routing**, each development workspace, training, and distributed inference workload is exposed using its own subdomain:

    ```bash
    https://<project>-<workload>.<CLUSTER_URL>
    ```
*   In **path-based routing**, development workspaces, training, and distributed inference workloads are exposed under a shared domain using URL paths:

    ```bash
    https://runai.mycorp.local/<project>/<workload>
    ```

This distinction determines the required FQDN structure, TLS certificates, and Gateway configuration described in the sections below.

## Gateway API with Host-Based Routing

In this configuration:

* Development workspaces, training, and distributed inference workloads are exposed using subdomains
* A wildcard FQDN is required (e.g., `*.runai.mycorp.local`)
* A wildcard TLS certificate is required for those workloads
* The Gateway includes listeners for:
  * The cluster domain (control plane access)
  * Wildcard subdomains (workspaces, training, and distributed inference workloads)

### Gateway Configuration

<Note>

**Note**

Routing is based on hostnames (subdomains). HTTPRoute resources for the control plane and workloads are created automatically by NVIDIA Run:ai and bind to the configured Gateway.

</Note>

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: runai-gateway
  namespace: runai
spec:
  gatewayClassName: kgateway
  listeners:
  - name: https-cluster-domain
    protocol: HTTPS
    port: 443
    hostname: "<CLUSTER_DOMAIN>"
    tls:
      mode: Terminate
      certificateRefs:
      - kind: Secret
        name: runai-cluster-domain-tls-secret

  - name: https-workloads
    protocol: HTTPS
    port: 443
    hostname: "*.<CLUSTER_DOMAIN>"
    tls:
      mode: Terminate
      certificateRefs:
      - kind: Secret
        name: runai-cluster-domain-star-tls-secret
```

## Gateway API with Path-Based Routing

In this configuration:

* Development workspaces, training, and distributed inference workloads are exposed under a single domain using URL paths
* A wildcard FQDN is **not required** for workspaces, training, and distributed inference workloads
* A wildcard TLS certificate is **not required** for workspaces, training, and distributed inference workloads
* The Gateway uses a single domain listener for these services

Ensure that host-based routing is disabled. For more details, see [Advanced cluster configurations](../cluster-config.md).

<Tabs>

<Tab title="Guided Installation">

Set workload routing to `path` in your `config.yaml`:

```yaml
cluster:
  workloadRouting: path
```

Re-run `runai-installer install` to apply the change.

</Tab>

<Tab title="Manual Helm">

Disable host-based routing:

```yaml
clusterConfig.global.subdomainSupport: false
```

</Tab>

</Tabs>

### Gateway Configuration

<Note>

**Note**

Routing is based on URL paths under a shared domain. HTTPRoute resources for the control plane and workloads are created automatically by NVIDIA Run:ai and bind to the configured Gateway.

</Note>

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: runai-gateway
  namespace: runai
spec:
  gatewayClassName: kgateway
  listeners:
  - name: https-cluster-domain
    protocol: HTTPS
    port: 443
    hostname: "<CLUSTER_DOMAIN>"
    tls:
      mode: Terminate
      certificateRefs:
      - kind: Secret
        name: runai-cluster-domain-tls-secret
```

## Standard Inference Routing with Gateway API

<Note>

**Note**

This section applies to standard inference (Knative, single-node) only. Other inference workloads (for example, distributed inference using serving frameworks such as LWS, Dynamo, KServe, or others) do not use the inference subdomain and do not require the configuration described in this section.

</Note>

Standard inference (Knative) workloads are always exposed using host-based routing via a dedicated inference subdomain, regardless of the selected routing mode for development workspaces, training, and distributed inference workloads.

Standard inference endpoints are accessed via dedicated subdomains, for example:

```bash
https://<service>.runai-inference.mycorp.local
```

To enable standard inference routing with Gateway API:

* A wildcard FQDN must be configured for inference (e.g., `*.runai-inference.mycorp.local`)
* A wildcard TLS certificate must be configured for that domain
* The Gateway must include a listener for the inference subdomain
* Knative Serving must be configured to route inference traffic from the Gateway to inference workloads

### TLS Certificate

Ensure that a wildcard TLS certificate for standard inference is created:

* Secret name: `runai-cluster-inference-tls-secret`
* Namespace: `runai`

Replace `/path/to/inference-fullchain.pem` and `/path/to/inference-private.pem` with the actual paths to your certificate and private key:

```bash
kubectl create secret tls runai-cluster-inference-tls-secret -n runai \
    --cert /path/to/inference-fullchain.pem \
    --key /path/to/inference-private.pem
```

### Gateway Configuration

<Note>

**Note**

Routing to inference workloads is based on hostnames (subdomains).

</Note>

Add a standard inference listener to the existing Gateway resource:

```yaml
- name: https-inference
  protocol: HTTPS
  port: 443
  hostname: "*.runai-inference.<CLUSTER_DOMAIN>"
  tls:
    mode: Terminate
    certificateRefs:
    - kind: Secret
      name: runai-cluster-inference-tls-secret
```

### Configure Knative Serving

NVIDIA Run:ai uses Knative Serving to support standard inference workloads. Inference traffic arrives at the Gateway and is forwarded to Knative Serving through Kourier, which routes it to the individual inference endpoints.

<Note>

**Note**

On OpenShift, Knative is installed through the OpenShift Serverless operator, which supports Knative 1.17 only.

</Note>

<Tabs>

<Tab title="Guided Installation">

Configure standard inference in your `config.yaml`:

```yaml
cluster:
  inference:
    wildcardDomain: "*.<INFERENCE_WILDCARD_DOMAIN>"

features:
  knative:
    mode: install

externalConnection:
  type: gatewayApi
  gatewayApi:
    name: <GATEWAY_NAME>
    namespace: <GATEWAY_NAMESPACE>
```

Re-run `runai-installer install` to apply the change. The installer installs Knative Serving and attaches it to the Gateway with an HTTPRoute automatically - no manual YAML is needed, and no wildcard certificate is required for Knative itself under Gateway API. See [`features`](../../installation/guided-installation/configuration-reference.md#feature-configuration) and [Common Prerequisite and Feature Fields](../../installation/guided-installation/configuration-reference.md#common-prerequisite-and-feature-fields).

</Tab>

<Tab title="Manual Helm">

The following steps install Knative Serving and configure the HTTPRoute to connect the Gateway to Kourier.

<Tabs>

<Tab title="Kubernetes">

1.  Install Knative Serving. Follow the [Installing Knative](https://knative.dev/docs/install/operator/knative-with-operators/) instructions or run the following, updating `{VERSION}` with a `knative-operator` release that supports the Knative Serving version you are using:

    ```bash
    helm repo add knative-operator https://knative.github.io/operator
    helm install knative-operator --create-namespace --namespace knativeoperator --version {VERSION} knative-operator/knative-operator
    ```
2.  Create the `knative-serving` namespace:

    ```bash
    kubectl create ns knative-serving
    ```
3.  Create a YAML file named `knative-serving.yaml` and replace the placeholder FQDN with your wildcard [inference FQDN](../../system-requirements/networking-setup.md#fqdn-and-dns) (for example, `runai-inference.mycorp.local`):

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
        observability:
          metrics-protocol: "prometheus"
          request-metrics-protocol: "http/protobuf"
          request-metrics-endpoint: "http://prometheus-operated.runai.svc.cluster.local:9090/api/v1/otlp/v1/metrics"
        domain:
          runai-inference.mycorp.local: "" # replace with the wildcard FQDN for Inference
        network:
          domainTemplate: '{{.Name}}-{{.Namespace}}.{{.Domain}}'
          ingress-class: kourier.ingress.networking.knative.dev
          default-external-scheme: https
      high-availability:
        replicas: 2
      ingress:
        kourier:
          enabled: true
    ```
4.  Apply the changes:

    ```bash
    kubectl apply -f knative-serving.yaml
    ```
5.  Create a YAML file named `knative-httproute.yaml` to route inference traffic from the Gateway to the Kourier service. Replace the FQDN placeholder with your wildcard [inference FQDN](../../system-requirements/networking-setup.md#fqdn-and-dns):

    ```yaml
    apiVersion: gateway.networking.k8s.io/v1
    kind: HTTPRoute
    metadata:
      name: knative-serving-httproute
      namespace: knative-serving
    spec:
      parentRefs:
        - name: runai-gateway
          namespace: runai
      hostnames:
        - "*.runai-inference.mycorp.local" # replace with the wildcard FQDN for Inference
      rules:
        - matches:
          - path:
              type: PathPrefix
              value: /
          backendRefs:
            - name: kourier
              port: 80
    ```
6.  Apply the changes:

    ```bash
    kubectl apply -f knative-httproute.yaml
    ```

</Tab>

<Tab title="OpenShift">

1. Install the OpenShift Serverless Operator. Follow the [Installing the OpenShift Serverless Operator](https://docs.redhat.com/en/documentation/red_hat_openshift_serverless/1.37/html/installing_openshift_serverless/install-serverless-operator) instructions. Once installed, follow the steps below.
2.  Create the `knative-serving` project:

    ```bash
    oc new-project knative-serving
    ```
3.  Create a YAML file named `knative-serving.yaml`:

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
4.  Apply the changes:

    ```bash
    oc apply -f knative-serving.yaml
    ```
5.  Create a YAML file named `knative-httproute.yaml` to route inference traffic from the Gateway to the Kourier service. Replace the FQDN placeholder with your wildcard [inference FQDN](../../system-requirements/networking-setup.md#fqdn-and-dns):

    <Note><p><strong>Note</strong></p><p>OpenShift's Ingress Operator automatically creates DNS records for Gateway listeners, so no manual DNS configuration is required for the inference subdomain.</p></Note>

    ```yaml
    apiVersion: gateway.networking.k8s.io/v1
    kind: HTTPRoute
    metadata:
      name: knative-serving-httproute
      namespace: knative-serving
    spec:
      parentRefs:
        - name: openshift-default
          namespace: openshift-ingress
      hostnames:
        - "*.runai-inference.mycorp.local" # replace with the wildcard FQDN for Inference
      rules:
        - matches:
          - path:
              type: PathPrefix
              value: /
          backendRefs:
            - name: kourier
              port: 80
    ```
6.  Apply the changes:

    ```bash
    oc apply -f knative-httproute.yaml
    ```

</Tab>

</Tabs>

</Tab>

</Tabs>

## Introduce the Gateway API to NVIDIA Run:ai Services

To enable NVIDIA Run:ai to route traffic through the configured Gateway, both the control plane and the cluster must be updated to reference the Gateway.

At this stage, traffic is still served through the existing Ingress. The Gateway is introduced alongside the current setup and will become active only after DNS is updated in a later step.

<Tabs>

<Tab title="Guided Installation">

Set `externalConnection.type: gatewayApi` in your `config.yaml`, for both the control plane and cluster:

```yaml
externalConnection:
  type: gatewayApi
  gatewayApi:
    name: <GATEWAY_NAME>
    namespace: <GATEWAY_NAMESPACE>
```

Re-run `runai-installer install` with the updated `config.yaml` against the control plane cluster and each NVIDIA Run:ai cluster to apply the change. See [`externalConnection`](../../installation/guided-installation/configuration-reference.md#external-connection).

</Tab>

<Tab title="Manual Helm">

### Configure the NVIDIA Run:ai Control Plane

```bash
helm upgrade runai-backend -n runai-backend runai-backend/control-plane \
  --reuse-values \
  --set-string global.gatewayAPI.enabled=true \
  --set-string global.gatewayAPI.parentRefs[0].name=<gateway-name> \
  --set-string global.gatewayAPI.parentRefs[0].namespace=<gateway-namespace>
```

### Configure the NVIDIA Run:ai Cluster

```bash
helm upgrade runai-cluster -n runai runai/runai-cluster \
  --reuse-values \
  --set-string clusterConfig.global.gatewayAPI.enabled=true \
  --set-string clusterConfig.global.gatewayAPI.name=<gateway-name> \
  --set-string clusterConfig.global.gatewayAPI.namespace=<gateway-namespace>
```

</Tab>

</Tabs>

## Verify Gateway Configuration

```bash
kubectl get gateway -n runai
```

```bash
kubectl get httproute -A
```

Test connectivity:

```bash
curl -H "Host: <CLUSTER_DOMAIN>" https://<gateway-ip>/
```

## Switch Traffic to Gateway

Before switching traffic, ensure that the NVIDIA Run:ai control plane and cluster are already configured to use the Gateway API.

Update DNS:

* `<CLUSTER_DOMAIN>` -> Gateway IP
* `*.<CLUSTER_DOMAIN>` -> Gateway IP

Traffic is routed based on DNS configuration. Until DNS records are updated to point to the Gateway IP, all traffic continues to be served through the existing Ingress.

```bash
nslookup <CLUSTER_DOMAIN>
```

```bash
curl https://<CLUSTER_DOMAIN>/
```

Disable Ingress on both the cluster and the control plane. There is no dedicated `externalConnection` field for this - Ingress and Gateway API can be enabled together, so disabling Ingress is a separate step.

<Tabs>

<Tab title="Guided Installation">

Set `ingress.enabled: false` under `customHelmValues` for both the cluster and control plane in your `config.yaml`:

```yaml
cluster:
  customHelmValues:
    clusterConfig.global.ingress.enabled: false
controlPlane:
  customHelmValues:
    global.ingress.enabled: false
```

Re-run `runai-installer install` with the updated `config.yaml` to apply the change. See [`cluster`](../../installation/guided-installation/configuration-reference.md#cluster-configuration) / [`controlPlane`](../../installation/guided-installation/configuration-reference.md#control-plane-configuration).

</Tab>

<Tab title="Manual Helm">

Disable Ingress on the cluster:

```bash
helm upgrade runai-cluster -n runai runai/runai-cluster \
  --reuse-values \
  --set-string clusterConfig.global.ingress.enabled=false
```

Disable Ingress on the control plane:

```bash
helm upgrade runai-backend -n runai-backend runai-backend/control-plane \
  --reuse-values \
  --set-string global.ingress.enabled=false
```

</Tab>

</Tabs>

## Rollback to Ingress

To roll back to Ingress, re-enable Ingress and disable Gateway API on both the cluster and the control plane. Then update DNS records to point back to the Ingress IP.

<Tabs>

<Tab title="Guided Installation">

In your `config.yaml`, set `externalConnection.type` back to `ingress` (with `ingress.className`), and remove or set to `true` the `ingress.enabled` override under `customHelmValues` for both the cluster and control plane:

```yaml
externalConnection:
  type: ingress
  ingress:
    className: <ingress-class>
```

Re-run `runai-installer install` with the updated `config.yaml` to apply the change.

</Tab>

<Tab title="Manual Helm">

Cluster:

```bash
helm upgrade runai-cluster -n runai runai/runai-cluster \
  --reuse-values \
  --set-string clusterConfig.global.ingress.enabled=true \
  --set-string clusterConfig.global.gatewayAPI.enabled=false
```

Control plane:

```bash
helm upgrade runai-backend -n runai-backend runai-backend/control-plane \
  --reuse-values \
  --set-string global.ingress.enabled=true \
  --set-string global.gatewayAPI.enabled=false
```

</Tab>

</Tabs>

Update DNS records for `<CLUSTER_DOMAIN>` and `*.<CLUSTER_DOMAIN>` to point back to the Ingress IP.
