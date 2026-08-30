# Configuring NUMA-Aware Scheduling

NUMA (Non-Uniform Memory Access) architecture divides the CPUs, memory, and GPUs within a physical server into multiple domains called NUMA nodes. Resources that span different NUMA nodes communicate over a higher-latency, lower-throughput interconnect, so workloads whose GPU and CPU allocations are split across NUMA boundaries can experience degraded performance.

NVIDIA Run:ai supports NUMA-aware scheduling at the node pool level. When enabled, the NVIDIA Run:ai Scheduler places workloads so that GPU, CPU, and CPU memory are allocated within the same NUMA node where possible, minimizing cross-NUMA fragmentation and improving performance. When disabled, the Scheduler places workloads using aggregate node capacity only, without considering NUMA topology.

NUMA-aware scheduling is especially important on high-density GPU servers such as DGX H100 or DGX B300, and in clusters where the Kubelet Topology Manager is configured to `restricted` or `single-numa-node` mode. These are more restrictive policies, and workloads may fail if NUMA-aware scheduling is not enabled on the corresponding node pool.

## Prerequisites

The following must be in place before enabling NUMA-aware scheduling on a node pool. These are configured outside of NVIDIA Run:ai.

<Info>

**Note**

Before setting up the prerequisites below, verify whether they are already configured in your cluster.

</Info>

### Node Feature Discovery (NFD) Topology Updater

NFD must be deployed with the topology updater enabled. The topology updater publishes per-node NUMA topology data as Kubernetes NodeResourceTopology custom resources, including available and allocated GPUs, CPUs, and memory per NUMA node, as well as the Kubelet Topology Manager policy and scope configured on each node. The NVIDIA Run:ai Scheduler reads these resources to make NUMA-aware placement decisions. See the [NFD Topology Updater documentation](https://kubernetes-sigs.github.io/node-feature-discovery/stable/usage/nfd-topology-updater.html) for deployment instructions.

### Kubelet Topology Manager Policy

The Kubelet Topology Manager policy controls how NUMA alignment is enforced for workload resource allocations. NVIDIA Run:ai reads this policy from the NFD topology data and takes it into account when scheduling. If the policy is set to `none` on the pool's nodes, enabling NUMA-aware scheduling has no practical effect. For a description of the available policies (`none`, `best-effort`, `restricted`, `single-numa-node`), see the [Kubernetes Topology Manager documentation](https://kubernetes.io/docs/tasks/administer-cluster/topology-manager/#topology-manager-policies).

The `restricted` and `single-numa-node` policies are the more restrictive modes. If the cluster admin has configured nodes in a node pool to one of these policies but NUMA-aware scheduling is not enabled on that node pool in NVIDIA Run:ai, workload submissions may fail intermittently. Without NUMA-aware scheduling enabled, the NVIDIA Run:ai Scheduler does not predict NUMA placement, which can result in pods being assigned to nodes where the Kubelet rejects them.

### Kubelet Topology Manager Scope

The `--topology-manager-scope` Kubelet setting controls whether NUMA alignment is computed per container (the default) or for the pod as a whole. NVIDIA Run:ai reads this setting from the NFD topology data alongside the topology manager policy and takes it into account when scheduling. Both `container` and `pod` scopes are supported. See [Topology Manager scopes](https://kubernetes.io/docs/tasks/administer-cluster/topology-manager/#topology-manager-scopes) for details.

### Additional Kubelet Requirements

For NUMA alignment to take effect, the following Kubelet settings must also be configured on the pool's nodes:

* [`--cpu-manager-policy=static`](https://kubernetes.io/docs/tasks/administer-cluster/cpu-management-policies/): required for the Topology Manager to align CPU allocations to NUMA nodes
* [`--memory-manager-policy=Static`](https://kubernetes.io/docs/tasks/administer-cluster/memory-manager/): required for the Topology Manager to align memory allocations to NUMA nodes

The Topology Manager only coordinates hints from the CPU manager and memory manager; without these policies set to static, NUMA alignment for CPU and memory has no effect.

## Enabling NUMA-Aware Scheduling

NUMA-aware scheduling is configured per node pool and is disabled by default. To enable it, toggle the **NUMA-aware scheduling** option when creating or editing a node pool. See [Node pools](https://run-ai-docs.nvidia.com/api/2.26/organizations/node-pools) for details.

<Info>

**Note**

The NUMA-aware scheduling toggle is only available for clusters running version 2.26 or later.

</Info>

## How It Works

When NUMA-aware scheduling is enabled on a node pool, the Scheduler reads each node's NUMA topology from NodeResourceTopology resources published by the NFD topology updater. This includes the Topology Manager policy and scope configured on each node, and the available and allocated GPUs, CPUs, and memory per NUMA node.

When placing a workload, the Scheduler predicts whether a node can satisfy the workload's GPU, CPU, and memory request within NUMA boundaries before the pod reaches the Kubelet. The goal is to determine placement compatibility early, minimize cross-NUMA fragmentation, and improve performance. Whether resources may span NUMA nodes depends on the Topology Manager policy configured on the node:

* **`best-effort`** - The Scheduler prefers NUMA-aligned placement but always admits the pod, even if resources must span NUMA nodes.
* **`restricted`** - Spanning across NUMA nodes is only allowed when every requested resource type (GPUs, CPUs, and memory) individually requires more than a single NUMA node. If one resource spans two nodes but another fits on one, the requests do not align and the pod is rejected. For example, on a server with 2 NUMA nodes and 4 GPUs each, a workload requesting 8 GPUs must span both nodes. If that workload also requests only 1 CPU and 100 GB of memory (both of which fit on a single NUMA node), the pod is rejected. The pod is admitted only when the CPU and memory requests are large enough to require both NUMA nodes as well, so all resources align to the same span.
* **`single-numa-node`** - All resources must fit within a single NUMA node. The workload is rejected if no single NUMA node can satisfy the full request.

If no node satisfies the workload's resource request within the policy constraints, the workload remains pending with a NUMA-specific event explaining the reason.

Enabling or disabling NUMA-aware scheduling does not affect workloads already running in the pool unless they are rescheduled.

## Best Practices

* Configure all nodes in the same node pool to the same Kubelet Topology Manager policy. Mixed policies within a pool can produce inconsistent scheduling results.
* For the lowest-risk approach to NUMA-aware scheduling, set all nodes to `best-effort`. The Kubelet attempts to align resources to a single NUMA node but always admits the pod, so workloads are never rejected due to NUMA constraints. This may produce less optimal performance than `restricted` or `single-numa-node`, but avoids placement failures.
* You can also configure all nodes to `best-effort` and keep NUMA-aware scheduling disabled in NVIDIA Run:ai. In this mode, the Kubelet still attempts to align resources to a single NUMA node on its own, but the NVIDIA Run:ai Scheduler does not predict NUMA placement. Workloads are never rejected due to NUMA constraints, though placement may be less optimal than with NUMA-aware scheduling enabled.

## Using API

To view the available actions, go to the [Node pools](https://run-ai-docs.nvidia.com/api/2.26/organizations/nodepools) API reference.
