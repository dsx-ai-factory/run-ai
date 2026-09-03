---
---

# Configuration

Day 1 Configuration covers the infrastructure configuration that makes NVIDIA Run:ai production-ready for your environment, once installation is complete. Which configurations apply depends on your environment and workload needs - not every section here is required for every deployment.

## Cluster and Control Plane Settings

* **[Node Roles](node-roles.md)** - Designates dedicated nodes for NVIDIA Run:ai system services, GPU workloads, and CPU-only workloads, so system services don't compete with GPU workloads for CPU cycles on expensive GPU machines.
* **[Control Plane Configuration](control-plane-config.md)** - Advanced Helm chart values for customizing the control plane deployment beyond its defaults.
* **[Cluster Configuration](cluster-config.md)** - Advanced Helm chart and `runaiconfig` values for customizing cluster behavior - scheduling policies, resource management, and system service tuning.
* **[Cluster Authentication](cluster-authentication.md)** - Configures the Kubernetes API server to authenticate users via the NVIDIA Run:ai identity provider (OIDC), so users can securely submit workloads directly with `kubectl`.

## Networking

* **[Kubernetes Gateway API](networking/kubernetes-gateway-api.md)** - An alternative to Ingress for routing external traffic, offering a more flexible and extensible routing model.
* **[Service Mesh](networking/service-mesh.md)** - Configuration required to run NVIDIA Run:ai alongside a service mesh with sidecar injection.
* **[External Access to Containers](networking/external-access-to-containers.md)** - Lets researchers connect to running containers remotely - for example, to run a Jupyter Notebook or connect an IDE for remote development.

## Security

* **[User Identity in Containers](security/user-identity-in-containers.md)** - Propagates the correct user identity into containers instead of running as root by default, which matters for both functionality (like network file system access) and security.
* **[Security Best Practices](security/security-best-practices.md)** - Actionable, platform-native and Kubernetes-mapped best practices for securely configuring and operating NVIDIA Run:ai.
* **[Secure Your Cluster](security/secure-your-cluster.md)** - The specific permissions NVIDIA Run:ai requires, for administrators and security officers evaluating the platform.

## Storage

* **[Storage](shared-storage.md)** - Sets up shared storage accessible to every worker node, so distributed training and shared datasets can be read and written consistently across nodes.

## Third-Party Integrations

* **[Supported Integrations](third-party-integrations/supported-integrations.md)** - Kubernetes components and external tools with built-in NVIDIA Run:ai support.
* **[Community Support Integrations](third-party-integrations/community-support-integrations.md)** - Integrations commonly used alongside NVIDIA Run:ai but not supported out of the box.
* **[Karpenter](third-party-integrations/karpenter.md)** - Guidance for running the Kubernetes cluster autoscaler Karpenter (optimizes for cost) alongside the NVIDIA Run:ai Scheduler (optimizes for fairness and utilization), since the two can conflict.
* **[Autonomous Job Recovery](third-party-integrations/autonomous-job-recovery.md)** - Integrates with AJR to detect and recover from hardware, network, and software failures in large-scale distributed training runs automatically.
