## runai datasource list

list data source assets

```
runai datasource list [flags]
```

### Examples

```
# List all data source assets
runai datasource list -A

# List data source assets in a specific project
runai datasource list -p <project_name>

# List data source assets with a specific output format
runai datasource list --yaml
```

### Options

```
  -A, --all[=true|false]                   Show resources from all projects
      --columns strings                    Specify which columns to display (comma-separated list)
  -h, --help[=true|false]                  help for list
      --json[=true|false]                  Output structure JSON
      --no-headers[=true|false]            Output structure table without headers
  -p, --project string                     Specify the project. Higher-scope data-sources in the same tenant are included. Use 'runai project set <project>' to set the default.
      --raw[=true|false]                   Return the unprocessed response from the API call
      --set-default-columns[=true|false]   [Experimental] Set the columns flag value to the default output
      --table[=true|false]                 Output structure table
      --type string                        Filter by data source type (nfs, pvc, s3, git, hostPath, configMap, secretVolume, dataVolume)
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

* [runai datasource](runai_datasource.md)	 - Data Source asset management

