# runai environment describe

describe an environment asset

```
runai environment describe ENVIRONMENT_NAME [flags]
```

### Examples

```
# Describe an environment asset with a default project
runai environment describe pytorch-env

# Describe an environment asset in a specific project
runai environment describe pytorch-env -p <project_name>

# Describe an environment asset with specific output format
runai environment describe pytorch-env -o json
```

### Options

```
  -h, --help             help for describe
  -o, --output string    Output format: table, json, or yaml. Default is table. (default "table")
  -p, --project string   Specify the project. Higher-scope environments in the same tenant are included. Use 'runai project set <project>' to set the default.
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

* [runai environment](runai_environment.md)	 - Environment asset management

