/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package forward

import (
	"context"
	"net/url"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/apis"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/portforward/options"
)

var logger *zap.Logger

type List struct {
	portsList []Forward
	ctx       context.Context
	workload  *apis.Workload
	token     string
	options   *options.Options
	url       *url.URL
}

func NewForwardList(ctx context.Context, workload *apis.Workload, options *options.Options, token string, url *url.URL) *List {
	logger = config.GetInstance().GetLogger()
	return &List{
		portsList: make([]Forward, 0),
		ctx:       ctx,
		workload:  workload,
		options:   options,
		token:     token,
		url:       url,
	}
}

func (fl *List) Close() error {
	for _, p := range fl.portsList {
		defer p.Close()
	}
	return nil
}

func (fl *List) Add(p *Forward) {
	logger.Debug("adding port forward", zap.Int32("local", p.local), zap.Int32("remote", p.remote))
	fl.portsList = append(fl.portsList, *p)
}

func (fl *List) BindLocalPort() error {
	var err error
	for i := range fl.portsList {
		logger.Debug("binding local port forward",
			zap.Int32("local", fl.portsList[i].local),
			zap.Int32("remote", fl.portsList[i].remote))
		err = fl.portsList[i].BindLocalPort(fl.ctx)
	}
	return err
}

func (fl *List) Run() error {
	eg, ctx := errgroup.WithContext(fl.ctx)
	for _, port := range fl.portsList {
		eg.Go(func() error {
			err := port.Run(ctx) //nolint:govet // tested and is safe
			_ = fl.Close()
			return err
		})
	}
	return eg.Wait()
}
