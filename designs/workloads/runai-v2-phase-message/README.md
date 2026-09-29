# Design: phase + phaseMessage for Run:ai workloads in workloads-v2 (temporary)

> Per-type details live in [`types/`](./types):
> [Training & Workspace](./types/training-workspace.md) · [Distributed Training](./types/distributed-training.md) · [Inference](./types/inference.md) · [Distributed Inference](./types/distributed-inference.md)

## TL;DR (the whole thing in 3 lines)
1. The **new** v2/anyworkload flow composes Run:ai workloads into **native engine resources** (Training/Workspace → `batch/v1 Job`) — the old run.ai workload CRDs are **gone from the cluster**, so there's no `status.phaseMessage` to read.
2. **EWI builds the full phase + phaseMessage in one place**: lifecycle states from the **engine + pods**, and the Failed/Degraded *cause* from the **WorkloadRunner's grouped issues/conditions** (EWI reads + watches the wlr).
3. cluster-sync holds **no build logic** — one guard to stop overriding EWI's phase. Old run.ai CRDs/controllers and `karta` untouched (reference only). **First PR = Training + Workspace.**

---

## Context — why
We migrate Run:ai workloads to **workloads-v2 / anyworkload** (WorkloadRunner). In the new flow the run.ai workload CRDs (`TrainingWorkload`, `InteractiveWorkload`, …) are **not created** — Composer-V2 converts them into native engine objects. We **reproduce** the same phases/messages **temporarily** in EWI, encapsulated for easy removal once a permanent owner exists.

> **Why not just use `wlr.status.phase`?** The wlr *does* emit a phase enum (`Creating/Pending/Initializing/Running/Stopped/Degraded/Failed/Completed/Suspended/Suspending/Resuming` — `workloadrunner_types.go:278`), and cluster-sync **already** syncs it via `overrideFromWorkloadRunner()`. Two gaps remain, and they're why EWI rebuilds:
> 1. **No message.** The wlr has **no `phaseMessage`/`message` field** — only `status.issues[]{Description, Group}`. The human string ("waiting for N pods", "failed due to a compute issue") does not exist anywhere and must be built.
> 2. **Phase isn't v1-parity, and must agree with the message.** The wlr phase is computed from the runner/plugins, not from Job+pods, so it doesn't reliably split v1's `Creating` vs `Pending` (needs pods) or match v1 `Initializing`/`Running` thresholds. Since the *message* is derived from Job+pods, the *phase* must be derived from the same inputs or the two can disagree (e.g. phase `Running` while the message still says "pulling image"). Building both in one place guarantees consistency.
>
> If a future wlr exposes a v1-accurate phase **and** a message, EWI can drop to message-only (or be deleted entirely) — see §7.

> **Old run.ai CRDs = reference only.** Their controllers are untouched; we read their *logic* (what phases/messages they emit) and reproduce it. We do not depend on them at runtime in the new flow.
>
> **Single build location:** all phase + phaseMessage construction lives in EWI. cluster-sync only stops discarding it.

---

## End-to-end v2 flow — Training (services + inputs/outputs)

### A) Control-plane → Cluster (submit / desired state)

