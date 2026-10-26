# The NVIDIA Run:ai Scheduler: Concepts and Principles

When a user [submits a workload](../../../workloads-in-nvidia-run-ai/workloads.md), it is directed to the designated Kubernetes cluster and managed by the NVIDIA Run:ai Scheduler. The Scheduler’s primary role is to optimize placement, allocating workloads to the most suitable nodes based on specific resource requirements, workload characteristics, and organizational policies. This ensures high utilization while strictly adhering to NVIDIA Run:ai’s fairness and quota management logic.

The Scheduler is platform-agnostic, supporting native Kubernetes workloads, NVIDIA Run:ai workloads, and various third-party frameworks. For a detailed breakdown of compatible types, see [Introduction to workloads](../../../workloads-in-nvidia-run-ai/introduction-to-workloads.md#workloads-in-nvidia-runai).

To better understand the logic driving these allocation decisions, get to know the key concepts, resource management and scheduling principles of the Scheduler.

## Workloads, Pod Groups and Sub Groups

### Workloads and Pod Groups

Workloads can range from a single pod running on individual nodes to distributed workloads using multiple pods, each running on a node (or part of a node). A large-scale training workload could use dozens of nodes; similarly, an inference workload could use many pods (replicas) across multiple nodes.

Every newly created pod is assigned to a pod group that represents one or multiple pods belonging to a single workload. For example, a distributed PyTorch training workload with 32 workers is grouped into a single pod group.

The NVIDIA Run:ai Scheduler is a workload scheduler. This means the Scheduler always uses the pod group associated with each pod and handles the pod as part of a larger group of pods, while taking into consideration the common characteristics of the pod group, such as [gang scheduling](concepts-and-principles.md#gang-scheduling), minimum replicas or workers, workload priority class, workload preemptibility policy, topology information, and more. These characteristics are applied consistently across the entire pod group.

### Multi-Level Sub Groups

Some workloads use a workload structure composed of multiple functions (functionality disaggregation). In this structure, each function can have one or more replicas, and each replica is made up of one or more pods, usually structured as leader and worker sets (distributed replicas). One example of such a workload structure is the NVIDIA Dynamo inference framework, which uses Grove as a Kubernetes API (CRD) and controller for the underlying pods. A single Dynamo workload is composed of three functions: an incoming router (gateway), a prefill, and a decoder. Each of these functions may have multiple replicas, where the prefill and decoder replicas are structured as leader-worker sets.

This workload structure requires framing sets of leader-worker replicas as gang-scheduled sub-groups, which may be further grouped into sets of scale replicas, forming a second level of sub-groups. Therefore, the NVIDIA Run:ai workload group structure, handled by the Scheduler, is a hierarchy consisting of a top-level pod group with subordinate sub-groups. In this hierarchical group structure, a sub-group always points to a higher-level sub-group and ultimately to the top-level pod group. By default, each pod group has at least one sub-group under which all workers or replicas are federated. The lowest level of sub-groups always federates individual pods.

## Scheduling Queue

A scheduling queue (or simply a queue) represents a scheduler primitive that manages the scheduling of workloads based on different parameters.

A queue is created for each [project/node pool pair](../../aiinitiatives/adapting-ai-initiatives.md#mapping-your-organization) and [department/node pool pair](../../aiinitiatives/adapting-ai-initiatives.md#mapping-your-organization). The NVIDIA Run:ai Scheduler supports hierarchical queueing, project queues are bound to department queues, per node pool. This allows an organization to manage per node pool quota, over quota and more parameters for projects and their associated departments.

## Resource Management

### Quota

Each project and department includes a set of deserved resource quotas, per node pool and resource type. For example, project “LLM-Train/Node Pool NV-H100” quota parameters specify the number of GPUs, CPUs(cores), and the amount of CPU memory that this project deserves to get when using this node pool. Non-preemptible workloads and the non-preemptible pods of semi-preemptible workloads (see [Priority and Preemption](concepts-and-principles.md#priority-and-preemption)) can only be scheduled if their requested resources are within the deserved resource quotas of their respective project/node pool and department/node pool.

#### Deserved Quota and Workload Scheduling

Deserved quota means that a project is entitled to use up to a maximum number of resources defined by its quota (for example, 10 GPUs). The Scheduler does not allow a project to exceed this limit within a given node pool for non-preemptible workloads and the non-preemptible pods of semi-preemptible workloads. A non-preemptible workload's pods, and the non-preemptible pods of a semi-preemptible workload, that are scheduled within the deserved quota are assured to keep running.

Preemptible workloads and the preemptible pods of semi-preemptible workloads, on the other hand, may be preempted by higher priority workloads using the same project / node pool resource quota.

[Background workloads](background-workloads.md) are excluded from quota and fairshare calculations entirely. They occupy resources but are invisible to the Scheduler when it plans, and are evicted when their resources are needed by other workloads.

#### When Deserved Quota Cannot Be Fully Satisfied

The Scheduler may not always be able to provide the full deserved quota for several reasons. For example, this may occur due to cluster fragmentation, additional workload constraints such as specific topology requirements, node failures, or situations where administrators assign projects or departments more quota than the physically available resources (that is, [over-subscription](concepts-and-principles.md#over-subscription)).

### Over Quota

Projects and departments can have a share in the unused resources of any node pool, beyond their quota of deserved resources. These resources are referred to as over quota resources. The administrator configures the [over quota parameters](../../aiinitiatives/organization/projects.md) per node pool for each project and department. Over quota resources can only be used by preemptible workloads and the preemptible pods of semi-preemptible workloads.

### Over Quota Weight

Projects can receive a share of the cluster/node pool unused resources when the over quota weight setting is enabled. The part each Project receives depends on its over quota **weight** value, and the total weights of all other projects over quota weights. The same applies to Departments. The administrator configures the [over quota weight parameters](../../aiinitiatives/organization/projects.md) per node pool for each project and department.

### Max GPU Device Allocation

Administrators can limit the maximum number of GPUs that a project can allocate per node pool (that is, deserved quota + over quota &lt;= max GPU device allocation). The same constraint applies at the department level.

A department’s max GPU device allocation effectively limits the total number of GPUs that can be allocated by all projects under that department. This applies even if the combined max GPU device allocations of the individual projects exceed the department’s max GPU device allocation.

### Project and Department Rank

<Note>

**Note**

NVIDIA Run:ai terminology for project priority and department priority has changed to project rank and department rank, across the UI and API.

</Note>

Administrators can assign each project/node pool pair a rank. The rank determines the project’s relative precedence for resources compared to its siblings within the same department and node pool pair. Projects with a higher rank are allocated resources first, both within their quota and over quota, and may also reclaim (preempt) resources from lower-ranked projects within the same department/node pool pair.

The same principle applies to departments: a department’s rank is relative to its sibling departments using the same node pool. Because a project’s rank is local to its associated department/node pool pair, a high-ranked project in a low-ranked department can have lower overall resource precedence than a lower-ranked project in a higher-ranked department.

### Multi-Level Quota System

Each project has a set of deserved resource quotas (GPUs, CPUs, and CPU memory) defined per node pool. Projects can exceed their deserved quota and receive a share of unused resources in the node pool beyond that quota. The same model applies at the department level.

The Scheduler first balances over-quota resources across departments and then, within each department, distributes those resources among its projects. A department’s deserved quota and over-quota limits constrain the total amount of resources that can be allocated by all projects within that department.

If a project still has available deserved quota but the department’s deserved quota is exhausted, the Scheduler does not allocate additional deserved resources to that project. The same rule applies to over-quota resources: over-quota capacity is first allocated to the department and only then divided among its projects.

### Resource Limit Parameter

Each project and department has a Limit parameter per resource type. This parameter defines the maximum amount of that resource that a project or department can allocate, effectively placing an upper bound on the combined deserved quota + over-quota resources.

The NVIDIA Run:ai API exposes the Limit parameter for all resource types at both the project and department levels. In the UI, this parameter is exposed only for GPUs and is referred to as Max GPU Device Allocation, as GPUs are typically the most critical resource in AI clusters.

By default, all projects and departments have a Limit value of Unlimited, meaning they can theoretically allocate resources up to the cluster’s total capacity. In practice, however, resource allocation is constrained by additional factors, such as quotas, over-quota distribution, and competition between departments and projects for unused resources.

### Over-Subscription

Over-subscription is a scenario where the sum of all deserved resource quotas surpasses the physical resources of the cluster or node pool. In this case, there may be scenarios in which the Scheduler cannot find matching nodes to all workload requests, even if those requests were within the deserved resource quota of their associated projects.

## Scheduling Principles

### Fairness (Fair Resource Distribution)

[Fairness](how-the-scheduler-works.md) is a major principle within the NVIDIA Run:ai scheduling system. It means that the NVIDIA Run:ai Scheduler always respects certain resource splitting rules (fairness) between projects and between departments.

### Fairshare and Fairshare Balancing

To implement fairness, the NVIDIA Run:ai Scheduler calculates a numerical value called fairshare for each project or department, per node pool. This value represents the sum of the project’s or department’s deserved resources (quota) plus its share of unused resources in that node pool (over-quota resources).

The Scheduler then takes the minimum of:

* the calculated fairshare, and
* the total resources requested by the project’s or department’s workloads in that node pool

and uses this value as the effective fairshare for the project or department.

The Scheduler aims to provide each project or department with the resources they deserve per node pool using two main parameters: deserved quota and deserved fairshare (that is, quota plus over-quota resources). If one project’s node pool queue is below its fairshare while another project’s node pool queue exceeds its fairshare, the Scheduler shifts resources between queues to rebalance fairness. This process may result in the preemption of some over-quota, preemptible workloads or the preemptible pods of semi-preemptible workloads.

### Time-Based Fairshare

Administrators can enable, per node pool, a fairshare mode called [time-based fairshare](how-the-scheduler-works.md#time-based-fairshare).

When this mode is enabled, the Scheduler continuously collects historical resource usage data for each project and department, per node pool. It evaluates each project’s and department’s GPU-hour consumption relative to its configured weight and uses this information to balance resource distribution more effectively across projects and departments over time.

This capability applies only to over-quota resources. Quota resources always have the highest allocation precedence. Only after all quota allocations are satisfied does the Scheduler distribute excess (over-quota) resources among departments and projects, according to their [ranks and weights](../../aiinitiatives/organization/projects.md).

This time-based fairshare calculation ensures that each project and department receives its fair share of resource processing time (GPUs, CPUs, and CPU memory) over extended periods. As a result, over-quota resource distribution becomes more stable and fair, reducing short-term imbalances.

### Priority and Preemption

NVIDIA Run:ai supports scheduling workloads using different priority and preemption policies:

* Workload's priority and preemption are two distinct parameters set in a workload. Workload's priority sets the scheduling precedence within a project, while workload's preemption sets whether the workload is preemptible, non-preemptible, or semi-preemptible.
* High priority workloads (pods) can preempt [lower priority workloads](concepts-and-principles.md#preemption-of-lower-priority-workloads-within-a-project) (pods) within the same scheduling queue (project), according to their preemption policy (i.e. if preemptible).
* If no explicit preemption policy parameter is set, the NVIDIA Run:ai Scheduler implicitly assumes any PriorityClass >= 100 is non-preemptible and any PriorityClass < 100 is preemptible.
* Cross project and cross department workload preemptions are referred to as [resource reclaim](concepts-and-principles.md#reclaim-of-resources-between-projects-and-departments) and are based on [fairness](concepts-and-principles.md#fairness-fair-resource-distribution) between queues rather than the priority of the workloads.

To make it easier for users to submit workloads, NVIDIA Run:ai preconfigured several Kubernetes PriorityClass objects. The NVIDIA Run:ai preset PriorityClass objects have their ‘preemptionPolicy’ always set to ‘PreemptLowerPriority’, regardless of their actual NVIDIA Run:ai preemption policy within the NVIDIA Run:ai platform.

A non-preemptible workload is only scheduled if in-quota and cannot be preempted after being scheduled, not even by a higher priority workload. A preemptible workload can go over-quota. It can either reserve resources within the quota, where those resources may be taken back by higher priority workloads within that project, or it can use temporary over-quota resources that may be taken back by other projects.

A semi-preemptible workload uses a mix of non-preemptible and preemptible pods, that is, workers or replicas. Pods within the min-members count, that is, the minimum scheduling gang, are non-preemptible, while pods above the min-members count are preemptible. Semi-preemptible workloads provide a minimum guaranteed service level for an inference or distributed training workload, while also using excess resources when demand exists and extra resources are available.

To see the *default* priority and preemption policy of a workload and for details on how to change the priority and preemption, see [Workload priority and preemption](workload-priority-control.md).

### Preemption of Lower Priority Workloads Within a Project

Workload priority is always respected within a project. This means higher priority workloads are scheduled before lower priority workloads. It also means that higher priority workloads may preempt lower priority workloads within the same project if the lower priority workloads are preemptible, or are the preemptible pods of a semi-preemptible workload.

### Reclaim of Resources Between Projects and Departments

[Reclaim](how-the-scheduler-works.md#reclaim-preemption-between-projects-and-departments) is an inter-project (and inter-department) scheduling action that takes back resources from one project (or department) that has used them as over quota, back to a project (or department) that deserves those resources as part of its deserved quota, or to balance fairness between projects, each to its fairshare (i.e. sharing fairly the portion of the unused resources).

### Gang Scheduling

Gang scheduling describes a scheduling principle in which a workload composed of multiple pods is either scheduled as a gang (at least the min-members number of pods are scheduled together) or fully pending (none of the pods are scheduled). Pods above the min-members count are not part of the gang. They are scheduled when resources are available and can remain pending without affecting the gang. Gang scheduling applies to a single pod group.

#### Multi-Level Gang Scheduling

The NVIDIA Run:ai Scheduler supports scheduling workloads that use multi-level pod-group structures, as described in [Workloads, pod groups, and sub groups](concepts-and-principles.md#workloads-pod-groups-and-sub-groups). To support these workloads, the Scheduler implements multi-level pod grouping.

The top level of the hierarchy is always a pod group, regardless of whether the workload is flat or hierarchical. Any grouping level below the top level is a sub-group. A workload may include multiple levels of sub-groups, as required, and each sub-group points to its immediate parent in the hierarchy.

A sub-group has parameters similar to those of a pod group. For example, it includes a min-members parameter, which indicates the minimum number of pods that must be scheduled for the sub-group to be considered ganged.

For a hierarchical workload to transition to the Running state, all sub-groups must be successfully ganged, including the top-level pod group.

#### Stale Gang Eviction

A pod group becomes stale when it has running pods (or sub-groups) but no longer meets its gang condition for that level, for example, when one or more pods fail or stop running while the rest continue, or when one or more sub-groups no longer meet their gang condition. The Scheduler detects the stale state and starts a grace period before evicting the remaining pods, giving the workload's controller time to recover the missing pods and restore the gang condition.

During the grace period, the running pods are not evicted. If enough pods return to a running state within the window, the pod group recovers and is no longer considered stale. If the grace period expires and the gang condition is still not met, the Scheduler evicts all running pods in the pod group, and it becomes eligible for rescheduling.

The grace period has a cluster-wide default and can be overridden per workload:

* **Cluster-wide default** - Configured with the `defaultStalenessGracePeriod` [advanced cluster configuration](../../../installation-and-configuration/configuration/cluster-config.md) parameter. Default: 60 seconds.
* **Per-workload override** - Set the `kai.scheduler/staleness-grace-period` annotation on the workload. The value is a duration string, a number followed by a unit (`s`, `m`, or `h`), for example `30s`, `5m`, or `1h`:
  * An invalid value is ignored, and the cluster-wide default applies.
  * A negative value disables stale gang eviction for that pod group entirely.
  * If the annotation is not set, the cluster-wide default applies.

### Placement Strategy - Bin-Pack and Spread

The administrator can set a [placement strategy](../../aiinitiatives/resources/node-pools.md#adding-a-new-node-pool), bin-pack or spread, of the Scheduler per node pool. For GPU based workloads, workloads can request both GPU and CPU resources. For CPU-only based workloads, workloads can request CPU resources only.

* **GPU workloads:**
  * **Node placement strategy** - Determines how workloads are distributed across nodes in the node pool:
    * **Bin-pack** - The Scheduler places as many workloads as possible in each GPU and node to use fewer resources and maximize GPU and node vacancy.
    * **Spread** - The Scheduler spreads workloads across as many GPUs and nodes as possible to minimize the load and maximize the available resources per workload.
  * **Device placement strategy** - Determines how workloads are distributed across GPU devices within each node:
    * **Bin-pack** - The Scheduler places as many workloads as possible on each GPU device to minimize the number of GPU devices used.
    * **Spread** - The Scheduler spreads workloads across GPU devices within each node to maximize the available resources per workload.
  * The Scheduler applies these strategies in order: first the node placement strategy to select nodes, then the device placement strategy to place pods across GPU devices within each selected node.
* **CPU workloads:**
  * **Bin-pack** - The Scheduler places as many workloads as possible in each CPU and node to use fewer resources and maximize CPU and node vacancy.
  * **Spread** - The Scheduler spreads workloads across as many CPUs and nodes as possible to minimize the load and maximize the available resources per workload.

## Workload Resources

The primary function of the NVIDIA Run:ai Scheduler is to match workloads with available nodes that satisfy their resource requirements and other constraints (such as node type or topology). Workload resources generally fall into three categories:

* **Kubernetes Native Resources** - CPUs and CPU memory.
* **Kubernetes Extended Resources** - Non-native Kubernetes resources such as GPUs, FPGAs, or networking adapters. Historically requested as static "extended resources" (for example, `nvidia.com/gpu: 2`). Starting with Kubernetes v1.34, workloads can leverage [Dynamic Resource Allocation (DRA)](concepts-and-principles.md#dynamic-resource-allocation-dra).
* **Kubernetes Storage Resources** - Persistent Volumes (PVs) and Persistent Volume Claims (PVCs), managed by the Kubernetes CSI subsystem.

### Dynamic Resource Allocation (DRA)

Dynamic Resource Allocation (DRA) is a Kubernetes-native mechanism that provides a more flexible alternative to static extended resources. Instead of requesting a fixed number of devices, workloads can express complex hardware requirements - such as GPU type, minimum GPU memory, GPU sharing, or priority lists - using `ResourceClaim` and `ResourceClaimTemplate` objects. `ResourceClaims` and `ResourceClaimTemplates` define the proprietary parameters of the requested hardware, such as GPUs or communication channels. DRA is GA in Kubernetes v1.34 and was Beta in v1.33, which is also supported.

NVIDIA Run:ai has introduced DRA support progressively. Each supported resource type in DRA requires the respective vendor to advertise their proprietary driver:

* **v2.20** - DRA support for `ComputeDomain` `ResourceClaims`
* **v2.25** - DRA support for NVIDIA GPUs
* **v3.0** - Translation of extended-resource requests into DRA claims

The NVIDIA Run:ai platform supports scheduling and presentation of workloads using DRA claims in two scenarios:

* Workloads submitted directly to the cluster
* Workloads submitted [via YAML](../../../workloads-in-nvidia-run-ai/submit-via-yaml.md) using NVIDIA Run:ai UI, API and CLI

In both scenarios, the platform schedules the workloads, accounts for the requested resources in quota management, and surfaces workload status and other parameters via the UI, API, and CLI.

#### Extended-Resource Translation

A translation layer in the Scheduler translates legacy extended-resource requests (for example, `nvidia.com/gpu: 2`) into DRA `ResourceClaim` requests automatically, so [NVIDIA Run:ai native workloads](../../../workloads-in-nvidia-run-ai/workload-types/native-workloads.md) and [supported workload types](../../../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md) can use DRA without submitting a claim directly.

Workloads that request extended resources and workloads that submit DRA claims can run in the same cluster when extended-resource translation is enabled. Translation requires all of the following:

* The NVIDIA DRA driver, deployed by the NVIDIA GPU Operator when it runs in `GPUCluster` mode (GPU Operator 26.7). In this mode, the GPU Operator does not deploy the device plugin, so workloads that request `nvidia.com/gpu` depend on translation.
* A `DeviceClass` that sets `spec.extendedResourceName` to the extended resource name in use, for example `nvidia.com/gpu`. The GPU Operator creates the `gpu.nvidia.com` `DeviceClass` but does not set this field, so the legacy name requires a separate `DeviceClass` created by the administrator.
* Kubernetes support for the DRA extended resource feature, which is generally available from Kubernetes 1.37.
* NVIDIA Run:ai v3.0 or later, which provides the Scheduler side of the translation.

For configuration details, refer to [Installing the NVIDIA DRA driver](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/26.7/dra-intro-install.html) and [Setting up DRA in a Kubernetes cluster](https://kubernetes.io/docs/tasks/configure-pod-container/assign-resources/set-up-dra-cluster/).

## Next Steps

Now that you have learned the key concepts and principles of the NVIDIA Run:ai Scheduler, see [how the Scheduler works](how-the-scheduler-works.md) - allocating pods to workloads, applying preemption mechanisms, and managing resources.
