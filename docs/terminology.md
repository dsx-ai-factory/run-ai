# Terminology

Reference for correct spelling, casing, and usage of NVIDIA Run:ai terms. When in doubt, check this page before writing.

## Product and Editions

| Term | Correct form | Notes |
|---|---|---|
| Product name | **NVIDIA Run:ai** | Always use the full name in prose. Never shorten to "Run:ai" in body text. Colon is part of the name. |
| In code | `runai` | CLI commands, API flags, code blocks, and inline code only |
| Commercial edition | **NVIDIA Run:ai Enterprise** | Full name on first mention per page |
| Open source edition | **NVIDIA Run:ai OSS** | Spell out "Open Source Software" on first mention |
| Deployment type | **Self-Hosted** | Capitalize; hyphenated |
| Deployment type | **SaaS** | All caps |
| Deployment type | **Multi-Tenant** | Capitalize; hyphenated |

## Platform Components

| Term | Correct form | Notes |
|---|---|---|
| Scheduler | **NVIDIA Run:ai Scheduler** | Full name on first mention per page; "the Scheduler" after |
| KAI Scheduler | **KAI Scheduler** | The underlying scheduling engine (Kubernetes AI Scheduler). Capitalize both words. Distinct from "NVIDIA Run:ai Scheduler" (the brand name). |
| Control plane | **NVIDIA Run:ai control plane** | Lowercase "control plane" |
| Cluster component | **NVIDIA Run:ai cluster** | Lowercase "cluster" |

## Organizational Hierarchy

| Term | Correct form | Notes |
|---|---|---|
| Cluster | **Cluster** | Capitalize when referring to the Run:ai construct; lowercase in generic Kubernetes context |
| Department | **Department** | Capitalize when referring to the Run:ai construct |
| Project | **Project** | Capitalize when referring to the Run:ai construct |
| Scope | **Scope** | Capitalize when referring to the Run:ai access scope concept |
| Tenant | **Tenant** | Capitalize when referring to the SaaS or Multi-Tenant construct |
| Namespace | `namespace` | Lowercase; Kubernetes concept |
| AI initiative | **AI initiative** | Lowercase "initiative." The conceptual framing around which Projects and Departments are organized. |
| Organizational tree | **organizational tree** | Lowercase. The logical hierarchy of accounts, clusters, departments, and projects. |
| Access rule | **access rule** | Lowercase. The RBAC assignment of a subject to a scope and role. |
| Subject | **subject** | Lowercase. In RBAC context: a user, SSO group, or service account. |
| Project rank | **project rank** | Lowercase. Official term replacing "project priority." The scheduling precedence of a Project relative to others. |
| Department rank | **department rank** | Lowercase. Official term replacing "department priority." |

## Workload Types

| Term | Correct form | Notes |
|---|---|---|
| Workspace | **Workspace** | Capitalize; NVIDIA Run:ai native workload type |
| Training | **Training** | Capitalize when referring to the Run:ai workload type generically |
| Standard training | **Standard training** | Capitalize "Standard," lowercase "training." The non-distributed Training workload type; used when distinguishing from Distributed Training. |
| Distributed training | **Distributed Training** | Capitalize both words |
| Distributed inference | **Distributed Inference** | Capitalize both words; matches Distributed Training pattern |
| Inference | **Inference** | Capitalize when referring to the Run:ai workload type |
| AI application | **AI application** | Lowercase "application" |
| Workload category | **workload category** | Lowercase. Three named values: **Build** (default for Workspaces), **Train** (default for Training), **Deploy** (default for Inference). Capitalize the category names. |
| Karta | **Karta** | Always capitalize. The platform's workload extensibility framework — a declarative contract for CRD-based workload types. |
| Supported workload type | **supported workload type** | Lowercase. Non-native workloads registered via the Workload Types API (for example, NVIDIA Dynamo via Grove). |
| Grove | **Grove** | Capitalize. The Kubernetes CRD and controller for NVIDIA Dynamo workloads. |
| NVIDIA Dynamo | **NVIDIA Dynamo** | Full name on first mention; "Dynamo" after. Inference framework. |
| NVCF | **NVCF** | Spell out "NVIDIA Cloud Functions (NVCF)" on first mention. |
| LeaderWorkerSet | **LeaderWorkerSet (LWS)** | Spell out on first mention. A Kubernetes workload type for distributed inference. |
| vLLM | **vLLM** | Mixed case exactly as shown. Inference serving framework. |

