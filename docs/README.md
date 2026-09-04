---
hidden: true
noIndex: true
layout:
  width: default
  title:
    visible: true
  description:
    visible: true
  tableOfContents:
    visible: true
  outline:
    visible: true
  pagination:
    visible: false
  metadata:
    visible: true
  tags:
    visible: true
  actions:
    visible: true
---

# NVIDIA Run:ai Self-Hosted Product Documentation

NVIDIA Run:ai is an open-source AI workload and GPU orchestration platform for managing shared AI infrastructure on Kubernetes.

NVIDIA Run:ai provides a consistent way to run development, training, and inference workloads across Kubernetes environments, whether deployed on-premises, in the cloud, or across hybrid infrastructure. It combines GPU scheduling, resource management, workload orchestration, governance, and observability to help organizations share GPU resources efficiently across teams and clusters.

<Note>
NVIDIA Run:ai is available as open source. The NVIDIA Run:ai Support-Entitled Release is available for organizations that require NVIDIA enterprise support and an SLA.
</Note>

## Getting started

<Cards>
  <Card
    title="Get started with NVIDIA Run:ai"
    icon="rocket"
    href="getting-started/overview.md"
  >
    Learn how NVIDIA Run:ai works and follow the quick start guides to deploy infrastructure, configure the platform, and run your first workloads.
  </Card>

  <Card
    title="Install NVIDIA Run:ai"
    icon="gears"
    href="installation-and-configuration/overview.md"
  >
    Review system requirements and deploy NVIDIA Run:ai using the guided installer or Helm.
  </Card>

  <Card
    title="Local Development"
    icon="square-terminal"
    href="https://github.com/dsx-ai-factory/run-ai/blob/master/local-dev.md"
  >
    Build NVIDIA Run:ai and run the full stack locally for development.
  </Card>
</Cards>

## Using NVIDIA Run:ai

<Cards>
  <Card
    title="Manage shared AI infrastructure"
    icon="sitemap"
    href="platform-management/aiinitiatives/adapting-ai-initiatives.md"
  >
    Organize GPU resources across departments and projects, allocate resources to teams, and control how shared capacity is used.
  </Card>

  <Card
    title="Run AI workloads"
    icon="laptop-code"
    href="workloads-in-nvidia-run-ai/introduction-to-workloads.md"
  >
    Run development, training, distributed training, and inference workloads on shared Kubernetes infrastructure.
  </Card>

  <Card
    title="Optimize GPU resources"
    icon="calendar-circle-plus"
    href="platform-management/runai-scheduler/scheduling/concepts-and-principles.md"
  >
    Share GPU capacity across workloads and use scheduling, priorities, and resource allocation controls to improve utilization.
  </Card>

  <Card
    title="Operate NVIDIA Run:ai at scale"
    icon="gears"
    href="installation-and-configuration/operations/README.md"
  >
    Manage multiple clusters, configure platform services, monitor system health and resource utilization, and maintain your NVIDIA Run:ai deployment.
  </Card>
</Cards>

## Resources

<Cards>
  <Card
    title="Videos"
    icon="video"
    href="resources/videos/README.md"
  >
    Watch NVIDIA Run:ai platform overviews, walkthroughs, and task-based demonstrations.
  </Card>

  <Card
    title="Blogs"
    icon="newspaper"
    href="resources/blogs.md"
  >
    Explore NVIDIA Run:ai articles and technical content.
  </Card>

  <Card
    title="NVIDIA On-Demand Sessions"
    icon="play"
    href="resources/nvidia-on-demand-sessions.md"
  >
    Watch NVIDIA Run:ai sessions, presentations, and technical talks.
  </Card>
</Cards>
