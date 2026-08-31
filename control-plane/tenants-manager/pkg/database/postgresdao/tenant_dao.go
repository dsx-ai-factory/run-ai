/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package postgresdao

import (
	"context"
	"errors"
	"strconv"

	"github.com/run-ai/runai/runai-common-packages/runaidb/orm/clause"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"

	"github.com/run-ai/runai/control-plane/tenants-manager/internal/cache"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/domain/models"
)

func NewTenantDao(tenantCacheManager cache.TenantsCache) *Tenant {
	return &Tenant{
		tenantCacheManager: tenantCacheManager,
	}
}

type Tenant struct {
	tenantCacheManager cache.TenantsCache
}

func (t *Tenant) Delete(ctx context.Context, tenantSpec models.Tenant) *dao.DaoError {
	_, err := DBwithCtx(ctx).Delete(&tenantSpec)
	if isNotFoundError(err) {
		return &dao.DaoError{
			Reason: dao.NotFound,
			Err:    err,
		}
	} else if err != nil {
		return &dao.DaoError{
			Reason: dao.General,
			Err:    err,
		}
	}
	if err := t.tenantCacheManager.DeleteTenant(ctx, int32(tenantSpec.ID)); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to delete tenant from cache. tenant id: %d, error: %v", tenantSpec.ID, err)
	}
	return nil
}

func (t *Tenant) HardDeleteById(ctx context.Context, tenantId uint) *dao.DaoError {
	_, err := DBwithCtx(ctx).Unscoped().Where("Id = ?", tenantId).Delete(&models.Tenant{})
	if isNotFoundError(err) {
		return &dao.DaoError{
			Reason: dao.NotFound,
			Err:    err,
		}
	} else if err != nil {
		return &dao.DaoError{
			Reason: dao.General,
			Err:    err,
		}
	}
	if err := t.tenantCacheManager.DeleteTenant(ctx, int32(tenantId)); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to delete tenant from cache. tenant id: %d, error: %v", tenantId, err)
	}
	return nil
}

func (t *Tenant) Create(ctx context.Context, tenantSpec models.Tenant) (*models.Tenant, *dao.DaoError) {
	_, err := DBwithCtx(ctx).Create(&tenantSpec)
	if isDuplicateKeyError(err) {
		return nil, &dao.DaoError{
			Reason: dao.Conflict,
			Err:    err,
		}
	} else if err != nil {
		return nil, &dao.DaoError{
			Reason: dao.General,
			Err:    err,
		}
	}
	return &tenantSpec, nil
}

func (t *Tenant) UpdateStatus(ctx context.Context, tenantId uint, status models.TenantStatus) (*models.Tenant, *dao.DaoError) {
	return t.UpdateByField(ctx, tenantId, "Status", status)
}

func (t *Tenant) UpdateByField(ctx context.Context, tenantId uint, paramName string, paramValue interface{}) (*models.Tenant, *dao.DaoError) {
	var tenant models.Tenant
	_, err := DBwithCtx(ctx).Model(&tenant).
		Table(TableNameTenants).
		Clauses(clause.Returning{}).
		Where("Id = ?", tenantId).
		Update(paramName, paramValue)

	if isDuplicateKeyError(err) {
		return nil, &dao.DaoError{
			Reason: dao.Conflict,
			Err:    err,
		}
	} else if isNotFoundError(err) {
		return nil, &dao.DaoError{
			Reason: dao.NotFound,
			Err:    err,
		}
	} else if err != nil {
		return nil, &dao.DaoError{
			Reason: dao.General,
			Err:    err,
		}
	}
	if err := t.tenantCacheManager.DeleteTenant(ctx, int32(tenantId)); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to delete tenant from cache. tenant id: %d, error: %v", tenantId, err)
	}

	return &tenant, nil
}

func (t *Tenant) GetByName(ctx context.Context, tenantName string) (*models.Tenant, *dao.DaoError) {
	return t.GetBy(ctx, "Name", tenantName)
}

func (t *Tenant) GetByAlias(ctx context.Context, alias string) (*models.Tenant, *dao.DaoError) {
	return t.GetBy(ctx, "Alias", alias)
}

func (t *Tenant) GetByNameOrAlias(ctx context.Context, nameOrAlias string) (*models.Tenant, *dao.DaoError) {
	tenant := &models.Tenant{}
	err := DBwithCtx(ctx).Table(TableNameTenants).
		Where("name = ? OR alias = ?", nameOrAlias, nameOrAlias).
		First(tenant)
	if err != nil {
		reason := dao.General
		if isNotFoundError(err) {
			reason = dao.NotFound
			err = errors.Join(err, dao.ErrNotFound)
		}
		return nil, &dao.DaoError{
			Reason: reason,
			Err:    err,
		}
	}
	return tenant, nil
}

