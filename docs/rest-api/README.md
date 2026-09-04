---
hidden: true
---

# Self-Hosted API and CLI Reference

NVIDIA Run:ai is a platform for accelerating and managing AI workloads in shared computing environments. It enables teams to dynamically schedule, allocate, and monitor GPU resources across clusters.

The NVIDIA Run:ai control plane exposes a **REST API** for programmatic access and a **CLI** for interactive and scripted use, covering workload submission, resource management, and administrative operations.

## REST API

<Cards>
<Card title="About the REST API" icon="gear-code" href="about-the-rest-api.md">
Overview of the NVIDIA Run:ai REST API, structure, and core concepts
</Card>
<Card title="Authenticate to the API" icon="code-pull-request" href="how-to-authenticate-to-the-api.md">
Create and use access tokens to authorize API requests
</Card>
<Card title="Using the REST API" icon="square-terminal" href="making-rest-api-requests.md">
Make your first requests and learn common request, response, and pagination patterns
</Card>
</Cards>

## Command Line Interface (CLI)

<Cards>
<Card title="Install and Configure the CLI" icon="square-terminal" href="../reference/cli/install-cli.md">
Install the NVIDIA Run:ai CLI and connect it to your cluster
</Card>
<Card title="CLI Command Reference" icon="rectangle-terminal" href="../reference/cli/runai/runai.md">
Browse the full list of `runai` commands, flags, and usage examples
</Card>
</Cards>

## Common Use Cases

Developers and administrators use the NVIDIA Run:ai API and CLI to:

* **Automate workload management** - Submit, suspend/resume, and monitor AI training and inference jobs programmatically or from scripts.
* **Optimize resource utilization** - Dynamically allocate GPU and CPU resources based on workload demand.
* **Integrate with existing systems** - Connect NVIDIA Run:ai with CI/CD pipelines, data science platforms, or custom dashboards via the API.
* **Manage multi-tenant environments** - Streamline access control, quotas, and user management across teams or projects.

Together, the API and CLI give you fine-grained control and scalability beyond the user interface, making it easier to embed NVIDIA Run:ai capabilities into your organization's AI infrastructure.
