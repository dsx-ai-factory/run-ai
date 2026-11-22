## runai compute list

list compute resource assets

```
runai compute list [flags]
```

### Examples

```
# List all compute resource assets
runai compute list -A

# List compute resource assets in a specific project
runai compute list -p <project_name>

# List compute resource assets with a specific output format
runai compute list --yaml
```

### Options

```
  -A, --all[=true|false]                   Show resources from all projects
      --columns strings                    Specify which columns to display (comma-separated list)
  -h, --help[=true|false]                  help for list
      --json[=true|false]                  Output structure JSON
      --no-headers[=true|false]            Output structure table without headers
  -p, --project string                     Specify the project. Higher-scope compute-resources in the same tenant are included. Use runai project set <project> to set the default
      --raw[=true|false]                   Return the unprocessed response from the API call
      --set-default-columns[=true|false]   [Experimental] Set the columns flag value to the default output
      --table[=true|false]                 Output structure table
      --yaml[=true|false]                  Output structure YAML
```

### Options inherited from parent commands

```
      --config-file string     config file name; can be set by environment variable RUNAI_CLI_CONFIG_FILE (default "config.json")
      --config-path string     config path; can be set by environment variable RUNAI_CLI_CONFIG_PATH
  -d, --debug[=true|false]     enable debug mode
  -q, --quiet[=true|false]     enable quiet mode, suppress all output except error messages
      --verbose[=true|false]   enable verbose mode
```

### SEE ALSO

* [runai compute](runai_compute.md)	 - Compute resource asset management

