# application-controller

The Application controller manages `Application` CRs in two modes:

1. **Provider mode** — is created from helm installation initiated in the cluster
   (helm install, etc). for those applications the controller task is mainly to track deployed resources 
   (ExternalWorkloads, PVCs, Services, etc.) and reflects their state into the Application status.
2. **Helm install mode** — when `run.ai/application` annotation is set to true, the controller owns the 
   full lifecycle: install, upgrade, and uninstall of the Helm chart. For this scenario, spec.helm should
   be provided, with the chart URL/ref and values.

---

## Architecture

```text
  Kubernetes API Server
  ┌─────────────────────────────────────────────────────┐
  │  Application CR                                     │
  │  ┌──────────────────┐  ┌───────────────────────┐    │
  │  │  spec.helm = nil │  │  spec.helm != nil     │    │
  │  │  (Provider mode) │  │  (Helm install mode)  │    │
  │  └────────┬─────────┘  └──────────┬────────────┘    │
  └───────────┼────────────────────────┼────────────────┘
              │                        │
              v                        v
  ┌───────────────────────────────────────────────────┐
  │              ApplicationController                │
  │                   Reconcile()                     │
  └───────────────┬───────────────────────────────────┘
                  │
        ┌─────────┴──────────┐
        │                    │
        v                    v
  ┌───────────┐    ┌─────────────────────────────────┐
  │ Provider  │    │    reconcileHelmDeployment()    │
  │  Mode     │    │                                 │
  │           │    │  1. Helm op in progress? → wait │
  │  Watches: │    │  2. Deleting?  → uninstall task │
  │  - EW     │    │  3. No release → install task   │
  │  - PVC    │    │  4. Spec changed → upgrade task │
  │  - CM     │    │  5. Matches desired → no-op     │
  │  - Svc    │    └──────────────┬──────────────────┘
  │  - Secret │                   │
  └───────────┘                   v
                       ┌──────────────────-───┐
                       │     HelmWorkers      │
                       │  5 worker goroutines │
                       │  job buffer: 100     │
                       └──────────┬────────-──┘
                                  │
                     ┌────────────┴────────-─────┐
                     │       Task.Execute()      │
                     │                           │
                     │  reportProgress()         │
                     │  ┌─────────────────────┐  │
                     │  │ queued              │  │
                     │  │      ↓              │  │
                     │  │ started             │  │
                     │  │      ↓              │  │
                     │  │ downloading         │  │
                     │  │      ↓              │  │
                     │  │ running             │  │
                     │  │      ↓              │  │
                     │  │ completed / failed  │  │
                     │  └─────────────────────┘  │
                     └────────────┬──────────────┘
                                  │
                                  v
                       ┌────────────────────-─┐
                       │   Status().Patch()   │
                       │  (Application CR)    │
                       └─────────────────-────┘
```text

---

## Helm Install Mode

Activated when `spec.helm` is set on the Application CR.

### Chart Sources

Two mutually exclusive ways to specify the chart:

| Field | Description |
|---|---|
| `spec.helm.chartRef` | Repository + chart name + version (HTTP or OCI) |
| `spec.helm.chartUrl` | Direct URL to a packaged chart (`.tgz`) |

### Authentication

Set `spec.helm.auth.secretRef` to reference a Secret in the same namespace:

| Secret type | Keys used | Auth method |
|---|---|---|
| `kubernetes.io/dockerconfigjson` | `.dockerconfigjson` | OCI registry |
| `kubernetes.io/tls` | `tls.crt`, `tls.key`, `ca.crt` | TLS client cert |
| `Opaque` | `username` + `password` | Basic auth |
| `Opaque` | `token` | Bearer token |
| `Opaque` | `ca.crt` | Custom CA |

### Lifecycle

1. **Finalizer** — `run.ai/helm-application` is added on first reconcile to ensure uninstall runs before the CR is deleted.
2. **Install** — triggered when no Helm release exists.
3. **Upgrade** — triggered when the deployed chart URL or values differ from the spec.
4. **Uninstall** — triggered when the CR deletion timestamp is set. The finalizer is removed after uninstall succeeds.
5. **No-op** — when the deployed release already matches the spec.

### Failure Retry Suppression

After a failed install or upgrade the controller backs off for `5 minutes` before retrying. A retry is also triggered immediately when `spec` changes (i.e. `metadata.generation` increases). The last failed generation is tracked in the `run.ai/helm-last-failed-generation` annotation.

### Restart Resilience

On controller restart, in-memory task tracking is lost. The controller detects in-flight Helm operations by checking the release status (`pending-install`, `pending-upgrade`, `pending-rollback`, `uninstalling`) and waits for them to settle before acting.

---

## Status

The Application status is visible via `kubectl get application`:

```text
NAME         ACTION    PHASE       STARTED   COMPLETED   AGE
my-app       install   completed   10m       9m          12m
```text

Use `-o wide` for the full message field.

### Status Fields

**`status.action`** — the current or last Helm operation:

| Field | Description |
|---|---|
| `type` | `install`, `upgrade`, or `uninstall` |
| `phase` | `queued` → `started` → `downloading` → `running` → `completed` / `failed` |
| `startedAt` | When the operation started |
| `completedAt` | When the operation finished |
| `message` | Human-readable status message |
| `error` | Error detail on failure |

**`status.helmStatus`** — state reported directly from Helm:

| Field | Description |
|---|---|
| `helmPhase` | Helm release status (e.g. `deployed`, `failed`) |
| `chartUrl` | The chart URL that was last deployed |
| `releaseVersion` | Helm release revision number |
| `helmError` | Error from Helm |
