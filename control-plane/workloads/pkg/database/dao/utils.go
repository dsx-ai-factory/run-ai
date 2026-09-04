/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

func ConstructDeleteQueryForProjectCluster(tx orm.DB, tableName string, tenantId uint, clusterId string, projectId string, from *uuid.UUID, to *uuid.UUID, excludingIds []string, workloadId string) orm.DB {
	return constructDeleteQuery(tx, tableName, tenantId, map[string]string{"cluster_id": clusterId, "project_id": projectId, "workload_id": workloadId}, from, to, excludingIds)
}

func constructDeleteQuery(tx orm.DB, tableName string, tenantId uint, columnFilters map[string]string, from *uuid.UUID, to *uuid.UUID, excludingIds []string) orm.DB {
	deleteQuery := tx.Table(tableName).
		Where("tenant_id = ?", tenantId)

	for filterByKey, filterByValue := range columnFilters {
		if filterByKey != "" && filterByValue != "" {
			deleteQuery = deleteQuery.Where(fmt.Sprintf("%s = ?", filterByKey), filterByValue)
		}
	}

	if from != nil {
		deleteQuery = deleteQuery.
			Where("id >= ?", from)
	}

	if to != nil {
		deleteQuery = deleteQuery.
			Where("id <= ?", to)
	}

	if len(excludingIds) > 0 {
		deleteQuery = deleteQuery.
			Not("id IN ?", excludingIds)
	}

	return deleteQuery
}

func createMapFromSlice(items []string) map[string]bool {
	newMap := map[string]bool{}
	for _, item := range items {
		newMap[item] = true
	}

	return newMap
}
