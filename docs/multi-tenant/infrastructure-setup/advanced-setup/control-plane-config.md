# Advanced Control Plane Configurations

## Helm Chart Values

The NVIDIA Run:ai control plane installation can be customized to support your environment via Helm [values files](https://helm.sh/docs/chart_template_guide/values_files/) or [Helm install](https://helm.sh/docs/helm/helm_install/) flags. Make sure to restart the relevant NVIDIA Run:ai pods so they can fetch the new configurations.

| Key                                                                                                                                                                                                                                                       | Change                           | Description                                                                                                                                                                                                                                                                |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `global.ingress.ingressClass`                                                                                                                                                                                                                             | Ingress class                    | NVIDIA Run:ai default is using NGINX. If your cluster has a different ingress controller, you can configure the ingress class to be created by NVIDIA Run:ai.                                                                                                              |
| `global.ingress.tlsSecretName`                                                                                                                                                                                                                            | TLS secret name                  | NVIDIA Run:ai requires the creation of a secret with [domain certificate](../../getting-started/installation/cp-system-requirements.md#domain-and-certificate-requirements). If the `runai-backend` namespace already had such a secret, you can set the secret name here. |
| `<service-name>.podLabels`                                                                                                                                                                                                                                | Pod labels                       | Set NVIDIA Run:ai and 3rd party services' [Pod Labels ](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels)in a format of key/value pairs                                                                                                            |
| `<service-name>.tolerations`                                                                                                                                                                                                                              | Pod tolerations                  | Set NVIDIA Run:ai and third-party services' [Pod Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) in list format. These tolerations apply only to the specified service.                                                        |
| `global.tolerations`                                                                                                                                                                                                                                      | Pod tolerations                  | Set NVIDIA Run:ai and 3rd party services' [Pod Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) in list format. These tolerations are applied globally to all supported services.                                               |
| `<service-name>.replicaCount`                                                                                                                                                                                                                             | Pod replicas                     | By default, NVIDIA Run:ai services are deployed with a single replica. Set this value to run the specified service with multiple pod replicas.                                                                                                                             |
| `global.replicaCount`                                                                                                                                                                                                                                     | Pod replicas                     | By default, NVIDIA Run:ai services are deployed with a single replica. Set this value to apply a global replica count to all services that support scaling.                                                                                                                |
| `global.requireDefaultPodAntiAffinity`                                                                                                                                                                                                                    | Enable default pod anti-affinity | <p>When enabled, NVIDIA Run:ai applies a default pod anti-affinity rule that attempts to prevent pods belonging to the same service from being scheduled on the same node.<br/>Default: <code>true</code></p>                                                               |
| <p><code>&#x3C;service-name></code><br/> <code>resources:</code><br/>  <code>limits:</code><br/>    <code>cpu: 500m</code><br/>    <code>memory: 512Mi</code><br/>  <code>requests:</code><br/>    <code>cpu: 250m</code><br/>    <code>memory: 256Mi</code></p> | Pod request and limits           | Set NVIDIA Run:ai and 3rd party services' resources                                                                                                                                                                                                                        |
| `disableIstioSidecarInjection.enabled`                                                                                                                                                                                                                    | Disable Istio sidecar injection  | Disable the automatic injection of Istio sidecars across the entire NVIDIA Run:ai Control Plane services                                                                                                                                                                   |
| `global.affinity`                                                                                                                                                                                                                                         | System nodes                     | <p>Sets the system nodes where NVIDIA Run:ai system-level services are scheduled<br/>Default: Prefer to schedule on nodes that are labeled with <code>node-role.kubernetes.io/runai-system</code></p>                                                                       |
| `global.customCA.enabled`                                                                                                                                                                                                                                 | Certificate authority            | Enables the use of a custom Certificate Authority (CA) in your deployment. When set to `true`, the system is configured to trust a user-provided CA certificate for secure communication.                                                                                  |

## Additional Third-Party Configurations

The NVIDIA Run:ai control plane chart includes multiple sub-charts of third-party components:

* Data store- [PostgreSQL](https://artifacthub.io/packages/helm/bitnami/postgresql) (`postgresql`)
* Metrics Store - [Thanos](https://artifacthub.io/packages/helm/bitnami/thanos) (`thanos`)
* Identity & Access Management - [Keycloakx](https://artifacthub.io/packages/helm/codecentric/keycloakx) (`keycloakx`)
* Analytics Dashboard - [Grafana](https://artifacthub.io/packages/helm/grafana/grafana) (`grafana`)
* Caching, Queue - [NATS](https://artifacthub.io/packages/helm/bitnami/nats) (`nats`)

<Info>

**Tip**

Click on any component to view its chart values and configurations.

</Info>

### PostgreSQL

If you have opted to connect to an [external PostgreSQL database](../../getting-started/installation/cp-system-requirements.md#external-postgres-database-optional), refer to the additional configurations table below. Adjust the following parameters based on your connection details:

1. Disable PostgreSQL deployment - `postgresql.enabled`
2. NVIDIA Run:ai connection details - `global.postgresql.auth`
3. Grafana connection details - `grafana.dbUser`, `grafana.dbPassword`

| Key                                           | Change                            | Description                                                                                                                                                                                   |
| --------------------------------------------- | --------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `postgresql.enabled`                          | PostgreSQL installation           | If set to `false`, PostgreSQL will not be installed.                                                                                                                                          |
| `global.postgresql.auth.host`                 | PostgreSQL host                   | Hostname or IP address of the PostgreSQL server.                                                                                                                                              |
| `global.postgresql.auth.port`                 | PostgreSQL port                   | Port number on which PostgreSQL is running.                                                                                                                                                   |
| `global.postgresql.auth.username`             | PostgreSQL username               | Username for connecting to PostgreSQL.                                                                                                                                                        |
| `global.postgresql.auth.password`             | PostgreSQL password               | Password for the PostgreSQL user specified by `global.postgresql.auth.username`.                                                                                                              |
| `global.postgresql.auth.postgresPassword`     | PostgreSQL default admin password | Password for the built-in PostgreSQL superuser (`postgres`).                                                                                                                                  |
| `global.postgresql.auth.existingSecret`       | Postgres Credentials (secret)     | Existing secret name with authentication credentials.                                                                                                                                         |
| `global.postgresql.auth.dbSslMode`            | Postgres connection SSL mode      | Set the SSL mode. See the full list in [Protection Provided in Different Modes](https://www.postgresql.org/docs/current/libpq-ssl.html#LIBPQ-SSL-PROTECTION). `Prefer` mode is not supported. |
| `postgresql.primary.initdb.password`          | PostgreSQL default admin password | Set the same password as in `global.postgresql.auth.postgresPassword` (if changed).                                                                                                           |
| `postgresql.primary.persistence.storageClass` | Storage class                     | The installation is configured to work with a specific storage class instead of the default one.                                                                                              |

### Thanos

<Info>

**Note**

This section applies to Kubernetes only.

</Info>

| Key                                       | Change        | Description                                                                                      |
| ----------------------------------------- | ------------- | ------------------------------------------------------------------------------------------------ |
| `thanos.receive.persistence.storageClass` | Storage class | The installation is configured to work with a specific storage class instead of the default one. |

### Keycloakx

The `keycloakx.adminUser` can only be set during the initial installation. The admin password can be changed later through the Keycloak UI, but you must also update the `keycloakx.adminPassword` value in the Helm chart using helm upgrade. Failing to update the Helm values after changing the password can lead to control plane services encountering errors.

| Key                        | Change                                                        | Description                                                                                           |
| -------------------------- | ------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| `keycloakx.adminUser`      | User name of the internal identity provider administrator     | This user is the administrator of Keycloak.                                                           |
| `keycloakx.adminPassword`  | Password of the internal identity provider administrator      | This password is for the administrator of Keycloak.                                                   |
| `keycloakx.existingSecret` | Keycloakx Credentials (secret)                                | Existing secret name with authentication credentials.                                                 |
| `global.keycloakx.host`    | KeyCloak (NVIDIA Run:ai internal identity provider) host path | Override the DNS for Keycloak. This can be used to access access Keycloack externally to the cluster. |

### Grafana

| Key                            | Change                                           | Description                                                         |
| ------------------------------ | ------------------------------------------------ | ------------------------------------------------------------------- |
| `grafana.db.existingSecret`    | Grafana database connection credentials (secret) | Existing secret name with authentication credentials.               |
| `grafana.dbUser`               | Grafana database username                        | Username for accessing the Grafana database.                        |
| `grafana.dbPassword`           | Grafana database password                        | Password for the Grafana database user.                             |
| `grafana.admin.existingSecret` | Grafana admin default credentials (secret)       | Existing secret name with authentication credentials.               |
| `grafana.adminUser`            | Grafana username                                 | Override the NVIDIA Run:ai default user name for accessing Grafana. |
| `grafana.adminPassword`        | Grafana password                                 | Override the NVIDIA Run:ai default password for accessing Grafana.  |
