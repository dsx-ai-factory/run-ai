# RUN-41388 Design Review Presentation

## Slide 0 — Final Requirements and Scope

Sunset the legacy workload stack without stopping existing workloads.

### Required Behavior

- Every new 3.0+ submission uses WorkloadRunner; there is no typed-CR fallback.
- Existing legacy workloads keep running and remain readable and deletable.
- Existing legacy workload CRDs are preserved during upgrade.
- Fresh installations do not install the six workload CRDs or five policy CRDs.
- The three legacy controllers and their RunaiConfig surfaces are absent from 3.0+; upgrade cleanup removes any exact legacy resources left by the previous release.
- Ownership protection moves to a surviving service before the old controllers disappear.
- Missing legacy CRDs are a valid state for every surviving cluster component.
- Shared resources needed by surviving components are not deleted with the old controllers.

### Scope Boundary

- This document describes fresh 3.0+ installation and upgrade from pre-3.0.
- Downgrade is handled separately; only its manual scenario remains in the E2E task.
- Optional deletion of retained workload CRDs is deferred and is not part of upgrade.

#### Speaker Note

The implementation keeps the data plane running while retiring the legacy management path.

---

<br>
<hr>
<br>

## Slide 1 — Final 3.0+ State

~~~mermaid
flowchart LR
    New["New workload request"] --> WR["WorkloadRunner only"]

    subgraph Removed["Removed from the 3.0+ product surface"]
        WC["workload-controller"]
        IWC["inference-workload-controller"]
        JC["job-controller"]
        LA["controller-only assets"]
    end

    subgraph Retained["Retained after upgrade"]
        CRD["Existing legacy CRDs"]
        CR["Existing workload CRs"]
        Pods["Existing children and pods"]
    end

    CR -->|"read + delete only"| Stale["Stale workload"]
    CR --> Pods
~~~

### What Is Removed

- **workload-controller**
- **inference-workload-controller**
- **job-controller** / RunaiJob controller operand
- Their RunaiConfig fields, defaults, and deployable source trees
- Their controller-only Deployments, Services, monitoring resources, TLS assets, and exact RBAC
- Legacy webhook rules that are not part of the ownership handoff

### What Is Retained

- Existing legacy workload CRDs, CRs, children, and pods on upgraded clusters
- Read-only synchronization of served legacy workloads
- Typed delete for stale workloads
- Ownership validation served by **anyworkload-controller**
- Shared **workloads-auth**, now owned by the AnyWorkload operand

#### Speaker Note

Removal targets the management stack, not the running workload objects.

---

<br>
<hr>
<br>

## Slide 2 — Merged Implementation Map

