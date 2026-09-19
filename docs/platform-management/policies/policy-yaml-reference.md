# Policy YAML Reference

A workload policy is an end-to-end solution for AI managers and administrators to control and simplify how workloads are submitted, setting best practices, enforcing limitations, and standardizing processes for AI projects within their organization.

This guide explains the policy YAML fields and the possible rules and defaults that can be set for each field.

<Info>

**Note**

This reference applies to NVIDIA Run:ai native workloads. To define policies for other supported workload types, see [Supported workload types policies](supported-workload-type-policies.md).

</Info>

## Policy YAML Fields - Reference Table

The policy fields are structured in a similar format to the workload API fields. The following tables represent a structured guide designed to help you understand and configure policies in a YAML format. It provides the fields, descriptions, defaults and rules for each workload type.

Click the link to view the value type of each field.

| Fields | Description | Value type | Supported NVIDIA Run:ai workload type |
| --- | --- | --- | --- |
| args | When set, contains the arguments sent along with the command. These override the entry point of the image in the created workload | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| command | A command to serve as the entry point of the container running the workload | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| createHomeDir | Instructs the system to create a temporary home directory for the user within the container. Data stored in this directory is not saved when the container exists. When the runAsUser flag is set to true, this flag defaults to true as well | <a href="policy-yaml-reference.md#value-types">boolean</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| environmentVariables | Set of environmentVariables to populate the container running the workload | <a href="policy-yaml-reference.md#value-types">itemized</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| image | Specifies the image to use when creating the container running the workload | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| imagePullPolicy | Specifies the pull policy of the image when starting t a container running the created workload. Options are: Always, Never, or IfNotPresent | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| imagePullSecrets | Specifies a list of references to Kubernetes secrets in the same namespace used for pulling container images. | <a href="policy-yaml-reference.md#value-types">array</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| workingDir | Container’s working directory. If not specified, the container runtime default is used, which might be configured in the container image | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| nodeType | Nodes (machines) or a group of nodes on which the workload runs | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| nodePools | A prioritized list of node pools for the scheduler to run the workload on. The scheduler always tries to use the first node pool before moving to the next one when the first is not available. | <a href="policy-yaml-reference.md#value-types">array</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| annotations | Set of annotations to populate into the container running the workload | <a href="policy-yaml-reference.md#value-types">itemized</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| labels | Set of labels to populate into the container running the workload | <a href="policy-yaml-reference.md#value-types">itemized</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| terminateAfterPreemtpion | Indicates whether the job should be terminated, by the system, after it has been preempted | <a href="policy-yaml-reference.md#value-types">boolean</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li></ul> |
| autoDeletionTimeAfterCompletionSeconds | Specifies the duration after which a finished workload (Completed or Failed) is automatically deleted. If this field is set to zero, the workload becomes eligible to be deleted immediately after it finishes. | <a href="policy-yaml-reference.md#value-types">integer</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li></ul> |
| terminationGracePeriodSeconds | The duration, in seconds, that a workload is allowed to continue running after a preemption request before it is forcibly terminated. The grace period acts as a buffer that allows the workload to reach a safe checkpoint before termination. The default value is 30 seconds and is limited by a system policy to 300 seconds (5 minutes) across all workloads. Administrators can override the default by creating a new policy at the desired scope. See <a href="policies-and-rules.md#system-policies">System policies</a> for more details. | <a href="policy-yaml-reference.md#value-types">integer</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li></ul> |
| backoffLimit | Specifies the number of retries before marking a workload as failed | <a href="policy-yaml-reference.md#value-types">integer</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li></ul> |
| restartPolicy | <p>Specify the restart policy of the workload pods. Default is empty, which is determine by the framework default</p><p>Enum: "Always" "Never" "OnFailure"</p> | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Distributed inference (API only)</li></ul> |
| cleanPodPolicy | <p>Specifies which pods will be deleted when the workload reaches a terminal state (completed/failed). The policy can be one of the following values:</p><ul><li><code>Running</code> - Only pods still running when a job completes (for example, parameter servers) will be deleted immediately. Completed pods will not be deleted so that the logs will be preserved. (Default for MPI)</li><li><code>All</code> - All (including completed) pods will be deleted immediately when the job finishes.</li><li><code>None</code> - No pods will be deleted when the job completes. It will keep running pods that consume GPU, CPU and memory over time. It is recommended to set to None only for debugging and obtaining logs from running pods. (Default for PyTorch)</li></ul> | <a href="policy-yaml-reference.md#value-types">string</a> | Distributed training |
| completions | Used with Hyperparameter Optimization. Specifies the number of successful pods the job should reach to be completed. The Job is marked as successful once the specified amount of pods has succeeded. | <a href="policy-yaml-reference.md#value-types">integer</a> | Standard training |
| parallelism | Used with Hyperparameters Optimization. Specifies the maximum desired number of pods the workload should run at any given time. | <a href="policy-yaml-reference.md#value-types">itemized</a> | Standard training |
| exposedUrls | Specifies a set of exported URLs (e.g. ingress) from the container running the created workload. | <a href="policy-yaml-reference.md#value-types">itemized</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li></ul> |
| relatedUrls | Specifies a set of URLs related to the workload. For example, a URL to an external server providing statistics or logging about the workload. | <a href="policy-yaml-reference.md#value-types">itemized</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li></ul> |
| PodAffinitySchedulingRule | Indicates if we want to use the Pod affinity rule as: the “hard” (required) or the “soft” (preferred) option. | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| podAffinityTopology | Specifies the Pod Affinity Topology to be used for scheduling the job. | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| category | Specifies the workload category assigned to the workload. Categories are used to classify and monitor different types of workloads within the NVIDIA Run:ai platform. | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| sshAuthMountPath | Specifies the directory where SSH keys are mounted | <a href="policy-yaml-reference.md#value-types">string</a> | Distributed training (MPI only) |
| mpiLauncherCreationPolicy | <p>Define s whether the MPI Launcher is created in parallel with the workers, or if its creation is postponed until all workers are in Ready state. This prevents failures when the launcher attempts to connect to workers that are not yet ready.</p><p>Enum: <code>AtStartup</code>, <code>WaitForWorkersReady</code></p> | <a href="policy-yaml-reference.md#value-types">string</a> | Distributed training (MPI only) |
| ports | Specifies a set of ports exposed from the container running the created workload. More information in <a href="policy-yaml-reference.md#ports-fields">Ports fields</a> below. | <a href="policy-yaml-reference.md#value-types">itemized</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li></ul> |
| preemptibility | Specifies whether the workload can be preempted by higher-priority workloads. Valid values are preemptible, semi-preemptible, and non-preemptible. |  | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| probes | Specifies the ReadinessProbe to use to determine if the container is ready to accept traffic. More information in <a href="policy-yaml-reference.md#probes-fields">Probes fields</a> below | - | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| tolerations | Toleration rules which apply to the pods running the workload. Toleration rules guide (but do not require) the system to which node each pod can be scheduled to or evicted from, based on matching between those rules and the set of taints defined for each Kubernetes node. | <a href="policy-yaml-reference.md#value-types">itemized</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| priorityClass | <p>Specifies the priority class of the workload. The default values are:</p><ul><li>Workspace - <code>High</code></li><li>Training / distributed training - <code>Low</code></li><li>Inference - <code>Very high</code></li></ul><p>You can change it to any of the following valid values to adjust the workload's scheduling behavior: <code>very-low</code>, <code>low</code>, <code>medium- low</code>, <code>medium</code>, <code>medium-high</code>, <code>high</code>, <code>very-high</code>.</p> | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| nodeAffinityRequired | If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to an update), the system may or may not try to eventually evict the pod from its node. | <a href="policy-yaml-reference.md#value-types">array</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| startupPolicy | <p>Determines when the worker pods should start during workload initialization.</p><ul><li><code>LeaderCreated</code>: Workers start after the leader pod is created.</li><li><code>LeaderReady</code>: Workers start only after the leader pod is ready.</li></ul><p>Default: "LeaderCreated"</p> | <a href="policy-yaml-reference.md#value-types">string</a> | Distributed inference (API only) |
| workers | Specifies the number of worker nodes to run. If set to 0, only the leader node will run, and no worker pods will be created. In this case, worker spec is not required. Default: 0 | <a href="policy-yaml-reference.md#value-types">integer</a> | Distributed inference (API only) |
| replicas | Specifies the number of leader-worker sets to deploy. Each replica represents a group consisting of one leader pod and multiple worker pods. For example, setting replicas: 3 will create 3 independent groups, each with its own leader and corresponding set of workers. Default: 1 | <a href="policy-yaml-reference.md#value-types">integer</a> | Distributed inference (API only) |
| replicas | The number of replicas to deploy. Default: 1 | <a href="policy-yaml-reference.md#value-types">integer</a> | NVIDIA NIM services (API only) |
| leader | Defines the pod specification for the leader. Must always be provided, regardless of the number of workers. | - | Distributed inference (API only) |
| worker | Defines the pod specification for the workers. Required only if the number of workers is greater than 0. | - | Distributed inference (API only) |
| multiNode | Defines whether the NIM service runs as a multi-node deployment. If workers is set to 1 or more, the service runs in multi-node. | - | NVIDIA NIM services (API only) |
| ngcAuthSecret | The name of a Kubernetes secret containing the NGC access credentials. The secret must contain a key named NGC_API_KEY with the API key as the value. | <a href="policy-yaml-reference.md#value-types">integer</a> | NVIDIA NIM services (API only) |
| storage | Contains all the fields related to storage configurations. More information in <a href="policy-yaml-reference.md#storage-fields">Storage fields</a> below. | - | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| security | Contains all the fields related to security configurations. More information in <a href="policy-yaml-reference.md#security-fields">Security fields</a> below. | - | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| compute | Contains all the fields related to compute configurations. More information in <a href="policy-yaml-reference.md#compute-fields">Compute fields </a>below. | - | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| tty | Whether this container should allocate a TTY for itself, also requires 'stdin' to be true | <a href="policy-yaml-reference.md#value-types">boolean</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li></ul> |
| stdin | Whether this container should allocate a buffer for stdin in the container runtime. If this is not set, reads from stdin in the container will always result in EOF | <a href="policy-yaml-reference.md#value-types">boolean</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li></ul> |
| numWorkers | the number of workers that will be allocated for running the workload. | <a href="policy-yaml-reference.md#value-types">integer</a> | Distributed training |
| distributedFramework | <p>The distributed training framework used in the workload.</p><p>Enum: "MPI" "PyTorch"</p> | <a href="policy-yaml-reference.md#value-types">string</a> | Distributed training |
| slotsPerWorker | Specifies the number of slots per worker used in hostfile. Defaults to 1. (applicable only for MPI) | <a href="policy-yaml-reference.md#value-types">integer</a> | Distributed training (MPI only) |
| minReplicas | The lower limit for the number of worker pods to which the training job can scale down. (applicable only for PyTorch) | <a href="policy-yaml-reference.md#value-types">integer</a> | Distributed training (PyTorch only) |
| maxReplicas | The upper limit for the number of worker pods that can be set by the autoscaler. Cannot be smaller than MinReplicas. (applicable only for PyTorch) | <a href="policy-yaml-reference.md#value-types">integer</a> | Distributed training (PyTorch only) |
| servingPort | Specifies the port for accessing the inference service. See <a href="policy-yaml-reference.md#serving-port-fields">Serving Port Fields</a>. | - | <ul><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| autoscaling | Specifies the minimum and maximum number of replicas to be scaled up and down to meet the changing demands of inference services. See <a href="policy-yaml-reference.md#autoscaling-fields">Autoscaling Fields</a>. | - | <ul><li>Inference</li><li>NVIDIA NIM services (API only)</li></ul> |
| servingConfiguration | Specifies the inference workload serving configuration. See <a href="policy-yaml-reference.md#serving-configuration-fields">Serving Configuration Fields</a>. | - | Inference |