```text
 USER (UI / CLI / API)
   │  POST /api/v2/workloads  { manifest (Training/Workspace), projectId, category... }
   ▼
┌──────────────────────────────────────────────────────────────────────────┐
│ workloads-manager  (CP, workloads-v2 API)                                  │
│   in : WorkloadV2CreateRequest                                              │
│   • validate, resolve project, GVK→workload-type, defaults, MNNVL/net       │
│   out: hands manifest to Composer-V2                                         │
└──────────────────────────────┬─────────────────────────────────────────────┘
                               ▼
┌──────────────────────────────────────────────────────────────────────────┐
│ Composer-V2  (CP)  composer_v2/iw_tw_workloads/iw_tw_to_job.go              │
│   in : Training/Workspace spec                                              │
│   • CONVERT → batch/v1 Job (+ run.ai labels/annotations, scheduler, podgroup)│
│   out: effective manifest = batch/v1 Job   ◄── NO run.ai CRD created         │
└──────────────────────────────┬─────────────────────────────────────────────┘
                               ▼ persist
                      ┌───────────────────────┐
                      │ workloads_queue  DB    │  UpdateCount = incremental cursor
                      └───────────┬───────────┘
                                  ▲  GET /workloads/v2/updates?clusterId&updatesAfter=N
┌─────────────────────────────────┴────────────────────────────────────────┐
│ cluster agent  (cluster, sync-in)                                          │
│   in : updates (batch/v1 Job manifests)                                     │
│   • wrapWithWorkloadRunner(): Job → spec.template ; mnnvl/net→instructions  │
│   out: WorkloadRunner (wlr) CRD                                             │
└──────────────────────────────┬─────────────────────────────────────────────┘
                               ▼
┌──────────────────────────────────────────────────────────────────────────┐
│ anyworkload-controller  (cluster)                                          │
│   in : wlr.spec.template (= batch/v1 Job) + instructions                    │
│   • apply template → CREATE batch/v1 Job ; run plugins (net/mnnvl)          │
│   out: Job ; wlr.status.phase + wlr.status.issues{Group: Compute/Storage/…} │
└──────────────────────────────┬─────────────────────────────────────────────┘
                               ▼
                   batch/v1 Job  →  Pods   (the engine; no run.ai CRD, no workload-controller)
```

### B) Cluster → Control-plane (observed status / where our change lives)

```text
   batch/v1 Job.status + Pods            WorkloadRunner.status
        │                                  • phase (Degraded/Failed/…)
        │                                  • issues[]{Description, Group}
        │                                  • plugins[].issues, conditions
        ▼                                              │
┌───────────────────────────────────────────────────────────────────┐
│ EWI  (external-workload-integrator)   [reads + WATCHES the wlr]     │
│   in : batch/v1 Job (top owner) + Pods + WorkloadRunner             │
│   • NEW: build the FULL phase + message (one place):               │
│       - lifecycle (engine+pods): Creating/Pending/Initializing/     │
│         Running/Completed, "waiting for N pods", image-pull, etc.   │
│       - Degraded/Failed cause: formatFromIssues(wlr.issues) →       │
│         "Workload failed due to a compute issue"                    │
│   • NEW: set annotation ewi-status                                  │
│   out: ExternalWorkload.observedStatus{phase, message} (+annotation)│
└───────────────────────────────┬────────────────────────────────────┘
                                ▼
┌──────────────────────────────────────────────────────────────────────────┐
│ cluster-sync  (cluster, sync-out)  external_workload_parser.go             │
│   in : ExternalWorkload.observedStatus                                      │
│   • Phase/PhaseMessage = observedStatus.{Phase,Message}                     │
│   • NEW (1 guard): in overrideFromWorkloadRunner(), if annotation set →     │
│        SKIP ONLY the phase override (lines that set Phase from wlr.phase).   │
│        Child-id join, wlr issues→AdditionalFields, URL fallback STILL run.   │
│   out: WorkloadSyncFields {phase, phaseMessage, conditions, ...} → CP       │
└──────────────────────────────┬─────────────────────────────────────────────┘
                               ▼
┌──────────────────────────────────────────────────────────────────────────┐
│ workloads-manager  (CP)                                                    │
│   out: persist; GET /api/v2/workloads/{id} → phase + phaseMessage (v1-parity)│
└──────────────────────────────────────────────────────────────────────────┘
```

---

## 1. Supported workload types + the engine resource the new flow creates

Focus now = **Training + Workspace**. Others for the full picture. **All phase+message built in EWI.**

| Workload type | New-flow engine EWI observes | Lifecycle msg | Failed/Degraded msg (built in EWI) | Detail |
|---|---|---|---|---|
| **Training** | `batch/v1 Job` → Pods | Job + **pods** | Job cond + **wlr** issue-cause | [training-workspace](./types/training-workspace.md) |
| **Workspace** | `batch/v1 Job` → Pods | Job + **pods** | Job cond + **wlr** issue-cause | [training-workspace](./types/training-workspace.md) |
| Distributed Training | kubeflow job (PyTorchJob…) → Pods | replicas + **pods** | Job cond + **wlr** issue-cause | [distributed-training](./types/distributed-training.md) |
| Inference | Knative Service → Pods | Knative conds + **pods** | Knative conds (+ **wlr** build error) | [inference](./types/inference.md) |
| Distributed Inference | LeaderWorkerSet → Pods | LWS status + **pods** | LWS counts (+ **wlr** build/MNNVL) | [distributed-inference](./types/distributed-inference.md) |

