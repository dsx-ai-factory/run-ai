# NVIDIA Run:ai CRD Reference

The cluster plane is CRD-driven. The `agent` reconciles control-plane state into cluster CRDs; `cluster-sync` watches those CRDs and reports status back. All CRDs listed here use the `run.ai` API group.

For an overview of how CRDs fit into the platform architecture, see the [Architecture and Components](architecture.md). For the services that manage these CRDs, see the [Service Reference](services.md).

## Workload Types

| Kind | API group | Description |
|------|-----------|-------------|
| [ExternalWorkload](https://github.com/dsx-ai-factory/run-ai/blob/main/runai-cluster/cluster/runai-chart/charts/crds/crds/run_ai_externalworkload.yaml) | `run.ai/v1alpha1` | A workload originating outside NVIDIA Run:ai (e.g. a plain Kubernetes Deployment) that the platform tracks for quota and visibility. |
| [WorkloadRunner](https://github.com/dsx-ai-factory/run-ai/blob/main/runai-cluster/cluster/runai-chart/charts/crds/crds/run.ai_workloadrunner.yaml) | `run.ai/v1` | Internal lower-level resource that materializes a workload's final pod spec; produced by workload controllers and consumed by the KAI Scheduler. |

## Cluster Configuration and Access

| Kind | API group | Description |
|------|-----------|-------------|
| [RunaiConfig](https://github.com/dsx-ai-factory/run-ai/blob/main/runai-cluster/cluster/runai-chart/charts/crds/crds/run_ai_v1_runaiconfig.yaml) | `run.ai/v1` | Top-level operator configuration object; controls settings for every platform component in the cluster. |
| [Cluster](https://github.com/dsx-ai-factory/run-ai/blob/main/runai-cluster/cluster/runai-chart/crds/run.ai_clusters.yaml) | `run.ai/v1` | Singleton that registers the cluster with the control plane and reports observed health conditions. |
| [AccessRule](https://github.com/dsx-ai-factory/run-ai/blob/main/runai-cluster/cluster/runai-chart/charts/crds/crds/run_ai_accessrule.yaml) | `run.ai/v1alpha1` | Binds a subject (user or group) to a named role within a scope (cluster, department, or project). |
| [DataVolume](https://github.com/dsx-ai-factory/run-ai/blob/main/runai-cluster/cluster/runai-chart/charts/crds/crds/run.ai_datavolumes.yaml) | `run.ai/v1alpha1` | Cluster-scoped reference to a source PVC that can be shared across multiple project namespaces as a reusable storage asset. |
| [Application](https://github.com/dsx-ai-factory/run-ai/blob/main/runai-cluster/cluster/runai-chart/charts/crds/crds/run_ai_application.yaml) | `run.ai/v1alpha1` | An OAuth2 client application registered with the cluster's identity manager. |

## Resource Management

CRDs in the `kai.resources` API group, reconciled today by `nodepool-controller` and `project-controller` (see [Service Reference](services.md)). `NodePool` and `ManagedNodesConfig` also still have a legacy `run.ai`-group version; migrating those onto `kai.resources` runs automatically and self-gates once complete.

| Kind | API group | Description |
|------|-----------|-------------|
| [NodePool](https://github.com/dsx-ai-factory/run-ai/blob/main/runai-cluster/cluster/runai-chart/charts/crds/crds/kai.resources_nodepools.yaml) | `kai.resources/v1alpha1` (legacy: `run.ai/v1alpha1`) | A partition of cluster nodes, tagged by `nodepool-controller`, that a workload can be scheduled onto; each node pool runs its own KAI Scheduler instance. |
| [Project](https://github.com/dsx-ai-factory/run-ai/blob/main/runai-cluster/cluster/runai-chart/charts/crds/crds/kai.resources_projects.yaml) | `kai.resources/v1alpha1` | Reconciled by `project-controller` into a namespace, queue, limit range, and image pull secret. |
| [Department](https://github.com/dsx-ai-factory/run-ai/blob/main/runai-cluster/cluster/runai-chart/charts/crds/crds/kai.resources_departments.yaml) | `kai.resources/v1alpha1` | The parent grouping of Projects, with its own quota and queues. |
| [ManagedNodesConfig](https://github.com/dsx-ai-factory/run-ai/blob/main/runai-cluster/cluster/runai-chart/charts/crds/crds/kai.resources_managednodesconfigs.yaml) | `kai.resources/v1alpha1` (legacy: `run.ai/v1alpha1`) | Cluster-wide selection of which nodes are under NVIDIA Run:ai management. |