| PR | Area | Final implemented result |
|---|---|---|
| [#25614 — 01 / Submission Cutover](https://github.com/dsx-ai-factory/run-ai/pull/25614) | Agent, cluster-sync, runtime-rules | WorkloadRunner-only submission, typed legacy delete-only handling, read-only legacy sync, and no legacy runtime actions |
| [#25615 — 02 / Ownership and Webhook Handoff](https://github.com/dsx-ai-factory/run-ai/pull/25615) | Workloads Manager, Agent, AnyWorkload, init-ca | Carries the original submitter to WorkloadRunner and moves ownership webhooks to AnyWorkload |
| [#25616 — 03 / Operator Dependencies](https://github.com/dsx-ai-factory/run-ai/pull/25616) | go-operator, Helm hooks | Moves **workloads-auth** ownership and decouples MPI and Training operators from legacy controller flags |
| [#25617 — 04 / Missing-CRD Readiness](https://github.com/dsx-ai-factory/run-ai/pull/25617) | Helm, CRD upgrader, surviving consumers | Removes all 11 legacy CRDs from installation packages and makes missing APIs safe |
| [#25618 — 05 / Controller Retirement](https://github.com/dsx-ai-factory/run-ai/pull/25618) | go-operator, cleanup Job, Helm hooks, project-controller | Introduces the Operator-managed cleanup Job, safety checks, condition tracking, and exact asset inventory |
| [#26398 — Fresh-Install Defaults Follow-up](https://github.com/dsx-ai-factory/run-ai/pull/26398) | Helm hooks | Historical follow-up that made the transition path independent of target semantic-version parsing |
| [#26564 — Remove Legacy Workload Controllers](https://github.com/dsx-ai-factory/run-ai/pull/26564) | SDK, go-operator, Helm, cleanup Job | Removes all three deployable controllers and RunaiConfig fields, migrates retained configuration to AnyWorkload, and makes cleanup resource-driven |
| [#26571 — Allow RestartRequest Webhook](https://github.com/dsx-ai-factory/run-ai/pull/26571) | Cleanup safety gate | Keeps the live RestartRequest validation webhook outside the retired-resource set |

### Review Result

- All eight implementation and follow-up PRs are merged.
- The presentation now describes the merged code, not the original proposal.
- Control Plane stale-workload UX remains a separate contract and is called out explicitly.

### File-by-File Audit Coverage

| PR | Production surfaces reviewed | Covered in |
|---|---|---|
| #25614 | Agent, cluster-sync cache/controllers/handlers, runtime-rules, go-operator Agent operand, RunaiConfig schema | Slide 3 |
| #25615 | Workloads Manager queue/API, Agent, AnyWorkload ownership, init-ca, OpenShift chart, and legacy ownership handoff | Slide 4 |
| #25616 | Managed operands, AnyWorkload Secret, legacy dependency removal, Helm post hooks, and RunaiConfig rendering | Slide 5 |
| #25617 | CRD packages/copy scripts, cluster-sync policy and workload registration, project blockers, exporter, runtime-rules, and then-transitional controller binaries | Slide 6 |
| #25618 | Initial go-operator cleanup reconciliation, clients, permissions, safety checks, finalizers, cleanup operand, chart RBAC, and RunaiConfig condition | Slides 7–9 |
| #26398 | Historical transition-default behavior later superseded by controller removal | Slides 5 and 11 |
| #26564 | Controller source and operand removal, RunaiConfig schema/config migration, resource detection, direct asset deletion, and final cleanup flow | Slides 5 and 7–11 |
| #26571 | RestartRequest webhook compatibility in the cleanup safety gate | Slide 8 |

#### Speaker Note

The next slides explain each PR's final behavior and how the parts connect.

---

<br>
<hr>
<br>

## Slide 3 — Submission Cutover and Stale Workloads

### 3.0+ Routing Rule

For a 3.0+ cluster, every new submission uses WorkloadRunner. If WorkloadRunner creation or validation fails, the request fails. It never falls back to a typed legacy CR.

The cluster has no WorkloadRunner enable flag in 3.0+. WorkloadRunner is the only submission path.

### Legacy Object Behavior

| Operation | 3.0+ behavior |
|---|---|
| Create a new workload | WorkloadRunner only |
| Read an existing legacy workload | Best-effort read; status may be stale |
| Delete an existing legacy workload | Supported through the existing typed identity |
| Suspend, resume, restart, or update a legacy workload | Rejected |
| Agent receives a typed legacy desired-state update | Delete is accepted; create/update are not recreated as legacy CRs |

### Surviving Cluster Components

- **cluster-sync** watches only legacy APIs that are actually served.
- It continues observation and synchronization but does not write legacy workload status.
- On 3.0+, **runtime-rules** keeps WorkloadRunner behavior but does not start legacy action, time-limit, or auto-delete behavior.
- Missing or deleted legacy objects are treated as already deleted where appropriate.
- The old WorkloadRunner enable value is no longer defaulted, exported to the Agent, or used by 3.0 runtime logic.
- Its RunaiConfig field remains only as inert compatibility data for rollback to 2.26; it is not a 3.0 submission switch.

### Control Plane Contract

- Stale workloads are shown as unmanaged, readable, and delete-only.
- Unsupported actions are hidden in the UI and rejected by the API.
- The final stale identity signal still needs confirmation; **sourceApi is empty** remains a question for Ofer.

#### Speaker Note

The cluster cutover is implemented in #25614. The stale UI and API response remain the Control Plane contract.

---

<br>
<hr>
<br>

## Slide 4 — Ownership and Webhook Handoff

### New Workload Ownership

~~~mermaid
flowchart LR
    User["Authenticated API submitter"] --> WM["Workloads Manager"]
    WM -->|"persist CreatedBy"| Queue[("Workload queue")]
    Queue -->|"WorkloadV2Update.createdBy"| Agent["runai-agent"]
    Agent -->|"owner annotation"| WR["WorkloadRunner"]
    WR -->|"UPDATE / DELETE"| AWC["AnyWorkload ownership webhook"]
~~~

- The Kubernetes creator is **runai-agent**, so admission identity alone is insufficient.
- Workloads Manager carries immutable **CreatedBy** through the queue.
- The Agent stamps the owner annotation on WorkloadRunner.
- Existing Control Plane-managed WorkloadRunners missing the annotation are lazily backfilled by the trusted Agent.
- The WorkloadRunner owner annotation is immutable after initialization.

### Legacy Ownership Handoff

- Ownership validation moves from the retired controllers to **anyworkload-controller**.
- It covers Training, Interactive, Distributed, Inference, and Distributed Inference workloads.
- Validation protects UPDATE and DELETE.
- Mutation protects UPDATE and restores the stored legacy **spec.userId** when changed or cleared.
- RunaiJob is not included in this ownership handoff.

### Authorization Decision

- The stored owner can update or delete the workload.
- Authorized internal service accounts and users with the required privileged access can continue.
- The trusted Agent may initialize missing ownership on a Control Plane-managed WorkloadRunner.
- An unowned WorkloadRunner may be deleted; other ownership changes remain protected.

### Webhook Desired State

- Only exact ownership rules may continue targeting served legacy workload APIs.
- Those rules must target **runai/anyworkload-controller**.
- Legacy non-ownership webhook rules are omitted.
- Rules for missing CRDs are filtered out by API discovery.
- On OpenShift, the surviving ownership webhook configurations use the service-CA injection flow.
- Cleanup stops any surviving legacy controller pods before validating the final webhook handoff.

#### Speaker Note

#25615 completes both sides of the handoff: owner data for new WorkloadRunners and surviving admission for retained legacy CRs.

---

<br>
<hr>
<br>

## Slide 5 — Fresh Installation and Operator Dependencies

### Fresh 3.0+ Installation

~~~mermaid
flowchart LR
    Install["Fresh installation path"] --> Render["Drop legacy controller config sections"]
    Render --> Desired["Render only surviving operands"]
    Desired --> Ready["Cluster starts without legacy controllers or 11 legacy CRDs"]
~~~

- The workload-controller, inference-workload-controller, and job-controller fields no longer exist in the 3.0 RunaiConfig schema.
- Helm removes those three sections from rendered RunaiConfig instead of setting controller flags.
- Their go-operator operands, deployable source trees, and controller-only chart resources are absent.
- No cleanup Job is requested when no exact legacy controller resources are detected.
- Legacy controller RBAC and Project RoleBindings are not rendered on a fresh installation.
- **legacyProjectPermissions** remains a separate access-rule-controller setting; controller retirement no longer changes it.

### Shared Dependencies That Must Survive

- **workloads-auth** is owned by the AnyWorkload operand.
- Existing Secret identity and data are adopted and preserved during the move.
- AnyWorkload health checks the live Secret and its client ID, client secret, cookie secret, and Redis password.
- MPI and Training operators reconcile independently from retired workload-controller configuration.
- Legacy networking, OAuth proxy, and image fields migrate only when the corresponding AnyWorkload field is unset; shared GPU-network acceleration preserves an enabled legacy requirement.
- Helm then removes the legacy controller sections, so old cluster-config values cannot recreate retired operands.

### Final Follow-Up

#26398 removed target-version parsing from the interim transition-default flow. #26564 then removed that flow entirely by deleting the legacy RunaiConfig fields and controller operands.

#### Speaker Note

#25616 moved shared dependencies; #26564 made controller absence structural rather than flag-driven.

---

<br>
<hr>
<br>

## Slide 6 — Legacy CRDs and Missing-API Readiness

### CRDs Removed From Installation Packages

Six workload CRDs:

- **trainingworkloads.run.ai**
- **interactiveworkloads.run.ai**
- **distributedworkloads.run.ai**
- **inferenceworkloads.run.ai**
- **distributedinferenceworkloads.run.ai**
- **runaijobs.run.ai**

Five policy CRDs:

- **trainingpolicies.run.ai**
- **interactivepolicies.run.ai**
- **distributedpolicies.run.ai**
- **inferencepolicies.run.ai**
- **distributedinferencepolicies.run.ai**

### Installation Semantics

| Path | CRD behavior |
|---|---|
| Fresh 3.0+ install | The 11 manifests are absent, so the CRDs are never installed |
| Upgrade from pre-3.0 | Existing CRDs remain because removing manifests does not delete live CRDs |
| Future cleanup | Explicit, opt-in, outside upgrade, and not yet scheduled |

### Surviving Consumers

- **cluster-sync** registers only served legacy workload and policy APIs.
- **workload-exporter** registers metrics controllers only for the five served typed legacy workload APIs; RunaiJob is not an exporter controller.
- **project-controller** treats a missing legacy API as no blockers, but real permission and API errors still block.
- Policy controllers skip missing policy APIs.
- The workload, inference-workload, and RunaiJob controller binaries and operands are removed; retained shared API packages remain available to surviving consumers.
- Legacy controller migrations are no longer started because their deployable controller processes no longer exist.
- DistInfW and RunaiJob remain project deletion blockers only when their APIs and objects exist.
- Only an absent API is optional. Discovery, permission, or list failures are not silently ignored.

#### Speaker Note

#25617 made missing APIs safe; #26564 subsequently removed the three transitional controller processes themselves.

---

<br>
<hr>
<br>

## Slide 7 — Upgrade Sequence

~~~mermaid
flowchart LR
    A["1. Normal upgrade"] --> B["2. Existing post-upgrade actions"]
    B --> C["3. Migrate retained config to AnyWorkload"]
    C --> D["4. Remove legacy RunaiConfig sections"]
    D --> E{"5. Exact legacy resources exist?"}
    E -->|"No"| F["No cleanup Job or condition"]
    E -->|"Yes"| G["6. Wait for RunaiConfig Ready"]
    G --> H["7. Operator creates cleanup Job"]
    H --> I{"8. Cleanup succeeds?"}
    I -->|"No"| J["Condition = Failed<br/>legacy assets may remain"]
    I -->|"Yes"| K["9. Assets removed<br/>Condition = Completed"]
~~~

### Trigger Contract

- The Operator detects cleanup need from an exact allowlist of live controller Deployments, Services, ServiceAccounts, TLS Secrets, RBAC, monitoring resources, legacy admission configurations, and Project RoleBindings.
- Detection tolerates NotFound and unavailable optional APIs, but other read failures stop reconciliation.
- If no exact legacy resource exists, the cleanup condition is cleared and no Job is created.
- If resources exist, the Operator waits for the current RunaiConfig generation to be Reconciled and Available before creating the Job.
- No source-version annotation, controller-state classification, or cleanup request flag participates in the final trigger.

### Important Correction From the Original Design

The cleanup is not a Helm post-upgrade hook and does not block the Helm upgrade. The continuously running go-operator derives cleanup from live cluster state after the normal upgrade.

#### Speaker Note

Resource presence is the initial trigger. The Job removes exact legacy assets asynchronously without depending on deleted controller configuration fields.

---

<br>
<hr>
<br>

## Slide 8 — Cleanup Job Contract

~~~mermaid
flowchart LR
    Start["1. Load live RunaiConfig"] --> Discover["2. Discover which 6 workload CRDs are served"]
    Discover --> Access["3. Verify management + workload permissions"]
    Access --> Stop["4. Delete 3 controller Deployments"]
    Stop --> Wait["5. Wait until controller pods stop"]
    Wait --> Webhooks["6. Validate surviving webhook state"]
    Webhooks --> Finalizers["7. Remove allowlisted finalizer from TW / IW / DW"]
    Finalizers --> Assets["8. Delete exact legacy assets"]
    Assets --> Done["9. Job completes"]
~~~

### What the Job Checks

- Each legacy workload CRD is discovered individually; a missing CRD is allowed.
- Management access can read RunaiConfig and create SelfSubjectAccessReviews.
- Workload-cluster access can discover CRDs, list and patch the three finalizer-bearing workload types, stop legacy controller pods, and delete the exact asset allowlist.
- No unexpected webhook rule targets a retired API.
- If ownership protection is enabled, every served owned legacy API is covered by an exact AnyWorkload ownership rule.
- If ownership protection is disabled, no webhook rule may continue targeting retired APIs.
- **runai-restartrequest-validation** remains valid because RestartRequest is served by AnyWorkload and is not a retired resource.

### What the Job Changes

- Deletes the **workload-controller**, **inference-workload-controller**, and **runai-job-controller** Deployments, then waits until their pods stop.
- Removes only **run.ai.workload/finalizer** from existing Training, Interactive, and Distributed Workloads.
- Preserves every other finalizer.
- Does not change finalizers on InferenceWorkload, DistributedInferenceWorkload, or RunaiJob.
- Does not wait for workloads to terminate or disappear.
- Deletes an exact allowlist of legacy Services, ServiceAccounts, TLS Secrets, Roles, RoleBindings, ServiceMonitors, VPAs, ClusterRoles, ClusterRoleBindings, and retired admission configurations.
- Deletes exact legacy Project RoleBindings identified by name, roleRef, and legacy controller ServiceAccount subject.
- Does not update RunaiConfig or delete legacy workload CRDs, workload CRs, children, or workload pods.

### Job Lifecycle

- Name: **runai-legacy-workload-cleanup**
- Uses the deployed go-operator image with **--legacy-workload-cleanup**.
- Uses a dedicated ServiceAccount and RBAC.
- **backoffLimit: 3**; there is no active deadline in the final Job spec.
- A failed Job sets the cleanup condition to **Failed** and may leave part of the exact asset inventory for retry.
- Deleting a failed Job allows the Operator to recreate it and retry while cleanup is incomplete.
- No source-version annotation or separate **legacyWorkloadCleanup** enable/request field exists; live resource presence, Job state, and the derived RunaiConfig condition drive the lifecycle.

### Cluster Placement

- On a single cluster, management and workload operations use the same Kubernetes API.
- When **NMC.RemoteAPIServerURL** is set, the Job reads RunaiConfig on the management cluster.
- Legacy-resource detection and deletion, controller shutdown, CRD discovery, workload listing and patching, webhook validation, and their permission checks run against the remote workload cluster.
- Management and workload permissions are checked separately before cleanup.

#### Speaker Note

The final Job is explicit and idempotent: it shuts down the retired controllers, preserves the surviving ownership and RestartRequest webhooks, prepares retained workloads, and deletes only allowlisted controller assets.

---

<br>
<hr>
<br>

## Slide 9 — Cleanup State and Asset Convergence

### RunaiConfig Condition

**LegacyWorkloadCleanupComplete** is derived from exact legacy-resource presence, RunaiConfig readiness, and Job state.

| Reason | Meaning |
|---|---|
| **WaitingForReady** | The upgraded RunaiConfig has not reached the reconciled and available generation yet |
| **Creating** | The Operator should create the cleanup Job |
| **Running** | The Job exists and is still running |
| **Failed** | The Job failed; deleting it permits an idempotent retry |
| **Completed** | The Job completed after removing the exact legacy asset inventory |

Before completion, if no cleanup Job and no exact legacy resource exist, the condition is cleared. Once **Completed**, it remains terminal. The **Blocked** reason constant remains in the SDK but is not emitted by the final resource-driven reconciler.

### Final Asset Boundary

~~~mermaid
flowchart LR
    Detect["Exact legacy resource detected"] --> Job["Cleanup Job"]
    Job --> Stop["Controllers stopped"]
    Job --> Delete["Controller assets + retired webhooks deleted"]
    Job --> Prepare["TW / IW / DW finalizer removed"]
    Stop --> End["Condition = Completed"]
    Delete --> End
    Prepare --> End
    Retained["Legacy CRDs, CRs, children, pods"] -. "retained" .-> End
~~~

- The legacy controller operands and the separate legacy-controller-RBAC operand no longer exist.
- Helm and go-operator no longer render controller Deployments or controller-only RBAC.
- The cleanup Job performs the explicit deletion needed for resources left by an older installation.
- Project-controller no longer generates controller-specific RoleBindings; the Job removes exact legacy bindings left in existing Project namespaces.
- Cleanup is retryable and tolerates resources already being absent.
- **workloads-auth** is safe because it is already owned by the AnyWorkload operand.

#### Speaker Note

The Job converges the retired asset inventory itself. The retained workload objects remain available for observation and delete-only handling.

---

<br>
<hr>
<br>

## Slide 10 — Existing Workloads After Controller Retirement

### Workload State

- Existing legacy CRs remain in Kubernetes.
- The cleanup Job does not issue delete requests for existing child resources or pods.
- If a workload was already terminating, removing its legacy finalizer allows the existing Kubernetes deletion flow to continue.
- Running workloads continue running but become unmanaged.
- Their displayed status may become stale because their old controllers no longer reconcile them.
- Users can read and delete them; other management actions are unsupported.

### Delete Behavior

- Typed delete uses the workload's existing stored Kubernetes identity.
- Agent delete treats NotFound or an unserved legacy API as already deleted.
- Kubernetes garbage collection continues to use the ownerReferences already present on existing objects.
- The cleanup flow does not repair arbitrary ownerReferences.

### Read-Only Observation

~~~mermaid
flowchart LR
    Legacy["Served legacy CR"] --> Sync["cluster-sync"]
    Sync --> CP["Control Plane read model"]
    CP --> UI["Stale workload: read + delete"]
    Rules["runtime-rules"] -. "no legacy actions" .-> Legacy
~~~

#### Speaker Note

The design deliberately separates observation from management: the system can still show an old workload without continuing to control it.

---

<br>
<hr>
<br>

## Slide 11 — Fresh Install vs Upgrade

| Area | Fresh 3.0+ installation | Upgrade from pre-3.0 |
|---|---|---|
| Legacy controller configuration | Three RunaiConfig sections are omitted | Retained settings migrate to AnyWorkload, then the legacy sections are removed |
| Cleanup trigger | No exact legacy resources, so no request | Presence of exact legacy controller resources |
| Cleanup Job | Not created when no legacy resources exist | Created and monitored by go-operator after RunaiConfig is ready |
| Six workload CRDs | Not installed | Existing CRDs remain |
| Five policy CRDs | Not installed | Existing CRDs remain |
| Existing workloads | None | Remain stale, readable, and delete-only |
| Legacy ownership rules | None when APIs are absent | Exact rules move to AnyWorkload for served APIs |
| Controller processes and assets | Not rendered | Deployments are stopped first; exact remaining assets are deleted by the Job |
| Legacy Project RoleBindings | Not rendered | Exact controller bindings are deleted by the Job |
| Controller-only RBAC | Not rendered | Exact legacy RBAC is deleted by the Job |
| Shared **workloads-auth** | Created by AnyWorkload | Adopted by AnyWorkload without losing existing data |

### Platform Difference

- The core flow is identical for OpenShift and non-OpenShift Kubernetes.
- Non-OpenShift uses the normal init-ca certificate and CA-bundle flow.
- OpenShift uses service-CA injection for the surviving ownership webhook configurations.
- Both platforms omit legacy admission configurations and rules that no longer have a surviving owner.
- NMC uses the same state machine, with RunaiConfig on the management cluster and legacy workload cleanup routed to the configured workload cluster.

#### Speaker Note

Fresh install starts directly in the final state. Upgrade uses resource-driven cleanup because controller assets from the previous release can still exist even though their configuration and operands are gone.

---

<br>
<hr>
<br>

## Slide 12 — Deferred Work and Open Questions

### Deferred Work

- Optional deletion of empty legacy workload CRDs is outside upgrade and is not scheduled yet.
- Downgrade implementation is outside this design; the manual downgrade scenario stays in the E2E task.

### Open Questions

1. Confirm the Control Plane stale-workload identity signal, including whether **sourceApi is empty** is sufficient, with Ofer.
2. Confirm the final UI label and API error for unsupported stale-workload actions.
3. Confirm MNNVL and existing ComputeDomain behavior after legacy controllers stop reconciling.
4. Decide which later release, operator setting, and empty-cluster checks will own optional CRD deletion.

### Implementation Validation

- [x] WorkloadRunner-only submission with no legacy fallback
- [x] Delete-only Agent behavior for typed legacy workloads
- [x] Read-only legacy synchronization without legacy runtime actions
- [x] Ownership and webhook handoff to AnyWorkload
- [x] Shared Secret and independent operator ownership fixed
- [x] All 11 legacy CRDs removed from fresh-install packages
- [x] Surviving consumers tolerate missing legacy APIs
- [x] All three deployable legacy controllers and RunaiConfig fields removed
- [x] Legacy workload configuration migrated to AnyWorkload with field-specific precedence
- [x] Resource-driven Operator cleanup Job and condition tracking
- [x] Controller shutdown and exact legacy asset deletion
- [x] RestartRequest webhook preserved outside the retired-resource set
- [x] Allowlisted TW/IW/DW finalizer removal
- [ ] Control Plane stale UX and identity contract confirmed end to end
- [ ] Full manual E2E matrix completed, including the separately owned downgrade scenario

#### Speaker Note

The cluster implementation, controller removal, and cleanup follow-ups are merged. The remaining items are product-contract confirmation, E2E validation, and deliberately deferred CRD cleanup.
