## runai training suspend

suspend a standard training workload

```
runai training suspend [WORKLOAD_NAME] [flags]
```

### Examples

```
# Suspend a workload
runai training standard suspend <standard-name>

# Suspend a workload in a specific project
runai training standard suspend <standard-name> -p <project-name>

# Suspend a workload by UUID
runai training standard suspend --uuid=<standard_uuid>
```

### Options

```
  -h, --help             help for suspend
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

* [runai training](runai_training.md)	 - training management

