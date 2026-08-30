package transport

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog"
)

// log must be wired to stderr — stdout carries MCP JSON-RPC frames.
func ServeStdio(ctx context.Context, srv *mcp.Server, log zerolog.Logger) error {
	log.Info().Msg("starting MCP server on stdio")
	return srv.Run(ctx, &mcp.StdioTransport{})
}
