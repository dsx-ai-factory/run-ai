# Before You Start

NVIDIA Run:ai provides [metrics and telemetry](metrics.md) for both physical cluster entities such as clusters, nodes, and node pools and application organization entities such as departments and projects. Metrics represent over-time data while telemetry represents current analytics data. This data is essential for monitoring and analyzing the performance and health of your platform.

## Consuming Metrics and Telemetry Data

Users can consume the data based on their permissions:

1. **API** - Access the data programmatically through the [NVIDIA Run:ai API](../../rest-api/about-the-rest-api.md).
2. **CLI** - Use the NVIDIA Run:ai [Command Line Interface](../../reference/cli/install-cli.md) to query and manage the data.
3. **UI** - Visualize the data through the NVIDIA Run:ai user interface.

### **API**

* **Metrics API** - Access over-time detailed analytics data programmatically.
* **Telemetry API** - Access current analytics data programmatically.

Refer to [metrics and telemetry](metrics.md) to see the full list of supported metrics and telemetry APIs.

### **CLI**

Use the `list` and `describe` commands to fetch and manage the data. See [CLI reference](../../reference/cli/runai/runai.md) for more details.

<figure><img src="../../assets/workload-telemetry-describe.png" alt=""/><figcaption><p>Describe a specific workload telemetry</p></figcaption></figure>

<figure><img src="../../assets/projects-telemetry-list.png" alt=""/><figcaption><p>List projects and view their telemetry and metrics</p></figcaption></figure>

### **UI Views**

Refer to [metrics and telemetry](metrics.md) to see the full list of supported metrics and telemetry.

* **Overview dashboard** - Provides a high-level summary of the cluster's health and performance, including key metrics such as GPU utilization, memory usage, and node status. Allows administrators to quickly identify any potential issues or areas for optimization. Offers advanced analytics capabilities for analyzing GPU usage patterns and identifying trends. Helps administrators optimize resource allocation and improve cluster efficiency.
* **Quota management** - Enables administrators to monitor and manage GPU quotas across the cluster. Includes features for setting and adjusting quotas, tracking usage, and receiving alerts when quotas are exceeded.
* **Workload visualizations** - Provides detailed insights into the resource usage and utilization of each GPU in the cluster. Includes metrics such as GPU memory utilization, core utilization, and power consumption. Allows administrators to identify GPUs that are under-utilized and overloaded.
* **Node and node pool visualizations** - Similar to workload visualizations, but focused on the resource usage and utilization of each GPU within a specific node or node pool. Helps administrators identify potential issues or bottlenecks at the node level.
* **Advanced NVIDIA metrics** - Provides access to a range of advanced NVIDIA metrics, such as GPU temperature, fan speed, and voltage. Enables administrators to monitor the health and performance of GPUs in greater detail. This data is available at the node and workload level.
