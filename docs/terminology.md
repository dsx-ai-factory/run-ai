# NVIDIA Run:ai Controlled Word List

This list defines approved NVIDIA Run:ai technical terms for explanatory text. It gives each term one product meaning and identifies alternatives that can make that meaning unclear.

This list is not a general English dictionary.

## Use the List

- Use the approved term with the meaning and form shown here.
- Treat an alternative as discouraged only when it refers to the same concept.
- Use a different term when it identifies a real difference. Define that difference at first use.
- Preserve literal identifiers, commands, output, API fields, quotations, and official third-party names.
- Apply the list to changed text. Report unrelated terminology debt in a focused follow-up.

## Product and Component Names

| Use | Use as | Meaning | Avoid for this meaning |
|---|---|---|---|
| `NVIDIA Run:ai` | Product name | The full product name. | RunAI, Run.ai, Run:AI |
| `Run:ai` | Short product name | Short form of NVIDIA Run:ai when the product is already established. | RunAI, Run.ai, Run:AI |
| `runai` | Technical noun | CLI commands, API flags, code blocks, and inline code only. | NVIDIA Run:ai or Run:ai in code contexts |
| `NVIDIA Run:ai OSS` | Product name | The open source edition. Spell out "Open Source Software" on first mention. | OSS alone without first spelling it out |
| `NVIDIA Run:ai Support-entitled Release` | Product name | The supported, non-OSS release. | NVIDIA Run:ai Enterprise (retired term) |
| `Self-hosted` | Deployment type | Customer-managed deployment. | Self-Hosted (use sentence case, not title case, for this compound term) |
| `SaaS` | Deployment type | NVIDIA-hosted, software-as-a-service deployment. | — |
| `Multi-tenant` | Deployment type | A host-organization-managed deployment serving multiple tenants. | Multi-Tenant (use sentence case, not title case, for this compound term) |
| `NVIDIA Run:ai control plane` | Architecture concept | The central NVIDIA Run:ai management layer. Use `control plane` when the full name is already clear. | Omitting "NVIDIA Run:ai" on first mention |
| `NVIDIA Run:ai cluster` | Architecture concept | Full form on first mention; "the cluster" after. | Omitting "NVIDIA Run:ai" on first mention |
| `NVIDIA Run:ai Scheduler` | Component name | The NVIDIA Run:ai component responsible for workload scheduling and resource allocation. Use `Scheduler` when the full name is already clear. | Run:ai scheduler when the named component is intended |

Qualify `control plane` when necessary to distinguish the NVIDIA Run:ai control plane from the Kubernetes control plane.

## Workloads

| Use | Use as | Meaning | Avoid for this meaning |
|---|---|---|---|
| `distributed inference workload` | Workload type | An inference workload that uses distributed execution. | distributed inference job |
| `distributed training workload` | Workload type | A workload used for distributed model training. | distributed training job |
| `externally submitted workload` | Product concept | A workload submitted outside NVIDIA Run:ai and subsequently managed by NVIDIA Run:ai. | external job |
| `inference workload` | Workload type | A workload used to serve a model for inference. | inference job |
| `native workload` | Product concept | A workload created using a workload type natively supported by NVIDIA Run:ai. | — |
| `non-preemptible` | Scheduling property | Describes a workload protected from scheduling preemption after it starts. | non-preemptive |
| `preemptibility` | Scheduling property | The property that determines whether a workload is preemptible, semi-preemptible, or non-preemptible. | — |
| `preemptible` | Scheduling property | Describes a workload that can be interrupted by the Scheduler. | preemptive |
| `semi-preemptible` | Scheduling property | Describes a workload whose minimum required replicas are protected from interruption, while any additional replicas behave as preemptible replicas. | — |
| `standard training workload` | Workload type | A non-distributed training workload. | standard training job |
| `supported workload type` | Product concept | A Kubernetes or ML workload type integrated with NVIDIA Run:ai. | native workload when the distinction matters |
| `training workload` | Workload type | A workload used to train a model. | training job |
| `workload` | Product concept | A unit of compute work submitted to or managed by NVIDIA Run:ai. | job, task |
| `workload type` | Product concept | A supported classification of workload. | job type |
| `workspace` | Workload type | An interactive workload for development and experimentation. | job |

Use `Job` when referring to a Kubernetes workload type whose official name contains `Job`, such as `Job`, `RayJob`, or `PyTorchJob`.

Preserve `job` in literal identifiers, API fields, commands, output, and other interfaces that use that term.

## Organization and Scope

| Use | Use as | Meaning | Avoid for this meaning |
|---|---|---|---|
| `account` | Product concept | The top-level NVIDIA Run:ai organizational scope. | tenant |
| `department` | Product concept | An organizational and resource-management scope that groups projects. | project |
| `Kubernetes namespace` | Kubernetes concept | A Kubernetes namespace. | project |
| `project` | Product concept | An organizational and resource-management scope under which workloads run. | namespace |
| `scope` | Product concept | The part of the NVIDIA Run:ai hierarchy in which an object or access assignment applies. | — |

