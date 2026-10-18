## runai inference distributed submit

submit a distributed inference workload

### Synopsis

Before using the flags, keep in mind:

By default (--worker-as-leader=true), worker pods inherit their configuration from the leader:
- You only need to specify worker-specific flags if you want worker pods to differ from the leader
- When --worker-as-leader=false is set, workers require their own explicit --worker-* flags

Command and Arguments:
- Use --command to set the leader's entrypoint command
- Use --arguments OR positional args (-- arg1 arg2) to set the leader's arguments
- Note: If both --arguments and -- are used, positional args (--) take precedence
- Use --worker-command and --worker-arguments to override worker's command/args

```
runai inference distributed submit [flags]
```

### Examples

```

# Submit a workload
runai inference distributed submit <workload-name> -p <project-name> -i tiangolo/uvicorn-gunicorn-fastapi:python3.9 --replicas 2 --workers 3 -g 1 --serving-port 80

# Submit a workload with a template
runai inference distributed submit <workload-name> -p <project-name> --template <template-name>

# Submit a workload where workers inherit leader settings but override specific fields
runai inference distributed submit <workload-name> -p <project-name> -i tiangolo/uvicorn-gunicorn-fastapi:python3.9 --environment ROLE=LEADER --replicas 3 --workers 7 -g 2 --serving-port 80 --worker-environment ROLE=WORKER

# Submit a workload with a different worker setup (workers not based on leader configuration)
runai inference distributed submit <workload-name> -p <project-name> -i tiangolo/uvicorn-gunicorn-fastapi:python3.9 --replicas 4 --workers 1 -g 8 --serving-port 80 --existing-pvc claimname=my-pvc,path=/data --worker-as-leader=false --worker-image tiangolo/uvicorn-gunicorn-fastapi:python3.9 --worker-gpu-devices-request 8
```

### Options

