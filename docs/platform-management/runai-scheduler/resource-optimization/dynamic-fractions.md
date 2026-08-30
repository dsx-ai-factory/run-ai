# Dynamic GPU Fractions

Many workloads utilize GPU resources intermittently, with long periods of inactivity. These workloads typically need GPU resources when they are running AI applications or debugging a model in development. Other workloads such as inference may utilize GPUs at lower rates than requested, but may demand higher resource usage during peak utilization. The disparity between resource request and actual resource utilization often leads to inefficient utilization of GPUs. This usually occurs when multiple workloads request resources based on their peak demand, despite operating below those peaks for the majority of their runtime.

To address this challenge, NVIDIA Run:ai has introduced dynamic GPU fractions. This feature optimizes GPU utilization by enabling workloads to dynamically adjust their resource usage. It allows users to specify a guaranteed fraction of GPU memory and compute resources with a higher limit that can be dynamically utilized when additional resources are requested.

## How Dynamic GPU Fractions Work

With dynamic GPU fractions, users can [submit workloads](../../../workloads-in-nvidia-run-ai/workloads.md) using GPU fraction Request and Limit which is achieved by leveraging the Kubernetes Request and Limit notations. You can either:

* Request a GPU fraction (portion) using a percentage of a GPU and specify a Limit
* Request a GPU memory size (GB, MB) and specify a Limit

When setting a GPU memory limit either as GPU fraction or GPU memory size, the Limit must be equal to or greater than the GPU fractional memory request. Both GPU fraction and GPU memory are translated into the actual requested memory size of the Request (guaranteed resources) and the Limit (burstable resources - non guaranteed).

For example, a user can specify a workload with a GPU fraction request of 0.25 GPU, and add a limit of up to 0.80 GPU. The NVIDIA Run:ai [Scheduler](../scheduling/how-the-scheduler-works.md) schedules the workload to a node that can provide the GPU fraction request (0.25), and then assigns the workload to a GPU. The GPU scheduler monitors the workload and allows it to occupy memory between 0 to 0.80 of the GPU memory (based on the Limit), where only 0.25 of the GPU memory is guaranteed to that workload. The rest of the memory (from 0.25 to 0.8) is “loaned” to the workload, as long as it is not needed by other workloads.

NVIDIA Run:ai automatically manages the state changes between Request and Limit as well as the reverse (when the balance needs to be "returned"), updating the workloads’ utilization vs. Request and Limit parameters in the [metrics pane for each workload](../../../workloads-in-nvidia-run-ai/workloads.md).

To guarantee fair quality of service between different workloads using the same GPU, NVIDIA Run:ai developed an extendable GPUOOMKiller (Out Of Memory Killer) component that guarantees the quality of service using Kubernetes semantics for resources of Request and Limit.

The OOMKiller capability requires adding CAP\_KILL capabilities to the dynamic GPU fractions and to the NVIDIA Run:ai core scheduling module (toolkit daemon). This capability is enabled by default.

<Info>

**Note**

Dynamic GPU fractions is enabled by default in the [cluster](../../../installation-and-configuration/post-installation-configuration/cluster-config.md). Disabling dynamic GPU fractions removes the CAP\_KILL capability.

</Info>

## Multi-GPU Dynamic Fractions

NVIDIA Run:ai also supports workload submission using multi-GPU dynamic fractions. Multi-GPU dynamic fractions work similarly to dynamic fractions on a single GPU workload, however, instead of a single GPU device, the NVIDIA Run:ai Scheduler allocates the same dynamic fraction pair (Request and Limit) on multiple GPU devices within the same node. For example, if practitioners develop a new model that uses 8 GPUs and requires 40GB of memory per GPU, but may want to burst out and consume up to the full GPU memory, they can allocate 8×40GB with multi-GPU fractions and a limit of 80GB (e.g. H100 GPU) instead of reserving the full memory of each GPU (e.g. 80GB). This leaves 40GB of GPU memory available on each of the 8 GPUs for other workloads within that node.This is useful during model development, where memory requirements are usually lower due to experimentation with smaller models or configurations.

This approach significantly improves GPU utilization and availability, enabling more precise and often smaller quota requirements for the end user. Time sharing where single GPUs can serve multiple workloads with dynamic fractions remains unchanged, only now, it serves multiple workloads using multi-GPUs per workload.

## Deployment Considerations

Some users restrict the use of Kubernetes containers with direct `hostPath` mounts due to stricter security policies and best practices. NVIDIA Run:ai offers an alternative for configuring fractions without relying on `hostPath`. Instead, you can enable device plugin–based host mounts by setting the `clusterConfig.global.devicePluginBindings` parameter to `true` (default is `false`, and uses the standard `hostPath` mount method). For details on how to configure this value using Helm or `runaiconfig`, see [Advanced cluster configurations](../../../installation-and-configuration/post-installation-configuration/cluster-config.md).

## Setting Dynamic GPU Fractions

<Info>

**Note**

Dynamic GPU fractions is disabled by default in the NVIDIA Run:ai UI. To use dynamic GPU fractions, it must be enabled by your Administrator, under **General Settings** → Resources → GPU resource optimization.

</Info>

