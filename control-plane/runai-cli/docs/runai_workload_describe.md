## runai workload describe

Describe a workload

```
runai workload describe WORKLOAD_NAME [flags]
```

### Options

```
      --compute             Show compute resources information (CPU, GPU, memory, etc.). (default true)
      --containers          Include container-level details in pod information.
      --event-limit int32   Limit the number of events displayed. Use -1 for no limit. (default 50)
      --events              Show recent events for the resource. (default true)
      --framework string    Filter by workload framework (TensorFlow, PyTorch, MPI, XGBoost, JAX).
      --general             Show general resource information. (default true)
  -h, --help                help for describe
      --networks            Show network-related information. (default true)
  -o, --output string       Output format: table, json, or yaml. Default is table. (default "table")
      --pod-limit int32     Limit the number of pods displayed. Use -1 for no limit. (default 10)
      --pods                Show information about pods associated with the resource. (default true)
  -p, --project string      Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default.
      --sortEvents string   Sort events in ascending or descending order. Valid values: asc, desc. (default "asc")
      --storage             Show storage information (PVCs, ConfigMaps, Secrets, etc.). (default true)
      --type string         The type of workload. Valid values: training, workspace, inference.
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

* [runai workload](runai_workload.md)	 - workload management

