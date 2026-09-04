# control-plane (chart) — Build Context

**Type:** helm
**Team:** Devops Team (`artifact-metadata.json`)
**Service name:** control-plane
**Artifact:** `control-plane.tgz` (packaged chart; `cache_artifacts: true`, `force_build: true`)

This is a Helm chart — no Go module, no Dockerfile, no binary. It packages and
templates the images that the other control-plane services build.

## Make targets

Targets come from the chart `Makefile` plus the shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`).

| Target | What it does |
|--------|-------------|
| `build` | Stamp registry+tag into `values.yaml` (`global.image`, `notificationsService.image`, `telemetryService.image`, and the `ubi` tag) via `yq`/`sed`, `helm dep up`, then `helm package` → `control-plane.tgz` with `--app-version`/`--version` = `VERSION` (default `0.0.0-devel`) |
| `extra` | `helm dep update` |
| `lint` | `build` + `helm lint` |
| `validate` | `build` + `helm template` (OpenShift distribution) → `manifests.yaml`, then `conftest` against `common/opa-policies/kubernetes` |
| `test` | `build` + `test-secrets`, then `helmunittest/helm-unittest` over `tests/*_test.yaml` |
| `test-secrets` | Secret-reference integrity tests across scenarios (`cd tests/integrity && go test ./...`) |

## Key build notes

- **Subchart dependencies** are declared in `Chart.yaml`, pinned in `Chart.lock`,
  and vendored as `.tgz` under `charts/`: runai-common-chart (file://), plus
  bitnami postgresql / thanos / nats / common, codecentric keycloakx,
  getunleash unleash-edge, and jfrog runai-telemetry. The bitnami `common`
  library must remain last among the bitnami deps.
- **`make build` rewrites `values.yaml` in place** (registry/tag stamping) — do
  not commit those local mutations.
- **`values.yaml` is ~139 KB**: a `global` block plus one top-level key per
  service (`identityManager`, `tenantsManager`, `clusterService`, `bffService`,
  `frontend`, `policyService`, `workloads`, `workloadsManager`, `orgUnitService`,
  `metricsService`, migrators, …) and one per subchart.
- **Deploy (what local-dev.sh actually runs):** `deploy_control_plane` does NOT
  call `make build`. It patches `Chart.yaml` `appVersion` → `0.0.0` (`CP_IMAGE_TAG`,
  restored on exit), ensures `charts/*.tgz` are present (only running
  `helm dependency build` if a tarball or `Chart.lock` entry is missing), then:
  `helm upgrade --install runai-backend control-plane/control-plane-chart -n <cp-ns> --create-namespace -f .local-dev/values-cp.yaml --set global.image.registry=localhost:<port>/<repo> --set global.image.tag=0.0.0 --set global.config.kubernetesDistribution=kubernetes --set unleashEdge.enabled=false --timeout 6m`.
  On upgrades it adds `--no-hooks` (preserve secrets) and force-restarts
  deployments/statefulsets, because every local image shares tag `0.0.0` so Helm
  sees no spec change. **Image tag `0.0.0` is load-bearing** — it must match
  `Chart.yaml` `appVersion` and the tags the CP services build.
- **FIPS:** `global.fipsMode` (`off|on|only`) → `GODEBUG=fips140=<mode>` on every container.
- **helm-unittest** suites and their snapshots live in `tests/`; `tests/integrity`
  is a Go/terratest secret-reference cross-document check.

## local-dev.sh iteration

This is a control-plane artifact:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
