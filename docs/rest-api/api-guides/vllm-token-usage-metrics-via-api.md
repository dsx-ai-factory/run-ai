# vLLM Token Usage Metrics via API

The vLLM token usage metrics report how many input and output tokens a workload served with vLLM processed over a period you choose. You can use them for chargeback, cost attribution, and capacity planning. NVIDIA Run:ai exposes these metrics through the Workloads API. Each request returns one result that covers the whole period, not a time series.

## Requirements

To use the vLLM token usage metrics, the following requirements apply:

* **Cluster version** - The cluster must run version 3.1 or later. The cluster starts recording token usage after it is upgraded to 3.1, so a period before the upgrade has no token usage.
* **vLLM workload** - The workload must be served with vLLM. NVIDIA Run:ai detects a vLLM workload automatically from the container image name, or you can set it explicitly with the `run.ai/vllm-workload` label.
* **vLLM version** - `VLLM_CACHED_PROMPT_TOKENS_TOTAL` requires vLLM 0.16.0 or later. The metric names were checked against vLLM 0.10.0 through 0.34.0.

## Token Usage Metrics

The following `metricType` values are available on the `GET /api/v1/workloads/{workloadId}/metrics` endpoint:

* `VLLM_PROMPT_TOKENS_TOTAL` - Input tokens processed in the period, including cached input tokens
* `VLLM_GENERATION_TOKENS_TOTAL` - Output tokens generated in the period
* `VLLM_CACHED_PROMPT_TOKENS_TOTAL` - Input tokens served from the prefix cache in the period. This value is already included in `VLLM_PROMPT_TOKENS_TOTAL`, so do not add it to the input total.
* `VLLM_FINISHED_ENGINE_REQUEST_COUNT` - Engine requests that finished in the period
* `VLLM_PROMPT_TOKENS_PER_REQUEST_AVG` - Average input tokens per finished engine request in the period
* `VLLM_GENERATION_TOKENS_PER_REQUEST_AVG` - Average output tokens per finished engine request in the period
* `VLLM_PROMPT_TOKENS_PER_REQUEST_PERCENTILES` - Approximate p50, p95, and p99 input tokens per finished engine request in the period, estimated from the vLLM histogram buckets
* `VLLM_GENERATION_TOKENS_PER_REQUEST_PERCENTILES` - Approximate p50, p95, and p99 output tokens per finished engine request in the period, estimated from the vLLM histogram buckets

All values are in tokens, except `VLLM_FINISHED_ENGINE_REQUEST_COUNT`, which is a number of requests.

## Requesting Token Usage

Set `start` and `end` to the period you want to report on. The period must be at least one second. The `numberOfSamples` parameter does not apply to these metrics. To request several metrics, separate the `metricType` values with commas.

The following example requests the input and output token totals for November:

```bash
curl -L 'https://<COMPANY-URL>/api/v1/workloads/<WORKLOAD-ID>/metrics?metricType=VLLM_PROMPT_TOKENS_TOTAL,VLLM_GENERATION_TOKENS_TOTAL&start=2026-11-01T00:00:00Z&end=2026-12-01T00:00:00Z' \
-H 'Authorization: Bearer <API-TOKEN>'
```

## Response Shape

The token totals, the request count, and the averages are returned under `measurements`, with one value for the period. The timestamp of the value is the end of the period:

```json
{
  "measurements": [
    {
      "type": "VLLM_PROMPT_TOKENS_TOTAL",
      "values": [
        { "value": "1843201", "timestamp": "2026-12-01T00:00:00Z" }
      ]
    }
  ]
}
```

The percentile metrics are returned under `histogram`, with one percentile map for the period:

```json
{
  "histogram": [
    {
      "type": "VLLM_PROMPT_TOKENS_PER_REQUEST_PERCENTILES",
      "values": [
        {
          "timestamp": "2026-12-01T00:00:00Z",
          "data": {
            "p50": "412",
            "p95": "1630",
            "p99": "3890"
          }
        }
      ]
    }
  ]
}
```

When there is no data for a metric, its `values` list is empty. Treat a missing value as unknown, not as zero. When no engine request finished in the period, `VLLM_FINISHED_ENGINE_REQUEST_COUNT` returns 0, and the averages and percentiles return no value.

## How Token Usage Is Calculated

NVIDIA Run:ai reads the vLLM token counters of every pod of the workload, in all node pools. Each pod and vLLM engine is calculated separately and then summed for the workload. This keeps pod restarts, scaling, and rolling updates from resetting or double counting the totals. Deleted pods still count for any period they overlap.

The request statistics combine the requests of all pods. For example, if one pod finishes one 100-token request and another pod finishes nine 10-token requests, the average is 19 tokens per request.

## Limitations

The following limitations apply to the vLLM token usage metrics:

* **Approximate totals** - Totals are calculated from samples collected at a regular interval, 30 seconds by default, and estimated at the start and end of the period. Missed collections and the estimates at the period boundaries can affect the result.
* **Short periods** - A period too short to hold two collected samples, about one minute at the default interval, returns no value.
* **Pods with too few samples** - A pod that has fewer than two collected samples in the period does not contribute to the totals. The response still returns a numeric total for the other pods.
* **Usage at the end of a pod's life** - Tokens served after a pod's last successful collection are not counted. This is usually up to one collection interval, but can be more if collections fail before the pod is removed.
* **Available history** - Usage is available only within the metrics retention period, 183 days by default, and only from the upgrade to cluster version 3.1. A successful response can cover only part of the requested period if older data has expired or the cluster was upgraded during the period.
* **Workload record cleanup** - Usage is available only while the workload record exists. By default, a deleted workload is removed 14 days after deletion, and a completed or failed workload is removed 28 days after it reaches that phase. These periods are configurable, and API access can end before the underlying data expires.
* **Cached input tokens** - If vLLM was upgraded to 0.16.0 or later during the period, or only some pods run a version that reports cached tokens, `VLLM_CACHED_PROMPT_TOKENS_TOTAL` covers only part of the usage. Subtracting it from `VLLM_PROMPT_TOKENS_TOTAL` then overstates the non-cached input.
* **Totals and request statistics do not reconcile** - The request statistics count the full size of requests that finished during the period. The token totals count tokens processed during the period, including requests that were cancelled or had not finished by the end of the period. The totals therefore do not equal the request count multiplied by the average.
* **Requests with `n` greater than 1** - vLLM handles an API call with `n` greater than 1 as `n` engine requests. The call counts as `n` requests, and its input is counted once per engine request, so `VLLM_PROMPT_TOKENS_TOTAL` can be higher than the input token count that the API call reports.
* **Cancelled requests** - Requests that the client cancels are not counted in the request statistics, but their processed tokens are included in the totals. Requests that vLLM itself aborts can be counted.
* **Percentiles** - The percentiles are estimated within the vLLM histogram buckets. If the bucket boundaries changed during the period, the percentiles are unreliable. Percentiles over long periods on workloads with many pods take longer to return than the other metrics.
