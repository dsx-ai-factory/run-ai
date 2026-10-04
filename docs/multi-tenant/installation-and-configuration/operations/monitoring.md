# Infrastructure Monitoring

The host organization is responsible for monitoring the health and performance of the NVIDIA Run:ai control plane and the underlying Kubernetes infrastructure. This ensures system stability, availability, and operational compliance across all tenant environments.

## What to Monitor

Monitoring focuses on three main infrastructure layers:

* **Control plane and cluster services** - This includes the NVIDIA Run:ai system components running in Kubernetes pods. The host organization must ensure these services are healthy and performant to maintain platform availability.
* **Kubernetes Cluster** - Monitor the overall state of the Kubernetes environment, including node health, API server availability, and core services. For general cluster guidance, refer to the [official Kubernetes documentation](https://kubernetes.io/docs/tasks/debug/debug-cluster/resource-usage-monitoring/)..
* **Host Infrastructure** - This represents the physical or virtual machines that host the Kubernetes nodes. The host organization is responsible for maintaining system-level health (CPU, memory, storage), OS patches, and networking. NVIDIA Run:ai does not require any special configurations at this level.

## Recommended Monitoring Setup with Grafana

The host organization is responsible for monitoring the health and performance of the NVIDIA Run:ai control plane and the underlying Kubernetes infrastructure.

We recommend using Grafana with Prometheus as the data source. Grafana offers a wide selection of prebuilt dashboards that provide insight into the control plane, cluster services, and node-level behavior.

Recommended dashboards include:

* **Kubernetes Cluster Monitoring** - Tracks overall cluster health, resource utilization, and node availability.
* **Node Exporter Full** - Displays detailed system-level metrics for each node (CPU, memory, disk, etc.).
* **Kubelet Metrics** - Monitors node-pod interactions and kubelet performance.
* **Kubernetes Resource Requests vs Limits** - Visualizes the gap between requested vs. actual resource usage for improved capacity planning.

These dashboards are available through the Grafana community and can be easily imported by ID. They help host organizations maintain observability across the multi-tenant control plane and ensure infrastructure-level SLAs are met.
