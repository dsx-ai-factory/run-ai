## runai application delete

Delete an AI application

```
runai application delete APPLICATION_NAME [flags]
```

### Examples

```
# Delete an application (with confirmation prompt)
runai application delete my-grafana -p team-alpha

# Delete without confirmation
runai application delete my-grafana -p team-alpha --force
```

### Options

```
      --force[=true|false]   Skip confirmation prompt
  -h, --help[=true|false]    help for delete
  -p, --project string       Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default.
  -y, --yes[=true|false]     Alias for --force
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

