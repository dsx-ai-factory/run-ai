/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package websocket

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/network"
)

type Dialer interface {
	Connect(ctx context.Context, url *url.URL, header *http.Header) (network.Conn, *http.Response, error)
}

type Handler struct {
	Dialer           Dialer
	BaseUrl          string
	MaxRetryDuration time.Duration
	TimeBetweenTries time.Duration
}
