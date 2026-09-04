# Deploy the MCP Server

## Requirements

* NVIDIA Run:ai cluster version 2.25 or later.
* The MCP server image is available on the [NGC catalog](https://catalog.ngc.nvidia.com/orgs/nvidia/runai/containers/runai-mcp-server) with guest access — no NGC account is required to pull it.

## Deployment Modes

The MCP server supports two deployment modes:

| Mode | Who sets it up | How the AI tool connects |
|---|---|---|
| **Local (STDIO)** | Each user independently | Docker container on the user's machine |
| **Remote (HTTP)** | Cluster administrator | Centrally deployed server, shared across users |

## Remote Mode (HTTP)

In remote mode, a cluster administrator deploys the MCP server centrally and exposes it at an FQDN. Users connect their AI tools to the shared server using OAuth, without managing service account credentials themselves.

### Network Requirements

The following network paths must be open:

* User's AI agent → MCP server
* MCP server → NVIDIA Run:ai control plane

### Deploy the MCP Server

Run the MCP server container with HTTP transport enabled:

```bash
docker run -d -p 8080:8080 \
  -e RUNAI_BASE_URL=https://<your-runai-url> \
  -e RUNAI_TRANSPORT=http \
  nvcr.io/nvidia/runai/runai-mcp-server
```

Expose the container at an FQDN accessible to users on your network. The MCP endpoint is available at `<fqdn>/mcp`.

## Configuration Reference

The following environment variables apply to both deployment modes unless noted otherwise:

| Variable | Description | Applies to | Default |
|---|---|---|---|
| `RUNAI_BASE_URL` | NVIDIA Run:ai control plane URL | Both | Required |
| `RUNAI_CLIENT_ID` | Service account client ID | STDIO only | Required |
| `RUNAI_CLIENT_SECRET` | Service account client secret | STDIO only | Required |
| `RUNAI_TRANSPORT` | Transport mode: `stdio` or `http` | Both | `stdio` |
| `RUNAI_MCP_LISTEN_PORT` | Listen port | HTTP only | `8080` |
| `RUNAI_ALLOW_WRITE_TOOLS` | Expose write (mutating) tools | Both | `false` |
