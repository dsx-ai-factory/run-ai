# Configure GB200 NVL72 Placement

Influence how the Scheduler places a distributed workload in a GB200 NVL72 node pool, and set the MNNVL behavior for each workload. For the prerequisites and how the platform configures MNNVL domains, refer to [Using GB200 NVL72 and Multi-Node NVLink Domains](../../../platform-management/aiinitiatives/resources/using-gb200.md).

## Fine-Tuning Distributed Workload Constraints

You can influence how the Scheduler places distributed workloads into GB200 MNNVL node pools using the **Topology** field available in the [distributed training](../../using-training/distributed-training-models.md) and [distributed inference](../../using-inference/distributed-inference.md) workload submission forms.

<Note>

**Note**

The following options are based on [inter-pod affinity rules](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/), which define how pods are grouped based on topology.

</Note>

* **Confine a workload to a single GB200 MNNVL domain** - To ensure the workload is scheduled within a single GB200 MNNVL domain (e.g., a GB200 NVL72 rack), apply a label with a Required policy using the MNNVL label key (`nvidia.com/gpu.clique`). This instructs the Scheduler to strictly place all pods within the same MNNVL domain. If the workload exceeds 18 pods (or 72 GPUs), the Scheduler will not be able to find a matching domain and will fail to schedule the workload.
* **Try to schedule a workload using a Preferred topology** - To guide the Scheduler to prioritize a specific topology without enforcing it, apply a label with a policy of Preferred. You can apply any label with a Preferred policy. These labels are treated with higher scheduling weight than the default Preferred pod affinity automatically applied by NVIDIA Run:ai for MNNVL.
* **Mandate a custom topology** - To force scheduling a workload into a custom topology, add a label with a policy of Required. This ensures the workload is strictly scheduled according to the specified topology. Keep in mind that using a Required policy can significantly constrain scheduling. If matching resources are not available, the Scheduler may fail to place the workload.

## Fine-tuning MNNVL per Workload

You can customize how the NVIDIA Run:ai platform applies the MNNVL feature to each distributed workload. This allows you to override the default behavior when needed. To configure this behavior, set the proprietary label key `run.ai/MNNVL` in the **General settings** section of the [distributed training](../../using-training/distributed-training-models.md) and [distributed inference](../../using-inference/distributed-inference.md) workload submission forms. The following values are supported:

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
