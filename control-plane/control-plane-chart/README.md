# control-plane

`control-plane` is the umbrella Helm chart that deploys the entire Run:ai
control plane onto a Kubernetes (or OpenShift) cluster as a single release
(`runai-backend`). It bundles every control-plane service (identity-manager,
tenants-manager, cluster-service, bff-service, frontend, policy-service,
workloads/workloads-manager, org-unit-service, metrics-service, …) together
with the infrastructure they depend on (PostgreSQL, Keycloak, NATS, Thanos) as
versioned subchart dependencies. The chart is owned by the **Devops Team** and
is the artifact `local-dev.sh` installs to bring a control plane up locally.

This directory is a Helm chart, not a Go service — there is no binary and no
Dockerfile here. It packages and templates the container images that the other
control-plane services build.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Chart dependencies (subcharts)](#chart-dependencies-subcharts)
- [Key values](#key-values)
- [Building and packaging](#building-and-packaging)
- [Deployment](#deployment)
- [Testing](#testing)
- [FIPS 140-3 mode](#fips-140-3-mode)
- [OpenShift TLS certificate configuration](#openshift-tls-certificate-configuration)

## Overview

The chart is responsible for:

- **Templating every control-plane service** — each service under
  `templates/<service>/` renders its Deployment/StatefulSet, Service,
  ConfigMaps, secrets, ingress/HTTPRoute and RBAC from the corresponding
  section of `values.yaml`.
- **Bundling infrastructure dependencies** — PostgreSQL, Keycloak (keycloakx),
  NATS and Thanos are pulled in as subcharts and gated behind `*.enabled`
  conditions so they can be swapped for external managed instances.
- **Global configuration fan-out** — the `global` block (image registry/tag,
  domain, ingress, FIPS mode, Kubernetes distribution, multi-tenancy, blob
  storage) is propagated to every managed container so a single value changes
  behaviour across all services.
- **Install/upgrade lifecycle** — helm hooks under `templates/helm-hooks/`
  handle secret generation, database/keycloak migrators and pre-install
  validation.

## Repository layout

```text
.
├── Chart.yaml              # Chart metadata + subchart dependency list
├── Chart.lock              # Pinned resolved dependency versions
├── values.yaml             # Default values for global + every service (~139 KB)
├── charts/                 # Vendored dependency tarballs (.tgz) resolved from Chart.yaml
├── templates/              # Per-service template dirs + shared helpers
│   ├── common/             # Shared named templates (_helpers)
│   ├── helm-hooks/         # Secret-generation / migrator / validation hooks
│   ├── identity-manager/   # One dir per control-plane service
│   ├── ...                 # (assets-service, bff-service, tenants-manager, …)
│   └── validate.yaml       # Fail-fast input validation
├── tests/                  # helm-unittest suites (*_test.yaml) + snapshots
│   └── integrity/          # Go/terratest secret-reference integrity tests
├── control-plane.tgz       # Packaged chart artifact (build output)
├── manifests.yaml          # Rendered output from `make validate` (OPA input)
├── Makefile
└── artifact-metadata.json
```

## Chart dependencies (subcharts)

Declared in [`Chart.yaml`](Chart.yaml) and pinned in [`Chart.lock`](Chart.lock);
the resolved tarballs live in [`charts/`](charts):

| Dependency | Version | Repository | Condition |
|------------|---------|------------|-----------|
| `runai-common-chart` | 1.0.62 | `file://../../runai-common-chart` | always |
| `postgresql` (bitnami) | 15.5.38 | charts.bitnami.com | `postgresql.enabled` |
| `keycloakx` (codecentric) | 7.2.0 | codecentric.github.io | `keycloakx.enabled` |
| `thanos` (bitnami) | 16.0.7 | charts.bitnami.com | `thanos.enabled` |
| `runai-telemetry` → alias `telemetryService` | 1.1.6 | configured in `Chart.yaml` | `telemetryService.enabled,supportEntitled` |
| `nats` (bitnami) | 9.0.28 | charts.bitnami.com | `nats.enabled` |
| `common` (bitnami) | 2.40.0 | charts.bitnami.com | `common.enabled` |
| `unleash-edge` → alias `unleashEdge` | 2.7.4 | getunleash.io | `unleashEdge.enabled` |

The bitnami `common` library chart must be listed last among the bitnami
dependencies — see the note in `Chart.yaml`.

## Key values

[`values.yaml`](values.yaml) is organised as a `global` block plus one top-level
key per service. Highlights:

- **`global.image.registry` / `global.image.tag`** — registry and tag applied to
  every control-plane image. `make build` stamps these from `REGISTRY_URL/REGISTRY`
  and `VERSION`; `local-dev.sh` overrides them to the local registry and tag `0.0.0`.
- **`global.domain`** — the ingress hostname the control plane is served on.
- **`global.ingress`** — ingress on/off, `ingressClass` (nginx/haproxy),
  TLS secret name, and OpenShift route config.
- **`global.fipsMode`** — `off | on | only` (see [FIPS](#fips-140-3-mode)).
- **`global.config.kubernetesDistribution`** — `kubernetes` or `openshift`;
  switches route vs ingress rendering and security contexts.
- **`global.management`** — bootstrap admin user/password (or `existingSecret`).
- **`supportEntitled`** — marks the release as an NVIDIA Run:ai Support-Entitled
  Release (`--set supportEntitled=true`, installer flag `--support-entitled`).
  Gates the `runai-telemetry` (call-home) subchart and its CUSTOMER_ID secrets;
  `telemetryService.enabled` can still opt out explicitly.
- **Per-service blocks** — `identityManager`, `tenantsManager`, `clusterService`,
  `bffService`, `frontend`, `policyService`, `workloads`, `workloadsManager`,
  `orgUnitService`, `metricsService`, `assetsService`, `auditService`,
  `catalogService`, `notificationsService`, `datavolumes`, `cliExposer`,
  `secretGenerator`, migrators (`authorizationMigrator`, `clusterMigrator`,
  `metricStoresMigrator`), etc. — each carries its own `image`, `resources`,
  `replicas`, config and secret references.
- **Subchart blocks** — `postgresql`, `keycloakx`, `nats`, `thanos`,
  `unleashEdge`, `telemetryService` — expose each dependency's own values.

## Building and packaging

Built through the [`Makefile`](Makefile), which includes
`../../build/makefile/control-plane.mk` and `index.mk`. Chart name is
`control-plane`; default `VERSION` is `0.0.0-devel`.

| Target | Description |
|--------|-------------|
| `make build` | Stamp `global.image` / `notificationsService.image` / `telemetryService.image` registry+tag and the `ubi` tag into `values.yaml` (via `yq`/`sed`), run `helm dep up`, then `helm package` the chart to `control-plane.tgz` (using `--app-version`/`--version` = `VERSION`) |
| `make extra` | `helm dep update` — refresh subchart dependencies |
| `make lint` | `build` then `helm lint` |
| `make validate` | `build`, then `helm template` (OpenShift distribution) into `manifests.yaml` and run `conftest` against the OPA policies under `common/opa-policies/kubernetes` |

`make build` rewrites `values.yaml` in place, so the packaged tarball always
carries the intended registry/tag rather than the checked-in defaults.

## Deployment

`local-dev.sh`'s `deploy_control_plane` installs this chart as release
`runai-backend`. It does **not** call `make build`; instead it:

1. Patches `Chart.yaml` `appVersion` to the built image tag `0.0.0`
   (`CP_IMAGE_TAG`), restoring it on exit.
2. Ensures dependency tarballs are present in `charts/` (running
   `helm dependency build` only if a tarball or a `Chart.lock` entry is missing).
3. Installs/upgrades with (abridged):

```bash
helm upgrade --install runai-backend control-plane/control-plane-chart \
  --namespace <cp-namespace> --create-namespace \
  -f .local-dev/values-cp.yaml \
  --set global.domain=<domain> \
  --set global.image.registry=localhost:<port>/<repo> \
  --set global.image.pullPolicy=IfNotPresent \
  --set global.image.tag=0.0.0 \
  --set global.ingress.ingressClass=nginx \
  --set global.config.kubernetesDistribution=kubernetes \
  --set unleashEdge.enabled=false \
  --timeout 6m
```

On upgrades (existing secrets) it adds `--no-hooks` to preserve passwords and
force-restarts deployments/statefulsets, since every local image shares tag
`0.0.0` and Helm would otherwise see no spec change.

To iterate on the chart without rebuilding images or the cluster:

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```

## Testing

Two test layers, wired into the `test` target:

| Target | Description |
|--------|-------------|
| `make test` | `build` + `test-secrets`, then run the [helm-unittest](https://github.com/helm-unittest/helm-unittest) suites in `tests/*_test.yaml` inside the `helmunittest/helm-unittest` container |
| `make test-secrets` | Cross-document integrity check — every service's `secretRef` must resolve to a declared `(secret, key)` pair across several values scenarios; backed by Go + terratest under `tests/integrity` |

The `tests/` directory holds the helm-unittest specs (ingress, HTTPRoute,
security contexts, per-service secret rendering, custom CA, support-entitled
features, …), their `__snapshot__/`, and shared `_values/` and `scenarios/`
fixtures.

## FIPS 140-3 mode

The control plane binaries ship with the Go 1.24+ native FIPS 140-3 validated
cryptographic module embedded (via `GOFIPS140=v1.0.0` at build time). The
runtime mode is controlled by `global.fipsMode`, which is passed through to every
container the chart manages as `GODEBUG=fips140=<mode>`:

```yaml
global:
  fipsMode: "on"   # off | on | only
```

| Mode   | Behavior                                                                            |
| ------ | ----------------------------------------------------------------------------------- |
| `off`  | FIPS module is not used; standard Go crypto. **Default** — non-regulated installs.  |
| `on`   | FIPS module is used for approved algorithms; non-approved algorithms still allowed. |
| `only` | FIPS-only mode; any non-approved algorithm panics.                                  |

## OpenShift TLS certificate configuration

When deploying on OpenShift, you can configure custom TLS certificates for the
routes that are created. This allows you to use a trusted certificate instead of
the default OpenShift certificate.

### Configuration

To configure a custom certificate for OpenShift routes, set the following values:

```yaml
global:
  config:
    kubernetesDistribution: openshift
  ingress:
    enabled: true
    openshift:
      route:
        tls:
          # Certificate in PEM format
          certificate: |
            -----BEGIN CERTIFICATE-----
            Your certificate content here
            -----END CERTIFICATE-----
          # Private key in PEM format
          key: |
            -----BEGIN PRIVATE KEY-----
            Your private key content here
            -----END PRIVATE KEY-----
```

If a certificate or key is not provided, OpenShift will use its default
certificate or key for the routes.

### Troubleshooting

If the certificates are not being applied correctly:

1. Ensure both certificate and key are provided in valid PEM format
2. Check that the certificate and key pair match
3. Verify that the certificate is valid for your domain
