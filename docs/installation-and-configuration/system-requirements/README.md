---
---

# System Requirements

This section describes the infrastructure you must provide before installing NVIDIA Run:ai - hardware, Kubernetes distribution, networking, and storage - whether you use the [guided installer](../guided-installation/README.md) or [manual Helm installation](../install-using-helm/install-control-plane.md). It follows the [shared responsibility model](../../overview.md#shared-responsibility): NVIDIA provides the Run:ai software, while you prepare and manage the Kubernetes environment it runs on.

Before reviewing the requirements, determine your [deployment model](../../overview.md#deployment-model), as it affects which infrastructure components must be provisioned once and which are required for each NVIDIA Run:ai cluster.

| Section | What it covers |
|---|---|
| [Deployment Model](deployment-model.md) | Same-cluster vs multi-cluster connectivity, networking, DNS, and TLS requirements |
| [Infrastructure Requirements](infrastructure-requirements.md) | Hardware, OS, Kubernetes distribution, storage, and permissions |
| [Networking Setup](networking-setup.md) | Load balancer, ingress controller, DNS, FQDN, and TLS certificates |
| [External Access to Critical Services](external-access-to-critical-services.md) | Inbound/outbound firewall rules and registry domain allowlist |
| [Third-Party Services](third-party-services.md) | Services deployed as part of installation, their registries, and how to replace them with externally managed instances |

## Next Steps

Once your environment is prepared, proceed to installation:

* [Guided Installation](../guided-installation/README.md) - Recommended for most installations.
* [Manual Helm](../install-using-helm/install-control-plane.md) - For organizations managing releases through their own tooling.
