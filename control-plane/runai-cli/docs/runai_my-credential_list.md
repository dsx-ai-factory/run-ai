## runai my-credential list

list personal credentials

```
runai my-credential list [flags]
```

### Examples

```
# List all personal credentials
runai my-credential list

# List only Docker registry credentials
runai my-credential list --type dockerRegistry
```

### Options

```
  -h, --help          help for list
      --type string   Filter by credential type (genericSecret, dockerRegistry, ngcApiKey)
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

* [runai my-credential](runai_my-credential.md)	 - manage personal credentials associated with the current user

