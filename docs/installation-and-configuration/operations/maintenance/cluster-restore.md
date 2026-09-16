# Backup and Restore

This guide outlines how to back up and restore an NVIDIA Run:ai deployment, including both the NVIDIA Run:ai cluster and control plane.

## Back Up the Cluster

The restoration or backup of NVIDIA Run:ai [advanced cluster configurations](../../configuration/cluster-config.md) which are stored locally on the Kubernetes cluster is optional and can be restored and backed up separately. As backup of data is not required, the backup procedure is optional for advanced deployments.

### Save Cluster Configurations

To back up the NVIDIA Run:ai cluster configurations, you should save both the Helm values and the runtime configuration (`runaiconfig`).

1.  **Back up Helm values** - Run the following command to export the Helm values used for deployment:

    ```bash
    helm get values runai-cluster -n runai > runai_cluster_values_backup.yaml
    ```
2.  **Back up the runtime configuration (`runaiconfig`)** - Run the following command to export the active runtime configuration:

    ```bash
    kubectl get runaiconfig runai -n runai -o yaml -o=jsonpath='{.spec}' > runaiconfig_backup.yaml
    ```
3. Save both backup files (`runai_cluster_values_backup.yaml` and `runaiconfig_backup.yaml`) externally so they can be retrieved later if needed.

## Restore the Cluster

In the event of a critical Kubernetes failure or alternatively, if you want to migrate an NVIDIA Run:ai cluster to a new Kubernetes environment, simply reinstall the NVIDIA Run:ai cluster. Once you have reinstalled and reconnected the cluster, projects, workloads and other cluster data are synced automatically. Follow the steps below to restore the NVIDIA Run:ai cluster on a new Kubernetes environment.

### Prerequisites

Before restoring the NVIDIA Run:ai cluster, it is essential to validate that it is both disconnected and uninstalled:

1. If the Kubernetes cluster is still available, [uninstall](../../installation/install-using-helm/uninstall.md) the NVIDIA Run:ai cluster. Make sure not to remove the cluster from the control plane.
2. Navigate to the **Clusters** grid in the NVIDIA Run:ai UI
3. Locate the cluster and verify its status is **Disconnected**

### Re-install the Cluster

1. Follow the NVIDIA Run:ai cluster [installation](../../installation/install-using-helm/helm-install.md) instructions and ensure all [prerequisites](../../installation/install-using-helm/prerequisites-and-features.md) are met.
2.  If you have a backup of the cluster configurations, reload it once the installation is complete:

    ```bash
    kubectl apply -f runaiconfig_backup.yaml -n runai
    ```
3. Navigate to the **Clusters** grid in the NVIDIA Run:ai UI
4. Locate the cluster and verify its status is **Connected**

### Restore Namespace and RoleBindings

If your cluster configuration disables automatic namespace creation for projects, you must manually:

* Re-create each project namespace
* Reapply the required role bindings for access control

For more information, see [Advanced cluster configurations](../../configuration/cluster-config.md).

## Back Up the Control Plane <a href="#database-storage" id="database-storage"></a>

### Database Storage <a href="#database-storage" id="database-storage"></a>

By default, NVIDIA Run:ai utilizes an internal PostgreSQL database to manage control plane data. This database resides on a Kubernetes Persistent Volume (PV). To safeguard against data loss, it's essential to implement a reliable backup strategy.

#### Backup Methods

Consider the following methods to back up the PostgreSQL database:

*   **PostgreSQL logical backup** - Use `pg_dump` to create a logical backup of the database. Replace `<password>` with the appropriate PostgreSQL password. For example:

    ```bash
    kubectl -n runai-backend exec -it runai-backend-postgresql-0 -- \
        env PGPASSWORD=<password> pg_dump -U postgres backend > cluster_name_db_backup.sql
    ```
* **Persistent volume backup** - Back up the entire PV that stores the PostgreSQL data.
* **Third-Party backup solutions** - Integrate with external backup tools that support Kubernetes and PostgreSQL to automate and manage backups effectively.

<Info>

**Note**

* To obtain your `PGPASSWORD=<password>`, run `helm get values runai-backend -n runai-backend --all`.
* NVIDIA Run:ai also supports an external PostgreSQL database. If you are using an PostgreSQL database, the above steps do not apply. For more details, see [Using an External PostgreSQL Instance](../../system-requirements/third-party-services.md#using-an-external-postgresql-instance).

</Info>

### Metrics Storage <a href="#metrics-storage" id="metrics-storage"></a>

NVIDIA Run:ai stores metrics history using [Thanos](https://github.com/thanos-io/thanos). Thanos is configured to write data to a persistent volume (PV). To protect against data loss, it is recommended to regularly back up this volume.

### Deployment Configurations <a href="#backing-up-control-plane-configuration" id="backing-up-control-plane-configuration"></a>

The NVIDIA Run:ai control plane installation can be [customized](../../configuration/control-plane-config.md) using `--set` flags during Helm deployment. These configuration overrides are preserved during upgrades but are **not retained** if Kubernetes is uninstalled or damaged. To ensure recovery, it's recommended to back up the full set of applied Helm customizations.

1.  To back up the control plane Helm values:

    ```bash
    helm get values runai-backend -n runai-backend  > runai_controlplane_values_backup.yaml
    ```
2.  You can retrieve the current configuration using:

    ```bash
    helm get values runai-backend -n runai-backend
    ```

## Restore the Control Plane

Follow the steps below to restore the control plane including previously backed-up data and configurations:

1. Recreate the Kubernetes environment - Begin by provisioning a new Kubernetes or OpenShift cluster that meets all NVIDIA Run:ai installation [requirements](../../system-requirements/README.md).
2. Restore Persistent Volumes - Recover the PVs and ensure these volumes are correctly reattached or restored from your backup solution:
   * PostgreSQL database - Stores control plane metadata
   * Thanos - Stores workload metrics and historical data
3. Reinstall the control plane - Install the NVIDIA Run:ai [control plane](../../installation/install-using-helm/install-control-plane.md) on the newly created cluster. During installation:
   * Use the saved Helm configuration overrides to preserve custom settings
   * Connect the control plane to the recovered PostgreSQL volume
   * Reconnect Thanos to the restored metrics volume

<Info>

**Note**

For external PostgreSQL databases, ensure the appropriate connection details and credentials are reconfigured. See [Using an External PostgreSQL Instance](../../system-requirements/third-party-services.md#using-an-external-postgresql-instance) for more details.

</Info>
