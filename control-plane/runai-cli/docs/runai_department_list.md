## runai department list

list available departments

```
runai department list [flags]
```

### Options

```
  -h, --help[=true|false]         help for list
      --json[=true|false]         Output structure JSON
      --limit int32               Maximum number of results to return. Used for pagination. (default 50)
      --no-headers[=true|false]   Output structure table without headers
      --offset int32              Offset for paginated results. Default is 0 (start from first result).
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

* [runai department](runai_department.md)	 - department management

