# Back Up Before Upgrading

Before upgrading NVIDIA Run:ai, back up your deployment so you can restore it if the upgrade fails or needs to be rolled back. Back up the following configuration and data. If you need to restore any of these afterward, see [Backup and Restore](../operations/maintenance/cluster-restore.md).

## Cluster Configuration

To back up the NVIDIA Run:ai cluster configurations, you should save both the Helm values and the runtime configuration (`runaiconfig`).

* **Back up Helm values** - Run the following command to export the Helm values used for deployment:

  ```bash
  helm get values runai-cluster -n runai > runai_cluster_values_backup.yaml
  ```

* **Back up the runtime configuration (`runaiconfig`)** - Run the following command to export the active runtime configuration:

  ```bash
  kubectl get runaiconfig runai -n runai -o yaml -o=jsonpath='{.spec}' > runaiconfig_backup.yaml
  ```

Save both backup files (`runai_cluster_values_backup.yaml` and `runaiconfig_backup.yaml`) externally so they can be retrieved later if needed.

## Control Plane Configuration

The NVIDIA Run:ai control plane installation can be customized using `--set` flags during Helm deployment. These configuration overrides are preserved during upgrades but are not retained if Kubernetes is uninstalled or damaged. Back up the full set of applied Helm customizations:

```bash
helm get values runai-backend -n runai-backend > runai_controlplane_values_backup.yaml
```

## Third-Party Services

NVIDIA Run:ai v3.0 changes how third-party service images are sourced: instead of NVIDIA Run:ai bundling these images, you now pull them yourself from the upstream registry or your own mirror.

### Database Storage

By default, NVIDIA Run:ai utilizes an internal PostgreSQL database to manage control plane data. This database resides on a Kubernetes Persistent Volume (PV). To protect against data loss, back up this database before upgrading.

#### Backup Methods

Use one of the following methods to back up the PostgreSQL database before upgrading:

* **PostgreSQL logical backup** - Use `pg_dump` to create a logical backup of the database. Replace `<password>` with the appropriate PostgreSQL password. For example:

  ```bash
  kubectl -n runai-backend exec -it runai-backend-postgresql-0 -- \
      env PGPASSWORD=<password> pg_dump -U postgres backend > cluster_name_db_backup.sql
  ```

* **Persistent volume backup** - Back up the entire PV that stores the PostgreSQL data.

* **Third-party backup solutions** - Integrate with external backup tools that support Kubernetes and PostgreSQL to automate and manage backups effectively.

<Note>

**Note**

* To obtain your `PGPASSWORD=<password>`, run `helm get values runai-backend -n runai-backend --all`.
* NVIDIA Run:ai also supports an external PostgreSQL database. If you are using an external PostgreSQL database, the above steps do not apply. For more details, see [external PostgreSQL database](../system-requirements/third-party-services.md#using-an-external-postgresql-instance).

</Note>

### Metrics Storage

NVIDIA Run:ai stores metrics history using Thanos. Thanos is configured to write data to a persistent volume (PV). To protect against data loss, back up this volume before upgrading.
