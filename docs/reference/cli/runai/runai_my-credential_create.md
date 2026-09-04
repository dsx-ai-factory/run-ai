# runai my-credential create

create a personal credential

### Synopsis

Create a personal credential. Set --type and pass data with --item using the keys below.

  genericSecret   Keys: key, value (repeat --item for multiple pairs). Requires cluster version 2.22+.
  dockerRegistry  Keys: user, password, url. Requires cluster version 2.22+.
  ngcApiKey       Keys: ngc-api-key. Requires cluster version 2.23+.

```
runai my-credential create SECRET_NAME --type <genericSecret|dockerRegistry|ngcApiKey> [flags]
```

### Examples

```
  runai my-credential create app-config --type genericSecret --item key=API_URL,value=https://api.example.com --item key=DB_PASS,value=secret
  runai my-credential create dockerhub-creds --type dockerRegistry --item user=myuser,password=secret,url=https://index.docker.io/v1/
  runai my-credential create ngc-prod --type ngcApiKey --item ngc-api-key=YOUR_KEY
```

### Options

```
      --description string   Optional description for the credential (e.g. "My API config" or a single word).
  -h, --help                 help for create
      --item stringArray     Key=value pairs; required keys depend on --type (see above). For genericSecret only, repeat to add multiple pairs to one credential.
      --type string          Credential type: genericSecret, dockerRegistry, or ngcApiKey (required)
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

