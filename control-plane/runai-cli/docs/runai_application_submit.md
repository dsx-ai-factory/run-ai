## runai application submit

Submit an AI application

```
runai application submit APPLICATION_NAME [flags]
```

### Examples

```
# Basic: submit from a chart URL
runai application submit my-grafana -p team-alpha \
  --chart-url https://charts.bitnami.com/bitnami/grafana-11.0.0.tgz

# With value overrides
runai application submit my-grafana -p team-alpha \
  --chart-url https://charts.bitnami.com/bitnami/grafana-11.0.0.tgz \
  --values ./my-values.yaml \
  --set replicaCount=2 \
  --set persistence.enabled=false
```

### Options

```
      --chart-url string    URL to a packaged Helm chart (.tgz)
      --credential string   NGC API key credential asset name for private chart URLs
  -h, --help                help for submit
  -p, --project string      Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default.
      --set stringArray     Inline key=value overrides (nested keys supported), merged on top of --values
      --user-credential     Treat --credential as a personal credential owned by the current user
      --values string       Path to a YAML values file for Helm chart overrides
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

