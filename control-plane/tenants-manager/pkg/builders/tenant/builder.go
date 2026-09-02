/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tenant

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/builders"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/builders/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/builders/composite"
	metricsbuilder "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/builders/metrics"
	metricStore "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/business/metric-store"
)

const (
	builderIdentifier = "tenant_builder"
)

type Builder struct {
	composite.CompositeBuilder
}

func NewBuilder(metricsStoreService metricStore.MetricStoreService) builders.Interface {
	var tenantBuilders []builders.Interface
	tenantBuilders = append(tenantBuilders,
		metricsbuilder.NewBuilder(metricsStoreService),
		auth.NewBuilder())
	return &Builder{
		composite.NewCompositeBuilder(tenantBuilders, composite.Parallel),
	}
}

func (h Builder) Name() string {
	return builderIdentifier
}
