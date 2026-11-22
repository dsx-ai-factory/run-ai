## runai application list

List AI applications

```
runai application list [flags]
```

### Examples

```
# List all AI applications from all projects
runai application list -A

# List AI applications in a specific project
runai application list -p <project_name>

# List AI applications with a specific output format
runai application list --yaml

# List AI applications with custom columns
runai application list --columns name,status,gpucomputereq
```

### Options

```
  -A, --all[=true|false]                   Show resources from all projects
      --columns strings                    Specify which columns to display (comma-separated list)
  -h, --help[=true|false]                  help for list
      --json[=true|false]                  Output structure JSON
      --max-items int32                    set the max number of items to return, default is all of the items
      --next-token next_token              set the token for requesting the next page
      --no-headers[=true|false]            Output structure table without headers
      --no-pagination[=true|false]         return a single page instead of full list, by default set to false
      --page-size int32                    set the single page size (max 500) (default 100)
  -p, --project string                     Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default.
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

* [runai application](runai_application.md)	 - ai application management

