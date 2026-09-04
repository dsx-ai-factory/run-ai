# runai login

login to the control plane

```
runai login [flags]
```

### Examples

```

  # Login using browser
  runai login

  # Login using SSO with remote browser
  runai login sso
  runai login remote-browser

  # Login using username and password without browser
  runai login user -u <username> 

  # Login using browser with specific port and host
  runai login --listen-port=43121 --listen-host=localhost

```

### Options

```
  -h, --help                 help for login
      --listen-host string   the host to listen on for the authentication callback (for browser mode only) (default "localhost")
      --listen-port int      the port to listen on for the authentication callback (for browser mode only) (default 43121)
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

* [runai](runai.md)	 - Run:ai Command-line Interface
* [runai login access-key](runai_login_access-key.md)	 - login as a service account or a user using an access key
* [runai login sso](runai_login_sso.md)	 - login using sso without browser
* [runai login user](runai_login_user.md)	 - login for local user without browser

