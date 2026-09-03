/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/status"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/apis"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/base"
	clusterUrl "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/url"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/websocket"
)

const (
	logsEndpoint = "logs"
)

type WorkloadLogsStreamer struct {
	*base.StreamBase
	clusterMetadata *status.Metadata
}

func NewWorkloadLogsStreamer(dialer websocket.Dialer, baseUrl string, maxRetryDuration, timeBetweenTries time.Duration, clusterMetadata *status.Metadata) *WorkloadLogsStreamer {
	return &WorkloadLogsStreamer{
		StreamBase:      base.NewStreamBase(dialer, baseUrl, maxRetryDuration, timeBetweenTries),
		clusterMetadata: clusterMetadata,
	}
}

func (wls *WorkloadLogsStreamer) Stream(ctx context.Context, workload *apis.Workload, options *Options) (io.ReadWriteCloser, error) {
	logsUrl, err := wls.buildUrl(workload, options)
	if err != nil {
		return nil, err
	}

	remote, err := wls.stream(ctx, logsUrl)
	if err != nil {
		return nil, err
	}

	if !base.IsMessageHandlerCapable(remote, wls.clusterMetadata) {
		return remote, nil
	}

	base.SetMessageKeepAlive(remote)
	return remote, nil
}

func (wls *WorkloadLogsStreamer) stream(ctx context.Context, logsUrl *url.URL) (io.ReadWriteCloser, error) {
	return wls.ConnectWithRetries(ctx, logsUrl, nil)
}

func (wls *WorkloadLogsStreamer) buildUrl(workload *apis.Workload, options *Options) (*url.URL, error) {
	var query *url.Values
	if options != nil {
		var err error
		query, err = options.ToQueryParams()
		if err != nil {
			return nil, err
		}
	}
	handler := wls.GetHandler()
	return clusterUrl.BuildClusterUrl(handler.BaseUrl, workload, logsEndpoint, query)
}
