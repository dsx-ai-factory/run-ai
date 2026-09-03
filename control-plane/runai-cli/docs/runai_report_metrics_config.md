## runai report metrics config

metrics configuration

```
runai report metrics config [flags]
```

### Options

```
      --age int          metrics max file age (default 14)
      --files int        metrics max file number (default 30)
  -h, --help             help for config
      --metrics enable   metrics enable flag (enabled|disabled)
      --size int         metrics max file size (default 10)
      --type reporter    report generated type (none|logger|local)
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

* [runai report metrics](runai_report_metrics.md)	 - [Experimental] metrics management

