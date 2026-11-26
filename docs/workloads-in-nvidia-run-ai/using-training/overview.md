# Overview

NVIDIA Run:ai provides GPU-accelerated training workloads for developing, training, and fine-tuning AI models. The platform supports standard and distributed training architectures, enabling AI practitioners to execute training jobs on a single node or distribute training across multiple nodes.

NVIDIA Run:ai supports popular machine learning frameworks, including PyTorch, TensorFlow, JAX, XGBoost, and MPI, and manages workload scheduling, resource allocation, and execution across shared GPU infrastructure.

## Use Cases

* **Model training and fine-tuning** - Train models from scratch or fine-tune pretrained models using GPU-accelerated compute resources.
* **Large-scale distributed training** - Train models across multiple nodes to accommodate larger models, datasets, and computational requirements.
* **Hyperparameter experimentation** - Execute multiple training runs, including parallel runs, to evaluate different model configurations.
* **Unattended training** - Submit training jobs that execute without continuous user interaction, including long-running workloads.

## Key Features and Capabilities

### Distributed Training and Performance

* **Flexible training architectures** - Run standard training workloads using a single main process or distributed training workloads with multiple coordinated processes across nodes.
* **Framework support** - Run distributed training with PyTorch, TensorFlow, JAX, XGBoost, and MPI.
* **Gang scheduling** - Schedule distributed training pods together to ensure that the required processes have the resources needed to run.
* **Topology-aware scheduling** - Place distributed training workers according to network topology constraints to reduce communication overhead and improve training efficiency.
* **Elastic training** - Configure minimum and maximum worker counts for supported PyTorch workloads, allowing the number of active workers to adjust to resource availability.

### Training Execution and Lifecycle

* **Multiple training runs** - Configure repeated training runs and control how many execute concurrently for standard training workloads.
* **Priority and preemption** - By default, training workloads run with low priority and are preemptible, allowing them to use resources beyond their project's quota when available.
* **Checkpointing and recovery** - Support long-running, preemptible training by saving checkpoints to persistent shared storage and configuring the training application to restore its state after an interruption.
* **Failure handling** - Configure retry limits and manage workload execution following failures. For distributed training, stale gang eviction handles situations where some pods remain running after others fail.

## Training Architectures

NVIDIA Run:ai supports two native training architectures.

### Standard Training

Standard training workloads execute a single main training process. They are suitable for workloads that do not require coordination across multiple nodes, including single-GPU and multi-GPU training on a single node.

Standard training workloads support multiple runs, with configurable parallelism for executing runs concurrently.

### Distributed Training

Distributed training workloads coordinate multiple training processes that can execute across different nodes. NVIDIA Run:ai supports master-and-worker and workers-only configurations, depending on the selected training framework.

Distributed training workloads also support topology-aware placement and, for supported PyTorch workloads, elastic worker scaling.

<Note>

**Note**

Multi-GPU training and distributed training are distinct concepts. Multi-GPU training uses multiple GPUs within a single node, while distributed training spans multiple nodes and requires coordination between training processes.

</Note>

## Workload Support and Framework Compatibility

### Native Training Workloads

NVIDIA Run:ai provides built-in support for standard and distributed training workloads through its native workload submission interfaces.

Native distributed training supports PyTorch, TensorFlow, JAX, XGBoost, and MPI.

### Additional Supported Training Workloads

NVIDIA Run:ai also supports training workloads from the Kubernetes and machine learning ecosystem, including:

* **Kubeflow training jobs** - PyTorchJob, TFJob, MPIJob, XGBoostJob, and JAXJob.
* **Ray workloads** - RayJob and RayCluster.
* **Other Kubernetes workloads** - JobSet and additional supported training workload types.

These workload types can be submitted using Kubernetes YAML manifests and managed through NVIDIA Run:ai scheduling, resource management, and monitoring capabilities.

For the complete list, refer to [Supported Workload Types](../workload-types/supported-workload-types.md).

## Getting Started

To create and configure a native training workload, refer to:

* [Standard Training](train-models.md)
* [Distributed Training](distributed-training-models.md)

For guidance on preserving progress during workload interruptions, refer to [Checkpointing Preemptible Training Workloads](checkpointing-preemptible-workloads.md).
