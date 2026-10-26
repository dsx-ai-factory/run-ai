# runai-cluster (runai-chart)

`runai-chart` is the umbrella **Helm chart** that installs the Run:AI cluster
side onto a Kubernetes cluster. It is not a service — it packages and wires
together every in-cluster component: the Run:AI operator (`go-operator`) that
owns the `RunaiConfig`, the CRD-upgrader and helm-hooks-handler jobs that
prepare the cluster on install/upgrade, cluster-scoped RBAC and priority
classes, OpenShift-specific resources, the `RunaiConfig`/`Cluster` custom
resources, and the scheduler stack pulled in as chart dependencies
(`kai-resource-management`, `karta`, and the CRDs subchart; `kai-scheduler`
arrives as a sub-dependency of KRM). The chart is named
`runai-cluster` (see [`Chart.yaml`](Chart.yaml)) and is owned by the
**Griffin Team**.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Chart dependencies](#chart-dependencies)
- [Building and packaging](#building-and-packaging)
- [Testing and validation](#testing-and-validation)
- [Configuration (values)](#configuration-values)
- [Lifecycle hooks](#lifecycle-hooks)

## Overview

The chart renders and manages:

- **Run:AI operator** (`runai-operator`, image `go-operator`) — the controller
  that reconciles `RunaiConfig` and brings up the cluster operands. Its
  Deployment, RBAC, network policy, and ServiceMonitor are in
  [`templates/`](templates).
- **Install/upgrade jobs** — `crdUpgrader` (runs `runai-crd-upgrader` to apply
  CRDs and wait for `Established`), `helmHooksHandler` (runs the lifecycle
  hooks), and `kartaMigrate` (migrates legacy `ResourceInterface` CRs into Karta
  CRs). These are wired as Helm hook `Job`s under
  [`templates/hooks/`](templates/hooks).
- **Core custom resources** — `RunaiConfig` (`templates/runaiconfig.yaml`), the
  installation `Cluster` (`templates/cluster.yaml`), and the cluster CRD
  ([`crds/run.ai_clusters.yaml`](crds); the full CRD set is packaged via the
  `crds` subchart).
- **Cluster-scoped RBAC** — one ClusterRole per controller under
  [`templates/clusterroles/`](templates/clusterroles) (agent, cluster-sync,
  project-controller, advice-restart-controller, …), plus user-auth RBAC.
- **Scheduling primitives** — Run:AI priority classes
  ([`templates/priorityclasses/`](templates/priorityclasses)) and KAI
  scale-adjust resources (`templates/kai-scale-adjust-resources.yaml`).
- **OpenShift support** — SecurityContextConstraints and admission-webhook
  resources ([`templates/ocp/`](templates/ocp)), gated by `openShift.mode`.
- **Observability** — ServiceMonitor, Grafana secret, training-operator image-tag
  configmap.

## Repository layout

```text
.
├── Chart.yaml               # Chart metadata + dependencies (kai-resource-management, karta, crds)
├── values.yaml              # Default customer-facing values
├── values.schema.json       # JSON Schema validating values.yaml
├── internal_values.yaml     # Internal-only values doc (NOT referenced by the chart)
├── crds/                    # run.ai_clusters CRD (installed ahead of templates)
├── charts/                  # Packaged dependency charts (.tgz) + crds subchart
│   ├── crds/                # Subchart bundling the full Run:AI CRD set
│   ├── kai-resource-management-*.tgz   # bundles kai-scheduler as its own dependency
│   ├── karta-*.tgz
│   └── runai-engine-*.tgz   # (legacy engine chart tarball)
├── templates/
│   ├── deployment.yaml          # runai-operator Deployment
│   ├── rbac.yaml, clusterroles/ # RBAC (per-controller ClusterRoles + user auth)
│   ├── runaiconfig.yaml         # RunaiConfig CR
│   ├── cluster.yaml             # installation Cluster CR
│   ├── priorityclasses/         # Run:AI PriorityClasses
│   ├── hooks/                   # Helm hook Jobs + RBAC (crd-upgrade, install/upgrade/delete, migrations)
│   ├── ocp/                     # OpenShift SCC + admission webhooks
│   ├── servicemonitor.yaml, networkpolicy.yaml, grafana-secret.yaml
│   ├── kai-scale-adjust-resources.yaml
│   ├── training-operator-tags-config-map.yaml
│   ├── _helpers.tpl, NOTES.txt
│   └── tests/                   # Helm test hooks (validate-crds)
├── tests/                   # helm-unittest suites (+ __snapshot__)
├── script/uploadChart.sh    # Chart publishing helper
├── Makefile
└── .helmignore
```

## Chart dependencies

Declared in [`Chart.yaml`](Chart.yaml) and vendored as tarballs under `charts/`:

| Dependency | Version | Condition | Source |
|------------|---------|-----------|--------|
| `kai-resource-management` | `~0.18.0` | `kai-resource-management.enabled` | `oci://ghcr.io/kai-scheduler/kai-resource-management` |
| `karta` | `0.2.6` | `karta.enabled` | `oci://ghcr.io/dsx-ai-factory/workload-map/karta` |
| `crds` | `0.0.0` | `development.includeCRDsInReleaseManifest` | `file://charts/crds` |

`helm dep up` (run by `make build`) resolves these into `charts/`.

`kai-resource-management` is a **range**, not a pin: each build takes the newest
KRM patch of that minor, resolved against the v-less OCI tag KRM publishes
alongside its `v`-prefixed one (Helm will not match a range against `v0.18.6`).
Because the resolved version is not recorded anywhere in this repository,
each GitHub Release carries a `subcharts.yaml` asset naming the versions that
build actually packaged. See
[KAI Scheduler dependency](../../kai-scheduler/README.md).

## Building and packaging

This is a Helm chart — there is **no Go binary**. The [`Makefile`](Makefile)
includes `../../../build/makefile/runai-cluster.mk` and `index.mk`, but overrides
the standard targets with chart-packaging recipes. `VERSION` defaults to
`0.0.0-dev`; `REGISTRY_URL`/`REGISTRY` default to `nvcr.io` / `nvidia/runai`.

| Target | Description |
|--------|-------------|
| `make build` | Stamps `VERSION` into `Chart.yaml` and the image tags/registries in `values.yaml` (via `yq`: `runai-operator`, `crdUpgrader`, `helmHooksHandler`, `kartaMigrate`), runs `helm dep up`, renders `helm template … > manifests.yaml`, then `helm package … → runai-cluster.tgz`. |
| `make test` | `make build`, then runs the helm-unittest suites in Docker (`helmunittest/helm-unittest:3.17.2-0.8.1` over `tests/**/*_test.yaml`). |
| `make lint` | `make build`, then `helm lint`. |
| `make validate` | `make build`, then renders with OpenShift enabled and runs OPA/`conftest` (`openpolicyagent/conftest:v0.59.0`) against `common/opa-policies/kubernetes`. |
| `make extra` | Stamps `VERSION` into `Chart.yaml` and runs `helm dep update` (CI "make extra" step; `make_extra_cmd: true` in `artifact-metadata.json`). |
| `make build-airgapped` | Placeholder (todo). |

The packaged artifact is `runai-cluster.tgz` (`path_to_artifacts` in
`artifact-metadata.json`).

### Local deploy (local-dev.sh)

`local-dev.sh` does **not** run `make build` for the chart. Its `deploy_runai_cluster`
step deploys the chart directory in place:

- It rewrites `appVersion` in [`Chart.yaml`](Chart.yaml) to the RC image tag
  **`0.0.0-dev`** (`RC_IMAGE_TAG`), then runs `helm upgrade --install
  runai-cluster runai-cluster/cluster/runai-chart` into the RC namespace.
- Image registries are overridden to the local Kind registry
  (`localhost:<port>/runai-local`) and all component tags to `0.0.0-dev`
  (`runai-operator`, `crdUpgrader`, `helmHooksHandler`, `kartaMigrate`), matching
  the images built by `build_go_service` (`runai-local/<name>:0.0.0-dev`).
- It passes `--set development.includeCRDsInReleaseManifest=true` (renders the
  `crds` subchart), sets the control-plane URL/credentials and `cluster.uid`, and
  layers `.local-dev/values-rc.yaml`.

To rebuild the RC images and re-deploy the chart locally:

```bash
./local-dev.sh --skip-build-cp --skip-cluster
```

## Testing and validation

- **Unit tests:** `make test` runs [helm-unittest](https://github.com/helm-unittest/helm-unittest)
  suites in `tests/` (with golden snapshots in `tests/__snapshot__/`) covering
  RBAC, KAI operator/ownership/namespaces, priority classes, global service
  account, OpenShift, and hooks.
- **Lint:** `make lint` → `helm lint`.
- **Policy validation:** `make validate` renders the chart (OpenShift enabled)
  and runs OPA `conftest` against the shared `common/opa-policies`.
- **In-cluster tests:** `templates/tests/validate-crds` provides a `helm test`
  hook that validates CRDs against a deployed release.

## Configuration (values)

Full defaults are in [`values.yaml`](values.yaml) and validated by
[`values.schema.json`](values.schema.json). Top-level sections:

| Section | Purpose |
|---------|---------|
| `global` | Cross-cutting settings: `fipsMode`, `customCA`, `additionalImagePullSecrets`, `image.registry`, `affinity`/`tolerations`, `metrics.disableInfrastructureMetricsRemoteWrite`, internal `serviceAccountName`. |
| `clusterConfig` | RFC 7396 JSON Merge Patch applied to `RunaiConfig` (override specific fields; `null` deletes). |
| `controlPlane` | `url`, `clientID` (default `runai-internal.installer`), `clientSecret` / `existingSecret` for CP auth. |
| `cluster` | `uid`, `rbac.escalate`. |
| `runai-operator` | Operator image/resources/replicas, `networkPolicy`, `metrics.serviceMonitor`, `config`. |
| `crdUpgrader` | `runai-crd-upgrader` job: `enabled`, `establishedTimeout`, image, resources. |
| `helmHooksHandler` | Hook job image, resources. |
| `kartaMigrate` | Karta migration job: `enabled`, `removeRid`, `activeDeadlineSeconds`, `backoffLimit`, image, resources. |
| `kai-resource-management` | Subchart values for KRM's three controllers and its KRMConfig mode, plus a nested `kai-scheduler` block (resource reservation namespace/SA, autoscaling, queues, shards, priority classes). |
| `karta` | Subchart values (operator, pod security context). |
| `operands` | Enable/RBAC for training operator, researcher service, cluster API, prometheus. |
| `userAuth` | RBAC for Run:AI user `kubectl` access. |
| `openShift` | `mode` (AutoDetect/Enabled/Disabled) + SecurityContextConstraints. |
| `engineToKaiMigration` / `resourceManagementMigration` | Gate one-off migration hooks (self-gating on legacy CRDs). |
| `trainingOperatorImageTags` | Map of training-operator image digests → versions. |
| `development` | `headless`, `includeCRDsInReleaseManifest`, `skipDependeciesChecks`, `enableProfiling`. |

`internal_values.yaml` documents internal-only values (e.g. `global.nmc.*`,
`runai-operator.disableRunaiConfigFinalizer`); it is **not** referenced by the
chart and its defaults have no effect.

## Lifecycle hooks

The install/upgrade/delete flow runs `helm-hooks-handler` and
`runai-crd-upgrader` as Helm hook `Job`s (templates under
[`templates/hooks/`](templates/hooks)):

- **CRD upgrade** (`crdUpgrader`) — applies and establishes CRDs.
- **pre-install / post-install** — CP handshake, base resources, CA, pull secrets.
- **pre-upgrade / pre-crd-upgrade / pre-resource-management-upgrade** — upgrade
  prep and migrations (engine→KAI SchedulingShard teardown, legacy department →
  KAI department migration).
- **post-upgrade** — queue-version validation and migration cleanup.
- **pre-delete** — cleanup before release removal.
- **kartaMigrate pre/post-upgrade** — copy `ResourceInterface` CRs → Karta CRs
  and (post-upgrade) remove the legacy RIs.

See [`helm-hooks-handler`](../helm-hooks-handler) and
[`runai-crd-upgrader`](../runai-crd-upgrader) for the job binaries these hooks run.
