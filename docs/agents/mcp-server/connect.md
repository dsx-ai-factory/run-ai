# Connect Your AI Tool

## Local Mode (STDIO)

In local mode, each user runs the MCP server as a Docker container on their own machine. The container authenticates to the NVIDIA Run:ai control plane using an NVIDIA Run:ai service account.

### Prerequisites

* Docker installed on your machine.
* An NVIDIA Run:ai service account with the appropriate permissions. See [Service Accounts](../../infrastructure-setup/authentication/service-accounts.md) for instructions.

### Credential Security

Store your service account credentials in environment variables rather than hardcoding them in config files. The configuration examples below pass them via `env` so they are resolved at runtime from your shell environment and not written to disk.

### Configure Your AI Tool

Replace the following placeholders before applying the configuration:

* `<your-runai-url>` with your NVIDIA Run:ai control plane URL
* `<client-id>` and `<client-secret>` with your service account credentials

<Tabs>

<Tab title="Claude Code">

Add the following to `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "runai": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "-e", "RUNAI_BASE_URL",
        "-e", "RUNAI_CLIENT_ID",
        "-e", "RUNAI_CLIENT_SECRET",
        "nvcr.io/nvidia/runai/runai-mcp-server"
      ],
      "env": {
        "RUNAI_BASE_URL": "https://<your-runai-url>",
        "RUNAI_CLIENT_ID": "<client-id>",
        "RUNAI_CLIENT_SECRET": "<client-secret>"
      }
    }
  }
}
```

</Tab>
<Tab title="Cursor">

Add the following to `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "runai": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "-e", "RUNAI_BASE_URL",
        "-e", "RUNAI_CLIENT_ID",
        "-e", "RUNAI_CLIENT_SECRET",
        "nvcr.io/nvidia/runai/runai-mcp-server"
      ],
      "env": {
        "RUNAI_BASE_URL": "https://<your-runai-url>",
        "RUNAI_CLIENT_ID": "<client-id>",
        "RUNAI_CLIENT_SECRET": "<client-secret>"
      }
    }
  }
}
```

</Tab>
<Tab title="VS Code / GitHub Copilot">

Add the following to `.vscode/mcp.json`:

```json
{
  "servers": {
    "runai": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "-e", "RUNAI_BASE_URL",
        "-e", "RUNAI_CLIENT_ID",
        "-e", "RUNAI_CLIENT_SECRET",
        "nvcr.io/nvidia/runai/runai-mcp-server"
      ],
      "env": {
        "RUNAI_BASE_URL": "https://<your-runai-url>",
        "RUNAI_CLIENT_ID": "<client-id>",
        "RUNAI_CLIENT_SECRET": "<client-secret>"
      }
    }
  }
}
```

</Tab>
<Tab title="Codex">

Add the following to `~/.codex/config.toml`:

```toml
[mcp_servers.runai]
command = "docker"
args = [
  "run", "--rm", "-i",
  "-e", "RUNAI_BASE_URL",
  "-e", "RUNAI_CLIENT_ID",
  "-e", "RUNAI_CLIENT_SECRET",
  "nvcr.io/nvidia/runai/runai-mcp-server"
]
env = { RUNAI_BASE_URL = "https://<your-runai-url>", RUNAI_CLIENT_ID = "<client-id>", RUNAI_CLIENT_SECRET = "<client-secret>" }
```

</Tab>

</Tabs>

## Remote Mode (HTTP)

Once your administrator has [deployed the MCP server](deploy.md), configure your AI tool using the FQDN they provide.

<Tabs>

<Tab title="Claude Code">

**Option 1 — CLI:**

```bash
claude mcp add --transport http --client-id runai-mcp runai https://<fqdn>/mcp
```

**Option 2 — Config file:**

Add the following to `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "runai": {
      "type": "http",
      "url": "https://<fqdn>/mcp",
      "oauth": {
        "clientId": "runai-mcp"
      }
    }
  }
}
```

</Tab>
<Tab title="Cursor">

Add the following to `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "runai": {
      "url": "https://<fqdn>/mcp",
      "auth": {
        "CLIENT_ID": "runai-mcp"
      }
    }
  }
}
```

</Tab>
<Tab title="VS Code / GitHub Copilot">

Add the following to `.vscode/mcp.json`:

```json
{
  "servers": {
    "runai": {
      "type": "http",
      "url": "https://<fqdn>/mcp",
      "oauth": {
        "clientId": "runai-mcp"
      }
    }
  }
}
```

</Tab>

</Tabs>