Preserve `tenant` in literal API fields and other interfaces that use that term.

Do not use `project` and `Kubernetes namespace` interchangeably. A project is an NVIDIA Run:ai concept even when it is associated with a Kubernetes namespace.

## Clusters and Nodes

| Use | Use as | Meaning | Avoid for this meaning |
|---|---|---|---|
| `cluster` | Product and Kubernetes concept | A Kubernetes cluster connected to NVIDIA Run:ai. | — |
| `node` | Kubernetes concept | A Kubernetes node that provides resources to workloads or platform services. | — |
| `node pool` | Product concept | A group of nodes treated as a resource pool for resource allocation and scheduling. | node type |
| `node type` | Scheduling concept | A label-based node classification used to constrain workloads to matching nodes. | node pool |
| `system node` | Infrastructure concept | A node designated to run NVIDIA Run:ai system services. | — |

Do not use `node pool` and `node type` interchangeably. A node type can identify nodes within a node pool.

## Quota and Resource Allocation

| Use | Use as | Meaning | Avoid for this meaning |
|---|---|---|---|
| `deserved quota` | Scheduling concept | Resources that a project or department is entitled to use according to its configured quota. | guaranteed quota, guaranteed resources |
| `fairshare` | Scheduling concept | The calculated resource entitlement of a scheduling queue. | fair share |
| `in quota` | Scheduling state | Resource use within deserved quota. | under quota |
| `in-quota` | Scheduling adjective | Describes resource use within deserved quota. | under-quota |
| `max GPU device allocation` | Resource-allocation property | The maximum number of GPU devices a project or department can allocate from a node pool. | GPU limit when the named setting is intended |
| `over quota` | Scheduling state | Resource use beyond deserved quota. | above quota |
| `over-fairshare` | Scheduling adjective | Describes a queue whose allocation exceeds its calculated fairshare. | above fairshare |
| `over-quota` | Scheduling adjective | Describes resource use beyond deserved quota. | above-quota |
| `over-quota resources` | Scheduling concept | Resources consumed beyond deserved quota. | — |
| `over-quota weight` | Scheduling property | A value used to calculate a queue's relative share of available over-quota resources. | priority |
| `over-subscription` | Resource-management concept | A state in which assigned deserved quota exceeds available physical resources. | oversubscription |
| `priority` | Workload property | Ordering used when scheduling workloads within a project. | rank |
| `quota` | Resource-management concept | A configured resource allocation for a project or department. | guaranteed quota |
| `rank` | Scheduling property | Ordering used when allocating resources among project or department queues. | priority |
| `scheduling queue` | Scheduling concept | The Scheduler representation of a project or department for resource allocation within a node pool. | job queue |
| `time-based fairshare` | Scheduling concept | Fairshare that accounts for historical resource consumption. | time-based fair share |

Do not use `rank` and `priority` interchangeably. Rank applies to resource allocation among project or department queues. Priority applies to workload scheduling within a project.

## Preemption and Resource Reclaim

| Use | Use as | Meaning | Avoid for this meaning |
|---|---|---|---|
| `preemption` | Scheduling concept | Interruption of a preemptible workload so its resources can be reallocated. | — |
| `reclaim` | Verb | Reallocate resources between queues to restore deserved quota or fairness. | preempt when the cross-queue action is intended |
| `resource reclaim` | Scheduling concept | Reallocation of resources between project or department queues to restore deserved quota or fairness. | priority preemption |

Do not describe cross-project or cross-department resource reclaim as priority-based preemption. Priority applies to workload scheduling within a project.

## GPU Resource Optimization

| Use | Use as | Meaning | Avoid for this meaning |
|---|---|---|---|
| `dynamic GPU fractions` | Resource-optimization concept | GPU fractions that can dynamically use additional available GPU memory up to a configured limit. | — |
| `GPU fraction` | Resource-optimization concept | A portion of a physical GPU allocated to a workload. | fractional GPU |
| `GPU fractions` | Feature or plural concept | NVIDIA Run:ai functionality for allocating portions of GPUs to workloads. | fractional GPUs |
| `GPU memory limit` | Resource property | The maximum GPU memory available to a workload using dynamic GPU fractions. | GPU memory request |
| `GPU memory request` | Resource property | The amount of GPU memory requested for a workload. | GPU memory limit |
| `GPU compute mode` | Resource-optimization concept | How workloads sharing a GPU divide the GPU compute. Either time-slicing or SM-sharing. | GPU memory fraction |
| `GPU memory fraction` | Resource-optimization concept | A portion of the GPU memory of a physical GPU allocated to a workload. | fractional GPU |
| `SM-sharing` | Resource-optimization concept | A GPU compute mode in which workloads run concurrently on separate streaming multiprocessors. | MPS |
| `time-slicing` | Resource-optimization concept | A GPU compute mode in which workloads take sequential turns on the GPU compute engines. | GPU fractions when compute sharing is intended |
| `multi-GPU fractions` | Resource-optimization concept | Fractional allocations across multiple GPU devices. | multi GPU fractions |
| `single-GPU fraction` | Resource-optimization concept | A fractional allocation from one GPU device. | single GPU fraction |

