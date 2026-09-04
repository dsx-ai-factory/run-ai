# Preparations

The following section provides the information needed to prepare for a NVIDIA Run:ai installation.

## Software Artifacts

NVIDIA Run:ai container images are hosted in the NVIDIA NGC container registry under `nvidia/runai`. You must have access to NVIDIA Run:ai artifacts in NGC in order to pull NVIDIA Run:ai container images. Create the required secret using your NGC API key. For information on creating an NGC API key and authenticating to NGC, see [NGC API keys](https://docs.nvidia.com/ngc/latest/ngc-catalog-user-guide.html#ngc-api-keys).

```bash
kubectl create secret docker-registry runai-reg-creds \
--docker-server=https://nvcr.io \
--docker-username='$oauthtoken' \
--docker-password=<NGC_API_KEY> \
--docker-email=<EMAIL> \
--namespace=runai-backend
```

## Private Docker Registry

This step is required only if you are installing NVIDIA Run:ai in an environment that uses a private Docker registry.

To access the organization's docker registry, set the registry's credentials (imagePullSecret).

Create the secret named `runai-reg-creds` based on your existing credentials. For more information, see [Pull an Image from a Private Registry](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/).

## Set Up Your Environment

### External Postgres Database (Optional)

If you have opted to use an [external PostgreSQL database](cp-system-requirements.md#external-postgres-database-optional), you need to perform initial setup to ensure successful installation. Follow these steps:

1.  Create a SQL script file, edit the parameters below, and save it locally:

    * Replace `<DATABASE_NAME>` with a dedicate database name for NVIDIA Run:ai in your PostgreSQL database.
    * Replace `<ROLE_NAME>` with a dedicated role name (user) for NVIDIA Run:ai database.
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

    * Replace `<POSTGRESQL_HOST>` with the PostgreSQL ip address or hostname.
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
