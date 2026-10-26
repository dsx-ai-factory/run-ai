/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadJSONResponse(t *testing.T) {
	resp := &http.Response{Body: io.NopCloser(strings.NewReader(`{"value":"ok"}`))}
	var result map[string]any
	require.NoError(t, ReadJSONResponse(resp, &result))
	assert.Equal(t, "ok", result["value"])
}

func TestReadJSONResponseString(t *testing.T) {
	resp := &http.Response{Body: io.NopCloser(strings.NewReader(`"ok"`))}
	var result string
	require.NoError(t, ReadJSONResponse(resp, &result))
	assert.Equal(t, "ok", result)
}

func TestReadJSONResponseRejectsMissingBody(t *testing.T) {
	assert.ErrorContains(t, ReadJSONResponse(nil, &map[string]any{}), "missing response body")
}

func TestReadJSONResponseRejectsTrailingData(t *testing.T) {
	for _, body := range []string{`{} {}`, `{} invalid`} {
		resp := &http.Response{Body: io.NopCloser(strings.NewReader(body))}
		assert.ErrorContains(t, ReadJSONResponse(resp, &map[string]any{}), "decode response body")
	}
}
