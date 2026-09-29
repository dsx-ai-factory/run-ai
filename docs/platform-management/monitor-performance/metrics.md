# Metrics and Telemetry

Metrics are numeric measurements recorded **over time** that are emitted from the NVIDIA Run:ai cluster and telemetry is a numeric measurement recorded in real-time when emitted from the NVIDIA Run:ai cluster.

## Scopes

NVIDIA Run:ai provides control-plane API which supports and aggregates analytics at various levels.

| Level      | Description                                                                                                                                                                                           |
| ---------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Cluster    | A cluster is a set of nodes pools and nodes. With Cluster metrics, metrics are aggregated at the Cluster level. In the NVIDIA Run:ai user interface, metrics are available in the Overview dashboard. |
| Node       | Data is aggregated at the node level.                                                                                                                                                                 |
| Node pool  | Data is aggregated at the node pool level.                                                                                                                                                            |
| Workload   | Data is aggregated at the workload level. In some workloads, e.g. with distributed workloads, these metrics aggregate data from all worker pods.                                                      |
| Pod        | The basic unit of execution.                                                                                                                                                                          |
| Project    | The basic organizational unit. Projects are the tool to implement resource allocation policies as well as the segregation between different initiatives.                                              |
| Department | Departments are a grouping of projects.                                                                                                                                                               |

## Supported Metrics

