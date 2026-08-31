/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metricstores

import (
	"context"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/database/models"
	api "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client/generated"
)

type MetricStoresDao interface {
	Create(ctx context.Context, metricStore models.MetricStore) (*models.MetricStore, error)
	GetByTenantId(ctx context.Context, tenantId int32) (*models.MetricStore, error)
	Delete(ctx context.Context, tenantId int32) error
}

type TenantMetricStoreGetter interface {
	GetExtendedMetricStore(ctx context.Context, tenantId int32) (*api.ExtendedMetricStore, error)
}

type DatasourceProvider interface {
	Create(ctx context.Context, datasourceId string) (*Datasource, error)
	Delete(ctx context.Context, datasourceId string) error
}
