/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package attach

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/apis"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/base"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/pseudo"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/status"
	clusterUrl "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/url"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/websocket"
)

const (
	attachEndpoint = "attach"
)

type WorkloadAttach struct {
	*base.StreamAttachExecute
	clusterMetadata *status.Metadata
}

func NewWorkloadAttach(dialer websocket.Dialer, baseUrl string, maxRetryDuration, timeBetweenTries time.Duration, clusterMetadata *status.Metadata) *WorkloadAttach {
	return &WorkloadAttach{
		StreamAttachExecute: base.NewStreamAttachExecute(base.NewStreamBase(dialer, baseUrl, maxRetryDuration, timeBetweenTries)),
		clusterMetadata:     clusterMetadata,
	}
}

func (wa *WorkloadAttach) Stream(ctx context.Context, workload *apis.Workload, options *pseudo.Options) (io.ReadWriteCloser, error) {
	execUrl, err := wa.buildUrl(workload, options)
	if err != nil {
		return nil, err
	}
	remote, err := wa.stream(ctx, execUrl, nil)
	if err != nil {
		return nil, err
	}

	return wa.SetupMessageHandlerAttachExecute(remote, wa.clusterMetadata, options.TTY, options.Height, options.Width)
}

func (wa *WorkloadAttach) stream(ctx context.Context, execUrl *url.URL, header *http.Header) (io.ReadWriteCloser, error) {
	return wa.ConnectWithRetries(ctx, execUrl, header)
}

func (wa *WorkloadAttach) buildUrl(workload *apis.Workload, options *pseudo.Options) (*url.URL, error) {
	var query *url.Values
	if options != nil {
		var err error
		query, err = options.ToQueryParams()
		if err != nil {
			return nil, err
		}
	}
	handler := wa.GetHandler()
	return clusterUrl.BuildClusterUrl(handler.BaseUrl, workload, attachEndpoint, query)
}
