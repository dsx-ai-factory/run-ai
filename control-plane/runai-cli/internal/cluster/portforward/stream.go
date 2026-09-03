/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package portforward

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/status"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/base"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/websocket"
)

type Stream struct {
	*base.StreamBase
	clusterMetadata *status.Metadata
}

func NewStream(dialer websocket.Dialer, baseUrl string, maxRetryDuration, timeBetweenTries time.Duration, clusterMetadata *status.Metadata) *Stream {
	return &Stream{
		StreamBase:      base.NewStreamBase(dialer, baseUrl, maxRetryDuration, timeBetweenTries),
		clusterMetadata: clusterMetadata,
	}
}

func (s *Stream) Stream(ctx context.Context, streamUrl *url.URL) (io.ReadWriteCloser, error) {
	conn, err := s.stream(ctx, streamUrl)
	if err != nil {
		return nil, err
	}

	if !base.IsMessageHandlerCapable(conn, s.clusterMetadata) {
		return conn, nil
	}

	base.SetMessageKeepAlive(conn)
	return conn, nil
}

func (s *Stream) stream(ctx context.Context, url *url.URL) (io.ReadWriteCloser, error) {
	return s.ConnectWithRetries(ctx, url, nil)
}
