package dao

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func ConstructDeleteQueryForProjectCluster(tx *gorm.DB, tableName string, tenantId uint, clusterId string, projectId string, from *uuid.UUID, to *uuid.UUID, excludingIds []string, workloadId string) *gorm.DB {
	return constructDeleteQuery(tx, tableName, tenantId, map[string]string{"cluster_id": clusterId, "project_id": projectId, "workload_id": workloadId}, from, to, excludingIds)
}

func constructDeleteQuery(tx *gorm.DB, tableName string, tenantId uint, columnFilters map[string]string, from *uuid.UUID, to *uuid.UUID, excludingIds []string) *gorm.DB {
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
