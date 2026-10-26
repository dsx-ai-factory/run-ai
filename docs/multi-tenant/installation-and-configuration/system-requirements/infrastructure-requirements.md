---
---

# Infrastructure Requirements

This page covers the hardware, operating system, Kubernetes, and storage requirements that must be in place before installation begins - both for the NVIDIA Run:ai control plane cluster, and for each tenant's Kubernetes cluster. Where a requirement differs between the control plane and a tenant cluster, it is called out separately.

<Note>

**Note**

A multi-tenant deployment is always multi-cluster: the control plane runs on its own dedicated Kubernetes cluster, and each tenant is connected to its own separate Kubernetes cluster. For more information, see [Deployment Model](deployment-model.md).

</Note>

| Category | Requirement | Applies To | Details |
|---|---|---|---|
| Hardware | Control plane nodes, tenant cluster nodes, and worker nodes provisioned | Control plane, tenant clusters, and worker nodes (sized separately) | See [Hardware](#hardware) |
| Operating System | Linux installed on all nodes | Control plane and tenant clusters (requirements differ) | See [Operating System](#operating-system) |
| Time Sync | Nodes synchronized using NTP | All nodes | See [Network Time Protocol](#network-time-protocol) |
| Distribution | Supported Kubernetes distribution installed and running | Control plane and tenant clusters (distributions differ) | See [Supported Distributions](#supported-distributions) |
| Runtime | Container runtime configured | All nodes | See [Container Runtime](#container-runtime) |
| Permissions | Cluster-admin access | All nodes | See [Permissions](#permissions) |
| Storage | Default storage class configured | Control plane | See [Default Storage Class](#default-storage-class) |

## Hardware

The following hardware requirements apply to all nodes in the deployment. By default, all NVIDIA Run:ai services run on all available nodes. For production deployments, set [node roles](../../../installation-and-configuration/configuration/node-roles.md) to separate system nodes from worker nodes, reduce downtime, and avoid consuming CPU cycles on expensive GPU machines.

### Supported Architectures

NVIDIA Run:ai supports the following CPU architectures:

* **x86** - Kubernetes and OpenShift
* **ARM** - Kubernetes and OpenShift

### NVIDIA Run:ai Control Plane - System Nodes

The following are the minimum requirements for NVIDIA Run:ai control plane system nodes. By default, all control plane services run on all available nodes in the control plane cluster. To designate nodes to NVIDIA Run:ai control plane services, see [System nodes](../../../installation-and-configuration/configuration/node-roles.md#system-nodes).

| Component  | Required Capacity |
|------------|-------------------|
| CPU        | 10 cores          |
| Memory     | 12 GB             |
| Disk space | 110 GB            |

### NVIDIA Run:ai Cluster - System Nodes

The following are the minimum requirements for a tenant cluster's system nodes. By default, all cluster services run on all available nodes in the tenant cluster. To designate nodes to NVIDIA Run:ai cluster services, see [System nodes](../../../installation-and-configuration/configuration/node-roles.md#system-nodes).

| Component  | Required Capacity |
|------------|-------------------|
| CPU        | 10 cores          |
| Memory     | 20 GB             |
| Disk space | 50 GB             |

### NVIDIA Run:ai Cluster - Worker Nodes

Each tenant cluster supports x86 and ARM CPUs, and any NVIDIA GPUs supported by the NVIDIA GPU Operator. The list of supported GPUs depends on the version of the NVIDIA GPU Operator installed in the tenant cluster.

<Note>

**Note**

* NVIDIA DGX Spark, NVIDIA Jetson, and workstations are not supported.
* vGPU is not supported. NVIDIA Run:ai currently supports GPU passthrough only.

</Note>

The following are the minimum hardware requirements for each worker node. To designate nodes to NVIDIA Run:ai workloads, see [Worker nodes](../../../installation-and-configuration/configuration/node-roles.md#worker-nodes).

| Component | Required Capacity |
|-----------|-------------------|
| CPU       | 2 cores           |
| Memory    | 4 GB              |

## Operating System

### NVIDIA Run:ai Control Plane

* Any **Linux** operating system supported by both Kubernetes and NVIDIA GPU Operator.
* Internal tests are performed on **Ubuntu 22.04**.

### NVIDIA Run:ai Cluster

* Any **Linux** operating system supported by both Kubernetes and NVIDIA GPU Operator.
* NVIDIA Run:ai on Google Kubernetes Engine (GKE) supports both Ubuntu and Container Optimized OS (COS).
  * COS is supported only with NVIDIA GPU Operator 24.6 and above, and NVIDIA Run:ai cluster version 2.19 and above.
* NVIDIA Run:ai on Elastic Kubernetes Service (EKS) does not support Bottlerocket or Amazon Linux.
* NVIDIA Run:ai on Oracle Kubernetes Engine (OKE) supports only Ubuntu.
* Internal tests are performed on **Ubuntu** and **CoreOS** for OpenShift.

## Network Time Protocol

All nodes must be synchronized using NTP (Network Time Protocol) for proper system functionality.

## Kubernetes

### Supported Distributions

The NVIDIA Run:ai control plane supports the following Kubernetes distributions:

* Vanilla Kubernetes
* Elastic Kubernetes Service (EKS)
* Google Kubernetes Engine (GKE)
* Azure Kubernetes Service (AKS)
* Oracle Kubernetes Engine (OKE)
* Rancher Kubernetes Engine 2 (RKE2)

Each tenant cluster supports the following Kubernetes distributions:

* Vanilla Kubernetes
* OpenShift Container Platform (OCP)
* Rancher Kubernetes Engine 2 (RKE2)

For the full list of compatible distributions and per-version compatibility, see the [Support Matrix](../../../installation-and-configuration/support-matrix.md#nvidia-runai-compatible-distributions).

### Container Runtime

NVIDIA Run:ai supports the following [container runtimes](https://kubernetes.io/docs/setup/production-environment/container-runtimes/):

* [Containerd](https://kubernetes.io/docs/setup/production-environment/container-runtimes/#containerd) (default in Kubernetes)
* [CRI-O](https://cri-o.io/) (default in OpenShift)

### Permissions

Installing the NVIDIA Run:ai control plane or a tenant cluster requires **cluster-admin** permissions on the target Kubernetes cluster. Ensure your `kubectl` context is configured with a user that has cluster-admin or equivalent rights before proceeding.

## Storage

A default storage class is required for the NVIDIA Run:ai control plane before installation.

### Default Storage Class

<Note>

**Note**

Applies to Kubernetes only.

</Note>

The NVIDIA Run:ai control plane requires a **default storage class** to create persistent volume claims for NVIDIA Run:ai storage. The storage class, as per Kubernetes standards, controls the reclaim behavior - whether NVIDIA Run:ai persistent data is saved or deleted when the NVIDIA Run:ai control plane is deleted.

<Note>

**Note**

For a simple (non-production) storage class example, see [Kubernetes Local Storage Class](https://kubernetes.io/docs/concepts/storage/storage-classes/#local). The storage class will set the directory `/opt/local-path-provisioner` to be used across all nodes as the path for provisioning persistent volumes. Then set the new storage class as default:

```bash
kubectl patch storageclass local-path -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
```

</Note>
