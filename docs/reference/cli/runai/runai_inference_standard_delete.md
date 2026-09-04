# runai inference standard delete

delete an inference workload

```
runai inference standard delete [WORKLOAD_NAME] [flags]
```

### Examples

```
# Delete a workload with a default project
runai inference delete <workload-name>

# Delete a workload with a specific project
runai inference delete <workload-name> -p <project-name>

# Delete a workload by UUID
runai inference delete --uuid=<workload-id> -p <project-name>
```

### Options

```
  -h, --help             help for delete
  -p, --project string   Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default.
  -u, --uuid string      The unique identifier (UUID) of the resource, as returned by the API.
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

* [runai inference standard](runai_inference_standard.md)	 - Runs a single inference process on one node. Suitable for smaller models or simpler inference tasks.

