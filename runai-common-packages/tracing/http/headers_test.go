/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package http

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetHeaders_SetsAllHeaders(t *testing.T) {
	req := httptest.NewRequest("POST", "http://some.url.com", nil)
	ctx := context.Background()
	for _, header := range tracingHeaders {
		ctx = context.WithValue(ctx, contextKey(header), fmt.Sprintf("%s-value", header))
	}
	SetHeaders(ctx, req)
	for _, header := range tracingHeaders {
		assert.Equal(t, req.Header.Get(header), fmt.Sprintf("%s-value", header))
	}
}

func TestSaveTracingHeadersInCtx_SavesHeadersToCtx(t *testing.T) {
	req := httptest.NewRequest("POST", "http://some.url.com", nil)
	ctx := context.Background()
	for _, header := range tracingHeaders {
		req.Header.Set(header, fmt.Sprintf("%s-value", header))
	}

	ctx = saveTracingHeadersInCtx(req, ctx)

	for _, header := range tracingHeaders {
		assert.Equal(t, ctx.Value(contextKey(header)).(string), fmt.Sprintf("%s-value", header))
	}
}

func TestSaveTracingHeadersInCtx_DontSavesEmptyHeadersToCtx(t *testing.T) {
	req := httptest.NewRequest("POST", "http://some.url.com", nil)
	ctx := context.Background()
	for _, header := range tracingHeaders {
		req.Header.Set(header, "")
	}

	ctx = saveTracingHeadersInCtx(req, ctx)

	for _, header := range tracingHeaders {
		assert.Equal(t, ctx.Value(contextKey(header)), nil)
	}
}