### Ports Fields

| Fields | Description | Value type | Supported NVIDIA Run:ai workload type |
| --- | --- | --- | --- |
| container | The port that the container running the workload exposes. | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li></ul> |
| serviceType | Specifies the default service exposure method for ports. the default shall be used for ports which do not specify service type. Options are: LoadBalancer, NodePort or ClusterIP. For more information see the <a href="../../installation-and-configuration/configuration/networking/external-access-to-containers.md">External Access to Containers</a> guide. | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li></ul> |
| external | The external port which allows a connection to the container port. If not specified, the port is auto-generated by the system. | <a href="policy-yaml-reference.md#value-types">integer</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li></ul> |
| toolType | The tool type that runs on this port. | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li></ul> |
| toolName | A name describing the tool that runs on this port. | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li></ul> |

### Probes Fields

| Fields | Description | Value type | Supported NVIDIA Run:ai workload type |
| --- | --- | --- | --- |
| readiness | Specifies the Readiness Probe to use to determine if the container is ready to accept traffic. | - | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |

<Accordion title="Readiness Field Details">

**Readiness Field Details**

* **Description:** Specifies the Readiness Probe to use to determine if the container is ready to accept traffic
* **Value type:** [itemized](policy-yaml-reference.md#value-types)
* **Example policy snippet:**

```yaml
defaults:
   probes:
     readiness:
         initialDelaySeconds: 2
```

| Spec readiness fields | Description | Value type |
| --- | --- | --- |
| initialDelaySeconds | Number of seconds after the container has started before liveness or readiness probes are initiated. | [integer](policy-yaml-reference.md#value-types) |
| periodSeconds | How often (in seconds) to perform the probe | [integer](policy-yaml-reference.md#value-types) |
| timeoutSeconds | Number of seconds after which the probe times out | [integer](policy-yaml-reference.md#value-types) |
| successThreshold | Minimum consecutive successes for the probe to be considered successful after having failed | [integer](policy-yaml-reference.md#value-types) |
| failureThreshold | When a probe fails, the number of times to try before giving up | [integer](policy-yaml-reference.md#value-types) |

</Accordion>

### Security Fields

| Fields | Description | Value type | Supported NVIDIA Run:ai workload type |
| --- | --- | --- | --- |
| uidGidSource | <p>Indicates the way to determine the user and group ids of the container. The options are:</p><ul><li><code>fromTheImage</code> - user and group IDs are determined by the docker image that the container runs. This is the default option.</li><li><code>custom</code> - user and group IDs can be specified in the environment asset and/or the workspace creation request.</li><li><code>fromIdpToken</code> - user and group IDs are automatically taken from the identity provider (IdP) token (available only in SSO-enabled installations).</li></ul><p>For more information, see <a href="../../installation-and-configuration/configuration/security/user-identity-in-containers.md">User identity in containers</a>.</p> | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| capabilities | The capabilities field allows adding a set of unix capabilities to the container running the workload. Capabilities are Linux distinct privileges traditionally associated with superuser which can be independently enabled and disabled | <a href="policy-yaml-reference.md#value-types">array</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| seccompProfileType | <p>Indicates which kind of seccomp profile is applied to the container. The options are:</p><ul><li>RuntimeDefault - the container runtime default profile should be used</li><li>Unconfined - no profile should be applied</li></ul> | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| runAsNonRoot | Indicates that the container must run as a non-root user. | <a href="policy-yaml-reference.md#value-types">boolean</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| readOnlyRootFilesystem | If true, mounts the container's root filesystem as read-only. | <a href="policy-yaml-reference.md#value-types">boolean</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| runAsUid | Specifies the Unix user id with which the container running the created workload should run. | <a href="policy-yaml-reference.md#value-types">integer</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| runasGid | Specifies the Unix Group ID with which the container should run. | <a href="policy-yaml-reference.md#value-types">integer</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| supplementalGroups | Comma separated list of groups that the user running the container belongs to, in addition to the group indicated by runAsGid. | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| allowPrivilegeEscalation | Allows the container running the workload and all launched processes to gain additional privileges after the workload starts | <a href="policy-yaml-reference.md#value-types">boolean</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li></ul> |
| privileged | Grants the container full access to the host, bypassing almost all container isolation; the container acts like root. Default is false.<br/><br/>This parameter is governed by a system policy, which enforces <code>privileged: false</code> by default and marks it as non-editable (<code>canEdit: false</code>). Containers cannot run in privileged mode unless an administrator explicitly updates the system policy to allow it. See <a href="policies-and-rules.md#system-policies">System policies</a> for more details. | <a href="policy-yaml-reference.md#value-types">boolean</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li></ul> |
| hostIpc | Whether to enable hostIpc. Defaults to false. | <a href="policy-yaml-reference.md#value-types">boolean</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li></ul> |
| hostNetwork | Whether to enable host network. | <a href="policy-yaml-reference.md#value-types">boolean</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li></ul> |

### Compute Fields

| Fields | Description | Value type | Supported NVIDIA Run:ai workload type |
| --- | --- | --- | --- |
| cpuCoreRequest | CPU units to allocate for the created workload (0.5, 1, .etc). The workload receives at least this amount of CPU. Note that the workload is not scheduled unless the system can guarantee this amount of CPUs to the workload. | <a href="policy-yaml-reference.md#value-types">number</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| cpuCoreLimit | Limitations on the number of CPUs consumed by the workload (0.5, 1, .etc). The system guarantees that this workload is not able to consume more than this amount of CPUs. | <a href="policy-yaml-reference.md#value-types">number</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| cpuMemoryRequest | The amount of CPU memory to allocate for this workload (1G, 20M, .etc). The workload receives at least this amount of memory. Note that the workload is not scheduled unless the system can guarantee this amount of memory to the workload | <a href="policy-yaml-reference.md#value-types">quantity</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| cpuMemoryLimit | Limitations on the CPU memory to allocate for this workload (1G, 20M, .etc). The system guarantees that this workload is not be able to consume more than this amount of memory. The workload receives an error when trying to allocate more memory than this limit. | <a href="policy-yaml-reference.md#value-types">quantity</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| largeShmRequest | A large /dev/shm device to mount into a container running the created workload (shm is a shared file system mounted on RAM). | <a href="policy-yaml-reference.md#value-types">boolean</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| gpuRequestType | Sets the unit type for GPU resources requests to either portion or memory. Only if <code>gpuDeviceRequest = 1</code>, the request type can be stated as <code>portion</code> or <code>memory</code>. | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| gpuPortionRequest | Specifies the fraction of GPU to be allocated to the workload, between 0 and 1. For backward compatibility, it also supports the number of gpuDevices larger than 1, currently provided using the gpuDevices field. | <a href="policy-yaml-reference.md#value-types">number</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| gpuDeviceRequest | Specifies the number of GPUs to allocate for the created workload. Only if <code>gpuDeviceRequest = 1</code>, the gpuRequestType can be defined. | <a href="policy-yaml-reference.md#value-types">integer</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| gpuPortionLimit | When a fraction of a GPU is requested, the GPU limit specifies the portion limit to allocate to the workload. The range of the value is from 0 to 1. | <a href="policy-yaml-reference.md#value-types">number</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| gpuMemoryRequest | Specifies GPU memory to allocate for the created workload. The workload receives this amount of memory. Note that the workload is not scheduled unless the system can guarantee this amount of GPU memory to the workload. | <a href="policy-yaml-reference.md#value-types">quantity</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| gpuMemoryLimit | Specifies a limit on the GPU memory to allocate for this workload. Should be no less than the gpuMemory. | <a href="policy-yaml-reference.md#value-types">quantity</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| extendedResources | Specifies values for extended resources. Extended resources are third-party devices (such as high-performance NICs, FPGAs, or InfiniBand adapters) that you want to allocate to your Job. | <a href="policy-yaml-reference.md#value-types">itemized</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |

### Storage Fields

| Fields | Description | Value type | Supported NVIDIA Run:ai workload type |
| --- | --- | --- | --- |
| dataVolume | Set of data volumes to use in the workload. Each data volume is mapped to a file-system mount point within the container running the workload. | <a href="policy-yaml-reference.md#value-types">itemized</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li></ul> |
| <a href="policy-yaml-reference.md#hostpath-field-details">hostPath</a> | Maps a folder to a file-system mount point within the container running the workload. | <a href="policy-yaml-reference.md#value-types">itemized</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li></ul> |
| <a href="policy-yaml-reference.md#git-field-details">git</a> | Details of the git repository and items mapped to it. | <a href="policy-yaml-reference.md#value-types">itemized</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li></ul> |
| <a href="policy-yaml-reference.md#pvc-field-details">pvc</a> | Specifies persistent volume claims to mount into a container running the created workload. | <a href="policy-yaml-reference.md#value-types">itemized</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| <a href="policy-yaml-reference.md#nfs-field-details">nfs</a> | Specifies NFS volume to mount into the container running the workload. | <a href="policy-yaml-reference.md#value-types">itemized</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li></ul> |
| <a href="policy-yaml-reference.md#s3-field-details">s3</a> | Specifies S3 buckets to mount into the container running the workload. | <a href="policy-yaml-reference.md#value-types">itemized</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li></ul> |
| configMapVolumes | Specifies ConfigMaps to mount as volumes into a container running the created workload. | <a href="policy-yaml-reference.md#value-types">itemized</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| secretVolume | Set of secret volumes to use in the workload. A secret volume maps a secret resource in the cluster to a file-system mount point within the container running the workload. | <a href="policy-yaml-reference.md#value-types">itemized</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| emptyDir | A list of emptyDir volumes to mount in the workload. | <a href="policy-yaml-reference.md#value-types">itemized</a> | <ul><li>Workspace</li><li>Standard training</li><li>Distributed training</li><li>Inference</li><li>Distributed inference (API only)</li></ul> |

#### Storage Field Examples

<Accordion title="hostPath Field Details">

**hostPath Field Details**

* **Description:** Maps a folder to a file system mount oint within the container running the workload
* **Value type:** [itemized](policy-yaml-reference.md#value-types)
* **Example policy snippet:**

```yaml
defaults:
  storage:
    hostPath:
      instances:
        - path: h3-path-1
          mountPath: h3-mount-1
        - path: h3-path-2
          mountPath: h3-mount-2
      attributes:
        - readOnly: true
```

| hostPath fields | Description | Value type |
| --- | --- | --- |
| name | Unique name to identify the instance. Primarily used for policy locked rules. | <a href="policy-yaml-reference.md#value-types">string</a> |
| path | Local path within the controller to which the host volume is mapped. | <a href="policy-yaml-reference.md#value-types">string</a> |
| readOnly | Force the volume to be mounted with read-only permissions. Defaults to false | <a href="policy-yaml-reference.md#value-types">boolean</a> |
| mountPath | <p>The path that the host volume is mounted to when in use. Enum:</p><ul><li>"None"</li><li>"HostToContainer"</li></ul> | <a href="policy-yaml-reference.md#value-types">string</a> |
| mountPropagation | Share this volume mount with other containers. If set to HostToContainer, this volume mount receives all subsequent mounts that are mounted to this volume or any of its subdirectories. In case of multiple hostPath entries, this field should have the same value for all of them. | <a href="policy-yaml-reference.md#value-types">string</a> |

</Accordion>

<Accordion title="Git Field Details">

**Git Field Details**

* **Description:** Details of the git repository and items mapped to it
* **Value type:** [itemized](policy-yaml-reference.md#value-types)
* **Example policy snippet:**

```yaml
defaults:
  storage:
    git:
      attributes:
        Repository: https://runai.public.github.com
      instances
        - branch: "master"
          path: /container/my-repository
          passwordSecret: my-password-secret
```

| Git fields | Description | Value type |
| --- | --- | --- |
| repository | URL to a remote git repository. The content of this repository is mapped to the container running the workload | [string](policy-yaml-reference.md#value-types) |
| revision | Specific revision to synchronize the repository from | [string](policy-yaml-reference.md#value-types) |
| path | Local path within the workspace to which the S3 bucket is mapped | [string](policy-yaml-reference.md#value-types) |
| secretName | Optional name of Kubernetes secret that holds your git username and password | [string](policy-yaml-reference.md#value-types) |
| username | If secretName is provided, this field should contain the key, within the provided Kubernetes secret, which holds the value of your git username. Otherwise, this field should specify your git username in plain text (example: myuser). | [string](policy-yaml-reference.md#value-types) |

</Accordion>

<Accordion title="PVC Field Details">

**PVC Field Details**

* **Description:** Specifies persistent volume claims to mount into a container running the created workload
* **Value type:** [itemized](policy-yaml-reference.md#value-types)
* **Example policy snippet:**

```yaml
defaults:
  storage:
    pvc:
      instances:
        - claimName: pvc-staging-researcher1-home
          existingPvc: true
          path: /myhome
          readOnly: false
          claimInfo:
            accessModes:
              readWriteMany: true
```

| Spec PVC fields | Description | Value type |
| --- | --- | --- |
| claimName (mandatory) | A given name for the PVC. Allowed referencing it across workspaces | [string](policy-yaml-reference.md#value-types) |
| ephemeral | Use **true** to set PVC to ephemeral. If set to **true**, the PVC is deleted when the workspace is stopped. | [boolean](policy-yaml-reference.md#value-types) |
| path | Local path within the workspace to which the PVC bucket is mapped | [string](policy-yaml-reference.md#value-types) |
| readonly | Permits read only from the PVC, prevents additions or modifications to its content | [boolean](policy-yaml-reference.md#value-types) |
| ReadwriteOnce | Requesting claim that can be mounted in read/write mode to exactly 1 host. If none of the modes are specified, the default is readWriteOnce. | [boolean](policy-yaml-reference.md#value-types) |
| size | Requested size for the PVC. Mandatory when existing PVC is false | [string](policy-yaml-reference.md#value-types) |
| storageClass | Storage class name to associate with the PVC. This parameter may be omitted if there is a single storage class in the system, or you are using the default storage class. Further details at [Kubernetes storage classes](https://kubernetes.io/docs/concepts/storage/storage-classes.). | [string](policy-yaml-reference.md#value-types) |
| readOnlyMany | Requesting claim that can be mounted in read-only mode to many hosts | [boolean](policy-yaml-reference.md#value-types) |
| readWriteMany | Requesting claim that can be mounted in read/write mode to many hosts | [boolean](policy-yaml-reference.md#value-types) |

</Accordion>

<Accordion title="NFS Field Details">

**NFS Field Details**

* **Description:** Specifies NFS volume to mount into the container running the workload
* **Value type:** [itemized](policy-yaml-reference.md#value-types)
* **Example policy snippet:**

```yaml
defaults:
 storage:
   nfs:
     instances:
       - path: nfs-path
         readOnly: true
         server: nfs-server
         mountPath: nfs-mount
rules:
  storage:
    nfs:
      instances:
        canAdd: false
```

| nfs fields | Description | Value type |
| --- | --- | --- |
| mountPath | The path that the NFS volume is mounted to when in use | [string](policy-yaml-reference.md#value-types) |
| path | Path that is exported by the NFS server | [string](policy-yaml-reference.md#value-types) |
| readOnly | Whether to force the NFS export to be mounted with read-only permissions | [boolean](policy-yaml-reference.md#value-types) |
| nfsServer | The hostname or IP address of the NFS server | [string](policy-yaml-reference.md#value-types) |

</Accordion>

<Accordion title="S3 Field Details">

**S3 Field Details**

* **Description:** Specifies S3 buckets to mount into the container running the workload
* **Value type:** [itemized](policy-yaml-reference.md#value-types)
* **Example policy snippet:**

```yaml
defaults:
  storage:
    s3:
      instances:
        - bucket: bucket-opt-1
          path: /s3/path
          accessKeySecret: s3-access-key
          secretKeyOfAccessKeyId: s3-secret-id
          secretKeyOfSecretKey: s3-secret-key
      attributes:
        url: https://amazonaws.s3.com
```

| s3 fields | Description | Value type |
| --- | --- | --- |
| Bucket | The name of the bucket | [string](policy-yaml-reference.md#value-types) |
| path | Local path within the workspace to which the S3 bucket is mapped | [string](policy-yaml-reference.md#value-types) |
| url | The URL of the S3 service provider. The default is the URL of the Amazon AWS S3 service | [string](policy-yaml-reference.md#value-types) |

</Accordion>

### Serving Port Fields

| Fields | Description | Value type | Supported NVIDIA Run:ai workload type |
| --- | --- | --- | --- |
| container / port | Specifies the port that the container running the inference service exposes | <a href="policy-yaml-reference.md#value-types">integer</a> | <ul><li>Inference</li><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| protocol | <p>Specifies the protocol used by the port. Defaults to http.</p><p>Enum: "http", "grpc"</p> | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| authorizationType | <p>Specifies the authorization type for serving port URL access. Defaults to public, which means no authorization is required. If set to authenticatedUsers, only authenticated NVIDIA Run:ai users are allowed to access the URL. If set to authorizedUsersOrGroups, only users or groups specified in authorizedUsers or authorizedGroups are allowed to access the URL. Supported from cluster version 2.19.</p><p>Enum: "public", "authenticatedUsers", "authorizedUsersOrGroups"</p> | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| authorizedUsers | Specifies the list of users that are allowed to access the URL. Note that authorizedUsers and authorizedGroups are mutually exclusive. | <a href="policy-yaml-reference.md#value-types">array</a> | <ul><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| authorizedGroups | Specifies the list of groups that are allowed to access the URL. Note that authorizedUsers and authorizedGroups are mutually exclusive. | <a href="policy-yaml-reference.md#value-types">array</a> | <ul><li>Inference</li><li>Distributed inference (API only)</li></ul> |
| clusterLocalAccessOnly | Configures the serving port URL to be available only on the cluster-local network, and not externally. Defaults to false. | <a href="policy-yaml-reference.md#value-types">boolean</a> | Inference |
| exposeExternally | Indicates whether the inference serving endpoint should be accessible outside the cluster. If set to true, the endpoint will be exposed externally. To enable external access, your administrator must configure the cluster as described in the <a href="../../installation-and-configuration/installation/install-using-helm/prerequisites-and-features.md#standard-inference">inference requirements</a> section. | <a href="policy-yaml-reference.md#value-types">boolean</a> | <ul><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| exposedUrl | The custom URL to use for the serving port. If empty (default), an autogenerated URL will be used. | <a href="policy-yaml-reference.md#value-types">string</a> | <ul><li>Distributed inference (API only)</li><li>NVIDIA NIM services (API only)</li></ul> |
| serviceType | The type of Kubernetes service to create for the inference deployment. Options include 'ClusterIP' (default), 'NodePort', 'LoadBalancer', and 'ExternalName'. Default: "ClusterIP" | <a href="policy-yaml-reference.md#value-types">string</a> | NVIDIA NIM services (API only) |
| grpcPort | The GRPC port that the container running the inference service exposes. | <a href="policy-yaml-reference.md#value-types">integer</a> | NVIDIA NIM services (API only) |
| metricsPort | The port where metrics are exposed, required only if it's different than the main port. | <a href="policy-yaml-reference.md#value-types">integer</a> | NVIDIA NIM services (API only) |
| exposedProtocol | The protocol to use for the exposed URL. If grpcPort is set, this defaults to grpc. Otherwise, it defaults to http. Enum: "http" "grpc" | <a href="policy-yaml-reference.md#value-types">string</a> | NVIDIA NIM services (API only) |

### Autoscaling Fields

| Fields | Description | Value type | Supported NVIDIA Run:ai workload type |
| --- | --- | --- | --- |
| metricThresholdPercentage | Specifies the percentage of metric threshold value to use for autoscaling. Defaults to 70. Applicable only with the 'throughput' and 'concurrency' metrics. | <a href="policy-yaml-reference.md#value-types">number</a> | Inference |
| minReplicas | Specifies the minimum number of replicas for autoscaling. Defaults to 1. Use 0 to allow scale-to-zero. | <a href="policy-yaml-reference.md#value-types">integer</a> | <ul><li>Inference</li><li>NVIDIA NIM services (API only)</li></ul> |
| maxReplicas | Specifies the maximum number of replicas for autoscaling. Defaults to minReplicas, or to 1 if minReplicas is set to 0. | <a href="policy-yaml-reference.md#value-types">integer</a> | <ul><li>Inference</li><li>NVIDIA NIM services (API only)</li></ul> |
| initialReplicas | Specifies the number of replicas to run when initializing the workload for the first time. Defaults to minReplicas, or to 1 if minReplicas is set to 0. | <a href="policy-yaml-reference.md#value-types">integer</a> | Inference |
| activationReplicas | Specifies the number of replicas to run when scaling-up from zero. Defaults to minReplicas, or to 1 if minReplicas is set to 0. | <a href="policy-yaml-reference.md#value-types">integer</a> | Inference |
| concurrencyHardLimit | Specifies the maximum number of requests allowed to flow to a single replica at any time. 0 means no limit. | <a href="policy-yaml-reference.md#value-types">integer</a> | Inference |
| scaleToZeroRetentionSeconds | Specifies the minimum amount of time (in seconds) that the last replica will remain active after a scale-to-zero decision. Defaults to 0. Available only if minReplicas is set to 0. | <a href="policy-yaml-reference.md#value-types">integer</a> | Inference |
| scaleDownDelaySeconds | Specifies the minimum amount of time (in seconds) that a replica will remain active after a scale-down decision | <a href="policy-yaml-reference.md#value-types">integer</a> | Inference |
| scaleWindowSeconds | The time window for autoscaling decisions, in seconds. Defaults to 300 seconds. | <a href="policy-yaml-reference.md#value-types">integer</a> | NVIDIA NIM services (API only) |
| metric | Specifies the metric to use for autoscaling. Mandatory if minReplicas &#x3C; maxReplicas, except for the special case where minReplicas is set to 0 and maxReplicas is set to 1, as in this case autoscaling decisions are made according to network activity rather than metrics. Use one of the built-in metrics of 'throughput', 'concurrency' or 'latency', or any other available custom metric. Only the 'throughput' and 'concurrency' metrics support scale-to-zero. | <a href="policy-yaml-reference.md#value-types">string</a> | Inference |
| metricThreshold | Specifies the threshold to use with the specified metric for autoscaling. Mandatory if metric is specified. | <a href="policy-yaml-reference.md#value-types">integer</a> | <ul><li>Inference</li><li>NVIDIA NIM services (API only)</li></ul> |

### Serving Configuration Fields

| Fields | Description | Value type | Supported NVIDIA Run:ai workload type |
| --- | --- | --- | --- |
| initializationTimeoutSeconds | Specifies the maximum time (in seconds) allowed for a workload to initialize and become ready. If the workload does not start within this time, it will be moved to failed state. | [integer](policy-yaml-reference.md#value-types) | Inference |
| requestTimeoutSeconds<br/> | Specifies the maximum time (in seconds) allowed to process an end-user request. If no response is returned within this time, the request will be ignored. | [integer](policy-yaml-reference.md#value-types) | Inference |

## Value Types

Each field has a specific value type. The following value types are supported.

| Value type | Description | Supported rule type | Defaults |
| --- | --- | --- | --- |
| Boolean | A binary value that can be either True or False | <ul><li><a href="policy-yaml-reference.md#rule-types">canEdit</a></li><li><a href="policy-yaml-reference.md#rule-types">required</a></li></ul> | true/false |
| String | A sequence of characters used to represent text. It can include letters, numbers, symbols, and spaces | <ul><li><a href="policy-yaml-reference.md#rule-types">canEdit</a></li><li><a href="policy-yaml-reference.md#rule-types">required</a></li><li><a href="policy-yaml-reference.md#rule-types">options</a></li></ul> | abc |
| Itemized | An ordered collection of items (objects), which can be of different types (all items in the list are of the same type). For further information see the chapter below the table. | <ul><li><a href="policy-yaml-reference.md#rule-types">canAdd</a></li><li><a href="policy-yaml-reference.md#rule-types">locked</a></li></ul> | See below |
| Integer | An Integer is a whole number without a fractional component. | <ul><li><a href="policy-yaml-reference.md#rule-types">canEdit</a></li><li><a href="policy-yaml-reference.md#rule-types">required</a></li><li><a href="policy-yaml-reference.md#rule-types">min</a></li><li><a href="policy-yaml-reference.md#rule-types">max</a></li><li><a href="policy-yaml-reference.md#rule-types">step</a></li><li><a href="policy-yaml-reference.md#rule-types">defaultFrom</a></li></ul> | 100 |
| Number | Capable of having non-integer values | <ul><li><a href="policy-yaml-reference.md#rule-types">canEdit</a></li><li><a href="policy-yaml-reference.md#rule-types">required</a></li><li><a href="policy-yaml-reference.md#rule-types">min</a></li><li><a href="policy-yaml-reference.md#rule-types">defaultFrom</a></li></ul> | 10.3 |
| Quantity | Holds a string composed of a number and a unit representing a quantity | <ul><li><a href="policy-yaml-reference.md#rule-types">canEdit</a></li><li><a href="policy-yaml-reference.md#rule-types">required</a></li><li><a href="policy-yaml-reference.md#rule-types">min</a></li><li><a href="policy-yaml-reference.md#rule-types">max</a></li><li><a href="policy-yaml-reference.md#rule-types">defaultFrom</a></li></ul> | 5M |
| Array | Set of values that are treated as one, as opposed to Itemized in which each item can be referenced separately. | <ul><li><a href="policy-yaml-reference.md#rule-types">canEdit</a></li><li><a href="policy-yaml-reference.md#rule-types">required</a></li></ul> | <ul><li>node-a</li><li>node-b</li><li>node-c</li></ul> |

## Itemized

Workload fields of type itemized have multiple instances, however in comparison to objects, each can be referenced by a key field. The key field is defined for each field.

Consider the following workload spec:

```yaml
spec:
  image: ubuntu
  compute:
    extendedResources:
      - resource: added/cpu
        quantity: 10
      - resource: added/memory
        quantity: 20M
```

In this example, extendedResources have two instances, each has two attributes: resource (the key attribute) and quantity.

In policy, the defaults and rules for itemized fields have two sub sections:

* Instances: default items to be added to the policy or rules which apply to an instance as a whole.
* Attributes: defaults for attributes within an item or rules which apply to attributes within each item.

Consider the following example:

```yaml
defaults:
  compute:
    extendedResources:
      instances: 
        - resource: default/cpu
          quantity: 5
        - resource: default/memory
          quantity: 4M
      attributes:
        quantity: 3
rules:
  compute:
    extendedResources:
      instances:
        locked: 
          - default/cpu
      attributes:
        quantity: 
          required: true
```

Assume the following workload submission is requested:

```yaml
spec:
  image: ubuntu
  compute:
    extendedResources:
      - resource: default/memory
        exclude: true
      - resource: added/cpu
      - resource: added/memory
        quantity: 5M
```

The effective policy for the above mentioned workload has the following extendedResources instances:

| Resource | Source of the instance | Quantity | Source of the attribute quantity |
| --- | --- | --- | --- |
| default/cpu | Policy defaults | 5 | The default of this instance in the policy defaults section |
| added/cpu | Submission request | 3 | The default of the quantity attribute from the attributes section |
| added/memory | Submission request | 5M | Submission request |

<Info>

**Note**

The default/memory is not populated to the workload, this is because it has been excluded from the workload using “exclude: true”.

</Info>

A workload submission request cannot exclude the default/cpu resource, as this key is included in the locked rules under the instances section.

## Rule Types

| Rule types | Description | Supported value types |
| --- | --- | --- |
| canAdd | Whether the submission request can add items to an itemized field other than those listed in the policy defaults for this field. | <a href="policy-yaml-reference.md#value-types">itemized</a> |
| locked | Set of items that the workload is unable to modify or exclude. In this example, a workload policy default is given to HOME and USER, that the submission request cannot modify or exclude from the workload. | <a href="policy-yaml-reference.md#value-types">itemized</a> |
| blocked | Blocks the field to prevent workloads from specifying any values for it. | <ul><li><a href="policy-yaml-reference.md#value-types">string</a></li><li><a href="policy-yaml-reference.md#value-types">boolean</a></li><li><a href="policy-yaml-reference.md#value-types">integer</a></li><li><a href="policy-yaml-reference.md#value-types">number</a></li><li><a href="policy-yaml-reference.md#value-types">quantity</a></li><li><a href="policy-yaml-reference.md#value-types">array</a></li></ul> |
| canEdit | Whether the submission request can modify the policy default for this field. In this example, it is assumed that the policy has default for imagePullPolicy. As canEdit is set to false, submission requests are not able to alter this default. | <ul><li><a href="policy-yaml-reference.md#value-types">string</a></li><li><a href="policy-yaml-reference.md#value-types">boolean</a></li><li><a href="policy-yaml-reference.md#value-types">integer</a></li><li><a href="policy-yaml-reference.md#value-types">number</a></li><li><a href="policy-yaml-reference.md#value-types">quantity</a></li><li><a href="policy-yaml-reference.md#value-types">array</a></li></ul> |
| required | When set to true, the workload must have a value for this field. The value can be obtained from policy defaults. If no value specified in the policy defaults, a value must be specified for this field in the submission request. | <ul><li><a href="policy-yaml-reference.md#value-types">string</a></li><li><a href="policy-yaml-reference.md#value-types">boolean</a></li><li><a href="policy-yaml-reference.md#value-types">integer</a></li><li><a href="policy-yaml-reference.md#value-types">number</a></li><li><a href="policy-yaml-reference.md#value-types">quantity</a></li><li><a href="policy-yaml-reference.md#value-types">array</a></li></ul> |
| min | The minimal value for the field | <ul><li><a href="policy-yaml-reference.md#value-types">integer</a></li><li><a href="policy-yaml-reference.md#value-types">number</a></li><li><a href="policy-yaml-reference.md#value-types">quantity</a></li></ul> |
| max | The maximal value for the field | <ul><li><a href="policy-yaml-reference.md#value-types">integer</a></li><li><a href="policy-yaml-reference.md#value-types">number</a></li><li><a href="policy-yaml-reference.md#value-types">quantity</a></li></ul> |
| step | The allowed gap between values for this field. In this example the allowed values are: 1, 3, 5, 7 | <ul><li><a href="policy-yaml-reference.md#value-types">integer</a></li><li><a href="policy-yaml-reference.md#value-types">number</a></li></ul> |
| options | Set of allowed values for this field | <a href="policy-yaml-reference.md#value-types">string</a> |
| defaultFrom | Set a default value for a field that will be calculated based on the value of another field | <ul><li><a href="policy-yaml-reference.md#value-types">integer</a></li><li><a href="policy-yaml-reference.md#value-types">number</a></li><li><a href="policy-yaml-reference.md#value-types">quantity</a></li></ul> |

### Rule Type Examples

<Accordion title="canAdd">

**canAdd**

```yaml
storage:
  hostPath:
     instances:
       canAdd: false
```

</Accordion>

<Accordion title="locked">

**locked**

```yaml
storage:
  hostPath:
    Instances:
      locked:
        - HOME
        - USER
```

</Accordion>

<Accordion title="canEdit">

**canEdit**

```yaml
imagePullPolicy:
    canEdit: false
```

</Accordion>

<Accordion title="required">

**required**

```yaml
image:
    required: true
```

</Accordion>

<Accordion title="min">

**min**

```yaml
compute:
  gpuDevicesRequest:
    min: 3
```

</Accordion>

<Accordion title="max">

**max**

```yaml
compute:
  gpuMemoryRequest:
     max: 2G
```

</Accordion>

<Accordion title="step">

**step**

```yaml
compute:
  cpuCoreRequest:
    min: 1
    max: 7
    Step: 2
```

</Accordion>

<Accordion title="options">

**options**

```yaml
image:
  options:
    - value: image-1
    - value: image-2
```

</Accordion>

<Accordion title="defaultFrom">

**defaultFrom**

```yaml
cpuCoreRequest:
  defaultFrom:
    field: compute.cpuCoreLimit
    factor: 0.5
```

</Accordion>

## Policy Spec Sections

For each field of a specific policy, you can specify both rules and defaults. A policy spec consists of the following sections:

* Rules
* Defaults
* Imposed Assets

### Rules

Rules set up constraints on workload policy fields. For example, consider the following policy:

```yaml
rules:
  compute:
    gpuDevicesRequest: 
      max: 8
  security:
    runAsUid: 
      min: 500
```

Such a policy restricts the maximum value for gpuDeviceRequests to 8, and the minimal value for runAsUid, provided in the security section to 500.

### Defaults

The defaults section is used for providing defaults for various workload fields. For example, consider the following policy:

```yaml
defaults:
  imagePullPolicy: Always
  security:
    runAsNonRoot: true
    runAsUid: 500
```

Assume a submission request with the following values:

* Image: ubuntu
* runAsUid: 501

The effective workload that runs has the following set of values:

| Field                 | Value  | Source             |
| --------------------- | ------ | ------------------ |
| Image                 | Ubuntu | Submission request |
| ImagePullPolicy       | Always | Policy defaults    |
| security.runAsNonRoot | true   | Policy defaults    |
| security.runAsUid     | 501    | Submission request |

<Info>

**Note**

It is possible to specify a rule for each field, which states if a submission request is allowed to change the policy default for that given field, for example:

```yaml
defaults:
imagePullPolicy: Always
security:
    runAsNonRoot: true
    runAsUid: 500
 rules:
 security:
    runAsUid:
    canEdit: false
```

If this policy is applied, the submission request above fails, as it attempts to change the value of secuirty.runAsUid from 500 (the policy default) to 501 (the value provided in the submission request), which is forbidden due to canEdit rule set to false for this field.

</Info>

### Imposed Assets

Default instances of a storage field can be provided using a datasource containing the details of this storage instance. To add such instances in the policy, specify those asset IDs in the imposedAssets section of the policy.

```yaml
defaults: null
rules: null
imposedAssets:
  - f12c965b-44e9-4ff6-8b43-01d8f9e630cc
```

Assets with references to credential assets (for example: private S3, containing reference to an AccessKey asset) cannot be used as imposedAssets.
