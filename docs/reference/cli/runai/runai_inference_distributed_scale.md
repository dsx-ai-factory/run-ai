# runai inference distributed scale

scale a distributed inference workload

```
runai inference distributed scale [WORKLOAD_NAME] [flags]
```

### Examples

```
# Scale a workload (replicas flag is required)
runai inference distributed scale <workload-name> --replicas 5
```

### Options

```
      --dry-run          If true, only print the object that would be sent, without sending it
  -h, --help             help for scale
  -p, --project string   Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default.
      --replicas int32   The number of replicas (sets of leader and workers) to run. Defaults to 1
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

