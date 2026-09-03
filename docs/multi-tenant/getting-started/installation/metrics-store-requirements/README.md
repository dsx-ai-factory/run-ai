# Metrics Store Requirements

In a  multi-tenant deployment, integrating a multi-tenant metrics store to is required to support:

* Tenant usage reporting
* System and workload monitoring

NVIDIA Run:ai components rely on Prometheus-compatible metrics for dashboards, APIs, and backend decision-making. The metrics store must:

* Support PromQL (Prometheus Query Language)
* Scale to support multiple tenants concurrently
* Retain time-series data reliably for reporting and analysis

## Supported Backends

| Option                           | Description                                                                                                                |
| -------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| Grafana Labs (hosted Prometheus) | Recommended managed service. Offers multi-tenancy, scalability, and ease of integration with NVIDIA Run:ai.                |
| Grafana Mimir (self-hosted)      | Supported self-managed alternative. The host organization is responsible for installation, configuration, and maintenance. |

NVIDIA Run:ai components will query metrics from the configured store to power dashboards, reports, and scheduling decisions. Ensure availability and performance SLAs are in place.

## Connecting Grafana Labs (Hosted Prometheus)

To connect NVIDIA Run:ai to Grafana Labs (hosted Prometheus), you need a Grafana Cloud Access Policy token. This token authenticates API requests and enables secure access to your metrics data.

1. Create an access token following the [Create access policies and tokens – Grafana Cloud Docs](https://grafana.com/docs/grafana-cloud/security-and-account-management/authentication-and-permissions/access-policies/create-access-policies/).
2. Create a values file (e.g., `grafanalabs-values.yaml`) with your hosted Prometheus endpoint and access token.
3. Add the file during [control plane installation](../install-control-plane.md#installation):

```yaml
thanos:
 enabled: false
tenantsManager:
 config:
   grafanaLab:
     accessToken: <<GRAFANA LAB ACCESS TOKEN>>
```

This configuration tells the NVIDIA Run:ai platform where to send PromQL queries for tenant insights and metrics.

## Grafana Mimir Integration

If you choose to use Grafana Mimir as your metrics store, follow the steps below to ensure compatibility, security, and observability.

### Installation and Configuration

Follow the official [Grafana Mimir](https://grafana.com/docs/helm-charts/mimir-distributed/latest/get-started-helm-charts/) Helm chart documentation for installation. NVIDIA Run:ai has optimized Mimir compatibility. You can review all configurable options here: [Mimir Configuration Parameters](https://grafana.com/docs/mimir/latest/configure/configuration-parameters/).

### Prerequisites

Make sure you have the following before deploying Mimir:

* **TLS certificate** (private and public) - Used to secure HTTPS access to the metrics store. This should be a dedicated certificate specifically for the Mimir deployment.
* **FQDN for Mimir access** (e.g., mimir.runai.hostorg.com) - This must resolve to the Mimir service endpoint. Use a dedicated domain reserved for the metrics store.

### Helm Values Template

NVIDA Run:ai provides a tested `values.yaml` configuration for Helm-based Mimir installation. See [NVIDIA Run:ai Mimir Helm Chart](nvidia-run-ai-mimir-helm-chart.md).

### Connecting Mimir to the NVIDIA Run:ai Control Plane

To integrate Mimir with the NVIDIA Run:ai control plane, include the required Mimir specific configuration values in your `values.yaml` file when installing or upgrading the control plane. See [Install the control plane](../install-control-plane.md) for more details:

```yaml
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
       url: <METRIC_STORE_WRITE_URL> # example: http://mimir-distributor.monitoring.svc:8080/api/v1/push

thanos:
 enabled: false

```

### Monitoring and Debugging Mimir

Grafana Labs offers a collection of dashboards and alerts for monitoring a self-hosted Mimir. NVIDIA Run:ai utilizes these dashboards to monitor our Mimir instance. To deploy these dashboards and alerts, see [About Grafana Mimir dashboards and alerts requirements](https://grafana.com/docs/mimir/latest/manage/monitor-grafana-mimir/requirements/).

