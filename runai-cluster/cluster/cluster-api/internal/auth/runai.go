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
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/common/auth/providers/oidc/configuration"
	"github.com/run-ai/runai/runai-cluster/common/auth/providers/oidc/userinfo"
)

const (
	OptionHeader      = "OPTIONS"
	authzExtractError = "authentication error: %s"
)

func CreateAuthMiddleware(jwksConfig *keyfunc.JWKS, oidcConfigurationProvider configuration.Provider,
	userInfoProvider userinfo.Provider, allowedOrigins []string,
) func(http.Handler) http.Handler {
	log.Debug().Msg("Creating authentication middleware")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Msg("Starting request processing")

			setCORSHeaders(w, r, allowedOrigins)

			if r.Method == OptionHeader {
				log.Debug().Msg("Handling OPTIONS request")

				w.WriteHeader(http.StatusOK)

				return
			}

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
