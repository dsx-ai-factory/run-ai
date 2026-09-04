/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs

import (
	"context"
	"io"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"go.uber.org/zap"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/apis"
)

type LogsStreamer interface {
	Stream(ctx context.Context, workload *apis.Workload, options *Options) (io.ReadWriteCloser, error)
}

type Proxy interface {
	Start(ctx context.Context, remote io.ReadWriteCloser, local io.ReadWriteCloser) error
}

type Printer struct {
	workloadStreamer LogsStreamer
	proxy            Proxy
	output           io.ReadWriteCloser
}

func NewPrinter(workloadStreamer LogsStreamer, proxy Proxy, output io.ReadWriteCloser) *Printer {
	return &Printer{
		workloadStreamer: workloadStreamer,
		proxy:            proxy,
		output:           output,
	}
}

func (p *Printer) Print(ctx context.Context, workload *apis.Workload, options *Options) error {
	stream, err := p.workloadStreamer.Stream(ctx, workload, options)
	if err != nil {
		return err
	}
	defer stream.Close()
	defer p.output.Close()
	err = p.proxy.Start(ctx, stream, p.output)
	if err != nil {
		if strings.Contains(err.Error(), "StatusNormalClosure") {
			logger := config.GetInstance().GetLogger()
			logger.Debug("normal connection closer", zap.Error(err))
			return nil
		}
		return err
	}
	return nil
}
