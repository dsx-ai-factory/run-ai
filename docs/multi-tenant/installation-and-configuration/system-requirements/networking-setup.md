---
---

# Networking Setup

NVIDIA Run:ai relies on your Kubernetes clusters' networking to route traffic between tenant users, the NVIDIA Run:ai control plane, and each tenant's NVIDIA Run:ai cluster. This section describes the networking components you must install and configure before installation:

* [Load Balancer](#load-balancer) - Exposes services outside the cluster through a single IP address.
* [Ingress Controller](#ingress-controller) - Routes external traffic to the NVIDIA Run:ai control plane and tenant cluster services.
* [FQDN and DNS](#fqdn-and-dns) - Domain names and DNS records used to route traffic by hostname.
* [TLS Certificates](#tls-certificates) - Secures HTTPS access to exposed services.

<Note>

**Note**

* A multi-tenant deployment is always multi-cluster: the control plane runs on its own dedicated Kubernetes cluster, and each tenant is connected to its own separate Kubernetes cluster. Every component below is installed once for the control plane cluster, and again for each tenant cluster. For more information, see [Deployment Model](deployment-model.md).
* FQDN and TLS certificates are not provisioned automatically by NVIDIA Run:ai. These must be prepared beforehand - for the control plane, before [installing the control plane](../installation/guided-installation/installation.md); for a tenant cluster, before [installing the cluster](../installation/guided-installation/create-a-tenant-and-cluster.md#install-the-cluster).

</Note>

## Load Balancer

<Note>

**Deployment model**

Install the load balancer once for the control plane cluster, and once for each tenant cluster.

</Note>

In Kubernetes, services of type `LoadBalancer` are used to expose applications outside the cluster through a single, stable IP address, providing a consistent entry point for external traffic. In managed cloud environments this capability is built-in, while in self-hosted and on-premise environments it must be provided explicitly.

MetalLB fulfills this role by allocating external IP addresses from a predefined pool and advertising them on the external network, enabling access to services running inside the cluster.

In NVIDIA Run:ai, this is required to support north-south traffic, including access to the NVIDIA Run:ai control plane, APIs, and UI, as well as externally exposed workloads - standard inference (Knative) endpoints, development workspaces, training, and distributed inference workloads running on a tenant cluster.

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

## Ingress Controller

<Note>

**Deployment model**

Install the ingress controller once for the control plane cluster, and once for each tenant cluster.

</Note>

Both the NVIDIA Run:ai control plane and each tenant cluster require a [Kubernetes Ingress Controller](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/) to be installed on their Kubernetes cluster.

* RKE2 comes with a pre-installed ingress controller.
* OpenShift comes with a pre-installed ingress controller.

There are multiple ways to install and configure an ingress controller. The following example demonstrates how to install and configure the [HAProxy Kubernetes Ingress Controller](https://www.haproxy.com/documentation/kubernetes-ingress/) using [helm](https://helm.sh/).

<Accordion title="Vanilla Kubernetes">

**Vanilla Kubernetes**

```bash
helm repo add haproxytech https://haproxytech.github.io/helm-charts
helm repo update
helm install haproxy-kubernetes-ingress haproxytech/kubernetes-ingress \
 --create-namespace \
 --namespace haproxy-controller \
 --set controller.kind=DaemonSet \
 --set controller.service.type=NodePort
```

<Note>

**Note**

For ingresses with TLS, HAProxy redirects HTTP requests to the port set in `controller.config.ssl-redirect-port`. The default is 8443, the controller's container port, which clients cannot reach through a NodePort Service. Set it to the HTTPS port that clients connect to. Use the HTTPS NodePort when clients reach the nodes directly, or 443 when an external load balancer forwards port 443 to the HTTPS NodePort. For example, add `--set-string controller.config.ssl-redirect-port=443` to the command above. For more information, refer to [ssl-redirect-port](https://www.haproxy.com/documentation/kubernetes-ingress/community/configuration-reference/configmap/) in the HAProxy Kubernetes Ingress Controller documentation.

</Note>

</Accordion>

<Accordion title="Managed Kubernetes (EKS, GKE, AKS)">

**Managed Kubernetes (EKS, GKE, AKS)**

```bash
helm repo add haproxytech https://haproxytech.github.io/helm-charts
helm repo update
helm install haproxy-kubernetes-ingress haproxytech/kubernetes-ingress \
 --create-namespace \
 --namespace haproxy-controller \
 --set controller.service.type=LoadBalancer
```

</Accordion>

<Accordion title="Oracle Kubernetes Engine (OKE)">

**Oracle Kubernetes Engine (OKE)**

```bash
helm repo add haproxytech https://haproxytech.github.io/helm-charts
helm repo update
helm install haproxy-kubernetes-ingress haproxytech/kubernetes-ingress \
 --create-namespace \
 --namespace haproxy-controller \
 --set controller.kind=DaemonSet \
 --set controller.service.type=LoadBalancer \
 --set controller.service.externalTrafficPolicy=Local \
 --set controller.service.annotations."oci-network-load-balancer\.oraclecloud\.com/is-preserve-source"="True" \
 --set controller.service.annotations."oci-network-load-balancer\.oraclecloud\.com/security-list-management-mode"=All \
 --set controller.service.annotations."oci\.oraclecloud\.com/load-balancer-type"=nlb
```

</Accordion>

## Routing Traffic to and from NVIDIA Run:ai Services

This section describes how to route traffic to and from NVIDIA Run:ai services. Proper traffic routing is required to enable secure access to the NVIDIA Run:ai control plane (via the UI or API) for every tenant, as well as external access to development workspaces, training workloads, and inference workloads running on tenant clusters. External access means the workload is reachable from outside the Kubernetes cluster, for example by a user's browser or client. See [FQDN and DNS](#fqdn-and-dns).

<Note>

**Note**

When exposed externally, inference workloads always use host-based routing via their own wildcard subdomain, regardless of whether workspaces and training workloads use host-based or path-based routing.

</Note>

NVIDIA Run:ai supports two routing approaches for exposing services: host-based routing and path-based routing. While path-based routing exposes multiple services under a shared domain using URL paths, host-based routing assigns each service its own subdomain. Since many development tools and applications expect to run at the root path, host-based routing avoids common compatibility issues and is therefore used by default in NVIDIA Run:ai.

NVIDIA Run:ai uses host-based routing by default, which relies on DNS and TLS configuration to securely expose services. To support this, three key components must be configured:

* A Fully Qualified Domain Name (FQDN) structure
* TLS certificates for secure communication
* Ingress rules connecting the FQDN and TLS certificates to route traffic by hostname - [Guided Installation](../installation/guided-installation/README.md) creates these automatically as part of installation, while [Manual Helm](../installation/install-using-helm/preparations.md) requires you to create the Ingress resource yourself; see the installation guide for your chosen path for details

Together, these components ensure that traffic is routed correctly and securely across all NVIDIA Run:ai services: host-based routing binds together the configured domains (FQDN), TLS certificates, and ingress rules to expose workloads externally, with each workload assigned its own subdomain and URL. For example:

```bash
https://<project>-<workload>.<CLUSTER_URL>
```

<Note>

**Note**

NVIDIA Run:ai also supports path-based routing, which you can use instead of the default host-based routing if it better fits your environment - [Guided Installation](../installation/guided-installation/README.md) accepts it as a configuration field, while [Manual Helm](../installation/install-using-helm/preparations.md) sets it via a Helm value; see the installation guide for your chosen path for the exact syntax.

</Note>

## FQDN and DNS

<Note>

**Note**

* The control plane's wildcard domain is configured once, at control plane install time.
* Each tenant cluster's domains are configured separately, per tenant, during onboarding.
* Applies to Kubernetes only - on OpenShift, a tenant cluster's domain is derived automatically from the cluster's route domain.

</Note>

NVIDIA Run:ai multi-tenant relies on Fully Qualified Domain Names (FQDNs) at two independent layers: a single wildcard domain for the control plane, shared across every tenant, and a separate domain per tenant cluster. FQDNs are needed for:

* Accessing the NVIDIA Run:ai control plane UI and API, for every tenant
* Enabling communication between the control plane and each tenant cluster
* Exposing each tenant's development workspaces, training, and distributed inference workloads via subdomains
* Exposing each tenant's standard inference workloads (Knative) via dedicated subdomains

### Control Plane Domain (Wildcard)

Tenants are created after the control plane is already installed and running - a host organization sends a `POST` request to the [`/api/v1/tenants`](../installation/guided-installation/create-a-tenant-and-cluster.md#create-a-tenant) endpoint at any time to onboard a new tenant. At that point, NVIDIA Run:ai automatically assigns the new tenant a subdomain of the control plane's domain - there is no manual DNS or certificate step per tenant.

Because tenant subdomains are minted on demand, the control plane domain must be backed by a **wildcard** DNS record, not a single fixed domain, so that every tenant subdomain that will ever be created already resolves. You supply this domain as part of control plane installation - [Guided Installation](../installation/guided-installation/README.md) accepts it as the `controlPlane.domain` configuration field, while [Manual Helm](../installation/install-using-helm/install-control-plane.md) applies it via the `global.domain` Helm value.

| Domain | Purpose | Example |
|---|---|---|
| Control plane domain (wildcard) | Access to the NVIDIA Run:ai control plane UI and API, for every tenant. | `*.runai.hostorg.com`, resolving requests such as `tenant-a.runai.hostorg.com` and `tenant-b.runai.hostorg.com` |

Configure the wildcard DNS record to resolve to the control plane cluster's load balancer IP address before [installing the control plane](../installation/guided-installation/installation.md).

```bash
*.runai.hostorg.com → <control plane cluster IP>
```

### Tenant Cluster Domain

Each tenant cluster has its own domain, independent of that tenant's control plane subdomain above. You supply these domains as part of the cluster installation - [Guided Installation](../installation/guided-installation/README.md) accepts the cluster domain as the `cluster.domain` configuration field, while [Manual Helm](../installation/install-using-helm/create-a-tenant-and-cluster.md) applies it via the Helm command returned by the platform when the cluster is registered.

| Domain | Purpose | Example |
|---|---|---|
| Tenant cluster domain | Control plane ↔ tenant cluster communication. | `runai.tenant-a-corp.com` |
| Workspaces, training, and distributed inference workloads | External access | `*.runai.tenant-a-corp.com` |
| Standard inference workloads (Knative) | External access | `*.runai-inference.tenant-a-corp.com` |

<Note>

**Note**

The inference wildcard (`*.runai-inference.tenant-a-corp.com`) is only required for standard inference (Knative, single-node) workloads exposed outside the cluster. Other inference workloads (for example, distributed inference using serving frameworks such as LWS, Dynamo, KServe, or others) do not use this domain. For cluster-local standard inference, skip the inference wildcard record and the matching inference wildcard certificate.

</Note>

All of a tenant cluster's domains resolve to that cluster's own public IP address, or to the cluster's load balancer IP in on-premises environments. This ensures that each workspace or workload is assigned a unique subdomain under the wildcard domains:

```bash
runai.tenant-a-corp.com             → <tenant cluster IP>
*.runai.tenant-a-corp.com           → <tenant cluster IP>
*.runai-inference.tenant-a-corp.com → <tenant cluster IP>
```

## TLS Certificates

As with FQDNs, TLS certificates are required at both the control plane layer and the tenant cluster layer. Obtain the certificate and private key files for each domain before installation. How they are applied depends on the installation path - [Guided Installation](../installation/guided-installation/README.md) accepts file paths in the configuration file; [Manual Helm](../installation/install-using-helm/preparations.md) requires creating Kubernetes secrets manually.

NVIDIA Run:ai multi-tenant requires up to four TLS certificates, each aligned with a specific domain:

* Control plane certificate (wildcard)
* Tenant cluster domain certificate
* Tenant workspaces, training, and distributed inference wildcard certificate
* Tenant standard inference wildcard certificate

### Control Plane Certificate (Wildcard)

You must provide a TLS certificate that matches your wildcard DNS domain (e.g., `*.runai.hostorg.com`). This certificate is used to secure HTTPS access to tenant-facing endpoints, ensuring that each tenant receives a secure URL when accessing NVIDIA Run:ai services.

### Tenant Cluster Domain Certificate

Each tenant cluster is its own separate Kubernetes cluster with its own domain, so each tenant requires its own certificate - unlike the control plane's single shared wildcard certificate above. To enable secure communication between the NVIDIA Run:ai control plane and a tenant cluster, configure a TLS certificate associated with that cluster's main domain (e.g. `runai.tenant-a-corp.com`). On OpenShift, NVIDIA Run:ai uses the OpenShift default Ingress router for serving; the TLS certificate configured for this router must be issued by a trusted CA.

### Tenant Workspaces, Training, and Distributed Inference Wildcard Certificate

<Note>

**Note**

For path-based routing, this certificate is not required.

</Note>

To allow secure access to a tenant's workspace, training, and distributed inference workloads via subdomains, configure a wildcard TLS certificate that matches that tenant cluster's domain (e.g. `*.runai.tenant-a-corp.com`). On OpenShift, a wildcard TLS certificate for these workloads is not required. OpenShift Routes handle TLS termination for workspace, training, and distributed inference endpoints using the platform's built-in routing and certificate management.

### Tenant Standard Inference Wildcard Certificate

<Note>

**Note**

The inference wildcard certificate is only required for standard inference (Knative, single-node) workloads exposed externally via wildcard FQDN. Other inference workloads (for example, distributed inference using serving frameworks such as LWS, Dynamo, KServe, or others) do not require this certificate. For cluster-local standard inference, skip this section.

</Note>

To securely expose a tenant's inference services over HTTPS, configure a wildcard TLS certificate for that tenant cluster's inference domain (e.g. `*.runai-inference.tenant-a-corp.com`). On OpenShift, a wildcard TLS certificate for inference workloads is not required. OpenShift Routes handle TLS termination for inference endpoints using the platform's built-in routing and certificate management.
