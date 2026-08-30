package mcphttp

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog"

	httpmetrics "github.com/run-ai/runai/runai-common-packages/metrics/http"
)

func newRouter(getServer func(context.Context) (*mcp.Server, error), cfg Config, log zerolog.Logger) chi.Router {
	log.Info().
		Int("port", cfg.ListenPort).
		Msg("starting MCP server on streamable-http")

	r := chi.NewRouter()
	r.Use(httpmetrics.HttpMiddleware)
	r.Handle(mcpPath, protectedMCPHandler(getServer, cfg.Verifier, log))
	r.Handle(wellKnownPath, metadataHandler(cfg.Verifier))
	r.Get("/healthz", healthz)

	return r
}
