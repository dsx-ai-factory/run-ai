/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

func SaveOrgUnitDataDefaultNodepoolLists(ctx context.Context, orgUnitDataEntities []models.OrgUnitData) error {
	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.Errorf("Failed to begin a session with the db: %v", tx.Error)
		panic(tx.Error)
	}

	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)
	tx.SavePoint("sp")

	for _, orgUnitData := range orgUnitDataEntities {
		result := tx.
			Table(models.OrgUnitDataTable).
			Where("id = ?", orgUnitData.ID).
			UpdateColumns(map[string]interface{}{
				"default_node_pool_list_effective":        orgUnitData.DefaultNodePoolListEffective,
				"default_node_pool_list_effective_origin": orgUnitData.DefaultNodePoolListEffectiveOrigin,
				"default_node_pool_list":                  orgUnitData.DefaultNodePoolList,
			})
		if result.Error != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update org unit %s default nodepool list: error: %s", orgUnitData.ID, result.Error.Error())
			tx.RollbackTo("sp")
		} else {
			tx.SavePoint("sp")
		}
	}

	result := tx.Commit()
	if result.Error != nil {
		logging.Errorf("Failed to commit the default nodepool deletion: error: %v", result.Error.Error())
		return result.Error
	}
	committed = true

	return nil
}

func updateUpdatedAt(ctx context.Context, orgUnitDataIds []string, tx *gorm.DB, subjectId string) error {
	updatedAt := time.Now()
	result := tx.Table(models.OrgUnitDataTable).
		Where("id in ?", orgUnitDataIds).
		Updates(map[string]interface{}{
			"updated_at": updatedAt,
			"updated_by": subjectId,
		})
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update org_unit_data updated at, error: %v", result.Error)
		return result.Error
	}
	return nil
}

func GetOrgUnitsNodepoolSummary(ctx context.Context, tenantId, nodepoolId string, nodepoolSummaryType api.OrgUnitsNodepoolSummaryType) ([]api.OrgUnitsNodepoolSummaryInner, error) {
	var result []api.OrgUnitsNodepoolSummaryInner

	projectsSummary, err := getOrgUnitsNodepoolSummaryInternal(ctx, tenantId, nodepoolId, nodepoolSummaryType, true)
	if err != nil {
		return nil, err
	}
	result = append(result, projectsSummary...)

	departmentsSummary, err := getOrgUnitsNodepoolSummaryInternal(ctx, tenantId, nodepoolId, nodepoolSummaryType, false)
	if err != nil {
		return nil, err
	}
	result = append(result, departmentsSummary...)

	return result, nil
}
