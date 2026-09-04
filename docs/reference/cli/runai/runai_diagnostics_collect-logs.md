# runai diagnostics collect-logs

collect diagnostic logs from the Kubernetes cluster for troubleshooting or sharing with NVIDIA Run:ai Support

```
runai diagnostics collect-logs [flags]
```

### Examples

```
# Collect logs from all default namespaces
runai diagnostics collect-logs

# Collect logs to a specific output directory
runai diagnostics collect-logs --output-dir /tmp/logs

# Collect logs without previous logs from pods
runai diagnostics collect-logs --no-previous

# Collect logs from specific namespaces
runai diagnostics collect-logs --namespaces runai,runai-backend

```

### Options

```
  -h, --help                   help for collect-logs
      --namespaces enumSlice   Comma-separated list of namespaces to collect logs from (default [runai,runai-reservation,runai-backend,training-operator,knative,gpu-operator,nim-operator])
      --no-previous            Do not collect previous logs from pods
      --output-dir string      Directory to output the archive file (default "./")
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

* [runai diagnostics](runai_diagnostics.md)	 - system diagnostics and troubleshooting tools

