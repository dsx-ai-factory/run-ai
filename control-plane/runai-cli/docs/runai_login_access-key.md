## runai login access-key

login as a service account or a user using an access key

```
runai login access-key [flags]
```

### Examples

```
  
  # Login interactive using access key
  runai login access-key

  # Login using access key credentials
  runai login access-key --client-id=<client_id> --secret=<secret> --interactive=disabled

  # Login and save access key credentials
  runai login access-key --client-id=<client_id> --secret=<secret> --interactive=disabled --save

```

### Options

```
      --client-id string     client ID for service account or user
  -h, --help                 help for access-key
      --interactive enable   set interactive mode (enabled|disabled)
      --save                 save access key credentials in config file
      --secret string        access key secret
      --secret-file string   use access key secret from file
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

* [runai login](runai_login.md)	 - login to the control plane

