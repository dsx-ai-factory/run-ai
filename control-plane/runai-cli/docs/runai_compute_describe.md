## runai compute describe

describe a compute resource asset

```
runai compute describe [COMPUTE_NAME] [flags]
```

### Examples

```
# Describe a compute resource asset with a default project
runai compute describe compute-large

# Describe a compute resource asset in a specific project
runai compute describe compute-large -p <project_name>

# Describe a compute resource asset with specific output format
runai compute describe compute-large -o json
```

### Options

```
  -h, --help             help for describe
  -o, --output string    Output format: table, json, or yaml. Default is table. (default "table")
  -p, --project string   Specify the project. Higher-scope compute-resources in the same tenant are included. Use runai project set <project> to set the default
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

* [runai compute](runai_compute.md)	 - Compute resource asset management

