/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/domain/models"
)

type TenantFilters struct {
	Name    string
	Alias   string
	Deleted bool
}

type Tenant interface {
	Create(ctx context.Context, tenantSpec models.Tenant) (*models.Tenant, *DaoError)
	Delete(ctx context.Context, tenantSpec models.Tenant) *DaoError
	HardDeleteById(ctx context.Context, tenantId uint) *DaoError
	UpdateStatus(ctx context.Context, tenantId uint, status models.TenantStatus) (*models.Tenant, *DaoError)
	UpdateSmg(ctx context.Context, tenantId uint, smg bool) (*models.Tenant, *DaoError)
	UpdateByField(ctx context.Context, tenantId uint, fieldName string, fieldValue interface{}) (*models.Tenant, *DaoError)
	GetByName(ctx context.Context, tenantName string) (*models.Tenant, *DaoError)
	GetByAlias(ctx context.Context, alias string) (*models.Tenant, *DaoError)
	GetByNameOrAlias(ctx context.Context, nameOrAlias string) (*models.Tenant, *DaoError)
	GetById(ctx context.Context, tenantId uint) (*models.Tenant, *DaoError)
	GetByIdIncludeDeleted(ctx context.Context, tenantId uint, includeDeleted bool) (*models.Tenant, *DaoError)
	GetAll(ctx context.Context, permittedScopes *authorization_api.PermittedScopes, filters *TenantFilters) (*[]models.Tenant, *DaoError)
	GetBy(ctx context.Context, paramName string, paramValue string) (*models.Tenant, *DaoError)
	GetTenantsCount(ctx context.Context, contractType string) (*int64, *DaoError)
}

type Logo interface {
	GetByTenantId(ctx context.Context, tenantId uint) (*[]byte, *DaoError)
	SetByTenantId(ctx context.Context, tenantId uint, logo *[]byte) *DaoError
	GetAll(ctx context.Context, offset int, limit int) (*[]models.Logo, *DaoError)
}

type MetricStore interface {
	GetByTenantId(ctx context.Context, tenantId uint) (*models.MetricStore, *DaoError)
}

type ProvisioningRun interface {
	Create(ctx context.Context, tenantId uint) (*models.ProvisioningRun, *DaoError)
	Complete(ctx context.Context, runId uint, status models.ProvisioningRunStatus, diagnostics string) *DaoError
	GetLatestForTenant(ctx context.Context, tenantId uint) (*models.ProvisioningRun, *DaoError)
}
