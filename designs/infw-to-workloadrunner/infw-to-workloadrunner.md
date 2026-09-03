# InfW → WorkloadRunner — Inference Migration

> **Scope:** Single-replica `InferenceWorkload` (InfW) only. `DistributedInferenceWorkload` (DistInfW) is the [sibling doc](./dist-infw-to-workloadrunner.md).
>
> **Companions:** [`infw-fields.html`](./infw-fields.html) — filterable field-by-field mapping table · [`infw-to-workloadrunner.slides.html`](./infw-to-workloadrunner.slides.html) — slide deck.
>
> **⚠ Schema-version note.** Every YAML example in this design uses the **target** `WorkloadRunner` schema — `instructions.networking.endpoints[]` for the Service + Ingress layer. **The networking plugin has no `InPod` mode.** Sidecar auth (what we've been calling "InPod") is composed at the **control-plane composer** level via Liran's design (still in review): the OAuth2-proxy sidecar is baked directly into the Knative Service template's pod spec (`spec.template.spec.template.spec.containers[]`), alongside the user container. **There is no `instructions.initContainer.*` field on `WorkloadRunner` — that's a CP composer responsibility, not a runtime plugin.** The networking plugin only sees the Service + Ingress shape and is unaware of the sidecar. The schema in `master` today is the v1 flat shape (`instructions.networking.services[]`, `MaxItems: 1`, workload-level `externalAccess`); nothing migrates until both [PR #19525](https://github.com/dsx-ai-factory/run-ai/pull/19525) (multi-endpoint networking) **and** Liran's CP composer design land — see §7.1.

## 1. TL;DR

1. **Composer is the heavy lift.** Once `InferenceFields` are translated to Knative annotations and the right `spec.template` is in place, Knative handles the runtime behaviour natively. We don't re-implement autoscaling, revisions, etc.
2. **The serving port pattern doesn't change.** Legacy made the OAuth2-proxy sidecar port-bearing so Knative's auto-Service routes through it. On the new path the **CP composer** bakes the sidecar directly into the Knative Service template's pod spec (Liran's design — in review) — the networking plugin is unaware. No new `corev1.Service` for InfW.
3. **Tools (`exposedUrls`) need multi-endpoint.** Legacy injects one OAuth2-proxy sidecar **per URL**, each with its own auth. The new path requires PR #19525 multi-endpoint support to land — that's the blocking dependency.

## 2. Architecture

```
[ User ]  POST runai API
      │
      ▼
[ Control-Plane composer ]  reads Knative-Service Karta, translates InferenceFields
      │  writes ↓
      ▼
WorkloadRunner CR
  ├─ spec.template:        serving.knative.dev/v1.Service
  └─ spec.instructions:    typed plugin inputs
                           (networking.endpoints[], gpuOptimizations, storage, …)
      │
      ▼
[ anyworkload-controller ]  applies the template, runs plugins
      │
      ▼
Knative Service ── Configuration ── Revision(s) ── Pods
                                                   ↑
                              OAuth2-proxy sidecar (composed into template by CP), port-bearing
```

The composer's responsibility ends at "produce a complete WLR." The cluster-side controller knows nothing about inference — it follows the Karta for the target kind, applies plugins, and reconciles.

### 2.1 Plugin landscape

| Plugin | Status | Drives | Responsible for |
|---|---|---|---|
| `networking` | v1 shipped · v2 pending ([PR #19525](https://github.com/dsx-ai-factory/run-ai/pull/19525)) | `instructions.networking` | Services + Ingress / Route / HTTPRoute. **No `InPod` mode and no sidecar awareness** — the OAuth2-proxy sidecar is composed into the template by the CP composer. **InfW depends on PR #19525 (multi-endpoint).** |
| `mnnvl` | shipped | `instructions.gpuOptimizations.mnnvl` | N/A for single-replica InfW |
| `storage` | design | `instructions.storage.newPvcs` | Lifecycle-bound new PVCs |
| `personal-credentials` | design | per-user secret synthesis | Per-user image-pull / env-var creds |

> **Not a plugin: CP composer responsibilities (Liran's design — in review).** Init-containers (`gitSync`, `s3`) and the OAuth2-proxy **sidecar** for auth are *not* a runtime plugin. The CP composer writes them directly into the template's pod spec (`spec.template.spec.template.spec.{initContainers,containers}` for Knative; `leaderTemplate.spec.{initContainers,containers}` for LWS). There is no `instructions.initContainer.*` field on `WorkloadRunner`. **InfW depends on this design landing.**

### 2.2 The Knative-Service Karta

The Karta for `serving.knative.dev/v1.Service` **already ships** via control-plane DB migrations — see `BuildKNativeServiceV1()` in [`control-plane/workloads/pkg/database/go_migrations/ri/source/knative.go:21`](../../control-plane/workloads/pkg/database/go_migrations/ri/source/knative.go). Shape:

```go
// trimmed for readability — see knative.go for the full source
StructureDefinition: {
  RootComponent: {
    Name: "knativeservice",
    Kind: { Group: "serving.knative.dev", Version: "v1", Kind: "Service" },
    StatusDefinition: { ConditionsDefinition + StatusMappings.Running on Ready/RoutesReady/ConfigurationsReady },
  },
  AdditionalChildKinds: [ { apps/v1.Deployment } ],
  ChildComponents: [
    {
      Name: "revision",
      Kind: { serving.knative.dev/v1.Revision },
      OwnerRef: "knativeservice",
      SpecDefinition: { PodTemplateSpecPath: ".spec.template" },
      ScaleDefinition: {
        MinReplicasPath: '.spec.template.metadata.annotations["autoscaling.knative.dev/min-scale"] // 1',
        MaxReplicasPath: '.spec.template.metadata.annotations["autoscaling.knative.dev/max-scale"]',
      },
    },
  ],
},
Instructions: {
  GangScheduling: { revision pods grouped by label "serving.knative.dev/revision" },
},
```

Key shape notes:

- **Root has no spec/pod paths.** All the pod-template plumbing hangs off the `revision` **child component** via `PodTemplateSpecPath: ".spec.template"` — the whole-template form, not the fragmented one.
- **No `Configuration` / `Revision` in `AdditionalChildKinds`** — only `apps/v1.Deployment`. The `revision` is a proper child component with its own scale definition.
- **Scale definition lives on the child**, not the root — it reads the autoscaling annotations off `.spec.template.metadata.annotations`.

**Anything the Karta doesn't expose is the composer's / builder's problem, not the Karta's.** We will not extend the Karta as part of this migration. If a field (e.g. autoscaling annotations, containerConcurrency, timeoutSeconds) doesn't have a dedicated path in the Karta, the InfW composer simply writes it into the right place inside `spec.template` directly. The DistInfW Karta (LWS) gets the same treatment — anything missing lives in that builder.

## 3. The WorkloadRunner CR — what the composer emits

```yaml
apiVersion: run.ai/v1alpha1
kind: WorkloadRunner
metadata:
  name: my-llm
  namespace: runai-team-a
  labels:
    runai/project: team-a
    runai/workload-kind: Inference
spec:
  template:
    apiVersion: serving.knative.dev/v1
    kind: Service
    spec:
      template:
        metadata:
          annotations:
            # InferenceFields.AutoScalingFields → Knative annotations
            autoscaling.knative.dev/class: kpa.autoscaling.knative.dev
            autoscaling.knative.dev/metric: rps
            autoscaling.knative.dev/target: "100"
            autoscaling.knative.dev/min-scale: "0"
            autoscaling.knative.dev/max-scale: "5"
            autoscaling.knative.dev/scale-to-zero-pod-retention-period: "200s"
        spec:
          containers:
          # ── CP composer injects the OAuth2-proxy sidecar here when auth is on ──
          - name: oauth2-proxy
            image: quay.io/oauth2-proxy/oauth2-proxy:...
            args: [--upstream=http://localhost:8080, --http-address=0.0.0.0:4180, ...]
            ports: [{ containerPort: 4180, name: h2c }]      # port-bearing → Knative routes here
          # ── user container ──
          - name: model
            image: my-llm:latest
            ports: [{ containerPort: 8080 }]                  # NOT port-bearing under the sidecar
            resources: { limits: { nvidia.com/gpu: "1", memory: 16Gi } }
          containerConcurrency: 0   # concurrencyHardLimit
          timeoutSeconds: 300       # requestTimeoutSeconds
  instructions:
    networking:                                   # network plugin: Service + Ingress only
      endpoints:
      - services:
        - ref: { name: my-llm, port: 80 }         # Knative auto-Service (targets 4180 = proxy)
        externalAccess:
          ingress: { spec: { rules: [...] } }
```

The auth fields (`authorizationType`, `authorizedUsers`, `authorizedGroups`) drive the **composer's** decision to inject the sidecar and to render its args. There is no auth block on `instructions.networking.endpoints[]` and no `instructions.initContainer.*` field. The sidecar simply appears in the template's `containers[]` next to the user container.

## 4. Auth-vocabulary mapping

The one place where translation is non-trivial: serving-port auth.

| InfW source | WLR `auth.authorizationType` | WLR list field |
|---|---|---|
| `public` | — (omit the auth block) | — |
| `authenticatedUsers` | `authenticatedUsers` | — |
| `authorizedUsersOrGroups` + users | `authorizedUsers` | `auth.authorizedUsers` |
| `authorizedUsersOrGroups` + groups | `authorizedGroups` | `auth.authorizedGroups` |

A workload that sets *both* the users-list and the groups-list under `authorizedUsersOrGroups` is ambiguous against the WLR contract — CP should reject with a validation error.

Source: `servingPort.authorizationType` + `servingPort.authorizedUsers[]` / `servingPort.authorizedGroups[]` on the Inference API DTO (`model_serving_port.go:26-30`).

## 5. End-to-end scenarios

For each scenario: what each layer creates, today vs. new. The principle: **don't create resources the target template already provides** — `services[].oneOf { create, ref }` is the lever.

### 5.1 Basic InfW (no auth, no tools, no NIM)

- **Knative auto-creates** a Service targeting the user-container's port.
- **Composer emits** `endpoints[0]` with `services[].ref` at the Knative auto-Service.
- **Plugin creates** only an Ingress / Route / HTTPRoute pointing at that referenced Service.
- **No sidecar, no new `corev1.Service`.**

```
external → Ingress → Knative auto-Service → user container
```

### 5.2 InfW + serving-port auth (sidecar composed by CP)

- **Knative auto-creates** the same Service. The composed-in sidecar becomes **port-bearing**, so the auto-Service routes through the proxy.
- **Composer (CP) does two things:**
  - Bakes the OAuth2-proxy as an extra container into `spec.template.spec.template.spec.containers[]` next to the user container (port-bearing, configured from the auth fields). This is the part of Liran's design (in review) we hard-depend on.
  - Emits `instructions.networking.endpoints[0]` with `services[].ref` at the Knative auto-Service + `externalAccess.ingress` (networking plugin: Service + Ingress).
- The **networking plugin is unaware** of the sidecar. The Knative auto-Service routes to whichever container is port-bearing — the composer puts the proxy there when auth is on.
- **No new `corev1.Service`.** Matches legacy exactly (`inference-workload-controller/internal/builders/utils.go:458-481`).

```
external → Ingress → Knative auto-Service → OAuth2-proxy sidecar → user container
```

### 5.3 InfW + NIM (label-triggered)

- **Trigger:** the `runai.nim/workload=true` label on the workload (not a spec field).
- **Composer** reads the label and:
  - Propagates it to the Knative pod-template labels (matches legacy `label_utils.go:19-30` `CopyNimLabel`).
  - Emits a second endpoint for the metrics port: `endpoints[1]` with `services[].create` (port-bearing, the metrics port), **no `externalAccess`** (scrape-only).
- **Plugin** creates the metrics `corev1.Service`. **No sidecar on the metrics endpoint** — NIM metrics bypass auth by legacy policy.

```
serving:  external → Ingress → ksvc auto-Service → (sidecar →) user container
metrics:  Prometheus → metrics Service → user container (port 8000)
```

Legacy reference: `inference-workload-controller/internal/builders/utils.go:562-617` (`BuildNIMMetricsService`).

### 5.4 InfW + Tools (`exposedUrls`, N entries each with own auth)

This is the multi-endpoint case. Each `exposedUrls[]` entry has its own port and its own auth.

- **Composer (CP):**
  - **Bakes one OAuth2-proxy sidecar per URL** into the Knative pod template's `containers[]` — each with its own proxy port and its own cookie prefix (same shape as legacy `workload-common/builders/oauth2_proxy.go:303` `ComposeOauthIngressUrl` at `inference-workload-controller/internal/builders/utils.go:483-504`).
  - Emits `instructions.networking.endpoints[i]` per `exposedUrls[i]` — each with `services[].ref` + `externalAccess.ingress` (networking plugin: per-URL Service + Ingress, `customUrl` honoured).

```
tool-1: external → Ingress → auto-Service:portA → sidecar-1 → user container
tool-2: external → Ingress → auto-Service:portB → sidecar-2 → user container
…
```

**Dependency:** This scenario does not work until PR #19525 (multi-endpoint networking) **and** Liran's CP composer design both land. See §7.1.

## 6. What each layer creates — at a glance

| Layer | Basic | + Auth | + NIM | + Tools (N URLs) |
|---|---|---|---|---|
| Knative auto-creates | Service → container | Service → sidecar → container | same as basic / auth | same as basic / auth (per-port routing via sidecars) |
| Composer emits | `endpoints[].ref` | `endpoints[].ref` + injects OAuth2-proxy sidecar in template `containers[]` | + `endpoints[metrics].create` (no ext.) | one `endpoints[i]` per URL + one composed-in sidecar per URL |
| Plugin: `corev1.Service` | — | — | metrics Service | — (auto-Service already routes) |
| Plugin: Ingress | yes | yes | yes for serving · no for metrics | yes (per endpoint) |
| Composer: sidecar(s) in template | — | 1 (port-bearing) | 1 if auth is on | N (one per URL) |

**InfW never gets a runai-created `corev1.Service` for the serving path** — auth or no auth. Knative's auto-Service is sufficient because the OAuth2-proxy sidecar (composed into the template by the CP composer) becomes port-bearing in front of the user container. Matches legacy.

## 7. Dependencies / sibling work

| What | How it touches InfW |
|---|---|
| Epic — sunset TW/IW/DW/InfW/RunaiJob | Parent |
| WorkloadRunner status / phase alignment | Phase mapping decision (status is out of scope for this doc) |
| TW/IW/RunaiJob CP-to-Batch | Sibling design; same composer vocabulary; OAuth2-proxy diverges (sidecar here vs Deployment there) |
| TW/IW/DW gap matrix | Generic field gaps tracked there, not here |
| Personal Credentials Plugin | Owns `imagePullSecrets` + env-creds. **Action:** composer iterates Karta `childComponents` for multi-component CRDs (LWS); InfW is single-pod so this is a no-op for our composer, but the plugin must still synthesize. |
| Storage Plugin | Owns new-PVC lifecycle (`pvcs`, `dataVolumes`) |
| **CP composer feature** (Liran, in review) — composes init-containers + sidecars (`gitSync`, `s3`, OAuth2-proxy) directly into the target template's pod spec. **Not a runtime plugin.** | **Blocking** — owns OAuth2-proxy sidecar composition. Without it, InfW auth can't be reproduced (no `InPod` mode in the networking plugin). |
| Networking plugin redesign — Multi-Endpoint ([PR #19525](https://github.com/dsx-ai-factory/run-ai/pull/19525)) | **Blocking** for tools / `exposedUrls`. Network plugin will only handle Service + Ingress; sidecar auth is the CP composer's job (no `InPod` mode in networking). |

### 7.1 The sidecar dependency is hard

InfW auth needs an OAuth2-proxy **sidecar** in the Knative pod template that becomes port-bearing in front of the user container — that's how Knative's auto-Service routes through the proxy and how legacy InfW already works today. Two consequences:

- **There is no `InPod` mode in the networking plugin.** Sidecar injection is the **CP composer's** job, via Liran's design. The networking plugin only creates Services + Ingress and is unaware of the sidecar. There is no `instructions.initContainer.*` field — the sidecar appears directly in `spec.template.spec.template.spec.containers[]`.
- **`Standalone` auth (separate Deployment + Service) — the networking plugin's only current auth mode — is not a viable substitute for inference.** With Standalone the Knative auto-Service still targets the unprotected user container, regressing every existing InfW. And for **tools** (`exposedUrls`), Standalone would require one extra Deployment + Service per URL, all separately wired into Ingress — operationally much worse than the per-URL sidecar legacy already ships.

**Decision: we hard-block on Liran's CP composer design landing.** No fallback to Standalone for inference.

## 8. Deliverables for InfW migration

1. **CP composer (InfW → WLR)** — translate every field per the [companion field-mapping table](./infw-fields.html). Heaviest of the composers — most of the Knative-annotation translation is here. **The composer also owns anything the Karta doesn't surface** (autoscaling annotations, `containerConcurrency`, `timeoutSeconds`, etc.) — write straight into `spec.template`.
2. **Knative-Service Karta** — already ships (`BuildKNativeServiceV1`). No work here. If the migration uncovers gaps, fix them in the composer/builder, not in the Karta.
3. **Dependencies** (sibling tracks):
   - **Networking plugin (PR #19525)** — multi-endpoint (`maxItems > 1` on `services[]` + per-endpoint `externalAccess`) — needed for `exposedUrls` and the metrics endpoint when NIM is on. **No InPod / sidecar work here.**
   - **CP composer init/sidecar feature (Liran — in review)** — composes git-sync init-containers, S3 init-containers, and the OAuth2-proxy sidecar (one per serving-port + one per exposed URL) directly into the target template's pod spec. **Not a runtime plugin; lives in the CP composer.**
4. **CP composer networking behaviour** (ours):
   - Emit `instructions.networking.endpoints[]` per the multi-endpoint schema — one endpoint per serving-port + one per `exposedUrls[]` entry.
   - **Serving endpoint uses `services[].ref`** at the Knative auto-Service, not `create`. Creating a parallel `corev1.Service` would be a wasteful object with no functional role.
   - Whenever `auth` is set, **the composer also injects an OAuth2-proxy sidecar** into the Knative pod template's `containers[]` (in the same composer) — port-bearing in front of the user container. The networking endpoint itself doesn't carry sidecar config.
   - Translate the InfW auth-vocabulary enum to the OAuth2-proxy args + the per-endpoint policy the composer renders (see §4).
   - Set `networking.knative.dev/visibility: cluster-local` and omit `externalAccess` when `isPrivateServiceUrl=true`.
   - **NIM metrics endpoint** — when the `runai.nim/workload=true` label is present, emit a second endpoint for the metrics port with `services[].create` and no `externalAccess`. **No sidecar for the metrics endpoint** (scrape-only).

## 9. Design decisions (locked)

1. **Translation location:** CP-side composer (v2) is the primary translator. Anything the CP can't compute statically falls to an **AnyWorkload plugin** on the cluster side. Concretely: enum mapping, annotation generation, vocab translation, NIM-label propagation → composer; live-cluster-state work (per-user secret synthesis, runtime decisions) → plugin.
2. **`initializationTimeoutSeconds`:** legacy InfW already maps this to the Knative annotation `serving.knative.dev/progress-deadline: '<N>s'` (`utils.go:96-97`). Composer mirrors — no investigation needed.

## 10. Out of scope

- **DistInfW** — see the sibling doc [`dist-infw-to-workloadrunner.md`](./dist-infw-to-workloadrunner.md).
- **Status surface** — revision names, external URLs, phase mapping, `PluginStatus` propagation.
- **AMD-GPU `NVIDIA_VISIBLE_DEVICES` quirks** — flagged as open in the audit; no conversion for non-NVIDIA vendors.

## 11. Fields not in use

Inference API fields the composer **accepts and ignores** (mirror legacy), or that are deprecated / internal-only / not surfaced on the WLR. Listing them so the rest of this design stays focused on fields we actually translate.

| Category | API field | Source | Reason it's not in use |
|---|---|---|---|
| dead | `compute.extendedResources[]` | `model_workspace_spec_spec_all_of_compute.go:32` | Legacy InfW silently drops this — no users have surfaced a need. Composer mirrors legacy (accept, ignore). Could in principle map to `resources.limits[<resource>]` on the new path — gated on a real use case. |
| deprecated | `autoscaling.initializationTimeoutSeconds` | `model_auto_scaling.go:41` | Deprecated in the API model itself (Go doc-string says *DEPRECATED*). Use `servingConfiguration.initializationTimeoutSeconds` instead — legacy InfW already maps it via the Knative `progress-deadline` annotation. |
| internal | `compute.gpuRequestType` | `model_workspace_spec_spec_all_of_compute.go:43` | API-side validation enum (`memory` vs `portion`) — tells the API/composer which of `gpuPortionRequest` / `gpuMemoryRequest` is authoritative. Doesn't appear on the WLR. |
| internal | `security.uidGidSource` | `model_inference_spec_spec_all_of_security.go:40` | API-side enum (`fromIdpToken` / `fromTheImage` / `custom`) — drives whether the literal `runAsUid` / `runAsGid` values are used, or resolved from the user's identity provider. Resolved by API/composer; the final WLR carries concrete `runAsUser` / `runAsGroup` values. |
| internal | *(no s3 field on the API)* | — | Legacy InfW CRD has `storage.s3` (composed as an init-container today). The Inference API does **not** expose s3 today — users mount S3 via PVC or external tooling. If/when the API adds it, the CP composer handles the init-container composition. |

**Composer behaviour:** mirror legacy — accept the fields, ignore the values, do not translate to the WLR. If we later decide to honour any of these, promote the row back into the main field-mapping table.

> Legacy InfW CRD also exposes fields the API does not (`usage`, `baseWorkload`, `username`/`userId`, `name`/`namePrefix`, `migProfile`, `nodePool`, `ingressUrl`, `restartPolicy`, `autoDeletionTimeAfterCompletionSeconds`, `terminationGracePeriodSeconds`, `preemptionLimit`, `mps` as a top-level value). These can't reach the composer because no API surface produces them — they're outside the scope of this migration.
