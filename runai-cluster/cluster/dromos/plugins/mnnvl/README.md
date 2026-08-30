# MNNVL Plugin

## Overview

The MNNVL Plugin enables Multi-Node NVLink (MNNVL) support for workloads by automatically mutating Pod specifications to include the necessary resource claims and scheduling constraints and also creates a ComputeDomain resource for the workload. This allows workloads to leverage NVIDIA's MNNVL DRA (Dynamic Resource Allocation) driver for GPU resources across multiple nodes with NVLink connectivity.

## How it Works

### Activation

The plugin is triggered when:
- `MnnvlMode` is set to `Enabled` in the WorkloadRunner specification
- The workload contains GPU resources
- The Karta exposes the MNNVL write-back paths: `containersPath` or `containerPath` or `resourceClaimsPath`, and `podAffinityPath`

### Resource Mutation

When activated, the plugin performs the following mutations:

1. **Compute Domain Creation**: Creates a ComputeDomain resource for the workload. If elastic CD is not enabled, the plugin calculates the number of nodes based on the number of pods with GPU resources and the scale definition of the workload using the RI scale definition.

1. **Resource Claim Injection**: Adds a `ResourceClaim` for containers that have gpu resources to the Pod spec - with the compute domain resource claim template.

1. **Pod Affinity Configuration**: The plugin adds preferred `PodAffinity` rules to encourage the scheduler to place a workload's pods on the same MNNVL clique (nodes with NVLink connectivity). The topology keys are the `GpuNetworkAccelerationLabelKeys` passed through the WorkloadRunner instructions. The affinity is resolved once per workload:
   - if a consumer supplied `Options.PodAffinityProvider`, the plugin uses its return value as-is (a nil return means no pod affinity - e.g. the consumer decided an external scheduler owns placement);
   - else the default applies: one preferred term per topology key selecting peers by `dromos.run.ai/mnnvl-workload-name=<runner.Name>`.

   The chosen affinity is merged into any user-authored pod affinity (idempotently, deduping by topology key), so user terms are preserved.

   The plugin stamps `dromos.run.ai/mnnvl-workload-name` by default, so its default affinity has a label to select. A consumer that supplies its own `PodAffinityProvider` (whose label is stamped elsewhere) turns the stamp off with `WithWorkloadLabelStampDisabled()`.

## Options (injection seam)

The plugin owns no host-specific label or scheduler conventions; `dromos.run.ai` is its own namespace. A consumer wires host-specific behavior through `mnnvl.Option` funcs passed to `NewBuilder`/`NewPlugin`:

- `WithPodAffinityProvider(func(runner, resource) *corev1.PodAffinity)`: use the return as the pod affinity instead of the default (e.g. to group by a different label, or to return nil when an external scheduler owns placement). The plugin applies the result verbatim.
- `WithWorkloadLabelStampDisabled()`: stop stamping `dromos.run.ai/mnnvl-workload-name` (stamped by default). Use it when `PodAffinityProvider` supplies an affinity whose label is stamped elsewhere.

## Karta Requirements

A `PodSpec` or `PodTemplateSpec` Karta exposes these targets directly. When the Karta uses a `FragmentedPodSpec`, it must define the following paths so the plugin can write its mutations back:

### Required for Resource Claims
- `containersPath` or `containerPath`
- `resourceClaimsPath`

### Required for Affinity Rules
- `podAffinityPath`: path to the pod affinity configuration

### Required for the workload-name label (stamped by default)
- `labelsPath`: where the plugin writes `dromos.run.ai/mnnvl-workload-name` (unless disabled with `WithWorkloadLabelStampDisabled()`)

If a required path is not defined for a `FragmentedPodSpec`, the plugin returns a validation error.
