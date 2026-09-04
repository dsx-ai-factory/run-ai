# workload-exporter

The workload-exporter is a service that exports metrics about pods and workloads of runai.
In the past, we used metrics-exporter to export metrics about pod-groups,
but stating version 2.17 we need metrics on the workload level, and not on the pod-group level.
Detailed metrics deign is described here: https://runai.atlassian.net/wiki/spaces/EN/pages/2730623054/Design+-+New+Workloads+Metrics

## metrics exported by workload-exporter:
* runai_workload_info
* runai_pod_info
* runai_scheduler_requested_gpus_per_pod
* runai_scheduler_requested_gpu_memory_mb_per_pod
* runai_scheduler_allocated_gpus_per_pod
* runai_scheduler_allocated_gpu_memory_mb_per_pod
* runai_scheduler_allocated_millicpus_per_pod
* runai_scheduler_allocated_memory_bytes_per_pod