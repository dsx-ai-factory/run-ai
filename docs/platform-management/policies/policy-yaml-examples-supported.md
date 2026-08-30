# Policy YAML Examples

This page provides YAML examples for supported workload type policies. These policies use a flat `rules` array where each rule specifies the field to control, the spec selector that identifies which part of the workload spec the rule applies to, and one or more enforcements.

For a description of all rule components, see [Supported workload type policies](supported-workload-type-policies.md).

## Rule Structure

```yaml
rules:
  - field: <field-name>
    specSelector: <spec-selector>
    itemSelector: <item-selector>    # optional; only for list-type fields
    selectorParams:                   # optional; required by some selectors
      <param>: <value>
    enforcements:
      - <enforcement-type>: <value>
```

## Deployment: Container Security Baseline

Requires an image on all containers, blocks privilege escalation, and caps CPU and memory limits.

```yaml
rules:
  - field: image
    specSelector: allContainers
    enforcements:
      - required: true
  - field: allowPrivilegeEscalation
    specSelector: allContainers
    enforcements:
      - set:
          value: false
  - field: resourceLimits
    specSelector: allContainers
    enforcements:
      - set:
          key: cpu
          value: "2"
      - set:
          key: memory
          value: "4Gi"
```

## PyTorchJob: Distributed Training Guardrails

Sets resource limits across all containers, pins the master replica count to 1, restricts the worker replica range, and limits retries on the job.

```yaml
rules:
  - field: resourceLimits
    specSelector: allContainers
    enforcements:
      - set:
          key: cpu
          value: "4"
      - set:
          key: memory
          value: "8Gi"
  - field: replicas
    specSelector: masterReplicaSpec
    enforcements:
      - set:
          value: 1
  - field: replicas
    specSelector: workerReplicaSpec
    enforcements:
      - range:
          min: 1
          max: 8
  - field: runPolicy-BackoffLimit
    specSelector: workloadSpec
    enforcements:
      - set:
          value: 3
```
