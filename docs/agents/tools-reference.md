# Tools Reference

The MCP server exposes read and write tools, organized by domain. Write tools are disabled by default, set `RUNAI_ALLOW_WRITE_TOOLS=true` to enable them. Each write tool supports a `dryRun` parameter to preview the change before applying it, and requires explicit confirmation before making any upstream change.

## Identity

| Tool | Type | Description |
|---|---|---|
| `whoami` | Read | Who am I authenticated as? Returns the current NVIDIA Run:ai principal (user or service account) and their permissions. |

## Credentials

| Tool | Type | Description |
|---|---|---|
| `list_credentials` | Read | List safe metadata for shared credentials and the caller's own user credentials, with filtering and pagination. Never returns credential values. |
| `manage_credentials` | Write | Create, update, or delete shared and user credentials (access keys, passwords, Docker registry, generic secrets, NGC API keys). Create accepts secret material but never returns it; update and delete require confirmation. |

## Workloads

| Tool | Type | Description |
|---|---|---|
| `list_workloads` | Read | List workloads with their ID, name, type, phase, project, cluster, node pools, priority, and preemptibility. Supports filtering by name, project, cluster, phase, type, node pool, or recent activity, and pagination. For detailed status, events, specs, pods, metrics, or history, use the relevant workload-specific tool. |
| `get_workload_status` | Read | What is this workload's current state and why? Returns phase, recent events, and the reason for any pending or failed condition. |
| `get_workload_pods` | Read | Which pods does this workload have, on which nodes, and in what state? Includes per-pod resource usage. |
| `get_workload_metrics` | Read | How has this workload's GPU, CPU, and memory usage changed over time? For inference workloads, also returns request throughput and latency. |
| `get_workload_history` | Read | What happened to this workload over time? Returns state transitions, revision history, and associated events. |
| `get_workload_spec` | Read | What did this workload request and where is it placed? Returns the submitted spec (compute resources, environment, volumes, and node pool affinity), the actually allocated resources, actual node placement, and any pending-scheduling messages. |
| `get_workload_effective_policy` | Read | What policy rules and defaults apply to a given workload type in a project? |
| `get_workloads_summary` | Read | How many workloads are running, pending, and failed? Scoped to a cluster, project, department, node pool, or the whole platform. |
| `get_node_pods` | Read | Which pods are running on a specific node and which workloads do they belong to? Returns up to 10 pods, each with its status, requested and allocated resources, and the workload it belongs to, plus the node's pool, status, and GPU type and count. |
| `submit_workload` | Write | Submit a new workload, either a native NVIDIA Run:ai workload (Training, Workspace, Distributed, Inference, or DistributedInference) or, on supported control-plane versions, a raw Kubernetes YAML manifest. A real submission requires confirmation. |
| `manage_workload_operations` | Write | Stop, restart, update, or permanently delete an existing workload. Stop and restart require the workload to support suspension; update applies only to native Inference or DistributedInference workloads. Every operation requires confirmation. |

## Cluster and Infrastructure

| Tool | Type | Description |
|---|---|---|
| `get_cluster_physical_inventory` | Read | What hardware does this cluster have? Returns GPU nodes, GPU models, and total, allocatable, allocated, and free GPUs, CPU cores, and memory. |
| `get_cluster_infrastructure_health` | Read | Are any nodes in a degraded state? Returns a point-in-time snapshot of nodes that are NotReady or Unknown, with their Kubernetes conditions, taints, and GPU counts, aggregated per cluster and per node pool. |
| `get_cluster_metrics` | Read | How has this cluster's GPU and CPU capacity and utilization changed over time? |
| `list_node_pools` | Read | What node pools exist in this cluster? Returns each pool's name, default status, lifecycle phase, and network topology. Does not include quota, metrics, or GPU capacity. |
| `manage_node_pool` | Write | Create, update, or delete a node pool in a cluster. Supports scheduling configuration, GPU resource optimization, and network topology binding. |
| `list_system_settings` | Read | List tenant-wide platform feature flags across workloads, authorization, cluster, organization, and other categories, with their current values. Optionally filter by category. Does not cover security or UI settings. |
| `update_system_settings` | Write | Toggle a tenant-wide platform feature flag on or off by specifying a settings category and key. Covers boolean feature flags across workloads, authorization, and cluster settings. |

## Resources and Organization

| Tool | Type | Description |
|---|---|---|
| `list_projects` | Read | List NVIDIA Run:ai project entities, by name, count, or parent department. Includes projects with no resource quota configured. For quota and allocation numbers, use `list_project_resources`. |
| `list_departments` | Read | List NVIDIA Run:ai department entities, by name or count. Includes departments with no resource quota configured. Each entity carries allocated and quota values for point-in-time over-quota checks. For per-node-pool quota, use `list_department_resources`. |
| `list_project_resources` | Read | What GPU, CPU, and memory quotas are configured for a project, and how much is currently allocated? Broken down per node pool. |
| `list_department_resources` | Read | What GPU, CPU, and memory quota is configured for a department on each node pool? Returns the department's own configured quota per node pool. Does not include child-project quotas or live allocation. |
| `get_org_unit_metrics` | Read | How has a project's or department's GPU, CPU, and memory allocation and utilization trended over time? |
| `manage_org_unit` | Write | Create, update, or delete a project or department, including quota per node pool and scheduler enforcement. Deleting a project permanently deletes all its workloads. |
