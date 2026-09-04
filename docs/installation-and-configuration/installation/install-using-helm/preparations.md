# Preparations

These steps prepare your environment for installation and include retrieving the required software artifacts, configuring access to container images, and preparing installation-specific configuration files where applicable (for example, in air-gapped environments).

## Software Artifacts

This section describes how to prepare the software artifacts required for installing the NVIDIA Run:ai [control plane](install-control-plane.md) and [cluster](helm-install.md). The steps depend on:

* Platform distribution - Kubernetes or OpenShift
* On-premise environment type - connected or air-gapped (i.e. not connected)

### Kubernetes

<Accordion title="Connected">

**Connected**

In connected environments, Kubernetes pulls NVIDIA Run:ai container images directly from a remote registry at runtime. To enable this, you must create a Kubernetes secret for registry access.

NVIDIA Run:ai container images are hosted in the NVIDIA NGC container registry under `nvidia/runai`. You must have access to NVIDIA Run:ai artifacts in NGC in order to pull NVIDIA Run:ai container images. Create the required secret using your NGC API key. For information on creating an NGC API key and authenticating to NGC, see [NGC API keys](https://docs.nvidia.com/ngc/latest/ngc-catalog-user-guide.html#ngc-api-keys).

```bash
kubectl create secret docker-registry runai-reg-creds \
--docker-server=https://nvcr.io \
--docker-username='$oauthtoken' \
--docker-password=<NGC_API_KEY> \
--docker-email=<EMAIL> \
--namespace=runai-backend
```

</Accordion>

<Accordion title="Air-gapped">

**Air-gapped**

In air-gapped environments, Kubernetes cannot pull container images from external registries during runtime. Instead, NVIDIA Run:ai provides an air-gapped installation package that contains all required images. These images are uploaded to an internal registry that your cluster can access.

**Download and Extract the Air-gapped Package**

Use your NGC API key to download the NVIDIA Run:ai air-gapped installation package from NVIDIA NGC. For information on creating an NGC API key and authenticating to NGC, see [NGC API keys](https://docs.nvidia.com/ngc/latest/ngc-catalog-user-guide.html#ngc-api-keys).

1. Browse the available package versions. Air-gapped packages are published as an NGC Resource. To view available versions, refer to the [NGC Resource](https://catalog.ngc.nvidia.com/orgs/nvidia/teams/runai/resources/runai-airgapp-package/version?) page and select the required version.
2.  Download the package using your NGC API key (authenticated download). NGC supports downloading resources via `curl`/`wget` using an `Authorization: Bearer` header. The following example uses `curl`. To download the package using `wget`, refer to [Downloading Authenticated Access Resources via WGET/cURL](https://docs.nvidia.com/ngc/latest/ngc-catalog-user-guide.html#downloading-authenticated-access-resources-via-wget-curl).

    *   Before running the `curl` command, set the following variables:

        ```shellscript
        NGC_CLI_API_KEY="nvapi-REDACTED"
        RUNAI_VERSION="<RUNAI_VERSION>"
        ```
    *   Run the below command:

        ```shellscript
        curl -LO --request GET \
          "https://api.ngc.nvidia.com/v2/org/nvidia/team/runai/resources/runai-airgapp-package/versions/$RUNAI_VERSION/files/runai-airgapped-package-$RUNAI_VERSION.tar.gz" \
          -H "Authorization: Bearer ${NGC_CLI_API_KEY}" \
          -H "Content-Type: application/json"
        ```

    For example, run the below to get the 2.26 air-gapped package:

    ```bash
    curl -LO --request GET \
      "https://api.ngc.nvidia.com/v2/org/nvidia/team/runai/resources/runai-airgapp-package/versions/2.26.14/files/runai-airgapped-package-2.26.14.tar.gz" \
      -H "Authorization: Bearer ${NGC_CLI_API_KEY}" \
      -H "Content-Type: application/json"
    ```
3. SSH into a node with `kubectl` access to the cluster and Docker installed.
4.  Extract the NVIDIA Run:ai package, replace `<VERSION>` in the command below and run:

    ```bash
    tar xvf runai-airgapp-package-<VERSION>.tar.gz
    ```

**Upload Images**

NVIDIA Run:ai assumes the existence of a Docker registry within your organization for hosting container images. The installation requires the network address and port for this registry (referred to as `<REGISTRY_URL>`).

1.  Upload images to a local Docker registry. Set the Docker registry address in the form of `NAME:PORT` (do not add `https`):

    ```bash
    export REGISTRY_URL=<DOCKER REGISTRY ADDRESS>
    ```
2.  Run the following script. You must have at least 20GB of free disk space to run. If Docker is configured to [run as non-root](https://docs.docker.com/engine/install/linux-postinstall/#manage-docker-as-a-non-root-user) then `sudo` is not required:

    ```bash
    sudo ./setup.sh
    ```

The script should create a file named `custom-env.yaml` which will be used during [control plane installation](install-control-plane.md#installation).

</Accordion>

### OpenShift

<Accordion title="Connected">

**Connected**

In connected environments, OpenShift pulls NVIDIA Run:ai container images directly from a remote registry at runtime. To enable this, you must create a Kubernetes secret for registry access.

NVIDIA Run:ai container images are hosted in the NVIDIA NGC container registry under `nvidia/runai`. You must have access to NVIDIA Run:ai artifacts in NGC in order to pull NVIDIA Run:ai container images. Create the required secret using your NGC API key. For information on creating an NGC API key and authenticating to NGC, see [NGC API keys](https://docs.nvidia.com/ngc/latest/ngc-catalog-user-guide.html#ngc-api-keys).

```bash
oc create secret docker-registry runai-reg-creds \
--docker-server=https://nvcr.io \
--docker-username='$oauthtoken' \
--docker-password=<NGC_API_KEY> \
--docker-email=<EMAIL> \
--namespace=runai-backend
```

</Accordion>

<Accordion title="Air-gapped">

**Air-gapped**

In air-gapped environments, OpenShift cannot pull container images from external registries during runtime. Instead, NVIDIA Run:ai provides an air-gapped installation package that contains all required images. These images are uploaded to an internal registry that your cluster can access.

**Download and Extract the Air-gapped Package**

Use your NGC API key to download the NVIDIA Run:ai air-gapped installation package from NVIDIA NGC. For information on creating an NGC API key and authenticating to NGC, see [NGC API keys](https://docs.nvidia.com/ngc/latest/ngc-catalog-user-guide.html#ngc-api-keys).

1. Browse the available package versions. Air-gapped packages are published as an NGC Resource. To view available versions, refer to the [NGC Resource](https://catalog.ngc.nvidia.com/orgs/nvidia/teams/runai/resources/runai-airgapp-package/version?) page and select the required version.
2.  Download the package using your NGC API key (authenticated download). NGC supports downloading resources via `curl`/`wget` using an `Authorization: Bearer` header. The following example uses `curl`. To download the package using `wget`, refer to [Downloading Authenticated Access Resources via WGET/cURL](https://docs.nvidia.com/ngc/latest/ngc-catalog-user-guide.html#downloading-authenticated-access-resources-via-wget-curl).

    *   Before running the `curl` command, set the following variables:

        ```shellscript
        NGC_CLI_API_KEY="nvapi-REDACTED"
        RUNAI_VERSION="<RUNAI_VERSION>"
        ```
    *   Run the below command:

        ```shellscript
        curl -LO --request GET \
          "https://api.ngc.nvidia.com/v2/org/nvidia/team/runai/resources/runai-airgapp-package/versions/$RUNAI_VERSION/files/runai-airgapped-package-$RUNAI_VERSION.tar.gz" \
          -H "Authorization: Bearer ${NGC_CLI_API_KEY}" \
          -H "Content-Type: application/json"
        ```

    For example, run the below to get the 2.26 air-gapped package:

    ```bash
    curl -LO --request GET \
      "https://api.ngc.nvidia.com/v2/org/nvidia/team/runai/resources/runai-airgapp-package/versions/2.26.14/files/runai-airgapped-package-2.26.14.tar.gz" \
      -H "Authorization: Bearer ${NGC_CLI_API_KEY}" \
      -H "Content-Type: application/json"
    ```
3. SSH into a node with `oc` access to the cluster and Docker installed.
4.  Extract the NVIDIA Run:ai package, replace `<VERSION>` in the command below and run:

    ```bash
    tar xvf runai-airgapp-package-<VERSION>.tar.gz
    ```

**Upload Images**

NVIDIA Run:ai assumes the existence of a Docker registry within your organization for hosting container images. The installation requires the network address and port for this registry (referred to as `<REGISTRY_URL>`).

1.  Upload images to a local Docker registry. Set the Docker registry address in the form of `NAME:PORT` (do not add `https`):

    ```bash
    export REGISTRY_URL=<DOCKER REGISTRY ADDRESS>
    ```
2.  Run the following script. You must have at least 20GB of free disk space to run. If Docker is configured to [run as non-root](https://docs.docker.com/engine/install/linux-postinstall/#manage-docker-as-a-non-root-user) then `sudo` is not required:

    ```bash
    sudo ./setup.sh
    ```

The script should create a file named `custom-env.yaml` which will be used during [control plane installation](install-control-plane.md#installation).

</Accordion>

## Private Docker Registry

This step is required only if you are installing NVIDIA Run:ai in an environment that uses a private Docker registry.

### Kubernetes

To access the organization's docker registry, set the registry's credentials (imagePullSecret).

Create the secret named `runai-reg-creds` based on your existing credentials. For more information, see [Pull an Image from a Private Registry](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/).

### OpenShift

To access the organization's docker registry, set the registry's credentials (imagePullSecret).

Create the secret named `runai-reg-creds` in the `runai-backend` namespace based on your existing credentials. The configuration will be copied over to the `runai` namespace at [cluster install](helm-install.md). For more information, see [Allowing pods to reference images from other secured registries](https://docs.openshift.com/container-platform/latest/openshift_images/managing_images/using-image-pull-secrets.html#images-allow-pods-to-reference-images-from-secure-registries_using-image-pull-secrets).

## Set Up Your Environment

### External Postgres Database

If you have opted to use an [external PostgreSQL database](cp-system-requirements.md#external-postgres-database-optional), you need to perform initial setup to ensure successful installation. Follow these steps:

1.  Create an SQL script file, edit the parameters below, and save it locally:

    * Replace `<DATABASE_NAME>` with a dedicated database name for NVIDIA Run:ai in your PostgreSQL database.
    * Replace `<ROLE_NAME>` with a dedicated role name (user) for the NVIDIA Run:ai database.
    * Replace `<ROLE_PASSWORD>` with a password for the new PostgreSQL role.
    * Replace `<GRAFANA_PASSWORD>` with the password to be set for Grafana integration.

    ```sql
    -- Create a new database for runai
    CREATE DATABASE <DATABASE_NAME>; 

    -- Create the role with login and password
    CREATE ROLE <ROLE_NAME>  WITH LOGIN PASSWORD '<ROLE_PASSWORD>'; 

    -- Grant all privileges on the database to the role
    GRANT ALL PRIVILEGES ON DATABASE <DATABASE_NAME> TO <ROLE_NAME>; 

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