```
      --annotation stringArray                                Set of annotations to populate into the container running the workload
      --arguments string                                      Specifies the arguments to pass to the container command
      --capability stringArray                                Add POSIX capabilities to the container. Defaults to the runtime's default set.
      --command string                                        Specifies the command to run in the container, overriding the image's default entrypoint
      --compute string                                        Specifies the compute resource asset to use for the workload.
      --configmap-map-volume stringArray                      Mount a ConfigMap as a volume. Format: name=CONFIGMAP_NAME,path=PATH,subpath=SUBPATH,default-mode=DEFAULT_MODE.
      --cpu-core-limit positiveFloat                          Maximum number of CPU cores allowed (e.g. 0.5, 1).
      --cpu-core-request positiveFloat                        Number of CPU cores to request (e.g. 0.5, 1).
      --cpu-memory-limit string                               Maximum memory allowed (e.g. 1G, 500M).
      --cpu-memory-request string                             Amount of memory to request (e.g. 1G, 500M).
      --create-home-dir                                       Create a temporary home directory for the container. Defaults to true when --run-as-user is set, false otherwise.
      --datasource stringArray                                Specifies the data source asset to attach to the workload. Format: type=hostPath|nfs|pvc|git|s3|configMap|secretVolume,name=NAME
      --datavolume stringArray                                Mount a data volume. Format: name=DATA_VOLUME_NAME,mountpath=MOUNT_PATH.
      --empty-dir-volume stringArray                          Mount an empty directory as a volume. Format: name=NAME,path=PATH,medium=MEDIUM,size-limit=SIZE.
      --env-my-credentials stringArray                        Set an environment variable from a user credential. Format: type=<dockerRegistry|genericSecret|ngcApiKey>,name=<ENV_VAR>,credential-name=<CREDENTIAL_NAME>[,key=<KEY>]. Key is required only for type=genericSecret. Requires cluster version 2.22+ (2.23+ for ngcApiKey).
      --env-pod-field-ref stringArray                         Set an environment variable from a pod field reference. Format: ENV_VARIABLE=FIELD_REFERENCE.
      --env-secret stringArray                                Set an environment variable from a Kubernetes secret. Format: ENV_VARIABLE=secret-name,key=secret-key.
      --environment stringArray                               Specifies the environment asset to use for the workload.
  -e, --environment-variable stringArray                      Set environment variables in the container. Format: --environment-variable name=value --environment-variable name-b=value-b.
      --exclude-node stringArray                              Nodes that will be excluded from use by the scheduler. Format: --exclude-node node-a --exclude-node node-b
      --existing-pvc stringArray                              Mount an existing PersistentVolumeClaim. Format: claimname=CLAIM_NAME,path=PATH. Auto-complete supported.
      --extended-resource stringArray                         Request access to a Kubernetes extended resource. Format: resource_name=quantity.
      --gpu-compute-mode string                               GPU compute sharing mode. Valid values: time-slicing, sm-sharing. Requires cluster version 3.0.14+.
  -g, --gpu-devices-request positiveInt                       Number of GPU devices to allocate for the workload (e.g. 1, 2).
      --gpu-memory-limit string                               Maximum GPU memory to allocate (e.g. 1G, 500M).
      --gpu-memory-portion-limit positiveFloat                Maximum GPU memory fraction allowed per device (between 0 and 1).
      --gpu-memory-portion-request positiveFloat              GPU memory fraction to allocate per device (between 0 and 1, e.g. 0.5).
      --gpu-memory-request string                             Amount of GPU memory to allocate (e.g. 1G, 500M).
      --gpu-request-type string                               Type of GPU request: portion, memory
  -h, --help                                                  help for submit
      --host-ipc                                              Enable host IPC for the container. Default: false.
      --host-network                                          Enable host networking for the container. Default: false.
  -i, --image string                                          The container image to use for the workload.
      --image-pull-my-credentials stringArray                 Use a user credential for authenticating image pulls. Format: type=<dockerRegistry|ngcApiKey>,name=<NAME>. Requires cluster version 2.22+ (dockerRegistry) or 2.23+ (ngcApiKey).
      --image-pull-policy string                              Image pull policy for the container. Valid values: Always, IfNotPresent, Never.
      --label stringArray                                     Set of labels to populate into the container running the workspace
      --large-shm                                             Request a large /dev/shm device to mount in the container. Useful for memory-intensive workloads.
      --mnnvl mnnvl-mode                                      Multi-Node NVLINK (MNNVL) mode controls GPU network acceleration for distributed workloads. Valid values: None, Preferred, Required
      --name-prefix string                                    Set defined prefix for the workload name and add index as a suffix
      --new-pvc stringArray                                   Create and mount a new volume. This volume is used only for the duration of the workload's lifecycle. Format: claimname=CLAIM_NAME,storageclass=STORAGE_CLASS,size=SIZE,path=PATH,accessmode-rwo,accessmode-rom,accessmode-rwm,ro,ephemeral.
      --node-pools stringArray                                Node pools to use for scheduling the job, ordered by priority. Format: --node-pools pool-a --node-pools pool-b
      --node-type string                                      Enforce node type affinity by setting a node-type label.
      --pod-running-timeout duration                          Timeout for pod to reach running state (e.g. 5s, 2m, 3h).
      --preemptibility preemptibility                         Specify whether the workload can be preempted by higher-priority workloads. Valid values: preemptible, semi-preemptible, non-preemptible. Overrides the default preemptibility for the workload type.
      --preferred-pod-topology-key string                     If possible, schedule all pods of this workload within the same topology domain defined by a Kubernetes node label key. Format: KEY.
      --priority string                                       Sets the workload’s scheduling priority. Valid values: very-low, low, medium-low, medium, medium-high, high, very-high. 
                                                              Overrides the default priority for the workload type. Changing priority does not update preemptibility automatically.
      --privileged                                            Grants the container full access to the host, bypassing almost all container isolation; the container acts like root.
  -p, --project string                                        Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default.
      --replicas int32                                        The number of replicas (sets of leader and workers) to run. Defaults to 1 (default 1)
      --required-pod-topology-key string                      Require scheduling pods of this workload within the same topology domain defined by a Kubernetes node label key. Format: KEY.
      --restart-policy distributed-inference-restart-policy   Restart policy for distributed inference workload. Valid values: RecreateGroupOnPodRestart (restarts all pods if any pod fails), None (no automatic restart). Defaults to RecreateGroupOnPodRestart
      --run-as-gid int                                        Group ID to run the container as.
      --run-as-non-root                                       Force the container to run as a non-root user.
      --run-as-uid int                                        User ID to run the container as.
      --run-as-user                                           Set the user and group IDs for the container. Uses local terminal credentials if not specified.
      --seccomp-profile string                                Seccomp profile for the container. Valid values: RuntimeDefault, Unconfined, or Localhost.
      --secret-volume stringArray                             Mount a Kubernetes Secret as a volume. Format: path=PATH,name=SECRET_RESOURCE_NAME.
      --serving-port string                                   Defines various attributes for the serving port. Usage formats: (1) Simplified format: --serving-port=CONTAINER_PORT (2) Full format: --serving-port=container=CONTAINER_PORT,[authorization-type=public|authenticatedUsers|authorizedUsers|authorizedGroups],[authorized-users=USER1:USER2:APP1...],[authorized-groups=GROUP1:GROUP2...],[expose-externally=true|false],[exposed-url=URL],[protocol=http]
      --startup-policy distributed-inference-startup-policy   Controls when worker pods start. Valid values: LeaderCreated (workers start when leader is created), LeaderReady (workers start when leader is ready). Defaults to LeaderCreated
      --supplemental-groups ints                              Comma-separated list of group IDs for the container user.
      --template string                                       The name of the template to use for submitting this workload. Values from the template and CLI flags are combined. Flags with matching keys replace template values; other flags are applied in addition.
      --termination-grace-period duration                     The length of time (like 5s or 2m, higher than zero) the workload's pod is expected to terminate gracefully upon probe failure. In case value is not specified, kubernetes default of 30 seconds applies (default 0s)
      --toleration stringArray                                Add Kubernetes tolerations. Format: operator=Equal|Exists,key=KEY,[value=VALUE],[effect=NoSchedule|NoExecute|PreferNoSchedule],[seconds=SECONDS].
      --user-group-source string                              How to determine user/group IDs. Valid values: fromTheImage, fromIdpToken.
      --wait-for-submit duration                              How long to wait for the workload to be created in the cluster. Default: 1m.
      --worker-allow-privilege-escalation                     Allow worker pods to gain additional privileges after starting.
      --worker-annotation stringArray                         Set of annotations to populate into worker pods running the workload
      --worker-arguments string                               Specifies the arguments to pass to worker pod command
      --worker-as-leader                                      Use leader configuration for workers unless overridden with --worker-* flags. Set to false when workers require explicit configuration. (default true)
      --worker-capability stringArray                         Add POSIX capabilities to worker pods. Defaults to the runtime's default set.
      --worker-command string                                 Specifies the command to run in worker pods, overriding the image's default entrypoint
      --worker-configmap-map-volume stringArray               Mount a ConfigMap as a volume for worker pods. Format: name=CONFIGMAP_NAME,path=PATH,subpath=SUBPATH,default-mode=DEFAULT_MODE.
      --worker-cpu-core-limit positiveFloat                   Maximum number of CPU cores allowed per worker pod (e.g. 0.5, 1).
      --worker-cpu-core-request positiveFloat                 Number of CPU cores to request per worker pod (e.g. 0.5, 1).
      --worker-cpu-memory-limit string                        Maximum memory allowed per worker pod (e.g. 1G, 500M).
      --worker-cpu-memory-request string                      Amount of memory to request per worker pod (e.g. 1G, 500M).
      --worker-create-home-dir                                Create a temporary home directory for worker pods. Defaults to true when --worker-run-as-uid is set, false otherwise.
      --worker-empty-dir-volume stringArray                   Mount an empty directory as a volume for worker pods. Format: path=PATH,medium=MEDIUM,size-limit=SIZE.
      --worker-env-pod-field-ref stringArray                  Set an environment variable from a pod field reference for worker pods. Format: ENV_VARIABLE=FIELD_REFERENCE.
      --worker-env-secret stringArray                         Set an environment variable from a Kubernetes secret for worker pods. Format: ENV_VARIABLE=secret-name,key=secret-key.
      --worker-environment stringArray                        Set environment variables in worker pods. Format: --worker-environment name=value --worker-environment name-b=value-b.
      --worker-exclude-node stringArray                       Nodes that will be excluded from use by worker pods. Format: --worker-exclude-node node-a --worker-exclude-node node-b
      --worker-existing-pvc stringArray                       Mount an existing PersistentVolumeClaim for worker pods. Format: claimname=CLAIM_NAME,path=PATH. Auto-complete supported.
      --worker-extended-resource stringArray                  Request access to a Kubernetes extended resource per worker pod. Format: resource_name=quantity.
      --worker-gpu-compute-mode string                        GPU compute sharing mode. Valid values: time-slicing, sm-sharing. Requires cluster version 3.0.14+.
      --worker-gpu-devices-request positiveInt                Number of GPU devices to allocate per worker pod (e.g. 1, 2).
      --worker-gpu-memory-limit string                        Maximum GPU memory to allocate per worker pod (e.g. 1G, 500M).
      --worker-gpu-memory-portion-limit positiveFloat         Maximum GPU memory fraction allowed per device for each worker pod (between 0 and 1).
      --worker-gpu-memory-portion-request positiveFloat       GPU memory fraction to allocate per device for each worker pod (between 0 and 1, e.g. 0.5).
      --worker-gpu-memory-request string                      Amount of GPU memory to allocate per worker pod (e.g. 1G, 500M).
      --worker-host-ipc                                       Enable host IPC for worker pods. Default: false.
      --worker-host-network                                   Enable host networking for worker pods. Default: false.
      --worker-image string                                   The container image to use for worker pods.
      --worker-image-pull-policy string                       Image pull policy for worker pods. Valid values: Always, IfNotPresent, Never.
      --worker-label stringArray                              Set of labels to populate into worker pods running the workload
      --worker-new-pvc stringArray                            Create and mount a new volume for worker pods. This volume is used only for the duration of the workload's lifecycle. Format: claimname=CLAIM_NAME,storageclass=STORAGE_CLASS,size=SIZE,path=PATH,accessmode-rwo,accessmode-rom,accessmode-rwm,ro,ephemeral.
      --worker-node-type string                               Enforce node type affinity for worker pods by setting a node-type label.
      --worker-preferred-pod-topology-key string              If possible, schedule all worker pods within the same topology domain defined by a Kubernetes node label key. Format: KEY.
      --worker-privileged                                     Grants worker pods full access to the host, bypassing almost all container isolation; the container acts like root.
      --worker-required-pod-topology-key string               Require scheduling worker pods within the same topology domain defined by a Kubernetes node label key. Format: KEY.
      --worker-run-as-gid int                                 Group ID to run worker pods as.
      --worker-run-as-non-root                                Force worker pods to run as a non-root user.
      --worker-run-as-uid int                                 User ID to run worker pods as.
      --worker-secret-volume stringArray                      Mount a Kubernetes Secret as a volume for worker pods. Format: path=PATH,name=SECRET_RESOURCE_NAME.
      --worker-supplemental-groups ints                       Comma-separated list of group IDs for worker pods.
      --worker-toleration stringArray                         Add Kubernetes tolerations for worker pods. Format: operator=Equal|Exists,key=KEY,[value=VALUE],[effect=NoSchedule|NoExecute|PreferNoSchedule],[seconds=SECONDS].
      --worker-working-dir string                             Working directory inside worker pods. Overrides the default working directory set in the image.
      --workers positiveInt                                   the number of workers that will be allocated for running the workload
      --working-dir string                                    Working directory inside the container. Overrides the default working directory set in the image.
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

* [runai inference distributed](runai_inference_distributed.md)	 - Runs multiple coordinated inference processes across multiple nodes. Required for models too large to run on a single node.