Do not use `GPU memory fractions`, `dynamic GPU memory fractions`, and `GPU compute modes` interchangeably. Memory fractions and compute modes are separate settings on a workload.

## Policies and Scheduling Rules

| Use | Use as | Meaning | Avoid for this meaning |
|---|---|---|---|
| `default` | Workload-policy concept | A value supplied by a workload policy when the workload does not specify one. | rule |
| `idle GPU time limit` | Scheduling rule | A scheduling rule that limits allowed GPU inactivity. | — |
| `imposed asset` | Workload-policy concept | A workload asset that a workload policy requires a workload to use. | default asset |
| `rule` | Workload-policy concept | A constraint applied by a workload policy. | default |
| `scheduling rule` | Product concept | A restriction on workload runtime or eligible nodes. | workload policy |
| `workload duration` | Scheduling rule | A scheduling rule that limits how long a workload can run. | — |
| `workload policy` | Product concept | A policy that controls workload submission through rules, defaults, or imposed assets. | scheduling rule |

Do not use `workload policy` and `scheduling rule` interchangeably.

## Workload Assets

| Use | Use as | Meaning | Avoid for this meaning |
|---|---|---|---|
| `compute resource` | Workload asset | A reusable specification of workload compute requirements. | — |
| `ConfigMap` | Kubernetes object | A Kubernetes ConfigMap. | data source |
| `container image` | Technical concept | An image used to create a workload container. | environment |
| `credential` | Workload asset | A reusable object containing authentication information required by a workload. | Kubernetes Secret |
| `data source` | Workload asset | A reusable definition for making data or storage available to a workload. | data volume |
| `data volume` | Workload storage concept | Storage configured for use by a workload. | data source |
| `environment` | Workload asset | A reusable configuration for the environment in which a workload runs. | container image |
| `PersistentVolumeClaim` (`PVC`) | Kubernetes object | A Kubernetes request for persistent storage. | data source |
| `Secret` | Kubernetes object | A Kubernetes Secret. | credential |
| `workload asset` | Product concept | A reusable NVIDIA Run:ai object used to configure workloads. | — |

Do not use an underlying Kubernetes or storage object as a synonym for an NVIDIA Run:ai workload asset.

## Workload Templates

| Use | Use as | Meaning | Avoid for this meaning |
|---|---|---|---|
| `one-time configuration` | Product concept | Configuration defined for a specific workload rather than saved as a reusable asset. | workload asset |
| `training template` | Template type | A workload template for training workloads. | — |
| `workload template` | Product concept | A reusable configuration for creating workloads. | workload asset |
| `workspace template` | Template type | A workload template for workspaces. | — |

Do not use `workload template` and `workload asset` interchangeably.

## Authentication and Authorization

| Use | Use as | Meaning | Avoid for this meaning |
|---|---|---|---|
| `access rule` | Authorization concept | Assignment of a role to a subject in a scope. | role |
| `authentication` | Security concept | Verification of an identity. | authorization |
| `authorization` | Security concept | Determination of what an authenticated subject can do. | authentication |
| `group` | Identity concept | A collection of users used for access management. | — |
| `identity provider` | Security concept | A system that authenticates identities used with NVIDIA Run:ai. Use `IdP` after defining it when appropriate. | — |
| `permission` | Authorization concept | An allowed action on an NVIDIA Run:ai entity. | role |
| `permission set` | Authorization concept | A reusable group of permissions. | role |
| `role` | Authorization concept | A defined set of permissions that can be assigned to a subject. | permission, access rule |
| `service account` | Identity concept | A non-human identity used for programmatic access. | application |
| `subject` | Authorization concept | A user, group, or service account to which access can be granted. | user when other subject types are possible |
| `user` | Identity concept | A human identity. | account |

`application` is legacy terminology for a service account. Preserve it when referring to the legacy concept or when it appears in literal interfaces.

## Maintain the List

The list has no fixed entry count. Expand it when repository usage proves that a term needs control, and remove an entry when the project no longer uses its concept or exact form. The section boundaries help writers find terms; they do not create separate vocabularies.

Add or change an entry only when it does at least one of these things:

- Resolves a recurring ambiguity across repository surfaces.
- Separates concepts whose difference affects behavior, security, support, or evidence.
- Preserves an exact product name or text form.

For each entry:

1. Confirm the term against the applicable CLI, code, tests, and current documentation. Confirm an external product name against its official form.
2. Select the shortest familiar term that preserves the current meaning.
3. Define one meaning. Add another row when the same word has a different controlled meaning.
4. List only alternatives that writers use for that same meaning.
5. Keep entries alphabetical within their section.

Do not add every acceptable English word. Do not use this list to rename a command, identifier, schema field, UI label, or third-party product. Such a rename needs its own behavior or interface decision.
