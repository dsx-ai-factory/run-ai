# GPU Memory Swap

NVIDIA Run:ai’s GPU memory swap helps administrators and AI practitioners to further increase the utilization of their existing GPU hardware by improving GPU sharing between AI initiatives and stakeholders. This is done by expanding the GPU physical memory to the CPU memory, typically an order of magnitude larger than that of the GPU.

Expanding the GPU physical memory helps the NVIDIA Run:ai system to put more workloads on the same GPU physical hardware, and to provide a smooth workload context switching between GPU memory and CPU memory, eliminating the need to kill workloads when the memory requirement is larger than what the GPU physical memory can provide.

## Benefits of GPU Memory Swap

There are several use cases where GPU memory swap can benefit and improve the user experience and the system's overall utilization.

### Sharing a GPU Between Multiple Interactive Workloads (Notebooks)

AI practitioners use notebooks to develop and test new AI models and to improve existing AI models. While developing or testing an AI model, notebooks use GPU resources intermittently, yet, required resources of the GPUs are pre-allocated by the notebook and cannot be used by other workloads after one notebook has already reserved them. To overcome this inefficiency, NVIDIA Run:ai introduced [dynamic GPU fractions](dynamic-memory-fractions.md).

When one or more workloads require more than their requested GPU resources, there’s a high probability not all workloads can run on a single GPU because the total memory required is larger than the physical size of the GPU memory.

With GPU memory swap, several workloads can run on the same GPU, even if the sum of their used memory is larger than the size of the physical GPU memory. GPU memory swap can swap in and out workloads interchangeably, allowing multiple workloads to each use the full amount of GPU memory. The most common scenario is for one workload to run on the GPU (for example, an interactive notebook), while other notebooks are either idle or using the CPU to develop new code (while not using the GPU). From a user experience point of view, the swap in and out is a smooth process since the notebooks do not notice that they are being swapped in and out of the GPU memory. On rare occasions, when multiple notebooks need to access the GPU simultaneously, slower workload execution may be experienced.

