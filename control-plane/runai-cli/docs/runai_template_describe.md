## runai template describe

describe a workload template

```
runai template describe [TEMPLATE_NAME] [flags]
```

### Examples

```
# Describe a template with a default project
runai template describe my-template

# Describe a template in a specific project
runai template describe my-template -p my-project

# Describe a template with specific output format
runai template describe my-template -o json
```

### Options

```
  -h, --help             help for describe
  -o, --output string    Output format: table, json, or yaml. Default is table. (default "table")
  -p, --project string   Specify the project. Higher-scope templates in the same tenant are included. Use 'runai project set <project>' to set the default.
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

