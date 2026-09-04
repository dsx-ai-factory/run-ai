## runai training standard logs

view logs of a standard training workload

```
runai training standard logs [WORKLOAD_NAME] [flags]
```

### Examples

```
# Get logs for a workload
runai training standard logs <workload-name>

# Get logs for a specific pod in a workload
runai training standard logs <workload-name> --pod=<pod-name>

# Get logs for a specific container in a workload
runai training standard logs <workload-name> --container=<container-name>

# Get the last 100 lines of logs
runai training standard logs <workload-name> --tail=100

# Get logs with timestamps
runai training standard logs <workload-name> --timestamps

# Follow the logs
runai training standard logs <workload-name> --follow

# Get logs for the previous instance of the workload
runai training standard logs <workload-name> --previous

# Limit the logs to 1024 bytes
runai training standard logs <workload-name> --limit-bytes=1024

# Get logs since the last 5 minutes
runai training standard logs <workload-name> --since=300s

# Get logs since a specific timestamp
runai training standard logs <workload-name> --since-time=2023-05-30T10:00:00Z

# Wait up to 30 seconds for a workload to be ready for logs
runai training standard logs <workload-name> --wait-timeout=30s
```

### Options

```
  -c, --container string        The name of the container within the pod.
  -f, --follow                  Stream logs in real time.
  -h, --help                    help for logs
      --limit-bytes int         Maximum bytes of logs to return. Defaults to no limit (e.g, 20, 2000).
      --name string             The name of the workload. Must match the name used at submission.
      --pod string              The pod ID. If not specified, the first pod will be used.
      --previous                Show logs from the previous instance of the pod, if it was restarted.
  -p, --project string          Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default.
      --since duration          Show logs newer than a relative duration (e.g., 5s, 2m, 3h).
      --since-time string       Show logs generated after a specific RFC3339 timestamp (e.g. 2025-05-25T14:30:00Z).
  -t, --tail int                The number of log lines to display from the end of the logs. Use -1 to show all lines. (default -1)
      --timestamps              Include timestamps in log output.
      --wait-timeout duration   Timeout while waiting for the workload to become ready for log streaming (e.g., 5s, 2m, 3h).
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

* [runai training standard](runai_training_standard.md)	 - standard training management

