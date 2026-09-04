# Inference — phase + phaseMessage

**Engine resource (new flow):** Knative `Service` → Pods.
**Sources:** Knative Service `status` (conditions, `LatestCreatedRevisionName`, `LatestReadyRevisionName`) + Pods (phase, container states) + build error (controller/wlr).
**Reference:** `inference-workload-controller/internal/status/aggregator/workload_phase.go` (`determineWorkloadPhase`, `generatePhaseMessage`, `generatePodPhaseMessage`, `AnalyzeServiceConditions`).

## How we reach each phase

| Phase | Reached when | Source |
|---|---|---|
| **Deleting** | `DeletionTimestamp` set | engine |
| **Creating** | `LatestCreatedRevisionName == ""` (no revision yet) | engine |
| **Pending** | revision created, pods waiting for resources | pods |
| **Initializing** | pods pulling images / container-creating / running-not-ready | pods |
| **Ready** | `LatestReadyRevisionName != ""` AND (*updating* OR all pods ready) | engine + pods |
| **Degraded** | *failure state* AND a previous **ready** revision exists | engine (+wlr build) |
| **Failed** | *failure state* AND **no** ready revision | engine (+wlr build) |

Definitions:
- **updating** = rollout in progress = `LatestReadyRevisionName != LatestCreatedRevisionName` (a newer revision is created but not yet the ready one).
- **failure state** = Knative `Service` condition `Ready == False` **or** `ConfigurationsReady == False` **or** `RoutesReady == False`. Split by ready-revision history: `Ready==False` with **no** prior ready revision → **Failed**; with a prior ready revision → **Degraded**.

## Phase → message → source → derivable

| Phase | phaseMessage | Source | Derivable? |
|---|---|---|---|
| Creating | "" | Knative `LatestCreatedRevision == nil` | ✅ engine |
| Pending | `"Workload is waiting for resources."` | pods Pending | ✅ pods |
| Initializing (pulling) | `"Workload is pulling container images."` | pod `ContainerCreating` | ✅ pods |
| Initializing (image err) | `"Failed to pull container image. Retrying…"` | pod `ImagePullBackOff`/`ErrImagePull` | ✅ pods |
| Initializing (running) | `"Container images pulled successfully. Workload is initializing."` | pod Running, not ready | ✅ pods |
| Ready | "" (or update-progress msg during rollout) | Knative `LatestReadyRevision` + pods ready | ✅ engine+pods |
| Failed (image) | `"Failed to pull container image."` | Knative cond reason "unable to fetch image" | ✅ engine |
| Failed (name too long) | `"The workload name is too long…"` | Knative cond message (63-char) | ✅ engine |
| Failed (timeout) | `"Timed out waiting for the workload to become ready."` | Knative `Failed` cond | ✅ engine |
| Failed (build) | `"Failed to create workload: <buildError>"` | controller build error | ⚠️ wlr/build |
| Degraded | `"Workload is ready. Update failed: …"` | Knative + build error | ⚠️ partly wlr |

Note: most inference failure **reasons live in Knative conditions** (engine) — only the build error needs the wlr.
