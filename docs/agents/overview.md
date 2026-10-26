# AI Agents

The NVIDIA Run:ai MCP server connects AI coding assistants directly to your NVIDIA Run:ai cluster through the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) open standard. This gives AI agents and assistants the ability to inspect workloads, query resource utilization, and diagnose scheduling issues, all through natural language, without switching context to the NVIDIA Run:ai UI or API.

## Use Cases

* **Workload debugging** - When an AI practitioner's training or inference job is stuck, queued, or failing, they can ask why, surfacing the workload's current status and recent events, drill into pod placement and resource allocation, inspect the submitted spec, or trace what happened over time with full revision history. For inference workloads, also check request throughput, latency trends, and revision history.
* **Capacity planning** - Platform admins can see how GPU quota is assigned versus actually used across projects, departments, and node pools, spot teams holding idle capacity, and track utilization trends over time.
* **Cluster state monitoring** - When workloads fail across multiple users, system admins can identify nodes that are NotReady or unreachable, inspect their Kubernetes conditions and taints, and see GPU counts per node pool at a point in time.
* **Cluster configuration** - Platform admins can create or update node pools, adjust scheduling and GPU optimization settings, and toggle tenant-wide feature flags through natural language without switching to the NVIDIA Run:ai UI.
* **Organization management** - Platform admins can create, update, and manage all projects and departments across the cluster, including quota assignments and organizational hierarchy, through natural language without switching to the NVIDIA Run:ai UI.
