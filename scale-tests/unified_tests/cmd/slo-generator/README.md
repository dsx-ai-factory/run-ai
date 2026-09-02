# SLO Generator

This tool generates performance metrics configuration from historical metrics data. It processes metrics data from Grafana Loki JSON files and calculates average values for each metric in each scale group.

## Usage

1. Place your metrics data JSON files in a directory. The files should contain metrics data in the Grafana Loki format.

2. Run the generator:
```bash
go run main.go --metrics-dir /path/to/metrics --output metrics_p95.json
```

## Output Format

The tool generates a JSON file with the following structure:

```json
{
  "scale_groups": [
    {
      "group_name": "scale-group-name",
      "metrics": [
        {
          "metric_name": "workload_first_pod_creation",
          "p95_value_ms": 29000
        },
        {
          "metric_name": "workload_pod_scheduling",
          "p95_value_ms": 49000
        }
        // ... other metrics
      ]
    }
    // ... other scale groups
  ]
}
```
run the query:
{job="workload_scale_test"} | json | scale_group=~"many-nodes-per-nodepool|many-projects|many-workloads|real-gpus" | json
In Loki and export the results as json into metrics/raw/grafan_metrics.json
The script will process this file and generate the output in the specified format.

## Metrics Processed


The tool processes the following metrics:
- `workload_first_pod_creation`: Time taken for the first pod to be created
- `workload_pod_scheduling`: Time taken for pod scheduling
- `workload_sync_to_cluster`: Time taken to sync to the cluster
- `workload_sync_to_control_plane`: Time taken to sync to the control plane
- `workload_whole_lifecycle`: Total time for the entire workload lifecycle

## Calculation Method

For each metric in each scale group:
1. All values from all reports are collected
2.  P95 is computed per metric for each scale group

## Command Line Options

- `--metrics-dir`: Directory containing metrics data files (required)
- `--output`: Output path for generated metrics configuration (default: "metrics_p95.json")
- `--debug`: Enable debug logging

## Example

```bash
# Generate metrics configuration with debug logging. run it from unified_tests directory
go run cmd/slo-generator/main.go --metrics-dir metrics/raw --output metrics/processed/metrics_p95.json --debug``` 