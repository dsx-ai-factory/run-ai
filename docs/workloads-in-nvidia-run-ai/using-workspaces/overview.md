# NVIDIA Run:ai Workspaces Overview

NVIDIA Run:ai provides interactive, containerized workspaces for AI development, experimentation, and data exploration. Workspaces combine development tools, compute resources, and data in a single environment, allowing AI practitioners to use their preferred tools and machine learning frameworks while NVIDIA Run:ai manages resource allocation, scheduling, and workload orchestration across shared infrastructure.

## Use Cases

* Develop, test, and debug AI models using interactive development environments such as JupyterLab and Visual Studio Code.
* Explore, analyze, and preprocess datasets, and develop data pipelines, before model training.
* Experiment with different model architectures, frameworks, and configurations, and validate approaches before running training or inference workloads.
* Provide teams with standardized development environments that include predefined tools, dependencies, compute resources, and storage.

## Key Features and Capabilities

NVIDIA Run:ai supports native workspaces and additional supported workload types for development. Feature availability can vary by workload type. For a detailed breakdown of which capabilities are supported for native workloads and supported workload types, see [Supported features](../workload-types/supported-features.md).

### Performance and Optimizations

* **Flexible resource allocation** - Assign GPU, CPU, and memory resources according to development and experimentation requirements.
* **GPU sharing** - Allocate fractional GPUs to development workloads that do not require a dedicated GPU. See [Request GPU memory fractions](../optimize-workload-scheduling-and-placement/gpu-sharing/request-gpu-memory-fractions.md).
* **Priority and scheduling** - By default, workspaces are assigned high priority and are non-preemptible, allowing practitioners to work without scheduling-related interruptions. A non-preemptible workspace cannot use resources beyond its project's deserved quota. See [Set priority and preemptibility](../optimize-workload-scheduling-and-placement/configure-priority-and-preemption.md).
* **Node pool selection** - Select node pools to target infrastructure that meets the workspace's resource requirements. See [Node pools](../../platform-management/aiinitiatives/resources/node-pools.md).

### Lifecycle Management

* **Custom development environments** - Use container images from NVIDIA NGC or custom registries, with configurable development tools, dependencies, and runtime settings.
* **Interactive access** - Connect to development tools running inside a workspace through configurable endpoints with user- and group-based access controls, or open a terminal in a running workspace. See [Connecting to a workload](../workloads.md#connecting-to-a-workload) and [Terminal](../workloads.md#terminal).
* **Reusable configurations** - Create workspaces from [workload templates](../workload-templates/workspace-templates/README.md) or reuse existing environment, compute resource, data source, and credential [workload assets](../workload-assets/overview.md).
* **Data and storage integration** - Mount datasets and configuration files, and attach persistent volumes to retain development files across workspace restarts.
* **Workload management** - Create, run, stop, connect to, and monitor workspaces. See [Workloads](../workloads.md).
* **Workload policies** - Apply organizational defaults and restrictions to workspace configurations. See [Policies and rules](../../platform-management/policies/policies-and-rules.md).

### Visualization and Monitoring

**Workload monitoring** - Monitor workspace status, resource utilization (GPU, CPU, and network), workload events, and container logs to understand resource consumption and troubleshoot workloads. See [Workloads](../workloads.md#show-hide-details).

## Workload Support and Framework Compatibility

### Native Workspaces

NVIDIA Run:ai provides native workspace workloads that use custom container images containing the development tools, libraries, and machine learning frameworks required for AI development. To create and configure a workspace, see [Running workspaces](running-workspace.md). For a quick start, see [Running Jupyter Notebooks using workspaces](quick-starts/jupyter-quickstart.md).

### Additional Supported Development Workloads

NVIDIA Run:ai supports additional workload types from the ML and Kubernetes ecosystem that are already registered in the platform and ready to use. Their default category is Build, which gives them the same default priority and preemptibility as native workspaces. Key workload types include the following:

* **Notebook** - Kubeflow notebook workloads for interactive machine learning development.
* **DevWorkspace** - Development environments defined through the DevWorkspace API.

These workloads can be submitted [via YAML](../submit-via-yaml.md). Available platform capabilities differ from those of native workspaces. For example, workload templates and workload assets are not available. See [Supported features](../workload-types/supported-features.md) for details, and [Supported workload types](../supported-workload-types/list-of-supported-workload-types.md) for the full list.

### Extending Development Workload Support

For emerging development tools and frameworks, Karta provides a declarative way to extend platform support. Administrators can register new workload types via the [Workload Types](api:main-api:POST/api/v1/workload-types) API, making them available across the organization without requiring platform updates or code changes. Once registered, these workloads can be submitted [via YAML](../submit-via-yaml.md) and managed with the same orchestration, monitoring, and scheduling capabilities as native workloads. See [Extending workload support with Karta](../supported-workload-types/extending-workload-support/README.md) for more details.
