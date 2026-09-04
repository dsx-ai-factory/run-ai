/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
)

type MetricQuerier interface {
	Query(ctx context.Context, tenantId int32, query string, timeRange *MetricInformationRange) (*api.GetMetrics200Response, error)
}

type MetricQueryBuilder interface {
	Build(ctx context.Context, metricInformation *MetricInformation) (string, error)
}

type ReportsDao interface {
	CreateReport(ctx context.Context, tenantId int32, report models.Report) (models.Report, *api.Error)
	GetReport(ctx context.Context, tenantId int32, reportId string) (models.Report, *api.Error)
	GetReportsByIds(ctx context.Context, tenantId int32, reportIds []string, filterBy ...queryapi.Filter) ([]models.Report, *api.Error)
	DeleteReport(ctx context.Context, tenantId int32, reportId string) *api.Error
	GetReports(ctx context.Context, tenantId int32, filterBy []queryapi.Filter, sortBy string, sortOrder string, offset int32, limit int32, search string) ([]models.Report, *api.Error)
	CountReports(ctx context.Context, tenantId int32, filterBy []queryapi.Filter, search string) (int64, *api.Error)
	UpdatePendingDelete(ctx context.Context, tenantId int32, reportIds []string) (updated []string, apiErr *api.Error)
}

type MetricStoreManager interface {
	Create(ctx context.Context, metricStore api.MetricStore) (*api.MetricStore, error)
	GetByTenantId(ctx context.Context, tenantId int32) (*api.MetricStore, error)
	DeleteByTenantId(ctx context.Context, tenantId int32) error
}

type TenantsClienter interface {
	TenantExist(ctx context.Context, tenantId int32) (bool, error)
}