**Pods are required for every type** — the rich lifecycle messages come from per-pod state, not the engine's top-level status (the engine only gives coarse Ready/NotReady/Active, not the *reason*). EWI already fetches pods via `getters.GetPodsOfWorkload`; pods RBAC is present.

Training vs Workspace: distinguished by category label (`Train` vs `Build`) and/or `spec.parallelism` (single- vs multi-pod init message).

---

## 2. Where we implement this — EWI only (build), cluster-sync (one guard)

**EWI** — the whole build:
- Engine is a `batch/v1 Job`, handled by `pkg/handlers/job_handler.go` (`setJobObservedStatus`), today coarse + message-less.
- New package `pkg/handlers/runaistatus/` with `ComputeStatus(engine, pods, wlr) → (phase, message)`:
  - lifecycle from engine + pods (reference: `workload_status_calculator.go`),
  - Degraded/Failed cause from `wlr.status.issues` / `plugins[].issues` / conditions. Groups are `Compute/Storage/Network/Internal` (`dromos/pkg/reconcile/issue/issue.go:15`). `Internal` is not user-facing — map it to a generic fallback ("Workload failed due to an internal issue" / omit the group) rather than rendering "due to a Internal issue".
- `job_handler.go` (gated to run.ai Jobs + feature flag): look up the wlr (via `WorkloadRunnerNameLabel`), call `ComputeStatus`, set **only** `observedStatus.{Phase,Message}` + the `ewi-status` annotation.
- **EWI controller setup:** add a **watch on `WorkloadRunner`** mapping each wlr → its child workload (Job) so wlr-only status changes (e.g. a plugin failure) re-trigger reconcile.

> ### ⚠️ Invariant — `K8sStatus` ≠ `Phase`; never overwrite the native status
> `observedStatus` carries **three** fields with two different meanings (`externalworkload_types.go`):
> - `Phase` — *"mapping native k8s statuses to pre-defined list of phases"* → the run:ai **normalized** enum (computed: `Creating/Pending/Initializing/Running/Degraded/Failed/…`).
> - `Message` — the human phaseMessage.
> - `K8sStatus` — *"native k8s status"* → the **raw engine** value (`Active/Succeeded/Failed/Suspended/Creating`), same convention every other handler follows (kubevirt→`vmi.Status.Phase`, ray→`State`, kubeflow→`JobRunning`…).
>
> `ComputeStatus` returns **only** `(phase, message)`. `K8sStatus` must keep coming from the native Job status (reuse the existing `setJobObservedStatus` path, then override Phase+Message). **Do not set `K8sStatus = string(phase)`** — that collapses the two axes and emits non-k8s values (`Degraded`, `Initializing`) into the raw field, which then surfaces downstream (`cluster-sync K8sPhase → CP → frontend fallback display`).
>
> ```text
>   Degraded workload →  Phase = "Degraded"   (run:ai, computed)
>                        Message = "...network issue"
>                        K8sStatus = "Active"   ◄── engine is STILL running; stays raw
> ```

> ### ⚠️ Hard constraint — the wlr is reference-only, NEVER a top owner
> The WorkloadRunner is the **controller-owner of the Job** (`ctrl.SetControllerReference(runner, template)` — `dromos/pkg/controller/reconcile_spec.go:209`), so it sits **one rung above** the Job in the ownership chain:
>
> ```text
>   WorkloadRunner ──owns──► Job ──owns──► Pods
>         ▲ controller:true        ▲
>         │                        └── EWI's top owner TODAY (Job is a registered GVK)
>         └── NOT a registered GVK → owner-walk stops at the Job ✓
> ```
>
> EWI picks the top owner by climbing ownerRefs and **stopping at the highest GVK it recognizes** (`supportedhandlers/supported_gvk_handlers.go`). The Job is the top owner *only because the wlr is not registered*.
>
> **Why the Job must stay top owner:** top owner = the object that *becomes* the ExternalWorkload (its UID / workload-id / labels). The Job is the engine (real lifecycle, owns the pods, v1-parity status); the v2 contract is **one ExternalWorkload per Job**, and cluster-sync/CP/the v2 API all key off that identity.
>
> **What would break the flow:** registering the wlr as a supported GVK — or adding a **Karta** for `run.ai/workloadrunners` — makes the owner-walk climb *past* the Job to the wlr. The ExternalWorkload would then be built for the **wlr** (new UID/identity), the Job demoted to a child, `ComputeStatus` no longer runs at the top, and the real workload is orphaned/duplicated in CP.
>
> **The trap:** the *normal* EWI way to "support a new type" **is** to register its GVK / add a Karta. Resist it here — we need the wlr's **data** (issues → failure cause), not its **identity**.
>
> **So, strictly:**
> - ✅ Read the wlr as a side-reference (via `WorkloadRunnerNameLabel`) — same as cluster-sync already does.
> - ✅ Watch the wlr **only** as a requeue trigger; the map function returns the **owning Job's** key.
> - ✅ RBAC = `get/list/watch` on the wlr (read), nothing implying adoption.
> - ❌ Do **not** add the wlr to `supported_gvk_handlers.go`; do **not** create a Karta for it.

