# Run Workloads Quick Start

This guide walks through running experiments and production workloads on NVIDIA Run:ai.

The quick start walks through the essential steps to begin using the platform, from initial access and project selection to launching a workspace and submitting your first workloads. The focus is on day-to-day workload execution and resource consumption, so you can experiment, train models, and deploy inference within your assigned project.

## Prerequisites

To begin, ensure you meet the following conditions:

* You have an active user account and credentials to access the NVIDIA Run:ai UI
* You are assigned to at least one project
* Your project has available resources to run workloads

## Getting Started

Choose a quick start based on your goal. Each scenario walks through a practical example so you can validate access, confirm resource availability, and understand how workloads run in your environment.

* [Run your first workspace](../../workloads-in-nvidia-run-ai/using-workspaces/quick-starts/jupyter-quickstart.md) - Launch a Jupyter notebook workspace for interactive development and experimentation. A guided tour is also available in the UI to help you familiarize yourself with the workspace experience.
* [Run a standard training workload](../../workloads-in-nvidia-run-ai/using-training/quick-starts/standard-training-quickstart.md) - Submit a standard training job to run a model training script on a single GPU.
* [Run a distributed training workload](../../workloads-in-nvidia-run-ai/using-training/quick-starts/distributed-training-quickstart.md) - Submit a distributed PyTorch training job and launch a multi-node training workload using an example PyTorch image.
* [Run a custom inference workload](../../workloads-in-nvidia-run-ai/using-inference/quick-starts/inference-quickstart.md) - Submit an inference workload and query the inference server to verify it is serving requests correctly.

## Understand Workload Capabilities

After completing the quick starts, explore the broader workload capabilities available in NVIDIA Run:ai. This helps you move beyond basic scenarios and take advantage of advanced scheduling, scaling, and configuration options.

* [Introduction to workloads](../../workloads-in-nvidia-run-ai/introduction-to-workloads.md) - How workloads are defined, scheduled, and executed in NVIDIA Run:ai.
* [Workload types and features](../../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md) - The different supported workload types and the capabilities available for each, including scaling, resource configuration, scheduling behavior, and other advanced options.
* [Workload assets](../../workloads-in-nvidia-run-ai/workload-assets/overview.md) - Shared resources used by workloads, such as environments, data sources, and credentials.
* [Workload templates](../../workloads-in-nvidia-run-ai/workload-templates/README.md) - Reusable configurations that help standardize and simplify workload creation.

## Run Workloads for Your Use Case

Once you understand the supported workload types and configuration options, proceed to the workload-specific documentation to configure and run workloads tailored for your project. Each workload section includes complete configuration examples and step-by-step instructions for the UI, API, and CLI.

* [Workspace](../../workloads-in-nvidia-run-ai/using-workspaces/running-workspace.md) - Interactive development environment for building and testing. Recommended for lightweight experimentation and debugging.
* [Training](../../workloads-in-nvidia-run-ai/using-training/train-models.md) - Workload for standard or distributed training models. Recommended for resource-intensive model development.
* [Inference](../../workloads-in-nvidia-run-ai/using-inference/nvidia-run-ai-inference-overview.md) - Deployment of an AI model for serving via an API. Recommended for production use.
* [Via YAML](../../workloads-in-nvidia-run-ai/submit-via-yaml.md) - Submission of a range of supported workload types using a standard Kubernetes YAML.

## Tutorials for End-to-End Workflows

Full end-to-end tutorials are available for deeper learning. These guides provide complete, practical examples that walk through development, training, and deployment workflows, showing how NVIDIA Run:ai features work together in real-world scenarios. See [Tutorials](../../tutorials/inference-tutorials/nvidia-nim-distributed.md) for more details.
