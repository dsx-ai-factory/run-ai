/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mcphttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog"

	"github.com/run-ai/runai/runai-mcp-server/internal/auth"
)

type Config struct {
	ListenPort int
	Verifier   *auth.JWKSVerifier
}

func Serve(ctx context.Context, getServer func(context.Context) (*mcp.Server, error), cfg Config, log zerolog.Logger) error {
	if cfg.ListenPort < 1 || cfg.ListenPort > 65535 {
		return fmt.Errorf("http: listen port must be 1–65535, got %d", cfg.ListenPort)
	}
	if cfg.Verifier == nil {
		return errors.New("http: verifier required")
	}

	r := newRouter(getServer, cfg, log)

	httpSrv := &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.ListenPort),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpSrv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("http: shutdown: %w", err)
		}
		return nil
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("http: serve: %w", err)
	}
}
