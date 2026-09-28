---
hidden: true
noIndex: true
---

# Cluster Restore

This section explains how to restore a NVIDIA Run:ai cluster on a different Kubernetes environment.

In the event of a critical Kubernetes failure or alternatively, if you want to migrate a NVIDIA Run:ai cluster to a new Kubernetes environment, simply reinstall the NVIDIA Run:ai cluster. Once you have reinstalled and reconnected the cluster - projects, workloads and other cluster data is synced automatically.

The restoration or back-up of NVIDIA Run:ai cluster [Advanced features](../../configuration/cluster-config.md) which are stored locally on the Kubernetes cluster is optional and they can be restored and backed-up separately.

## Backup

As back-up of data is not required, the backup procedure is optional for advanced deployments, as explained above.

### Backup Cluster Configurations

To backup NVIDIA Run:ai cluster configurations:

1.  Run the following command in your terminal:

    ```bash
    kubectl get runaiconfig runai -n runai -o yaml -o=jsonpath='{.spec}' > runaiconfig_backup.yaml
    ```
2. Once the `runaiconfig_back.yaml` back-up file is created, save the file externally, so that it can be retrieved later.

## Restore

Follow the steps below to restore the NVIDIA Run:ai cluster on a new Kubernetes environment.

### Prerequisites

Before restoring the NVIDIA Run:ai cluster, it is essential to validate that it is both disconnected and uninstalled.

1. If the Kubernetes cluster is still available, [uninstall](../../installation/install-using-helm/uninstall.md) the NVIDIA Run:ai cluster - make sure not to remove the cluster from the Control Plane
2. Navigate to the Cluster page in the NVIDIA Run:ai platform
3. Search for the cluster, and make sure its status is **Disconnected**

### Re-installing NVIDIA Run:ai Cluster

1. Follow the NVIDIA Run:ai cluster installation instructions and ensure all prerequisites are met
2.  If you have a back-up of the cluster configurations, reload it once the installation is complete

    ```bash
    kubectl apply -f runaiconfig_backup.yaml -n runai
    ```
3. Navigate to the Cluster page in the NVIDIA Run:ai platform
4. Search for the cluster, and make sure its status is **Connected**
