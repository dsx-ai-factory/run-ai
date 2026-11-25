/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/MicahParks/keyfunc"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/auth/oidc/configuration"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/auth/oidc/userinfo"
	"github.com/rs/zerolog/log"
)

const (
	OptionHeader      = "OPTIONS"
	authzExtractError = "authentication error: %s"
)

// CreateCORSMiddleware must be registered on the root router, before any route: chi answers
// an OPTIONS request to a GET-only route with 405 before route-group middleware runs, so a
// CORS check placed in a group never sees the browser's preflight.
func CreateCORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			setCORSHeaders(w, r, allowedOrigins)

			// Preflights carry no credentials, so they are answered here, before authentication.
			if r.Method == OptionHeader {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func CreateAuthMiddleware(jwksConfig *keyfunc.JWKS, oidcConfigurationProvider configuration.Provider,
	userInfoProvider userinfo.Provider,
) func(http.Handler) http.Handler {
	log.Debug().Msg("Creating authentication middleware")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Msg("Starting request processing")

			var authInfo *Info

			token, err := ExtractAuthTokenFromRequest(r)
			if err != nil {
				log.Error().Err(err).Msg("Failed to extract auth token from request")

				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(err.Error()))

				return
			}

			log.Debug().Msg("Successfully extracted auth token")

			authInfo, err = GetAuthInfo(jwksConfig, userInfoProvider, token)
			if err != nil {
				log.Error().Err(err).Msg("Failed to get auth info")

				w.WriteHeader(http.StatusUnauthorized)
				_, _ = fmt.Fprintf(w, authzExtractError, err.Error())

				return
			}

			log.Debug().Interface("authInfo", authInfo).Msg("Successfully obtained auth info")

			ctx := ComposeAuthContext(r.Context(), authInfo)
			rWithAuthClient := r.WithContext(ctx)

			log.Info().Msg("Request authenticated successfully")

			next.ServeHTTP(w, rWithAuthClient)
		})
	}
}

func setCORSHeaders(w http.ResponseWriter, req *http.Request, allowedOrigins []string) {
	origin := req.Header.Get("Origin")
	if origin == "" {
		return
	}
	w.Header().Set("Vary", "Origin")
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			// The UI sends requests withCredentials; safe only because origin is an exact allowlist match.
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
			w.Header().Set("Access-Control-Allow-Headers",
				"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, authinfo")
			return
		}
	}
	log.Warn().Str("origin", origin).Msg("CORS request from non-allowed origin")
}

func OriginFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}
