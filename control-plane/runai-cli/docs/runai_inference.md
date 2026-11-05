## runai inference

inference management

### Options

```
  -h, --help   help for inference
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
* [runai inference bash](runai_inference_bash.md)	 - open a bash shell in an inference workload
* [runai inference delete](runai_inference_delete.md)	 - delete an inference workload
* [runai inference describe](runai_inference_describe.md)	 - describe an inference workload
* [runai inference distributed](runai_inference_distributed.md)	 - Runs multiple coordinated inference processes across multiple nodes. Required for models too large to run on a single node.
* [runai inference exec](runai_inference_exec.md)	 - execute a command in an inference workload
* [runai inference list](runai_inference_list.md)	 - list inference workloads
* [runai inference logs](runai_inference_logs.md)	 - view logs of an inference workload
* [runai inference nim](runai_inference_nim.md)	 - [Experimental] Runs NVIDIA NIM (NVIDIA Inference Microservices) workloads. Optimized for deploying foundation models.
* [runai inference port-forward](runai_inference_port-forward.md)	 - forward one or more local ports to an inference workload
* [runai inference recommend](runai_inference_recommend.md)	 - Recommend GPU allocation for a NIM test/dev workload as JSON
* [runai inference standard](runai_inference_standard.md)	 - Runs a single inference process on one node. Suitable for smaller models or simpler inference tasks.
* [runai inference submit](runai_inference_submit.md)	 - submit an inference workload
* [runai inference update](runai_inference_update.md)	 - update an inference workload

