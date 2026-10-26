---
hidden: true
---

# Multi-Tenant API Reference

NVIDIA Run:ai multi-tenant lets a host organization operate a secure, scalable multi-tenant control plane to deliver AI Platform-as-a-Service (PaaS) to multiple tenant organizations.

The NVIDIA Run:ai control plane exposes a **REST API** for programmatic access, covering tenant management, cluster registration, and platform-level administrative operations.

## REST API

<Cards>
<Card title="About the REST API" icon="gear-code" href="about-the-rest-api.md">
Overview of the NVIDIA Run:ai Multi-Tenant REST API, structure, and core concepts
</Card>
<Card title="API Access Setup" icon="code-pull-request" href="api-access-setup.md">
Create and use access tokens to authorize API requests
</Card>
<Card title="Using the REST API" icon="square-terminal" href="making-rest-api-requests.md">
Make your first requests and learn common request, response, and pagination patterns
</Card>
</Cards>

## Common Use Cases

Host organization administrators use the NVIDIA Run:ai Multi-Tenant API to:

* **Manage tenants** - Create, update, and offboard tenants and their configuration programmatically or from scripts.
* **Register and manage clusters** - Provision and register each tenant's Kubernetes cluster with the control plane.
* **Configure platform-level settings** - Manage authentication, access control, and other control-plane-wide settings across tenants.
* **Monitor and audit operations** - Track multi-tenant operational activity for governance and compliance.

Together, the API gives you fine-grained control and scalability beyond the user interface, making it easier to embed NVIDIA Run:ai multi-tenant capabilities into your organization's tenant onboarding and operations workflows.
