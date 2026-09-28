# NVIDIA Run:ai at Scale

Operating NVIDIA Run:ai at scale ensures that the system can efficiently handle fluctuating workloads while maintaining optimal performance. As clusters grow - whether due to an increasing number of nodes or a surge in workload demand - NVIDIA Run:ai services must be appropriately tuned to support large-scale environments.

This guide outlines the best practices for optimizing NVIDIA Run:ai for high-performance deployments, including NVIDIA Run:ai system services configurations, vertical scaling (adjusting CPU and memory resources) and where applicable, horizontal scaling (replicas).

## NVIDIA Run:ai Services

### Vertical Scaling

Each of the NVIDIA Run:ai containers has default resource requirements that reflect an average customer load. With significantly larger cluster loads, certain NVIDIA Run:ai services will require more CPU and memory resources. NVIDIA Run:ai supports configuring these resources for each NVIDIA Run:ai service group separately. For instructions and more information, see [NVIDIA Run:ai services resource management](../advanced-setup/cluster-config.md#nvidia-run-ai-services-resource-management).

#### Scheduling Services

The scheduling services group should be scaled together with the number of nodes and the number of workloads handled by the Scheduler (running / pending). These resource recommendations are based on internal benchmarks performed on stressed environments:

| Scale (nodes/workloads) | CPU (request) | Memory (request) |
| --- | --- | --- |
| Small - 30 / 480 | 1 | 1GB |
| Medium - 100 / 1600 | 2 | 2GB |
| Large - 500 / 8500 | 2 | 7GB |

#### Sync and Workload Services

The sync and workload service groups are less sensitive for scale. The recommendation for large or intensive environments is set to the following:

| CPU (request-limit) | Memory (request-limit) |
| --- | --- |
| 1-2 | 1GB-2GB |

### Horizontal Scaling

By default, NVIDIA Run:ai cluster services are deployed with a single replica. For large scale and intensive environments it is recommended to scale the NVIDIA Run:ai services horizontally by increasing the number of replicas. For more information, see [NVIDIA Run:ai services replicas](../advanced-setup/cluster-config.md#nvidia-run-ai-services-replicas).

## Metrics Collection

NVIDIA Run:ai relies on [Prometheus](../../installation-and-configuration/installation/guided-installation/prerequisites-and-features.md#prometheus) to scrape cluster metrics and forward them to the NVIDIA Run:ai control plane. The volume of metrics generated is directly proportional to the number of nodes, workloads, and projects in the system. When operating at scale—reaching hundreds, and thousands of nodes and projects—the system generates a significant volume of metrics which can place a strain on the cluster and the network bandwidth.

To mitigate this impact, it is recommended to tune the Prometheus [remote-write](https://prometheus.io/docs/specs/remote_write_spec/) configurations. See [remote write tuning](https://prometheus.io/docs/practices/remote_write/#remote-write-tuning) to read more about the tuning parameters available via the remote write configuration and refer to this [article](https://last9.io/blog/optimizing-prometheus-remote-write-performance-guide/) for optimizing Prometheus remote write performance.

You can apply the remote-write configurations required as described in [advanced cluster configurations.](../advanced-setup/cluster-config.md#prometheus)

The following example demonstrates the recommended approach in NVIDIA Run:ai for tuning **Prometheus remote-write** configurations:

```yaml
remoteWrite:
  queueConfig:
    capacity: 5000
    maxSamplesPerSend: 1000
    maxShards: 100
```
