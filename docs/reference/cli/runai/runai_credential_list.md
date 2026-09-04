# runai credential list

list credential assets

```
runai credential list [flags]
```

### Examples

```
# List all credential assets
runai credential list -A

# List credential assets in a specific project
runai credential list -p <project_name>

# List credential assets filtered by type
runai credential list --type <type>

# List credential assets with a specific output format
runai credential list --yaml
```

### Options

```
  -A, --all                   Show resources from all projects
      --columns strings       Specify which columns to display (comma-separated list)
  -h, --help                  help for list
      --json                  Output structure JSON
      --no-headers            Output structure table without headers
  -p, --project string        Specify the project. Higher-scope credentials in the same tenant are included. Use 'runai project set <project>' to set the default.
      --raw                   Return the unprocessed response from the API call
      --set-default-columns   [Experimental] Set the columns flag value to the default output
      --table                 Output structure table
      --type string           Filter by credential type (accessKey, dockerRegistry, password, genericSecret, ngcApiKey).
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

* [runai credential](runai_credential.md)	 - Credential asset management

