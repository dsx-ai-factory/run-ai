---
---

# Infrastructure Requirements

The following requirements cover what must be configured in your Kubernetes cluster before installation begins. Requirements are listed separately for the NVIDIA Run:ai control plane and the NVIDIA Run:ai cluster(s) - see [Deployment Model](deployment-model.md) for how each is provisioned. Ensure each item is in place before proceeding with either installation path.

| Category | Requirement | Details |
|---|---|---|
| Hardware | Control plane nodes, cluster nodes, and worker nodes provisioned | See [Hardware](#hardware) |
| Operating System | Linux installed on all nodes | See [Operating System](#operating-system) |
| Time Sync | Nodes synchronized using NTP | See [Network Time Protocol](#network-time-protocol) |
| Distribution | Supported Kubernetes distribution installed and running | See [Supported Distributions](#supported-distributions) |
| Runtime | Container runtime configured | See [Container Runtime](#container-runtime) |
| Permissions | Cluster-admin access | See [Permissions](#permissions) |
| Storage | Default storage class configured | See [Default Storage Class](#default-storage-class) |
| Storage | Shared storage accessible from all worker nodes | See [Shared Storage](#shared-storage) |

## Hardware

The following hardware requirements apply to all nodes in the deployment. By default, all NVIDIA Run:ai services run on all available nodes. For production deployments, set [node roles](../../configuration/node-roles.md) to separate system nodes from worker nodes, reduce downtime, and avoid consuming CPU cycles on expensive GPU machines. If deploying the NVIDIA Run:ai control plane and NVIDIA Run:ai cluster on the same Kubernetes cluster, the system node requirements are additive - provision nodes that meet both sets of requirements combined.

### Supported Architectures

NVIDIA Run:ai supports the following CPU architectures:

* **x86** - Kubernetes and OpenShift
* **ARM** - Kubernetes and OpenShift

### Control Plane - System Nodes

The following are the minimum requirements for NVIDIA Run:ai control plane system nodes. By default, all control plane services run on all available nodes in the control plane cluster. To designate nodes to NVIDIA Run:ai control plane services, see [System nodes](../../configuration/node-roles.md#system-nodes).

| Component  | Required Capacity |
|------------|-------------------|
| CPU        | 10 cores          |
| Memory     | 12 GB             |
| Disk space | 110 GB            |

### Cluster - System Nodes

The following are the minimum requirements for NVIDIA Run:ai cluster system nodes. By default, all cluster services run on all available nodes in the cluster. To designate nodes to NVIDIA Run:ai cluster services, see [System nodes](../../configuration/node-roles.md#system-nodes).

| Component  | Required Capacity |
|------------|-------------------|
| CPU        | 10 cores          |
| Memory     | 20 GB             |
| Disk space | 50 GB             |

### Cluster - Worker Nodes

The NVIDIA Run:ai cluster supports x86 and ARM CPUs, and any NVIDIA GPUs supported by the NVIDIA GPU Operator. The list of supported GPUs depends on the version of the NVIDIA GPU Operator installed in the cluster. For supported GPU Operator versions, see the [Support Matrix](../../support-matrix.md#operator-and-framework-versions). For the list of supported GPUs, see [Supported NVIDIA Data Center GPUs and Systems](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/platform-support.html#supported-nvidia-data-center-gpus-and-systems).

<Info>

**Note**

* NVIDIA DGX Spark, NVIDIA Jetson, and workstations are not supported.
* vGPU is not supported. NVIDIA Run:ai currently supports GPU passthrough only.

</Info>

The following are the minimum hardware requirements for each worker node. To designate nodes to NVIDIA Run:ai workloads, see [Worker nodes](../../configuration/node-roles.md#worker-nodes).

| Component | Required Capacity |
|-----------|-------------------|
| CPU       | 2 cores           |
| Memory    | 4 GB              |

## Operating System

* Any **Linux** operating system supported by both Kubernetes and NVIDIA GPU Operator.
* NVIDIA Run:ai on Google Kubernetes Engine (GKE) supports both Ubuntu and Container Optimized OS (COS).
  * COS is supported only with NVIDIA GPU Operator 24.6 and above, and NVIDIA Run:ai cluster version 2.19 and above.
* NVIDIA Run:ai on Elastic Kubernetes Service (EKS) does not support Bottlerocket or Amazon Linux.
* NVIDIA Run:ai on Oracle Kubernetes Engine (OKE) supports only Ubuntu.
* Internal tests are performed on **Ubuntu 22.04** and **CoreOS** for OpenShift.

## Network Time Protocol

Nodes must be synchronized using NTP (Network Time Protocol) for proper system functionality.

## Kubernetes

### Supported Distributions

The following Kubernetes distributions are supported for both the NVIDIA Run:ai control plane and NVIDIA Run:ai cluster:

* Vanilla Kubernetes
* OpenShift Container Platform (OCP)
* Elastic Kubernetes Service (EKS)
* Google Kubernetes Engine (GKE)
* Azure Kubernetes Service (AKS)
* Oracle Kubernetes Engine (OKE)
* Rancher Kubernetes Engine 2 (RKE2)

<Info>

**Note**

The latest release of NVIDIA Run:ai supports **Kubernetes 1.34 to 1.36** and **OpenShift 4.18 to 4.21**.

</Info>

For the full list of compatible distributions and per-version compatibility, see the [Support Matrix](../support-matrix.md#nvidia-run-ai-compatible-distributions).

### Partner-Compatible Distributions

The following distributions are tested and validated by the partner, who is responsible for maintaining compatibility with NVIDIA Run:ai:

* Crusoe Managed Kubernetes (CMK)
* Mirantis k0rdent
* Rafay platform
* Spectro Cloud
* vCluster
* VMware vSphere Kubernetes Service (VKS)

For per-partner version compatibility, see the [Support Matrix](../support-matrix.md#partner-compatible-distributions).

### Container Runtime

NVIDIA Run:ai supports the following [container runtimes](https://kubernetes.io/docs/setup/production-environment/container-runtimes/):

* [Containerd](https://kubernetes.io/docs/setup/production-environment/container-runtimes/#containerd) (default in Kubernetes)
* [CRI-O](https://cri-o.io/) (default in OpenShift)

### Permissions

Both the guided installer and the manual Helm path require **cluster-admin** permissions on the target Kubernetes cluster. Ensure your `kubectl` context is configured with a user that has cluster-admin or equivalent rights before proceeding.

## Storage

Two storage requirements apply before installation: a default storage class for the NVIDIA Run:ai control plane, and shared storage for the NVIDIA Run:ai cluster.

### Default Storage Class

<Info>

**Note**

Applies to Kubernetes only.

</Info>

The NVIDIA Run:ai control plane requires a **default storage class** to create persistent volume claims for NVIDIA Run:ai storage. The storage class, as per Kubernetes standards, controls the reclaim behavior - whether NVIDIA Run:ai persistent data is saved or deleted when the NVIDIA Run:ai control plane is deleted.

<Info>

**Note**

For a simple (non-production) storage class example, see [Kubernetes Local Storage Class](https://kubernetes.io/docs/concepts/storage/storage-classes/#local). The storage class will set the directory `/opt/local-path-provisioner` to be used across all nodes as the path for provisioning persistent volumes. Then set the new storage class as default:

```bash
kubectl patch storageclass local-path -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
```

</Info>

### Shared Storage

The NVIDIA Run:ai cluster requires shared storage so that training workloads can access data from any worker node in a uniform way - including training data, code, checkpoints, weights, and other machine-learning artifacts. Typical protocols are Network File Storage (NFS) or Network-Attached Storage (NAS). For more information, see [Shared Storage](../configuration/shared-storage.md).
