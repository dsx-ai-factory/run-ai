---
---

# Control Plane Installation

This page describes how to install the NVIDIA Run:ai control plane in multi-tenant mode, on its own dedicated Kubernetes cluster.

## Before You Start

Complete [Preparations](preparations.md) before continuing, including the system requirements for your Kubernetes cluster, the external metrics store, any required registry credentials, and downloading the `runai-installer` binary. The installer deploys and configures the control plane itself; it does not provision this underlying infrastructure for you.

## Create a Configuration File

<Note>

**Configuration Templates**

The steps below build the minimum configuration for a multi-tenant control plane installation on standard Kubernetes. Depending on your Kubernetes distribution and networking setup, you may require additional configuration.

* Once you understand the minimum configuration, see [Configuration Templates](templates.md) - if your use case matches one, use it as your starting point instead of building the configuration from scratch.
* See [Prerequisites and Features](prerequisites-and-features.md) to understand the components required, depending on the workload types you plan to enable.
* See [Configuration Reference](configuration-reference.md) for the full set of supported fields and descriptions, to build a configuration from scratch.

</Note>

Create a `cp-config.yaml` file for the control plane cluster. Use the domain, TLS certificate, and networking values you prepared in [Networking Setup](../../system-requirements/networking-setup.md) when completing the configuration below.

At a minimum, configure:

1. Define the installation:

   - `install.mode` - Set to `runai_control_plane_only`.
   - `install.connectivity` - Set to `connected`, for an installation with outbound internet access.

   ```yaml
   install:
     mode: runai_control_plane_only
     connectivity: connected
   ```

