# EW / EWI / Cluster Sync survival across WLR restart

Complementary design to [design.md](./design.md), which deferred this section.

## Problem

A WLR restart deletes and recreates the child (e.g. PyTorchJob). The child owns the
ExternalWorkload, so the EW is garbage-collected, cluster-sync reports the delete, and the
CP drops the workload with its history. The recreated EW comes back with the same name and
the same `workloadId` — so this is not an identity problem, it is a **delete that should
never have fired**.

## Why the EW dies

- EWI sets the EW owner to the child object itself
  ([external_workload_handler.go:473-497](../../runai-cluster/cluster/external-workload-integrator/pkg/externalworkloads/external_workload_handler.go)).
- The WLR is only a label, `run.ai/workload-runner-name` (same file, `setWorkloadRunnerLabel`).
  `workload_runner_watcher.go` states it: the WLR is *"reference-only; never a top owner"*.
- Child deleted → GC deletes the EW → cluster-sync `DeleteFunc` → `DeleteWorkload(workloadId)`.

## Two delete vectors, not one

This is what rules out the "just guard the delete hop" option:

1. **Event delete** — cluster-sync `DeleteFunc` → `HandleDelete` → `DeleteWorkload`.
2. **Periodic full sync** — `SyncWorkloads` is range-authoritative: the CP deletes every
   workload in the `from`/`to` range that is not in the payload or in `IdsToIgnore`
   ([workloads_service.go](../../control-plane/workloads/pkg/services/workloads_service.go),
   `DbSyncWorkloads`).

Guarding only (1) still loses the workload to any full sync that lands mid-restart.

## Decision: the WLR owns the EW

For WLR-owned children, EWI sets the **WLR** as the EW's ownerReference instead of the child.

The EW then outlives child recreation and both delete vectors become moot by construction,
rather than by two guards that must stay in sync. It also matches the object model: the WLR
is the thing with the stable lifetime and the stable `workloadId`; the child now churns by
design.

Non-WLR workloads keep today's behavior — the child stays the owner.

### Options considered

| Option | Verdict |
|---|---|
| **WLR owns the EW** | **Chosen.** Covers both vectors. EW lifetime matches workload lifetime. |
| Guard the delete hop in cluster-sync (skip `DeleteWorkload` while `managedBy` is set) | Rejected. Covers vector 1 only. |
| Re-parent the EW just before the child delete | Rejected. Same end state as the chosen option, but only during the restart window, and it puts EW knowledge in the restart reconciler. |

## What we report when there is no child

Two cases, both now possible because the EW survives:

- **Child deleted mid-restart.** Keep the last known EW spec; drive phase from the WLR. This
  works with no new trigger: the EW already exists, and the WLR watcher already enqueues it.
- **Child never created** (a plugin or build failed). Nothing creates an EW here today —
  `createExternalWorkloadForObject` takes the child, and the WLR watcher only enqueues EWs
  that already exist. Covering this needs a new WLR-keyed trigger that synthesizes child
  identity from `spec.template`. That is a capability addition rather than a consequence of
  the ownership change, so it is **deferred to open question 2**: scope first, mechanism after.

The WLR already carries the reason in `status.issues` and `status.plugins[].issues`, which is
what the phase is derived from.

## How we parse it

**Dual parsing, and it already exists.** `kubeflowjob_runaistatus.go` computes distributed
status from the child (`JobStatus`, replica specs, pods) and overlays WLR critical issues via
`extractCriticalIssues(wlr)`; `runai_status.go` does the same for inference. The split is
already: **child = structure and counts, WLR = issues and phase.**

We keep that split and add one rule: when the child is absent, parse the WLR alone. No new
parsing layer, no per-framework duplication.

## Consequences

- Existing EWs need their ownerRef flipped from child to WLR — see open questions.
- `setWorkloadRunnerLabel` becomes partly redundant once the WLR is a real ownerRef, but the
  label stays: it is the list index used by `enqueueWorkloadsForWorkloadRunner` and others.
