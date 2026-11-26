# Request GPU Memory Fractions

Request a portion of GPU memory, or an absolute amount of GPU memory, for a workload, with or without a memory Limit. For how GPU memory fractions work and their prerequisites, refer to [GPU memory fractions](../../resource-optimization/memory-fractions.md) and [dynamic GPU memory fractions](../../resource-optimization/dynamic-memory-fractions.md).

## GPU Memory Fractions

### Setting GPU Memory Fractions


Using the [compute resources](../../../../workloads-in-nvidia-run-ai/workload-assets/compute-resources.md) asset, you can define the compute requirements by specifying your requested GPU portion or GPU memory, and use it with [NVIDIA Run:ai native workloads](../../../../workloads-in-nvidia-run-ai/workload-types/workload-types.md) for single-GPU and multi-GPU memory fractions.

* **Single-GPU memory fractions** - Define the compute requirement to run 1 GPU device, by specifying either a fraction (percentage) of the overall memory or a memory request (GB, MB).
* **Multi-GPU memory fractions** - Define the compute requirement to run multiple GPU devices, by specifying either a fraction (percentage) of the overall memory or a memory request (GB, MB).

### Setting GPU Memory Fractions via YAML

This method applies to [supported workload types](../../../../workloads-in-nvidia-run-ai/supported-workload-types/list-of-supported-workload-types.md) and to [workloads submitted externally](../../../../workloads-in-nvidia-run-ai/workload-types/supported-features.md#externally-submitted-kubernetes-workloads), outside of NVIDIA Run:ai. To enable GPU memory fractions for workloads submitted via Kubernetes YAML, use the annotations described below.

<Note>

**Note**

Make sure the [default scheduler](../../scheduling/default-scheduler.md) is set to `runai-scheduler`.

</Note>

#### Annotation Prefixes

* Annotations without a prefix (`gpu-fraction`, `gpu-memory`, `gpu-fraction-num-devices`, `gpu-fraction-container-name`) are Run:ai's own convention for the memory Request, consumed by the Scheduler rather than passed to CUDA. GPU memory portions are a Scheduler convention, because CUDA expresses memory only as an absolute size; the Scheduler resolves the portion against the GPU it selected and passes an absolute size to the node.
* `nvidia.com/container.<container-name>.gpu-memory.request` is the official NVIDIA-native alternative for an absolute-size Request only - it reaches the node and is applied directly to CUDA.
* There is no `nvidia.com/` equivalent for a portion Request; use `gpu-fraction` for that.

#### Annotation Reference

GPU memory fractions can be assigned to only one container in the pod. An annotation identifies that container in one of two ways:

* Embedded in the key, as `container.<container-name>` - used only by `nvidia.com/container.<container-name>.gpu-memory.request`.
* Named separately, by `gpu-fraction-container-name`, defaulting to the first container in the pod if omitted.

| Annotation | Input format | Purpose | Used with |
|---|---|---|---|
| `gpu-fraction` | A portion of GPU memory as a double-precision floating-point number. For example, `0.25`, `0.75`. | Requests a portion of GPU memory. | Optionally `gpu-fraction-num-devices` and `gpu-fraction-container-name`. No container-scoped form. |
| `gpu-memory` | Memory size in MiB. For example, `2500`, `4096`. | Requests an absolute amount of GPU memory. | Optionally `gpu-fraction-num-devices` and `gpu-fraction-container-name`. Alternative: `nvidia.com/container.<container-name>.gpu-memory.request`. |
| `gpu-fraction-num-devices` | The number of GPU devices to allocate using the specified `gpu-fraction` or `gpu-memory` value. Defaults to 1 if omitted. | Requests multiple GPU devices for the fractional container. | Any Request form - `gpu-fraction`, `gpu-memory`, or `nvidia.com/container.<container-name>.gpu-memory.request`. The number of GPU devices is a pod-level property, so this annotation has no container-scoped form. |
| `gpu-fraction-container-name` | The name of a container in the pod. Defaults to the first container (index 0) if omitted. | Selects which container receives the GPU allocation when the request is made with `gpu-fraction` or `gpu-memory`. | Not used with `nvidia.com/container.<container-name>.gpu-memory.request` - that form names the container in its key. If both are present, the names must match or the request is rejected. |
| `nvidia.com/container.<container-name>.gpu-memory.request` | Memory size in Kubernetes resource quantity format. For example, `40G`. | Requests an absolute amount of GPU memory. This is the official NVIDIA alternative to `gpu-memory`. | Optionally `gpu-fraction-num-devices`. |

<Note>

**Note**

* To set how a workload shares GPU compute, refer to the annotations described in [GPU compute modes](../../resource-optimization/compute-modes.md).
* Dynamic GPU memory fractions add a memory limit to the request. Refer to [dynamic GPU memory fractions](../../resource-optimization/dynamic-memory-fractions.md) for those annotations.

</Note>

#### Examples

* Requests an absolute amount of GPU memory on 4 GPU devices:

  ```yaml
  metadata:
    annotations:
      gpu-memory: "4000"   # MiB
      # optional; 1 if omitted - number of fractional GPU devices
      gpu-fraction-num-devices: "4"
  spec:
    containers:
      - name: app
  ```

* Requests the same absolute amount of GPU memory using the official NVIDIA-native annotation instead of `gpu-memory`. The container name moves into the key, and device count still uses the unprefixed `gpu-fraction-num-devices`:

  ```yaml
  metadata:
    annotations:
      nvidia.com/container.app.gpu-memory.request: "40G"
      # optional; 1 if omitted - number of fractional GPU devices
      gpu-fraction-num-devices: "4"
  spec:
    containers:
      - name: app
  ```

* Requests a portion of GPU memory on 2 GPU devices:

  ```yaml
  metadata:
    annotations:
      gpu-fraction: "0.5"
      # optional; 1 if omitted - number of fractional GPU devices
      gpu-fraction-num-devices: "2"
  spec:
    containers:
      - name: app
  ```

* Complete example: creates a pod that requests 2 GPU devices, each requesting 50% of memory:

  ```yaml
  apiVersion: v1
  kind: Pod
  metadata:
    annotations:
      user: test
      gpu-fraction: "0.5"
      gpu-fraction-num-devices: "2"
      # Specify which container should receive the GPU memory fraction allocation
      # By default, the first container (index 0) receives the GPU allocation
      gpu-fraction-container-name: "app"
    labels:
      runai/queue: test
    name: multi-fractional-pod-job
    namespace: test
  spec:
    containers:
    - image: <image>
      imagePullPolicy: Always
      name: app
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

## Dynamic GPU Memory Fractions

### Setting Dynamic GPU Memory Fractions


Using the [compute resources](../../../../workloads-in-nvidia-run-ai/workload-assets/compute-resources.md) asset, you can define the compute requirements by specifying your requested GPU memory portion or GPU memory size, and set a Limit. You can then use the compute resource with [NVIDIA Run:ai native workloads](../../../../workloads-in-nvidia-run-ai/workload-types/workload-types.md) for single-GPU and multi-GPU dynamic memory fractions. In addition, you can view the workloads' utilization compared to Request and Limit parameters in the [metrics pane for each workload](../../../../workloads-in-nvidia-run-ai/workloads.md).

* **Single-GPU dynamic memory fractions** - Define the compute requirement to run 1 GPU device, by specifying either a portion (percentage) of the overall memory or a memory request (GB, MB) with a Limit.
* **Multi-GPU dynamic memory fractions** - Define the compute requirement to run multiple GPU devices, by specifying either a portion (percentage) of the overall memory or a memory request (GB, MB) with a Limit.

<Note>

**Note**

Dynamic GPU memory fractions is disabled by default. To set a memory Limit through the NVIDIA Run:ai UI or CLI, it must be enabled by your administrator under **General settings** → **Resources** → **GPU resource optimization**. While the setting is disabled, the UI does not offer a Limit field on a compute resource, and the CLI rejects a submission that specifies one.

The setting does not apply to supported workload types submitted via YAML or to externally submitted workloads, since their Limit is set directly with a Kubernetes annotation rather than through the UI or CLI. Do not rely on this setting to prevent the use of memory Limits across a cluster.

</Note>

### Setting Dynamic GPU Memory Fractions via YAML

This method applies to [supported workload types](../../../../workloads-in-nvidia-run-ai/supported-workload-types/list-of-supported-workload-types.md) and to [workloads submitted externally](../../../../workloads-in-nvidia-run-ai/workload-types/supported-features.md#externally-submitted-kubernetes-workloads), outside of NVIDIA Run:ai. To enable dynamic GPU memory fractions for workloads submitted via Kubernetes YAML, use the annotations described below.

<Note>

**Note**

Make sure the [default scheduler](../../scheduling/default-scheduler.md) is set to `runai-scheduler`.

</Note>


#### Annotation Prefixes

* Annotations without a prefix (`gpu-fraction`, `gpu-memory`, `gpu-fraction-num-devices`, `gpu-fraction-container-name`) set the Request, and are Run:ai's own convention, consumed by the Scheduler rather than passed to CUDA.
* `nvidia.com/container.<container-name>.gpu-memory.request` is the official NVIDIA-native alternative for an absolute-size Request.
* `nvidia.com/container.<container-name>.gpu-memory.limit` sets the Limit for an absolute-size Request - the annotation reaches the node and is applied directly to CUDA.
* `kai.scheduler/container.<container-name>.gpu-memory.portion.limit` sets the Limit for a portion Request - the annotation is consumed by the Scheduler only and is never passed to the node. GPU memory portions are a Scheduler convention, because CUDA expresses memory only as an absolute size; the Scheduler resolves the portion against the GPU it selected and passes an absolute size to the node.

#### Annotation Reference

A dynamic GPU memory fraction can be assigned to only one container in the pod. The Request sets the guaranteed memory, and the Limit sets the ceiling the workload can grow into. An annotation identifies its container in one of two ways:

* Embedded in the key, as `container.<container-name>` - used by every `nvidia.com/` and `kai.scheduler/` annotation.
* Named separately, by `gpu-fraction-container-name`, defaulting to the first container in the pod if omitted.

| Annotation | Input format | Purpose | Used with |
|---|---|---|---|
| `gpu-fraction` | A portion of GPU memory as a double-precision floating-point number. For example, `0.25`, `0.75`. | Requests a portion of GPU memory. | Optionally `gpu-fraction-num-devices` and `gpu-fraction-container-name`. No container-scoped form. Limit: `kai.scheduler/container.<container-name>.gpu-memory.portion.limit`. |
| `gpu-memory` | Memory size in MiB. For example, `2500`, `4096`. | Requests an absolute amount of GPU memory. | Optionally `gpu-fraction-num-devices` and `gpu-fraction-container-name`. Alternative Request form: `nvidia.com/container.<container-name>.gpu-memory.request`. Limit: `nvidia.com/container.<container-name>.gpu-memory.limit`. |
| `gpu-fraction-num-devices` | The number of GPU devices to allocate using the specified `gpu-fraction` or `gpu-memory` value. Defaults to 1 if omitted. | Requests multiple GPU devices for the fractional container. | Any Request form - `gpu-fraction`, `gpu-memory`, or `nvidia.com/container.<container-name>.gpu-memory.request`. The number of GPU devices is a pod-level property, so this annotation has no container-scoped form. |
| `gpu-fraction-container-name` | The name of a container in the pod. Defaults to the first container (index 0) if omitted. | Selects which container receives the GPU allocation when the request is made with `gpu-fraction` or `gpu-memory`. | Not used with an annotation that names the container in its key. If both are present, the names must match or the request is rejected. |
| `nvidia.com/container.<container-name>.gpu-memory.request` | Memory size in Kubernetes resource quantity format. For example, `40G`. | Requests an absolute amount of GPU memory. This is the official NVIDIA alternative to `gpu-memory`. | Optionally `gpu-fraction-num-devices`. Limit: `nvidia.com/container.<container-name>.gpu-memory.limit`. |
| `nvidia.com/container.<container-name>.gpu-memory.limit` | Memory size in Kubernetes resource quantity format. For example, `80G`. | Sets the memory Limit for a request made as an absolute memory size. | Pairs with an absolute-size Request - `gpu-memory` or `nvidia.com/container.<container-name>.gpu-memory.request`. Pairing with a portion Request causes the Limit to be interpreted as the wrong kind of value. |
| `kai.scheduler/container.<container-name>.gpu-memory.portion.limit` | A portion of GPU memory as a double-precision floating-point number. For example, `0.8`. | Sets the memory Limit for a request made as a GPU memory portion. | Pairs with the portion Request - `gpu-fraction`. Pairing with an absolute-size Request causes the Limit to be interpreted as the wrong kind of value. |

<Note>

**Note**

* The Limit does not affect the Scheduler. The Scheduler places the workload according to its Request. The Limit governs only how far the workload can then grow at runtime.
* To set how a workload shares GPU compute, refer to the annotations described in [GPU compute modes](../../resource-optimization/compute-modes.md).

</Note>

#### Examples

* Requests an absolute amount of GPU memory with a Limit, on 4 GPU devices:

  ```yaml
  metadata:
    annotations:
      gpu-memory: "4000"   # MiB
      nvidia.com/container.app.gpu-memory.limit: "80G"
      # optional; 1 if omitted - number of fractional GPU devices
      gpu-fraction-num-devices: "4"
  spec:
    containers:
      - name: app          # the <container-name> used in the limit annotation
  ```

* Requests the same Request and Limit using the official NVIDIA-native annotation for the Request instead of `gpu-memory`. The container name moves into the key, and device count still uses the unprefixed `gpu-fraction-num-devices`:

  ```yaml
  metadata:
    annotations:
      nvidia.com/container.app.gpu-memory.request: "40G"
      nvidia.com/container.app.gpu-memory.limit: "80G"
      # optional; 1 if omitted - number of fractional GPU devices
      gpu-fraction-num-devices: "4"
  spec:
    containers:
      - name: app
  ```

* Requests a portion of GPU memory with a Limit, on 2 GPU devices:

  ```yaml
  metadata:
    annotations:
      gpu-fraction: "0.5"
      kai.scheduler/container.app.gpu-memory.portion.limit: "0.8"
      # optional; 1 if omitted - number of fractional GPU devices
      gpu-fraction-num-devices: "2"
  spec:
    containers:
      - name: app          # the <container-name> used in the limit annotation
  ```
