# Advanced Cluster Configurations

Advanced cluster configurations allow you to customize your NVIDIA Run:ai cluster deployment to support your environment. Some settings may be required for deployment, while others can be fine-tuned to align with organizational policies, security requirements, or other operational preferences.

By adjusting these configurations, you can influence system behavior, including functionality, scheduling policies, and resource management, giving you greater control over how the cluster operates. This article provides guidance on configuring and managing these settings so you can adapt your NVIDIA Run:ai cluster to your organization’s needs.

## Configuration Scope

The Helm chart provides the complete set of configuration options for the NVIDIA Run:ai cluster. Some options, such as `global.affinity`, are available only through Helm. The rest of the configurable settings are grouped under `clusterConfig`. These `clusterConfig` settings can be applied through Helm as part of your deployment or upgrade process.

The `clusterConfig` subset can also be managed at runtime through the `runaiconfig` Custom Resource (under `spec`). For details, see [Modify cluster configurations at runtime](cluster-config.md#modify-cluster-configurations-at-runtime).

At runtime, `runaiconfig` is the source of truth for the active cluster configuration. If a configuration key is defined in both Helm and `runaiconfig` and the values differ, a Helm upgrade will overwrite the `runaiconfig` value to match the chart.

<Info>

**Note**

The approach remains backward compatible so existing clusters configured via `runaiconfig` continue to work. However, when using Helm, you should manage configurations exclusively through Helm values. Mixing Helm values and manual `runaiconfig` edits is not recommended.

</Info>

## Helm Chart Values

The NVIDIA Run:ai cluster installation can be customized to support your environment via Helm [values files](https://helm.sh/docs/chart_template_guide/values_files/) or [Helm install](https://helm.sh/docs/helm/helm_install/) flags. For example:

```yaml
# values.yaml
global:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
          - matchExpressions:
              - key: node-role.kubernetes.io/runai-system
                operator: Exists
```

## Modify Cluster Configurations at Runtime

The `clusterConfig` subset of settings can also be managed at runtime via the `runaiconfig` [Kubernetes Custom Resource](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/).

1.  To edit the cluster configurations, run:

    ```bash
    kubectl edit runaiconfig runai -n runai
    ```
2.  To see the full `runaiconfig` object structure, use:

    ```bash
    kubectl get crds/runaiconfigs.run.ai -n runai -o yaml
    ```

When using `runaiconfig`, the `clusterConfig` values appear under `spec`.

## Configurations

The following configurations allow you to enable or disable features, control permissions, and customize the behavior of your NVIDIA Run:ai cluster

<Info>

**Note**

* Keys that start with `clusterConfig` are available both in Helm and at runtime in `runaiconfig` (under `spec`). All other keys are available only through Helm.
* At runtime, `runaiconfig` reflects the active configuration. If specific `clusterConfig` values set in `runaiconfig` differ from those in Helm, a subsequent Helm upgrade will overwrite only those fields to align with the chart.

</Info>

### Helm Only Configurations

| Key                                              | Description                                                                                                                                                                                                                                                                                                               |
| ------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `global.image.registry` _(string)_               | <p>Global Docker image registry<br/>Default: <code>""</code></p>                                                                                                                                                                                                                                                           |
| `global.additionalImagePullSecrets` _(list)_     | <p>List of image pull secrets references<br/>Default: <code>[]</code></p>                                                                                                                                                                                                                                                  |
| `global.affinity` _(object)_                     | <p>Sets the system nodes where NVIDIA Run:ai system-level services are scheduled. Using global.affinity will overwrite the <a href="node-roles.md">node roles</a> set set using <code>kubectl</code>.<br/>Default: Prefer to schedule on nodes that are labeled with <code>node-role.kubernetes.io/runai-system</code></p> |
| `global.tolerations` _(object)_                  | Configure Kubernetes tolerations for NVIDIA Run:ai system-level services                                                                                                                                                                                                                                                  |
| `global.additionalJobLabels` _(object)_          | <p>Set NVIDIA Run:ai and 3rd party services' <a href="https://kubernetes.io/docs/concepts/overview/working-with-objects/labels">Pod Labels </a>in a format of key/value pairs.<br/>Default: <code>""</code></p>                                                                                                            |
| `global.additionalJobAnnotations` _(object)_     | <p>Set NVIDIA Run:ai and 3rd party services' <a href="https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/">Annotations</a> in a format of key/value pairs.<br/>Default: <code>""</code></p>                                                                                                     |
| `global.customCA.enabled`                        | Enables the use of a custom Certificate Authority (CA) in your deployment. When set to `true`, the system is configured to trust a user-provided CA certificate for secure communication.                                                                                                                                 |
| `global.customCAGit.enabled`                     | Enables the use of a custom Certificate Authority (CA) for Git data sources. When set to `true`, the system uses the global CA certificate defined at installation unless overridden using `global.customCAGit.secret.name`.                                                                                              |
| `global.customCAGit.secret.name`                 | Specifies the name of the Kubernetes secret that contains a custom CA certificate for Git data sources. Overrides the default global CA when `global.customCAGit.enabled` is set to `true`.                                                                                                                               |
| `global.customCAS3.enabled`                      | Enables the use of a custom Certificate Authority (CA) for S3 data sources. When set to `true`, the system uses the global CA certificate defined at installation unless overridden using `global.customS3Git.secret.name`.                                                                                               |
| `global.customCAS3.secret.name`                  | Specifies the name of the Kubernetes secret that contains a custom CA certificate for S3 data sources. Overrides the default global CA when `global.customCAS3.enabled` is set to `true`.                                                                                                                                 |
| `openShift.securityContextConstraints.create`    | Enables the deployment of Security Context Constraints (SCC). Disable for CIS compliance. Default: `true`                                                                                                                                                                                                                 |
| `researcherService.ingress.tlsSecret` _(string)_ | Existing secret key where cluster [TLS certificates](../system-requirements/networking-setup.md#tls-certificates) are stored (non-OpenShift) Default: `runai-cluster-domain-tls-secret`                                                                                   |
| `researcherService.route.tlsSecret` _(string)_   | Existing secret key where cluster [TLS certificates](../system-requirements/networking-setup.md#tls-certificates) are stored (OpenShift only) Default: `""`                                                                                                               |
| `controlPlane.existingSecret`                    | Specifies the name of the existing Kubernetes secret where the cluster’s `clientSecret` used for secure connection with the control plane is stored.                                                                                                                                                                      |
| `controlPlane.secretKeys.clientSecret`           | Specifies the key within the `controlPlane.existingSecret` that stores the cluster’s `clientSecret` used for secure connection with the control plane.                                                                                                                                                                    |

### All Configurations (Helm and runaiconfig)

| Helm Key | runaiconfig Key | Description |
| -------- | --------------- | ----------- |
| `clusterConfig.global.nodeAffinity.restrictScheduling` _(boolean)_ | `spec.global.nodeAffinity.restrictScheduling` _(boolean)_ | Enables setting [node roles](node-roles.md) and restricting workload scheduling to designated nodes. Default: `false` |
| `clusterConfig.global.ingress.ingressClass` | `spec.global.ingress.ingressClass` | NVIDIA Run:ai uses NGINX as the default ingress controller. If your cluster has a different ingress controller, you can configure the ingress class to be created by NVIDIA Run:ai. |
| `clusterConfig.global.subdomainSupport` _(boolean)_ | `spec.global.subdomainSupport` _(boolean)_ | Enables host-based routing, exposing workload URLs as subdomains on the cluster's [FQDN](../system-requirements/networking-setup.md#fqdn-and-dns). To use path-based routing instead, set to `false`. For setup instructions, see [Host-Based Routing (Ingress Resource)](../installation/install-using-helm/preparations.md#host-based-routing-ingress-resource). For more details, see [External Access to Containers](networking/external-access-to-containers.md). Default: `true` |
| `clusterConfig.global.devicePluginBindings` _(boolean)_ | `spec.global.devicePluginBindings` _(boolean)_ | Instruct NVIDIA Run:ai fractions to use device plugin for host mount instead of NVIDIA Run:ai fractions using explicit host path mount configuration on the pod. See [GPU fractions](../../platform-management/runai-scheduler/resource-optimization/fractions.md) and [dynamic GPU fractions](../../platform-management/runai-scheduler/resource-optimization/dynamic-fractions.md). Default: `false` |
| `clusterConfig.global.enableWorkloadOwnershipProtection` _(boolean)_ | `spec.global.enableWorkloadOwnershipProtection` _(boolean)_ | Prevents users within the same project from deleting workloads created by others. Default: `false` |
| `clusterConfig.global.requireDefaultPodAntiAffinity` _(boolean)_ | `spec.global.requireDefaultPodAntiAffinity` _(boolean)_ | When enabled, NVIDIA Run:ai applies a default pod anti-affinity rule that attempts to prevent pods belonging to the same service from being scheduled on the same node. Default: `true` |
| `clusterConfig.project-controller.createNamespaces` _(boolean)_ | `spec.project-controller.createNamespaces` _(boolean)_ | Allows Kubernetes namespace creation for new projects. Default: `true` |
| `clusterConfig.project-controller.CreateRoleBindings` _(boolean)_ | `spec.project-controller.CreateRoleBindings` _(boolean)_ | Specifies if role bindings should be created in the project's namespace. Default: `true` |
| `clusterConfig.project-controller.limitRange` _(boolean)_ | `spec.project-controller.limitRange` _(boolean)_ | Specifies if limit ranges should be defined for projects. Default: `true` |
| `clusterConfig.project-controller.clusterWideSecret` _(boolean)_ | `spec.project-controller.clusterWideSecret` _(boolean)_ | Allows Kubernetes Secrets creation at the cluster scope. See [Credentials](../../workloads-in-nvidia-run-ai/workload-assets/credentials.md#creating-secrets-in-advance) for more details. Default: `true` |
| `clusterConfig.workload-controller.failureResourceCleanupPolicy` | `spec.workload-controller.failureResourceCleanupPolicy` | NVIDIA Run:ai cleans the workload's unnecessary resources. `All` - removes all resources of the failed workload; `None` - retains all resources; `KeepFailing` - removes all resources except those that encountered issues. Default: `All` |
| `clusterConfig.workload-controller.externalAuthUrlEnabled` _(boolean)_ | `spec.workload-controller.externalAuthUrlEnabled` _(boolean)_ | Determines whether to use the OAuth2 proxy when submitting workloads that expose an external URL. Set to `false` to disable. Default: `true` |
| `clusterConfig.inference-workload-controller.externalAuthUrlEnabled` _(boolean)_ | `spec.inference-workload-controller.externalAuthUrlEnabled` _(boolean)_ | Determines whether to use the OAuth2 proxy when submitting inference workloads that expose an external URL. Set to `false` to disable. Default: `true` |
| `clusterConfig.workload-controller.GPUNetworkAccelerationEnabled` | `spec.workload-controller.GPUNetworkAccelerationEnabled` | Enables GPU network acceleration for workloads managed by the workload controller. See [Using GB200 NVL72 and Multi-Node NVLink Domains](../../platform-management/aiinitiatives/resources/using-gb200.md) for more details. Default: `false` |
| `clusterConfig.inference-workload-controller.GPUNetworkAccelerationEnabled` | `spec.inference-workload-controller.GPUNetworkAccelerationEnabled` | Enables GPU network acceleration for workloads managed by the inference workload controller. See [Using GB200 NVL72 and Multi-Node NVLink Domains](../../platform-management/aiinitiatives/resources/using-gb200.md) for more details. Default: `false` |
| `clusterConfig.anyworkload-controller.GPUNetworkAccelerationEnabled` | `spec.anyworkload-controller.GPUNetworkAccelerationEnabled` | Enables GPU network acceleration for [supported workload types](../../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md). See [Using GB200 NVL72 and Multi-Node NVLink Domains](../../platform-management/aiinitiatives/resources/using-gb200.md) for more details. Default: `false` |
| `clusterConfig.anyworkload-controller.openshiftRouteConfiguration.terminationPolicy` _(string)_ | `spec.anyworkload-controller.openshiftRouteConfiguration.terminationPolicy` _(string)_ | Sets the TLS termination policy for the OpenShift `Route` objects that NVIDIA Run:ai creates for workloads. Set to `passthrough` to terminate TLS inside the pod for end-to-end encryption instead of at the OpenShift Router. Applies to OpenShift clusters only, across all workload types. For more details, see [External Access to Containers](networking/external-access-to-containers.md#openshift-route-tls-termination). Default: `edge` |
| `clusterConfig.workload-controller.additionalPodLabels` | `spec.workload-controller.additionalPodLabels` | Set additional workload [Pod Labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels) in a format of key/value pairs. Default: `""` |
| `clusterConfig.mps-server.enabled` _(boolean)_ | `spec.mps-server.enabled` _(boolean)_ | Enabled when using [NVIDIA MPS](https://docs.nvidia.com/deploy/mps/index.html). Default: `false` |
| `clusterConfig.daemonSetsTolerations` _(object)_ | `spec.daemonSetsTolerations` _(object)_ | Configure Kubernetes tolerations for NVIDIA Run:ai daemonSets / engine. |
| `clusterConfig.runai-container-toolkit.enabled` _(boolean)_ | `spec.runai-container-toolkit.enabled` _(boolean)_ | Enables workloads to use [GPU fractions](../../platform-management/runai-scheduler/resource-optimization/fractions.md). Default: `true` |
| `clusterConfig.runai-container-toolkit.logLevel` | `spec.runai-container-toolkit.logLevel` | Specifies the NVIDIA Run:ai-container-toolkit logging level: `SPAM`, `DEBUG`, `INFO`, `NOTICE`, `WARN`, or `ERROR`. Default: `INFO` |
| `clusterConfig.node-scale-adjuster.args.gpuMemoryToFractionRatio` _(object)_ | `spec.node-scale-adjuster.args.gpuMemoryToFractionRatio` _(object)_ | A scaling-pod requesting a single GPU device will be created for every 1 to 10 pods requesting fractional GPU memory (1/gpuMemoryToFractionRatio). This value represents the ratio (0.1-0.9) of fractional GPU memory (any size) to GPU fraction (portion) conversion. Default: `0.1` |
| `clusterConfig.global.core.dynamicFractions.enabled` _(boolean)_ | `spec.global.core.dynamicFractions.enabled` _(boolean)_ | Enables [dynamic GPU fractions](../../platform-management/runai-scheduler/resource-optimization/dynamic-fractions.md). Default: `true` |
| `clusterConfig.global.core.swap.enabled` _(boolean)_ | `spec.global.core.swap.enabled` _(boolean)_ | Enables [memory swap](../../platform-management/runai-scheduler/resource-optimization/memory-swap.md) for GPU workloads. Default: `false` |
| `clusterConfig.global.core.swap.biDirectional` _(string)_ | `spec.global.core.swap.biDirectional` _(string)_ | Sets the read/write memory mode of GPU memory swap to bi-directional (fully duplex). This produces higher performance (typically +80%) vs. uni-directional (simplex) read-write operations. For more details, see [GPU memory swap](../../platform-management/runai-scheduler/resource-optimization/memory-swap.md). Default: `false` |
| `clusterConfig.global.core.swap.mode` _(string)_ | `spec.global.core.swap.mode` _(string)_ | Sets the GPU to CPU memory swap method to use UVA and optimized memory prefetch for optimized performance in some scenarios. For more details, see [GPU memory swap](../../platform-management/runai-scheduler/resource-optimization/memory-swap.md). Default: not set. To enable, set `mode=mapped`. |
| `clusterConfig.global.core.nodeScheduler.enabled` _(boolean)_ | `spec.global.core.nodeScheduler.enabled` _(boolean)_ | Enables the [node-level scheduler](../../platform-management/runai-scheduler/resource-optimization/node-level-scheduler.md). Default: `false` |
| `clusterConfig.global.core.timeSlicing.mode` _(string)_ | `spec.global.core.timeSlicing.mode` _(string)_ | Sets the [GPU time-slicing mode](../../platform-management/runai-scheduler/resource-optimization/time-slicing.md). `timesharing` - all pods on a GPU share compute time evenly; `strict` - each pod gets an exact time slice based on its memory fraction; `fair` - each pod gets an exact time slice and unused compute time is split evenly. Default: `timesharing` |
| `clusterConfig.runai-scheduler.args.fullHierarchyFairness` _(boolean)_ | `spec.runai-scheduler.args.fullHierarchyFairness` _(boolean)_ | Enables fairness between departments, on top of projects fairness. Default: `true` |
| `clusterConfig.runai-scheduler.args.defaultStalenessGracePeriod` | `spec.runai-scheduler.args.defaultStalenessGracePeriod` | Sets the timeout before the scheduler evicts a stale pod-group (gang) that went below its min-members in running state. `0s` - immediately; `-1` - never. Default: `60s` |
| `clusterConfig.runai-scheduler.args.verbosity` _(int)_ | `spec.runai-scheduler.args.verbosity` _(int)_ | Configures the level of detail in the logs generated by the scheduler service. Default: `4` |
| `clusterConfig.pod-grouper.args.gangSchedulingKnative` _(boolean)_ | `spec.pod-grouper.args.gangSchedulingKnative` _(boolean)_ | Enables gang scheduling for inference workloads. For backward compatibility with versions earlier than v2.19, set to `false`. Default: `false` |
| `clusterConfig.pod-grouper.args.gangScheduleArgoWorkflow` _(boolean)_ | `spec.pod-grouper.args.gangScheduleArgoWorkflow` _(boolean)_ | Groups all pods of a single ArgoWorkflow workload into a single Pod-Group for gang scheduling. Default: `true` |
| `clusterConfig.limitRange.cpuDefaultRequestCpuLimitFactorNoGpu` _(string)_ | `spec.limitRange.cpuDefaultRequestCpuLimitFactorNoGpu` _(string)_ | Sets a default ratio between the CPU request and the limit for workloads without GPU requests. Default: `0.1` |
| `clusterConfig.limitRange.memoryDefaultRequestMemoryLimitFactorNoGpu` _(string)_ | `spec.limitRange.memoryDefaultRequestMemoryLimitFactorNoGpu` _(string)_ | Sets a default ratio between the memory request and the limit for workloads without GPU requests. Default: `0.1` |
| `clusterConfig.limitRange.cpuDefaultRequestGpuFactor` _(string)_ | `spec.limitRange.cpuDefaultRequestGpuFactor` _(string)_ | Sets a default amount of CPU allocated per GPU when the CPU is not specified. Default: `100` |
| `clusterConfig.limitRange.cpuDefaultLimitGpuFactor` _(int)_ | `spec.limitRange.cpuDefaultLimitGpuFactor` _(int)_ | Sets a default CPU limit based on the number of GPUs requested when no CPU limit is specified. Default: none |
| `clusterConfig.limitRange.memoryDefaultRequestGpuFactor` _(string)_ | `spec.limitRange.memoryDefaultRequestGpuFactor` _(string)_ | Sets a default amount of memory allocated per GPU when the memory is not specified. Default: `100Mi` |
| `clusterConfig.limitRange.memoryDefaultLimitGpuFactor` _(string)_ | `spec.limitRange.memoryDefaultLimitGpuFactor` _(string)_ | Sets a default memory limit based on the number of GPUs requested when no memory limit is specified. Default: none |

### Third-Party Service Images

The NVIDIA Run:ai cluster bundles Valkey, oauth2-proxy, and git-sync. For what these services do, see [Third-Party Services](../system-requirements/third-party-services.md).

| Helm Key | runaiconfig Key | Description |
| -------- | --------------- | ----------- |
| `clusterConfig.cluster-redis.image.registry` | `spec.cluster-redis.image.registry` | Container registry hostname for the Valkey image. |
| `clusterConfig.cluster-redis.image.name` | `spec.cluster-redis.image.name` | Image name for Valkey. |
| `clusterConfig.cluster-redis.image.tag` | `spec.cluster-redis.image.tag` | Image tag for Valkey. |
| `clusterConfig.cluster-redis.image.pullPolicy` | `spec.cluster-redis.image.pullPolicy` | Image pull policy for Valkey. |
| `clusterConfig.anyworkload-controller.runaiOauth2ProxyImage.registry` | `spec.anyworkload-controller.runaiOauth2ProxyImage.registry` | Container registry hostname for the oauth2-proxy image used by all workload types. |
| `clusterConfig.anyworkload-controller.runaiOauth2ProxyImage.name` | `spec.anyworkload-controller.runaiOauth2ProxyImage.name` | Image name for oauth2-proxy (all workload types). |
| `clusterConfig.anyworkload-controller.runaiOauth2ProxyImage.tag` | `spec.anyworkload-controller.runaiOauth2ProxyImage.tag` | Image tag for oauth2-proxy (all workload types). |
| `clusterConfig.anyworkload-controller.runaiOauth2ProxyImage.pullPolicy` | `spec.anyworkload-controller.runaiOauth2ProxyImage.pullPolicy` | Image pull policy for oauth2-proxy (all workload types). |
| `clusterConfig.inference-workload-controller.runaiOauth2ProxyImage.registry` | `spec.inference-workload-controller.runaiOauth2ProxyImage.registry` | Container registry hostname for the oauth2-proxy image used by inference workloads. |
| `clusterConfig.inference-workload-controller.runaiOauth2ProxyImage.name` | `spec.inference-workload-controller.runaiOauth2ProxyImage.name` | Image name for oauth2-proxy (inference workloads). |
| `clusterConfig.inference-workload-controller.runaiOauth2ProxyImage.tag` | `spec.inference-workload-controller.runaiOauth2ProxyImage.tag` | Image tag for oauth2-proxy (inference workloads). |
| `clusterConfig.inference-workload-controller.runaiOauth2ProxyImage.pullPolicy` | `spec.inference-workload-controller.runaiOauth2ProxyImage.pullPolicy` | Image pull policy for oauth2-proxy (inference workloads). |
| `clusterConfig.workload-controller.runaiOauth2ProxyImage.registry` | `spec.workload-controller.runaiOauth2ProxyImage.registry` | Container registry hostname for the oauth2-proxy image used by standard workloads. |
| `clusterConfig.workload-controller.runaiOauth2ProxyImage.name` | `spec.workload-controller.runaiOauth2ProxyImage.name` | Image name for oauth2-proxy (standard workloads). |
| `clusterConfig.workload-controller.runaiOauth2ProxyImage.tag` | `spec.workload-controller.runaiOauth2ProxyImage.tag` | Image tag for oauth2-proxy (standard workloads). |
| `clusterConfig.workload-controller.runaiOauth2ProxyImage.pullPolicy` | `spec.workload-controller.runaiOauth2ProxyImage.pullPolicy` | Image pull policy for oauth2-proxy (standard workloads). |
| `clusterConfig.inference-workload-controller.gitSyncImage.registry` | `spec.inference-workload-controller.gitSyncImage.registry` | Container registry hostname for the git-sync image used by inference workloads. |
| `clusterConfig.inference-workload-controller.gitSyncImage.name` | `spec.inference-workload-controller.gitSyncImage.name` | Image name for git-sync (inference workloads). |
| `clusterConfig.inference-workload-controller.gitSyncImage.tag` | `spec.inference-workload-controller.gitSyncImage.tag` | Image tag for git-sync (inference workloads). |
| `clusterConfig.inference-workload-controller.gitSyncImage.pullPolicy` | `spec.inference-workload-controller.gitSyncImage.pullPolicy` | Image pull policy for git-sync (inference workloads). |
| `clusterConfig.workload-controller.gitSyncImage.registry` | `spec.workload-controller.gitSyncImage.registry` | Container registry hostname for the git-sync image used by standard workloads. |
| `clusterConfig.workload-controller.gitSyncImage.name` | `spec.workload-controller.gitSyncImage.name` | Image name for git-sync (standard workloads). |
| `clusterConfig.workload-controller.gitSyncImage.tag` | `spec.workload-controller.gitSyncImage.tag` | Image tag for git-sync (standard workloads). |
| `clusterConfig.workload-controller.gitSyncImage.pullPolicy` | `spec.workload-controller.gitSyncImage.pullPolicy` | Image pull policy for git-sync (standard workloads). |

### NVIDIA Run:ai Services Resource Management

NVIDIA Run:ai cluster includes many different services. To simplify resource management, the configuration structure allows you to configure the containers CPU / memory resources for each service individually or group of services together.

<Info>

**Note**

For resource recommendations, see [Vertical scaling](../operations/scaling.md#vertical-scaling). To automatically adjust resources based on actual usage, see [Vertical Pod Autoscaling (VPA)](../operations/vpa.md). Note that static resource configuration remains the primary mechanism. VPA works alongside it by generating recommendations or dynamically applying adjustments depending on the configured update mode.

</Info>

| Service Group      | Description                                                                                                      | NVIDIA Run:ai containers                                                                                                                                                                             |
| ------------------ | ---------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| SchedulingServices | Containers associated with the NVIDIA Run:ai Scheduler                                                           | Scheduler, StatusUpdater, MetricsExporter, PodGrouper, PodGroupAssigner, PodGroupController, QueueController, NodePoolController, Binder, DevicePlugin                                               |
| SyncServices       | Containers associated with syncing updates between the NVIDIA Run:ai cluster and the NVIDIA Run:ai control plane | Agent, ClusterSync, AssetsSync                                                                                                                                                                       |
| WorkloadServices   | Containers associated with submitting NVIDIA Run:ai workloads                                                    | WorkloadController, JobController, WorkloadOverseer, ExternalWorkloadIntegrator, ClusterRedis, ClusterAPI, InferenceWorkloadController, ResearcherService, SharedObjectsController, WorkloadExporter |

Apply the following configuration in order to change resources request and limit for a group of services:

<Tabs>

<Tab title="Helm">

```yaml
clusterConfig:
  global:
   <service-group-name>: # schedulingServices | syncServices | workloadServices
     resources:
       limits:
         cpu: 1000m
         memory: 1Gi
       requests:
         cpu: 100m
         memory: 512Mi
```

</Tab>

<Tab title="runaiconfig">

```yaml
spec:
  global:
   <service-group-name>: # schedulingServices | syncServices | workloadServices
     resources:
       limits:
         cpu: 1000m
         memory: 1Gi
       requests:
         cpu: 100m
         memory: 512Mi
```

</Tab>

</Tabs>

Or, apply the following configuration in order to change resources request and limit for each service individually:

<Tabs>

<Tab title="Helm">

```yaml
clusterConfig:
  <service-name>: # for example: pod-grouper
    resources:
      limits:
        cpu: 1000m
        memory: 1Gi
      requests:
        cpu: 100m
        memory: 512Mi
```

</Tab>

<Tab title="runaiconfig">

```yaml
spec:
  <service-name>: # for example: pod-grouper
    resources:
      limits:
        cpu: 1000m
        memory: 1Gi
      requests:
        cpu: 100m
        memory: 512Mi
```

</Tab>

</Tabs>

### NVIDIA Run:ai Services Replicas

By default, all NVIDIA Run:ai containers are deployed with a single replica. Some services support multiple replicas for redundancy and performance.

To simplify configuring replicas, a global replicas configuration can be set and is applied to all supported services:

<Tabs>

<Tab title="Helm">

```yaml
clusterConfig:
  global: 
    replicaCount: 1 # default
```

</Tab>

<Tab title="runaiconfig">

```yaml
spec:
  global: 
    replicaCount: 1 # default
```

</Tab>

</Tabs>

This can be overwritten for specific services (if supported). Services without the `replicas` configuration does not support replicas:

<Tabs>

<Tab title="Helm">

```yaml
clusterConfig:
  <service-name>: # for example: pod-grouper
    replicas: 1 # default
```

</Tab>

<Tab title="runaiconfig">

```yaml
spec:
  <service-name>: # for example: pod-grouper
    replicas: 1 # default
```

</Tab>

</Tabs>

### Prometheus

NVIDIA Run:ai uses two separate Prometheus instances:

* The Prometheus instance used for metrics collection and alerting.
* A dedicated Prometheus instance used the NVIDIA Run:ai Scheduler for tracking historical resource usage and enabling time-based fairshare.

#### Metrics Collection

The Prometheus instance in NVIDIA Run:ai is used for metrics collection and alerting.

The configuration scheme follows the official [PrometheusSpec](https://prometheus-operator.dev/docs/api-reference/api/#monitoring.coreos.com/v1.PrometheusSpec) and supports additional custom configurations. The PrometheusSpec schema is available using the `spec.prometheus.spec` configuration.

A common use case using the PrometheusSpec is for metrics retention. This prevents metrics loss during potential connectivity issues and can be achieved by configuring local temporary metrics retention. For more information, see [Prometheus Storage](https://prometheus.io/docs/prometheus/latest/storage/#storage):

<Tabs>

<Tab title="Helm">

```yaml
clusterConfig:  
  prometheus:
    spec: # PrometheusSpec
      retention: 2h # default 
      retentionSize: 20GB
```

</Tab>

<Tab title="runaiconfig">

```yaml
spec:  
  prometheus:
    spec: # PrometheusSpec
      retention: 2h # default 
      retentionSize: 20GB
```

</Tab>

</Tabs>

In addition to the PrometheusSpec schema, some custom NVIDIA Run:ai configurations are also available:

* Additional labels - Set additional labels for NVIDIA Run:ai's [built-in alerts](../operations/observability/system-monitoring.md#built-in-alerts) sent by Prometheus.
* Log level configuration - Configure the `logLevel` setting for the Prometheus container.
* Image override - Use `prometheus.spec.image` to manually specify the Prometheus image reference. Due to a known issue, the `imageRegistry` setting in the Prometheus Helm chart is ignored. To pull the image from a different registry, specify the full image reference. Default: `quay.io/prometheus/prometheus`.
* Image pull secrets - Use `prometheus.spec.imagePullSecrets` to list Kubernetes image pull secrets in the `runai` namespace. This is particularly relevant for air-gapped installations where pulling Prometheus images requires authentication. Default: `[]`.
* Advanced metrics (GPU profiling metrics) - Use `prometheus.spec.config.advancedMetricsEnabled` to activate GPU profiling metrics from NVIDIA DCGM. When enabled, Prometheus collects and aggregates advanced GPU performance data such as SM activity, memory bandwidth, and tensor core utilization. For setup instructions, see [GPU profiling metrics](../../platform-management/monitor-performance/gpu-profiling-metrics.md).
* OTLP receiver - Use `prometheus.config.enableOTLPReceiver` to enable or disable the OpenTelemetry Protocol (OTLP) receiver on the NVIDIA Run:ai Prometheus instance. When enabled, Knative Serving can push inference request metrics to Prometheus over OTLP, surfacing them in the NVIDIA Run:ai API and UI. Default: `true`.

<Tabs>

<Tab title="Helm">

```yaml
clusterConfig:  
  prometheus:
    logLevel: info # debug | info | warn | error
    additionalAlertLabels:
      - env: prod # example
    config:
      enableOTLPReceiver: true # default: true
```

</Tab>

<Tab title="runaiconfig">

```yaml
spec:  
  prometheus:
    logLevel: info # debug | info | warn | error
    additionalAlertLabels:
      - env: prod # example
    config:
      enableOTLPReceiver: true # default: true
```

</Tab>

</Tabs>

#### Time-Based Fairshare

Time-based fairshare relies on historical resource usage metrics collected by a dedicated Prometheus instance. By default, persistent storage is not configured for this Prometheus instance. As a result, if the Prometheus pod restarts or fails, all historical usage data is lost. In this scenario, the Scheduler recalculates fairshare from scratch and makes scheduling decisions without taking previously collected data into account until new data is accumulated.

To preserve historical usage data across restarts, you can enable persistent storage for the Prometheus instance used for time-based fairshare. It is recommended to use network-based storage that is accessible from all nodes in the cluster.

To enable persistent storage, configure the following cluster settings:

<Tabs>

<Tab title="Helm">

```bash
clusterConfig:
  historical-usage-prometheus:
    enablePersistentStorage: true #default false
    storageSize: 50Gi # default 50Gi
    storageClassName: fast-ssd # defaults to the default storage class configured in the cluster  
```

</Tab>

<Tab title="runaiconfig">

```bash
spec:
  historical-usage-prometheus:
    enablePersistentStorage: true #default false
    storageSize: 50Gi # default 50Gi
    storageClassName: fast-ssd # defaults to the default storage class configured in the cluster
```

</Tab>

</Tabs>

### NVIDIA Run:ai Managed Nodes

To include or exclude specific nodes from running workloads within a cluster managed by NVIDIA Run:ai, use the `nodeSelectorTerms` flag. For additional details, see [Kubernetes nodeSelector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector).

Label the nodes using the below:

* key - Label key (e.g., zone, instance-type).
* operator - Operator defining the inclusion/exclusion condition (In, NotIn, Exists, DoesNotExist).
* values - List of values for the key when using In or NotIn.

The below example shows how to include NVIDIA GPUs only and exclude all other GPU types in a cluster with mixed nodes, based on product type GPU label:

<Tabs>

<Tab title="Helm">

```yaml
clusterConfig:   
  global:
     managedNodes:
       inclusionCriteria:
          nodeSelectorTerms:
          - matchExpressions:
            - key: nvidia.com/gpu.product  
              operator: Exists
```

</Tab>

<Tab title="runaiconfig">

```yaml
spec:   
  global:
     managedNodes:
       inclusionCriteria:
          nodeSelectorTerms:
          - matchExpressions:
            - key: nvidia.com/gpu.product  
              operator: Exists
```

</Tab>

</Tabs>

### Custom Certificate Authority for Git and S3

To override the default global CA used by the system and inject a custom CA certificate for Git or S3 data sources, follow these steps:

1.  Create a Kubernetes secret with the custom CA certificate:

    ```bash
    kubectl -n runai create secret generic runai-cluster-git-ca \
        --from-file=runai-ca-git.pem=<ca_bundle_path>
    kubectl label secret runai-cluster-git-ca -n runai run.ai/cluster-wide=true run.ai/name=runai-ca-cert --overwrite
    ```
2. When installing the cluster, make sure the following flags are added to the helm command. See [Install cluster](../../installation-and-configuration/installation/install-using-helm/helm-install.md).

<Tabs>

<Tab title="Git">

```bash
--set global.customCAGit.enabled=true \
--set global.customCAGit.secret.name=<secret-name>
```

</Tab>

<Tab title="S3">

```bash
--set global.customCAS3.enabled=true \
--set global.customCAS3.secret.name=<secret-name>
```

</Tab>

</Tabs>
