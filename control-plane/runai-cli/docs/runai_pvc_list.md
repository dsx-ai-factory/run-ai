## runai pvc list

List PVC

```
runai pvc list [flags]
```

### Examples

```
# List of PVC by project with table output
runai pvc list -p <project-name>

# List of PVC by project with table JSON output
runai pvc list -p <project-name> --json

# List of PVC by project with YAML  output
runai pvc list -p <project-name> --yaml
```

### Options

```
  -h, --help[=true|false]         help for list
      --json[=true|false]         Output structure JSON
      --no-headers[=true|false]   Output structure table without headers
  -p, --project string            Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default.
      --table[=true|false]        Output structure table
      --yaml[=true|false]         Output structure YAML
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

* [runai pvc](runai_pvc.md)	 - PVC management

