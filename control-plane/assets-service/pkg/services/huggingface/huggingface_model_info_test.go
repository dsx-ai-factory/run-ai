/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package huggingface

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type trackingBody struct {
	io.ReadCloser
	closed bool
}

func (b *trackingBody) Close() error {
	b.closed = true
	return b.ReadCloser.Close()
}

type fileRoundTripper func(*http.Request) (*http.Response, error)

func (f fileRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestModelFileApiCallClosesBodyAndSetsTimeout(t *testing.T) {
	body := &trackingBody{ReadCloser: io.NopCloser(strings.NewReader("ok"))}
	var timeout time.Duration
	original := newHuggingFaceHTTPClient
	t.Cleanup(func() { newHuggingFaceHTTPClient = original })
	newHuggingFaceHTTPClient = func() *http.Client {
		client := original()
		timeout = client.Timeout
		client.Transport = fileRoundTripper(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusForbidden,
				Body:       body,
				Header:     make(http.Header),
				Request:    r,
			}, nil
		})
		return client
	}

	status, err := (&ExternalApi{}).modelFileApiCall(context.Background(), "https://example.test/<HF_MODEL>", "org/model", nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, status)
	require.True(t, body.closed)
	require.Equal(t, huggingFaceHTTPTimeout, timeout)
}
