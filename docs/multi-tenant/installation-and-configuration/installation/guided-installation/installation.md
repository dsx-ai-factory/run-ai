---
---

# Control Plane Installation

This page describes how to install the NVIDIA Run:ai control plane in multi-tenant mode, on its own dedicated Kubernetes cluster.

## Before You Start

Complete [Preparations](preparations.md) before continuing, including the system requirements for your Kubernetes cluster, the external metrics store, any required registry credentials, and downloading the `runai-installer` binary. The installer deploys and configures the control plane itself; it does not provision this underlying infrastructure for you.

## Create a Configuration File

<Note>

**Configuration Templates**

The steps below build the minimum configuration for a multi-tenant control plane installation on standard Kubernetes. Depending on your Kubernetes distribution, workload routing, networking setup, and the prerequisites and workload features you need, you may require additional configuration.

If your use case matches one of the [Configuration Templates](templates.md), use that template instead of building the configuration from scratch. See [Prerequisites and Features](prerequisites-and-features.md) for the components required by different workload types, and the [Configuration Reference](configuration-reference.md) for all supported fields and options.

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

   - `controlPlane.domain` - The wildcard FQDN for the NVIDIA Run:ai UI and API. See [FQDN and DNS](../../system-requirements/networking-setup.md#fqdn-and-dns).
   - `controlPlane.tls` - The wildcard certificate and private key for the control plane domain, as `certPath` and `keyPath`. See [TLS Certificates](../../system-requirements/networking-setup.md#tls-certificates).
   - `controlPlane.multiTenant` - Set to `true` to install the control plane in multi-tenant mode.
   - `controlPlane.version` - The control plane version to install. Omit it to install the latest supported version.
   - `controlPlane.customHelmValues` - Merges Helm values into the control plane chart, for settings not otherwise exposed by the installer - including the control plane administrator credentials and the external metrics store connection. See [Control Plane Configuration](../../../../installation-and-configuration/configuration/control-plane-config.md) for the full set of keys.
   - `controlPlane.customHelmValues.supportEntitled` - For NVIDIA Run:ai Support-Entitled Release, set to `true`.
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
       metricsService:
         config:
           datasourceUrl: <METRIC_STORE_READ_URL> # example: https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom
       tenantsManager:
         config:
           defaultMetricStore:
             read:
               auth:
                 basic:
                   password: <METRIC_STORE_READ_PASSWORD>
                   username: "1434499"
               url: <METRIC_STORE_READ_URL> # example: https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom
             useXscopeHeader: true
             write:
               auth:
                 basic:
                   password: <METRIC_STORE_WRITE_PASSWORD>
                   username: "1434499"
               url: <METRIC_STORE_WRITE_URL> # example: https://prometheus-prod-56-prod-us-east-2.grafana.net/api/prom
       thanos:
         enabled: false
       authorization:
         config:
           isCSPEnv: true
   ```

4. Configure the networking:

   - `externalConnection` - Configures how traffic reaches the control plane's UI and API through a Kubernetes Ingress Controller or Gateway API implementation already running on your cluster. See [Networking Setup](../../system-requirements/networking-setup.md).

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
    metricsService:
      config:
        datasourceUrl: <METRIC_STORE_READ_URL> # example: https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom
    tenantsManager:
      config:
        defaultMetricStore:
          read:
            auth:
              basic:
                password: <METRIC_STORE_READ_PASSWORD>
                username: "1434499"
            url: <METRIC_STORE_READ_URL> # example: https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom
          useXscopeHeader: true
          write:
            auth:
              basic:
                password: <METRIC_STORE_WRITE_PASSWORD>
                username: "1434499"
            url: <METRIC_STORE_WRITE_URL> # example: https://prometheus-prod-56-prod-us-east-2.grafana.net/api/prom
    thanos:
      enabled: false
    authorization:
      config:
        isCSPEnv: true

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

- **Run the installer** - Once you're ready, run it against the control plane cluster:

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

## Enable Telemetry Service (Support-Entitled)

NVIDIA Run:ai includes telemetry for Support-Entitled Release installations, reporting initial metrics including platform usage and adoption signals, GPU utilization and allocation, and cluster health signals to support proactive troubleshooting and continued product improvement. By default, telemetry is not associated with your organization. You can opt in to identified telemetry by providing an [NGC Service API Key](https://docs.nvidia.com/ngc/latest/ngc-catalog-user-guide.html#generating-a-service-api-key) (SAK) in your configuration file, which associates the telemetry with your NCA ID.

Configure the SAK through `controlPlane.customHelmValues`, using one of the following approaches:

<Accordion title="Pass the SAK Value Directly">

**Pass the SAK Value Directly**

Set the SAK as a plaintext value in the configuration file. The chart creates and manages the Kubernetes Secret for you.

```yaml
controlPlane:
  ...
  customHelmValues:
    supportEntitled: true
    telemetryService:
      telemetryAuth:
        token: <SAK_TOKEN>
```

<Warning>

**Warning**

The SAK value is stored in plaintext in the configuration file. Protect the file accordingly, or use the pre-created Secret approach below.

</Warning>

Re-run the installer against the control plane cluster to apply the configuration.

</Accordion>

<Accordion title="Reference a Pre-Created Secret">

**Reference a Pre-Created Secret**

Create the Secret containing the SAK yourself before running the installer:

```bash
kubectl create secret generic runai-telemetry-sak \
  --namespace runai-backend \
  --from-literal=TELEMETRY_TOKEN=<SAK_TOKEN>
```

Then reference it by name in the configuration file - the installer does not create or manage this Secret. Always set the secret key name explicitly, even if it matches the default (`TELEMETRY_TOKEN`):

```yaml
controlPlane:
  ...
  customHelmValues:
    supportEntitled: true
    telemetryService:
      telemetryAuth:
        existingSecret: runai-telemetry-sak
        secretKey: TELEMETRY_TOKEN
```

Re-run the installer against the control plane cluster to apply the configuration.

</Accordion>

## Retrieve Control Plane Credentials

Retrieve the generated control plane administrator username and password - you'll need these for [API Access Setup](api-access-setup.md) to authenticate your API requests:

```bash
kubectl get configmap runai-backend-tenants-manager -n runai-backend -o jsonpath='{.data.ADMIN_USERNAME}'
kubectl get secret runai-backend-gen-tenants-manager -n runai-backend -o jsonpath='{.data.ADMIN_PASSWORD}' | base64 -d
```

## Command Reference and Troubleshooting

See [Command Reference and Troubleshooting](command-reference.md) for installer flags and common troubleshooting scenarios.
