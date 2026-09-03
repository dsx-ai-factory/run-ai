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
      --columns strings       Specify which columns to display (comma-separated list)
      --external-only         Excludes native workload types and returns only NVIDIA Run:ai supported workload types and user-defined workload types.
  -h, --help                  help for list
      --json                  Output structure JSON
      --no-headers            Output structure table without headers
      --raw                   Return the unprocessed response from the API call
      --set-default-columns   [Experimental] Set the columns flag value to the default output
      --table                 Output structure table
      --yaml                  Output structure YAML
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

