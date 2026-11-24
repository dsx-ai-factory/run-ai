/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/utils"
)

const nodepoolsPathPrefix = "/api/v1/node-pools/"

func addFullHostToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		value := r.Header.Get("X-Forwarded-Host")
		ctx = context.WithValue(ctx, utils.FullHostKey, value)

		newR := r.WithContext(ctx)
		next.ServeHTTP(w, newR)
	})
}

// markNetworkTopologyIdSent lets a node pool PATCH keep the topology when networkTopologyId is omitted and clear it
// only on an explicit null. The body is put back for the generated controller, which decodes it again.
func markNetworkTopologyIdSent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || !strings.HasPrefix(r.URL.Path, nodepoolsPathPrefix) || r.Body == nil {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewReader(body))
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		var fields map[string]json.RawMessage
		if json.Unmarshal(body, &fields) == nil {
			_, sent := fields["networkTopologyId"]
			r = r.WithContext(context.WithValue(r.Context(), utils.NetworkTopologyIdSentKey, sent))
		}
		next.ServeHTTP(w, r)
	})
}
