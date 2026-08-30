package database

import (
	"github.com/run-ai/runai/control-plane/org-units-migrator/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

func GetTenantClusterVersions(tenantId int32) (map[string]string, error) {
	clusters := []models.Cluster{}

	result := globalClient.
		Where("tenant_id = ?", tenantId).
		Find(&clusters)
	if result.Error != nil {
		logging.Errorf("Failed to query tenant clusters for tenant %d. err: %v", tenantId, result.Error)
		return nil, result.Error
	}

	mapClusterIdToVersion := map[string]string{}

	for _, cluster := range clusters {
		mapClusterIdToVersion[cluster.Id] = cluster.Version
	}

	return mapClusterIdToVersion, nil
}
