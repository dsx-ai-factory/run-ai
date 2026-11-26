Workloads are both native platform workloads (Workspaces, Training, and Inference) and workloads that originate from third-party ML frameworks, tools, or the broader Kubernetes ecosystem. For more details on the supported workloads, see [Introduction to workloads](/run-ai/self-hosted/documentation/workloads-in-nvidia-run-ai/introduction-to-workloads).

Use the Workloads endpoints to list, retrieve, and count workloads of all types in your environment, and to view their endpoints, structure, telemetry, and metrics.

## List, Retrieve, and Count Workloads

* List active workloads with their details
* Retrieve a single workload by its workload ID
* Count workloads

## Workload Telemetry

Retrieve aggregated data across workloads by telemetry type: workload count, GPU allocation, pending time distribution, and idle allocated GPUs. You can filter by cluster, node pool, department, project, and workload phase.

## Workload Metrics

Retrieve metrics for a single workload, such as GPU utilization and GPU memory usage. For inference workloads, you can also retrieve metrics from the inference server:

* **NVIDIA NIM** - Request throughput, latency, and token usage. Refer to [NIM observability metrics via API](/run-ai/self-hosted/reference/rest-api/api-guides/nim-observability-metrics-via-api).
* **vLLM** - Request concurrency, latency, and token usage, including input and output token totals for a period you choose. Refer to [vLLM token usage metrics via API](/run-ai/self-hosted/reference/rest-api/api-guides/vllm-token-usage-metrics-via-api).

## Workload Endpoints

Retrieve the network endpoints that a workload exposes.

## Workload Structure Elements

Retrieve the hierarchy of a workload as a list of structure elements. For each element, you can get its details, including its aggregated requested and allocated resources, and its metrics. You can also count the elements of a workload.
