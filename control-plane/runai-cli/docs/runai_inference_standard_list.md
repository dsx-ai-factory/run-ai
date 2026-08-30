## runai inference standard list

list inference workloads

```
runai inference standard list [flags]
```

### Examples

```
# List all inference workloads
runai inference standard list -A

# List workloads with default project
runai inference standard list

# List workloads in a specific project
runai inference standard list -p <project-name>

# List all workloads with a specific output format
runai inference standard list --yaml

# List workloads with pagination
runai inference standard list --limit 20 --offset 40
```

### Options

```
  -A, --all                     Show resources from all projects
      --columns strings         Specify which columns to display (comma-separated list)
      --deleted                 Return only resources that have been deleted.
  -h, --help                    help for list
      --json                    Output structure JSON
      --max-items int32         set the max number of items to return, default is all of the items
      --next-token next_token   set the token for requesting the next page
      --no-headers              Output structure table without headers
      --no-pagination           return a single page instead of full list, by default set to false
      --page-size int32         set the single page size (max 500) (default 100)
  -p, --project string          Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default.
      --raw                     Return the unprocessed response from the API call
      --set-default-columns     [Experimental] Set the columns flag value to the default output
      --status string           Filter by workload state (e.g., Pending, Running, Completed, Failed, Stopped, Initializing).
      --table                   Output structure table
      --yaml                    Output structure YAML
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

