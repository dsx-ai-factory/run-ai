## runai training bash

open a bash shell in a standard training workload

```
runai training bash [WORKLOAD_NAME] [flags]
```

### Examples

```
# Open a bash shell in the workload's main worker
runai training standard bash <workload-name>

# Open a bash shell in a specific worker of the workload
runai training standard bash <workload-name> --pod <pod-name>
```

### Options

```
  -c, --container string               The name of the container within the pod.
  -h, --help                           help for bash
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

* [runai training](runai_training.md)	 - training management

