---
---

# Networking Setup

NVIDIA Run:ai relies on your Kubernetes cluster's networking to route traffic between users, the control plane, and NVIDIA Run:ai clusters. This section describes the networking components you must install and configure before installation:

* [Load Balancer](#load-balancer) - Exposes services outside the cluster through a single IP address.
* [Ingress Controller](#ingress-controller) - Routes external traffic to the NVIDIA Run:ai control plane and cluster services.
* [FQDN and DNS](#fqdn-and-dns) - Domain names and DNS records used to route traffic by hostname.
* [TLS Certificates](#tls-certificates) - Secures HTTPS access to exposed services.
* [Local Certificate Authority (CA)](#local-certificate-authority-ca) - Required if your certificates are signed by a private CA.

As per the [deployment model](deployment-model.md), each component is installed once or once per NVIDIA Run:ai cluster.

## Load Balancer

<Info>

**Deployment model**

Install the load balancer once for a same-cluster deployment, or on each cluster for a multi-cluster deployment. For more information, see [Deployment Model](deployment-model.md).

</Info>

In Kubernetes, services of type `LoadBalancer` are used to expose applications outside the cluster through a single, stable IP address, providing a consistent entry point for external traffic. In managed cloud environments this capability is built-in, while in self-hosted and on-premise deployments it must be provided explicitly.

MetalLB fulfills this role by allocating external IP addresses from a predefined pool and advertising them on the external network, enabling access to services running inside the cluster.

In NVIDIA Run:ai, this is required to support north-south traffic, including access to the NVIDIA Run:ai control plane, APIs, UI, inference endpoints, and externally exposed development workspaces and training workloads.

1. Reserve a range of IP addresses (recommended a full 32 subnet) for example: `172.20.10.0-172.20.10.255`
2. Install MetalLB:

    ```bash
    helm repo add metallb https://metallb.github.io/metallb
    helm repo update
    helm install metallb metallb/metallb --version 0.15.3 --namespace metallb-system --create-namespace
    ```

3. Create a YAML file named `metalLB-config.yaml` and replace `<IPADDRESS-RANGE-START>-<IPADDRESS-RANGE-END>` with the reserved range of IP addresses:

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

4. Apply the YAML:

    ```bash
    kubectl apply -f metalLB-config.yaml
    ```

## Ingress Controller

<Info>

**Deployment model**

Install the ingress controller once for a same-cluster deployment, or on each cluster for a multi-cluster deployment. For more information, see [Deployment Model](deployment-model.md).

</Info>

NVIDIA Run:ai requires a [Kubernetes Ingress Controller](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/) to be installed on the Kubernetes cluster. OpenShift and RKE2 come with a pre-installed ingress controller. Once installed, you'll reference this controller when configuring NVIDIA Run:ai - the guided installer accepts it as a configuration field, while the Helm path sets it via a Helm value. See the installation guide for your chosen path for the exact syntax.

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

## How Routing Works

This section describes how to route traffic to and from NVIDIA Run:ai services. Proper traffic routing is required to enable secure access to the NVIDIA Run:ai control plane (via the UI or API), as well as external access to development workspaces, training workloads, and inference workloads. External access means the workload is reachable from outside the Kubernetes cluster, for example by a user's browser or client. Inference workloads can also be kept cluster-local, reachable only from other services inside the cluster, instead of exposed externally; see [FQDN and DNS](#fqdn-and-dns).

NVIDIA Run:ai supports two routing approaches for exposing services: host-based routing and path-based routing. While path-based routing exposes multiple services under a shared domain using URL paths, host-based routing assigns each service its own subdomain. Since many development tools and applications expect to run at the root path, host-based routing avoids common compatibility issues and is therefore used by default in NVIDIA Run:ai.

NVIDIA Run:ai uses host-based routing by default, which relies on DNS and TLS configuration to securely expose services. To support this, three key components must be configured:

* A Fully Qualified Domain Name (FQDN) structure
* TLS certificates for secure communication
* Ingress rules connecting the FQDN and TLS certificates to route traffic by hostname - how these are created depends on your installation path; see the installation guide for your chosen path for details

Together, these components ensure that traffic is routed correctly and securely across all NVIDIA Run:ai services.

<Info>

**Note**

NVIDIA Run:ai also supports path-based routing, which you can use instead of the default host-based routing if it better fits your environment - the guided installer accepts it as a configuration field, while the Helm path sets it via a Helm value; see the installation guide for your chosen path for the exact syntax. In this case, the workspace and training wildcard certificate is not required.

</Info>

Host-based routing binds together the configured domains (FQDN), TLS certificates, and ingress rules to expose workloads externally. With NVIDIA Run:ai host-based routing, workloads are exposed using subdomains, so each workload is assigned its own URL. For example:

```bash
https://<project>-<workload>.<CLUSTER_URL>
```

## FQDN and DNS

<Info>

**Deployment model**

Configure wildcard FQDNs once for a same-cluster deployment, or separately for each NVIDIA Run:ai cluster in a multi-cluster deployment. Applies to Kubernetes only - on OpenShift, the domain is derived automatically from the cluster's route domain. For more information, see [Deployment Model](deployment-model.md).

</Info>

NVIDIA Run:ai services rely on Fully Qualified Domain Names (FQDNs) to route traffic between system components and to expose workloads externally. FQDNs are needed for:

* Enabling communication between the NVIDIA Run:ai control plane and the NVIDIA Run:ai cluster
* Exposing development workspaces and training workloads via subdomains
* Exposing inference workloads via dedicated subdomains

Configure domain names for each of the following:

* NVIDIA Run:ai control plane ↔ NVIDIA Run:ai cluster communication. Example: `runai.mycorp.local`. The IP address of this domain must be resolvable within the organization's private network.
* Workspace and training workloads (external access). Example: `*.runai.mycorp.local`, or `*.runai-a.mycorp.local` for cluster A in a multi-cluster deployment.
* Inference workloads (external access). Example: `*.runai-inference.mycorp.local`, or `*.runai-a-inference.mycorp.local` for cluster A in a multi-cluster deployment.

<Info>

**Note**

The inference wildcard (`*.runai-inference.mycorp.local`) is only required if you plan to expose inference workloads outside the cluster. For cluster-local inference, skip the inference wildcard record and the matching inference wildcard certificate.

</Info>

Configure the following DNS records. Both records should resolve to the same cluster public IP address, or to the cluster's load balancer IP in on-premises environments. This ensures that each workspace or workload is assigned a unique subdomain under the wildcard domains:

```bash
*.runai.mycorp.local           → <cluster IP>
*.runai-inference.mycorp.local → <cluster IP>
```

The following diagram shows how each request hostname maps to a DNS record, TLS certificate, and destination:

<img src="../../assets/dns-tls-mapping.svg" alt="DNS and TLS mapping: how request hostnames resolve to DNS records, TLS certificates, and destinations" />

Create these DNS records in your own DNS provider before installation. You then supply the domain values to NVIDIA Run:ai as part of installation - the guided installer accepts them as fields in the configuration file, while the Helm path applies them via Helm values.

## TLS Certificates

<Info>

**Deployment model**

The number of certificates required depends on your deployment model and Kubernetes distribution. For a same-cluster Kubernetes deployment, three certificates are required (plus an inference wildcard if exposing inference workloads externally). For multi-cluster, the control plane requires one certificate, and each NVIDIA Run:ai cluster requires two (plus an inference wildcard if exposing inference workloads externally). For OpenShift, one router certificate per cluster - no other certificates are required. For more information, see [Deployment Model](deployment-model.md).

</Info>

TLS certificates secure communication between NVIDIA Run:ai components and enable HTTPS access to exposed services. Depending on your deployment model, some or all of the following are required:

* **Control plane certificate** - Secures HTTPS access to the NVIDIA Run:ai UI and API.
* **Cluster domain certificate** - Secures communication between the NVIDIA Run:ai control plane and the NVIDIA Run:ai cluster.
* **Workspace and training wildcard certificate** - Secures external access to workspace and training workload subdomains.
* **Inference wildcard certificate** - Inference workloads can be exposed externally or kept cluster-local. This certificate secures external access to inference workload subdomains and is required only when exposing inference workloads externally.

Obtain the certificate and private key files for each domain before installation. How they are applied depends on the installation path - the guided installer accepts file paths in the configuration file; the Helm path requires creating Kubernetes secrets manually. If certificates from a trusted CA are not available, the guided installer can generate self-signed certificates instead.

## Local Certificate Authority (CA)

<Info>

**Note**

Required for organizations that cannot use a publicly trusted certificate authority, and mandatory for air-gapped installations.

</Info>

A local certificate authority (CA) serves as the root certificate for organizations that cannot use a publicly trusted certificate authority. If your TLS certificates are signed by a private or local CA, you must provide the CA's certificate chain as part of installation. In air-gapped environments, this is required for the installation to succeed.

How you provide the CA chain depends on the installation path - the guided installer accepts a `tls.caChainPath` field in the configuration file; the Helm path requires creating a Kubernetes secret manually. See the installation guide for your chosen path for the exact procedure.
