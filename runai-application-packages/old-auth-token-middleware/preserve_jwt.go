/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package old_auth_token_middleware

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	HeaderAuthorization            = "Authorization"
	ContextJwtToken     contextKey = "JwtToken"
	BearerPrefix                   = "Bearer "
)

// PreserveJwtToken is not to be used ideally. it is intended for the transition period where
// not all services can work with the new X-Runai-SubjectId/TenantId headers and require the original JWT token
func PreserveJwtToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		value := r.Header.Get(HeaderAuthorization)
		if strings.HasPrefix(value, BearerPrefix) {
			ctx = context.WithValue(ctx, ContextJwtToken, value[len(BearerPrefix):])
		} else {
			ctx = context.WithValue(ctx, ContextJwtToken, "")
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
