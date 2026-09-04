# runai application describe

Describe an AI application

```
runai application describe APPLICATION_NAME [flags]
```

### Examples

```
# Describe an AI application
runai application describe my-grafana -p team-alpha

# Describe in JSON format
runai application describe my-grafana -p team-alpha -o json
```

### Options

```
  -h, --help             help for describe
  -o, --output string    Output format: table, json, or yaml. Default is table. (default "table")
  -p, --project string   Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default.
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

* [runai application](runai_application.md)	 - ai application management

