## runai mpi submit

submit a mpi training workload

```
runai mpi submit [flags]
```

### Examples

```
# Submit a workload
runai training mpi submit <workload-name> -p <project-name> -i ubuntu

# Submit a workload with arguments
runai training mpi submit <workload-name> -p <project-name> -i ubuntu -- ls -la

# Submit a workload with a template
runai training mpi submit <workload-name> -p <project-name> --template <template-name>

# Submit a workload with an environment asset
runai training mpi submit <workload-name> -p <project-name> --environment <environment-asset-name>

# Submit a workload with a custom command
runai training mpi submit <workload-name> -p <project-name> -i ubuntu --command -- echo "Hello, World"

# Submit a workload with a field reference 
runai training mpi submit <workload-name> -p <project-name> -i ubuntu --env-pod-field-ref "PROJECT=metadata.labels['project']"

# Submit a workload master args with worker args
runai training mpi submit <workload-name> -p <project-name> -i ubuntu --master-args "-a master_arg_a -b master-arg_b'" -- '-a worker_arg_a'

# Submit a workload master command with worker args
runai training mpi submit <workload-name> -p <project-name> -i ubuntu --master-command "echo -e 'master command'" -- '-a worker_arg_a'

# Submit a workload master command with worker command
runai training mpi submit <workload-name> -p <project-name> -i ubuntu --master-command "echo -e 'master command'" --command -- echo -e 'worker command'
```

### Options

