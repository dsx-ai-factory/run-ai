# composer_v2 — agent guide

Scope: everything under `composer/composer_v2/` and its shared helpers in
`composer/` (`FieldsComposer`, `StepRunner`, `setters_impl.go`,
`setters_impl_k8s.go`, `init_containers/`, ...).

**`composer_v1/` is legacy — do not modify it, and do not copy patterns
from it.** The old design doc at `composer_v1/runai_workloads/README.md`
predates the v2 layout and is obsolete; this file supersedes it.

## What the composer does

Converts control-plane workload API objects into the concrete Kubernetes CRDs
(RunaiJob, Kubeflow TrainingJob, Knative Service, LeaderWorkerSet,
NIMService, ...) that the cluster consumes. Two responsibilities:

1. **Validate** — reject invalid workloads in the control plane rather than
   letting them fail at admission or runtime in the cluster.
2. **Project** — copy each source field to the right place in the target CRD,
   in the canonical form the target expects.

## The building blocks

### Compose function

`ComposeXxx(src, target, setter, cookie) func() *workloadsApi.Error`

A method — usually on `*FieldsComposer`, occasionally on a concrete
composer. Owns **one source field**. Responsibilities:

- Validate the source (range, format, allowed enum, mutual exclusion, ...).
- Normalize the source into the **canonical form the setter expects** (e.g.
  join a slice into a comma-separated string, project a nested object into a
  small info struct).
- Hand the normalized value to the setter, along with `target` and `cookie`.

Compose is **generic across workload types**. It must not know Knative /
RunaiJob / LWS field paths. Anything CRD-specific belongs in the setter.

### Setter function

`func XxxToYyy(src ProcessedType, target any, cookie any) *workloadsApi.Error`

An **independent, package-level function** (not a method on any composer).
Owns writing the processed value into one specific target shape. A single
Compose function can be paired with many setters:

- `StringToString` — write into a plain `*string`.
- `StringToStringField` — write into a `field_groups.StringField`.
- `MapToMapDict` — write as an entry in a labels/annotations map (cookie
  supplies the map key).
- `StringToEnvVar` — write as a `corev1.EnvVar` on a container.
- `UrlsToK8sInitContainers` — install an oauth2-proxy chain into
  `podSpec.InitContainers` + `podSpec.Volumes`.

The setter **must type-assert** its `target` and `cookie` on entry and
return a 500 on mismatch — both are `any`, so a mistyped caller would
otherwise panic:

```go
func servingPortToLWS(src *composer.ServingPortInfo, target any, rawCookie any) *workloadsApi.Error {
    lwsTarget, ok := target.(*LeaderWorkerSetTarget)
    if !ok {
        return workloadsApi.NewError(http.StatusInternalServerError,
            fmt.Sprintf("servingPortToLWS: target type %T is not *LeaderWorkerSetTarget", target))
    }
    cookie, ok := rawCookie.(*lwsNetworkingCookie)
    if !ok {
        return workloadsApi.NewError(http.StatusInternalServerError,
            fmt.Sprintf("servingPortToLWS: cookie type %T is not *lwsNetworkingCookie", rawCookie))
    }
    // ... write into lwsTarget using cookie.serviceName / cookie.lwsName / ...
}
```

### processFields function

`func (c *ConcreteComposer) processXxxFields(...) func() *workloadsApi.Error`

A method on a concrete workload composer (`InferenceToKnativeComposer`,
`DistributedInferenceToLWSComposer`, ...) that groups the Compose calls for
one subsection of the workload spec — `processComputeFields`,
`processStorageFields`, `processNetworkingFields`, `processSecurityFields`,
`processPlacementFields`, `processReplicaFields`, `processDistributedFields`,
`processModelStoreFields`, `processServingConfigFields`, ...

Each returns a batch step that:

- Opens a `runner := &composer.StepRunner{}`.
- Calls `runner.Batch(step1, step2, ..., stepN)` with the Compose functions.
- Returns `runner.Error()`.

`processXxxFields` steps may themselves be small methods on the concrete
composer when the local orchestration is non-trivial — see
`InferenceToKnativeComposer.buildAuthProxySidecar` and
`DistributedInferenceToLWSComposer.buildAuthProxySidecar` for the shape.

### The cookie

Free-form `any` argument threaded from Compose to setter. Purpose: hand the
setter **the small bit of routing / naming info it needs to do its job
that is not the source data itself**. Examples:

- `"run.ai/image"` — the map key for `MapToMapDict`.
- `*lwsNetworkingCookie{config, projectName, workloadName, lwsName,
  serviceName}` — naming context for the LWS serving-port setter.
- `nil` — most calls, when the setter needs nothing extra.

**Cookie rules — non-negotiable:**

- **Never carry target fields in the cookie.** If the setter needs to read
  or mutate a target field, receive it via `target`, not via `cookie`.
  Conflating the two turns the cookie into a shadow target and blocks reuse.
- **Keep it minimal.** Only what this specific setter needs. If two setters
  need different subsets, don't merge them into one big cookie.
