# GPU Profiling Metrics

This guide describes how to enable advanced GPU profiling metrics from **NVIDIA Data Center GPU Manager (DCGM)**. These metrics provide deep visibility into GPU performance, including SM utilization, memory bandwidth, tensor core activity, and compute pipeline behavior - extending beyond standard GPU utilization metrics. For more details on metric definitions, see [NVIDIA profiling metrics](https://docs.nvidia.com/datacenter/dcgm/latest/user-guide/feature-overview.html#metrics).

## Available Metrics

Once enabled, NVIDIA Run:ai exposes the following GPU profiling metrics:

* SM activity - SM active cycles and occupancy
* Memory bandwidth - DRAM active cycles
* Compute pipelines - FP16/FP32/FP64 and tensor core activity
* Graphics engine - GR engine utilization
* PCIe/NVLink - Data transfer rates

NVIDIA Run:ai automatically aggregates these metrics at multiple levels:

* Per GPU device
* Per pod
* Per workload
* Per node

## Configuring the DCGM Exporter

The **DCGM Exporter** is responsible for exposing GPU performance metrics to Prometheus. To configure it for advanced metrics:

1.  Create the metrics configuration file and save it as `dcgm-metrics.csv`:

    ```csv
    # DCGM FIELD, Prometheus metric type, help message

    # Clocks
    DCGM_FI_DEV_SM_CLOCK,  gauge, SM clock frequency (in MHz).
    DCGM_FI_DEV_MEM_CLOCK, gauge, Memory clock frequency (in MHz).

    # Temperature
    DCGM_FI_DEV_MEMORY_TEMP, gauge, Memory temperature (in C).
    DCGM_FI_DEV_GPU_TEMP,    gauge, GPU temperature (in C).

    # Power
    DCGM_FI_DEV_POWER_USAGE,              gauge, Power draw (in W).
    DCGM_FI_DEV_TOTAL_ENERGY_CONSUMPTION, counter, Total energy consumption since boot (in mJ).

    # PCIE
    DCGM_FI_DEV_PCIE_REPLAY_COUNTER, counter, Total number of PCIe retries.

    # Utilization
    DCGM_FI_DEV_GPU_UTIL,      gauge, GPU utilization (in %).
    DCGM_FI_DEV_MEM_COPY_UTIL, gauge, Memory utilization (in %).
    DCGM_FI_DEV_ENC_UTIL,      gauge, Encoder utilization (in %).
    DCGM_FI_DEV_DEC_UTIL ,     gauge, Decoder utilization (in %).

    # Errors
    DCGM_FI_DEV_XID_ERRORS, gauge, Value of the last XID error encountered.

    # Memory
    DCGM_FI_DEV_FB_FREE, gauge, Framebuffer memory free (in MiB).
    DCGM_FI_DEV_FB_USED, gauge, Framebuffer memory used (in MiB).

    # NVLink
    DCGM_FI_DEV_NVLINK_BANDWIDTH_TOTAL, counter, Total number of NVLink bandwidth counters for all lanes.
    DCGM_FI_DEV_NVLINK_BANDWIDTH_L0,    counter, The number of bytes of active NVLink rx or tx data including both header and payload.

    # vGPU
    DCGM_FI_DEV_VGPU_LICENSE_STATUS, gauge, vGPU License status

    # Remapped rows
    DCGM_FI_DEV_UNCORRECTABLE_REMAPPED_ROWS, counter, Number of remapped rows for uncorrectable errors
    DCGM_FI_DEV_CORRECTABLE_REMAPPED_ROWS,   counter, Number of remapped rows for correctable errors
    DCGM_FI_DEV_ROW_REMAP_FAILURE,           gauge,   Whether remapping of rows has failed

    # Labels
    DCGM_FI_DRIVER_VERSION, label, Driver Version

    # DCP Profiling Metrics (Advanced)
    DCGM_FI_PROF_GR_ENGINE_ACTIVE,   gauge, Ratio of time the graphics engine is active (in %).
    DCGM_FI_PROF_SM_ACTIVE,          gauge, The ratio of cycles an SM has at least 1 warp assigned (in %).
    DCGM_FI_PROF_SM_OCCUPANCY,       gauge, The ratio of number of warps resident on an SM (in %).
    DCGM_FI_PROF_PIPE_TENSOR_ACTIVE, gauge, Ratio of cycles the tensor (HMMA) pipe is active (in %).
    DCGM_FI_PROF_DRAM_ACTIVE,        gauge, Ratio of cycles the device memory interface is active sending or receiving data (in %).
    DCGM_FI_PROF_PIPE_FP64_ACTIVE,   gauge, Ratio of cycles the fp64 pipes are active (in %).
    DCGM_FI_PROF_PIPE_FP32_ACTIVE,   gauge, Ratio of cycles the fp32 pipes are active (in %).
    DCGM_FI_PROF_PIPE_FP16_ACTIVE,   gauge, Ratio of cycles the fp16 pipes are active (in %).
    DCGM_FI_PROF_PCIE_TX_BYTES,      gauge, The rate of data transmitted over the PCIe bus - including both protocol headers and data payloads - in bytes per second.
    DCGM_FI_PROF_PCIE_RX_BYTES,      gauge, The rate of data received over the PCIe bus - including both protocol headers and data payloads - in bytes per second.
    DCGM_FI_PROF_NVLINK_TX_BYTES,    gauge, The number of bytes of active NvLink tx (transmit) data including both header and payload.
    DCGM_FI_PROF_NVLINK_RX_BYTES,    gauge, The number of bytes of active NvLink rx (read) data including both header and payload
    ```
2.  Create the following Helm values file and save it as `extended-dcgm-metrics-values.yaml`:

    ```yaml
    dcgmExporter:
      config:
        name: metrics-config
      env:
        - name: DCGM_EXPORTER_COLLECTORS
          value: /etc/dcgm-exporter/dcgm-metrics.csv
    ```
3.  Run the following to create the ConfigMap and upgrade the GPU operator:

    ```bash
    # Get GPU Operator version
    GPU_OPERATOR_VERSION=$(helm ls -A | grep gpu-operator | awk '{ print $10 }')

    # Create ConfigMap with metrics configuration
    kubectl create configmap metrics-config -n gpu-operator --from-file=dcgm-metrics.csv

    # Upgrade GPU Operator with new configuration
    helm upgrade -i gpu-operator nvidia/gpu-operator \
      -n gpu-operator \
      --version $GPU_OPERATOR_VERSION \
      --reuse-values \
      -f extended-dcgm-metrics-values.yaml
    ```

## Enabling NVIDIA Run:ai Metric Aggregation

Enable NVIDIA Run:ai to create enriched metrics from the DCGM profiling data. This configures Prometheus recording rules that aggregate raw DCGM metrics per pod, workload, and node. See [Advanced cluster configurations](../../installation-and-configuration/post-installation-configuration/cluster-config.md#prometheus) for more details.

*   **Using Helm** - Set the following value in your `values.yaml` file under `clusterConfig` and upgrade the chart:

    ```bash
    clusterConfig:  
      prometheus:
        spec: # PrometheusSpec
          config:
            advancedMetricsEnabled: true
    ```
*   **Using runaiconfig at runtime** - Use the following `kubectl` patch command:

    ```bash
    kubectl patch runaiconfig runai -n runai \
      --type=merge \
      -p '{
        "spec": {
          "prometheus": {
            "config": {
              "advancedMetricsEnabled": true
            }
          }
        }
      }'
    ```

## Enabling GPU Profiling Metrics Settings

GPU profiling metrics are disabled by default. To enable:

1. Go to **General settings** and navigate to Analytics
2. Enable **GPU profiling metrics**
3. Metrics become visible under the **Workloads** and **Nodes** pages once active

## Verification

<Tabs>

<Tab title="UI">

Workloads:

1. Navigate to **Workload manager** → Workloads
2. Click a row in the Workloads table and then click the SHOW DETAILS button at the upper-right side of the action bar. The details pane appears, presenting the **Metrics** tab.
3. In the **Type** dropdown, verify the **GPU profiling** option is available

Nodes:

1. Navigate to **Resources** → Nodes
2. Click a row in the Nodes table and then click the SHOW DETAILS button at the upper-right side of the action bar. The details pane appears, presenting the **Metrics** tab.
3. In the **Type** dropdown, verify the **GPU profiling** option is available

</Tab>

<Tab title="CLI v2">

```bash
# Check DCGM exporter pods are running
kubectl get pods -n gpu-operator -l app=nvidia-dcgm-exporter

# Check for advanced metrics in Prometheus (if accessible)

# Look for metrics like: DCGM_FI_PROF_SM_ACTIVE, DCGM_FI_PROF_DRAM_ACTIVE

# And Run:ai enriched metrics like: runai_gpu_sm_active_per_pod_per_gpu
```

</Tab>

<Tab title="API">

* Workloads - To view the GPU profiling metrics per pod, refer to the [Pods](https://run-ai-docs.nvidia.com/api/2.26/workloads/pods) API
* Nodes - To view the GPU profiling metrics for nodes, refer to the [Nodes](https://run-ai-docs.nvidia.com/api/2.26/organizations/nodes) API

</Tab>

</Tabs>
