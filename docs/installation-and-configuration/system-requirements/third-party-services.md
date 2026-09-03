# Third-Party Services

Third-party services are a set of supporting services deployed as part of the NVIDIA Run:ai installation, including Keycloak for authentication and identity management, PostgreSQL as the centralized database, Traefik as the API gateway, NATS for asynchronous communication, and others. This page lists those services, their configuration options, and how to replace select services with externally managed instances. For supported versions of each service, see the [Support Matrix](../support-matrix.md#third-party-service-versions).

## Image Registry

* **Open Source and NVIDIA Supported installations** - All third-party service images are pulled from the official upstream registries on Docker Hub (`docker.io`), Quay (`quay.io`), and the Kubernetes registry (`registry.k8s.io`). No registry credentials are required for these images. Ensure that your cluster nodes can reach these registries, see [External Access to Critical Services](external-access-to-critical-services.md) for the required outbound access.
* **Gov-ready installations** - Customers must obtain hardened, FIPS-enabled images from [Chainguard](https://www.chainguard.dev/) for each supported third-party component. Chainguard images must match the supported version of each component for the NVIDIA Run:ai release in use.
* **Air-gapped installations** - NVIDIA Run:ai provides a script to download and mirror the images.

## Control Plane Services

| Component | Supports External Instance | Description |
| --------- | -------------------------- | ----------- |
| Keycloak | No | Provides identity federation, SSO, and OAuth2/OIDC authentication for the NVIDIA Run:ai platform |
| PostgreSQL | Yes | Stores all platform data including workloads, projects, users, and configuration |
| Traefik | No | API gateway for token verification and micro-service routing |
| NATS | Yes | Message broker for caching and asynchronous communication between platform services |
| Thanos | Yes | Long-term metrics storage for the NVIDIA Run:ai control plane |

### Keycloak

Keycloak provides authentication and identity management for the NVIDIA Run:ai control plane. It is deployed as part of the control plane Helm chart and is not supported as an external service.

The Helm chart uses [Keycloakx](https://artifacthub.io/packages/helm/codecentric/keycloakx) (`keycloakx`).

<Info>

**Note**

The `keycloakx.adminUser` can only be set during the initial installation. The admin password can be changed later through the Keycloak UI, but you must also update `keycloakx.adminPassword` in the Helm chart using `helm upgrade`.

</Info>

| Key | Change | Description |
| --- | ------ | ----------- |
| `keycloakx.adminUser` | Admin username | Defines the username for the Keycloak administrator. Can only be set during initial installation. |
| `keycloakx.adminPassword` | Admin password | Defines the password for the Keycloak administrator. |
| `keycloakx.existingSecret` | Credentials secret | Existing secret name with authentication credentials. |
| `global.keycloakx.host` | Keycloak host path | Overrides the DNS for Keycloak. Use this to access Keycloak externally to the cluster. |

#### Changing the Keycloak Admin Password

1. Open the Keycloak UI at: `https://<runai-domain>/auth`
2. Sign in with your existing admin credentials
3. Go to **Users** and select **admin** (or your admin username)
4. Open **Credentials** and select **Reset password**
5. Set the new password and click **Save**
6. Run `helm upgrade` and set `keycloakx.adminPassword` to the new password

<Info>

**Note**

Failing to update the Helm values after changing the password can cause control plane services to encounter errors.

</Info>

### PostgreSQL

PostgreSQL is the centralized database for the NVIDIA Run:ai control plane. It is deployed as part of the control plane Helm chart using [Bitnami PostgreSQL](https://artifacthub.io/packages/helm/bitnami/postgresql) (`postgresql`). Only PostgreSQL version 16 is supported.

#### Using an External PostgreSQL Instance

To replace the deployed PostgreSQL with an external managed database, set `postgresql.enabled: false`, configure `global.postgresql.auth.*` with the external server connection details, and update `grafana.dbUser` and `grafana.dbPassword`.

#### Configuration

| Key | Change | Description |
| --- | ------ | ----------- |
| `postgresql.enabled` | PostgreSQL installation | If set to `false`, PostgreSQL will not be installed. Use this when replacing with an external instance. |
| `global.postgresql.auth.host` | PostgreSQL host | Hostname or IP address of the PostgreSQL server. Required when using an external instance. |
| `global.postgresql.auth.port` | PostgreSQL port | Port number on which PostgreSQL is running. |
| `global.postgresql.auth.database` | PostgreSQL database name | Name of the database to use. Required when using an external instance. |
| `global.postgresql.auth.username` | PostgreSQL username | Username for connecting to PostgreSQL. |
| `global.postgresql.auth.password` | PostgreSQL password | Password for the PostgreSQL user. |
| `global.postgresql.auth.postgresPassword` | PostgreSQL admin password | Password for the built-in PostgreSQL superuser (`postgres`). |
| `global.postgresql.auth.existingSecret` | Credentials secret | Existing secret name with authentication credentials. |
| `global.postgresql.auth.dbSslMode` | SSL mode | Set the SSL mode. See [Protection Provided in Different Modes](https://www.postgresql.org/docs/current/libpq-ssl.html#LIBPQ-SSL-PROTECTION). `Prefer` mode is not supported. |
| `postgresql.primary.initdb.password` | PostgreSQL admin password | Set the same password as `global.postgresql.auth.postgresPassword` if changed. |
| `postgresql.primary.persistence.storageClass` | Storage class | Use a specific storage class instead of the default. |

#### Changing PostgreSQL Passwords

1. Connect to the PostgreSQL pod:

   ```bash
   kubectl exec -it runai-backend-postgresql-0 -n runai-backend -- bash
   psql -U postgres
   ```

2. Update the passwords:

   ```sql
   ALTER USER "postgres" WITH PASSWORD '<new-postgres-password>';
   ALTER USER "user" WITH PASSWORD '<new-user-password>';
   ```

3. Find the Kubernetes secret (default name: `runai-backend-gen-postgresql`) and update it with the new base64-encoded passwords. The keys are typically `postgres-password` and `password`.

4. Restart all deployments to apply the changes:

   ```bash
   kubectl rollout restart deployment -n runai-backend
   ```

<Info>

**Note**

Failing to restart the pods after updating the secret can cause control plane services to encounter errors.

</Info>

### Traefik

Traefik serves as the API gateway for the NVIDIA Run:ai control plane, handling token verification and micro-service routing. It is deployed as part of the control plane Helm chart and is not supported as an external service.

### NATS

NATS provides caching and asynchronous communication for the NVIDIA Run:ai control plane. It is deployed using [Bitnami NATS](https://artifacthub.io/packages/helm/bitnami/nats) (`nats`).

#### Using an External NATS Instance

To replace the deployed NATS with an external instance, set `nats.enabled: false` and provide the connection URLs and credentials for both cache and pub/sub operations.

#### Configuration

| Key | Change | Description |
| --- | ------ | ----------- |
| `nats.enabled` | NATS installation | If set to `false`, NATS will not be installed. Use this when replacing with an external instance. |
| `global.cache.url` | Cache URL | URL of the NATS instance for cache operations. Required when using an external instance. |
| `global.cache.readUrl` | Cache read URL | Read-only URL of the NATS instance. |
| `global.cache.auth.existingSecret` | Cache credentials | Existing secret name with cache authentication credentials. |
| `global.pubsub.url` | Pub/sub URL | URL of the NATS instance for async communication. Required when using an external instance. |
| `global.pubsub.auth.existingSecret` | Pub/sub credentials | Existing secret name with pub/sub authentication credentials. |

### Thanos

Thanos provides single-tenant metrics storage for the NVIDIA Run:ai control plane. It is deployed using [Bitnami Thanos](https://artifacthub.io/packages/helm/bitnami/thanos) (`thanos`).

#### Using an External Thanos Instance

To replace the deployed Thanos with an external metrics store, set `thanos.enabled: false` and provide the read and write URLs for the external instance.

#### Configuration

| Key | Change | Description |
| --- | ------ | ----------- |
| `thanos.enabled` | Thanos installation | If set to `false`, Thanos will not be installed. Use this when replacing with an external instance. |
| `thanos.receive.persistence.storageClass` | Storage class | Use a specific storage class instead of the default. |
| `metricsService.config.datasourceProvider.staticDatasource.read.url` | Metrics read URL | URL for reading metrics from the Thanos instance. Required when using an external instance. |
| `metricsService.config.datasourceProvider.staticDatasource.write.url` | Metrics write URL | URL for writing metrics to the Thanos instance. Required when using an external instance. |
| `metricsService.config.datasourceProvider.staticDatasource.existingSecret` | Metrics credentials | Existing secret name with Thanos authentication credentials. |

## Cluster Services

| Component | Supports External Instance | Description |
| --------- | -------------------------- | ----------- |
| Valkey | No | In-memory cache for workload state and session data |
| oauth2-proxy | No | Authenticates and authorizes access to workloads that expose external URLs |
| git-sync | No | Syncs Git repositories as data sources into running workloads |

Cluster services do not support external instances. Their upstream images can be overridden via Helm or the `RunaiConfig` custom resource. For the full list of configurable fields and instructions, see [Advanced Cluster Configurations](../configuration/cluster-config.md#third-party-service-images).

### Valkey

Valkey provides workload caching for the NVIDIA Run:ai cluster. It is deployed as part of the cluster Helm chart and is not supported as an external service.

#### Changing the Upstream Image

<Tabs>
<Tab title="Helm">

```yaml
clusterConfig:
  cluster-redis:
    image:
      registry: <registry>
      name: <image-name>
      tag: <tag>
```

</Tab>
<Tab title="runaiconfig">

```yaml
spec:
  cluster-redis:
    image:
      registry: <registry>
      name: <image-name>
      tag: <tag>
```

</Tab>
</Tabs>

### oauth2-proxy

oauth2-proxy handles workload authentication and authorization for the NVIDIA Run:ai cluster. It is deployed as part of the cluster Helm chart and is not supported as an external service.

#### Changing the Upstream Image

<Tabs>
<Tab title="Helm">

```yaml
clusterConfig:
  anyworkload-controller:
    runaiOauth2ProxyImage:
      registry: <registry>
      name: <image-name>
      tag: <tag>
  inference-workload-controller:
    runaiOauth2ProxyImage:
      registry: <registry>
      name: <image-name>
      tag: <tag>
  workload-controller:
    runaiOauth2ProxyImage:
      registry: <registry>
      name: <image-name>
      tag: <tag>
```

</Tab>
<Tab title="runaiconfig">

```yaml
spec:
  anyworkload-controller:
    runaiOauth2ProxyImage:
      registry: <registry>
      name: <image-name>
      tag: <tag>
  inference-workload-controller:
    runaiOauth2ProxyImage:
      registry: <registry>
      name: <image-name>
      tag: <tag>
  workload-controller:
    runaiOauth2ProxyImage:
      registry: <registry>
      name: <image-name>
      tag: <tag>
```

</Tab>
</Tabs>

### git-sync

git-sync is a Git data source utility for workloads running on the NVIDIA Run:ai cluster. It is deployed as part of the cluster Helm chart and is not supported as an external service.

#### Changing the Upstream Image

<Tabs>
<Tab title="Helm">

```yaml
clusterConfig:
  inference-workload-controller:
    gitSyncImage:
      registry: <registry>
      name: <image-name>
      tag: <tag>
  workload-controller:
    gitSyncImage:
      registry: <registry>
      name: <image-name>
      tag: <tag>
```

</Tab>
<Tab title="runaiconfig">

```yaml
spec:
  inference-workload-controller:
    gitSyncImage:
      registry: <registry>
      name: <image-name>
      tag: <tag>
  workload-controller:
    gitSyncImage:
      registry: <registry>
      name: <image-name>
      tag: <tag>
```

</Tab>
</Tabs>
