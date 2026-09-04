# NVIDIA Run:ai Native Workloads

NVIDIA Run:ai workloads are ready-to-use workload types provided natively in the platform, including Workspaces, Training, and Inference. Native workloads include built-in support for orchestration, scheduling, and policy controls, ensuring optimization, governance, and security standards. These workloads are fully integrated and immediately available as part of the platform. For more details on feature support, see [Supported features](supported-features.md).

## ML Lifecycle and Workload Mapping
In the world of machine learning (ML), the journey from raw data to actionable insights is a complex process that spans multiple stages. Each stage of the AI lifecycle requires different tools, resources, and frameworks to ensure optimal performance. NVIDIA Run:ai simplifies this process by offering specialized workload types tailored to each phase, facilitating a smooth transition across various stages of the ML workflows.

The ML lifecycle usually begins with the experimental work on data and exploration of different modeling techniques to identify the best approach for accurate predictions. At this stage, resource consumption is usually moderate as experimentation is done on a smaller scale. As confidence grows in the model's potential and its accuracy, the demand for compute resources increases. This is especially true during the training phase, where vast amounts of data need to be processed, particularly with complex models such as large language models (LLMs), with their huge parameter sizes, that often require distributed training across multiple GPUs to handle the intensive computational load.

Finally, once the model is ready, it moves to the inference stage, where it is deployed to make predictions on new, unseen data. NVIDIA Run:ai's workload types are designed to correspond with the natural stages of this lifecycle. They are structured to align with the specific resource and framework requirements of each phase, ensuring that AI researchers and data scientists can focus on advancing their models without worrying about infrastructure management.

## Workspaces: The Experimentation Phase

The **Workspace** is where data scientists conduct initial research, experiment with different data sets, and test various algorithms. This is the most flexible stage in the ML lifecycle, where models and data are explored, tuned, and refined. The value of workspaces lies in the flexibility they offer, allowing the researcher to iterate quickly without being constrained by rigid infrastructure.

*   **Framework flexibility**

    Workspaces support a variety of machine learning frameworks, as researchers need to experiment with different tools and methods.
*   **Resource requirements**

    Workspaces are often lighter on resources compared to the training phase, but they still require significant computational power for data processing, analysis, and model iteration.

    Hence, the default for the NVIDIA Run:ai workspaces considerations is to allow scheduling those workloads without the ability to preempt them once the resources were allocated. However, this non-preemptible state doesn’t allow utilizing more resources outside of the project’s deserved quota.

See [Running workspaces](../using-workspaces/running-workspace.md) to learn more about how to submit a workspace via the NVIDIA Run:ai platform. For quick starts, see [Running Jupyter Notebook using workspaces](../using-workspaces/quick-starts/jupyter-quickstart.md).

## Training: Scaling Resources for Model Development

As models mature and the need for more robust data processing and model training increases, NVIDIA Run:ai facilitates this shift through the Training workload. This phase is resource-intensive, often requiring distributed computing and high-performance clusters to process vast data sets and train models.

*   **Training architecture**

    For training workloads NVIDIA Run:ai allows you to specify the architecture - standard or distributed. The distributed architecture is relevant for larger data sets and more complex models that require utilizing multiple nodes. For the distributed architecture, NVIDIA Run:ai allows you to specify different configurations for the master and workers and select which framework to use - PyTorch, XGBoost, MPI, TensorFlow and JAX. In addition, as part of the distributed configuration, NVIDIA Run:ai enables the researchers to schedule their distributed workloads on nodes within the same region, zone, placement group, or any other topology.
*   **Resource requirements**

    Training tasks demand high memory, compute power, and storage. NVIDIA Run:ai ensures that the allocated resources match the scale of the task and allows those workloads to utilize more compute resources than the project’s deserved quota. Make sure that if you wish your training workload not to be preempted, specify the number of GPUs that are in your quota.

See [Standard training](../using-training/train-models.md) and [Distributed training](../using-training/distributed-training-models.md) to learn more about how to submit a training workload via the NVIDIA Run:ai UI. For quick starts, see [Run your first standard training](../using-training/quick-starts/standard-training-quickstart.md) and [Run your first distributed training](../using-training/quick-starts/distributed-training-quickstart.md).

<Info>

**Note**

Multi-GPU training and distributed training are two distinct concepts. Multi-GPU training uses multiple GPUs within a single node, whereas distributed training spans multiple nodes and typically requires coordination between them.

</Info>

## Inference: Deploying and Serving Models

Once a model is trained and validated, it moves to the Inference stage, where it is deployed to make predictions (usually in a production environment). This phase is all about efficiency and responsiveness, as the model needs to serve real-time or batch predictions to end-users or other systems.

*   **Inference-specific use cases**

    Naturally, inference workloads are required to change and adapt to the ever-changing demands to meet SLA. For example, additional replicas may be deployed, manually or automatically, to increase compute resources as part of a horizontal scaling approach or a new version of the deployment may need to be rolled out without affecting the running services.
*   **Resource requirements**

    Inference models differ in size and purpose, leading to varying computational requirements. For example, small OCR models can run efficiently on CPUs, whereas LLMs typically require significant GPU memory for deployment and serving. Inference workloads are considered production-critical and are given the highest priority to ensure compliance with SLAs. Additionally, NVIDIA Run:ai ensures that inference workloads cannot be preempted, maintaining consistent performance and reliability.

See [Deploy a custom inference workload](../using-inference/nvidia-run-ai-inference-overview.md) to learn more about how to submit an inference workload via the NVIDIA Run:ai UI. For a quick start, see [Run your first custom inference workload](../using-inference/quick-starts/inference-quickstart.md).
