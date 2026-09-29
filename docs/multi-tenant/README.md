---
hidden: true
noIndex: true
---

# NVIDIA Run:ai Multi-Tenant Product Documentation

NVIDIA Run:ai is an open-source AI workload and GPU orchestration platform for managing shared AI infrastructure on Kubernetes.

NVIDIA Run:ai provides a consistent way to run development, training, and inference workloads across Kubernetes environments, whether deployed on-premises, in the cloud, or across hybrid infrastructure. It combines GPU scheduling, resource management, workload orchestration, governance, and observability to help organizations share GPU resources efficiently across teams and clusters.

<Note>
**Support-Entitled Release**

NVIDIA Run:ai multi-tenant is available as open source. The NVIDIA Run:ai Support-Entitled Release is available for organizations that require [NVIDIA enterprise support and an SLA](../getting-started/product-support-policy.md).
</Note>

## Getting Started

<Cards>
  <Card
    title="Install NVIDIA Run:ai Multi-Tenant"
    icon="gears"
    href="installation-and-configuration/overview.md"
  >
    Review system requirements and deploy NVIDIA Run:ai multi-tenant using the guided installer or Helm.
  </Card>

  <Card
    title="Local Development"
    icon="square-terminal"
    href="../../local-dev.md"
  >
    Build NVIDIA Run:ai and run the full stack locally for development.
  </Card>
</Cards>

## Using NVIDIA Run:ai Multi-Tenant

<Cards>
  <Card
    title="Configure authentication and authorization"
    icon="sitemap"
    href="installation-and-configuration/configuration/authentication/roles.md"
  >
    Manage roles and access rules for administrators and tenants.
  </Card>

  <Card
    title="Operate at scale"
    icon="gears"
    href="installation-and-configuration/operations/scaling.md"
  >
    Tune NVIDIA Run:ai system services for scale and high availability as your number of tenants grows.
  </Card>

  <Card
    title="Monitor your deployment"
    icon="calendar-circle-plus"
    href="../installation-and-configuration/operations/observability/system-monitoring.md"
  >
    Monitor system health and resource utilization across the control plane and tenant clusters.
  </Card>
</Cards>

## Quick Links

<Cards>
  <Card
    title="What's New"
    icon="sparkles"
    href="../getting-started/whats-new/whats-new-3-0.md"
  >
    See the features and enhancements in the latest release.
  </Card>

  <Card
    title="Upgrade"
    icon="arrow-up"
    href="installation-and-configuration/upgrade.md"
  >
    Upgrade an existing NVIDIA Run:ai multi-tenant deployment to a new version.
  </Card>

  <Card
    title="Product Version Life Cycle"
    icon="calendar"
    href="../getting-started/product-support-policy.md#supported-versions"
  >
    Plan upgrades based on support timelines.
  </Card>
</Cards>
