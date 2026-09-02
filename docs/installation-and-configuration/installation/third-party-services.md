# Third-Party Services

NVIDIA Run:ai bundles a set of third-party services as part of its installation. This page lists those services, where they are deployed, whether they can be replaced with an externally managed instance, and how to configure each one.

## Image Registry

For Open Source and Enterprise installations, all third-party service images are pulled from the official upstream registries on Docker Hub (`docker.io`). No registry credentials are required for these images.

For Gov-ready installations, customers must obtain hardened, FIPS-enabled images from [Chainguard](https://www.chainguard.dev/) for each supported third-party component. Chainguard images must match the supported version of each component for the NVIDIA Run:ai release in use.

For air-gapped installations, NVIDIA Run:ai provides a script to download and mirror the images. See [Air-Gap Installation](install-using-helm/preparations.md).

## Services

| Component | Deployed On | Supports External Instance | Description |
| --------- | ----------- | -------------------------- | ----------- |
| Keycloak | NVIDIA Run:ai control plane | No | Authentication and identity management |
| PostgreSQL | NVIDIA Run:ai control plane | Yes | Centralized database |
| Traefik | NVIDIA Run:ai control plane | No | API gateway — token verification and micro-service routing |
| NATS | NVIDIA Run:ai control plane | Yes | Caching and async communication |
| Thanos | NVIDIA Run:ai control plane | Yes | Single-tenant metrics |
| Valkey | NVIDIA Run:ai cluster | Planned | Workload cache |
| oauth2-proxy | NVIDIA Run:ai cluster | No | Workload authentication and authorization |

## Assumptions

* Each NVIDIA Run:ai release is tested and validated against a single minor version of each third-party dependency. For example, NVIDIA Run:ai 2.26.x is validated with the Keycloak 26.4.x family, whether using the open-source community distribution or a commercial distribution (such as Red Hat's build of Keycloak) based on the same version line.

* NVIDIA does not modify the base images of third-party components. Any component can be replaced with a compatible image, provided it uses one of the supported versions. Customers can provide any distributor image as long as it is based on a supported version.

* All third-party images are based on official open-source upstream images pulled from Docker Hub (`docker.io`).

* NVIDIA Run:ai supports the following deployment types:
  * **Open Source** — Uses a pinned "latest supported" third-party version, pulled directly from `docker.io`.
  * **Enterprise** — Uses a pinned "latest supported" third-party version, pulled directly from `docker.io`.
  * **Air-gapped** — Uses a pinned "latest supported" third-party version. NVIDIA Run:ai provides a script to download and mirror the images.
  * **Gov-ready** — Customers must provide hardened, FIPS-enabled images from Chainguard (or equivalent) at the supported version for each component, where available.

## Keycloak

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

### Changing the Keycloak Admin Password

1. Open the Keycloak UI at: `https://<runai-domain>/auth`
2. Sign in with your existing admin credentials
3. Go to **Users** and select **admin** (or your admin username)
4. Open **Credentials** → **Reset password**
5. Set the new password and click **Save**
6. Run `helm upgrade` and set `keycloakx.adminPassword` to the new password

<Info>

**Note**

Failing to update the Helm values after changing the password can cause control plane services to encounter errors.

</Info>

## PostgreSQL

PostgreSQL is the centralized database for the NVIDIA Run:ai control plane. It is deployed as part of the control plane Helm chart using [Bitnami PostgreSQL](https://artifacthub.io/packages/helm/bitnami/postgresql) (`postgresql`). Only PostgreSQL version 16 is supported.

PostgreSQL supports an external instance. To use an external database, disable the bundled deployment and provide connection details as described below.

### Using an External PostgreSQL Instance

1. Disable the bundled PostgreSQL deployment: set `postgresql.enabled: false`
2. Configure the NVIDIA Run:ai connection: set `global.postgresql.auth.*`
3. Configure the Grafana connection: set `grafana.dbUser` and `grafana.dbPassword`

| Key | Change | Description |
| --- | ------ | ----------- |
| `postgresql.enabled` | PostgreSQL installation | If set to `false`, PostgreSQL will not be installed. |
| `global.postgresql.auth.host` | PostgreSQL host | Hostname or IP address of the PostgreSQL server. |
| `global.postgresql.auth.port` | PostgreSQL port | Port number on which PostgreSQL is running. |
| `global.postgresql.auth.username` | PostgreSQL username | Username for connecting to PostgreSQL. |
| `global.postgresql.auth.password` | PostgreSQL password | Password for the PostgreSQL user. |
| `global.postgresql.auth.postgresPassword` | PostgreSQL admin password | Password for the built-in PostgreSQL superuser (`postgres`). |
| `global.postgresql.auth.existingSecret` | Credentials secret | Existing secret name with authentication credentials. |
| `global.postgresql.auth.dbSslMode` | SSL mode | Set the SSL mode. See [Protection Provided in Different Modes](https://www.postgresql.org/docs/current/libpq-ssl.html#LIBPQ-SSL-PROTECTION). `Prefer` mode is not supported. |
| `postgresql.primary.initdb.password` | PostgreSQL admin password | Set the same password as `global.postgresql.auth.postgresPassword` if changed. |
| `postgresql.primary.persistence.storageClass` | Storage class | Use a specific storage class instead of the default. |

## Traefik

Traefik serves as the API gateway for the NVIDIA Run:ai control plane, handling token verification and micro-service routing. It is deployed as part of the control plane Helm chart and is not supported as an external service.

## NATS

NATS provides caching and asynchronous communication for the NVIDIA Run:ai control plane. It is deployed using [Bitnami NATS](https://artifacthub.io/packages/helm/bitnami/nats) (`nats`). NATS supports an external instance.

<Info>

**Note**

External NATS configuration options are coming soon.

</Info>

## Thanos

Thanos provides single-tenant metrics storage for the NVIDIA Run:ai control plane. It is deployed using [Bitnami Thanos](https://artifacthub.io/packages/helm/bitnami/thanos) (`thanos`). Thanos supports an external instance.

| Key | Change | Description |
| --- | ------ | ----------- |
| `thanos.receive.persistence.storageClass` | Storage class | Use a specific storage class instead of the default. |

<Info>

**Note**

External Thanos configuration options are coming soon.

</Info>

## Valkey

Valkey provides workload caching for the NVIDIA Run:ai cluster. It is deployed as part of the cluster Helm chart. Support for an external Valkey instance is planned but not yet available.

## oauth2-proxy

oauth2-proxy handles workload authentication and authorization for the NVIDIA Run:ai cluster. It is deployed as part of the cluster Helm chart and is not supported as an external service.
