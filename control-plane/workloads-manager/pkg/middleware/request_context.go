/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

type contextKey string

const (
	RequestPathKey contextKey = "request_path"
	RequestBodyKey contextKey = "request_body"
)

// RequestContextMiddleware captures the request path and body and stores them in the context
func RequestContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Store request path
		ctx = context.WithValue(ctx, RequestPathKey, r.URL.Path)

		// Read and store request body (if present)
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil && len(bodyBytes) > 0 {
				ctx = context.WithValue(ctx, RequestBodyKey, bodyBytes)
				// Restore the body for downstream handlers
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}

		// Continue with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestPath retrieves the request path from context
func GetRequestPath(ctx context.Context) string {
	if path, ok := ctx.Value(RequestPathKey).(string); ok {
		return path
	}
	return ""
}

// GetRequestBody retrieves the request body from context
func GetRequestBody(ctx context.Context) []byte {
	if body, ok := ctx.Value(RequestBodyKey).([]byte); ok {
		return body
	}
	return nil
}
