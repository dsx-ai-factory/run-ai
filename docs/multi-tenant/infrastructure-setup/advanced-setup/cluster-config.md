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

| Key                                              | Description                                                                                                                                                                                                                                                                                                           |
| ------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `global.image.registry` _(string)_               | <p>Global Docker image registry<br/>Default: <code>""</code></p>                                                                                                                                                                                                                                                       |
| `global.additionalImagePullSecrets` _(list)_     | <p>List of image pull secrets references<br/>Default: <code>[]</code></p>                                                                                                                                                                                                                                              |
| `global.affinity` _(object)_                     | <p>Sets the system nodes where NVIDIA Run:ai system-level services are scheduled. Using global.affinity will overwrite the <a href="node-roles.md">node roles</a> set using <code>kubectl</code>.<br/>Default: Prefer to schedule on nodes that are labeled with <code>node-role.kubernetes.io/runai-system</code></p> |
| `global.tolerations` _(object)_                  | Configure Kubernetes tolerations for NVIDIA Run:ai system-level services                                                                                                                                                                                                                                              |
| `global.additionalJobLabels` _(object)_          | <p>Set NVIDIA Run:ai and 3rd party services' <a href="https://kubernetes.io/docs/concepts/overview/working-with-objects/labels">Pod Labels </a>in a format of key/value pairs.<br/>Default: <code>""</code></p>                                                                                                        |
| `global.additionalJobAnnotations` _(object)_     | <p>Set NVIDIA Run:ai and 3rd party services' <a href="https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/">Annotations</a> in a format of key/value pairs.<br/>Default: <code>""</code></p>                                                                                                 |
| `global.customCA.enabled`                        | Enables the use of a custom Certificate Authority (CA) in your deployment. When set to `true`, the system is configured to trust a user-provided CA certificate for secure communication.                                                                                                                             |
| `global.customCAGit.enabled`                     | Enables the use of a custom Certificate Authority (CA) for Git data sources. When set to `true`, the system uses the global CA certificate defined at installation unless overridden using `global.customCAGit.secret.name`.                                                                                          |
| `global.customCAGit.secret.name`                 | Specifies the name of the Kubernetes secret that contains a custom CA certificate for Git data sources. Overrides the default global CA when `global.customCAGit.enabled` is set to `true`.                                                                                                                           |
| `global.customCAS3.enabled`                      | Enables the use of a custom Certificate Authority (CA) for S3 data sources. When set to `true`, the system uses the global CA certificate defined at installation unless overridden using `global.customS3Git.secret.name`.                                                                                           |
| `global.customCAS3.secret.name`                  | Specifies the name of the Kubernetes secret that contains a custom CA certificate for S3 data sources. Overrides the default global CA when `global.customCAS3.enabled` is set to `true`.                                                                                                                             |
| `openShift.securityContextConstraints.create`    | Enables the deployment of Security Context Constraints (SCC). Disable for CIS compliance. Default: `true`                                                                                                                                                                                                             |
| `researcherService.ingress.tlsSecret` _(string)_ | Existing secret key where cluster [TLS certificates](../../organization-life-cycle/organization-onboarding/system-requirements.md#nvidia-run-ai-tls-certificates) are stored (non-OpenShift) Default: `runai-cluster-domain-tls-secret`                                                                               |
| `researcherService.route.tlsSecret` _(string)_   | Existing secret key where cluster [TLS certificates](../../organization-life-cycle/organization-onboarding/system-requirements.md#nvidia-run-ai-tls-certificates) are stored (OpenShift only) Default: `""`                                                                                                           |

### All Configurations (Helm and runaiconfig)

<table><thead><tr><th>Helm Key</th><th>runaiconfig Key</th><th>Description</th></tr></thead><tbody><tr><td><code>clusterConfig.global.nodeAffinity.restrictScheduling</code> <em>(boolean)</em></td><td><code>spec.global.nodeAffinity.restrictScheduling</code> <em>(boolean)</em></td><td>Enables setting <a href="node-roles.md">node roles</a> and restricting workload scheduling to designated nodes<br/>Default: <code>false</code></td></tr><tr><td><code>clusterConfig.global.ingress.ingressClass</code></td><td><code>spec.global.ingress.ingressClass</code></td><td>NVIDIA Run:ai uses NGINX as the default ingress controller. If your cluster has a different ingress controller, you can configure the ingress class to be created by NVIDIA Run:ai.</td></tr><tr><td><code>clusterConfig.global.subdomainSupport</code> <em>(boolean)</em></td><td><code>spec.global.subdomainSupport</code> <em>(boolean)</em></td><td>Enables host-based routing, exposing workload URLs as subdomains on the cluster's <a href="../../organization-life-cycle/organization-onboarding/system-requirements.md#fully-qualified-domain-name-fqdn">FQDN</a>. To use path-based routing instead, set to <code>false</code>. For setup instructions, see <a href="../../organization-life-cycle/organization-onboarding/system-requirements.md#host-based-routing-default">System requirements</a>. For details, see <a href="external-access-to-containers.md">External Access to Containers</a>.<br/>Default: <code>true</code></td></tr><tr><td><code>clusterConfig.global.devicePluginBindings</code> <em>(boolean)</em></td><td><code>spec.global.devicePluginBindings</code> <em>(boolean)</em></td><td>Instruct NVIDIA Run:ai fractions to use device plugin for host mount instead of NVIDIA Run:ai fractions using explicit host path mount configuration on the pod. See <a href="https://app.gitbook.com/s/MeXTfFihZ2OyRbNxZBY3/platform-management/runai-scheduler/resource-optimization/fractions">GPU fractions</a> and <a href="https://app.gitbook.com/s/MeXTfFihZ2OyRbNxZBY3/platform-management/runai-scheduler/resource-optimization/dynamic-fractions">dynamic GPU fractions</a>.<br/>Default: <code>false</code></td></tr><tr><td><code>clusterConfig.global.enableWorkloadOwnershipProtection</code> <em>(boolean)</em></td><td><code>spec.global.enableWorkloadOwnershipProtection</code> <em>(boolean)</em></td><td>Prevents users within the same project from deleting workloads created by others. This enhances workload ownership security and ensures better collaboration by restricting unauthorized modifications or deletions.<br/>Default: <code>false</code></td></tr><tr><td><code>clusterConfig.global.requireDefaultPodAntiAffinity</code> <em>(boolean)</em></td><td><code>spec.global.requireDefaultPodAntiAffinity</code> <em>(boolean)</em></td><td>When enabled, NVIDIA Run:ai applies a default pod anti-affinity rule that attempts to prevent pods belonging to the same service from being scheduled on the same node.<br/>Default: <code>true</code></td></tr><tr><td><code>clusterConfig.project-controller.createNamespaces</code> <em>(boolean)</em></td><td><code>spec.project-controller.createNamespaces</code> <em>(boolean)</em></td><td>Allows Kubernetes namespace creation for new projects<br/>Default: <code>true</code></td></tr><tr><td><code>clusterConfig.project-controller.CreateRoleBindings</code> <em>(boolean)</em></td><td><code>spec.project-controller.CreateRoleBindings</code> <em>(boolean)</em></td><td>Specifies if role bindings should be created in the project's namespace<br/>Default: <code>true</code></td></tr><tr><td><code>clusterConfig.project-controller.limitRange</code> <em>(boolean)</em></td><td><code>spec.project-controller.limitRange</code> <em>(boolean)</em></td><td>Specifies if limit ranges should be defined for projects<br/>Default: <code>true</code></td></tr><tr><td><code>clusterConfig.project-controller.clusterWideSecret</code> <em>(boolean)</em></td><td><code>spec.project-controller.clusterWideSecret</code> <em>(boolean)</em></td><td>Allows Kubernetes Secrets creation at the cluster scope. See <a href="https://app.gitbook.com/s/MeXTfFihZ2OyRbNxZBY3/workloads-in-nvidia-run-ai/workload-assets/credentials#creating-secrets-in-advance">Credentials</a> for more details.<br/>Default: <code>true</code></td></tr><tr><td><code>clusterConfig.workload-controller.failureResourceCleanupPolicy</code></td><td><code>spec.workload-controller.failureResourceCleanupPolicy</code></td><td><p>NVIDIA Run:ai cleans the workload's unnecessary resources:</p><ul><li><code>All</code> - Removes all resources of the failed workload</li><li><code>None</code> - Retains all resources</li><li><code>KeepFailing</code> - Removes all resources except for those that encountered issues (primarily for debugging purposes)</li></ul><p>Default: <code>All</code></p></td></tr><tr><td><code>clusterConfig.workload-controller.externalAuthUrlEnabled</code> <em>(boolean)</em></td><td><code>spec.workload-controller.externalAuthUrlEnabled</code> <em>(boolean)</em></td><td>Determines whether to use the OAuth2 proxy when submitting workloads that expose an external URL. Set to <code>false</code> to disable.<br/>Default: <code>true</code></td></tr><tr><td><code>clusterConfig.inference-workload-controller.externalAuthUrlEnabled</code> <em>(boolean)</em></td><td><code>spec.inference-workload-controller.externalAuthUrlEnabled</code> <em>(boolean)</em></td><td>Determines whether to use the OAuth2 proxy when submitting inference workloads that expose an external URL. Set to <code>false</code> to disable.<br/>Default: <code>true</code></td></tr><tr><td><code>clusterConfig.workload-controller.GPUNetworkAccelerationEnabled</code></td><td><code>spec.workload-controller.GPUNetworkAccelerationEnabled</code></td><td>Enables GPU network acceleration for workloads managed by the workload controller. See <a href="https://app.gitbook.com/s/MeXTfFihZ2OyRbNxZBY3/platform-management/aiinitiatives/resources/using-gb200">Using GB200 NVL72 and Multi-Node NVLink Domains</a> for more details.<br/>Default: <code>false</code></td></tr><tr><td><code>clusterConfig.inference-workload-controller.GPUNetworkAccelerationEnabled</code></td><td><code>spec.inference-workload-controller.GPUNetworkAccelerationEnabled</code></td><td>Enables GPU network acceleration for workloads managed by the inference workload controller. See <a href="https://app.gitbook.com/s/MeXTfFihZ2OyRbNxZBY3/platform-management/aiinitiatives/resources/using-gb200">Using GB200 NVL72 and Multi-Node NVLink Domains</a> for more details.<br/>Default: <code>false</code></td></tr><tr><td><code>clusterConfig.anyworkload-controller.GPUNetworkAccelerationEnabled</code></td><td><code>spec.anyworkload-controller.GPUNetworkAccelerationEnabled</code></td><td>Enables GPU network acceleration for <a href="https://app.gitbook.com/s/MeXTfFihZ2OyRbNxZBY3/workloads-in-nvidia-run-ai/workload-types/supported-workload-types">supported workload types</a>. See <a href="https://app.gitbook.com/s/MeXTfFihZ2OyRbNxZBY3/platform-management/aiinitiatives/resources/using-gb200">Using GB200 NVL72 and Multi-Node NVLink Domains</a> for more details.<br/>Default: <code>false</code></td></tr><tr><td><code>clusterConfig.workload-controller.additionalPodLabels</code></td><td><code>spec.workload-controller.additionalPodLabels</code></td><td>Set additional workload's <a href="https://kubernetes.io/docs/concepts/overview/working-with-objects/labels">Pod Labels </a>in a format of key/value pairs. Default: <code>""</code></td></tr><tr><td><code>clusterConfig.mps-server.enabled</code> <em>(boolean)</em></td><td><code>spec.mps-server.enabled</code> <em>(boolean)</em></td><td>Enabled when using <a href="https://docs.nvidia.com/deploy/mps/index.html">NVIDIA MPS</a><br/>Default: <code>false</code></td></tr><tr><td><code>clusterConfig.daemonSetsTolerations</code> <em>(object)</em></td><td><code>spec.daemonSetsTolerations</code> <em>(object)</em></td><td>Configure Kubernetes tolerations for NVIDIA Run:ai daemonSets / engine</td></tr><tr><td><code>clusterConfig.runai-container-toolkit.enabled</code> <em>(boolean)</em></td><td><code>spec.runai-container-toolkit.enabled</code> <em>(boolean)</em></td><td>Enables workloads to use <a href="https://app.gitbook.com/s/MeXTfFihZ2OyRbNxZBY3/platform-management/runai-scheduler/resource-optimization/fractions">GPU fractions</a><br/>Default: <code>true</code></td></tr><tr><td><code>clusterConfig.runai-container-toolkit.logLevel</code> <em>(boolean)</em></td><td><code>spec.runai-container-toolkit.logLevel</code> <em>(boolean)</em></td><td>Specifies the NVIDIA Run:ai-container-toolkit logging level: either 'SPAM', 'DEBUG', 'INFO', 'NOTICE', 'WARN', or 'ERROR'<br/>Default: <code>INFO</code></td></tr><tr><td><code>clusterConfig.node-scale-adjuster.args.gpuMemoryToFractionRatio</code> <em>(object)</em></td><td><code>spec.node-scale-adjuster.args.gpuMemoryToFractionRatio</code> <em>(object)</em></td><td>A scaling-pod requesting a single GPU device will be created for every 1 to 10 pods requesting fractional GPU memory (1/gpuMemoryToFractionRatio). This value represents the ratio (0.1-0.9) of fractional GPU memory (any size) to GPU fraction (portion) conversion.<br/>Default: <code>0.1</code></td></tr><tr><td><code>clusterConfig.global.core.dynamicFractions.enabled</code> <em>(boolean)</em></td><td><code>spec.global.core.dynamicFractions.enabled</code> <em>(boolean)</em></td><td>Enables <a href="https://app.gitbook.com/s/MeXTfFihZ2OyRbNxZBY3/platform-management/runai-scheduler/resource-optimization/dynamic-fractions">dynamic GPU fractions</a><br/>Default: <code>true</code></td></tr><tr><td><code>clusterConfig.global.core.swap.enabled</code> <em>(boolean)</em></td><td><code>spec.global.core.swap.enabled</code> <em>(boolean)</em></td><td>Enables <a href="https://app.gitbook.com/s/MeXTfFihZ2OyRbNxZBY3/platform-management/runai-scheduler/resource-optimization/memory-swap">memory swap</a> for GPU workloads<br/>Default: <code>false</code></td></tr><tr><td><code>clusterConfig.global.core.swap.limits.cpuRam</code> <em>(string)</em></td><td><code>spec.global.core.swap.limits.cpuRam</code> <em>(string)</em></td><td>Sets the CPU memory size used to swap GPU workloads<br/>Default:<code>100Gi</code></td></tr><tr><td><code>clusterConfig.global.core.swap.limits.reservedGpuRam</code> <em>(string)</em></td><td><code>spec.global.core.swap.limits.reservedGpuRam</code> <em>(string)</em></td><td>Sets the reserved GPU memory size used to swap GPU workloads<br/>Default: <code>2Gi</code></td></tr><tr><td><code>clusterConfig.global.core.swap.biDirectional</code> <em>(string)</em></td><td><code>spec.global.core.swap.biDirectional</code> <em>(string)</em></td><td>Sets the read/write memory mode of GPU memory swap to bi-directional (fully duplex). This produces higher performance (typically +80%) vs. uni-directional (simplex) read-write operations. For more details, see <a href="https://app.gitbook.com/s/MeXTfFihZ2OyRbNxZBY3/platform-management/runai-scheduler/resource-optimization/memory-swap">GPU memory swap</a>.<br/>Default: <code>false</code></td></tr><tr><td><code>clusterConfig.global.core.swap.mode</code> <em>(string)</em></td><td><code>spec.global.core.swap.mode</code> <em>(string)</em></td><td>Sets the GPU to CPU memory swap method to use UVA and optimized memory prefetch for optimized performance in some scenarios. For more details, see <a href="https://app.gitbook.com/s/MeXTfFihZ2OyRbNxZBY3/platform-management/runai-scheduler/resource-optimization/memory-swap">GPU memory swap</a>.<br/>Default: None. The parameter is not set by default. To add this parameter set <code>mode=mapped</code> .</td></tr><tr><td><code>clusterConfig.global.core.nodeScheduler.enabled</code> <em>(boolean)</em></td><td><code>spec.global.core.nodeScheduler.enabled</code> <em>(boolean)</em></td><td>Enables the <a href="https://app.gitbook.com/s/MeXTfFihZ2OyRbNxZBY3/platform-management/runai-scheduler/resource-optimization/node-level-scheduler">node-level scheduler</a><br/>Default: <code>false</code></td></tr><tr><td><code>clusterConfig.global.core.timeSlicing.mode</code> <em>(string)</em></td><td><code>spec.global.core.timeSlicing.mode</code> <em>(string)</em></td><td><p>Sets the <a href="https://app.gitbook.com/s/MeXTfFihZ2OyRbNxZBY3/platform-management/runai-scheduler/resource-optimization/time-slicing">GPU time-slicing mode</a>. Possible values:</p><ul><li><code>timesharing</code> - all pods on a GPU share the GPU compute time evenly.</li><li><code>strict</code> - each pod gets an exact time slice according to its memory fraction value.</li><li><code>fair</code> - each pod gets an exact time slice according to its memory fraction value and any unused GPU compute time is split evenly between the running pods.</li></ul><p>Default: <code>timesharing</code></p></td></tr><tr><td><code>clusterConfig.runai-scheduler.args.fullHierarchyFairness</code> <em>(boolean)</em></td><td><code>spec.runai-scheduler.args.fullHierarchyFairness</code> <em>(boolean)</em></td><td>Enables fairness between departments, on top of projects fairness<br/>Default: <code>true</code></td></tr><tr><td><code>clusterConfig.runai-scheduler.args.defaultStalenessGracePeriod</code></td><td><code>spec.runai-scheduler.args.defaultStalenessGracePeriod</code></td><td><p>Sets the timeout in seconds before the scheduler evicts a stale pod-group (gang) that went below its min-members in running state:</p><ul><li><code>0s</code> - Immediately (no timeout)</li><li><code>-1</code> - Never</li></ul><p>Default: <code>60s</code></p></td></tr><tr><td><code>clusterConfig.runai-scheduler.args.verbosity</code> <em>(int)</em></td><td><code>spec.runai-scheduler.args.verbosity</code> <em>(int)</em></td><td>Configures the level of detail in the logs generated by the scheduler service<br/>Default: <code>4</code></td></tr><tr><td><code>clusterConfig.pod-grouper.args.gangSchedulingKnative</code> <em>(boolean)</em></td><td><code>spec.pod-grouper.args.gangSchedulingKnative</code> <em>(boolean)</em></td><td>Enables gang scheduling for inference workloads.For backward compatibility with versions earlier than v2.19, change the value to false<br/>Default: <code>false</code></td></tr><tr><td><code>clusterConfig.pod-grouper.args.gangScheduleArgoWorkflow</code> <em>(boolean)</em></td><td><code>spec.pod-grouper.args.gangScheduleArgoWorkflow</code> <em>(boolean)</em></td><td>Groups all pods of a single ArgoWorkflow workload into a single Pod-Group for gang scheduling<br/>Default: <code>true</code></td></tr><tr><td><code>clusterConfig.limitRange.cpuDefaultRequestCpuLimitFactorNoGpu</code> <em>(string)</em></td><td><code>spec.limitRange.cpuDefaultRequestCpuLimitFactorNoGpu</code> <em>(string)</em></td><td>Sets a default ratio between the CPU request and the limit for workloads without GPU requests<br/>Default: <code>0.1</code></td></tr><tr><td><code>clusterConfig.limitRange.memoryDefaultRequestMemoryLimitFactorNoGpu</code> <em>(string)</em></td><td><code>spec.limitRange.memoryDefaultRequestMemoryLimitFactorNoGpu</code> <em>(string)</em></td><td>Sets a default ratio between the memory request and the limit for workloads without GPU requests<br/>Default: <code>0.1</code></td></tr><tr><td><code>clusterConfig.limitRange.cpuDefaultRequestGpuFactor</code> <em>(string)</em></td><td><code>spec.limitRange.cpuDefaultRequestGpuFactor</code> <em>(string)</em></td><td>Sets a default amount of CPU allocated per GPU when the CPU is not specified<br/>Default: <code>100</code></td></tr><tr><td><code>clusterConfig.limitRange.cpuDefaultLimitGpuFactor</code> <em>(int)</em></td><td><code>spec.limitRange.cpuDefaultLimitGpuFactor</code> <em>(int)</em></td><td>Sets a default CPU limit based on the number of GPUs requested when no CPU limit is specified<br/>Default: <code>NO DEFAULT</code></td></tr><tr><td><code>clusterConfig.limitRange.memoryDefaultRequestGpuFactor</code> <em>(string)</em></td><td><code>spec.limitRange.memoryDefaultRequestGpuFactor</code> <em>(string)</em></td><td>Sets a default amount of memory allocated per GPU when the memory is not specified<br/>Default: <code>100Mi</code></td></tr><tr><td><code>clusterConfig.limitRange.memoryDefaultLimitGpuFactor</code> <em>(string)</em></td><td><code>spec.limitRange.memoryDefaultLimitGpuFactor</code> <em>(string)</em></td><td>Sets a default memory limit based on the number of GPUs requested when no memory limit is specified<br/>Default: <code>NO DEFAULT</code></td></tr></tbody></table>

### NVIDIA Run:ai Services Resource Management

NVIDIA Run:ai cluster includes many different services. To simplify resource management, the configuration structure allows you to configure the containers CPU / memory resources for each service individually or group of services together.

<Info>

**Note**

For resource recommendations, see [Vertical scaling](../procedures/scaling.md#vertical-scaling). To automatically adjust resources based on actual usage, see [Vertical Pod Autoscaling (VPA)](vpa.md). Note that static resource configuration remains the primary mechanism. VPA works alongside it by generating recommendations or dynamically applying adjustments depending on the configured update mode.

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

For resource recommendations, see [Vertical scaling](../procedures/scaling.md#vertical-scaling).

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

* Additional labels - Set additional labels for NVIDIA Run:ai's [built-in alerts](../procedures/system-monitoring.md#built-in-alerts) sent by Prometheus.
* Log level configuration - Configure the `logLevel` setting for the Prometheus container.
* Image override - Use `prometheus.spec.image` to manually specify the Prometheus image reference. Due to a known issue, the `imageRegistry` setting in the Prometheus Helm chart is ignored. To pull the image from a different registry, specify the full image reference. Default: `quay.io/prometheus/prometheus`.
* Image pull secrets - Use `prometheus.spec.imagePullSecrets` to list Kubernetes image pull secrets in the `runai` namespace. This is particularly relevant for air-gapped installations where pulling Prometheus images requires authentication. Default: `[]`.
* Advanced metrics (GPU profiling metrics) - Use `prometheus.spec.config.advancedMetricsEnabled` to activate GPU profiling metrics from NVIDIA DCGM. When enabled, Prometheus collects and aggregates advanced GPU performance data such as SM activity, memory bandwidth, and tensor core utilization. For setup instructions, see [GPU profiling metrics](https://app.gitbook.com/s/fvuZQSF7VCx1Azh0LF0H/platform-management/monitor-performance/gpu-profiling-metrics).
* OTLP receiver - Use `prometheus.config.enableOTLPReceiver` to enable or disable the OpenTelemetry Protocol (OTLP) receiver on the NVIDIA Run:ai Prometheus instance. The OTLP receiver is required for inference request metrics when using Knative Serving 1.19 or later. Default: `true`.

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
