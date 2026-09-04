# runai inference nim delete

delete a nim inference workload

```
runai inference nim delete [WORKLOAD_NAME] [flags]
```

### Examples

```
# Delete a workload with a default project
runai inference nim delete <workload-name>

# Delete a workload with a specific project
runai inference nim delete <workload-name> -p <project-name>

# Delete a workload by UUID
runai inference nim delete --uuid=<workload-id> -p <project-name>
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

* [runai inference nim](runai_inference_nim.md)	 - [Experimental] Runs NVIDIA NIM (NVIDIA Inference Microservices) workloads. Optimized for deploying foundation models.

