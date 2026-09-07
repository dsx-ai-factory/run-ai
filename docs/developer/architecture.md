# Architecture and Components

NVIDIA Run:ai is a Kubernetes-native platform for deploying, scheduling, and managing AI workloads across shared GPU infrastructure. It extends Kubernetes with AI-specific capabilities for resource management, scheduling, governance, and workload lifecycle management, while Kubernetes remains the underlying orchestration layer. Run:ai runs on standard Kubernetes clusters backed by NVIDIA GPU infrastructure: GPU nodes, high-speed interconnects (NVLink, InfiniBand, RoCE), and high-performance shared storage.

The platform is organized into three layers that work together: the interfaces used to interact with the platform, a control plane that centrally manages configuration and state, and cluster components that enforce that state on each Kubernetes cluster running workloads.

![NVIDIA Run:ai system overview](../assets/system-overview.png)

## User Interfaces

NVIDIA Run:ai exposes three interfaces to administrators, AI practitioners, and external systems: the web UI, CLI, and REST API. All three provide access to the same underlying platform capabilities: submitting and managing workloads, administering the platform, and integrating with external tools and automation.

- **Web UI** - The primary interface for administrators and AI practitioners. Submit and monitor workloads, manage organizational structure and access, and view usage and performance data.
- **CLI** - Submits and manages workloads and queries cluster state from the command line, for interactive use or scripting.
- **REST API** - The programmatic interface underlying the UI and CLI, used to integrate NVIDIA Run:ai with external tools and automated pipelines.

## Control Plane Services

The control plane is the centralized management layer of NVIDIA Run:ai. It provides the APIs and services that configure, monitor, and manage AI environments across one or more Kubernetes clusters from a single place, and because it's decoupled from any single cluster, one control plane can manage multiple clusters across different locations. Internally, control plane services are independently deployable microservices that communicate over HTTP and a pub/sub event bus. See the [Service Reference](services.md) for the complete list of control-plane services.

- **Identity and Access** - Handles authentication, identity provider integration, and RBAC enforcement across all services. Every request to NVIDIA Run:ai, through the UI, CLI, or API, is authenticated and then authorized before it's allowed to proceed. [identity-manager](../../control-plane/identity-manager) authenticates the requester, resolving external identity providers, local users, and service accounts, and issuing and revoking tokens. [authorization](../../control-plane/authorization) then evaluates that identity against RBAC roles and access rules to resolve the effective permissions for the request.
- **Organizational Structure** - Manages the hierarchy of tenants, departments, and projects that scopes resource quotas and access rules. Compute is organized into a physical layer (nodes and node pools) and an organizational layer (departments and projects) connected through quota. [org-unit-service](../../control-plane/org-unit-service) persists that hierarchy and the quota allocated at each level, and syncs it to every managed cluster so scheduling and quota enforcement always resolve against the same source of truth.
- **Workload Lifecycle** - Handles workload submission, policy enforcement, asset management, and persistent storage. When a workload is submitted, [workloads-manager](../../control-plane/workloads-manager) validates it against effective policy, resolves any assets it references, and generates the Kubernetes manifests that get synced down to the cluster. [assets-service](../../control-plane/assets-service) manages those reusable assets, environments, compute resources, credentials, data sources, so workloads can reference a standardized configuration instead of repeating it.
- **Observability** - Collects and serves GPU and workload metrics, audit logs, and event-driven notifications. [metrics-service](../../control-plane/metrics-service) queries per-tenant, per-cluster Prometheus stores and serves GPU and workload metrics over REST. [audit-service](../../control-plane/audit-service) ingests user and system action events from the pub/sub bus and persists them for query.
- **Platform Delivery** - Serves the web UI, CLI binaries, and the cluster-facing APIs that track node and Kubernetes object state. [frontend](../../control-plane/frontend) is the web UI. [cluster-service](../../control-plane/cluster-service) maintains control-plane state for cluster-scoped resources, currently node inventory, and syncs it from connected clusters.

## Cluster Components

