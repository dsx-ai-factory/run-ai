package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "runai-mcp",
		Short: "Run:ai MCP server",
		Long:  "Run:ai Model Context Protocol server. Exposes Run:ai control-plane operations as MCP tools.",
	}
	serveCmd, err := newServeCmd()
	if err != nil {
		return nil, err
	}
	cmd.AddCommand(serveCmd)
	return cmd, nil
}
