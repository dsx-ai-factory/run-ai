## runai project create

Create a new project

```
runai project create PROJECT_NAME [flags]
```

### Examples

```
# Basic Project Creation
runai project create research-team-a --resources gpu-deserved=2

# Project with Single Nodepool
runai project create ml-team \
  --resources nodepool=gpu-pool,gpu-deserved=2,gpu-limit=4,priority=High \
  --description "Machine Learning Team Project"

# Project with Multiple Nodepools
runai project create data-science \
  --resources nodepool=gpu-pool-a,gpu-deserved=2,gpu-limit=4,priority=High \
  --resources nodepool=gpu-pool-b,gpu-deserved=1,gpu-limit=2,priority=Medium \
  --description "Data Science Team Project"

# Project with Custom Namespace
runai project create research \
  --resources nodepool=gpu-pool,gpu-deserved=1,gpu-limit=2,priority=Medium \
  --namespace research-ns \
  --description "Research Team Project"

# Project with Over-Quota
runai project create over-quota \
  --resources gpu-deserved=5,gpu-limit=10,priority=Medium

# Project with Unlimited Over-Quota
runai project create unlimited-over-quota \
  --resources gpu-deserved=0,gpu-limit=-1
```

### Options

```
      --department string     Department for the project (default "default")
      --description string    Description for the project
  -h, --help                  help for create
      --namespace string      Namespace for the project
      --resources resources   Resources for the project (can be specified multiple times). Format: nodepool=<name>,gpu-deserved=<num>,gpu-limit=<num>,priority=<level>
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

* [runai project](runai_project.md)	 - project management

