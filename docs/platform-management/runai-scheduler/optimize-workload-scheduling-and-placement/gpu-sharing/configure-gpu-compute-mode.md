# Configure GPU Compute Mode

Set how a workload shares GPU compute with the other workloads on the same GPU. For how the modes work and how to choose between them, refer to [GPU compute modes](../../resource-optimization/compute-modes.md).

The compute mode is set per workload with the `nvidia.com/container.<container-name>.gpu-compute.mode` annotation, whether you submit through the UI, the CLI, the API, or directly in YAML. Set this in addition to the [GPU memory fraction](../../resource-optimization/memory-fractions.md) or [dynamic GPU memory fraction](../../resource-optimization/dynamic-memory-fractions.md) request - the compute mode has no effect without a memory fraction already requested on the same container. After the workload is submitted, the compute mode is visible in the workload's [YAML tab](../../../../workloads-in-nvidia-run-ai/workloads.md#yaml). For a full step-by-step walkthrough of each interface, refer to [Launching Workloads with GPU Compute Modes](../../resource-optimization/quick-starts/compute-modes-quickstart.md).

## Annotation Reference

Substitute `<container-name>` with the name of the container receiving the GPU memory fraction. A GPU fraction is always allocated to a single container, and NVIDIA Run:ai currently supports one fractional container per pod. This is the same container named by `gpu-fraction-container-name`, or the first container in the pod when that annotation is omitted. If both are used, the container names must be identical or the request is rejected.

| Value | Behavior |
|---|---|
| `time-slicing` | Workloads take sequential turns on the GPU compute engines. This is the default and can be omitted. |
| `sm-sharing` | Workloads run concurrently on separate streaming multiprocessors. |

If the annotation is omitted, time-slicing applies.

This annotation applies the same way to [supported workload types](../../../../workloads-in-nvidia-run-ai/supported-workload-types/list-of-supported-workload-types.md) submitted via YAML and to [workloads submitted externally](../../../../workloads-in-nvidia-run-ai/workload-types/supported-features.md#externally-submitted-kubernetes-workloads), outside of NVIDIA Run:ai.

## Example

The following example requests an absolute amount of GPU memory and enables SM-sharing:

```yaml
metadata:
  annotations:
    gpu-memory: "20000"   # MiB
    # "time-slicing" (the default, may be omitted) or "sm-sharing"
    nvidia.com/container.app.gpu-compute.mode: "sm-sharing"
spec:
  containers:
    - name: app          # the <container-name> used in the mode annotation
```
