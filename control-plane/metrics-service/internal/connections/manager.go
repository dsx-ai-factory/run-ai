/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package connections

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/errors"
)

type Manager struct {
	getter MetricStoreGetter
}

func NewManager(getter MetricStoreGetter) *Manager {
	return &Manager{getter: getter}
}

func (c *Manager) GetConnection(ctx context.Context, tenantId int32) (*Connection, error) {
	metricStore, err := c.getter.GetByTenantId(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	if metricStore == nil {
		return nil, fmt.Errorf("%s %d", errors.MetricStoreNotFound, tenantId)
	}
	return newConnection(metricStore.MetricStoreDetails.Read)
}