## Resources and Quotas

| Term | Correct form | Notes |
|---|---|---|
| Graphics processing unit | **GPU** | Always uppercase |
| Central processing unit | **CPU** | Always uppercase |
| Node pool | **Node Pool** | Capitalize when referring to the Run:ai construct |
| Default node pool | **default node pool** | Lowercase. Every cluster has one node pool named "default" created automatically. |
| GPU quota | **GPU quota** | Lowercase "quota" |
| Deserved quota | **deserved quota** | Lowercase. The guaranteed resource allocation a Project is entitled to. |
| Over-quota | **over-quota** (adjective); **over quota** (noun) | Hyphenate when used before a noun ("over-quota resources"); two words when used as a noun ("running in over quota"). |
| Over-quota weight | **over-quota weight** | Lowercase, uppercase GPU. Controls how over-quota resources are distributed among competing Projects. |
| GPU fraction | **GPU fraction** | Lowercase "fraction" |
| Dynamic GPU fractions | **dynamic GPU fractions** | Lowercase except GPU |
| Multi-GPU fractions | **multi-GPU fractions** | Lowercase "multi," uppercase GPU, lowercase "fractions." |
| GPU time-slicing | **GPU time-slicing** | Hyphenated, uppercase GPU. NVIDIA-native mechanism for sharing GPU compute time. |
| GPU memory swap | **GPU memory swap** | Lowercase "memory swap." Extends GPU physical memory using CPU memory. |
| Max GPU device allocation | **max GPU device allocation** | Lowercase except GPU. The maximum number of GPU devices a Project or Department can allocate. |
| Multi-Instance GPU | **MIG** | Spell out "Multi-Instance GPU (MIG)" on first mention |
| Dynamic Resource Allocation | **Dynamic Resource Allocation (DRA)** | Spell out on first mention, then "DRA." Kubernetes-native mechanism for flexible hardware requests. |
| Persistent volume claim | **PVC** | Spell out on first mention |
| Network file system | **NFS** | Spell out on first mention |

## Scheduler Concepts

| Term | Correct form | Notes |
|---|---|---|
| Fairshare | **fairshare** | One word, all lowercase in prose. The scheduling principle of distributing resources proportionally. Not "fair share" (two words). |
| Fairshare balancing | **fairshare balancing** | Lowercase. The ongoing process of redistributing resources to maintain fairshare across Projects. |
| Point-in-time fairshare | **point-in-time fairshare** | Lowercase. The default fairshare calculation mode. |
| Time-based fairshare | **time-based fairshare** | Lowercase. Fairshare mode that accounts for historical resource usage. |
| Gang scheduling | **gang scheduling** | Lowercase. Scheduling all pods of a workload simultaneously. |
| Multi-level gang scheduling | **multi-level gang scheduling** | Lowercase. |
| Pod group | **pod group** | Two words, lowercase. A set of pods that must be scheduled together. |
| Sub-group | **sub-group** | Hyphenated, lowercase. A subset of a pod group. |
| Consolidation | **consolidation** | Lowercase. The scheduler action of rearranging workloads to free resources before preemption. |
| Preemption | **preemption** | Lowercase. The act of stopping a lower-priority workload to free resources for a higher-priority one. |
| Reclaim | **reclaim** | Lowercase. Inter-project or inter-department preemption to reclaim deserved quota. |
| Preemptible | **preemptible** | Lowercase adjective. Workload property indicating it can be preempted. |
| Non-preemptible | **non-preemptible** | Hyphenated, lowercase adjective. |
| Workload priority | **workload priority** | Lowercase. The scheduling precedence of a workload within a Project. |
| Priority class | **priority class** / `PriorityClass` | Lowercase in prose; `PriorityClass` (CamelCase in code) when referring to the Kubernetes object. |
| Bin-pack | **bin-pack** | Hyphenated. The placement strategy that concentrates workloads on as few nodes as possible. Use "bin packing" only as a gerund ("the scheduler uses bin packing"). |
| Spread | **spread** | Lowercase. The placement strategy that distributes workloads across nodes. Opposite of bin-pack. |
| Scheduling queue | **scheduling queue** | Lowercase. The internal queue (one per Project/Node Pool pair and Department/Node Pool pair). |
| Over-subscription | **over-subscription** | Hyphenated, lowercase. When the sum of all deserved quotas exceeds available physical resources. |
| NUMA-aware scheduling | **NUMA-aware scheduling** | Uppercase NUMA (acronym), lowercase rest. |
| Topology-aware scheduling | **topology-aware scheduling** | Lowercase. Network-topology-based placement optimization. |
| Network topology | **network topology** | Lowercase. Platform construct associated with Node Pools for accelerating distributed workloads. |

