/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog"

	"github.com/run-ai/runai/runai-mcp-server/internal/auth"
	mcpmetrics "github.com/run-ai/runai/runai-mcp-server/internal/metrics"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools"
)

const serverName = "runai-mcp"

// serverVersion is injected at build time via -ldflags (see Makefile).
var serverVersion = "0.0.0-dev"

type Config struct {
	RunaiBaseURL string
	Auth         auth.Provider
	Logger       zerolog.Logger
	ReadOnly     bool
}

func New(ctx context.Context, cfg Config) (*mcp.Server, error) {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)

	client := runaiclient.New(cfg.RunaiBaseURL, cfg.Auth)

	apiVersion, err := probeControlPlaneVersion(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("probe control-plane version: %w", err)
	}
	cfg.Logger.Info().Str("api_version", apiVersion).Msg("control-plane version detected")

	registry := runaimcp.New(srv, runaimcp.Options{
		Logger:     cfg.Logger,
		ReadOnly:   cfg.ReadOnly,
		APIVersion: apiVersion,
	})
	tools.Register(registry, client)

	knownTools := make(map[string]bool, len(registry.ToolNames()))
	for _, name := range registry.ToolNames() {
		knownTools[name] = true
	}
	srv.AddReceivingMiddleware(mcpmetrics.MCPMiddleware(knownTools))

	return srv, nil
}

func probeControlPlaneVersion(ctx context.Context, client *runaiclient.Client) (string, error) {
	resp, _, err := client.BFF.VersionAPI.GetVersion(ctx).Execute()
	if err != nil {
		return "", err
	}
	if resp == nil || resp.Version == "" {
		return "", fmt.Errorf("empty version in /_version response")
	}
	return resp.Version, nil
}
