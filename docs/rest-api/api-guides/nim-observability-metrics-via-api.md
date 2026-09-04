# NIM Observability Metrics via API

The NIM observability metrics provides programmatic access to key runtime and performance metrics for workloads deployed as NIM. These metrics extend beyond traditional infrastructure monitoring by offering workload-specific insights such as request throughput, latency and token usage (for LLMs). To support monitoring and optimization of these workloads, NVIDIA Run:ai exposes NIM observability metrics directly through the Workloads and Pods APIs.

NVIDIA Run:ai collects NIM observability metrics from workloads to generate charts and insights at the control plane level and does not expose the raw NIM metrics directly as-is. For more details, see [Observability for NVIDIA NIM for LLMs](https://docs.nvidia.com/nim/large-language-models/latest/observability.html).

## NIM Metrics Availability

To enable observability, NVIDIA NIM workloads must be deployed in one of the supported forms:

* **NVIDIA Run:ai** - Submit workloads through the NVIDIA Run:ai platform.
* **NIM Operator** - Recommended for enterprise deployments with full lifecycle management.
* **Helm Chart** - For Kubernetes-native deployment and integration with existing cluster tooling.
* **Container Image** - A direct deployment option where NIM runs as a container image. This approach requires two conditions to ensure workloads are correctly identified and monitored:
  * Apply the label `run.ai/nim-workload: "true"`
  * Define a service port with the name `serving-port`

## NIM-Specific Request Metrics

* `NIM_NUM_REQUESTS_RUNNING` - Number of requests currently running on GPU
* `NIM_NUM_REQUESTS_WAITING` - Number of requests waiting to be processed
* `NIM_NUM_REQUEST_MAX` - Max number of requests that can be run concurrently by the model
* `NIM_REQUEST_SUCCESS_TOTAL` - Number of successful requests, requests with finish reason “stop” or “length” are counted
* `NIM_REQUEST_FAILURE_TOTAL` - Number of failed requests, requests with other finish reason are counted
* `NIM_GPU_CACHE_USAGE_PERC` - GPU KV-cache usage - 1 means 100 percent usage

## NIM Latency Metrics

Histogram and percentile-based latency metrics provide detailed insights into request performance:

* `NIM_TIME_TO_FIRST_TOKEN_SECONDS` - Histogram of time to first token in seconds
* `NIM_E2E_REQUEST_LATENCY_SECONDS` - Histogram of end to end request latency in seconds
* `NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES` - Percentiles (p50, p90, p99) of time to first token in seconds
* `NIM_E2E_REQUEST_LATENCY_SECONDS_PERCENTILES` - Percentiles (p50, p90, p99) of end to end request latency in seconds

## Response Shape by Metric Type

The API returns metrics as either time-series gauges (under `measurements`) or histogram/percentile metrics (under `histogram`).

### Histogram-based NIM Metrics

If you request any of the following `metricType` values, the results appear in the `histogram` field:

* `NIM_TIME_TO_FIRST_TOKEN_SECONDS`
* `NIM_E2E_REQUEST_LATENCY_SECONDS`
* `NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES`
* `NIM_E2E_REQUEST_LATENCY_SECONDS_PERCENTILES`

### Percentile Variants

For metrics ending with `_PERCENTILES`, each entry is a timestamped percentile map:

```json
{
  "histogram": [
    {
      "type": "NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES",
      "values": [
        {
          "timestamp": "2025-08-20T10:51:26Z",
          "data": {
            "p50": 0.0533,
            "p90": 0.0533,
            "p99": 10.0
          }
        }
      ]
    }
  ]
}
```

### Raw Histogram Buckets

For non-percentile histogram metrics, each entry is a timestamped bucket → count map:

```json
{
  "histogram": [
    {
      "type": "NIM_TIME_TO_FIRST_TOKEN_SECONDS",
      "values": [
        {
          "timestamp": "2025-08-20T10:51:26Z",
          "data": {
            "0.001": 0,
            "0.005": 0,
            "+Inf": 6
          }
        }
      ]
    }
  ]
}
```

### Non-Histogram Metrics (Gauges)

All other metric types are returned as time-series measurements with numeric values per timestamp:

```json
{
  "measurements": [
    {
      "type": "GPU_UTILIZATION",
      "values": [
        { "value": "1",  "timestamp": "2025-08-20T10:51:26Z" },
        { "value": "10", "timestamp": "2025-08-20T10:52:26Z" },
        { "value": "11", "timestamp": "2025-08-20T10:53:26Z" }
      ]
    }
  ]
}
```

### Mixed Requests

When multiple `metricType` values are requested, the response may include both sections:

```json
{
  "measurements": [ /* non-histogram series here */ ],
  "histogram":    [ /* histogram/percentiles here */ ]
}
```
