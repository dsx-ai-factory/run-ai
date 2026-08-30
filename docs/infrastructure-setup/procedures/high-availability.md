# High Availability

This guide outlines the best practices for configuring the NVIDIA Run:ai platform to ensure high availability and maintain service continuity during system failures or under heavy load. The goal is to reduce downtime and eliminate single points of failure by leveraging Kubernetes best practices alongside NVIDIA Run:ai specific configuration options. The NVIDIA Run:ai platform relies on two fundamental high availability strategies:

* **Use of system nodes** - Assigning multiple dedicated nodes for critical system services ensures control, resource isolation, and enables system-level scaling.
* **Replication of core and third-party services** - Configuring multiple replicas of essential services, including both platform and third-party components, distributes workloads and reduces single points of failure. If a component fails on one node, requests can seamlessly route to another instance.

## System Nodes

The NVIDIA Run:ai platform allows you to dedicate specific nodes (system nodes) exclusively for core platform services. This approach provides improved operational isolation and easier resource management.

Ensure that at least **three system nodes** are configured to support high availability. If you use only a single node for core services, horizontally scaled components will not be distributed, resulting in a single point of failure. See [NVIDIA Run:ai system nodes](../../installation-and-configuration/post-installation-configuration/node-roles.md) for more details. This practice applies to both the NVIDIA Run:ai cluster and control plane.

## Service Replicas <a href="#undefined" id="undefined"></a>

### Control Plane Service Replicas

The NVIDIA Run:ai control plane runs in the `runai-backend` namespace and consists of multiple Kubernetes [Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) and [StatefulSets](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/). To achieve high availability, it is recommended to configure multiple replicas during installation or upgrade using Helm flags.

In addition, the control plane supports autoscaling for certain services to handle variable load and improve system resiliency. Autoscaling can be enabled or configured during installation or upgrade using Helm flags.

#### Deployments

By default, NVIDIA Run:ai deployments are deployed with a single replica. To achieve high availability, it is recommended to configure multiple replicas for core NVIDIA Run:ai services. For more information, see [Advanced control plane configuration](../../installation-and-configuration/post-installation-configuration/control-plane-config.md).

#### StatefulSets

NVIDIA Run:ai uses the following third-party components which are managed as Kubernetes StatefulSets. For more information, see [Advanced control plane configurations](../../installation-and-configuration/post-installation-configuration/control-plane-config.md):

* **PostgreSQL** - The internal PostgreSQL cannot be scaled horizontally. To connect NVIDIA Run:ai to an external PostgreSQL service which can be configured for high availability, see [External Postgres Database](../../installation-and-configuration/installation/install-using-helm/cp-system-requirements.md#external-postgres-database-optional).
*   **Thanos** - To enable Thanos autoscaling, use the following NVIDIA Run:ai control plane helm flags:

    ```bash
    --set thanos.query.autoscaling.enabled=true \  
    --set thanos.query.autoscaling.maxReplicas=2 \
    --set thanos.query.autoscaling.minReplicas=2 
    ```
*   **Keycloak** **-** By default, Keycloak sets a minimum of 3 pods and will scale to more on transaction load. To scale Keycloak, use the following NVIDIA Run:ai control plane helm flags:

    ```bash
    --set keycloakx.autoscaling.enabled=true
    ```

### Cluster Services Replicas <a href="#nvidia-run-ai-services-replicas" id="nvidia-run-ai-services-replicas"></a>

By default, NVIDIA Run:ai cluster services are deployed with a single replica. To achieve high availability, it is recommended to configure multiple replicas for core NVIDIA Run:ai services. For more information, see [NVIDIA Run:ai services replicas](../../installation-and-configuration/post-installation-configuration/cluster-config.md#nvidia-run-ai-services-replicas).