- Deleting the WLR now GCs the EW directly, which is the intended lifetime.
- The `Restarting` phase needs no special handling here — cluster-sync degrades it to
  `Unknown` and the UI shows the last known state, as `design.md` already accepts (designs/restart-controller-workloadrunner/design.md).

## Open questions

1. Is the startup backfill in [Rollout](#rollout) enough for EWs that exist at upgrade time,
   or does migration need an explicit gate? Depends on how long EWI is down during a rolling
   upgrade.
2. Should a WLR-driven create path exist at all — and if so, is it limited to the parity
   scope (PyTorch via Heimdall) or every WLR? It needs a new WLR-keyed reconcile trigger
   (see [What we report when there is no child](#what-we-report-when-there-is-no-child)),
   so scope should be settled before the mechanism is designed.
3. e2e: `design.md` plans to re-enable restart e2e with the new implementation. Add a case
   that asserts CP history survives a restart.

---

# Appendix: implementation

## How WLR is excluded as top owner today

Not by the drop list. `kindsToDrop` holds the legacy v2 workloads (DW, TrainingWorkload, …);
**WLR is not in it**. WLR is excluded only by being **absent from `supportedGVKsHandlers`**:

`getTopResourceOwnerWithHandler` walks Pod → PyTorchJob → WLR, finds WLR has no owners (true
top) and no handler, and returns the pair `(WLR, PyTorchJob)`. Every call site then uses the
second value, `topOwnerWithHandlerMeta` — the child.

**So we do not add WLR to `supportedGVKsHandlers`.** That would move EW identity, name and
parsing onto the WLR — a far larger change, and it would break every child-keyed lookup. The
child stays the "top owner with handler".

## The change is ownership-only

Only `getOwnerReferences()` changes: if the child has a WLR ownerRef, return that; otherwise
today's behavior. This deliberately splits two things that are conflated today:

- **child** — what the EW *is about*: identity, name, spec, parsing.
- **WLR** — what keeps the EW *alive*.

## The blocker: EW lookup is keyed on the child UID

This is the part that makes the ownerRef flip insufficient on its own.

- EWs are found by a field index on `Spec.OriginalWorkloadUID` (`GetExternalWorkloadOfObject`).
- The EW name, `ew-<Kind>-<childName>`, is **stable** across restart.

After a restart the child returns with the same name but a **new UID**. So:

1. lookup by new child UID → NotFound
2. EWI takes the create path
3. it generates the same name as the surviving EW → **AlreadyExists**

The surviving EW is orphaned by index while its name blocks the replacement. Required
alongside the ownerRef change:

- on child create, if an EW with that generated name already exists **and its WLR UID matches
  the current WLR**, **re-bind** it — update `Spec.OriginalWorkloadUID` and
  `Spec.CreationTimestamp` to the new child instead of creating.
- **Match on UID, not `run.ai/workload-runner-name`.** The name is reusable and EW names are
  deterministic, so a workload deleted and resubmitted under the same name would otherwise
  adopt the dead workload's history in the window before GC removes the old EW. This means
  stamping a WLR UID label alongside the existing name label.
- alternatively, index WLR-owned EWs by WLR UID rather than child UID. Cleaner, but touches
  more call sites.

## Other touchpoints

| Area | Impact |
|---|---|
| `setWorkloadRunnerLabel` | Keep. Still the list index for `enqueueWorkloadsForWorkloadRunner`. |
| `workload_runner_watcher.go` | Comment *"never a top owner"* becomes wrong; the enqueue logic itself stays valid. |
| `job_handler`, `kubeflowjob_runaistatus`, `runai_status`, `networkstatus` | Unaffected — all read the WLR by label, not by ownership. |
| `runai_status_pod_watcher` | **Affected.** Its pod→EW mapping is `workloadId`-based and stable, but it also calls `gvkFromExternalWorkloadOwner` (line 148) — see below. |
| cluster-sync `WasSynced` | Keyed on EW UID, which now stops churning. Small improvement. |
| structure updater | Pod → EW mapping is `workloadId`-based, unaffected. |

## Rollout

`updateExternalWorkloadForObject` already calls `getOwnerReferences`, so existing EWs have
their ownerRef rewritten on their next reconcile.

That still leaves a window. GC acts on ownerRefs independently of EWI, so a child deleted
before EWI has reconciled that EW — during a rolling upgrade, say — collects the EW as it
does today. The window degrades to **current behavior** rather than introducing new loss, so
it does not warrant a migration gate; but it is cheap to close by rewriting ownerRefs during
EWI's initial sync on startup, which is the preferred option.

The re-bind path must ship **in the same version** as the ownerRef flip — shipping the flip
alone produces the name collision above.

## Tests

- restart: EW UID unchanged, CP workload and history survive.
- re-bind: EW is reused, not duplicated, when the child returns with a new UID.
- re-bind refuses on WLR UID mismatch: a same-name resubmit does not adopt the previous
  workload's EW or its history.
- re-bind refuses with no `workloadId` annotation: CP identity is left unchanged.
- child GVK still resolves after the ownerRef flip: kubeflow workloads keep being enqueued by
  `runai_status_pod_watcher` and the WLR watcher, and the CP distributed framework stays
  `PyTorch`.
- non-WLR workloads: owner is still the child.

## Cluster-sync and other ownerRef readers

The EW ownerRef is not private to EWI. Three places read it; **two regress**.

**1. Distributed framework goes empty (cluster-sync).**
`distributedFrameworkForExternalWorkload` tries `Spec.Type` first, then falls back to
ownerRefs. For v1-cast workloads — our parity scope — `Spec.Type` is the *native* kind
(`Distributed`), not `PyTorchJob`, so the first lookup returns `""` and the ownerRef fallback
is what actually resolves `PyTorch`. After the flip the ownerRef Kind is `WorkloadRunner`,
which maps to `""`, and the framework field empties out in the CP and UI.

**2. Kubeflow workloads stop being enqueued (EWI).**
`gvkFromExternalWorkloadOwner` matches the ownerRef's APIVersion against
`Spec.Group`/`Spec.Version` (`kubeflow.org/v1`). A `run.ai/v1alpha1` WLR ref never matches, so
it returns `false` and both the pod watcher and the WLR watcher silently skip the workload —
status goes stale with no error.

Both have the same root cause and the same fix: **derive the child GVK from
`Spec.Group`/`Spec.Version` plus the child kind, not from ownerRefs.** The EW already stores
group and version; the kind needs to be carried explicitly rather than inferred from the owner.

EWs created before that field exists will not carry it, so both readers need a fallback to
the current ownerRef behavior until every EW has been rewritten — or the same startup
backfill as the ownerRef flip (see Rollout).

(`revision_handler.go` also reads ownerRefs, but for Knative Revisions — unrelated.)

## What cluster-sync does *not* need

- **Both delete vectors close on their own.** `getParsedExternalWorkloadsInNamespace` lists
  live EWs, so a surviving EW keeps appearing in the full-sync payload and is never reaped.
  No guard, no `managedBy` awareness, no changes to `HandleDelete`.
- **`overrideFromWorkloadRunner` is already WLR-aware** and reaches it by label. It also
  registers the WLR UID in `ChildrenIds`, which is what joins runner events to the workload —
  unaffected, and part of why history survives.
- **`WasSynced`** is keyed on the EW UID, which now stops churning: fewer redundant upserts.

**Re-bind invariant.** `workloadId` is `GetRunaiWorkloadId(Spec.OriginalWorkloadUID,
annotations)` — stable only while the `workloadId` annotation is present. Re-bind **must not
proceed without it**: updating `Spec.OriginalWorkloadUID` with the annotation missing falls
the ID back to the child UID and changes CP identity on every restart, the exact failure this
design exists to prevent. EWI writes the annotation only when it differs from the child UID
(`external_workload_handler.go:421-424`), so workloads not submitted through the CP will not
have one — those keep today's behavior rather than being re-bound.
