/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mock

import (
	"context"
	"net/http"
)

type FakeProvider struct {
	TokenValue string
	TokenErr   error
}

func (f *FakeProvider) Token(_ context.Context) (string, error) {
	return f.TokenValue, f.TokenErr
}

func (f *FakeProvider) RoundTrip(req *http.Request) (*http.Response, error) {
	return http.DefaultTransport.RoundTrip(req)
}

func (f *FakeProvider) HttpClient() *http.Client {
	return &http.Client{Transport: f}
}
