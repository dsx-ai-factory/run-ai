---
---

# Configuration Templates

Use these templates to start with a complete `config.yaml` instead of building one from scratch. Choose the template that best matches your networking setup, then replace the example values with those prepared for your environment.

| Template | Use this when |
|---|---|
| [Full Installation](#full-installation) | You want every setting a control plane install needs in one place - metrics store, Support-Entitled options, and Ingress-based networking. |
| [Gateway API Installation](#gateway-api-installation) | Your environment uses Kubernetes Gateway API instead of an Ingress controller. |
| [Mimir Metrics Store](#mimir-metrics-store) | Your external metrics store is Grafana Mimir instead of Grafana Labs. |

## Before You Use a Template

Complete [Preparations](preparations.md) before using a template. This includes completing the applicable [System Requirements](../../system-requirements/README.md) and having the required wildcard domain, TLS certificate, networking values, registry credentials, and other installation inputs ready.

* See [Control Plane Installation](installation.md#create-a-configuration-file) to understand the minimum configuration needed before adapting a template.
* See [Configuration Reference](configuration-reference.md) for the full set of supported fields and descriptions.

## Full Installation

```yaml
install:
  mode: runai_control_plane_only
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

controlPlane:
  enabled: true
  version: ""
  domain: <WILDCARD_DOMAIN>
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  multiTenant: true
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

## Gateway API Installation

```yaml
install:
  mode: runai_control_plane_only
  connectivity: connected

registry:
  type: ngc
  tokenPath: /path/to/ngc-token # Support-Entitled Release only

controlPlane:
  enabled: true
  version: ""
  domain: <WILDCARD_DOMAIN>
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  multiTenant: true
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
  type: gatewayApi
  gatewayApi:
    name: <GATEWAY_NAME>
    namespace: <GATEWAY_NAMESPACE>
```

## Mimir Metrics Store

Use this in place of the Grafana Labs `customHelmValues` shown in [Control Plane Installation](installation.md) when your external metrics store is Grafana Mimir instead:

```yaml
controlPlane:
  enabled: true
  version: ""
  domain: <WILDCARD_DOMAIN>
  tls:
    certPath: /path/to/cp-cert.pem
    keyPath: /path/to/cp-key.pem
  multiTenant: true
  customHelmValues:
    metricsService.config.datasourceUrl: <METRIC_STORE_READ_URL> # example: http://mimir-query-frontend.monitoring.svc:8080/prometheus
    tenantsManager.config.defaultMetricStore.read.auth.basic.password: ''
    tenantsManager.config.defaultMetricStore.read.auth.basic.username: ''
    tenantsManager.config.defaultMetricStore.read.url: <METRIC_STORE_READ_URL> # example: http://mimir-query-frontend.monitoring.svc:8080/prometheus
    tenantsManager.config.defaultMetricStore.useXscopeHeader: true
    tenantsManager.config.defaultMetricStore.write.auth.basic.password: ''
    tenantsManager.config.defaultMetricStore.write.auth.basic.username: ''
    tenantsManager.config.defaultMetricStore.write.url: <<REPLACE ME WITH MIMIR ENDPOINT>> # example: http://mimir-distributor.monitoring.svc:8080/api/v1/push
    thanos.enabled: false
    authorization.config.isCSPEnv: true

externalConnection:
  type: ingress
  ingress:
    className: <YOUR_INGRESS_CLASS>
```
