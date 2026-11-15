/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
)

const chartInspectionJSONLimit = 64 << 10

// Bound the body before the generated JSON controller's unbounded io.ReadAll.
func chartInspectionBodyLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/ai-applications/charts/inspect" {
			next.ServeHTTP(w, r)
			return
		}
		reject := func(status int, message string) {
			_ = api.EncodeJSONResponse(api.NewError(status, message, nil), &status, w)
		}
		if r.ContentLength > chartInspectionJSONLimit {
			reject(http.StatusRequestEntityTooLarge, "chart inspection JSON exceeds the 64 KiB limit")
			return
		}
		if r.Body == nil {
			next.ServeHTTP(w, r)
			return
		}
		defer r.Body.Close()
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, chartInspectionJSONLimit))
		if err != nil {
			var limitError *http.MaxBytesError
			if errors.As(err, &limitError) {
				reject(http.StatusRequestEntityTooLarge, "chart inspection JSON exceeds the 64 KiB limit")
			} else {
				reject(http.StatusBadRequest, "failed to read chart inspection JSON")
			}
			return
		}
		if !json.Valid(body) {
			reject(http.StatusBadRequest, "chart inspection requires a single valid JSON document")
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		next.ServeHTTP(w, r)
	})
}
