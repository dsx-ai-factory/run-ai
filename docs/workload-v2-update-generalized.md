# Workload-V2 update — generalized across typed workload APIs

> **Inference** update generalized to all five typed workload APIs. Goal: support **update**
> through the Workload-V2 / `workloads-manager` (WM) flow, reusing the `workloads_v2_snapshots` store
> as the spec source. Inference ships **now**; the rest are **later** but share the same
> machinery.

## The approach in one line

Build **one shared, snapshot-based update orchestrator**; each workload type plugs in only its
`submitter.Update` (which reuses that type's *create* composer). Wire **inference first**.

## Target APIs

| API | When |
|---|---|
| `/api/v1/workloads/inferences` | **now** |
| `/api/v1/workloads/workspaces` | later |
| `/api/v1/workloads/trainings` | later |
| `/api/v1/workloads/distributed` | later |
| `/api/v1/workloads/distributed-inferences` | later |

## Readiness per API (what each still needs)

| API | Snapshot on create? | Update endpoint today? | Work to enable update-via-WM |
|---|---|---|---|
| **inferences** | ✅ (WS/T/D/I) | ✅ `UpdateInferenceSpec` | implement `AnyWorkloadSubmitterInference.Update` |
| **workspaces** | ✅ | ❌ none | add update endpoint + `InterTrain.Update` |
| **trainings** | ✅ | ❌ none | add update endpoint + `InterTrain.Update` (same submitter as workspace) |
| **distributed** | ✅ (incl `master_spec`) | ❌ none | add update endpoint + `Distributed.Update` (merge **master + worker**) |
| **distributed-inferences** | ❌ none — `Dinf.CanCreate` hardcoded `false` → native fallback | ✅ `UpdateDistributedInferenceSpec` | **blocked**: first make DistInf *create* go V2 (composer) + write a snapshot, then `Dinf.Update` |

> Why the asymmetry: the snapshot row is written **only** when the workloads service (WS) takes the V2
> path, and **only for WS/T/D/I** (`persistWorkloadV2Snapshot` is the sole snapshot writer). DistInf is
> hardcoded to the native fallback, and the WM-direct endpoints (nim-services, raw YAML) never go
> through WS — so none of them produce a snapshot.

---

## Unified update flow (one path, type-parameterized)

```
 PUT /api/v1/workloads/{type}/{id}/spec      type ∈ {workspaces,trainings,distributed,inferences,distributed-inferences}
        │ [workloads svc]  Update<Type>Spec  →  shared orchestrator
        ▼ resolve workload
        ├─ found in crud_workloads (native)  ─►  native update (merge → recompose → upsert)   ◄ works today
        │
        └─ NotFound → V2 path:
              DbGetWorkloadV2Snapshot(id, type).ToWorkload()   → full SupersetSpec (+ MasterSpec)
                    │   ❌ no snapshot (DistInf / raw-YAML / nim) → cannot update via WM (see table)
                    ▼ submitter.Update(partialSpec)             ← PER-TYPE plug-in
                       mergeUpdateSpecIntoFullSpec(spec, partial)         ▲ SHARED (map-based merge)
                       <type>.Compose → manifest                ╮ per type — SAME composer used on create:
                          workspaces / trainings → RunaiJob      │   InterTrain
                          distributed            → dist. CRD     │   Distributed (master + worker)
                          inferences             → Knative       │   InferenceToKnative
                          distributed-inferences → dinf CRD      ╯   Dinf (composer not built yet)
                       UpdateWorkloadV2{ Manifest, Config } → WM          ▲ SHARED (existing generic update)
                       DbUpsertWorkloadV2Snapshot(merged)                 ▲ SHARED (write-back)
        → agent POST /api/v2/workloads/updates → WorkloadRunner.spec.template → cluster
```

## What is shared vs per-type

**Shared — built once, used by all five (this is the bulk of the work):**
- the orchestrator: resolve native (`crud_workloads`) vs V2 (snapshot), and route;
- snapshot **read** — `DbGetWorkloadV2Snapshot(...).ToWorkload()` → the same `models_crud.Workload` the native path uses;
- the **merge** — generalize the existing inference `copyUpdateSpecIntoFullSpec` into a type-agnostic
  `mergeUpdateSpecIntoFullSpec(full *SupersetSpec, partial any)` (it already works on maps internally);
- the **send** — existing generic `UpdateWorkloadV2` (full manifest replace);
- the **write-back** — `DbUpsertWorkloadV2Snapshot(merged)`.

**Per-type — the only new code each "later" API needs:**
- its `submitter.Update`: call the same composer the type already uses on `Create`;
- (for ws/train/distributed) the `Update<Type>Spec` endpoint, which just calls the shared orchestrator.

So shipping **inference now** builds essentially all the shared machinery; each later type is mostly
"wire its composer into `Update`" (+ an endpoint where one doesn't exist).

---

## Implementation plan (chosen: build the shared orchestrator generically, inference first)

1. **Generalize the merge** — turn inference's `copyUpdateSpecIntoFullSpec` into a shared
   `mergeUpdateSpecIntoFullSpec(full *SupersetSpec, partial any)` (and a master-spec variant for distributed).
2. **Shared orchestrator** — a `updateWorkloadV2(ctx, workloadType, id, partial, submitter)` helper that
   does: native-vs-snapshot resolve → snapshot read → `submitter.Update(partial)` → `sendUpdateToWorkloadMgr`
   → `DbUpsertWorkloadV2Snapshot`. Have `UpdateInferenceSpec` call it.
3. **Pass the partial spec into the submitter** — extend `AnyWorkloadSubmitter.Update(...)` to take the
   partial (currently it isn't passed); update the other stubs to the new signature.
4. **Wire inference** — implement `AnyWorkloadSubmitterInference.Update` (snapshot `ToWorkload` → merge →
   `CopySpecs` → `InferenceToKnativeComposer.Compose` → `WorkloadV2UpdateRequest`).
5. **Reliability** — make `persistWorkloadV2Snapshot` reliable for update-eligible workloads (or fall back
   to an error when the snapshot is missing); add the update write-back.

Later types reuse 1–2–3 and only add their `submitter.Update` (+ endpoint).

## Open items / blockers
- **Best-effort snapshot** today → make reliable, or fall back when missing.
- **Write-back** → snapshot is create-only today; update must upsert.
- **Distributed** → merge `master_spec` + `master_same_as_worker`, not just the worker spec.
- **Distributed-inference** → no snapshot until its V2 *create* path + composer exist (currently `CanCreate=false`).
- **Missing endpoints** → workspaces/trainings/distributed have no spec-update endpoint yet.
- **Dependency** → `workloads_v2_snapshots` must land.

## Key code references
| Concern | Location |
|---|---|
| Snapshot store + DAO + `ToWorkload()` | `control-plane/workloads/pkg/database/{models_crud,dao_crud}/workloads_v2_snapshots.go` |
| Snapshot writer (create) | `services_crud/create.go` → `persistWorkloadV2Snapshot` |
| Inference update orchestration | `services_crud/update_inference_spec.go` |
| Submitters + registry | `services_crud/any_workload_submitter*.go` |
| Merge (to generalize) | `update_inference_spec.go` → `copyUpdateSpecIntoFullSpec` |
| Generic WM update (send) | `services_crud/submit_any_workload.go` → `sendUpdateToWorkloadMgr` → WM `UpdateWorkloadV2` |
| Per-type composers | `composer_v2/{inference_workloads,runai_workloads}/...` |
