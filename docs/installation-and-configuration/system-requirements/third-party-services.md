# Third-Party Services

Third-party services are a set of supporting services deployed as part of the NVIDIA Run:ai installation, including Keycloak for authentication and identity management, PostgreSQL as the centralized database, Traefik as the API gateway, NATS for asynchronous communication, and others. This page lists those services and how to replace select services with externally managed instances, including any preparation required before installation. For supported versions of each service, see the [Support Matrix](../support-matrix.md#third-party-service-versions). For configuration of the NVIDIA Run:ai-managed instances themselves (built-in tuning, password rotation, and so on), see [Advanced Control Plane Configurations](../configuration/control-plane-config.md).

## Externally Managed Instances

By default, NVIDIA Run:ai deploys every third-party service itself as part of installation. Some services can instead be replaced with an externally managed instance - one you already operate outside the NVIDIA Run:ai installation (for example, a managed cloud database). In that case, NVIDIA Run:ai connects to your existing instance instead of deploying its own.

Each service's table below has a **Supports External Instance** column indicating whether this is available. Where it is, disable that service's built-in deployment and provide its connection details and credentials - see that service's own section for the exact configuration keys.

## Image Registry

Third-party service images (PostgreSQL, Keycloak, and others) are sourced from their official upstream registries: Docker Hub (`docker.io`), Quay (`quay.io`), and the Kubernetes registry (`registry.k8s.io`).

* **Connected installations** - Images are pulled directly from these registries during installation, for Open Source and NVIDIA Run:ai Support-Entitled installations. No registry credentials are required. Ensure that your cluster nodes can reach these registries, see [External Access to Critical Services](external-access-to-critical-services.md) for the required outbound access.
* **Air-gapped installations** - For Open Source and NVIDIA Run:ai Support-Entitled installations, images are pulled from these registries on a connected machine, packed into a tarball along with the platform images, then loaded into your internal registry.

## NVIDIA Run:ai Control Plane Services

| Component | Supports External Instance | Description |
| --------- | -------------------------- | ----------- |
| Keycloak | No | Provides identity federation, SSO, and OAuth2/OIDC authentication for the NVIDIA Run:ai platform |
| PostgreSQL | Yes | Stores all platform data including workloads, projects, users, and configuration |
| Traefik | No | API gateway for token verification and micro-service routing |
| NATS | Yes | Message broker for caching and asynchronous communication between platform services |
| Thanos | Yes | Long-term metrics storage for the NVIDIA Run:ai control plane |

### Keycloak

Keycloak provides authentication and identity management for the NVIDIA Run:ai control plane. It is deployed as part of the control plane Helm chart using [Keycloakx](https://artifacthub.io/packages/helm/codecentric/keycloakx) (`keycloakx`), and is not supported as an external service. For Helm configuration fields and changing the admin password, see [Advanced Control Plane Configurations](../configuration/control-plane-config.md#keycloak).

### PostgreSQL

PostgreSQL is the centralized database for the NVIDIA Run:ai control plane. It is deployed as part of the control plane Helm chart using [Bitnami PostgreSQL](https://artifacthub.io/packages/helm/bitnami/postgresql) (`postgresql`). Only PostgreSQL version 16 is supported. For built-in PostgreSQL tuning and changing passwords, see [Advanced Control Plane Configurations](../configuration/control-plane-config.md#postgresql).

#### Using an External PostgreSQL Instance

To replace the deployed PostgreSQL with an external managed database, set `postgresql.enabled: false` and configure `global.postgresql.auth.*` with the external server connection details.

Before installing, provision the database and role on your external PostgreSQL server:

1.  Create an SQL script file, edit the parameters below, and save it locally:

    * Replace `<DATABASE_NAME>` with a dedicated database name for NVIDIA Run:ai in your PostgreSQL database.
    * Replace `<ROLE_NAME>` with a dedicated role name (user) for the NVIDIA Run:ai database.
    * Replace `<ROLE_PASSWORD>` with a password for the new PostgreSQL role.

    ```sql
    -- Create a new database for runai
    CREATE DATABASE <DATABASE_NAME>; 

    -- Create the role with login and password
    CREATE ROLE <ROLE_NAME>  WITH LOGIN PASSWORD '<ROLE_PASSWORD>'; 

    -- Grant all privileges on the database to the role
    GRANT ALL PRIVILEGES ON DATABASE <DATABASE_NAME> TO <ROLE_NAME>; 
    ```
2.  Run the following command on a machine where PostgreSQL client (`pgsql`) is installed:

    * Replace `<POSTGRESQL_HOST>` with the PostgreSQL IP address or hostname.
    * Replace `<POSTGRESQL_USER>` with the PostgreSQL username.
    * Replace `<POSTGRESQL_PORT>` with the port number where PostgreSQL is running.
    * Replace `<POSTGRESQL_DB>` with the name of your PostgreSQL database.
    * Replace `<SQL_FILE>` with the path to the SQL script created in the previous step.

    ```bash
    psql --host <POSTGRESQL_HOST> \
    --user <POSTGRESQL_USER> \
    --port <POSTGRESQL_PORT> \
    --dbname <POSTGRESQL_DB> \
    -a -f <SQL_FILE> \
    ```

Then configure the chart to use the provisioned database.

#### Configuring External PostgreSQL

| Key | Change | Description |
| --- | ------ | ----------- |
| `postgresql.enabled` | PostgreSQL installation | If set to `false`, PostgreSQL will not be installed. Use this when replacing with an external instance. |
| `global.postgresql.auth.host` | PostgreSQL host | Hostname or IP address of the PostgreSQL server. Required when using an external instance. |
| `global.postgresql.auth.port` | PostgreSQL port | Port number on which PostgreSQL is running. |
| `global.postgresql.auth.database` | PostgreSQL database name | Name of the database to use. Required when using an external instance. |
| `global.postgresql.auth.username` | PostgreSQL username | Username for connecting to PostgreSQL. |
| `global.postgresql.auth.password` | PostgreSQL password | Password for the PostgreSQL user. |
| `global.postgresql.auth.existingSecret` | Credentials secret | Existing secret name with authentication credentials. |
| `global.postgresql.auth.dbSslMode` | SSL mode | Set the SSL mode. See [Protection Provided in Different Modes](https://www.postgresql.org/docs/current/libpq-ssl.html#LIBPQ-SSL-PROTECTION). `Prefer` mode is not supported. |

### Traefik

Traefik serves as the API gateway for the NVIDIA Run:ai control plane, handling token verification and micro-service routing. It is deployed as part of the control plane Helm chart and is not supported as an external service. For additional configuration, see [Advanced Control Plane Configurations](../configuration/control-plane-config.md).

### NATS

NATS provides caching and asynchronous communication for the NVIDIA Run:ai control plane. It is deployed using [Bitnami NATS](https://artifacthub.io/packages/helm/bitnami/nats) (`nats`).

#### Using an External NATS Instance

To replace the deployed NATS with an external instance, set `nats.enabled: false` and provide the connection URLs and credentials for both cache and pub/sub operations.

#### Configuring External NATS

| Key | Change | Description |
| --- | ------ | ----------- |
| `nats.enabled` | NATS installation | If set to `false`, NATS will not be installed. Use this when replacing with an external instance. |
| `global.cache.url` | Cache URL | URL of the NATS instance for cache operations. Required when using an external instance. |
| `global.cache.readUrl` | Cache read URL | Read-only URL of the NATS instance. |
| `global.cache.auth.existingSecret` | Cache credentials | Existing secret name with cache authentication credentials. |
| `global.pubsub.url` | Pub/sub URL | URL of the NATS instance for async communication. Required when using an external instance. |
| `global.pubsub.auth.existingSecret` | Pub/sub credentials | Existing secret name with pub/sub authentication credentials. |

### Thanos

Thanos provides single-tenant metrics storage for the NVIDIA Run:ai control plane. It is deployed using [Bitnami Thanos](https://artifacthub.io/packages/helm/bitnami/thanos) (`thanos`). For built-in Thanos storage tuning, see [Advanced Control Plane Configurations](../configuration/control-plane-config.md#thanos).

#### Using an External Thanos Instance

To replace the deployed Thanos with an external metrics store, set `thanos.enabled: false` and provide the read and write URLs for the external instance.

#### Configuring External Thanos

| Key | Change | Description |
| --- | ------ | ----------- |
| `thanos.enabled` | Thanos installation | If set to `false`, Thanos will not be installed. Use this when replacing with an external instance. |
| `metricsService.config.datasourceProvider.staticDatasource.read.url` | Metrics read URL | URL for reading metrics from the Thanos instance. Required when using an external instance. |
| `metricsService.config.datasourceProvider.staticDatasource.write.url` | Metrics write URL | URL for writing metrics to the Thanos instance. Required when using an external instance. |
| `metricsService.config.datasourceProvider.staticDatasource.existingSecret` | Metrics credentials | Existing secret name with Thanos authentication credentials. |

## NVIDIA Run:ai Cluster Services

NVIDIA Run:ai cluster services support the workloads running on a NVIDIA Run:ai cluster, rather than the platform itself. Each is deployed automatically as part of the cluster Helm chart and does not support an external instance. Their upstream images can be overridden via Helm or the `RunaiConfig` custom resource. For the full list of configurable fields and instructions, see [Advanced Cluster Configurations](../configuration/cluster-config.md#third-party-service-images).

| Component | Supports External Instance | Description |
| --------- | -------------------------- | ----------- |
| Valkey | No | In-memory cache for workload state and session data |
| oauth2-proxy | No | Authenticates and authorizes access to workloads that expose external URLs |
| git-sync | No | Syncs Git repositories as data sources into running workloads |