- Cookies MAY be used as a **shared out-parameter** between two Batch steps
  (`cookie.servingURL` filled by `servingPortToLWS`, read later by the
  sidecar builder) — but even then the value must be setter-relevant
  routing/naming info, not a slice of the target.

### StepRunner

`StepRunner.Batch(steps...)` runs each step sequentially and short-circuits
on the first error. Every Compose / process step therefore returns
`func() *workloadsApi.Error` — a *factory* that produces the closure Batch
will call:

```go
func (c *ConcreteComposer) ComposeSomething(...) func() *workloadsApi.Error {
    return func() *workloadsApi.Error {
        // validate, normalize, call setter
    }
}
```

`runner.Run(step)` is the eager single-step alias.

If step N returns a value step N+1 needs, capture a **shared struct** with
both closures:

```go
authInfo := composer.PortAuthInfo{}
runner.Batch(
    dinfComposer.composeServingPort(&projected, target, &cookie, &authInfo),
    dinfComposer.buildAuthProxySidecar(&projected, target, &cookie, &authInfo),
)
return runner.Error()
```

Both closures capture `&authInfo`; the first fills it, the second reads it,
and Batch's sequential ordering guarantees the first has finished before
the second runs. Don't route this through the cookie.

## Common practices — must be kept

### One source → one target — split for fan-out

A Compose function projects **one source field** onto **one target**. When
the same source needs to land in two places, use **two Compose calls** with
the same source and different target/setter pairs, over one Compose that
writes into two places.

When the second call needs derived data from the first, use the "cooked
struct" pattern — the first Compose fills a small info struct as an
out-parameter, the second Compose consumes it:

```go
var gpuSizingInfo GpuSizingInfo
ComposeGpuSizing(src.GpuRequestType, src.GpuDevicesRequest, ...,
    &target.ComputeSection, gpuSizingSetter, nil, &gpuSizingInfo)
ComposeGpuMemoryLimit(gpuSizingInfo, &target.Annotation,
    composer_v2.MapToMapDict, "run.ai/memory-limit")
```

If both Composes run inside a `runner.Batch(...)`, capture the info struct
in a shared variable as in the StepRunner example above.

### Write for reuse — prefer FieldsComposer over concrete methods

The same source field usually appears in more than one workload type. The
canonical home for a Compose function is **`FieldsComposer`** (top-level
`composer/` package), so every concrete composer can reuse it via
`itc.fc.ComposeXxx(...)`.

- Concrete composers should define as few local Compose functions as
  possible. When you add one, ask: could this be a `FieldsComposer` method
  every workload type could share?
- If a specific composer needs to tweak a shared `FieldsComposer` call
  (add validation, choose a different setter), prefer a **thin wrapper
  method** on the concrete composer that delegates to `fc.ComposeXxx`, over
  copy-pasting the compose body.

### Composers own Compose + processFields; setters stand alone

- `Compose*` and `process*Fields` are **member functions** — either of
  `*FieldsComposer` (shared) or of a concrete workload composer (specific).
- Setters (`XxxToYyy`) are **package-level functions with no receiver**.
  They are stateless and reusable across composers.

### Where the shared setters live

- `composer/setters_impl.go` — generic setters (`StringToString`,
  `BoolToMapString`, `MapToMapDict`, `StringToEnvVar`, ...).
- `composer/setters_impl_k8s.go` — setters that target `*corev1.PodSpec`
  and other core Kubernetes types.
- `composer/composer_v2/setters_init_containers.go` — setters that install
  init containers on a `*corev1.PodSpec` (oauth2 proxies, username loader,
  volumes).
- `composer/composer_v2/setters_impl_runner.go` — setters that target the
  workload runner endpoints/config.
- Workload-specific setters in `setters.go` inside each sub-composer —
  only when the shape truly can't be generalized (e.g. `servingPortToLWS`).

### Aggregator targets

When a setter produces multiple related artifacts (init container + volume
+ volume mount), the natural target is the Kubernetes type that already
owns them all — usually `*corev1.PodSpec`. Don't invent multi-bucket
buffer types just to carry artifacts back to the caller.

### Cluster-side knobs via a config interface in the cookie

Setters that produce k8s artifacts often need values the cluster side reads
from env vars (image tags, custom-CA toggles, ...). Pass a
`composer.ClusterStorageConfig` via the cookie; swapping the source
(cluster defaults → per-cluster overrides) becomes a one-line call-site
change.

## Directory layout

```
composer/
  fields_composer*.go          shared Compose methods on *FieldsComposer
  step_runner.go               StepRunner.Batch idiom
  setters_impl.go              generic setters (any target)
  setters_impl_k8s.go          k8s / PodSpec setters
  cluster_storage_config.go    ClusterStorageConfig interface
  validations.go               shared validators
  init_containers/             oauth2-proxy, username-loader, volumes builders
                               and the shared BuildOauth2ProxySidecar helper
  composer_v2/
    setters_init_containers.go     init-container installer setters
    setters_impl_runner.go         runner-target setters
    cluster_based_options.go       ClusterBasedOptions plumbing
    iw_tw_workloads/               Interactive / Training  → RunaiJob (reference)
    inference_workloads/           Inference               → Knative Service
    distributed_to_kubeflow/       Distributed training    → Kubeflow TrainingJob
    dinf_to_lws/                   Distributed inference   → LeaderWorkerSet
    nim_workloads/                 NIM inference           → NIMService
  composer_v1/                   LEGACY — do not modify
```