```
      --allow-privilege-escalation                          Allow the container to gain additional privileges after starting.
      --annotation stringArray                              Set of annotations to populate into the container running the workload
      --attach                                              Wait for the pod to start running, then attach to it as if 'runai attach' was called. Implies --tty and --stdin.
      --auto-deletion-time-after-completion duration        Automatically delete a completed job after a specified duration (e.g. 5s, 2m, 3h). (default 0s)
      --backoff-limit int32                                 Number of times to retry a failed job before marking it as failed.
      --capability stringArray                              Add POSIX capabilities to the container. Defaults to the runtime's default set.
      --clean-pod-policy string                             Specifies which pods will be deleted when the workload reaches a terminal state (completed/failed). Valid values are None, All, Running.
                                                            * Running - Only pods still running when a job completes (for example, parameter servers) will be deleted immediately. Completed pods will not be deleted so that the logs will be preserved.
                                                            * All     - All (including completed) pods will be deleted immediately when the job finishes.
                                                            * None    - No pods will be deleted when the job completes. It will keep running pods that consume GPU, CPU and memory over time. 
                                                                        It is recommended to set to None only for debugging and obtaining logs from running pods. (Default)
  -c, --command                                             If true, override the image's entrypoint with the command supplied after '--'
      --compute string                                      Specifies the compute resource asset to use for the workload.
      --configmap-map-volume stringArray                    Mount a ConfigMap as a volume. Format: name=CONFIGMAP_NAME,path=PATH,subpath=SUBPATH,default-mode=DEFAULT_MODE.
      --cpu-core-limit positiveFloat                        Maximum number of CPU cores allowed (e.g. 0.5, 1).
      --cpu-core-request positiveFloat                      Number of CPU cores to request (e.g. 0.5, 1).
      --cpu-memory-limit string                             Maximum memory allowed (e.g. 1G, 500M).
      --cpu-memory-request string                           Amount of memory to request (e.g. 1G, 500M).
      --create-home-dir                                     Create a temporary home directory for the container. Defaults to true when --run-as-user is set, false otherwise.
      --datasource stringArray                              Specifies the data source asset to attach to the workload. Format: type=hostPath|nfs|pvc|git|s3|configMap|secretVolume,name=NAME
      --datavolume stringArray                              Mount a data volume. Format: name=DATA_VOLUME_NAME,mountpath=MOUNT_PATH.
      --empty-dir-volume stringArray                        Mount an empty directory as a volume. Format: name=NAME,path=PATH,medium=MEDIUM,size-limit=SIZE.
      --env-my-credentials stringArray                      Set an environment variable from a user credential. Format: type=<dockerRegistry|genericSecret|ngcApiKey>,name=<ENV_VAR>,credential-name=<CREDENTIAL_NAME>[,key=<KEY>]. Key is required only for type=genericSecret. Requires cluster version 2.22+ (2.23+ for ngcApiKey).
      --env-pod-field-ref stringArray                       Set an environment variable from a pod field reference. Format: ENV_VARIABLE=FIELD_REFERENCE.
      --env-secret stringArray                              Set an environment variable from a Kubernetes secret. Format: ENV_VARIABLE=secret-name,key=secret-key.
      --environment stringArray                             Specifies the environment asset to use for the workload.
  -e, --environment-variable stringArray                    Set environment variables in the container. Format: --environment-variable name=value --environment-variable name-b=value-b.
      --exclude-node stringArray                            Nodes that will be excluded from use by the scheduler. Format: --exclude-node node-a --exclude-node node-b
      --existing-pvc stringArray                            Mount an existing PersistentVolumeClaim. Format: claimname=CLAIM_NAME,path=PATH. Auto-complete supported.
      --extended-resource stringArray                       Request access to a Kubernetes extended resource. Format: resource_name=quantity.
      --external-url stringArray                            Expose a URL from the workload container. Format: container=PORT,url=https://external.runai.com,authusers=user1,authgroups=group1.
      --git-sync stringArray                                Mount a Git repository into the container. Format: name=NAME,repository=REPO,path=PATH,secret=SECRET,rev=REVISION.
      --gpu-compute-mode string                             GPU compute sharing mode. Valid values: time-slicing, sm-sharing. Requires cluster version 3.0.14+.
  -g, --gpu-devices-request positiveInt                     Number of GPU devices to allocate for the workload (e.g. 1, 2).
      --gpu-memory-limit string                             Maximum GPU memory to allocate (e.g. 1G, 500M).
      --gpu-memory-portion-limit positiveFloat              Maximum GPU memory fraction allowed per device (between 0 and 1).
      --gpu-memory-portion-request positiveFloat            GPU memory fraction to allocate per device (between 0 and 1, e.g. 0.5).
      --gpu-memory-request string                           Amount of GPU memory to allocate (e.g. 1G, 500M).
      --gpu-request-type string                             Type of GPU request: portion, memory
  -h, --help                                                help for submit
      --host-ipc                                            Enable host IPC for the container. Default: false.
      --host-network                                        Enable host networking for the container. Default: false.
      --host-path stringArray                               Mount a host path as a volume. Format: path=PATH,mount=MOUNT,mount-propagation=None|HostToContainer,readwrite.
  -i, --image string                                        The container image to use for the workload.
      --image-pull-my-credentials stringArray               Use a user credential for authenticating image pulls. Format: type=<dockerRegistry|ngcApiKey>,name=<NAME>. Requires cluster version 2.22+ (dockerRegistry) or 2.23+ (ngcApiKey).
      --image-pull-policy string                            Image pull policy for the container. Valid values: Always, IfNotPresent, Never.
      --label stringArray                                   Set of labels to populate into the container running the workspace
      --large-shm                                           Request a large /dev/shm device to mount in the container. Useful for memory-intensive workloads.
      --launcher-creation-policy launcher-creation-policy   Launcher creation policy. Valid values: AtStartup, WaitForWorkersReady
      --master-args string                                  Specifies the arguments to pass to the master pod container command
      --master-command string                               Specifies the command to run in the master pod container, overriding the image's default entrypoint. The command can include arguments following it.
      --master-cpu-core-limit positiveFloat                 Maximum number of CPU cores allowed for the master pod (e.g., 0.5, 1).
      --master-cpu-core-request positiveFloat               Number of CPU cores to request for the master pod (e.g., 0.5, 1).
      --master-cpu-memory-limit string                      Maximum memory allowed for the master pod (e.g., 1G, 500M).
      --master-cpu-memory-request string                    Amount of memory to request for the master pod (e.g., 1G, 500M).
      --master-datasource stringArray                       Specifies the data source asset to attach to the master pod. Format: type=hostPath|nfs|pvc|git|s3|configMap|secretVolume,name=NAME
      --master-environment-variable stringArray             Set environment variables in the master container. --master-environment-variable name=value -master-environment-variable name-b=value-b.
      --master-extended-resource stringArray                Request access to a Kubernetes extended resource. Format: resource_name=quantity.
      --master-no-pvcs                                      Do not mount any persistent volumes in the master pod
      --master-restart-policy restart-policy                Restart policy for master pod. Valid values: Always, OnFailure, Never.
      --name-prefix string                                  Set defined prefix for the workload name and add index as a suffix
      --new-pvc stringArray                                 Create and mount a new volume. This volume is used only for the duration of the workload's lifecycle. Format: claimname=CLAIM_NAME,storageclass=STORAGE_CLASS,size=SIZE,path=PATH,accessmode-rwo,accessmode-rom,accessmode-rwm,ro,ephemeral.
      --nfs stringArray                                     Mount an NFS volume. Format: path=PATH,server=SERVER,mountpath=MOUNT_PATH,readwrite.
      --node-pools stringArray                              Node pools to use for scheduling the job, ordered by priority. Format: --node-pools pool-a --node-pools pool-b
      --node-type string                                    Enforce node type affinity by setting a node-type label.
      --pod-running-timeout duration                        Timeout for pod to reach running state (e.g. 5s, 2m, 3h).
      --port stringArray                                    Expose ports from the workload container. Format: service-type=NodePort,container=80,external=8080.
      --preemptibility preemptibility                       Specify whether the workload can be preempted by higher-priority workloads. Valid values: preemptible, semi-preemptible, non-preemptible. Overrides the default preemptibility for the workload type.
      --preferred-pod-topology-key string                   If possible, schedule all pods of this workload within the same topology domain defined by a Kubernetes node label key. Format: KEY.
      --priority string                                     Sets the workload’s scheduling priority. Valid values: very-low, low, medium-low, medium, medium-high, high, very-high. 
                                                            Overrides the default priority for the workload type. Changing priority does not update preemptibility automatically.
      --privileged                                          Grants the container full access to the host, bypassing almost all container isolation; the container acts like root.
  -p, --project string                                      Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default.
      --required-pod-topology-key string                    Require scheduling pods of this workload within the same topology domain defined by a Kubernetes node label key. Format: KEY.
      --restart-policy restart-policy                       Restart policy for worker pods. Valid values: Always, OnFailure, Never.
      --run-as-gid int                                      Group ID to run the container as.
      --run-as-non-root                                     Force the container to run as a non-root user.
      --run-as-uid int                                      User ID to run the container as.
      --run-as-user                                         Set the user and group IDs for the container. Uses local terminal credentials if not specified.
      --s3 stringArray                                      Mount an S3 bucket as a volume. Format: name=NAME,bucket=BUCKET,path=PATH,accesskey=ACCESS_KEY,url=URL.
      --seccomp-profile string                              Seccomp profile for the container. Valid values: RuntimeDefault, Unconfined, or Localhost.
      --secret-volume stringArray                           Mount a Kubernetes Secret as a volume. Format: path=PATH,name=SECRET_RESOURCE_NAME.
      --slots-per-worker int32                              Number of slots to allocate for each worker
      --ssh-auth-mount-path string                          Specifies the directory where SSH keys are mounted.
      --stdin                                               Keep stdin open on the container(s) in the pod, even if nothing is attached.
      --supplemental-groups ints                            Comma-separated list of group IDs for the container user.
      --template string                                     The name of the template to use for submitting this workload. Values from the template and CLI flags are combined. Flags with matching keys replace template values; other flags are applied in addition.
      --termination-grace-period duration                   The length of time (like 5s or 2m, higher than zero) the workload's pod is expected to terminate gracefully upon probe failure. In case value is not specified, kubernetes default of 30 seconds applies (default 0s)
      --toleration stringArray                              Add Kubernetes tolerations. Format: operator=Equal|Exists,key=KEY,[value=VALUE],[effect=NoSchedule|NoExecute|PreferNoSchedule],[seconds=SECONDS].
  -t, --tty                                                 Allocate a TTY for the container. Useful for interactive workloads.
      --user-group-source string                            How to determine user/group IDs. Valid values: fromTheImage, fromIdpToken.
      --wait-for-submit duration                            How long to wait for the workload to be created in the cluster. Default: 1m.
      --workers positiveInt                                 the number of workers that will be allocated for running the workload
      --working-dir string                                  Working directory inside the container. Overrides the default working directory set in the image.
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

* [runai mpi](runai_mpi.md)	 - alias for mpi management

