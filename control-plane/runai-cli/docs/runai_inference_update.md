## runai inference update

update an inference workload

```
runai inference update [WORKLOAD_NAME] [flags]
```

### Examples

```
# Update a workload with a new image
runai inference update <workload-name> -p <project-name> -i ubuntu

# Update a workload with a new autoscaling configuration
runai inference update <workload-name> -p <project-name> --max-replicas=5 --min-replicas=3 --metric=concurrency --metric-threshold=10

```

### Options

```
      --activation-replicas int32                  The number of replicas to run when scaling-up from zero. Defaults to minReplicas, or to 1 if minReplicas is set to 0
  -c, --command                                    If true, override the image's entrypoint with the command supplied after '--'
      --concurrency-hard-limit int32               The maximum number of requests allowed to flow to a single replica at any time. 0 means no limit
      --cpu-core-limit positiveFloat               Maximum number of CPU cores allowed (e.g. 0.5, 1).
      --cpu-core-request positiveFloat             Number of CPU cores to request (e.g. 0.5, 1).
      --cpu-memory-limit string                    Maximum memory allowed (e.g. 1G, 500M).
      --cpu-memory-request string                  Amount of memory to request (e.g. 1G, 500M).
      --create-home-dir                            Create a temporary home directory for the container. Defaults to true when --run-as-user is set, false otherwise.
      --environment stringArray                    Specifies the environment asset to use for the workload.
  -e, --environment-variable stringArray           Set environment variables in the container. Format: --environment-variable name=value --environment-variable name-b=value-b.
      --exclude-node stringArray                   Nodes that will be excluded from use by the scheduler. Format: --exclude-node node-a --exclude-node node-b
      --extended-resource stringArray              Request access to a Kubernetes extended resource. Format: resource_name=quantity.
      --gpu-compute-mode string                    GPU compute sharing mode. Valid values: time-slicing, sm-sharing. Requires cluster version 3.0.14+.
      --gpu-devices-request positiveInt            Number of GPU devices to allocate for the workload (e.g. 1, 2).
      --gpu-memory-limit string                    Maximum GPU memory to allocate (e.g. 1G, 500M).
      --gpu-memory-portion-limit positiveFloat     Maximum GPU memory fraction allowed per device (between 0 and 1).
      --gpu-memory-portion-request positiveFloat   GPU memory fraction to allocate per device (between 0 and 1, e.g. 0.5).
      --gpu-memory-request string                  Amount of GPU memory to allocate (e.g. 1G, 500M).
      --gpu-request-type string                    Type of GPU request: portion, memory
  -h, --help                                       help for update
  -i, --image string                               The container image to use for the workload.
      --image-pull-policy string                   Image pull policy for the container. Valid values: Always, IfNotPresent, Never.
      --initial-replicas int32                     The number of replicas to run when initializing the workload for the first time. Defaults to minReplicas, or to 1 if minReplicas is set to 0
      --initialization-timeout-seconds int32       The maximum amount of time (in seconds) to wait for the container to become ready
      --large-shm                                  Request a large /dev/shm device to mount in the container. Useful for memory-intensive workloads.
      --max-replicas int32                         The maximum number of replicas for autoscaling. Defaults to minReplicas, or to 1 if minReplicas is set to 0
      --metric string                              Autoscaling metric is required if minReplicas < maxReplicas, except when minReplicas = 0 and maxReplicas = 1. Use 'throughput', 'concurrency', 'latency', or custom metrics.
      --metric-threshold int32                     The threshold to use with the specified metric for autoscaling. Mandatory if metric is specified
      --metric-threshold-percentage float32        The percentage of metric threshold value to use for autoscaling. Defaults to 70. Applicable only with the 'throughput' and 'concurrency' metrics
      --min-replicas int32                         The minimum number of replicas for autoscaling. Defaults to 1. Use 0 to allow scale-to-zero
      --node-pools stringArray                     Node pools to use for scheduling the job, ordered by priority. Format: --node-pools pool-a --node-pools pool-b
  -p, --project string                             Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default.
      --request-timeout-seconds int32              The maximum time (in seconds) allowed to process an end-user request. If no response is returned within this time, the request will be ignored
      --scale-down-delay-seconds int32             The minimum amount of time (in seconds) that a replica will remain active after a scale-down decision
      --scale-to-zero-retention-seconds int32      The minimum amount of time (in seconds) that the last replica will remain active after a scale-to-zero decision. Defaults to 0. Available only if minReplicas is set to 0
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

