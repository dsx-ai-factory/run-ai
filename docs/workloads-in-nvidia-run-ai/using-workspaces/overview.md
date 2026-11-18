# Overview

NVIDIA Run:ai provides interactive, containerized workspaces for AI development and experimentation. Workspaces combine development tools, compute resources, and data in a single environment, allowing AI practitioners to develop, test, and debug models without managing the underlying infrastructure.

## Use Cases

* **Interactive AI development** - Develop, test, and debug models using familiar tools such as JupyterLab and Visual Studio Code.
* **Model experimentation** - Explore model architectures, experiment with frameworks, and validate approaches before running training or inference workloads.
* **Data exploration and preparation** - Analyze datasets, preprocess data, and develop data pipelines using interactive notebooks.
* **Standardized development environments** - Provide teams with preconfigured environments containing the frameworks, libraries, and tools required for their projects.

## Key Features and Capabilities

* **Custom development environments** - Use container images from NVIDIA NGC or custom registries, with configurable development tools, dependencies, and runtime settings.
* **Reusable configurations** - Create workspaces from predefined templates or reuse existing environment, compute resource, and data configurations.
* **Interactive tool access** - Connect to development tools running inside a workspace through configurable endpoints with user- and group-based access controls.
* **Flexible compute allocation** - Assign GPU, CPU, and memory resources according to development requirements, including fractional GPU allocations.
* **Data and storage integration** - Mount datasets and configuration files, and attach persistent storage to retain development files across workspace restarts.
* **Workspace lifecycle management** - Create, start, stop, and monitor workspaces through NVIDIA Run:ai.

## Getting Started

To create and configure a workspace, refer to [Running workspaces](running-workspace.md).