**cluster-sync** — no build, one narrowly-scoped guard:
- `overrideFromWorkloadRunner()` does **four** things today: (1) override Phase from `wlr.Status.Phase`, (2) register the wlr as a child so its events join the workload, (3) convert `wlr` issues → `AdditionalFields`, (4) external-URL fallback.
- The guard touches **only (1)**: when the `ewi-status` annotation is set, skip the **phase override** so EWI's phase+message survive. **(2), (3), (4) keep running** — otherwise we'd regress wlr event-joining, the structured issues sent to CP, and the URL fallback.
- Note this means the wlr is read in **two** places (EWI + cluster-sync); see §3 for why that's an accepted trade for the "single build location."

### Robustness, rollback & edge cases
- **wlr missing / racing / GC'd / empty issues:** EWI may reconcile a Job before the wlr exists (anyworkload-controller creates it async), after it's GC'd, or while `wlr.status.issues` is nil/empty. The wlr contributes **only** the Failed/Degraded *cause*, so in all these cases EWI still builds the lifecycle phase+message **from engine+pods alone** — a missing/empty wlr just yields no cause suffix (e.g. job-level `"job reached failed state"` without the `"due to a <group> issue"` tail). The wlr watch re-triggers reconcile when issues later appear. **Status never blocks on the wlr.**
- **Unknown phase → safe default:** if `ComputeStatus` hits an engine/wlr state that doesn't map to a known `ExternalWorkloadPhase`, log a warning and default to `Pending` (never panic). Lets the enum be extended without breaking EWI.
- **Defensive wlr→Job map:** the watch map function enqueues only when the wlr has a valid **controller ownerRef to a Job**; broken/orphaned/hand-edited ownerRef → return empty (skip enqueue) + warn, so a GC-lagged wlr can't cause reconcile churn or enqueue a nonexistent Job. Don't assume `ctrl.SetControllerReference` completed synchronously.
- **Runtime top-owner guard:** before `ComputeStatus`, assert the top-owner GVK is **not** `run.ai/workloadrunners`. If it ever is (wlr accidentally registered as a supported GVK / Karta), log a **critical** error and skip — fail fast rather than corrupt CP identity. Complements the §6.4 regression test and enforces the §2 hard constraint at runtime.
- **Annotation trust (why no signature/admission):** `ewi-status` lives on an internal CR written by trusted controllers; existing RBAC already restricts who can write `ExternalWorkload` status, and EWI writes phase+message+annotation in **one** status update (cluster-sync only *reads* it — no concurrent writer, eventually consistent, EWI re-reconciles on drift). So no ownerRef-signature/admission-webhook is warranted for this internal path.
- **Gate OFF / removal fallback:** feature flag off → EWI does **not** compute or annotate, and cluster-sync's `overrideFromWorkloadRunner()` runs its **original** phase override (full v1-prior fallback — safe rollback). The cluster-sync guard must honor the annotation **only while the gate is on**, so a stale `ewi-status` annotation left on an old ExternalWorkload after rollback/removal can't freeze its status. (Removal need not strip annotations — the gate-scoped guard makes them inert.)

---

## 3. How we keep it encapsulated (this is temporary)

