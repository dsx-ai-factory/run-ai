# NVIDIA Run:ai SaaS Product Documentation

Find all the product information, step-by-step guides, and reference documentation for deploying and operating NVIDIA Run:ai in SaaS cloud environments.

## New to NVIDIA Run:ai

Begin with the quick start guides to set up the platform and run your first workloads.

<Cards>
  <Card title="Infrastructure Administrators" href="getting-started/quick-starts/infra-admin-quick-start.md" icon="fa-gears">
    Install the platform, complete the onboarding wizard, and set up cluster infrastructure
  </Card>
  <Card title="Platform Administrators" href="getting-started/quick-starts/platform-admin-quick-start.md" icon="fa-sitemap">
    Configure org structure, access control, resource quotas, and governance policies
  </Card>
  <Card title="AI Practitioners" href="getting-started/quick-starts/ai-practitioner-quick-start.md" icon="fa-laptop-code">
    Launch workspaces, run training, and deploy inference
  </Card>
</Cards>

---

## Already Managing a Deployment

Review the latest changes and upgrade guidance before making updates to your environment.

<Cards>
  <Card title="What's New" href="getting-started/whats-new-for-nvidia-run-ai-saas.md" icon="fa-bullhorn">
    New features, enhancements, and hotfixes
  </Card>
  <Card title="Upgrade" href="getting-started/installation/install-using-helm/upgrade.md" icon="fa-circle-up">
    Step-by-step instructions to upgrade your cluster
  </Card>
  <Card title="Product Version Life Cycle" href="support-policy/product-version-life-cycle.md" icon="fa-calendar">
    Plan upgrades based on support timelines
  </Card>
</Cards>

---

## Workloads (AI Lifecycle )

Step-by-step guides for running workloads at each stage of the AI lifecycle - from interactive experimentation to model training and production deployment.

<Cards>
  <Card title="Workspaces (Experiment)" href="workloads-in-nvidia-run-ai/using-workspaces/running-workspace.md" icon="fa-laptop">
    Create an interactive environment for model development and experimentation
  </Card>
  <Card title="Training (Standard)" href="workloads-in-nvidia-run-ai/using-training/train-models.md" icon="fa-gears">
    Submit a standard training workload to build and train your model
  </Card>
  <Card title="Training (Distributed)" href="workloads-in-nvidia-run-ai/using-training/distributed-training-models.md" icon="fa-network-wired">
    Run a distributed training workload across multiple nodes
  </Card>
  <Card title="Inference (deploy)" href="workloads-in-nvidia-run-ai/using-inference/nvidia-run-ai-inference-overview.md" icon="fa-rocket">
    Deploy and serve AI models with autoscaling and multi-node support
  </Card>
  <Card title="Supported workload types (via YAML )" href="workloads-in-nvidia-run-ai/submit-via-yaml.md" icon="fa-file-code">
    Submit supported workload types (such as Dynamo) using a YAML manifest via the UI
  </Card>
  <Card title="Tutorials (end-to-end)" href="tutorials/inference-tutorials/" icon="fa-book-open">
    Complete workflows that connect workspace → training → deployment
  </Card>
</Cards>

---

## Browse by Topic

Navigate directly to the documentation area most relevant to your role or task.

**Install, set up and monitor**

- [Install SaaS](../installation-and-configuration/installation/install-using-helm/system-requirements.md)
- [Set authenticated access](../infrastructure-setup/authentication/overview.md)
- [Set node roles and advanced cluster configurations](../installation-and-configuration/post-installation-configuration/node-roles.md)
- [Monitor, manage and restore clusters](../infrastructure-setup/procedures/clusters.md)
- [Monitor your platform](../infrastructure-setup/procedures/monitoring.md)

**Manage organizations and resources**

- [Map and set up your organizations](../platform-management/aiinitiatives/adapting-ai-initiatives.md#mapping-your-organization)
- [Set up and assign your resources](../platform-management/aiinitiatives/adapting-ai-initiatives.md#mapping-your-resources)
- [Manage permissions](../platform-management/aiinitiatives/adapting-ai-initiatives.md#assigning-users-to-projects-and-departments)
- [Create and manage policies](../platform-management/policies/policies-and-rules.md)
- [Monitor performance and health](../platform-management/monitor-performance/before-you-start.md)

**Build, train and deploy models**

- [Learn more about workloads and workload types](../workloads-in-nvidia-run-ai/introduction-to-workloads.md)
- [Prepare workload assets](../workloads-in-nvidia-run-ai/workload-assets/overview.md)
- [Build your model using workspaces](../workloads-in-nvidia-run-ai/using-workspaces/running-workspace.md)
- [Train your model using standard or distributed training workloads](../workloads-in-nvidia-run-ai/using-training/train-models.md)
- [Deploy your model with inference workloads](../workloads-in-nvidia-run-ai/using-inference/custom-inference.md)

**Scheduling and resource optimization**

- [Learn the NVIDIA Run:ai Scheduler concepts and principles](../platform-management/runai-scheduler/scheduling/concepts-and-principles.md)
- [Understand more about how the Scheduler works](../platform-management/runai-scheduler/scheduling/how-the-scheduler-works.md)
- [Explore different resource optimizations](../platform-management/runai-scheduler/resource-optimization/fractions.md)

**Use the CLI**

- [Install and configure the CLI](reference/cli/install-cli.md)
- [See the full list of commands and examples](reference/cli/runai/)

**Quick starts**

- [Run Jupyter Notebook using workspaces](../workloads-in-nvidia-run-ai/using-workspaces/quick-starts/jupyter-quickstart.md)
- [Run your first distributed training workload](../workloads-in-nvidia-run-ai/using-training/quick-starts/distributed-training-quickstart.md)
- [Launch workloads with dynamic GPU fractions](../platform-management/runai-scheduler/resource-optimization/quick-starts/dynamic-gpu-fractions-quickstart.md)
