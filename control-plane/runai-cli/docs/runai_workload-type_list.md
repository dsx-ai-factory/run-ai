## runai workload-type list

list available workload types

```
runai workload-type list [flags]
```

### Examples

```
  # List all workload types
  runai workload-type list

  # List only external workload types
  runai workload-type list --external-only

  # List with specific columns
  runai workload-type list --columns name,category,priority

  # Output as JSON
  runai workload-type list --json
```

### Options

```
      --columns strings                    Specify which columns to display (comma-separated list)
      --external-only[=true|false]         Excludes native workload types and returns only NVIDIA Run:ai supported workload types and user-defined workload types.
  -h, --help[=true|false]                  help for list
      --json[=true|false]                  Output structure JSON
      --no-headers[=true|false]            Output structure table without headers
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

* [runai workload-type](runai_workload-type.md)	 - workload type management