```text
EWI:  runaistatus/  (one package, single entry point)
        ComputeStatus(engine, pods, wlr) → (phase, message)   ← ALL build logic here
      job_handler.go: gate + wlr lookup + call + annotation
      controller setup: watch WorkloadRunner → enqueue child workload
cluster-sync: one guard in overrideFromWorkloadRunner() — skip ONLY the phase override
              when annotation set; (2) child-join, (3) issues→AdditionalFields, (4) URL keep running
```
- All construction in **one** EWI package; everything else is wiring/gates.
- One feature gate + run.ai-label gate → zero impact on non-Run:ai workloads.
- **Accepted trade-off — wlr read twice:** cluster-sync already reads the wlr (for issues→`AdditionalFields`, child-join, URL). EWI now reads it too, because only EWI also has the **pods**, and the phaseMessage must combine lifecycle (pods) + cause (wlr) in **one** consistent place. We keep cluster-sync's existing wlr reads intact rather than move them, to keep this change small and removable. CP still receives the structured issues (from cluster-sync) **and** the human message (from EWI).
- **Removal = delete `runaistatus/`, revert the `job_handler.go` gate + the wlr watch + the cluster-sync phase-override guard + the new RBAC.** No karta/CRD entanglement.

---

## 4. Per-type phase + message
The full per-type matrices (how we reach each phase + message/source/feasibility) live in [`types/`](./types). Each file has two tables: **how we reach each phase** and **phase → message → source → derivable**.

### What actually needs the WorkloadRunner (the only non-engine bits)
- **Run:ai issue-cause** ("Workload failed/degraded due to a `<compute/storage/network>` issue"; `Internal` → generic fallback, not surfaced as a group) — Training, Workspace, Distributed Training.
- **Build errors** — Inference, Distributed Inference.
- **MNNVL / ComputeDomain not-ready**, missing extra resources — Distributed Inference (+ any MNNVL workload).
Everything else (lifecycle phases, init/image-pull/replica messages, engine-native failure reasons) comes from the **engine + pods**.

