# Detailed design — Runtime-Rules support for WorkloadRunner

This document complements [design.md](./design.md) with file paths, line references, and concrete code shapes. Read the high-level doc first.

## Table of Contents
1. [Existing Runtime-rules Internals](#1-existing-runtime-rules-internals)
2. [Kind → Category Mapping](#2-kind--category-mapping)
3. [WorkloadTask Schema](#3-workloadtask-schema)
4. [Watch Registration](#4-watch-registration)
5. [Conversion: WorkloadRunner → JobWorkload](#5-conversion-workloadrunner--jobworkload)
6. [Fetch and Update Dispatch](#6-fetch-and-update-dispatch)
7. [Suspend Implementation Detail (per-Kind suspend field table)](#7-suspend-implementation-detail-per-kind-suspend-field-table)
8. [Workload Identity](#8-workload-identity)
9. [Idle Detection](#9-idle-detection)
10. [Auto-Delete-After-Completion Extension](#10-auto-delete-after-completion-extension)
11. [AutoDeletionTimeAfterCompletionSeconds Source — Attached Resource Annotation](#11-autodeletiontimeaftercompletionseconds-source--attached-resource-annotation)
12. [RBAC](#12-rbac)
13. [Tests](#13-tests)
14. [Rollout and Feature Flag](#14-rollout-and-feature-flag)
15. [Critical Files](#15-critical-files)
16. [Verification](#16-verification)
17. [Risks and Open Questions](#17-risks-and-open-questions)

---

## 1. Existing Runtime-rules Internals

The runtime-rules is a switch-driven controller, **not** polymorphic per workload kind. Key control points:

- **Entry**: `runai-cluster/cluster/runtime-rules/runtime_rules.go:90-118` registers three `controllers.NewWorkloadController(..., Kind)` watches.
- **Per-kind controller**: `pkg/controllers/workload_controller.go` — generic, parametrized by Kind.
- **Event handler**: `pkg/handlers/workload_handler.go` converts a CRD event into a `WorkloadTask` (in `pkg/queuesManager/priority_queue.go:21`, `WorkloadType` is a string field).
- **Switch points keyed off `WorkloadTask.WorkloadType`**:
  - `pkg/queuesManager/priority_queue_utils.go:126-150` — `fetchWorkload`.
  - `pkg/queuesManager/priority_queue_utils.go:152-170` — `updateWorkload` (today dispatches to `UpdateTW` / `UpdateIW` / `UpdateDW`).
  - `pkg/queuesManager/priority_queue_utils.go:315-334` — `getLimitsByWorkloadType` picks per-category time/idle knobs from `ProjectSpec`.
- **Deactivation**: `pkg/queuesManager/priority_queue_utils.go:49-96` sets `jw.Spec.Active = false` then calls `updateWorkload`.
- **Idle detection**: `pkg/queuesManager/priority_queue_utils.go:287-308` — Prometheus query keyed by workload name + id, gated on `WorkloadRequestedGpu`.
- **Conversion**: `pkg/util/functions.go:127+` — `ConvertToJobWorkload` unifies the three legacy CRDs into a `JobWorkload`.
- **Tests**: Ginkgo + envtest under `test/`.

## 2. Kind → Category Mapping

New file `pkg/util/category_mapping.go`. Allowlist matches the kinds the runtime-rules already supports — nothing more:

| Wrapped Kind | Inferred `WorkloadType` |
|---|---|
| `batch/v1.Job` | `KindTrainingWorkload` **OR** `KindInteractiveWorkload` — disambiguated by the category label on the runner (since `Job` backs both legacy types in the new path). |
| `MPIJob`, `PyTorchJob`, `TFJob`, `XGBoostJob`, `JAXJob` | `KindDistributedWorkload` |

Anything else → ignored (return `ok=false`).

Exported function:

```go
func InferCategory(runner *v1alpha1.WorkloadRunner) (category string, ok bool)
```

Algorithm:
1. Decode `Spec.Template.Raw` enough to read the wrapped Kind.
2. Distributed kinds → return `KindDistributedWorkload, true`.
3. Wrapped Kind `batch/v1.Job` → read `runner.Labels[<category label>]` (placeholder constant — see §17 open question; expected values: `training` / `interactive` or `workspace`). Map to `KindTrainingWorkload` or `KindInteractiveWorkload`. If absent or unknown → `ok=false`.
4. Any other Kind → `ok=false`.

The handler skips enqueue when `ok=false`.

## 3. WorkloadTask Schema

`pkg/queuesManager/priority_queue.go:14-27` — add one field:

```go
SourceKind string // "" or constants.KindWorkloadRunner
```

`WorkloadType` continues to be the inferred *category* string (`KindTrainingWorkload` / `KindInteractiveWorkload` / `KindDistributedWorkload`) so `getLimitsByWorkloadType` (`priority_queue_utils.go:315-334`) keeps working untouched. `SourceKind` only routes fetch / update dispatch.

Add `KindWorkloadRunner = v1alpha1.WorkloadRunnerKind` to `runai-cluster/cluster/common/constants/workload_constants.go`.

## 4. Watch Registration

- `cmd/main.go`: add `runnerv1alpha1.AddToScheme(scheme)` and parse `DISABLE_WORKLOAD_RUNNER_RULES` (bool, default `false`).
- `runtime_rules.go:90-118`: unless the feature is disabled, register a fourth controller:

  ```go
  if !cfg.DisableWorkloadRunnerRules {
      workloadRunnerController := controllers.NewWorkloadController(
          wo.K8sClient, workloadHandler, constants.KindWorkloadRunner)
      go workloadRunnerController.Start(ctxWithCancel, stopper)
  }
  ```
- `pkg/controllers/workload_controller.go:84-95` `getObjectList`: return `&v1alpha1.WorkloadRunnerList{}` for the new kind.
- Extend `PushProjectWorkloadsToProjectQueue` (`priority_queue_utils.go:176`) to also list WorkloadRunners (`kubeConfig.ListWLR`) and push each through the new conversion branch.

## 5. Conversion: WorkloadRunner → JobWorkload

`pkg/util/functions.go:127+` `ConvertToJobWorkload`: add a `*v1alpha1.WorkloadRunner` branch. Change its signature to also return `sourceKind string` so the handler can stamp it onto the `WorkloadTask`. Because the new branch needs to read the attached resource (for GPU + auto-delete annotation), the helper takes a `ctx context.Context` and a `client.Client`:

```go
func CopyRunnerToJobWorkload(
    ctx context.Context,
    k8sClient client.Client,
    runner *v1alpha1.WorkloadRunner,
    jw *jobs.JobWorkload,
) error
```

Field mapping:

| `JobWorkload` field | Source | Notes |
|---|---|---|
| `Kind` | inferred category (§2) | `KindTrainingWorkload` / `KindInteractiveWorkload` / `KindDistributedWorkload`. |
| `Name`, `Namespace`, `UID`, `Labels`, `Annotations` | copied from the `WorkloadRunner` | Standard metadata. |
| `Status.StartTime` | parse RFC3339 from `runner.Annotations["runai/start-time"]`; fall back to `runner.CreationTimestamp` if absent | Maintained by the runtime-rules itself (see §5a below) — stamped on `Running`/`Degraded`, cleared on the runtime-rules' own suspend action and on `Completed`/`Failed`. Used to key `WorkloadCache` so resume creates a fresh cache entry — see design §3.1. |
| `Status.CompletionTime` | parse RFC3339 from `runner.Annotations["runai/completion-time"]` | Maintained by the runtime-rules itself (see §5a below) — stamped on `Completed`/`Failed`, cleared on `Running`/`Degraded`. **Not** stamped on `Stopped` or `Suspended`. |
| `Spec.Active` | derived from `runner.Status.Phase` | Terminal phases (`Stopped` / `Failed` / `Completed` / `Suspended` / `Unknown`) → inactive (skip enforcement). Active phases (`Creating` / `Pending` / `Initializing` / `Running` / `Degraded`) → active. |
| `Spec.AutoDeletionTimeAfterCompletionSeconds` | parse int from the annotation on the **attached resource** (key: `runai/auto-deletion-time-after-completion-seconds`), fetched via the GVK + Name + Namespace published on `WorkloadRunner.Spec` | Stamped by the **control-plane** (not the runtime-rules). Missing or invalid → `nil` (no auto-delete, parity with legacy). See §11. |
| `Status.OwnedResources` | synthesized — one entry with `Kind` = wrapped Template Kind | Lets the existing `GetWorkloadValues` allowlist check (`util/functions.go:46-52`) pass without changes. |

### 5a. Annotation lifecycle (runtime-rules owns `runai/start-time` and `runai/completion-time`)

A new helper in the workload event handler runs on every observed `WorkloadRunner` event before the enqueue / dispatch step:

```go
func (h *workloadHandler) reconcileTimestampAnnotations(
    ctx context.Context, runner *v1alpha1.WorkloadRunner,
) error
```

State machine (matches design §3.1):

| Observed phase | Action |
|---|---|
| `Running` or `Degraded` | If `runai/start-time` is absent, patch it to `time.Now().Format(time.RFC3339)`. If `runai/completion-time` is present, patch it out. |
| `Completed` or `Failed` | If `runai/completion-time` is absent, patch it to `time.Now().Format(time.RFC3339)`. If `runai/start-time` is present, patch it out. |
| `Suspended` / `Stopped` / `Unknown` / `Creating` / `Pending` / `Initializing` | no-op (the runtime-rules' own suspend action clears `runai/start-time` at the moment it patches the attached resource — that's the only timestamp write outside this helper). |

Use a single JSON-patch per reconcile (combined add + remove ops) to avoid double-writes. RFC3339 was picked over Unix seconds for human readability in `kubectl describe`.

### GPU detection — read the attached resource's pod template

The runtime-rules fetches the attached resource (GVK + Name + Namespace from `WorkloadRunner.Spec` — same reference `anyworkload-controller` uses) and runs the existing legacy-CRD GPU check on its pod template. No per-Kind branch in the runtime-rules, no dependency on a status condition published by `anyworkload-controller`.

New helper `util.AttachedResourceHasRequestedGpu(ctx, k8sClient, runner) (bool, error)`:
1. Resolve the attached resource's GVK + Name + Namespace from `runner.Spec`.
2. Fetch via `unstructured.Unstructured` + `client.Get`.
3. Walk to the pod template (`spec.template.spec.containers` for `batch/v1.Job`; `spec.<role>.template.spec.containers` for Kubeflow). The walk uses the same allowlist of paths the runtime-rules already maintains for legacy CRDs — no new per-Kind code.
4. Return true if any container requests `nvidia.com/gpu` (or the configured GPU resource name).

### Phase-aware skip in the handler

- **v1→v2 cast annotation absent** → skip outright. This is the scope gate per design §3 point 1. Increment `runai_workloadrunner_not_v1_cast_skipped_total`.
- If `runner.Status.Phase` ∈ {`Stopped`, `Failed`, `Completed`, `Suspended`, `Unknown`} → do not enqueue for time-limit/idle enforcement (the auto-delete controller has its own phase guard, §10).
- If `InferCategory` returns `ok=false` → skip; increment `runai_workloadrunner_unknown_category_total`.
- If the `workloadId` annotation is missing → log and skip (still useful for downstream metric correlation even though suspend no longer needs it).

## 6. Fetch and Update Dispatch

### `fetchWorkload`
`priority_queue_utils.go:126-150` — branch on `SourceKind`:

```go
if workloadTask.SourceKind == constants.KindWorkloadRunner {
    runner, err := kubeConfig.GetWLR(ctx, k8sClient, ns, name)
    if err != nil { return jw, err }
    if err := util.CopyRunnerToJobWorkload(ctx, k8sClient, &runner, jw); err != nil {
        return jw, err
    }
    return jw, nil
}
// existing switch unchanged
```

Add `kubeConfig.GetWLR` / `ListWLR` to `pkg/kubeConfig/` mirroring `GetTW` / `ListTW`.

### `updateWorkload` (suspend dispatch)
`priority_queue_utils.go:152-170` — branch on `SourceKind`:

```go
if workloadTask.SourceKind == constants.KindWorkloadRunner {
    return suspendAttachedResource(ctx, k8sClient, workloadTask)
}
// existing switch (UpdateTW / UpdateIW / UpdateDW) unchanged
```

`suspendAttachedResource` is a new helper in `pkg/queuesManager/suspend_attached.go`:

1. Fetch the `WorkloadRunner` (already in cache via the watch).
2. Patch the WorkloadRunner: clear `runai/start-time` annotation (closes the suspend ↔ resume loop, design §3.1).
3. Resolve GVK + Name + Namespace of the attached resource from `runner.Spec`.
4. Fetch the attached resource via `unstructured.Unstructured`.
5. Patch the suspend field per Kind:
    | Kind | Patch |
    |---|---|
    | `batch/v1.Job` | `spec.suspend = true` |
    | `kubeflow.org/v1.MPIJob` / `PyTorchJob` / `TFJob` / `XGBoostJob` / `JAXJob` | `spec.runPolicy.suspend = true` |
6. Use a server-side apply or strategic-merge patch to avoid clobbering unrelated spec fields.

Error policy:
- 404 on the attached resource → treat as success (already gone; nothing to suspend).
- Conflict (HTTP 409) → re-queue at the next tick; the priority queue already handles transient retries.
- Forbidden → return error; the operator/admin needs to grant RBAC.

Refactor the deactivate block (`priority_queue_utils.go:57-74`) into a small `deactivateWorkload(workloadTask, jw)` helper that skips the `jw.Spec.Active = false` mutation when `SourceKind == KindWorkloadRunner` — the in-memory field is meaningless for runners; the suspend signal flows through the attached resource.

## 7. Suspend implementation detail (per-Kind suspend field table)

(Section 7 previously hosted the control-plane HTTP client. That client is removed — suspend is now a K8s patch, see §6 above. The per-Kind suspend-field table lives here for reference.)

| GVK | Suspend field path | Type |
|---|---|---|
| `batch/v1.Job` | `spec.suspend` | `bool` |
| `kubeflow.org/v1.MPIJob` | `spec.runPolicy.suspend` | `bool` |
| `kubeflow.org/v1.PyTorchJob` | `spec.runPolicy.suspend` | `bool` |
| `kubeflow.org/v1.TFJob` | `spec.runPolicy.suspend` | `bool` |
| `kubeflow.org/v1.XGBoostJob` | `spec.runPolicy.suspend` | `bool` |
| `kubeflow.org/v1.JAXJob` | `spec.runPolicy.suspend` | `bool` |

Anything outside this allowlist returns "unsupported Kind" and the runner is skipped (matches the §2 category allowlist). The allowlist is a single constant table in `pkg/queuesManager/suspend_attached.go` — one place to update if a new Kind comes in scope.

Env config additions (parsed in `cmd/main.go`):

| Variable | Default | Purpose |
|---|---|---|
| `ENABLE_WORKLOAD_RUNNER_RULES` | `false` | Master feature flag for everything in this design. |

No control-plane HTTP client, no OAuth2 credentials, no v2 endpoint. The deployment manifest gets only the new feature-flag env var.

## 8. Workload Identity

The runtime-rules no longer needs the control-plane UUID to suspend a runner (suspend is a K8s patch on the attached resource — no API call to identify by ID). It still reads `util.GetWorkloadID` (`util/functions.go:154-162`) for metric labels and log correlation. Missing `workloadId` annotation → log a warning but proceed; this is a relaxation from the prior draft where it was a hard skip.

## 9. Idle Detection

No code change required. The Prometheus query `prometheusQuerier.GetWorkloadIdleDuration(name, id)` keys on `workload_id` / `workloadName` labels, which the control-plane already stamps onto pods owned by WorkloadRunners. Verify in tests (§13b).

## 10. Auto-Delete-After-Completion Extension

The auto-delete loop lives at `pkg/controllers/autodeletecompleted_controller.go` + `pkg/handlers/autodeletecompleted_handler.go` and is structurally independent of the time-limit queue. Today it:

- Watches `TrainingWorkloadList`, `InteractiveWorkloadList`, `DistributedWorkloadList`.
- Calls `ConvertToJobWorkload` on the event payload to get a generic `JobWorkload`.
- Considers a workload eligible when `Status.CompletionTime != nil && Spec.AutoDeletionTimeAfterCompletionSeconds != nil`.
- Sleeps until `CompletionTime + AutoDeletionTimeAfterCompletionSeconds`, then `client.Delete(ctx, workload)`.

Extension for WorkloadRunner — minimal, since the handler is already field-driven on the unified `JobWorkload`:

1. **Watch addition** (`pkg/controllers/autodeletecompleted_controller.go:41-58` `initWorkloadWatchers`): behind the same `ENABLE_WORKLOAD_RUNNER_RULES` flag, also watch `WorkloadRunnerList`.
2. **Scope gate** (`handleWorkloadEvent`): skip outright if the runner does not carry the v1→v2 cast annotation (design §3 point 1).
3. **Kind dispatch** (`handleWorkloadEvent` lines 114-124 and `getWorkload` / `getDeletionTimestamp` switches 83-114): add a `KindWorkloadRunner` branch routed through `ConvertToJobWorkload`. The conversion fetches the attached resource to read the `runai/auto-deletion-time-after-completion-seconds` annotation (see §11).
4. **No handler logic change**: eligibility check and elapsed-time arithmetic stay the same — both fields are populated for runners by the conversion helper (§5).
5. **Deletion**: call `client.Delete(ctx, runner)` directly on the `WorkloadRunner` object — no control-plane API call, no Karta. `anyworkload-controller` cascades cleanup through `ownerReferences` on the attached resource.
6. **Phase guard**: only treat a runner as "completed" when `Status.Phase ∈ {Completed, Failed}`. `Stopped` and `Suspended` are terminal-for-enforcement but do **not** trigger auto-delete — `Stopped` is typically user- or operator-initiated and the workload may still be resumed or inspected; `Suspended` may be resumed.
7. **Completion-time source**: read `runai/completion-time` from the `WorkloadRunner` annotation that the runtime-rules itself maintains (see §5a). On the very first observation of `Completed`/`Failed`, the annotation-lifecycle helper stamps the annotation before the auto-delete handler reads it (both run in the same handler chain).

## 11. AutoDeletionTimeAfterCompletionSeconds Source — Attached Resource Annotation

Decided: the **control-plane** stamps the value as an annotation on the **attached resource** (not on the `WorkloadRunner`). This matches how the GPU-presence signal flows (both pieces of submission metadata live on the attached resource at submission time).

- Control-plane stamps `runai/auto-deletion-time-after-completion-seconds: "<seconds>"` on the attached resource (`batch/v1.Job` / `MPIJob` / `PyTorchJob` / etc.) during submission, in whichever service materializes the attached resource.
- Runtime-rules fetches the attached resource (GVK + Name + Namespace from `WorkloadRunner.Spec`), reads `attached.Annotations[runai/auto-deletion-time-after-completion-seconds]`, parses to int32. Missing or invalid → `nil` (no auto-delete, parity with legacy when the field is unset).
- **Files touched**:
  - control-plane: whichever service stamps the attached-resource manifest (out of scope for the runtime-rules change; the control-plane team owns this).
  - `runai-cluster/common/constants/constants.go` — annotation key constant (shared between control-plane and runtime-rules).
  - `runtime-rules/pkg/util/functions.go` — read in `CopyRunnerToJobWorkload` (see §5).

Why the attached resource and not the `WorkloadRunner`: this annotation already conceptually belongs to the submitted workload manifest, alongside the GPU resource request. Putting it on the attached resource keeps "submission-time metadata" and "runtime-lifecycle metadata" (start/completion times, owned by the runtime-rules) on separate carriers. The auto-delete handler logic is unchanged either way.

## 12. RBAC

Runtime-rules ClusterRole additions:

- `get / list / watch / patch` on `run.ai/v1alpha1/workloadrunners` (patch for annotation lifecycle).
- `delete` on `run.ai/v1alpha1/workloadrunners` (auto-delete path).
- `get / patch` on `batch/v1/jobs` (suspend the attached resource for `Job`-wrapped runners).
- `get / patch` on `kubeflow.org/v1/mpijobs`, `pytorchjobs`, `tfjobs`, `xgboostjobs`, `jaxjobs` (suspend the attached resource for distributed-wrapped runners).

## 13. Tests

New file `test/workload_runner_test.go` (Ginkgo + envtest). No fake HTTP client needed — suspend is observable as a K8s patch on the attached resource. Use envtest's in-process API server to assert both the WorkloadRunner annotation lifecycle and the attached resource's suspend field.

Time-limit / idle / suspend tests:

| # | Scenario | Expectation |
|---|---|---|
| a | WorkloadRunner wrapping `batch/v1.Job` with v1-cast annotation, category=`training`, attached Job's `creationTimestamp` past `TrainingJobTimeLimitSecs`. | Attached `Job.Spec.Suspend == true`; runner's `runai/start-time` annotation cleared; no patch to the runner spec. |
| b | WorkloadRunner wrapping `batch/v1.Job` with v1-cast, category=`interactive`, fake Prometheus reports idle > `InteractiveJobMaxIdleDurationSecs`. | Attached `Job.Spec.Suspend == true`; suspend dispatched with `WorkloadType=KindInteractiveWorkload`. |
| c | WorkloadRunner wrapping `MPIJob` with v1-cast, past `TrainingJobTimeLimitSecs`. | Attached `MPIJob.Spec.RunPolicy.Suspend == true`. |
| d | WorkloadRunner wrapping `Deployment` (out of scope Kind) with v1-cast. | No task enqueued; attached resource untouched. |
| e | WorkloadRunner wrapping `batch/v1.Job` with v1-cast, missing or unknown category label. | No task enqueued. |
| f | **No v1-cast annotation** on a `batch/v1.Job`-backed runner that would otherwise be governed. | No task enqueued; scope-gate counter incremented. |
| g | K8s patch on attached resource returns 409 once then 200. | Retry succeeds at the next queue tick. |
| h | Coexistence: one `TrainingWorkload` and one v1-cast runner-wrapped `batch/v1.Job` (training) in the same project. | Legacy → `Spec.Active=false` patch; runner → attached `Job.Spec.Suspend=true` patch. Both fire under the same project limits. |
| i | Suspend → Resume loop. Runtime-rules suspends; test sets `Job.Spec.Suspend=false` and re-marks Phase=Running. | Runtime-rules stamps a fresh `runai/start-time` annotation; cache key flips; the new run gets the full allowance and is not immediately re-suspended. |
| j | Annotation lifecycle: phase Running → Failed → Running (resume after failure). | Sequence of annotation states matches §5a table on every transition. |

Auto-delete tests (extensions to `test/autodeletecompleted_test.go`):

| # | Scenario | Expectation |
|---|---|---|
| k | Runner with v1-cast and attached resource carrying `AutoDeletionTimeAfterCompletionSeconds=0`, reaching `Phase=Completed`. | `WorkloadRunner` deleted immediately via `client.Delete`. |
| l | Same with value `> 0`. | Deleted after the delay. |
| m | Attached resource annotation unset. | Never deleted (parity with legacy). |
| n | Runner in `Phase=Suspended`. | NOT deleted (suspend ≠ completion; may be resumed). |
| o | Runner reaching `Phase=Failed`. | Deleted after the delay. |
| p | Runner reaching `Phase=Stopped`. | NOT deleted (stop ≠ completion). |
| q | Runner without v1-cast annotation reaching `Phase=Completed`. | NOT deleted (scope gate). |

## 14. Rollout and Feature Flag

- Env flag `ENABLE_WORKLOAD_RUNNER_RULES` (default `false`) gates:
  - The fourth time-limit watch.
  - The fourth auto-delete watch.
  - Inclusion of WorkloadRunners in `PushProjectWorkloadsToProjectQueue`.
- Without the flag, behavior is byte-identical to today.
- **Per-workload scope gate** (orthogonal to the cluster-level flag): only runners carrying the v1→v2 cast annotation are governed. This is the design's intentional restriction — runners born directly as v2 are not in scope for this change.
- **Helm**: in `cluster/charts/runai-cluster/templates/runtime-rules/deployment.yaml` add the `ENABLE_WORKLOAD_RUNNER_RULES` env var. Update the ClusterRole (see §12).
- **Deprecation**: keep legacy CRD watches indefinitely — `WorkloadRunner` does not subsume them at submission time. The legacy paths can be retired when `EnableWorkloadRunner` becomes the only submission path (a separate effort).

## 15. Critical Files

### Runtime-rules (this change)
- `runai-cluster/cluster/runtime-rules/cmd/main.go` — env, scheme, flag.
- `runai-cluster/cluster/runtime-rules/runtime_rules.go` — fourth watch.
- `runai-cluster/cluster/runtime-rules/pkg/controllers/workload_controller.go` — `getObjectList` extension.
- `runai-cluster/cluster/runtime-rules/pkg/handlers/workload_handler.go` — v1-cast scope gate, category-label allowlist, skip rules, annotation-lifecycle helper (`reconcileTimestampAnnotations`).
- `runai-cluster/cluster/runtime-rules/pkg/util/functions.go` — `ConvertToJobWorkload`, `AttachedResourceHasRequestedGpu`, attached-resource fetcher.
- `runai-cluster/cluster/runtime-rules/pkg/util/category_mapping.go` (NEW).
- `runai-cluster/cluster/runtime-rules/pkg/util/attached_resource.go` (NEW) — GVK + Name + Namespace resolution from `WorkloadRunner.Spec`; unstructured fetch.
- `runai-cluster/cluster/runtime-rules/pkg/kubeConfig/` — `GetWLR`, `ListWLR`.
- `runai-cluster/cluster/runtime-rules/pkg/queuesManager/priority_queue.go` — `SourceKind` field.
- `runai-cluster/cluster/runtime-rules/pkg/queuesManager/priority_queue_utils.go` — fetch / update branches, `deactivateWorkload` extraction.
- `runai-cluster/cluster/runtime-rules/pkg/queuesManager/suspend_attached.go` (NEW) — per-Kind suspend patch on the attached resource.
- `runai-cluster/cluster/runtime-rules/pkg/controllers/autodeletecompleted_controller.go` — WorkloadRunner watch + kind dispatch.
- `runai-cluster/cluster/runtime-rules/pkg/handlers/autodeletecompleted_handler.go` — kind dispatch in `getWorkload` / `getDeletionTimestamp`; phase guard; `client.Delete` on the runner.
- `runai-cluster/cluster/runtime-rules/test/workload_runner_test.go` (NEW — time-limit + idle + annotation-lifecycle cases).
- `runai-cluster/cluster/runtime-rules/test/autodeletecompleted_test.go` — extended with WorkloadRunner cases.
- `cluster/charts/runai-cluster/templates/runtime-rules/deployment.yaml` + ClusterRole.

### Upstream (separate PRs, prerequisites)
- **control-plane** — single owner for all submission-time stamping:
  - Stamp the v1→v2 cast annotation on every `WorkloadRunner` produced by migrating a legacy v1 submission.
  - Stamp the `runai/workload-category` label on `Job`-backed runners.
  - Stamp `runai/auto-deletion-time-after-completion-seconds: "<seconds>"` on the attached resource.
  - Confirm GPU resource requests on the attached resource's pod template (verify parity with legacy submissions).
- **No `anyworkload-controller` change**. The previous draft's prerequisites on this team (publish `GPURequested` condition, stamp the two timestamp annotations) are removed by this design.
- **No `workloads-manager` change**. The previous draft's v2 action endpoint dependency is removed by this design.

## 16. Verification

1. Spin up a dev cluster with `ENABLE_WORKLOAD_RUNNER_RULES=1`. Confirm the runtime-rules ServiceAccount can `list / patch workloadrunners` and `patch jobs.batch` (`kubectl auth can-i`).
2. Set `TrainingJobTimeLimitSecs=60` on a project; submit a training workload via the legacy v1 path that is then cast to v2 by the control-plane (verify the v1-cast annotation is present and that the runner wraps a `batch/v1.Job`). Wait > 60s; expect the attached `Job.Spec.Suspend` to flip to `true`, the runner's `runai/start-time` annotation to be cleared, and `Status.Phase` to transition to `Suspended`.
3. Set `InteractiveJobMaxIdleDurationSecs=120`, submit an interactive workload (`batch/v1.Job`-wrapped, category=`interactive`, v1-cast), do not produce GPU activity. After ~2 min the attached `Job.Spec.Suspend == true`.
4. Submit a legacy `TrainingWorkload` directly; confirm the existing `Spec.Active=false` suspend still works (no regression).
5. **Scope-gate**: submit a workload directly as v2 (no v1-cast annotation); confirm it is not governed — the runtime-rules skips it and increments `runai_workloadrunner_not_v1_cast_skipped_total`.
6. **Resume**: after a time-limit suspend (step 2), clear `Spec.Suspend` on the attached Job; confirm the runtime-rules stamps a fresh `runai/start-time` and the new run is not immediately re-suspended.
7. **Auto-delete**: submit a runner-wrapped workload with `AutoDeletionTimeAfterCompletionSeconds=10` on the attached resource (set by control-plane at submission), let it complete (`anyworkload-controller` transitions `Phase=Completed`); confirm the runtime-rules stamps `runai/completion-time`, then ~10s later issues `client.Delete` on the WorkloadRunner; verify the attached resource is cleaned up via owner-reference cascade.
8. Tests: `make test` in `runtime-rules/` (Ginkgo) and a Helm template diff to confirm new env / RBAC.

## 17. Risks and Open Questions

1. **v1→v2 cast annotation key and value vocabulary**. The exact annotation key (working assumption `runai/v1-cast`) and whether presence is the gate vs. a specific value (e.g., `"true"`) must be pinned with the control-plane team. Until pinned, the runtime-rules treats absence as out-of-scope (the conservative direction).
2. **Category label name and values for `Job`-backed runners**. Confirm with the control-plane team. Pinned in whichever control-plane service produces the v1-cast `WorkloadRunner` and the constants in `runai-cluster/common/constants/constants.go`.
3. **AttachedResource annotation contract (control-plane responsibility)**. The control-plane must stamp `runai/auto-deletion-time-after-completion-seconds` on the attached resource at submission time. The team and exact code path that materializes the attached resource must be identified before the runtime-rules change can be activated. A missing annotation means "no auto-delete" — same as the legacy `nil` case — so this is recoverable but the user-facing feature is silently lost for those workloads.
4. **Per-Kind suspend field drift**. Suspend uses `Spec.Suspend` on `batch/v1.Job` and `Spec.RunPolicy.Suspend` on the Kubeflow operators. If a future Kind in scope changes its suspend contract, the table in §7 needs an update. Mitigation: a Kind-allowlist envtest case for each entry, so a Kind regression fails CI in the runtime-rules test suite.
5. **Annotation-lifecycle race with `anyworkload-controller`**. Both the runtime-rules and `anyworkload-controller` reconcile the `WorkloadRunner` on phase transitions. The runtime-rules is the **sole writer** for `runai/start-time` and `runai/completion-time`, but it reads `Status.Phase` which `anyworkload-controller` writes. Stale-read risk: the runtime-rules could observe `Phase=Running` from an outdated cache while the runner is already `Failed`. Mitigation: the annotation-lifecycle helper is idempotent (set-only-if-absent + clear-only-if-present); a delayed observation re-runs and converges. Worst case: a one-tick delay in stamping `completion-time`, which trickles into a one-tick delay in auto-delete — acceptable.
6. **Phase model stability**. `WorkloadRunnerPhase` (`workloadrunner_types.go:252-275`) is part of the contract. New phases must be reviewed against the runtime-rules active-vs-terminal split. Mitigation: default unknown phases to "skip enforcement" and emit a `runai_workloadrunner_unknown_phase_total` counter.
7. **`workloadId` annotation key is unscoped** (`"workloadId"`, not `run.ai/workload-id`). Was a hard skip in the prior draft because we needed the UUID for the v2 API call. Relaxed to a warning now because suspend/delete are K8s-native; the UUID is only used for metric correlation. A future rename by whichever component stamps it would degrade observability but not break enforcement.