Each Kubernetes cluster connected to NVIDIA Run:ai runs a set of cluster components that extend Kubernetes with AI-specific capabilities, installed and managed by the [go-operator](../../runai-cluster/cluster/go-operator), which owns their full lifecycle: deploy, configure, and update. The cluster plane is CRD-driven: components reconcile control-plane desired state into cluster CRDs such as `RunaiConfig`, `AccessRule`, and `WorkloadRunner`, and report status back through the same CRDs. See the [Service Reference](services.md) and [CRD Reference](crds.md) for the complete list of cluster components and CRDs.

- **Data Sync** - [agent](../../runai-cluster/cluster/agent) pulls desired state for every resource kind, projects, workloads, access rules, node pools, credentials, from the control plane and reconciles each into the cluster. [cluster-sync](../../runai-cluster/cluster/cluster-sync) watches those CRDs and live Kubernetes objects and pushes their state back up to the control plane. This two-channel pattern, desired state down, live state up, is how the two planes stay synchronized, over an outbound-only, secured connection so clusters never need to be publicly reachable.
- **Scheduling** - [admission-controller](../../runai-cluster/cluster/admission-controller) intercepts pod creation and injects Run:ai scheduling attributes, labels, annotations, resource patches, before a pod is admitted. The actual placement decision is made by the Scheduler, built on the external KAI Scheduler (see External Dependencies below), which applies fair-share, preemption, bin-packing, and topology-aware placement rules. Each node pool runs its own Scheduler instance, so scheduling decisions for one hardware group never compete with another.
- **Workload Lifecycle** - [anyworkload-controller](../../runai-cluster/cluster/anyworkload-controller) translates a workload's CRD into its underlying pods and manages their full lifecycle: creation, status, teardown.
- **RBAC and Project Isolation** - [accessrule-controller](../../runai-cluster/cluster/accessrule-controller) reconciles `AccessRule` CRDs into native Kubernetes RoleBindings in the corresponding project namespace. [shared-objects-controller](../../runai-cluster/cluster/shared-objects-controller) replicates shared objects, primarily PVCs from `DataVolume` CRDs, into every project or department namespace that references them.

## Architectural Flow

A request moves through the platform in a loop: down from the control plane to the cluster, through execution, and back up as status.

1. A user or external system submits a request through the web UI, CLI, or REST API.
2. The control plane authenticates the request and updates the desired state it owns: the workload itself, plus any organizational resources, policies, or configuration it depends on.
3. Cluster components sync that desired state down to the relevant Kubernetes cluster and reconcile it into Kubernetes resources.
4. The admission controller and Scheduler validate the request, apply policy, and decide where and how the workload runs.
5. Kubernetes runs the workload on the available compute, GPU, networking, and storage infrastructure.
6. Cluster components watch workload and cluster state and sync status back up to the control plane.
7. The control plane surfaces that updated state through the NVIDIA Run:ai interfaces, closing the loop with centralized visibility across every connected cluster.

## External Dependencies

NVIDIA Run:ai's cluster components build on external open-source projects, pulled in as versioned Helm chart dependencies rather than built from Run:ai source. NVIDIA Run:ai packages, configures, and manages these dependencies as part of the platform, so no separate installation or maintenance is required to use them.

- **[KAI Scheduler](https://github.com/kai-scheduler/KAI-Scheduler)** - An open-source, CNCF-donated AI-aware scheduling engine that applies fair-share, preemption, bin-packing, and multi-node placement rules to GPU workloads. It's the engine underlying the NVIDIA Run:ai Scheduler, integrated with Run:ai-specific configuration. It's now a sub-dependency of KRM's node pool controller. **TBD: check Helm chart** - Confirm whether KAI Scheduler is still also pulled in as a direct chart dependency elsewhere.
- **[Karta](https://github.com/run-ai/karta)** - A CRD-based resource abstraction layer with its own operator, used to represent and reconcile workload resource attachments in the cluster.
- **KAI Resource Management (KRM)** *(rolling out)* - Takes over node pool and project/department-to-namespace reconciliation, migrating them onto `kai.resources` CRDs behind a staged, feature-flagged rollout; legacy CRDs are still supported alongside it this release. Pod group scheduling coordination has not yet moved to KRM and remains in the go-operator's kai-operator.

<Note>
These are open-source projects. A custom build of any dependency can be built and integrated in place of the default version. See [External Dependencies](external-dependencies.md) for source repositories, chart versions, and integration steps.
</Note>
