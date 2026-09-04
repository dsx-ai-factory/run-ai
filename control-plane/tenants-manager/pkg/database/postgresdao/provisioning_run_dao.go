/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package postgresdao

import (
	"context"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/domain/models"
)

const TableNameProvisioningRuns = "provisioning_runs"

type ProvisioningRunDao struct{}

func NewProvisioningRunDao() *ProvisioningRunDao {
	return &ProvisioningRunDao{}
}

func (r *ProvisioningRunDao) Create(ctx context.Context, tenantId uint) (*models.ProvisioningRun, *dao.DaoError) {
	run := models.ProvisioningRun{
		TenantId:  tenantId,
		StartedAt: time.Now().UTC(),
		Status:    models.ProvisioningRunRunning,
	}
	_, err := DBwithCtx(ctx).Create(&run)
	if err != nil {
		return nil, &dao.DaoError{Reason: dao.General, Err: err}
	}
	return &run, nil
}

func (r *ProvisioningRunDao) Complete(ctx context.Context, runId uint, status models.ProvisioningRunStatus, diagnostics string) *dao.DaoError {
	now := time.Now().UTC()
	_, err := DBwithCtx(ctx).Model(&models.ProvisioningRun{}).
		Table(TableNameProvisioningRuns).
		Where("id = ?", runId).
		Updates(map[string]interface{}{
			"ended_at":    now,
			"status":      status,
			"diagnostics": diagnostics,
		})
	if err != nil {
		return &dao.DaoError{Reason: dao.General, Err: err}
	}
	return nil
}

func (r *ProvisioningRunDao) GetLatestForTenant(ctx context.Context, tenantId uint) (*models.ProvisioningRun, *dao.DaoError) {
	var run models.ProvisioningRun
	err := DBwithCtx(ctx).Table(TableNameProvisioningRuns).
		Where("tenant_id = ?", tenantId).
		Order("ended_at DESC NULLS LAST").
		First(&run)
	if err != nil {
		if isNotFoundError(err) {
			return nil, &dao.DaoError{Reason: dao.NotFound, Err: err}
		}
		return nil, &dao.DaoError{Reason: dao.General, Err: err}
	}
	return &run, nil
}
