# Introduction to AI Applications

An AI application represents a high-level logical grouping of all Kubernetes resources (workloads, services, ConfigMaps, etc.) that together deliver a functional AI solution, such as a Retrieval-Augmented Generation (RAG) system or an image classification service.

## Key Characteristics

* **Unified purpose** - All grouped resources work together to deliver a single, identifiable functional goal.
* **Resource aggregation** - The AI application groups various NVIDIA Run:ai workloads and standard Kubernetes resources, providing a holistic view and management layer for complex, multi-component AI systems.
* **Simplified orchestration and observability** - Grouping related components into an AI Application makes it easier to identify resource dependencies and monitor the health and performance of the entire functional unit. This can help streamline operational tasks such as troubleshooting or scaling.
* **Collaboration and organization** - AI applications create a clear organizational boundary and naming structure for different functional AI systems running on the platform, aiding in organization and team collaboration.

## Deploying an AI Application with Helm

AI applications are deployed using Kubernetes Helm charts into an NVIDIA Run:ai project. You can submit an AI application directly from the NVIDIA Run:ai platform by selecting a chart from the NGC catalog or providing a custom Helm chart URL. See [AI applications](ai-applications.md) for step-by-step instructions.

<Info>

**Note**

AI applications can also be deployed externally by running a Helm chart directly into an NVIDIA Run:ai project namespace. For guidance on creating or configuring Helm charts, see the official [Helm](https://helm.sh/docs/intro/using_helm/) documentation.

</Info>

Once the chart is deployed:

* **Workload discovery** - Workloads created by the chart are automatically detected by NVIDIA Run:ai and appear in the [Workloads](../workloads-in-nvidia-run-ai/workloads.md) table.
* **Scheduling and management** - These workloads are scheduled, managed, and monitored according to NVIDIA Run:ai policies and capabilities. See [Supported features](../workloads-in-nvidia-run-ai/workload-types/supported-features.md#externally-submitted-kubernetes-workloads) for more details.
* **Application grouping** - NVIDIA Run:ai identifies related workloads and resources and groups them into a single AI application based on their shared context.
* **Aggregated visibility** - Resource consumption, status, and health are aggregated and presented at the AI application level, allowing the system to be viewed and analyzed as a single cohesive entity. See [AI applications](ai-applications.md) for more details.