## Policy and Governance

| Term | Correct form | Notes |
|---|---|---|
| Workload policy | **workload policy** | Lowercase. The platform construct for controlling workload submission across scopes. |
| Scheduling rules | **scheduling rules** | Lowercase. Policy restrictions on workload resources or runtime duration. |
| Effective policy | **effective policy** | Lowercase. The reconciled result of all policies from system through cluster, department, and project scopes. |
| Imposed assets | **imposed assets** | Lowercase. Policy construct for assets automatically attached to any workload in scope. |
| System policy | **system policy** | Lowercase. Platform-level baseline policies. |
| Grace period | **grace period** | Lowercase. Duration after preemption request before forced termination (default 30s; system policy caps at 5 minutes). |
| Idle GPU timeout | **idle GPU timeout** | Lowercase, uppercase GPU. Scheduling rule that stops workloads with an idle GPU after a set duration. |

## Infrastructure

| Term | Correct form | Notes |
|---|---|---|
| NVLink | **NVLink** | Mixed case exactly. NVIDIA's GPU interconnect technology. |
| MNNVL | **MNNVL** | All caps. Spell out "Multi-Node NVLink (MNNVL)" on first mention. |

## Access and Authentication

| Term | Correct form | Notes |
|---|---|---|
| Single sign-on | **SSO** | Spell out on first mention |
| OpenID Connect | **OIDC** | Spell out on first mention |
| Role-based access control | **RBAC** | Spell out on first mention |
| Access key | **access key** | Lowercase |
| Client secret | **client secret** | Lowercase |
| SSO group | **SSO group** | Lowercase "group" |
| Service account | **service account** | Lowercase |

## Third-Party Tools and Ecosystem

| Term | Correct form | Notes |
|---|---|---|
| Kubernetes | **Kubernetes** | Never "k8s" in prose; use `k8s` only in code |
| Helm | **Helm** | Capitalize |
| kubectl | `kubectl` | Lowercase; always in code style |
| NVIDIA GPU Operator | **NVIDIA GPU Operator** | Full name; capitalize "Operator" |
| NVIDIA Network Operator | **NVIDIA Network Operator** | Full name |
| NVIDIA NIM | **NVIDIA NIM** | Full name on first mention; "NIM" after |
| NGC catalog | **NGC catalog** | Uppercase "NGC"; lowercase "catalog" |
| NGC API key | **NGC API key** | Uppercase "NGC" and "API"; lowercase "key" |
| NIM Operator | **NIM Operator** | Capitalize "Operator" |

## CLI and API

| Term | Correct form | Notes |
|---|---|---|
| CLI version | **CLI v2** | Not "CLI v2.0", "new CLI", or "the CLI" without qualifier on first mention |
| API reference | **REST API Reference** | Capitalize all words in the nav label |
| API endpoint | lowercase in prose | e.g. "the workloads endpoint"; capitalize only in nav/UI labels |

## UI Labels and Navigation

Use bold for UI element names in prose. Capitalize exactly as shown in the product.

| UI element | Correct form |
|---|---|
| Top-level nav item | **Workload manager** |
| Sub-item | **Workloads** |
| Settings area | **General settings** |
| Button (action) | **+NEW WORKLOAD**, **SAVE & CONTINUE**, **DELETE** |
| Form field | Match exact label from UI, in bold |

Navigation paths use → as the separator: go to **Workload manager** → **Workloads**.
