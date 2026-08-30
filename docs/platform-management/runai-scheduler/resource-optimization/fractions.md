# GPU Fractions

To submit a [workload](../../../workloads-in-nvidia-run-ai/workloads.md) with GPU resources in Kubernetes, you typically need to specify an integer number of GPUs. However, workloads often require diverse GPU memory and compute requirements or even use GPUs intermittently depending on the application (such as inference workloads, training workloads or notebooks at the model-creation phase). Additionally, GPUs are becoming increasingly powerful, offering more processing power and larger memory capacity for applications. Despite the increasing model sizes, the increasing capabilities of GPUs allow them to be effectively shared among multiple users or applications.

NVIDIA Run:ai’s GPU fractions provide an agile and easy-to-use method to share a GPU or multiple GPUs across workloads. With GPU fractions, you can divide the GPU/s memory into smaller chunks and share the GPU/s compute resources between different workloads and users, resulting in higher GPU utilization and more efficient resource allocation.

## Benefits of GPU Fractions

Utilizing GPU fractions to share GPU resources among multiple workloads provides numerous advantages for both platform administrators and practitioners, including improved efficiency, resource optimization, and enhanced user experience.

* For the AI practitioner:
  * **Reduced wait time** - Workloads with smaller GPU requests are more likely to be scheduled quickly, minimizing delays in accessing resources.
  * **Increased workload capacity** - More workloads can be run using the same admin-defined GPU [quota](../scheduling/concepts-and-principles.md#quota) and available unused resources - [over quota](../scheduling/concepts-and-principles.md#over-quota).
* For the platform administrator:
  * **Improved GPU utilization** - Sharing GPUs across workloads increases the utilization of individual GPUs, resulting in better overall platform efficiency.
  * **Higher resource availability** - More users gain access to GPU resources, ensuring better distribution.
  * **Enhanced workload throughput** - More workloads can be served per GPU, ensuring maximum output from existing hardware.
  * **Optimized scheduling** - Smaller and dynamic resource allocations gives the [Scheduler](../scheduling/concepts-and-principles.md) a higher chance of finding GPU resources for incoming workloads.

## Quota Planning with GPU Fractions

When planning the quota distribution for your [projects](../../aiinitiatives/organization/projects.md) and [departments](../../aiinitiatives/organization/departments.md), using fractions gives the platform administrator the ability to allocate more precise quota per project and department, assuming the usage of GPU fractions or enforcing it with [pre-defined policies](../../policies/policy-yaml-reference.md) or [compute resource](../../../workloads-in-nvidia-run-ai/workload-assets/compute-resources.md) templates.

For example, in an organization with a department budgeted for **two nodes of 8×H100 GPUs** and a team of 32 researchers:

* Allocating 0.5 GPU per researcher ensures all researchers have access to GPU resources.
* Using fractions enables researchers to run smaller workloads intermittently within their quota or go over their quota by using temporary over quota resources with higher resource demanding workloads.
* Using GPUs for notebook-based model development, where GPUs are not continuously active and can be shared among multiple users.

For more details on mapping your organization and resources, see [Adapting AI initiatives to your organization](../../aiinitiatives/adapting-ai-initiatives.md).

## How GPU Fractions Work

When a workload is submitted, the Scheduler finds a node with a GPU that can satisfy the requested GPU portion or GPU memory, then it schedules the pod to that node. The NVIDIA Run:ai GPU fractions logic, running locally on each NVIDIA Run:ai worker node, allocates the requested memory size on the selected GPU. **Each pod uses its own separate virtual memory address space.** NVIDIA Run:ai’s GPU fractions logic enforces the requested memory size, so no workload can use more than requested, and no workload can run over another workload’s memory. This gives users the experience of a ‘logical GPU’ per workload.

While [MIG](../../aiinitiatives/resources/mig-profiles.md) requires administrative work to configure every MIG slice, where a slice is a fixed chunk of memory, GPU fractions allow dynamic and fully flexible allocation of GPU memory chunks. By default, GPU fractions use NVIDIA’s time-slicing to share the GPU compute runtime. You can also use the [NVIDIA Run:ai GPU time-slicing](time-slicing.md) which allows dynamic and fully flexible splitting of the GPU compute time.

NVIDIA Run:ai GPU fractions are agile and dynamic allowing a user to allocate and free GPU fractions during the runtime of the system, at any size between zero to the maximum GPU portion (100%) or memory size (up to the maximum memory size of a GPU).

The NVIDIA Run:ai Scheduler can work alongside other schedulers. In order to avoid collisions with other schedulers, the NVIDIA Run:ai Scheduler creates special reservation pods. Once a workload is submitted requesting a fraction of a GPU, NVIDIA Run:ai will create a pod in a dedicated runai-reservation namespace with the full GPU as a resource, allowing other schedulers to understand that the GPU is reserved.

<Info>

**Note**

* Splitting a GPU into fractions may generate some fragmentation of the GPU memory. The [Scheduler](../scheduling/concepts-and-principles.md) will try to consolidate GPU resources where feasible (i.e. preemptible workloads).
* Using [bin-pack](../../aiinitiatives/resources/node-pools.md#adding-a-new-node-pool) as a scheduling placement strategy can also reduce GPU fragmentation.
* Using [dynamic GPU fractions ](dynamic-fractions.md)ensures that even small unused fragments of GPU memory are utilized by workloads.

</Info>

## Multi-GPU Fractions

NVIDIA Run:ai also supports workload submission using multi-GPU fractions. Multi-GPU fractions work similarly to single-GPU fractions, however, the NVIDIA Run:ai Scheduler allocates the same fraction size on multiple GPU devices within the same node. For example, if practitioners develop a new model that uses 8 GPUs and requires 40GB of memory per GPU, they can allocate 8×40GB with multi-GPU fractions instead of reserving the full memory of each GPU (e.g. 80GB). This leaves 40GB of GPU memory available on each of the 8 GPUs for other workloads within that node.

Time sharing where single GPUs can serve multiple workloads with fractions remains unchanged, only now, it serves multiple workloads using multi-GPUs per workload, single-GPU per workload, or a mix of both.

## Deployment Considerations

* Selecting a GPU portion using percentages as units does not guarantee the exact memory size. This means 50% of an A-100-40GB is 20GB while 50% of an A-100-80 is 40GB. To have better control over the exact allocated memory, specify the exact memory size, i.e. 40GB.
* Using NVIDIA Run:ai GPU fractions controls the memory split (i.e. 0.5 GPU means 50% of the GPU memory) but not the compute (processing time). To split the compute time, see [NVIDIA Run:ai’s GPU time slicing](time-slicing.md).
* NVIDIA Run:ai GPU fractions and [MIG mode](../../aiinitiatives/resources/mig-profiles.md) cannot be used on the same node.
* Some users restrict the use of Kubernetes containers with direct `hostPath` mounts due to stricter security policies and best practices. NVIDIA Run:ai offers an alternative for configuring fractions without relying on `hostPath`. Instead, you can enable device plugin–based host mounts by setting the `clusterConfig.global.devicePluginBindings` parameter to `true` (default is `false`, and uses the standard `hostPath` mount method). For details on how to configure this value using Helm or `runaiconfig`, see [Advanced cluster configurations](../../../installation-and-configuration/post-installation-configuration/cluster-config.md).

## Setting GPU Fractions

Using the [compute resources](../../../workloads-in-nvidia-run-ai/workload-assets/compute-resources.md) asset, you can define the compute requirements by specifying your requested GPU portion or GPU memory, and use it with any of the [NVIDIA Run:ai workload types](../../../workloads-in-nvidia-run-ai/workload-types/README.md) for single GPU and multi-GPU fractions.

* **Single-GPU fractions** - Define the compute requirement to run 1 GPU device, by specifying either a fraction (percentage) of the overall memory or a memory request (GB, MB).
* **Multi-GPU fractions** - Define the compute requirement to run multiple GPU devices, by specifying either a fraction (percentage) of the overall memory or a memory request (GB, MB).

## Setting GPU Fractions via YAML

To enable GPU fractions for workloads submitted via Kubernetes YAML, use the following annotations to define the GPU fraction configuration. You can configure either `gpu-fraction` or `gpu-memory`.

GPU fractions can be assigned to only one container in the pod. By default, GPU fraction allocation is applied to the first container (index 0) in the pod. Set the `gpu-fraction-container-name` annotation to specify which container should consume the fractional GPU resources. The specified container can be the main container, a sidecar, or an init container, any container other than the default first container.

<Info>

**Note**

Make sure the [default scheduler](../scheduling/default-scheduler.md) is set to `runai-scheduler`.

</Info>

| Variable | Input Format | Where to Set |
| --- | --- | --- |
| `gpu-fraction` | A portion of GPU memory as a double-precision floating-point number. Example: `0.25`, `0.75`. | Pod annotation (`metadata.annotations`) |
| `gpu-memory` | Memory size in MiB. Example: `2500,` `4096`. The `gpu-memory` values are always in MiB. | Pod annotation (`metadata.annotations`) |
| `gpu-fraction-num-devices` | The number of GPU devices to allocate using the specified `gpu-fraction` or `gpu-memory` value. Set this annotation only if you want to request multiple GPU devices. | Pod annotation (`metadata.annotations`) |
| `gpu-fraction-container-name` | By default, GPU fraction allocation is applied to the first container (index 0) in the pod. Set this annotation to specify a different container to receive the GPU allocation. | Pod annotation (`metadata.annotations`) |

The following example YAML creates a pod that requests 2 GPU devices, each requesting 50% of memory (`gpu-fraction: "0.5"`) .

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
