/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

// ReadJSONResponse decodes a generated API client's buffered response body.
func ReadJSONResponse(resp *http.Response, out any) error {
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("missing response body")
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("decode response body: multiple JSON values")
		}
		return fmt.Errorf("decode response body: %w", err)
	}
	return nil
}

// ReadAPIResponse maps generated-client errors and decodes a JSON object.
// For operations with no response content, it closes the body and returns {}.
func ReadAPIResponse(resp *http.Response, err error, noContent bool) (map[string]any, error) {
	if err != nil {
		return nil, runaimcp.MapHTTPError(resp, err)
	}
	if noContent {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		return map[string]any{}, nil
	}
	out := map[string]any{}
	if err := ReadJSONResponse(resp, &out); err != nil {
		return nil, fmt.Errorf("read API response: %w: %w", err, runaimcp.ErrUpstream)
	}
	if out == nil {
		return nil, fmt.Errorf("API returned null instead of a JSON object: %w", runaimcp.ErrUpstream)
	}
	return out, nil
}
