# NodeAffinity Plugin

## Overview

The NodeAffinity Plugin merges the node affinity declared in `spec.instructions.nodeAffinity` into the pod affinity of every component of the rendered child. It is how an external controller constrains a workload's scheduling without knowing its framework: the caller writes one standard `corev1.NodeAffinity` on the WorkloadRunner, and the plugin lands it on every pod spec the workload's Karta exposes.

The primary consumer is the restart flow: the restart reconciler patches node exclusions (`kubernetes.io/hostname NotIn [...]`) into the instruction, and the recreated child comes back scheduled away from the faulty nodes.

```yaml
kind: WorkloadRunner
spec:
  instructions:
    nodeAffinity:
      nodeAffinity:                # standard corev1.NodeAffinity
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
          - matchExpressions:
            - key: kubernetes.io/hostname
              operator: NotIn
              values: [node-a, node-b]
```

## How it Works

### Activation

The plugin is triggered when:
- `NODE_AFFINITY_ENABLED` is not set to `false` (enabled by default)
- `spec.instructions.nodeAffinity.nodeAffinity` is set and carries at least one term (an absent or term-less instruction is a no-op, which is also how an applied instruction is removed)

### Resource Mutation (`ResourceMutatePlugin`)

The instruction is normalized (requirement values sorted and de-duplicated) and merged into each pod-bearing component's node affinity:

- **Required** terms are ANDed with the workload's own selector via a cartesian product of terms. A containment short-circuit keeps the merge idempotent, so the framework's multi-round mutation loop converges instead of squaring the term count on every round.
- **Preferred** terms are appended, skipping terms already present (values compared as sets - the scheduler sums the weights of matching terms, so a duplicate would silently double the preference).
- The workload's own affinity is never replaced, only merged into.

The merge distinguishes the three no-term shapes the way Kubernetes does: a `nil` selector and one with zero terms place no constraint, while a selector whose every term is empty matches no nodes and wins the merge outright.

### Validation (`ValidatorPlugin`)

On any WorkloadRunner write that carries the instruction, the webhook rejects what the cluster could not act on - failing here means a restart patch dies **before** the running child is deleted (the dry-run equivalent):

- structural rules matching what the apiserver enforces on a pod's node affinity (operator/values consistency, `matchFields` restricted to `metadata.name` with one In/NotIn DNS-name value). The scheduler's stricter selector compiler is deliberately not used: the apiserver admits label values that fail label validation, so the instruction must too.
- empty node selector terms are rejected outright (they match no nodes and would otherwise suppress the instruction)
- ceilings the CRD cannot express on an embedded corev1 type: at most 16 terms per side and 256 values per requirement (the cartesian merge multiplies term counts), preferred weights in 1-100
- the template's Karta must be able to carry node affinity (see below); a missing Karta is not an admission error - the controller reports it on status

### Status (`StatusPlugin`)

`status.plugins["node-affinity"]` publishes the normalized applied affinity in `additionalInfo` with a `NodeAffinityApplied` condition, so the constraint a workload runs under is visible without reading the child. The info is cleared when the instruction is removed or was not applied.

## Karta Requirements

Every pod-bearing component of the workload's Karta must expose a place to write node affinity:

- a `PodSpecPath` or `PodTemplateSpecPath` spec definition works as-is
- a `FragmentedPodSpecDefinition` must define `nodeAffinityPath`

A Karta with no pod-bearing component at all, or a fragmented component missing `nodeAffinityPath`, fails validation at admission and is reported as a critical plugin issue at reconcile time.
