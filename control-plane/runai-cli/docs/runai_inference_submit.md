## runai inference submit

submit an inference workload

```
runai inference submit [flags]
```

### Examples

```

# Submit a workload with scale to zero
runai inference submit <workload-name> -p <project-name> -i ghcr.io/knative/helloworld-go --gpu-devices-request 1 
--serving-port=8080 --min-replicas=0 --max-replicas=1 

# Submit a workload with arguments
runai inference submit <workload-name> -p <project-name> -i vllm/vllm-openai:latest --gpu-devices-request 1 --serving-port=8000 -- --model Qwen/Qwen3-0.6B

# Submit a workload with a template
runai inference submit <workload-name> -p <project-name> --template <template-name>

# Submit a workload with an asset
runai inference submit <workload-name> -p <project-name> --environment <environment-asset-name>

# Submit a workload with autoscaling and authorization
runai inference submit <workload-name> -p <project-name> -i ghcr.io/knative/helloworld-go  --gpu-devices-request 1
--serving-port=container=8080,authorization-type=authorizedUsersOrGroups,authorized-users=user1:user2:app1,protocol=http 
--min-replicas=1 --max-replicas=4 --metric=concurrency  --metric-threshold=100 
```

### Options

```
      --activation-replicas int32                  The number of replicas to run when scaling-up from zero. Defaults to minReplicas, or to 1 if minReplicas is set to 0
      --annotation stringArray                     Set of annotations to populate into the container running the workload
      --capability stringArray                     Add POSIX capabilities to the container. Defaults to the runtime's default set.
  -c, --command                                    If true, override the image's entrypoint with the command supplied after '--'
      --compute string                             Specifies the compute resource asset to use for the workload.
      --concurrency-hard-limit int32               The maximum number of requests allowed to flow to a single replica at any time. 0 means no limit
      --configmap-map-volume stringArray           Mount a ConfigMap as a volume. Format: name=CONFIGMAP_NAME,path=PATH,subpath=SUBPATH,default-mode=DEFAULT_MODE.
      --cpu-core-limit positiveFloat               Maximum number of CPU cores allowed (e.g. 0.5, 1).
      --cpu-core-request positiveFloat             Number of CPU cores to request (e.g. 0.5, 1).
      --cpu-memory-limit string                    Maximum memory allowed (e.g. 1G, 500M).
      --cpu-memory-request string                  Amount of memory to request (e.g. 1G, 500M).
      --create-home-dir                            Create a temporary home directory for the container. Defaults to true when --run-as-user is set, false otherwise.
      --datasource stringArray                     Specifies the data source asset to attach to the workload. Format: type=hostPath|nfs|pvc|git|s3|configMap|secretVolume,name=NAME
      --datavolume stringArray                     Mount a data volume. Format: name=DATA_VOLUME_NAME,mountpath=MOUNT_PATH.
      --empty-dir-volume stringArray               Mount an empty directory as a volume. Format: name=NAME,path=PATH,medium=MEDIUM,size-limit=SIZE.
      --env-my-credentials stringArray             Set an environment variable from a user credential. Format: type=<dockerRegistry|genericSecret|ngcApiKey>,name=<ENV_VAR>,credential-name=<CREDENTIAL_NAME>[,key=<KEY>]. Key is required only for type=genericSecret. Requires cluster version 2.22+ (2.23+ for ngcApiKey).
      --env-pod-field-ref stringArray              Set an environment variable from a pod field reference. Format: ENV_VARIABLE=FIELD_REFERENCE.
      --env-secret stringArray                     Set an environment variable from a Kubernetes secret. Format: ENV_VARIABLE=secret-name,key=secret-key.
      --environment stringArray                    Specifies the environment asset to use for the workload.
  -e, --environment-variable stringArray           Set environment variables in the container. Format: --environment-variable name=value --environment-variable name-b=value-b.
      --exclude-node stringArray                   Nodes that will be excluded from use by the scheduler. Format: --exclude-node node-a --exclude-node node-b
      --existing-pvc stringArray                   Mount an existing PersistentVolumeClaim. Format: claimname=CLAIM_NAME,path=PATH. Auto-complete supported.
      --extended-resource stringArray              Request access to a Kubernetes extended resource. Format: resource_name=quantity.
      --external-url stringArray                   Expose a URL from the workload container. Format: container=PORT,url=https://external.runai.com,authusers=user1,authgroups=group1.
      --git-sync stringArray                       Mount a Git repository into the container. Format: name=NAME,repository=REPO,path=PATH,secret=SECRET,rev=REVISION.
      --gpu-compute-mode string                    GPU compute sharing mode. Valid values: time-slicing, sm-sharing. Requires cluster version 3.0.14+.
  -g, --gpu-devices-request positiveInt            Number of GPU devices to allocate for the workload (e.g. 1, 2).
      --gpu-memory-limit string                    Maximum GPU memory to allocate (e.g. 1G, 500M).
      --gpu-memory-portion-limit positiveFloat     Maximum GPU memory fraction allowed per device (between 0 and 1).
      --gpu-memory-portion-request positiveFloat   GPU memory fraction to allocate per device (between 0 and 1, e.g. 0.5).
      --gpu-memory-request string                  Amount of GPU memory to allocate (e.g. 1G, 500M).
      --gpu-request-type string                    Type of GPU request: portion, memory
  -h, --help                                       help for submit
      --host-ipc                                   Enable host IPC for the container. Default: false.
      --host-network                               Enable host networking for the container. Default: false.
      --host-path stringArray                      Mount a host path as a volume. Format: path=PATH,mount=MOUNT,mount-propagation=None|HostToContainer,readwrite.
  -i, --image string                               The container image to use for the workload.
      --image-pull-my-credentials stringArray      Use a user credential for authenticating image pulls. Format: type=<dockerRegistry|ngcApiKey>,name=<NAME>. Requires cluster version 2.22+ (dockerRegistry) or 2.23+ (ngcApiKey).
      --image-pull-policy string                   Image pull policy for the container. Valid values: Always, IfNotPresent, Never.
      --initial-replicas int32                     The number of replicas to run when initializing the workload for the first time. Defaults to minReplicas, or to 1 if minReplicas is set to 0
      --initialization-timeout-seconds int32       The maximum amount of time (in seconds) to wait for the container to become ready
      --label stringArray                          Set of labels to populate into the container running the workspace
      --large-shm                                  Request a large /dev/shm device to mount in the container. Useful for memory-intensive workloads.
      --max-replicas int32                         The maximum number of replicas for autoscaling. Defaults to minReplicas, or to 1 if minReplicas is set to 0
      --metric string                              Autoscaling metric is required if minReplicas < maxReplicas, except when minReplicas = 0 and maxReplicas = 1. Use 'throughput', 'concurrency', 'latency', or custom metrics.
      --metric-threshold int32                     The threshold to use with the specified metric for autoscaling. Mandatory if metric is specified
      --metric-threshold-percentage float32        The percentage of metric threshold value to use for autoscaling. Defaults to 70. Applicable only with the 'throughput' and 'concurrency' metrics
      --min-replicas int32                         The minimum number of replicas for autoscaling. Defaults to 1. Use 0 to allow scale-to-zero
      --name-prefix string                         Set defined prefix for the workload name and add index as a suffix
      --new-pvc stringArray                        Create and mount a new volume. This volume is used only for the duration of the workload's lifecycle. Format: claimname=CLAIM_NAME,storageclass=STORAGE_CLASS,size=SIZE,path=PATH,accessmode-rwo,accessmode-rom,accessmode-rwm,ro,ephemeral.
      --nfs stringArray                            Mount an NFS volume. Format: path=PATH,server=SERVER,mountpath=MOUNT_PATH,readwrite.
      --node-pools stringArray                     Node pools to use for scheduling the job, ordered by priority. Format: --node-pools pool-a --node-pools pool-b
      --node-type string                           Enforce node type affinity by setting a node-type label.
      --pod-running-timeout duration               Timeout for pod to reach running state (e.g. 5s, 2m, 3h).
      --port stringArray                           Expose ports from the workload container. Format: service-type=NodePort,container=80,external=8080.
      --preemptibility preemptibility              Specify whether the workload can be preempted by higher-priority workloads. Valid values: preemptible, semi-preemptible, non-preemptible. Overrides the default preemptibility for the workload type.
      --preferred-pod-topology-key string          If possible, schedule all pods of this workload within the same topology domain defined by a Kubernetes node label key. Format: KEY.
      --priority string                            Sets the workload’s scheduling priority. Valid values: very-low, low, medium-low, medium, medium-high, high, very-high. 
                                                   Overrides the default priority for the workload type. Changing priority does not update preemptibility automatically.
      --privileged                                 Grants the container full access to the host, bypassing almost all container isolation; the container acts like root.
  -p, --project string                             Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default.
      --request-timeout-seconds int32              The maximum time (in seconds) allowed to process an end-user request. If no response is returned within this time, the request will be ignored
      --required-pod-topology-key string           Require scheduling pods of this workload within the same topology domain defined by a Kubernetes node label key. Format: KEY.
      --run-as-gid int                             Group ID to run the container as.
      --run-as-non-root                            Force the container to run as a non-root user.
      --run-as-uid int                             User ID to run the container as.
      --run-as-user                                Set the user and group IDs for the container. Uses local terminal credentials if not specified.
      --scale-down-delay-seconds int32             The minimum amount of time (in seconds) that a replica will remain active after a scale-down decision
      --scale-to-zero-retention-seconds int32      The minimum amount of time (in seconds) that the last replica will remain active after a scale-to-zero decision. Defaults to 0. Available only if minReplicas is set to 0
      --seccomp-profile string                     Seccomp profile for the container. Valid values: RuntimeDefault, Unconfined, or Localhost.
      --secret-volume stringArray                  Mount a Kubernetes Secret as a volume. Format: path=PATH,name=SECRET_RESOURCE_NAME.
      --serving-port string                        Defines various attributes for the serving port. Usage formats: (1) Simplified format: --serving-port=CONTAINER_PORT (2) Full format: --serving-port=container=CONTAINER_PORT,[authorization-type=public|authenticatedUsers|authorizedUsersOrGroups],[authorized-users=USER1:USER2:APP1...],[authorized-groups=GROUP1:GROUP2...],[cluster-local-access-only],[protocol=http|grpc]
      --supplemental-groups ints                   Comma-separated list of group IDs for the container user.
      --template string                            The name of the template to use for submitting this workload. Values from the template and CLI flags are combined. Flags with matching keys replace template values; other flags are applied in addition.
      --toleration stringArray                     Add Kubernetes tolerations. Format: operator=Equal|Exists,key=KEY,[value=VALUE],[effect=NoSchedule|NoExecute|PreferNoSchedule],[seconds=SECONDS].
      --user-group-source string                   How to determine user/group IDs. Valid values: fromTheImage, fromIdpToken.
      --wait-for-submit duration                   How long to wait for the workload to be created in the cluster. Default: 1m.
      --working-dir string                         Working directory inside the container. Overrides the default working directory set in the image.
```

### Options inherited from parent commands

```
      --config-file string   config file name; can be set by environment variable RUNAI_CLI_CONFIG_FILE (default "config.json")
      --config-path string   config path; can be set by environment variable RUNAI_CLI_CONFIG_PATH
  -d, --debug                enable debug mode
  -q, --quiet                enable quiet mode, suppress all output except error messages
      --verbose              enable verbose mode
```

### SEE ALSO

* [runai inference](runai_inference.md)	 - inference management