2. Configure the registry:

   - `registry.type` - Set to `ngc`.
   - `registry.tokenPath` - For NVIDIA Run:ai Support-Entitled Release installations, provide the path to your NGC Personal Key. Open Source installations do not require registry credentials. See [Registry Credentials](preparations.md#registry-credentials).

   ```yaml
   registry:
     type: ngc
     tokenPath: /path/to/ngc-token # Support-Entitled Release only
   ```

3. Define the control plane:

   - `controlPlane.domain` - The wildcard FQDN for the NVIDIA Run:ai UI and API. See [Domains and TLS Certificates](preparations.md#domains-and-tls-certificates).
   - `controlPlane.tls` - The wildcard certificate and private key for the control plane domain, as `certPath` and `keyPath`. See [Domains and TLS Certificates](preparations.md#domains-and-tls-certificates).
   - `controlPlane.multiTenant` - Set to `true` to install the control plane in multi-tenant mode.
   - `controlPlane.version` - The control plane version to install. Omit it to install the latest supported version.
   - `controlPlane.customHelmValues` - Merges Helm values into the control plane chart, for settings not otherwise exposed by the installer - including the control plane administrator credentials and the external metrics store connection. See [Advanced Control Plane Configurations](../../../../installation-and-configuration/configuration/control-plane-config.md) for the full set of keys.
   - `controlPlane.customHelmValues.supportEntitled` - For NVIDIA Run:ai Support-Entitled Release, must be set to `true`.
   - `controlPlane.customHelmValues.telemetryService.enabled` - For NVIDIA Run:ai Support-Entitled Release, must be set to `true` to enable telemetry.
   - `controlPlane.customHelmValues.telemetryService.telemetryAuth` - Optional, Support-Entitled Release only. Associates NVIDIA Run:ai's telemetry with your NCA ID - by default, telemetry is not associated with your organization. See [Registry Credentials](preparations.md#registry-credentials) for how to generate the NGC Service API Key. Provide it one of two ways:
     - `token: <SERVICE_API_KEY>` - Sets the Service API Key directly; the chart creates and manages the Kubernetes secret for you. Stored in plaintext in the configuration file, so protect it accordingly, or use `existingSecret` instead.
     - `existingSecret: <SECRET_NAME>` / `secretKey: <KEY_NAME>` - References a secret you create yourself beforehand - the installer does not create or manage it.
   - The external metrics store connection - required for every multi-tenant control plane. The example below configures Grafana Labs; for Mimir, see the [Mimir Metrics Store](templates.md#mimir-metrics-store) template. See [Metrics Store Requirements](../../system-requirements/metrics-store-requirements/README.md) for both.

   ```yaml
   controlPlane:
     enabled: true
     domain: <WILDCARD_DOMAIN>
     tls:
       certPath: /path/to/cp-cert.pem
       keyPath: /path/to/cp-key.pem
     multiTenant: true
     version: "" # Omit to install the latest supported version
     customHelmValues:
       supportEntitled: true # Support-Entitled Release only
       telemetryService.enabled: true # Support-Entitled Release only
       telemetryService.telemetryAuth.token: <SERVICE_API_KEY> # Optional
       metricsService.config.datasourceUrl: <METRIC_STORE_READ_URL> # example: https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom
       tenantsManager.config.defaultMetricStore.read.auth.basic.password: <METRIC_STORE_READ_PASSWORD>
       tenantsManager.config.defaultMetricStore.read.auth.basic.username: "1434499"
       tenantsManager.config.defaultMetricStore.read.url: <METRIC_STORE_READ_URL> # example: https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom
       tenantsManager.config.defaultMetricStore.useXscopeHeader: true
       tenantsManager.config.defaultMetricStore.write.auth.basic.password: <METRIC_STORE_WRITE_PASSWORD>
       tenantsManager.config.defaultMetricStore.write.auth.basic.username: "1434499"
       tenantsManager.config.defaultMetricStore.write.url: <METRIC_STORE_WRITE_URL> # example: https://prometheus-prod-56-prod-us-east-2.grafana.net/api/prom
       thanos.enabled: false
       authorization.config.isCSPEnv: true
   ```

4. Configure external access to the control plane:

   - `externalConnection` - Configures how traffic reaches the control plane's UI and API from outside the cluster, through a Kubernetes Ingress Controller or Gateway API implementation already running on your cluster. See [Networking Setup](../../system-requirements/networking-setup.md).

   ```yaml
   externalConnection:
     type: ingress
     ingress:
       className: <YOUR_INGRESS_CLASS>
   ```

The following example combines these settings to install the NVIDIA Run:ai control plane in multi-tenant mode:

```yaml
install:
  mode: runai_control_plane_only
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

controlPlane:
  enabled: true
  domain: <WILDCARD_DOMAIN>
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  multiTenant: true
  version: ""
  customHelmValues:
    supportEntitled: true # Support-Entitled Release only
    metricsService.config.datasourceUrl: <METRIC_STORE_READ_URL> # example: https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom
    tenantsManager.config.defaultMetricStore.read.auth.basic.password: <METRIC_STORE_READ_PASSWORD>
    tenantsManager.config.defaultMetricStore.read.auth.basic.username: "1434499"
    tenantsManager.config.defaultMetricStore.read.url: <METRIC_STORE_READ_URL> # example: https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom
    tenantsManager.config.defaultMetricStore.useXscopeHeader: true
    tenantsManager.config.defaultMetricStore.write.auth.basic.password: <METRIC_STORE_WRITE_PASSWORD>
    tenantsManager.config.defaultMetricStore.write.auth.basic.username: "1434499"
    tenantsManager.config.defaultMetricStore.write.url: <METRIC_STORE_WRITE_URL> # example: https://prometheus-prod-56-prod-us-east-2.grafana.net/api/prom
    thanos.enabled: false
    authorization.config.isCSPEnv: true

externalConnection:
  type: ingress
  ingress:
    className: <YOUR_INGRESS_CLASS>
```

## Run the Installer

- **Validate first** - Before applying any changes, check your configuration against the control plane cluster:

  ```bash
  ./runai-installer install \
    --config cp-config.yaml \
    --kubeconfig /path/to/cp-kubeconfig \
    --preflight-only
  ```

- **Preview the plan** - Go further and preview the full installation plan without applying it:

  ```bash
  ./runai-installer install \
    --config cp-config.yaml \
    --kubeconfig /path/to/cp-kubeconfig \
    --dry-run
  ```

- **Run the installer** - Once you are ready, run it against the control plane cluster:

  ```bash
  ./runai-installer install \
    --config cp-config.yaml \
    --kubeconfig /path/to/cp-kubeconfig \
    --logs-dir ./logs
  ```

When the installation completes, verify that the control plane is reachable at its domain (for example, `https://runai.hostorg.com`) before continuing.

## Verify the Installation

After the installer run completes successfully, verify that the NVIDIA Run:ai pods are running:

```bash
kubectl get pods -n runai-backend
```

All pods should reach `Running` or `Completed` status.

## Retrieve Control Plane Credentials

Retrieve the generated control plane administrator username and password - you'll need these for [API Access Setup](api-access-setup.md) to authenticate your API requests:

```bash
kubectl get configmap runai-backend-tenants-manager -n runai-backend -o jsonpath='{.data.ADMIN_USERNAME}'
kubectl get secret runai-backend-gen-tenants-manager -n runai-backend -o jsonpath='{.data.ADMIN_PASSWORD}' | base64 -d
```

## Command Reference and Troubleshooting

See [Command Reference and Troubleshooting](command-reference.md) for installer flags and common troubleshooting scenarios.
