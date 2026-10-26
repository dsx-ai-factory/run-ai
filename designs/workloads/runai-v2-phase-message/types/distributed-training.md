# Distributed Training — phase + phaseMessage

**Engine resource (new flow):** kubeflow `PyTorchJob` / `TFJob` / `MPIJob` / `XGBoostJob` / `JAXJob` → Pods.
**Sources:** kubeflow `JobStatus` (`ReplicaStatuses`, conditions `JobCreated`/`Running`/`Succeeded`/`Failed`/`Suspended`) + Pods + `wlr.status.issues` (for cause).
**Reference:** `workload-controller/internal/jobs/distributed_status.go` (`GetReadyPods`/`GetFailedPods`/`GetPendingPods`/`GetInitializingPods`/`IsInitializing`).

## How we reach each phase

| Phase | Reached when | Source |
|---|---|---|
| **Creating** | `ReplicaStatuses` empty AND not initializing/running/pending/finished | engine |
| **Pending** | `GetPendingPods>0` | engine + pods |
| **Initializing** | `GetInitializingPods>0` AND `GetReadyPods<1` | engine + pods |
| **Running** | `GetReadyPods>0` | engine + pods |
| **Completed** | condition `JobSucceeded` = True | engine |
| **Stopped** (Suspended) | condition `JobSuspended` = True | engine |
| **Failed** | condition `JobFailed` = True | engine |
| **Degraded** | running AND the wlr reports critical issues / a plugin failed | engine + **wlr** |

## Phase → message → source → derivable

| Phase | phaseMessage | Source | Derivable? |
|---|---|---|---|
| Creating | "" | `ReplicaStatuses` empty | ✅ engine |
| Pending | "" | pending pods>0 | ✅ engine+pods |
| Initializing | `"Workload is initializing, minimum ready pods not reached, waiting for N pods to be ready"` (N = minimum ready pods required to start = `1`, same hardcoded threshold as `GetReadyPods<1`; **not** total replicas) | init pods>0 & ready<1 | ✅ engine+pods |
| Running | "" | ready pods>0 | ✅ engine+pods |
| Completed | "" | cond `JobSucceeded` | ✅ engine |
| Stopped (Suspended) | "" | cond `JobSuspended` | ✅ engine |
| Failed (job-level) | failed cond msg / `"job reached failed state"` | cond `JobFailed` | ✅ engine |
| **Failed (cause)** | `"Workload failed due to a <group> issue"` (`Internal`→generic fallback, no group word) | wlr issues (`Group`: Compute/Storage/Network/Internal) | ⚠️ wlr |
| **Degraded** | `"Workload is in degraded state due to a <group> issue"` (`Internal`→generic fallback, no group word) | wlr issues (`Group`: Compute/Storage/Network/Internal) | ⚠️ wlr |