func (t *Tenant) GetById(ctx context.Context, tenantId uint) (*models.Tenant, *dao.DaoError) {
	tenant, err := t.tenantCacheManager.GetTenant(ctx, int32(tenantId), func() (*models.Tenant, error) {
		tenant, err := t.GetBy(ctx, "Id", strconv.FormatUint(uint64(tenantId), 10))
		if err == (*dao.DaoError)(nil) {
			return tenant, nil
		}
		return nil, err
	})
	if err == nil {
		return tenant, nil
	}
	if daoError, ok := err.(*dao.DaoError); ok {
		return nil, daoError
	}
	return nil, &dao.DaoError{
		Reason: dao.General,
		Err:    err,
	}
}

func (t *Tenant) GetByIdIncludeDeleted(ctx context.Context, tenantId uint, includeDeleted bool) (*models.Tenant, *dao.DaoError) {
	tenant, err := t.tenantCacheManager.GetTenant(ctx, int32(tenantId), func() (*models.Tenant, error) {
		tenant, err := t.GetByIncludeDeleted(ctx, "Id", strconv.FormatUint(uint64(tenantId), 10), includeDeleted)
		if err == (*dao.DaoError)(nil) {
			return tenant, nil
		}
		return nil, err
	})
	if err == nil {
		return tenant, nil
	}
	if daoError, ok := err.(*dao.DaoError); ok {
		return nil, daoError
	}
	return nil, &dao.DaoError{
		Reason: dao.General,
		Err:    err,
	}
}

func (t *Tenant) GetAll(ctx context.Context, permittedScopes *authorization_api.PermittedScopes, filters *dao.TenantFilters) (*[]models.Tenant, *dao.DaoError) {
	tenants := &[]models.Tenant{}
	query := DBwithCtx(ctx).Table(TableNameTenants)
	if filters != nil && filters.Deleted {
		query = query.Unscoped()
	}
	if permittedScopes != nil {
		if permittedScopes.System != nil && *permittedScopes.System {
			// system sees everything — no filter
		} else if len(permittedScopes.Tenants) > 0 {
			query = query.Where("Id IN ?", permittedScopes.Tenants)
		} else {
			query = query.Where("false")
		}
	}
	if filters != nil {
		if filters.Name != "" && filters.Alias != "" {
			query = query.Where("name = ? OR alias = ?", filters.Name, filters.Alias)
		} else if filters.Name != "" {
			query = query.Where("name = ?", filters.Name)
		} else if filters.Alias != "" {
			query = query.Where("alias = ?", filters.Alias)
		}
	}
	if err := query.Find(tenants); err != nil {
		return nil, &dao.DaoError{
			Reason: dao.General,
			Err:    err,
		}
	}
	return tenants, nil
}

func (t *Tenant) GetByIncludeDeleted(ctx context.Context, paramName string, paramValue string, includeDeleted bool) (*models.Tenant, *dao.DaoError) {
	tenant := &models.Tenant{}
	tx := DBwithCtx(ctx).Table(TableNameTenants)
	whereClause := paramName + " = ?"
	err := t.getResult(tx, whereClause, paramValue, tenant, includeDeleted)
	if err != nil {
		reason := dao.General
		if isNotFoundError(err) {
			reason = dao.NotFound
			err = errors.Join(err, dao.ErrNotFound)
		}
		return nil, &dao.DaoError{
			Reason: reason,
			Err:    err,
		}
	}
	return tenant, nil
}

func (t *Tenant) GetBy(ctx context.Context, paramName string, paramValue string) (*models.Tenant, *dao.DaoError) {
	return t.GetByIncludeDeleted(ctx, paramName, paramValue, false)
}

func (t *Tenant) getResult(tx orm.DB, query string, paramValue string, tenant *models.Tenant, includeDeleted bool) error {
	if includeDeleted {
		return tx.Unscoped().Where(query, paramValue).First(tenant)
	}
	return tx.Where(query, paramValue).First(tenant)
}

func (t *Tenant) GetTenantsCount(ctx context.Context, contractType string) (*int64, *dao.DaoError) {
	var rowCount int64
	query := DBwithCtx(ctx).Table(TableNameTenants)
	if contractType != "" {
		query = query.Where("contract_type = ?", contractType)
	}
	if err := query.Count(&rowCount); err != nil {
		return nil, &dao.DaoError{
			Reason: dao.General,
			Err:    err,
		}
	}
	return &rowCount, nil
}

func (t *Tenant) UpdateSmg(ctx context.Context, tenantId uint, smg bool) (*models.Tenant, *dao.DaoError) {
	return t.UpdateByField(ctx, tenantId, "Smg", smg)
}
