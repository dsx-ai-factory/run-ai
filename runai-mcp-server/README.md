# runai-mcp-server

An [MCP](https://modelcontextprotocol.io) server that exposes Run:ai control-plane operations as tools for AI assistants.

## Configuration

| Flag | Env var | Default | Description |
|---|---|---|---|
| `--runai-url` | `RUNAI_BASE_URL` | — | Run:ai control-plane base URL (**required**) |
| _(env-only)_ | `RUNAI_CLIENT_ID` | — | Service-account client ID (**required** for `stdio`) |
| _(env-only)_ | `RUNAI_CLIENT_SECRET` | — | Service-account client secret (**required** for `stdio`) |
| `--transport` | `RUNAI_TRANSPORT` | `stdio` | MCP transport: `stdio` or `http` |
| `--allow-write-tools` | `RUNAI_ALLOW_WRITE_TOOLS` | `false` | Expose tools that mutate Run:ai state |
| `--listen-port` | `RUNAI_MCP_LISTEN_PORT` | `8080` | Port to listen on (HTTP transport only) |

**stdio transport** authenticates to the Run:ai API using the OAuth2 client credentials flow (`RUNAI_CLIENT_ID` / `RUNAI_CLIENT_SECRET`). Credentials are env-only — not accepted on the command line.

**HTTP transport** delegates authentication to the MCP client. The client obtains a Bearer token from the Run:ai identity manager and presents it on each request; the server validates the token and forwards it downstream.

## Running locally

**stdio** (default):

```bash
RUNAI_BASE_URL=https://<cluster-url> \
RUNAI_CLIENT_ID=<client-id> \
RUNAI_CLIENT_SECRET=<client-secret> \
go run ./cmd/runai-mcp serve
```

**HTTP transport** — client credentials are not required; inbound Bearer tokens are validated automatically via OIDC discovery from `RUNAI_BASE_URL`:

```bash
RUNAI_BASE_URL=https://<cluster-url> \
go run ./cmd/runai-mcp serve --transport http --listen-port 8080
```

## MCP Inspector

```bash
make inspect \
  RUNAI_BASE_URL=https://<cluster-url> \
  RUNAI_CLIENT_ID=<client-id> \
  RUNAI_CLIENT_SECRET=<client-secret>
```

## Client configuration

### Claude Code

Add to `~/.claude/settings.json` (global) or `.claude/settings.json` (project):

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
        "registry.example.com/runai-mcp-server:<tag>",
      ],
      "env": {
        "RUNAI_BASE_URL": "https://<cluster-url>",
        "RUNAI_CLIENT_ID": "<client-id>",
        "RUNAI_CLIENT_SECRET": "<client-secret>"
      }
    }
  }
}
```

### Cursor

Add to `~/.cursor/mcp.json` (global) or `.cursor/mcp.json` (project):

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
        "registry.example.com/runai-mcp-server:<tag>",
      ],
      "env": {
        "RUNAI_BASE_URL": "https://<cluster-url>",
        "RUNAI_CLIENT_ID": "<client-id>",
        "RUNAI_CLIENT_SECRET": "<client-secret>"
      }
    }
  }
}
```

## Adding new tools

See [internal/tools/README.md](internal/tools/README.md) for the tool-authoring guide — naming and design guidelines, the `runaimcp` framework, error categorization, and the lint test contract.

## Building

```bash
# Local binaries (linux/amd64 + linux/arm64)
make build

# Docker images
make docker DOCKER_TAG=1.0.0
```

## Docker

```bash
docker run --rm -i \
  -e RUNAI_BASE_URL=https://<cluster-url> \
  -e RUNAI_CLIENT_ID=<client-id> \
  -e RUNAI_CLIENT_SECRET=<client-secret> \
  registry.example.com/runai-mcp-server:0.0.0-dev serve
```
