# NVIDIA Run:ai MCP Server

The NVIDIA Run:ai MCP server connects AI coding assistants directly to your NVIDIA Run:ai cluster through the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) open standard. This gives AI agents and assistants the ability to inspect workloads, interact with cluster resources, and diagnose scheduling issues, all through natural language, without switching context to the NVIDIA Run:ai UI or API.

### Use Cases

* **Workload debugging** - When an AI practitioner's training or inference job is stuck, queued, or failing, they can ask why - surfacing the workload's current status and recent events, drill into pod placement and resource allocation, inspect the submitted spec, or trace what happened over time with full revision history. For inference workloads, also check request throughput, latency trends, and revision history.
* **Capacity planning** - Platform admins can see how GPU quota is assigned versus actually used across projects, departments, and node pools, spot teams holding idle capacity, and track utilization trends over time.
* **Cluster state monitoring** - When workloads fail across multiple users, system admins can identify nodes that are NotReady or unreachable, inspect their Kubernetes conditions and taints, and see GPU counts per node pool at a point in time.
* **Cluster configuration** (write tools) - Platform admins can create or update node pools, adjust scheduling and GPU optimization settings, and toggle tenant-wide feature flags through natural language without switching to the NVIDIA Run:ai UI.
* **Organization management** (write tools) - Platform admins can create, update, and manage all projects and departments across the cluster, including quota assignments and organizational hierarchy, through natural language without switching to the NVIDIA Run:ai UI.

## Requirements