| Metric name in API               | Applicable API endpoint                                                            | Metric name in UI per grid                                           | Applicable UI grid                                                        |
| -------------------------------- | ---------------------------------------------------------------------------------- | -------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| `ALLOCATED_GPU`                  | <ul><li>Clusters</li><li>Node pools</li></ul>                                      | <ul><li>GPU devices (allocated)</li><li>Allocated GPUs</li></ul>     | <ul><li>Overview dashboard</li><li>Node pools</li></ul>                   |
| `AVG_WORKLOAD_WAIT_TIME`         | <ul><li>Clusters</li><li>Node pools</li></ul>                                      |                                                                      |                                                                           |
| `CPU_LIMIT_CORES`                | Workloads                                                                          | CPU limit                                                            | Workloads                                                                 |
| `CPU_MEMORY_LIMIT_BYTES`         | Workloads                                                                          | CPU memory limit                                                     | Workloads                                                                 |
| `CPU_MEMORY_REQUEST_BYTES`       | Workloads                                                                          | CPU memory request                                                   | Workloads                                                                 |
| `CPU_MEMORY_USAGE_BYTES`         | <ul><li>Workloads</li><li>Pods</li></ul>                                           | CPU memory usage                                                     | Workloads                                                                 |
| `CPU_MEMORY_UTILIZATION`         | <ul><li>Clusters</li><li>Node pools</li><li>Nodes</li></ul>                        | CPU memory utilization                                               | <ul><li>Overview dashboard</li><li>Node pools</li><li>Nodes</li></ul>     |
| `CPU_REQUEST_CORES`              | Workloads                                                                          | CPU request                                                          | Workloads                                                                 |
| `CPU_USAGE_CORES`                | <ul><li>Nodes</li><li>Workloads</li><li>Pods</li></ul>                             | CPU usage                                                            | Workloads                                                                 |
| `CPU_UTILIZATION`                | <ul><li>Clusters</li><li>Node pools</li><li>Nodes</li></ul>                        | <ul><li>CPU compute utilization</li><li>CPU utilization</li></ul>    | <ul><li>Overview dashboard and Node pools</li><li>Nodes</li></ul>         |
| `GPU_ALLOCATION`                 | <ul><li>Workloads</li><li>Projects</li><li>Departments</li></ul>                   | GPU devices (allocated)                                              | Overview dashboard                                                        |
| `GPU_MEMORY_REQUEST_BYTES`       | Workloads                                                                          | GPU memory request                                                   | Workloads                                                                 |
| `GPU_MEMORY_USAGE_BYTES`         | <ul><li>Workloads</li><li>Pods</li><li>Nodes</li></ul>                             | GPU memory usage                                                     | Workloads                                                                 |
| `GPU_MEMORY_USAGE_BYTES_PER_GPU` | <ul><li>Nodes</li><li>Pods</li></ul>                                               | GPU memory usage per GPU                                             | Workloads per pod                                                         |
| `GPU_MEMORY_UTILIZATION`         | <ul><li>Clusters</li><li>Node pools</li></ul>                                      | GPU memory utilization                                               | <ul><li>Overview dashboard</li><li>Node pools</li></ul>                   |
| `GPU_MEMORY_UTILIZATION_PER_GPU` | Nodes                                                                              | GPU memory utilization per GPU                                       | Nodes                                                                     |
| `GPU_QUOTA`                      | <ul><li>Clusters</li><li>Node pools</li><li>Projects</li><li>Departments</li></ul> | Quota                                                                | Quota management                                                          |
| `GPU_UTILIZATION`                | <ul><li>Clusters</li><li>Node pools</li><li>Workloads</li><li>Pods</li></ul>       | GPU compute utilization                                              | <ul><li>Overview dashboard</li><li>Node pools</li><li>Workloads</li></ul> |
| `GPU_UTILIZATION_PER_GPU`        | <ul><li>Nodes</li><li>Pods</li></ul>                                               | GPU utilization per GPU                                              | Nodes                                                                     |
| `TOTAL_GPU`                      | <ul><li>Clusters</li><li>Node pools</li></ul>                                      | <ul><li>GPU devices total</li><li>Total GPUs</li></ul>               | <ul><li>Overview dashboard</li><li>Node pools</li></ul>                   |
| `TOTAL_GPU_NODES`                | <ul><li>Clusters</li><li>Node pools</li></ul>                                      |                                                                      |                                                                           |
| `GPU_UTILIZATION_DISTRIBUTION`   | <ul><li>Clusters</li><li>Node pools</li></ul>                                      | GPU utilization distribution                                         | Node pools                                                                |
| `UNALLOCATED_GPU`                | <ul><li>Clusters</li><li>Node pools</li></ul>                                      | <ul><li>GPU devices (unallocated)</li><li>Unallocated GPUs</li></ul> | <ul><li>Overview dashboard</li><li>Node pools</li></ul>                   |
| `CPU_QUOTA_MILLICORES`           | <ul><li>Projects</li><li>Departments</li></ul>                                     |                                                                      |                                                                           |
| `CPU_MEMORY_QUOTA_MB`            | <ul><li>Projects</li><li>Departments</li></ul>                                     |                                                                      |                                                                           |
| `CPU_ALLOCATION_MILLICORES`      | <ul><li>Projects</li><li>Departments</li></ul>                                     |                                                                      |                                                                           |
| `CPU_MEMORY_ALLOCATION_MB`       | <ul><li>Projects</li><li>Departments</li></ul>                                     |                                                                      |                                                                           |
| `POD_COUNT`                      | Workloads                                                                          |                                                                      |                                                                           |
| `RUNNING_POD_COUNT`              | Workloads                                                                          |                                                                      |                                                                           |
| `NVLINK_BANDWIDTH_TOTAL`         | <ul><li>Nodes</li><li>Pods</li></ul>                                               |                                                                      | <ul><li>Nodes</li><li>Workloads per pod</li></ul>                         |

### GPU Profiling

