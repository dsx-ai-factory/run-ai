## runai pytorch describe

describe a pytorch training workload

```
runai pytorch describe [WORKLOAD_NAME] [flags]
```

### Examples

```
# Describe a workload with a default project
runai training pytorch describe <workload-name>

# Describe a workload in a specific project
runai training pytorch describe <workload-name> -p <project-name>

# Describe a workload by UUID
runai training pytorch describe --uuid=<workload-id>

# Describe a workload with specific output format
runai training pytorch describe <workload-name> -o json

# Describe a workload with specific sections
runai training pytorch describe <workload-name> --general --compute --pods --events --networks

# Describe a workload with container details and custom limits
runai training pytorch describe <workload-name> --containers --pod-limit 20 --event-limit 100
```

### Options

```
      --compute             Show compute resources information (CPU, GPU, memory, etc.). (default true)
      --containers          Include container-level details in pod information.
      --event-limit int32   Limit the number of events displayed. Use -1 for no limit. (default 50)
      --events              Show recent events for the resource. (default true)
      --general             Show general resource information. (default true)
  -h, --help                help for describe
      --networks            Show network-related information. (default true)
  -o, --output string       Output format: table, json, or yaml. Default is table. (default "table")
      --pod-limit int32     Limit the number of pods displayed. Use -1 for no limit. (default 10)
      --pods                Show information about pods associated with the resource. (default true)
  -p, --project string      Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default.
      --sortEvents string   Sort events in ascending or descending order. Valid values: asc, desc. (default "asc")
      --storage             Show storage information (PVCs, ConfigMaps, Secrets, etc.). (default true)
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

* [runai pytorch](runai_pytorch.md)	 - alias for pytorch management

