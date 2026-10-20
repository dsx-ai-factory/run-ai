Distributed Training, is the ability to split the training of a model among multiple processors.
It is often a necessity when multi-GPU training no longer applies;
typically when you require more GPUs than exist on a single node.
Each such split is a pod (see definition above). NVIDIA Run:ai spawns an additional launcher process that manages and
coordinates the other worker pods. For more information, see [Distributed training](/run-ai/self-hosted/documentation/workloads-in-nvidia-run-ai/using-training/distributed-training-models).
