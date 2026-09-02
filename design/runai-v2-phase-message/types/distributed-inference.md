# Distributed Inference — phase + phaseMessage

**Engine resource (new flow):** `LeaderWorkerSet` (LWS) → Pods.
**Sources:** LWS `status` (replicas, ready) + per-group/pod phases + build error / `ComputeDomain` (MNNVL) for cause.
**Reference:** `inference-workload-controller/internal/status/distributed/{distributed_status.go,lws_status.go}` (`determineLwsPhase`, `GetPhaseMessage`, `determinePodPhase`).

## How we reach each phase

| Phase | Reached when | Source |
|---|---|---|
| **Deleting** | `DeletionTimestamp` set | engine |
| **Creating** | `lws == nil` (LWS not created yet) | engine |
| **Pending** | `PendingCount == Replicas` — every group is *pending*: its leader pod is either **not yet created** (no Pod resource) **or** has `pod.status.phase == Pending` | engine + pods |
| **Initializing** | `ReadyCount != Replicas` (some groups not ready) | engine + pods |
| **Ready** | `ReadyCount == Replicas` | engine + pods |
| **Failed** | `FailedCount>0` AND `ReadyCount==0` | engine + pods |
| **Degraded** | `FailedCount>0` AND `ReadyCount>0` | engine + pods |

Per-group/pod phase comes from pod `.Status.Phase` + container waiting reasons (`ContainerCreating`/`ImagePullBackOff`/`CreateContainerError`…) and the leader (index `0`) readiness.

## Phase → message → source → derivable

| Phase | phaseMessage | Source | Derivable? |
|---|---|---|---|
| Creating | "" | `lws == nil` | ✅ engine |
| Pending | `"Replicas: N; Pending: N"` | LWS replicas + pods | ✅ engine+pods |
| Initializing | `"Replicas: N; Ready: x; Initializing: y; …"` | LWS replicas + pod phases | ✅ engine+pods |
| Ready | `"Replicas: N; Ready: N"` | ready == replicas | ✅ engine+pods |
| Failed | `"Replicas: N; Failed: x"` | failed>0 & ready==0 | ✅ engine+pods |
| Degraded | `"Replicas: N; Ready: x; Failed: y"` | failed>0 & ready>0 | ✅ engine+pods |
| **Failed/Degraded (cause)** | build error / MNNVL not ready / missing resources | build err / `ComputeDomain` / resources | ⚠️ wlr/external |

Note: replica-count messages are **fully engine-derivable**; only the underlying *cause* (build/MNNVL) needs the wlr/external.
