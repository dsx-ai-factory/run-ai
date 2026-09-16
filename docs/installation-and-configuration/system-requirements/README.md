---
---

# System Requirements

This section describes the infrastructure you must provide before installing NVIDIA Run:ai - hardware, Kubernetes distribution, networking, and storage - and the third-party services the installation deploys for you, whether you use the [Guided Installation](../installation/guided-installation/README.md) or [Manual Helm](../installation/install-using-helm/preparations.md) path. It follows the [shared responsibility model](../overview.md#shared-responsibility): NVIDIA provides the Run:ai software, while you prepare and manage the Kubernetes environment it runs on.

<Note>

**Note**

Before reviewing the requirements, determine your [deployment model](deployment-model.md), as it affects which infrastructure components must be provisioned once and which are required for each NVIDIA Run:ai cluster.

</Note>

Complete each of the following before installing NVIDIA Run:ai:

| Section | What to prepare |
|---|---|
| [Deployment Model](deployment-model.md) | Same-cluster vs multi-cluster connectivity, networking, DNS, and TLS requirements |
| [Infrastructure Requirements](infrastructure-requirements.md) | Hardware, OS, Kubernetes distribution, storage, and permissions |
| [Networking Setup](networking-setup.md) | Load balancer, ingress controller, DNS, FQDN, and TLS certificates |
| [Third-Party Services](third-party-services.md) | Services deployed as part of installation, their registries, and how to replace them with externally managed instances |
| [External Access to Critical Services](external-access-to-critical-services.md) | Inbound/outbound firewall rules and registry domain allowlist |

## Next Steps

Once your environment is prepared, proceed to installation:

* [Guided Installation](../installation/guided-installation/README.md) - Recommended for most installations.
* [Manual Helm](../installation/install-using-helm/preparations.md) - For organizations managing releases through their own tooling.
