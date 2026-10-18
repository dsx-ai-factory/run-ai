## runai inference nim submit

submit a nim inference workload

### Synopsis

Before using the flags, keep in mind:

Model Store Configuration:
- You must select exactly one model store option: --model-existing-pvc, --model-new-pvc, or --model-nim-cache.

Autoscaling vs Fixed Replicas:
- Use --replicas for a fixed number of instances.
- Use --min-replicas, --max-replicas, --metric for autoscaling.

Multi-Node Inference:
- Use --workers to specify the number of worker nodes per leader.
- Setting workers > 0 automatically enables multi-node mode (1 Leader + N Workers).
- Note: Multi-node (--workers) cannot be used with autoscaling (--min-replicas/--max-replicas).

```
runai inference nim submit [flags]
```

### Examples

```

# Submit a Llama3 model using an existing PVC
runai inference nim submit <workload-name> -p <project-name> -i nvcr.io/nim/meta/llama3-8b-instruct:latest --ngc-auth-secret my-ngc-secret --model-existing-pvc claimname=my-model-pvc --serving-port 8000 -g 1

# Submit with autoscaling
runai inference nim submit <workload-name> -p <project-name> -i nvcr.io/nim/meta/llama3-8b-instruct:latest --ngc-auth-secret my-ngc-secret --model-nim-cache "name=my-cache,profile=10gb" --min-replicas 1 --max-replicas 5 --metric concurrency --metric-threshold 10 --serving-port 8000

# Submit multi-node model (1 Leader + 2 Workers)
runai inference nim submit <workload-name> -p <project-name> -i nvcr.io/nim/meta/llama3-8b-instruct:latest --ngc-auth-secret my-ngc-secret --model-new-pvc "claimname=my-model-pvc,size=100Gi,accessmode-rwo" --workers 2 --replicas 1 --serving-port 8000
```

### Options

```
      --annotation stringArray                     Set of annotations to populate into the container running the workload
      --category string                            Workload category
      --cpu-core-limit positiveFloat               Maximum number of CPU cores allowed (e.g. 0.5, 1).
      --cpu-core-request positiveFloat             Number of CPU cores to request (e.g. 0.5, 1).
      --cpu-memory-limit string                    Maximum memory allowed (e.g. 1G, 500M).
      --cpu-memory-request string                  Amount of memory to request (e.g. 1G, 500M).
  -e, --environment-variable stringArray           Set environment variables in the container. Format: --environment-variable name=value --environment-variable name-b=value-b.
      --gpu-compute-mode string                    GPU compute sharing mode. Valid values: time-slicing, sm-sharing. Requires cluster version 3.0.14+.
  -g, --gpu-devices-request positiveInt            Number of GPU devices to allocate for the workload (e.g. 1, 2).
      --gpu-memory-limit string                    Maximum GPU memory to allocate (e.g. 1G, 500M).
      --gpu-memory-portion-limit positiveFloat     Maximum GPU memory fraction allowed per device (between 0 and 1).
      --gpu-memory-portion-request positiveFloat   GPU memory fraction to allocate per device (between 0 and 1, e.g. 0.5).
      --gpu-memory-request string                  Amount of GPU memory to allocate (e.g. 1G, 500M).
      --gpu-request-type string                    Type of GPU request: portion, memory
  -h, --help                                       help for submit
  -i, --image string                               The container image to use for the workload.
      --image-pull-policy string                   Image pull policy for the container. Valid values: Always, IfNotPresent, Never.
      --image-pull-secret stringArray              Image pull secrets
      --label stringArray                          Set of labels to populate into the container running the workspace
      --max-replicas int32                         Maximum number of replicas for autoscaling. Defaults to min-replicas or 1
      --metric string                              Autoscaling metric (e.g. cpu, http_requests_total). Required when min-replicas < max-replicas
      --metric-threshold int32                     The threshold to use with the specified metric for autoscaling. Mandatory if metric is specified
      --min-replicas int32                         Minimum number of replicas for autoscaling (default 1)
      --model-existing-pvc string                  Mount an existing PVC for the model store. Format: claimname=CLAIM_NAME
      --model-new-pvc string                       Create and mount a new PVC for the model store. Format: [claimname=CLAIM_NAME],size=SIZE,[storageclass=STORAGE_CLASS],[accessmode-rwo|accessmode-rwm|accessmode-rom],[ro]
      --model-nim-cache string                     NIM Cache configuration. Format: name=CACHE_NAME,[profile=PROFILE]
      --ngc-auth-secret string                     Name of the secret containing NGC API key
      --preemptibility preemptibility              Specify whether the workload can be preempted by higher-priority workloads. Valid values: preemptible, semi-preemptible, non-preemptible. Overrides the default preemptibility for the workload type.
      --priority string                            Sets the workload’s scheduling priority. Valid values: very-low, low, medium-low, medium, medium-high, high, very-high. 
                                                   Overrides the default priority for the workload type. Changing priority does not update preemptibility automatically.
  -p, --project string                             Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default.
      --readiness-probe string                     Readiness probe. Format: port=PORT,[path=PATH],[host=HOST],[scheme=HTTP|HTTPS],[initial-delay=SECONDS],[period=SECONDS],[timeout=SECONDS],[success=THRESHOLD],[failure=THRESHOLD]. Port is required
      --replicas int32                             Number of replicas to run (default 1)
      --run-as-gid int                             Group ID to run the container as.
      --run-as-uid int                             User ID to run the container as.
      --scale-window-seconds int32                 The duration (in seconds) for which the autoscaler considers past metrics when making scaling decisions.
      --serving-port string                        Serving port options. Simplified: --serving-port=PORT. Full: --serving-port=port=PORT,[grpc-port=GRPC_PORT],[metrics-port=METRICS_PORT],[expose-externally],[exposed-url=URL],[exposed-protocol=http|grpc],[service-type=ClusterIP|NodePort|LoadBalancer|ExternalName]
      --toleration stringArray                     Add Kubernetes tolerations. Format: operator=Equal|Exists,key=KEY,[value=VALUE],[effect=NoSchedule|NoExecute|PreferNoSchedule],[seconds=SECONDS].
      --use-name-as-prefix                         Use the provided workload name as a prefix and generate a unique suffix
      --workers int32                              Number of worker nodes for multi-node NIM. Cannot be used with autoscaling (--min-replicas/--max-replicas)
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

* [runai inference nim](runai_inference_nim.md)	 - [Experimental] Runs NVIDIA NIM (NVIDIA Inference Microservices) workloads. Optimized for deploying foundation models.

