# Node Roles

This guide explains how to designate specific node roles in a Kubernetes cluster to ensure optimal performance and reliability in production deployments.

For optimal performance in production clusters, it is essential to avoid extensive CPU usage on GPU nodes where possible. This can be done by ensuring the following:

* NVIDIA Run:ai system-level services run on dedicated CPU-only nodes.
* Workloads that do not request GPU resources (e.g. Machine Learning jobs) are executed on CPU-only nodes.

NVIDIA Run:ai services are scheduled on the defined node roles by applying [Kubernetes Node Affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity) using node labels .

## Configure Node Roles

The following node roles can be configured on the cluster:

* **System node:** Reserved for NVIDIA Run:ai system-level services.
* **GPU Worker node:** Dedicated for GPU-based workloads.
* **CPU Worker node:** Used for CPU-only workloads.

### System Nodes

NVIDIA Run:ai system nodes run system-level services required to operate. This can be done via [Kubectl](https://kubernetes.io/docs/reference/kubectl/).

By default, NVIDIA Run:ai applies a node affinity rule to prefer nodes that are labeled with `node-role.kubernetes.io/runai-system` for system services scheduling. You can modify the default node affinity rule by:

* Editing the `global.affinity` configuration parameter as detailed in [Advanced cluster configurations](cluster-config.md).
* Editing the `global.affinity` configuration as detailed in [Advanced control plane configurations](control-plane-config.md) for self-hosted deployments.

To set a system role for a node in your Kubernetes cluster using Kubectl, follow these steps:

1. Use the `kubectl get nodes` command to list all the nodes in your cluster and identify the name of the node you want to modify.
2.  Run one of the following commands to label the node with its role:

    ```bash
    kubectl label nodes <node-name> node-role.kubernetes.io/runai-system=true
    kubectl label nodes <node-name> node-role.kubernetes.io/runai-system=false
    ```

<Info>

**Note**

* To ensure high availability and prevent a single point of failure, it is recommended to configure at least three system nodes in your cluster.
* By default, Kubernetes master nodes are configured to prevent workloads from running on them as a best-practice measure to safeguard control plane stability. While this restriction is generally recommended, certain NVIDIA reference architectures allow adding tolerations to the NVIDIA Run:ai deployment so critical system services can run on these nodes.

</Info>

### Worker Nodes

NVIDIA Run:ai worker nodes run user-submitted workloads and system-level [DeamonSets](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/) required to operate. This can be managed via [Kubectl](https://kubernetes.io/docs/reference/kubectl/).

By default, GPU workloads are scheduled on GPU nodes based on the `nvidia.com/gpu.present` label. When `clusterConfig.global.nodeAffinity.restrictScheduling` is set to true via the [Advanced cluster configurations](cluster-config.md):

* GPU Workloads are scheduled with node affinity rule to require nodes that are labeled with `node-role.kubernetes.io/runai-gpu-worker`
* CPU-only Workloads are scheduled with node affinity rule to require nodes that are labeled with `node-role.kubernetes.io/runai-cpu-worker`

To set a worker role for a node in your Kubernetes cluster using Kubectl, follow these steps:

1. Validate the `clusterConfig.global.nodeAffinity.restrictScheduling` is set to true in the cluster’s [Configurations](cluster-config.md).
2. Use the `kubectl get nodes` command to list all the nodes in your cluster and identify the name of the node you want to modify.
3.  Run one of the following commands to label the node with its role. Replace the label and value (`true`/`false`) to enable or disable GPU/CPU roles as needed:

    ```bash
    kubectl label nodes <node-name> node-role.kubernetes.io/runai-gpu-worker=true
    kubectl label nodes <node-name> node-role.kubernetes.io/runai-cpu-worker=false
    ```
