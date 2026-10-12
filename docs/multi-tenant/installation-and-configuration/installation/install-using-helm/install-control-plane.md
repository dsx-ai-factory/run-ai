# Control Plane Installation

This page describes how to install the NVIDIA Run:ai control plane on your Kubernetes cluster using Helm. The control plane provides the central management layer for NVIDIA Run:ai multi-tenant, handling tenant management, resource and access management, as well as workload submission and monitoring across tenant clusters.

This procedure includes:

* Configuring key settings such as domain name, ingress, and administrator credentials
* Deploying the control plane into the `runai-backend` namespace

By completing this process, the NVIDIA Run:ai control plane will be running in your cluster and accessible via the configured domain, ready to onboard tenants.

## Before You Start

Before installing the NVIDIA Run:ai control plane, complete [Preparations](preparations.md).

## Permissions

As part of the installation, you will be required to install the NVIDIA Run:ai control plane [Helm chart](https://helm.sh/). The Helm chart requires Kubernetes administrator permissions. You can review the exact objects that are created by the chart using the `--dry-run` flag.

## Installation

<Note>

**Note**

* To customize the installation based on your environment, see [Advanced Control Plane Configurations](../../../../installation-and-configuration/configuration/control-plane-config.md).
* PostgreSQL and Keycloakx are installed with default usernames and passwords. To change the default credentials, see [Advanced Control Plane Configurations](../../../../installation-and-configuration/configuration/control-plane-config.md).

</Note>

1. Create a `values.yaml` file using the appropriate example below based on your [metrics backend](../../system-requirements/metrics-store-requirements/README.md#supported-backends):

<Tabs>

<Tab title="YAML with Grafana Labs">

```yaml
global:
  domain: <RUNAI_DOMAIN> # example: "runai.hostorg.com"
  multiTenancy: 
    enabled: true
  management:
    user: <ADMIN-USER>
    password: <ADMIN-PASSWORD>
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

</Tab>

<Tab title="YAML with Mimir">

```yaml
global:
  domain: <RUNAI_DOMAIN> # example: "runai.hostorg.com"
  multiTenancy:
    enabled: true
  management:
    user: <ADMIN-USER>
    password: <ADMIN-PASSWORD>
metricsService:
  config:
   datasourceUrl: <METRIC_STORE_READ_URL> # example: http://mimir-query-frontend.monitoring.svc:8080/prometheus
tenantsManager:
  config:
    defaultMetricStore:
      read:
        auth:
          basic:
            password: ''
            username: ''
        url: <METRIC_STORE_READ_URL> # example: http://mimir-query-frontend.monitoring.svc:8080/prometheus
      useXscopeHeader: true
      write:
        auth:
          basic:
            password: ''
            username: ''
        url: <<REPLACE ME WITH MIMIR ENDPOINT>> # example: http://mimir-distributor.monitoring.svc:8080/api/v1/push
thanos:
  enabled: false
authorization:
  config:
    isCSPEnv: true
```

</Tab>

</Tabs>

2. Run the following commands to install the NVIDIA Run:ai control plane using your configured `values.yaml` file. Use the commands for the release type you are installing.

<Tabs>

<Tab title="Open Source">

**Open Source**

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai-oss --force-update
helm repo update
helm upgrade -i runai-backend -n runai-backend runai/control-plane \
    -f /path/to/your/values.yaml 
```

</Tab>

<Tab title="Support-Entitled">

**Support-Entitled**

In addition to the values in your `values.yaml` file, update the Helm commands as follows:

* Add `--set supportEntitled=true` to the `helm upgrade` command.
* Add `--set telemetryService.enabled=true` to the `helm upgrade` command to enable telemetry.
* Add `--set global.imagePullSecrets[0].name=runai-reg-creds` to the `helm upgrade` command to reference the secret created in [Registry Credentials](preparations.md#registry-credentials).
* Optional: associate NVIDIA Run:ai's telemetry with your NCA ID by adding one of the following to the `helm upgrade` command. See [Registry Credentials](preparations.md#registry-credentials) for how to generate the NGC Service API Key. By default, telemetry is not associated with your organization.
  - `--set telemetryService.telemetryAuth.token=<SERVICE_API_KEY>` - Passes the Service API Key directly. Visible in shell history and process listings.
  - `--set telemetryService.telemetryAuth.existingSecret=<SECRET_NAME> --set telemetryService.telemetryAuth.secretKey=<KEY_NAME>` - References a secret you create yourself beforehand - the chart does not create or manage it.

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai-oss --force-update
helm repo update
helm upgrade -i runai-backend -n runai-backend runai/control-plane \
    -f /path/to/your/values.yaml \
    --set supportEntitled=true \
    --set telemetryService.enabled=true \
    --set global.imagePullSecrets[0].name=runai-reg-creds \
    --set telemetryService.telemetryAuth.token=<SERVICE_API_KEY>
```

</Tab>

</Tabs>

<Note>

**Note**

* To install a specific version, add `--version <VERSION>` to the install command. You can find all available versions by using the `helm search repo runai/control-plane --versions` command.
* Use the `--dry-run` flag to gain an understanding of what is being installed before the actual installation.

</Note>

## Next Steps

Set up [API Access](../guided-installation/api-access-setup.md) to authenticate your API requests, then onboard your first tenant. See [Create a Tenant and Cluster](create-a-tenant-and-cluster.md). Before installing a tenant's cluster, make sure you have completed the required setup in [Prerequisites and Features](prerequisites-and-features.md) on that tenant's own Kubernetes cluster.
