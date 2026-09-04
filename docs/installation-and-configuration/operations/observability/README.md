---
---

# Observability

Deploying NVIDIA Run:ai in mission-critical environments requires proper monitoring and maintenance of resources to ensure workloads run and are deployed as expected.

This section describes how to monitor NVIDIA Run:ai, including [collecting logs](logs-collection.md) and [reviewing audit logs](event-history.md) to meet compliance requirements.

In addition to monitoring NVIDIA Run:ai resources, it is also highly recommended to monitor NVIDIA Run:ai runs on Kubernetes, which manages containerized applications. In particular, focus on three main layers:

## NVIDIA Run:ai Control Plane and Cluster Services

This is the highest layer and includes the parts of NVIDIA Run:ai pods, which run in containers managed by Kubernetes.

## Kubernetes Cluster

This layer includes the main Kubernetes system that runs and manages NVIDIA Run:ai components. Important elements to monitor include:

* The health of the cluster and nodes (machines in the cluster).
* The status of key Kubernetes services, such as the API server. For detailed information on managing clusters, see the [official Kubernetes documentation](https://kubernetes.io/docs/tasks/debug/debug-cluster/resource-usage-monitoring/).

## Host Infrastructure

This is the base layer, representing the actual machines (virtual or physical) that make up the cluster IT teams need to handle:

* Managing CPU, memory, and storage
* Keeping the operating system updated
* Setting up the network and balancing the load

NVIDIA Run:ai does not require any special configurations at this level.

