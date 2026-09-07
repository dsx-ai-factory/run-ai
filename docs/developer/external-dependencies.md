# External Dependencies

The NVIDIA Run:ai cluster plane pulls in components from external repositories as Helm chart dependencies. These components are not built from the Run:ai source - they are consumed as versioned OCI charts and integrated at deploy time.

The chart dependencies are declared in [`runai-chart/Chart.yaml`](runai-cluster/cluster/runai-chart/Chart.yaml) and documented in the [Chart dependencies](runai-cluster/cluster/runai-chart/README.md#chart-dependencies) section of the runai-chart README.

## KAI Scheduler

**Source:** [`github.com/kai-scheduler/KAI-Scheduler`](https://github.com/kai-scheduler/KAI-Scheduler)
**Chart:** `oci://ghcr.io/kai-scheduler/kai-scheduler`

The KAI (Kubernetes AI) Scheduler is an open-source, CNCF-donated AI-aware scheduling engine. It applies fair-share, preemption, bin-packing, and multi-node placement rules to GPU workloads.

NVIDIA Run:ai integrates KAI with specific configuration overrides - the `go-operator` owns `kai-config` and the `kai-nodepool-controller` creates one KAI SchedulingShard per node pool. KAI's default queue, shard, and config-deployer components are disabled.

Once `kai-resource-management` is added as a chart dependency, KAI Scheduler will become a sub-dependency of KRM rather than a direct entry in `runai-chart/Chart.yaml`.

## Karta

**Source:** `oci://ghcr.io/run-ai/karta`
**Chart:** `oci://ghcr.io/run-ai/karta`

Karta is a CRD-based resource abstraction layer with its own operator. It replaced the legacy `ResourceInterface` (RI) CRD as the way Run:ai represents and reconciles workload resource attachments in the cluster. The `karta-migrate` Helm hook handles the one-time migration from RI to Karta CRs on upgrade.

## KAI Resource Management *(coming)*

**Source:** TBD
**Chart:** TBD

KAI Resource Management (KRM) is a planned external component that will own node pool management, pod group scheduling coordination, and project-to-namespace reconciliation. When it ships, it replaces the following components currently built from the Run:ai cluster source:

- `nodepool-controller` - manages NodePool CRDs and per-pool scheduler instances
- `pod-group-assigner` - assigns PodGroups to node pools for multi-node-pool scheduling
- `project-controller` - reconciles Project CRDs into Namespace, Queue, LimitRange, and RBAC resources

KRM will also take ownership of the Department, Project, and NodePool CRDs (currently in the `run.ai` and `kai.resources` API groups), and will pull in KAI Scheduler as a sub-dependency.

---

## Integrating a Custom Version

To replace any of these dependencies with a custom build:

1. Fork the relevant external repository and make your changes.
2. Build and publish your custom chart to an OCI registry.
3. Update the dependency entry in [`runai-chart/Chart.yaml`](runai-cluster/cluster/runai-chart/Chart.yaml) - change `repository` to your registry and `version` to your build.
4. Run `helm dep update` in `runai-cluster/cluster/runai-chart/` to pull in your version.
5. Deploy with `helm upgrade` or via `local-dev.sh`.
