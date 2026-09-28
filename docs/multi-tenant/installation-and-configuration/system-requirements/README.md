---
---

# System Requirements

This section describes the infrastructure you must provide before installing NVIDIA Run:ai multi-tenant - hardware, Kubernetes distribution, networking, and storage - and the third-party services the installation deploys for you, whether you use [Guided Installation](../installation/guided-installation/README.md) or [Manual Helm](../installation/install-using-helm/preparations.md). It follows the [shared responsibility model](../overview.md#shared-responsibility): NVIDIA provides the Run:ai software, while the host organization prepares and manages the Kubernetes environment it runs on.

Complete each of the following before installing NVIDIA Run:ai multi-tenant:

| Section | What to prepare |
|---|---|
| [Deployment Model](deployment-model.md) | Control plane and tenant cluster connectivity, networking, DNS, and TLS requirements |
| [Infrastructure Requirements](infrastructure-requirements.md) | Hardware, OS, Kubernetes distribution, storage, and permissions |
| [Networking Setup](networking-setup.md) | Load balancer, ingress controller, DNS, FQDN, and TLS certificates |
| [Third-Party Services](third-party-services.md) | Services deployed as part of installation, their registries, and how to replace them with externally managed instances |
| [Metrics Store Requirements](metrics-store-requirements/README.md) | External metrics backend required to aggregate metrics across tenants |
| [External Access to Critical Services](external-access-to-critical-services.md) | Inbound/outbound firewall rules and registry domain allowlist |

## Next Steps

Once your environment is prepared, proceed to installation:

* **[Guided Installation](../installation/guided-installation/README.md)** - Recommended for most deployments.
* **[Manual Helm](../installation/install-using-helm/install-control-plane.md)** - For organizations managing deployments through their own tooling.
