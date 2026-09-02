## runai training mpi resume

resume mpi training

```
runai training mpi resume [WORKLOAD_NAME] [flags]
```

### Examples

```
# Resume a workload
runai training mpi resume <mpi-name>

# Resume a workload in a specific project
runai training mpi resume <mpi-name> -p <project-name>

# Resume a workload by UUID
runai training mpi resume --uuid=<mpi_uuid>
```

### Options

```
  -h, --help             help for resume
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

* [runai training mpi](runai_training_mpi.md)	 - mpi management

