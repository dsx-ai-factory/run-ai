---
---

# Command Reference and Troubleshooting

The `runai-installer` commands and flags are identical across same-cluster, multi-cluster, connected, and air-gapped installations - run each command against the relevant cluster with its own `--config` and `--kubeconfig`.

## Global Flags

These flags apply to all `runai-installer` subcommands.

| Flag             | Description                                     | Default           |
| ---------------- | ----------------------------------------------- | ----------------- |
| `--logs-dir`     | Directory where installer log files are written | Current directory |
| `--step-timeout` | Maximum time allowed for each installation step | `7m`              |

## `install` Command

The `install` command installs NVIDIA Run:ai from a configuration file.

```bash
./runai-installer install [flags]
```

| Flag | Description | Default |
|---|---|---|
| `--config` | Path to the installation configuration file **(required)** | - |
| `--dry-run` | Plan the installation without applying any changes | `false` |
| `--preflight-only` | Run preflight checks and exit without installing | `false` |
| `--skip-preflight` | Skip preflight checks and proceed with installation | `false` |
| `-y`, `--yes` | Run non-interactively without confirmation prompts | `false` |
| `--output` | Output format: `text` or `json` | `text` on a TTY, `json` otherwise |
| `--events-file` | Write NDJSON installation events to a file | - |
| `--status-file` | Write the final installation summary to a file | - |
| `--on-failure` | Failure policy: `abort` or `skip` | `abort` |
| `--kubeconfig` | Path to a kubeconfig file | - |
| `--ngc-token-path` | Path to an NGC API key file (overrides `registry` in the configuration file) | - |
| `--execution-timeout` | Maximum time allowed for the installation | `50m` |

## Troubleshooting

### Cluster Cannot Reach the Control Plane

If cluster registration fails or a cluster does not show as **Connected**, verify that the NVIDIA Run:ai cluster's nodes can reach the control plane FQDN over HTTPS (port 443):

```bash
curl -v https://<control-plane-domain>
```

### Installation Step Timed Out

Increase the per-step timeout for slow networks or large image pulls:

```bash
./runai-installer install \
  --config config.yaml \
  --step-timeout 15m
```

### View Installer Logs

```bash
# Stream logs in real time
tail -f ./logs/runai-installer.log

# Show only errors
grep -i error ./logs/runai-installer.log

# Filter Helm operations
grep "\[Helm\]" ./logs/runai-installer.log

# Filter Kubernetes API calls
grep "\[K8s\]" ./logs/runai-installer.log
```

### Registry Authentication Failed

Verify that the NGC API key file is valid and readable:

```bash
cat /path/to/ngc-token
```

The file must contain only the API key. The `runai-installer` CLI automatically trims leading and trailing whitespace.