Using the [compute resources](../../../workloads-in-nvidia-run-ai/workload-assets/compute-resources.md) asset, you can define the compute requirements by specifying your requested GPU portion or GPU memory, and set a Limit. You can then use the compute resource with any of the [NVIDIA Run:ai workload types](../../../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md) for single and multi-GPU dynamic fractions. In addition, you will be able to view the workloads’ utilization vs. Request and Limit parameters in the [metrics pane for each workload](../../../workloads-in-nvidia-run-ai/workloads.md).

* **Single dynamic GPU fractions** - Define the compute requirement to run 1 GPU device, by specifying either a fraction (percentage) of the overall memory or specifying the memory request (GB, MB) with a Limit. The limit must be equal to or greater than the GPU fractional memory request.
* **Multi-GPU dynamic fractions** - Define the compute requirement to run multiple GPU devices, by specifying either a fraction (percentage) of the overall memory or specifying the memory request (GB, MB) with a Limit. The limit must be equal to or greater than the GPU fractional memory request.

<Info>

**Note**

When setting a workload with dynamic GPU fractions, (for example, when using it with GPU Request or GPU memory Limits), you practically make the workload burstable. This means it can use memory that is not guaranteed for that workload and is susceptible to an ‘OOM Kill’ signal if the actual owner of that memory requires it back. This applies to non-preemptible workloads as well. For that reason, it is recommended that you use dynamic GPU fractions with Interactive workloads running Notebooks. Notebook pods are not evicted when their GPU process is OOM Kill’ed. This behavior is the same as standard Kubernetes burstable CPU workloads.

</Info>

## Setting Dynamic GPU Fractions via YAML

To enable dynamic GPU fractions for workloads submitted via Kubernetes YAML, use the following annotations to define the GPU fraction configuration. You can configure either `gpu-fraction` or `gpu-memory`. You must also set the `RUNAI_GPU_MEMORY_LIMIT` environment variable in the container to enforce the memory limit.

GPU fractions can be assigned to only one container in the pod. By default, GPU fraction allocation is applied to the first container (index 0) in the pod. Set the `gpu-fraction-container-name` annotation to to specify which container should consume the fractional GPU resources. The specified container can be the main container, a sidecar, or an init container, any container other than the default first container.

<Info>

**Note**

Make sure the [default scheduler](../scheduling/default-scheduler.md) is set to `runai-scheduler`.

</Info>

| Variable | Input Format | Where to Set |
| --- | --- | --- |
| <code>gpu-fraction</code> | A portion of GPU memory as a double-precision floating-point number. Example: <code>0.25</code>, <code>0.75</code>. | Pod annotation (<code>metadata.annotations</code>) |
| <code>gpu-memory</code> | Memory size in MiB. Example: <code>2500,</code> <code>4096</code>. The <code>gpu-memory</code> values are always in MiB. | Pod annotation (<code>metadata.annotations</code>) |
| <code>gpu-fraction-num-devices</code> | The number of GPU devices to allocate using the specified <code>gpu-fraction</code> or <code>gpu-memory</code> value. Set this annotation only if you want to request multiple GPU devices. | Pod annotation (<code>metadata.annotations</code>) |
| <code>gpu-fraction-container-name</code> | By default, GPU fraction allocation is applied to the first container (index 0) in the pod. Set this annotation to specify a different container to receive the GPU allocation. | Pod annotation (<code>metadata.annotations</code>) |
| <code>RUNAI_GPU_MEMORY_LIMIT</code> | <ul><li>To use for <code>gpu-fraction</code> - Specify a double-precision floating-point number. Example: <code>0.95</code></li><li>To use for <code>gpu-memory</code> - Specify a Kubernetes resource quantity format. Example: <code>500000000</code>, <code>2500M</code></li></ul><p>The limit must be equal to or greater than the GPU fractional memory request.</p> | Environment variable in the container |

The following example YAML creates a pod that requests 2 GPU devices, each requesting 50% of memory (`gpu-fraction: "0.5"`) and allows usage of up to 95% (`RUNAI_GPU_MEMORY_LIMIT: "0.95"`) if available.

```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    user: test
    gpu-fraction: "0.5"
    gpu-fraction-num-devices: "2"
    # Specify which container should receive the GPU fraction allocation
    # By default, the first container (index 0) receives the GPU allocation
    # Use this annotation to specify a different container by name
    gpu-fraction-container-name: "gpu-workloads"
  labels:
    runai/queue: test
  name: multi-fractional-pod-job
  namespace: test
spec:
  containers:
  - image: gcr.io/run-ai-demo/quickstart-cuda
    imagePullPolicy: Always
    name: job
    env:
    - name: RUNAI_VERBOSE
      value: "1"
    - name: RUNAI_GPU_MEMORY_LIMIT
      value: "0.95"
    resources:
      limits:
        cpu: 200m
        memory: 200Mi
      requests:
        cpu: 100m
        memory: 100Mi
    securityContext:
      capabilities:
        drop: ["ALL"]
  schedulerName: runai-scheduler
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 5
```

## Using CLI

To view the available actions, go to the [CLI v2 reference](../../../reference/cli/runai/) and run according to your workload.

## Using API

To view the available actions, go to the [API reference](https://run-ai-docs.nvidia.com/api/2.26/) and run according to your workload.
