## runai inference distributed

Runs multiple coordinated inference processes across multiple nodes. Required for models too large to run on a single node.

### Options

```
  -h, --help   help for distributed
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

* [runai inference](runai_inference.md)	 - inference management
* [runai inference distributed bash](runai_inference_distributed_bash.md)	 - open a bash shell in a distributed inference workload
* [runai inference distributed delete](runai_inference_distributed_delete.md)	 - delete a distributed inference workload
* [runai inference distributed describe](runai_inference_distributed_describe.md)	 - describe a distributed inference workload
* [runai inference distributed exec](runai_inference_distributed_exec.md)	 - execute a command in a distributed inference workload
* [runai inference distributed list](runai_inference_distributed_list.md)	 - list distributed inference workloads
* [runai inference distributed logs](runai_inference_distributed_logs.md)	 - view logs of a distributed inference workload
* [runai inference distributed port-forward](runai_inference_distributed_port-forward.md)	 - forward one or more local ports to a distributed inference workload
* [runai inference distributed scale](runai_inference_distributed_scale.md)	 - scale a distributed inference workload
* [runai inference distributed submit](runai_inference_distributed_submit.md)	 - submit a distributed inference workload

