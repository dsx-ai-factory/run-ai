/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mcphttp

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
	"github.com/rs/zerolog"

	mcpauth "github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/auth"
)

const (
	mcpPath       = "/mcp"
	wellKnownPath = "/.well-known/oauth-protected-resource" + mcpPath
)

var (
	bearerMethodsSupported = []string{"header"}
	scopesSupported        = []string{"openid", "email"}
)

func protectedMCPHandler(getServer func(context.Context) (*mcp.Server, error), verifier *mcpauth.JWKSVerifier, log zerolog.Logger) http.Handler {
	var (
		mu         sync.RWMutex
		mcpHandler http.Handler
	)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		h := mcpHandler
		mu.RUnlock()

		if h == nil {
			mu.Lock()
			if mcpHandler == nil {
				srv, err := getServer(r.Context())
				if err != nil {
					mu.Unlock()
					log.Error().Err(err).Msg("mcp server init failed")
					http.Error(w, "server not ready", http.StatusServiceUnavailable)
					return
				}
				mcpHandler = mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return srv }, nil)
			}
			h = mcpHandler
			mu.Unlock()
		}
		h.ServeHTTP(w, r)
	})

	secured := auth.RequireBearerToken(verifier.Verify, nil)(mcpauth.RawTokenPropagator(inner))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secured.ServeHTTP(&bearerChallengeWriter{
			ResponseWriter: w,
			challenge:      "Bearer resource_metadata=" + resourceURL(r) + wellKnownPath,
		}, r)
	})
}

type bearerChallengeWriter struct {
	http.ResponseWriter
	challenge string
}

func (w *bearerChallengeWriter) WriteHeader(code int) {
	if code == http.StatusUnauthorized {
		w.ResponseWriter.Header().Set("WWW-Authenticate", w.challenge)
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *bearerChallengeWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func metadataHandler(verifier *mcpauth.JWKSVerifier) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth.ProtectedResourceMetadataHandler(&oauthex.ProtectedResourceMetadata{
			Resource:               resourceURL(r) + mcpPath,
			AuthorizationServers:   []string{verifier.Issuer()},
			BearerMethodsSupported: bearerMethodsSupported,
			ScopesSupported:        scopesSupported,
		}).ServeHTTP(w, r)
	})
}

func resourceURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	} else if proto := r.Header.Get("X-Forwarded-Proto"); proto == "http" || proto == "https" {
		scheme = proto
	}
	host := r.Host
	if fwd := r.Header.Get("X-Forwarded-Host"); fwd != "" && !strings.ContainsAny(fwd, "\r\n") {
		host = fwd
	}
	return scheme + "://" + host
}