## Composers at a glance

| Concrete composer                     | Package                                     | Target CRD           | Notes |
|---------------------------------------|---------------------------------------------|----------------------|-------|
| `InterTrainToJobComposer`             | `composer_v2/iw_tw_workloads`               | RunaiJob             | **Reference implementation.** Interactive workspaces + single-pod training. |
| `InferenceToKnativeComposer`          | `composer_v2/inference_workloads`           | Knative Service      | Knative admits one port-declaring `spec.container`; the oauth2 proxy replaces the user container as the port owner when authentication is enabled. |
| `DistributedToKubeflowComposer`       | `composer_v2/distributed_to_kubeflow`       | Kubeflow TrainingJob | Multi-replica training (PyTorch / TF / MPI / XGBoost / ...). |
| `DistributedInferenceToLWSComposer`   | `composer_v2/dinf_to_lws`                   | LeaderWorkerSet      | LWS-based distributed inference; oauth2 proxy is a peer sidecar next to the leader container. |
| `NimServiceComposer`                  | `composer_v2/nim_workloads`                 | NIMService (CRD)     | NVIDIA NIM inference. |

All five share `*FieldsComposer` for cross-workload Compose logic and
`StepRunner.Batch` for orchestration.

## Lessons learned (subtle bugs)

Real footguns hit while working in this code. Read before making changes
to similar patterns.

### Pointer-into-slice invalidation across `append`

If a caller captures `container := &podSpec.Containers[0]` and later a
helper calls `podSpec.Containers = append(podSpec.Containers, ...)`, the
append **may reallocate the backing array** and strand the caller's pointer
on the old array. Silent bug: writes through the captured pointer land in
unused memory; the goldens then show the fields the caller thought it had
cleared still populated.

Two fixes work; the second is the one used in `BuildOauth2ProxySidecar`:

- Refetch inside the callback (`&podSpec.Containers[i]`) after all appends.
- Run any callback that mutates caller-held pointers **before** the append.
  See `ServingProxySpec.CustomizeServingProxy` — its doc comment explains
  why it must run pre-install.

### Typed cookies at the setter boundary

A cookie declared as `any` and then re-wrapped along the way easily becomes
`*interface{}` (e.g. passing `&someInterfaceVar` to a
`*Concrete`-expecting setter). Two rules:

- Setters must declare their expected cookie type with a
  `_, ok := rawCookie.(*Foo); ok` check and return a 500 on mismatch — not
  panic.
- Helper closures that thread the cookie should type their parameter as
  `*Concrete`, not `any` — see `composeServingPort`'s explicit
  `*lwsNetworkingCookie` parameter.

### FieldsComposer field propagation

When you add a field to `FieldsComposer` (e.g. `Namespace`), **every**
concrete composer's constructor must populate it. Missing one produces a
silent, downstream bug — we shipped a malformed oauth2 cookie prefix
(`InferenceWorkload--my-llm-0` instead of
`InferenceWorkload-runai-team-a-my-llm-0`) because
`NewInferenceToKnativeComposer` forgot to set `fc.Namespace`.

Prefer `itc.fc.ProjectName` / `itc.fc.Namespace` over per-composer copies
of the same fields.

### Cookie stays minimal — never target fields

Cookies must only carry setter-scoped naming/routing info. If the setter
needs to read or mutate a target field, it should go through `target`, not
through the cookie. When we conflated the two — an
`oauth2ProxySidecarContext` carrying `namespace`, `projectName`,
`workloadName`, `config` — unifying the inference and dinf sidecar builders
required a full restructuring, because the "context" was really a
mini-target masquerading as a cookie.

### Sharing state across Batch steps

`StepRunner.Batch` steps run sequentially, so a shared out-parameter
captured by both closures is safe and readable. Don't reach for a cookie
for this — a plain struct captured by both closures is clearer than
routing state through an `any` cookie.

## Existing rules (still in force)

1. **Use existing `fc.*` helpers.** Don't hand-roll field copies. See
   `fields_composer.go` for the catalog: `CopyString`, `CopyInt`,
   `CopyBoolean`, `CopyArray`, `CopyStringToMap`, `ComposeImage`,
   `ComposeReadinessProbe`, `ComposeAutoScaling`, ...
2. **Use existing setters.** See `setters_impl.go` and
   `setters_impl_k8s.go` for `StringToString`, `BoolToMapString`,
   `MapToMapDict`, `StringToEnvVar`, ...
3. **Workload-specific logic lives in setters, not in Compose.** If you
   find yourself writing Knative field paths inside a Compose function,
   stop — move it to a setter.
4. **Follow `iw_tw_workloads/` as the reference implementation.**
5. **Wrap Compose calls in `runner.Batch(...)`.** Don't chain
   `if err != nil` checks manually.
