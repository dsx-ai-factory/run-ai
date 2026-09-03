---
---

# Operations

Day 2 Operations covers the ongoing operational tasks that keep NVIDIA Run:ai healthy and your clusters running smoothly.

## Cluster Management

* **[Clusters](clusters.md)** - View and manage your clusters, including status, from the NVIDIA Run:ai UI.
* **[Scaling](scaling.md)** - Best practices for tuning NVIDIA Run:ai system services as clusters grow in nodes or workload demand, including vertical and horizontal scaling.
* **[High Availability](high-availability.md)** - Reduce downtime and eliminate single points of failure using dedicated system nodes and replicated core and third-party services.
* **[Vertical Pod Autoscaling (VPA)](vpa.md)** - Automatically right-sizes CPU and memory requests for cluster-level services based on observed usage, reducing the need for manual tuning.

## Maintenance

* **[Node Maintenance](maintenance/nodes-maintenance.md)** - Manage planned and unplanned node downtime while maintaining service continuity for running workloads.
* **[Backup and Restore](maintenance/cluster-restore.md)** - Back up and restore an NVIDIA Run:ai deployment, covering both the cluster and control plane.

## Observability

* **[System Monitoring](observability/system-monitoring.md)** - Configure health alerts for NVIDIA Run:ai clusters and connect them to your organization's alert-management systems.
* **[Logs Collection](observability/logs-collection.md)** - Collect NVIDIA Run:ai logs for support, including enabling verbose logging for Prometheus and the Scheduler.
* **[Event History](observability/event-history.md)** - Audit log API and UI event history tracking changes to business objects like clusters, projects, and assets.
