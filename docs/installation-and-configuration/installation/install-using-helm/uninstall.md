# Uninstall

## Uninstall the Control Plane

To delete the control plane, run:

```bash
helm uninstall runai-backend -n runai-backend
```

## Uninstall the Cluster

To uninstall the NVIDIA Run:ai cluster, run the following [helm](https://helm.sh/) command in your terminal:

```bash
helm uninstall runai-cluster -n runai
```

To remove the NVIDIA Run:ai cluster from the NVIDIA Run:ai platform, see [Removing a cluster](../../operations/clusters.md#removing-a-cluster).

<Info>

**Note**

Uninstall of NVIDIA Run:ai cluster from the Kubernetes cluster does **not** delete existing projects, departments or workloads submitted by users.

</Info>
