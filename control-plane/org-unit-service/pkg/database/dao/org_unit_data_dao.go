/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

func SaveOrgUnitDataDefaultNodepoolLists(ctx context.Context, orgUnitDataEntities []models.OrgUnitData) error {
	tx := runaidb.DB().WithContext(ctx).Begin()

	committed := false
	defer orm.FinalizeTransaction(tx, &committed)
	tx.SavePoint("sp")

	for _, orgUnitData := range orgUnitDataEntities {
		columns := map[string]interface{}{
			"default_node_pool_list_effective":        orgUnitData.DefaultNodePoolListEffective,
			"default_node_pool_list_effective_origin": orgUnitData.DefaultNodePoolListEffectiveOrigin,
			"default_node_pool_list":                  orgUnitData.DefaultNodePoolList,
		}
		if _, err := tx.Table(models.OrgUnitDataTable).Where("id = ?", orgUnitData.ID).UpdateColumns(columns); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update org unit %s default nodepool list: error: %s", orgUnitData.ID, err.Error())
			tx.RollbackTo("sp")
		} else {
			tx.SavePoint("sp")
		}
	}

	if err := tx.Commit(); err != nil {
		logging.Errorf("Failed to commit the default nodepool deletion: error: %v", err.Error())
		return err
	}
	committed = true

	return nil
}

func updateUpdatedAt(ctx context.Context, orgUnitDataIds []string, tx orm.DB, subjectId string) error {
	updatedAt := time.Now()
	if _, err := tx.Table(models.OrgUnitDataTable).Where("id in ?", orgUnitDataIds).Updates(map[string]interface{}{"updated_at": updatedAt, "updated_by": subjectId}); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update org_unit_data updated at, error: %v", err)
		return err
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
