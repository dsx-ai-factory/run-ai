## runai template list

list all workload templates

```
runai template list [flags]
```

### Examples

```
# List all templates from all scopes
runai template list -A

# List templates in a specific project (includes higher-scope templates)
runai template list -p <project-name>

# List templates with pagination
runai template list --page-size 50 --max-items 100

# List templates with a specific output format
runai template list --yaml

# List templates with custom columns
runai template list --columns name,scopeType,workloadType,createdAt
```

### Options

```
  -A, --all                     Show templates from all scopes
      --columns strings         Specify which columns to display (comma-separated list)
  -h, --help                    help for list
      --json                    Output structure JSON
      --max-items int32         set the max number of items to return, default is all of the items
      --next-token next_token   set the token for requesting the next page
      --no-headers              Output structure table without headers
      --no-pagination           return a single page instead of full list, by default set to false
      --page-size int32         set the single page size (max 500) (default 100)
  -p, --project string          Specify the project. Higher-scope templates in the same tenant are included. Use 'runai project set <project>' to set the default.
      --raw                     Return the unprocessed response from the API call
      --set-default-columns     [Experimental] Set the columns flag value to the default output
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

* [runai template](runai_template.md)	 - workload template management

