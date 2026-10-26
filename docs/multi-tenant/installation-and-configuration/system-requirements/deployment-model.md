---
---

# Deployment Model

The NVIDIA Run:ai control plane and every tenant's NVIDIA Run:ai cluster are always deployed on separate Kubernetes clusters: the control plane runs on its own dedicated Kubernetes cluster, managed by the host organization, and each tenant's NVIDIA Run:ai cluster runs on its own separate Kubernetes cluster. Each tenant cluster requires its own load balancer, ingress controller, FQDNs, and TLS certificates.

## Connectivity

NVIDIA Run:ai multi-tenant involves two distinct traffic flows, each with its own connectivity requirements:

* **Management traffic (outbound)** - Each tenant's NVIDIA Run:ai cluster services initiate the connection to the control plane FQDN (for example, `tenant-a.runai.hostorg.com`) over HTTPS (port 443). In restricted networks, allow outbound HTTPS from each tenant cluster's nodes to the control plane FQDN.
* **User access traffic (inbound)** - Users and clients require inbound access to:
  * The NVIDIA Run:ai control plane, for UI and API access, per tenant.
  * Each tenant's NVIDIA Run:ai cluster, for access to workspaces, training, and inference workloads exposed through that cluster's load balancer and ingress controller.

Each tenant cluster must be reachable by its own users - this is why the load balancer, ingress controller, and cluster-specific FQDNs and TLS certificates are required per tenant cluster. The control plane's wildcard domain and certificate are provisioned once, covering every tenant.

The following diagram shows the traffic flows across the control plane and tenant clusters:

<img src="../../../assets/traffic-multi-cluster.svg" alt="Multi-tenant traffic flow: control plane and tenant cluster workloads on separate Kubernetes clusters, connected via management traffic" />

## Infrastructure by Component

Use the tables below to determine how many of each networking, DNS, and TLS component to provision before you begin installation.

* **Single** - One component, provisioned once for the control plane, covering every tenant.
* **Per tenant cluster** - One component required for each tenant cluster.

For setup instructions, see [Networking Setup](networking-setup.md).

### Load Balancer and Ingress

#### Kubernetes

| Component | Control Plane | Tenant Cluster |
|---|---|---|
| Load balancer | Single | Per tenant cluster |
| Ingress controller | Single | Per tenant cluster |

#### OpenShift

OpenShift ships with a pre-installed router that serves as both load balancer and ingress controller, so no separate installation is required. This applies to tenant clusters running on OpenShift.

| Component | Tenant Cluster |
|---|---|
| Router (built-in) | Per tenant cluster |

### DNS

| Component | Control Plane | Tenant Cluster |
|---|---|---|
| Control plane domain (wildcard) | Single | - |
| Tenant cluster domain | - | Per tenant cluster |
| Workspace and training wildcard domain | - | Per tenant cluster |
| Inference wildcard domain<sup>1</sup> | - | Per tenant cluster |

<sup>1</sup> Required only if inference workloads are exposed externally.

### TLS Certificates

#### Kubernetes

| Component | Control Plane | Tenant Cluster |
|---|---|---|
| Control plane certificate (wildcard) | Single | - |
| Cluster domain certificate | - | Per tenant cluster |
| Workspace and training wildcard certificate | - | Per tenant cluster |
| Inference wildcard certificate<sup>1</sup> | - | Per tenant cluster |

<sup>1</sup> Required only if inference workloads are exposed externally.

#### OpenShift

This applies to tenant clusters running on OpenShift.

| Component | Tenant Cluster |
|---|---|
| Router certificate | Per tenant cluster |
