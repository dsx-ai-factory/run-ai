# Before You Install

This page covers the infrastructure that must be in place before starting an NVIDIA Run:ai installation, regardless of the installation path chosen. These are the foundational requirements that apply to both the guided installer and the manual Helm path.

## Requirements Summary

| Category | Requirement | Details |
|---|---|---|
| Hardware | Control plane nodes, cluster nodes, and worker nodes provisioned | See [Hardware](#hardware) |
| Operating System | Linux installed on all nodes | See [Operating System](#operating-system) |
| Kubernetes | Distribution installed and running | See [Supported Distributions](#supported-distributions) |
| Kubernetes | Container runtime configured | See [Container Runtime](#container-runtime) |
| Kubernetes | Cluster-admin permissions | See [Permissions](#permissions) |
| Storage | Default storage class configured | See [Default Storage Class](#default-storage-class) |
| Storage | Shared storage accessible from all worker nodes | See [Shared Storage](#shared-storage) |
| Networking | FQDN and DNS configured | See [FQDN and DNS](#fqdn-and-dns) |
| Networking | TLS certificate files obtained | See [TLS Certificates](#tls-certificates) |
| Registry | Registry credentials available | See [Registry Credentials](#registry-credentials) |
| OpenShift | Pod Security Admission configured | See [Kubernetes Pod Security Admission](#kubernetes-pod-security-admission) |

## Hardware

The following hardware requirements apply to all nodes in the deployment. By default, all NVIDIA Run:ai services run on all available nodes. For production deployments, set [node roles](../post-installation-configuration/node-roles.md) to separate system nodes from worker nodes, reduce downtime, and avoid consuming CPU cycles on expensive GPU machines.

### Architecture

* **x86** — Supported for Kubernetes and OpenShift
* **ARM** — Supported for Kubernetes and OpenShift

### Control Plane — System Nodes

The following are the minimum requirements for NVIDIA Run:ai control plane system nodes. By default, all control plane services run on all available nodes in the control plane cluster.

| Component  | Required Capacity |
|------------|-------------------|
| CPU        | 10 cores          |
| Memory     | 12 GB             |
| Disk space | 110 GB            |

<Info>

**Note**

To designate nodes to NVIDIA Run:ai control plane services, follow the instructions in [System nodes](../post-installation-configuration/node-roles.md#system-nodes).

</Info>

### Cluster — System Nodes

The following are the minimum requirements for NVIDIA Run:ai cluster system nodes. By default, all cluster services run on all available nodes in the cluster.

| Component  | Required Capacity |
|------------|-------------------|
| CPU        | 10 cores          |
| Memory     | 20 GB             |
| Disk space | 50 GB             |

<Info>

**Note**

To designate nodes to NVIDIA Run:ai cluster services, follow the instructions in [System nodes](../post-installation-configuration/node-roles.md#system-nodes).

</Info>

### Cluster — Worker Nodes

The NVIDIA Run:ai cluster supports x86 and ARM CPUs, and any NVIDIA GPUs supported by the NVIDIA GPU Operator. The list of supported GPUs depends on the version of the NVIDIA GPU Operator installed in the cluster. For the list of supported GPUs, see [Supported NVIDIA Data Center GPUs and Systems](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/platform-support.html#supported-nvidia-data-center-gpus-and-systems).

<Info>

**Note**

* NVIDIA DGX Spark, NVIDIA Jetson, and workstations are not supported.
* vGPU is not supported. NVIDIA Run:ai currently supports GPU passthrough only.

</Info>

The following are the minimum hardware requirements for each worker node:

| Component | Required Capacity |
|-----------|-------------------|
| CPU       | 2 cores           |
| Memory    | 4 GB              |

<Info>

**Note**

To designate nodes to NVIDIA Run:ai workloads, follow the instructions in [Worker nodes](../post-installation-configuration/node-roles.md#worker-nodes).

</Info>

<Info>

**Note**

If deploying the NVIDIA Run:ai control plane and NVIDIA Run:ai cluster on the same Kubernetes cluster, the system node requirements are additive — provision nodes that meet both sets of requirements combined.

</Info>

## Operating System

* Any **Linux** operating system supported by both Kubernetes and NVIDIA GPU Operator.
* NVIDIA Run:ai on Google Kubernetes Engine (GKE) supports both Ubuntu and Container Optimized OS (COS).
  * COS is supported only with NVIDIA GPU Operator 24.6 and above, and NVIDIA Run:ai cluster version 2.19 and above.
* NVIDIA Run:ai on Elastic Kubernetes Service (EKS) does not support Bottlerocket or Amazon Linux.
* NVIDIA Run:ai on Oracle Kubernetes Engine (OKE) supports only Ubuntu.
* Internal tests are performed on **Ubuntu 22.04** and **CoreOS** for OpenShift.

Nodes must be synchronized using NTP (Network Time Protocol) for proper system functionality.

## Kubernetes

### Supported Distributions

The following Kubernetes distributions are supported for both the NVIDIA Run:ai control plane and NVIDIA Run:ai cluster:

* Vanilla Kubernetes
* OpenShift Container Platform (OCP)
* Elastic Kubernetes Engine (EKS)
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
* vCluster
* VMware vSphere Kubernetes Service (VKS)

For per-partner version compatibility, see the [Support Matrix](../support-matrix.md#partner-compatible-distributions).

### Container Runtime

NVIDIA Run:ai supports the following [container runtimes](https://kubernetes.io/docs/setup/production-environment/container-runtimes/):

* [Containerd](https://kubernetes.io/docs/setup/production-environment/container-runtimes/#containerd) (default in Kubernetes)
* [CRI-O](https://cri-o.io/) (default in OpenShift)

### Permissions

The installer requires **cluster-admin** permissions on the target Kubernetes cluster. Ensure your `kubectl` context is configured with a user that has cluster-admin or equivalent rights before running the installer.

## Storage

### Default Storage Class

<Info>

**Note**

Applies to Kubernetes only.

</Info>

The NVIDIA Run:ai control plane requires a **default storage class** to create persistent volume claims for NVIDIA Run:ai storage. The storage class, as per Kubernetes standards, controls the reclaim behavior — whether NVIDIA Run:ai persistent data is saved or deleted when the NVIDIA Run:ai control plane is deleted.

<Info>

**Note**

For a simple (non-production) storage class example, see [Kubernetes Local Storage Class](https://kubernetes.io/docs/concepts/storage/storage-classes/#local). The storage class will set the directory `/opt/local-path-provisioner` to be used across all nodes as the path for provisioning persistent volumes. Then set the new storage class as default:

```bash
kubectl patch storageclass local-path -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
```

</Info>

### Shared Storage

NVIDIA Run:ai workloads must be able to access data from any worker node in a uniform way, to access training data and code as well as save checkpoints, weights, and other machine-learning-related artifacts. Typical protocols are Network File Storage (NFS) or Network-attached storage (NAS). NVIDIA Run:ai cluster supports both. For more information, see [Shared Storage](../../infrastructure-setup/procedures/shared-storage.md).

## Networking

### FQDN and DNS

<Info>

**Note**

Applies to Kubernetes only.

</Info>

NVIDIA Run:ai services rely on Fully Qualified Domain Names (FQDNs) to route traffic between system components and to expose workloads externally. FQDNs are needed for:

* Enabling communication between the control plane and the cluster
* Exposing development workspaces and training workloads via subdomains
* Exposing inference workloads via dedicated subdomains

Configure domain names for each of the following:

* Control plane ↔ cluster communication. Example: `runai.mycorp.local`. The IP address of this domain must be resolvable within the organization's private network.
* Workspace and training workloads (external access). Example: `*.runai.mycorp.local`
* Inference workloads (external access). Example: `*.runai-inference.mycorp.local`

<Info>

**Note**

The inference wildcard (`*.runai-inference.mycorp.local`) is only required if you plan to expose inference workloads outside the cluster. For cluster-local inference, skip the inference wildcard record and the matching inference wildcard certificate.

</Info>

Since NVIDIA Run:ai uses host-based routing, wildcard DNS records must be configured to enable external access to workloads. Both records should resolve to the same cluster public IP address, or to the cluster's load balancer IP in on-premises environments:

```bash
*.runai.mycorp.local           → <cluster IP or load balancer IP>
*.runai-inference.mycorp.local → <cluster IP or load balancer IP>
```

### TLS Certificates

TLS certificates secure communication between NVIDIA Run:ai components and enable HTTPS access to exposed services. NVIDIA Run:ai requires three TLS certificates, each aligned with a specific domain:

* **Cluster domain certificate** — secures communication between the control plane and the cluster
* **Workspaces and training wildcard certificate** — secures external access to workspace and training workload subdomains
* **Inference wildcard certificate** — secures external access to inference workload subdomains (only required if exposing inference workloads externally)

Obtain the certificate and private key files for each domain before installation. How they are applied depends on the installation path — the guided installer accepts file paths in the configuration file; the Helm path requires creating Kubernetes secrets manually. If certificates from a trusted CA are not available, the guided installer can generate self-signed certificates instead.

## Registry Credentials

The installer pulls NVIDIA Run:ai images and Helm charts from a container registry. Prepare the appropriate credentials before installation:

| Offering | Registry | Credentials needed |
|---|---|---|
| Open Source | NVIDIA NGC (public) | None — NGC is publicly accessible for Run:ai artifacts |
| Enterprise | NVIDIA NGC | NGC API key |
| Air-gapped (any) | Local registry | Registry URL, username, and password (or token) |

For air-gapped installations, you will also need the Run:ai installation bundle downloaded and accessible on the machine running the installer.

## Kubernetes Pod Security Admission

NVIDIA Run:ai supports `restricted` policy for [Pod Security Admission](https://kubernetes.io/docs/concepts/security/pod-security-admission/) (PSA) on OpenShift only. Other Kubernetes distributions are only supported with `privileged` policy.

For NVIDIA Run:ai on OpenShift to run with PSA `restricted` policy:

* Label the `runai` namespace with the following labels:

  ```bash
  pod-security.kubernetes.io/audit=privileged
  pod-security.kubernetes.io/enforce=privileged
  pod-security.kubernetes.io/warn=privileged
  ```

* Workloads submitted through NVIDIA Run:ai must comply with the PSA restricted policy restrictions. This can be enforced using [Policies](../../platform-management/policies/native-workload-policies.md).

## Next Steps

Once your environment is prepared, review [Third-Party Services](third-party-services.md) to understand the bundled services and decide if you are bringing any external instances. Then proceed to installation:

* [Guided installation](guided-installation/install-with-the-installer.md) — recommended for most installations.
* [Manual installation](install-using-helm/install-control-plane.md) — for organizations managing releases through their own tooling.
