# Storage Plugin

## Overview

The Storage Plugin reconciles `PersistentVolumeClaim` resources whose lifecycle is tied to a `WorkloadRunner`. 

The plugin does **not** mount the PVC into pods. Volume / `volumeMount` wiring is the workload template's responsibility. The plugin only materializes the PVC resources.

## How it Works

### Activation

The plugin is triggered when `spec.instructions.storage.newPvcs` contains at least one entry.

### PVC Reconciliation

For each PVC in `newPvcs`:

1. **Ownership check**: if a PVC with the same name already exists in the apiserver and is **not** controlled by this WorkloadRunner (different controller UID, or no controller ownerRef at all), the plugin emits a critical issue and marks itself `Failed`. No write is attempted.
1. **OwnerRef wiring**: The runner is set as the ownerRef on the PVC's.

### Status / Phase Tree

A phase derived from the worst-case state across all PVCs in `newPvcs`:

| Plugin Phase     | Condition | Trigger |
|------------------|-----------|---------|
| `Disabled`       | False     | No `newPvcs` declared |
| `Creating`       | False     | At least one PVC has not been created yet |
| `Initializing`   | False     | All PVCs exist but at least one is not `Bound` (e.g. `Pending`) |
| `Ready`          | True      | All PVCs are `Bound` |
| `Failed`         | False     | At least one PVC is `Lost`, or a critical issue was emitted (e.g. ownership conflict) |

Priority on mixed states: `Lost` > `NotCreated` > `NotBound` > `Bound`.

### Watcher

The plugin registers `Owns(&corev1.PersistentVolumeClaim{})`.

## Configuration

```yaml
spec:
  instructions:
    storage:
      newPvcs:
        - metadata:
            name: training-data
            namespace: my-namespace    # optional; must equal WorkloadRunner.namespace if set
          spec:
            accessModes:
              - ReadWriteOnce
            storageClassName: standard
            resources:
              requests:
                storage: 50Gi
```

The `metadata` and `spec` fields are the standard `corev1.PersistentVolumeClaim` schema; any field the apiserver accepts is passed through unchanged.

## Validation Rules

- Enforced by the plugin's webhook validator. 
- Storage-class existence and size caps are not validated by the plugin; the apiserver and PV provisioner enforce those at PVC creation time.
