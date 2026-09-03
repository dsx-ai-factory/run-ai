# runai workload-type describe

Describe a workload type

```
runai workload-type describe WORKLOAD_TYPE_NAME [flags]
```

### Examples

```
  # Describe a workload type by name
  runai workload-type describe MyWorkloadType

  # Output as JSON
  runai workload-type describe MyWorkloadType --output json

  # Output as YAML
  runai workload-type describe MyWorkloadType -o yaml
```

### Options

```
  -h, --help            help for describe
  -o, --output string   Output format: table, json, or yaml. Default is table. (default "table")
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

* [runai workload-type](runai_workload-type.md)	 - workload type management

