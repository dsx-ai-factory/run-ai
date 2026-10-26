# Using GB200 NVL72 and Multi-Node NVLink Domains

Multi-Node NVLink (MNNVL) systems, including NVIDIA GB200, NVIDIA GB200 NVL72 and its derivatives are fully supported by the NVIDIA Run:ai platform.

Kubernetes does not natively recognize NVIDIA’s MNNVL architecture, which makes managing and scheduling workloads across these high-performance domains more complex. The NVIDIA Run:ai platform simplifies this by abstracting the complexity of MNNVL configuration. Without this abstraction, optimal performance on a GB200 NVL72 system would require deep knowledge of NVLink domains, their hardware dependencies, and manual configuration for each distributed workload. NVIDIA Run:ai automates these steps, ensuring high performance with minimal effort. While GB200 NVL72 supports all [workload types](../../../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md), distributed workloads - training and inference - benefit most from its accelerated GPU networking capabilities.

<Note>

**Note**

Distributed workloads include both distributed training and distributed inference.

</Note>

To learn more about GB200, MNNVL and related NVIDIA technologies, refer to the following:

* [NVIDIA GB200 NVL72](https://www.nvidia.com/en-us/data-center/gb200-nvl72/)
* [NVIDIA Blackwell datasheet](https://nvdam.widen.net/s/wwnsxrhm2w/blackwell-datasheet-3384703)
* [NVIDIA Multi-Node NVLink Systems](https://docs.nvidia.com/multi-node-nvlink-systems/)

## Benefits of Using GB200 NVL72 with NVIDIA Run:ai

The NVIDIA Run:ai platform enables administrators, researchers, and MLOps engineers to fully leverage GB200 NVL72 systems and other NVLink-based domains without requiring deep knowledge of hardware configurations or NVLink topologies. Key capabilities include:

* **Automatic detection and labeling**
  * Detects GB200 NVL72 nodes and identifies MNNVL domains (e.g., GB200 NVL72 racks).
  * Automatically detects whether a node pool contains GB200 NVL72.
  * Supports manual override of GB200 MNNVL detection and label key for future compatibility and improved resiliency.
* **Simplified distributed workload submission**
  * Allows seamless submission of distributed workloads into GB200-based node pools, eliminating all the complexities involved with that operation on top of GB200 MNNVL domains.
  * Abstracts away the complexity of configuring workloads for NVL domains.
  * Automatically optimizes distributed workloads performance.
* **Flexible support for NVLink domain variants**
  * Compatible with current and future NVL domain configurations.
  * Supports any number of domains or GB200 racks.
* **Enhanced monitoring and visibility**
  * Provides detailed NVIDIA Run:ai dashboards for monitoring GB200 nodes and MNNVL domains by node pool.
* **Control and customization**
  * Offers manual override and label configuration for greater resiliency and future-proofing.
  * Enables advanced users to fine-tune GB200 scheduling behavior based on workload requirements.
* **Full support for elastic distributed workloads**
  * Elastic distributed workloads (auto-scaling or dynamically sized distributed workloads) are fully supported on GB200 NVL72 and MNNVL domains.
  * The NVIDIA Run:ai platform automatically applies ComputeDomain configuration and topology-aware scheduling rules to ensure elastic workloads scale within the same NVLink domain while maintaining optimal performance.
  * Scaling events such as adding or removing worker replicas respect NVLink domain boundaries and inherit the same high-bandwidth placement optimizations as fixed-size distributed workloads.

## Prerequisites

* **Kubernetes version** - Requires Kubernetes 1.32 or later.
* **NVIDIA GPU Operator** - Install the NVIDIA GPU Operator. See the [NVIDIA GPU Operator](../../../installation-and-configuration/installation/guided-installation/prerequisites-and-features.md#nvidia-gpu-operator) section for installation instructions.
* **NVIDIA Dynamic Resource Allocation (DRA) driver** - Install the NVIDIA DRA driver, which provides support for GB200 accelerated networking resources and the ComputeDomain feature. Depending on the GPU Operator mode, the DRA driver is installed separately (`ClusterPolicy` mode, the default) or deployed by the GPU Operator (`GPUCluster` mode, GPU Operator 26.7). For details, refer to [NVIDIA Dynamic Resource Allocation (DRA) Driver](../../../installation-and-configuration/installation/guided-installation/prerequisites-and-features.md#nvidia-dynamic-resource-allocation-dra-driver).
* **NVIDIA Network Operator** - Install the NVIDIA Network Operator. See the [NVIDIA Network Operator](../../../installation-and-configuration/installation/guided-installation/prerequisites-and-features.md#nvidia-network-operator) section for installation instructions.
* **Enable GPU network acceleration** - After installation, update the cluster configuration to enable GPU network acceleration by setting the appropriate flag for the relevant controller. Enabling either flag triggers an update of the corresponding NVIDIA Run:ai controller deployment and automatically restarts the controller to apply the change. Configure the flag, `GPUNetworkAccelerationEnabled=true`, for the controller that applies to your workloads. For details on how to configure this value using Helm or `runaiconfig`, see [Advanced cluster configurations](../../../installation-and-configuration/configuration/cluster-config.md).

## Configuring and Managing GB200 NVL72 Domains

Administrators must define dedicated node pools that align with GB200 NVL72 rack topologies. These node pools ensure that workloads are isolated to nodes with NVLink interconnects and are not scheduled on incompatible hardware. Each node pool can be manually configured in the NVIDIA Run:ai platform and associated with specific node labels. Two key configurations are required for each node pool:

* **Node Labels** - Identify nodes equipped with GB200.
* **MNNVL Domain Discovery** - Specify how the platform detects whether the node pool includes NVLink-connected nodes.

To create a node pool with GPU network acceleration, see [Node pools](../../aiinitiatives/resources/node-pools.md).

### Identifying GB200 Nodes

To enable the [NVIDIA Run:ai Scheduler](../../runai-scheduler/scheduling/how-the-scheduler-works.md) to recognize GB200-based nodes, administrators must:

* Use the default node label provided by the NVIDIA GPU Operator - `nvidia.com/gpu.clique`.
* Or, apply a custom label that clearly marks the node as GB200/MNNVL capable.

This node label serves as the basis for identifying appropriate nodes and ensuring workloads are scheduled on the correct hardware.

### Enabling MNNVL Domain Discovery

The administrator can configure how the NVIDIA Run:ai platform detects MNNVL domains for each node pool. The available options include:

* **Auto-detect** - Uses the default label key `nvidia.com/gpu.clique`, or a custom label key specified by the administrator. The NVIDIA Run:ai platform automatically discovers MNNVL domains within node pools. If a node is labeled with the MNNVL label key, the NVIDIA Run:ai platform indicates this node pool as MNNVL detected. MNNVL detected node pools are treated differently by the NVIDIA Run:ai platform when submitting a distributed workload.
* **MNNVL is present** - Explicitly indicate that the node pool contains MNNVL nodes
* **MNNVL is not present** - Explicitly indicate that the node pool does not contain MNNVL nodes

When auto-detect is enabled, all GB200 nodes that are part of the same physical rack (NVL72 or other future topologies) are part of the same NVL Domain and automatically labeled by the GPU Operator with a common label using a unique label value per domain and sub-domain. The default label key set by the NVIDIA GPU Operator is `nvidia.com/gpu.clique` and its value consists of - `<NVL Domain ID (ClusterUUID)>.<Clique ID>` :

* The **NVL Domain ID (ClusterUUID)** is a unique identifier that represents the physical NVL domain, for example, a physical GB200 NVL72 rack.
* The **Clique ID** denotes a logical MNNVL sub-domain. A clique represents a further logical split of the MNNVL into smaller domains that enable secure, fast, and isolated communication between pods running on different GB200 nodes within the same GB200 NVL72.

The [Nodes table](../../aiinitiatives/resources/nodes.md) provides more information on which GB200 NVL72 domain each node belongs to, and which Clique ID it is associated with.

### Submitting Distributed Workloads

When a distributed workload is submitted to an MNNVL node pool, the NVIDIA Run:ai platform automates several key configuration steps to ensure optimal workload execution:

* **ComputeDomain creation** - The NVIDIA Run:ai platform creates a ComputeDomain Custom Resource Definition (CRD), which is a proprietary resource used to manage NVLink-based domain assignments.
* **Resource Claim injection** - A reference to the ComputeDomain is automatically added to the workload specification as a resource claim, allowing the Scheduler to link the workload to a specific NVLink domain.
* **Network topology injection** - If network topology is configured on the GB200 node pool, the NVIDIA Run:ai platform adds a `NetworkTopology` request to the distributed workloads. This applies a default Preferred constraint at the lowest defined topology level, ensuring pods are placed as close as possible in the network hierarchy. The Scheduler uses the defined topology hierarchy to minimize communication overhead and improve workload performance, delivering more optimal scheduling results than the `PodAffinity` plugin. While pod affinity evaluates placement one pod at a time, network topology is aware of all pods in the workload and optimizes their placement accordingly. See [Accelerating workloads with network topology-aware scheduling](topology-aware-scheduling/workload-level-topology.md) for more details.
* **Pod affinity configuration** (if network topology is not configured) - Pod affinity is applied using a Preferred policy with the MNNVL label key (e.g., `nvidia.com/gpu.clique`) as the topology key. This ensures that pods within the distributed workload are located on nodes with NVLink interconnects.
* **Node affinity configuration** - Node affinity is also applied using a Preferred policy based on the same label key, further guiding the Scheduler to place workloads within the correct node group.

These additional steps are crucial for the creation of underlying HW resources (also known as IMEX channels) and stickiness of the distributed workload to MNNVL topologies and nodes. When a distributed workload is stopped or evicted, the platform automatically removes the corresponding ComputeDomain.

## Best Practices for MNNVL Node Pool Management

* Attaching a network topology to an MNNVL node pool, especially one that includes the MNNVL level label (e.g. GB200 rack) can significantly improve the Scheduler's optimal placement of distributed workloads.
* Avoid using the hostname label in network topology for MNNVL node pools. Assigning more than one pod to the same hostname will cause the request to fail, since a compute domain can schedule only one pod per node. Instead, rely on higher-level topology labels to ensure proper placement and avoid scheduling conflicts.
* When submitting a distributed workload, you should explicitly specify a list of one or more MNNVL node pools, or a list of one or more non-MNNVL node pools. A mix of MNNVL and non-MNNVL node pools is not supported. A GB200 MNNVL node pool is a pool that contains at least one node belonging to an MNNVL domain.
* Other workload types (not distributed) can include a list of mixed MNNVL and non-MNNVL node pools, from which the Scheduler will choose.
* MNNVL node pools can include any size of MNNVL domains (i.e. NVL72 and any future domain size) and support any Grace-Blackwell models (GB200 and any future models).
* To support the submission of larger distributed workloads, it is recommended to group as many GB200 racks as possible into fewer node pools. When possible, use a single GB200 node pool, unless there is a specific operational reason to divide resources across multiple node pools.
* When submitting distributed workloads with the controller pod set as a distinct non-GPU workload, the MNNVL feature should be used with the default Preferred mode as explained in the below section.

## Fine-Tuning Distributed Workload Constraints

You can influence how the Scheduler places distributed workloads into GB200 MNNVL node pools using the **Topology** field available in the [distributed training](../../../workloads-in-nvidia-run-ai/using-training/distributed-training-models.md) and [distributed inference](../../../workloads-in-nvidia-run-ai/using-inference/distributed-inference.md) workload submission forms.

<Note>

**Note**

The following options are based on [inter-pod affinity rules](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/), which define how pods are grouped based on topology.

</Note>

* **Confine a workload to a single GB200 MNNVL domain** - To ensure the workload is scheduled within a single GB200 MNNVL domain (e.g., a GB200 NVL72 rack), apply a label with a Required policy using the MNNVL label key (`nvidia.com/gpu.clique`). This instructs the Scheduler to strictly place all pods within the same MNNVL domain. If the workload exceeds 18 pods (or 72 GPUs), the Scheduler will not be able to find a matching domain and will fail to schedule the workload.
* **Try to schedule a workload using a Preferred topology** - To guide the Scheduler to prioritize a specific topology without enforcing it, apply a label with a policy of Preferred. You can apply any label with a Preferred policy. These labels are treated with higher scheduling weight than the default Preferred pod affinity automatically applied by NVIDIA Run:ai for MNNVL.
* **Mandate a custom topology** - To force scheduling a workload into a custom topology, add a label with a policy of Required. This ensures the workload is strictly scheduled according to the specified topology. Keep in mind that using a Required policy can significantly constrain scheduling. If matching resources are not available, the Scheduler may fail to place the workload.

## Fine-tuning MNNVL per Workload

You can customize how the NVIDIA Run:ai platform applies the MNNVL feature to each distributed workload. This allows you to override the default behavior when needed. To configure this behavior, set the proprietary label key `run.ai/MNNVL` in the **General settings** section of the [distributed training](../../../workloads-in-nvidia-run-ai/using-training/distributed-training-models.md) and [distributed inference](../../../workloads-in-nvidia-run-ai/using-inference/distributed-inference.md) workload submission forms. The following values are supported:

* **None** - Disables the MNNVL feature for the workload. The platform does not create a ComputeDomain and no pod affinity or node affinity is applied by default.
* **Preferred** (default) _-_ Indicates that MNNVL feature is preferred but not required. This is the default behavior when submitting a distributed workload:
  * If the workload is submitted to a non-MNNVL node pool, then the NVIDIA Run:ai platform does not add a ComputeDomain, ComputeDomain claim, pod affinity or node affinity for MNNVL nodes.
  * Otherwise, if the workload is submitted to a 'MNNVL' node pool, then the NVIDIA Run:ai platform automatically adds: ComputeDomain, ComputeDomain claim, NodeAffinity and PodAffinity both with a Preferred policy and using the MNNVL label.
  * If you manually add an additional Preferred topology label, it will be given higher scheduling weight than the default embedded pod affinity (which has weight = 1).
* **Required** - Enforces a strict use of MNNVL domains for the workload. The workload must be scheduled on MNNVL supported nodes:
  * The NVIDIA Run:ai platform creates a ComputeDomain and ComputeDomain claim.
  * The NVIDIA Run:ai platform will automatically add a node affinity rule with a Required policy using the appropriate label.
  * Pod affinity is set to Preferred by default, but you can override it manually with a Required pod affinity rule using the MNNVL label key or another custom label.
  * If any of the targeted node pools do not support MNNVL or if the workload (or any of its pods) does not request GPU resources, the workload will fail to run.

## Known Limitations and Compatibility

* If the DRA driver is not installed correctly in the cluster, particularly if the required CRDs are missing, and the MNNVL feature is enabled in the NVIDIA Run:ai platform, the workload controller will enter a crash loop. This will continue until the DRA driver is properly installed with all necessary CRDs or the MNNVL feature is disabled in the NVIDIA Run:ai platform.
* To run workloads on a GB200 node pool (i.e., a node pool MNNVL-enabled), the workload must explicitly request that node pool. To prevent unintentional use of MNNVL node pools, administrators must ensure these node pools are not included in any project's default list of node pools.
* Only one distributed workload per node can use GB200 accelerated networking resources. If GPUs remain unused on that node, other workload types may still utilize them.
* Workloads created in versions earlier than 2.21 do not include GB200 MNNVL node pools and are therefore not expected to experience compatibility issues.
* If a node pool that was previously used in a workload submission is later updated to include GB200 nodes (i.e., becomes a mixed node pool), the workload submitted before version 2.21 will not use any accelerated networking resources, although it may still run on GB200 nodes.