* NVIDIA Run:ai cluster version 2.25 or later.
* The MCP server image is available on the [NGC catalog](https://catalog.ngc.nvidia.com/orgs/nvidia/runai/containers/runai-mcp-server) with guest access — no NGC account is required to pull it.

## Deployment Modes

The MCP server supports two deployment modes:

| Mode | Who sets it up | How the AI tool connects |
|---|---|---|
| **Local (STDIO)** | Each user independently | Docker container on the user's machine |
| **Remote (HTTP)** | Cluster administrator | Centrally deployed server, shared across users |

## Local Mode (STDIO)

In local mode, each user runs the MCP server as a Docker container on their own machine. The container authenticates to the NVIDIA Run:ai control plane using an NVIDIA Run:ai service account.

### Prerequisites

* Docker installed on your machine.
* An NVIDIA Run:ai service account with the appropriate permissions. See [Service Accounts](../infrastructure-setup/authentication/service-accounts.md) for instructions.

### Credential Security

Store your service account credentials in environment variables rather than hardcoding them in config files. The configuration examples below pass them via `env` so they are resolved at runtime from your shell environment and not written to disk.

### Configure Your AI Tool

Replace the following placeholders before applying the configuration:

* `<your-runai-url>` with your NVIDIA Run:ai control plane URL
* `<client-id>` and `<client-secret>` with your service account credentials

<Tabs>

<Tab title="Claude Code">

Add the following to `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "runai": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "-e", "RUNAI_BASE_URL",
        "-e", "RUNAI_CLIENT_ID",
        "-e", "RUNAI_CLIENT_SECRET",
        "nvcr.io/nvidia/runai/runai-mcp-server"
      ],
      "env": {
        "RUNAI_BASE_URL": "https://<your-runai-url>",
        "RUNAI_CLIENT_ID": "<client-id>",
        "RUNAI_CLIENT_SECRET": "<client-secret>"
      }
    }
  }
}
```

</Tab>
<Tab title="Cursor">

Add the following to `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "runai": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "-e", "RUNAI_BASE_URL",
        "-e", "RUNAI_CLIENT_ID",
        "-e", "RUNAI_CLIENT_SECRET",
        "nvcr.io/nvidia/runai/runai-mcp-server"
      ],
      "env": {
        "RUNAI_BASE_URL": "https://<your-runai-url>",
        "RUNAI_CLIENT_ID": "<client-id>",
        "RUNAI_CLIENT_SECRET": "<client-secret>"
      }
    }
  }
}
```

</Tab>
<Tab title="VS Code / GitHub Copilot">

Add the following to `.vscode/mcp.json`:

```json
{
  "servers": {
    "runai": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "-e", "RUNAI_BASE_URL",
        "-e", "RUNAI_CLIENT_ID",
        "-e", "RUNAI_CLIENT_SECRET",
        "nvcr.io/nvidia/runai/runai-mcp-server"
      ],
      "env": {
        "RUNAI_BASE_URL": "https://<your-runai-url>",
        "RUNAI_CLIENT_ID": "<client-id>",
        "RUNAI_CLIENT_SECRET": "<client-secret>"
      }
    }
  }
}
```

</Tab>
<Tab title="Codex">

Add the following to `~/.codex/config.toml`:

```toml
[mcp_servers.runai]
command = "docker"
args = [
  "run", "--rm", "-i",
  "-e", "RUNAI_BASE_URL",
  "-e", "RUNAI_CLIENT_ID",
  "-e", "RUNAI_CLIENT_SECRET",
  "nvcr.io/nvidia/runai/runai-mcp-server"
]
env = { RUNAI_BASE_URL = "https://<your-runai-url>", RUNAI_CLIENT_ID = "<client-id>", RUNAI_CLIENT_SECRET = "<client-secret>" }
```

</Tab>

</Tabs>

## Remote Mode (HTTP)

In remote mode, a cluster administrator deploys the MCP server centrally and exposes it at an FQDN. Users connect their AI tools to the shared server using OAuth, without managing service account credentials themselves.

### Network Requirements

The following network paths must be open:

* User's AI agent → MCP server
* MCP server → NVIDIA Run:ai control plane

### Deploy the MCP Server (Administrators)

Run the MCP server container with HTTP transport enabled:

```bash
docker run -d -p 8080:8080 \
  -e RUNAI_BASE_URL=https://<your-runai-url> \
  -e RUNAI_TRANSPORT=http \
  nvcr.io/nvidia/runai/runai-mcp-server
```

Expose the container at an FQDN accessible to users on your network. The MCP endpoint is available at `<fqdn>/mcp`.

### Connect Your AI Tool (Users)

Once your administrator provides the server FQDN, configure your AI tool:

<Tabs>

<Tab title="Claude Code">

**Option 1 — CLI:**

```bash
claude mcp add --transport http --client-id runai-mcp runai https://<fqdn>/mcp
```

**Option 2 — Config file:**

Add the following to `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "runai": {
      "type": "http",
      "url": "https://<fqdn>/mcp",
      "oauth": {
        "clientId": "runai-mcp"
      }
    }
  }
}
```

</Tab>
<Tab title="Cursor">

Add the following to `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "runai": {
      "url": "https://<fqdn>/mcp",
      "auth": {
        "CLIENT_ID": "runai-mcp"
      }
    }
  }
}
```

</Tab>
<Tab title="VS Code / GitHub Copilot">

Add the following to `.vscode/mcp.json`:

```json
{
  "servers": {
    "runai": {
      "type": "http",
      "url": "https://<fqdn>/mcp",
      "oauth": {
        "clientId": "runai-mcp"
      }
    }
  }
}
```

</Tab>

</Tabs>

## Configuration Reference

The following environment variables apply to both deployment modes unless noted otherwise:

| Variable | Description | Applies to | Default |
|---|---|---|---|
| `RUNAI_BASE_URL` | NVIDIA Run:ai control plane URL | Both | Required |
| `RUNAI_CLIENT_ID` | Service account client ID | STDIO only | Required |
| `RUNAI_CLIENT_SECRET` | Service account client secret | STDIO only | Required |
| `RUNAI_TRANSPORT` | Transport mode: `stdio` or `http` | Both | `stdio` |
| `RUNAI_MCP_LISTEN_PORT` | Listen port | HTTP only | `8080` |
| `RUNAI_ALLOW_WRITE_TOOLS` | Expose write (mutating) tools | Both | `false` |

## Available Tools

The MCP server exposes read and write tools, organized by domain. Write tools are disabled by default — set `RUNAI_ALLOW_WRITE_TOOLS=true` to enable them. Each write tool supports a `dryRun` parameter to preview the change before applying it, and requires explicit confirmation before making any upstream change.

### Identity

| Tool | Type | Description |
|---|---|---|
| `whoami` | Read | Who am I authenticated as? Returns the current NVIDIA Run:ai principal (user or service account) and their permissions. |

### Workload Visibility

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

### Cluster and Infrastructure

| Tool | Type | Description |
|---|---|---|
| `get_cluster_physical_inventory` | Read | What hardware does this cluster have? Returns GPU nodes, GPU models, and total, allocatable, allocated, and free GPUs, CPU cores, and memory. |
| `get_cluster_infrastructure_health` | Read | Are any nodes in a degraded state? Returns a point-in-time snapshot of nodes that are NotReady or Unknown, with their Kubernetes conditions, taints, and GPU counts, aggregated per cluster and per node pool. |
| `get_cluster_metrics` | Read | How has this cluster's GPU and CPU capacity and utilization changed over time? |
| `list_node_pools` | Read | What node pools exist in this cluster? Returns each pool's name, default status, lifecycle phase, and network topology. Does not include quota, metrics, or GPU capacity. |
| `manage_node_pool` | Write | Create, update, or delete a node pool in a cluster. Supports scheduling configuration, GPU resource optimization, and network topology binding. |
| `list_system_settings` | Read | List tenant-wide platform feature flags across workloads, authorization, cluster, organization, and other categories, with their current values. Optionally filter by category. Does not cover security or UI settings. |
| `update_system_settings` | Write | Toggle a tenant-wide platform feature flag on or off by specifying a settings category and key. Covers boolean feature flags across workloads, authorization, and cluster settings. |

### Resources and Organization

| Tool | Type | Description |
|---|---|---|
| `list_projects` | Read | List NVIDIA Run:ai project entities — by name, count, or parent department. Includes projects with no resource quota configured. For quota and allocation numbers, use `list_project_resources`. |
| `list_departments` | Read | List NVIDIA Run:ai department entities — by name or count. Includes departments with no resource quota configured. Each entity carries allocated and quota values for point-in-time over-quota checks. For per-node-pool quota, use `list_department_resources`. |
| `list_project_resources` | Read | What GPU, CPU, and memory quotas are configured for a project, and how much is currently allocated? Broken down per node pool. |
| `list_department_resources` | Read | What GPU, CPU, and memory quota is configured for a department on each node pool? Returns the department's own configured quota per node pool. Does not include child-project quotas or live allocation. |
| `get_org_unit_metrics` | Read | How has a project's or department's GPU, CPU, and memory allocation and utilization trended over time? |
| `manage_org_unit` | Write | Create, update, or delete a project or department, including quota per node pool and scheduler enforcement. Deleting a project permanently deletes all its workloads. |
