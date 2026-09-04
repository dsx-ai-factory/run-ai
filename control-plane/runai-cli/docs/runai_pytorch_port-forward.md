## runai pytorch port-forward

forward one or more local ports to a pytorch training workload

```
runai pytorch port-forward [WORKLOAD_NAME] [flags]
```

### Examples

```
# Forward connections from localhost:8080 to a workload on port 8090:
runai training pytorch port-forward <workload-name> --port 8080:8090 --address localhost

# Forward connections from 0.0.0.0:8080 to a workload on port 8080:
runai training pytorch port-forward <workload-name> --port 8080 --address 0.0.0.0 [requires privileges]

# Forward multiple connections from localhost:8080 to a workload on port 8090 and from localhost:6443 to a workload on port 443:
runai training pytorch port-forward <workload-name> --port 8080:8090 --port 6443:443 --address localhost
```

### Options

```
      --address string                 Bind the workload to a specific local interface or host. E.g., --address localhost or --address 0.0.0.0. (default "localhost")
  -h, --help                           help for port-forward
      --pod string                     Workload pod ID for port-forward, default: master
      --pod-running-timeout duration   Timeout for pod to reach running state (e.g. 5s, 2m, 3h).
      --port stringArray               port
  -p, --project string                 Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default.
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

* [runai pytorch](runai_pytorch.md)	 - alias for pytorch management

