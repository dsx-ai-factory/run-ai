/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package postgresdao

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/domain/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type Logo struct{}

func (plc *Logo) GetByTenantId(ctx context.Context, tenantId uint) (*[]byte, *dao.DaoError) {
	logoModel := &models.Logo{
		TenantId: tenantId,
	}
	err := DBwithCtx(ctx).First(logoModel, "tenant_id = ?", logoModel.TenantId)
	if err == nil {
		return logoModel.LogoPng, nil
	}

	if isNotFoundError(err) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Logo not found for tenant %d", tenantId)
		return nil, &dao.DaoError{
			Reason: dao.NoContent,
			Err:    err,
		}

	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get logo for tenant %d: %v", tenantId, err)
	return nil, &dao.DaoError{
		Reason: dao.General,
		Err:    err,
	}
}

func (plc *Logo) GetAll(ctx context.Context, offset int, limit int) (*[]models.Logo, *dao.DaoError) {
	var logos []models.Logo

	err := DBwithCtx(ctx).Model(&models.Logo{}).Order("tenant_id ASC").Offset(offset).Limit(limit).Find(&logos)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get logos (offset: %d, limit: %d): %v", offset, limit, err)
		return nil, &dao.DaoError{
			Reason: dao.General,
			Err:    err,
		}
	}

	return &logos, nil
}

func (plc *Logo) SetByTenantId(ctx context.Context, tenantId uint, logo *[]byte) *dao.DaoError {
	logoModel := &models.Logo{
		TenantId: tenantId,
		LogoPng:  logo,
	}
	_, err := DBwithCtx(ctx).FirstOrCreate(logoModel, "tenant_id = ?", logoModel.TenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to save logo for tenant %d: %v", tenantId, err)
		return &dao.DaoError{
			Reason: dao.General,
			Err:    err,
		}
	}
	logoModel.LogoPng = logo
	_, saveErr := DBwithCtx(ctx).Save(logoModel)
	if saveErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to save logo for tenant %d: %v", tenantId, saveErr)
		return &dao.DaoError{
			Reason: dao.General,
			Err:    saveErr,
		}
	}
	return nil
}
