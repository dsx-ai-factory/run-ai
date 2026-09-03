# runai project

project management

```
runai project [flags]
```

### Options

```
  -h, --help                    help for project
      --interactive enable      set interactive mode (enabled|disabled)
      --max-items int32         set the max number of items to return, default is all of the items
      --next-token next_token   set the token for requesting the next page
      --no-pagination           return a single page instead of full list, by default set to false
      --page-size int32         set the single page size (max 500) (default 100)
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
* [runai project create](runai_project_create.md)	 - Create a new project
* [runai project list](runai_project_list.md)	 - list available project
* [runai project set](runai_project_set.md)	 - set default project name