### Mechanism + RBAC + enum gap
- `ComputeStatus(engine, pods, wlr)` mirrors the v1 decision tree (per type).
- **RBAC (new, small):** add `get/list/watch` on `run.ai` `workloadrunners` (+ `/status`) to EWI; pods + jobs already present. K8s RBAC can't filter by label, so EWI's wlr **informer uses a `run.ai/workload-name` label selector** — narrows the cached/watched set to Run:ai workloads (lower memory, defense-in-depth vs other tenants' wlrs).
- **Enum gap:** `ExternalWorkloadPhase` lacks `Suspended/Suspending/Resuming/Terminating`; Jobs need only `Creating/Pending/Initializing/Running/Completed/Failed/Stopped` (+ `Degraded`) — all present. Extend later only if needed.

---

## 5. Task list — Training + Workspace (this PR)

- [ ] **T1 (EWI)** — `pkg/handlers/runaistatus/`: `ComputeStatus(job, pods, wlr)` reproducing the v1 decision tree (lifecycle from Job+pods; Failed/Degraded cause from wlr grouped issues).
- [ ] **T2 (EWI)** — `job_handler.go`: gate (run.ai label + feature flag) → wlr lookup + call `ComputeStatus`, set `observedStatus.{Phase,Message}` + the `ewi-status` annotation.
- [ ] **T3 (EWI)** — add **WorkloadRunner watch** (wlr → child Job mapping) so wlr-only status changes reconcile. **Map function must return the owning Job's key** and defensively skip (empty result + warn) when the wlr has no valid controller ownerRef to a Job (race/orphan/GC-lag); **do NOT register the wlr as a supported GVK / Karta** (see §2 hard constraint — keeps the Job as top owner).
- [ ] **T4 (EWI)** — add `workloadrunners` (+`/status`) `get/list/watch` to the EWI ClusterRole, and apply a `run.ai/workload-name` **label selector on the wlr informer**.
- [ ] **T5 (EWI)** — Training vs Workspace disambiguation (category `Train`/`Build` and/or `spec.parallelism`).
- [ ] **T6 (cluster-sync)** — one guard in `overrideFromWorkloadRunner()`: when annotation set, skip **only** the phase override (`workloadSyncFields.Phase = …wlr.Status.Phase`, `external_workload_parser.go:514-516`); leave child-join, issues→`AdditionalFields`, and URL fallback running. Add a regression test asserting those three still populate when annotated.
- [ ] **T7** — Unit tests: `ComputeStatus` over Job+pod+wlr-issue fixtures → exact v1 phase/message; **wlr missing/empty issues → lifecycle still built from engine+pods (no cause suffix)**; non-Run:ai unchanged; cluster-sync keeps EWI values when annotated; **gate OFF → EWI no-ops and cluster-sync resumes its original override; stale annotation + gate off → status still updates (not frozen)**.
- [ ] **T8** — E2E: submit Training + Workspace via v2; assert v2 API returns v1 phase + non-empty phaseMessage across lifecycle and a forced Failed/Degraded case.

---

## 6. Verification
1. Unit tests beside `runaistatus/`, `job_handler.go`, and the cluster-sync guard.
2. E2E (env-in-a-click): submit Training + Workspace; confirm `observedStatus.message` + v2 API `phase`/`phaseMessage` match v1 — including a degraded/failed case (force a plugin/issue).
3. Regression: non-Run:ai Jobs and other types unchanged (gate off / annotation absent).
4. **Top-owner invariant:** assert the ExternalWorkload is still built for the **Job** (not the wlr) after adding the watch + RBAC — the wlr must never appear as a top owner. Backed at runtime by the top-owner guard (§2) that skips + logs critical if the wlr is ever the top owner.

---

## 7. Out of scope (after approval)
- **Distributed Training** (kubeflow job → `kubeflowjob_handler.go`), **Inference** (Knative), **Distributed Inference** (LeaderWorkerSet): same EWI build per engine, follow-up PRs.
- **Long-term home**: move build out of EWI; delete `runaistatus/` + watch + guard + RBAC.

---

## Critical files
| Purpose | Path |
|---|---|
| **New — EWI build module** | `runai-cluster/cluster/external-workload-integrator/pkg/handlers/runaistatus/` |
| **Edit — EWI Job handler (gate + wlr lookup + call)** | `.../external-workload-integrator/pkg/handlers/job_handler.go` |
| **Edit — EWI controller setup (wlr watch)** | `.../external-workload-integrator/` controller/manager setup |
| **Edit — EWI RBAC (add workloadrunners)** | `runai-chart/templates/clusterroles/external-workload-integrator.yaml` |
| EWI pod helpers (reuse) | `.../external-workload-integrator/pkg/handlers/commonparsers/common_handler.go` |
| **Edit — cluster-sync guard** | `runai-cluster/cluster/cluster-sync/pkg/handlers/parsers/external_workload_parser.go` (`overrideFromWorkloadRunner`, phase override at `:514-516` only) |
| wlr status / Issue{Description,Group} types | `runai-cluster/cluster/dromos/api/v1alpha1/workloadrunner_types.go`; groups in `dromos/pkg/reconcile/issue/issue.go`; population in `dromos/pkg/reconcile/issues.go` |
| Reference — v1 phase/message logic | `runai-cluster/cluster/workload-controller/pkg/status/aggregator/workload_status_calculator.go` |
| Proves Job is the engine | `control-plane/workloads-and-policies/composer_v2/iw_tw_workloads/iw_tw_to_job.go` |
| Proves wlr controller-owns the Job (why wlr is 1 rung above top owner) | `runai-cluster/cluster/dromos/pkg/controller/reconcile_spec.go:209` |
| EWI top-owner GVK registry (wlr must NOT be added here) | `runai-cluster/cluster/external-workload-integrator/pkg/supportedhandlers/supported_gvk_handlers.go` |

## 🧠 Remember it as
"All phase+message build in EWI: lifecycle from engine+pods, Failed/Degraded cause from the WorkloadRunner's grouped issues (EWI now reads + watches the wlr, +small RBAC). The wlr already has a phase but **no message** and it isn't v1-parity, so EWI rebuilds both from the same inputs. cluster-sync stops overriding **only the phase** — its wlr issues→AdditionalFields, child-join and URL fallback keep running. The wlr controller-owns the Job, so it's one rung above the top owner — keep it **reference-only** (read + watch), never a registered GVK/Karta, or the Job stops being the top owner. Training + Workspace first; delete the module when a permanent owner exists."
