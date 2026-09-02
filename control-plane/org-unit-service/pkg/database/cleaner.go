/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"fmt"

	org_unit_config "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/config"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

const (
	cronJobName = "org_unit_cleaner"

	deleteResourcesQueryFormat = "DELETE FROM resources " +
		"WHERE deleted_at IS NOT NULL AND deleted_at < NOW() - INTERVAL '%d' DAY"

	deleteProjectsQueryFormat = "DELETE FROM project " +
		"WHERE deleted_at IS NOT NULL AND deleted_at < NOW() - INTERVAL '%d' DAY"

	deleteDepartmentsQueryFormat = "DELETE FROM department " +
		"WHERE deleted_at IS NOT NULL AND deleted_at < NOW() - INTERVAL '%d' DAY"

	// overtime_data and the node-type join tables have no independent lifecycle.
	// They are parent-driven: cleaned whenever their org_unit_data parent is
	// soft-deleted past the retention window. Must run before deleteOrgUnitData
	// since the FK has no ON DELETE CASCADE.
	deleteOvertimeDataQueryFormat = "DELETE FROM overtime_data " +
		"WHERE org_unit_data_id IN (" +
		"  SELECT id FROM org_unit_data " +
		"  WHERE deleted_at IS NOT NULL AND deleted_at < NOW() - INTERVAL '%d' DAY" +
		")"

	deleteOrgUnitTrainNodeTypeQueryFormat = "DELETE FROM org_unit_train_node_type " +
		"WHERE org_unit_data_id IN (" +
		"  SELECT id FROM org_unit_data " +
		"  WHERE deleted_at IS NOT NULL AND deleted_at < NOW() - INTERVAL '%d' DAY" +
		")"

	deleteOrgUnitWorkspaceNodeTypeQueryFormat = "DELETE FROM org_unit_workspace_node_type " +
		"WHERE org_unit_data_id IN (" +
		"  SELECT id FROM org_unit_data " +
		"  WHERE deleted_at IS NOT NULL AND deleted_at < NOW() - INTERVAL '%d' DAY" +
		")"

	// Remove org_unit_data rows whose six referencing tables have all been cleaned.
	deleteOrgUnitDataQueryFormat = "DELETE FROM org_unit_data " +
		"WHERE deleted_at IS NOT NULL " +
		"  AND deleted_at < NOW() - INTERVAL '%d' DAY " +
		"  AND NOT EXISTS (SELECT 1 FROM project                   WHERE project.org_unit_data_id                   = org_unit_data.id) " +
		"  AND NOT EXISTS (SELECT 1 FROM department                WHERE department.org_unit_data_id                = org_unit_data.id) " +
		"  AND NOT EXISTS (SELECT 1 FROM resources                 WHERE resources.org_unit_data_id                 = org_unit_data.id) " +
		"  AND NOT EXISTS (SELECT 1 FROM overtime_data             WHERE overtime_data.org_unit_data_id             = org_unit_data.id) " +
		"  AND NOT EXISTS (SELECT 1 FROM org_unit_train_node_type     WHERE org_unit_train_node_type.org_unit_data_id     = org_unit_data.id) " +
		"  AND NOT EXISTS (SELECT 1 FROM org_unit_workspace_node_type WHERE org_unit_workspace_node_type.org_unit_data_id = org_unit_data.id)"
)

func InitCleaner(stopCh <-chan struct{}) {
	conf := runai_config.Get[*org_unit_config.DBCleanerConfig]()
	if !conf.CleanerEnabled {
		logging.Infof("org-unit DB cleaner is disabled")
		return
	}
	if conf.SoftDeletedDaysToLive <= 0 {
		logging.Errorf("org-unit DB cleaner: soft_deleted_days_to_live must be > 0, got %d; cleaner will not run",
			conf.SoftDeletedDaysToLive)
		return
	}
	logging.Infof("Initializing org-unit DB cleaner: interval=%dh, ttl=%dd",
		conf.CleanerIntervalHours, conf.SoftDeletedDaysToLive)

	dao_utils.InitCleaner(stopCh, dao_utils.DB(), cronJobName, conf.CleanerIntervalHours,
		BuildCleanerQueries(conf.SoftDeletedDaysToLive))
}

func BuildCleanerQueries(daysToLive int) []string {
	if daysToLive <= 0 {
		return nil
	}
	return []string{
		fmt.Sprintf(deleteResourcesQueryFormat, daysToLive),
		fmt.Sprintf(deleteProjectsQueryFormat, daysToLive),
		fmt.Sprintf(deleteDepartmentsQueryFormat, daysToLive),
		fmt.Sprintf(deleteOvertimeDataQueryFormat, daysToLive),
		fmt.Sprintf(deleteOrgUnitTrainNodeTypeQueryFormat, daysToLive),
		fmt.Sprintf(deleteOrgUnitWorkspaceNodeTypeQueryFormat, daysToLive),
		fmt.Sprintf(deleteOrgUnitDataQueryFormat, daysToLive),
	}
}
