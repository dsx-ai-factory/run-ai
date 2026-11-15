# runai-cluster (runai-chart) — Build Context

**Type:** helm (Helm chart — no Go binary)
**Team:** Griffin Team (`artifact-metadata.json` → `jira.team_name`; Slack `@griffin`)
**Chart name:** `runai-cluster` (see `Chart.yaml`)

## What it is

Umbrella Helm chart for the Run:AI cluster side. Packages the `runai-operator`
(`go-operator`) Deployment/RBAC, the `crdUpgrader` + `helmHooksHandler` +
`kartaMigrate` lifecycle Jobs, cluster-scoped ClusterRoles, priority classes,
`RunaiConfig`/`Cluster` CRs, OpenShift SCC/webhooks, ServiceMonitor, and the
scheduler stack as dependencies (`kai-resource-management ~0.18.0`, `karta 0.2.6`,
`crds` subchart; `kai-scheduler` arrives as a sub-dependency of KRM). Produces
`runai-cluster.tgz`.

## Make targets

Makefile overrides the shared targets with chart recipes (`VERSION?=0.0.0-dev`,
`REGISTRY_URL/REGISTRY = nvcr.io/nvidia/runai`). Includes
`../../../build/makefile/runai-cluster.mk`, `index.mk`; `DOCKER_WORK_DIR=/local/runai-cluster`.

| Target | What it does |
|--------|-------------|
| build  | `yq` stamps `VERSION` into `Chart.yaml` + image tag/registry in `values.yaml` (runai-operator, crdUpgrader, helmHooksHandler, kartaMigrate), `helm dep up`, `helm template > manifests.yaml`, `helm package → runai-cluster.tgz` |
| test   | `build` then helm-unittest in Docker (`helmunittest/helm-unittest:3.17.2-0.8.1`, `tests/**/*_test.yaml`) |
| lint   | `build` then `helm lint` |
| validate | `build` then render (OpenShift enabled) + OPA `conftest` (`openpolicyagent/conftest:v0.59.0`) over `common/opa-policies/kubernetes` |
| extra  | stamp `VERSION` into `Chart.yaml` + `helm dep update` (CI make-extra step) |
| build-airgapped | placeholder (todo) |

## Key build notes

- **No Go build / no `build-go`.** This is `artifact_type: helm`. Artifact is
  `runai-cluster.tgz` (`path_to_artifacts`); `force_build: true`,
  `needs_to_wait: true`, `make_extra_cmd: true`, `cache_artifacts: true`.
- **Version/tag: `0.0.0-dev`.** `Chart.yaml` `version` is `0.0.0-dev`; `appVersion`
  is the placeholder `CHART_APP_VERSION` in-repo. `make build` sets `version`;
  component image tags in `values.yaml` are `0.0.0-dev`.
- **local-dev.sh does NOT `make build` the chart.** `deploy_runai_cluster`:
  - rewrites `Chart.yaml` `appVersion` to `RC_IMAGE_TAG` (`0.0.0-dev`),
  - `helm upgrade --install runai-cluster <chart dir>` into the RC namespace,
  - overrides registries to `localhost:<port>/runai-local` and all component tags
    to `0.0.0-dev` (runai-operator, crdUpgrader, helmHooksHandler, kartaMigrate),
  - `--set development.includeCRDsInReleaseManifest=true`, sets `controlPlane.*`
    and `cluster.uid`, layers `.local-dev/values-rc.yaml`.
- **Dependencies** (`Chart.yaml`, vendored tarballs in `charts/`):
  `kai-resource-management` (`oci://ghcr.io/kai-scheduler/kai-resource-management`
  `~0.18.0` — a **range**, so `helm dep up` takes the newest patch of that minor;
  it bundles `kai-scheduler` itself), `karta`
  (`oci://ghcr.io/dsx-ai-factory/workload-map/karta` 0.2.6), `crds` (`file://charts/crds` 0.0.0, gated
  by `development.includeCRDsInReleaseManifest`). `runai-engine-*.tgz` is a legacy
  tarball also present.
- **CRDs:** `crds/run.ai_clusters.yaml` applied ahead of templates; the full CRD
  set is bundled in the `charts/crds` subchart. Keep in sync with
  `runai-crd-upgrader/crds`.
- **Values schema:** `values.schema.json` validates `values.yaml`.
  `internal_values.yaml` is documentation only — NOT referenced by the chart.
- **Hooks:** `templates/hooks/*-job.yaml` run `helm-hooks-handler` (one flag per
  hook) and `runai-crd-upgrader`; `kartaMigrate` pre/post-upgrade jobs. Migration
  hooks self-gate via `engineToKaiMigration` / `resourceManagementMigration`.
- **Tests:** helm-unittest suites in `tests/` with golden snapshots in
  `tests/__snapshot__/`. `templates/tests/validate-crds` is a `helm test` hook.

## local-dev.sh iteration

runai-cluster chart re-deploy:
```bash
./local-dev.sh --skip-build-cp --skip-cluster
```