Notebooks typically use the GPU intermittently, therefore with high probability, only one workload (for example, an [interactive notebook](../../../workloads-in-nvidia-run-ai/workload-types/native-workloads.md#workspaces-the-experimentation-phase)), will use the GPU at a time. The more notebooks the system puts on a single GPU, the higher the chances are that there will be more than one notebook requiring the GPU resources at the same time. Admins have a significant role here in fine tuning the number of notebooks running on the same GPU, based on specific use patterns and required SLAs.

### Sharing a GPU Between Inference/Interactive Workloads and Training Workloads

A single GPU can be shared between an [interactive or inference workload](../../../workloads-in-nvidia-run-ai/workload-types/native-workloads.md) (for example, a Jupyter notebook, image recognition services, or an LLM service), and a training workload that is not time-sensitive or delay-sensitive. At times when the inference/interactive workload uses the GPU, both training and inference/interactive workloads share the GPU resources, each running part of the time swapped-in to the GPU memory, and swapped-out into the CPU memory the rest of the time.

Whenever the inference/interactive workload stops using the GPU, the swap mechanism swaps out the inference/interactive workload GPU data to the CPU memory. Kubernetes wise, the pod is still alive and running using the CPU. This allows the training workload to run faster when the inference/interactive workload is not using the GPU, and slower when it does, thus sharing the same resource between multiple workloads, fully utilizing the GPU at all times, and maintaining uninterrupted service for both workloads.

### Serving Inference Warm Models with GPU Memory Swap

Running multiple[ inference models](../../../workloads-in-nvidia-run-ai/workload-types/native-workloads.md#inference-deploying-and-serving-models) is a demanding task and you will need to ensure that your SLA is met. You need to provide high performance and low latency, while maximizing GPU utilization. This becomes even more challenging when the exact model usage patterns are unpredictable. You must plan for the agility of inference services and strive to keep models on standby in a ready state rather than an idle state.

NVIDIA Run:ai’s GPU memory swap feature enables you to load multiple models to a single GPU, where each can use up to the full amount GPU memory. Using an application load balancer, the administrator can control to which server each inference request is sent. Then the GPU can be loaded with multiple models, where the model in use is loaded into the GPU memory and the rest of the models are swapped-out to the CPU memory. The swapped models are stored as ready models to be loaded when required. GPU memory swap always maintains the context of the workload (model) on the GPU so it can easily and quickly switch between models. This is unlike industry standard model servers that load models from scratch into the GPU whenever required.

## How GPU Memory Swap Works

Swapping the workload’s GPU memory to and from the CPU is performed simultaneously and synchronously for all GPUs used by the workload. In some cases, if workloads specify a memory limit smaller than a full GPU memory size, multiple workloads can run in parallel on the same GPUs, maximizing the utilization and shortening the response times.

In other cases, workloads will run serially, with each workload running for a few seconds before the system swaps them in/out. If multiple workloads occupy more than the GPU physical memory and attempt to run simultaneously, memory swapping will occur. In this scenario, each workload will run part of the time on the GPU while being swapped out to the CPU memory the other part of the time, slowing down the execution of the workloads. Therefore, it is important to evaluate whether memory swapping is suitable for your specific use cases, weighing the benefits against the potential for slower execution time. To better understand the benefits and use cases of GPU memory swap, refer to the detailed sections below. This will help you determine how to best utilize GPU swap for your workloads and achieve optimal performance.

The workload MUST use [dynamic GPU fractions](dynamic-memory-fractions.md). This means the workload’s memory Request is less than a full GPU, but it may add a GPU memory Limit to allow the workload to effectively use the full GPU memory. The NVIDIA Run:ai Scheduler allocates the dynamic fraction pair (Request and Limit) on single or multiple GPU devices in the same node.

To enable GPU memory swap, you must first configure the cluster with the required global setting, `global.core.swap.enabled`. After the cluster supports swap, you can enable GPU memory swap at the node-pool level. When creating or updating a node pool via the [Node pools](https://run-ai-docs.nvidia.com/api/2.26/organizations/nodepools) API, set the `gpuResourceOptimization.swapEnabled` parameter to `true`. NVIDIA Run:ai automatically applies the required `run.ai/swap-enabled=true` label to the nodes as part of this configuration. You can also use the `gpuResourceOptimization.cpuSwapMemorySize` field to specify the reserved CPU memory size for serving swapped GPU memory. See [Enabling and configuring GPU memory swap](memory-swap.md#enabling-and-configuring-gpu-memory-swap).

## Multi-GPU Memory Swap

NVIDIA Run:ai also supports workload submission using multi-GPU memory swap. Multi-GPU memory swap works similarly to single GPU memory swap, but instead of swapping memory for a single GPU workload, it swaps memory for workloads across multiple GPUs simultaneously and synchronously.

The NVIDIA Run:ai Scheduler allocates the same dynamic GPU fraction pair (Request and Limit) on multiple GPU devices in the same node. For example, if you want to run two LLM models, each consuming 8 GPUs that are not used simultaneously, you can use GPU memory swap to share their GPUs. This approach allows multiple models to be stacked on the same node.

The following outlines the advantages of stacking multiple models on the same node:

* **Maximizes GPU utilization** - Efficiently uses available GPU resources by enabling multiple workloads to share GPUs.
* **Improves cold start times** - Loading large LLM models to a node and its GPUs can take several minutes during a “cold start”. Using memory swap turns this process into a “warm start” that takes only a fraction of a second to a few seconds (depending on the model size and the GPU model).
* **Increases GPU availability** - Frees up and maximizes GPU availability for additional workloads (and users), enabling better resource sharing.
* **Smaller quota requirements** - Enables more precise and often smaller quota requirements for the end user.

## Deployment Considerations

* A pod created before the GPU memory swap feature was enabled in that cluster, cannot be scheduled to a swap-enabled node. A proper event is generated in case no matching node is found. Users must re-submit those pods to make them swap-enabled.
* GPU memory swap cannot be enabled if the NVIDIA Run:ai strict or fair time-slicing is used. GPU memory swap can only be used with the default NVIDIA time-slicing mechanism.
* CPU RAM size cannot be decreased once GPU memory swap is enabled.

## Enabling and Configuring GPU Memory Swap

Before configuring GPU memory swap, dynamic GPU fractions must be enabled. Dynamic GPU fractions enable you to make your workloads burstable as well as maximize your workloads’ performance and GPU utilization within a single node.

To enable GPU memory swap:

1. Enable the required global cluster setting. For more details, see [Advanced cluster configurations](../../../installation-and-configuration/post-installation-configuration/cluster-config.md):
   *   **Using Helm** - Set the following value in your `values.yaml` file under `clusterConfig` and upgrade the chart:

       ```yaml
       clusterConfig:
         global:
           core:
             swap:
               enabled: true
       ```
   *   **Using runaiconfig at runtime** - Use the following `kubectl` patch command:

       ```bash
       kubectl patch -n runai runaiconfigs.run.ai/runai --type='merge' --patch '{"spec":{"global":{"core":{"swap":{"enabled": true}}}}}'
       ```
2. When creating or updating a node pool via the [Node pools](https://run-ai-docs.nvidia.com/api/2.26/organizations/nodepools) API, set the `gpuResourceOptimization.swapEnabled` parameter to `true`. The required `run.ai/swap-enabled=true` label is automatically applied to the nodes.
3. Specify the desired swap memory limit using the `gpuResourceOptimization.cpuSwapMemorySize` field. The default amount of CPU RAM reserved for Swap is **100GB**. CPU memory is shared across all GPUs on a Kubernetes node (GPU server), therefore when setting this parameter, administrators should set this value based on both the number of GPUs in the server and the number of workloads expected to use Swap. For example, consider a GPU server with 8×H100 GPUs, each with 80 GB of memory, and 4 LLM workloads sharing those GPUs. If each LLM consumes 40 GB per GPU, the required CPU memory reservation is:
   * 8 GPUs × 40 GB = 320 GB per LLM
   * 320 GB × 4 LLM workloads = 1,280 GB (1.2 TB) of CPU RAM needed for Swap

See the [Node pools](https://run-ai-docs.nvidia.com/api/2.26/organizations/nodepools) API for more details. For example:

```bash
{
  "name": "v100",
  "labelKey": "node-type",
  "labelValue": "type-x",
  "clusterId": "d73a738f-fab3-430a-8fa3-xxxx",
  "gpuResourceOptimization": {
    "swapEnabled": true,
    "cpuSwapMemorySize": "100GB"
  }
}
```

### Configuring System Reserved GPU Resources

Swappable workloads require reserving a small part of the GPU memory for non-swappable allocations like binaries and GPU context. To avoid getting out-of-memory (OOM) errors due to non-swappable memory regions, the system reserves a 2GiB of GPU RAM memory by default, effectively truncating the total size of the GPU memory. For example, a 16GiB T4 will appear as 14GiB on a swap-enabled node. The exact reserved size is application-dependent, and 2GiB is a safe assumption for 2-3 applications sharing and swapping on a GPU.

This value can be changed when creating or updating a node pool via the API using the `gpuResourceOptimization.reservedGpuMemoryForSwapOperations` parameter. See the [Node pools](https://run-ai-docs.nvidia.com/api/2.26/organizations/nodepools) API for more details. For example:

```bash
{
  "name": "v100",
  "labelKey": "node-type",
  "labelValue": "type-x",
  "clusterId": "d73a738f-fab3-430a-8fa3-xxxx",
  "gpuResourceOptimization": {
    "swapEnabled": true,
    "cpuSwapMemorySize": "100GB",
    "reservedGpuMemoryForSwapOperations": "2GB"
  }
}
```

### Performance Optimizations

#### Using Bi-Directional GPU-CPU Memory Swap Read-Write Operations

To optimize GPU-to-CPU memory swap performance, administrators can enable full-duplex bi-directional read-write operations using `clusterConfig.global.core.swap.biDirectional.enabled`. See [Advanced cluster configurations](../../../installation-and-configuration/post-installation-configuration/cluster-config.md) for more details:

1.  **Using Helm** - Set the following value in your `values.yaml` file under `clusterConfig` and upgrade the chart:

    ```yaml
    clusterConfig:
      global:
        core:
          swap:
            biDirectional:
              enabled: true
    ```
2.  **Using runaiconfig at runtime** - Use the following `kubectl` patch command:

    ```bash
    kubectl patch -n runai runaiconfigs.run.ai/runai --type='merge' --patch '{"spec":{"global":{"core":{"swap":{"biDirectional": {"enabled": true}}}}}}'
    ```

Setting the read/write memory mode of GPU memory swap to bi-directional (full duplex) produces higher performance (typically +80%) vs. uni-directional (simplex) read-write operations.

#### Using UVA Based GPU-CPU Memory Mapped Swap Access

This setting enables the use of Unified Virtual Addressing (UVA) and early memory prefetching for GPU-to-CPU memory swap. It is especially effective when the GPU memory Request is close to the GPU memory Limit (e.g., Request = 90%, Limit = 100%). Enabling mapped mode can improve memory swap performance by +80–90% on newer GPUs such as H100 or B100 and up to +400% on GPUs such as A10 and L40.

Administrators can enable this setting using the `clusterConfig.global.core.swap.mode=mapped`. See [Advanced cluster configurations](../../../installation-and-configuration/post-installation-configuration/cluster-config.md) for more details:

1.  **Using Helm** - Set the following value in your `values.yaml` file under `clusterConfig` and upgrade the chart:

    ```yaml
    clusterConfig:
      global:
        core:
          swap:
            mode: mapped
    ```
2.  **Using runaiconfig at runtime** - Use the following `kubectl` patch command:

    ```bash
    kubectl patch -n runai runaiconfigs.run.ai/runai --type='merge' --patch '{"spec":{"global":{"core":{"swap":{"mode": "mapped"}}}}}'
    ```

### Preventing Your Workloads from Getting Swapped

If you prefer your workloads not to be swapped into CPU memory, you can specify on the pod an anti-affinity to `run.ai/swap-enabled=true` node label when submitting your workloads and the Scheduler will ensure not to use swap-enabled nodes. An alternative way is to set swap on a dedicated node pool and not use this node pool for workloads you prefer not to swap.

### What Happens When the CPU Reserved Memory for GPU Swap Is Exhausted?

CPU memory is limited, and since a single CPU serves multiple GPUs on a node, this number is usually between 2 to 8. For example, when using 80GB of GPU memory, each swapped workload consumes up to 80GB (but may use less) assuming each GPU is shared between 2-4 workloads. In this example, you can see how the swap memory can become very large. Therefore, we give administrators a way to limit the size of the CPU reserved memory for swapped GPU memory on each swap-enabled node as shown in [enabling and configuring GPU memory swap](memory-swap.md#enabling-and-configuring-gpu-memory-swap).
