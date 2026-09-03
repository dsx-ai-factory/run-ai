# Install the Control Plane

## System and Network Requirements
Before installing the NVIDIA Run:ai control plane, validate that the [system requirements](cp-system-requirements.md) and [network requirements ](network-requirements.md)are met.

## Permissions

As part of the installation, you will be required to install the NVIDIA Run:ai control plane [Helm chart](https://helm.sh/). The Helm charts require Kubernetes administrator permissions. You can review the exact objects that are created by the charts using the `--dry-run` on both helm charts.

## Installation

<Info>

**Note**

* To customize the installation based on your environment, see [Advanced control plane configurations](../../infrastructure-setup/advanced-setup/control-plane-config.md).
* PostgreSQL and Keycloakx are installed with default usernames and passwords. To change the default credentials, see [Additional third-party configurations](../../infrastructure-setup/advanced-setup/control-plane-config.md#additional-third-party-configurations).

</Info>

1. Create a `values.yaml` file using the appropriate example below based on your [metrics backend](metrics-store-requirements/#supported-backend):

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

2. Run the following commands to install the NVIDIA Run:ai control plane using your configured `values.yaml` file:

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai --force-update \
  --username='$oauthtoken' \
  --password=<NGC_API_KEY>
helm repo update
helm upgrade -i runai-backend -n runai-backend runai/control-plane \ 
    -f /path/to/your/values.yaml 
```

<Info>

**Note**

* To install a specific version, add `--version <VERSION>` to the install command. You can find all available versions by using the `helm search repo runai/control-plane --versions` command.
* Use the `--dry-run` flag to gain an understanding of what is being installed before the actual installation.

</Info>
