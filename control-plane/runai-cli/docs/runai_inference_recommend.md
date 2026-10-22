## runai inference recommend

Recommend GPU allocation for a NIM test/dev workload as JSON

### Synopsis

Recommend GPU allocation for an exact NIM image tag using the selected cluster. Does not submit a workload or reserve resources.

```
runai inference recommend [flags]
```

### Options

```
      --gpu-type string   Exact cluster GPU type (default: all reported types)
  -h, --help              help for recommend
      --image string      Untagged nvcr.io/nim/organization/model image
      --tag string        Exact image tag
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

