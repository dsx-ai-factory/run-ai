/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"
	"net/http"
)

type Provider interface {
	Token(ctx context.Context) (string, error)
	http.RoundTripper
	HttpClient() *http.Client
}
