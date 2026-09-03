# Training & Workspace — phase + phaseMessage

**Engine resource (new flow):** `batch/v1 Job` → Pods (Composer-V2 `iw_tw_to_job.go`; no run.ai CRD).
**Sources:** Job `.status` (`Active/Ready/Succeeded/Failed/Initialized/Pending`, conditions `Complete`/`Failed`/`Suspended`) + Pods + `wlr.status.issues` (for cause). `Spec.Parallelism` + category (`Train`/`Build`) decide single- vs multi-pod wording.
**Reference (reproduce, don't import):** `workload-controller/pkg/status/aggregator/workload_status_calculator.go` (`handleWorkloadPhase`, `getInitializingReasonMessages`) and `internal/jobs/runaijob.go`.

## How we reach each phase

| Phase | Reached when | Source |
|---|---|---|
| **Creating** | Job exists, no pods started yet (no `Active`/`Initialized`/`Ready`, no terminal condition) — default | engine |
| **Pending** | pods exist but none initialized/ready (`GetPendingPods>0`) | engine + pods |
| **Initializing** | `Initialized>0` AND `Ready < MinimumReadyReplicas` (hardcoded Go constant `=1`, **not** read from Job `.spec`) | engine + pods |
| **Running** | `Ready>0` (≥ minimum) | engine + pods |
| **Completed** | Job condition `Complete`/`Succeeded` = True | engine |
| **Stopped** (Suspended) | Job condition `Suspended` = True | engine |
| **Failed** | Job condition `Failed` = True | engine |
| **Degraded** | running/initialized AND the wlr reports critical issues / a plugin failed | engine + **wlr** |

Notes: `Creating` vs `Pending` needs pods. `Suspending`/`Resuming` are v1 transients not reliably derivable from a Job → map to `Stopped`/`Initializing` (or omit).

## Phase → message → source → derivable

| Phase | phaseMessage | Source | Derivable? |
|---|---|---|---|
| Creating | "" | Job: no pods started yet | ✅ engine |
| Pending | "" | pods pending; none initialized | ✅ engine+pods |
| Initializing (single — Workspace / Training `parallelism=1`) | `"Workload is initializing, the pod is still not ready"` | `Ready<1 & Initialized>0`; category `Build` or parallelism=1 | ✅ engine+pods |
| Initializing (multi — Training) | `"Workload is initializing, minimum ready pods not reached, waiting for N pods to be ready"` (N = `MinimumReadyReplicas`, hardcoded constant `=1`) | `Ready<min & Initialized>0`; parallelism>1 | ✅ engine+pods |
| Running | "" | `Ready>0` | ✅ engine+pods |
| Completed | "" | Job cond `Complete`/Succeeded | ✅ engine |
| Stopped (Suspended) | "" | Job cond `Suspended` | ✅ engine |
| Failed (job-level) | `"job reached failed state"` (+ failed cond msg) | Job cond `Failed` | ✅ engine |
| **Failed (cause)** | `"Workload failed due to a <group> issue"` (plural: `"… due to <g1>, <g2> issues"`) | wlr `Issue.Group` (Compute/Storage/Network; `Internal`→generic fallback, no group word) | ⚠️ wlr |
| **Degraded** | `"Workload is in degraded state due to a <group> issue"` | wlr issues / plugins | ⚠️ wlr |

> **`K8sStatus` stays native.** The table above is `Phase` + `Message` (computed). The separate `observedStatus.K8sStatus` field keeps the **raw Job status** (`Active/Succeeded/Failed/Suspended/Creating`) — e.g. a `Degraded` workload is still `K8sStatus=Active`. Compute it from the Job (reuse `setJobObservedStatus`), never from the run:ai phase. See README §2 invariant.
