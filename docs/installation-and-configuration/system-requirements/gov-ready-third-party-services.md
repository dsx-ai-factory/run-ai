# Gov-Ready Third-Party Services

Third-party services are a set of supporting services deployed as part of the NVIDIA Run:ai installation, including Keycloak for authentication and identity management, PostgreSQL as the centralized database, Traefik as the API gateway, NATS for asynchronous communication, and others. A Gov-ready installation replaces the default images of these services with FIPS-validated hardened images. This page lists the services you change, the registry you pull the hardened images from, and the Helm values to set, including any preparation required before installation. Complete these steps before installing the control plane and cluster. For supported versions of each service, refer to the [Support Matrix](../support-matrix.md#third-party-service-versions).

## Image Registry

Gov-ready third-party images are FIPS-enabled Docker Hardened Images (DHI). You pull them from Docker Hub (`docker.io`). For the outbound access required to reach it, refer to [External Access to Critical Services](external-access-to-critical-services.md#gov-ready).

### Getting Access

FIPS-enabled Docker Hardened Images require a paid Docker subscription (DHI Select or DHI Enterprise). To get access:

1. Subscribe to Docker Hardened Images.
2. Mirror the repositories of the images listed on this page to your Docker Hub organization. The `registry` and `repository` values in the Helm values on this page are the ones of your mirrored repositories.
3. Create an access token with at least read-only access to those repositories. Use an Organization Access Token for an organization, or a Personal Access Token for a personal account.

For the exact steps, refer to the Docker documentation on [using Docker Hardened Images](https://docs.docker.com/dhi/how-to/use/).

### Providing Credentials

The Docker Hub credentials work the same way in [Guided Installation](../installation/guided-installation/README.md) and [Manual Helm](../installation/install-using-helm/preparations.md). How you provide them depends on whether your environment is connected or air-gapped.

<Accordion title="Connected">

**Connected**

1. Create the control plane and cluster namespaces if they do not exist yet. In Manual Helm, [Preparations](../installation/install-using-helm/preparations.md) already includes this step. In Guided Installation, the installer normally creates the namespaces, so create them now only to have a place for the secret. The installer uses the existing namespaces.

   ```bash
   kubectl create namespace runai-backend
   kubectl create namespace runai
   ```

2. Create a Kubernetes image pull secret named `dhi-reg-creds` from your access token in each namespace. You can use another name, but it must differ from `runai-reg-creds`, which holds the NVIDIA Run:ai registry credentials. Use your Docker ID as the username for a Personal Access Token, or your organization name for an Organization Access Token.

   ```bash
   kubectl create secret docker-registry dhi-reg-creds \
   --docker-server=docker.io \
   --docker-username=<docker-id-or-organization> \
   --docker-password=<access-token> \
   --docker-email=<email> \
   --namespace=<namespace>
   ```

During the installation, reference the secrets in the Helm values. List both secrets: `runai-reg-creds` for the NVIDIA Run:ai images from NGC, and `dhi-reg-creds` for the hardened images from Docker Hub.

* **Manual Helm** - You create `runai-reg-creds` as described in [Software Artifacts](../installation/install-using-helm/preparations.md#software-artifacts). Add the values to the control plane and cluster values you pass to Helm.
* **Guided Installation** - The installer creates `runai-reg-creds` for you. Add the values under `controlPlane.customHelmValues` and `cluster.customHelmValues` in the installer configuration.

```yaml
# Control plane
global:
  imagePullSecrets:
    - name: runai-reg-creds
    - name: dhi-reg-creds
keycloakx:
  imagePullSecrets:
    - name: runai-reg-creds
    - name: dhi-reg-creds
```

```yaml
# Cluster
global:
  additionalImagePullSecrets:
    - runai-reg-creds
    - dhi-reg-creds
```

</Accordion>

<Accordion title="Air-gapped">

**Air-gapped**

Your cluster cannot reach Docker Hub, so you do not create a Kubernetes secret. Instead, add the hardened images and your Docker Hub credentials when you prepare the air-gapped bundle, in either Guided Installation or Manual Helm.

You do this in the pack configuration, the file you pass to `airgap pack`. Add an `imageOverrides` entry for each hardened image, so that the bundle contains it instead of the default image, and a `registryAuth` entry with your Docker Hub credentials. Use your Docker ID as the username for a Personal Access Token, or your organization name for an Organization Access Token, and the access token as the password.

```yaml
imageOverrides:
  # Control plane
  - name: keycloak
    source: docker.io/<organization>/<dhi-keycloak-repository>:<fips-ready-tag>
  - name: postgres
    source: docker.io/<organization>/<dhi-postgres-repository>:<fips-ready-tag>
  - name: traefik
    source: docker.io/<organization>/<dhi-traefik-repository>:<fips-ready-tag>
  - name: nats
    source: docker.io/<organization>/<dhi-nats-repository>:<fips-ready-tag>
  - name: thanos
    source: docker.io/<organization>/<dhi-thanos-repository>:<fips-ready-tag>
  # Cluster
  - name: valkey
    source: docker.io/<organization>/<dhi-valkey-repository>:<fips-ready-tag>
  - name: git-sync
    source: docker.io/<organization>/<dhi-git-sync-repository>:<fips-ready-tag>

registryAuth:
  - registry: docker.io
    username: <docker-id-or-organization>
    passwordPath: /path/to/docker-access-token
```

Each `source` is the full image reference of the hardened image in your mirrored repository, including the tag. The `postgres` entry is the PostgreSQL wait image.

Then pack, load, and install as usual. For the full procedure, refer to [Air-Gapped Bundle](../installation/guided-installation/air-gapped/preparations.md#air-gapped-bundle) in Guided Installation, or [Preparations](../installation/install-using-helm/preparations.md) in Manual Helm. For the pack configuration fields, refer to [Pack an Air-Gapped Bundle](../installation/guided-installation/configuration-reference.md#pack-an-air-gapped-bundle).

</Accordion>

## Where to Set the Values

Write the values in a file or in the installer configuration, depending on the installation path:

* **Manual Helm** - Save the values in a YAML file and pass it with `-f` to the `helm upgrade` commands in [Install the Control Plane](../installation/install-using-helm/install-control-plane.md#installation) and [Install the Cluster](../installation/install-using-helm/helm-install.md#installation). Use one file for the control plane values and one for the cluster values, because they belong to different charts.
* **Guided Installation** - Add the control plane values under `controlPlane.customHelmValues` and the cluster values under `cluster.customHelmValues` in the installer configuration.

<Note>

**Note**

In air-gapped installations, you do not set the image `registry`, `repository`, and `tag` values in the sections below. You enter the image locations in the pack configuration, and the values generated by the load step set them for the install. For details, refer to the Air-gapped section in [Providing Credentials](#providing-credentials).

</Note>

## Control Plane Services

Set the following values in the control plane values, as described in [Where to Set the Values](#where-to-set-the-values).

In a Gov-ready installation, PostgreSQL must be an externally managed instance. NATS and Thanos can also be replaced with externally managed instances, as in a standard installation. The image changes for NATS and Thanos apply when you use the instances deployed with NVIDIA Run:ai.

| Component | Supports External Instance | Description | Gov-Ready Change |
| --------- | -------------------------- | ----------- | ---------------- |
| Keycloak | No | Provides identity federation, SSO, and OAuth2/OIDC authentication for the NVIDIA Run:ai platform | Replace the image, add the `fips` feature to the start command, and run as user ID `65532` |
| PostgreSQL | Yes. Required for Gov-ready. | Stores all platform data including workloads, projects, users, and configuration | Use your own FIPS-ready instance, and replace the PostgreSQL wait image |
| Traefik | No | API gateway for token verification and micro-service routing | Replace the image |
| NATS | Yes | Message broker for caching and asynchronous communication between platform services | Replace the image and override the command |
| Thanos | Yes | Long-term metrics storage for the NVIDIA Run:ai control plane | Replace the image |

### Keycloak

Keycloak provides authentication and identity management for the NVIDIA Run:ai control plane. It is deployed as part of the control plane Helm chart using [Keycloakx](https://artifacthub.io/packages/helm/codecentric/keycloakx) (`keycloakx`), and is not supported as an external service. For Helm configuration fields and changing the admin password, refer to [Advanced Control Plane Configurations](../configuration/control-plane-config.md#keycloak).

Replace the Keycloak image, add the `fips` feature to the start command, and run the container as user ID `65532`. The hardened image restricts its files to that user, and Keycloak does not start without it. Compared with the default start command, the only difference is `fips` added to `--features`. Because Helm replaces the whole `command` value, the example shows the full command.

```yaml
keycloakx:
  image:
    repository: <registry>/<dhi-keycloak-repository>
    tag: <fips-ready-tag>                # tested with 26.6
  securityContext:
    runAsUser: 65532
  command:
    - /bin/sh
    - -c
    - |
      KC_BOOTSTRAP_ADMIN_USERNAME="$(cat /run/secrets/keycloak/admin-username)" \
      KC_BOOTSTRAP_ADMIN_PASSWORD="$(cat /run/secrets/keycloak/admin-password)" \
      exec /opt/keycloak/bin/kc.sh start \
        --features=token-exchange,admin-fine-grained-authz:v1,fips
```

### PostgreSQL

PostgreSQL is the centralized database for the NVIDIA Run:ai control plane. Gov-ready installations always use an externally managed PostgreSQL instance that is FIPS-ready. You provide and manage the instance. Only PostgreSQL version 16 is supported.

#### Using an External PostgreSQL Instance

To use your external instance, set `postgresql.enabled: false`, configure `global.postgresql.auth.*` with the connection details of your server, and update `grafana.dbUser` and `grafana.dbPassword`.

Before installing, provision the database, role, and Grafana user on your PostgreSQL server. NVIDIA Run:ai uses the role to create a user for each service, and each service has its own schema in the database. Each service user has permissions only on its own schema. The role owns the database, so that it has administrative permissions over it:

1.  Create an SQL script file, edit the parameters below, and save it locally:

    * Replace `<DATABASE_NAME>` with a dedicated database name for NVIDIA Run:ai in your PostgreSQL database.
    * Replace `<ROLE_NAME>` with a dedicated role name (user) for the NVIDIA Run:ai database.
    * Replace `<ROLE_PASSWORD>` with a password for the new PostgreSQL role.
    * Replace `<GRAFANA_PASSWORD>` with the password to be set for Grafana integration.

    ```sql
    -- Create the role with login and password
    CREATE ROLE <ROLE_NAME> WITH LOGIN PASSWORD '<ROLE_PASSWORD>';

    -- Create a new database for runai, owned by the role
    CREATE DATABASE <DATABASE_NAME> OWNER <ROLE_NAME>;

    -- Connect to the newly created database
    \c <DATABASE_NAME>

    -- grafana
    CREATE ROLE grafana WITH LOGIN PASSWORD '<GRAFANA_PASSWORD>';
    CREATE SCHEMA grafana authorization grafana;
    ALTER USER grafana set search_path='grafana';
    -- Exit psql
    \q
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
| `postgresql.enabled` | PostgreSQL installation | Set to `false` so that PostgreSQL is not installed. |
| `global.postgresql.auth.host` | PostgreSQL host | Hostname or IP address of the PostgreSQL server. |
| `global.postgresql.auth.port` | PostgreSQL port | Port number on which PostgreSQL is running. |
| `global.postgresql.auth.database` | PostgreSQL database name | Name of the database to use. |
| `global.postgresql.auth.username` | PostgreSQL username | Username for connecting to PostgreSQL. |
| `global.postgresql.auth.password` | PostgreSQL password | Password for the PostgreSQL user. |
| `global.postgresql.auth.postgresPassword` | PostgreSQL admin password | Required, even though it is not used with an external instance. Set any value. |
| `global.postgresql.auth.existingSecret` | Credentials secret | Existing secret name with authentication credentials. Alternative to `global.postgresql.auth.password`. The keys are `password` and `postgres-password`. |
| `global.postgresql.auth.dbSslMode` | SSL mode | Set the SSL mode. See [Protection Provided in Different Modes](https://www.postgresql.org/docs/current/libpq-ssl.html#LIBPQ-SSL-PROTECTION). `Prefer` mode is not supported. |

#### PostgreSQL Wait Image

NVIDIA Run:ai services use a PostgreSQL client image in init containers to wait until the database is available. This image is required even when you provide your own PostgreSQL instance. Replace it with the hardened image:

```yaml
global:
  postgresImage:
    registry: <registry>
    repository: <dhi-postgres-repository>
    tag: <fips-ready-tag>                # tested with 16.15
```

### Traefik

Traefik serves as the API gateway for the NVIDIA Run:ai control plane, handling token verification and micro-service routing. It is deployed as part of the control plane Helm chart and is not supported as an external service.

Replace the Traefik image:

```yaml
traefik:
  image:
    registry: <registry>
    repository: <dhi-traefik-repository>
    tag: <fips-ready-tag>                # tested with 3.7.11
```

### NATS

NATS provides caching and asynchronous communication for the NVIDIA Run:ai control plane. It is deployed using [Bitnami NATS](https://artifacthub.io/packages/helm/bitnami/nats) (`nats`).

Replace the NATS image and override the command. The hardened image places the binary at `/nats-server`, so the default `nats-server` command cannot start it.

```yaml
nats:
  image:
    registry: <registry>
    repository: <dhi-nats-repository>
    tag: <fips-ready-tag>
  command:
    - /nats-server
```

#### Using an External NATS Instance

To replace the deployed NATS with an external instance, set `nats.enabled: false` and provide the connection URLs and credentials for both cache and pub/sub operations.

| Key | Change | Description |
| --- | ------ | ----------- |
| `nats.enabled` | NATS installation | If set to `false`, NATS will not be installed. Use this when replacing with an external instance. |
| `global.cache.url` | Cache URL | URL of the NATS instance for cache operations. Required when using an external instance. |
| `global.cache.readUrl` | Cache read URL | Read-only URL of the NATS instance. |
| `global.cache.auth.existingSecret` | Cache credentials | Existing secret name with cache authentication credentials. |
| `global.pubsub.url` | Pub/sub URL | URL of the NATS instance for async communication. Required when using an external instance. |
| `global.pubsub.auth.existingSecret` | Pub/sub credentials | Existing secret name with pub/sub authentication credentials. |

### Thanos

Thanos provides single-tenant metrics storage for the NVIDIA Run:ai control plane. It is deployed using [Bitnami Thanos](https://artifacthub.io/packages/helm/bitnami/thanos) (`thanos`). For built-in Thanos storage tuning, refer to [Advanced Control Plane Configurations](../configuration/control-plane-config.md#thanos).

Replace the Thanos image:

```yaml
thanos:
  image:
    registry: <registry>
    repository: <dhi-thanos-repository>
    tag: <fips-ready-tag>                # tested with 0.42
```

#### Using an External Thanos Instance

To replace the deployed Thanos with an external metrics store, set `thanos.enabled: false` and provide the read and write URLs for the external instance.

| Key | Change | Description |
| --- | ------ | ----------- |
| `thanos.enabled` | Thanos installation | If set to `false`, Thanos will not be installed. Use this when replacing with an external instance. |
| `metricsService.config.datasourceProvider.staticDatasource.read.url` | Metrics read URL | URL for reading metrics from the Thanos instance. Required when using an external instance. |
| `metricsService.config.datasourceProvider.staticDatasource.write.url` | Metrics write URL | URL for writing metrics to the Thanos instance. Required when using an external instance. |
| `metricsService.config.datasourceProvider.staticDatasource.existingSecret` | Credentials secret | Existing secret name with Thanos authentication credentials. |

## Cluster Services

NVIDIA Run:ai cluster services support the workloads running on a NVIDIA Run:ai cluster, rather than the platform itself. Each is deployed automatically as part of the cluster Helm chart and does not support an external instance. Set the following values in the cluster values, as described in [Where to Set the Values](#where-to-set-the-values).

| Component | Supports External Instance | Description | Gov-Ready Change |
| --------- | -------------------------- | ----------- | ---------------- |
| Valkey | No | In-memory cache for workload state and session data | Replace the image |
| git-sync | No | Syncs Git repositories as data sources into running workloads | Replace the image |

### Valkey

```yaml
thirdPartyImages:
  valkey:
    registry: <registry>
    repository: <dhi-valkey-repository>
    tag: <fips-ready-tag>                # tested with 9.1.1-fips
```

### git-sync

```yaml
thirdPartyImages:
  git-sync:
    registry: <registry>
    repository: <dhi-git-sync-repository>
    tag: <fips-ready-tag>                # tested with 4.7.1-fips
```

## Next Steps

Continue with [Gov-Ready Installation](../installation/gov-ready-installation.md).
