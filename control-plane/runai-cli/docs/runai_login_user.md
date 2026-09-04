## runai login user

login for local user without browser

### Synopsis

Login to the control plane using a local user without browser

```
runai login user [flags]
```

### Examples

```

# Login with a username. the password will be prompted via stdin afterward (recommended)
runai login user -u <username>

# Login with a username and plain password (not recommended for security reasons)
runai login user --user=user --password=pass

# Login with a username and password (not recommended for security reasons)
runai login user -u=user -p=pass

```

### Options

```
  -h, --help              help for user
  -p, --password string   plaintext password of the given username. not recommended for security reasons
  -u, --user string       the username to login with
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

