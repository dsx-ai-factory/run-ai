package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog"
	"github.com/run-ai/runai/runai-mcp-server/internal/auth"
	mcphttp "github.com/run-ai/runai/runai-mcp-server/internal/http"
	mcpmetrics "github.com/run-ai/runai/runai-mcp-server/internal/metrics"
	"github.com/run-ai/runai/runai-mcp-server/internal/server"
	"github.com/run-ai/runai/runai-mcp-server/internal/transport"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type transportType string

const (
	transportStdio transportType = "stdio"
	transportHTTP  transportType = "http"
)

type serveConfig struct {
	transport       transportType
	runaiBaseURL    string
	clientID        string
	clientSecret    string
	allowWriteTools bool

	listenPort  int
	metricsPort int
}

func newServeCmd() (*cobra.Command, error) {
	v := viper.New()
	v.BindEnv("transport", "RUNAI_TRANSPORT")
	v.BindEnv("runai-url", "RUNAI_BASE_URL")
	v.BindEnv("client-id", "RUNAI_CLIENT_ID")
	v.BindEnv("client-secret", "RUNAI_CLIENT_SECRET")
	v.BindEnv("allow-write-tools", "RUNAI_ALLOW_WRITE_TOOLS")
	v.BindEnv("listen-port", "RUNAI_MCP_LISTEN_PORT")
	v.BindEnv("metrics-port", "RUNAI_MCP_METRICS_PORT")
	v.SetDefault("listen-port", 8080)
	v.SetDefault("metrics-port", 9090)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the MCP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := serveConfig{
				transport:       transportType(v.GetString("transport")),
				runaiBaseURL:    v.GetString("runai-url"),
				clientID:        v.GetString("client-id"),
				clientSecret:    v.GetString("client-secret"),
				allowWriteTools: v.GetBool("allow-write-tools"),
				listenPort:      v.GetInt("listen-port"),
				metricsPort:     v.GetInt("metrics-port"),
			}
			return runServe(cmd.Context(), cfg)
		},
	}

	cmd.Flags().String("transport", string(transportStdio), "MCP transport: stdio | http")
	cmd.Flags().String("runai-url", "", "Run:ai control-plane base URL (env: RUNAI_BASE_URL)")
	cmd.Flags().Bool("allow-write-tools", false, "Expose tools that mutate Run:ai state. Off by default — only read-only tools are registered.")
	cmd.Flags().Int("listen-port", 0, "HTTP transport: port to listen on (env: RUNAI_MCP_LISTEN_PORT, default 8080)")
	cmd.Flags().Int("metrics-port", 0, "HTTP transport: port to expose Prometheus metrics on (env: RUNAI_MCP_METRICS_PORT, default 9090)")
	for _, name := range []string{
		"transport", "runai-url", "allow-write-tools",
		"listen-port", "metrics-port",
	} {
		if err := v.BindPFlag(name, cmd.Flags().Lookup(name)); err != nil {
			return nil, fmt.Errorf("serve: BindPFlag %q: %w", name, err)
		}
	}

	return cmd, nil
}

func runServe(ctx context.Context, cfg serveConfig) error {
	if cfg.runaiBaseURL == "" {
		return fmt.Errorf("runai base URL is required (--runai-url or RUNAI_BASE_URL)")
	}

	logger := newLogger()

	authProvider, err := buildAuthProvider(ctx, cfg)
	if err != nil {
		return err
	}

	serverCfg := server.Config{
		RunaiBaseURL: cfg.runaiBaseURL,
		Auth:         authProvider,
		Logger:       logger,
		ReadOnly:     !cfg.allowWriteTools,
	}

	switch cfg.transport {
	case transportStdio:
		mcpServer, err := server.New(ctx, serverCfg)
		if err != nil {
			return fmt.Errorf("build mcp server: %w", err)
		}
		return transport.ServeStdio(ctx, mcpServer, logger)
	case transportHTTP:
		if err := mcpmetrics.Init(cfg.metricsPort); err != nil {
			return fmt.Errorf("init metrics: %w", err)
		}
		verifier, err := auth.NewJWKSVerifier(ctx, auth.JWKSVerifierConfig{
			IssuerURL:  cfg.runaiBaseURL,
			HTTPClient: http.DefaultClient,
			Logger:     logger,
		})
		if err != nil {
			return fmt.Errorf("init jwks verifier: %w", err)
		}
		return mcphttp.Serve(ctx, func(ctx context.Context) (*mcp.Server, error) {
			return server.New(ctx, serverCfg)
		}, mcphttp.Config{
			ListenPort: cfg.listenPort,
			Verifier:   verifier,
		}, logger)
	}
	return fmt.Errorf("unknown transport %q (expected: %s | %s)", cfg.transport, transportStdio, transportHTTP)
}

func newLogger() zerolog.Logger {
	return zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
}

func buildAuthProvider(ctx context.Context, cfg serveConfig) (auth.Provider, error) {
	switch cfg.transport {
	case transportStdio:
		if cfg.clientID == "" || cfg.clientSecret == "" {
			return nil, fmt.Errorf("stdio transport requires RUNAI_CLIENT_ID and RUNAI_CLIENT_SECRET environment variables")
		}
		cc, err := auth.NewClientCredentials(ctx, cfg.runaiBaseURL, cfg.clientID, cfg.clientSecret)
		if err != nil {
			return nil, fmt.Errorf("init auth: %w", err)
		}
		return cc, nil
	case transportHTTP:
		return auth.NewBearerPassthrough(), nil
	default:
		return nil, fmt.Errorf("unknown transport %q", cfg.transport)
	}
}
