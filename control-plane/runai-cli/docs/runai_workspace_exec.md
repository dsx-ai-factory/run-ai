## runai workspace exec

execute a command in a workspace workload

```
runai workspace exec [WORKLOAD_NAME] [flags]
```

### Examples

```
# Execute bash in the workload's main worker
runai workspace exec <workload-name> --tty --stdin -- /bin/bash 

# Execute ls command in the workload's main worker
runai workspace exec <workload-name> -- ls

# Execute a command in a specific worker of the workload
runai workspace exec <workload-name> --pod <pod-name> -- nvidia-smi
```

### Options

```
  -c, --container string               The name of the container within the pod.
  -h, --help[=true|false]              help for exec
      --pod string                     The pod ID. If not specified, the first pod will be used.
      --pod-running-timeout duration   Timeout for pod to reach running state (e.g. 5s, 2m, 3h).
  -p, --project string                 Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default.
  -i, --stdin[=true|false]             Pass stdin to the container
  -t, --tty[=true|false]               Stdin is a TTY
      --wait-timeout duration          Timeout while waiting for the workload to become ready for log streaming (e.g., 5s, 2m, 3h).
```

### Options inherited from parent commands

```
      --config-file string     config file name; can be set by environment variable RUNAI_CLI_CONFIG_FILE (default "config.json")
      --config-path string     config path; can be set by environment variable RUNAI_CLI_CONFIG_PATH
  -d, --debug[=true|false]     enable debug mode
  -q, --quiet[=true|false]     enable quiet mode, suppress all output except error messages
      --verbose[=true|false]   enable verbose mode
```

### SEE ALSO

* [runai workspace](runai_workspace.md)	 - workspace management