NVIDIA provides extended metrics as shown [here](https://docs.nvidia.com/datacenter/dcgm/latest/user-guide/feature-overview.html#profiling-metrics).

<Note>

**Note**

GPU profiling metrics are disabled by default. If unavailable, your administrator must enable it under **General settings** → Analytics → GPU profiling metrics. Before enabling, the administrator must configure GPU profiling through the DCGM Exporter and NVIDIA Run:ai Prometheus integration. For configuration steps, see [GPU profiling metrics](gpu-profiling-metrics.md).

</Note>

| Metric name in API | Applicable API endpoint | Metric name in UI | Applicable UI table |
| --- | --- | --- | --- |
| <code>GPU_FP16_ENGINE_ACTIVITY_PER_GPU</code> | <ul><li>Pods</li><li>Nodes</li></ul> | GPU FP16 engine activity | <ul><li>Workloads</li><li>Nodes</li></ul> |
| <code>GPU_FP32_ENGINE_ACTIVITY_PER_GPU</code> | <ul><li>Pods</li><li>Nodes</li></ul> | GPU FP32 engine activity | <ul><li>Workloads</li><li>Nodes</li></ul> |
| <code>GPU_FP64_ENGINE_ACTIVITY_PER_GPU</code> | <ul><li>Pods</li><li>Nodes</li></ul> | GPU FP64 engine activity | <ul><li>Workloads</li><li>Nodes</li></ul> |
| <code>GPU_GRAPHICS_ENGINE_ACTIVITY_PER_GPU</code> | <ul><li>Pods</li><li>Nodes</li></ul> | Graphics engine activity | <ul><li>Workloads</li><li>Nodes</li></ul> |
| <code>GPU_MEMORY_BANDWIDTH_UTILIZATION_PER_GPU</code> | <ul><li>Pods</li><li>Nodes</li></ul> | Memory bandwidth utilization | <ul><li>Workloads</li><li>Nodes</li></ul> |
| <code>GPU_NVLINK_RECEIVED_BANDWIDTH_PER_GPU</code> | <ul><li>Pods</li><li>Nodes</li></ul> | NVLink received bandwidth | <ul><li>Workloads</li><li>Nodes</li></ul> |
| <code>GPU_NVLINK_TRANSMITTED_BANDWIDTH_PER_GPU</code> | <ul><li>Pods</li><li>Nodes</li></ul> | NVLink transmitted bandwidth | <ul><li>Workloads</li><li>Nodes</li></ul> |
| <code>GPU_PCIE_RECEIVED_BANDWIDTH_PER_GPU</code> | <ul><li>Pods</li><li>Nodes</li></ul> | PCIe received bandwidth | <ul><li>Workloads</li><li>Nodes</li></ul> |
| <code>GPU_PCIE_TRANSMITTED_BANDWIDTH_PER_GPU</code> | <ul><li>Pods</li><li>Nodes</li></ul> | PCIe transmitted bandwidth | <ul><li>Workloads</li><li>Nodes</li></ul> |
| <code>GPU_SM_ACTIVITY_PER_GPU</code> | <ul><li>Pods</li><li>Nodes</li></ul> | GPU SM activity | <ul><li>Workloads</li><li>Nodes</li></ul> |
| <code>GPU_SM_OCCUPANCY_PER_GPU</code> | <ul><li>Pods</li><li>Nodes</li></ul> | GPU SM occupancy | <ul><li>Workloads</li><li>Nodes</li></ul> |
| <code>GPU_TENSOR_ACTIVITY_PER_GPU</code> | <ul><li>Pods</li><li>Nodes</li></ul> | GPU tensor activity | <ul><li>Workloads</li><li>Nodes</li></ul> |
| <code>GPU_OOMKILL_SWAP_OUT_OF_RAM_COUNT_PER_GPU</code> | Nodes | OOMKill swap out of RAM count | Nodes |
| <code>GPU_OOMKILL_BURST_COUNT_PER_GPU</code> | Nodes | OOMKill burst count | Nodes |
| <code>GPU_OOMKILL_IDLE_COUNT_PER_GPU</code> | Nodes | OOMKill idle count | Nodes |
| <code>GPU_SWAP_MEMORY_BYTES_PER_GPU</code> | Pods | GPU swap memory | Workloads |

### NVIDIA NIM

NVIDIA NIM metrics provide workload-level observability, including key runtime and performance data such as request throughput, latency, and token usage for LLMs. See [NIM observability metrics via API](../../rest-api/api-guides/nim-observability-metrics-via-api.md) for more details.

| Metric name in API | Applicable API endpoint | Metric name in UI | Applicable UI table |
| --- | --- | --- | --- |
| <code>NIM_NUM_REQUESTS_RUNNING</code> | <ul><li>Pods</li><li>Workloads</li></ul> | Request concurrency by status | Workloads |
| <code>NIM_NUM_REQUESTS_WAITING</code> | <ul><li>Pods</li><li>Workloads</li></ul> | Request concurrency by status | Workloads |
| <code>NIM_NUM_REQUEST_MAX</code> | <ul><li>Pods</li><li>Workloads</li></ul> | Request concurrency by status | Workloads |
| <code>NIM_REQUEST_SUCCESS_TOTAL</code> | <ul><li>Pods</li><li>Workloads</li></ul> | Request count by status | Workloads |
| <code>NIM_REQUEST_FAILURE_TOTAL</code> | <ul><li>Pods</li><li>Workloads</li></ul> | Request count by status | Workloads |
| <code>NIM_GPU_CACHE_USAGE_PERC</code> | <ul><li>Pods</li><li>Workloads</li></ul> | GPU KV cache utilization | Workloads |
| <code>NIM_TIME_TO_FIRST_TOKEN_SECONDS</code> | <ul><li>Pods</li><li>Workloads</li></ul> | Time to first token (TTFT) | Workloads |
| <code>NIM_E2E_REQUEST_LATENCY_SECONDS</code> | <ul><li>Pods</li><li>Workloads</li></ul> | End to end request latency | Workloads |
| <code>NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES</code> | <ul><li>Pods</li><li>Workloads</li></ul> | Time to first token (TTFT) by percentiles | Workloads |
| <code>NIM_E2E_REQUEST_LATENCY_SECONDS_PERCENTILES</code> | <ul><li>Pods</li><li>Workloads</li></ul> | End to end request latency by percentiles | Workloads |

### vLLM

vLLM metrics provide workload-level observability for large language model (LLM) inference workloads served with the vLLM inference engine, including request concurrency, token throughput, latency, and GPU KV-cache utilization. NVIDIA Run:ai detects a vLLM workload automatically from the container image name, or you can set it explicitly with the `run.ai/vllm-workload` label.

| Metric name in API | Applicable API endpoint | Metric name in UI per grid | Applicable UI grid |
| --- | --- | --- | --- |
| <code>VLLM_NUM_REQUESTS_RUNNING</code> | <ul><li>Pods</li><li>Workloads</li></ul> | Request concurrency by status | Workloads |
| <code>VLLM_NUM_REQUESTS_WAITING</code> | <ul><li>Pods</li><li>Workloads</li></ul> | Request concurrency by status | Workloads |
| <code>VLLM_REQUEST_SUCCESS</code> | <ul><li>Pods</li><li>Workloads</li></ul> | Request count by status | Workloads |
| <code>VLLM_REQUEST_SUCCESS_RATE</code> | <ul><li>Pods</li><li>Workloads</li></ul> | Request success rate | Workloads |
| <code>VLLM_KV_CACHE_USAGE_PERC</code> | <ul><li>Pods</li><li>Workloads</li></ul> | GPU KV cache utilization | Workloads |
| <code>VLLM_PROMPT_TOKENS_PER_SECOND</code> | <ul><li>Pods</li><li>Workloads</li></ul> | Token throughput | Workloads |
| <code>VLLM_GENERATION_TOKENS_PER_SECOND</code> | <ul><li>Pods</li><li>Workloads</li></ul> | Token throughput | Workloads |
| <code>VLLM_TIME_TO_FIRST_TOKEN_SECONDS</code> | <ul><li>Pods</li><li>Workloads</li></ul> | Time to first token (TTFT) | Workloads |
| <code>VLLM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES</code> | <ul><li>Pods</li><li>Workloads</li></ul> | Time to first token (TTFT) by percentiles | Workloads |
| <code>VLLM_E2E_REQUEST_LATENCY_SECONDS</code> | <ul><li>Pods</li><li>Workloads</li></ul> | End to end request latency | Workloads |
| <code>VLLM_E2E_REQUEST_LATENCY_SECONDS_PERCENTILES</code> | <ul><li>Pods</li><li>Workloads</li></ul> | End to end request latency by percentiles | Workloads |
| <code>VLLM_REQUEST_QUEUE_TIME_SECONDS_PERCENTILES</code> | <ul><li>Pods</li><li>Workloads</li></ul> | Queue time percentiles | Workloads |
| <code>VLLM_REQUEST_PREFILL_TIME_SECONDS_PERCENTILES</code> | <ul><li>Pods</li><li>Workloads</li></ul> | Latency phase breakdown | Workloads |
| <code>VLLM_REQUEST_DECODE_TIME_SECONDS_PERCENTILES</code> | <ul><li>Pods</li><li>Workloads</li></ul> | Latency phase breakdown | Workloads |
| <code>VLLM_INTER_TOKEN_LATENCY_SECONDS_PERCENTILES</code> | <ul><li>Pods</li><li>Workloads</li></ul> | Latency phase breakdown | Workloads |

## Supported Telemetry

| Metric                              | Applicable API endpoint                                          | Metric name in UI              | Applicable UI table                                          |
| ----------------------------------- | ---------------------------------------------------------------- | ------------------------------ | ------------------------------------------------------------ |
| `WORKLOADS_COUNT`                   | Workloads                                                        |                                |                                                              |
| `ALLOCATED_GPUS`                    | Nodes                                                            | Allocated GPUs                 | Nodes                                                        |
| `GPU_allocation`                    | <ul><li>Workloads</li><li>Projects</li><li>Departments</li></ul> |                                |                                                              |
| `READY_GPU_NODES`                   | Nodes                                                            | Ready / Total GPU nodes        | Overview dashboard                                           |
| `READY_GPUS`                        | Nodes                                                            | Ready / Total GPU devices      | Overview dashboard                                           |
| `TOTAL_GPU_NODES`                   | Nodes                                                            | Ready / Total GPU nodes        | Overview dashboard                                           |
| `TOTAL_GPUS`                        | Nodes                                                            | Ready / Total GPU devices      | Overview dashboard                                           |
| `IDLE_ALLOCATED_GPUS`               | Nodes                                                            | Idle allocated GPU devices     | Overview dashboard                                           |
| `FREE_GPUS`                         | Nodes                                                            | Free GPU devices               | Nodes                                                        |
| `TOTAL_CPU_CORES`                   | Nodes                                                            | CPU (Cores)                    | Nodes                                                        |
| `USED_CPU_CORES`                    | Nodes                                                            |                                |                                                              |
| `ALLOCATED_CPU_CORES`               | <ul><li>Nodes</li><li>Projects</li><li>Departments</li></ul>     | <p>Allocated CPU cores<br/></p> | Nodes                                                        |
| `TOTAL_GPU_MEMORY_BYTES`            | Nodes                                                            | GPU memory                     | Nodes                                                        |
| `USED_GPU_MEMORY_BYTES`             | Nodes                                                            | Used GPU memory                | Nodes                                                        |
| `TOTAL_CPU_MEMORY_BYTES`            | Nodes                                                            | CPU memory                     | Nodes                                                        |
| `USED_CPU_MEMORY_BYTES`             | Nodes                                                            | Used CPU memory                | Nodes                                                        |
| `ALLOCATED_CPU_MEMORY_BYTES`        | <ul><li>Nodes</li><li>Projects</li><li>Departments</li></ul>     | Allocated CPU memory           | <ul><li>Nodes</li><li>Projects</li><li>Departments</li></ul> |
| `GPU_QUOTA`                         | <ul><li>Projects</li><li>Departments</li></ul>                   | GPU quota                      | <ul><li>Projects</li><li>Departments</li></ul>               |
| `CPU_QUOTA`                         | <ul><li>Projects</li><li>Departments</li></ul>                   |                                |                                                              |
| `MEMORY_QUOTA`                      | <ul><li>Projects</li><li>Departments</li></ul>                   |                                |                                                              |
| `GPU_ALLOCATION_NON_PREEMPTIBLE`    | <ul><li>Projects</li><li>Departments</li></ul>                   |                                |                                                              |
| `CPU_ALLOCATION_NON_PREEMPTIBLE`    | <ul><li>Projects</li><li>Departments</li></ul>                   |                                |                                                              |
| `MEMORY_ALLOCATION_NON_PREEMPTIBLE` | <ul><li>Projects</li><li>Departments</li></ul>                   |                                |                                                              |
