## runai inference distributed exec

execute a command in a distributed inference workload

### Synopsis

If no pod is specified, a leader pod will be used.

```
runai inference distributed exec [WORKLOAD_NAME] [flags]
```

### Examples

```
# Execute bash in the workload
runai inference distributed exec <workload-name> --tty --stdin -- /bin/bash 

# Execute ls command in the workload
runai inference distributed exec <workload-name> -- ls

# Execute a command in a specific pod of the workload
runai inference distributed exec <workload-name> --pod <pod-name> -- nvidia-smi
```

### Options

```
  -c, --container string               The name of the container within the pod.
  -h, --help                           help for exec
      --pod string                     The pod ID. If not specified, the first pod will be used.
      --pod-running-timeout duration   Timeout for pod to reach running state (e.g. 5s, 2m, 3h).
  -p, --project string                 Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default.
  -i, --stdin                          Pass stdin to the container
  -t, --tty                            Stdin is a TTY
      --wait-timeout duration          Timeout while waiting for the workload to become ready for log streaming (e.g., 5s, 2m, 3h).
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

