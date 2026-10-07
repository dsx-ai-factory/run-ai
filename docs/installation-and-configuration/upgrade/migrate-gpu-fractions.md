# Migrate GPU Fractions

GPU fractions remain supported in v3.0, now powered by NVIDIA CUDA instead of the previous proprietary engine. This makes the capability part of the standard NVIDIA stack, with broader compatibility and long-term maintenance by NVIDIA.

The upgrade affects every cluster running fractional GPU workloads. The submission method determines whether the workload's request needs to be rewritten before it is resubmitted:

| Workload type | Anything else required |
| --- | --- |
| [Native workload types](../../workloads-in-nvidia-run-ai/workload-types/native-workloads.md) | No changes - the request and limit are still set through the same UI, API, and CLI fields. NVIDIA Run:ai converts the workload to the new annotation format automatically. Existing automation, pipelines, and saved templates continue to work unchanged. |
| [Supported workload types](../../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md), submitted as a Kubernetes manifest | The manifest itself requires a change only for dynamic fractions - a workload using a GPU fraction that declares a memory limit in addition to a request. A manifest that requests a fraction without a limit needs no change. See [Migrate Dynamic Fractions YAML](#migrate-dynamic-fractions-yaml). |

## Prerequisites

| Component | Required version | Notes |
| --- | --- | --- |
| CUDA | 13.4 or later | Delivers the CUDA-native GPU sharing implementation |
| CUDA driver | r615 or later | Must be installed on all GPU nodes |
| GPU Operator | 26.7.1 or later | Deploys the CUDA sharing components cluster-wide. Does not default to CUDA driver r615 - set the driver version explicitly |
| NRI plugin | Disabled (`cdi.nriPluginEnabled` set to `false`) | Set when installing or configuring the GPU Operator. GPU memory fractions do not function while the NRI plugin is enabled |

## Migrate Dynamic Fractions YAML

This section applies only to [supported workload types](../../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md) whose manifest sets a fraction limit. [Native workload types](../../workloads-in-nvidia-run-ai/workload-types/native-workloads.md) use the same UI, API, and CLI fields as before - NVIDIA Run:ai converts them to the new annotation format automatically.

In releases up to v2.26, a dynamic fraction was expressed in two places: the fraction request was an annotation, while the fraction limit was the container environment variable `RUNAI_GPU_MEMORY_LIMIT`. In v3.0, the limit becomes an annotation alongside the request.

<Warning>

**Warning**

Every supported workload type that sets `RUNAI_GPU_MEMORY_LIMIT` must be migrated. The environment variable is no longer honored in v3.0 - a workload that still uses it is treated as a request without a limit, silently losing its burst ceiling.

</Warning>

Fraction request annotations (`gpu-fraction`, `gpu-memory`, `gpu-fraction-num-devices`, `gpu-fraction-container-name`) are unchanged in v3.0 - see [Dynamic Fractions](../../platform-management/runai-scheduler/resource-optimization/dynamic-memory-fractions.md) for their definitions.

The fraction limit replaces the environment variable with one of two annotations, depending on whether the request uses an absolute size or a portion:

| Up to v2.26 | v3.0 | Value format |
| --- | --- | --- |
| `env: RUNAI_GPU_MEMORY_LIMIT` | `nvidia.com/container.<container>.gpu-memory.limit` | Absolute size in Kubernetes units, for example `"80G"` |
| `env: RUNAI_GPU_MEMORY_LIMIT` | `kai.scheduler/container.<container>.gpu-memory.portion.limit` | Portion, for example `"0.8"` |

Substitute `<container>` with the name of the container receiving the fraction - the same container named by `gpu-fraction-container-name`, or the first container in the pod if that annotation is omitted. Match the limit form to the request form: pair `gpu-memory` with the absolute limit annotation, and `gpu-fraction` with the portion limit annotation.

**Example - absolute memory size:**

```yaml
# Before (up to v2.26)
metadata:
  annotations:
    gpu-memory: "4000"            # Memory size in MiB
    gpu-fraction-num-devices: "2"
    gpu-fraction-container-name: "app"
spec:
  containers:
  - name: app
    env:
    - name: RUNAI_GPU_MEMORY_LIMIT   # REMOVE: no longer honored in v3.0
      value: "8000M"

# After (v3.0)
metadata:
  annotations:
    gpu-memory: "4000"                                    # Memory size in MiB
    nvidia.com/container.app.gpu-memory.limit: "8000M"     # replaces the env var
    gpu-fraction-num-devices: "2"
spec:
  containers:
  - name: app
```

**Example - portion of GPU memory:**

```yaml
# Before (up to v2.26)
metadata:
  annotations:
    gpu-fraction: "0.5"
    gpu-fraction-num-devices: "2"
    gpu-fraction-container-name: "app"
spec:
  containers:
  - name: app
    env:
    - name: RUNAI_GPU_MEMORY_LIMIT   # REMOVE: no longer honored in v3.0
      value: "0.8"

# After (v3.0)
metadata:
  annotations:
    gpu-fraction: "0.5"
    kai.scheduler/container.app.gpu-memory.portion.limit: "0.8"   # replaces the env var
    gpu-fraction-num-devices: "2"
spec:
  containers:
  - name: app
```

## Choose the Upgrade Sequence

The CUDA driver, GPU Operator, and NVIDIA Run:ai upgrades can happen in either order. The order chosen determines only how long fractions stay unavailable:

|  | Sequence | Fractions availability |
| --- | --- | --- |
| **Option A: GPU stack first, then Run:ai** | Upgrade the CUDA driver and GPU Operator first, then upgrade NVIDIA Run:ai. | Available a short while after the NVIDIA Run:ai upgrade. Recommended where a long fractions outage is unacceptable. |
| **Option B: Run:ai first, then GPU stack** | Upgrade NVIDIA Run:ai first, then upgrade the stack when ready. | Unavailable until the stack upgrade completes, then enabled automatically. Useful when the NVIDIA Run:ai upgrade and the node-level driver work must be scheduled separately. |

## Step-by-Step

**Step 1 - Suspend or delete running and pending workloads using GPU fractions**

* For [native workload types](../../workloads-in-nvidia-run-ai/workload-types/native-workloads.md), suspend the workload through the NVIDIA Run:ai UI, CLI, or API.
* For [supported workload types](../../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md) submitted as a Kubernetes manifest, delete the workload but retain its manifest.

No workload using a GPU fraction may be running or pending when the NVIDIA Run:ai upgrade begins.

**Step 2 - Migrate Dynamic Fractions YAML**

For supported workload types whose manifest sets a fraction limit, update those manifests now, before resubmission - see [Migrate Dynamic Fractions YAML](#migrate-dynamic-fractions-yaml) above. Native workload types keep the same UI, API, and CLI fields - NVIDIA Run:ai converts them to the new annotation format automatically. Doing this while the workload is already suspended or deleted avoids a second disruption.

**Step 3 - Upgrade the CUDA driver on all GPU nodes**

Upgrade to r615 or later (or r620 when available). This is the core dependency for CUDA-based sharing. Skip this step if you are following Option B.

**Step 4 - Upgrade the GPU Operator**

Upgrade to 26.7.1 or later. This deploys the CUDA sharing components and updates related node-level components across the cluster. In the GPU Operator configuration:

1. Set the driver version to r615 or later (`driver.version`). GPU Operator 26.7.1 does not default to CUDA driver r615, so do not rely on the default.
2. Set `cdi.nriPluginEnabled` to `false` (`--set cdi.nriPluginEnabled=false`, or `cdi.nriPluginEnabled: false` in the `ClusterPolicy` spec on OpenShift). GPU memory fractions do not function while the NRI plugin is enabled.

Skip this step if you are following Option B.

**Step 5 - Upgrade NVIDIA Run:ai**

Follow [Upgrade NVIDIA Run:ai](upgrade.md) to upgrade the control plane and cluster. If you completed Steps 3 and 4 first (Option A), GPU sharing is CUDA-based as soon as this upgrade completes.

**Step 6 - Complete the stack upgrade (Option B only)**

If you upgraded NVIDIA Run:ai first, upgrade the CUDA driver and GPU Operator as described in Steps 3 and 4. GPU sharing becomes CUDA-based once both are in place.

**Step 7 - Resubmit workloads using GPU fractions**

Native workload types cannot resume their original suspended workload on the new CUDA-based fractions stack:

1. Copy each suspended workload - using **Make a Copy** in the NVIDIA Run:ai UI, or by resubmitting the same configuration through the API or CLI.
2. Give the copy a different name so it can coexist with the original during verification. Submit it as-is, with the same fields and configuration - no changes are required.
3. Verify the replacement reaches `Running` and holds its fraction as expected.
4. Delete the original suspended workload once the replacement is verified.

Supported workload types were deleted, not suspended, in Step 1, so there is no original workload left to delete afterward:

1. Submit the migrated manifest from Step 2 through the NVIDIA Run:ai UI, API, or CLI.
2. Verify the workload reaches `Running` and holds its fraction limit as expected.

## Verify GPU Fractions

* Confirm all GPU nodes report CUDA driver r615 or later. A single node on an older driver accepts workloads not using GPU fractions but rejects workloads using GPU fractions, producing scheduling failures that are easy to misdiagnose. Verify this explicitly rather than assuming it from the GPU Operator version, since GPU Operator 26.7.1 does not default to r615.
* Confirm the GPU Operator is at 26.7.1 or later, that `cdi.nriPluginEnabled` is set to `false`, and that its pods are `Running` across the cluster (`kubectl get pods -n gpu-operator`).
* Confirm the platform reports v3.0.
* Submit a small test workload requesting a fraction and confirm it reaches `Running` rather than remaining `Pending`.
* For a migrated dynamic fraction workload, confirm the limit annotation is present on the pod and that the container is held to its ceiling - a missed limit annotation runs normally but without a burst ceiling, so verify explicitly rather than inferring from a successful start.
* Resubmit and confirm production workloads using GPU fractions, including any using SM-sharing.

## Troubleshoot GPU Fractions

| Symptom | Likely cause | Resolution |
| --- | --- | --- |
| Workloads using GPU fractions stay `Pending` after the upgrade; workloads not using GPU fractions schedule normally | A prerequisite is not yet satisfied, most often the CUDA driver on one or more nodes, or a GPU Operator still below 26.7.1 | Check the driver on every GPU node and the GPU Operator version - fractions enable automatically once both are correct |
| Only some workloads using GPU fractions are `Pending`; others run | Mixed driver versions across the node pool | Audit the driver version node by node rather than cluster-wide, and upgrade any node below r615 |
| GPU Operator is at 26.7.1 or later, but GPU nodes still report a driver older than r615 | GPU Operator 26.7.1 does not default to CUDA driver r615 | Set the driver version to r615 or later explicitly in the GPU Operator configuration (`driver.version`) - do not rely on the GPU Operator default |
| Workloads using GPU fractions do not function even though the GPU Operator and driver versions are correct | The NRI plugin is enabled in the GPU Operator (`cdi.nriPluginEnabled=true`) | Set `cdi.nriPluginEnabled` to `false` in the GPU Operator configuration and confirm the GPU Operator pods restart |
| A dynamic fraction workload runs but is not held to its limit | `RUNAI_GPU_MEMORY_LIMIT` is still in the manifest and was not replaced by the annotation; v3.0 ignores the variable | Migrate the manifest as described above. This failure is silent - audit manifests rather than waiting for a symptom |
| The limit annotation is present but has no effect | The container name in the annotation key does not match a container in the pod, or does not match `gpu-fraction-container-name` | Confirm the `<container>` element of the annotation key matches the container name exactly, including case |
| The limit is interpreted as the wrong kind of value | An absolute limit was paired with a portion request, or the reverse | Pair `gpu-memory` with the `nvidia.com/…gpu-memory.limit` key, and `gpu-fraction` with the `kai.scheduler/…portion.limit` key |
| Workloads fail to start after enabling SM-sharing | Insufficient SM capacity for the number of concurrent workloads using GPU fractions requested | Revert to time-slicing by removing the annotation, then reintroduce SM-sharing with fewer concurrent fractions |
| OpenShift cluster cannot install GPU Operator 26.7.1 | The OCP version is outside the GPU Operator 26.7.1 support matrix | Upgrade OCP to a supported version. NVIDIA Run:ai v3.0 can be installed in the meantime; fractions remain disabled until the stack is compatible |

## Optional: Configure GPU Compute Sharing Mode

NVIDIA Run:ai v3.0 introduces a choice of how workloads using GPU fractions share GPU compute, set per workload with the `nvidia.com/gpu-compute-sharing-mode` annotation. This is optional new functionality - you do not need to configure it to complete the upgrade, and omitting it preserves existing behavior.

Workloads migrated from v2.26 or below retain time-slicing behavior without any change. Adopting SM-sharing is an opt-in decision that can be made after the upgrade.

| Mode | Annotation value | Behavior | When to use |
| --- | --- | --- | --- |
| Time-slicing (default) | `"time-slicing"` | Workloads take sequential turns on the GPU compute engines | The default; fully isolated error domain, behaves as GPU sharing did before v3.0 |
| SM-sharing | `"sm-sharing"` | Workloads run concurrently on separate streaming multiprocessors (SMs) | Opt-in; uses a shared MPS context for all workloads sharing a GPU, with a shared error domain |
