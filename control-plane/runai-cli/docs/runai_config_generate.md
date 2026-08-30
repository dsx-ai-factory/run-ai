## runai config generate

generate config file

```
runai config generate [flags]
```

### Examples

```
# JSON configuration file
runai config generate -f path/to/config.json
# YAML configuration file
runai config generate -f path/to/config.yaml

```

### Options

```
  -f, --file string   file path for generating the configuration
  -h, --help          help for generate
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

* [runai config](runai_config.md)	 - configuration management

