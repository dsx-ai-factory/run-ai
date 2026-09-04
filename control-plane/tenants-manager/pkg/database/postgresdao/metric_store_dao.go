/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package postgresdao

import (
	"context"
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/domain/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type MetricStore struct{}

func (*MetricStore) GetByTenantId(ctx context.Context, tenantId uint) (*models.MetricStore, *dao.DaoError) {
	msModel := &models.MetricStore{
		TenantId: tenantId,
	}
	err := DBwithCtx(ctx).First(msModel, "tenant_id = ?", msModel.TenantId)
	if err == nil {
		return msModel, nil
	}

	if isNotFoundError(err) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("MetricStore not found for tenant %d", tenantId)
		return nil, &dao.DaoError{
			Reason: dao.NotFound,
			Err:    errors.Join(err, dao.ErrNotFound),
		}
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get MetricStore for tenant %d:%v", tenantId, err)
	return nil, &dao.DaoError{
		Reason: dao.General,
		Err:    err,
	}
}
