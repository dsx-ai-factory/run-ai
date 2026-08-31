/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package exec

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/apis"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/base"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/pseudo"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/status"
	clusterUrl "github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/url"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/websocket"
)

const (
	execEndpoint  = "exec"
	HeaderCommand = "Command"
)

type WorkloadExec struct {
	*base.StreamAttachExecute
	clusterMetadata *status.Metadata
}

func NewWorkloadExec(dialer websocket.Dialer, baseUrl string, maxRetryDuration, timeBetweenTries time.Duration, clusterMetadata *status.Metadata) *WorkloadExec {
	return &WorkloadExec{
		StreamAttachExecute: base.NewStreamAttachExecute(base.NewStreamBase(dialer, baseUrl, maxRetryDuration, timeBetweenTries)),
		clusterMetadata:     clusterMetadata,
	}
}

func (we *WorkloadExec) Stream(ctx context.Context, workload *apis.Workload, options *pseudo.Options) (io.ReadWriteCloser, error) {
	execUrl, err := we.buildUrl(workload, options)
	if err != nil {
		return nil, err
	}
	execHeader := we.buildHeader(options)
	remote, err := we.stream(ctx, execUrl, execHeader)
	if err != nil {
		return nil, err
	}

	return we.SetupMessageHandlerAttachExecute(remote, we.clusterMetadata, options.TTY, options.Height, options.Width)
}

func (we *WorkloadExec) stream(ctx context.Context, execUrl *url.URL, header *http.Header) (io.ReadWriteCloser, error) {
	return we.ConnectWithRetries(ctx, execUrl, header)
}

func (we *WorkloadExec) buildUrl(workload *apis.Workload, options *pseudo.Options) (*url.URL, error) {
	var query *url.Values
	if options != nil {
		var err error
		query, err = options.ToQueryParams()
		if err != nil {
			return nil, err
		}
	}
	handler := we.GetHandler()
	return clusterUrl.BuildClusterUrl(handler.BaseUrl, workload, execEndpoint, query)
}

func (we *WorkloadExec) buildHeader(options *pseudo.Options) *http.Header {
	header := &http.Header{}
	header.Set(HeaderCommand, options.Command)
	return header
}
